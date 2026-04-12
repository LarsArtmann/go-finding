# go-finding

A Go library for unified static analysis findings. Provides a common data model and pipeline for tools that detect code issues.

## Purpose

Seven tools detect issues. Zero tools **route them to remediation**.

This SDK solves:

- Each tool invents its own types for findings
- No standardized way to apply fixes
- Manual loop: run tool → read output → fix → re-run

## Features

- **Unified Finding type** - Common model for all static analysis tools
- **Severity levels** - info, warning, error, critical
- **Fix strategies** - none, suggest, direct (deterministic), ai
- **Position tracking** - File, line, column with range support
- **SARIF 2.1.0 output** - Standard interchange format
- **LSP integration** - Diagnostic conversion for IDE support
- **go/analysis compatibility** - Convert to/from standard Go analyzer types
- **Report merging** - Combine findings from multiple tools
- **Filtering & grouping** - Query findings flexibly

## Installation

```bash
go get github.com/larsartmann/go-finding
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/larsartmann/go-finding"
)

func main() {
    // Create a finding
    f := finding.Finding{
        ID:       finding.GenerateID("my-tool", "unused-var", finding.Position{File: "main.go", Line: 42}),
        Rule:     "unused-var",
        ToolName: "my-tool",
        Message:  "variable x is unused",
        Severity: finding.SeverityWarning,
        Position: finding.Position{File: "main.go", Line: 42, Column: 5},
    }

    // Create a report
    report := finding.NewReport(finding.ToolInfo{Name: "my-tool", Version: "1.0.0"})
    report.AddFinding(f)
    report.ComputeSummary()

    // Output as SARIF
    sarif, _ := report.ToSARIF()
    fmt.Println(string(sarif))
}
```

## Core Types

### Finding

```go
type Finding struct {
    ID       string   // "tool:rule:file:line:col"
    Rule     string   // "STRONG_ID", "clone-detected"
    ToolName string   // "branching-flow", "art-dupl"
    Message  string   // Human-readable description
    Severity Severity // info, warning, error, critical
    Position Position // Where the issue is

    // Fix information
    FixStrategy FixStrategy // none, suggest, direct, ai
    Suggestion  string      // Human-readable fix
    BeforeCode  string      // Code before fix
    AfterCode   string      // Code after fix

    // Context
    Category   string       // "security", "style", "duplication"
    Confidence float64      // 0.0-1.0
    Related    []RelatedRef // Linked findings
}
```

### Report

```go
type Report struct {
    Tool     ToolInfo  // Name, version
    Findings []Finding // All findings
    Summary  Summary   // Aggregated stats
}
```

## Converting from Existing Types

### From go/analysis.Diagnostic

```go
import "golang.org/x/tools/go/analysis"

diag := &analysis.Diagnostic{...}
finding := finding.FromDiagnostic(diag, pass.Fset, "my-analyzer")
```

### To SARIF

```go
report := finding.NewReport(finding.ToolInfo{Name: "my-tool"})
// ... add findings ...

sarifJSON, err := report.ToSARIF()
```

### To LSP Diagnostic

```go
lspDiag := finding.ToLSP()
```

## Filtering

```go
// By severity
errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))

// By fix strategy
autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))

// Chained filters
important := finding.Filter(findings,
    finding.BySeverityAtLeast(finding.SeverityWarning),
    finding.NotSuppressed,
    finding.HasFix,
)

// Group by file
byFile := finding.GroupByFile(findings)
```

## Merging Reports

```go
// From multiple tools
merged := finding.Merge([]*Report{report1, report2, report3},
    finding.WithDeduplication(true),
)
```

## Tools Using This SDK

- **art-dupl** - Code duplication detection
- **branching-flow** - Go code quality analyzer
- **hierarchical-errors** - Error handling pattern detector
- **go-auto-upgrade** - Dependency upgrade automation
- **golangci-lint-auto-configure** - Linter configuration
- **BuildFlow** - Build pipeline validation

## Related Projects

- [go-business-rules](https://github.com/larsartmann/go-business-rules) - Runtime validation (uses same Severity type)
- [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) - Static Analysis Results Interchange Format

## License

MIT
