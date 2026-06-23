package finding

import (
	"testing"
)

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
	groups := GroupBy(findings, func(f Finding) string { return string(f.ToolName) })
	assertGroupLen(t, groups, "a", 2, "GroupBy tool 'a'")
	assertGroupLen(t, groups, "b", 1, "GroupBy tool 'b'")
}

func TestGroupBy_Empty(t *testing.T) {
	t.Parallel()

	groups := GroupBy(nil, func(f Finding) string { return string(f.ToolName) })
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
		{ID: FindingID(string(SeverityInfo)), Severity: SeverityInfo},
		{ID: FindingID(string(SeverityCritical)), Severity: SeverityCritical},
		{ID: FindingID(string(SeverityWarning)), Severity: SeverityWarning},
		{ID: FindingID(string(SeverityError)), Severity: SeverityError},
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
