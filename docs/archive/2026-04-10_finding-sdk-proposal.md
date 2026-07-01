# Finding SDK — Unified Pipeline & Data Model Proposal

**Status:** Draft v2 | **Date:** 2026-04-10

---

## The Problem We're Actually Solving

Seven tools detect issues. Zero tools **route them to remediation**.

Current workflow:

```
run tool → read output → fix manually → re-run → new issues → repeat 5×
```

The loop is manual, lossy, and incomplete. Each tool invents its own types:

| Project                          | Finding Type                                                                                           | Severity                                            | Position                                             | Fix Model                                                                                                                       | Output                                       |
| -------------------------------- | ------------------------------------------------------------------------------------------------------ | --------------------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| **golangci-lint-auto-configure** | `LinterRecommendation`, `ValidationError`                                                              | `LinterPriority` (Critical/High/Medium/Optional)    | Config path only (no source lines)                   | `AutoFix bool` on linter metadata                                                                                               | JSON, HTML                                   |
| **BuildFlow**                    | `BinaryViolation`, `TODOViolation`, `HierarchicalErrorsViolation`, etc.                                | `ViolationSeverity` (Critical/High/Medium/Low/Info) | `GetFile()`, `GetLine()`, `GetColumn()`              | `AutoFix bool` in config; `Suggestion string` on violations                                                                     | JSON, SARIF, HTML                            |
| **rules**                        | `analysis.Analyzer` diagnostics (no custom type)                                                       | None (delegated to host framework)                  | `token.Pos` via `go/analysis`                        | None (report-only)                                                                                                              | Text (via `go vet`)                          |
| **art-dupl**                     | `Clone`, `CloneGroup`                                                                                  | `CloneSeverity` (low/medium/high/critical)          | `LineNumber`, `BytePosition`, `Filename` (no column) | None; `Suggestion` text only                                                                                                    | JSON, SARIF, HTML, CSV, text                 |
| **branching-flow**               | `StrongIDViolation`, `BoolBlindnessViolation`, `PrimitiveTypeViolation`, `DuplicateGroup`, `Detection` | `Severity` (critical/high/medium/low)               | `SourceLocation` (file, line, column)                | Rich: `StrongIDSuggestion{BeforeCode, AfterCode}`, `BitFlagSuggestion`, `EnumSuggestion`, `CompositionSuggestion`; `--fix` flag | JSON, SARIF, HTML, Markdown, text            |
| **go-auto-upgrade**              | `Change`, `Warning`                                                                                    | None (implicit: change vs warning vs error)         | `PathString`, `LineInt` (no column)                  | All changes are auto-fixable via AST rewrite; `Result.Content` has new code                                                     | Text (slog)                                  |
| **hierarchical-errors**          | `ErrorViolation`, `ErrorFlow`, `ErrorHierarchy`                                                        | `Severity` (low/medium/high)                        | `token.Position` (file, line, column, offset)        | `Suggestion string`; SARIF `Fix` structs (descriptive only)                                                                     | JSON, SARIF, HTML, DOT, Mermaid, agent, text |

---

## Why Not Just SARIF?

SARIF 2.1.0 is excellent as an **interchange format for reporting**. 4 of 7 tools already emit it. GitHub Code Scanning, Azure DevOps, and VS Code consume it natively.

SARIF is **insufficient** for three things this SDK must do:

| Gap                        | Why it matters                                                                                                                                                                                                             |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Fix strategy**           | SARIF has `fixes[]` with `artifactChanges`, but can't distinguish "apply mechanically" from "AI should figure it out" from "no fix possible". The pipeline needs to know _how_ to remediate, not just _that_ a fix exists. |
| **Cross-tool correlation** | golangci-lint and branching-flow may flag the same line. SARIF has no merge protocol, no cross-run identity, no dedup semantics.                                                                                           |
| **Pipeline state**         | SARIF is a terminal snapshot. It can't express "fix → verify → re-detect → fix again → stable". The pipeline needs mutable working state, not serialized output.                                                           |

**Design stance:** SARIF is the output format. The SDK is the in-memory working representation + pipeline engine. SARIF is generated from `Report`, not the other way around.

---

## Relationship to Existing Go Standards

This SDK must **align with, not replace** existing Go ecosystem types.

### `go/analysis.Diagnostic` and `go/analysis.SuggestedFix`

Three tools (rules, art-dupl via golangci-lint integration, branching-flow via `--fix`) live inside the `go/analysis` framework. Its types:

```go
type Diagnostic struct {
    Pos     token.Pos
    Message string
    Code    string
    Category string
    Related []RelatedInformation
    SuggestedFixes []SuggestedFix
}
type SuggestedFix struct {
    Message   string
    TextEdits []TextEdit
}
```

**Alignment:** `Finding` is a superset. Every `Diagnostic` maps to a `Finding`. The SDK provides `FromDiagnostic()` but adds `Severity`, `FixStrategy`, `Confidence`, and `Related` chains that `Diagnostic` lacks.

### LSP `Diagnostic` + `CodeAction`

hierarchical-errors already has an LSP server. LSP types:

```go
type Diagnostic struct {
    Range    Range
    Severity DiagnosticSeverity  // Error=1, Warning=2, Information=3, Hint=4
    Source   string
    Code     any
    Message  string
}
```

**Alignment:** `Finding` → LSP `Diagnostic` is a lossy conversion (loses fix strategy, confidence, related). The SDK provides `ToLSPDiagnostic()` for LSP consumers, but the `Finding` carries more.

### `golangci-lint` JSON Output

De facto interchange format for Go linters. Every tool that integrates with golangci-lint speaks this:

```json
{ "Pos": "file.go:42:5", "Text": "message", "FromLinter": "gosec", "Severity": "warning" }
```

**Alignment:** The SDK parses golangci-lint JSON into `Finding` via `FromGolangciLintJSON()`. This is how BuildFlow already integrates external linters.

### `go-business-rules` (this project)

This project has `Violation` with severity, message, and context. `Finding` is a **parallel concept** for static analysis findings, not a replacement. `Violation` is runtime validation output; `Finding` is static analysis output. They share `Severity` semantics but serve different loops.

---

## The Pipeline (Why This SDK Exists)

```
┌─────────────────────────────────────────────────────────┐
│                     Pipeline                            │
│                                                         │
│  ┌──────────┐    ┌─────────┐    ┌──────┐    ┌────────┐  │
│  │ Detect   │───→│ Triage  │───→│ Fix  │───→│ Verify │  │
│  │ (tools)  │    │ (route) │    │ (act)│    │ (re-run│──┤
│  └──────────┘    └─────────┘    └──────┘    └────────┘  │
│       ↑                                     │           │
│       └─────────────────────────────────────┘           │
│                  (loop until stable)                     │
└─────────────────────────────────────────────────────────┘
```

### Detect

Run N tools, each producing `[]Finding`:

```go
findings, _ := pipeline.Detect(ctx, ".",
    artdupl.Detector{},
    branchingflow.Detector{},
    hierarchicalerrors.Detector{},
    goupgrade.Detector{},
)
```

### Triage

Route findings by fix strategy:

```go
direct  := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))
suggest := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategySuggest))
none    := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyNone))
```

### Fix

```go
// Direct fixes: apply deterministically
applied, remaining := pipeline.ApplyDirectFixes(ctx, direct)

// Suggest fixes: route to AI
aiFixed := pipeline.ApplyAIFixes(ctx, suggest, aiClient)
```

### Verify

```go
// Re-run detection on modified files
newFindings := pipeline.Verify(ctx, applied.ModifiedFiles())
// If new findings exist, loop
```

### The value proposition

Without the SDK: run tool → read → fix → re-run → repeat manually 5×.
With the SDK: one command, automated loop, direct fixes applied, AI fixes routed, remaining reported.

---

## Core Types

### Severity

```go
type Severity string

const (
    SeverityInfo     Severity = "info"
    SeverityWarning  Severity = "warning"
    SeverityError    Severity = "error"
    SeverityCritical Severity = "critical"
)
```

**Mapping from existing projects:**

| Project                               | Native                  | Maps To                         |
| ------------------------------------- | ----------------------- | ------------------------------- |
| BuildFlow `ViolationSeverityCritical` | Critical                | `critical`                      |
| BuildFlow `ViolationSeverityHigh`     | High                    | `error`                         |
| BuildFlow `ViolationSeverityMedium`   | Medium                  | `warning`                       |
| BuildFlow `ViolationSeverityLow`      | Low                     | `info`                          |
| BuildFlow `ViolationSeverityInfo`     | Info                    | `info`                          |
| art-dupl `CloneSeverityCritical`      | critical                | `critical`                      |
| art-dupl `CloneSeverityHigh`          | high                    | `error`                         |
| art-dupl `CloneSeverityMedium`        | medium                  | `warning`                       |
| art-dupl `CloneSeverityLow`           | low                     | `info`                          |
| branching-flow `SeverityCritical`     | critical                | `critical`                      |
| branching-flow `SeverityHigh`         | high                    | `error`                         |
| branching-flow `SeverityMedium`       | medium                  | `warning`                       |
| branching-flow `SeverityLow`          | low                     | `info`                          |
| hierarchical-errors `SeverityHigh`    | high                    | `error`                         |
| hierarchical-errors `SeverityMedium`  | medium                  | `warning`                       |
| hierarchical-errors `SeverityLow`     | low                     | `info`                          |
| go-auto-upgrade `Warning`             | (implicit)              | `warning`                       |
| go-auto-upgrade `Change`              | (implicit)              | `info`                          |
| go-auto-upgrade `Error`               | (implicit)              | `error`                         |
| LSP `DiagnosticSeverity` (1-4)        | Error/Warning/Info/Hint | `error`/`warning`/`info`/`info` |
| `go/analysis`                         | (none)                  | defaults to `warning`           |

### FixStrategy

```go
type FixStrategy string

const (
    FixStrategyNone    FixStrategy = "none"    // No fix available
    FixStrategySuggest FixStrategy = "suggest" // Human-readable suggestion, not machine-applicable
    FixStrategyDirect  FixStrategy = "direct"  // Deterministic code transformation (AST rewrite, formatter, etc.)
    FixStrategyAI      FixStrategy = "ai"      // Requires AI/LLM to generate context-aware fix
)
```

**Mapping from existing projects:**

| Project                              | Scenario                                    | FixStrategy                          |
| ------------------------------------ | ------------------------------------------- | ------------------------------------ |
| art-dupl                             | Clone detected (no suggestion)              | `none`                               |
| art-dupl                             | `Suggestion` text present                   | `suggest`                            |
| branching-flow                       | `StrongIDSuggestion{BeforeCode, AfterCode}` | `suggest`                            |
| branching-flow                       | `--fix` flag + deterministic rewrite        | `direct`                             |
| BuildFlow                            | `AutoFix=true` + formatter step             | `direct`                             |
| BuildFlow                            | `Suggestion string` on violation            | `suggest`                            |
| go-auto-upgrade                      | `Result.Content` with AST rewrite           | `direct`                             |
| go-auto-upgrade                      | `Warning` (manual review needed)            | `suggest`                            |
| hierarchical-errors                  | `Suggestion string` on ErrorViolation       | `suggest`                            |
| golangci-lint-auto-configure         | `AutoFix bool` on linter                    | `direct` (via `golangci-lint --fix`) |
| rules                                | No fixes                                    | `none`                               |
| `go/analysis.SuggestedFix` present   | Has `TextEdits`                             | `direct`                             |
| `go/analysis` without `SuggestedFix` | Report-only                                 | `none`                               |

### Suppression

Every tool has its own suppression mechanism. The SDK unifies them:

```go
type Suppression struct {
    Kind      SuppressionKind // InSource, InConfig, InReview
    Rule      string          // Which rule is suppressed
    Reason    string          // Why (from comment or config)
    ExpiresAt *time.Time      // Optional: temporary suppressions with expiry
}

type SuppressionKind string

const (
    SuppressionInSource SuppressionKind = "in-source" // //nolint, //lint:ignore, etc.
    SuppressionInConfig SuppressionKind = "in-config" // Config file rules
    SuppressionInReview SuppressionKind = "in-review" // Accepted as false positive after review
)
```

**Existing mechanisms:**

| Project             | Mechanism                                                                                     | Maps To                   |
| ------------------- | --------------------------------------------------------------------------------------------- | ------------------------- |
| golangci-lint       | `//nolint` comments                                                                           | `in-source`               |
| branching-flow      | `//lint:ignore STRONG_ID`                                                                     | `in-source`               |
| hierarchical-errors | `//nolint` + config-based rules with `FilePattern`, `FunctionPattern`, `LineStart`, `LineEnd` | `in-source` + `in-config` |
| BuildFlow           | Config exclusions                                                                             | `in-config`               |
| art-dupl            | None                                                                                          | (missing)                 |
| go-auto-upgrade     | None                                                                                          | (missing)                 |
| rules               | None                                                                                          | (missing)                 |

### Position

```go
type Position struct {
    File   string `json:"file"`
    Line   int    `json:"line,omitempty"`   // 1-based; 0 = not set
    Column int    `json:"column,omitempty"` // 1-based; 0 = not set
    Offset int    `json:"offset,omitempty"` // byte offset; 0 = not set
}

type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end,omitempty"`
}
```

**Alignment with existing types:**

| Existing Type                          | Conversion                                                                     |
| -------------------------------------- | ------------------------------------------------------------------------------ |
| `token.Position` (Go stdlib)           | `Position{File: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}` |
| `go/analysis` `token.Pos`              | Requires `pass.Fset.Position(pos)` then as above                               |
| branching-flow `SourceLocation`        | `Position{File: l.FilePath(), Line: l.Line(), Column: l.Column()}`             |
| art-dupl `LineNumber` + `BytePosition` | `Position{File: f, Line: int(ln), Offset: int(bp)}`                            |
| LSP `Range` (0-based)                  | `Line+1, Column+1` when converting from LSP                                    |
| SARIF `Region` (1-based)               | Direct mapping                                                                 |

### Finding

```go
type Finding struct {
    // Identity
    ID       string `json:"id"`                // Stable, unique identifier (e.g., "branching-flow:STRONG_ID:pkg/types.go:42:5")
    Rule     string `json:"rule"`              // Rule/check name (e.g., "STRONG_ID", "clone-detected", "silent-swallow")
    ToolName string `json:"toolName"`          // Source tool name (e.g., "branching-flow", "art-dupl")

    // Core
    Message  string   `json:"message"`          // Human-readable description
    Severity Severity `json:"severity"`         // info, warning, error, critical
    Position Position `json:"position"`         // Where the issue is

    // Classification
    Category  string `json:"category,omitempty"`  // Domain: "security", "style", "duplication", "error-handling", etc.
    Tag       string `json:"tag,omitempty"`       // Sub-classification: "phantom-type", "bool-blindness", "clone", etc.

    // Fix
    FixStrategy FixStrategy `json:"fixStrategy"`              // none, suggest, direct, ai
    Suggestion  string      `json:"suggestion,omitempty"`     // Human-readable fix description
    BeforeCode  string      `json:"beforeCode,omitempty"`     // Code before the fix
    AfterCode   string      `json:"afterCode,omitempty"`      // Code after the fix (for direct fixes, this IS the fix)

    // Context
    Range       Range           `json:"range,omitempty"`       // For span-based findings (clones, selections)
    Snippet     string          `json:"snippet,omitempty"`     // Surrounding code context
    Confidence  float64         `json:"confidence,omitempty"`  // 0.0-1.0 (art-dupl, branching-flow composition)
    Related     []RelatedRef    `json:"related,omitempty"`     // Related findings (clone groups, error flows)
    Suppression *Suppression    `json:"suppression,omitempty"` // If suppressed, why and how

    // Extensibility
    Metadata map[string]string `json:"metadata,omitempty"` // Tool-specific key-value pairs
}
```

### RelatedRef

For linking findings (clone groups, error flows, duplicate types):

```go
type RelatedRef struct {
    FindingID string   `json:"findingId"`           // ID of the related finding
    Relation  string   `json:"relation"`            // "clone-of", "wraps", "duplicates", "causes", "fixes"
    Position  Position `json:"position,omitempty"`  // Quick access to the related location
}
```

### Report

Top-level container for a tool run:

```go
type Report struct {
    Tool     ToolInfo  `json:"tool"`
    Findings []Finding `json:"findings"`
    Summary  Summary   `json:"summary"`
}

type ToolInfo struct {
    Name    string `json:"name"`
    Version string `json:"version,omitempty"`
}

type Summary struct {
    Total          int                 `json:"total"`
    BySeverity     map[Severity]int    `json:"bySeverity"`
    ByCategory     map[string]int      `json:"byCategory,omitempty"`
    ByFixStrategy  map[FixStrategy]int `json:"byFixStrategy,omitempty"`
    FilesAffected  int                 `json:"filesAffected,omitempty"`
    DurationMs     int64               `json:"durationMs,omitempty"`
    Suppressed     int                 `json:"suppressed,omitempty"`
}
```

---

## Cross-Tool Conversion

Converters live **in each tool**, not in the SDK. Tools depend on the SDK; the SDK depends on nothing.

```
finding (SDK)              artdupl (tool)
├── finding.go             ├── finding.go  ← func ToFindings(result) []finding.Finding
├── severity.go            └── ...
├── fix_strategy.go
├── ...
```

Each tool adds a single file that converts its native types to `finding.Finding`:

### art-dupl → Finding

```go
func ToFindings(result *artdupl.Result) []finding.Finding {
    var out []finding.Finding
    for _, group := range result.CloneGroups {
        for _, clone := range group.Clones {
            out = append(out, finding.Finding{
                ID:          fmt.Sprintf("art-dupl:clone:%s:%d", clone.Filename, clone.StartLine),
                Rule:        "clone-detected",
                ToolName:    "art-dupl",
                Message:     fmt.Sprintf("Duplicate code (%d tokens)", group.Size),
                Severity:    cloneSeverity(group.Severity),
                Position:    finding.Position{File: clone.Filename, Line: clone.StartLine},
                Category:    finding.CategoryDuplication,
                Tag:         "clone",
                FixStrategy: finding.FixStrategySuggest,
                Range: finding.Range{
                    Start: finding.Position{File: clone.Filename, Line: clone.StartLine},
                    End:   finding.Position{File: clone.Filename, Line: clone.EndLine},
                },
                Related:  cloneRelated(group, clone),
                Metadata: map[string]string{"hash": clone.Hash, "tokens": fmt.Sprint(group.Size)},
            })
        }
    }
    return out
}
```

### branching-flow → Finding

```go
func ToFindings(violations []core.StrongIDViolation) []finding.Finding {
    var out []finding.Finding
    for _, v := range violations {
        out = append(out, finding.Finding{
            ID:          fmt.Sprintf("branching-flow:STRONG_ID:%s:%d:%d", v.Location.FilePath(), v.Location.Line(), v.Location.Column()),
            Rule:        "STRONG_ID",
            ToolName:    "branching-flow",
            Message:     v.Message,
            Severity:    bfSeverity(v.Severity),
            Position:    finding.Position{File: v.Location.FilePath(), Line: v.Location.Line(), Column: v.Location.Column()},
            Category:    finding.CategoryTypeSafety,
            Tag:         "phantom-type",
            FixStrategy: finding.FixStrategySuggest,
            BeforeCode:  v.Suggestion.BeforeCode,
            AfterCode:   v.Suggestion.AfterCode,
        })
    }
    return out
}
```

### go/analysis.Diagnostic → Finding (SDK-provided)

Since the `go/analysis` framework is stdlib-adjacent, the SDK provides this converter:

```go
func FromDiagnostic(d analysis.Diagnostic, fset *token.FileSet, toolName string) Finding {
    pos := fset.Position(d.Pos)
    fs := FixStrategyNone
    if len(d.SuggestedFixes) > 0 {
        fs = FixStrategyDirect
    }
    return Finding{
        ID:          fmt.Sprintf("%s:%s:%s:%d:%d", toolName, d.Code, pos.Filename, pos.Line, pos.Column),
        Rule:        d.Code,
        ToolName:    toolName,
        Message:     d.Message,
        Severity:    SeverityWarning,
        Position:    Position{File: pos.Filename, Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
        Category:    d.Category,
        FixStrategy: fs,
    }
}
```

---

## SARIF Mapping

Every `Finding` maps directly to SARIF 2.1.0. The SDK generates SARIF from `Report`:

| Finding field            | SARIF path                                                               |
| ------------------------ | ------------------------------------------------------------------------ |
| `Rule`                   | `result.ruleId`                                                          |
| `Message`                | `result.message.text`                                                    |
| `Severity` → SARIF level | `result.level` (info→note, warning→warning, error→error, critical→error) |
| `Position.File`          | `result.locations[0].physicalLocation.artifactLocation.uri`              |
| `Position.Line`          | `result.locations[0].physicalLocation.region.startLine`                  |
| `Position.Column`        | `result.locations[0].physicalLocation.region.startColumn`                |
| `Range.End.Line`         | `result.locations[0].physicalLocation.region.endLine`                    |
| `Range.End.Column`       | `result.locations[0].physicalLocation.region.endColumn`                  |
| `Suggestion`             | `result.fixes[0].description.text`                                       |
| `BeforeCode`/`AfterCode` | `result.fixes[0].artifactChanges[0].replacements[0]`                     |
| `Related`                | `result.relatedLocations`                                                |
| `Metadata`               | `result.properties`                                                      |
| `Confidence`             | `result.rank` (0.0-100.0)                                                |
| `ToolName`/`Version`     | `run.tool.driver.name`/`version`                                         |
| `Category`               | `rule.properties.category`                                               |
| `Suppression`            | `result.suppressions`                                                    |

---

## Standard Categories

```go
const (
    CategorySecurity      = "security"
    CategoryStyle         = "style"
    CategoryPerformance   = "performance"
    CategoryCorrectness   = "correctness"
    CategoryComplexity    = "complexity"
    CategoryDuplication   = "duplication"
    CategoryErrorHandling = "error-handling"
    CategoryMigration     = "migration"
    CategoryTypeSafety    = "type-safety"
    CategoryStructure     = "structure"
    CategoryConfiguration = "configuration"
    CategoryDocumentation = "documentation"
    CategoryTesting       = "testing"
)
```

---

## File Structure

```
finding/
├── finding.go          # Finding, Position, Range, RelatedRef types
├── severity.go         # Severity enum, validation, ordering
├── fix_strategy.go     # FixStrategy enum
├── suppression.go      # Suppression, SuppressionKind types
├── report.go           # Report, ToolInfo, Summary types
├── category.go         # Standard category constants
├── id.go               # Finding ID generation (tool:rule:file:line:col)
├── filter.go           # Query/filter (BySeverity, ByCategory, ByFixStrategy, etc.)
├── merge.go            # Merge multiple Reports (dedup, correlate)
├── sarif.go            # Report → SARIF 2.1.0 conversion
├── diagnostic.go       # go/analysis.Diagnostic → Finding converter
├── lsp.go              # Finding → LSP Diagnostic conversion
├── json.go             # JSON marshal/unmarshal helpers
└── finding_test.go
```

No `converters/` directory. Converters live in each tool as a single file.

---

## Integration Points

### How Each Tool Would Use This

| Tool                             | Integration                                                                              | Migration Cost                                                                                                |
| -------------------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **art-dupl**                     | Add `finding.go` with `ToFindings()`; output `Report` as JSON alongside existing formats | Low: additive, no existing types changed                                                                      |
| **branching-flow**               | Add `finding.go` with `ToFindings()`; add `--format finding` flag                        | Low: additive, existing output unchanged                                                                      |
| **BuildFlow**                    | Add `Finding` as alternative to `PrioritizedViolation`; adapters can output either       | Medium: `ValidationResult[T PrioritizedViolation]` is wired into 40+ steps, don't rip out — add parallel path |
| **go-auto-upgrade**              | Add `finding.go` with `ToFindings()`; emit `Report` JSON                                 | Low: additive                                                                                                 |
| **hierarchical-errors**          | Add `finding.go` with `ToFindings()`; use `Finding` in LSP server                        | Medium: LSP server already has its own mapping, but `Finding` → LSP Diagnostic is provided                    |
| **golangci-lint-auto-configure** | Add `finding.go` with `ToFindings()` for recommendations                                 | Low: additive                                                                                                 |
| **rules**                        | Future: custom analyzers emit `Finding` via `FromDiagnostic()`                           | Low: just adds output format                                                                                  |

### Consumer Use Cases

1. **Unified report** — Run all tools, merge `Report`s, output single SARIF for GitHub Code Scanning
2. **Auto-fix pipeline** — Filter `FixStrategy == "direct"`, apply, re-run until stable
3. **AI fix pipeline** — Filter `FixStrategy == "suggest"`, send to AI with `BeforeCode`/`AfterCode` context, apply
4. **Quality dashboard** — Aggregate `Summary` across tools, track trends over time
5. **IDE integration** — Consume `Report` via LSP, show all-tool diagnostics in one view
6. **Cross-tool dedup** — Merge reports, detect same-location findings from different tools, surface root cause

---

## Open Questions

1. **Repository name**: `finding`? `finding-sdk`? `go-finding`?
2. **Stable ID format**: Should IDs be deterministic hashes or readable strings? Proposal uses readable `"tool:rule:file:line:col"` but hash-based IDs might be more robust for deduplication
3. **Breaking change from existing types**: Should BuildFlow replace `PrioritizedViolation` or add `Finding` alongside it? Recommendation: **add alongside** — don't rip out wired types
4. **AI fix metadata**: Should `FixStrategyAI` have additional fields (model, prompt template, context window)? Recommendation: start with `Metadata` key-value, add structured fields when the pipeline is real
5. **Suppression expiry**: Should temporary suppressions expire? Useful for "ignore for now, revisit in 30 days" but adds complexity. Recommendation: include the field, start without enforcement
6. **go-business-rules relationship**: This project has `Violation` with `Severity`. `Finding` is for static analysis; `Violation` is for runtime validation. They share severity semantics but serve different loops. Should they share the same `Severity` type? Recommendation: yes — extract `Severity` to a shared package if both projects import it

---

## Implementation Roadmap

### Phase 0: Prove the Pipeline Exists (1 day)

Before building anything, validate the concept with a throwaway script:

```bash
# Run 3 tools, parse JSON output, apply direct fixes, re-run, report what's left
./prove-pipeline.sh ./my-project
```

If this loop is useful manually, the SDK is worth building. If not, stop here.

### Phase 1: Core Types (1-2 days)

- `finding.go`, `severity.go`, `fix_strategy.go`, `suppression.go`, `report.go`, `category.go`, `id.go`
- `filter.go` (query helpers)
- `json.go` (serialization)
- `diagnostic.go` (go/analysis converter)
- Tests

### Phase 2: Output Formats (1 day)

- `sarif.go` (Report → SARIF 2.1.0)
- `lsp.go` (Finding → LSP Diagnostic)
- Tests with SARIF schema validation

### Phase 3: Pipeline (2-3 days)

- `merge.go` (dedup, correlate across tools)
- Pipeline engine: detect → triage → fix → verify loop
- Direct fix application (write `AfterCode` to files)
- AI fix routing (structured prompt from `Finding` fields)

### Phase 4: Tool Integration (2-3 days per tool)

Each tool adds a single `finding.go` file with `ToFindings()` converter.
Start with the tools that have the simplest types: go-auto-upgrade, art-dupl, branching-flow.

### Phase 5: BuildFlow Integration (3-5 days)

BuildFlow is the hardest because `ValidationResult[T PrioritizedViolation]` is deeply wired.
Strategy: add `Finding` as a parallel output path, don't replace `PrioritizedViolation`.
Over time, migrate steps to produce `Finding` natively.
