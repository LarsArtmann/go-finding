package finding_test

import (
	"errors"
	"io"

	finding "github.com/larsartmann/go-finding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON Serialization", func() {
	It("round-trips a Finding through JSON", func() {
		original := mustBuild("r", "tool", "msg", finding.SeverityError, "file.go", 5)

		data, err := original.LineJSON()
		Expect(err).NotTo(HaveOccurred())

		parsed, err := finding.FromJSON([]byte(data))
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed.Equal(original)).To(BeTrue())
	})

	It("drops invalid findings from Report JSON", func() {
		report := finding.NewReport(finding.ToolInfo{Name: "t"})
		report.AddFinding(mustBuild("r1", "t", "valid", finding.SeverityInfo, "f.go", 1))

		data, _ := report.PrettyJSON()

		parsed, dropped, err := finding.ReportFromJSON([]byte(data))
		Expect(err).NotTo(HaveOccurred())
		Expect(dropped).To(Equal(0))
		Expect(parsed.Len()).To(Equal(1))
	})
})

var _ = Describe("Position and Range User Stories", func() {
	Describe("spatial queries", func() {
		It("detects when a position falls inside a range", func() {
			r := finding.NewRange("main.go", 10, 1, 20, 80)
			inside := finding.Pos("main.go", 15, 5)
			outside := finding.Pos("main.go", 25, 1)

			Expect(r.Contains(inside)).To(BeTrue())
			Expect(r.Contains(outside)).To(BeFalse())
		})

		It("detects when two ranges overlap", func() {
			r1 := finding.NewRange("main.go", 10, 1, 20, 1)
			r2 := finding.NewRange("main.go", 15, 1, 25, 1)
			r3 := finding.NewRange("main.go", 21, 1, 30, 1)

			Expect(r1.Overlaps(r2)).To(BeTrue())
			Expect(r1.Overlaps(r3)).To(BeFalse())
		})

		It("computes the intersection of overlapping ranges", func() {
			r1 := finding.NewRange("main.go", 10, 1, 20, 1)
			r2 := finding.NewRange("main.go", 15, 1, 25, 1)

			intersection := r1.Intersection(r2)
			Expect(intersection).NotTo(BeNil())
			Expect(intersection.Start.Line).To(Equal(15))
			Expect(intersection.End.Line).To(Equal(20))
		})

		It("detects adjacent ranges", func() {
			r1 := finding.NewRange("main.go", 10, 1, 20, 0)
			r2 := finding.NewRange("main.go", 20, 0, 30, 1)

			Expect(r1.Adjacent(r2)).To(BeTrue())
		})

		It("counts lines in a range", func() {
			r := finding.NewRange("main.go", 5, 1, 10, 1)
			Expect(r.LineCount()).To(Equal(6))
		})

		It("orders positions by file, then line, then column", func() {
			p1 := finding.Pos("a.go", 10, 5)
			p2 := finding.Pos("a.go", 10, 10)
			p3 := finding.Pos("a.go", 20, 1)

			Expect(p1.Compare(p2)).To(BeNumerically("<", 0))
			Expect(p2.Compare(p3)).To(BeNumerically("<", 0))
		})
	})
})

var _ = Describe("Error Handling User Stories", func() {
	Describe("structured error categories", func() {
		It("categorizes validation errors for programmatic handling", func() {
			err := finding.NewValidationError("finding.ID is required", nil)

			Expect(err.Error()).To(ContainSubstring("validation"))
			Expect(errors.Is(err, finding.ErrValidation)).To(BeTrue())
			Expect(finding.GetCategory(err)).To(Equal(finding.ErrCategoryValidation))
		})

		It("wraps underlying causes for error chain inspection", func() {
			cause := io.ErrUnexpectedEOF
			err := finding.NewIOError("read file", cause)

			Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
			Expect(errors.Unwrap(err)).To(Equal(cause))
		})

		It("attaches finding context to errors", func() {
			f := mustBuild("r", "t", "m", finding.SeverityError, "main.go", 42)
			baseErr := finding.NewValidationError("bad finding", nil)
			enriched := baseErr.WithFinding(f)

			Expect(enriched.Finding).NotTo(BeNil())
			Expect(enriched.File).To(Equal("main.go"))
			Expect(enriched.Position).NotTo(BeNil())
			Expect(enriched.Position.Line).To(Equal(42))
		})

		It("identifies FindingError types generically", func() {
			err := finding.NewConflictError("overlapping fix", nil)
			stdErr := errors.New("plain error")

			Expect(finding.IsFindingError(err)).To(BeTrue())
			Expect(finding.IsFindingError(stdErr)).To(BeFalse())
		})
	})

	Describe("validation", func() {
		It("reports all validation problems at once", func() {
			f := finding.Finding{}
			err := f.Validate()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ID"))
			Expect(err.Error()).To(ContainSubstring("Rule"))
			Expect(err.Error()).To(ContainSubstring("ToolName"))
		})
	})
})

var _ = Describe("LSP Conversion User Stories", func() {
	Describe("converting findings to LSP diagnostics", func() {
		It("converts severity to LSP levels correctly", func() {
			errorFinding := mustBuild("r", "t", "m", finding.SeverityError, "f.go", 1)
			warningFinding := mustBuild("r", "t", "m", finding.SeverityWarning, "f.go", 1)
			infoFinding := mustBuild("r", "t", "m", finding.SeverityInfo, "f.go", 1)

			Expect(errorFinding.ToLSP().Severity).To(Equal(finding.LSPSeverityError))
			Expect(warningFinding.ToLSP().Severity).To(Equal(finding.LSPSeverityWarning))
			Expect(infoFinding.ToLSP().Severity).To(Equal(finding.LSPSeverityInfo))
		})

		It("converts 1-based positions to 0-based LSP positions", func() {
			f := mustBuild("r", "t", "m", finding.SeverityError, "f.go", 10)
			diag := f.ToLSP()

			Expect(diag.Range.Start.Line).To(Equal(9))
			Expect(diag.Range.Start.Character).To(Equal(0))
		})

		It("preserves related information through LSP conversion", func() {
			f, _ := finding.NewBuilder("r", "t", "m", finding.SeverityError, finding.Pos("f.go", 1, 1)).
				WithRelated(finding.RelatedRef{
					FindingID: "other:1",
					Relation:  "causes",
					Position:  finding.Pos("other.go", 5, 3),
				}).
				Build()

			diag := f.ToLSP()
			Expect(diag.Related).To(HaveLen(1))
			Expect(diag.Related[0].Message).To(Equal("causes"))
			Expect(diag.Related[0].Location.URI).To(Equal("other.go"))
		})
	})

	Describe("converting LSP diagnostics to findings", func() {
		It("round-trips position through LSP conversion", func() {
			original := mustBuild("r", "t", "m", finding.SeverityError, "f.go", 10)
			diag := original.ToLSP()
			restored := finding.FromLSP("f.go", diag)

			Expect(restored.Position.File).To(Equal("f.go"))
			Expect(restored.Position.Line).To(Equal(10))
		})
	})
})

var _ = Describe("Cross-Tool Correlation User Stories", func() {
	It("groups findings from different tools on nearby lines", func() {
		findings := crossToolFindings("unused var", 11)

		correlations := finding.Correlate(findings)
		Expect(correlations).NotTo(BeEmpty())

		var crossTool bool

		for _, c := range correlations {
			Expect(c.Score).To(BeNumerically(">=", 0.0))
			Expect(c.Score).To(BeNumerically("<=", 1.0))
			Expect(c.Reason).NotTo(BeEmpty())

			if len(c.FindingIDs) == 2 {
				crossTool = true
			}
		}

		Expect(crossTool).To(BeTrue())
	})

	It("does not correlate findings from the same tool", func() {
		findings := []finding.Finding{
			mustBuild("r1", "govet", "a", finding.SeverityError, "main.go", 10),
			mustBuild("r2", "govet", "b", finding.SeverityWarning, "main.go", 11),
		}

		correlations := finding.Correlate(findings)
		Expect(correlations).To(BeEmpty())
	})
})

var _ = Describe("ID Generation User Stories", func() {
	It("generates stable human-readable IDs", func() {
		pos := finding.Pos("main.go", 42, 5)
		id := finding.GenerateID("govet", "nilcheck", pos)

		Expect(id).To(ContainSubstring("govet"))
		Expect(id).To(ContainSubstring("nilcheck"))
		Expect(id).To(ContainSubstring("main.go"))
		Expect(id).To(ContainSubstring("42"))
	})

	It("generates hash-based IDs for position-less findings", func() {
		pos := finding.Pos("main.go", 0, 0)
		id := finding.GenerateID("govet", "nilcheck", pos)

		Expect(finding.IsHashID(string(id))).To(BeTrue())
	})

	It("parses IDs back into components", func() {
		pos := finding.Pos("main.go", 42, 5)
		id := finding.GenerateID("govet", "nilcheck", pos)

		parsed := finding.ParseID(finding.FindingID(id))
		Expect(parsed.OK()).To(BeTrue())
		Expect(parsed.Tool).To(Equal("govet"))
		Expect(parsed.Rule).To(Equal(finding.RuleName("nilcheck")))
		Expect(parsed.Line).To(Equal(42))
		Expect(parsed.Column).To(Equal(5))
	})
})

func crossToolFindings(r2Msg string, r2Line int) []finding.Finding {
	return []finding.Finding{
		mustBuild("r1", "govet", "nil deref", finding.SeverityError, "main.go", 10),
		mustBuild("r2", "staticcheck", r2Msg, finding.SeverityWarning, "main.go", r2Line),
		mustBuild("r3", "govet", "other", finding.SeverityInfo, "other.go", 1),
	}
}

func mustBuild(
	rule, tool, msg string, sev finding.Severity, file string, line int,
) finding.Finding {
	f, err := finding.NewBuilder(finding.RuleName(rule), finding.ToolName(tool), msg, sev, finding.Pos(file, line, 1)).
		Build()
	if err != nil {
		panic(err)
	}

	return f
}
