// pipeline demonstrates running the detection-fix-verify loop.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// staticDetector simulates a tool that always finds one issue.
type staticDetector struct{ file string }

func (staticDetector) Name() string { return "demo-detector" }

func (d staticDetector) Detect(_ context.Context) ([]finding.Finding, error) {
	return []finding.Finding{
		{
			ID:          finding.ID("demo:unused-var:" + d.file + ":2:2"),
			Rule:        "unused-var",
			ToolName:    "demo",
			Message:     "variable x is unused",
			Severity:    finding.SeverityWarning,
			Position:    finding.Position{File: d.file, Line: 2, Column: 2},
			BeforeCode:  "x := 42",
			AfterCode:   "",
			FixStrategy: finding.FixStrategyDirect,
		},
	}, nil
}

// fatal prints the error to stderr and exits.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func main() {
	tmpDir, err := os.MkdirTemp("", "pipeline-example-*")
	if err != nil {
		fatal(err)

		return
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}
	defer cleanup()

	srcFile := filepath.Join(tmpDir, "main.go")

	err = os.WriteFile(srcFile, []byte("package main\n\nx := 42\n"), 0o644)
	if err != nil {
		fatal(err)

		return
	}

	cfg := pipeline.Config{
		MaxIterations:     1,
		VerifyAfterFix:    false,
		ParallelDetectors: false,
	}

	p, err := pipeline.New(cfg, tmpDir, staticDetector{file: "main.go"})
	if err != nil {
		fatal(err)

		return
	}

	result, err := p.Run(context.Background())
	if err != nil {
		fatal(err)

		return
	}

	fmt.Printf("Iterations: %d\n", result.TotalIterations)
	fmt.Printf("Stable: %v\n", result.Stable())

	for _, it := range result.Iterations {
		fmt.Printf("  Iteration %d: %d found, %d applied, %d conflicts\n",
			it.Number, it.FindingsFound, it.Applied, it.Conflicts)
	}
}
