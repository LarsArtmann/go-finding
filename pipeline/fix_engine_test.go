package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func makeRangeFix(
	file string,
	startLine, startCol, endLine, endCol int,
	before, after string,
) finding.Finding {
	return finding.Finding{
		BeforeCode: before, AfterCode: after,
		Range:    finding.NewRangePtr(file, startLine, startCol, endLine, endCol),
		Position: finding.Pos(file, startLine, startCol),
	}
}

func TestFixEngine_Apply_EmptyInput(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()

	lines, applied := engine.Apply(nil, nil)
	g.Expect(lines).To(BeNil())
	g.Expect(applied).To(Equal(0))

	lines, applied = engine.Apply([]string{}, nil)
	g.Expect(lines).To(BeEmpty())
	g.Expect(applied).To(Equal(0))
}

func TestFixEngine_Apply_NoMatchingFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	engine := NewFixEngine()
	lines := []string{
		"package main",
		"",
		"func main() {}",
	}

	fixes := []finding.Finding{
		{
			BeforeCode: "nonexistent", AfterCode: "replacement",
			Position: finding.Position{File: "a.go", Line: 1},
		},
	}

	result, applied := engine.Apply(lines, fixes)
	g.Expect(result).To(Equal(lines))
	g.Expect(applied).To(Equal(0))
}

func TestFixEngine_Apply_FixWithNoCode(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	engine := NewFixEngine()
	lines := []string{"package main"}

	fixes := []finding.Finding{
		{Position: finding.Position{File: "a.go", Line: 1}},
	}

	result, applied := engine.Apply(lines, fixes)
	g.Expect(result).To(Equal(lines))
	g.Expect(applied).To(Equal(0))
}

func TestPartitionFixes(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		r, s := partitionFixes(nil)
		g.Expect(r).To(BeNil())
		g.Expect(s).To(BeNil())
	})

	t.Run("range fix with valid end", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		fixes := []finding.Finding{
			{
				BeforeCode: "old", AfterCode: "new",
				Range:    finding.NewRangePtr("a.go", 1, 1, 3, 1),
				Position: finding.Pos("a.go", 1, 1),
			},
		}
		r, s := partitionFixes(fixes)
		g.Expect(r).To(HaveLen(1))
		g.Expect(s).To(BeEmpty())
	})

	t.Run("string fix without range", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		fixes := []finding.Finding{
			{BeforeCode: "old", AfterCode: "new", Position: finding.Pos("a.go", 1, 1)},
		}
		r, s := partitionFixes(fixes)
		g.Expect(r).To(BeEmpty())
		g.Expect(s).To(HaveLen(1))
	})

	t.Run("fix with no code is skipped", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		fixes := []finding.Finding{
			{Position: finding.Pos("a.go", 1, 1)},
		}
		r, s := partitionFixes(fixes)
		g.Expect(r).To(BeEmpty())
		g.Expect(s).To(BeEmpty())
	})
}

func TestApplyRangeFixes_SingleLine(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"package main", "", "func main() {", "\told()", "}"}
	fixes := []finding.Finding{
		makeRangeFix("a.go", 4, 2, 4, 7, "old()", "new()"),
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	g.Expect(applied).To(Equal(1))
	g.Expect(result[3]).To(Equal("\tnew()"))
}

func TestApplyRangeFixes_MultiLine(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"package main", "", "func old() {", "\treturn", "}", "", "func main() {}"}
	fixes := []finding.Finding{
		{
			BeforeCode: "func old() {\n\treturn\n}",
			AfterCode:  "func new() {\n\treturn 42\n}",
			Range:      finding.NewRangePtr("a.go", 3, 1, 5, 1),
			Position:   finding.Pos("a.go", 3, 1),
		},
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	g.Expect(applied).To(Equal(1))
	g.Expect(result[2]).To(Equal("func new() {"))
	g.Expect(result[3]).To(Equal("\treturn 42"))
	g.Expect(result[4]).To(Equal("}"))
}

func TestApplyRangeFixes_OutOfBounds(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"package main"}
	fixes := []finding.Finding{
		makeRangeFix("a.go", 100, 1, 200, 1, "old", "new"),
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	g.Expect(applied).To(Equal(0))
	g.Expect(result).To(Equal(lines))
}

func TestApplyRangeFixes_DescendingOrder(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{
		"package main",
		"line2: old",
		"line3: old",
		"line4: old",
	}
	fixes := []finding.Finding{
		makeRangeFix("a.go", 2, 8, 2, 11, "old", "fix1"),
		makeRangeFix("a.go", 4, 8, 4, 11, "old", "fix2"),
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	g.Expect(applied).To(Equal(2))
	g.Expect(result[1]).To(Equal("line2: fix1"))
	g.Expect(result[3]).To(Equal("line4: fix2"))
}

func TestApplyStringFixes_ReplaceFirst(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"old code here"}
	fixes := []finding.Finding{
		{BeforeCode: "old", AfterCode: "new", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	g.Expect(applied).To(Equal(1))
	g.Expect(result[0]).To(Equal("new code here"))
}

func TestApplyStringFixes_Insertion(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"package main", "", "func main() {}"}
	fixes := []finding.Finding{
		{AfterCode: "\tinserted", Position: finding.Pos("a.go", 3, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	g.Expect(applied).To(Equal(1))
	g.Expect(result).To(HaveLen(4))
	g.Expect(result[2]).To(Equal("\tinserted"))
}

func TestApplyStringFixes_NotFound(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	lines := []string{"package main"}
	fixes := []finding.Finding{
		{BeforeCode: "nonexistent", AfterCode: "replacement", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	g.Expect(applied).To(Equal(0))
	g.Expect(result).To(Equal(lines))
}

func TestReplaceNearestToLine(t *testing.T) {
	t.Parallel()

	content := "line1: X\nline2: X\nline3: X"

	t.Run("first occurrence when targetLine is 0", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		result := replaceNearestToLine(content, "X", "Y", 0)
		g.Expect(result).To(Equal("line1: Y\nline2: X\nline3: X"))
	})

	t.Run("nearest to target line", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		result := replaceNearestToLine(content, "X", "Y", 3)
		g.Expect(result).To(Equal("line1: X\nline2: X\nline3: Y"))
	})

	t.Run("not found returns unchanged", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		result := replaceNearestToLine(content, "Z", "Y", 1)
		g.Expect(result).To(Equal(content))
	})
}

func TestLineDistance(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	content := "a\nb\nc"

	g.Expect(lineDistance(content, 0, 1)).To(Equal(0))
	g.Expect(lineDistance(content, 0, 2)).To(Equal(1))
	g.Expect(lineDistance(content, 2, 1)).To(Equal(1))
	g.Expect(lineDistance(content, 0, 3)).To(Equal(2))
}
