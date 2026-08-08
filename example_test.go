package finding_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/larsartmann/go-finding"
)

func ExampleNewFinding() {
	pos := finding.Pos("main.go", 42, 5)
	f := finding.NewFinding(
		"nilcheck", "govet", "possible nil dereference",
		finding.SeverityError, pos, 0,
	)
	fmt.Println(f.ID)
	fmt.Println(f.Rule)
	fmt.Println(f.Severity)
	fmt.Println(f.Position)

	// Output:
	// govet:nilcheck:main.go:42:5
	// nilcheck
	// error
	// main.go:42:5
}

// ExampleBuilder demonstrates the fluent Finding builder API. This block
// intentionally mirrors examples/builder/main.go: the testable example feeds
// the godoc with a verifiable `// Output:` snapshot, while the demo program
// is a standalone binary for `go run`. Sharing the snippet between them
// would require an indirection that obscures both forms.
func ExampleBuilder() {
	builder := finding.NewBuilder(
		"staticcheck", "SA1000", "invalid regex",
		finding.SeverityError, finding.Pos("pkg/validate.go", 24, 8),
	)
	builder.WithCategory(finding.CategoryCorrectness)
	builder.WithConfidence(0.95)
	builder.WithBeforeCode("oldPattern")
	builder.WithAfterCode("newPattern")
	builder.WithFixStrategy(finding.FixStrategyDirect)

	f, err := builder.Build()
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(f.Rule)
	fmt.Println(f.ToolName)
	fmt.Println(f.Category)
	fmt.Println(f.HasFix())

	// Output:
	// staticcheck
	// SA1000
	// correctness
	// true
}

func ExampleGenerateID() {
	pos := finding.Position{File: "main.go", Line: 42, Column: 5}
	id := finding.GenerateID("govet", "printf", pos)
	fmt.Println(id)

	// Hash-based ID when line is 0
	posNoLine := finding.Position{File: "main.go"}
	hashID := finding.GenerateID("govet", "printf", posNoLine)
	fmt.Println(finding.IsHashID(string(hashID)))

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

	combined := finding.Filter(
		findings,
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

func ExampleCombine() {
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

	merged := finding.Combine([]*finding.Report{r1, r2})
	fmt.Println("Total:", merged.Summary.Total)

	// Output:
	// Total: 2
}

func ExampleCombine_deduplication() {
	duplicate := finding.Finding{
		ID:       "same-id",
		Rule:     "R1",
		Position: finding.Position{File: "a.go"},
	}

	r1 := finding.NewReport(finding.ToolInfo{Name: "tool-a"})
	r1.AddFinding(duplicate)

	r2 := finding.NewReport(finding.ToolInfo{Name: "tool-b"})
	r2.AddFinding(duplicate)

	merged := finding.Combine([]*finding.Report{r1, r2})
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
		fmt.Printf("%.1f: %s\n", c.Score, c.Reason)
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

func ExampleFindingError() {
	err := finding.NewValidationError("invalid input", nil)
	fmt.Println(finding.IsFindingError(err))
	fmt.Println(finding.CategoryOf(err))

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
	return finding.NewFinding(
		finding.RuleName(rule),
		finding.ToolName(tool),
		msg,
		sev,
		finding.Pos(finding.FilePath(file), line, col),
		0,
	)
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
					finding.Position{File: finding.FilePath(d.File), Line: d.Line},
				),
				Rule:     finding.RuleName(d.Rule),
				ToolName: "mylint",
				Message:  d.Message,
				Severity: finding.SeverityError,
				Category: finding.CategoryForLinter(d.Rule),
				Position: finding.Position{File: finding.FilePath(d.File), Line: d.Line},
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

func Example_basic() {
	// Create a finding
	f := finding.Finding{
		ID: finding.GenerateID(
			"my-linter",
			"unused-import",
			finding.Position{File: "main.go", Line: 5},
		),
		Rule:        "unused-import",
		ToolName:    "my-linter",
		Message:     "import \"fmt\" is unused",
		Severity:    finding.SeverityWarning,
		Position:    finding.Pos("main.go", 5, 2),
		Category:    finding.CategoryStyle,
		FixStrategy: finding.FixStrategyDirect,
		BeforeCode:  `import "fmt"`,
		AfterCode:   "",
	}

	// Create a report
	report := finding.NewReport(finding.ToolInfo{Name: "my-linter", Version: "1.0.0"})
	report.AddFinding(f)
	report.ComputeSummary()

	// Print summary
	fmt.Printf("Tool: %s\n", report.Tool.Name)
	fmt.Printf("Total findings: %d\n", report.Summary.Total)
	fmt.Printf("Warnings: %d\n", report.Summary.BySeverity[finding.SeverityWarning])

	// Output:
	// Tool: my-linter
	// Total findings: 1
	// Warnings: 1
}

func Example_filter() {
	findings := []finding.Finding{
		{ID: "1", Severity: finding.SeverityError, Rule: "nil-pointer", ToolName: "analyzer"},
		{ID: "2", Severity: finding.SeverityWarning, Rule: "unused-var", ToolName: "analyzer"},
		{ID: "3", Severity: finding.SeverityInfo, Rule: "comment-style", ToolName: "analyzer"},
	}

	// Filter for errors only
	errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))
	fmt.Printf("Errors: %d\n", len(errors))

	// Filter for severity >= warning
	warningsAndErrors := finding.Filter(
		findings,
		finding.BySeverityAtLeast(finding.SeverityWarning),
	)
	fmt.Printf("Warnings and Errors: %d\n", len(warningsAndErrors))

	// Output:
	// Errors: 1
	// Warnings and Errors: 2
}

func Example_merge() {
	// Reports from different tools
	r1 := finding.NewReport(finding.ToolInfo{Name: "linter-a"})
	r1.AddFinding(finding.Finding{
		ID:       "a:rule1:file.go:10:5",
		Severity: finding.SeverityError,
		Position: finding.Position{File: "file.go", Line: 10},
	})

	r2 := finding.NewReport(finding.ToolInfo{Name: "linter-b"})
	r2.AddFinding(finding.Finding{
		ID:       "b:rule2:file.go:20:3",
		Severity: finding.SeverityWarning,
		Position: finding.Position{File: "file.go", Line: 20},
	})

	// Merge reports
	merged := finding.Combine([]*finding.Report{r1, r2})
	merged.ComputeSummary()

	fmt.Printf("Total: %d\n", merged.Summary.Total)
	fmt.Printf("Files: %d\n", merged.Summary.FilesAffected)

	// Output:
	// Total: 2
	// Files: 1
}

// Example_simpleCLI demonstrates a simple CLI tool using the finding library.
func Example_simpleCLI() {
	// Simulate findings from a tool
	findings := []finding.Finding{
		{
			ID:          "linter:unused-import:main.go:3:2",
			Rule:        "unused-import",
			ToolName:    "my-linter",
			Message:     "import \"fmt\" is unused",
			Severity:    finding.SeverityWarning,
			Position:    finding.Pos("main.go", 3, 2),
			Category:    finding.CategoryStyle,
			FixStrategy: finding.FixStrategyDirect,
			BeforeCode:  `import "fmt"`,
			AfterCode:   "",
		},
		{
			ID:          "linter:unused-var:main.go:10:5",
			Rule:        "unused-var",
			ToolName:    "my-linter",
			Message:     "variable x is unused",
			Severity:    finding.SeverityWarning,
			Position:    finding.Pos("main.go", 10, 5),
			Category:    finding.CategoryStyle,
			FixStrategy: finding.FixStrategySuggest,
			Suggestion:  "Remove the variable or use it",
		},
		{
			ID:          "linter:nil-pointer:auth.go:45:12",
			Rule:        "nil-pointer",
			ToolName:    "my-linter",
			Message:     "potential nil pointer dereference",
			Severity:    finding.SeverityError,
			Position:    finding.Position{File: "auth.go", Line: 45, Column: 12},
			Category:    finding.CategorySecurity,
			FixStrategy: finding.FixStrategyNone,
		},
	}

	// Create report
	report := finding.NewReport(finding.ToolInfo{Name: "my-linter", Version: "1.0.0"})
	report.AddFindings(findings)
	report.ComputeSummary()

	// Filter for actionable items
	autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))
	suggestions := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategySuggest))

	// Output summary
	fmt.Printf("=== Analysis Summary ===\n")
	fmt.Printf("Total findings: %d\n", report.Summary.Total)
	fmt.Printf("Auto-fixable: %d\n", len(autoFixable))
	fmt.Printf("Need manual review: %d\n", len(suggestions))
	fmt.Printf("Errors: %d\n", report.Summary.BySeverity[finding.SeverityError])
	fmt.Printf("Warnings: %d\n", report.Summary.BySeverity[finding.SeverityWarning])

	// Output SARIF for CI integration
	sarif, err := report.ToSARIF()
	if err != nil {
		log.Fatal(err)
	}

	_ = sarif // In real tool, write to file

	// Output:
	// === Analysis Summary ===
	// Total findings: 3
	// Auto-fixable: 1
	// Need manual review: 1
	// Errors: 1
	// Warnings: 2
}

// Example_mergingShows unified report from multiple tools.
func Example_merging() {
	// Tool A: Linter
	toolA := finding.NewReport(finding.ToolInfo{Name: "linter", Version: "1.0"})
	toolA.AddFinding(finding.Finding{
		ID:       "linter:unused:main.go:10",
		Rule:     "unused",
		ToolName: "linter",
		Message:  "unused variable",
		Severity: finding.SeverityWarning,
		Position: finding.Position{File: "main.go", Line: 10},
	})

	// Tool B: Security Scanner
	toolB := finding.NewReport(finding.ToolInfo{Name: "security", Version: "2.0"})
	toolB.AddFinding(finding.Finding{
		ID:       "security:sql-inject:db.go:45",
		Rule:     "sql-inject",
		ToolName: "security",
		Message:  "SQL injection vulnerability",
		Severity: finding.SeverityCritical,
		Position: finding.Position{File: "db.go", Line: 45},
	})

	// Merge into unified report
	merged := finding.Combine([]*finding.Report{toolA, toolB})
	merged.ComputeSummary()

	fmt.Printf("Unified Report:\n")
	fmt.Printf("Total: %d findings from %d tools\n", merged.Summary.Total, 2)
	fmt.Printf("By severity: critical=%d, warning=%d\n",
		merged.Summary.BySeverity[finding.SeverityCritical],
		merged.Summary.BySeverity[finding.SeverityWarning])

	// Output:
	// Unified Report:
	// Total: 2 findings from 2 tools
	// By severity: critical=1, warning=1
}

func ExampleParseConfidence() {
	c, err := finding.ParseConfidence("high")
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(c)

	// Empty string defaults to low
	low, _ := finding.ParseConfidence("")
	fmt.Println(low)

	// Decimals work too
	custom, _ := finding.ParseConfidence("0.42")
	fmt.Println(custom)

	// Output:
	// high
	// low
	// 0.42
}

func ExampleTemplate_Builder() {
	tmpl := finding.NewTemplate("my-linter").
		WithCategory(finding.CategoryStyle).
		WithFixStrategy(finding.FixStrategySuggest)

	f := tmpl.Builder("R1", "bad pattern", finding.SeverityWarning, finding.Pos("demo.go", 42, 3)).
		WithConfidence(finding.ConfidenceHigh).
		WithSuggestion("use humanize.Bytes instead").
		MustBuild()

	fmt.Println(f.Rule)
	fmt.Println(f.ToolName)
	fmt.Println(f.Category)
	fmt.Println(f.Confidence)
	fmt.Println(f.Suggestion)

	// Output:
	// R1
	// my-linter
	// style
	// high
	// use humanize.Bytes instead
}
