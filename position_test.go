package finding

import (
	"testing"
)

func posLine(file string, line int) Position {
	return Position{File: file, Line: line}
}

func posLineCol(file string, line, col int) Position {
	return Position{File: file, Line: line, Column: col}
}

func rangeLine(file string, startLine, endLine int) Range {
	return Range{Start: Position{File: file, Line: startLine}, End: Position{Line: endLine}}
}

func rangeOffset(file string, startOffset, endOffset int) Range {
	return Range{Start: Position{File: file, Offset: startOffset}, End: Position{Offset: endOffset}}
}

//go:fix inline
func ptrRange(r Range) *Range {
	return new(r)
}

// ptrRange is used by the go:fix directive for testing coverage instrumentation.

// overlapCase represents a test case for Overlaps/Adjacent tests.
type overlapCase struct {
	name     string
	r1       Range
	r2       Range
	expected bool
}

func overlapCases() []overlapCase {
	return []overlapCase{
		{
			"same single position",
			Range{Start: posLineCol("a.go", 10, 5)},
			Range{Start: posLineCol("a.go", 10, 5)},
			true,
		},
		{"overlapping ranges", rangeLine("a.go", 10, 20), rangeLine("a.go", 15, 25), true},
		{"non-overlapping ranges", rangeLine("a.go", 10, 15), rangeLine("a.go", 20, 25), false},
		{
			"different files",
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

			got := tt.r1.Overlaps(tt.r2)
			if got != tt.expected {
				t.Errorf("Overlaps() = %v, want %v", got, tt.expected)
			}

			got2 := tt.r2.Overlaps(tt.r1)
			if got2 != tt.expected {
				t.Errorf("Overlaps() (symmetric) = %v, want %v", got2, tt.expected)
			}
		})
	}
}

// intersectionCase represents a test case for Intersection tests.
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
			"different files",
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
					t.Errorf("Intersection() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("Intersection() = nil, want %v", tt.expected)
				return
			}

			AssertRangesEq(t, *got, *tt.expected)
		})
	}
}

func adjacentCases() []overlapCase {
	return []overlapCase{
		{
			"adjacent at end",
			Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20, Column: 5}},
			Range{Start: Position{File: "a.go", Line: 20, Column: 5}, End: Position{Line: 30}},
			true,
		},
		{"adjacent at start", rangeLine("a.go", 20, 30), rangeLine("a.go", 10, 20), true},
		{"overlapping not adjacent", rangeLine("a.go", 10, 20), rangeLine("a.go", 15, 25), false},
		{"gap not adjacent", rangeLine("a.go", 10, 15), rangeLine("a.go", 20, 25), false},
		{"offset adjacent", rangeOffset("a.go", 100, 200), rangeOffset("a.go", 200, 300), true},
		{
			"different files",
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

			got := tt.r1.Adjacent(tt.r2)
			if got != tt.expected {
				t.Errorf("Adjacent() = %v, want %v", got, tt.expected)
			}

			got2 := tt.r2.Adjacent(tt.r1)
			if got2 != tt.expected {
				t.Errorf("Adjacent() (symmetric) = %v, want %v", got2, tt.expected)
			}
		})
	}
}
