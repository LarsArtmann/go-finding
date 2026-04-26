package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// filterTestCase represents a filter test with input findings and expected count.
type filterTestCase struct {
	name     string
	input    []Finding
	preds    []FilterFunc
	expected int
}

// newFilterCase creates a filter test case with a single predicate.
func newFilterCase(name string, input []Finding, pred FilterFunc, expected int) filterTestCase {
	return filterTestCase{name: name, input: input, preds: []FilterFunc{pred}, expected: expected}
}

// runFilterCase runs a filter test case.
func runFilterCase(t *testing.T, tc filterTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()

		result := Filter(tc.input, tc.preds...)
		if len(result) != tc.expected {
			t.Errorf("Filter() = %d, want %d", len(result), tc.expected)
		}
	})
}

// makeFixStrategyFindings creates test findings with varying FixStrategy values.
func makeFixStrategyFindings(id2, id3 FixStrategy) []Finding {
	return []Finding{
		{ID: "1", FixStrategy: FixStrategyDirect},
		{ID: "2", FixStrategy: id2},
		{ID: "3", FixStrategy: id3},
	}
}

// makeFindingsWithSeverity creates findings with specified severities.
func makeFindingsWithSeverity(sev ...Severity) []Finding {
	findings := make([]Finding, len(sev))
	for i, s := range sev {
		findings[i] = Finding{ID: string(rune('1' + i)), Severity: s}
	}

	return findings
}

// makeFindingsWithCategory creates findings with specified categories.
func makeFindingsWithCategory(cats ...Category) []Finding {
	findings := make([]Finding, len(cats))
	for i, c := range cats {
		findings[i] = Finding{ID: string(rune('1' + i)), Category: c}
	}

	return findings
}

func TestFilter_Empty(t *testing.T) {
	t.Parallel()

	result := Filter(nil, BySeverity(SeverityError))
	AssertEmpty(t, result, "Filter(nil)")
}

func TestFilter_NoPredicates(t *testing.T) {
	t.Parallel()

	findings := []Finding{{ID: "1"}, {ID: "2"}}

	result := Filter(findings)
	if len(result) != 2 {
		t.Errorf("Filter with no predicates = %d, want 2", len(result))
	}
}

func TestFilter_MultiplePredicates(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError, ToolName: "tool1"},
		{ID: "2", Severity: SeverityError, ToolName: "tool2"},
		{ID: "3", Severity: SeverityWarning, ToolName: "tool1"},
	}

	result := Filter(findings, BySeverity(SeverityError), ByTool("tool1"))
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("Filter(Severity=error, Tool=tool1) = %v, want [1]", result)
	}
}

func TestBySeverity(t *testing.T) {
	t.Parallel()

	findings := makeFindingsWithSeverity(SeverityError, SeverityWarning, SeverityError)
	runFilterCase(t, newFilterCase("error", findings, BySeverity(SeverityError), 2))
}

func TestBySeverityAtLeast(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityInfo},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityError},
		{ID: "4", Severity: SeverityCritical},
	}
	runFilterCase(
		t,
		newFilterCase("at least warning", findings, BySeverityAtLeast(SeverityWarning), 3),
	)
}

func TestByCategory(t *testing.T) {
	t.Parallel()

	findings := makeFindingsWithCategory(CategorySecurity, CategoryStyle, CategorySecurity)
	runFilterCase(t, newFilterCase("security", findings, ByCategory(CategorySecurity), 2))
}

func TestByFixStrategy(t *testing.T) {
	t.Parallel()

	findings := makeFixStrategyFindings(FixStrategyNone, FixStrategyDirect)
	runFilterCase(t, newFilterCase("direct", findings, ByFixStrategy(FixStrategyDirect), 2))
}

func TestByTool(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet"},
		{ID: "2", ToolName: "staticcheck"},
	}
	runFilterCase(t, newFilterCase("govet", findings, ByTool("govet"), 1))
}

func TestByRule(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Rule: "nilcheck"},
		{ID: "2", Rule: "unused"},
	}
	runFilterCase(t, newFilterCase("nilcheck", findings, ByRule("nilcheck"), 1))
}

func TestByFile(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Position: Position{File: "a.go"}},
		{ID: "2", Position: Position{File: "b.go"}},
	}
	runFilterCase(t, newFilterCase("a.go", findings, ByFile("a.go"), 1))
}

func TestNotSuppressed(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1"},
		{ID: "2", Suppression: &Suppression{Kind: SuppressionInSource}},
	}
	runFilterCase(t, newFilterCase("not suppressed", findings, NotSuppressed, 1))
}

func TestHasFix(t *testing.T) {
	t.Parallel()

	findings := makeFixStrategyFindings(FixStrategyNone, FixStrategyAI)
	runFilterCase(t, newFilterCase("has fix", findings, HasFix, 2))
}

func TestHasSuggestion(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Suggestion: "fix it"},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
		{ID: "3"},
	}
	runFilterCase(t, newFilterCase("has suggestion", findings, HasSuggestion, 2))
}

func assertGroupLen[K comparable](
	t *testing.T,
	groups map[K][]Finding,
	key K,
	want int,
	msg string,
) {
	t.Helper()

	if len(groups[key]) != want {
		t.Errorf("%s = %d, want %d", msg, len(groups[key]), want)
	}
}

func TestGroupByCustom(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "a"},
		{ID: "2", ToolName: "b"},
		{ID: "3", ToolName: "a"},
	}
	groups := GroupBy(findings, func(f Finding) string { return f.ToolName })
	assertGroupLen(t, groups, "a", 2, "GroupBy tool 'a'")
	assertGroupLen(t, groups, "b", 1, "GroupBy tool 'b'")
}

func TestGroupBy_Empty(t *testing.T) {
	t.Parallel()

	groups := GroupBy(nil, func(f Finding) string { return f.ToolName })
	assert.Empty(t, groups, "GroupBy(nil)")
}

func TestGroupByFile(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Position: Position{File: "a.go"}},
		{ID: "2", Position: Position{File: "b.go"}},
		{ID: "3", Position: Position{File: "a.go"}},
	}
	groups := GroupByFile(findings)
	assertGroupLen(t, groups, "a.go", 2, "GroupByFile 'a.go'")
}

func TestGroupBySeverity(t *testing.T) {
	t.Parallel()

	findings := makeFindingsWithSeverity(SeverityError, SeverityWarning, SeverityError)
	groups := GroupBySeverity(findings)
	assertGroupLen(t, groups, SeverityError, 2, "GroupBySeverity error")
	assertGroupLen(t, groups, SeverityWarning, 1, "GroupBySeverity warning")
}

func TestGroupByCategory(t *testing.T) {
	t.Parallel()

	findings := makeFindingsWithCategory(CategorySecurity, CategoryStyle, CategorySecurity)
	groups := GroupByCategory(findings)
	assertGroupLen(t, groups, CategorySecurity, 2, "GroupByCategory security")
}

func TestSortByPosition(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "b", Position: Position{File: "b.go", Line: 5, Column: 1}},
		{ID: "a2", Position: Position{File: "a.go", Line: 10, Column: 5}},
		{ID: "a1", Position: Position{File: "a.go", Line: 10, Column: 1}},
		{ID: "a0", Position: Position{File: "a.go", Line: 3, Column: 1}},
	}

	SortByPosition(findings)

	want := []string{"a0", "a1", "a2", "b"}
	AssertFindingsIDs(t, findings, want)
}

func TestSortBySeverity(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "info", Severity: SeverityInfo},
		{ID: "critical", Severity: SeverityCritical},
		{ID: "warning", Severity: SeverityWarning},
		{ID: "error", Severity: SeverityError},
	}

	SortBySeverity(findings)

	want := []string{"critical", "error", "warning", "info"}
	AssertFindingsIDs(t, findings, want)
}
