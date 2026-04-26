package finding

import "testing"

func addCat(r *Report, id, cat, msg string) {
	r.AddFinding(Finding{ID: id, Category: Category(cat), Message: msg})
}

func addRule(r *Report, id, rule, msg string) {
	r.AddFinding(Finding{ID: id, Rule: rule, Message: msg})
}

func addFix(r *Report, id string, fs FixStrategy, msg string) {
	r.AddFinding(Finding{ID: id, FixStrategy: fs, Message: msg})
}

func addSev(r *Report, id string, sev Severity, msg string) {
	r.AddFinding(Finding{ID: id, Severity: sev, Message: msg})
}

func TestReportActiveFindings(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Message: "active"})
	r.AddFinding(Finding{
		ID:          "2",
		Message:     "suppressed",
		Suppression: &Suppression{Reason: "test"},
	})

	active := r.ActiveFindings()
	requireLenEq(t, len(active), 1, "active findings")

	if active[0].ID != "1" {
		t.Errorf("expected active finding ID '1', got %q", active[0].ID)
	}
}

func TestReportActiveFindings_AllActive(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})

	active := r.ActiveFindings()
	requireLenEq(t, len(active), 2, "active findings")
}

func TestReportActiveFindings_None(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})

	active := r.ActiveFindings()
	assertIntEq(t, len(active), 0, "active findings")
}

func TestReportByCategory(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addCat(r, "1", "security", "a")
	addCat(r, "2", "style", "b")
	addCat(r, "3", "security", "c")

	sec := r.ByCategory("security")
	requireLenEq(t, len(sec), 2, "security findings")

	style := r.ByCategory("style")
	requireLenEq(t, len(style), 1, "style findings")

	none := r.ByCategory("nonexistent")
	AssertEmpty(t, none, "nonexistent category")
}

func TestReportByFixStrategy(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addFix(r, "1", FixStrategyDirect, "a")
	addFix(r, "2", FixStrategyNone, "b")
	addFix(r, "3", FixStrategyDirect, "c")

	direct := r.ByFixStrategy(FixStrategyDirect)
	requireLenEq(t, len(direct), 2, "direct findings")

	none := r.ByFixStrategy(FixStrategyNone)
	requireLenEq(t, len(none), 1, "none findings")
}

func TestReportFindByID(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "find-me", Message: "target"})
	r.AddFinding(Finding{ID: "other", Message: "other"})

	found := r.FindByID("find-me")
	if found == nil {
		t.Fatal("expected to find finding")
	}

	if found.Message != "target" {
		t.Errorf("expected message 'target', got %q", found.Message)
	}

	notFound := r.FindByID("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestReportBySeverity(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addSev(r, "1", SeverityError, "a")
	addSev(r, "2", SeverityWarning, "b")
	addSev(r, "3", SeverityError, "c")

	errors := r.BySeverity(SeverityError)
	requireLenEq(t, len(errors), 2, "error findings")

	warnings := r.BySeverity(SeverityWarning)
	requireLenEq(t, len(warnings), 1, "warning findings")
}

func TestReportFindByRule(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addRule(r, "1", "SA1000", "a")
	addRule(r, "2", "SA2000", "b")
	addRule(r, "3", "SA1000", "c")

	matches := r.FindByRule("SA1000")
	requireLenEq(t, len(matches), 2, "SA1000 findings")

	none := r.FindByRule("nonexistent")
	AssertEmpty(t, none, "nonexistent rule")
}

func TestReportLen(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	if r.Len() != 0 {
		t.Fatalf("expected 0 findings, got %d", r.Len())
	}

	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})

	if r.Len() != 2 {
		t.Fatalf("expected 2 findings, got %d", r.Len())
	}
}

func TestReportAll(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})
	r.AddFinding(Finding{ID: "3", Message: "c"})

	collected := make([]Finding, 0, len(r.Findings))
	for f := range r.All() {
		collected = append(collected, f)
	}

	requireLenEq(t, len(collected), 3, "collected findings")

	if collected[0].ID != "1" || collected[1].ID != "2" || collected[2].ID != "3" {
		t.Errorf("unexpected order: %v", collected)
	}
}

func TestReportAll_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})

	count := 0
	for range r.All() {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 findings, got %d", count)
	}
}

func TestReportAll_BreakEarly(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	for i := range 10 {
		r.AddFinding(Finding{ID: string(rune('A' + i)), Message: "finding"})
	}

	count := 0
	for range r.All() {
		count++
		if count == 3 {
			break
		}
	}

	if count != 3 {
		t.Errorf("expected 3 iterations before break, got %d", count)
	}
}
