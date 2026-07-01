package pipeline_test

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func ExamplePipeline() {
	detector := pipeline.NamedDetectorFunc(
		"example",
		func(_ context.Context) ([]finding.Finding, error) {
			return []finding.Finding{
				{
					ID:          "example:R1:main.go:1:1",
					Rule:        "R1",
					ToolName:    "example",
					Message:     "example finding",
					Severity:    finding.SeverityWarning,
					Position:    finding.Pos("main.go", 1, 1),
					FixStrategy: finding.FixStrategyNone,
				},
			}, nil
		},
	)

	cfg := pipeline.Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		Timeout:           5 * time.Second,
	}

	p, err := pipeline.New(cfg, ".", detector)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	result, err := p.Run(context.Background())
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("Iterations:", result.TotalIterations)
	fmt.Println("Findings:", result.TotalDetected)

	// Output:
	// Iterations: 1
	// Findings: 1
}

func ExampleGeneratedFileFilter() {
	findings := []finding.Finding{
		{
			ID:       finding.GenerateID("nilcheck", "R1", finding.Pos("main.go", 42, 5)),
			Rule:     "R1",
			ToolName: "govet",
			Message:  "possible nil dereference",
			Severity: finding.SeverityError,
			Position: finding.Pos("main.go", 42, 5),
		},
	}

	// Without config the filter is disabled — all findings pass through.
	noop, err := pipeline.NewGeneratedFileFilter(nil)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(noop.Name())

	result, err := noop.Transform(context.Background(), findings)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("passed:", len(result))

	// With FilterAll the filter is enabled and will evaluate files.
	optCfg, err := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	active, err := pipeline.NewGeneratedFileFilter(nil, optCfg)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(active.Name())

	// Output:
	// generated-file-filter
	// passed: 1
	// generated-file-filter
}
