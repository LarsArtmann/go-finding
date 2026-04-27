package finding_test

import (
	"fmt"

	"github.com/larsartmann/go-finding"
)

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
	merged := finding.Merge([]*finding.Report{r1, r2})
	merged.ComputeSummary()

	fmt.Printf("Total: %d\n", merged.Summary.Total)
	fmt.Printf("Files: %d\n", merged.Summary.FilesAffected)

	// Output:
	// Total: 2
	// Files: 1
}
