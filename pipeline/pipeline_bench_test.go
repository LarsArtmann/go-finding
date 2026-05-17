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

func benchPipelineDryRun(b *testing.B, findingCount int) {
	b.Helper()
	findings := generateFindings(findingCount)

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		DryRun:            true,
	}
	det := NamedDetectorFunc("bench", func(_ context.Context) ([]finding.Finding, error) {
		return findings, nil
	})

	p, err := New(cfg, b.TempDir(), det)
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

func BenchmarkPipeline_DryRun_100(b *testing.B)   { benchPipelineDryRun(b, 100) }
func BenchmarkPipeline_DryRun_1000(b *testing.B)  { benchPipelineDryRun(b, 1000) }
func BenchmarkPipeline_DryRun_10000(b *testing.B) { benchPipelineDryRun(b, 10000) }

func benchPipelineWithCorrelation(b *testing.B, findingCount int) {
	b.Helper()
	findings := generateFindings(findingCount)

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		DryRun:            true,
		CorrelateFindings: true,
	}
	det := NamedDetectorFunc("bench", func(_ context.Context) ([]finding.Finding, error) {
		return findings, nil
	})

	p, err := New(cfg, b.TempDir(), det)
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

func BenchmarkPipeline_Correlate_100(b *testing.B)   { benchPipelineWithCorrelation(b, 100) }
func BenchmarkPipeline_Correlate_1000(b *testing.B)  { benchPipelineWithCorrelation(b, 1000) }
func BenchmarkPipeline_Correlate_10000(b *testing.B) { benchPipelineWithCorrelation(b, 10000) }

func benchPipelineParallel(b *testing.B, findingCount, detectorCount int) {
	b.Helper()
	findings := generateFindings(findingCount)

	detectors := make([]Detector, detectorCount)
	for i := range detectors {
		detectors[i] = NamedDetectorFunc(
			fmt.Sprintf("det-%d", i),
			func(_ context.Context) ([]finding.Finding, error) {
				return findings, nil
			},
		)
	}

	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: true,
		DryRun:            true,
	}

	p, err := New(cfg, b.TempDir(), detectors...)
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

func BenchmarkPipeline_Parallel_5Det_1k(b *testing.B)   { benchPipelineParallel(b, 1000, 5) }
func BenchmarkPipeline_Parallel_5Det_10k(b *testing.B)  { benchPipelineParallel(b, 10000, 5) }
func BenchmarkPipeline_Parallel_10Det_1k(b *testing.B)  { benchPipelineParallel(b, 1000, 10) }
func BenchmarkPipeline_Parallel_10Det_10k(b *testing.B) { benchPipelineParallel(b, 10000, 10) }
