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

## Simple Fixes (Core Package)

For the 80% case — BeforeCode→AfterCode string replacement on files — use the core package's `ApplySimpleFixes`. No pipeline import needed:

```go
results := finding.ApplySimpleFixes(findingsWithDirectFixes)
for file, fileResults := range results {
    for _, r := range fileResults {
        if r.Applied {
            fmt.Printf("Fixed %s in %s\n", r.FindingID, file)
        } else {
            fmt.Printf("Skipped %s: %s\n", r.FindingID, r.Reason)
        }
    }
}
```

`ApplySimpleFixes` reads each file, applies `strings.Replace` with count=1 per finding, and writes back. Findings without BeforeCode/AfterCode are skipped automatically.

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

### Per-Finding Outcomes

`ApplyWithOutcomes` tells you exactly what happened to each finding — applied, refused, conflicted, or failed. A provider that matches a finding but produces zero edits is reported as `refused` instead of being silently indistinguishable from success:

```go
result := engine.ApplyWithOutcomes(content, findings)

for _, o := range result.Outcomes {
    switch o.Status {
    case pipeline.FixOutcomeApplied:
    case pipeline.FixOutcomeNoChange:
    case pipeline.FixOutcomeRefused:
        fmt.Printf("provider refused %s\n", o.Finding.ID)
    case pipeline.FixOutcomeConflict:
    case pipeline.FixOutcomeInvalid:
    case pipeline.FixOutcomeFailed:
        fmt.Printf("failed %s: %v\n", o.Finding.ID, o.Err)
    }
}

failed := result.FailedOutcomes == nil // via HasErrors(); OutcomeFor(id) for lookups
```

Outcomes come in input order, one per input finding. `result.OutcomeFor(id)` looks up a single finding; `result.OutcomeCounts()` tallies by status; `result.HasErrors()` reports provider failures.

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
    // The failing file is restored; files fixed earlier keep their fixes.
    // Inspect report := applier.ApplyWithReport(...) for details.
    return err
}

// When done, cleanup backups:
err = applier.Cleanup()
```

### Rollback Policy

By default (`RollbackPolicyFailingFile`), a file failure restores only the failing file: independent fixes in other files stay applied. Soft per-finding failures (provider resolve errors, refused findings) never abort the run or roll anything back. For all-or-nothing semantics, opt in:

```go
applier.SetRollbackPolicy(pipeline.RollbackPolicyAllFiles)
```

or via pipeline `Config.FixRollbackAllFiles`, or the config-file field `fixRollbackAllFiles`.

### With Shift Maps

`ApplyWithShiftMap` returns line-shift maps so you can adjust positions of remaining findings after edits:

```go
count, applied, shiftMaps, err := applier.ApplyWithShiftMap(ctx, findings)
for file, sm := range shiftMaps {
    fmt.Printf("File %s shifted by %d lines\n", file, sm.TotalShift())
}
```

### With a Full Report

`ApplyWithReport` combines applied findings, per-finding outcomes, shift maps, and the list of rolled-back files:

```go
report, err := applier.ApplyWithReport(ctx, findings)
fmt.Printf("applied %d, rolled back %v\n", report.Applied, report.RolledBack)
for _, o := range report.FailedOutcomes() {
    fmt.Printf("failed %s: %v\n", o.Finding.ID, o.Err)
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
