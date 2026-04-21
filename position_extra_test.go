package finding

import (
	"testing"
)

type containsTest struct {
	name       string
	r          Range
	p          Position
	shouldCont bool
}

func runContainsTests(t *testing.T, tests []containsTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if r := tt.r.Contains(tt.p); r != tt.shouldCont {
				t.Errorf("Contains() = %v, want %v", r, tt.shouldCont)
			}
		})
	}
}

func rng(file string, startLine, startCol, endLine, endCol int) Range {
	return NewRange(file, startLine, startCol, endLine, endCol)
}

func rngOffset(file string, startOffset, endOffset int) Range {
	return Range{
		Start: Position{File: file, Offset: startOffset},
		End:   Position{File: file, Offset: endOffset},
	}
}

func rngLine(file string, startLine, endLine int) Range {
	return Range{
		Start: Position{File: file, Line: startLine},
		End:   Position{Line: endLine},
	}
}

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
		{"Positive offset", Position{Offset: 42}, true},
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

	t.Run("column outside range", func(t *testing.T) {
		t.Parallel()

		t.Run("on end line", func(t *testing.T) {
			t.Parallel()

			if rng("a.go", 10, 5, 10, 15).Contains(Pos("a.go", 10, 20)) {
				t.Error("expected column 20 to be outside range 5-15")
			}
		})

		t.Run("on start line", func(t *testing.T) {
			t.Parallel()

			r := rng("a.go", 10, 5, 20, 15)
			if r.Contains(Pos("a.go", 10, 3)) {
				t.Error("expected column 3 to be outside range starting at column 5")
			}
		})
	})

	t.Run("offset-only range", func(t *testing.T) {
		t.Parallel()

		r := rngOffset("a.go", 10, 50)
		if !r.Contains(Position{File: "a.go", Offset: 30}) {
			t.Error("expected Position at offset 30 to be in range 10-50")
		}

		if r.Contains(Position{File: "a.go", Offset: 5}) {
			t.Error("expected Position at offset 5 to be outside range 10-50")
		}
	})

	t.Run("column within range on same line", func(t *testing.T) {
		t.Parallel()

		r := rng("a.go", 10, 5, 10, 15)
		if !r.Contains(Pos("a.go", 10, 10)) {
			t.Error("expected column 10 to be in range 5-15")
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

		r1 := rngOffset("a.go", 50, 0)
		r2 := rngOffset("a.go", 0, 50)
		if !r1.Overlaps(r2) {
			t.Error("ranges should overlap when one end encompasses the other start")
		}
	})

	t.Run("point range overlaps with range", func(t *testing.T) {
		t.Parallel()

		point := rngOffset("a.go", 50, -1)
		span := rngOffset("a.go", 30, 70)
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

		r1 := rngOffset("a.go", 100, 200)
		r2 := rngOffset("a.go", 150, 250)

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

		r1 := rngOffset("a.go", 100, -1)
		r2 := rngOffset("a.go", 100, -1)
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

		r1 := rngOffset("a.go", 100, 150)
		r2 := rngOffset("a.go", 200, 250)
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
			rngLine("", 10, 20),
			rngLine("", 10, 20),
			true,
		},
		{
			"different start line",
			rngLine("", 10, 20),
			rngLine("", 15, 20),
			false,
		},
		{
			"different end line",
			rngLine("", 10, 20),
			rngLine("", 10, 25),
			false,
		},
		{
			"different files same lines",
			rngLine("a.go", 10, 20),
			rngLine("b.go", 10, 20),
			true,
		},
		{
			"same start and end line (single line)",
			rngLine("", 5, 5),
			rngLine("", 5, 5),
			true,
		},
		{
			"zero lines",
			rngLine("", 0, 0),
			rngLine("", 0, 0),
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

	runContainsTests(t, []containsTest{
		{
			"Position before range",
			rngLine("a.go", 10, 20),
			Pos("a.go", 5, 0),
			false,
		},
		{
			"Position after range",
			rngLine("a.go", 10, 20),
			Pos("a.go", 25, 0),
			false,
		},
		{
			"Position within range no column",
			rngLine("a.go", 10, 20),
			Pos("a.go", 15, 0),
			true,
		},
		{
			"zero Position line falls through to offset",
			rngLine("a.go", 10, 0),
			Pos("a.go", 0, 0),
			true,
		},
	})
}
