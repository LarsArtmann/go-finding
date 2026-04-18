package finding

import "testing"

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
		{"missing start offset", Range{Start: Position{File: "a.go"}, End: Position{Offset: 100}}, 0},
		{"missing end offset", Range{Start: Position{Offset: 10}, End: Position{}}, 0},
		{"valid range", Range{Start: Position{Offset: 50}, End: Position{Offset: 100}}, 50},
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
