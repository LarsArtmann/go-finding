package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFixEngine_Apply_EmptyInput(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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
	return finding.Finding{
		BeforeCode: before, AfterCode: after,
		Range:    finding.NewRangePtr(file, startLine, startCol, endLine, endCol),
		Position: finding.Pos(file, startLine, startCol),
	}
}

func twoLineRangeFixes() ([]byte, []finding.Finding) {
	return []byte("line1: old\nline2: old\nline3: old"),
		[]finding.Finding{
			makeRangeFix("a.go", 1, 8, 1, 11, "old", "fix1"),
			makeRangeFix("a.go", 3, 8, 3, 11, "old", "fix2"),
		}
}

func makeOffsetFix(id, before, after string, startOff, endOff int) finding.Finding {
	return finding.Finding{
		ID:         id,
		BeforeCode: before,
		AfterCode:  after,
		Range: &finding.Range{
			Start: finding.Position{File: "a.go", Offset: startOff},
			End:   finding.Position{File: "a.go", Offset: endOff},
		},
		Position: finding.Pos("a.go", 4, 2),
	}
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
			t.Parallel()
			g := NewWithT(t)
			engine := NewFixEngine()
			result, _, count := engine.Apply([]byte(tt.content), []finding.Finding{tt.fix})
			g.Expect(count).To(Equal(tt.wantCount))
			g.Expect(string(result)).To(Equal(tt.wantResult))
		})
	}
}

func TestFixEngine_Apply_LineRange_MultiLine(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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
			t.Parallel()
			g := NewWithT(t)
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
	t.Parallel()
	g := NewWithT(t)

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
			t.Parallel()
			g := NewWithT(t)

			fixes := []finding.Finding{
				{BeforeCode: "X", AfterCode: "Y", Position: finding.Pos("a.go", tt.line, tt.col)},
			}
			result, _, count := engine.Apply(content, fixes)
			g.Expect(count).To(Equal(1))
			g.Expect(string(result)).To(Equal(tt.want))
		})
	}
}

func TestFixEngine_Apply_ByteOffset(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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

func TestLineColToOffset(t *testing.T) {
	t.Parallel()

	content := []byte("a\nb\nc")

	cases := []struct {
		name       string
		line, col  int
		wantOffset int
		wantErr    bool
	}{
		{"line 1, col 1", 1, 1, 0, false},
		{"line 2, col 1", 2, 1, 2, false},
		{"line 3, col 1", 3, 1, 4, false},
		{"line beyond file", 10, 1, 0, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			idx := buildLineOffsetIndex(content)
			offset, err := indexLineColToOffset(idx, len(content), tt.line, tt.col)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(offset).To(Equal(tt.wantOffset))
			}
		})
	}
}

func TestFindAllOccurrences(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	content := []byte("abcabc")
	needle := []byte("abc")

	positions := findAllOccurrences(content, needle)
	g.Expect(positions).To(Equal([]int{0, 3}))
}

func TestOffsetLineDistance(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	content := []byte("a\nb\nc")

	g.Expect(offsetLineDistance(content, 0, 1)).To(Equal(0))
	g.Expect(offsetLineDistance(content, 0, 2)).To(Equal(1))
	g.Expect(offsetLineDistance(content, 2, 1)).To(Equal(1))
	g.Expect(offsetLineDistance(content, 0, 3)).To(Equal(2))
}

func TestFixEdit_Overlaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		a, b    FixEdit
		overlap bool
	}{
		{"non-overlapping", FixEdit{Offset: 0, Length: 5}, FixEdit{Offset: 5, Length: 5}, false},
		{"overlapping", FixEdit{Offset: 0, Length: 6}, FixEdit{Offset: 5, Length: 5}, true},
		{
			"zero-length at same offset",
			FixEdit{Offset: 10, Length: 0},
			FixEdit{Offset: 10, Length: 0},
			true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(tt.a.Overlaps(tt.b)).To(Equal(tt.overlap))
		})
	}
}

func TestFixEdit_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid edit", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: 0, Length: 5}.Validate()).NotTo(HaveOccurred())
	})

	t.Run("negative offset", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: -1}.Validate()).To(HaveOccurred())
	})

	t.Run("negative length", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: 0, Length: -1}.Validate()).To(HaveOccurred())
	})
}

func TestFixEdit_Helpers(t *testing.T) {
	t.Parallel()

	t.Run("EndOffset", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: 10, Length: 5}.EndOffset()).To(Equal(15))
	})

	t.Run("IsInsert", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: 10, Length: 0, Replacement: []byte("x")}.IsInsert()).To(BeTrue())
		g.Expect(FixEdit{Offset: 10, Length: 1}.IsInsert()).To(BeFalse())
	})

	t.Run("IsDelete", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		g.Expect(FixEdit{Offset: 10, Length: 5}.IsDelete()).To(BeTrue())
		g.Expect(FixEdit{Offset: 10, Length: 5, Replacement: []byte("x")}.IsDelete()).To(BeFalse())
	})
}

func TestFixEngine_Providers(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	providers := engine.Providers()
	g.Expect(providers).To(HaveLen(3))
	g.Expect(providers[0].Name()).To(Equal("byte-offset"))
	g.Expect(providers[1].Name()).To(Equal("line-column"))
	g.Expect(providers[2].Name()).To(Equal("substring"))
}

func TestFixEngine_ApplyWithConflicts_NoConflicts(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content, fixes := twoLineRangeFixes()

	applied, conflicts, result, _ := engine.ApplyWithConflicts(content, fixes)
	g.Expect(applied).To(HaveLen(2))
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(string(result)).To(Equal("line1: fix1\nline2: old\nline3: fix2"))
}

func TestFixEngine_ApplyWithConflicts_OverlappingEdits(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {\n\told()\n}")

	fixes := overlappingOffsetFixes()

	applied, conflicts, result, _ := engine.ApplyWithConflicts(content, fixes)
	g.Expect(applied).To(HaveLen(1))
	g.Expect(applied[0].ID).To(Equal("fix1"))
	g.Expect(conflicts).To(HaveLen(1))
	g.Expect(conflicts[0].Reason).To(Equal("overlapping edit"))
	g.Expect(conflicts[0].Finding.ID).To(Equal("fix2"))
	g.Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
}

func TestFixEdit_JSON(t *testing.T) {
	t.Parallel()

	t.Run("marshal and unmarshal round-trip", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		edit := FixEdit{Offset: 10, Length: 5, Replacement: []byte("hello")}

		data, err := json.Marshal(edit)
		g.Expect(err).NotTo(HaveOccurred())

		var got FixEdit
		g.Expect(json.Unmarshal(data, &got)).NotTo(HaveOccurred())
		g.Expect(got.Offset).To(Equal(10))
		g.Expect(got.Length).To(Equal(5))
		g.Expect(got.Replacement).To(Equal([]byte("hello")))
	})

	t.Run("pure deletion omits replacement", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		edit := FixEdit{Offset: 10, Length: 5}
		data, err := json.Marshal(edit)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(data)).NotTo(ContainSubstring("replacement"))
	})

	t.Run("source not included in JSON", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		edit := FixEdit{
			Offset:      10,
			Length:      5,
			Replacement: []byte("x"),
			Source:      finding.Finding{ID: "test-123"},
		}
		data, err := json.Marshal(edit)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(data)).NotTo(ContainSubstring("test-123"))
	})
}

func TestFixEdit_SARIFProperties(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		edit := FixEdit{Offset: 42, Length: 10, Replacement: []byte("new code")}
		props := edit.ToSARIFProperties()

		got := FixEditFromSARIFProperties(props)
		g.Expect(got).NotTo(BeNil())
		g.Expect(got.Offset).To(Equal(42))
		g.Expect(got.Length).To(Equal(10))
		g.Expect(got.Replacement).To(Equal([]byte("new code")))
	})

	t.Run("missing offset returns nil", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		props := map[string]string{"go-finding/edit/length": "5"}
		got := FixEditFromSARIFProperties(props)
		g.Expect(got).To(BeNil())
	})

	t.Run("pure deletion has no replacement", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		edit := FixEdit{Offset: 10, Length: 5}
		props := edit.ToSARIFProperties()
		g.Expect(props).NotTo(HaveKey("go-finding/edit/replacement"))

		got := FixEditFromSARIFProperties(props)
		g.Expect(got).NotTo(BeNil())
		g.Expect(got.Replacement).To(BeNil())
	})
}

func TestFilterConflictingEdits(t *testing.T) {
	t.Parallel()

	t.Run("no conflicts", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		engine := NewFixEngine()
		content, fixes := twoLineRangeFixes()

		result, errs := FilterConflictingEdits(content, fixes, engine)
		g.Expect(errs).To(BeEmpty())
		g.Expect(result).To(HaveLen(2))
	})

	t.Run("overlapping edits filtered", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		engine := NewFixEngine()
		content := []byte("package main\n\nfunc main() {\n\told()\n}")
		fixes := overlappingOffsetFixes()

		result, errs := FilterConflictingEdits(content, fixes, engine)
		g.Expect(errs).To(BeEmpty())
		g.Expect(result).To(HaveLen(1))
		g.Expect(result[0].ID).To(Equal("fix1"))
	})
}

func TestFixEngine_MultipleLineEdits_SameFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()

	content := []byte("line1: foo\nline2: bar\nline3: baz\nline4: qux\n")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 1, 7, 1, 10, "foo", "FIXED1"),
		makeRangeFix("a.go", 2, 7, 2, 10, "bar", "FIXED2"),
		makeRangeFix("a.go", 3, 7, 3, 10, "baz", "FIXED3"),
		makeRangeFix("a.go", 4, 7, 4, 10, "qux", "FIXED4"),
	}

	applied, conflicts, result, providerErrors := engine.ApplyWithConflicts(content, fixes)
	g.Expect(providerErrors).To(BeEmpty())
	g.Expect(applied).To(HaveLen(4))
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(string(result)).To(Equal(
		"line1: FIXED1\nline2: FIXED2\nline3: FIXED3\nline4: FIXED4\n",
	))
}

func TestFixEngine_LineEdits_DifferentLengths(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()

	content := []byte("a: short\nb: medium\nc: longer\n")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 1, 3, 1, 8, "short", "MUCH-LONGER-REPLACEMENT"),
		makeRangeFix("a.go", 3, 3, 3, 9, "longer", "X"),
	}

	applied, conflicts, result, providerErrors := engine.ApplyWithConflicts(content, fixes)
	g.Expect(providerErrors).To(BeEmpty())
	g.Expect(applied).To(HaveLen(2))
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(string(result)).To(Equal("a: MUCH-LONGER-REPLACEMENT\nb: medium\nc: X\n"))
}
