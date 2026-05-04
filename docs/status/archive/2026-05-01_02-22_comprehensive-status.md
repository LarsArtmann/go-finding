# Comprehensive Status Report — go-finding

**Date:** 2026-05-01 02:22
**Session:** TODO_LIST.md rebuild + full codebase audit
**Author:** Crush (AI assistant)
**Previous Report:** 2026-04-30 03:40

---

## Executive Summary

go-finding is in **strong shape**. All tests pass with race detection, zero lint issues, and high coverage across all packages. The codebase is well-structured with a mature pipeline, comprehensive test suites, and good CI infrastructure.

The primary gap is **version management**: the codebase has accumulated breaking changes (notably `NewFinding` signature changes) but the version is still `v0.1.3`. Several architectural decisions need finalization before an API-stable `v0.2.0` release.

This session rebuilt `TODO_LIST.md` from a full audit of all 56 `.md` files in the repository, cross-referencing every TODO claim against actual code.

---

## 1. Health Check

### Tests

| Package                                    | Status   | Coverage  |
| ------------------------------------------ | -------- | --------- |
| Root (`github.com/larsartmann/go-finding`) | ✅ PASS  | **99.5%** |
| `cmd/go-finding`                           | ✅ PASS  | **95.4%** |
| `pipeline`                                 | ✅ PASS  | **98.0%** |
| `internal/detectors`                       | ✅ PASS  | **96.1%** |
| `examples` (compile check)                 | ✅ PASS  | N/A       |
| **Race detector**                          | ✅ Clean | —         |

### Lint

| Tool                      | Result          |
| ------------------------- | --------------- |
| `golangci-lint run ./...` | ✅ **0 issues** |

### CI

| Job                          | Status        |
| ---------------------------- | ------------- |
| Test (Go 1.26, ubuntu/macos) | ✅ Configured |
| Coverage enforcement         | ✅ Total ≥75% |
| `golangci-lint`              | ✅ Configured |
| `govulncheck`                | ✅ Configured |
| Stress test (`-count=20`)    | ✅ Configured |

### Benchmarks

| Operation                  | Before (ns/op) | After (ns/op) | Δ               | allocs (before→after) |
| -------------------------- | -------------- | ------------- | --------------- | --------------------- |
| `GenerateID`               | 43             | 53            | baseline        | 2→2                   |
| `Filter` (1000)            | 49,750         | 53,722        | baseline        | 1→1                   |
| **`Merge` (dedup, 5×200)** | **408,277**    | **172,900**   | **2.4x faster** | **46→20**             |
| **`MergeNoDedup` (3×500)** | **442,624**    | **217,977**   | **2.0x faster** | **24→17**             |
| `ToSARIF`                  | 278,814        | 349,024       | baseline        | 1508→1508             |
| `FromSARIF`                | 469,279        | 575,333       | baseline        | 2233→2233             |
| `Correlate`                | 55,941         | 70,008        | baseline        | 37→37                 |
| `Clone`                    | 65             | 74            | baseline        | 2→2                   |
| `FindingKey`               | 3              | 3             | unchanged       | 0→0                   |
| Parallel detection         | 5,504,439      | 5,479,774     | unchanged       | 256→256               |

**Merge optimization:** Pre-allocated the findings slice to total capacity and skipped per-item mutex locking. Eliminated ~10 slice reallocations and 1000 mutex lock/unlock cycles per merge.

### Code Stats

- **~19,919 lines** total Go code
- **0 lint issues**
- **4 packages** with tests
- **1 phantom dep**: `golang.org/x/tools` (12MB) via `diagnostic.go`

---

## 2. A) FULLY DONE ✅

These items are verified complete in the current codebase — confirmed by reading actual code files, not just status reports.

### Core Types & API

- ✅ `Finding` type with full field set (Identity, Classification, Fix, Context, Metadata)
- ✅ `Finding.Builder` fluent API with `Build() (Finding, error)` return
- ✅ `Finding.Key()` stable identity method
- ✅ `Finding.Equal()` with epsilon-based float comparison
- ✅ `Finding.Validate()` method for structural validation
- ✅ `Finding.Clone()` deep copy
- ✅ `NewFinding()` with 6-param signature including confidence clamping to [0.0, 1.0]
- ✅ `Severity` type with `Compare()`, `LessThan()`, `IsValid()`, `String()`
- ✅ `FixStrategy` with `HasFix()` correctly handling AI as Suggest-equivalent
- ✅ `Tag` type with standard constants + `Tags []Tag` field (deprecated `Tag string`)
- ✅ `Category` type with standard constants and `IsStandard()`
- ✅ `Suppression` with `IsExpired()`, `IsValid()`
- ✅ `Range` with `Overlaps()`, `Intersection()`, `Adjacent()`, `Contains()`, `LineCount()`
- ✅ `Position` with zero-value handling
- ✅ `RelatedRef` for cross-finding links
- ✅ `Correlation` type with camelCase JSON tags

### Report

- ✅ `Report` with thread-safe `AddFinding()`/`AddFindings()` (sync.Mutex)
- ✅ `Report.All()`, `FindByID()`, `FindByRule()`, `ActiveFindings()`
- ✅ `Report.Filter(pred func(Finding) bool) *Report`
- ✅ `Report.Map(fn func(Finding) Finding) *Report`
- ✅ `Report.WriteSARIF(w io.Writer) error`
- ✅ `Report.WriteSARIFFiltered(w io.Writer, ...) error`
- ✅ `Report.Summary` auto-computed
- ✅ Package-level `Merge()` with deduplication strategies
- ✅ `Correlate()` with O(n²) hard-limited to 10k

### Pipeline

- ✅ Full pipeline: detect → triage → fix → verify loop
- ✅ Parallel detection via `errgroup` with `ParallelDetectors` config
- ✅ `FixApplier` with backup/rollback support, line-based + string replacement
- ✅ `FixEngine` for range-based and string-based fixes
- ✅ `FileBackup` with enable/disable, backup+restore
- ✅ Conflict detection: `FilterConflictingFixes()` + `AnalyzeConflicts()`
- ✅ Verification stage: re-run detectors, diff findings
- ✅ Retry with exponential backoff using `math/rand/v2`
- ✅ Partial success: collect from failed detectors
- ✅ `CorrelateFindings` config field wired into pipeline
- ✅ `callbackMu sync.Mutex` for concurrent `OnFinding` safety
- ✅ `PipelineResult.PartialErrors` for per-detector failures
- ✅ `PipelineResult.Metrics` auto-populated

### CLI

- ✅ Functional CLI with govet + staticcheck detectors
- ✅ Text, JSON, SARIF output formats
- ✅ YAML/JSON config file support
- ✅ Severity filtering, timeout, max-iterations
- ✅ CPU/memory profiling
- ✅ Graceful degradation on detector failures
- ✅ Metrics summary to stderr

### Output Formats

- ✅ SARIF 2.1.0 export with round-trip loss documentation
- ✅ `FindingsFromSARIF()` import with helper decomposition
- ✅ `FuzzFindingsFromSARIF` — 1.6M execs, zero panics
- ✅ LSP Diagnostic conversion
- ✅ `go/analysis.Diagnostic` conversion

### Infrastructure

- ✅ `.github/workflows/ci.yml` — test, coverage, lint, govulncheck, stress
- ✅ `.github/workflows/release.yml` — tag-triggered GoReleaser
- ✅ `.goreleaser.yml` — cross-platform (linux/darwin/windows, amd64+arm64)
- ✅ `go.work` for local development
- ✅ `scripts/coverage-check.sh` — per-package thresholds
- ✅ `version.go` with semver constants
- ✅ `.gitignore` for binaries
- ✅ `examples/` with compile checks
- ✅ `docs/integration-guide.md`
- ✅ `docs/release-procedure.md`
- ✅ `docs/architecture-decisions.md`
- ✅ `CONTRIBUTING.md`
- ✅ `FEATURES.md` (new — comprehensive feature inventory)

### Testing

- ✅ 99.5% root package, 98.0% pipeline, 96.1% detectors, 95.4% CLI
- ✅ Builder.Build() error-path tests
- ✅ Finding.Equal field-mismatch tests (18 cases)
- ✅ DeduplicateStrategies behavior diff test
- ✅ SARIF import + round-trip tests
- ✅ FixEngine tests (262 lines)
- ✅ FileBackup tests (160 lines)
- ✅ FixApplier tests (19 functions)
- ✅ RetryConfig.Validate tests (7 functions)
- ✅ Verifier error-path tests
- ✅ `Range.Contains` edge cases
- ✅ Property-based tests for ID generation
- ✅ Fuzz tests for SARIF and merge
- ✅ Benchmarks for all hot paths
- ✅ No `t.Parallel()` on tests that mutate globals

---

## 3. B) PARTIALLY DONE ⚠️

### API Stability

- ⚠️ **`NewFinding` API pattern** — Signature changed (5→6 params with confidence), but no decision on long-term pattern (functional options? builder-only?). Breaking change happened without formal decision.
- ⚠️ **`Tag` deprecation** — `Tag string` field marked deprecated, `Tags []Tag` added, but tests and examples still use `Tag` extensively. `WithTag()` builder method not deprecated.
- ⚠️ **`FixStrategyAI`** — Split brain fixed in `HasFix()`, but no user-facing documentation. ADR exists in `architecture-decisions.md` but semantics not in godoc or USAGE_GUIDE.

### Coverage

- ⚠️ **`WriteSARIF` error path** — 75% coverage. Main logic covered but `io.Writer` failure path untested.
- ⚠️ **`detectPartialSequential/Parallel`** — 90–94% coverage. Context cancellation paths untested.
- ⚠️ **`setupProfiling`** — 88.5%. Memory profile file creation error path untested.
- ⚠️ **`NewGoVetDetector`/`NewStaticcheckDetector`** — 80–90%. Binary-not-found error paths partially covered.

### Documentation

- ⚠️ **SARIF round-trip losses** — Documented in `sarif.go` godoc but NOT in user-facing docs (USAGE_GUIDE, README).
- ⚠️ **`FixStrategyAI` semantics** — In architecture-decisions.md but not in user-facing docs.

---

## 4. C) NOT STARTED ⬜

### P0 — Must Do

- ⬜ **Bump version to v0.2.0** — Version still `v0.1.3` despite breaking changes
- ⬜ **Release `[Unreleased]` in CHANGELOG.md** — Move to versioned entry
- ⬜ **Decide `NewFinding` API pattern** — Functional options vs current approach
- ⬜ **Extract `diagnostic.go` to subpackage** — Removes 12MB dep from core
- ⬜ **API stability review** — Audit all exported symbols for v1.0.0 lock

### P1 — Should Do

- ⬜ **Fix `Pipeline.Run()` mutability** — Mutates internal state per invocation
- ⬜ **Add `Properties map[string]any`** — Structured round-trip alongside `Metadata map[string]string`
- ⬜ **Add `Suppression.IsActive()`** — Combined expiry + validity check
- ⬜ **Add `Report.Merge(other *Report)` method** — In-place merge
- ⬜ **Confidence strong type** — `type Confidence float64` with validation
- ⬜ **Refactor CLI `run()` for testability** — Inject io.Writer + FlagSet
- ⬜ **Integration tests for real govet/staticcheck** — Currently all mocked
- ⬜ **WriteSARIF error-path test** — failingWriter pattern
- ⬜ **Detect context-cancel tests** — sequential + parallel paths

### P2 — Nice to Have

- ⬜ **Per-package coverage thresholds in CI** — Script exists but not wired to CI
- ⬜ **Benchmark regression tracking** — No `scripts/bench-compare.sh`
- ⬜ **SARIF schema validation test** — Requires downloading JSON schema
- ⬜ **`go/analysis` reverse conversion** — Noted in README as unsupported
- ⬜ **Nix migration** — Full proposal exists, Phase 0 not started
- ⬜ **`Finding` struct sub-grouping** — Deferred to v2 (breaking change)
- ⬜ **Consumer migration guide** — For v0.1.3 → v0.2.0 upgrade

---

## 5. D) TOTALLY FUCKED UP 💥

### Version Management

- 💥 **Version is still `v0.1.3`** despite multiple breaking API changes. `NewFinding` went from 5 to 6 parameters. `Builder.Build()` changed from panic to error return. `Tags []Tag` was added. Anyone updating `go-finding` will have broken code and no migration guide.
- 💥 **`CHANGELOG.md` has `[Unreleased]`** section that has been accumulating changes since April 28. No versioned release has been cut.

### Architectural Debt

- 💥 **`Pipeline.Run()` is stateful** — Mutates `p.findings` and `p.iterations` on each call. Creating a Pipeline and calling `Run()` twice shares state between invocations. This is documented but not enforced. Either make it immutable or make it a one-shot.
- 💥 **12MB phantom dependency** — `golang.org/x/tools` is imported solely for `diagnostic.go`'s `go/analysis` integration. Every consumer of the core `finding` package pulls in 12MB of transitive deps for a single file they may never use.

### Admitted Past Failures (from status reports)

- 💥 **`NewFinding` breaking change was shipped without decision** — Multiple status reports (04-30 02:59 especially) call this out as a "Totally Fucked Up" moment. The signature was changed from 5 to 6 params without deciding on the long-term API pattern (functional options, builder-only, etc.).

---

## 6. E) WHAT WE SHOULD IMPROVE

### Process

1. **Stop deferring version bumps** — Every breaking change should bump the minor version immediately. `v0.1.3` → `v0.2.0` should have happened weeks ago.
2. **Make decisions before shipping** — `NewFinding` signature change, `FixStrategyAI` semantics, API stability — these were debated across 10+ sessions without resolution while code was already changed.
3. **One source of truth for TODOs** — Before this session, TODOs were scattered across 56 `.md` files with massive duplication and contradictory statuses. The rebuilt `TODO_LIST.md` is now the single source.
4. **Reduce status report proliferation** — 26 status reports in 2 days (04-29 to 04-30). Many are redundant. Consolidate to one per session.

### Code

5. **Extract `diagnostic.go`** — The 12MB `golang.org/x/tools` dependency is the single biggest dependency hygiene issue. Move to `finding/analysis` subpackage.
6. **`Pipeline.Run()` immutability** — Either return a new `PipelineResult` without mutating the Pipeline, or document that Pipeline is single-use.
7. **`Confidence` as bare `float64`** — No compile-time protection against `Finding{Confidence: 1.5}`. Strong type or builder-only construction.
8. **`Tag` deprecation is half-done** — Field deprecated but builder method and tests not updated. Either commit to the migration or revert.

### Architecture

9. **Plugin architecture for detectors** — `knownDetectorBuilders` is a hardcoded map. Runtime registration would allow external tools to integrate without forking.
10. **Pipeline middleware** — Stages are hardcoded (detect → triage → fix → verify). Custom stage injection would enable use cases like AI fix routing, custom filtering, etc.

---

## 7. F) TOP #25 THINGS TO DO NEXT

Ordered by impact × effort ratio (highest first):

| #   | Task                                                         | Priority | Effort | Impact   | Package            |
| --- | ------------------------------------------------------------ | -------- | ------ | -------- | ------------------ |
| 1   | **Bump version to v0.2.0** in `version.go`                   | P0       | 2min   | Critical | root               |
| 2   | **Release `[Unreleased]` in CHANGELOG.md** as v0.2.0         | P0       | 5min   | Critical | docs               |
| 3   | **Write consumer migration guide** (v0.1.3 → v0.2.0)         | P0       | 30min  | High     | docs               |
| 4   | **Extract `diagnostic.go` to `finding/analysis` subpackage** | P0       | 60min  | High     | root               |
| 5   | **Decide `NewFinding` API pattern** and document in ADR      | P0       | 30min  | High     | root               |
| 6   | **API stability review** — audit all exported symbols        | P0       | 60min  | High     | all                |
| 7   | **Fix `Pipeline.Run()` mutability** — single-use or copy     | P1       | 30min  | High     | pipeline           |
| 8   | **Deprecate `WithTag()` builder method**                     | P1       | 5min   | Medium   | root               |
| 9   | **Add `Suppression.IsActive()` method**                      | P1       | 10min  | Medium   | root               |
| 10  | **Add `WriteSARIF` error-path test** (failingWriter)         | P1       | 15min  | Medium   | root               |
| 11  | **Add `detectPartial*` context-cancel tests**                | P1       | 20min  | Medium   | pipeline           |
| 12  | **Integration tests with real govet/staticcheck**            | P1       | 60min  | High     | internal/detectors |
| 13  | **Refactor CLI `run()` for testability**                     | P1       | 45min  | Medium   | cmd                |
| 14  | **Add `Properties map[string]any` to Finding**               | P1       | 20min  | High     | root               |
| 15  | **Add per-package coverage thresholds to CI**                | P2       | 20min  | Medium   | CI                 |
| 16  | **SARIF schema validation test**                             | P2       | 30min  | Medium   | root               |
| 17  | **Benchmark regression tracking script**                     | P2       | 30min  | Low      | scripts            |
| 18  | **Confidence strong type**                                   | P1       | 45min  | Medium   | root               |
| 19  | **Add Nix section to CONTRIBUTING.md**                       | P2       | 15min  | Low      | docs               |
| 20  | **Document SARIF round-trip losses in user-facing docs**     | P2       | 15min  | Medium   | docs               |
| 21  | **`FixApplier` cross-iteration persistence design**          | P1       | 60min  | High     | pipeline           |
| 22  | **Error wrapping consistency audit**                         | P2       | 30min  | Medium   | all                |
| 23  | **Unify `Tag` deprecation** — migrate tests to `Tags`        | P1       | 45min  | Medium   | tests              |
| 24  | **Profile performance at 10k+ findings**                     | P2       | 60min  | Medium   | pipeline           |
| 25  | **Consumer migration guide** (v0.1.3 → v0.2.0)               | P0       | 30min  | Critical | docs               |

---

## 8. G) TOP #1 QUESTION

**Should `v0.2.0` be cut NOW with the current state, or should the `diagnostic.go` extraction and `NewFinding` API decision happen first?**

Arguments for cutting now:

- Breaking changes are already shipped but unreleased
- Consumers who update get breakage regardless of version number
- A v0.2.0 tag at least signals "breaking changes here"

Arguments for waiting:

- Extracting `diagnostic.go` is itself a breaking change (import path changes)
- `NewFinding` API pattern is undecided — another signature change would mean v0.3.0
- Better to batch all breaking changes into one release

This decision blocks items 1–6 in the Top 25 list and determines whether this is a "stabilize and release" cycle or a "break more things then stabilize" cycle.

---

## 9. TODO_LIST.md Status

The `TODO_LIST.md` was fully rebuilt this session:

| Category                   | Count            |
| -------------------------- | ---------------- |
| P0 (Must Do)               | 5 open           |
| P1 (Should Do)             | 14 open          |
| P2 (Nice to Have)          | 12 open          |
| P3 (Future/Deferred)       | 22 open          |
| Verified Completed         | 55 items         |
| Out of Scope               | 11 items         |
| **Source files processed** | **56 .md files** |

All items were cross-referenced against actual code to verify completion status.

---

## 10. Files Changed This Session

| File                                                   | Change                                                                                                               |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------- |
| `TODO_LIST.md`                                         | Rebuilt from 56-file audit. 266 insertions, 174 deletions. Deduplicated ~300+ scattered TODOs into prioritized list. |
| `report.go`                                            | Added `newReportWithCapacity()` and `addFindingUnchecked()` for internal batch operations                            |
| `merge.go`                                             | Optimized `Merge()`: pre-count findings, pre-allocate slice, skip per-item mutex lock                                |
| `docs/status/2026-05-01_02-22_comprehensive-status.md` | This file                                                                                                            |

---

## 11. Coverage Trend

| Package   | 04-28 | 04-30 | 05-01                |
| --------- | ----- | ----- | -------------------- |
| Root      | 93.4% | 99.1% | **99.5%**            |
| Pipeline  | 96.4% | 97.8% | **98.0%**            |
| Detectors | —     | 95.8% | **96.1%**            |
| CLI       | 78.0% | 81.1% | **95.4%**            |
| **Total** | —     | —     | **97.3%** (weighted) |

---

## 12. Dependency Profile

```
golang.org/x/tools   → 12MB (diagnostic.go only — extraction candidate)
golang.org/x/sync    → errgroup for parallel detection
gopkg.in/yaml.v3     → YAML config file parsing (CLI only)
```

Core `finding` package imports only stdlib except via `diagnostic.go`.

---

_Assisted-by: Crush <crush@charm.land>_
