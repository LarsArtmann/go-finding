# Migrating to go-finding v1.0.0

This guide helps consumers upgrade from v0.7.x to v1.0.0.

**Summary:** v1.0.0 removes deprecated APIs and resolves zero-value ambiguities.
Most consumers need only a few mechanical changes.

---

## 1. Report.Findings → FindingsSnapshot()

The public `Report.Findings` slice is now unexported. Direct field access no longer
compiles.

**Before (v0.7.x):**

```go
for _, f := range report.Findings {
    fmt.Println(f)
}
```

**After (v1.0.0):**

```go
for _, f := range report.FindingsSnapshot() { // deep-cloned, thread-safe
    fmt.Println(f)
}

// Or use the iterator:
for f := range report.All() {
    fmt.Println(f)
}
```

| Old API                | New API                     | Notes                                      |
| ---------------------- | --------------------------- | ------------------------------------------ |
| `report.Findings`      | `report.FindingsSnapshot()` | Deep-cloned slice, safe for concurrent use |
| `report.Findings[i]`   | `report.FindByID(id)`       | Returns a copy (`*Finding`)                |
| `len(report.Findings)` | `report.Len()`              | —                                          |

**Why:** `Findings` was a public slice that bypassed the mutex, risking data races.
The new APIs are thread-safe by design.

---

## 2. Report.Merge() → Report.MergeInto()

`Merge()` mutated the receiver in place. `MergeInto()` returns a new Report without
modifying either input.

**Before:**

```go
report.Merge(other) // mutates report
```

**After:**

```go
merged := report.MergeInto(other) // both inputs unchanged
```

---

## 3. OnStage → StageHooks

The single `OnStage` callback is replaced by the `StageHooks` slice, which supports
before/after events and abort capability.

**Before:**

```go
config.OnStage = func(stage pipeline.Stage, iter, count int) {
    log.Printf("stage %s: %d findings", stage, count)
}
```

**After:**

```go
config.StageHooks = []pipeline.StageHook{
    pipeline.StageHookFunc(func(e pipeline.StageEvent) error {
        log.Printf("stage %s (after): %d findings", e.Stage, e.FindingsCount)
        return nil // return error to abort the pipeline
    }),
}
```

---

## 4. Metrics.RecordFix() → Metrics.RecordFixes(1)

**Before:**

```go
metrics.RecordFix()
```

**After:**

```go
metrics.RecordFixes(1)
```

---

## 5. CountBySeverity() → Report.CountBySeverity()

The free function is removed. Use the method on Report.

**Before:**

```go
counts := finding.CountBySeverity(findings)
```

**After:**

```go
counts := report.CountBySeverity()
```

---

## 6. Position Zero-Value Semantics (OWNER DECISION PENDING)

v1.0.0 resolves the `Position.Offset = 0` ambiguity. The chosen approach (sentinel
`-1`, pointer, or flag field) will be documented here once decided. See
`docs/RELEASE_CRITERIA.md` Blocker 1.

**Current behavior (v0.7.x):** `Offset = 0` is both "byte 0" and "unset".
`HasOffset()` returns `true` for both.

---

## 7. Branded Primitive Types

Identity fields now use branded types that prevent accidental mixing at compile time:

| Field                    | Old Type   | New Type   |
| ------------------------ | ---------- | ---------- |
| `Finding.ID`             | `string`   | `ID`       |
| `Finding.Rule`           | `string`   | `RuleName` |
| `Finding.ToolName`       | `string`   | `ToolName` |
| `Position.File`          | `string`   | `FilePath` |
| `FindingError.File`      | `string`   | `FilePath` |
| `FixGroup.File`          | `string`   | `FilePath` |
| `Suppression.Rule`       | `string`   | `RuleName` |
| `RelatedRef.FindingID`   | `string`   | `ID`       |
| `Correlation.FindingIDs` | `[]string` | `[]ID`     |
| `ParsedID.Tool`          | `string`   | `ToolName` |
| `ParsedID.Rule`          | `string`   | `RuleName` |
| `ParsedID.File`          | `string`   | `FilePath` |

**Before:**

```go
f, err := finding.NewBuilder("nilcheck", "govet", "msg", finding.SeverityError, pos).Build()
```

**After:**

```go
f, err := finding.NewBuilder(
    finding.RuleName("nilcheck"), finding.ToolName("govet"), "msg", finding.SeverityError, pos,
).Build()
```

### Position.File → FilePath

`Position.File` is now `FilePath` (a named `string` type). This prevents accidentally assigning an ID, rule name, or tool name to a file path field.

**String literals auto-convert** — no change needed:

```go
pos := finding.Position{File: "main.go", Line: 10} // still compiles
```

**String variables need wrapping:**

```go
path := getPath()
pos := finding.Position{File: finding.FilePath(path), Line: 10}
```

**Constructors accept `FilePath`:**

```go
pos := finding.Pos(finding.FilePath("main.go"), 10, 5)
r := finding.NewRange(finding.FilePath("main.go"), 1, 1, 5, 10)
```

**`ByFile` filter accepts `FilePath`:**

```go
filtered := finding.Filter(findings, finding.ByFile(finding.FilePath("main.go")))
```

**`GroupByFile` returns `map[FilePath][]Finding`:**

```go
groups := finding.GroupByFile(findings) // map[FilePath][]Finding
for file, fs := range groups {
    fmt.Println(string(file), len(fs)) // cast only for non-string APIs
}
```

**`FromLSP` accepts `FilePath`:**

```go
f := finding.FromLSP(finding.FilePath("file:///main.go"), diag)
```

JSON serialization is identical (marshals as plain string). The type named `ID` (not `FindingID`) avoids the `finding.FindingID` stutter.

---

## 8. FindingProcessor → FindingTransformer

Renamed for clarity. `Process()` is now `Transform()`.

| Old API              | New API                                                      |
| -------------------- | ------------------------------------------------------------ |
| `FindingProcessor`   | `FindingTransformer`                                         |
| `ProcessorFunc`      | `TransformerFunc`                                            |
| `NamedProcessorFunc` | `NamedTransformerFunc`                                       |
| `.Process(ctx, fs)`  | `.Transform(ctx, fs)`                                        |
| `Config.Processors`  | `Config.Processors` (type changed to `[]FindingTransformer`) |

---

## 9. Other API Renames

| Old API             | New API                            | Location               |
| ------------------- | ---------------------------------- | ---------------------- |
| `GetCategory(err)`  | `CategoryOf(err)`                  | `errors.go`            |
| `ConflictInfo`      | `Conflict`                         | `pipeline/conflict.go` |
| `LSPRelatedInfo`    | `LSPRelated`                       | `lsp.go`               |
| `SeverityAliases()` | `LookupSeverityAlias(name)`        | `severity.go`          |
| (direct map edit)   | `RegisterSeverityAlias(name, sev)` | `severity.go`          |

`GetCategory()`, `ConflictInfo`, and `LSPRelatedInfo` are kept as deprecated wrappers until v1.0.0 final.

---

## 10. FixStrategy Normalization (OWNER DECISION PENDING)

v1.0.0 will normalize the empty string `""` to `FixStrategyNone` (`"none"`). See
`docs/RELEASE_CRITERIA.md` Blocker 2.

---

## 11. SARIF Export API (v1.1.0)

`ToSARIFFiltered` and `WriteSARIFFiltered` are removed in v1.1.0. Use the functional
options pattern instead:

**Before:**

```go
data, err := report.ToSARIFFiltered(finding.SeverityWarning)
err = report.WriteSARIFFiltered(ctx, w, finding.SeverityWarning)
```

**After:**

```go
data, err := report.ToSARIFWithOpts(finding.WithMinSeverity(finding.SeverityWarning))
err = report.WriteSARIFWithOpts(ctx, w, finding.WithMinSeverity(finding.SeverityWarning))

// Include suppressed findings in SARIF output:
data, err := report.ToSARIFWithOpts(
    finding.WithIncludeSuppressed(),
    finding.WithMinSeverity(finding.SeverityWarning),
)
```

---

## 12. CLI Flag Rename (v1.1.0)

`-severity` renamed to `-min-severity` for honesty (it's a minimum, not an exact match).
Deprecated alias `-severity` retained.

---

## 13. Multi-Module Workspace (v1.1.0)

The project is now split into 4 independently versioned Go modules:

| Module   | Path                                               | Dependencies        |
| -------- | -------------------------------------------------- | ------------------- |
| Core     | `github.com/larsartmann/go-finding`                | stdlib only         |
| Pipeline | `github.com/larsartmann/go-finding/pipeline`       | x/sync, gogenfilter |
| Analysis | `github.com/larsartmann/go-finding/analysis`       | x/tools             |
| CLI      | `github.com/larsartmann/go-finding/cmd/go-finding` | yaml, go-output     |

Consumers import only what they need. The core module has zero external production dependencies.

---

## Migration Checklist

- [ ] Replace all `report.Findings` access with `FindingsSnapshot()`, `All()`, or `FindByID()`
- [ ] Replace `report.Merge(other)` with `merged := report.MergeInto(other)`
- [ ] Replace `config.OnStage` with `config.StageHooks`
- [ ] Replace `metrics.RecordFix()` with `metrics.RecordFixes(1)`
- [ ] Replace `finding.CountBySeverity(findings)` with `report.CountBySeverity()`
- [ ] Add branded type conversions: `finding.RuleName("x")`, `finding.ToolName("x")`, `finding.ID("x")`
- [ ] Replace `FindingProcessor` with `FindingTransformer`, `.Process()` with `.Transform()`
- [ ] Replace `GetCategory(err)` with `CategoryOf(err)`
- [ ] Replace `ConflictInfo` with `Conflict`
- [ ] Replace `LSPRelatedInfo` with `LSPRelated`
- [ ] Replace direct `SeverityAliases` map edits with `RegisterSeverityAlias()`
- [ ] Run `go test ./...` — all tests should pass after these changes

---

## Need Help?

- Check `docs/API_STABILITY.md` for per-symbol stability classification
- Check `docs/architecture-decisions.md` ADR #11 for the full breaking-change rationale
- Run `go vet ./...` to catch remaining deprecated API usage
