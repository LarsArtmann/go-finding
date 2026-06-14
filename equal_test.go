package finding

import (
	"testing"
)

func TestPosition_Equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b Position
		want bool
	}{
		{
			"identical",
			Position{File: "a.go", Line: 1, Column: 2, Offset: 3},
			Position{File: "a.go", Line: 1, Column: 2, Offset: 3},
			true,
		},
		{"different file", Position{File: "a.go"}, Position{File: "b.go"}, false},
		{"different line", Position{File: "a.go", Line: 1}, Position{File: "a.go", Line: 2}, false},
		{
			"different column",
			Position{File: "a.go", Column: 1},
			Position{File: "a.go", Column: 2},
			false,
		},
		{
			"different offset",
			Position{File: "a.go", Offset: 1},
			Position{File: "a.go", Offset: 2},
			false,
		},
		{"both empty", Position{}, Position{}, true},
	}

	RunEqualTests(t, tests, func(a, b Position) bool { return a.Equal(b) }, "Position")
}

func TestRange_Equal(t *testing.T) {
	t.Parallel()

	lineEq := func(sl, el int) Range {
		return Range{Start: Position{Line: sl}, End: Position{Line: el}}
	}

	tests := []struct {
		name string
		a, b Range
		want bool
	}{
		{"identical", lineEq(1, 3), lineEq(1, 3), true},
		{"different start", lineEq(1, 3), lineEq(2, 3), false},
		{"different end", lineEq(1, 3), lineEq(1, 4), false},
		{
			"both empty end",
			Range{Start: Position{File: "a.go", Line: 1}},
			Range{Start: Position{File: "a.go", Line: 1}},
			true,
		},
	}

	RunEqualTests(t, tests, func(a, b Range) bool { return a.Equal(b) }, "Range")
}

func testFindingBase() Finding {
	return Finding{
		ID: "tool:rule:file.go:1:1", Rule: "rule", ToolName: "tool",
		Message: "msg", Severity: SeverityError,
		Position:    Position{File: "file.go", Line: 1},
		Category:    CategorySecurity,
		FixStrategy: FixStrategyDirect,
		Suggestion:  "fix it",
		Confidence:  0.9,
		Metadata:    map[string]string{"key": "val"},
	}
}

func testFindingWithRange(f Finding, line int) Finding {
	f.Range = &Range{Start: Position{File: "f.go", Line: line}}

	return f
}

func testFindingEqualCase(
	t *testing.T,
	name string,
	a, b Finding,
	want bool,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Parallel()

		if got := a.Equal(b); got != want {
			t.Errorf("Finding.Equal() = %v, want %v", got, want)
		}
	})
}

func testFindingEqualCases(t *testing.T, tests []struct {
	name string
	a, b Finding
	want bool
},
) {
	t.Helper()

	for _, tt := range tests {
		testFindingEqualCase(t, tt.name, tt.a, tt.b, tt.want)
	}
}

func TestFinding_Equal(t *testing.T) {
	t.Parallel()

	base := testFindingBase()
	rangedFinding := testFindingWithRange(testFindingBase(), 1)

	tests := []struct {
		name string
		a, b Finding
		want bool
	}{
		{"identical", base, base, true},
		{
			"different ID",
			base,
			func() Finding {
				f := testFindingBase()
				f.ID = "other"

				return f
			}(),
			false,
		},
		{
			"different severity",
			base,
			func() Finding {
				f := testFindingBase()
				f.Severity = SeverityWarning

				return f
			}(),
			false,
		},
		{
			"different position",
			base,
			func() Finding {
				f := testFindingBase()
				f.Position.Line = 99

				return f
			}(),
			false,
		},
		{
			"different metadata",
			base,
			func() Finding {
				f := testFindingBase()
				f.Metadata = map[string]string{"key": "val", "k2": "v2"}

				return f
			}(),
			false,
		},
		{"nil vs non-nil range", base, testFindingWithRange(testFindingBase(), 0), false},
		{"same range", rangedFinding, rangedFinding, true},
		{
			"different related",
			base,
			func() Finding {
				f := testFindingBase()
				f.Related = []RelatedRef{{FindingID: "x"}}

				return f
			}(),
			false,
		},
	}

	testFindingEqualCases(t, tests)
}
