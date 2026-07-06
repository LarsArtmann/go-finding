package pipeline

import (
	"context"
	"errors"
	"path/filepath"
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
func countingOnFix(ptr *int) func(finding.Finding, bool) {
	return func(_ finding.Finding, wasApplied bool) {
		if wasApplied {
			*ptr++
		}
	}
}

func collectingOnFix(ids *[]string) func(finding.Finding, bool) {
	return func(f finding.Finding, wasApplied bool) {
		if wasApplied {
			*ids = append(*ids, string(f.ID))
		}
	}
}

func directFix(id, rule, tool, msg, before, after, file string, line int) finding.Finding {
	return finding.Finding{
		ID:          finding.ID(id),
		Rule:        finding.RuleName(rule),
		ToolName:    finding.ToolName(tool),
		Message:     msg,
		BeforeCode:  before,
		AfterCode:   after,
		Position:    finding.Position{File: finding.FilePath(file), Line: line},
		FixStrategy: finding.FixStrategyDirect,
	}
}

// TestOnFix_FiresOnlyForAppliedFixes verifies C-1: OnFix callback fires
// exactly once per actually-applied fix, not once per safeFix.
func TestOnFix_FiresOnlyForAppliedFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	// File contains two occurrences.
	original := "package main\n\nfunc main() {\n\told()\n\told()\n}\n" //nolint:dupword // test fixture
	writeTestFile(t, testFile, []byte(original))

	// One fix that targets the second occurrence at line 5.
	fix := directFix("fix1", "r1", "tool", "replace old", "old()", "new()", "fixme.go", 5)

	var onFixCalls int

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFix:             countingOnFix(&onFixCalls),
	}

	det := mockDetWithFindings("tool", fix)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// One fix was applied.
	g.Expect(result.Iterations[0].Applied).To(Equal(1))
	// OnFix should have been called exactly once for the applied fix.
	g.Expect(onFixCalls).To(Equal(1))
}

// TestOnFix_SkipsUnappliedFixes verifies that when a fix is skipped
// (e.g., BeforeCode not found), OnFix is not called with wasApplied=true.
func TestOnFix_SkipsUnappliedFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	// Fix with BeforeCode that doesn't exist in the file.
	fix := directFix("fix1", "", "", "", "nonexistent", "replacement", "fixme.go", 1)

	var appliedCount int

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFix:             countingOnFix(&appliedCount),
	}

	det := mockDetWithFindings("tool", fix)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(appliedCount).To(Equal(0))
}

// TestOnFix_ReportsCorrectAppliedFindings verifies that when some fixes fail
// and others succeed, OnFix is called with the actually-applied findings, not
// just the first N fixes from the input slice.
func TestOnFix_ReportsCorrectAppliedFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	writeTestFile(t, testFile, []byte("package main\n\nfunc main() {\n\tfirst()\n\tsecond()\n}\n"))

	// fixA fails (BeforeCode not found), fixB succeeds.
	fixA := directFix(
		"fixA",
		"r1",
		"tool",
		"replace first",
		"nonexistent()",
		"newFirst()",
		"fixme.go",
		4,
	)
	fixB := directFix(
		"fixB",
		"r2",
		"tool",
		"replace second",
		"second()",
		"newSecond()",
		"fixme.go",
		5,
	)

	var appliedIDs []string

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFix:             collectingOnFix(&appliedIDs),
	}

	det := mockDetWithFindings("tool", fixA, fixB)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Iterations[0].Applied).To(Equal(1))
	g.Expect(appliedIDs).To(Equal([]string{"fixB"}))
}

// TestPipelineRun_ParallelDetectorError verifies that a detector error in
// parallel mode propagates correctly from detectParallel.
func TestPipelineRun_ParallelDetectorError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = true

	expectedErr := errors.New("parallel detector failed")
	d1 := &mockDetector{name: "ok", findings: []finding.Finding{{ID: "1"}}}
	d2 := &mockDetector{name: "fail", err: expectedErr}

	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("parallel detection")))
}

// TestApplyTriage_AllConflicts verifies that when all fixes are skipped by
// FilterConflictingFixes (no file info), applyTriage returns nil without
// calling applyDirectFixes.
func TestApplyTriage_AllConflicts(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()

	// Fixes without file info are skipped entirely by FilterConflictingFixes.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			Rule:        "r1",
			ToolName:    "t1",
			Message:     "m1",
			Position:    finding.Position{Line: 3, Column: 1}, // File is empty
			FixStrategy: finding.FixStrategyDirect,
			BeforeCode:  "func",
			AfterCode:   "fn",
		},
	}

	cfg := DefaultConfig()

	p, err := New(cfg, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	iter := Iteration{Number: 1}

	err = p.applyTriage(context.Background(), fixes, &iter)
	if err != nil {
		t.Fatalf("applyTriage: %v", err)
	}

	g.Expect(iter.Conflicts).To(Equal(1))
	g.Expect(iter.Applied).To(Equal(0))
}

// TestApplyTriage_ApplyError verifies that applyTriage returns an error when
// applyDirectFixes fails (e.g., target file does not exist).
func TestApplyTriage_ApplyError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()

	// A single fix pointing to a non-existent file.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			Rule:        "r1",
			ToolName:    "t1",
			Message:     "m1",
			Position:    finding.Position{File: "missing.go", Line: 1, Column: 1},
			FixStrategy: finding.FixStrategyDirect,
			BeforeCode:  "old",
			AfterCode:   "new",
		},
	}

	cfg := DefaultConfig()

	p, err := New(cfg, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	iter := Iteration{Number: 1}
	err = p.applyTriage(context.Background(), fixes, &iter)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("apply fixes")))
}
