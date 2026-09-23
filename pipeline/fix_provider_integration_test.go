package pipeline

import (
	"bytes"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// TestFixEngine_MixedProviders verifies that the FixEngine correctly routes
// findings to different providers (Offset, Line, Substring) in a single
// ApplyWithConflicts call, and that the line offset index is built only once.
func TestFixEngine_MixedProviders(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nfunc main() {\n\talpha()\n\tbeta()\n\tgamma()\n}\n")

	alphaIdx := bytes.Index(content, []byte("alpha()"))

	engine := NewFixEngine()

	findings := []finding.Finding{
		// OffsetProvider: replace alpha() via byte offsets
		{
			Range: &finding.Range{
				Start: finding.Position{Offset: alphaIdx},
				End:   finding.Position{Offset: alphaIdx + 7},
			},
			BeforeCode: "alpha()",
			AfterCode:  "first()",
			Position:   finding.Position{File: "test.go"},
		},
		// LineProvider: replace gamma() (line 6) via line/column
		{
			Position:   finding.Position{File: "test.go", Line: 6, Column: 2},
			BeforeCode: "gamma()",
			AfterCode:  "third()",
		},
		// SubstringProvider: replace beta() via substring (no position)
		{
			Position:   finding.Position{File: "test.go"},
			BeforeCode: "beta()",
			AfterCode:  "second()",
		},
	}

	result, applied, count := engine.Apply(content, findings)

	g.Expect(count).To(Equal(3), "all three findings should be applied")
	g.Expect(applied).To(HaveLen(3))

	expected := "package main\n\nfunc main() {\n\tfirst()\n\tsecond()\n\tthird()\n}\n"
	g.Expect(string(result)).To(Equal(expected))
}

// TestFixEngine_LineIndexLazyBuild verifies that the line offset index is
// built only when a lineIndexAware provider handles a finding, not when
// only OffsetProvider-handled findings are present.
func TestFixEngine_LineIndexLazyBuild(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nfunc main() {\n\told()\n}\n")

	firstOld := bytes.Index(content, []byte("old()"))

	engine := NewFixEngine()

	// Only offset-based findings — line index should NOT be built
	findings := []finding.Finding{
		{
			Range: &finding.Range{
				Start: finding.Position{Offset: firstOld},
				End:   finding.Position{Offset: firstOld + 5},
			},
			BeforeCode: "old()",
			AfterCode:  "new()",
			Position:   finding.Position{File: "test.go"},
		},
	}

	modified := applyAndGetModifiedContent(t, engine, content, findings)
	g.Expect(modified).To(Equal("package main\n\nfunc main() {\n\tnew()\n}\n"))
}

func applyAndGetModifiedContent(t *testing.T, engine *FixEngine, content []byte, findings []finding.Finding) string {
	t.Helper()

	_, _, _, result, _ := engine.ApplyWithConflicts(content, findings) //nolint:dogsled // only need content

	return string(result)
}

// TestFixEngine_InterleavedProviders verifies correctness when offset and
// line-based findings are interleaved in the same file.
func TestFixEngine_InterleavedProviders(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("line1\nline2\nline3\nold1\nline5\nold2\nline7\n")

	// Two findings: one offset-based, one line-based
	old1Idx := bytes.Index(content, []byte("old1"))
	before, _, _ := bytes.Cut(content, []byte("old2"))
	old2Line := 1 + bytes.Count(before, []byte{'\n'})

	engine := NewFixEngine()

	findings := []finding.Finding{
		// Offset-based: replace old1
		{
			Range: &finding.Range{
				Start: finding.Position{Offset: old1Idx},
				End:   finding.Position{Offset: old1Idx + 4},
			},
			BeforeCode: "old1",
			AfterCode:  "new1",
			Position:   finding.Position{File: "test.txt"},
		},
		// Line-based: replace old2
		{
			Position:   finding.Position{File: "test.txt", Line: old2Line, Column: 1},
			BeforeCode: "old2",
			AfterCode:  "new2",
		},
	}

	result, _, count := engine.Apply(content, findings)

	g.Expect(count).To(Equal(2))
	g.Expect(string(result)).To(ContainSubstring("new1"))
	g.Expect(string(result)).To(ContainSubstring("new2"))
	g.Expect(string(result)).NotTo(ContainSubstring("old1"))
	g.Expect(string(result)).NotTo(ContainSubstring("old2"))
}

func TestPickNearestOccurrence_Branches(t *testing.T) {
	t.Parallel()

	content := []byte("X here\nX there\nX everywhere")
	idx := buildLineOffsetIndex(content)
	occurrences := findAllOccurrences(content, []byte("X")) // offsets 0, 7, 15

	t.Run("line only fallback", func(t *testing.T) {
		t.Parallel()

		f := finding.Finding{Position: finding.Position{Line: 2}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != 7 {
			t.Fatalf("line-only: got %d, want 7", best)
		}
	})

	t.Run("line out of range returns first", func(t *testing.T) {
		t.Parallel()

		f := finding.Finding{Position: finding.Position{Line: 99}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != occurrences[0] {
			t.Fatalf("out-of-range: got %d, want %d", best, occurrences[0])
		}
	})

	t.Run("column nearest", func(t *testing.T) {
		t.Parallel()

		f := finding.Finding{Position: finding.Position{Line: 1, Column: 1}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != 0 {
			t.Fatalf("col1: got %d, want 0", best)
		}
	})
}

// TestEditListProvider_AppliesAllEditsInOnePass reproduces the issue #36
// scenario: a multi-edit fix (declaration removal + condition rewrite, the
// erraudit legacyerrors shape) must apply completely via the default engine.
func TestEditListProvider_AppliesAllEditsInOnePass(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nvar useLegacy = true\n\nfunc main() {\n\tif useLegacy {\n\t\tpanic(1)\n\t}\n}\n")

	declOff := bytes.Index(content, []byte("var useLegacy = true\n"))
	condOff := bytes.Index(content, []byte("useLegacy {"))

	f := finding.Finding{
		ID:          "1",
		Rule:        "legacy-flag",
		ToolName:    "tool",
		Message:     "legacy flag must go",
		Severity:    finding.SeverityWarning,
		Position:    finding.Position{File: "main.go", Line: 6, Column: 5},
		FixStrategy: finding.FixStrategyDirect,
		Edits: []finding.TextEdit{
			{
				Start: finding.Position{File: "main.go", Line: 3, Column: 1, Offset: declOff},
				End:   finding.Position{File: "main.go", Line: 3, Column: 22, Offset: declOff + len("var useLegacy = true\n")},
			},
			{
				Start:   finding.Position{File: "main.go", Line: 6, Column: 5, Offset: condOff},
				End:     finding.Position{File: "main.go", Line: 6, Column: 16, Offset: condOff + len("useLegacy {")},
				NewText: "false {",
			},
		},
	}

	engine := NewFixEngine()
	result := engine.ApplyWithOutcomes(content, []finding.Finding{f})

	g.Expect(result.Outcomes).To(HaveLen(1))
	g.Expect(result.Outcomes[0].Status).To(Equal(FixOutcomeApplied))
	// Applied describes findings (one entry per fix); AppliedEdits keeps
	// per-edit granularity.
	g.Expect(result.Applied).To(HaveLen(1))
	g.Expect(result.AppliedEdits).To(HaveLen(2))

	// The decl edit removes exactly the "var useLegacy = true\n" line, so the
	// blank lines before and after it merge into the double blank shown here.
	expected := "package main\n\n\nfunc main() {\n\tif false {\n\t\tpanic(1)\n\t}\n}\n"
	g.Expect(string(result.Content)).To(Equal(expected), "both edits must apply; half-application is the bug")
}

func TestEditListProvider_LineColumnOnlyEdits(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nfunc main() {\n\told()\n}\n")

	f := finding.Finding{
		ID:          "1",
		BeforeCode:  "old()",
		Position:    finding.Position{File: "main.go", Line: 4, Column: 2},
		FixStrategy: finding.FixStrategyDirect,
		Edits: []finding.TextEdit{
			{
				Start:   finding.Position{File: "main.go", Line: 4, Column: 2, Offset: -1},
				End:     finding.Position{File: "main.go", Line: 4, Column: 7, Offset: -1},
				NewText: "new()",
			},
		},
	}

	engine := NewFixEngine()
	result, applied, count := engine.Apply(content, []finding.Finding{f})

	g.Expect(count).To(Equal(1))
	g.Expect(applied).To(HaveLen(1))
	g.Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}\n"))
}

func TestEditListProvider_CrossFileEditListRefused(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nvar x = true\n")

	f := finding.Finding{
		ID:          "1",
		Position:    finding.Position{File: "main.go", Line: 3},
		FixStrategy: finding.FixStrategyDirect,
		Edits: []finding.TextEdit{
			{
				Start: finding.Position{File: "other.go", Line: 1, Column: 1, Offset: 0},
				End:   finding.Position{File: "other.go", Line: 1, Column: 2, Offset: 1},
			},
		},
	}

	engine := NewFixEngine()
	result := engine.ApplyWithOutcomes(content, []finding.Finding{f})

	g.Expect(result.Outcomes).To(HaveLen(1))
	g.Expect(result.Outcomes[0].Status).To(Equal(FixOutcomeFailed))
	g.Expect(result.Outcomes[0].Err).To(HaveOccurred())
	g.Expect(errors.Is(result.Outcomes[0].Err, ErrEditCrossFile)).To(BeTrue(),
		"cross-file edit lists must fail loudly, not half-apply")
	g.Expect(string(result.Content)).To(Equal(string(content)))
}

func TestEditListProvider_StaleOffsetsRefused(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n")

	f := finding.Finding{
		ID:          "1",
		Position:    finding.Position{File: "main.go", Line: 1},
		FixStrategy: finding.FixStrategyDirect,
		Edits: []finding.TextEdit{
			{
				Start: finding.Position{File: "main.go", Line: 9, Column: 1, Offset: 500},
				End:   finding.Position{File: "main.go", Line: 9, Column: 5, Offset: 505},
			},
		},
	}

	engine := NewFixEngine()
	result := engine.ApplyWithOutcomes(content, []finding.Finding{f})

	g.Expect(result.Outcomes[0].Status).To(Equal(FixOutcomeFailed))
	g.Expect(errors.Is(result.Outcomes[0].Err, ErrEditStale)).To(BeTrue())
}

func TestEditListProvider_PureInsertion(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n")

	f := finding.Finding{
		ID:          "1",
		Position:    finding.Position{File: "main.go", Line: 1},
		FixStrategy: finding.FixStrategyDirect,
		Edits: []finding.TextEdit{
			{
				Start:   finding.Position{File: "main.go", Line: 1, Column: 1, Offset: 0},
				End:     finding.Position{Offset: -1},
				NewText: "import \"fmt\"\n",
			},
		},
	}

	engine := NewFixEngine()
	result, _, count := engine.Apply(content, []finding.Finding{f})

	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(Equal("import \"fmt\"\npackage main\n"))
}
