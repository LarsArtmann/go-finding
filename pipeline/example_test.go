package pipeline_test

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
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

func ExampleNewFlightRecorderHook() {
	hook, err := pipeline.NewFlightRecorderHook(pipeline.DefaultFlightRecorderConfig())
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("enabled:", hook.Enabled())

	hook.Close()

	fmt.Println("enabled after close:", hook.Enabled())

	// Output:
	// enabled: true
	// enabled after close: false
}

func ExampleConfigFile_ResolveFlightRecorder() {
	// Config with flight recorder enabled.
	data := []byte(`{
		"flightRecorder": {
			"enabled": true,
			"slowStageThreshold": "10s"
		}
	}`)

	var cf pipeline.ConfigFile
	if err := json.Unmarshal(data, &cf); err != nil {
		fmt.Println("error:", err)

		return
	}

	hook, err := cf.ResolveFlightRecorder()
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	if hook != nil {
		fmt.Println("flight recorder created")
		hook.Close()
	} else {
		fmt.Println("no flight recorder")
	}

	// Output:
	// flight recorder created
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

// ExampleMetrics demonstrates recording fix outcomes and reading them back
// from a snapshot, the way the pipeline's fix stage does.
func ExampleMetrics() {
	m := pipeline.NewMetrics()

	m.RecordOutcome(pipeline.FixOutcomeApplied)
	m.RecordOutcome(pipeline.FixOutcomeApplied)
	m.RecordOutcome(pipeline.FixOutcomeRefused)
	m.RecordFixes(2)

	snap := m.Snapshot()

	fmt.Println("applied:", snap.OutcomeCounts[pipeline.FixOutcomeApplied])
	fmt.Println("refused:", snap.OutcomeCounts[pipeline.FixOutcomeRefused])
	fmt.Println("fixes:", snap.FixesApplied)

	// Output:
	// applied: 2
	// refused: 1
	// fixes: 2
}

// ExampleFixApplier_ApplyDryRun demonstrates plan-without-apply: the dry run
// reports the outcomes a real fix run would produce (including shift maps)
// while leaving every file untouched and creating no backups.
func ExampleFixApplier_ApplyDryRun() {
	dir, err := os.MkdirTemp("", "dryrun-*")
	if err != nil {
		fmt.Println("error:", err)

		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	file := filepath.Join(dir, "dry.go")
	const original = "package main\n\nfunc main() {\n\told()\n}\n"
	if err := os.WriteFile(file, []byte(original), 0o600); err != nil {
		fmt.Println("error:", err)

		return
	}

	applier, err := pipeline.NewFixApplier(dir)
	if err != nil {
		fmt.Println("error:", err)

		return
	}
	defer func() { _ = applier.Close() }()

	fix, ferr := finding.NewBuilder(
		finding.RuleName("rename"), finding.ToolName("example"),
		"rename old",
		finding.SeverityWarning,
		finding.Pos("dry.go", 4, 2),
	).
		WithFixStrategy(finding.FixStrategyDirect).
		WithBeforeCode("old()").
		WithAfterCode("new()").
		Build()
	if ferr != nil {
		fmt.Println("error:", ferr)

		return
	}

	report, err := applier.ApplyDryRun(context.Background(), []finding.Finding{fix})
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	for _, oc := range report.Outcomes {
		fmt.Println(oc.Finding.Rule, oc.Status)
	}

	data, _ := os.ReadFile(file)
	fmt.Println("unchanged:", string(data) == original)

	// Output:
	// rename applied
	// unchanged: true
}
