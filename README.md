# go-finding

A Go library providing a unified data model and pipeline for static analysis tools.

[![CI](https://github.com/larsartmann/go-finding/actions/workflows/ci.yml/badge.svg)](https://github.com/larsartmann/go-finding/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/go-finding.svg)](https://pkg.go.dev/github.com/larsartmann/go-finding)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/go-finding)](https://goreportcard.com/report/github.com/larsartmann/go-finding)
[![codecov](https://codecov.io/gh/larsartmann/go-finding/branch/master/graph/badge.svg)](https://codecov.io/gh/larsartmann/go-finding)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why

Seven tools detect issues. Zero tools **route them to remediation**.

Each tool invents its own types for findings. There is no standardized way to apply fixes. The manual loop — run tool, read output, fix, re-run — is slow and error-prone.

**go-finding** solves this with:

- **Unified Finding type** — Common model for all static analysis tools
- **Pipeline** — Automated detect → triage → fix → verify loop
- **SARIF 2.1.0** — Standard interchange format for CI/CD integration
- **LSP diagnostics** — IDE integration out of the box

## Installation

```bash
go get github.com/larsartmann/go-finding
```

Requires Go 1.26 or later.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/larsartmann/go-finding"
)

func main() {
    f := finding.NewFinding(
        "unused-var", "my-tool",
        "variable x is unused",
        finding.SeverityWarning,
        finding.Position{File: "main.go", Line: 42, Column: 5},
    )

    report := finding.NewReport(finding.ToolInfo{Name: "my-tool", Version: "1.0.0"})
    report.AddFinding(f)
    report.ComputeSummary()

    sarif, _ := report.ToSARIF()
    fmt.Println(string(sarif))
}
```

## Builder API

Construct findings fluently with the `Builder`:

```go
f := finding.NewBuilder("nilcheck", "govet", "possible nil deref",
    finding.SeverityError, finding.Pos("main.go", 42, 5)).
    WithFixStrategy(finding.FixStrategyDirect).
    WithBeforeCode("x.foo").
    WithAfterCode("x.foo()").
    WithConfidence(0.95).
    Build()
```

## Core Types

| Type          | Purpose                                                    |
| ------------- | ---------------------------------------------------------- |
| `Finding`     | A single issue: ID, rule, severity, position, fix strategy |
| `Report`      | Thread-safe container for findings with summary statistics |
| `Severity`    | `info` / `warning` / `error` / `critical`                  |
| `FixStrategy` | `none` / `suggest` / `direct` / `ai`                       |
| `Position`    | File, line, column location                                |
| `Range`       | Start and end positions with geometric operations          |
| `Category`    | `security`, `style`, `performance`, `correctness`, etc.    |

## Filtering

```go
errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))

autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))

important := finding.Filter(findings,
    finding.BySeverityAtLeast(finding.SeverityWarning),
    finding.NotSuppressed,
    finding.HasFix,
)

byFile := finding.GroupByFile(findings)
bySeverity := finding.GroupBySeverity(findings)
```

## Merging

Combine reports from multiple tools with deduplication:

```go
merged := finding.Merge([]*Report{govet, staticcheck, custom},
    finding.WithDeduplication(true),
)
```

Cross-tool correlation finds related findings:

```go
correlations := finding.Correlate(allFindings)
for _, c := range correlations {
    fmt.Printf("%.1f: %s\n", c.Confidence, c.Reason)
}
```

## Pipeline

The `pipeline` package provides a detect → triage → fix → verify loop:

```go
detector := pipeline.NamedDetectorFunc("my-tool", func(ctx context.Context) ([]finding.Finding, error) {
    return []finding.Finding{...}, nil
})

cfg := pipeline.Config{
    MaxIterations:     5,
    ParallelDetectors: true,
    Timeout:           10 * time.Minute,
    VerifyAfterFix:    true,
    GracefulDegradation: true,
    DryRun:            false,
}

p := pipeline.New(cfg, ".", detector)
result, err := p.Run(context.Background())

fmt.Printf("Iterations: %d, Findings: %d, Stable: %v\n",
    result.TotalIterations, result.TotalDetected, result.Stable)
```

### Pipeline Features

| Feature                | Description                                         |
| ---------------------- | --------------------------------------------------- |
| **Parallel detection** | errgroup-based concurrent detector execution        |
| **Conflict detection** | Overlapping fixes filtered before application       |
| **Fix application**    | AST-aware with text fallback, backup/rollback       |
| **Verification**       | Re-run detectors to confirm fixes                   |
| **Retry**              | Exponential backoff for flaky detectors             |
| **Partial success**    | Continue with findings from successful detectors    |
| **Metrics**            | Optional timing and count collection with snapshots |
| **Dry run**            | Detect + triage without applying fixes              |

### Custom Detector

```go
type MyDetector struct{}

func (d *MyDetector) Name() string { return "my-detector" }

func (d *MyDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
    findings := []finding.Finding{
        finding.NewFinding("RULE001", "my-detector", "issue found",
            finding.SeverityError,
            finding.Position{File: "main.go", Line: 10}),
    }
    return findings, nil
}
```

## SARIF

```go
// Export
sarifJSON, err := report.ToSARIF()

// Parse SARIF from another tool
findings, err := finding.FindingsFromSARIF(sarifJSON)
```

Round-trip fidelity is preserved. `SeverityCritical` maps to SARIF `"error"` (SARIF 2.1.0 has no critical level); the original severity is stored in `Properties["go-finding/severity"]`.

## LSP Diagnostics

```go
lspDiag := f.ToLSP()

// From LSP diagnostic
f := finding.FromLSP(lspDiag, "my-tool")
```

## go/analysis Integration

```go
// From go/analysis Diagnostic
f := finding.FromDiagnostic(diag, pass.Fset, "my-analyzer")

// Convert back to analysis.Diagnostic
ad := finding.AnalysisDiagnostic(f)
```

## JSON

```go
// Serialize
data, err := f.ToJSON()

// Deserialize with validation
f, err := finding.FromJSON(data)

// Line-delimited JSON stream
data, dropped := finding.LineJSON(findings)

// Pretty-printed report
data, err := report.PrettyJSON()
```

## Error Handling

Structured errors with categories:

```go
err := finding.NewValidationError("invalid severity", nil)
err := finding.NewIOError("read file", cause).WithPosition(pos)
err := finding.NewConflictError("overlapping fixes", cause)

finding.IsFindingError(err)
finding.GetCategory(err) // "validation", "io", "conflict", etc.
```

## Tools Using This SDK

- [art-dupl](https://github.com/larsartmann/art-dupl) — Code duplication detection
- [branching-flow](https://github.com/larsartmann/branching-flow) — Go code quality analyzer
- [hierarchical-errors](https://github.com/larsartmann/hierarchical-errors) — Error handling pattern detector
- [go-auto-upgrade](https://github.com/larsartmann/go-auto-upgrade) — Dependency upgrade automation
- [golangci-lint-auto-configure](https://github.com/larsartmann/golangci-lint-auto-configure) — Linter configuration

## CLI

```bash
go install github.com/larsartmann/go-finding/cmd/go-finding@latest

go-finding run --format sarif --output results.sarif
go-finding run --format json --config config.yaml
```

## Development

```bash
just test       # Run tests with -race
just bench      # Run benchmarks
just lint       # golangci-lint
just cover      # Coverage report
just check      # All checks (fmt + lint + test)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Project Stats

| Package   | Coverage  |
| --------- | --------- |
| Root      | 98.5%     |
| Pipeline  | 94.8%     |
| Detectors | 96.1%     |
| CLI       | 74.2%     |
| **Total** | **94.8%** |

## Related Projects

- [go-business-rules](https://github.com/larsartmann/go-business-rules) — Runtime validation (uses same Severity type)
- [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) — Static Analysis Results Interchange Format

## License

[MIT](LICENSE)
