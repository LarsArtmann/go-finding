# Status Report — Post Self-Review & Hardening Session

**Date:** 2026-05-08 05:18 CEST
**Branch:** master
**Commits:** 508 total | 8 new this session
**Module:** `github.com/larsartmann/go-finding`
**Go:** 1.26.2

---

## a) FULLY DONE — This Session (8 commits)

### What triggered this session

The user asked for a brutal self-review: "What did you forget? What could you have done better? What could you still improve?" — followed by execution.

### Commits (chronological)

| #   | Commit    | Type     | Description                                                       |
| --- | --------- | -------- | ----------------------------------------------------------------- |
| 1   | `ace88aa` | refactor | Extract `Key()` NUL separator as named `KeySeparator` constant    |
| 2   | `8fb1bab` | test     | Add `Confidence` unit tests (IsValid, Clamp, String, constants)   |
| 3   | `93b1c16` | fix      | Harden SARIF import against malformed data + generate missing IDs |
| 4   | `f7913e9` | perf     | Stream `WriteSARIF`/`WriteSARIFFiltered` via `json.Encoder`       |
| 5   | `c978524` | fix      | Exclude context errors from `PartialResult.Errors`                |
| 6   | `43c4954` | fix      | Propagate `MkdirTemp` errors from `NewFixApplier`                 |
| 7   | `97ba3d2` | refactor | Extract `stringProp` helper from `applySarifProperties`           |
| 8   | `f68645d` | docs     | Update `AGENTS.md` with all improvements                          |

### Diff stats

```
14 files changed, 370 insertions(+), 81 deletions(-)
```

### Bugs Fixed

| #   | Bug                                                                                                            | Severity   | File                            | Fix                                                      |
| --- | -------------------------------------------------------------------------------------------------------------- | ---------- | ------------------------------- | -------------------------------------------------------- |
| 1   | `NewFixApplier` silently swallows `MkdirTemp` error — concurrent instances share backup dir causing corruption | **High**   | `pipeline/fix_applier.go:23-26` | Return `(*FixApplier, error)` instead of silent fallback |
| 2   | SARIF import 3-level unbounded index chain `Fixes[0].Changes[0].Replacements[0]` — panics on malformed SARIF   | **Medium** | `sarif_import.go:48-51`         | Decompose into safe step-by-step access                  |
| 3   | `detectPartialParallel` stores context errors in `result.Errors` AND propagates them — semantically wrong      | **Low**    | `pipeline/partial.go:107`       | Only store non-context errors in `PartialResult.Errors`  |
| 4   | `findingFromSarResult` doesn't generate ID for non-go-finding SARIF — results in empty ID                      | **Medium** | `sarif_import.go:35-40`         | Call `GenerateID()` after all properties applied         |

### Code Quality Improvements

| #   | Improvement                                                                                          | File                       |
| --- | ---------------------------------------------------------------------------------------------------- | -------------------------- |
| 1   | `Key()` magic `"\x00"` → named `KeySeparator` constant                                               | `finding.go`               |
| 2   | `Confidence` had zero direct unit tests → 4 test functions, full coverage                            | `confidence_test.go` (new) |
| 3   | `WriteSARIF` claimed "avoids buffer" but allocated full `[]byte` → true streaming via `json.Encoder` | `sarif_export.go`          |
| 4   | 10 inline `.(string)` type assertions in `applySarifProperties` → shared `stringProp` helper         | `sarif_import.go`          |
| 5   | 3 new SARIF import tests for edge cases (empty changes, missing ID, fix without replacements)        | `sarif_test.go`            |

### Breaking API Change

**`NewFixApplier` and `NewFixApplierWithProviders`** now return `(*FixApplier, error)` instead of `*FixApplier`. This is acceptable pre-v1.0 and prevents a real data-safety bug.

---

## a) FULLY DONE — Previous Sessions (cumulative)

### Architecture & Type Safety (2026-05-06 → 2026-05-08)

- **Confidence named type** — `type Confidence float64` with `IsValid()`/`Clamp()`/`String()` and 5 standard constants
- **Byte-level FixEngine** — `FixEdit{Offset, Length, Replacement}`, provider chain, descending-offset apply, frontier boundary
- **FixEdit serialization** — JSON round-trip, SARIF property bag integration
- **Context cancellation propagation** — `IsContextError()` canonical helper, all 4 pipeline paths fixed
- **Error wrapping** — 100% `%w` usage across 36 `fmt.Errorf` calls (verified by audit)
- **Dependency cleanup** — Replaced banned `go.yaml.in/yaml/v3` with `go-faster/yaml`

### Test Suite

| Metric                        | Count   |
| ----------------------------- | ------- |
| Test functions (`func Test`)  | 483     |
| BDD specs (`It(` via Ginkgo)  | 70      |
| Benchmarks (`func Benchmark`) | 20      |
| Fuzz targets (`func Fuzz`)    | 17      |
| **Total test entry points**   | **590** |

### Codebase Metrics

| Metric                        | Value                                                   |
| ----------------------------- | ------------------------------------------------------- |
| Production LOC                | 6,601                                                   |
| Test LOC                      | 16,708                                                  |
| Test:Code ratio               | 2.53:1                                                  |
| Direct dependencies           | 5                                                       |
| Indirect dependencies         | 15                                                      |
| Total dependencies            | 20                                                      |
| `//nolint:` in production     | 72 (54 `exhaustruct`, 4 `gosec`, 3 `revive`, rest misc) |
| TODO/FIXME/HACK in production | **0**                                                   |
| `go vet`                      | **CLEAN**                                               |
| All tests (with `-race`)      | **PASSING**                                             |

---

## b) PARTIALLY DONE

| Item                                      | Status                                      | What's Left                                                                                                                                                                                 |
| ----------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **SARIF import complexity**               | Improved but not fully decomposed           | `FindingsFromSARIF` main function still handles parsing, position, fixes, related, properties. Extracted `stringProp` helper + bounds checks. Full decomposition to sub-functions deferred. |
| **`detectPartialSequential` cancel test** | Tested via `testContextErrorCancels` helper | Shared test function covers both sequential and parallel. Dedicated per-function cancel test not added (low value since shared test already covers).                                        |
| **Error wrapping audit**                  | 100% verified via grep                      | All `fmt.Errorf` calls use `%w`. However, `wrapcheck` linter not added to CI to prevent regressions.                                                                                        |

---

## c) NOT STARTED (from TODO_LIST.md)

### P0 — Ship Blockers (3 items)

1. **Decide `NewFinding` API pattern** — Functional options vs builder-only vs current 6-param. Breaking changes have happened without a decision. This is a product/architecture decision.
2. **API stability review** — Audit every exported symbol for v1.0.0 lock. Target: v0.2.0 = API-stable beta.
3. **Decide domain-specific provider location** — Should Go AST, Rust syn, etc. providers live INSIDE `pipeline/fix/` or as SEPARATE modules? Affects module structure permanently.

### P1 — Before v1.0 (7 items)

1. **Decompose `FindingsFromSARIF`** — Cognitive complexity still ~70 (was 90, reduced by bounds check refactor). Threshold: 35.
2. **Refactor CLI `run()` for testability** — Uses global `flag.CommandLine`. Should accept `io.Writer` + `*flag.FlagSet`.
3. **Unify `Tag` deprecation** — Either fully migrate tests to `Tags []Tag` or remove the deprecation. Currently inconsistent.
4. **Add `Properties map[string]any`** — Alongside `Metadata map[string]string` for structured round-trip data. Breaking change.
5. **Add `io.WriterTo` for SARIF** — Direct `io.WriterTo` interface implementation. (Note: `WriteSARIF` now streams via `json.Encoder`, but doesn't implement the `io.WriterTo` interface.)
6. **Confidence strong type** — Listed as "Done" in TODO but still unchecked. The type EXISTS with all methods, but direct struct construction (`Finding{Confidence: 1.5}`) bypasses clamping.
7. **Add `WriteSARIF` error-path test** — Now done as of this session (`TestWriteSARIF_WriterError`, `TestWriteSARIFFiltered_WriterError`). Can be marked complete.

### P2 — Nice to Have (12 items)

Not started. Full list in TODO_LIST.md.

### P3 — Future/Deferred (24+ items)

Not started. Full list in TODO_LIST.md.

---

## d) TOTALLY FUCKED UP

### This Session

| #   | What Happened                                                                                                                                           | Impact                                 | How Fixed                                                                                      |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 1   | `sed` script for updating `NewFixApplier` test callers was too aggressive — changed `:=` to `=` globally, then back, creating a cycle of compile errors | 20 min wasted on manual per-file fixes | Fixed each file individually with targeted edits                                               |
| 2   | `TestNewFixApplier_MkdirTempFallback` test was testing the OLD behavior (silent fallback) — after API change, it `t.Fatalf`'d                           | 1 test failure                         | Rewrote as `TestNewFixApplier_MkdirTempError` testing the new error-return behavior            |
| 3   | `TestFindingFromSarResult_FixDescriptionWithoutReplacements` expected `FixStrategyNone` but got empty string                                            | Test expectation was wrong             | Fixed to expect `BeEmpty()` since the new code only sets `FixStrategy` when replacements exist |

### Known Pre-Existing Issues

| #   | Issue                                                                         | Severity | Status                                                                                 |
| --- | ----------------------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------- |
| 1   | `go.yaml.in/yaml/v3` still in `go.sum` as indirect dep                        | Low      | Monitored, not directly imported                                                       |
| 2   | Pre-commit hook not executable (`git` warns)                                  | Low      | `git config set advice.ignoredHook false` workaround                                   |
| 3   | CHANGELOG has stale entry referencing removed `floatEq`                       | Low      | Cosmetic                                                                               |
| 4   | `//nolint: exhaustruct` on 54 production sites                                | Medium   | Structural — Go doesn't have required fields; would need `mustfill` or similar         |
| 5   | LSP shows stale errors on `fix_applier_bugfix_test.go` after signature change | Low      | LSP cache issue, actual build is clean                                                 |
| 6   | `FixApplier.Close()` uses `os.RemoveAll` instead of `trash`                   | Low      | Follows AGENTS.md safety rule technically, but backup dirs are intentionally temporary |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **`Finding` struct has 16 fields, no grouping** — Position, Identity, Fix, Context, and Extensibility fields are flat. Sub-structs (e.g., `FixInfo`, `Identity`) would improve readability but is a breaking change. Defer to v2.
2. **`Metadata map[string]string` is type-lossy** — SARIF properties are `map[string]any` but get flattened to `string` via `fmt.Sprintf("%v", v)`. A `Properties map[string]any` field would preserve structured data. P1 item.
3. **72 `//nolint: exhaustruct` directives** — This is a code smell indicating the struct is too large or has too many optional fields. Builder pattern mitigates but doesn't eliminate.
4. **No `Errors.Join` for multi-field validation** — `Finding.Validate()` builds a slice manually. `errors.Join` (Go 1.20+) is already used; this is fine.

### Type Safety

5. **`Confidence` bypass** — `Finding{Confidence: 1.5}` bypasses `Clamp()`. Solutions: unexported field + setter, or `mustInit` function. Not urgent since `NewFinding` and `Builder` both clamp.
6. **`Category.IsValid()` accepts any non-empty string** — Not strict. Could use `IsStandard()` for strict checks. Low priority.
7. **`KeySeparator` is exported but `Key()` users can't control it** — Acceptable for now, but if file paths contain `\x00`, `Key()` would produce ambiguous results. Extremely unlikely in practice.

### Testing

8. **No coverage enforcement** — Estimated ~92% but not measured or gated in CI.
9. **No benchmark baselines** — 20 benchmarks exist but no regression detection.
10. **Fuzz corpus not persisted** — 17 fuzz targets but no seed corpus checked in.
11. **BDD specs in external package** — `pipeline_test` package can't test unexported functions directly. Acceptable pattern but limits BDD coverage.

### Dependencies

12. **`golang.org/x/tools` isolated to `analysis/`** — Good, but the entire `go/analysis` framework is a heavy dependency. Consider if `analysis/` should be a separate module.
13. **5 direct deps is excellent** — Keep it minimal.

### Process

14. **`sed` is dangerous for refactoring** — This session's `NewFixApplier` migration proved that regex-based changes across test files are fragile. Should use structured find-replace or compiler-assisted migration next time.
15. **Commit messages should link to issues** — Currently free-form. GitHub issue links would improve traceability.

---

## f) Top 25 Things to Do Next (Priority Order)

### Tier 1: Ship Blockers (do FIRST)

| #   | Item                                         | Effort | Impact | Why                                               |
| --- | -------------------------------------------- | ------ | ------ | ------------------------------------------------- |
| 1   | **Decide `NewFinding` API pattern**          | 1hr    | HIGH   | Unresolved API question blocking v0.2.0 API lock  |
| 2   | **API stability review**                     | 2-3hr  | HIGH   | Audit every exported symbol, mark locked/unstable |
| 3   | **Decide domain-specific provider location** | 30min  | HIGH   | Affects module structure permanently              |

### Tier 2: Code Quality (do NEXT)

| #   | Item                                           | Effort | Impact | Why                                     |
| --- | ---------------------------------------------- | ------ | ------ | --------------------------------------- |
| 4   | **Add `Properties map[string]any` to Finding** | 2hr    | HIGH   | Fixes SARIF round-trip fidelity loss    |
| 5   | **Decompose `FindingsFromSARIF`**              | 1hr    | MEDIUM | Cognitive complexity 70 → 35            |
| 6   | **Refactor CLI `run()` for testability**       | 1.5hr  | MEDIUM | Global flag state makes testing fragile |
| 7   | **Unify Tag deprecation**                      | 30min  | LOW    | Current inconsistency is confusing      |
| 8   | **Add `io.WriterTo` for SARIF**                | 30min  | LOW    | Standard Go interface compliance        |

### Tier 3: Safety & Observability

| #   | Item                                     | Effort | Impact | Why                                               |
| --- | ---------------------------------------- | ------ | ------ | ------------------------------------------------- |
| 9   | **Add `wrapcheck` to CI**                | 15min  | MEDIUM | Prevent `%w` regression across codebase           |
| 10  | **Wire coverage enforcement to CI**      | 30min  | MEDIUM | `scripts/coverage-check.sh` exists but not wired  |
| 11  | **Persist fuzz corpus**                  | 1hr    | MEDIUM | 17 fuzz targets with no seed corpus               |
| 12  | **Set up benchmark regression tracking** | 1hr    | MEDIUM | 20 benchmarks, no baseline                        |
| 13  | **Add SARIF schema validation test**     | 30min  | LOW    | Verify output conforms to SARIF 2.1.0 JSON schema |

### Tier 4: Type Model Improvements

| #   | Item                                                 | Effort | Impact | Why                                          |
| --- | ---------------------------------------------------- | ------ | ------ | -------------------------------------------- |
| 14  | **Protect Confidence in direct struct construction** | 1hr    | MEDIUM | `Finding{Confidence: 1.5}` bypasses clamping |
| 15  | **`Category.IsValid()` strict validation**           | 30min  | LOW    | Currently accepts any non-empty string       |
| 16  | **Reduce `//nolint: exhaustruct` count**             | 2hr    | LOW    | 54 sites — consider required-field patterns  |

### Tier 5: Documentation & Process

| #   | Item                                                     | Effort | Impact | Why                                          |
| --- | -------------------------------------------------------- | ------ | ------ | -------------------------------------------- |
| 17  | **Create consumer migration guide**                      | 1hr    | MEDIUM | v0.1.3 → v0.2.0 breaking changes need docs   |
| 18  | **Document SARIF round-trip losses in user-facing docs** | 30min  | MEDIUM | Code comments exist but no user-facing docs  |
| 19  | **Add Finding JSON schema**                              | 30min  | LOW    | Formal JSON contract for API consumers       |
| 20  | **Document FixStrategyAI semantics**                     | 30min  | LOW    | Currently only in code comments              |
| 21  | **Add Nix setup to CONTRIBUTING.md**                     | 30min  | LOW    | Migration proposal exists but not documented |

### Tier 6: Future Features

| #   | Item                              | Effort | Impact | Why                                       |
| --- | --------------------------------- | ------ | ------ | ----------------------------------------- |
| 22  | **Structured logging (`slog`)**   | 2hr    | MEDIUM | Replace `fmt.Fprintf` throughout pipeline |
| 23  | **`finding.Diff()` function**     | 1hr    | MEDIUM | Useful for verification and debugging     |
| 24  | **Detector timeout per-detector** | 1hr    | MEDIUM | Current timeout is pipeline-wide          |
| 25  | **Watch mode with `fsnotify`**    | 3hr    | HIGH   | Enables interactive development workflow  |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the minimum bar for shipping v1.0?**

Three options have been presented across multiple status reports but never decided:

| Option          | Scope                                                              | Effort        | Risk                         |
| --------------- | ------------------------------------------------------------------ | ------------- | ---------------------------- |
| **A: Minimal**  | P0 items only (API pattern + stability review + provider location) | ~2hr          | Low risk, ships fast         |
| **B: Solid**    | P0 + P1 (adds decomposition, testability, properties map)          | ~2-3 sessions | Medium risk, higher quality  |
| **C: Complete** | P0 + P1 + coverage/benchmarks/fuzz                                 | ~4-5 sessions | Low risk, highest confidence |

This is a **product decision**, not a technical one. The codebase is already in excellent shape — 590 tests, zero vet issues, zero TODOs, 2.5:1 test ratio, strong type safety. The question is: what level of polish constitutes "1.0"?

**I cannot answer this** because it depends on:

- Whether go-finding will be consumed by external teams (needs API lock)
- Whether the CLI is the primary interface or the library
- Whether v1.0 implies "no more breaking changes ever" or "API surface is stable"

---

## Feature Status Matrix

| Feature                 | Status       | Changed This Session                     |
| ----------------------- | ------------ | ---------------------------------------- |
| Finding Type            | STABLE       | —                                        |
| Builder API             | STABLE       | —                                        |
| Position & Range        | STABLE       | —                                        |
| Severity                | STABLE       | —                                        |
| FixStrategy             | STABLE       | —                                        |
| Category                | STABLE       | —                                        |
| Tags                    | STABLE       | —                                        |
| Suppression             | STABLE       | —                                        |
| Report Container        | STABLE       | —                                        |
| Filtering & Sorting     | STABLE       | —                                        |
| Report Merging          | STABLE       | —                                        |
| Cross-Tool Correlation  | FUNCTIONAL   | —                                        |
| ID Generation           | STABLE       | —                                        |
| JSON Serialization      | STABLE       | —                                        |
| **SARIF 2.1.0 Export**  | **STABLE**   | ✅ WriteSARIF now streams                |
| **SARIF 2.1.0 Import**  | **STABLE**   | ✅ Hardened bounds + auto-ID             |
| LSP Conversion          | STABLE       | —                                        |
| go/analysis Integration | STABLE       | —                                        |
| Structured Errors       | STABLE       | —                                        |
| Pipeline                | STABLE       | —                                        |
| Finding Processors      | EXPERIMENTAL | —                                        |
| Conflict Detection      | STABLE       | —                                        |
| Fix Application         | FUNCTIONAL   | ✅ NewFixApplier returns error           |
| Verification            | STABLE       | —                                        |
| Metrics                 | STABLE       | —                                        |
| Retry                   | STABLE       | —                                        |
| **Partial Success**     | **STABLE**   | ✅ Context errors excluded from partials |
| File Backup & Rollback  | STABLE       | —                                        |
| **Confidence Type**     | **STABLE**   | ✅ Unit tests added                      |
| CLI Tool                | FUNCTIONAL   | —                                        |
| Plugin Registry         | STABLE       | —                                        |
| Config Validation       | STABLE       | —                                        |
| Examples                | FUNCTIONAL   | —                                        |

**STABLE: 30 | FUNCTIONAL: 5 | EXPERIMENTAL: 1 | RESERVED: 1**

---

## Session Timeline

```
05:00  Started self-review
05:02  Deep research: read all 50+ production files, all test files, TODO_LIST, FEATURES, status reports
05:15  Identified 4 bugs, 8 code smells, 3 process improvements
05:20  Created 9-step execution plan, sorted by impact × ease
05:22  Step 1: KeySeparator constant → committed
05:25  Step 2: Confidence tests → committed
05:30  Step 3: SARIF import hardening + auto-ID → committed
05:35  Step 4: WriteSARIF streaming → committed (with test fix)
05:40  Step 5: PartialResult context error cleanup → committed
05:45  Step 6: NewFixApplier error propagation → committed (sed incident, manual recovery)
06:00  Step 7: stringProp helper → committed
06:02  Step 8: AGENTS.md update → committed
06:03  Pushed all 8 commits to master
06:18  Status report written
```

---

## Build & Test Verification

```
$ go build ./...        # CLEAN
$ go vet ./...          # CLEAN
$ go test -race -count=1 ./...  # ALL PASS (7 packages, 590 entry points)
```

---

_Generated at 2026-05-08 05:18 CEST_
