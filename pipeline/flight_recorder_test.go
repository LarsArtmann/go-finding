package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

	path, err := hook.Snapshot("manual-test")
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

	_, err := hook.Snapshot("post-close")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(errors.Is(err, ErrFlightRecorderNotEnabled)).To(gomega.BeTrue())
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

	path, err := hook.Snapshot("post-run")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
	g.Expect(path).To(gomega.BeAnExistingFile())
}

func TestFlightRecorderHook_MultipleSnapshots(t *testing.T) {
	hook := newTestFlightRecorderHook(t, DefaultFlightRecorderConfig())
	defer hook.Close()

	g := gomega.NewWithT(t)

	path1, err := hook.Snapshot("first")
	g.Expect(err).To(gomega.Not(gomega.HaveOccurred()))

	path2, err := hook.Snapshot("second")
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
		{"", ""},
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
