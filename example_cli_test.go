package finding_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/larsartmann/go-finding"
)

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
			Position:    finding.Position{File: "main.go", Line: 3, Column: 2},
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
			Position:    finding.Position{File: "main.go", Line: 10, Column: 5},
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

// Example_resultHandling shows using Result[T] for error handling.
func Example_resultHandling() {
	// Function that might fail
	parseFinding := func(data []byte) finding.Result[finding.Finding] {
		var f finding.Finding

		err := json.Unmarshal(data, &f)
		if err != nil {
			return finding.Err[finding.Finding](err)
		}

		if !f.IsValid() {
			return finding.Err[finding.Finding](errors.New("invalid finding"))
		}

		return finding.Ok(f)
	}

	// Valid JSON
	validJSON := `{"id":"test","rule":"R","toolName":"T","message":"M","severity":"warning","position":{"file":"f.go"}}`

	result := parseFinding([]byte(validJSON))
	if result.IsOk() {
		finding := result.Value()
		fmt.Printf("Parsed: %s\n", finding.ID)
	} else {
		fmt.Printf("Error: %v\n", result.Error())
	}

	// Invalid JSON
	invalidJSON := `{"invalid": true}`

	result2 := parseFinding([]byte(invalidJSON))
	finding2 := result2.ValueOr(finding.Finding{ID: "default"})
	fmt.Printf("Using default: %s\n", finding2.ID)

	// Output:
	// Parsed: test
	// Using default: default
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
	merged := finding.Merge([]*finding.Report{toolA, toolB})
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
