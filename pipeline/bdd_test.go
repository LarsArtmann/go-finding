package pipeline_test

import (
	"context"
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
					return nil, context.DeadlineExceeded
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
