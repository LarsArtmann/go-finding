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
		Suppression: &Suppression{Kind: SuppressionInSource, Rule: "test", Reason: string(TagTest)},
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

func TestReportGroupFindings(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", GroupID: "group-a", Message: "a1"})
	r.AddFinding(Finding{ID: "2", GroupID: "group-a", Message: "a2"})
	r.AddFinding(Finding{ID: "3", GroupID: "group-b", Message: "b1"})
	r.AddFinding(Finding{ID: "4", Message: "no group"})
	r.AddFinding(Finding{
		ID:          "5",
		GroupID:     "group-c",
		Message:     "suppressed",
		Suppression: &Suppression{Kind: SuppressionInSource, Rule: "test", Reason: string(TagTest)},
	})

	groups := r.GroupFindings()

	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2", len(groups))
	}

	if len(groups["group-a"]) != 2 {
		t.Errorf("group-a findings = %d, want 2", len(groups["group-a"]))
	}

	if len(groups["group-b"]) != 1 {
		t.Errorf("group-b findings = %d, want 1", len(groups["group-b"]))
	}
}

func TestReportGroupFindings_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Message: "no group"})

	if groups := r.GroupFindings(); groups != nil {
		t.Errorf("GroupFindings = %v, want nil", groups)
	}
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
		Suppression: &Suppression{Kind: SuppressionInSource, Rule: "rule1", Reason: "intentional"},
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

func TestNewReportFromFindings(t *testing.T) {
	t.Parallel()

	tool := ToolInfo{Name: "test", Version: "1.0"}
	findings := []Finding{
		{
			ID: "1", Rule: "R1", ToolName: "test", Message: "m1",
			Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
			Category: CategorySecurity, FixStrategy: FixStrategyDirect,
		},
		{
			ID: "2", Rule: "R2", ToolName: "test", Message: "m2",
			Severity: SeverityWarning, Position: Position{File: "b.go", Line: 2},
			Category: CategoryStyle, FixStrategy: FixStrategySuggest,
		},
	}

	r := NewReportFromFindings(tool, findings)

	if r.Len() != 2 {
		t.Fatalf("Len = %d, want 2", r.Len())
	}

	if r.Tool.Name != "test" {
		t.Errorf("Tool.Name = %q, want %q", r.Tool.Name, "test")
	}

	assertSummaryField(t, "Total", r.Summary.Total, 2)
	assertSummarySeverity(t, r, SeverityError, 1)
	assertSummarySeverity(t, r, SeverityWarning, 1)
}

func TestNewReportFromFindings_Empty(t *testing.T) {
	t.Parallel()

	r := NewReportFromFindings(ToolInfo{Name: "test"}, nil)

	if r.Len() != 0 {
		t.Errorf("Len = %d, want 0", r.Len())
	}

	assertSummaryField(t, "Total", r.Summary.Total, 0)
}

func TestNewReportFromFindings_EqualsManual(t *testing.T) {
	t.Parallel()

	tool := ToolInfo{Name: "test"}
	findings := []Finding{
		{
			ID: "1", Rule: "R1", ToolName: "test", Message: "m1",
			Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
		},
		{
			ID: "2", Rule: "R2", ToolName: "test", Message: "m2",
			Severity: SeverityInfo, Position: Position{File: "b.go", Line: 3},
		},
	}

	r1 := NewReportFromFindings(tool, findings)

	r2 := NewReport(tool)
	r2.AddFindings(findings)
	r2.ComputeSummary()

	if r1.Len() != r2.Len() {
		t.Fatalf("Len mismatch: NewReportFromFindings=%d, manual=%d", r1.Len(), r2.Len())
	}

	if r1.Summary.Total != r2.Summary.Total {
		t.Errorf("Summary.Total mismatch: %d vs %d", r1.Summary.Total, r2.Summary.Total)
	}

	if r1.Summary.BySeverity[SeverityError] != r2.Summary.BySeverity[SeverityError] {
		t.Errorf("BySeverity[Error] mismatch: %d vs %d",
			r1.Summary.BySeverity[SeverityError], r2.Summary.BySeverity[SeverityError])
	}
}

func TestReportGroupFindingsSorted(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", GroupID: "group-z", Message: "z1"})
	r.AddFinding(Finding{ID: "2", GroupID: "group-a", Message: "a1"})
	r.AddFinding(Finding{ID: "3", GroupID: "group-a", Message: "a2"})
	r.AddFinding(Finding{ID: "4", GroupID: "group-m", Message: "m1"})
	r.AddFinding(Finding{ID: "5", Message: "no group"})

	groups := r.GroupFindingsSorted()

	if len(groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(groups))
	}

	wantOrder := []GroupID{"group-a", "group-m", "group-z"}
	for i, want := range wantOrder {
		if groups[i].ID != want {
			t.Errorf("groups[%d].ID = %q, want %q", i, groups[i].ID, want)
		}
	}

	if len(groups[0].Findings) != 2 || groups[0].Findings[0].ID != "2" || groups[0].Findings[1].ID != "3" {
		t.Errorf("group-a members = %v, want [2 3] in report order", groups[0].Findings)
	}
}

func TestReportGroupFindingsSorted_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})

	if groups := r.GroupFindingsSorted(); groups != nil {
		t.Errorf("GroupFindingsSorted = %v, want nil", groups)
	}
}

func TestReport_WithFinding_Chaining(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "tool"})
	got := r.WithFinding(Finding{Message: "one"}).WithFinding(Finding{Message: "two"})

	if got != r {
		t.Fatal("WithFinding must return the same report for chaining")
	}

	if n := len(r.FindingsSnapshot()); n != 2 {
		t.Errorf("findings after chaining = %d, want 2", n)
	}
}
