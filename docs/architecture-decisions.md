# Architecture Decisions

**Status:** Open for review — no action required unless you disagree.

These decisions need product input. They are documented here for visibility.

---

## 1. FixStrategyAI Semantics

**Current behavior:** `FixStrategyAI` behaves identically to `FixStrategySuggest`:

- `HasFix()` returns true only if `AfterCode != ""`
- Pipeline triage treats it as a suggestion (no auto-apply)
- `NeedsAI()` returns true (distinguishing method)

**Question:** Should `FixStrategyAI` mean:

- **(A)** "An AI could fix this" (capability) → `HasFix()` should always return true
- **(B)** "An AI has generated a suggested fix" (artifact) → current behavior is correct

**Recommendation:** Keep (B). It matches the existing SARIF round-trip behavior and doesn't over-promise.

---

## 2. Stable ID Format

**Current:** `tool:rule:file:line:col` (human-readable)

**Options:**

- (A) Keep readable strings (current)
- (B) SHA-256 hash (collision-resistant but opaque)
- (C) Both: readable format with optional hash suffix

**Recommendation:** Keep (A) for v1. Add hash in v2 if collisions become a problem.

---

## 3. Repository Name

**Current:** `go-finding` (GitHub) / `finding` (Go module path: `github.com/larsartmann/go-finding`)

**Options:** `finding`, `finding-sdk`, `go-finding`

**Recommendation:** Keep `go-finding`. It's clear, unique, and follows Go naming conventions.

---

## 4. Suppression Expiry

**Current:** `Suppression.ExpiresAt` field exists but is not enforced.

**Question:** Should the pipeline skip expired suppressions automatically?

**Recommendation:** Defer to v1.1. Add `IsActive()` method that checks expiry, but don't auto-filter.

---

## 5. API Stability for v1.0.0

**Current:** v0.2.1 — API-stable beta. All three conditions below are resolved.

**Status:** All blocking decisions resolved:

- [x] FixStrategyAI decision (#1 above) — Resolved: keep as artifact (B)
- [x] Builder.Build() error return is stable — Returns `(Finding, error)` since v0.2.0
- [x] ID format is finalized (#2 above) — Resolved: keep readable format (A) for v1

**Next target:** v1.0.0 = production release with API stability guarantee.

---

## 6. NewFinding API Pattern

**Decision:** Keep the current 6-parameter constructor + Builder pattern.

**Rationale:**

- `NewFinding(rule, toolName, message, severity, pos, confidence)` provides required fields with auto-generated ID and clamped confidence.
- `NewBuilder(rule, toolName, message, severity, pos)` provides a fluent API for optional fields.
- Functional options were considered but rejected — they add allocation and complexity for no ergonomics gain over the builder.
- The builder pattern is idiomatic Go and already in use.

**Status:** Resolved — no changes needed.

---

## 7. Domain-Specific FixProvider Location

**Decision:** Domain-specific providers (Go AST, Rust syn, etc.) should live in **separate modules** (`pipeline/fix/` is not appropriate for multi-language providers).

**Rationale:**

- Providers are domain-specific and would import heavy dependencies (go/ast, Rust parser, etc.).
- Separate modules allow consumers to import only what they need.
- The `FixProvider` interface in `pipeline/` is the contract; implementations live outside.
- Example: `github.com/larsartmann/go-finding-provider-goast` as a separate repo.

**Status:** Resolved — providers live in separate modules.

---

## 8. Properties map[string]any — Rejected

**Decision:** NOT adding `Properties map[string]any` to the `Finding` struct. `Metadata map[string]string` is the sole extensibility field.

**Rationale:**

- `map[string]string` is fully typed, lossless for interchange (SARIF, JSON, CLI, env vars).
- `map[string]any` requires type assertions at every read site, violating type safety.
- Complex values can be JSON-serialized into string values.
- This decision preserves the simplicity and type safety of the struct.
- See commit `024b6a3` for the detailed rationale.

**Status:** WONTFIX — intentionally rejected.

---

## 9. go-sarif Library vs Hand-Rolled SARIF

**Question:** Should we replace the hand-rolled SARIF implementation with `github.com/owenrumney/go-sarif/v3`?

**Answer: No.** The library is well-maintained but introduces an impedance mismatch that makes our code worse, not better.

---

### 9.1 The Core Problem: Impedance Mismatch

go-sarif is **SARIF-centric**. go-finding is **Finding-centric**. The two domain models point in opposite directions.

Our API asks: "I have a `Finding`. Give me SARIF."

go-sarif's API asks: "I have a SARIF `Run`. Add a `Result` with a `Rule` and `Location`."

This means every conversion becomes an adapter problem, not a simple mapping.

---

### 9.2 Concrete Code Comparison

**Current export (hand-rolled):**

```go
func (r *Report) ToSARIF() ([]byte, error) {
    data, err := json.MarshalIndent(r.sarifLog(), "", "  ")
    return data, err
}

func findingToSARIF(f Finding) SarifResult {
    return SarifResult{
        RuleID:     f.Rule,
        Level:      severityToSARIFLevel(f.Severity),
        Message:    SarifMessage{Text: f.Message},
        Locations:  sarifLocations(f),
        Fixes:      sarifFixes(f),
        Related:    sarifRelatedLocs(f),
        Properties: sarifProperties(f),
    }
}
```

**Equivalent with go-sarif (what we'd have to write):**

```go
func (r *Report) ToSARIF() ([]byte, error) {
    rep := report.NewV22Report()
    run := sarif.NewRunWithInformationURI(r.Tool.Name, "")

    for _, f := range r.Findings {
        if f.IsSuppressed() {
            continue
        }

        result := run.CreateResultForRule(f.Rule).
            WithLevel(severityToSARIFLevel(f.Severity)).
            WithMessage(sarif.NewTextMessage(f.Message))

        // Location
        result.AddLocation(sarif.NewLocationWithPhysicalLocation(
            sarif.NewPhysicalLocation().
                WithArtifactLocation(sarif.NewSimpleArtifactLocation(f.Position.File)).
                WithRegion(sarif.NewRegion().
                    WithStartLine(f.Position.Line).
                    WithStartColumn(f.Position.Column)),
        ))

        // Property bag for round-trip
        pb := sarif.NewPropertyBag()
        pb.Add("go-finding/id", f.ID)
        pb.Add("go-finding/severity", string(f.Severity))
        pb.Add("go-finding/fixStrategy", string(f.FixStrategy))
        pb.Add("go-finding/toolName", f.ToolName)
        pb.Add("go-finding/category", string(f.Category))
        // ... more properties ...
        result.WithProperties(pb)

        // Fix
        if f.HasFix() {
            result.AddFix(sarif.NewFix().
                WithDescription(sarif.NewTextMessage(f.Suggestion)).
                WithArtifactChanges([]sarif.ArtifactChange{
                    sarif.NewArtifactChange().
                        WithArtifactLocation(sarif.NewSimpleArtifactLocation(f.Position.File)).
                        WithReplacements([]sarif.Replacement{
                            sarif.NewReplacement().
                                WithDeletedRegion(sarif.NewRegion().
                                    WithStartLine(f.Position.Line).
                                    WithStartColumn(f.Position.Column).
                                    WithEndLine(f.Position.Line).
                                    WithEndColumn(f.Position.Column)).
                                WithInsertedText(sarif.NewMultiformatMessageString().
                                    WithText(f.AfterCode)),
                        }),
                }))
        }
    }

    rep.AddRun(run)

    var buf bytes.Buffer
    if err := rep.PrettyWrite(&buf); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

The go-sarif version is **~3x more code**, harder to read (deeply nested builder chains with `[]sarif.ArtifactChange{{...}}` soup), and loses the clarity of struct literal initialization.

---

### 9.3 Round-Trip Property Bag: Extra Indirection

Our hand-rolled `SarifResult` has `Properties map[string]any` directly:

```go
result.Properties = map[string]any{
    sarifPropID:       f.ID,
    sarifPropSeverity: string(f.Severity),
}
```

go-sarif wraps this in a `PropertyBag` struct:

```go
type PropertyBag struct {
    Properties Properties `json:"properties,omitempty"`  // map[string]interface{}
    Tags       []string   `json:"tags"`
}
```

So the adapter must map `map[string]any` ↔ `*PropertyBag` at every read/write site. This is not simpler — it is an extra layer of indirection for zero benefit.

---

### 9.4 Loss of First-Class Features

| Feature              | Hand-rolled                                             | go-sarif                               |
| -------------------- | ------------------------------------------------------- | -------------------------------------- |
| Streaming output     | `json.NewEncoder(w)` natively                           | `PrettyWrite(&buf)` buffers everything |
| Context cancellation | `WriteSARIF(ctx, w)` checks `ctx.Err()` first           | No context support                     |
| Reader-based import  | `FindingsFromReader(ctx, r)` streams via `json.Decoder` | `FromBytes(data)` requires full buffer |

Adapting these would require wrapping go-sarif's API, adding even more adapter code.

---

### 9.5 Dependency Argument

Adding `go-sarif` means every consumer of `go-finding` transitively depends on it. Our design principle #1 is "core types depend only on stdlib." The `finding` package currently has **zero** third-party dependencies in the root package. Breaking that for a builder API we do not need would be a regression.

---

### 9.6 What go-sarif Offers That We Don't Need

| go-sarif "benefit"     | Reality for us                                                           |
| ---------------------- | ------------------------------------------------------------------------ |
| Full SARIF 2.2 support | We only need 2.1.0. No demand for 2.2.                                   |
| Schema validation      | Would be nice, but blocked by 7K-line schema file, not by implementation |
| 100+ types             | We use 15. The other 85 are cognitive overhead                           |
| Active maintenance     | Our 470 LOC need near-zero maintenance — SARIF 2.1.0 is stable           |
| "Standard" library     | We are not building SARIF reports — we are converting Findings ↔ SARIF   |

---

### 9.7 Summary Comparison

| Dimension             | Hand-rolled                           | go-sarif v3                     |
| --------------------- | ------------------------------------- | ------------------------------- |
| Extra dependencies    | 0                                     | +1 module                       |
| Lines we maintain     | ~470                                  | ~0 (but adapter layer ~300-400) |
| Spec coverage         | Subset we use                         | Full 2.1.0 + 2.2.0              |
| Schema validation     | None                                  | Built-in `Validate()`           |
| go-finding round-trip | Native (property bag designed for it) | Requires adapter                |
| Streaming I/O         | Native (`json.Encoder`/`Decoder`)     | Buffered (`PrettyWrite`)        |
| Context cancellation  | First-class                           | None                            |
| API shape             | Finding-centric                       | SARIF-centric                   |
| Test coverage         | 100% (unit + fuzz, 1.6M execs)        | External                        |

---

### 9.8 When to Reconsider

- If we need **SARIF 2.2.0** features (currently no demand)
- If we need **schema validation** in production (currently blocked by schema size, not implementation)
- If we need **code flows, graphs, or taxonomies** (not on roadmap)
- If the hand-rolled implementation grows beyond ~1000 LOC (suggests we are reimplementing too much)

---

### 9.9 Final Decision

**Keep hand-rolled.** go-sarif is a well-maintained library for SARIF-first applications. go-finding is a Finding-first library that happens to interchange with SARIF. The adapter layer would be larger than our current implementation, harder to read, and would add a dependency we explicitly designed the project to avoid.

**Status:** Resolved — keep hand-rolled.

---

## 10. Report.Findings Encapsulation

**Date:** 2026-06-08
**Status:** Proposed (requires v1.0 milestone)

### Context

`Report.Findings` is a public `[]Finding` slice. External code can bypass the `sync.RWMutex` and cause data races:

```go
r.Findings[0].Severity = finding.SeverityCritical  // No lock held
r.Findings = append(r.Findings, f)                   // Data race with AddFinding
```

### Migration Path (Already Built)

- `FindingsSnapshot()` — returns deep-cloned slice under RLock (v0.4.2+)
- `All()` — returns `iter.Seq[Finding]` holding RLock during iteration
- `FindByID()` — single finding lookup under RLock
- `Filter()` — returns new filtered report
- `Map()` — returns new transformed report

### Options

| # | Approach | Breaking? | Effort | Safe? |
|---|----------|-----------|--------|-------|
| A | Make `Findings` unexported, keep everything else | Yes | Low | Yes |
| B | Keep `Findings` exported with documented caveat | No | None | No |
| C | Replace `Findings` with `[]Finding` getter method | Yes | Medium | Yes |

### Recommendation

**Option A** for v1.0. Rename `Findings` to `findings` (unexported). The existing accessor methods (`FindingsSnapshot`, `All`, `FindByID`, `Filter`, `Map`) provide all needed access patterns. Direct mutation was never documented as safe.

### Migration Guide (v1.0)

```go
// Before (v0.x)
for _, f := range report.Findings { ... }

// After (v1.0)
for f := range report.All() { ... }
// or
snapshot := report.FindingsSnapshot()
```

### Decision

**Deferred to v1.0.** This is a breaking change that must happen before the v1.0 stability guarantee. All necessary migration infrastructure is already in place.
