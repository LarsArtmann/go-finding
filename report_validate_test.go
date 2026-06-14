package finding

import (
	"sync"
	"testing"
	"time"
)

func validFinding(rule, tool, msg string) Finding {
	return NewFinding(rule, tool, msg, SeverityWarning, Pos("test.go", 1, 1), 0.5)
}

func TestReport_FindingsSnapshot(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "snapshot-test"})
	r.AddFinding(Finding{
		ID:       "1",
		Rule:     "R1",
		ToolName: "test",
		Message:  "msg1",
		Severity: SeverityError,
		Position: Pos("a.go", 1, 1),
		Tags:     []Tag{"security"},
		Metadata: map[string]string{"key": "value"},
	})
	r.AddFinding(Finding{
		ID:       "2",
		Rule:     "R2",
		ToolName: "test",
		Message:  "msg2",
		Severity: SeverityWarning,
		Position: Pos("b.go", 2, 1),
	})

	snapshot := r.FindingsSnapshot()

	if len(snapshot) != 2 {
		t.Fatalf("FindingsSnapshot() len = %d, want 2", len(snapshot))
	}

	snapshot[0].Message = "modified"

	if r.Findings[0].Message == "modified" {
		t.Error("FindingsSnapshot() did not clone: modification leaked back to report")
	}

	snapshot[0].Tags[0] = "modified-tag"
	if r.Findings[0].Tags[0] == "modified-tag" {
		t.Error("FindingsSnapshot() did not deep-clone Tags")
	}

	snapshot[0].Metadata["key"] = "modified-value"
	if r.Findings[0].Metadata["key"] == "modified-value" {
		t.Error("FindingsSnapshot() did not deep-clone Metadata")
	}
}

func TestReport_FindingsSnapshot_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "empty"})
	snapshot := r.FindingsSnapshot()

	if len(snapshot) != 0 {
		t.Errorf("FindingsSnapshot() len = %d, want 0 for empty report", len(snapshot))
	}
}

func TestReport_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "race-test"})

	const (
		writers           = 4
		readers           = 8
		findingsPerWriter = 50
	)

	var wg sync.WaitGroup

	for w := range writers {
		wg.Go(func() {
			for i := range findingsPerWriter {
				f := Finding{
					ID:       GenerateID("tool", "rule", Pos("file.go", w*100+i, 1)),
					Rule:     "rule",
					ToolName: "tool",
					Message:  "test",
					Severity: SeverityWarning,
					Position: Pos("file.go", w*100+i, 1),
				}
				r.AddFinding(f)
			}
		})
	}

	for rd := range readers {
		wg.Go(func() {
			_ = rd // use loop variable

			for i := range findingsPerWriter * 2 {
				_ = i // use loop variable

				_ = r.Len()
				_ = r.ActiveFindings()
				_ = r.BySeverity(SeverityWarning)
				_ = r.ByCategory(CategorySecurity)
				_ = r.ByFixStrategy(FixStrategyNone)
				_ = r.FindByID("nonexistent")
				_ = r.FindByRule("rule")
				_ = r.Filter(NotSuppressed)
				_ = r.Map(func(f Finding) Finding { return f })
				r.ComputeSummary()
			}
		})
	}

	wg.Wait()

	expected := writers * findingsPerWriter
	if r.Len() != expected {
		t.Errorf("Len() = %d, want %d", r.Len(), expected)
	}
}

func TestReport_All_IteratorReleasesLock(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "iter-test"})
	r.AddFinding(MakeSimpleFinding("1", SeverityError))
	r.AddFinding(MakeSimpleFinding("2", SeverityWarning))

	count := 0
	for range r.All() {
		count++

		break
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	r.AddFinding(MakeSimpleFinding("3", SeverityInfo))

	if r.Len() != 3 {
		t.Errorf("Len() = %d after break + add, want 3", r.Len())
	}
}

func suppressionFinding(id string, expiresAt *time.Time) Finding {
	return Finding{
		ID:       id,
		Rule:     "rule",
		ToolName: "tool",
		Message:  "test",
		Severity: SeverityWarning,
		Position: Pos("file.go", 1, 1),
		Suppression: &Suppression{
			Kind:      SuppressionInConfig,
			Rule:      "rule",
			Reason:    "expired",
			ExpiresAt: expiresAt,
		},
	}
}

func TestReport_ComputeSummaryAt_Deterministic(t *testing.T) {
	t.Parallel()

	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	expiredSuppression := suppressionFinding("exp1", &past)

	activeSuppression := suppressionFinding("act1", &future)

	r := NewReport(ToolInfo{Name: "deterministic"})
	r.AddFinding(expiredSuppression)
	r.AddFinding(activeSuppression)

	r.ComputeSummaryAt(now)

	if r.Summary.Suppressed != 1 {
		t.Errorf("Suppressed = %d, want 1 (only the non-expired one)", r.Summary.Suppressed)
	}

	afterBoth := time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)
	r.ComputeSummaryAt(afterBoth)

	if r.Summary.Suppressed != 0 {
		t.Errorf("Suppressed = %d after both expired, want 0", r.Summary.Suppressed)
	}
}
