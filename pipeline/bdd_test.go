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

func runSingleDetectorPipeline(
	dryRun bool,
	detector pipeline.Detector,
) (*pipeline.PipelineResult, error) {
	cfg := pipeline.DefaultConfig()
	cfg.DryRun = dryRun
	cfg.MaxIterations = 1

	p, err := pipeline.New(cfg, ".", detector)
	if err != nil {
		return nil, err
	}

	return p.Run(context.Background())
}

var _ = Describe("Pipeline Lifecycle", func() {
	Describe("detect → triage → fix → verify loop", func() {
		DescribeTable(
			"detects and triages findings",
			func(detector pipeline.Detector, dryRun bool) {
				result, err := runSingleDetectorPipeline(dryRun, detector)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.TotalDetected).To(BeNumerically(">=", 1))
			},
			Entry(
				"auto-fixable issues in dry-run mode",
				fixerDetector("unused import", 5, `"os"`, `""`),
				true,
			),
			Entry(
				"non-fixable issues",
				singleFindingDetector(
					"linter",
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
				),
				false,
			),
			Entry(
				"dry run prevents applying fixes",
				fixerDetector("fixable", 1, "old", "new"),
				true,
			),
		)
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

			detector := singleFindingDetector(
				"t",
				mustBuild(
					"r1",
					"t",
					"msg",
					finding.SeverityInfo,
					"f.go",
					1,
					finding.FixStrategyNone,
					"",
					"",
				),
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
			result, err := fn.Process(context.Background(), []finding.Finding{
				mustBuild("r1", "t", "m", finding.SeverityError, "f.go", 1,
					finding.FixStrategyNone, "", ""),
				mustBuild("r2", "t", "m", finding.SeverityInfo, "f.go", 2,
					finding.FixStrategyNone, "", ""),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Rule).To(Equal("r1"))
		})

		It("NamedProcessorFunc sets the name", func() {
			p := pipeline.NamedProcessorFunc(
				"severity-filter",
				pipeline.ProcessorFunc(func(findings []finding.Finding) []finding.Finding {
					return findings
				}),
			)
			Expect(p.Name()).To(Equal("severity-filter"))
		})

		It("chains processors between detection and triage in the pipeline", func() {
			var processed [][]finding.Finding

			filter := pipeline.NamedProcessorFunc(
				"only-errors",
				pipeline.ProcessorFunc(func(findings []finding.Finding) []finding.Finding {
					processed = append(processed, findings)

					return finding.Filter(findings, finding.BySeverity(finding.SeverityError))
				}),
			)

			detector := pipeline.NamedDetectorFunc(
				"multi",
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

func singleFindingDetector(name string, f finding.Finding) pipeline.Detector {
	return pipeline.NamedDetectorFunc(name, func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{f}, nil
	})
}

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

func fixerDetector(msg string, line int, before, after string) pipeline.Detector {
	return singleFindingDetector("fixer", mustBuild(
		"r1", "fixer", msg, finding.SeverityWarning, "main.go", line,
		finding.FixStrategyDirect, before, after,
	))
}

func assertCanHandle(
	provider pipeline.FixProvider,
	before, after string,
	line, col int,
	expected bool,
) {
	f := codeFix(before, after, line, col)
	Expect(provider.CanHandle(f)).To(Equal(expected))
}

func lineRangeFix(
	before, after, file string,
	startLine, startCol, endLine, endCol int,
) finding.Finding {
	return finding.Finding{
		BeforeCode: before,
		AfterCode:  after,
		Range:      finding.NewRangePtr(file, startLine, startCol, endLine, endCol),
		Position:   finding.Pos(file, startLine, startCol),
	}
}

func codeFix(before, after string, line, col int) finding.Finding {
	return finding.Finding{
		BeforeCode: before,
		AfterCode:  after,
		Position:   finding.Pos("a.go", line, col),
	}
}

func offsetFix(before, after string, startOff, endOff int) finding.Finding {
	return finding.Finding{
		BeforeCode: before,
		AfterCode:  after,
		Range: &finding.Range{
			Start: finding.Position{File: "a.go", Offset: startOff},
			End:   finding.Position{File: "a.go", Offset: endOff},
		},
		Position: finding.Pos("a.go", 4, 2),
	}
}
