package finding

import "testing"

func TestRangeLinesEq(t *testing.T) {
	t.Parallel()

	lineEq := func(aStart, aEnd, bStart, bEnd int, want bool) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			a := Range{Start: Position{Line: aStart}, End: Position{Line: aEnd}}
			b := Range{Start: Position{Line: bStart}, End: Position{Line: bEnd}}
			if got := RangeLinesEq(a, b); got != want {
				t.Errorf("RangeLinesEq() = %v, want %v", got, want)
			}
		}
	}

	t.Run("equal lines", lineEq(10, 20, 10, 20, true))
	t.Run("equal lines different columns", lineEq(10, 20, 10, 20, true))
	t.Run("different start lines", lineEq(10, 20, 11, 20, false))
	t.Run("different end lines", lineEq(10, 20, 10, 21, false))
	t.Run("zero values", lineEq(0, 0, 0, 0, true))
}
