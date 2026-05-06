package finding

import "testing"

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
	a.AddFinding(Finding{
		ID: "f1", Rule: "R1", Severity: SeverityError,
		Position: Position{File: "a.go"},
	})
	a.AddFinding(Finding{
		ID: "f2", Rule: "R2", Severity: SeverityWarning,
		Position: Position{File: "b.go"},
	})
	a.ComputeSummary()

	b := NewReport(ToolInfo{Name: "tool-b"})
	b.AddFinding(Finding{
		ID: "f3", Rule: "R3", Severity: SeverityInfo,
		Position: Position{File: "a.go"},
	})
	b.ComputeSummary()

	a.Merge(b)

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
	a.AddFinding(Finding{
		ID: "f1", Rule: "R1", Severity: SeverityError,
		Position: Position{File: "a.go"},
	})
	a.ComputeSummary()

	b := NewReport(ToolInfo{Name: "tool-b"})
	b.ComputeSummary()

	a.Merge(b)

	if len(a.Findings) != 1 {
		t.Errorf("Findings length = %d, want 1", len(a.Findings))
	}
}
