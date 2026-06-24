package finding

import (
	"testing"
)

const (
	filterTestFileA = "a.go"
	filterTestFileB = "b.go"
)

func threeSevFindings() []Finding {
	return []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityInfo},
	}
}

type filterTestCase struct {
	name     string
	input    []Finding
	preds    []FilterFunc
	expected int
}

func newFilterCase(name string, input []Finding, pred FilterFunc, expected int) filterTestCase {
	return filterTestCase{name: name, input: input, preds: []FilterFunc{pred}, expected: expected}
}

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

func makeFixStrategyFindings(id2, id3 FixStrategy) []Finding {
	return []Finding{
		{ID: "1", FixStrategy: FixStrategyDirect, AfterCode: "fixed()"},
		{ID: "2", FixStrategy: id2},
		{ID: "3", FixStrategy: id3},
	}
}

func makeFindingsWithSeverity(sev ...Severity) []Finding {
	findings := make([]Finding, len(sev))
	for i, s := range sev {
		findings[i] = Finding{ID: ID(string(rune('1' + i))), Severity: s}
	}

	return findings
}

func makeFindingsWithCategory(cats ...Category) []Finding {
	findings := make([]Finding, len(cats))
	for i, c := range cats {
		findings[i] = Finding{ID: ID(string(rune('1' + i))), Category: c}
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
		{ID: "1", Position: Position{File: filterTestFileA}},
		{ID: "2", Position: Position{File: filterTestFileB}},
	}
	runFilterCase(t, newFilterCase(filterTestFileA, findings, ByFile(filterTestFileA), 1))
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
	runFilterCase(t, newFilterCase("has fix", findings, WithFix, 1))
}

func TestHasSuggestion(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Suggestion: "fix it"},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
		{ID: "3"},
	}
	runFilterCase(t, newFilterCase("has suggestion", findings, WithSuggestion, 2))
}
