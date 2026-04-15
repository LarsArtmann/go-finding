package finding

import "testing"

func TestMerge_Empty(t *testing.T) {
	t.Parallel()

	merged := Merge(nil)
	if merged == nil {
		t.Fatal("Merge(nil) = nil, want non-nil report")
	}
	if len(merged.Findings) != 0 {
		t.Errorf("Merge(nil) findings = %d, want 0", len(merged.Findings))
	}
}

func TestMerge_SingleReport(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "tool1"})
	r.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	merged := Merge([]*Report{r})
	if merged.Tool.Name != "tool1" {
		t.Errorf("single report merge tool = %q, want %q", merged.Tool.Name, "tool1")
	}
	if len(merged.Findings) != 1 {
		t.Errorf("single report merge findings = %d, want 1", len(merged.Findings))
	}
}

func TestMerge_MultipleReports(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "2", Severity: SeverityWarning, Position: Position{File: "b.go"}})

	merged := Merge([]*Report{r1, r2})
	merged.ComputeSummary()

	if merged.Tool.Name != "merged" {
		t.Errorf("merged tool name = %q, want %q", merged.Tool.Name, "merged")
	}
	if len(merged.Findings) != 2 {
		t.Errorf("merged findings = %d, want 2", len(merged.Findings))
	}
	if merged.Summary.FilesAffected != 2 {
		t.Errorf("merged files = %d, want 2", merged.Summary.FilesAffected)
	}
}

func TestMerge_WithDeduplication(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go", Line: 10}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "1", Severity: SeverityWarning, Position: Position{File: "a.go", Line: 10}})

	merged := Merge([]*Report{r1, r2}, WithDeduplication(true))
	merged.ComputeSummary()

	if len(merged.Findings) != 1 {
		t.Errorf("deduplicated merge = %d, want 1", len(merged.Findings))
	}
}

func TestMerge_WithoutDeduplication(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "1", Severity: SeverityWarning, Position: Position{File: "a.go"}})

	merged := Merge([]*Report{r1, r2}, WithDeduplication(false))
	merged.ComputeSummary()

	if len(merged.Findings) != 2 {
		t.Errorf("non-deduplicated merge = %d, want 2", len(merged.Findings))
	}
}

func TestMerge_DeduplicateByPosition(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "a", Severity: SeverityError, Position: Position{File: "a.go", Line: 10, Column: 5}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "b", Severity: SeverityWarning, Position: Position{File: "a.go", Line: 10, Column: 5}})

	merged := Merge([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByPosition))
	merged.ComputeSummary()

	if len(merged.Findings) != 1 {
		t.Errorf("dedup by position = %d, want 1", len(merged.Findings))
	}
}

func TestMerge_DeduplicateByRule(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "a", Rule: "nilcheck", Severity: SeverityError, Position: Position{File: "a.go", Line: 10}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "b", Rule: "nilcheck", Severity: SeverityWarning, Position: Position{File: "a.go", Line: 10}})

	merged := Merge([]*Report{r1, r2}, WithDeduplicateBy(DeduplicateByRule))
	merged.ComputeSummary()

	if len(merged.Findings) != 1 {
		t.Errorf("dedup by rule = %d, want 1", len(merged.Findings))
	}
}

func TestDedupKey(t *testing.T) {
	t.Parallel()

	f := Finding{ID: "test-id", Rule: "rule1", Position: Position{File: "a.go", Line: 10, Column: 5}}

	tests := []struct {
		name string
		by   DeduplicateBy
		want string
	}{
		{"by ID", DeduplicateByID, "test-id"},
		{"by position", DeduplicateByPosition, "a.go:10:5"},
		{"by rule", DeduplicateByRule, "rule1:a.go:10:5"},
		{"default", DeduplicateBy(99), "test-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := dedupKey(f, MergeOptions{DeduplicateBy: tt.by})
			if got != tt.want {
				t.Errorf("dedupKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCorrelate(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet", Rule: "nilcheck", Position: Position{File: "a.go", Line: 10}},
		{ID: "2", ToolName: "staticcheck", Rule: "nilcheck", Position: Position{File: "a.go", Line: 12}},
		{ID: "3", ToolName: "govet", Rule: "unused", Position: Position{File: "a.go", Line: 50}},
	}

	correlations := Correlate(findings)
	if len(correlations) == 0 {
		t.Error("Correlate() = 0 correlations, expected at least 1")
	}

	for _, c := range correlations {
		if c.Confidence <= 0 || c.Confidence > 1 {
			t.Errorf("correlation confidence = %f, want (0, 1]", c.Confidence)
		}
	}
}

func TestCorrelate_TooFewFindings(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet", Position: Position{File: "a.go", Line: 10}},
	}

	correlations := Correlate(findings)
	if len(correlations) != 0 {
		t.Errorf("Correlate() = %d, want 0 for single finding", len(correlations))
	}
}
