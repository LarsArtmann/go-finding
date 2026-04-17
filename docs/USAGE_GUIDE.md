# go-finding Usage Guide

A Go library providing a unified data model and pipeline for static analysis tools.

## Installation

```bash
go get github.com/larsartmann/go-finding
```

## Quick Start

### As a Library

```go
package main

import (
    "fmt"

    "github.com/larsartmann/go-finding"
)

func main() {
    // Create a finding
    f := finding.Finding{
        ID:          finding.GenerateID("mytool", "RULE001", finding.Position{File: "main.go", Line: 42, Column: 5}),
        Rule:        "RULE001",
        ToolName:    "mytool",
        Message:     "unused variable",
        Severity:    finding.SeverityWarning,
        Category:    finding.CategoryCorrectness,
        FixStrategy: finding.FixStrategySuggest,
        Suggestion:  "Remove the unused variable",
        Position:    finding.Position{File: "main.go", Line: 42, Column: 5},
    }

    // Create a report
    report := finding.NewReport(finding.ToolInfo{Name: "mytool", Version: "1.0.0"})
    report.AddFinding(f)
    report.ComputeSummary()

    // Output as JSON
    data, _ := json.MarshalIndent(report, "", "  ")
    fmt.Println(string(data))

    // Output as SARIF 2.1.0
    sarif, _ := report.ToSARIF()
    fmt.Println(string(sarif))
}
```

### As a CLI Tool

```bash
# Install
go install github.com/larsartmann/go-finding/cmd/go-finding@latest

# Run analysis on current directory
go-finding -dir .

# Output as JSON
go-finding -format json

# Output as SARIF
go-finding -format sarif

# Filter by minimum severity
go-finding -severity error

# Use a config file
go-finding -config config.yaml

# Enable verification after fixes
go-finding -verify

# Profile performance
go-finding -cpuprof cpu.prof -memprof mem.prof
```

## CLI Flags

| Flag              | Default | Description                                              |
| ----------------- | ------- | -------------------------------------------------------- |
| `-dir`            | `.`     | Root directory to analyze                                |
| `-format`         | `text`  | Output format: `text`, `json`, `sarif`                   |
| `-severity`       | `info`  | Minimum severity: `info`, `warning`, `error`, `critical` |
| `-max-iterations` | `5`     | Maximum pipeline iterations                              |
| `-parallel`       | `true`  | Run detectors in parallel                                |
| `-verify`         | `false` | Verify fixes by re-running detectors                     |
| `-timeout`        | `10m`   | Pipeline timeout (Go duration format)                    |
| `-config`         | `""`    | Path to YAML or JSON config file                         |
| `-cpuprof`        | `""`    | Write CPU profile to file                                |
| `-memprof`        | `""`    | Write memory profile to file                             |

## Configuration File

Config files support both YAML and JSON (detected by extension):

```yaml
# config.yaml
maxIterations: 5
parallelDetectors: true
verifyAfterFix: false
timeout: "10m"

detectors:
  - name: govet
    args: {}
  - name: staticcheck
    args: {}
```

Equivalent JSON:

```json
{
  "maxIterations": 5,
  "parallelDetectors": true,
  "verifyAfterFix": false,
  "timeout": "10m",
  "detectors": [
    { "name": "govet", "args": {} },
    { "name": "staticcheck", "args": {} }
  ]
}
```

## Core Types

### Finding

The central type representing a single issue:

```go
type Finding struct {
    ID          string            // Stable unique identifier
    Rule        string            // Rule/check name
    ToolName    string            // Source tool name
    Message     string            // Human-readable description
    Severity    Severity          // info, warning, error, critical
    Position    Position          // Where the issue is
    Category    Category          // Domain classification
    FixStrategy FixStrategy       // none, suggest, direct, ai
    Suggestion  string            // Human-readable fix description
    BeforeCode  string            // Code before the fix
    AfterCode   string            // Code after the fix
    Range       *Range            // For span-based findings
    Related     []RelatedRef      // Related findings
    Suppression *Suppression      // If suppressed
    Metadata    map[string]string // Tool-specific key-value pairs
}
```

### Severity

```go
finding.SeverityInfo       // "info"
finding.SeverityWarning    // "warning"
finding.SeverityError      // "error"
finding.SeverityCritical   // "critical"
```

Severity supports comparison:

```go
finding.SeverityError.GreaterThan(finding.SeverityWarning) // true
finding.SeverityInfo.LessThan(finding.SeverityCritical)    // true
```

### Category

Standard categories:

| Constant                | Value              |
| ----------------------- | ------------------ |
| `CategorySecurity`      | `"security"`       |
| `CategoryStyle`         | `"style"`          |
| `CategoryPerformance`   | `"performance"`    |
| `CategoryCorrectness`   | `"correctness"`    |
| `CategoryComplexity`    | `"complexity"`     |
| `CategoryDuplication`   | `"duplication"`    |
| `CategoryErrorHandling` | `"error-handling"` |
| `CategoryMigration`     | `"migration"`      |
| `CategoryTypeSafety`    | `"type-safety"`    |
| `CategoryStructure`     | `"structure"`      |
| `CategoryConfiguration` | `"configuration"`  |
| `CategoryDocumentation` | `"documentation"`  |
| `CategoryTesting`       | `"testing"`        |
| `CategoryUnused`        | `"unused"`         |

Custom categories are valid — `Category` is a string type.

### FixStrategy

```go
finding.FixStrategyNone    // "none"    — no fix available
finding.FixStrategySuggest // "suggest" — human-readable suggestion
finding.FixStrategyDirect  // "direct"  — can be automatically applied
finding.FixStrategyAI      // "ai"      — requires AI assistance
```

### Position and Range

```go
// Position is 1-based; 0 means not set
pos := finding.Position{File: "main.go", Line: 42, Column: 5}

// Range represents a span
rng := finding.Range{
    Start: finding.Position{File: "main.go", Line: 40, Column: 1},
    End:   finding.Position{File: "main.go", Line: 45, Column: 10},
}

// Geometry operations
rng.Contains(pos)       // true if position is within range
rng.Overlaps(otherRng)  // true if ranges share positions
rng.Adjacent(otherRng)  // true if ranges are immediately adjacent
rng.Intersection(other) // overlapping region, or nil
```

## ID Generation

IDs are stable, unique identifiers in the format `tool:rule:file:line:col`:

```go
// Positional ID (human-readable)
id := finding.GenerateID("govet", "printf", finding.Position{File: "main.go", Line: 42, Column: 5})
// → "govet:printf:main.go:42:5"

// Hash-based ID (when line is 0)
id := finding.GenerateID("govet", "printf", finding.Position{File: "main.go"})
// → "govet:printf:<16-char-hex>"

// Parse an ID back
tool, rule, file, line, col, ok := finding.ParseID("govet:printf:main.go:42:5")

// Check if an ID is hash-based
finding.IsHashID("govet:printf:a1b2c3d4e5f6a1b2") // true
```

## Filtering and Grouping

### Composable Filters

```go
// Filter by severity threshold
warnings := finding.Filter(findings, finding.BySeverityAtLeast(finding.SeverityWarning))

// Filter by exact severity
errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))

// Filter by category
security := finding.Filter(findings, finding.ByCategory(finding.CategorySecurity))

// Filter by tool
govet := finding.Filter(findings, finding.ByTool("govet"))

// Filter by rule
rule := finding.Filter(findings, finding.ByRule("printf"))

// Filter by file
file := finding.Filter(findings, finding.ByFile("main.go"))

// Filter by fix strategy
fixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))

// Combine predicates
result := finding.Filter(findings,
    finding.BySeverityAtLeast(finding.SeverityError),
    finding.ByCategory(finding.CategorySecurity),
)
```

### Grouping

```go
// Group by file
byFile := finding.GroupByFile(findings) // map[string][]Finding

// Group by severity
bySev := finding.GroupBySeverity(findings) // map[Severity][]Finding

// Group by category
byCat := finding.GroupByCategory(findings) // map[Category][]Finding

// Custom grouping
byRule := finding.GroupBy(findings, func(f finding.Finding) string {
    return f.Rule
})
```

## Merging and Deduplication

```go
// Merge multiple reports with deduplication by ID (default)
merged := finding.Merge([]*finding.Report{report1, report2})

// Merge without deduplication
merged = finding.Merge(reports, finding.WithDeduplication(false))

// Deduplicate by position instead of ID
merged = finding.Merge(reports, finding.WithDeduplicateBy(finding.DeduplicateByPosition))

// Deduplicate by rule+position
merged = finding.Merge(reports, finding.WithDeduplicateBy(finding.DeduplicateByRule))
```

## Correlation

Find related findings across different tools:

```go
correlations := finding.Correlate(allFindings)
for _, c := range correlations {
    fmt.Printf("Related: %v (reason: %s, confidence: %.2f)\n",
        c.FindingIDs, c.Reason, c.Confidence)
}
```

Correlation uses heuristics: same file + nearby lines (within 5 lines) from different tools.

## Reports

```go
// Create
report := finding.NewReport(finding.ToolInfo{Name: "mytool", Version: "1.0.0"})

// Add findings
report.AddFinding(f)
report.AddFindings(multipleFindings)

// Compute summary statistics
report.ComputeSummary()

// Access summary
fmt.Println(report.Summary.Total)
fmt.Println(report.Summary.BySeverity)
fmt.Println(report.Summary.FilesAffected)

// Convenience methods
active := report.ActiveFindings()
errors := report.BySeverity(finding.SeverityError)
security := report.ByCategory(finding.CategorySecurity)
fixable := report.ByFixStrategy(finding.FixStrategyDirect)
found := report.FindByID("govet:printf:main.go:42:5")
```

## SARIF Output

```go
// Full SARIF 2.1.0 output
sarif, err := report.ToSARIF()

// Filtered by minimum severity
sarif, err = report.ToSARIFFiltered(finding.SeverityWarning)
```

## Pipeline

The pipeline orchestrates the detect → triage → fix → verify loop.

### Writing a Detector

Implement the `Detector` interface:

```go
type Detector interface {
    Name() string
    Detect(ctx context.Context) ([]finding.Finding, error)
}
```

Using a function:

```go
detector := pipeline.NamedDetectorFunc("my-tool", func(ctx context.Context) ([]finding.Finding, error) {
    // Run your analysis...
    return findings, nil
})
```

Full struct implementation:

```go
type MyDetector struct {
    dir string
}

func (d *MyDetector) Name() string { return "my-detector" }

func (d *MyDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
    // Analyze code in d.dir
    return findings, nil
}
```

### Running the Pipeline

```go
cfg := pipeline.Config{
    MaxIterations:     5,
    ParallelDetectors: true,
    VerifyAfterFix:    false,
    Timeout:           10 * time.Minute,
    Metrics:           pipeline.NewMetrics(),
}

p := pipeline.New(cfg, ".", detector1, detector2)
result, err := p.Run(ctx)

// Inspect results
fmt.Println("Stable:", result.Stable)
fmt.Println("Iterations:", result.TotalIterations)
fmt.Println("Findings:", result.FinalFindingCount)

for _, iter := range result.Iterations {
    fmt.Printf("Iteration %d: %d findings, %d fixes applied\n",
        iter.Number, iter.FindingsFound, iter.Applied)
    for _, f := range iter.Findings() {
        fmt.Printf("  - %s: %s\n", f.Position.String(), f.Message)
    }
}
```

### Pipeline Configuration

```go
cfg := pipeline.Config{
    MaxIterations:     5,          // Prevent infinite loops
    ParallelDetectors: true,       // Run detectors concurrently
    VerifyAfterFix:    true,       // Re-run detectors after fixes
    Timeout:           10 * time.Minute,
    GracefulDegradation: true,     // Continue on detector failures
    Metrics:           pipeline.NewMetrics(),

    // Retry flaky detectors
    Retry: &pipeline.RetryConfig{
        MaxRetries: 3,
        BaseDelay:  100 * time.Millisecond,
        MaxDelay:   5 * time.Second,
    },

    // Callbacks
    OnFinding: func(f finding.Finding) {
        log.Printf("Found: %s", f.ID)
    },
    OnFix: func(f finding.Finding, applied bool) {
        if applied {
            log.Printf("Fixed: %s", f.ID)
        }
    },
    OnIteration: func(iter int, findings []finding.Finding) {
        log.Printf("Iteration %d: %d findings", iter, len(findings))
    },
}
```

### Retry

```go
retryCfg := pipeline.DefaultRetryConfig() // 3 retries, 100ms base, 5s max
wrapped := pipeline.NewRetryDetector(myDetector, retryCfg)
```

### Partial Success

```go
cfg := pipeline.Config{
    GracefulDegradation: true,
    // ...
}

// Access partial results
result, err := p.DetectPartial(ctx)
if result.HasErrors() {
    for name, detErr := range result.Errors {
        log.Printf("Detector %s failed: %v", name, detErr)
    }
}
// result.Findings contains results from successful detectors
```

### Metrics

```go
m := pipeline.NewMetrics()
cfg := pipeline.Config{Metrics: m, /* ... */}

// After running
snapshot := m.Snapshot()
fmt.Println("Duration:", snapshot.TotalDuration)
fmt.Println("Findings:", snapshot.TotalFindings)
fmt.Println("Fixes:", snapshot.TotalFixes)
```

## Profiling

```bash
# Generate CPU profile
go-finding -cpuprof cpu.prof -dir .

# Generate memory profile
go-finding -memprof mem.prof -dir .

# Analyze
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Error Handling

Structured errors with categories:

```go
// Create typed errors
err := finding.NewValidationError("invalid severity", nil)
err = finding.NewIOError("read file", ioErr)
err = finding.NewParseError("parse config", parseErr)
err = finding.NewConflictError("overlapping fixes", nil)
err = finding.NewInternalError("unexpected state", nil)

// Add context
err = err.WithPosition(finding.Position{File: "main.go", Line: 42})
err = err.WithFinding(f)

// Inspect
finding.IsFindingError(err)          // true
finding.GetCategory(err)             // "validation"
finding.IsCategory(err, finding.ErrCategoryIO) // false

// Type assertion
var fe *finding.FindingError
if errors.As(err, &fe) {
    fmt.Println(fe.Category, fe.Message)
}
```

## Example Detectors

See the `examples/` directory for complete detector implementations:

- **`examples/govet/`** — Wraps `go vet -json` output
- **`examples/staticcheck/`** — Wraps staticcheck JSON output
- **`examples/artdupl/`** — Wraps art-dupl clone detection
- **`examples/branching/`** — Code complexity analysis with refactoring suggestions
