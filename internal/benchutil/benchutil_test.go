package benchutil

import (
	"bytes"
	"testing"
)

func TestFindOldOccurrences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content []byte
		want    []int
	}{
		{
			name:    "single occurrence",
			content: []byte("old()"),
			want:    []int{0},
		},
		{
			name:    "multiple occurrences",
			content: []byte("old() and old() again"),
			want:    []int{0, 10},
		},
		{
			name:    "no occurrences",
			content: []byte("nothing here"),
			want:    nil,
		},
		{
			name:    "empty content",
			content: []byte{},
			want:    nil,
		},
		{
			name:    "adjacent occurrences",
			content: []byte("old()old()"),
			want:    []int{0, 5},
		},
		{
			name:    "occurrence at end",
			content: []byte("prefix old()"),
			want:    []int{7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FindOldOccurrences(tt.content)
			if !equalInts(got, tt.want) {
				t.Errorf("FindOldOccurrences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOffsetToLineNumber(t *testing.T) {
	t.Parallel()

	content := []byte("line1\nline2\nline3\n")

	tests := []struct {
		name   string
		offset int
		want   int
	}{
		{name: "start of file", offset: 0, want: 1},
		{name: "end of line 1", offset: 5, want: 1},
		{name: "start of line 2", offset: 6, want: 2},
		{name: "mid line 2", offset: 8, want: 2},
		{name: "start of line 3", offset: 12, want: 3},
		{name: "empty content offset 0", offset: 0, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := content
			if tt.name == "empty content offset 0" {
				c = []byte{}
			}

			got := OffsetToLineNumber(c, tt.offset)
			if got != tt.want {
				t.Errorf("OffsetToLineNumber(%d) = %d, want %d", tt.offset, got, tt.want)
			}
		})
	}
}

func TestColumnOfOffset(t *testing.T) {
	t.Parallel()

	content := []byte("hello\nworld\n")

	tests := []struct {
		name   string
		offset int
		want   int
	}{
		{name: "first column line 1", offset: 0, want: 1},
		{name: "mid line 1", offset: 2, want: 3},
		{name: "end of line 1", offset: 4, want: 5},
		{name: "newline char (end of line 1)", offset: 5, want: 6},
		{name: "first column line 2", offset: 6, want: 1},
		{name: "mid line 2", offset: 8, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ColumnOfOffset(content, tt.offset)
			if got != tt.want {
				t.Errorf("ColumnOfOffset(%d) = %d, want %d", tt.offset, got, tt.want)
			}
		})
	}
}

func TestColumnOfOffset_StartOfFile(t *testing.T) {
	t.Parallel()

	got := ColumnOfOffset([]byte("x"), 0)
	if got != 1 {
		t.Errorf("ColumnOfOffset(0) = %d, want 1", got)
	}
}

func TestPickEvenly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		positions []int
		n         int
		want      []int
	}{
		{name: "empty positions", positions: nil, n: 5, want: nil},
		{name: "n is zero", positions: []int{1, 2, 3}, n: 0, want: nil},
		{name: "n negative", positions: []int{1, 2, 3}, n: -1, want: nil},
		{name: "n equals len", positions: []int{10, 20, 30}, n: 3, want: []int{10, 20, 30}},
		{name: "n greater than len", positions: []int{10, 20}, n: 5, want: []int{10, 20}},
		{name: "n less than len", positions: []int{0, 1, 2, 3, 4, 5}, n: 2, want: []int{0, 3}},
		{name: "single position", positions: []int{42}, n: 1, want: []int{42}},
		{name: "single position n3", positions: []int{42}, n: 3, want: []int{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := PickEvenly(tt.positions, tt.n)
			if !equalInts(got, tt.want) {
				t.Errorf("PickEvenly(%v, %d) = %v, want %v", tt.positions, tt.n, got, tt.want)
			}
		})
	}
}

// TestBenchutilIntegration verifies the 4 functions work together correctly,
// reproducing the exact pattern used in the benchmark files.
func TestBenchutilIntegration(t *testing.T) {
	t.Parallel()

	content := bytes.Join([][]byte{
		[]byte("package main\n\nfunc main() {"),
		[]byte("\told()"),
		[]byte("\told()"),
		[]byte("\told()"),
		[]byte("}"),
	}, []byte("\n"))

	occurrences := FindOldOccurrences(content)
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences, got %d", len(occurrences))
	}

	selected := PickEvenly(occurrences, 2)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(selected))
	}

	for _, pos := range selected {
		line := OffsetToLineNumber(content, pos)
		col := ColumnOfOffset(content, pos)

		if line < 1 {
			t.Errorf("line = %d, want >= 1", line)
		}

		if col < 1 {
			t.Errorf("col = %d, want >= 1", col)
		}

		got := content[pos : pos+len("old()")]
		if !bytes.Equal(got, []byte("old()")) {
			t.Errorf("position %d = %q, want the needle", pos, got)
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
