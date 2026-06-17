package finding

import (
	"testing"
)

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

		assertRangeStartOffset(t, *got, 150)
		assertRangeEndOffset(t, *got, 200)
	})

	t.Run("offset with negative end defaults to start", func(t *testing.T) {
		t.Parallel()

		r1 := rngOffset("a.go", 100, -1)
		r2 := rngOffset("a.go", 100, -1)

		got := r1.Intersection(r2)
		if got == nil {
			t.Fatal("expected non-nil intersection for same point")
		}

		assertRangeStartOffset(t, *got, 100)
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
			rangeLine("", 10, 20),
			rangeLine("", 10, 20),
			true,
		},
		{
			"different start line",
			rangeLine("", 10, 20),
			rangeLine("", 15, 20),
			false,
		},
		{
			"different end line",
			rangeLine("", 10, 20),
			rangeLine("", 10, 25),
			false,
		},
		{
			"different files same lines",
			rangeLine("a.go", 10, 20),
			rangeLine("b.go", 10, 20),
			true,
		},
		{
			"same start and end line (single line)",
			rangeLine("", 5, 5),
			rangeLine("", 5, 5),
			true,
		},
		{
			"zero lines",
			rangeLine("", 0, 0),
			rangeLine("", 0, 0),
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
			rangeLine("a.go", 10, 20),
			Pos("a.go", 5, 0),
			false,
		},
		{
			"Position after range",
			rangeLine("a.go", 10, 20),
			Pos("a.go", 25, 0),
			false,
		},
		{
			"Position within range no column",
			rangeLine("a.go", 10, 20),
			Pos("a.go", 15, 0),
			true,
		},
		{
			"Position with no line and no offset cannot be contained",
			rangeLine("a.go", 10, 0),
			Pos("a.go", 0, 0),
			false,
		},
		{
			"range with no end line uses column check",
			rangeLine("a.go", 10, 0),
			Pos("a.go", 10, 0),
			true,
		},
	})
}

func TestRangeContains_ZeroValue(t *testing.T) {
	t.Parallel()

	var zero Range
	if zero.Contains(Position{File: "a.go", Line: 1}) {
		t.Error("zero-value Range should not contain any position")
	}
}
