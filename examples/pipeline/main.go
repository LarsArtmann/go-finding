// pipeline demonstrates running the detection-fix-verify loop.
package main

import (
	"context"
	"fmt"
	"log"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// staticDetector simulates a tool that always finds one issue.
type staticDetector struct{}

func (staticDetector) Name() string { return "demo-detector" }

func (staticDetector) Detect(_ context.Context) ([]finding.Finding, error) {
	return []finding.Finding{
		{
			ID:          "demo:unused-var:main.go:10:2",
			Rule:        "unused-var",
			ToolName:    "demo",
			Message:     "variable x is unused",
			Severity:    finding.SeverityWarning,
			Position:    finding.Position{File: "main.go", Line: 10, Column: 2},
			BeforeCode:  "x := 42",
			AfterCode:   "",
			FixStrategy: finding.FixStrategyDirect,
		},
	}, nil
}

func main() {
	cfg := pipeline.Config{
		MaxIterations:     1,
		VerifyAfterFix:    false,
		ParallelDetectors: false,
	}

	p, err := pipeline.New(cfg, ".", staticDetector{})
	if err != nil {
		log.Fatal(err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Iterations: %d\n", result.TotalIterations)
	fmt.Printf("Stable: %v\n", result.Stable)

	for _, it := range result.Iterations {
		fmt.Printf("  Iteration %d: %d found, %d applied, %d conflicts\n",
			it.Number, it.FindingsFound, it.Applied, it.Conflicts)
	}
}
