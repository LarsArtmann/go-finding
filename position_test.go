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

type containsTest struct {
	name       string
	r          Range
	p          Position
	shouldCont bool
}

func runContainsTests(t *testing.T, tests []containsTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if r := tt.r.Contains(tt.p); r != tt.shouldCont {
				t.Errorf("Contains() = %v, want %v", r, tt.shouldCont)
			}
		})
	}
}

func rng(file string, startLine, startCol, endLine, endCol int) Range {
	return NewRange(FilePath(file), startLine, startCol, endLine, endCol)
}

func rngOffset(file string, startOffset, endOffset int) Range {
	return Range{
		Start: Position{File: FilePath(file), Offset: startOffset},
		End:   Position{File: FilePath(file), Offset: endOffset},
	}
}

func pos(line int) Position {
	return Position{Line: line}
}

func posOffset(offset int) Position {
	return Position{Offset: offset}
}

func rngPos(startLine, endLine int) Range {
	return Range{Start: pos(startLine), End: pos(endLine)}
}

func TestRangeLineCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want int
	}{
		{"no line info", Range{Start: Position{File: "a.go"}}, 0},
		{"single line no end", Range{Start: Position{File: "a.go", Line: 5}}, 1},
		{"single line with end", rngPos(5, 5), 1},
		{"multi line", rngPos(10, 20), 11},
		{"end < start", rngPos(20, 10), 11},
		{"zero start line", rngPos(0, 5), 0},
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
		{"no offsets", Range{Start: Position{File: "a.go", Line: 1}, End: pos(3)}, 0},
		{
			"missing start offset",
			Range{Start: Position{File: "a.go", Offset: -1}, End: posOffset(100)},
			0,
		},
		{"missing end offset", Range{Start: posOffset(10), End: posOffset(-1)}, 0},
		{"valid range", Range{Start: posOffset(50), End: posOffset(100)}, 50},
		{"zero start offset", Range{Start: posOffset(0), End: posOffset(100)}, 100},
		{"end lt start", Range{Start: posOffset(100), End: posOffset(50)}, 0},
		{"same start and end", Range{Start: posOffset(50), End: posOffset(50)}, 0},
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

func TestPosition_HasOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Position
		want bool
	}{
		{"zero offset (valid)", Position{Offset: 0}, true},
		{"Positive offset", Position{Offset: 42}, true},
		{"negative offset (unset)", Position{Offset: -1}, false},
		{"default zero value", Position{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.p.HasOffset(); got != tt.want {
				t.Errorf("HasOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeContains_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("different file returns false", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go", Line: 5}, End: Position{Line: 10}}

		p := Position{File: "b.go", Line: 7}
		if r.Contains(p) {
			t.Error("expected Contains to return false for different file")
		}
	})

	t.Run("column outside range", func(t *testing.T) {
		t.Parallel()

		t.Run("on end line", func(t *testing.T) {
			t.Parallel()

			if rng("a.go", 10, 5, 10, 15).Contains(Pos("a.go", 10, 20)) {
				t.Error("expected column 20 to be outside range 5-15")
			}
		})

		t.Run("on start line", func(t *testing.T) {
			t.Parallel()

			r := rng("a.go", 10, 5, 20, 15)
			if r.Contains(Pos("a.go", 10, 3)) {
				t.Error("expected column 3 to be outside range starting at column 5")
			}
		})
	})

	t.Run("offset-only range", func(t *testing.T) {
		t.Parallel()

		r := rngOffset("a.go", 10, 50)
		assertRangeContains(t, r, 30, "a.go", true)
		assertRangeContains(t, r, 5, "a.go", false)
	})

	t.Run("column within range on same line", func(t *testing.T) {
		t.Parallel()

		r := rng("a.go", 10, 5, 10, 15)
		if !r.Contains(Pos("a.go", 10, 10)) {
			t.Error("expected column 10 to be in range 5-15")
		}
	})

	t.Run("no line info uses offset", func(t *testing.T) {
		t.Parallel()

		r := Range{
			Start: Position{File: "a.go", Offset: 100},
			End:   Position{Offset: 200},
		}
		assertRangeContains(t, r, 150, "a.go", true)
	})

	t.Run("no line no offset uses offset check", func(t *testing.T) {
		t.Parallel()

		r := Range{Start: Position{File: "a.go"}}
		p := Position{File: "a.go"}
		// Both have Offset=0 (default), which means containsByOffset is reached
		// since hasLineRange is false (both Line=0). Offset=0 >= 0, so the check runs.
		if !r.Contains(p) {
			t.Error("default Offset=0 means both have offset, so containsByOffset runs")
		}
	})
}

func TestRangeEndOrStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want Position
	}{
		{
			name: "end set with line info returns end",
			r:    Range{Start: Pos("a.go", 1, 1), End: Pos("a.go", 2, 5)},
			want: Pos("a.go", 2, 5),
		},
		{
			name: "end unset (Line=0) returns start",
			r:    Range{Start: Pos("a.go", 1, 1)},
			want: Pos("a.go", 1, 1),
		},
		{
			name: "end with line=0 but offset set still returns start (line check)",
			r:    Range{Start: Pos("a.go", 1, 1), End: Position{Offset: 10}},
			want: Pos("a.go", 1, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.EndOrStart(); !got.Equal(tt.want) {
				t.Errorf("EndOrStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeEndOffsetOrStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want int
	}{
		{
			name: "end with offset returns end offset",
			r:    Range{Start: Position{File: "a.go", Offset: 5}, End: Position{File: "a.go", Offset: 15}},
			want: 15,
		},
		{
			name: "end unset (Offset=-1) returns start offset",
			r:    Range{Start: Position{File: "a.go", Offset: 5}, End: Position{File: "a.go", Offset: -1}},
			want: 5,
		},
		{
			name: "end with offset=0 returns 0",
			r:    Range{Start: Position{File: "a.go", Offset: 5}, End: Position{File: "a.go", Offset: 0}},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.EndOffsetOrStart(); got != tt.want {
				t.Errorf("EndOffsetOrStart() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFilePos(t *testing.T) {
	t.Parallel()

	p := FilePos("config.yaml")

	if p.File != FilePath("config.yaml") {
		t.Errorf("File = %q, want %q", p.File, "config.yaml")
	}

	if p.Line != 0 {
		t.Errorf("Line = %d, want 0", p.Line)
	}

	if p.Column != 0 {
		t.Errorf("Column = %d, want 0", p.Column)
	}

	if p.Offset != OffsetUnknown {
		t.Errorf("Offset = %d, want %d (OffsetUnknown)", p.Offset, OffsetUnknown)
	}

	if !p.HasFile() {
		t.Error("HasFile() = false, want true")
	}
}

func TestFileLevelPosition_Validate(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "tool:rule:config.yaml",
		Rule:     "rule",
		ToolName: "tool",
		Message:  "missing required field",
		Severity: SeverityError,
		Position: FilePos("config.yaml"),
	}

	if err := f.Validate(); err != nil {
		t.Errorf("file-level finding Validate() = %v, want nil", err)
	}

	if !f.IsValid() {
		t.Error("file-level finding IsValid() = false, want true")
	}
}

func TestEmptyFile_Validate(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "1",
		Rule:     "rule",
		ToolName: "tool",
		Message:  "msg",
		Severity: SeverityError,
		Position: Position{},
	}

	if err := f.Validate(); err == nil {
		t.Error("empty-file finding Validate() = nil, want error")
	}
}
