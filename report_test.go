package finding

import (
	"strings"
	"testing"
)

func assertByCategoryCount(t *testing.T, r *Report, cat string, want int) {
	t.Helper()

	if got := len(r.ByCategory(Category(cat))); got != want {
		t.Errorf("%s findings = %d, want %d", cat, got, want)
	}
}

func addCat(r *Report, id, cat, msg string) {
	r.AddFinding(Finding{ID: ID(id), Category: Category(cat), Message: msg})
}

func addRule(r *Report, id, rule, msg string) {
	r.AddFinding(Finding{ID: ID(id), Rule: RuleName(rule), Message: msg})
}

func addFix(r *Report, id string, fs FixStrategy, msg string) {
	r.AddFinding(Finding{ID: ID(id), FixStrategy: fs, Message: msg})
}

func addSev(r *Report, id string, sev Severity, msg string) {
	r.AddFinding(Finding{ID: ID(id), Severity: sev, Message: msg})
}

func TestReportActiveFindings(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Message: "active"})
	r.AddFinding(Finding{
		ID:          "2",
		Message:     "suppressed",
		Suppression: &Suppression{Reason: string(TagTest)},
	})

	active := r.ActiveFindings()
	if len(active) != 1 {
		t.Fatalf("active findings = %d, want 1", len(active))
	}

	if active[0].ID != "1" {
		t.Errorf("active[0].ID = %q, want %q", active[0].ID, "1")
	}
}

func TestReportActiveFindings_AllActive(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})

	active := r.ActiveFindings()
	if len(active) != 2 {
		t.Errorf("active findings = %d, want 2", len(active))
	}
}

func TestReportActiveFindings_None(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})

	active := r.ActiveFindings()
	if len(active) != 0 {
		t.Errorf("active findings = %d, want 0", len(active))
	}
}

func TestReportByCategory(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	addCat(r, "1", "security", "a")
	addCat(r, "2", "style", "b")
	addCat(r, "3", "security", "c")

	assertByCategoryCount(t, r, "security", 2)
	assertByCategoryCount(t, r, "style", 1)
	AssertEmpty(t, r.ByCategory("nonexistent"), "nonexistent category")
}

func TestReportByFixStrategy(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	addFix(r, "1", FixStrategyDirect, "a")
	addFix(r, "2", FixStrategyNone, "b")
	addFix(r, "3", FixStrategyDirect, "c")

	if len(r.ByFixStrategy(FixStrategyDirect)) != 2 {
		t.Errorf("direct findings = %d, want 2", len(r.ByFixStrategy(FixStrategyDirect)))
	}

	if len(r.ByFixStrategy(FixStrategyNone)) != 1 {
		t.Errorf("none findings = %d, want 1", len(r.ByFixStrategy(FixStrategyNone)))
	}
}

func TestReportFindByID(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "find-me", Message: "target"})
	r.AddFinding(Finding{ID: "other", Message: "other"})

	found := r.FindByID("find-me")
	if found == nil {
		t.Fatal("expected to find finding")
	}

	if found.Message != "target" {
		t.Errorf("Message = %q, want %q", found.Message, "target")
	}

	if r.FindByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestReportBySeverity(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	addSev(r, "1", SeverityError, "a")
	addSev(r, "2", SeverityWarning, "b")
	addSev(r, "3", SeverityError, "c")

	if len(r.BySeverity(SeverityError)) != 2 {
		t.Errorf("error findings = %d, want 2", len(r.BySeverity(SeverityError)))
	}

	if len(r.BySeverity(SeverityWarning)) != 1 {
		t.Errorf("warning findings = %d, want 1", len(r.BySeverity(SeverityWarning)))
	}
}

func TestReportFindByRule(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	addRule(r, "1", "SA1000", "a")
	addRule(r, "2", "SA2000", "b")
	addRule(r, "3", "SA1000", "c")

	if len(r.FindByRule("SA1000")) != 2 {
		t.Errorf("SA1000 findings = %d, want 2", len(r.FindByRule("SA1000")))
	}

	AssertEmpty(t, r.FindByRule("nonexistent"), "nonexistent rule")
}

func TestReportLen(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	if r.Len() != 0 {
		t.Errorf("empty report Len = %d, want 0", r.Len())
	}

	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})

	if r.Len() != 2 {
		t.Errorf("report with 2 findings Len = %d, want 2", r.Len())
	}
}

func TestReportAll(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})
	r.AddFinding(Finding{ID: "3", Message: "c"})

	collected := make([]Finding, 0, len(r.findings))
	for f := range r.All() {
		collected = append(collected, f)
	}

	AssertFindingsLenAndIDs(t, collected, []string{"1", "2", "3"}, "collected")
}

func TestReportAll_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})

	count := 0
	for range r.All() {
		count++
	}

	if count != 0 {
		t.Errorf("empty report iteration = %d, want 0", count)
	}
}

func addTestFinding(r *Report, id, rule string, sev Severity, file string) {
	r.AddFinding(Finding{
		ID: ID(id), Rule: RuleName(rule), Severity: sev,
		Position: Position{File: FilePath(file)},
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
		ID: ID(id), Rule: RuleName(rule), ToolName: "test", Message: msg,
		Severity: sev, Position: Position{File: FilePath(file), Line: line},
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

	merged := a.MergeInto(b)

	if len(merged.findings) != 3 {
		t.Errorf("Findings length = %d, want 3", len(merged.findings))
	}

	if merged.Summary.Total != 3 {
		t.Errorf("Summary.Total = %d, want 3", merged.Summary.Total)
	}

	if merged.Summary.BySeverity[SeverityError] != 1 {
		t.Errorf("BySeverity[Error] = %d, want 1", merged.Summary.BySeverity[SeverityError])
	}

	if merged.Summary.BySeverity[SeverityWarning] != 1 {
		t.Errorf("BySeverity[Warning] = %d, want 1", merged.Summary.BySeverity[SeverityWarning])
	}

	if merged.Summary.BySeverity[SeverityInfo] != 1 {
		t.Errorf("BySeverity[Info] = %d, want 1", merged.Summary.BySeverity[SeverityInfo])
	}

	if merged.Tool.Name != "tool-a" {
		t.Errorf("Tool.Name = %q, want %q", merged.Tool.Name, "tool-a")
	}
}

func TestReport_Merge_Empty(t *testing.T) {
	t.Parallel()

	a := NewReport(ToolInfo{Name: "tool-a"})
	addTestFinding(a, "f1", "R1", SeverityError, "a.go")
	a.ComputeSummary()

	b := NewReport(ToolInfo{Name: "tool-b"})
	b.ComputeSummary()

	merged := a.MergeInto(b)

	if len(merged.findings) != 1 {
		t.Errorf("Findings length = %d, want 1", len(a.findings))
	}
}

func TestToolInfo_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		err := (ToolInfo{Name: "govet"}).Validate()
		if err != nil {
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

		err := r.Validate()
		if err != nil {
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
