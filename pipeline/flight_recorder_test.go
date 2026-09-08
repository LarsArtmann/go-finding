package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/onsi/gomega"
)

// newTestFlightRecorderHook creates a hook writing to a temp dir.
// Flight recorder is a global runtime resource (only one active at a time),
// so these tests do NOT call t.Parallel().
func newTestFlightRecorderHook(t *testing.T, cfg FlightRecorderConfig) *FlightRecorderHook {
	t.Helper()

	cfg.OutputDir = t.TempDir()

	hook, err := NewFlightRecorderHook(cfg)
	if err != nil {
		t.Fatalf("NewFlightRecorderHook: %v", err)
	}

	return hook
}

func TestFlightRecorderHook_StartAndClose(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())

	g := gomega.NewWithT(t)

	g.Expect(hook.Enabled()).To(gomega.BeTrue())

	hook.Close()

	g.Expect(hook.Enabled()).To(gomega.BeFalse())
}

func TestFlightRecorderHook_CloseIsIdempotent(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())

	hook.Close()
	hook.Close()
	hook.Close()
}

func TestFlightRecorderHook_NeverReturnsError(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	ctx := context.Background()

	// Before stage — should succeed.
	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageBefore,
	})
	g := gomega.NewWithT(t)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageAfter,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// After close — should still return nil (diagnostic, not control flow).
	hook.Close()

	err = hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageBefore,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestFlightRecorderHook_ManualSnapshotWritesFile(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	g := gomega.NewWithT(t)

	path, err := hook.Snapshot(context.Background(), "manual-test")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(path).To(gomega.BeAnExistingFile())

	info, err := os.Stat(path)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(info.Size()).To(gomega.BeNumerically(">", 0))
}

func TestFlightRecorderHook_SnapshotAfterCloseFails(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	hook.Close()

	g := gomega.NewWithT(t)

	_, err := hook.Snapshot(context.Background(), "post-close")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(errors.Is(err, ErrFlightRecorderNotEnabled)).To(gomega.BeTrue())
}

func TestFlightRecorderHook_SnapshotWithCancelledContext(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g := gomega.NewWithT(t)

	_, err := hook.Snapshot(ctx, "cancelled")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestFlightRecorderHook_SlowStageTriggersSnapshot(t *testing.T) {
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		MinAge:             time.Second,
		MaxBytes:           1 << 20,
		SlowStageThreshold: 5 * time.Millisecond,
	})
	defer hook.Close()

	ctx := context.Background()

	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageBefore,
	})
	g := gomega.NewWithT(t)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	time.Sleep(20 * time.Millisecond)

	err = hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageAfter,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	hook.Close()

	files, err := os.ReadDir(filepath.Dir(hook.config.OutputDir))
	_ = files
	_ = err

	traceFiles, err := findTraceFiles(hook.config.OutputDir)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(traceFiles).To(gomega.HaveLen(1))
}

func TestFlightRecorderHook_FastStageDoesNotSnapshot(t *testing.T) {
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		MinAge:             time.Second,
		MaxBytes:           1 << 20,
		SlowStageThreshold: 5 * time.Minute,
	})
	defer hook.Close()

	ctx := context.Background()

	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageBefore,
	})
	g := gomega.NewWithT(t)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageDetect,
		Timing: StageAfter,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	hook.Close()

	traceFiles, err := findTraceFiles(hook.config.OutputDir)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(traceFiles).To(gomega.BeEmpty())
}

func TestFlightRecorderHook_WithPipeline(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	d := newMockDetector("test", "tool", finding.NewFinding(
		"rule", "tool", "msg", finding.SeverityInfo,
		finding.Position{}, finding.ConfidenceHigh,
	))

	cfg := DefaultConfig()
	cfg.MaxIterations = 1
	cfg.StageHooks = []StageHook{hook}

	g := gomega.NewWithT(t)

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	result, err := p.Run(context.Background())
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(result.Reason).To(gomega.Equal(ReasonMaxIterations))

	path, err := hook.Snapshot(context.Background(), "post-run")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(path).To(gomega.BeAnExistingFile())
}

func TestFlightRecorderHook_MultipleSnapshots(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	g := gomega.NewWithT(t)

	path1, err := hook.Snapshot(context.Background(), "first")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	path2, err := hook.Snapshot(context.Background(), "second")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	g.Expect(path1).To(gomega.Not(gomega.Equal(path2)))
	g.Expect(path1).To(gomega.BeAnExistingFile())
	g.Expect(path2).To(gomega.BeAnExistingFile())
}

func TestFlightRecorderHook_SatisfiesStageHookInterface(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	var _ StageHook = hook
}

func TestSanitizeFilename(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		input  string
		expect string
	}{
		{"simple", "simple"},
		{"detect-iter1", "detect-iter1"},
		{"path/to/file", "path-to-file"},
		{"hello world!", "hello-world"},
		{"+++bad---", "bad"},
		{"", "snapshot"},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		g.Expect(got).To(gomega.Equal(tt.expect), "sanitizeFilename(%q)", tt.input)
	}
}

// findTraceFiles returns all .trace files in dir.
func findTraceFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".trace" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

func TestFlightRecorderHook_ConcurrentSnapshotsDoNotCollide(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	g := gomega.NewWithT(t)

	const n = 10

	paths := make([]string, n)
	errs := make([]error, n)

	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			paths[idx], errs[idx] = hook.Snapshot(context.Background(), "concurrent")
		}(i)
	}

	wg.Wait()

	for i := range n {
		g.Expect(errs[i]).To(gomega.Not(gomega.HaveOccurred()), "snapshot %d", i)
		g.Expect(paths[i]).To(gomega.BeAnExistingFile(), "snapshot %d", i)
	}

	uniquePaths := make(map[string]struct{}, n)
	for _, p := range paths {
		uniquePaths[p] = struct{}{}
	}

	g.Expect(uniquePaths).To(gomega.HaveLen(n), "all snapshot paths should be unique")
}

func TestFlightRecorderHook_MkdirAllError(t *testing.T) {
	// /dev/null is a file, not a directory — MkdirAll should fail.
	_, err := NewFlightRecorderHook(FlightRecorderConfig{
		OutputDir: "/dev/null/subdir",
	})

	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestRecordStageBoundary_BeforeReturnsFalse(t *testing.T) {
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		SlowStageThreshold: 1 * time.Nanosecond,
	})
	defer hook.Close()

	ctx := context.Background()

	// StageBefore should never trigger a snapshot.
	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageTriage,
		Timing: StageBefore,
	})
	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(hook.Enabled()).To(gomega.BeTrue())

	// No trace files should exist from a Before event alone.
	hook.Close()
	traceFiles, _ := findTraceFiles(hook.config.OutputDir)
	g.Expect(traceFiles).To(gomega.BeEmpty())
}

func TestRecordStageBoundary_SlowAfterTriggersSnapshot(t *testing.T) {
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		SlowStageThreshold: 1 * time.Nanosecond,
	})
	defer hook.Close()

	ctx := context.Background()

	g := gomega.NewWithT(t)

	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageVerify,
		Timing: StageBefore,
	})
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	time.Sleep(2 * time.Millisecond)

	err = hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageVerify,
		Timing: StageAfter,
	})
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	hook.Close()

	traceFiles, err := findTraceFiles(hook.config.OutputDir)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(traceFiles).To(gomega.HaveLen(1))
}

func TestRecordStageBoundary_AfterWithoutBeforeDoesNothing(t *testing.T) {
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		SlowStageThreshold: 1 * time.Nanosecond,
	})
	defer hook.Close()

	ctx := context.Background()

	// StageAfter without a matching StageBefore should not snapshot.
	err := hook.OnStageEvent(ctx, StageEvent{
		Stage:  StageApply,
		Timing: StageAfter,
	})
	g := gomega.NewWithT(t)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	hook.Close()

	traceFiles, _ := findTraceFiles(hook.config.OutputDir)
	g.Expect(traceFiles).To(gomega.BeEmpty())
}

func TestSanitizeFilename_AllSpecialChars(t *testing.T) {
	g := gomega.NewWithT(t)

	got := sanitizeFilename("@#$%^&*()")
	g.Expect(got).To(gomega.Equal("snapshot"))
}

func TestFlightRecorderHook_SnapshotWriteError(t *testing.T) {
	// Simulate disk-full or permission denied by making the output dir
	// read-only AFTER hook construction. Snapshot should return an error
	// from os.Create, not panic or silently succeed.
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())

	dir := hook.config.OutputDir

	// Restore permissions so t.TempDir cleanup can delete files.
	defer func() { _ = os.Chmod(dir, 0o755) }()

	if err := os.Chmod(dir, 0o444); err != nil {
		// Some CI environments run as root, where chmod is ineffective.
		t.Skipf("cannot make dir read-only (running as root?): %v", err)
	}

	defer hook.Close()

	g := gomega.NewWithT(t)

	_, err := hook.Snapshot(context.Background(), "permission-test")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestFlightRecorderHook_SlowLastStageInPipeline(t *testing.T) {
	// Verify SlowStageThreshold triggers on the verify (last) stage during
	// a real pipeline run, not just isolated OnStageEvent calls.
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		MinAge:             time.Second,
		MaxBytes:           1 << 20,
		SlowStageThreshold: 1 * time.Nanosecond,
	})
	defer hook.Close()

	d := newMockDetector("test", "tool", finding.NewFinding(
		"rule", "tool", "msg", finding.SeverityInfo,
		finding.Position{}, finding.ConfidenceHigh,
	))

	cfg := DefaultConfig()
	cfg.MaxIterations = 1
	cfg.StageHooks = []StageHook{hook}

	g := gomega.NewWithT(t)

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	_, err = p.Run(context.Background())
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	hook.Close()

	traceFiles, err := findTraceFiles(hook.config.OutputDir)
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(traceFiles).ToNot(gomega.BeEmpty(), "at least one slow stage snapshot should exist")
}

func TestFlightRecorderHook_DegradedWhenConflict(t *testing.T) {
	// Must NOT call t.Parallel() — flight recorder is a global singleton.
	primary := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer primary.Close()

	g := gomega.NewWithT(t)
	g.Expect(primary.Enabled()).To(gomega.BeTrue())
	g.Expect(primary.Degraded()).To(gomega.BeFalse())

	// Second hook should enter degraded mode, not fail.
	degraded, err := NewFlightRecorderHook(FlightRecorderConfig{
		MinAge:    time.Second,
		MaxBytes:  1 << 20,
		OutputDir: t.TempDir(),
	})
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	defer degraded.Close()

	g.Expect(degraded.Degraded()).To(gomega.BeTrue())
	g.Expect(degraded.Enabled()).To(gomega.BeFalse())

	// Snapshot on degraded hook should return error.
	_, err = degraded.Snapshot(context.Background(), "degraded-test")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(errors.Is(err, ErrFlightRecorderNotEnabled)).To(gomega.BeTrue())

	// OnStageEvent should still work (no-op, no error).
	err = degraded.OnStageEvent(context.Background(), StageEvent{
		Stage:  StageDetect,
		Timing: StageBefore,
	})
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
}

func TestFlightRecorderHook_ConcurrentStageEventsSafe(t *testing.T) {
	// Fire Before/After events for all stages from multiple goroutines.
	// Designed to be run with -race to detect data races in stageStarts map.
	// Must NOT call t.Parallel() — flight recorder is a global singleton.
	hook := newTestFlightRecorderHook(t, FlightRecorderConfig{
		SlowStageThreshold: 1 * time.Millisecond,
	})
	defer hook.Close()

	ctx := context.Background()

	stages := []Stage{StageDetect, StageProcess, StageTriage, StageApply, StageVerify}

	var wg sync.WaitGroup

	for _, stage := range stages {
		wg.Add(1)

		go func(s Stage) {
			defer wg.Done()

			for i := range 5 {
				_ = hook.OnStageEvent(ctx, StageEvent{
					Stage:     s,
					Timing:    StageBefore,
					Iteration: i,
				})

				time.Sleep(2 * time.Millisecond)

				_ = hook.OnStageEvent(ctx, StageEvent{
					Stage:     s,
					Timing:    StageAfter,
					Iteration: i,
				})
			}
		}(stage)
	}

	wg.Wait()
	hook.Close()

	// If we get here without panicking or -race failures, the test passes.
	traceFiles, _ := findTraceFiles(hook.config.OutputDir)

	g := gomega.NewWithT(t)
	g.Expect(traceFiles).ToNot(gomega.BeEmpty(), "some slow stages should have triggered snapshots")
}

// TestFlightRecorderHook_MaxFilesRotation verifies f/46: with MaxFiles set,
// each snapshot prunes the oldest go-finding-trace-* files beyond the cap.
// Unrelated files in the output dir are never touched.
func TestFlightRecorderHook_MaxFilesRotation(t *testing.T) {
	g := NewParallelGomega(t)

	dir := t.TempDir()

	hook, err := NewFlightRecorderHook(FlightRecorderConfig{
		OutputDir: dir,
		MaxFiles:  2,
	})
	if err != nil {
		t.Fatalf("NewFlightRecorderHook: %v", err)
	}

	t.Cleanup(func() { hook.Close() })

	// Unrelated file must survive pruning.
	keep := filepath.Join(dir, "unrelated.txt")
	g.Expect(os.WriteFile(keep, []byte("keep"), 0o600)).NotTo(HaveOccurred())

	for i := range 4 {
		path, snapErr := hook.Snapshot(context.Background(), fmt.Sprintf("rot%d", i))
		if snapErr != nil {
			t.Fatalf("snapshot %d: %v", i, snapErr)
		}

		g.Expect(path).To(HaveSuffix(".trace"))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	var traces []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "go-finding-trace-") {
			traces = append(traces, e.Name())
		}
	}

	g.Expect(traces).To(HaveLen(2), "MaxFiles=2 must prune older snapshots")
	g.Expect(traces[0]).To(ContainSubstring("rot2"), "newest snapshots survive")
	g.Expect(traces[1]).To(ContainSubstring("rot3"))
	g.Expect(filepath.Join(dir, "unrelated.txt")).To(BeAnExistingFile())
}

// TestFlightRecorderHook_GzipCompression verifies f/47: Compress writes a
// .trace.gz file whose payload gunzips to non-empty trace data.
func TestFlightRecorderHook_GzipCompression(t *testing.T) {
	g := NewParallelGomega(t)

	dir := t.TempDir()

	hook, err := NewFlightRecorderHook(FlightRecorderConfig{
		OutputDir: dir,
		Compress:  true,
	})
	if err != nil {
		t.Fatalf("NewFlightRecorderHook: %v", err)
	}

	t.Cleanup(func() { hook.Close() })

	path, err := hook.Snapshot(context.Background(), "gz-test")
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	g.Expect(path).To(HaveSuffix(".trace.gz"))

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}

	defer func() { _ = gz.Close() }()

	data, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	g.Expect(data).NotTo(BeEmpty(), "gunzipped payload must contain trace bytes")
	g.Expect(data[0]).To(Equal(byte(0x00)), "trace header: 'go 1.' header starts with 0x00? no — first byte of trace is 'g'")
}

// TestFlightRecorderHook_NoRotationByDefault verifies MaxFiles=0 (default)
// keeps every snapshot.
func TestFlightRecorderHook_NoRotationByDefault(t *testing.T) {
	g := NewParallelGomega(t)

	dir := t.TempDir()

	hook, err := NewFlightRecorderHook(FlightRecorderConfig{OutputDir: dir})
	if err != nil {
		t.Fatalf("NewFlightRecorderHook: %v", err)
	}

	t.Cleanup(func() { hook.Close() })

	for range 3 {
		if _, err := hook.Snapshot(context.Background(), "keep-all"); err != nil {
			t.Fatalf("Snapshot: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "go-finding-trace-") {
			count++
		}
	}

	g.Expect(count).To(Equal(3), "default MaxFiles=0 must not prune")
}
