package finding

import (
	"testing"
)

func TestFinding_Equal_FieldMismatch(t *testing.T) {
	t.Parallel()

	base := testFindingBase()

	tests := []struct {
		name string
		a, b Finding
		want bool
	}{
		{
			"different rule",
			base,
			func() Finding {
				f := testFindingBase()
				f.Rule = "other-rule"

				return f
			}(),
			false,
		},
		{
			"different tool name",
			base,
			func() Finding {
				f := testFindingBase()
				f.ToolName = "other-tool"

				return f
			}(),
			false,
		},
		{
			"different message",
			base,
			func() Finding {
				f := testFindingBase()
				f.Message = "other-msg"

				return f
			}(),
			false,
		},
		{
			"different category",
			base,
			func() Finding {
				f := testFindingBase()
				f.Category = CategoryPerformance

				return f
			}(),
			false,
		},
		{
			"different tags",
			base,
			func() Finding {
				f := testFindingBase()
				f.Tags = []Tag{"other-tag"}

				return f
			}(),
			false,
		},
		{
			"different fix strategy",
			base,
			func() Finding {
				f := testFindingBase()
				f.FixStrategy = FixStrategySuggest

				return f
			}(),
			false,
		},
	}

	testFindingEqualCases(t, tests)
}

func TestFinding_Equal_FieldMismatch_CodeAndMeta(t *testing.T) {
	t.Parallel()

	base := testFindingBase()

	tests := []struct {
		name string
		a, b Finding
		want bool
	}{
		{
			"different suggestion",
			base,
			func() Finding {
				f := testFindingBase()
				f.Suggestion = "other-suggestion"

				return f
			}(),
			false,
		},
		{
			"different before code",
			base,
			func() Finding {
				f := testFindingBase()
				f.BeforeCode = "old"

				return f
			}(),
			false,
		},
		{
			"different after code",
			base,
			func() Finding {
				f := testFindingBase()
				f.AfterCode = "new"

				return f
			}(),
			false,
		},
		{
			"different snippet",
			base,
			func() Finding {
				f := testFindingBase()
				f.Snippet = "other-snippet"

				return f
			}(),
			false,
		},
		{
			"different group id",
			base,
			func() Finding {
				f := testFindingBase()
				f.GroupID = "other-group"

				return f
			}(),
			false,
		},
		{
			"different confidence",
			base,
			func() Finding {
				f := testFindingBase()
				f.Confidence = 0.5

				return f
			}(),
			false,
		},
	}

	testFindingEqualCases(t, tests)
}

func TestFinding_Equal_Suppression(t *testing.T) {
	t.Parallel()

	a := Finding{ID: "a", Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R1"}}
	b := Finding{ID: "a", Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R1"}}
	c := Finding{ID: "a"}
	d := Finding{ID: "a", Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R2"}}

	if !a.Equal(b) {
		t.Error("same suppression should be equal")
	}

	if a.Equal(c) {
		t.Error("nil vs non-nil suppression should not be equal")
	}

	if a.Equal(d) {
		t.Error("different suppression should not be equal")
	}
}

func TestPosition_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b Position
		want int
	}{
		{"equal", Pos("a.go", 1, 2), Pos("a.go", 1, 2), 0},
		{"different file", Pos("a.go", 0, 0), Pos("b.go", 0, 0), -1},
		{"different file reverse", Pos("b.go", 0, 0), Pos("a.go", 0, 0), 1},
		{"different line", Pos("a.go", 1, 0), Pos("a.go", 2, 0), -1},
		{"different line reverse", Pos("a.go", 2, 0), Pos("a.go", 1, 0), 1},
		{"different column", Pos("a.go", 1, 1), Pos("a.go", 1, 2), -1},
		{"both zero", Position{}, Position{}, 0},
	}

	RunCompareTests(t, tests, func(a, b Position) int { return a.Compare(b) }, "Position")
}

func TestRange_Compare(t *testing.T) {
	t.Parallel()

	rng := func(f string, sl, el int) Range {
		return NewRange(FilePath(f), sl, 0, el, 0)
	}

	tests := []struct {
		name string
		a, b Range
		want int
	}{
		{"equal", rng("a.go", 1, 3), rng("a.go", 1, 3), 0},
		{"earlier start", rng("a.go", 1, 3), rng("a.go", 2, 3), -1},
		{"later start", rng("a.go", 2, 3), rng("a.go", 1, 3), 1},
		{"same start diff end", rng("a.go", 1, 2), rng("a.go", 1, 3), -1},
		{"different file", NewRange("a.go", 1, 0, 0, 0), NewRange("b.go", 1, 0, 0, 0), -1},
	}

	RunCompareTests(t, tests, func(a, b Range) int { return a.Compare(b) }, "Range")
}
