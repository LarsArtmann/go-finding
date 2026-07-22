# Consumer Migration Guide: v1.2 → v1.3

> **Audience:** Projects consuming `github.com/larsartmann/go-finding`
> **Goal:** Replace boilerplate with the 11 new v1.3.0 APIs

v1.3.0 is **additive only** — zero breaking changes to existing types or signatures. You don't need to change anything to upgrade. But you *can* simplify your code by adopting these APIs.

---

## 1. Report Creation: `NewReportFromFindings`

**Before (4 lines):**

```go
report := finding.NewReport(finding.ToolInfo{Name: "my-tool"})
report.AddFindings(findings)
report.ComputeSummary()
```

**After (1 line):**

```go
report := finding.NewReportFromFindings(finding.ToolInfo{Name: "my-tool"}, findings)
```

---

## 2. Safe Building: `BuildOrDefault`

**Before (error swallowing):**

```go
func buildFinding(rule, tool, msg string, sev finding.Severity, pos finding.Position) finding.Finding {
    f, err := finding.NewBuilder(finding.RuleName(rule), finding.ToolName(tool), msg, sev, pos).Build()
    if err != nil {
        return finding.Finding{}
    }
    return f
}
```

**After (built-in):**

```go
f := finding.NewBuilder(finding.RuleName(rule), finding.ToolName(tool), msg, sev, pos).
    BuildOrDefault()
```

---

## 3. Batch Finding Creation: `Template`

**Before (custom factory):**

```go
func newMigrationFinding(rule, msg string, pos finding.Position) finding.Finding {
    f, _ := finding.NewBuilder(finding.RuleName(rule), "migrator", msg, finding.SeverityWarning, pos).
        WithCategory(finding.CategoryStyle).
        WithFixStrategy(finding.FixStrategySuggest).
        Build()
    return f
}
```

**After (reusable template):**

```go
tmpl := finding.NewTemplate("migrator").
    WithCategory(finding.CategoryStyle).
    WithFixStrategy(finding.FixStrategySuggest)

f1 := tmpl.Build("rule-1", "msg 1", finding.SeverityWarning, finding.Pos("a.go", 10, 1))
f2 := tmpl.Build("rule-2", "msg 2", finding.SeverityInfo, finding.Pos("b.go", 20, 5))
```

---

## 4. File-Level Positions: `FilePos`

**Before (manual construction):**

```go
pos := finding.Position{File: finding.FilePath("config.yaml"), Offset: -1}
```

**After (semantic constructor):**

```go
pos := finding.FilePos("config.yaml")
```

---

## 5. Severity Mapping: `SeverityFromLevel`

**Before (switch statement):**

```go
func mapSeverity(level string) finding.Severity {
    switch strings.ToLower(level) {
    case "error", "high":
        return finding.SeverityError
    case "warn", "warning", "medium":
        return finding.SeverityWarning
    case "info", "low":
        return finding.SeverityInfo
    default:
        return finding.SeverityInfo
    }
}
```

**After (built-in with alias support):**

```go
sev := finding.SeverityFromLevel(level, finding.SeverityInfo)
```

---

## 6. Simple Fixes: `ApplySimpleFixes`

**Before (pipeline for simple replacements):**

```go
// Set up pipeline just for BeforeCode→AfterCode replacement
applier, _ := pipeline.NewFixApplier(rootDir)
// ... configure FixEngine, providers, etc.
```

**After (core package, no pipeline needed):**

```go
results := finding.ApplySimpleFixes(findingsWithDirectFixes)
for file, fileResults := range results {
    for _, r := range fileResults {
        if r.Applied {
            fmt.Printf("Fixed %s in %s\n", r.FindingID, file)
        }
    }
}
```

---

## 7. External Tool Integration: `CheckBinary` + `RunCmd`

**Before (raw exec with manual error wrapping):**

```go
path, err := exec.LookPath("golangci-lint")
if err != nil {
    return fmt.Errorf("golangci-lint not found: %w", err)
}
cmd := exec.CommandContext(ctx, path, "run", "--out-format", "json", "./...")
output, err := cmd.Output()
if err != nil {
    return fmt.Errorf("golangci-lint failed: %w", err)
}
```

**After (standardized with finding errors):**

```go
_, _ = finding.CheckBinary("golangci-lint") // wraps NewIOError
output, err := finding.RunCmd(ctx, "golangci-lint", "run", "--out-format", "json", "./...")
```

---

## 8. Rich Text Output: `FormatTextRich`

`FormatText` retains its original `[SEVERITY]` format. For emoji badges:

```go
finding.FormatTextRich(os.Stdout, findings)
// Output: main.go:42:5 🟠 ERROR  rule: message [security]
//   💡 suggestion text
```

---

## Quick Reference

| Boilerplate Pattern | v1.3.0 Replacement |
|---|---|
| NewReport + AddFindings + ComputeSummary | `NewReportFromFindings` |
| SafeBuildFinding / buildFinding wrapper | `Builder.BuildOrDefault()` |
| newMigrationFinding / IssueBuilderFactory | `NewTemplate` + `Template.Build` |
| `Position{File: f, Offset: -1}` | `FilePos(f)` |
| `mapSeverity()` switch | `SeverityFromLevel(level, fallback)` |
| Pipeline FixEngine for simple replacements | `ApplySimpleFixes(findings)` |
| `exec.LookPath` + `exec.CommandContext` | `CheckBinary` + `RunCmd` |
| `[SEVERITY]` format with emoji badges | `FormatTextRich(w, findings)` |

---

_Assisted-by: Crush <crush@charm.land>_
