package finding_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func ExampleRange() {
	r := finding.NewRange("main.go", 10, 1, 15, 20)

	p := finding.Position{File: "main.go", Line: 12, Column: 5}
	fmt.Println("Contains:", r.Contains(p))
	fmt.Println("Valid:", r.IsValid())
	fmt.Println("HasEnd:", r.HasEnd())

	// Output:
	// Contains: true
	// Valid: true
	// HasEnd: true
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

func ExampleFindingError() {
	err := finding.NewValidationError("invalid input", nil)
	fmt.Println(finding.IsFindingError(err))
	fmt.Println(finding.GetCategory(err))

	ioErr := finding.NewIOError("read file", errors.New("permission denied"))
	fmt.Println(ioErr.Error())

	// Output:
	// true
	// validation
	// [io] read file: permission denied
}

func newExampleFinding(
	rule, tool, msg string,
	sev finding.Severity,
	file string,
	line, col int,
) finding.Finding {
	return finding.NewFinding(finding.RuleName(rule), finding.ToolName(tool), msg, sev, finding.Pos(file, line, col), 0)
}

func ExampleDiff() {
	before := []finding.Finding{
		newExampleFinding("rule-a", "tool", "msg a", finding.SeverityError, "a.go", 1, 1),
		newExampleFinding("rule-b", "tool", "msg b", finding.SeverityWarning, "b.go", 2, 1),
	}
	after := []finding.Finding{
		newExampleFinding("rule-a", "tool", "msg a", finding.SeverityError, "a.go", 1, 1),
		newExampleFinding("rule-c", "tool", "msg c", finding.SeverityInfo, "c.go", 3, 1),
	}

	result := finding.Diff(before, after)
	fmt.Println("Added:", len(result.Added))
	fmt.Println("Removed:", len(result.Removed))
	fmt.Println("Unchanged:", len(result.Unchanged))

	// Output:
	// Added: 1
	// Removed: 1
	// Unchanged: 1
}

func ExampleFormatText() {
	findings := []finding.Finding{
		newExampleFinding(
			"nilcheck",
			"govet",
			"possible nil dereference",
			finding.SeverityError,
			"main.go",
			42,
			5,
		),
		newExampleFinding(
			"unused",
			"staticcheck",
			"unused variable",
			finding.SeverityWarning,
			"util.go",
			10,
			3,
		),
	}

	finding.FormatText(os.Stdout, findings) //nolint:errcheck

	// Output:
	// main.go:42:5 [ERROR] nilcheck: possible nil dereference
	// util.go:10:3 [WARNING] unused: unused variable
}

func ExampleFormatMarkdown() {
	findings := []finding.Finding{
		newExampleFinding(
			"nilcheck", "govet", "possible nil dereference",
			finding.SeverityError, "main.go", 42, 5,
		),
	}

	finding.FormatMarkdown(os.Stdout, findings) //nolint:errcheck

	// Output:
	// | Location | Severity | Rule | Message |
	// |----------|----------|------|--------|
	// | main.go:42:5 | error | nilcheck | possible nil dereference |
}

func ExampleGeneratedFileFilter() {
	findings := []finding.Finding{
		newExampleFinding(
			"nilcheck", "govet", "possible nil dereference",
			finding.SeverityError, "main.go", 42, 5,
		),
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

func ExampleParseSeverity() {
	s, err := finding.ParseSeverity("warn")
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(s)

	s2 := finding.MustParseSeverity("high")
	fmt.Println(s2)

	// Output:
	// warning
	// error
}

func ExampleCategoryForLinter() {
	cat := finding.CategoryForLinter("gosec")
	fmt.Println(cat)

	cat2 := finding.CategoryForLinter("gocyclo")
	fmt.Println(cat2)

	cat3 := finding.CategoryForLinter("unknown")
	fmt.Println(cat3)

	// Output:
	// security
	// complexity
	// correctness
}

func ExampleParseCategory() {
	cat, err := finding.ParseCategory("security")
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(cat)

	cat2 := finding.MustParseCategory("error-handling")
	fmt.Println(cat2)

	// Output:
	// security
	// error-handling
}

func ExampleToolAdapter() {
	type lintOutput struct {
		Diagnostics []struct {
			Rule    string `json:"rule"`
			Message string `json:"message"`
			File    string `json:"file"`
			Line    int    `json:"line"`
		} `json:"diagnostics"`
	}

	parse := func(data []byte) (lintOutput, error) {
		var out lintOutput

		err := json.Unmarshal(data, &out)

		return out, err
	}

	convert := func(out lintOutput) ([]finding.Finding, error) {
		findings := make([]finding.Finding, 0, len(out.Diagnostics))
		for _, d := range out.Diagnostics {
			findings = append(findings, finding.Finding{
				ID: finding.GenerateID(
					"mylint",
					finding.RuleName(d.Rule),
					finding.Position{File: d.File, Line: d.Line},
				),
				Rule:     finding.RuleName(d.Rule),
				ToolName: "mylint",
				Message:  d.Message,
				Severity: finding.SeverityError,
				Category: finding.CategoryForLinter(d.Rule),
				Position: finding.Position{File: d.File, Line: d.Line},
			})
		}

		return findings, nil
	}

	run := func(_ context.Context) ([]byte, error) {
		return json.Marshal(lintOutput{
			Diagnostics: []struct {
				Rule    string `json:"rule"`
				Message string `json:"message"`
				File    string `json:"file"`
				Line    int    `json:"line"`
			}{
				{Rule: "no-unused", Message: "unused variable", File: "main.go", Line: 10},
			},
		})
	}

	adapter := finding.NewToolAdapter("mylint", run, parse, convert)

	findings, err := adapter.Detect(context.Background())
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println("count:", len(findings))
	fmt.Println("rule:", findings[0].Rule)

	// Output:
	// count: 1
	// rule: no-unused
}

func ExampleNewIntervalIndex() {
	type Item struct{ Name string }

	idx := finding.NewIntervalIndex([]finding.Interval[Item]{
		{Start: 1, End: 10, Value: Item{"alpha"}},
		{Start: 5, End: 15, Value: Item{"beta"}},
		{Start: 20, End: 30, Value: Item{"gamma"}},
	})

	overlaps := idx.Query(8, 12)
	for _, iv := range overlaps {
		fmt.Println(iv.Value.Name)
	}

	// Output:
	// alpha
	// beta
}

func ExampleMergeIter() {
	r1 := finding.NewReport(finding.ToolInfo{Name: "tool-a"})
	r1.AddFinding(finding.Finding{ID: "1", Rule: "r1", ToolName: "tool-a", Severity: finding.SeverityWarning})

	r2 := finding.NewReport(finding.ToolInfo{Name: "tool-b"})
	r2.AddFinding(finding.Finding{ID: "2", Rule: "r2", ToolName: "tool-b", Severity: finding.SeverityError})

	for f := range finding.MergeIter([]*finding.Report{r1, r2}) {
		fmt.Printf("%s: %s\n", f.ToolName, f.ID)
	}

	// Output:
	// tool-a: 1
	// tool-b: 2
}

func ExampleNewDetectorRegistry() {
	registry := finding.NewDetectorRegistry()
	registry.MustRegister("govet", func() finding.Detector {
		return finding.NamedDetectorFunc("govet", func(_ context.Context) ([]finding.Finding, error) {
			return []finding.Finding{{ID: "v1", Rule: "printf", Severity: finding.SeverityError}}, nil
		})
	})

	fmt.Println(registry.Has("govet"))
	fmt.Println(registry.Has("ghost"))
	fmt.Println(registry.Names())

	det, _ := registry.Build("govet")
	fmt.Println(det.Name())

	// Output:
	// true
	// false
	// [govet]
	// govet
}
