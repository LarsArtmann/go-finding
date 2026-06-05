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
    // Run your tool, parse output, return findings
    return []finding.Finding{
        {
            ID:       finding.GenerateID("mytool", "RULE001", finding.Pos("main.go", 10, 5)),
            Rule:     "RULE001",
            ToolName: "mytool",
            Message:  "potential nil dereference",
            Severity: finding.SeverityError,
            Position: finding.Pos("main.go", 10, 5),
        },
    }, nil
}
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

See `internal/detectors/govet.go` for the production implementation.

## Using FindingProcessor for Preprocessing

Run transforms between detection and triage:

```go
filter, _ := pipeline.NewGeneratedFileFilter(nil, gogenfilter.WithFilterOptions(gogenfilter.FilterAll))

cfg := pipeline.Config{
    Processors: []pipeline.FindingProcessor{filter},
}
```

Processors run in order. `ProcessorFunc` wraps a simple function; `NamedProcessorFunc` adds a name for logging.

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
