package finding

import (
	"testing"
)

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
			r1:       Range{Start: Position{File: "a.go", Line: 10, Column: 5}},
			r2:       Range{Start: Position{File: "a.go", Line: 10, Column: 5}},
			expected: true,
		},
		{
			name:     "overlapping ranges",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			r2:       Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 25}},
			expected: true,
		},
		{
			name:     "non-overlapping ranges",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 15}},
			r2:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 25}},
			expected: false,
		},
		{
			name:     "different files",
			r1:       Range{Start: Position{File: "a.go", Line: 10}},
			r2:       Range{Start: Position{File: "b.go", Line: 10}},
			expected: false,
		},
		{
			name:     "contained range",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 30}},
			r2:       Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 20}},
			expected: true,
		},
		{
			name:     "touching at boundary",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			r2:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 30}},
			expected: true,
		},
		{
			name:     "offset-based overlap",
			r1:       Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: 200}},
			r2:       Range{Start: Position{File: "a.go", Offset: 150}, End: Position{Offset: 250}},
			expected: true,
		},
		{
			name:     "offset-based no overlap",
			r1:       Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: 150}},
			r2:       Range{Start: Position{File: "a.go", Offset: 200}, End: Position{Offset: 250}},
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
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			r2:       Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 25}},
			expected: &Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 20}},
		},
		{
			name:     "non-overlapping returns nil",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 15}},
			r2:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 25}},
			expected: nil,
		},
		{
			name:     "contained range",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 30}},
			r2:       Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 20}},
			expected: &Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 20}},
		},
		{
			name:     "single point overlap",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			r2:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 30}},
			expected: &Range{Start: Position{File: "a.go", Line: 20}, End: Position{}},
		},
		{
			name:     "different files",
			r1:       Range{Start: Position{File: "a.go", Line: 10}},
			r2:       Range{Start: Position{File: "b.go", Line: 10}},
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
			r1:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 30}},
			r2:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			expected: true,
		},
		{
			name:     "overlapping not adjacent",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 20}},
			r2:       Range{Start: Position{File: "a.go", Line: 15}, End: Position{Line: 25}},
			expected: false,
		},
		{
			name:     "gap not adjacent",
			r1:       Range{Start: Position{File: "a.go", Line: 10}, End: Position{Line: 15}},
			r2:       Range{Start: Position{File: "a.go", Line: 20}, End: Position{Line: 25}},
			expected: false,
		},
		{
			name:     "offset adjacent",
			r1:       Range{Start: Position{File: "a.go", Offset: 100}, End: Position{Offset: 200}},
			r2:       Range{Start: Position{File: "a.go", Offset: 200}, End: Position{Offset: 300}},
			expected: true,
		},
		{
			name:     "different files",
			r1:       Range{Start: Position{File: "a.go", Line: 10, Column: 5}},
			r2:       Range{Start: Position{File: "b.go", Line: 10, Column: 5}},
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
