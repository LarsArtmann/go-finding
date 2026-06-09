package pipeline

import (
	"testing"

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
