package finding

import "testing"

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

	tests := []struct {
		name string
		a, b Range
		want bool
	}{
		{
			"identical",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			true,
		},
		{
			"different start",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 2}, End: Position{File: "a.go", Line: 3}},
			false,
		},
		{
			"different end",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 4}},
			false,
		},
		{
			"both empty end",
			Range{Start: Position{File: "a.go", Line: 1}},
			Range{Start: Position{File: "a.go", Line: 1}},
			true,
		},
	}

	RunEqualTests(t, tests, func(a, b Range) bool { return a.Equal(b) }, "Range")
}

func TestFinding_Equal(t *testing.T) {
	t.Parallel()

	newBase := func() Finding {
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

	tests := []struct {
		name string
		a, b Finding
		want bool
	}{
		{
			"identical",
			newBase(),
			newBase(),
			true,
		},
		{
			"different ID",
			newBase(),
			func() Finding { f := newBase(); f.ID = "other"; return f }(),
			false,
		},
		{
			"different severity",
			newBase(),
			func() Finding { f := newBase(); f.Severity = SeverityWarning; return f }(),
			false,
		},
		{
			"different position",
			newBase(),
			func() Finding { f := newBase(); f.Position.Line = 99; return f }(),
			false,
		},
		{
			"different metadata",
			newBase(),
			func() Finding { f := newBase(); f.Metadata = map[string]string{"key": "val", "k2": "v2"}; return f }(),
			false,
		},
		{
			"nil vs non-nil range",
			newBase(),
			func() Finding { f := newBase(); f.Range = &Range{Start: Position{File: "f.go"}}; return f }(),
			false,
		},
		{
			"same range",
			func() Finding { f := newBase(); f.Range = &Range{Start: Position{File: "f.go", Line: 1}}; return f }(),
			func() Finding { f := newBase(); f.Range = &Range{Start: Position{File: "f.go", Line: 1}}; return f }(),
			true,
		},
		{
			"different related",
			newBase(),
			func() Finding { f := newBase(); f.Related = []RelatedRef{{FindingID: "x"}}; return f }(),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Finding.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
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
		{
			"equal",
			Position{File: "a.go", Line: 1, Column: 2},
			Position{File: "a.go", Line: 1, Column: 2},
			0,
		},
		{"different file", Position{File: "a.go"}, Position{File: "b.go"}, -1},
		{"different file reverse", Position{File: "b.go"}, Position{File: "a.go"}, 1},
		{"different line", Position{File: "a.go", Line: 1}, Position{File: "a.go", Line: 2}, -1},
		{
			"different line reverse",
			Position{File: "a.go", Line: 2},
			Position{File: "a.go", Line: 1},
			1,
		},
		{
			"different column",
			Position{File: "a.go", Line: 1, Column: 1},
			Position{File: "a.go", Line: 1, Column: 2},
			-1,
		},
		{"both zero", Position{}, Position{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Compare(tt.b); got != tt.want {
				t.Errorf("Position.Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRange_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b Range
		want int
	}{
		{
			"equal",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			0,
		},
		{
			"earlier start",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 2}, End: Position{File: "a.go", Line: 3}},
			-1,
		},
		{
			"later start",
			Range{Start: Position{File: "a.go", Line: 2}, End: Position{File: "a.go", Line: 3}},
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			1,
		},
		{
			"same start different end",
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 2}},
			Range{Start: Position{File: "a.go", Line: 1}, End: Position{File: "a.go", Line: 3}},
			-1,
		},
		{
			"different file",
			Range{Start: Position{File: "a.go", Line: 1}},
			Range{Start: Position{File: "b.go", Line: 1}},
			-1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Compare(tt.b); got != tt.want {
				t.Errorf("Range.Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}
