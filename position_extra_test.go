package finding

import (
	"testing"
)

func TestPositionIsValidExtra(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Position
		want bool
	}{
		{"with file", Position{File: "a.go", Line: 10}, true},
		{"empty file", Position{File: "", Line: 10}, false},
		{"file only", Position{File: "a.go"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionStringExtra(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Position
		want string
	}{
		{"file:line:col", Position{File: "a.go", Line: 10, Column: 5}, "a.go:10:5"},
		{"file:line", Position{File: "a.go", Line: 10}, "a.go:10"},
		{"file only", Position{File: "a.go"}, "a.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.p.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRangeIsValidExtra(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want bool
	}{
		{"valid start", Range{Start: Position{File: "a.go", Line: 10}}, true},
		{"empty start", Range{Start: Position{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeHasEnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want bool
	}{
		{"with end", Range{Start: Position{File: "a.go"}, End: Position{Line: 20}}, true},
		{"without end", Range{Start: Position{File: "a.go"}}, false},
		{"end line zero", Range{Start: Position{File: "a.go"}, End: Position{Line: 0}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.HasEnd(); got != tt.want {
				t.Errorf("HasEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeContainsExtra(t *testing.T) {
	t.Parallel()

	r := NewRange("a.go", 10, 5, 20, 15)

	tests := []struct {
		name string
		p    Position
		want bool
	}{
		{"inside", Position{File: "a.go", Line: 15, Column: 10}, true},
		{"start position", Position{File: "a.go", Line: 10, Column: 5}, true},
		{"end position", Position{File: "a.go", Line: 20, Column: 15}, true},
		{"different file", Position{File: "b.go", Line: 15}, false},
		{"before range", Position{File: "a.go", Line: 5}, false},
		{"after range", Position{File: "a.go", Line: 25}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := r.Contains(tt.p); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeRangeWithOffsets(file string, startOff, endOff int) Range {
	return Range{
		Start: Position{File: file, Offset: startOff},
		End:   Position{Offset: endOff},
	}
}

func TestRangeIntersectionByOffset(t *testing.T) {
	t.Parallel()

	rng1 := makeRangeWithOffsets("a.go", 100, 200)
	rng2 := makeRangeWithOffsets("a.go", 150, 250)

	inter := rng1.Intersection(rng2)
	if inter == nil {
		t.Fatal("expected non-nil intersection")
	}

	assertReportField(t, "Start.Offset", inter.Start.Offset, 150)
	assertReportField(t, "End.Offset", inter.End.Offset, 200)
}
