package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

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
	lineIndex := buildLineOffsetIndex(content)

	g.Expect(offsetLineDistance(lineIndex, 0, 1)).To(Equal(0))
	g.Expect(offsetLineDistance(lineIndex, 0, 2)).To(Equal(1))
	g.Expect(offsetLineDistance(lineIndex, 2, 1)).To(Equal(1))
	g.Expect(offsetLineDistance(lineIndex, 0, 3)).To(Equal(2))
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

	applied, _, conflicts, result, _ := engine.ApplyWithConflicts(content, fixes)
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

	applied, _, conflicts, result, _ := engine.ApplyWithConflicts(content, fixes)
	g.Expect(applied).To(HaveLen(1))
	g.Expect(applied[0].ID).To(Equal(finding.FindingID("fix1")))
	g.Expect(conflicts).To(HaveLen(1))
	g.Expect(conflicts[0].Reason).To(Equal("overlapping edit"))
	g.Expect(conflicts[0].Finding.ID).To(Equal(finding.FindingID("fix2")))
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
		g.Expect(result[0].ID).To(Equal(finding.FindingID("fix1")))
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

	applied, _, conflicts, result, providerErrors := engine.ApplyWithConflicts(content, fixes)
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

	applied, _, conflicts, result, providerErrors := engine.ApplyWithConflicts(content, fixes)
	g.Expect(providerErrors).To(BeEmpty())
	g.Expect(applied).To(HaveLen(2))
	g.Expect(conflicts).To(BeEmpty())
	g.Expect(string(result)).To(Equal("a: MUCH-LONGER-REPLACEMENT\nb: medium\nc: X\n"))
}
