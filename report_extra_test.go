package finding

import "testing"

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

	if r.Summary.Total != 3 {
		t.Errorf("Total = %d, want 3", r.Summary.Total)
	}

	if r.Summary.BySeverity[SeverityError] != 1 {
		t.Errorf("BySeverity[Error] = %d, want 1", r.Summary.BySeverity[SeverityError])
	}

	if r.Summary.BySeverity[SeverityWarning] != 1 {
		t.Errorf("BySeverity[Warning] = %d, want 1", r.Summary.BySeverity[SeverityWarning])
	}

	if r.Summary.BySeverity[SeverityInfo] != 1 {
		t.Errorf("BySeverity[Info] = %d, want 1", r.Summary.BySeverity[SeverityInfo])
	}

	if r.Summary.ByCategory[CategorySecurity] != 1 {
		t.Errorf("ByCategory[Security] = %d, want 1", r.Summary.ByCategory[CategorySecurity])
	}

	if r.Summary.ByCategory[CategoryStyle] != 1 {
		t.Errorf("ByCategory[Style] = %d, want 1", r.Summary.ByCategory[CategoryStyle])
	}

	if r.Summary.ByFixStrategy[FixStrategyDirect] != 1 {
		t.Errorf("ByFixStrategy[Direct] = %d, want 1", r.Summary.ByFixStrategy[FixStrategyDirect])
	}

	if r.Summary.ByFixStrategy[FixStrategySuggest] != 1 {
		t.Errorf("ByFixStrategy[Suggest] = %d, want 1", r.Summary.ByFixStrategy[FixStrategySuggest])
	}

	if r.Summary.ByFixStrategy[FixStrategyNone] != 1 {
		t.Errorf("ByFixStrategy[None] = %d, want 1", r.Summary.ByFixStrategy[FixStrategyNone])
	}

	if r.Summary.Suppressed != 1 {
		t.Errorf("Suppressed = %d, want 1", r.Summary.Suppressed)
	}

	if r.Summary.FilesAffected != 2 {
		t.Errorf("FilesAffected = %d, want 2", r.Summary.FilesAffected)
	}
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

	if r.Summary.Total != 2 {
		t.Errorf("Total = %d, want 2", r.Summary.Total)
	}

	if r.Summary.ByCategory[CategorySecurity] != 1 {
		t.Errorf("ByCategory[Security] = %d, want 1", r.Summary.ByCategory[CategorySecurity])
	}

	if r.Summary.ByCategory[CategoryStyle] != 1 {
		t.Errorf("ByCategory[Style] = %d, want 1", r.Summary.ByCategory[CategoryStyle])
	}
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

func addFinding(r *Report, id, rule, msg string, sev Severity, file string, line int, cat Category, fs FixStrategy) {
	r.AddFinding(Finding{
		ID: id, Rule: rule, ToolName: "test", Message: msg,
		Severity: sev, Position: Position{File: file, Line: line},
		Category: cat, FixStrategy: fs,
	})
}
