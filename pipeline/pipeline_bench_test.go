package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/larsartmann/go-finding"
)

func generateFindings(n int) []finding.Finding {
	findings := make([]finding.Finding, n)
	for i := range findings {
		findings[i] = finding.Finding{
			ID:       fmt.Sprintf("tool:rule:file.go:%d:1", i),
			Rule:     "SA1000",
			ToolName: "bench",
			Message:  fmt.Sprintf("benchmark finding %d", i),
			Severity: []finding.Severity{finding.SeverityInfo, finding.SeverityWarning, finding.SeverityError, finding.SeverityCritical}[i%4],
			Position: finding.Position{
				File:   fmt.Sprintf("file%d.go", i%50),
				Line:   i + 1,
				Column: 1,
			},
			FixStrategy: []finding.FixStrategy{finding.FixStrategyNone, finding.FixStrategySuggest, finding.FixStrategyDirect}[i%3],
			Category:    []finding.Category{finding.CategorySecurity, finding.CategoryStyle, finding.CategoryPerformance}[i%3],
		}
	}
	return findings
}

type benchConfig struct {
	findings          int
	parallel          bool
	dryRun            bool
	correlateFindings bool
	detectorCount     int
}

func runBenchPipeline(b *testing.B, cfg benchConfig) {
	b.Helper()
	findings := generateFindings(cfg.findings)

	pipelineCfg := Config{
		MaxIterations:     1,
		ParallelDetectors: cfg.parallel,
		DryRun:            cfg.dryRun,
		CorrelateFindings: cfg.correlateFindings,
	}

	detectors := make([]Detector, 0, max(cfg.detectorCount, 1))
	for i := range max(cfg.detectorCount, 1) {
		name := "bench"
		if cfg.detectorCount > 1 {
			name = fmt.Sprintf("det-%d", i)
		}
		f := findings
		detectors = append(detectors, NamedDetectorFunc(name,
			func(_ context.Context) ([]finding.Finding, error) {
				return f, nil
			}))
	}

	p, err := New(pipelineCfg, b.TempDir(), detectors...)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		if _, err := p.Run(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPipeline_DryRun_100(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 100, dryRun: true})
}

func BenchmarkPipeline_DryRun_1000(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 1000, dryRun: true})
}

func BenchmarkPipeline_DryRun_10000(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 10000, dryRun: true})
}

func BenchmarkPipeline_Correlate_100(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 100, dryRun: true, correlateFindings: true})
}

func BenchmarkPipeline_Correlate_1000(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 1000, dryRun: true, correlateFindings: true})
}

func BenchmarkPipeline_Correlate_10000(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 10000, dryRun: true, correlateFindings: true})
}

func BenchmarkPipeline_Parallel_5Det_1k(b *testing.B) {
	runBenchPipeline(b, benchConfig{findings: 1000, parallel: true, dryRun: true, detectorCount: 5})
}

func BenchmarkPipeline_Parallel_5Det_10k(b *testing.B) {
	runBenchPipeline(
		b,
		benchConfig{findings: 10000, parallel: true, dryRun: true, detectorCount: 5},
	)
}

func BenchmarkPipeline_Parallel_10Det_1k(b *testing.B) {
	runBenchPipeline(
		b,
		benchConfig{findings: 1000, parallel: true, dryRun: true, detectorCount: 10},
	)
}

func BenchmarkPipeline_Parallel_10Det_10k(b *testing.B) {
	runBenchPipeline(
		b,
		benchConfig{findings: 10000, parallel: true, dryRun: true, detectorCount: 10},
	)
}
