package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/trace"
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

	// Logger receives snapshot lifecycle events (written, failed).
	// If nil, events are silently dropped.
	Logger *slog.Logger
}

// DefaultFlightRecorderConfig returns sensible defaults for pipeline
// flight recording.
func DefaultFlightRecorderConfig() FlightRecorderConfig {
	//nolint:exhaustruct // SlowStageThreshold and Logger are intentionally zero/nil by default
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
	snapshotWg    sync.WaitGroup
}

// NewFlightRecorderHook creates and starts a flight recorder.
//
// Returns an error if the recorder cannot be started (e.g., another
// flight recorder is already active — only one may exist at a time).
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
		return nil, fmt.Errorf("start flight recorder: %w", err)
	}

	return &FlightRecorderHook{ //nolint:exhaustruct // mu, snapshotCount, closed, snapshotWg are zero-valued intentionally
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

	if h.closed {
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

// asyncSnapshot captures a trace snapshot to a file in a background goroutine.
// Uses snapshotWg so Close can wait for in-flight snapshots.
func (h *FlightRecorderHook) asyncSnapshot(ctx context.Context, event StageEvent) {
	h.mu.Lock()
	num := h.snapshotCount
	h.snapshotCount++
	h.mu.Unlock()

	h.snapshotWg.Go(func() {
		path, err := h.writeSnapshot(num, fmt.Sprintf("%s-iter%d", event.Stage, event.Iteration))
		if err != nil {
			h.log(ctx, slog.LevelError, "flight recorder snapshot failed", slog.String("error", err.Error()))

			return
		}

		h.log(ctx, slog.LevelInfo, "flight recorder snapshot written",
			slog.String("path", path),
			slog.String("stage", string(event.Stage)),
			slog.Int("iteration", event.Iteration),
		)
	})
}

// Snapshot captures the current flight recorder buffer to a file and
// returns the file path. The reason string is included in the filename
// for identification.
//
// Returns ErrFlightRecorderNotEnabled if the recorder is closed or disabled.
func (h *FlightRecorderHook) Snapshot(reason string) (string, error) {
	h.mu.Lock()

	if h.closed || !h.fr.Enabled() {
		h.mu.Unlock()

		return "", ErrFlightRecorderNotEnabled
	}

	num := h.snapshotCount
	h.snapshotCount++
	h.mu.Unlock()

	return h.writeSnapshot(num, reason)
}

// writeSnapshot creates a trace file and writes the buffered trace data.
func (h *FlightRecorderHook) writeSnapshot(num int, reason string) (string, error) {
	filename := fmt.Sprintf("go-finding-trace-%03d-%s.trace", num, sanitizeFilename(reason))
	path := filepath.Join(h.config.OutputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create trace file %s: %w", path, err)
	}

	if _, err := h.fr.WriteTo(f); err != nil {
		_ = f.Close()

		return "", fmt.Errorf("write trace to %s: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close trace file %s: %w", path, err)
	}

	return path, nil
}

// Enabled reports whether the flight recorder is active and capturing.
func (h *FlightRecorderHook) Enabled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return !h.closed && h.fr.Enabled()
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

	return strings.Trim(result, "-")
}
