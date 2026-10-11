# Comprehensive Status Report — go-finding

**Date:** 2026-05-06 13:10 CET
**Session:** Brutal Self-Review + Execution
**Branch:** master (1 commit ahead of origin)
**Build:** GREEN | **Tests:** 502 PASS, 0 FAIL | **Vet:** CLEAN | **Lint:** CLEAN
**Project:** v0.2.1 | Go 1.26.2 | Nix-based dev env

---

## a) FULLY DONE — Completed This Session

### 1. Replace hardcoded `"test.go"` strings with constant (commit `1bc1458`)

- **File:** `analysis/analysis_test.go`
- **What:** 12 hardcoded `"test.go"` string literals replaced with existing `testFilename` constant
- **Why:** `goconst` lint warning. One const definition, used everywhere.

### 2. Remove deprecated `ConflictDetector` and `Verifier` wrapper types (commit `34d09d5`)

- **Files:** `pipeline/conflict.go`, `pipeline/verify.go`, `pipeline/pipeline.go`, `pipeline/conflict_test.go`, `pipeline/verify_test.go`
- **What:** Deleted 33 lines of deprecated wrapper structs (`ConflictDetector`, `NewConflictDetector()`, `Verifier`, `NewVerifier()`) that were zero-value structs delegating to package-level functions. All callers updated to use `DetectConflicts()` and `Verify()` directly.
- **Why:** Accumulated `// Deprecated` types with no removal timeline. They added zero value — just indirection to the same functions.

### 3. Make `Report` zero-value thread-safe (commit `aacd636`)

- **Files:** `report.go`, `merge.go`
- **What:** Changed `mu *sync.Mutex` to `mu sync.Mutex` (value type). Removed `new(sync.Mutex)` allocations from constructors. Simplified `lock()`/`unlock()` to direct calls (removed nil checks). Updated godoc.
- **Why:** `Report{}` without calling `NewReport()` was silently unsafe for concurrent use. Now the zero value is safe — `sync.Mutex` zero value is valid and unlocked. Tradeoff: `Report` must not be copied by value (already used as `*Report` everywhere).

### 4. Split CLI `main.go` into 3 files (commit `e23cf8a`)

- **Files:** `cmd/go-finding/main.go` (194 lines), `cmd/go-finding/config.go` (224 lines, NEW), `cmd/go-finding/registry.go` (59 lines, NEW)
- **What:** Extracted config loading, output formatting, severity parsing, detector registry into separate files. Each file has clear single responsibility.
- **Why:** 458-line god file. `main()` mixed entry logic with config parsing, output formatting, and detector registration.

### 5. Update TODO_LIST.md and AGENTS.md (commit `69494ca`)

- **Files:** `TODO_LIST.md`, `AGENTS.md`
- **What:** Marked 5 items completed in TODO_LIST.md. Updated AGENTS.md CLI file table (3 rows), added "Report zero-value safe" design principle.

### 6. Fix `go vet copylocks` regression (commit `58fb0d3`, NOT YET PUSHED)

- **File:** `json_test.go`
- **What:** Changed `json.Marshal(orig)` to `json.Marshal(&orig)` in 2 test functions. The `Report` value mutex change made value copies trigger `copylocks`.
- **Why:** Caught during status report verification. `go vet` had been clean before this fix.

### Verified Completed (from previous sessions, confirmed in code audit)

- FixEngine byte-level redesign with plugin architecture
- SARIF split into 3 files (`sarif_types.go`, `sarif_export.go`, `sarif_import.go`)
- Pipeline split into `pipeline.go` + `adapters.go` + `config.go`
- Report.Merge, Suppression.IsActive, Tag.IsStandard
- BDD test suite (44+ ginkgo specs)
- FindingProcessor composable transforms
- Pipeline single-use contract documented
- FixApplier cleanup with `defer Close()`
- go/analysis subpackage extraction

---

## b) PARTIALLY DONE — In Progress / Needs Follow-up

### 1. Triage Logic Centralization

- **Status:** Identified, not started
- **What:** "Is this fixable?" is answered differently in 3 places:
  - `Finding.HasFix()` (`finding.go:135`) — checks `Strategy != FixStrategyNone`
  - `Pipeline.triage()` (`pipeline/pipeline.go:346`) — applies category-based rules
  - `FixEngine.Apply()` (`pipeline/fix_engine.go:64`) — filters by fix availability
- **Why it matters:** These definitions overlap but aren't identical. A finding could be "fixable" by one definition but not another. This is a correctness risk.

### 2. Confidence Strong Type

- **Status:** Attempted, deliberately reverted
- **What:** Created `confidence.go` with `type Confidence float64`, then discovered 28 call sites across 12 files. Immediately deleted.
- **Why it matters:** `Finding{Confidence: 1.5}` compiles silently (valid range is 0.0–1.0). Strong type would catch this at compile time.
- **Decision:** Deferred to v1.0 API stabilization. 28 call sites = breaking change across entire API surface.

### 3. `diagnostic.go` Deprecated Wrappers

- **Status:** Deprecated but still present
- **What:** Root package `diagnostic.go` still imports `golang.org/x/tools` and provides backward-compatible wrappers around `analysis/` subpackage.
- **Why still there:** Backward compatibility. Removing these is a breaking API change.

---

## c) NOT STARTED — Tracked in TODO_LIST.md

### P0 — Must Do Before v1.0

| # | Item                                     | Location             | Impact                               |
| - | ---------------------------------------- | -------------------- | ------------------------------------ |
| 1 | Decide `NewFinding` API pattern          | `finding.go`         | Breaking API decision blocking v1.0  |
| 2 | API stability review                     | All exported symbols | Required for v1.0 lock               |
| 3 | Centralize triage logic                  | 3 files              | Correctness risk                     |
| 4 | Decide domain-specific provider location | Architecture         | Affects module structure permanently |

### P1 — Should Do Before v1.0

| #  | Item                                               | Location                      | Impact                                  |
| -- | -------------------------------------------------- | ----------------------------- | --------------------------------------- |
| 5  | Replace `go.yaml.in/yaml/v3` with `go-faster/yaml` | `cmd/go-finding/config.go:12` | Banned library per how-to-golang policy |
| 6  | Decompose `FindingsFromSARIF` (CC 90)              | `sarif_import.go`             | Cognitive complexity 90 (threshold 35)  |
| 7  | Error wrapping consistency audit                   | Various                       | `wrapcheck` linter catches gaps         |
| 8  | Refactor CLI `run()` for testability               | `cmd/go-finding/main.go`      | Uses global flag state                  |
| 9  | Add `Properties map[string]any`                    | `finding.go`                  | Structured SARIF round-trip             |
| 10 | `Confidence` strong type (28 call sites)           | `finding.go:37`               | Type safety                             |
| 11 | Add `WriteSARIF` error-path tests                  | `sarif_test.go`               | 75% coverage, failing writer untested   |
| 12 | Context-cancel tests for partial detection         | `pipeline/partial_test.go`    | Cancel paths untested                   |
| 13 | Unify `Tag` deprecation                            | Various test files            | Inconsistent state                      |

### P2 — Nice to Have

| #  | Item                                       | Location             |
| -- | ------------------------------------------ | -------------------- |
| 14 | SARIF schema validation test               | `sarif_test.go`      |
| 15 | Benchmark regression tracking              | CI                   |
| 16 | `io.WriterTo` for SARIF                    | `sarif.go`           |
| 17 | Nix flake migration                        | Full project         |
| 18 | Structured logging (`slog`)                | `cmd/` + `pipeline/` |
| 19 | `Category.IsValid()` strict validation     | `category.go`        |
| 20 | Document SARIF round-trip losses for users | User-facing docs     |

---

## d) TOTALLY FUCKED UP — Things That Are Wrong

### 1. Triage Split Brain (HIGH)

Three different definitions of "is this fixable?" across the codebase. A finding can pass one check and fail another, leading to:

- Findings marked fixable that never get fixed
- FixEngine silently skipping valid fixes
- Inconsistent behavior between CLI and library use

**This is the #1 correctness risk.**

### 2. `go.yaml.in/yaml/v3` Is a Banned Library (MEDIUM)

The `how-to-golang` policy explicitly bans `go.yaml.in/yaml/v3`. It's currently used in `cmd/go-finding/config.go:12`. Should be replaced with `go-faster/yaml`. This only affects the CLI, not the core library.

### 3. `Metadata map[string]string` Loses Type Information (MEDIUM)

SARIF round-trip via `Metadata` flattens structured data to strings. Number `42` becomes `"42"`. Booleans become `"true""/"false"`. There's no way to recover the original types. `Properties map[string]any` would fix this but hasn't been added yet.

### 4. `Confidence` Is a Bare `float64` (MEDIUM)

`Finding{Confidence: 1.5}` compiles. Valid range is 0.0–1.0. No compile-time protection. `NewFinding()` clamps, but direct struct construction bypasses it. 28 call sites make this a significant migration.

### 5. `FindingsFromSARIF` Cognitive Complexity 90 (MEDIUM)

The SARIF import function has cognitive complexity of 90 (threshold is 35). It's been partially decomposed but is still a maintenance hazard.

---

## e) WHAT WE SHOULD IMPROVE — Architecture & Quality

### Critical Architecture

1. **Centralize triage logic** — One canonical "is fixable?" function, called from all 3 locations. No more split brain.
2. **Decide `NewFinding` API** — Functional options vs builder-only vs current 6-param. Every week without a decision is another breaking change shipped without a plan.
3. **API stability review** — Audit every exported symbol. Lock the API for v1.0. This is the only path to production trust.

### Type Safety

4. **`Confidence` strong type** — 28 call sites but worth it. `Finding{Confidence: 1.5}` must not compile.
5. **`Properties map[string]any`** — Stop losing type data in SARIF round-trips.
6. **`Category.IsValid()` strict validation** — Currently accepts any non-empty string.

### Dependency Hygiene

7. **Replace `go.yaml.in/yaml/v3`** — Banned library. Use `go-faster/yaml`.
8. **Remove deprecated `diagnostic.go` wrappers** — Only exists for backward compat. Pre-v1.0 is the time to break.

### Test Coverage Gaps

9. **`WriteSARIF` error-path tests** — `failingWriter` pattern, currently 75% coverage.
10. **Context-cancel tests** — `detectPartialSequential` (90%) and `detectPartialParallel` (94.1%) have untested cancel paths.
11. **Benchmark regression tracking** — Baselines captured but no CI enforcement.

### Code Quality

12. **Decompose `FindingsFromSARIF`** — CC 90 → multiple focused functions.
13. **CLI `run()` testability** — Inject `io.Writer` + `*flag.FlagSet` instead of globals.
14. **Error wrapping consistency** — `%w` everywhere, enforced by `wrapcheck`.

---

## f) Top 25 Things to Get Done Next

| Priority | #  | Item                                               | Effort | Impact | Category      |
| -------- | -- | -------------------------------------------------- | ------ | ------ | ------------- |
| P0       | 1  | Centralize triage logic (split brain fix)          | M      | HIGH   | Correctness   |
| P0       | 2  | Decide `NewFinding` API pattern                    | S      | HIGH   | Architecture  |
| P0       | 3  | API stability review for v1.0 lock                 | M      | HIGH   | Architecture  |
| P0       | 4  | Replace `go.yaml.in/yaml/v3` with `go-faster/yaml` | S      | MEDIUM | Dependencies  |
| P0       | 5  | Remove deprecated `diagnostic.go` wrappers         | S      | MEDIUM | Cleanup       |
| P1       | 6  | `Confidence` strong type (28 call sites)           | L      | HIGH   | Type Safety   |
| P1       | 7  | Add `Properties map[string]any` to Finding         | M      | MEDIUM | Type Safety   |
| P1       | 8  | Wire FixProviders through CLI                      | M      | MEDIUM | Features      |
| P1       | 9  | Decompose `FindingsFromSARIF` (CC 90→<35)          | M      | MEDIUM | Code Quality  |
| P1       | 10 | `WriteSARIF` error-path tests                      | S      | MEDIUM | Testing       |
| P1       | 11 | Context-cancel tests (partial detection)           | S      | MEDIUM | Testing       |
| P1       | 12 | Error wrapping consistency audit                   | S      | LOW    | Code Quality  |
| P1       | 13 | Refactor CLI `run()` for testability               | M      | MEDIUM | Code Quality  |
| P1       | 14 | Decide domain-specific provider location           | S      | HIGH   | Architecture  |
| P2       | 15 | `Category.IsValid()` strict validation             | S      | MEDIUM | Type Safety   |
| P2       | 16 | Unify `Tag` deprecation (remove `WithTag`)         | S      | LOW    | Cleanup       |
| P2       | 17 | SARIF schema validation test                       | M      | LOW    | Testing       |
| P2       | 18 | `io.WriterTo` for SARIF streaming                  | S      | LOW    | Performance   |
| P2       | 19 | Benchmark regression tracking in CI                | S      | LOW    | Tooling       |
| P2       | 20 | Structured logging (`slog`) in cmd/pipeline        | M      | MEDIUM | Quality       |
| P2       | 21 | Nix flake migration (replace justfile)             | L      | MEDIUM | Tooling       |
| P2       | 22 | Document SARIF round-trip losses for users         | S      | LOW    | Documentation |
| P3       | 23 | Plugin architecture for detectors                  | M      | MEDIUM | Features      |
| P3       | 24 | Styled CLI output (`lipgloss`)                     | M      | LOW    | UX            |
| P3       | 25 | Progress reporting to Pipeline callbacks           | S      | MEDIUM | Features      |

---

## g) Top #1 Question I Cannot Figure Out Myself

### What is the v1.0 Release Criteria?

I cannot answer: **What is the minimum bar for shipping v1.0?**

The TODO list has 30+ open items across P0–P3. Without explicit v1.0 criteria, every session adds more items than it completes. Specific questions:

1. **Is API lock a prerequisite?** If yes, the `NewFinding` API decision, `Confidence` type, and `Properties` map must all happen first. That's a significant breaking-change window.
2. **Is triage split brain a release blocker?** It's the #1 correctness risk, but it only manifests when fix strategies and pipeline triage disagree — an edge case.
3. **What about `go.yaml.in/yaml/v3`?** It's a banned dependency per project policy, but it only affects the CLI, not the library. Ship v1.0 with it and fix in v1.1?
4. **Is 90%+ test coverage required?** Current coverage is strong overall but has specific gaps (SARIF error paths, context cancellation).
5. **Should v1.0 ship without domain-specific providers (Go AST, Rust syn)?** The FixProvider plugin architecture is done, but no language-specific providers exist yet.

**My recommendation:** Define a v1.0 release criteria document with explicit checkboxes. Without it, we're optimizing for completeness over shipping.

---

## Session Metrics

| Metric               | Value                                        |
| -------------------- | -------------------------------------------- |
| Production files     | 44 (+2 from CLI split)                       |
| Test files           | 60                                           |
| Production LOC       | 6,662 (net decrease from deprecated removal) |
| Test LOC             | 16,547                                       |
| Test:Code ratio      | 2.49:1                                       |
| Tests passing        | 502                                          |
| Commits this session | 6                                            |
| Commits unpushed     | 1 (`58fb0d3` — copylocks fix)                |
| LSP stale warnings   | 13 (all stale gopls cache, `go vet` clean)   |

---

## Unpushed Commit

```
58fb0d3 fix(test): resolve go vet copylocks warning from Report sync.Mutex
```

This commit fixes `json_test.go` to use `json.Marshal(&orig)` instead of `json.Marshal(orig)`, resolving the `copylocks` warning introduced by the `sync.Mutex` value type change. Build + vet + tests all green. **Not pushed** — awaiting user decision.

---

_Assisted-by: Crush <crush@charm.land>_
