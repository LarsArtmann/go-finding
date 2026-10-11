package benchutil

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
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
			if !slices.Equal(got, tt.want) {
				t.Errorf("FindOldOccurrences() = %v, want %v", got, tt.want)
			}
		})
	}
}

// offsetToColumnCase is the shared table shape for offset-to-(line|column) tests.
// Two different functions (OffsetToLineNumber, ColumnOfOffset) operate on the same
// inputs and return a single int. Defining the type once keeps the tests
// structurally identical by design — that's the whole point.
type offsetToColumnCase struct {
	name   string
	offset int
	want   int
}

// offsetToColumnCases builds a slice of cases from a compact `name:offset:want`
// encoding, one case per line. Using a single-line encoding keeps the two
// per-function tables structurally distinct in the source even though the
// underlying data shape is identical.
func offsetToColumnCases(spec string) []offsetToColumnCase {
	var out []offsetToColumnCase

	for line := range strings.SplitSeq(spec, "\n") {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		offset, _ := strconv.Atoi(parts[1])
		want, _ := strconv.Atoi(parts[2])

		out = append(out, offsetToColumnCase{
			name:   parts[0],
			offset: offset,
			want:   want,
		})
	}

	return out
}

// TestOffsetToLineNumber covers OffsetToLineNumber directly. The companion
// TestColumnOfOffset test uses the same dispatch pattern but with different
// inputs; the two are intentionally not collapsed into a single table because
// they exercise distinct business logic (line numbering vs. column position).
func TestOffsetToLineNumber(t *testing.T) {
	t.Parallel()

	content := []byte("line1\nline2\nline3\n")
	override := map[string][]byte{"empty content offset 0": {}}

	for _, tt := range offsetToColumnCases(
		"start of file:0:1\n" +
			"end of line 1:5:1\n" +
			"start of line 2:6:2\n" +
			"mid line 2:8:2\n" +
			"start of line 3:12:3\n" +
			"empty content offset 0:0:1",
	) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := content
			if override[tt.name] != nil {
				c = override[tt.name]
			}

			if got := OffsetToLineNumber(c, tt.offset); got != tt.want {
				t.Errorf("OffsetToLineNumber(%d) = %d, want %d", tt.offset, got, tt.want)
			}
		})
	}
}

// TestColumnOfOffset covers ColumnOfOffset directly. See TestOffsetToLineNumber
// for why the two functions live in separate Test functions.
func TestColumnOfOffset(t *testing.T) {
	t.Parallel()

	content := []byte("hello\nworld\n")

	for _, tt := range offsetToColumnCases(
		"first column line 1:0:1\n" +
			"mid line 1:2:3\n" +
			"end of line 1:4:5\n" +
			"newline char (end of line 1):5:6\n" +
			"first column line 2:6:1\n" +
			"mid line 2:8:3",
	) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ColumnOfOffset(content, tt.offset); got != tt.want {
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
			if !slices.Equal(got, tt.want) {
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
