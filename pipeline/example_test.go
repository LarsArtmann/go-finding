package pipeline_test

import (
	"context"
	"fmt"

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
