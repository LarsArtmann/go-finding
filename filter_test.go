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
		{ID: "1", FixStrategy: FixStrategyDirect},
		{ID: "2", FixStrategy: id2},
		{ID: "3", FixStrategy: id3},
	}
}

func makeFindingsWithSeverity(sev ...Severity) []Finding {
	findings := make([]Finding, len(sev))
	for i, s := range sev {
		findings[i] = Finding{ID: string(rune('1' + i)), Severity: s}
	}

	return findings
}

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
	runFilterCase(t, newFilterCase("has fix", findings, HasFix, 1))
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
	if len(groups) != 0 {
		t.Errorf("GroupBy(nil) = %d groups, want 0", len(groups))
	}
}

func TestGroupByFile(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Position: Position{File: filterTestFileA}},
		{ID: "2", Position: Position{File: filterTestFileB}},
		{ID: "3", Position: Position{File: filterTestFileA}},
	}
	groups := GroupByFile(findings)
	assertGroupLen(t, groups, filterTestFileA, 2, "GroupByFile 'a.go'")
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
		{ID: "b", Position: Position{File: filterTestFileB, Line: 5, Column: 1}},
		{ID: "a2", Position: Position{File: filterTestFileA, Line: 10, Column: 5}},
		{ID: "a1", Position: Position{File: filterTestFileA, Line: 10, Column: 1}},
		{ID: "a0", Position: Position{File: filterTestFileA, Line: 3, Column: 1}},
	}

	SortByPosition(findings)

	want := []string{"a0", "a1", "a2", "b"}
	AssertFindingsIDs(t, findings, want)
}

func TestSortBySeverity(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: string(SeverityInfo), Severity: SeverityInfo},
		{ID: string(SeverityCritical), Severity: SeverityCritical},
		{ID: string(SeverityWarning), Severity: SeverityWarning},
		{ID: string(SeverityError), Severity: SeverityError},
	}

	SortBySeverity(findings)

	want := []string{
		string(SeverityCritical),
		string(SeverityError),
		string(SeverityWarning),
		string(SeverityInfo),
	}
	AssertFindingsIDs(t, findings, want)
}

func TestFilterInPlace(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityInfo},
		{ID: "3", Severity: SeverityError},
	}

	result := FilterInPlace(findings, BySeverity(SeverityError))
	AssertFindingsLenAndIDs(t, result, []string{"1", "3"}, "FilterInPlace")
}

func TestFilterInPlace_NoPredicates(t *testing.T) {
	t.Parallel()

	findings := []Finding{{ID: "1"}, {ID: "2"}}
	result := FilterInPlace(findings)

	if len(result) != 2 {
		t.Fatalf("FilterInPlace() = %d, want 2", len(result))
	}

	if &findings[0] != &result[0] {
		t.Error("FilterInPlace without predicates should return same slice")
	}
}

func TestFilterInPlace_AllFiltered(t *testing.T) {
	t.Parallel()

	findings := []Finding{{ID: "1", Severity: SeverityInfo}}
	result := FilterInPlace(findings, BySeverity(SeverityError))

	if len(result) != 0 {
		t.Errorf("FilterInPlace() = %d, want 0", len(result))
	}
}

func TestAnyOf(t *testing.T) {
	t.Parallel()

	result := Filter(threeSevFindings(), AnyOf(BySeverity(SeverityError), BySeverity(SeverityInfo)))
	if len(result) != 2 {
		t.Errorf("AnyOf(error, info) = %d findings, want 2", len(result))
	}
}

func TestAnyOf_SingleMatch(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
	}

	result := Filter(findings, AnyOf(BySeverity(SeverityError)))
	AssertFindingsIDs(t, result, []string{"1"})
}

func TestNegate(t *testing.T) {
	t.Parallel()

	result := Filter(threeSevFindings(), Negate(BySeverity(SeverityError)))
	if len(result) != 2 {
		t.Errorf("Negate(BySeverity(error)) = %d, want 2", len(result))
	}
}

func TestByConfidence(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Confidence: ConfidenceHigh},
		{ID: "2", Confidence: ConfidenceLow},
		{ID: "3", Confidence: ConfidenceHigh},
	}

	result := Filter(findings, ByConfidence(ConfidenceHigh))
	AssertFindingsIDs(t, result, []string{"1", "3"})
}

func TestByConfidenceAtLeast(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Confidence: ConfidenceNone},
		{ID: "2", Confidence: ConfidenceMedium},
		{ID: "3", Confidence: ConfidenceHigh},
		{ID: "4", Confidence: ConfidenceFull},
	}

	result := Filter(findings, ByConfidenceAtLeast(ConfidenceMedium))
	AssertFindingsIDs(t, result, []string{"2", "3", "4"})
}
