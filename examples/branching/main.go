package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/larsartmann/go-finding"
)

type branchingIssue struct {
	file    string
	line    int
	message string
	rule    string
	before  string
	after   string
}

type BranchingDetector struct {
	dir string
}

func NewBranchingDetector(dir string) *BranchingDetector {
	return &BranchingDetector{dir: dir}
}

func (d *BranchingDetector) Name() string {
	return "branching"
}

func (d *BranchingDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	issues := detectBranchingIssues(d.dir)
	findings := make([]finding.Finding, 0, len(issues))
	for _, issue := range issues {
		f := finding.Finding{
			ID:          fmt.Sprintf("branching:%s:%s:%d", issue.rule, issue.file, issue.line),
			Rule:        issue.rule,
			ToolName:    "branching",
			Message:     issue.message,
			Severity:    finding.SeverityWarning,
			Category:    finding.CategoryComplexity,
			Position:    finding.Position{File: issue.file, Line: issue.line},
			FixStrategy: finding.FixStrategyDirect,
			Confidence:  0.85,
			BeforeCode:  issue.before,
			AfterCode:   issue.after,
			Suggestion:  fmt.Sprintf("Apply the suggested refactoring to simplify the branching logic."),
		}
		findings = append(findings, f)
	}
	return findings, nil
}

func detectBranchingIssues(dir string) []branchingIssue {
	var issues []branchingIssue

	// Example: detect if-else chains that could be switch statements
	issues = append(issues, branchingIssue{
		file:    filepath.Join(dir, "example.go"),
		line:    1,
		message: "if-else chain could be replaced with a switch statement",
		rule:    "if-else-to-switch",
		before: `if x == 1 {
	foo()
} else if x == 2 {
	bar()
} else if x == 3 {
	baz()
} else {
	qux()
}`,
		after: `switch x {
case 1:
	foo()
case 2:
	bar()
case 3:
	baz()
default:
	qux()
}`,
	})

	// Example: detect nested conditionals that could use early returns
	issues = append(issues, branchingIssue{
		file:    filepath.Join(dir, "example.go"),
		line:    20,
		message: "nested conditional can be simplified with early return",
		rule:    "nested-early-return",
		before: `func process(data *Data) error {
	if data != nil {
		if data.IsValid() {
			if data.Size > 0 {
				return doProcess(data)
			}
			return fmt.Errorf("empty data")
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("nil data")
}`,
		after: `func process(data *Data) error {
	if data == nil {
		return fmt.Errorf("nil data")
	}
	if !data.IsValid() {
		return fmt.Errorf("invalid data")
	}
	if data.Size <= 0 {
		return fmt.Errorf("empty data")
	}
	return doProcess(data)
}`,
	})

	return issues
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolve path: %v", err)
	}

	detector := NewBranchingDetector(absDir)
	findings, err := detector.Detect(context.Background())
	if err != nil {
		log.Fatalf("detect: %v", err)
	}

	if len(findings) == 0 {
		fmt.Println("No branching issues found.")
		return
	}

	fmt.Printf("Found %d branching issues:\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  %s:%d: %s [%s]\n", f.Position.File, f.Position.Line, f.Message, f.Rule)
		fmt.Printf("    Fix strategy: %s\n", f.FixStrategy)
		if f.HasFix() {
			fmt.Printf("    Automated fix available (before/after code provided)\n")
		}
	}
}
