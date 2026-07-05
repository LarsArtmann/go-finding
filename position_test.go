package finding

import (
	"testing"
)

const posTestDiffFiles = "different files"

func posLine(file string, line int) Position {
	return Position{File: FilePath(file), Line: line}
}

func posLineCol(file string, line, col int) Position {
	return Position{File: FilePath(file), Line: line, Column: col}
}

func rangeLine(file string, startLine, endLine int) Range {
	return Range{
		Start: Position{File: FilePath(file), Line: startLine, Offset: -1},
		End:   Position{Line: endLine, Offset: -1},
	}
}

func rangeOffset(file string, startOffset, endOffset int) Range {
	return Range{Start: Position{File: FilePath(file), Offset: startOffset}, End: Position{Offset: endOffset}}
}

type overlapCase struct {
	name     string
	r1       Range
	r2       Range
	expected bool
}

func overlapCases() []overlapCase {
	return []overlapCase{
		{
			name:     "same single position",
			r1:       Range{Start: posLineCol("a.go", 10, 5)},
			r2:       Range{Start: posLineCol("a.go", 10, 5)},
			expected: true,
		},
		{"overlapping ranges", rangeLine("a.go", 10, 20), rangeLine("a.go", 15, 25), true},
		{"non-overlapping ranges", rangeLine("a.go", 10, 15), rangeLine("a.go", 20, 25), false},
		{
			posTestDiffFiles,
			Range{Start: posLine("a.go", 10)},
			Range{Start: posLine("b.go", 10)},
			false,
		},
		{"contained range", rangeLine("a.go", 10, 30), rangeLine("a.go", 15, 20), true},
		{"touching at boundary", rangeLine("a.go", 10, 20), rangeLine("a.go", 20, 30), true},
		{
			"offset-based overlap",
			rangeOffset("a.go", 100, 200),
			rangeOffset("a.go", 150, 250),
			true,
		},
		{
			"offset-based no overlap",
			rangeOffset("a.go", 100, 150),
			rangeOffset("a.go", 200, 250),
			false,
		},
	}
}

func TestRangeOverlaps(t *testing.T) {
	t.Parallel()

	for _, tt := range overlapCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r1.Overlaps(tt.r2); got != tt.expected {
				t.Errorf("Overlaps() = %v, want %v", got, tt.expected)
			}

			if got := tt.r2.Overlaps(tt.r1); got != tt.expected {
				t.Errorf("Overlaps() reversed = %v, want %v", got, tt.expected)
			}
		})
	}
}

type intersectionCase struct {
	name     string
	r1       Range
	r2       Range
	expected *Range
}

func intersectionCases() []intersectionCase {
	return []intersectionCase{
		{
			"overlapping ranges",
			rangeLine("a.go", 10, 20),
			rangeLine("a.go", 15, 25),
			new(rangeLine("a.go", 15, 20)),
		},
		{"non-overlapping returns nil", rangeLine("a.go", 10, 15), rangeLine("a.go", 20, 25), nil},
		{
			"contained range",
			rangeLine("a.go", 10, 30),
			rangeLine("a.go", 15, 20),
			new(rangeLine("a.go", 15, 20)),
		},
		{
			"single point overlap",
			rangeLine("a.go", 10, 20),
			rangeLine("a.go", 20, 30),
			new(Range{Start: Position{File: "a.go", Line: 20}}),
		},
		{
			posTestDiffFiles,
			Range{Start: posLine("a.go", 10)},
			Range{Start: posLine("b.go", 10)},
			nil,
		},
	}
}

func TestRangeIntersection(t *testing.T) {
	t.Parallel()

	for _, tt := range intersectionCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r1.Intersection(tt.r2)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}

				return
			}

			if got == nil {
				t.Fatalf("expected non-nil intersection")
			}

			AssertRangesEq(t, *got, *tt.expected)
		})
	}
}

func adjacentCases() []overlapCase {
	return []overlapCase{
		{
			name: "adjacent at end",
			r1: Range{
				Start: Position{File: "a.go", Line: 10},
				End:   Position{Line: 20, Column: 5},
			},
			r2: Range{
				Start: Position{File: "a.go", Line: 20, Column: 5},
				End:   Position{Line: 30},
			},
			expected: true,
		},
		{"adjacent at start", rangeLine("a.go", 20, 30), rangeLine("a.go", 10, 20), true},
		{"overlapping not adjacent", rangeLine("a.go", 10, 20), rangeLine("a.go", 15, 25), false},
		{"gap not adjacent", rangeLine("a.go", 10, 15), rangeLine("a.go", 20, 25), false},
		{"offset adjacent", rangeOffset("a.go", 100, 200), rangeOffset("a.go", 200, 300), true},
		{
			posTestDiffFiles,
			Range{Start: posLineCol("a.go", 10, 5)},
			Range{Start: posLineCol("b.go", 10, 5)},
			false,
		},
	}
}

func TestRangeAdjacent(t *testing.T) {
	t.Parallel()

	for _, tt := range adjacentCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r1.Adjacent(tt.r2); got != tt.expected {
				t.Errorf("Adjacent() = %v, want %v", got, tt.expected)
			}

			if got := tt.r2.Adjacent(tt.r1); got != tt.expected {
				t.Errorf("Adjacent() reversed = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPosition_IsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"zero value has offset 0", Position{}, false},
		{"file only", Position{File: "a.go"}, false},
		{"line only", Position{Line: 1}, false},
		{"offset -1 unset", Position{Offset: -1}, true},
		{"all unset sentinel", Position{File: "", Line: 0, Column: 0, Offset: -1}, true},
		{"full position", Position{File: "a.go", Line: 1, Column: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.pos.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPosition_HasLocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"zero value", Position{}, false},
		{"file only", Position{File: "a.go"}, false},
		{"line only", Position{Line: 1}, false},
		{"file and line", Position{File: "a.go", Line: 1}, true},
		{"full position", Position{File: "a.go", Line: 5, Column: 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.pos.HasLocation(); got != tt.want {
				t.Errorf("HasLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRange_IsSingleLine(t *testing.T) {
	t.Parallel()

	start := Position{File: "a.go", Line: 5}
	tests := []struct {
		name string
		r    Range
		want bool
	}{
		{"no start line", Range{}, false},
		{"single position", Range{Start: start}, true},
		{"same line", Range{Start: start, End: Position{Line: 5}}, true},
		{"multi line", Range{Start: start, End: Position{Line: 10}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.IsSingleLine(); got != tt.want {
				t.Errorf("IsSingleLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRange_IsInverted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want bool
	}{
		{"zero range", Range{}, false},
		{"start only", Range{Start: Position{Line: 1}}, false},
		{"valid range", Range{Start: Position{Line: 1}, End: Position{Line: 5}}, false},
		{"same line", Range{Start: Position{Line: 3}, End: Position{Line: 3}}, false},
		{"inverted lines", Range{Start: Position{Line: 5}, End: Position{Line: 1}}, true},
		{
			"inverted columns",
			Range{Start: Position{Line: 3, Column: 10}, End: Position{Line: 3, Column: 2}},
			true,
		},
		{
			"same line same column",
			Range{Start: Position{Line: 3, Column: 5}, End: Position{Line: 3, Column: 5}},
			false,
		},
		{"end line zero", Range{Start: Position{Line: 5}, End: Position{Line: 0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.IsInverted(); got != tt.want {
				t.Errorf("IsInverted() = %v, want %v", got, tt.want)
			}
		})
	}
}
