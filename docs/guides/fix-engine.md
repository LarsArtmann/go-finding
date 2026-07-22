# Fix Engine Guide

The FixEngine applies byte-level edits to source code based on `Finding` data. This guide covers all usage patterns, from simple content-level fixes to disk-based application with backup and rollback.

## Table of Contents

- [Quick Start: ApplyToContent](#quick-start-applytocontent)
- [Simple Fixes (Core Package)](#simple-fixes-core-package)
- [Standalone FixEngine](#standalone-fixengine)
- [Disk-Based FixApplier](#disk-based-fixapplier)
- [Custom FixProviders](#custom-fixproviders)
- [Conflict Detection](#conflict-detection)
- [Go AST Provider](#go-ast-provider)
- [Full Pipeline Integration](#full-pipeline-integration)

---

## Quick Start: ApplyToContent

The simplest API for applying fixes to in-memory content:

```go
import "github.com/larsartmann/go-finding/pipeline"

content, err := os.ReadFile("main.go")
if err != nil { return err }

findings := []finding.Finding{
    {
        Rule:        "unused-var",
        ToolName:    "govet",
        Message:     "x declared but not used",
        Severity:    finding.SeverityWarning,
        Position:    finding.Position{File: "main.go", Line: 10, Column: 5},
        FixStrategy: finding.FixStrategyDirect,
        BeforeCode:  "x := 42",
        AfterCode:   "_ = x",
    },
}

result, applied := pipeline.ApplyToContent(content, findings)
fmt.Printf("Applied %d fixes\n", applied)
```

`ApplyToContent` uses the default FixProvider chain (Offset → Line → Substring) and returns the modified content. No filesystem access — perfect for editor buffers, git blobs, or API responses.

---

## Standalone FixEngine

For more control over provider selection and conflict reporting:

```go
engine := pipeline.NewFixEngine()

// Apply returns (modifiedContent, appliedFindings, count)
result, applied, count := engine.Apply(content, findings)

// For conflict information:
appliedFindings, appliedEdits, conflicts, result, errs := engine.ApplyWithConflicts(content, findings)
for _, c := range conflicts {
    fmt.Printf("skipped conflicting fix at %s\n", c.Finding.Position)
}
```

### Custom Providers

```go
engine := pipeline.NewFixEngineWithProviders(
    &goast.Provider{},    // AST-aware Go fixes
    &pipeline.OffsetProvider{},   // byte offset fallback
    &pipeline.LineProvider{},     // line:column fallback
)
```

Providers are tried in order. The first `CanHandle(finding)` wins.

---

## Disk-Based FixApplier

For filesystem operations with backup and automatic rollback:

```go
applier, err := pipeline.NewFixApplier(".")
if err != nil { return err }

// Apply reads files from disk, writes fixes, creates backups
count, err := applier.Apply(ctx, findings)
if err != nil {
    // All modified files are automatically rolled back
    return err
}

// When done, cleanup backups:
err = applier.Cleanup()
```

### With Shift Maps

`ApplyWithShiftMap` returns line-shift maps so you can adjust positions of remaining findings after edits:

```go
count, applied, shiftMaps, err := applier.ApplyWithShiftMap(ctx, findings)
for file, sm := range shiftMaps {
    fmt.Printf("File %s shifted by %d lines\n", file, sm.TotalShift())
}
```

---

## Conflict Detection

When two fixes overlap, the FixEngine skips the conflicting one:

```go
applied, edits, conflicts, result, errs := engine.ApplyWithConflicts(content, findings)
// conflicts contains findings whose byte ranges overlapped with applied edits
```

For byte-level precision (detecting overlaps between offset-based and line-based fixes):

```go
// In Pipeline config:
cfg.ByteLevelConflictDetection = true
```

---

## Go AST Provider

The `goast.Provider` parses Go source into an AST for precise, structure-aware fixes:

```go
import "github.com/larsartmann/go-finding/pipeline/goast"

engine := pipeline.NewFixEngineWithProviders(
    &goast.Provider{},
    &pipeline.OffsetProvider{},
)
```

The GoAST provider:

- Only handles `.go` files with code changes (`BeforeCode` or `AfterCode`)
- Parses the file to find exact byte offsets from AST node positions
- Falls back to text-based providers for non-Go files

---

## FixProvider Chain

The default chain resolves fix locations in this order:

| Provider              | Match Criteria         | Resolution                |
| --------------------- | ---------------------- | ------------------------- |
| **OffsetProvider**    | `Position.Offset >= 0` | Direct byte offset        |
| **LineProvider**      | `Position.Line > 0`    | Line+column → byte offset |
| **SubstringProvider** | `BeforeCode != ""`     | Find substring in content |

Each provider:

1. `CanHandle(finding) bool` — Can this provider resolve this finding?
2. `Edits(content, finding) ([]FixEdit, error)` — Produce byte-level edits

### Custom Provider Example

```go
type MyProvider struct{}

func (*MyProvider) Name() string { return "my-provider" }
func (*MyProvider) CanHandle(f finding.Finding) bool {
    return f.HasCodeChange() && /* domain check */
}
func (*MyProvider) Edits(content []byte, f finding.Finding) ([]pipeline.FixEdit, error) {
    // Return byte-level edits
    return []pipeline.FixEdit{{
        Start:   startOffset,
        End:     endOffset,
        NewText: []byte(f.AfterCode),
    }}, nil
}
```

---

## Full Pipeline Integration

For detect → triage → fix → verify loops:

```go
p, err := pipeline.New(pipeline.Config{
    MaxIterations: 3,
    FixProviders:  []pipeline.FixProvider{&goast.Provider{}},
}, ".", detector1, detector2)

result, err := p.Run(ctx)
// result.Iterations contains per-iteration findings and fixes
// result.Metrics has timing and fix counts
```

The Pipeline uses FixApplier internally with the registered providers.
