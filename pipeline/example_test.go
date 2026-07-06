package pipeline_test

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func ExampleConfigFromFile() {
	data := []byte(`{
		"maxIterations": 3,
		"parallelDetectors": true,
		"timeout": "30s"
	}`)

	cfg, err := pipeline.ConfigFromFile(data)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(cfg.MaxIterations)
	fmt.Println(cfg.ParallelDetectors)

	// Output:
	// 3
	// true
}

func ExampleNewLineShiftMap() {
	// Original content: 3 lines.
	original := []byte("line1\nline2\nline3\n")

	// Insert a new line before line 3 (byte offset 12).
	edits := []pipeline.FixEdit{
		{Offset: 12, Length: 0, Replacement: []byte("inserted\n")},
	}

	shiftMap := pipeline.NewLineShiftMap(original, edits)

	// Line 3 shifts to line 4 after inserting one line before it.
	fmt.Println(shiftMap.ShiftedLine(3))

	// Output:
	// 4
}

func ExampleDetectorRegistry_integration() {
	// Register a detector in a DetectorRegistry, then drive a Pipeline.
	registry := finding.NewDetectorRegistry()
	registry.MustRegister("example", func() finding.Detector {
		return finding.NamedDetectorFunc("example", func(_ context.Context) ([]finding.Finding, error) {
			return []finding.Finding{
				{ID: "1", Rule: "demo", Severity: finding.SeverityWarning},
			}, nil
		})
	})

	// Build and confirm.
	det, err := registry.Build("example")
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(det.Name())

	// Output:
	// example
}

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
