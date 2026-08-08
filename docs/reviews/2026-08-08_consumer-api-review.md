# Consumer API Review: go-humanize-linter + go-linter-sdk

**Date:** 2026-08-08
**Scope:** How `go-humanize-linter` and `go-linter-sdk` use `go-finding`, and what improvements would let them do the same or more with less code.

---

## Summary

go-humanize-linter is a 9-rule linter built on go-linter-sdk (v0.1.0) + go-finding (v1.4.1). go-linter-sdk is a thin abstraction layer (3 files, ~240 LOC) that eliminates the converter layer by having rules emit `finding.Finding` directly.

Both projects work well today, but they reinvent several patterns that belong in the upstream libraries. This review identifies **8 concrete improvements** (3 already implemented in go-finding, 5 proposed for go-linter-sdk) that would eliminate ~200 LOC of consumer boilerplate.

---

## What go-finding Already Provides (but consumers don't use)

### 1. `finding.Template` — partially used

The humanize-linter defines `makeFindingWithConfidence` (23 LOC, called 12+ times) which wraps `finding.NewBuilder(...)` with category + fix strategy + confidence + suggestion. go-finding already ships `Template` for exactly this, but `Template.Build()` returns a final `Finding` — callers can't add per-finding confidence or suggestion.

**Fix implemented:** Added `Template.Builder()` which returns a pre-configured `*Builder` for chaining.

### 2. `gotoken.LineColToPos` — not discovered

The humanize-linter's `plugin/plugin.go` reimplements `gotoken.LineColToPos` as `findingToTokenPos` (25 LOC). The existing `gotoken` helper already handles nil-checks, line-bounds, and column offsets.

**Action:** Consumer should use `gotoken.LineColToPos(tokenFile, f.Position.Line, f.Position.Column)` instead.

---

## Improvements Implemented in go-finding (this session)

### 3. `ParseConfidence(string) (Confidence, error)` — `confidence.go`

**Before (consumer reinvents):**

```go
// go-humanize-linter/confidence.go (14 LOC)
func ParseConfidenceLevel(level string) (finding.Confidence, error) {
    switch level {
    case "low", "":   return finding.ConfidenceLow, nil
    case "medium":    return finding.ConfidenceMedium, nil
    case "high":      return finding.ConfidenceHigh, nil
    case "full":      return finding.ConfidenceFull, nil
    }
    return finding.ConfidenceNone, ErrInvalidConfidence
}
```

**After:**

```go
c, err := finding.ParseConfidence("high") // one-liner, no consumer code
```

Also accepts decimals (`"0.42"`), is case-insensitive, trims whitespace, and defaults empty string to `ConfidenceLow`. Error is `ErrInvalidConfidence` sentinel, matchable via `errors.Is`.

**Impact:** Eliminates `confidence.go` entirely (14 LOC + test file). Every CLI linter with `--min-confidence` can delete this pattern.

### 4. `Template.Builder(rule, msg, sev, pos) *Builder` — `finding_builder.go`

**Before (consumer reinvents):**

```go
// go-humanize-linter/pattern_helpers.go (23 LOC, called 12+ times)
func makeFindingWithConfidence(ruleID, message, suggestion string, line, col int, filePath string, conf finding.Confidence) finding.Finding {
    b := finding.NewBuilder(
        finding.RuleName(ruleID),
        finding.ToolName("go-humanize-linter"),
        message,
        finding.SeverityWarning,
        finding.Pos(finding.FilePath(filePath), line, col),
    ).
        WithCategory(finding.CategoryStyle).
        WithFixStrategy(finding.FixStrategySuggest).
        WithConfidence(conf)
    if suggestion != "" {
        b = b.WithSuggestion(suggestion)
    }
    return b.MustBuild()
}
```

**After:**

```go
// One-time setup
var tmpl = finding.NewTemplate("go-humanize-linter").
    WithCategory(finding.CategoryStyle).
    WithFixStrategy(finding.FixStrategySuggest)

// Per finding (no wrapper function needed)
f := tmpl.Builder(rule, msg, finding.SeverityWarning, pos).
    WithConfidence(conf).
    WithSuggestion(suggestion).
    MustBuild()
```

**Impact:** Eliminates `makeFindingWithConfidence` entirely (23 LOC + calls become 1 line shorter). `Build` delegates to `Builder().BuildOrDefault()` for backward compat.

---

## Proposed Improvements for go-linter-sdk

### 5. Registry should accept a tool name (HIGH IMPACT)

**Problem:** `Registry.Run()` hardcodes the tool name as `"linter"` (`registry.go:158`). Findings are attributed to a generic "linter" tool, not the actual linter. The humanize-linter works around this by constructing findings with the correct tool name in every `NewBuilder` call, but the report metadata (`finding.ToolInfo`) is wrong.

**Proposed:**

```go
// Option A: Constructor option
registry := linter.NewRegistry(linter.WithToolName("go-humanize-linter"))

// Option B: Run option
report, err := registry.Run(ctx, dir, linter.WithToolInfo(finding.ToolInfo{
    Name:    "go-humanize-linter",
    Version: version,
}))
```

**Impact:** Correct tool attribution in reports and SARIF output. Enables the SDK to stamp findings with the correct tool name (see #6).

### 6. Rule should provide a finding builder factory (HIGHEST IMPACT)

**Problem:** Every rule in humanize-linter repeats the rule ID, tool name, severity, and category in every `finding.NewBuilder(...)` call — even though all four are already in `RuleMeta`. The SDK has this metadata but doesn't use it to help build findings.

**Proposed:**

```go
// On RuleFunc or RuleMeta:
func (r RuleFunc) NewFinding(message string, pos finding.Position) *finding.Builder {
    return finding.NewBuilder(
        finding.RuleName(r.Meta.ID),
        toolName,  // from registry or RuleMeta
        message,
        r.Meta.Sev,
        pos,
    ).
        WithCategory(finding.Category(r.Meta.Cat))
}
```

**Impact:** Rules emit findings without repeating identity fields:

```go
Run: func(ctx context.Context, dir string) ([]finding.Finding, error) {
    // ...
    return []finding.Finding{
        rule.NewFinding("manual byte formatting", pos).
            WithConfidence(finding.ConfidenceHigh).
            WithSuggestion("use humanize.Bytes").
            MustBuild(),
    }, nil
}
```

This is the **single biggest boilerplate eliminator** — it would remove the tool name constant, severity, category, and rule ID from every builder call across all 9 rules.

### 7. Enable/disable rule filtering helper (MEDIUM IMPACT)

**Problem:** Two near-identical implementations exist in the humanize-linter:

- `buildRegistry()` in `cmd/go-humanize-linter/main.go:425-455` (30 LOC)
- `filterRules()` in `plugin/plugin.go:261-281` (20 LOC)

Both iterate over `AllRules()`, check `disable[rule.Meta.ID]` and `enable[rule.Meta.ID]`, with identical logic.

**Proposed:**

```go
// In go-linter-sdk:
func (r *Registry) Filter(enableIDs, disableIDs []string) []Rule
// or:
func FilterRules(all []RuleFunc, enable, disable map[string]bool) []RuleFunc
```

**Impact:** Eliminates ~50 LOC of duplicate filtering logic. Every linter with `--enable`/`--disable` flags needs this.

### 8. Confidence-aware exit code (MEDIUM IMPACT)

**Problem:** The SDK provides `ExitCodeFromReport()` (binary: 0 clean / 1 any finding). The humanize-linter needs tiered exit codes (0 clean / 1 high+full / 2 medium+low) for CI, so it reimplements this as `exitCodeFromReport()` in `main.go:494-512` (18 LOC).

**Proposed:**

```go
// Option A: Dedicated function
func ExitCodeByConfidence(report *finding.Report, threshold finding.Confidence) int

// Option B: Configurable
func ExitCodeFromReportWithOpts(report *finding.Report, opts ...ExitOption) int
```

**Impact:** Eliminates `exitCodeFromReport` in the consumer (18 LOC). Any confidence-aware linter would benefit.

### 9. `IsEnabledByDefault` should be consulted at runtime (LOW IMPACT, design change)

**Problem:** `IsEnabledByDefault()` is declared on the `Rule` interface and `OptIn()` creates disabled-by-default rules, but **neither `Registry.Run` nor `DetectorsFromRegistry` consult it**. Both iterate `registry.All()` and run every rule unconditionally. The flag is purely informational.

**Proposed:** Either:

- Have `Registry.Run` skip rules where `!rule.IsEnabledByDefault()` unless explicitly enabled
- Or document clearly that `IsEnabledByDefault` is consumer-only metadata, not runtime behavior

**Impact:** Clarifies the SDK contract. If made runtime-active, eliminates consumer-side enable/disable logic for opt-in rules.

---

## Quantified Impact

| Improvement               | Where         | LOC eliminated per consumer                     |
| ------------------------- | ------------- | ----------------------------------------------- |
| `ParseConfidence` (done)  | go-finding    | ~14 LOC + tests                                 |
| `Template.Builder` (done) | go-finding    | ~23 LOC factory + 12 call sites shortened       |
| Tool name in Registry     | go-linter-sdk | Correct attribution + enables #6                |
| Rule.NewFinding factory   | go-linter-sdk | ~5 LOC × 9 rules = 45 LOC                       |
| Filter helper             | go-linter-sdk | ~50 LOC (two duplicates)                        |
| Confidence exit code      | go-linter-sdk | ~18 LOC                                         |
| **Total**                 |               | **~150 LOC eliminated from go-humanize-linter** |

---

## What NOT to Change

- **Branded types** (`FilePath`, `RuleName`, `ToolName`) — the verbosity is intentional compile-time safety. The `string()` conversions in the consumer are the right tradeoff.
- **`gotoken` stdlib-only design** — correctly keeps go/token coupling in a separate package. Adding `finding.Position` here would break the boundary.
- **Registry sequential Run** — the parallel path (`DetectorsFromRegistry`) already exists. Adding concurrency to `Run` would duplicate the pipeline's value proposition.
- **`GOEXPERIMENT=jsonv2`** — this is a Go ecosystem constraint, not a go-finding decision. It goes away when Go stabilizes json/v2.
