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

**Current:** `Suppression.ExpiresAt` field exists and is enforced via `IsActive(now)` which checks expiry.

**Status:** Resolved in v1.1. `IsActive()` checks validity + expiry; `IsSuppressedAt()` uses it for pipeline filtering.

---

## 5. API Stability for v1.0.0

**Current:** v1.3.0 — API-stable. All three conditions below are resolved.

**Status:** All blocking decisions resolved:

- [x] FixStrategyAI decision (#1 above) — Resolved: keep as artifact (B)
- [x] Builder.Build() error return is stable — Returns `(Finding, error)` since v0.2.0
- [x] ID format is finalized (#2 above) — Resolved: keep readable format (A) for v1

**Status:** v1.0.0 released. API stability guarantee published (`docs/API_STABILITY.md`).

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

**Status:** Resolved — GoAST provider lives at `pipeline/goast/` within the pipeline module (uses only stdlib `go/parser`, no external deps). Future non-Go providers should be separate repos.

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

    for _, f := range r.FindingsSnapshot() {
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

Adding `go-sarif` means every consumer of `go-finding` transitively depends on it. The `finding` package intentionally keeps its dependency surface minimal. Adding a heavy SARIF library for a builder API we do not need would be a regression.

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
**Status:** Implemented in v1.0.0 — Option A adopted.

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

| # | Approach                                          | Breaking? | Effort | Safe? |
| - | ------------------------------------------------- | --------- | ------ | ----- |
| A | Make `Findings` unexported, keep everything else  | Yes       | Low    | Yes   |
| B | Keep `Findings` exported with documented caveat   | No        | None   | No    |
| C | Replace `Findings` with `[]Finding` getter method | Yes       | Medium | Yes   |

### Recommendation

**Option A adopted in v1.0.0.** `Findings` renamed to `findings` (unexported). The existing accessor methods (`FindingsSnapshot`, `All`, `FindByID`, `Filter`, `Map`) provide all needed access patterns. Direct mutation was never documented as safe.

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

**Implemented in v1.0.0.** All necessary migration infrastructure was in place since v0.7.0.

---

## 11. v1.0.0 Breaking Changes Plan

**Status:** Implemented in v1.0.0.

### Consolidated v1.0 Breaking Changes

All changes below have been implemented:

| #  | Change                                  | Status  | Migration                                       |
| -- | --------------------------------------- | ------- | ----------------------------------------------- |
| 1  | `Report.Findings` unexported            | ✅ Done | Use `FindingsSnapshot()`, `All()`, `FindByID()` |
| 2  | `Report.Merge` removed                  | ✅ Done | Use `MergeInto(other)`                          |
| 3  | `RecordFix()` removed                   | ✅ Done | Use `RecordFixes(1)`                            |
| 4  | SARIF types remain unexported           | ✅ Done | No change needed                                |
| 5  | `Position.Offset` zero-value semantics  | ✅ Done | Use `HasOffset()`; `-1` sentinel for unset      |
| 6  | `Config.OnStage` removed                | ✅ Done | Use `StageHooks`                                |
| 7  | `CountBySeverity()` free function       | ✅ Done | Use `Report.CountBySeverity()`                  |
| 8  | `SeverityAliases()` removed             | ✅ Done | Use `LookupSeverityAlias()`                     |
| 9  | `GetCategory()` removed                 | ✅ Done | Use `CategoryOf()`                              |
| 10 | `HasFix`/`HasSuggestion` free functions | ✅ Done | Use `WithFix`/`WithSuggestion`                  |

### Position/Range Zero-Value Decision (Item 5)

**Problem:** `Position.Offset = 0` is ambiguous — it means both "byte 0" (valid)
and passes `HasOffset() == true`, while also satisfying `IsZero() == true`.

**Options:**

| Option | Approach                                   | Breaking?           | Clarity |
| ------ | ------------------------------------------ | ------------------- | ------- |
| A      | Change `Offset` default to `-1` (sentinel) | Yes — serialization | High    |
| B      | Add `OffsetSet bool` field                 | Yes — struct size   | Medium  |
| C      | Keep as-is, document the trap              | No                  | Low     |

**Recommendation adopted:** Option A in v0.9.0/v1.0.0. `Offset` defaults to `-1` (sentinel for unset).

### Internal Migration Progress (v0.x → v1.0)

The following refactors prepare for v1.0 without breaking consumers:

1. **`findingsLocked()` accessor** — All internal code uses `r.findingsLocked()`
   instead of `r.Findings` directly. At v1.0, only this method changes.
2. **`readFindings()` for thread-safe reads** — External code should already
   use `FindingsSnapshot()` or `readFindings()`.
3. **Deprecated methods** — `Merge` and `RecordFix` have godoc deprecation
   notices pointing to replacements.

### Decision

**Plan executed in v1.0.0.** All items implemented in a single release.

---

## 12. Canonical Finding Identity

**Decision:** `GenerateID()` output is the canonical finding identity. Two findings are identical if and only if their IDs are equal.

**Rationale:**

- `GenerateID` uses `tool:rule:file:line:col` — deterministic and collision-resistant.
- `Key()` is a fallback for findings without an ID; it includes `Message` in the composite key, which is a deliberate difference (pre-GenerateID consumers needed message-based fallback).
- `dedupKey()` strategies (ByID, ByPosition, ByRule) are deliberate relaxations of canonical identity for merge scenarios, not competing definitions.

**Implementation:**

- `Finding.Key()` documented as fallback only; references this ADR.
- `Finding.Equal()` compares all fields including ID.
- `Diff()` and `MergeIter()` use ID as the primary key.

**Status:** Resolved — GenerateID is canonical.

---

## 13. Category/Tags Relationship — Deprecation Plan

**Decision:** Category and Tags will coexist until v1.0.0. Category will be deprecated in favor of Tags (strictly more expressive). The validation invariant (added in this session) prevents conflicts in the meantime.

**Current behavior:**

- `Finding.Category` is a single-value classification (one domain).
- `Finding.Tags` is a multi-value classification (multiple labels).
- Six values are defined as constants in both types (security, style, performance, correctness, complexity, documentation).
- `Validate()` now rejects findings where Category and Tags contain conflicting standard categories.

**Migration plan (v1.0.0):**

1. v0.8.0: Add `// Deprecated:` comment to `Finding.Category` field. (Done — invariant active.)
2. v0.9.0: Add `Finding.PrimaryCategory()` method that returns `Tags[0]` or derived primary.
3. v1.0.0: Remove `Category` field. Consumers migrate to `Tags`. `Summary.ByCategory` becomes `Summary.ByTag`.

**Status:** Open — invariant active, full deprecation deferred to v1.0.0 batch.

---

## 14. sync.Pool for Line Offset Index

**Status:** Evaluated — **SKIP**.

**Context:** The `[]int` line offset index is built per file by `buildLineOffsetIndex`
and reused across all findings in that file via the `lineIndexAware` interface
and the lazy `*[]int` pointer in `resolveEdits`.

**Evaluation:**

- Index size: ~8 bytes/line (int on 64-bit). A 10K-line file: ~80KB.
- One allocation per file, immediately eligible for GC after `ApplyWithConflicts` returns.
- `sync.Pool` would reuse the backing array across files, avoiding repeated allocation.
- BUT: the index content is different per file (different newline positions), so the
  array would need to be fully rewritten each time. `sync.Pool` only saves the allocation
  overhead, not the fill cost.
- Go's small-object allocator is highly optimized for this size class.
- Profiling shows the index allocation is <0.1% of total FixEngine time.

**Conclusion:** The complexity of pool lifecycle management (reset, Put/Get, potential
for stale data) is not justified for a <0.1% improvement. Revisit only if profiling
shows line index allocation as a hot path on very large batches (>100K files).

---

## 15. go-error-family as Core Dependency

**Status:** Accepted — v1.4.0.

**Context:** The core `finding` package historically depended only on the Go standard
library. Error handling was purely internal: `FindingError` had sentinel errors
(`ErrValidation`, `ErrIO`, etc.) and `ErrorCategory` classification, but consumers had
no standardized way to classify go-finding errors alongside errors from other libraries.

The `go-error-family` library (`github.com/larsartmann/go-error-family`) provides a
unified error classification system with `Family` values (Rejection, Conflict,
Transient, Infrastructure) and `Classify()` for routing errors to retry, logging, or
user-facing strategies. Integrating it into go-finding allows consumers to handle
go-finding errors uniformly with all their other errors.

**Decision:** Accept `go-error-family` v0.9.0 as a direct production dependency of the
core module. `FindingError` now implements two `go-error-family` interfaces:

- `errorfamily.Coded` via `ErrorCode()` — returns `"finding.<category>"` (e.g., `"finding.validation"`)
- `errorfamily.Classified` via `ErrorFamily()` — maps categories to families:
  - Validation, Parse → Rejection
  - Conflict → Conflict
  - IO → Transient
  - Internal → Infrastructure

**Tradeoffs:**

- **Gain:** Consumers can call `errorfamily.Classify(err)` on any error, including
  go-finding errors, and get a consistent family for retry/log/escalation decisions.
  No more per-library error type switching.
- **Gain:** Error codes (`"finding.io"`, `"finding.parse"`) are stable, structured
  identifiers for observability and dashboards.
- **Cost:** Core module gains one external dependency. The "zero external deps"
  principle is retired in favor of a small, deliberate dependency surface.
- **Cost:** Consumers who do not use `go-error-family` are unaffected — `ErrorCode()`
  and `ErrorFamily()` are additive methods that do not change existing behavior.

**Why not vendor or inline?** The `go-error-family` library is small and focused, but
it is actively developed. Inlining would freeze the API surface and create a
maintenance burden. Vendoring would duplicate the code. A direct dependency is the
simplest correct choice.

**Reversibility:** Fully reversible. The two methods (`ErrorCode`, `ErrorFamily`) are
additive. Removing the dependency would only require deleting the import and the two
methods. No existing consumer code would break.

## 16. Per-File Rollback as `FixApplier` Default

**Status:** Accepted — v1.7.0 (D1 sign-off 2026-09-08).

**Context:** Until v1.6.x, `FixApplier` treated every multi-file fix run as
all-or-nothing: one hard failure on the last file rolled back every file modified
earlier in the run, discarding clean, verified fixes. Issue #28 documented the
concrete failure: a single unresolvable finding on the final file erased all
applied fixes from every previous file. Additionally, a provider resolve error on
any finding (a soft, per-finding failure) triggered the same global rollback even
when other edits in that file had applied cleanly.

Consumers of batch fix tooling expect failure isolation proportional to the
failure: one bad file should cost that file, not the whole run.

**Decision:** `RollbackPolicyFailingFile` (restore only the failing file) is the
default rollback policy. `RollbackPolicyAllFiles` preserves the legacy
all-or-nothing behavior and is opt-in via `FixApplier.SetRollbackPolicy`,
`pipeline.Config.FixRollbackAllFiles`, or the config-file field
`fixRollbackAllFiles` (CLI: `-fix-rollback-all`). Soft per-finding failures —
provider resolve errors, refused findings — never abort a run: applied edits are
written, failures surface via `ApplyReport.Outcomes` and the joined error return.

**Tradeoffs:**

- **Gain:** Failure isolation matches user expectation; issue #28's scenario now
  loses one file instead of the whole run.
- **Gain:** The `RollbackPolicy` type makes failure semantics explicit and
  testable instead of implicit in control flow.
- **Cost:** Consumers that relied on all-or-nothing semantics must opt back in
  explicitly. This is a behavior change shipped in a minor release — justified
  because the old behavior was a documented bug (#28), the migration is one
  line, and `docs/guides/consumer-migration-v1.7.md` covers it.
- **Cost:** `ApplyReport.RolledBack` lists every backed-up file on rollback,
  including files that were backed up but never modified (content no-ops) —
  a reporting nuance tests must account for.

**Why not default to `RollbackPolicyAllFiles` with an opt-in per-file flag?** The
default is what most consumers get; shipping the buggy-equivalent default and
asking everyone to opt into correctness inverts the burden. Flipping the default
later (after consumers migrate to the opt-out) would churn consumers twice.

**Reversibility:** Reversible at the API level (swap the default constant) but
not at the ecosystem level once consumers depend on per-file semantics — which
is the point: the decision freezes the correct default.

## 17. Outcome Metrics as Pipeline-Native Aggregates

**Status:** Accepted — v1.7.0.

**Context:** With per-finding fix outcomes (issue #27), consumers need
aggregate visibility: how many findings applied, how many were refused, how
many failed. Two places could own that aggregation — the caller reducing
`ApplyReport.Outcomes` itself, or the pipeline's existing `Metrics` collector,
which already tracks stage durations, detector times, and applied-fix counts.

**Decision:** `Metrics.RecordOutcome(status)` / `Metrics.OutcomeCounts()` own
the aggregation. The pipeline's fix stage records one outcome per finding
automatically (`pipeline_detect.go`), `MetricsSnapshot.OutcomeCounts` carries a
point-in-time copy in `PipelineResult`, and the CLI prints a `Fix outcomes:`
summary in canonical status order (nonzero counts only).

**Tradeoffs:**

- **Gain:** Every pipeline consumer gets outcome telemetry for free, including
  CLI users; no per-caller reduction boilerplate.
- **Gain:** Metrics already has the mutex discipline and snapshot pattern —
  counts ride the existing infrastructure with no new locking surface.
- **Cost:** The counts are aggregate-only; callers needing per-finding detail
  still read `ApplyReport.Outcomes`. Metrics is not a finding store.
- **Cost:** Double bookkeeping risk — `fixesApplied` (count) and
  `applied`-outcome counts can diverge if future stages record one but not the
  other. Mitigated by the property test pinning the invariant at the engine
  level and the fix stage recording both from one `ApplyReport`.

**Alternatives:** Callback-based aggregation (`OnOutcome` per finding) was
rejected as redundant with `StageHooks` and heavier than a counter.

**Reversibility:** Additive; removing the recording would only drop telemetry.

## 18. Typed Outcome Errors via `*finding.FindingError`

**Status:** Accepted — v1.7.0.

**Context:** Failed fix outcomes initially carried `fmt.Errorf`-wrapped
provider errors. Consumers filtering outcome failures by error type or
position had to parse message strings; `errors.Is`/`errors.As` stopped at the
wrapper; and the pipeline's error-classification story
(`errorfamily.Classify`, ADR #15) did not apply to fix failures.

**Decision:** Failed `FixOutcome.Err` values are `*finding.FindingError`
(parse category) constructed with the finding's position attached, wrapping
the original provider error as `Cause`. The `errors.Is`/`errors.As` chain
reaches the provider cause; `errorfamily.Classify` routes them as Rejection.
Serialization stays message-string-only (errors do not carry identity through
JSON — match on status, not error value, after a round-trip).

**Tradeoffs:**

- **Gain:** Uniform error classification across the library (detection and
  fixing produce the same error shape); position-bearing failures render
  actionable messages for free.
- **Cost:** Parse category is a slight stretch for provider failures (they are
  "resolution" failures, not syntax errors). Rejection-family routing is still
  correct (do not retry automatically), and the cause chain preserves the
  precise origin.
- **Cost:** Typed errors in a public result struct freeze `*finding.FindingError`
  into the outcome API contract.

**Alternatives:** A dedicated `OutcomeError` type was rejected — it would
duplicate `FindingError`'s fields (position, category, cause) without adding
information.

**Reversibility:** Low: swapping the error type later would break consumers
matching on `*finding.FindingError`. The type is stable since v1.4.0 (ADR #15).
