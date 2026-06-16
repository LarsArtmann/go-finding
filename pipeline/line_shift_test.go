package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/onsi/gomega"
)

func TestNewLineShiftMap_NoEdits(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	m := NewLineShiftMap([]byte("hello\nworld\n"), nil)
	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1))
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2))
}

func TestNewLineShiftMap_PureInsertion(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// "line1\nline2\nline3\nline4\n"
	// Insert 2 new lines before "line3" (offset 12)
	original := []byte("line1\nline2\nline3\nline4\n")

	edit := FixEdit{
		Offset:      12,
		Length:      0, // pure insertion
		Replacement: []byte("new1\nnew2\n"),
	}

	m := NewLineShiftMap(original, []FixEdit{edit})
	// effectOffset = 12 (pure insertion: shift starts at offset)
	// lineOffsets = [0, 6, 12, 18]

	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1)) // offset 0 < 12: not shifted
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2)) // offset 6 < 12: not shifted
	g.Expect(m.ShiftedLine(3)).To(gomega.Equal(5)) // offset 12 <= 12: shifted +2
	g.Expect(m.ShiftedLine(4)).To(gomega.Equal(6)) // offset 18 <= 18: shifted +2
}

func TestNewLineShiftMap_ReplacementNoLineChange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	original := []byte("line1\nline2\n")

	edit := FixEdit{
		Offset:      6,
		Length:      6,
		Replacement: []byte("other\n"),
	}

	m := NewLineShiftMap(original, []FixEdit{edit})
	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1))
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2))
	g.Expect(m.Entries()).To(gomega.BeEmpty()) // delta=0, no entries
}

func TestNewLineShiftMap_ReplacementAddsLines(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Replace "line2\n" (1 line) with "line2a\nline2b\n" (2 lines)
	original := []byte("line1\nline2\nline3\n")

	edit := FixEdit{
		Offset:      6,
		Length:      6,
		Replacement: []byte("line2a\nline2b\n"),
	}

	m := NewLineShiftMap(original, []FixEdit{edit})
	// effectOffset = 6+6 = 12 (replacement: shift after replaced region)
	// lineOffsets = [0, 6, 12]

	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1)) // offset 0 < 12: not shifted
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2)) // offset 6 < 12: not shifted (replacement line)
	g.Expect(m.ShiftedLine(3)).To(gomega.Equal(4)) // offset 12 <= 12: shifted +1
}

func TestNewLineShiftMap_Deletion(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Delete "line3\n" (offset 12, length 6)
	original := []byte("line1\nline2\nline3\nline4\n")

	edit := FixEdit{
		Offset: 12,
		Length: 6,
	}

	m := NewLineShiftMap(original, []FixEdit{edit})
	// effectOffset = 18, delta = -1
	// lineOffsets = [0, 6, 12, 18]

	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1)) // offset 0 < 18: not shifted
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2)) // offset 6 < 18: not shifted
	// Line 3 (offset 12) is within the deleted range; returns original
	g.Expect(m.ShiftedLine(4)).To(gomega.Equal(3)) // offset 18 <= 18: shifted -1
}

func TestNewLineShiftMap_MultipleEdits(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	original := []byte("line1\nline2\nline3\nline4\nline5\n")

	// Replace line 2 with 2 lines: effectOffset=12, delta=+1
	edit1 := FixEdit{
		Offset:      6,
		Length:      6,
		Replacement: []byte("line2a\nline2b\n"),
	}

	// Delete line 5: effectOffset=30, delta=-1
	edit2 := FixEdit{
		Offset: 24,
		Length: 6,
	}

	m := NewLineShiftMap(original, []FixEdit{edit1, edit2})

	g.Expect(m.ShiftedLine(1)).To(gomega.Equal(1)) // offset 0: no entries affect
	g.Expect(m.ShiftedLine(2)).To(gomega.Equal(2)) // offset 6 < 12: not shifted
	g.Expect(m.ShiftedLine(3)).To(gomega.Equal(4)) // offset 12 <= 12: +1, 12 < 30: not -1
	g.Expect(m.ShiftedLine(4)).To(gomega.Equal(5)) // offset 18 <= 18: +1, 18 < 30: not -1
	// Line 5 (offset 24) is within deleted range of edit2
}

func TestNewLineShiftMap_Entries(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	original := []byte("a\nb\nc\n")
	edit := FixEdit{
		Offset:      2,
		Length:      0,
		Replacement: []byte("x\ny\n"),
	}

	m := NewLineShiftMap(original, []FixEdit{edit})
	g.Expect(m.Entries()).To(gomega.HaveLen(1))
	g.Expect(m.Entries()[0].Delta).To(gomega.Equal(2))
}

func TestLineShiftMap_ShiftedPosition_LineOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Insert 2 lines before line 3.
	original := []byte("line1\nline2\nline3\nline4\n")
	edit := FixEdit{Offset: 12, Replacement: []byte("new1\nnew2\n")}
	m := NewLineShiftMap(original, []FixEdit{edit})

	// Line 4 shifted to 6, column unchanged.
	pos := finding.Position{Line: 4, Column: 5}
	shifted := m.ShiftedPosition(pos)
	g.Expect(shifted.Line).To(gomega.Equal(6))
	g.Expect(shifted.Column).To(gomega.Equal(5))
}

func TestLineShiftMap_ShiftedPosition_ColumnShift(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Single-line edit: replace "ab" with "abcde" at offset 6 on line 2.
	// "line1\nab here\nline3\n"
	// Line 2 starts at offset 6. "ab" is at offset 6-7.
	original := []byte("line1\nab here\nline3\n")
	edit := FixEdit{Offset: 6, Length: 2, Replacement: []byte("abcde")}
	m := NewLineShiftMap(original, []FixEdit{edit})

	// effectOffset = 6+2 = 8. byteDelta = 5-2 = 3. delta = 0.
	// Position on line 2, column 8 (after "ab here" - the space is col 3, 'h' col 4).
	// Original line 2: "ab here" — col 1='a', 2='b', 3=' ', 4='h'...
	// A position at col 4 ('h') is at byte offset 6+3 = 9. effectOffset 8 <= 9.
	// Same line, so column shifts by +3 → col 7.
	pos := finding.Position{Line: 2, Column: 4}
	shifted := m.ShiftedPosition(pos)
	g.Expect(shifted.Line).To(gomega.Equal(2)) // no line change
	g.Expect(shifted.Column).To(gomega.Equal(7))

	// Position on line 3: line unchanged, column unchanged.
	pos3 := finding.Position{Line: 3, Column: 1}
	shifted3 := m.ShiftedPosition(pos3)
	g.Expect(shifted3.Line).To(gomega.Equal(3))
	g.Expect(shifted3.Column).To(gomega.Equal(1))
}

func TestLineShiftMap_ShiftedRange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	// Insert 1 line before line 3.
	original := []byte("line1\nline2\nline3\nline4\n")
	edit := FixEdit{Offset: 12, Replacement: []byte("new\n")}
	m := NewLineShiftMap(original, []FixEdit{edit})

	r := &finding.Range{
		Start: finding.Position{Line: 3, Column: 1},
		End:   finding.Position{Line: 4, Column: 5},
	}
	shifted := m.ShiftedRange(r)
	g.Expect(shifted.Start.Line).To(gomega.Equal(4))
	g.Expect(shifted.End.Line).To(gomega.Equal(5))
	g.Expect(shifted.Start.Column).To(gomega.Equal(1))
	g.Expect(shifted.End.Column).To(gomega.Equal(5))
}

func TestLineShiftMap_ShiftedRange_Nil(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	m := NewLineShiftMap([]byte("x\n"), nil)
	g.Expect(m.ShiftedRange(nil)).To(gomega.BeNil())
}

func TestLineShiftMap_ShiftedPosition_NoEntries(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	m := NewLineShiftMap([]byte("a\nb\n"), nil)
	pos := finding.Position{Line: 2, Column: 3}
	shifted := m.ShiftedPosition(pos)
	g.Expect(shifted).To(gomega.Equal(pos))
}
