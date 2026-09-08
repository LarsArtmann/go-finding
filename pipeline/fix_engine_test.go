package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFixEngine_Apply_EmptyInput(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()

	result, applied, count := engine.Apply(nil, nil)
	g.Expect(result).To(BeNil())
	g.Expect(applied).To(BeNil())
	g.Expect(count).To(Equal(0))

	result, applied, count = engine.Apply([]byte{}, nil)
	g.Expect(result).To(BeEmpty())
	g.Expect(applied).To(BeNil())
	g.Expect(count).To(Equal(0))
}

func TestFixEngine_Apply_NoMatchingFixes(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {}")

	fixes := []finding.Finding{
		{
			BeforeCode: "nonexistent", AfterCode: "replacement",
			Position: finding.Position{File: "a.go", Line: 1},
		},
	}

	result, applied, count := engine.Apply(content, fixes)
	g.Expect(string(result)).To(Equal(string(content)))
	g.Expect(applied).To(BeNil())
	g.Expect(count).To(Equal(0))
}

func TestFixEngine_Apply_FixWithNoCode(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main")

	fixes := []finding.Finding{
		{Position: finding.Position{File: "a.go", Line: 1}},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(string(result)).To(Equal("package main"))
	g.Expect(count).To(Equal(0))
}

func makeRangeFix(
	file string,
	startLine, startCol, endLine, endCol int,
	before, after string,
) finding.Finding {
	return rangeFix(file, startLine, startCol, endLine, endCol, before, after)
}

func twoLineRangeFixes() ([]byte, []finding.Finding) {
	return []byte("line1: old\nline2: old\nline3: old"),
		[]finding.Finding{
			makeRangeFix("a.go", 1, 8, 1, 11, "old", "fix1"),
			makeRangeFix("a.go", 3, 8, 3, 11, "old", "fix2"),
		}
}

func makeOffsetFix(id, before, after string, startOff, endOff int) finding.Finding {
	return offsetFixWithID(id, before, after, startOff, endOff)
}

func overlappingOffsetFixes() []finding.Finding {
	return []finding.Finding{
		makeOffsetFix("fix1", "old", "new", 28, 33),
		makeOffsetFix("fix2", "old()", "replaced()", 28, 33),
	}
}

func TestFixEngine_Apply_LineRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		content    string
		fix        finding.Finding
		wantCount  int
		wantResult string
	}{
		{
			"single line",
			"package main\n\nfunc main() {\n\told()\n}",
			makeRangeFix("a.go", 4, 2, 4, 7, "old()", "new()"),
			1,
			"package main\n\nfunc main() {\n\tnew()\n}",
		},
		{
			"out of bounds",
			"package main",
			makeRangeFix("a.go", 100, 1, 200, 1, "old", "new"),
			0,
			"package main",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)
			engine := NewFixEngine()
			result, _, count := engine.Apply([]byte(tt.content), []finding.Finding{tt.fix})
			g.Expect(count).To(Equal(tt.wantCount))
			g.Expect(string(result)).To(Equal(tt.wantResult))
		})
	}
}

func TestFixEngine_Apply_LineRange_MultiLine(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc old() {\n\treturn\n}\n\nfunc main() {}")

	fixes := []finding.Finding{
		{
			BeforeCode: "func old() {\n\treturn\n}",
			AfterCode:  "func new() {\n\treturn 42\n}",
			Range:      finding.NewRangePtr("a.go", 3, 1, 5, 1),
			Position:   finding.Pos("a.go", 3, 1),
		},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(
		Equal("package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}"),
	)
}

func TestFixEngine_Apply_LineRange_DescendingOrder(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main\nline2: old\nline3: old\nline4: old")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 2, 8, 2, 11, "old", "fix1"),
		makeRangeFix("a.go", 4, 8, 4, 11, "old", "fix2"),
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(2))
	g.Expect(string(result)).To(Equal("package main\nline2: fix1\nline3: old\nline4: fix2"))
}

func TestFixEngine_Apply_Substring(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		content    string
		before     string
		after      string
		wantCount  int
		wantResult string
	}{
		{"replace", "old code here", "old", "new", 1, "new code here"},
		{"not found", "package main", "nonexistent", "replacement", 0, "package main"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)
			engine := NewFixEngine()
			fixes := []finding.Finding{
				{BeforeCode: tt.before, AfterCode: tt.after, Position: finding.Pos("a.go", 1, 1)},
			}
			result, _, count := engine.Apply([]byte(tt.content), fixes)
			g.Expect(count).To(Equal(tt.wantCount))
			g.Expect(string(result)).To(Equal(tt.wantResult))
		})
	}
}

func TestFixEngine_Apply_SubstringInsertion(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {}")

	fixes := []finding.Finding{
		{AfterCode: "\tinserted", Position: finding.Pos("a.go", 3, 1)},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(Equal("package main\n\n\tinserted\nfunc main() {}"))
}

func TestFixEngine_Apply_NearestLineMatch(t *testing.T) {
	t.Parallel()

	engine := NewFixEngine()
	content := []byte("line1: X\nline2: X\nline3: X")

	cases := []struct {
		name      string
		line, col int
		want      string
	}{
		{"first occurrence when targetLine is 0", 0, 0, "line1: Y\nline2: X\nline3: X"},
		{"nearest to target line", 3, 0, "line1: X\nline2: X\nline3: Y"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			fixes := []finding.Finding{
				{BeforeCode: "X", AfterCode: "Y", Position: finding.Pos("a.go", tt.line, tt.col)},
			}
			result, _, count := engine.Apply(content, fixes)
			g.Expect(count).To(Equal(1))
			g.Expect(string(result)).To(Equal(tt.want))
		})
	}
}

func TestFixEngine_Apply_NearestColumnMatch(t *testing.T) {
	t.Parallel()

	// Two occurrences of "X" on the SAME line; column must disambiguate.
	content := []byte("X and X here")

	engine := NewFixEngine()

	cases := []struct {
		name string
		col  int
		want string
	}{
		{"first X via column 1", 1, "Y and X here"},
		{"second X via column 7", 7, "X and Y here"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			fixes := []finding.Finding{
				{BeforeCode: "X", AfterCode: "Y", Position: finding.Pos("a.go", 1, tt.col)},
			}
			result, _, count := engine.Apply(content, fixes)
			g.Expect(count).To(Equal(1))
			g.Expect(string(result)).To(Equal(tt.want))
		})
	}
}

func TestFixEngine_Apply_ByteOffset(t *testing.T) {
	g := NewParallelGomega(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {\n\told()\n}")

	// "old()" starts at byte offset 29 (offset 28 is '\t').
	// Line=0 prevents LineProvider from handling this — forces OffsetProvider.
	fixes := []finding.Finding{
		{
			BeforeCode: "old()",
			AfterCode:  "new()",
			Range: &finding.Range{
				Start: finding.Position{File: "a.go", Offset: 29},
				End:   finding.Position{File: "a.go", Offset: 34},
			},
			Position: finding.Pos("a.go", 0, 0),
		},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
}

func TestFixEngine_Apply_WithCustomProvider(t *testing.T) {
	g := NewParallelGomega(t)

	custom := &upperProvider{}
	engine := NewFixEngineWithProviders(custom)

	content := []byte("hello world")
	fixes := []finding.Finding{
		{
			BeforeCode:  "hello",
			AfterCode:   "HELLO",
			Position:    finding.Pos("a.go", 1, 1),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	result, applied, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(applied).To(HaveLen(1))
	g.Expect(string(result)).To(Equal("HELLO world"))
}

// upperProvider is a test provider that only handles direct fixes.
type upperProvider struct{}

func (upperProvider) Name() string { return "test-upper" }

func (upperProvider) CanHandle(f finding.Finding) bool {
	return f.FixStrategy == finding.FixStrategyDirect && f.BeforeCode != ""
}

func (upperProvider) Edits(_ []byte, f finding.Finding) ([]FixEdit, error) {
	return []FixEdit{
		{
			Offset:      0,
			Length:      len(f.BeforeCode),
			Replacement: []byte(f.AfterCode),
			Source:      f,
		},
	}, nil
}

// TestApplyWithOutcomes_FailedOutcomeWrapsErrPositionUnresolvable verifies
// that a provider failure surfaces as a failed outcome whose error chain is
// matchable with errors.Is — so consumers can react to specific failure
// causes instead of string matching.
func TestApplyWithOutcomes_FailedOutcomeWrapsErrPositionUnresolvable(t *testing.T) {
	g := NewParallelGomega(t)

	content := []byte("package main\n\nold()\n")
	fix := finding.Finding{
		ID:         "test:rule:unresolvable",
		Rule:       "rule",
		ToolName:   "test",
		Message:    "fix",
		BeforeCode: "never-present()",
		AfterCode:  "new()",
		Position:   finding.Pos("test.go", 1000, 1),
	}

	result := NewFixEngine().ApplyWithOutcomes(content, []finding.Finding{fix})

	g.Expect(result.Outcomes).To(HaveLen(1))
	g.Expect(result.Outcomes[0].Status).To(Equal(FixOutcomeFailed))
	g.Expect(result.Outcomes[0].Err).NotTo(BeNil())
	g.Expect(result.Outcomes[0].Err).To(MatchError(ErrPositionUnresolvable))
	g.Expect(result.HasErrors()).To(BeTrue())
	g.Expect(result.OutcomeFor(fix.ID)).NotTo(BeNil())
}
