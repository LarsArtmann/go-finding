package finding

import (
	"fmt"
	"testing"

	"testing/quick"
)

// correlateRangeFinding builds a finding with a line range on the given file.
func correlateRangeFinding(id, tool string, file FilePath, start, end int) Finding {
	return Finding{
		ID:       ID(id),
		ToolName: ToolName(tool),
		Rule:     "r",
		Message:  "m",
		Severity: SeverityWarning,
		Position: Pos(file, start, 1),
		Range:    &Range{Start: Pos(file, start, 1), End: Pos(file, end, 1)},
	}
}

// TestCorrelate_ZeroLinePointsSkipped pins the zero-line guard in the
// proximity strategy: point findings without a line never correlate.
func TestCorrelate_ZeroLinePointsSkipped(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "a", ToolName: "toolA", Rule: "r", Message: "m", Severity: SeverityWarning, Position: Pos("x.go", 0, 1)},
		{ID: "b", ToolName: "toolB", Rule: "r", Message: "m", Severity: SeverityWarning, Position: Pos("x.go", 3, 1)},
	}

	if got := Correlate(findings); len(got) != 0 {
		t.Fatalf("zero-line point findings must not correlate, got %d", len(got))
	}
}

// TestCorrelate_SameToolPointInRange pins the same-tool guard in the
// range-and-point strategy: a point inside a range of the SAME tool is
// not correlated.
func TestCorrelate_SameToolPointInRange(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		correlateRangeFinding("range", "sameTool", "x.go", 4, 6),
		{ID: "point", ToolName: "sameTool", Rule: "r", Message: "m", Severity: SeverityWarning, Position: Pos("x.go", 5, 1)},
	}

	if got := Correlate(findings); len(got) != 0 {
		t.Fatalf("same-tool point-in-range must not correlate, got %d", len(got))
	}
}

// TestCorrelate_MaxCorrelationsCap pins the O(n^2) hang guard: 150 pairwise-
// overlapping range findings with pairwise-distinct tools would produce
// >10000 correlations; the result must be capped at exactly maxCorrelations.
func TestCorrelate_MaxCorrelationsCap(t *testing.T) {
	t.Parallel()

	findings := make([]Finding, 0, 150)
	for i := range 150 {
		findings = append(findings,
			correlateRangeFinding(fmt.Sprintf("f%d", i), fmt.Sprintf("tool%d", i), "x.go", 1, 2))
	}

	if got, want := len(Correlate(findings)), maxCorrelations; got != want {
		t.Fatalf("cap: got %d correlations, want %d", got, want)
	}
}

// TestNewIntervalIndex_EmptyInput pins the empty-index contract: queries on
// an index built from no intervals answer with no overlaps.
func TestNewIntervalIndex_EmptyInput(t *testing.T) {
	t.Parallel()

	idx := NewIntervalIndex[int](nil)
	if got := idx.Query(0, 100); got != nil {
		t.Fatalf("empty index must answer nil, got %v", got)
	}
}

// TestNewIntervalIndex_SortsByStartThenEnd pins the comparator: equal starts
// order by end, and queries see that order.
func TestNewIntervalIndex_SortsByStartThenEnd(t *testing.T) {
	t.Parallel()

	intervals := []Interval[string]{
		{Start: 5, End: 10, Value: "wide"},
		{Start: 1, End: 3, Value: "early"},
		{Start: 5, End: 6, Value: "narrow"},
	}

	idx := NewIntervalIndex(intervals)
	hits := idx.Query(5, 6)
	if len(hits) != 2 {
		t.Fatalf("expected 2 overlaps, got %d", len(hits))
	}

	if hits[0].Value != "narrow" || hits[1].Value != "wide" {
		t.Fatalf("equal starts must sort by end: got %v then %v", hits[0].Value, hits[1].Value)
	}
}

// TestCorrelate_QuickSanity keeps a quick-check lens on Correlate: any
// correlation it produces must reference real finding IDs.
//
// Consciously accepted uncovered guards (checked 2026-09-23): the
// f2.Position.Line == 0 check inside correlateByProximity's inner loop is
// unreachable through Correlate (ascending line sort places zero-line
// findings first, where the outer-loop guard already skips them), and the
// proximity-strategy maxCorrelations return is shadowed by the earlier
// overlap cap. Both remain as defense in depth.
func TestCorrelate_QuickSanity(t *testing.T) {
	t.Parallel()

	f := func(a, b uint8) bool {
		f1 := correlateRangeFinding("a", "toolA", "x.go", int(a)%20+1, int(a)%20+2)
		f2 := correlateRangeFinding("b", "toolB", "x.go", int(b)%20+1, int(b)%20+2)
		for _, c := range Correlate([]Finding{f1, f2}) {
			for _, id := range c.FindingIDs {
				if id != "a" && id != "b" {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Fatalf("quick: %v", err)
	}
}

// TestOverlapLength_DisjointReturnsZero pins the disjoint guard of
// overlapLength directly (unreachable through Correlate with sorted,
// overlapping-only inputs).
func TestOverlapLength_DisjointReturnsZero(t *testing.T) {
	t.Parallel()

	if got := overlapLength(1, 2, 5, 6); got != 0 {
		t.Fatalf("disjoint ranges must have zero overlap, got %d", got)
	}

	if got := overlapLength(5, 6, 1, 2); got != 0 {
		t.Fatalf("inverted argument order must have zero overlap, got %d", got)
	}
}
