package finding

import "testing"

func TestRangeLinesEq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b Range
		want bool
	}{
		{
			"equal lines",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			true,
		},
		{
			"equal lines different columns",
			Range{Start: Position{Line: 10, Column: 5}, End: Position{Line: 20, Column: 8}},
			Range{Start: Position{Line: 10, Column: 3}, End: Position{Line: 20, Column: 9}},
			true,
		},
		{
			"different start lines",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 11}, End: Position{Line: 20}},
			false,
		},
		{
			"different end lines",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 10}, End: Position{Line: 21}},
			false,
		},
		{
			"zero values",
			Range{},
			Range{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := RangeLinesEq(tt.a, tt.b); got != tt.want {
				t.Errorf("RangeLinesEq() = %v, want %v", got, tt.want)
			}
		})
	}
}
