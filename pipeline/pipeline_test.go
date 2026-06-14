package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// pipelineTestFinding creates a Finding with test values at the specified column.
func pipelineTestFinding(column int) finding.Finding {
	return finding.Finding{
		ID:          "test:1",
		Rule:        "test-rule",
		ToolName:    "test-tool",
		Message:     "test message",
		Severity:    finding.SeverityWarning,
		Position:    finding.Pos("test.go", 1, column),
		FixStrategy: finding.FixStrategySuggest,
		AfterCode:   "fixed()",
	}
}

// TestPipelineRun_NoFindings tests that pipeline completes when no findings.
func TestPipelineRun_NoFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = false
	detector := &mockDetector{name: "test", findings: nil}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Stable()).To(BeTrue())
	g.Expect(result.Reason).To(Equal(ReasonStable))
	g.Expect(result.TotalIterations).To(Equal(1))
}

// TestPipelineRun_WithFindings tests pipeline with findings.
func TestPipelineRun_WithFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = false
	config.MaxIterations = 3 // Limit to avoid running 5 iterations

	findings := []finding.Finding{pipelineTestFinding(1)}

	detector := &mockDetector{name: "test", findings: findings}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Stable()).To(BeFalse())
	g.Expect(result.TotalIterations).To(Equal(config.MaxIterations))

	g.Expect(result.Iterations).To(HaveLen(config.MaxIterations))

	iter := result.Iterations[0]
	g.Expect(iter.FindingsFound).To(Equal(1))

	g.Expect(iter.SuggestFixes).To(Equal(1))

	g.Expect(iter.SuggestedFindings()).To(HaveLen(1))
}

// TestPipelineRun_DetectorError tests error handling from detector.
func TestPipelineRun_DetectorError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = false

	expectedErr := errors.New("detector failed")
	detector := &mockDetector{name: "test", err: expectedErr}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	_, err = p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, expectedErr)).To(BeTrue())
}

// TestDetectSequential_ContextCancellation verifies that detectSequential
// returns an error when the context is cancelled.
func TestDetectSequential_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	p := &Pipeline{
		detectors: []Detector{&mockDetector{name: "slow", delay: time.Second}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.detectSequential(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

// TestPipelineRun_ContextCancellation tests context cancellation.
func TestPipelineRun_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = false

	// Detector that takes time
	detector := &mockDetector{
		name:  "slow",
		delay: 100 * time.Millisecond,
	}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
	g.Expect(result.Reason).To(Equal(ReasonCancelled))
}

// TestPipelineRun_Timeout tests pipeline timeout.
func TestPipelineRun_Timeout(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = false
	config.Timeout = 50 * time.Millisecond

	detector := slowTestDetector("slow", 500*time.Millisecond)

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	result, err := p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
	g.Expect(result.Reason).To(Equal(ReasonTimeout))
}

// TestPipelineRun_MaxIterations tests max iteration limit.
func TestPipelineRun_MaxIterations(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.MaxIterations = 2
	config.ParallelDetectors = false

	// Findings that won't be auto-fixed (suggest strategy)
	findings := []finding.Finding{pipelineTestFinding(0)}

	detector := &mockDetector{name: "test", findings: findings}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.TotalIterations).To(Equal(config.MaxIterations))
	g.Expect(result.Reason).To(Equal(ReasonMaxIterations))
}

// TestPipelineRun_Parallel tests parallel detection.
func TestPipelineRun_Parallel(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = true

	d1 := &mockDetector{
		name: "d1",
		findings: []finding.Finding{
			testFinding("1", "r1", "t1", "m1", finding.SeverityInfo, "a.go"),
		},
	}
	d2 := &mockDetector{
		name: "d2",
		findings: []finding.Finding{
			{
				ID:       "2",
				Rule:     "r2",
				ToolName: "t2",
				Message:  "m2",
				Severity: finding.SeverityWarning,
				Position: finding.Position{File: "b.go"},
			},
		},
	}

	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Iterations).NotTo(BeEmpty())

	assertFindingsFound(t, result, 2, "expected 2 findings")
}

// TestDetectParallel_SuppressionConsistency verifies that parallel and
// sequential detection produce identical results when findings are suppressed.
