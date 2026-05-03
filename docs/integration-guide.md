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

See `internal/detectors/govet.go` for a complete example of wrapping `go vet` JSON output.
