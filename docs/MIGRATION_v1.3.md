# Migrating to go-finding v1.3.0

This guide helps consumers upgrade from v1.2.x to v1.3.0.

**Summary:** v1.3.0 adds 12 consumer-driven convenience APIs based on a full audit of 22 consumer projects. All additions are purely additive — no breaking changes. Consumers don't _need_ to change anything, but adopting these APIs eliminates boilerplate patterns that every consumer reinvents.

---

## Quick Reference: What Changed

| New API                      | Replaces                                             | Section                                           |
| ---------------------------- | ---------------------------------------------------- | ------------------------------------------------- |
| `Builder.BuildOrDefault()`   | `Build()` + error-swallowing boilerplate             | [§1](#1-buildorbuild-eliminates-error-swallowing) |
| `Template` factory           | Per-finding builder configuration                    | [§2](#2-template-factory)                         |
| `NewReportFromFindings()`    | `NewReport()` + `AddFindings()` + `ComputeSummary()` | [§3](#3-newreportfromfindings)                    |
| `FilePos()`                  | `Position{File: ..., Line: 0}`                       | [§4](#4-filepos-file-level-positions)             |
| `SeverityFromLevel()`        | Consumer-side `mapSeverity()` switches               | [§5](#5-severityfromlevel)                        |
| `ApplySimpleFixes()`         | Manual file-read + `strings.Replace` + file-write    | [§6](#6-applysimplefixes)                         |
| `CheckBinary()` / `RunCmd()` | Manual `exec.LookPath` + `exec.CommandContext`       | [§7](#7-checkbinary--runcmd)                      |
| `FormatTextRich()`           | Emoji-badged output                                  | [§8](#8-formattextrich)                           |
| `FormatTable()`              | Manual `text/tabwriter` formatting                   | [§9](#9-formattable)                              |
| `Severity.PriorityString()`  | Consumer-side severity→level switches                | [§10](#10-severityprioritystring)                 |

---

## 1. BuildOrDefault: Eliminates Error Swallowing

Many consumers swallow `Build()` errors by returning a zero-value Finding. `BuildOrDefault()` makes this explicit.

**Before (v1.2.x):**

```go
f, err := builder.Build()
if err != nil {
    // Swallow error, use zero-value
    f = finding.Finding{}
}
```

**After (v1.3.0):**

```go
f := builder.BuildOrDefault()
```

Returns zero-value `Finding{}` on validation error. Use `Build()` when you need the error.

---

## 2. Template Factory

Pre-configures common builder fields so you can stamp them once and build many findings.

**Before (v1.2.x):**

```go
// Each finding repeats the same tool/category/strategy
f1 := finding.NewBuilder("rule1", "my-linter", "msg1", finding.SeverityError, pos1).
    WithCategory(finding.CategoryCorrectness).
    WithFixStrategy(finding.FixStrategyDirect).
    WithTags(finding.TagStyle).
    BuildOrDefault()

f2 := finding.NewBuilder("rule2", "my-linter", "msg2", finding.SeverityWarning, pos2).
    WithCategory(finding.CategoryCorrectness).
    WithFixStrategy(finding.FixStrategyDirect).
    WithTags(finding.TagStyle).
    BuildOrDefault()
```

**After (v1.3.0):**

```go
tmpl := finding.NewTemplate("my-linter").
    WithCategory(finding.CategoryCorrectness).
    WithFixStrategy(finding.FixStrategyDirect).
    WithTags(finding.TagStyle)

f1 := tmpl.Build("rule1", "msg1", finding.SeverityError, pos1)
f2 := tmpl.Build("rule2", "msg2", finding.SeverityWarning, pos2)
```

---

## 3. NewReportFromFindings

One-step report creation instead of 4 lines.

**Before (v1.2.x):**

```go
report := finding.NewReport("my-tool")
report.AddFindings(findings)
report.ComputeSummary()
```

**After (v1.3.0):**

```go
report := finding.NewReportFromFindings("my-tool", findings)
```

---

## 4. FilePos: File-Level Positions

File-level findings (no line/column) now have an explicit constructor.

**Before (v1.2.x):**

```go
pos := finding.Position{File: "go.mod"} // Line=0, Column=0
```

**After (v1.3.0):**

```go
pos := finding.FilePos("go.mod")
```

Both produce the same result, but `FilePos` is explicit about intent. `validateIdentity()` now uses `Position.HasFile()` (File != "") instead of `Position.IsValid()` (File != "" && Line > 0), so file-level findings pass validation.

---

## 5. SeverityFromLevel

Maps severity strings (canonical + aliases) to `Severity`. Eliminates consumer-side severity switches.

**Before (v1.2.x):**

```go
func mapSeverity(level string) finding.Severity {
    switch strings.ToLower(level) {
    case "error", "high", "fatal":
        return finding.SeverityError
    case "warning", "medium", "warn":
        return finding.SeverityWarning
    case "info", "low":
        return finding.SeverityInfo
    default:
        return finding.SeverityWarning // fallback
    }
}
```

**After (v1.3.0):**

```go
sev := finding.SeverityFromLevel(level, finding.SeverityWarning)
```

Supports canonical names (`error`, `warning`, `info`, `critical`) and aliases (`high`, `medium`, `low`, `fatal`, `crit`, `optional`).

---

## 6. ApplySimpleFixes

BeforeCode→AfterCode string replacement for the 80% case where consumers don't need the full pipeline FixEngine.

**Before (v1.2.x):**

```go
for _, f := range findings {
    if f.BeforeCode == "" || f.AfterCode == "" {
        continue
    }
    content, err := os.ReadFile(string(f.Position.File))
    if err != nil { return err }
    newContent := strings.Replace(content, f.BeforeCode, f.AfterCode, 1)
    err = os.WriteFile(string(f.Position.File), newContent, 0644)
    if err != nil { return err }
}
```

**After (v1.3.0):**

```go
results := finding.ApplySimpleFixes(findings)
for file, fileResults := range results {
    for _, r := range fileResults {
        if r.Applied { /* success */ }
    }
}
```

No pipeline import needed. Lives in the core `finding` package. See [Fix Engine Guide](guides/fix-engine.md#simple-fixes-core-package).

---

## 7. CheckBinary / RunCmd

External tool helpers for the "run CLI tool → parse JSON" pattern.

**Before (v1.2.x):**

```go
path, err := exec.LookPath("golangci-lint")
if err != nil {
    return fmt.Errorf("golangci-lint not found: %w", err)
}
cmd := exec.CommandContext(ctx, path, "run", "--json", "./...")
output, err := cmd.Output()
if err != nil {
    return fmt.Errorf("run golangci-lint: %w", err)
}
```

**After (v1.3.0):**

```go
if err := finding.CheckBinary("golangci-lint"); err != nil {
    return err
}
output, err := finding.RunCmd(ctx, "golangci-lint", "run", "--json", "./...")
if err != nil {
    return err
}
```

`CheckBinary` returns `NewIOError` on failure. `RunCmd` accepts context and returns output.

---

## 8. FormatTextRich

Rich text formatter with emoji severity badges and category display.

**Before (v1.2.x):**

```go
finding.FormatText(w, findings)
// Output: [WARNING] main.go:10:5 — unused variable
```

**After (v1.3.0):**

```go
finding.FormatTextRich(w, findings)
// Output: ⚠️ WARNING — main.go:10:5 — unused variable (Correctness)
```

`FormatText` is retained for backward compatibility.

---

## 9. FormatTable

Severity-badged table output.

**Before (v1.2.x):**

```go
w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
for _, f := range findings {
    fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.Severity, f.Position, f.Rule, f.Message)
}
w.Flush()
```

**After (v1.3.0):**

```go
finding.FormatTable(os.Stdout, findings)
// Columns: SEVERITY, LOCATION, RULE, MESSAGE
```

---

## 10. Severity.PriorityString

Reverse mapping: Severity → priority string for SARIF/SARIF-like output.

**Before (v1.2.x):**

```go
func severityToLevel(sev finding.Severity) string {
    switch sev {
    case finding.SeverityCritical: return "critical"
    case finding.SeverityError: return "high"
    case finding.SeverityWarning: return "medium"
    case finding.SeverityInfo: return "low"
    default: return "medium"
    }
}
```

**After (v1.3.0):**

```go
level := sev.PriorityString()
```

---

## No Breaking Changes

v1.3.0 is purely additive. All existing v1.2.x code compiles and works unchanged. The convenience APIs are opt-in — adopt them incrementally as you refactor consumer code.
