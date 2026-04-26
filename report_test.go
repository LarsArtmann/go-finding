package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	require.Len(t, active, 1, "active findings")
	assert.Equal(t, "1", active[0].ID)
}

func TestReportActiveFindings_AllActive(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})

	active := r.ActiveFindings()
	assert.Len(t, active, 2, "active findings")
}

func TestReportActiveFindings_None(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	active := r.ActiveFindings()
	assert.Empty(t, active, "active findings")
}

func TestReportByCategory(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addCat(r, "1", "security", "a")
	addCat(r, "2", "style", "b")
	addCat(r, "3", "security", "c")

	assert.Len(t, r.ByCategory("security"), 2, "security findings")
	assert.Len(t, r.ByCategory("style"), 1, "style findings")
	AssertEmpty(t, r.ByCategory("nonexistent"), "nonexistent category")
}

func TestReportByFixStrategy(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addFix(r, "1", FixStrategyDirect, "a")
	addFix(r, "2", FixStrategyNone, "b")
	addFix(r, "3", FixStrategyDirect, "c")

	assert.Len(t, r.ByFixStrategy(FixStrategyDirect), 2, "direct findings")
	assert.Len(t, r.ByFixStrategy(FixStrategyNone), 1, "none findings")
}

func TestReportFindByID(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{ID: "find-me", Message: "target"})
	r.AddFinding(Finding{ID: "other", Message: "other"})

	found := r.FindByID("find-me")
	require.NotNil(t, found, "expected to find finding")
	assert.Equal(t, "target", found.Message)

	assert.Nil(t, r.FindByID("nonexistent"), "expected nil for nonexistent ID")
}

func TestReportBySeverity(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addSev(r, "1", SeverityError, "a")
	addSev(r, "2", SeverityWarning, "b")
	addSev(r, "3", SeverityError, "c")

	assert.Len(t, r.BySeverity(SeverityError), 2, "error findings")
	assert.Len(t, r.BySeverity(SeverityWarning), 1, "warning findings")
}

func TestReportFindByRule(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	addRule(r, "1", "SA1000", "a")
	addRule(r, "2", "SA2000", "b")
	addRule(r, "3", "SA1000", "c")

	assert.Len(t, r.FindByRule("SA1000"), 2, "SA1000 findings")
	AssertEmpty(t, r.FindByRule("nonexistent"), "nonexistent rule")
}

func TestReportLen(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	assert.Zero(t, r.Len(), "empty report")

	r.AddFinding(Finding{ID: "1", Message: "a"})
	r.AddFinding(Finding{ID: "2", Message: "b"})
	assert.Equal(t, 2, r.Len(), "report with 2 findings")
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

	assert.Len(t, collected, 3, "collected findings")
	assert.Equal(t, "1", collected[0].ID)
	assert.Equal(t, "2", collected[1].ID)
	assert.Equal(t, "3", collected[2].ID)
}

func TestReportAll_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})

	count := 0
	for range r.All() {
		count++
	}

	assert.Zero(t, count, "empty report iteration")
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

	assert.Equal(t, 3, count, "iterations before break")
}
