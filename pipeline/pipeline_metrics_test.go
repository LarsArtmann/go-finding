package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func BenchmarkParallelDetection(b *testing.B) {
	for _, parallel := range []bool{false, true} {
		name := "sequential"
		if parallel {
			name = "parallel"
		}

		b.Run(name, func(b *testing.B) {
			config := DefaultConfig()
			config.ParallelDetectors = parallel

			// Create multiple slow detectors
			detectors := make([]Detector, 5)
			for i := range detectors {
				detectors[i] = &mockDetector{
					name:  "bench",
					delay: 1 * time.Millisecond,
					findings: []finding.Finding{
						testFinding("bench", "r", "t", "m", finding.SeverityInfo, "x.go"),
					},
				}
			}

			b.ResetTimer()

			for range b.N {
				p, err := New(config, b.TempDir(), detectors...)
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = p.Run(ctx)

				cancel()
			}
		})
	}
}

func TestIteration_Findings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []finding.Finding{
		{Message: "a", Severity: finding.SeverityError},
		{Message: "b", Severity: finding.SeverityWarning},
	}
	iter := Iteration{
		Number:   1,
		findings: findings,
		suggest:  []finding.Finding{{Message: "s", FixStrategy: finding.FixStrategySuggest}},
	}

	got := iter.Findings()
	g.Expect(got).To(HaveLen(2))

	assertFindingMessage := func(idx int, want string) {
		t.Helper()

		msg := got[idx].Message
		g.Expect(msg).To(Equal(want))
	}
	assertFindingMessage(0, "a")
	assertFindingMessage(1, "b")

	suggested := iter.SuggestedFindings()
	g.Expect(suggested).To(HaveLen(1))

	g.Expect(suggested[0].Message).To(Equal("s"))
}

func TestIteration_Findings_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	iter := Iteration{Number: 1}
	g.Expect(iter.Findings()).To(BeNil())

	g.Expect(iter.SuggestedFindings()).To(BeNil())
}

func TestIoErrorAt(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := ioErrorAt("read failed", os.ErrNotExist, "foo.go")

	var fe *finding.FindingError
	g.Expect(errors.As(err, &fe)).To(BeTrue())

	assertFindingErrorIO(t, fe, "foo.go")
}

func TestPipelineRun_PartialErrorsSurfaced(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	goodDetector := mockDetectorWithFinding("good", "F1", "r", "good", "m")
	badDetector := &mockDetector{name: "bad", err: errors.New("boom")}

	config := Config{
		MaxIterations:       1,
		GracefulDegradation: true,
	}

	p, err := New(config, t.TempDir(), goodDetector, badDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.PartialErrors).NotTo(BeNil())

	_, ok := result.PartialErrors["bad"]
	g.Expect(ok).To(BeTrue())

	_, ok = result.PartialErrors["good"]
	g.Expect(ok).To(BeFalse())
}

func TestPipelineRun_MetricsInResult(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	m := NewMetrics()
	config := Config{
		MaxIterations: 1,
		Metrics:       m,
	}

	detector := &mockDetector{name: "test", findings: nil}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Metrics.TotalDuration).NotTo(BeZero())

	g.Expect(result.Metrics.FixesApplied).To(Equal(0))
}

func TestPipelineRun_DirectFixStabilizes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fix := makeFixFinding("fix1", "old()", "new()", "fixme.go", 4)

	callCount := 0
	det := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		callCount++
		if callCount == 1 {
			return []finding.Finding{fix}, nil
		}

		return nil, nil
	})

	cfg := Config{MaxIterations: 5, ParallelDetectors: false}

	p, err := New(cfg, tmpDir, NamedDetectorFunc("tool", det))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Stable()).To(BeTrue())

	g.Expect(result.TotalIterations).To(Equal(2))

	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	g.Expect(string(content)).To(Equal(want))
}

func testContextErrorCancels(t *testing.T, parallel bool) {
	t.Helper()
	g := NewWithT(t)

	goodDetector := &mockDetector{
		name:     "good",
		findings: []finding.Finding{{ID: "f1", Rule: "r", ToolName: "t", Message: "m"}},
	}
	badDetector := &mockDetector{
		name: "bad",
		err:  context.Canceled,
	}

	config := DefaultConfig()
	config.GracefulDegradation = true
	config.ParallelDetectors = parallel

	p, err := New(config, t.TempDir(), goodDetector, badDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())

	g.Expect(result.PartialErrors).
		To(BeNil(), "context errors should not be stored as partial errors")
}

func TestDetectPartialSequential_ContextErrorPropagates(t *testing.T) {
	t.Parallel()
	testContextErrorCancels(t, false)
}

func TestDetectPartialParallel_ContextErrorPropagates(t *testing.T) {
	t.Parallel()
	testContextErrorCancels(t, true)
}

func TestVerify_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	slowDetector := &mockDetector{name: "slow", delay: time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Verify(ctx, []Detector{slowDetector}, nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

func TestIsContextError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"canceled", context.Canceled, true},
		{"deadline", context.DeadlineExceeded, true},
		{"wrapped canceled", fmt.Errorf("wrap: %w", context.Canceled), true},
		{"other error", errors.New("other"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsContextError(tt.err); got != tt.want {
				t.Errorf("IsContextError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
