package pipeline_test

import (
	"context"
	"errors"
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPipelineBDD(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline BDD Suite")
}

var _ = Describe("Pipeline Lifecycle", func() {
	Describe("detect → triage → fix → verify loop", func() {
		Context("when a detector finds auto-fixable issues", func() {
			It("detects and triages them correctly in dry-run mode", func() {
				detector := pipeline.NamedDetectorFunc("fixer",
					func(_ context.Context) ([]finding.Finding, error) {
						return []finding.Finding{
							mustBuild(
								"r1",
								"fixer",
								"unused import",
								finding.SeverityWarning,
								"main.go",
								5,
								finding.FixStrategyDirect,
								`"os"`,
								`""`,
							),
						}, nil
					})

				cfg := pipeline.DefaultConfig()
				cfg.DryRun = true
				cfg.MaxIterations = 1

				p, err := pipeline.New(cfg, ".", detector)
				Expect(err).NotTo(HaveOccurred())

				result, err := p.Run(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalDetected).To(BeNumerically(">=", 1))
			})
		})

		Context("when a detector finds non-fixable issues", func() {
			It("reports findings without attempting fixes", func() {
				detector := pipeline.NamedDetectorFunc(
					"linter",
					func(_ context.Context) ([]finding.Finding, error) {
						return []finding.Finding{
							mustBuild(
								"r1",
								"linter",
								"complex function",
								finding.SeverityInfo,
								"main.go",
								10,
								finding.FixStrategyNone,
								"",
								"",
							),
						}, nil
					},
				)

				cfg := pipeline.DefaultConfig()
				cfg.MaxIterations = 1

				p, err := pipeline.New(cfg, ".", detector)
				Expect(err).NotTo(HaveOccurred())

				result, err := p.Run(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalDetected).To(BeNumerically(">=", 1))
			})
		})

		Context("when dry run is enabled", func() {
			It("detects and triages but does not apply fixes", func() {
				detector := pipeline.NamedDetectorFunc(
					"fixer",
					func(_ context.Context) ([]finding.Finding, error) {
						return []finding.Finding{
							mustBuild(
								"r1",
								"fixer",
								"fixable",
								finding.SeverityWarning,
								"main.go",
								1,
								finding.FixStrategyDirect,
								"old",
								"new",
							),
						}, nil
					},
				)

				cfg := pipeline.DefaultConfig()
				cfg.DryRun = true
				cfg.MaxIterations = 1

				p, err := pipeline.New(cfg, ".", detector)
				Expect(err).NotTo(HaveOccurred())

				result, err := p.Run(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalDetected).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("graceful degradation", func() {
		It("collects findings from working detectors when one fails", func() {
			goodDetector := pipeline.NamedDetectorFunc(
				"good",
				func(_ context.Context) ([]finding.Finding, error) {
					return []finding.Finding{
						mustBuild(
							"r1",
							"good",
							"found something",
							finding.SeverityWarning,
							"f.go",
							1,
							finding.FixStrategyNone,
							"",
							"",
						),
					}, nil
				},
			)

			badDetector := pipeline.NamedDetectorFunc(
				"bad",
				func(_ context.Context) ([]finding.Finding, error) {
					return nil, errors.New("detector crashed")
				},
			)

			cfg := pipeline.DefaultConfig()
			cfg.GracefulDegradation = true
			cfg.MaxIterations = 1

			p, err := pipeline.New(cfg, ".", goodDetector, badDetector)
			Expect(err).NotTo(HaveOccurred())

			result, err := p.Run(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(result.TotalDetected).To(BeNumerically(">=", 1))
			Expect(result.PartialErrors).To(HaveLen(1))
			Expect(result.PartialErrors).To(HaveKey("bad"))
		})
	})

	Describe("conflict detection", func() {
		It("prevents overlapping fixes from being applied together", func() {
			findings := []finding.Finding{
				mustBuild("r1", "t", "fix A", finding.SeverityError, "f.go", 5,
					finding.FixStrategyDirect, "old", "newA"),
				mustBuild("r2", "t", "fix B", finding.SeverityError, "f.go", 5,
					finding.FixStrategyDirect, "old", "newB"),
			}

			conflicts := pipeline.AnalyzeConflicts(findings)
			Expect(conflicts).NotTo(BeEmpty())
		})
	})

	Describe("callbacks", func() {
		It("fires OnFinding for each detected finding", func() {
			var found []finding.Finding
			detector := pipeline.NamedDetectorFunc(
				"t",
				func(_ context.Context) ([]finding.Finding, error) {
					return []finding.Finding{
						mustBuild("r1", "t", "msg", finding.SeverityInfo, "f.go", 1,
							finding.FixStrategyNone, "", ""),
					}, nil
				},
			)

			cfg := pipeline.DefaultConfig()
			cfg.MaxIterations = 1
			cfg.OnFinding = func(f finding.Finding) {
				found = append(found, f)
			}

			p, _ := pipeline.New(cfg, ".", detector)
			_, err := p.Run(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(HaveLen(1))
			Expect(found[0].Rule).To(Equal("r1"))
		})
	})

	Describe("finding processors", func() {
		It("ProcessorFunc adapter works", func() {
			fn := pipeline.ProcessorFunc(func(findings []finding.Finding) []finding.Finding {
				return findings[:1]
			})
			Expect(fn.Name()).To(Equal("anonymous"))
			result := fn.Process([]finding.Finding{
				mustBuild("r1", "t", "m", finding.SeverityError, "f.go", 1,
					finding.FixStrategyNone, "", ""),
				mustBuild("r2", "t", "m", finding.SeverityInfo, "f.go", 2,
					finding.FixStrategyNone, "", ""),
			})
			Expect(result).To(HaveLen(1))
			Expect(result[0].Rule).To(Equal("r1"))
		})

		It("NamedProcessorFunc sets the name", func() {
			p := pipeline.NamedProcessorFunc("severity-filter",
				pipeline.ProcessorFunc(func(findings []finding.Finding) []finding.Finding {
					return findings
				}),
			)
			Expect(p.Name()).To(Equal("severity-filter"))
		})

		It("chains processors between detection and triage in the pipeline", func() {
			var processed [][]finding.Finding
			filter := pipeline.NamedProcessorFunc("only-errors",
				pipeline.ProcessorFunc(func(findings []finding.Finding) []finding.Finding {
					processed = append(processed, findings)
					return finding.Filter(findings, finding.BySeverity(finding.SeverityError))
				}),
			)

			detector := pipeline.NamedDetectorFunc("multi",
				func(_ context.Context) ([]finding.Finding, error) {
					return []finding.Finding{
						mustBuild("r1", "multi", "error", finding.SeverityError, "f.go", 1,
							finding.FixStrategyNone, "", ""),
						mustBuild("r2", "multi", "warning", finding.SeverityWarning, "f.go", 2,
							finding.FixStrategyNone, "", ""),
						mustBuild("r3", "multi", "info", finding.SeverityInfo, "f.go", 3,
							finding.FixStrategyNone, "", ""),
					}, nil
				},
			)

			cfg := pipeline.DefaultConfig()
			cfg.DryRun = true
			cfg.MaxIterations = 1
			cfg.Processors = []pipeline.FindingProcessor{filter}

			p, err := pipeline.New(cfg, ".", detector)
			Expect(err).NotTo(HaveOccurred())

			result, err := p.Run(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(processed).To(HaveLen(1))
			Expect(processed[0]).To(HaveLen(3))

			Expect(result.TotalDetected).To(BeNumerically(">=", 1))
		})
	})
})

func mustBuild(
	rule, tool, msg string, sev finding.Severity, file string, line int,
	fs finding.FixStrategy, before, after string,
) finding.Finding {
	f, err := finding.NewBuilder(rule, tool, msg, sev, finding.Pos(file, line, 1)).
		WithFixStrategy(fs).
		WithBeforeCode(before).
		WithAfterCode(after).
		Build()
	if err != nil {
		panic(err)
	}
	return f
}

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
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 28},
					End:   finding.Position{File: "a.go", Offset: 33},
				},
				Position: finding.Pos("a.go", 4, 2),
			}
			Expect(provider.CanHandle(f)).To(BeTrue())
		})

		It("rejects findings without byte-offset range", func() {
			f := finding.Finding{
				BeforeCode: "old",
				AfterCode:  "new",
				Position:   finding.Pos("a.go", 4, 2),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
		})

		It("rejects findings with no before/after code", func() {
			f := finding.Finding{
				Range:    finding.NewRangePtr("a.go", 4, 2, 4, 7),
				Position: finding.Pos("a.go", 4, 2),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
		})

		It("produces a byte-level edit from offset range", func() {
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 29},
					End:   finding.Position{File: "a.go", Offset: 34},
				},
				Position: finding.Pos("a.go", 4, 2),
			}
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(edits[0].Offset).To(Equal(29))
			Expect(edits[0].Length).To(Equal(5))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil when BeforeCode doesn't match content at offset", func() {
			f := finding.Finding{
				BeforeCode: "WRONG",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 29},
					End:   finding.Position{File: "a.go", Offset: 34},
				},
				Position: finding.Pos("a.go", 4, 2),
			}
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
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Position:   finding.Pos("a.go", 4, 2),
				Range:      finding.NewRangePtr("a.go", 4, 2, 4, 7),
			}
			Expect(provider.CanHandle(f)).To(BeTrue())
		})

		It("rejects findings with no line number", func() {
			f := finding.Finding{
				BeforeCode: "old",
				AfterCode:  "new",
				Position:   finding.Pos("a.go", 0, 0),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
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
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range:      finding.NewRangePtr("a.go", 4, 2, 4, 7),
				Position:   finding.Pos("a.go", 4, 2),
			}
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil for out-of-bounds line numbers", func() {
			f := finding.Finding{
				BeforeCode: "old",
				AfterCode:  "new",
				Position:   finding.Pos("a.go", 100, 1),
			}
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(BeNil())
		})
	})

	Describe("SubstringProvider", func() {
		var provider pipeline.SubstringProvider

		BeforeEach(func() {
			provider = pipeline.SubstringProvider{}
		})

		It("handles findings with BeforeCode", func() {
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Position:   finding.Pos("a.go", 1, 1),
			}
			Expect(provider.CanHandle(f)).To(BeTrue())
		})

		It("rejects findings without BeforeCode", func() {
			f := finding.Finding{
				AfterCode: "new()",
				Position:  finding.Pos("a.go", 1, 1),
			}
			Expect(provider.CanHandle(f)).To(BeFalse())
		})

		It("finds substring in content and produces edit", func() {
			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Position:   finding.Pos("a.go", 1, 1),
			}
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(string(edits[0].Replacement)).To(Equal("new()"))
		})

		It("returns nil when BeforeCode not found in content", func() {
			f := finding.Finding{
				BeforeCode: "NONEXISTENT",
				AfterCode:  "new()",
				Position:   finding.Pos("a.go", 1, 1),
			}
			edits, err := provider.Edits(content, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(BeNil())
		})

		It("picks the occurrence nearest to the target line when ambiguous", func() {
			multiContent := []byte("line1: X\nline2: X\nline3: X")
			f := finding.Finding{
				BeforeCode: "X",
				AfterCode:  "Y",
				Position:   finding.Pos("a.go", 3, 0),
			}
			edits, err := provider.Edits(multiContent, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(edits).To(HaveLen(1))
			Expect(edits[0].Offset).To(Equal(25))
		})
	})

	Describe("provider chain precedence", func() {
		It("OffsetProvider takes precedence over LineProvider and SubstringProvider", func() {
			engine := pipeline.NewFixEngine()

			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range: &finding.Range{
					Start: finding.Position{File: "a.go", Offset: 29},
					End:   finding.Position{File: "a.go", Offset: 34},
				},
				Position: finding.Pos("a.go", 4, 2),
			}

			result, applied, count := engine.Apply(content, []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
		})

		It("LineProvider is used when no byte offsets are available", func() {
			engine := pipeline.NewFixEngine()

			f := finding.Finding{
				BeforeCode: "old()",
				AfterCode:  "new()",
				Range:      finding.NewRangePtr("a.go", 4, 2, 4, 7),
				Position:   finding.Pos("a.go", 4, 2),
			}

			result, applied, count := engine.Apply(content, []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("package main\n\nfunc main() {\n\tnew()\n}"))
		})

		It("SubstringProvider is the fallback", func() {
			engine := pipeline.NewFixEngine()

			f := finding.Finding{
				BeforeCode: "old",
				AfterCode:  "new",
				Position:   finding.Pos("a.go", 1, 1),
			}

			result, applied, count := engine.Apply([]byte("old code"), []finding.Finding{f})
			Expect(count).To(Equal(1))
			Expect(applied).To(HaveLen(1))
			Expect(string(result)).To(Equal("new code"))
		})
	})
})
