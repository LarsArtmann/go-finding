package finding

import (
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

	collected := make([]Finding, 0, len(r.Findings))
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
