# Tool Integration Guide

This guide shows how to integrate a static analysis tool with `go-finding`.

## Implementing the Detector Interface

```go
package mytool

import (
    "context"
    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-finding/pipeline"
)

type MyDetector struct {
    dir string
}

func NewMyDetector(dir string) pipeline.Detector {
    return &MyDetector{dir: dir}
}

func (d *MyDetector) Name() string {
    return "mytool"
}

func (d *MyDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
    // Use the Builder API for validated findings (recommended).
    f, err := finding.NewBuilder(
        finding.RuleName("RULE001"), finding.ToolName("mytool"),
        "potential nil dereference",
        finding.SeverityError,
        finding.Pos("main.go", 10, 5),
    ).
        WithCategory(finding.CategoryCorrectness).
        WithConfidence(finding.ConfidenceHigh).
        Build()
    if err != nil {
        return nil, err
    }

    return []finding.Finding{f}, nil
}
```

## Quick Detection (no fix loop)

For tools that only need detection and output (no automated fixing), use the
convenience function:

```go
findings, err := pipeline.Detect(ctx, detector1, detector2)
// → []finding.Finding, ready for Report/SARIF/LSP output
```

This runs detectors concurrently, filters suppressed findings, and returns the
combined result. For fix application, retry logic, metrics, or stage hooks,
use the full [Pipeline](#using-with-the-pipeline) instead.

## Applying Fixes to In-Memory Content

If you already have file content in memory (e.g. from a git blob or editor buffer),
use `ApplyToContent` instead of the filesystem-bound `FixApplier`:

```go
content, _ := os.ReadFile(path)
result, applied := pipeline.ApplyToContent(content, findings)
if applied > 0 {
    _ = os.WriteFile(path, result, 0o644)
}
```

## Recommended: Type-Alias Pattern

For tools with their own domain types (e.g. `Issue`, `Violation`), the cleanest
integration is Go type aliases. This eliminates all conversion boilerplate:

```go
package mytool

import "github.com/larsartmann/go-finding"

// Alias go-finding types to your domain names — zero conversion overhead.
type Issue = finding.Finding
type Severity = finding.Severity
type RuleCategory = finding.Category

// Your rule system produces go-finding types directly.
func DetectIssue(file string, line int) Issue {
    return Issue{
        ID:       finding.GenerateID("mytool", "RULE001", finding.Pos(file, line, 0)),
        Rule:     "RULE001",
        ToolName: "mytool",
        Severity: Severity(finding.SeverityWarning),
        Position: finding.Pos(file, line, 0),
    }
}
```

This pattern means your `[]Issue` IS `[]finding.Finding` — no conversion needed
when passing to `Report.AddFindings()`, `WriteSARIF()`, or `pipeline.Detect()`.

## Migration: Types-Only to Pipeline

If you start with just the data types and later want the pipeline:

1. Implement `finding.Detector` on your analyzer
2. Replace manual orchestration with `pipeline.Detect(ctx, detectors...)`
3. Optionally upgrade to full `pipeline.New()` + `Run()` for fix application

```go
// Before: manual orchestration
findings := myAnalyzer.Run()
report := finding.NewReport(toolInfo)
report.AddFindings(findings)

// After: pipeline convenience (same result, plus parallel + cancellation)
findings, _ := pipeline.Detect(ctx, myDetector)
report := finding.NewReport(toolInfo)
report.AddFindings(findings)

// Later: full pipeline with fix application
cfg := pipeline.DefaultConfig()
p, _ := pipeline.New(cfg, ".", myDetector)
result, _ := p.Run(ctx)
```

## Using with the Pipeline

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-finding/pipeline"
)

func main() {
    config := pipeline.DefaultConfig()
    config.ParallelDetectors = true
    config.OnFinding = func(f finding.Finding) {
        fmt.Fprintf(os.Stderr, "found: %s\n", f.ID)
    }

    p, err := pipeline.New(config, ".", myDetector, otherDetector)
    if err != nil {
        panic(err)
    }

    result, err := p.Run(context.Background())
    if err != nil {
        panic(err)
    }

    // Output as SARIF
    report := finding.NewReport(finding.ToolInfo{Name: "my-tool"})
    for _, iter := range result.Iterations {
        report.AddFindings(iter.Findings())
    }
    report.ComputeSummary()

    data, err := report.ToSARIF()
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
}
```

## Using with the CLI

Register your detector and build the CLI:

```go
package main

import "github.com/larsartmann/go-finding/cmd/go-finding"

func main() {
    gofinding.RegisterDetector("mytool", func(dir string) pipeline.Detector {
        return NewMyDetector(dir)
    })
    gofinding.Main()
}
```

Then use: `go-finding -detectors=mytool -dir=./...`

## Converting Tool Output to Findings

Each tool has different output formats. Here's the pattern:

1. **Run the tool** as a subprocess (or call its API)
2. **Parse the output** (JSON, text, etc.)
3. **Map each issue** to a `Finding`:
   - `ID`: Use `GenerateID()` for deterministic IDs
   - `Rule`: The rule/check that was violated
   - `ToolName`: Your tool's name
   - `Message`: Human-readable description
   - `Severity`: Map to `SeverityInfo`/`Warning`/`Error`/`Critical`
   - `Position`: File, line, column
   - `FixStrategy`: `None`, `Suggest`, `Direct`, or `AI`
   - `Suggestion` + `BeforeCode`/`AfterCode`: If the tool provides fix suggestions

## Example: Wrapping `govet`

A complete example of wrapping `go vet -json` output as a Detector:

```go
package detectors

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os/exec"
    "strconv"
    "strings"

    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-finding/pipeline"
)

func NewGoVetDetector(dir string) pipeline.Detector {
    return pipeline.NamedDetectorFunc("govet", func(ctx context.Context) ([]finding.Finding, error) {
        cmd := exec.CommandContext(ctx, "go", "vet", "-json", "./...")
        cmd.Dir = dir

        out, err := cmd.Output()
        if err != nil {
            var exitErr *exec.ExitError
            if errors.As(err, &exitErr) && len(out) > 0 {
                return parseGoVetJSON(out, dir), nil
            }
            return nil, fmt.Errorf("run go vet: %w", err)
        }

        return parseGoVetJSON(out, dir), nil
    })
}

func parseGoVetJSON(data []byte, dir string) []finding.Finding {
    var diagnostics map[string]json.RawMessage
    if err := json.Unmarshal(data, &diagnostics); err != nil {
        return nil
    }

    var findings []finding.Finding
    for name, raw := range diagnostics {
        var entries []struct {
            Posn    string `json:"posn"`
            Message string `json:"message"`
        }
        if err := json.Unmarshal(raw, &entries); err != nil {
            continue
        }
        for _, e := range entries {
            pos := parsePosn(e.Posn, dir)
            findings = append(findings, finding.Finding{
                ID:          finding.GenerateID("govet", name, pos),
                Rule:        name,
                ToolName:    "govet",
                Message:     e.Message,
                Severity:    finding.SeverityWarning,
                Position:    pos,
                Category:    finding.CategoryCorrectness,
                FixStrategy: finding.FixStrategySuggest,
                Confidence:  finding.ConfidenceHigh,
            })
        }
    }
    return findings
}
```

Key patterns:

- `NamedDetectorFunc` wraps a function as a `Detector` interface
- `GenerateID` creates stable, deterministic IDs from tool + rule + position
- `exec.CommandContext` respects context cancellation
- Non-zero exit with output is treated as partial success, not hard failure
- `Category`, `FixStrategy`, and `Confidence` are set from domain knowledge

See `cmd/go-finding/internal/detectors/govet.go` for the production implementation.

## Using FindingTransformer for Preprocessing

Run transforms between detection and triage:

```go
filter, _ := pipeline.NewGeneratedFileFilter(nil, gogenfilter.WithFilterOptions(gogenfilter.FilterAll))

cfg := pipeline.Config{
    Processors: []pipeline.FindingTransformer{filter},
}
```

Processors run in order. `TransformerFunc` wraps a simple function; `NamedTransformerFunc` adds a name for logging.

## Using FixProvider for Custom Fix Resolution

When the default provider chain (Offset → Line → Substring) isn't enough:

```go
type ASTProvider struct{}

func (p *ASTProvider) Name() string { return "ast" }

func (p *ASTProvider) Resolve(ctx context.Context, content []byte, f finding.Finding) ([]pipeline.FixEdit, error) {
    // Parse AST, find exact byte offsets, return edits
    return []pipeline.FixEdit{{
        Offset:      byteOffset,
        Length:      byteLength,
        Replacement: []byte(f.AfterCode),
    }}, nil
}

applier, _ := pipeline.NewFixApplierWithProviders(rootDir, &ASTProvider{})
```

## SARIF Export with Suppressed Findings

By default, `ToSARIF()` excludes suppressed findings. Use `WithIncludeSuppressed()`
to emit them with SARIF suppression arrays for full round-trip fidelity:

```go
// Default: suppressed findings dropped
data, _ := report.ToSARIF()

// Include suppressed findings with suppression metadata
data, _ := report.ToSARIFWithOpts(finding.WithIncludeSuppressed())

// Combined with severity filtering
data, _ := report.ToSARIFWithOpts(
    finding.WithIncludeSuppressed(),
    finding.WithMinSeverity(finding.SeverityWarning),
)
```

## LSP Conversion with Full Fidelity

`Finding.ToLSP()` populates `LSPDiagnostic.Data` with all go-finding-specific fields
(ID, FixStrategy, Confidence, Category, Tags, code data). `FromLSP()` restores them:

```go
diag := finding.ToLSP(myFinding)
// diag.Data.ID, diag.Data.FixStrategy, diag.Data.Confidence all preserved

restored := finding.FromLSP("file:///main.go", diag)
// restored.ID, restored.FixStrategy, restored.Confidence all match original
```

For plain LSP diagnostics without Data (e.g., from gopls), `FromLSP` works normally
with sensible defaults (FixStrategyNone, auto-generated ID).

## FilePath Branded Type

`Position.File` uses `FilePath` (a named `string` type), not raw `string`. This provides
compile-time safety preventing accidental assignment of IDs, rule names, or tool names
to file path fields.

```go
// Correct — string literal auto-converts
pos := finding.Position{File: "main.go", Line: 10}

// Correct — explicit conversion for string variables
path := getPath()
pos := finding.Position{File: finding.FilePath(path), Line: 10}

// Constructors accept FilePath
pos := finding.Pos(finding.FilePath("main.go"), 10, 5)
r := finding.NewRange(finding.FilePath("main.go"), 1, 1, 5, 10)
```

JSON serialization is identical to plain string — `FilePath` marshals as `"main.go"`.
