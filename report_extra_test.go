package finding

import "testing"

func TestComputeSummary(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test", Version: "1.0"})
	r.AddFinding(Finding{
		ID: "1", Rule: "R1", ToolName: "test", Message: "m1",
		Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
		Category: CategorySecurity, FixStrategy: FixStrategyDirect,
	})
	r.AddFinding(Finding{
		ID: "2", Rule: "R2", ToolName: "test", Message: "m2",
		Severity: SeverityWarning, Position: Position{File: "a.go", Line: 2},
		Category: CategoryStyle, FixStrategy: FixStrategySuggest,
	})
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
	r.AddFinding(Finding{
		ID: "1", Rule: "R1", ToolName: "test", Message: "m",
		Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
		Category: CategorySecurity, FixStrategy: FixStrategyDirect,
	})
	r.ComputeSummary()

	if r.Summary.ByCategory[CategorySecurity] != 1 {
		t.Fatalf("first call: ByCategory[Security] = %d, want 1", r.Summary.ByCategory[CategorySecurity])
	}

	r.AddFinding(Finding{
		ID: "2", Rule: "R2", ToolName: "test", Message: "m",
		Severity: SeverityWarning, Position: Position{File: "a.go", Line: 2},
		Category: CategoryStyle, FixStrategy: FixStrategyNone,
	})
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
