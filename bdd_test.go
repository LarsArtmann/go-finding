package finding_test

import (
	"context"
	"testing"
	"time"

	finding "github.com/larsartmann/go-finding"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBDD(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD Suite")
}

var _ = Describe("Finding Lifecycle", func() {
	Describe("creating findings", func() {
		Context("with the Builder API", func() {
			It("produces a valid finding from minimal input", func() {
				f, err := finding.NewBuilder(
					"nilcheck", "govet", "possible nil dereference",
					finding.SeverityError, finding.Pos("main.go", 42, 5),
				).Build()

				Expect(err).NotTo(HaveOccurred())
				Expect(f.IsValid()).To(BeTrue())
				Expect(f.Rule).To(Equal(finding.RuleName("nilcheck")))
				Expect(f.ToolName).To(Equal(finding.ToolName("govet")))
				Expect(f.Severity).To(Equal(finding.SeverityError))
				Expect(f.Position.File).To(Equal("main.go"))
				Expect(f.Position.Line).To(Equal(42))
			})

			It("auto-generates a stable ID", func() {
				f, _ := finding.NewBuilder(
					"rule", "tool", "msg",
					finding.SeverityWarning, finding.Pos("file.go", 10, 1),
				).Build()

				Expect(f.ID).NotTo(BeEmpty())
				Expect(string(f.ID)).To(ContainSubstring("tool"))
				Expect(string(f.ID)).To(ContainSubstring("rule"))
			})

			It("clamps confidence to [0.0, 1.0]", func() {
				f, _ := finding.NewBuilder(
					"r", "t", "m", finding.SeverityInfo, finding.Pos("f.go", 1, 1),
				).WithConfidence(1.5).Build()

				Expect(f.NormalizedConfidence()).To(BeNumerically("<=", 1.0))
			})

			It("rejects invalid state from Build", func() {
				_, err := finding.NewBuilder(
					"", "", "", finding.Severity("bogus"), finding.Pos("", 0, 0),
				).Build()
				Expect(err).To(HaveOccurred())
			})

			It("panics on invalid state from MustBuild", func() {
				Expect(func() {
					finding.NewBuilder(
						"", "", "", finding.Severity("bogus"), finding.Pos("", 0, 0),
					).MustBuild()
				}).To(Panic())
			})
		})
	})

	Describe("fix information", func() {
		It("detects when a finding has a direct fix", func() {
			f, _ := finding.NewBuilder(
				"r", "t", "m", finding.SeverityError, finding.Pos("f.go", 1, 1),
			).
				WithFixStrategy(finding.FixStrategyDirect).
				WithBeforeCode("x.foo").
				WithAfterCode("x.foo()").
				Build()

			Expect(f.HasFix()).To(BeTrue())
			Expect(f.HasSuggestion()).To(BeTrue())
		})

		It("detects when a suggestion-only finding has no auto-fix", func() {
			f, _ := finding.NewBuilder(
				"r", "t", "m", finding.SeverityWarning, finding.Pos("f.go", 1, 1),
			).
				WithFixStrategy(finding.FixStrategySuggest).
				WithSuggestion("consider using x.foo() instead").
				Build()

			Expect(f.HasFix()).To(BeFalse())
			Expect(f.HasSuggestion()).To(BeTrue())
		})

		It("generates a unified-diff preview", func() {
			f, _ := finding.NewBuilder(
				"r", "t", "m", finding.SeverityError, finding.Pos("f.go", 1, 1),
			).
				WithBeforeCode("old code").
				WithAfterCode("new code").
				Build()

			preview := f.Preview()
			Expect(preview).To(ContainSubstring("- old code"))
			Expect(preview).To(ContainSubstring("+ new code"))
		})
	})
})

var _ = Describe("Suppression Lifecycle", func() {
	It("marks a finding as suppressed with a valid reason", func() {
		past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		f, _ := finding.NewBuilder(
			"r", "t", "m", finding.SeverityInfo, finding.Pos("f.go", 1, 1),
		).WithSuppression(finding.Suppression{
			Kind:   finding.SuppressionInSource,
			Rule:   "r",
			Reason: "intentional",
		}).Build()

		Expect(f.IsSuppressedAt(past)).To(BeTrue())
	})

	It("re-activates a finding after suppression expiry", func() {
		expired := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		now := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)

		f, _ := finding.NewBuilder(
			"r", "t", "m", finding.SeverityInfo, finding.Pos("f.go", 1, 1),
		).WithSuppression(finding.Suppression{
			Kind:      finding.SuppressionInConfig,
			Rule:      "r",
			Reason:    "temporary",
			ExpiresAt: &expired,
		}).Build()

		Expect(f.IsSuppressedAt(now)).To(BeFalse())
	})
})

var _ = Describe("Report Filtering and Aggregation", func() {
	var report *finding.Report

	BeforeEach(func() {
		report = finding.NewReport(finding.ToolInfo{Name: "test", Version: "1.0"})

		findings := []finding.Finding{
			mustBuild("r1", "govet", "nil deref", finding.SeverityError, "a.go", 10),
			mustBuild("r2", "govet", "unused var", finding.SeverityWarning, "b.go", 20),
			mustBuild("r3", "staticcheck", "simplify", finding.SeverityInfo, "a.go", 30),
			mustBuild("r4", "staticcheck", "perf issue", finding.SeverityCritical, "c.go", 5),
		}
		for _, f := range findings {
			report.AddFinding(f)
		}

		report.ComputeSummary()
	})

	It("computes correct totals", func() {
		Expect(report.Len()).To(Equal(4))
		Expect(report.Summary.FilesAffected).To(Equal(3))
	})

	It("filters by severity", func() {
		errors := report.BySeverity(finding.SeverityError)
		Expect(errors).To(HaveLen(1))
		Expect(errors[0].Rule).To(Equal(finding.RuleName("r1")))
	})

	It("filters by tool", func() {
		results := report.Filter(finding.ByTool("govet"))
		Expect(results.Len()).To(Equal(2))
	})

	It("groups by file", func() {
		groups := finding.GroupByFile(report.FindingsSnapshot())
		Expect(groups).To(HaveLen(3))
		Expect(groups["a.go"]).To(HaveLen(2))
	})

	It("sorts by severity (most severe first)", func() {
		sorted := make([]finding.Finding, len(report.FindingsSnapshot()))
		copy(sorted, report.FindingsSnapshot())
		finding.SortBySeverity(sorted)
		Expect(sorted[0].Severity).To(Equal(finding.SeverityCritical))
	})

	It("iterates all findings including suppressed", func() {
		count := 0
		for range report.All() {
			count++
		}

		Expect(count).To(Equal(4))
	})
})

var _ = Describe("SARIF Round-Trip Fidelity", func() {
	It("preserves all Finding fields through export and import", func() {
		original := mustBuild("SA1000", "staticcheck", "use fmt.Sprintf",
			finding.SeverityWarning, "main.go", 42)

		report := finding.NewReport(finding.ToolInfo{Name: "test"})
		report.AddFinding(original)
		report.ComputeSummary()

		sarifData, err := report.ToSARIF()
		Expect(err).NotTo(HaveOccurred())

		parsed, err := finding.FindingsFromSARIF(context.Background(), sarifData)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed).To(HaveLen(1))

		Expect(parsed[0].Rule).To(Equal(finding.RuleName("SA1000")))
		Expect(parsed[0].ToolName).To(Equal(finding.ToolName("staticcheck")))
		Expect(parsed[0].Message).To(Equal("use fmt.Sprintf"))
		Expect(parsed[0].Position.File).To(Equal("main.go"))
		Expect(parsed[0].Position.Line).To(Equal(42))
	})

	It("excludes suppressed findings from SARIF output", func() {
		suppressed := mustBuild("r1", "t", "m", finding.SeverityInfo, "f.go", 1)
		suppressed.Suppression = &finding.Suppression{
			Kind: finding.SuppressionInSource, Rule: "r1", Reason: "ok",
		}

		report := finding.NewReport(finding.ToolInfo{Name: "test"})
		report.AddFinding(suppressed)
		report.AddFinding(mustBuild("r2", "t", "active", finding.SeverityWarning, "f.go", 2))
		report.ComputeSummary()

		sarifData, err := report.ToSARIF()
		Expect(err).NotTo(HaveOccurred())

		parsed, err := finding.FindingsFromSARIF(context.Background(), sarifData)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed).To(HaveLen(1))
		Expect(parsed[0].Rule).To(Equal(finding.RuleName("r2")))
	})

	It("maps SARIF levels back to severity", func() {
		Expect(finding.FromSARIFLevel("error")).To(Equal(finding.SeverityError))
		Expect(finding.FromSARIFLevel("warning")).To(Equal(finding.SeverityWarning))
		Expect(finding.FromSARIFLevel("note")).To(Equal(finding.SeverityInfo))
	})
})

var _ = Describe("Report Merging and Deduplication", func() {
	It("merges two reports and deduplicates by ID", func() {
		r1 := finding.NewReport(finding.ToolInfo{Name: "govet"})
		r1.AddFinding(mustBuild("r1", "govet", "msg", finding.SeverityWarning, "a.go", 10))

		r2 := finding.NewReport(finding.ToolInfo{Name: "govet"})
		r2.AddFinding(mustBuild("r1", "govet", "msg", finding.SeverityWarning, "a.go", 10))
		r2.AddFinding(mustBuild("r2", "govet", "other", finding.SeverityError, "b.go", 20))

		merged := finding.Combine(
			[]*finding.Report{r1, r2},
			finding.WithDeduplication(true),
			finding.WithDeduplicateBy(finding.DeduplicateByID),
		)
		merged.ComputeSummary()

		Expect(merged.Len()).To(Equal(2))
	})

	It("correlates findings from different tools on the same file", func() {
		findings := crossToolFindings("unused", 12)

		correlations := finding.Correlate(findings)
		Expect(correlations).NotTo(BeEmpty())

		var found bool

		for _, c := range correlations {
			if len(c.FindingIDs) == 2 {
				found = true

				break
			}
		}

		Expect(found).To(BeTrue())
	})
})
