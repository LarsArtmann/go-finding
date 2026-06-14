package pipeline_test

import (
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FixProvider Contract", func() {
	var content []byte

	BeforeEach(func() {
		content = []byte("package main\n\nfunc main() {\n\told()\n}")
	})

	Describe("OffsetProvider", func() {
		var provider pipeline.OffsetProvider

		BeforeEach(func() {
			provider = pipeline.OffsetProvider{}
		})

		It("handles findings with byte-offset range info", func() {
			f := offsetFix("old()", "new()", 28, 33)
			Expect(provider.CanHandle(f)).To(BeTrue())
		})

		It("rejects findings without byte-offset range", func() {
			assertCanHandle(provider, "old", "new", 4, 2, false)
		})

		It("rejects findings with no before/after code", func() {
			f := finding.Finding{
				Range:    finding.NewRangePtr("a.go", 4, 2, 4, 7),
				Position: finding.Pos("a.go", 4, 2),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
		})

		It("produces a byte-level edit from offset range", func() {
			f := offsetFix("old()", "new()", 29, 34)
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(edits[0].Offset).To(Equal(29))
			Expect(edits[0].Length).To(Equal(5))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil when BeforeCode doesn't match content at offset", func() {
			f := offsetFix("WRONG", "new()", 29, 34)
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(BeNil())
		})
	})

	Describe("LineProvider", func() {
		var provider pipeline.LineProvider

		BeforeEach(func() {
			provider = pipeline.LineProvider{}
		})

		It("handles findings with line/column position", func() {
			f := lineRangeFix("old()", "new()", "a.go", 4, 2, 4, 7)
			Expect(provider.CanHandle(f)).To(BeTrue())
		})

		It("rejects findings with no line number", func() {
			assertCanHandle(provider, "old", "new", 0, 0, false)
		})

		It("handles insertion-only fixes (no BeforeCode)", func() {
			f := finding.Finding{
				AfterCode: "\tinserted",
				Position:  finding.Pos("a.go", 3, 1),
			}
			Expect(provider.CanHandle(f)).To(BeTrue())

			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(edits[0].Length).To(Equal(0))
			Expect(edits[0].IsInsert()).To(BeTrue())
		})

		It("produces range-based edits when Range has end", func() {
			f := lineRangeFix("old()", "new()", "a.go", 4, 2, 4, 7)
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil for out-of-bounds line numbers", func() {
			f := codeFix("old", "new", 100, 1)
			edits, err := provider.Edits(content, f)
			Expect(err).To(MatchError(pipeline.ErrPositionUnresolvable))
			Expect(edits).To(BeNil())
		})
	})

	Describe("SubstringProvider", func() {
		var provider pipeline.SubstringProvider

		BeforeEach(func() {
			provider = pipeline.SubstringProvider{}
		})

		It("handles findings with BeforeCode", func() {
			assertCanHandle(provider, "old()", "new()", 1, 1, true)
		})

		It("rejects findings without BeforeCode", func() {
			f := finding.Finding{
				AfterCode: "new()",
				Position:  finding.Pos("a.go", 1, 1),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
		})

		It("finds substring in content and produces edit", func() {
			f := codeFix("old()", "new()", 1, 1)
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil when BeforeCode not found in content", func() {
			f := codeFix("NONEXISTENT", "new()", 1, 1)
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(BeNil())
		})

		It("picks the occurrence nearest to the target line when ambiguous", func() {
			multiContent := []byte("line1: X\nline2: X\nline3: X")
			f := codeFix("X", "Y", 3, 0)
			edits, err := provider.Edits(multiContent, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(edits[0].Offset).To(Equal(25))
		})
	})

	Describe("provider chain precedence", func() {
		It("OffsetProvider takes precedence over LineProvider and SubstringProvider", func() {
			engine := pipeline.NewFixEngine()

			f := offsetFix("old()", "new()", 29, 34)

			result, applied, count := engine.Apply(content, []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
		})

		It("LineProvider is used when no byte offsets are available", func() {
			engine := pipeline.NewFixEngine()

			f := lineRangeFix("old()", "new()", "a.go", 4, 2, 4, 7)

			result, applied, count := engine.Apply(content, []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
		})

		It("SubstringProvider is the fallback", func() {
			engine := pipeline.NewFixEngine()

			f := codeFix("old", "new", 1, 1)

			result, applied, count := engine.Apply([]byte("old code"), []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("new code"))
		})
	})
})
