package finding

import (
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

// TestCorrelate_GuardBranches drives the branch guards left uncovered by the
// BDD suite: nil ranges, zero lines, same-tool point findings, the
// maxCorrelations cap, and the disjoint-range overlap guard.
func TestCorrelate_GuardBranches(t *testing.T) {
	t.Parallel()

	t.Run("nil-range and zero-line findings are skipped", func(t *testing.T) {
		t.Parallel()

		findings := []Finding{
			correlateRangeFinding("a", "toolA", "x.go", 1, 3),
			{ID: "norange", ToolName: "toolB", Rule: "r", Message: "m", Severity: SeverityWarning, Position: Pos("x.go", 1, 1)},
			correlateRangeFinding("zeroline", "toolC", "x.go", 1, 3),
		}
		findings[2].Position.Line = 0

		got := Correlate(findings)
		if len(got) != 0 {
			t.Fatalf("expected no correlations from uncorrelatable findings, got %d", len(got))
		}
	})

	t.Run("disjoint ranges do not correlate", func(t *testing.T) {
		t.Parallel()

		findings := []Finding{
			correlateRangeFinding("a", "toolA", "x.go", 1, 2),
			correlateRangeFinding("b", "toolB", "x.go", 50, 51),
		}

		got := Correlate(findings)
		if len(got) != 0 {
			t.Fatalf("disjoint ranges must not correlate, got %d", len(got))
		}
	})

	t.Run("same-tool point findings do not correlate", func(t *testing.T) {
		t.Parallel()

		sameLine := Finding{
			ID: "p1", Rule: "r", Message: "m", Severity: SeverityWarning,
			ToolName: "sameTool",
			Position: Pos("x.go", 5, 1),
		}
		findings := []Finding{
			correlateRangeFinding("range", "otherTool", "x.go", 4, 6),
			sameLine,
			{ID: "p2", ToolName: "sameTool", Rule: "r", Message: "m", Severity: SeverityWarning, Position: Pos("x.go", 5, 3)},
		}

		got := Correlate(findings)
		for _, c := range got {
			for _, id := range c.FindingIDs {
				if id == "p2" {
					t.Fatalf("same-tool point finding must not correlate: %+v", c)
				}
			}
		}
	})
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

// TestCorrelate_QuickSanity keeps a quick-check lens on Correlate: any pair
// of findings must produce correlations whose IDs reference real findings.
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
