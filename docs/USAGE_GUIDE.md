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
    "encoding/json"
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
go-finding -min-severity error

# Use a config file
go-finding -config config.yaml

# Enable verification after fixes
go-finding -verify

# Profile performance
go-finding -cpuprof cpu.prof -memprof mem.prof

# Filter out findings from auto-generated files
# (sqlc, protobuf, mockgen, templ, wire, stringer, etc.)
go-finding -filter-generated

# Filter only sqlc and protobuf generated files
go-finding -filter-generated -filter-generated-types sqlc,protobuf

# Filter all generated files, but exclude vendor from scope
go-finding -filter-generated -generated-exclude "**/vendor/**"
```

## CLI Flags

| Flag                      | Default | Description                                                              |
| ------------------------- | ------- | ------------------------------------------------------------------------ |
| `-dir`                    | `.`     | Root directory to analyze                                                |
| `-format`                 | `text`  | Output format: `text`, `markdown`, `csv`, `tsv`, `json`, `sarif`         |
| `-min-severity`           | `info`  | Minimum severity: `info`, `warning`, `error`, `critical`                 |
| `-max-iterations`         | `5`     | Maximum pipeline iterations                                              |
| `-parallel`               | `true`  | Run detectors in parallel                                                |
| `-verify`                 | `false` | Verify fixes by re-running detectors                                     |
| `-timeout`                | `10m`   | Pipeline timeout (Go duration format)                                    |
| `-config`                 | `""`    | Path to YAML or JSON config file                                         |
| `-cpuprof`                | `""`    | Write CPU profile to file                                                |
| `-memprof`                | `""`    | Write memory profile to file                                             |
| `-filter-generated`       | `false` | Filter out findings from auto-generated files                            |
| `-filter-generated-types` | `all`   | Comma-separated generator types: `all`, `sqlc`, `templ`, `mockgen`, etc. |
| `-generated-exclude`      | `""`    | Comma-separated glob patterns to exclude from filtering                  |
| `-generated-include`      | `""`    | Comma-separated glob patterns restricting filter scope                   |

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

# Filter findings from auto-generated files
# filterGenerated: true
# filterGenTypes: "sqlc,protobuf"
# generatedExclude:
#   - "**/vendor/**"
# generatedInclude:
#   - "pkg/**"
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
    ID          ID                // Stable unique identifier (branded type)
    Rule        RuleName          // Rule/check name (branded type)
    ToolName    ToolName          // Source tool name (branded type)
    Message     string            // Human-readable description
    Severity    Severity          // info, warning, error, critical
    Position    Position          // Where the issue is
    Category    Category          // Domain classification
    FixStrategy FixStrategy       // none, suggest, direct, ai
    Suggestion  string            // Human-readable fix description
    BeforeCode  string            // Code before the fix
    AfterCode   string            // Code after the fix
    Range       *Range            // For span-based findings
    Related     []RelatedRef      // Related findings (with optional Range)
    Suppression *Suppression      // If suppressed
    Metadata    map[string]string // Tool-specific key-value pairs
}
```

### RelatedRef

Related findings link to issues in other locations:

```go
finding.RelatedRef{
    ID:       "govet:printf:main.go:10:3",
    Relation: "duplicate",
    Message:  "same issue reported here",
    Range:    &finding.Range{Start: pos1, End: pos2}, // optional span
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

#### Pipeline behavior

| Strategy  | `HasFix()`                  | `IsAutoFixable()` | `CanAutoApply()` | `NeedsAI()` |
| --------- | --------------------------- | ----------------- | ---------------- | ----------- |
| `none`    | false                       | false             | false            | false       |
| `suggest` | true (requires `AfterCode`) | false             | false            | false       |
| `direct`  | true (requires `AfterCode`) | true              | true             | false       |
| `ai`      | true (requires `AfterCode`) | false             | false            | true        |

- `FixStrategyAI` is a **reserved placeholder**. No AI backend exists yet. Pipeline triage treats it like `FixStrategySuggest` (no auto-apply). Set `NeedsAI()` to `true` so consumers can identify findings that need AI-powered remediation when an AI backend becomes available.
- `FixStrategyDirect` is the only strategy the pipeline auto-applies. It requires both `BeforeCode` and `AfterCode` to be set.

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
// Combine multiple reports with deduplication by ID (default)
merged := finding.Combine([]*finding.Report{report1, report2})

// Combine without deduplication
merged = finding.Combine(reports, finding.WithDeduplication(false))

// Deduplicate by position instead of ID
merged = finding.Combine(reports, finding.WithDeduplicateBy(finding.DeduplicateByPosition))

// Deduplicate by rule+position
merged = finding.Combine(reports, finding.WithDeduplicateBy(finding.DeduplicateByRule))
```

## Correlation

Find related findings across different tools:

```go
correlations := finding.Correlate(allFindings)
for _, c := range correlations {
    fmt.Printf("Related: %v (reason: %s, confidence: %.2f)\n",
        c.FindingIDs, c.Reason, c.Score)
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

// Streaming (no intermediate buffer allocation)
err = report.WriteSARIF(os.Stdout)
err = report.WriteSARIFFiltered(os.Stdout, finding.SeverityWarning)
```

### SARIF Round-Trip Fidelity

Importing SARIF that was exported by go-finding preserves all fields via the `properties` bag (`go-finding/*` prefix):

```go
// Export → Import round-trip
sarif, _ := report.ToSARIF()
findings, _ := finding.FindingsFromSARIF(sarif)
// All fields preserved: ID, Severity, Category, Tags, Confidence, etc.
```

**Known limitations:**

- Suppressed findings are excluded from export (lossy)
- `SeverityCritical` maps to SARIF `"error"` (no critical level in SARIF 2.1.0); original severity preserved in properties
- Non-go-finding SARIF (from other tools) imports with best-effort mapping; unknown fields land in `Metadata`

### SARIF Import

```go
findings, err := finding.FindingsFromSARIF(sarifData)
```

Imported findings get auto-generated IDs if not present in the SARIF data.

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

p, err := pipeline.New(cfg, ".", detector1, detector2)
if err != nil {
    log.Fatal(err)
}
result, err := p.Run(ctx)

// Inspect results
fmt.Println("Stable:", result.Stable())
fmt.Println("Iterations:", result.TotalIterations)
fmt.Println("Findings:", result.TotalDetected)

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
fmt.Println("Fixes applied:", snapshot.FixesApplied)
for detector, count := range snapshot.FindingsFound {
    fmt.Printf("  %s: %d findings\n", detector, count)
}
```

## Generated File Filtering

The `GeneratedFileFilter` processor automatically removes findings from auto-generated Go source files. It uses [gogenfilter](https://github.com/LarsArtmann/gogenfilter) for two-phase detection: filename-based first (zero I/O), then content-based when needed.

### Supported Generators

`all`, `counterfeiter`, `deepcopy-gen`, `easyjson`, `ent`, `generic`, `go-enum`, `go-swagger`, `gqlgen`, `mockgen`, `mockery`, `moq`, `msgp`, `oapi-codegen`, `protobuf`, `sqlc`, `stringer`, `templ`, `wire`. Use `all` to enable all detectors. `generic` matches any `// Code generated by` comment.

### Library Usage

```go
import (
    "github.com/LarsArtmann/gogenfilter/v3"
    "github.com/larsartmann/go-finding/pipeline"
)

// Create the filter with all generators
opt, _ := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
filter, err := pipeline.NewGeneratedFileFilter(nil, opt)
if err != nil {
    log.Fatal(err)
}

// Add to pipeline config
cfg := pipeline.Config{
    Processors: []pipeline.FindingTransformer{filter},
    // ...
}

// Filter specific generators only
sqlcOpt, _ := gogenfilter.WithFilterOptions(gogenfilter.FilterSQLC, gogenfilter.FilterProtobuf)
filter, _ = pipeline.NewGeneratedFileFilter(nil, sqlcOpt)

// With include/exclude patterns
filter, _ = pipeline.NewGeneratedFileFilter(nil,
    opt,
    gogenfilter.WithExcludePatterns("**/vendor/**", "**/third_party/**"),
    gogenfilter.WithIncludePatterns("pkg/**"),
)
```

### CLI Usage

```bash
# Enable with all generators (default)
go-finding -filter-generated

# Only filter sqlc and protobuf
go-finding -filter-generated -filter-generated-types "sqlc,protobuf"

# Exclude vendor from filtering scope
go-finding -filter-generated -generated-exclude "**/vendor/**"
```

### Config File

```yaml
filterGenerated: true
filterGenTypes: "sqlc,protobuf"
generatedExclude:
  - "**/vendor/**"
generatedInclude:
  - "pkg/**"
```

### Graceful Degradation

If a file referenced by a finding cannot be read (e.g., deleted between detection and filtering), the finding is **kept** rather than dropped. A warning is logged via `slog` if a logger is provided. Findings without a file path (`Position.File == ""`) are always kept.

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
finding.CategoryOf(err)             // "validation"
finding.IsCategory(err, finding.ErrCategoryIO) // false

// Type assertion
var fe *finding.FindingError
if errors.As(err, &fe) {
    fmt.Println(fe.Category, fe.Message)
}
```

## Built-in Detectors

The library ships with detector implementations in `internal/detectors/`:

- **`govet`** — Wraps `go vet -json` output (`NewGoVetDetector(dir)`)
- **`staticcheck`** — Wraps staticcheck JSON output (`NewStaticcheckDetector(dir)`)

## Builder API

For complex finding construction with validation, use the Builder:

```go
f, err := finding.NewBuilder("unused-var", "my-tool", "variable x is unused",
    finding.SeverityWarning, finding.Pos("main.go", 5, 2)).
    WithID("custom-id").
    WithCategory(finding.CategoryUnused).
    WithConfidence(finding.ConfidenceHigh).
    WithFixStrategy(finding.FixStrategyDirect).
    WithBeforeCode("x := 1").
    WithAfterCode("_ = x").
    WithTags(finding.TagUnused, finding.TagBug).
    WithMetadata(map[string]string{"source": "staticcheck"}).
    Build()

// MustBuild panics on validation error (use in tests/init code)
f2 := finding.NewBuilder("rule", "tool", "msg",
    finding.SeverityError, finding.Pos("a.go", 1, 1)).MustBuild()
```

`Build()` calls `Validate()` and returns detailed per-field errors.

## Suppression

Findings can be suppressed with a reason and optional TTL:

```go
expiry := time.Now().Add(24 * time.Hour)
f.Suppression = &finding.Suppression{
    Kind:      finding.SuppressionKindInSource,
    Rule:      "unused-var",
    Reason:    "intentionally unused in test",
    ExpiresAt: &expiry,
}

f.IsSuppressed()                    // true
f.Suppression.IsActive(time.Now())  // true (not expired)
```

Three suppression kinds: `in-source`, `in-config`, `in-review`.

## Confidence

`Confidence` is a named float64 type on a 0.0–1.0 scale:

```go
f.Confidence = finding.ConfidenceHigh  // 0.75
f.Confidence.IsValid()                 // true
f.Confidence.Clamp()                   // ensures [0.0, 1.0]
```

Named constants: `ConfidenceNone` (0.0), `ConfidenceLow` (0.25), `ConfidenceMedium` (0.5), `ConfidenceHigh` (0.75), `ConfidenceFull` (1.0).

## Tags

Tags provide multi-label classification for findings:

```go
f.Tags = finding.Tags{finding.TagSecurity, finding.TagBug}

f.Tags.Contains(finding.TagSecurity) // true
f.Tags.IsValid()                      // true (all tags are valid)
```

Standard tags: `TagSecurity`, `TagBug`, `TagPerformance`, `TagStyle`, `TagDeprecated`, `TagExperimental`, `TagUnused`, `TagDuplicate`, `TagComplexity`, `TagVulnerability`, `TagCompatibility`.

## Diff

Compare two finding sets:

```go
result := finding.Diff(beforeFindings, afterFindings)
fmt.Println(result.Stats())     // "+2 -1 ~0 =3"
fmt.Println(result.HasChanges()) // true
```

`DiffResult` contains `Added`, `Removed`, `Modified` (with before/after pairs), and `Unchanged` slices.
`Modified` tracks findings with the same ID but different content via `ModifiedPair{Before, After}`.
All slices are sorted by ID.

### Sorting

```go
finding.SortFindingsByID(findings) // in-place sort by finding ID
```

## Conflict Detection

The pipeline subpackage provides conflict detection for overlapping fixes:

```go
groups, conflicts := pipeline.DetectConflicts(fixes)
for _, g := range groups {
    fmt.Printf("Safe group in %s: %d fixes\n", g.File, len(g.Fixes))
}

// Detailed conflict analysis
infos := pipeline.AnalyzeConflicts(fixes)
for _, info := range infos {
    fmt.Printf("%s conflicts with %d others: %s\n",
        info.Finding.ID, len(info.ConflictsWith), info.Reason)
}
```

### Byte-Level Conflict Detection

Enable precise byte-level conflict detection (opt-in, position-based is default):

```go
cfg := pipeline.Config{
    ByteLevelConflictDetection: true,
    // ...
}
```

When enabled, the pipeline uses the FixEngine to resolve exact byte offsets and detects
overlapping edits at the byte level rather than comparing position ranges.

## Fix Engine and Providers

The pipeline subpackage includes a byte-level fix engine with composable providers:

```go
// In-memory engine
engine := pipeline.NewFixEngine()
result, err := engine.Apply(content, fixes)

// Filesystem applier with backup/rollback
applier, err := pipeline.NewFixApplier(rootDir)
defer applier.Close()
applied, err := applier.Apply(ctx, fixes)
```

**Default provider chain** (tried in order):

1. `OffsetProvider` — byte offset ranges
2. `LineProvider` — line/column positions
3. `SubstringProvider` — BeforeCode text matching

**Custom providers** for domain-specific transformations (e.g., Go AST):

```go
applier, err := pipeline.NewFixApplierWithProviders(rootDir, myASTProvider)
```

### FixEdit

`FixEdit` represents a byte-level edit operation:

```go
edit := pipeline.FixEdit{
    Offset:     120,              // byte offset in file
    Length:     5,                // bytes to replace
    Replacement: []byte("world"), // new content
}
```

Edits are applied in descending offset order against the original content snapshot, so multiple edits to the same file are correct.

### TriageFunc

Customize how the pipeline categorizes findings into safe fixes vs conflicts:

```go
cfg := pipeline.Config{
    TriageFunc: func(findings []finding.Finding) *pipeline.TriageResult {
        var safe, conflicts []finding.Finding
        for _, f := range findings {
            if f.Confidence >= finding.ConfidenceHigh && f.HasCodeChange() {
                safe = append(safe, f)
            } else {
                conflicts = append(conflicts, f)
            }
        }
        return &pipeline.TriageResult{SafeFixes: safe, Conflicts: conflicts}
    },
}
```

`DefaultTriageFunc` preserves existing behavior (uses `HasFix()` / `IsAutoFixable()` checks).

## LSP Diagnostics

Convert findings to LSP Diagnostics for IDE integration:

```go
diag := f.ToLSP()
// diag.Severity, diag.Message, diag.Range, diag.Code, diag.Source

// Convert back
f := finding.FromLSP("file:///path/to/main.go", lspDiag)
```

LSP round-trip is lossless via `LSPDiagnosticData`: ID, Severity (including Critical),
FixStrategy, Confidence, BeforeCode, AfterCode, Suggestion, Snippet, Suppression,
Metadata, Category, Tags, and RelatedRef.FindingID are all preserved in `diag.Data`.
Diagnostic tags (unnecessary, deprecated) are also preserved via Metadata keys.

## Formatting

Human-readable output:

```go
err := finding.FormatText(os.Stdout, findings)    // single-line per finding
err := finding.FormatMarkdown(os.Stdout, findings) // markdown table
```

Both return errors instead of silently swallowing encoding failures.

## JSON

```go
// Single finding
f, err := finding.FromJSON(data) // validates on import

// Report
r, dropped, err := finding.ReportFromJSON(data) // drops invalid findings
fmt.Printf("%d invalid findings dropped", dropped)

// Pretty-print
pretty, err := report.PrettyJSON()         // all findings
pretty, err := report.PrettyJSONFiltered() // excludes suppressed
```

## go/analysis Integration

Convert from the standard Go analysis framework:

```go
import "github.com/larsartmann/go-finding/analysis"

f := analysis.FromDiagnostic(diag, pass.Fset, "my-analyzer", "RULE001")
```
