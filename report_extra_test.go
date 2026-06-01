package finding

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func addTestFinding(r *Report, id, rule string, sev Severity, file string) {
	r.AddFinding(Finding{
		ID: id, Rule: rule, Severity: sev,
		Position: Position{File: file},
	})
}

func assertCategoryCount(t *testing.T, r *Report, cat Category, want int) {
	t.Helper()
	if got := r.Summary.ByCategory[cat]; got != want {
		t.Errorf("ByCategory[%s] = %d, want %d", cat, got, want)
	}
}

func TestComputeSummary(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test", Version: "1.0"})
	addFinding(r, "1", "R1", "m1", SeverityError, "a.go", 1, CategorySecurity, FixStrategyDirect)
	addFinding(r, "2", "R2", "m2", SeverityWarning, "a.go", 2, CategoryStyle, FixStrategySuggest)
	r.AddFinding(Finding{
		ID: "3", Rule: "R3", ToolName: "test", Message: "m3",
		Severity: SeverityInfo, Position: Position{File: "b.go", Line: 1},
		FixStrategy: FixStrategyNone,
		Suppression: &Suppression{Kind: SuppressionInSource, Reason: "intentional"},
	})

	r.ComputeSummary()

	assertSummaryField(t, "Total", r.Summary.Total, 3)
	assertSummarySeverity(t, r, SeverityError, 1)
	assertSummarySeverity(t, r, SeverityWarning, 1)
	assertSummarySeverity(t, r, SeverityInfo, 1)

	assertCategoryCount(t, r, CategorySecurity, 1)
	assertCategoryCount(t, r, CategoryStyle, 1)

	assertSummaryFixStrategy(t, r, FixStrategyDirect, 1)
	assertSummaryFixStrategy(t, r, FixStrategySuggest, 1)
	assertSummaryFixStrategy(t, r, FixStrategyNone, 1)

	assertSummarySuppressed(t, r, 1)
	assertSummaryField(t, "FilesAffected", r.Summary.FilesAffected, 2)
}

func TestComputeSummary_Resets(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addFinding(r, "1", "R1", "m", SeverityError, "a.go", 1, CategorySecurity, FixStrategyDirect)
	r.ComputeSummary()

	if r.Summary.ByCategory[CategorySecurity] != 1 {
		t.Fatalf(
			"first call: ByCategory[Security] = %d, want 1",
			r.Summary.ByCategory[CategorySecurity],
		)
	}

	addFinding(r, "2", "R2", "m", SeverityWarning, "a.go", 2, CategoryStyle, FixStrategyNone)
	r.ComputeSummary()

	assertSummaryField(t, "Total", r.Summary.Total, 2)

	assertCategoryCount(t, r, CategorySecurity, 1)
	assertCategoryCount(t, r, CategoryStyle, 1)
}

func TestComputeSummary_PreservesDurationMs(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{
		ID: "1", Rule: "R1", ToolName: "test", Message: "m",
		Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
	})

	r.Summary.DurationMs = 1234
	r.ComputeSummary()

	if r.Summary.DurationMs != 1234 {
		t.Errorf(
			"DurationMs = %d, want 1234 (externally set value should be preserved)",
			r.Summary.DurationMs,
		)
	}
}

func addFinding(
	r *Report,
	id, rule, msg string,
	sev Severity,
	file string,
	line int,
	cat Category,
	fs FixStrategy,
) {
	r.AddFinding(Finding{
		ID: id, Rule: rule, ToolName: "test", Message: msg,
		Severity: sev, Position: Position{File: file, Line: line},
		Category: cat, FixStrategy: fs,
	})
}

func TestReport_Merge(t *testing.T) {
	t.Parallel()

	a := NewReport(ToolInfo{Name: "tool-a"})
	addTestFinding(a, "f1", "R1", SeverityError, "a.go")
	addTestFinding(a, "f2", "R2", SeverityWarning, "b.go")
	a.ComputeSummary()

	b := NewReport(ToolInfo{Name: "tool-b"})
	addTestFinding(b, "f3", "R3", SeverityInfo, "a.go")
	b.ComputeSummary()

	a.bCombine(b)

	if len(a.Findings) != 3 {
		t.Errorf("Findings length = %d, want 3", len(a.Findings))
	}

	if a.Summary.Total != 3 {
		t.Errorf("Summary.Total = %d, want 3", a.Summary.Total)
	}

	if a.Summary.BySeverity[SeverityError] != 1 {
		t.Errorf("BySeverity[Error] = %d, want 1", a.Summary.BySeverity[SeverityError])
	}

	if a.Summary.BySeverity[SeverityWarning] != 1 {
		t.Errorf("BySeverity[Warning] = %d, want 1", a.Summary.BySeverity[SeverityWarning])
	}

	if a.Summary.BySeverity[SeverityInfo] != 1 {
		t.Errorf("BySeverity[Info] = %d, want 1", a.Summary.BySeverity[SeverityInfo])
	}

	if a.Tool.Name != "tool-a" {
		t.Errorf("Tool.Name = %q, want %q", a.Tool.Name, "tool-a")
	}
}

func TestReport_Merge_Empty(t *testing.T) {
	t.Parallel()

	a := NewReport(ToolInfo{Name: "tool-a"})
	addTestFinding(a, "f1", "R1", SeverityError, "a.go")
	a.ComputeSummary()

	b := NewReport(ToolInfo{Name: "tool-b"})
	b.ComputeSummary()

	a.bCombine(b)

	if len(a.Findings) != 1 {
		t.Errorf("Findings length = %d, want 1", len(a.Findings))
	}
}

func TestToolInfo_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		if err := (ToolInfo{Name: "govet"}).Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})

	t.Run("empty_name", func(t *testing.T) {
		t.Parallel()

		err := ToolInfo{}.Validate()
		if err == nil {
			t.Fatal("Validate() = nil, want error")
		}
	})
}

func TestReport_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		r := NewReport(ToolInfo{Name: "test"})
		r.AddFinding(validFinding("R1", "test", "msg"))

		if err := r.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})

	t.Run("empty_tool_name", func(t *testing.T) {
		t.Parallel()

		r := NewReport(ToolInfo{})

		err := r.Validate()
		if err == nil {
			t.Fatal("Validate() = nil, want error")
		}

		if !strings.Contains(err.Error(), "ToolInfo.Name") {
			t.Errorf("Validate() = %v, want ToolInfo.Name error", err)
		}
	})

	t.Run("invalid_finding", func(t *testing.T) {
		t.Parallel()

		r := NewReport(ToolInfo{Name: "test"})
		r.AddFinding(Finding{})

		err := r.Validate()
		if err == nil {
			t.Fatal("Validate() = nil, want error")
		}

		if !strings.Contains(err.Error(), "findings[0]") {
			t.Errorf("Validate() = %v, want findings[0] error", err)
		}
	})

	t.Run("multiple_invalid_findings", func(t *testing.T) {
		t.Parallel()

		r := NewReport(ToolInfo{Name: "test"})
		r.AddFinding(Finding{})
		r.AddFinding(Finding{})

		err := r.Validate()
		if err == nil {
			t.Fatal("Validate() = nil, want error")
		}

		if !strings.Contains(err.Error(), "findings[0]") {
			t.Errorf("Validate() missing findings[0]: %v", err)
		}

		if !strings.Contains(err.Error(), "findings[1]") {
			t.Errorf("Validate() missing findings[1]: %v", err)
		}
	})
}

func validFinding(rule, tool, msg string) Finding {
	return NewFinding(rule, tool, msg, SeverityWarning, Pos("test.go", 1, 1), 0.5)
}

func TestReport_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "race-test"})

	const writers = 4
	const readers = 8
	const findingsPerWriter = 50

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
