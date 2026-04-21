package finding

import "testing"

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
	if len(active) != 1 {
		t.Fatalf("expected 1 active finding, got %d", len(active))
	}

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
	if len(active) != 2 {
		t.Fatalf("expected 2 active findings, got %d", len(active))
	}
}

func TestReportActiveFindings_None(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})

	active := r.ActiveFindings()
	if len(active) != 0 {
		t.Errorf("expected empty slice for empty findings, got %v", active)
	}
}

func TestReportByCategory(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Category: "security", Message: "a"})
	r.AddFinding(Finding{ID: "2", Category: "style", Message: "b"})
	r.AddFinding(Finding{ID: "3", Category: "security", Message: "c"})

	sec := r.ByCategory("security")
	if len(sec) != 2 {
		t.Fatalf("expected 2 security findings, got %d", len(sec))
	}

	style := r.ByCategory("style")
	if len(style) != 1 {
		t.Fatalf("expected 1 style finding, got %d", len(style))
	}

	none := r.ByCategory("nonexistent")
	AssertEmpty(t, none, "nonexistent category")
}

func TestReportByFixStrategy(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", FixStrategy: FixStrategyDirect, Message: "a"})
	r.AddFinding(Finding{ID: "2", FixStrategy: FixStrategyNone, Message: "b"})
	r.AddFinding(Finding{ID: "3", FixStrategy: FixStrategyDirect, Message: "c"})

	direct := r.ByFixStrategy(FixStrategyDirect)
	if len(direct) != 2 {
		t.Fatalf("expected 2 direct findings, got %d", len(direct))
	}

	none := r.ByFixStrategy(FixStrategyNone)
	if len(none) != 1 {
		t.Fatalf("expected 1 none finding, got %d", len(none))
	}
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
	r.AddFinding(Finding{ID: "1", Severity: SeverityError, Message: "a"})
	r.AddFinding(Finding{ID: "2", Severity: SeverityWarning, Message: "b"})
	r.AddFinding(Finding{ID: "3", Severity: SeverityError, Message: "c"})

	errors := r.BySeverity(SeverityError)
	if len(errors) != 2 {
		t.Fatalf("expected 2 error findings, got %d", len(errors))
	}

	warnings := r.BySeverity(SeverityWarning)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning finding, got %d", len(warnings))
	}
}

func TestReportFindByRule(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Rule: "SA1000", Message: "a"})
	r.AddFinding(Finding{ID: "2", Rule: "SA2000", Message: "b"})
	r.AddFinding(Finding{ID: "3", Rule: "SA1000", Message: "c"})

	matches := r.FindByRule("SA1000")
	if len(matches) != 2 {
		t.Fatalf("expected 2 SA1000 findings, got %d", len(matches))
	}

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

	if len(collected) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(collected))
	}

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
