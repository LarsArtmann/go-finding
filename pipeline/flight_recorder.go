package pipeline

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/trace"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrFlightRecorderNotEnabled is returned when Snapshot is called on a
// closed or disabled flight recorder.
var ErrFlightRecorderNotEnabled = errors.New("flight recorder: not enabled or already closed")

const (
	defaultFRMinAge   = 30 * time.Second
	defaultFRMaxBytes = 4 << 20 // 4 MiB
)

// FlightRecorderConfig configures the pipeline's execution trace flight recorder.
//
// The flight recorder continuously buffers Go runtime execution trace data
// in memory. At any point you can request a snapshot of the buffered trace
// to capture the last few seconds of execution — invaluable for diagnosing
// slow detectors, fix contention, or unexpected pipeline stalls.
//
// See: https://go.dev/blog/flight-recorder
type FlightRecorderConfig struct {
	// MinAge is how long trace data is reliably retained in the buffer.
	// Set to ~2x the longest stage duration you want to debug.
	// Default: 30s.
	MinAge time.Duration

	// MaxBytes limits the in-memory trace buffer size.
	// Default: 4 MiB (4 << 20).
	MaxBytes uint64

	// SlowStageThreshold, when non-zero, triggers an automatic snapshot
	// when any single stage execution exceeds this duration.
	// Default: 0 (manual snapshot only).
	SlowStageThreshold time.Duration

	// OutputDir is where snapshot .trace files are written.
	// Default: os.TempDir().
	OutputDir string

	// MaxFiles caps how many snapshot files are retained in OutputDir.
	// After each successful snapshot, the oldest go-finding-trace-* files
	// beyond this cap are deleted. 0 means unlimited (default) — long
	// -trace-slow runs should set it to avoid unbounded disk growth.
	MaxFiles int

	// Compress wraps snapshot output in gzip (.trace.gz suffix).
	// Default: false (plain .trace files open directly with `go tool trace`).
	// Compressed files must be gunzipped first.
	Compress bool

	// Logger receives snapshot lifecycle events (written, failed).
	// If nil, events are silently dropped.
	Logger *slog.Logger
}

// DefaultFlightRecorderConfig returns sensible defaults for pipeline
// flight recording.
func DefaultFlightRecorderConfig() FlightRecorderConfig {
	//nolint:exhaustruct_v5 // SlowStageThreshold and Logger are intentionally zero/nil by default
	return FlightRecorderConfig{
		MinAge:    defaultFRMinAge,
		MaxBytes:  defaultFRMaxBytes,
		OutputDir: os.TempDir(),
	}
}

// FlightRecorderHook implements StageHook to capture Go execution trace
// snapshots when pipeline stages exceed configurable thresholds.
//
// It wraps Go 1.25's runtime/trace.FlightRecorder, which continuously
// buffers trace data in memory and lets you snapshot the last few seconds
// on demand.
//
// Create with NewFlightRecorderHook, register via Config.StageHooks,
// and call Close after the pipeline finishes to release the recorder.
//
// OnStageEvent NEVER returns an error — trace collection is purely
// diagnostic and must not affect pipeline control flow.
type FlightRecorderHook struct {
	fr     *trace.FlightRecorder
	config FlightRecorderConfig

	mu            sync.Mutex
	stageStarts   map[Stage]time.Time
	snapshotCount int
	closed        bool
	degraded      bool
	snapshotWg    sync.WaitGroup

	// writeMu serializes snapshot file operations: WriteTo calls (which are
	// NOT safe for concurrent use) and prune listing/deletion, so concurrent
	// snapshots cannot race each other's rotation.
	writeMu sync.Mutex
}

// NewFlightRecorderHook creates and starts a flight recorder.
//
// If another flight recorder is already active (Go's runtime allows only one
// at a time), the hook enters degraded mode: it returns successfully but all
// snapshot operations are silently skipped. Check Degraded() to detect this.
//
// Returns an error if the output directory cannot be created or if the recorder
// fails to start for reasons other than a singleton conflict.
func NewFlightRecorderHook(config FlightRecorderConfig) (*FlightRecorderHook, error) {
	if config.MinAge <= 0 {
		config.MinAge = defaultFRMinAge
	}

	if config.MaxBytes == 0 {
		config.MaxBytes = defaultFRMaxBytes
	}

	if config.OutputDir == "" {
		config.OutputDir = os.TempDir()
	}

	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create trace output dir %s: %w", config.OutputDir, err)
	}

	recorder := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   config.MinAge,
		MaxBytes: config.MaxBytes,
	})

	if err := recorder.Start(); err != nil {
		if strings.Contains(err.Error(), "flight recorder already enabled") {
			if config.Logger != nil {
				config.Logger.Warn(
					"flight recorder degraded: another recorder is already active; snapshots will be skipped",
					slog.String("error", err.Error()),
				)
			}

			return &FlightRecorderHook{ //nolint:exhaustruct_v5 // zero-valued fields are intentional
				fr:          recorder,
				config:      config,
				degraded:    true,
				stageStarts: make(map[Stage]time.Time),
			}, nil
		}

		return nil, fmt.Errorf("start flight recorder: %w", err)
	}

	return &FlightRecorderHook{ //nolint:exhaustruct_v5 // mu, snapshotCount, closed, degraded, snapshotWg are zero-valued intentionally
		fr:          recorder,
		config:      config,
		stageStarts: make(map[Stage]time.Time),
	}, nil
}

// OnStageEvent implements StageHook. It tracks per-stage durations and
// auto-snapshots when SlowStageThreshold is exceeded.
//
// This method never returns an error — trace collection is diagnostic
// and must not affect pipeline control flow. Snapshot I/O errors are
// logged to the configured Logger.
func (h *FlightRecorderHook) OnStageEvent(ctx context.Context, event StageEvent) error {
	shouldSnapshot := h.recordStageBoundary(event)

	if shouldSnapshot {
		h.asyncSnapshot(ctx, event)
	}

	return nil
}

// recordStageBoundary updates stage timing state and returns true if the
// stage exceeded the slow threshold.
func (h *FlightRecorderHook) recordStageBoundary(event StageEvent) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed || h.degraded {
		return false
	}

	if event.Timing == StageBefore {
		h.stageStarts[event.Stage] = time.Now()

		return false
	}

	if event.Timing != StageAfter {
		return false
	}

	start, ok := h.stageStarts[event.Stage]
	if !ok {
		return false
	}

	delete(h.stageStarts, event.Stage)

	if h.config.SlowStageThreshold <= 0 {
		return false
	}

	return time.Since(start) > h.config.SlowStageThreshold
}

func (h *FlightRecorderHook) nextSnapshotNumber() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	num := h.snapshotCount
	h.snapshotCount++

	return num
}

// asyncSnapshot captures a trace snapshot to a file in a background goroutine.
// Uses snapshotWg so Close can wait for in-flight snapshots.
// The snapshot write uses context.Background() (not the pipeline ctx) because
// the pipeline's run context may be cancelled before the goroutine executes
// (defer cancel() in Pipeline.Run fires on return). Diagnostic snapshots must
// complete regardless of pipeline lifecycle.
func (h *FlightRecorderHook) asyncSnapshot(ctx context.Context, event StageEvent) {
	num := h.nextSnapshotNumber()

	h.snapshotWg.Go(func() {
		path, err := h.writeSnapshot(context.Background(), num, fmt.Sprintf("%s-iter%d", event.Stage, event.Iteration))
		if err != nil {
			h.log(ctx, slog.LevelError, "flight recorder snapshot failed", slog.String("error", err.Error()))

			return
		}

		h.log(
			ctx, slog.LevelInfo, "flight recorder snapshot written",
			slog.String("path", path),
			slog.String("stage", string(event.Stage)),
			slog.Int("iteration", event.Iteration),
		)
	})
}

// Snapshot captures the current flight recorder buffer to a file and
// returns the file path. The reason string is included in the filename
// for identification. The context is checked for cancellation before
// writing; if cancelled, no file is created.
//
// Returns ErrFlightRecorderNotEnabled if the recorder is closed or disabled.
// Returns context.Cause(ctx) if the context is cancelled before writing.
// Returns a wrapped os/Create or WriteTo error if the filesystem fails
// (e.g., disk full, permission denied). In such cases the partially-written
// file is closed but may remain on disk with incomplete data.
func (h *FlightRecorderHook) Snapshot(ctx context.Context, reason string) (string, error) {
	h.mu.Lock()

	if h.closed || h.degraded || !h.fr.Enabled() {
		h.mu.Unlock()

		return "", ErrFlightRecorderNotEnabled
	}

	h.mu.Unlock()
	num := h.nextSnapshotNumber()

	return h.writeSnapshot(ctx, num, reason)
}

// writeSnapshot creates a trace file and writes the buffered trace data.
// The writeMu serializes concurrent WriteTo calls because
// runtime/trace.FlightRecorder.WriteTo is NOT safe for concurrent use.
// The context is checked for cancellation before the potentially slow WriteTo call.
func (h *FlightRecorderHook) writeSnapshot(ctx context.Context, num int, reason string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("snapshot cancelled before write: %w", err)
	}

	suffix := ".trace"
	if h.config.Compress {
		suffix = ".trace.gz"
	}

	filename := fmt.Sprintf("go-finding-trace-%03d-%s%s", num, sanitizeFilename(reason), suffix)
	path := filepath.Join(h.config.OutputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create trace file %s: %w", path, err)
	}

	if h.config.Compress {
		gz := gzip.NewWriter(f)

		h.writeMu.Lock()
		_, writeErr := h.fr.WriteTo(gz)
		h.writeMu.Unlock()

		if writeErr == nil {
			writeErr = gz.Close()
		} else {
			_ = gz.Close()
		}

		if writeErr != nil {
			_ = f.Close()

			return "", fmt.Errorf("write trace to %s: %w", path, writeErr)
		}
	} else {
		h.writeMu.Lock()
		_, writeErr := h.fr.WriteTo(f)
		h.writeMu.Unlock()

		if writeErr != nil {
			_ = f.Close()

			return "", fmt.Errorf("write trace to %s: %w", path, writeErr)
		}
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close trace file %s: %w", path, err)
	}

	h.pruneSnapshots(ctx)

	return path, nil
}

// traceSnapshotPrefix/traceSnapshotSuffixes identify this hook's snapshot
// files during pruning. Only files matching BOTH are ever deleted.
const traceSnapshotPrefix = "go-finding-trace-"

// pruneSnapshots deletes the oldest snapshot files beyond config.MaxFiles.
// Best-effort: deletion errors are logged, never returned — pruning must not
// fail a successful snapshot. A no-op when MaxFiles is unset.
//
// Holds writeMu for the entire list-and-delete pass: two concurrent snapshots
// would otherwise prune overlapping file sets simultaneously (double-remove
// ENOENT noise) and modtime ties could prune nondeterministically.
func (h *FlightRecorderHook) pruneSnapshots(ctx context.Context) {
	if h.config.MaxFiles <= 0 {
		return
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	entries, err := os.ReadDir(h.config.OutputDir)
	if err != nil {
		h.log(ctx, slog.LevelWarn, "flight recorder prune: read output dir failed", slog.String("error", err.Error()))

		return
	}

	type snapshotFile struct {
		path    string
		modTime time.Time
	}

	var snapshots []snapshotFile

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasPrefix(name, traceSnapshotPrefix) {
			continue
		}

		if !strings.HasSuffix(name, ".trace") && !strings.HasSuffix(name, ".trace.gz") {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		snapshots = append(
			snapshots,
			snapshotFile{path: filepath.Join(h.config.OutputDir, name), modTime: info.ModTime()},
		)
	}

	if len(snapshots) <= h.config.MaxFiles {
		return
	}

	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].modTime.Before(snapshots[j].modTime) })

	for _, old := range snapshots[:len(snapshots)-h.config.MaxFiles] {
		if err := os.Remove(old.path); err != nil {
			h.log(ctx, slog.LevelWarn, "flight recorder prune: delete failed",
				slog.String("path", old.path), slog.String("error", err.Error()))

			continue
		}

		h.log(ctx, slog.LevelInfo, "flight recorder pruned old snapshot", slog.String("path", old.path))
	}
}

// Degraded reports whether the hook is operating in degraded mode because
// another flight recorder was already active when NewFlightRecorderHook was
// called. In degraded mode, all snapshot operations are silently skipped.
// The hook is safe to register as a StageHook — it simply does nothing.
func (h *FlightRecorderHook) Degraded() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.degraded
}

// Enabled reports whether the flight recorder is active and capturing.
//
// Note: There is an inherent TOCTOU window between calling Enabled() and
// acting on the result — the recorder may be closed by another goroutine
// between the check and the use. Callers that need atomicity should use
// Snapshot() directly (which checks closed state under the mutex) rather
// than pre-checking with Enabled().
func (h *FlightRecorderHook) Enabled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return !h.degraded && !h.closed && h.fr.Enabled()
}

// Close stops the flight recorder and waits for any in-flight snapshots
// to complete. Safe to call multiple times.
func (h *FlightRecorderHook) Close() {
	h.mu.Lock()

	if h.closed {
		h.mu.Unlock()

		return
	}

	h.closed = true
	h.mu.Unlock()

	// Wait for in-flight async snapshots to finish writing before
	// stopping the recorder, since WriteTo and Stop access the same
	// internal FlightRecorder state.
	h.snapshotWg.Wait()

	h.fr.Stop()
}

func (h *FlightRecorderHook) log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if h.config.Logger != nil {
		h.config.Logger.LogAttrs(ctx, level, msg, attrs...)
	}
}

// sanitizeFilename replaces non-alphanumeric characters with hyphens.
func sanitizeFilename(s string) string {
	var b strings.Builder

	b.Grow(len(s))

	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	result := b.String()

	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	result = strings.Trim(result, "-")
	if result == "" {
		return "snapshot"
	}

	return result
}
