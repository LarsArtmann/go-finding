package finding

import (
	"testing"
)

// rangeLine creates a Range with line-based positions in the same file.
func rangeLine(file string, startLine, endLine int) Range {
	return Range{Start: Position{File: file, Line: startLine}, End: Position{Line: endLine}}
}

// rangeOffset creates a Range with offset-based positions in the same file.
func rangeOffset(file string, startOffset, endOffset int) Range {
	return Range{Start: Position{File: file, Offset: startOffset}, End: Position{Offset: endOffset}}
}

// ptrRange returns a pointer to a Range (for test expected values).
func ptrRange(r Range) *Range {
	return &r
}

// posLine creates a Position with a line number.
func posLine(file string, line int) Position {
	return Position{File: file, Line: line}
}

// posLineCol creates a Position with line and column.
func posLineCol(file string, line, col int) Position {
	return Position{File: file, Line: line, Column: col}
}

func TestRangeOverlaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		r1       Range
		r2       Range
		expected bool
	}{
		{
			name:     "same single position",
			r1:       Range{Start: posLineCol("a.go", 10, 5)},
			r2:       Range{Start: posLineCol("a.go", 10, 5)},
			expected: true,
		},
		{
			name:     "overlapping ranges",
			r1:       rangeLine("a.go", 10, 20),
			r2:       rangeLine("a.go", 15, 25),
			expected: true,
		},
		{
			name:     "non-overlapping ranges",
			r1:       rangeLine("a.go", 10, 15),
			r2:       rangeLine("a.go", 20, 25),
			expected: false,
		},
		{
			name:     "different files",
			r1:       Range{Start: posLine("a.go", 10)},
			r2:       Range{Start: posLine("b.go", 10)},
			expected: false,
		},
		{
			name:     "contained range",
			r1:       rangeLine("a.go", 10, 30),
			r2:       rangeLine("a.go", 15, 20),
			expected: true,
		},
		{
			name:     "touching at boundary",
			r1:       rangeLine("a.go", 10, 20),
			r2:       rangeLine("a.go", 20, 30),
			expected: true,
		},
		{
			name:     "offset-based overlap",
			r1:       rangeOffset("a.go", 100, 200),
			r2:       rangeOffset("a.go", 150, 250),
			expected: true,
		},
		{
			name:     "offset-based no overlap",
			r1:       rangeOffset("a.go", 100, 150),
			r2:       rangeOffset("a.go", 200, 250),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r1.Overlaps(tt.r2)
			if got != tt.expected {
				t.Errorf("Overlaps() = %v, want %v", got, tt.expected)
			}

			// Test symmetry
			got2 := tt.r2.Overlaps(tt.r1)
			if got2 != tt.expected {
				t.Errorf("Overlaps() (symmetric) = %v, want %v", got2, tt.expected)
			}
		})
	}
}

func TestRangeIntersection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		r1       Range
		r2       Range
		expected *Range
	}{
		{
			name:     "overlapping ranges",
			r1:       rangeLine("a.go", 10, 20),
			r2:       rangeLine("a.go", 15, 25),
			expected: ptrRange(rangeLine("a.go", 15, 20)),
		},
		{
			name:     "non-overlapping returns nil",
			r1:       rangeLine("a.go", 10, 15),
			r2:       rangeLine("a.go", 20, 25),
			expected: nil,
		},
		{
			name:     "contained range",
			r1:       rangeLine("a.go", 10, 30),
			r2:       rangeLine("a.go", 15, 20),
			expected: ptrRange(rangeLine("a.go", 15, 20)),
		},
		{
			name:     "single point overlap",
			r1:       rangeLine("a.go", 10, 20),
			r2:       rangeLine("a.go", 20, 30),
			expected: ptrRange(Range{Start: Position{File: "a.go", Line: 20}}),
		},
		{
			name:     "different files",
			r1:       Range{Start: posLine("a.go", 10)},
			r2:       Range{Start: posLine("b.go", 10)},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r1.Intersection(tt.r2)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("Intersection() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("Intersection() = nil, want %v", tt.expected)
				return
			}

			if got.Start.Line != tt.expected.Start.Line || got.End.Line != tt.expected.End.Line || got.Start.File != tt.expected.Start.File {
				t.Errorf("Intersection() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRangeAdjacent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		r1       Range
		r2       Range
		expected bool
	}{
		{
			name:     "adjacent at end",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20, Column: 5}},
			r2:       Range{Start: Position{File: "a.go", Line: 20, Column: 5}, End: Position{Line: 30}},
			expected: true,
		},
		{
			name:     "adjacent at start",
			r1:       rangeLine("a.go", 20, 30),
			r2:       rangeLine("a.go", 10, 20),
			expected: true,
		},
		{
			name:     "overlapping not adjacent",
			r1:       rangeLine("a.go", 10, 20),
			r2:       rangeLine("a.go", 15, 25),
			expected: false,
		},
		{
			name:     "gap not adjacent",
			r1:       rangeLine("a.go", 10, 15),
			r2:       rangeLine("a.go", 20, 25),
			expected: false,
		},
		{
			name:     "offset adjacent",
			r1:       rangeOffset("a.go", 100, 200),
			r2:       rangeOffset("a.go", 200, 300),
			expected: true,
		},
		{
			name:     "different files",
			r1:       Range{Start: posLineCol("a.go", 10, 5)},
			r2:       Range{Start: posLineCol("b.go", 10, 5)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r1.Adjacent(tt.r2)
			if got != tt.expected {
				t.Errorf("Adjacent() = %v, want %v", got, tt.expected)
			}

			// Test symmetry
			got2 := tt.r2.Adjacent(tt.r1)
			if got2 != tt.expected {
				t.Errorf("Adjacent() (symmetric) = %v, want %v", got2, tt.expected)
			}
		})
	}
}
