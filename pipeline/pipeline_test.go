package pipeline

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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

	err = p.applyTriage(context.Background(), fixes, &iter, &PipelineResult{})
	if err != nil {
		t.Fatalf("applyTriage: %v", err)
	}

	g.Expect(iter.Conflicts).To(Equal(1))
	g.Expect(iter.Applied).To(Equal(0))
}

// TestApplyTriage_ApplyError verifies that applyTriage returns an error when
// applyDirectFixes fails (e.g., target file does not exist).
func TestApplyTriage_ApplyError(t *testing.T) {
	g := NewParallelGomega(t)

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
	err = p.applyTriage(context.Background(), fixes, &iter, &PipelineResult{})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("apply fixes")))
}

// TestPipeline_MetricsAvailableOnErrorPath verifies that metrics are populated
// even when the pipeline stops due to an error. Regression for pipeline.go:193 —
// metrics were only set on the success path.
func TestPipeline_MetricsAvailableOnErrorPath(t *testing.T) {
	g := NewParallelGomega(t)

	detectorErr := errors.New("detector exploded")
	d := &mockDetector{name: "broken", err: detectorErr}

	cfg := DefaultConfig()
	cfg.MaxIterations = 3
	cfg.Metrics = NewMetrics()

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).ToNot(HaveOccurred())

	result, runErr := p.Run(context.Background())
	g.Expect(runErr).To(HaveOccurred())

	// Metrics must be populated despite the error.
	g.Expect(result.Metrics.StartTime).ToNot(BeZero())
	g.Expect(result.Metrics.EndTime).ToNot(BeZero())
	g.Expect(result.Reason).To(Equal(ReasonError))
}

// TestPipeline_StageAfterHookErrorAborts verifies that a StageAfter hook error
// aborts the pipeline. Regression for pipeline_iteration.go:31 — StageAfter
// errors were silently discarded with `_ =`.
func TestPipeline_StageAfterHookErrorAborts(t *testing.T) {
	g := NewParallelGomega(t)

	abortErr := errors.New("after-hook-abort")

	hook := StageHookFunc(func(_ context.Context, event StageEvent) error {
		if event.Timing == StageAfter && event.Stage == StageDetect {
			return abortErr
		}

		return nil
	})

	d := newMockDetector("test", "tool", finding.NewFinding(
		"rule", "tool", "msg", finding.SeverityInfo,
		finding.Position{}, finding.ConfidenceHigh,
	))

	cfg := DefaultConfig()
	cfg.MaxIterations = 1
	cfg.StageHooks = []StageHook{hook}

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).ToNot(HaveOccurred())

	_, runErr := p.Run(context.Background())
	g.Expect(runErr).To(HaveOccurred())
	g.Expect(runErr).To(MatchError(ContainSubstring("after")))
}

// TestPipeline_TotalDetectedDeduplicated verifies that TotalDetected counts
// unique findings only, not duplicates across iterations. Regression for
// pipeline.go:249 — previously used raw p.findings (accumulated, not deduped).
func TestPipeline_TotalDetectedDeduplicated(t *testing.T) {
	g := NewParallelGomega(t)

	// Same finding returned every iteration.
	sameFinding := finding.NewFinding(
		"rule-x", "tool", "msg", finding.SeverityInfo,
		finding.Position{File: "a.go", Line: 1}, finding.ConfidenceHigh,
	)

	d := newMockDetector("test", "tool", sameFinding)

	cfg := DefaultConfig()
	cfg.MaxIterations = 3

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).ToNot(HaveOccurred())

	result, err := p.Run(context.Background())
	g.Expect(err).ToNot(HaveOccurred())

	// The same finding is detected 3 times (once per iteration) but
	// TotalDetected must report 1 (deduplicated).
	g.Expect(result.TotalDetected).To(Equal(1), "TotalDetected must count unique findings only")
}

// TestPipeline_SuggestFindingsShiftedAfterDirectFix verifies that suggest
// findings have their positions shifted when a direct fix in the same iteration
// adds lines. Regression for pipeline_detect.go:237 — previously only
// iter.findings was shifted, not iter.suggest.
func TestPipeline_SuggestFindingsShiftedAfterDirectFix(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()

	// "line1\nold\nline3\nline4\n" — replacing "old" (1 line) with 3 lines
	// shifts everything after line 2 by +2.
	content := []byte("line1\nold\nline3\nline4\n")
	writeTestFile(t, filepath.Join(tmpDir, "test.go"), content)

	directFix := finding.Finding{
		ID:          "fix-1",
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "replace old",
		Severity:    finding.SeverityInfo,
		Position:    finding.Pos("test.go", 2, 1),
		FixStrategy: finding.FixStrategyDirect,
		BeforeCode:  "old",
		AfterCode:   "new1\nnew2\nnew3",
	}

	// Suggest finding at line 3 — should shift to line 5 (+2) after fix.
	suggestFinding := finding.Finding{
		ID:          "suggest-1",
		Rule:        "r2",
		ToolName:    "tool",
		Message:     "consider refactoring",
		Severity:    finding.SeverityInfo,
		Position:    finding.Pos("test.go", 3, 1),
		FixStrategy: finding.FixStrategySuggest,
		Suggestion:  "use a helper function",
	}

	cfg := DefaultConfig()

	p, err := New(cfg, tmpDir)
	g.Expect(err).ToNot(HaveOccurred())

	iter := Iteration{
		Number:  1,
		suggest: []finding.Finding{suggestFinding},
	}

	err = p.applyTriage(context.Background(), []finding.Finding{directFix}, &iter, &PipelineResult{})
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(iter.suggest).To(HaveLen(1))
	g.Expect(iter.suggest[0].Position.Line).To(
		Equal(5),
		"suggest finding line should shift +2 after direct fix added 2 lines",
	)
}

// TestOnFixOutcome_CarriesExactStatuses verifies D3: the OnFixOutcome
// callback receives the per-finding FixOutcomeStatus (and error for failed
// resolutions), distinguishing applied from refused outcomes that the
// boolean OnFix collapses.
func TestOnFixOutcome_CarriesExactStatuses(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	appliedFix := directFix("fix-applied", "r1", "tool", "replace old", "old()", "new()", "fixme.go", 4)
	refusedFix := directFix("fix-refused", "r1", "tool", "absent", "nonexistent", "new()", "fixme.go", 5)

	type call struct {
		id     finding.ID
		status FixOutcomeStatus
		failed bool
	}

	var calls []call

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFixOutcome: func(f finding.Finding, status FixOutcomeStatus, err error) {
			calls = append(calls, call{id: f.ID, status: status, failed: err != nil})
		},
	}

	p, err := New(cfg, tmpDir, mockDetWithFindings("tool", appliedFix, refusedFix))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	byID := make(map[finding.ID]call, len(calls))
	for _, c := range calls {
		byID[c.id] = c
	}

	g.Expect(byID[finding.ID("fix-applied")].status).To(Equal(FixOutcomeApplied))
	g.Expect(byID[finding.ID("fix-applied")].failed).To(BeFalse())

	g.Expect(byID[finding.ID("fix-refused")].status).To(Equal(FixOutcomeRefused))
	g.Expect(byID[finding.ID("fix-refused")].failed).To(BeFalse(),
		"refusal is a decision, not an error")
}

// TestOnFixOutcome_EventOrdering pins the observable callback contract:
// within a fix stage, every OnFixOutcome call happens before any legacy OnFix
// call, each finding fires exactly once per callback (no double-fire), and
// OnFixOutcome sees every outcome (applied AND refused) while legacy OnFix
// only reports applied fixes. Refactors must not silently reorder these.
func TestOnFixOutcome_EventOrdering(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	writeTestFile(t, testFile, []byte("package main\n\nfunc main() {\n\told()\n}\n"))

	applied := directFix("ord-applied", "r1", "tool", "replace old", "old()", "new()", "fixme.go", 4)
	refused := directFix("ord-refused", "r1", "tool", "absent", "nonexistent", "new()", "fixme.go", 5)

	type event struct {
		kind   string // "outcome" | "legacy"
		id     finding.ID
		status FixOutcomeStatus
		appl   bool
	}

	var events []event
	onFixOutcomeCalls := map[finding.ID]int{}
	onFixCalls := map[finding.ID]int{}

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFixOutcome: func(f finding.Finding, status FixOutcomeStatus, _ error) {
			events = append(events, event{kind: "outcome", id: f.ID, status: status})
			onFixOutcomeCalls[f.ID]++
		},
		OnFix: func(f finding.Finding, wasApplied bool) {
			events = append(events, event{kind: "legacy", id: f.ID, appl: wasApplied})
			onFixCalls[f.ID]++
		},
	}

	p, err := New(cfg, tmpDir, mockDetWithFindings("tool", applied, refused))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := p.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(onFixOutcomeCalls).To(HaveLen(2), "one outcome event per finding")
	for id, n := range onFixOutcomeCalls {
		g.Expect(n).To(Equal(1), "OnFixOutcome double-fired for %s", id)
	}
	g.Expect(onFixCalls[finding.ID("ord-applied")]).To(Equal(1))
	g.Expect(onFixCalls[finding.ID("ord-refused")]).To(Equal(1),
		"legacy OnFix fires for every safe fix (false when not applied)")
	for _, e := range events {
		if e.kind == "legacy" && e.id == finding.ID("ord-applied") {
			g.Expect(e.appl).To(BeTrue())
		}
		if e.kind == "legacy" && e.id == finding.ID("ord-refused") {
			g.Expect(e.appl).To(BeFalse())
		}
	}

	firstLegacy := -1
	for i, e := range events {
		if e.kind == "legacy" {
			firstLegacy = i
			break
		}
	}
	g.Expect(firstLegacy).To(BeNumerically(">", 0), "legacy OnFix must not fire before OnFixOutcome")
	for i := 0; i < firstLegacy; i++ {
		g.Expect(events[i].kind).To(Equal("outcome"),
			"all OnFixOutcome events must precede the first OnFix event")
	}

	statuses := map[finding.ID]FixOutcomeStatus{}
	for _, e := range events {
		if e.kind == "outcome" {
			statuses[e.id] = e.status
		}
	}
	g.Expect(statuses[finding.ID("ord-applied")]).To(Equal(FixOutcomeApplied))
	g.Expect(statuses[finding.ID("ord-refused")]).To(Equal(FixOutcomeRefused))
}

// TestOnFixOutcome_ConcurrentPipelines pins callback safety across concurrent
// pipeline runs sharing user callbacks: exactly one OnFixOutcome and one
// legacy OnFix event per fixable finding per run — no lost, duplicated, or
// interleaved-corrupted events. Temp dirs and fixture files are created up
// front; only Run executes in the goroutines. Run under -race.
func TestOnFixOutcome_ConcurrentPipelines(t *testing.T) {
	g := NewParallelGomega(t)

	const pipelines = 8
	const findingsPerPipeline = 2

	dirs := make([]string, pipelines)
	for i := range dirs {
		dirs[i] = t.TempDir()
		writeTestFile(t, filepath.Join(dirs[i], "fixme.go"),
			[]byte("package main\n\nfunc main() {\n\told()\n}\n"))
	}

	var mu sync.Mutex
	outcomeCalls := map[string]int{}
	legacyCalls := map[string]int{}

	var wg sync.WaitGroup
	for i := range pipelines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			prefix := fmt.Sprintf("p%d-", n)
			applied := directFix(prefix+"applied", "r1", "tool", "replace old", "old()", "new()", "fixme.go", 4)
			refused := directFix(prefix+"refused", "r1", "tool", "absent", "nonexistent", "new()", "fixme.go", 5)

			cfg := Config{
				MaxIterations:     1,
				ParallelDetectors: false,
				OnFixOutcome: func(f finding.Finding, _ FixOutcomeStatus, _ error) {
					mu.Lock()
					outcomeCalls[string(f.ID)]++
					mu.Unlock()
				},
				OnFix: func(f finding.Finding, _ bool) {
					mu.Lock()
					legacyCalls[string(f.ID)]++
					mu.Unlock()
				},
			}

			p, err := New(cfg, dirs[n], mockDetWithFindings("tool", applied, refused))
			if err != nil {
				t.Errorf("pipeline %d: New: %v", n, err)
				return
			}
			if _, err := p.Run(context.Background()); err != nil {
				t.Errorf("pipeline %d: Run: %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	g.Expect(outcomeCalls).To(HaveLen(pipelines*findingsPerPipeline),
		"one outcome entry per finding per pipeline")
	g.Expect(legacyCalls).To(HaveLen(pipelines*findingsPerPipeline),
		"one legacy entry per finding per pipeline")
	for id, n := range outcomeCalls {
		g.Expect(n).To(Equal(1), "OnFixOutcome fired %d times for %s", n, id)
	}
	for id, n := range legacyCalls {
		g.Expect(n).To(Equal(1), "OnFix fired %d times for %s", n, id)
	}
}
