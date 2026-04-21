package finding

import (
	"testing"
)

func TestRangeLineCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want int
	}{
		{"no line info", Range{Start: Position{File: "a.go"}}, 0},
		{"single line no end", Range{Start: Position{File: "a.go", Line: 5}}, 1},
		{"single line with end", Range{Start: Position{Line: 5}, End: Position{Line: 5}}, 1},
		{"multi line", Range{Start: Position{Line: 10}, End: Position{Line: 20}}, 11},
		{"end < start", Range{Start: Position{Line: 20}, End: Position{Line: 10}}, 0},
		{"zero start line", Range{Start: Position{Line: 0}, End: Position{Line: 5}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.LineCount(); got != tt.want {
				t.Errorf("LineCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRangeLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want int
	}{
		{"no offsets", Range{Start: Position{File: "a.go", Line: 1}, End: Position{Line: 3}}, 0},
		{
			"missing start offset",
			Range{Start: Position{File: "a.go", Offset: -1}, End: Position{Offset: 100}},
			0,
		},
		{"missing end offset", Range{Start: Position{Offset: 10}, End: Position{Offset: -1}}, 0},
		{"valid range", Range{Start: Position{Offset: 50}, End: Position{Offset: 100}}, 50},
		{"zero start offset", Range{Start: Position{Offset: 0}, End: Position{Offset: 100}}, 100},
		{"end lt start", Range{Start: Position{Offset: 100}, End: Position{Offset: 50}}, 0},
		{"same start and end", Range{Start: Position{Offset: 50}, End: Position{Offset: 50}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.Length(); got != tt.want {
				t.Errorf("Length() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPosition_HasOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Position
		want bool
	}{
		{"zero offset (valid)", Position{Offset: 0}, true},
		{"positive offset", Position{Offset: 42}, true},
		{"negative offset (unset)", Position{Offset: -1}, false},
		{"default zero value", Position{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.p.HasOffset(); got != tt.want {
				t.Errorf("HasOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeContains_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("different file returns false", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 5}, End: Position{Line: 10}}
		p := Position{File: "b.go", Line: 7}
		if r.Contains(p) {
			t.Error("expected Contains to return false for different file")
		}
	})

	t.Run("offset-only range", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Offset: 10},
			End:   Position{File: "a.go", Offset: 50},
		}
		if !r.Contains(Position{File: "a.go", Offset: 30}) {
			t.Error("expected position at offset 30 to be in range 10-50")
		}

		if r.Contains(Position{File: "a.go", Offset: 5}) {
			t.Error("expected position at offset 5 to be outside range 10-50")
		}
	})

	t.Run("column within range on same line", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Line: 10, Column: 5},
			End:   Position{File: "a.go", Line: 10, Column: 15},
		}
		p := Position{File: "a.go", Line: 10, Column: 10}
		if !r.Contains(p) {
			t.Error("expected column 10 to be in range 5-15")
		}
	})

	t.Run("column outside range on end line", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Line: 10, Column: 5},
			End:   Position{File: "a.go", Line: 10, Column: 15},
		}
		p := Position{File: "a.go", Line: 10, Column: 20}
		if r.Contains(p) {
			t.Error("expected column 20 to be outside range 5-15")
		}
	})

	t.Run("column outside range on start line", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Line: 10, Column: 5},
			End:   Position{File: "a.go", Line: 20, Column: 15},
		}
		p := Position{File: "a.go", Line: 10, Column: 3}
		if r.Contains(p) {
			t.Error("expected column 3 to be outside range starting at column 5")
		}
	})

	t.Run("no line info uses offset", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Offset: 100},
			End:   Position{Offset: 200},
		}
		if !r.Contains(Position{File: "a.go", Offset: 150}) {
			t.Error("expected offset 150 to be in range 100-200")
		}
	})

	t.Run("no line no offset uses offset check", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go"}}
		p := Position{File: "a.go"}
		// Both have Offset=0 (default), which means containsByOffset is reached
		// since hasLineRange is false (both Line=0). Offset=0 >= 0, so the check runs.
		if !r.Contains(p) {
			t.Error("default Offset=0 means both have offset, so containsByOffset runs")
		}
	})
}

func TestRangeOverlaps_OffsetOnly(t *testing.T) {
	t.Parallel()

	t.Run("no end offset (zero value) defaults to start", func(t *testing.T) {
		t.Parallel()

		r1 := Range{Start: Position{File: "a.go", Offset: 50}, End: Position{Offset: 0}}
		r2 := Range{Start: Position{File: "a.go", Offset: 0}, End: Position{Offset: 50}}
		if !r1.Overlaps(r2) {
			t.Error("ranges should overlap when one end encompasses the other start")
		}
	})

	t.Run("point range overlaps with range", func(t *testing.T) {
		t.Parallel()

		point := Range{Start: Position{File: "a.go", Offset: 50}, End: Position{Offset: -1}}
		span := Range{Start: Position{File: "a.go", Offset: 30}, End: Position{Offset: 70}}
		if !point.Overlaps(span) {
			t.Error("point at 50 should overlap with range 30-70")
		}
	})

	t.Run("negative start offset returns false", func(t *testing.T) {
		t.Parallel()

		r1 := Range{Start: Position{File: "a.go", Offset: -1}}
		r2 := Range{Start: Position{File: "a.go", Offset: 50}}
		if r1.Overlaps(r2) {
			t.Error("negative offset should not overlap")
		}
	})
}

func TestRangeIntersection_OffsetOnly(t *testing.T) {
	t.Parallel()

	t.Run("overlapping offset ranges", func(t *testing.T) {
		t.Parallel()

		r1 := Range{
			Start: Position{File: "a.go", Offset: 100},
			End:   Position{Offset: 200},
		}
		r2 := Range{
			Start: Position{File: "a.go", Offset: 150},
			End:   Position{Offset: 250},
		}

		got := r1.Intersection(r2)
		if got == nil {
			t.Fatal("expected non-nil intersection")
		}

		if got.Start.Offset != 150 {
			t.Errorf("Start.Offset = %d, want 150", got.Start.Offset)
		}

		if got.End.Offset != 200 {
			t.Errorf("End.Offset = %d, want 200", got.End.Offset)
		}
	})

	t.Run("offset with negative end defaults to start", func(t *testing.T) {
		t.Parallel()

		r1 := Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: -1}}
		r2 := Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: -1}}
		got := r1.Intersection(r2)
		if got == nil {
			t.Fatal("expected non-nil intersection for same point")
		}

		if got.Start.Offset != 100 {
			t.Errorf("Start.Offset = %d, want 100", got.Start.Offset)
		}
	})

	t.Run("non-overlapping offset ranges return nil", func(t *testing.T) {
		t.Parallel()

		r1 := Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: 150}}
		r2 := Range{Start: Position{File: "a.go", Offset: 200}, End: Position{Offset: 250}}
		if got := r1.Intersection(r2); got != nil {
			t.Errorf("expected nil intersection, got %v", got)
		}
	})
}

func TestRangeIntersection_LineOnly_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("same start line no end", func(t *testing.T) {
		t.Parallel()

		r1 := Range{Start: Position{File: "a.go", Line: 10}}
		r2 := Range{Start: Position{File: "a.go", Line: 10}}
		got := r1.Intersection(r2)
		if got == nil {
			t.Fatal("expected non-nil intersection")
		}

		if got.Start.Line != 10 {
			t.Errorf("Start.Line = %d, want 10", got.Start.Line)
		}
	})
}

func TestRangeLinesEq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    Range
		b    Range
		want bool
	}{
		{
			"equal start and end lines",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			true,
		},
		{
			"different start line",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 15}, End: Position{Line: 20}},
			false,
		},
		{
			"different end line",
			Range{Start: Position{Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{Line: 10}, End: Position{Line: 25}},
			false,
		},
		{
			"different files same lines",
			Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			Range{Start: Position{File: "b.go", Line: 10}, End: Position{Line: 20}},
			true, // files are ignored by RangeLinesEq
		},
		{
			"same start and end line (single line)",
			Range{Start: Position{Line: 5}, End: Position{Line: 5}},
			Range{Start: Position{Line: 5}, End: Position{Line: 5}},
			true,
		},
		{
			"zero lines",
			Range{Start: Position{Line: 0}, End: Position{Line: 0}},
			Range{Start: Position{Line: 0}, End: Position{Line: 0}},
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

func TestRangeHasLineRange(t *testing.T) {
	t.Parallel()

	t.Run("position before range", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}}
		p := Position{File: "a.go", Line: 5}
		if r.Contains(p) {
			t.Error("position before range should not be contained")
		}
	})

	t.Run("position after range", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}}
		p := Position{File: "a.go", Line: 25}
		if r.Contains(p) {
			t.Error("position after range should not be contained")
		}
	})

	t.Run("position within range no column", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}}
		p := Position{File: "a.go", Line: 15}
		if !r.Contains(p) {
			t.Error("position at line 15 should be in range 10-20")
		}
	})

	t.Run("zero position line falls through to offset", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 10}}
		p := Position{File: "a.go", Line: 0}
		// Both have Offset=0 (Go zero value). containsByOffset runs because
		// the "both have line info" check fails. Since both offsets are 0
		// and 0 >= 0 (HasOffset is true), the offset check returns true.
		// This is a known semantic issue: Offset defaults to 0 which looks
		// like a valid offset.
		if !r.Contains(p) {
			t.Error("with default Offset=0 on both, containsByOffset returns true")
		}
	})
}
