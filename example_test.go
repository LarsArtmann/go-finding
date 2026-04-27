package finding_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func ExampleGenerateID() {
	pos := finding.Position{File: "main.go", Line: 42, Column: 5}
	id := finding.GenerateID("govet", "printf", pos)
	fmt.Println(id)

	// Hash-based ID when line is 0
	posNoLine := finding.Position{File: "main.go"}
	hashID := finding.GenerateID("govet", "printf", posNoLine)
	fmt.Println(finding.IsHashID(hashID))

	// Output:
	// govet:printf:main.go:42:5
	// true
}

func ExampleParseID() {
	p := finding.ParseID("govet:printf:main.go:42:5")
	if !p.OK() {
		fmt.Println("invalid ID")

		return
	}

	fmt.Printf("tool=%s rule=%s file=%s line=%d col=%d\n", p.Tool, p.Rule, p.File, p.Line, p.Column)

	// Output:
	// tool=govet rule=printf file=main.go line=42 col=5
}

func ExampleFilter() {
	findings := []finding.Finding{
		{
			ID:       "1",
			Rule:     "R1",
			Severity: finding.SeverityInfo,
			Position: finding.Position{File: "a.go"},
		},
		{
			ID:       "2",
			Rule:     "R2",
			Severity: finding.SeverityError,
			Position: finding.Position{File: "b.go"},
		},
		{
			ID:       "3",
			Rule:     "R1",
			Severity: finding.SeverityWarning,
			Position: finding.Position{File: "a.go"},
		},
	}

	errors := finding.Filter(findings, finding.BySeverityAtLeast(finding.SeverityError))
	fmt.Println("Errors:", len(errors))

	fromA := finding.Filter(findings, finding.ByFile("a.go"))
	fmt.Println("In a.go:", len(fromA))

	combined := finding.Filter(findings,
		finding.ByRule("R1"),
		finding.ByFile("a.go"),
	)
	fmt.Println("R1 in a.go:", len(combined))

	// Output:
	// Errors: 1
	// In a.go: 2
	// R1 in a.go: 2
}

func ExampleGroupByFile() {
	findings := []finding.Finding{
		{ID: "1", Position: finding.Position{File: "a.go"}},
		{ID: "2", Position: finding.Position{File: "b.go"}},
		{ID: "3", Position: finding.Position{File: "a.go"}},
	}

	byFile := finding.GroupByFile(findings)
	fmt.Println("a.go:", len(byFile["a.go"]))
	fmt.Println("b.go:", len(byFile["b.go"]))

	// Output:
	// a.go: 2
	// b.go: 1
}

func ExampleMerge() {
	r1 := finding.NewReport(finding.ToolInfo{Name: "tool-a"})
	r1.AddFinding(finding.Finding{
		ID:       "govet:printf:main.go:10:3",
		Rule:     "printf",
		ToolName: "govet",
		Message:  "fmt.Printf format error",
		Severity: finding.SeverityWarning,
		Position: finding.Pos("main.go", 10, 3),
	})

	r2 := finding.NewReport(finding.ToolInfo{Name: "tool-b"})
	r2.AddFinding(finding.Finding{
		ID:       "staticcheck:SA1000:main.go:20:1",
		Rule:     "SA1000",
		ToolName: "staticcheck",
		Message:  "invalid regular expression",
		Severity: finding.SeverityError,
		Position: finding.Pos("main.go", 20, 1),
	})

	merged := finding.Merge([]*finding.Report{r1, r2})
	fmt.Println("Total:", merged.Summary.Total)

	// Output:
	// Total: 2
}

func ExampleMerge_deduplication() {
	duplicate := finding.Finding{ID: "same-id", Rule: "R1", Position: finding.Position{File: "a.go"}}

	r1 := finding.NewReport(finding.ToolInfo{Name: "tool-a"})
	r1.AddFinding(duplicate)

	r2 := finding.NewReport(finding.ToolInfo{Name: "tool-b"})
	r2.AddFinding(duplicate)

	merged := finding.Merge([]*finding.Report{r1, r2})
	fmt.Println("After dedup:", merged.Summary.Total)

	// Output:
	// After dedup: 1
}

func ExampleCorrelate() {
	findings := []finding.Finding{
		{
			ID: "govet:printf:main.go:10:3", Rule: "printf",
			ToolName: "govet", Message: "format error",
			Severity: finding.SeverityWarning,
			Position: finding.Pos("main.go", 10, 3),
		},
		{
			ID: "staticcheck:SA1000:main.go:12:1", Rule: "SA1000",
			ToolName: "staticcheck", Message: "invalid regex",
			Severity: finding.SeverityError,
			Position: finding.Pos("main.go", 12, 1),
		},
	}

	correlations := finding.Correlate(findings)
	fmt.Println("Correlations:", len(correlations))

	for _, c := range correlations {
		fmt.Printf("%.1f: %s\n", c.Confidence, c.Reason)
	}

	// Output:
	// Correlations: 1
	// 0.6: same file, nearby lines
}

func ExampleNewReport() {
	report := finding.NewReport(finding.ToolInfo{Name: "mytool", Version: "1.0.0"})
	report.AddFinding(finding.Finding{
		ID:          "mytool:RULE001:main.go:5:1",
		Rule:        "RULE001",
		ToolName:    "mytool",
		Message:     "unused variable",
		Severity:    finding.SeverityWarning,
		Category:    finding.CategoryCorrectness,
		FixStrategy: finding.FixStrategySuggest,
		Suggestion:  "Remove the unused variable",
		Position:    finding.Pos("main.go", 5, 1),
	})
	report.ComputeSummary()

	fmt.Println("Total:", report.Summary.Total)
	fmt.Println("Files:", report.Summary.FilesAffected)

	// Output:
	// Total: 1
	// Files: 1
}

func ExampleReport_ToSARIF() {
	report := finding.NewReport(finding.ToolInfo{Name: "mytool", Version: "1.0.0"})
	pos := finding.Pos("main.go", 1, 1)
	f := finding.Finding{
		Severity: finding.SeverityWarning,
		ID:       "mytool:R1:main.go:1:1",
		Rule:     "R1",
		ToolName: "mytool",
		Message:  "test finding",
		Position: pos,
	}
	report.AddFinding(f)

	data, err := report.ToSARIF()
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	var log struct {
		Version string `json:"version"`
	}

	_ = json.Unmarshal(data, &log)
	fmt.Println("SARIF version:", log.Version)

	// Output:
	// SARIF version: 2.1.0
}

func ExampleReport_ActiveFindings() {
	report := finding.NewReport(finding.ToolInfo{Name: "tool"})
	report.AddFinding(finding.Finding{
		ID: "1", Rule: "R1", Position: finding.Position{File: "a.go"},
	})
	report.AddFinding(finding.Finding{
		ID: "2", Rule: "R2", Position: finding.Position{File: "b.go"},
		Suppression: &finding.Suppression{Kind: finding.SuppressionInSource, Rule: "R2"},
	})

	active := report.ActiveFindings()
	fmt.Println("Active:", len(active))

	// Output:
	// Active: 1
}

func ExampleSeverity() {
	fmt.Println(finding.SeverityInfo)
	fmt.Println(finding.SeverityWarning)
	fmt.Println(finding.SeverityError)
	fmt.Println(finding.SeverityCritical)

	fmt.Println(finding.SeverityError.GreaterThan(finding.SeverityWarning))
	fmt.Println(finding.SeverityInfo.LessThan(finding.SeverityCritical))

	// Output:
	// info
	// warning
	// error
	// critical
	// true
	// true
}

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
	fmt.Println("Findings:", result.FinalFindingCount)

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
