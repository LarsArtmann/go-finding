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

func TestFixEngine_Apply_LineRange_SingleLine(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {\n\told()\n}")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 4, 2, 4, 7, "old()", "new()"),
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
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
		Equal("package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}"))
}

func TestFixEngine_Apply_LineRange_OutOfBounds(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("package main")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 100, 1, 200, 1, "old", "new"),
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(0))
	g.Expect(string(result)).To(Equal("package main"))
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

func TestFixEngine_Apply_SubstringReplace(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("old code here")

	fixes := []finding.Finding{
		{BeforeCode: "old", AfterCode: "new", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(1))
	g.Expect(string(result)).To(Equal("new code here"))
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

func TestFixEngine_Apply_SubstringNotFound(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("package main")

	fixes := []finding.Finding{
		{BeforeCode: "nonexistent", AfterCode: "replacement", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, count := engine.Apply(content, fixes)
	g.Expect(count).To(Equal(0))
	g.Expect(string(result)).To(Equal("package main"))
}

func TestFixEngine_Apply_NearestLineMatch(t *testing.T) {
	t.Parallel()
	engine := NewFixEngine()
	content := []byte("line1: X\nline2: X\nline3: X")

	t.Run("first occurrence when targetLine is 0", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		fixes := []finding.Finding{
			{BeforeCode: "X", AfterCode: "Y", Position: finding.Pos("a.go", 0, 0)},
		}
		result, _, count := engine.Apply(content, fixes)
		g.Expect(count).To(Equal(1))
		g.Expect(string(result)).To(Equal("line1: Y\nline2: X\nline3: X"))
	})

	t.Run("nearest to target line", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		fixes := []finding.Finding{
			{BeforeCode: "X", AfterCode: "Y", Position: finding.Pos("a.go", 3, 0)},
		}
		result, _, count := engine.Apply(content, fixes)
		g.Expect(count).To(Equal(1))
		g.Expect(string(result)).To(Equal("line1: X\nline2: X\nline3: Y"))
	})
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

	t.Run("line 1, col 1", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		offset, err := lineColToOffset(content, 1, 1)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(offset).To(Equal(0))
	})

	t.Run("line 2, col 1", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		offset, err := lineColToOffset(content, 2, 1)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(offset).To(Equal(2))
	})

	t.Run("line 3, col 1", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		offset, err := lineColToOffset(content, 3, 1)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(offset).To(Equal(4))
	})

	t.Run("line beyond file", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		_, err := lineColToOffset(content, 10, 1)
		g.Expect(err).To(HaveOccurred())
	})
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

	t.Run("non-overlapping", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		a := FixEdit{Offset: 0, Length: 5}
		b := FixEdit{Offset: 5, Length: 5}
		g.Expect(a.Overlaps(b)).To(BeFalse())
	})

	t.Run("overlapping", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		a := FixEdit{Offset: 0, Length: 6}
		b := FixEdit{Offset: 5, Length: 5}
		g.Expect(a.Overlaps(b)).To(BeTrue())
	})

	t.Run("zero-length at same offset", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		a := FixEdit{Offset: 10, Length: 0}
		b := FixEdit{Offset: 10, Length: 0}
		g.Expect(a.Overlaps(b)).To(BeTrue())
	})
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
	content := []byte("line1: old\nline2: old\nline3: old")

	fixes := []finding.Finding{
		makeRangeFix("a.go", 1, 8, 1, 11, "old", "fix1"),
		makeRangeFix("a.go", 3, 8, 3, 11, "old", "fix2"),
	}

	applied, conflicts, result := engine.ApplyWithConflicts(content, fixes)
	g.Expect(applied).To(HaveLen(2))
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(string(result)).To(Equal("line1: fix1\nline2: old\nline3: fix2"))
}

func TestFixEngine_ApplyWithConflicts_OverlappingEdits(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	engine := NewFixEngine()
	content := []byte("package main\n\nfunc main() {\n\told()\n}")

	fixes := []finding.Finding{
		{
			ID:         "fix1",
			BeforeCode: "old",
			AfterCode:  "new",
			Range: &finding.Range{
				Start: finding.Position{File: "a.go", Offset: 28},
				End:   finding.Position{File: "a.go", Offset: 33},
			},
			Position: finding.Pos("a.go", 4, 2),
		},
		{
			ID:         "fix2",
			BeforeCode: "old()",
			AfterCode:  "replaced()",
			Range: &finding.Range{
				Start: finding.Position{File: "a.go", Offset: 28},
				End:   finding.Position{File: "a.go", Offset: 33},
			},
			Position: finding.Pos("a.go", 4, 2),
		},
	}

	applied, conflicts, result := engine.ApplyWithConflicts(content, fixes)
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
		content := []byte("line1: old\nline2: old\nline3: old")
		fixes := []finding.Finding{
			makeRangeFix("a.go", 1, 8, 1, 11, "old", "fix1"),
			makeRangeFix("a.go", 3, 8, 3, 11, "old", "fix2"),
		}

		result := FilterConflictingEdits(content, fixes, engine)
		g.Expect(result).To(HaveLen(2))
	})

	t.Run("overlapping edits filtered", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		engine := NewFixEngine()
		content := []byte("package main\n\nfunc main() {\n\told()\n}")
		fixes := []finding.Finding{
			{
				ID:         "fix1",
				BeforeCode: "old",
				AfterCode:  "new",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 28},
					End:   finding.Position{File: "a.go", Offset: 33},
				},
				Position: finding.Pos("a.go", 4, 2),
			},
			{
				ID:         "fix2",
				BeforeCode: "old()",
				AfterCode:  "replaced()",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 28},
					End:   finding.Position{File: "a.go", Offset: 33},
				},
				Position: finding.Pos("a.go", 4, 2),
			},
		}

		result := FilterConflictingEdits(content, fixes, engine)
		g.Expect(result).To(HaveLen(1))
		g.Expect(result[0].ID).To(Equal("fix1"))
	})
}
