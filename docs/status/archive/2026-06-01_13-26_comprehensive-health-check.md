# Status Report — go-finding

**Date:** 2026-06-01 13:26 CEST
**Version:** v0.4.2 (tagged)
**Since Last Report:** 2026-05-27 (gogenfilter v3 followup)

---

## Executive Summary

go-finding is **healthy and production-ready** for a v0.x library. All tests pass (with `-race`), zero lint warnings, coverage is excellent across all packages. The library has 7,905 lines of production Go and 19,048 lines of tests (2.4:1 test:code ratio) across 114 Go files. 220 commits since May 2026, 8 tagged releases.

**One active bug discovered during this review:** Pipeline benchmarks are broken due to the single-use `Run()` guard introduced in a previous session. 10 of 13 pipeline benchmarks fail with `"pipeline: Run already called"`. The benchmark helper creates one `Pipeline` but calls `Run()` `b.N` times.

---

## a) FULLY DONE

| Area                                       | Status   | Evidence                                                            |
| ------------------------------------------ | -------- | ------------------------------------------------------------------- |
| Core data model (Finding, Position, Range) | Complete | 98.7% coverage, STABLE                                              |
| Builder API                                | Complete | Fluent construction, validation, `MustBuild()`                      |
| Severity (4 levels)                        | Complete | Comparison operators, SARIF mapping                                 |
| FixStrategy (none/suggest/direct)          | Complete | Auto-fix for `direct`, RESERVED for `ai`                            |
| Category (14 standard + custom)            | Complete | `IsValid()`, `IsSecurity()`                                         |
| Tags (multi-label)                         | Complete | `IsStandard()`, `IsValid()`                                         |
| Suppression with TTL                       | Complete | `IsActive(now)`, expiry enforcement                                 |
| Report container (thread-safe)             | Complete | Mutex-protected, `All()` returns `iter.Seq`                         |
| Filtering & sorting                        | Complete | 11 predicates, `Negate`, `AnyOf`, `FilterInPlace` with GC safety    |
| Report merging & dedup                     | Complete | 3 strategies (ID, Position, Rule)                                   |
| Cross-tool correlation                     | Complete | Simple heuristic, capped at 10K                                     |
| ID generation & parsing                    | Complete | Hash-based fallback, Windows paths, length-prefixed collision-safe  |
| JSON serialization                         | Complete | Streaming, lossy `FromJSON` with `FilterInvalid`                    |
| SARIF 2.1.0 export/import                  | Complete | Round-trip via property bag, streaming output                       |
| LSP conversion                             | Complete | Bidirectional, documented lossiness                                 |
| go/analysis integration                    | Complete | Bidirectional Diagnostic ↔ Finding                                  |
| Structured errors                          | Complete | 5 categories, `errors.Is` support, `WithFinding`/`WithPosition`     |
| Pipeline (detect→fix→verify)               | Complete | Iterative loop, configurable, single-use guard                      |
| Finding processors                         | Complete | `ProcessorFunc`, `NamedProcessorFunc`, generated file filter        |
| Conflict detection                         | Complete | Overlapping fix detection, conservative resolution                  |
| FixEngine (byte-level)                     | Complete | Descending offset, frontier boundary                                |
| FixProvider chain                          | Complete | Offset → Line → Substring + custom                                  |
| FixApplier (filesystem)                    | Complete | Backup/rollback, path traversal protection, permission preservation |
| Verification stage                         | Complete | Diff-based: fixed/remaining/new                                     |
| Metrics collection                         | Complete | Thread-safe, snapshots, `TotalDuration` guard                       |
| Retry (exp backoff)                        | Complete | `math/rand/v2` jitter, per-detector timeouts                        |
| Partial success                            | Complete | Graceful degradation, `FormatPartialErrors`                         |
| File backup & rollback                     | Complete | Permission-preserving, `RollbackAll`                                |
| CLI tool                                   | Complete | 4 output formats, YAML config, profiling, plugin registry           |
| Generated file filtering                   | Complete | `gogenfilter/v3` integration, CLI flags                             |
| Examples (3)                               | Complete | Basic, Builder, Pipeline — compile-tested                           |
| `Confidence` named type                    | Complete | `IsValid()`, `Clamp()`, `Compare()`, `String()`                     |
| Diff function                              | Complete | `DiffResult.HasChanges()`, `Stats()`, `ModifiedPair`                |
| FormatText / FormatMarkdown                | Complete | Error-returning, UTF-8 safe, markdown escaping                      |
| Nix flake                                  | Complete | `flake.nix` with build, test, lint, bench                           |
| All tests pass                             | Complete | `go test -race -count=1 ./...` — green                              |
| Zero lint warnings                         | Complete | `nix run .#lint` — "0 issues"                                       |
| Open-source release prep                   | Complete | LICENSE, CONTRIBUTING, Code of Conduct                              |
| Research: go-error-family                  | Complete | Decision: do not adopt (documented reasoning)                       |

---

## b) PARTIALLY DONE

| Area                 | What's Done                                     | What's Missing                                                                                                          |
| -------------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Pipeline benchmarks  | FixEngine benchmarks pass (4/4)                 | Pipeline benchmarks broken (10/13 FAIL) — `runBenchPipeline` reuses single Pipeline instance, but `Run()` is single-use |
| Test coverage        | Root 98.7%, Pipeline 96.0%, Detectors 96.1%     | CLI at 69.8% — integration test gaps                                                                                    |
| Documentation        | doc.go, USAGE_GUIDE, FEATURES, README, AGENTS   | doc.go ~40% complete; USAGE_GUIDE not updated for v0.4.x                                                                |
| SARIF round-trip     | Export/import with property bag                 | `FixStrategySuggest` without `AfterCode` loses suggestion text                                                          |
| Error consistency    | `fix_applier.go` uses structured `FindingError` | `pipeline.go` still uses raw `fmt.Errorf` — inconsistent                                                                |
| `Equal()` in Finding | Works correctly with 9-condition boolean        | Complex, could use early-return helper for readability                                                                  |
| CI/CD                | Nix flake for local builds                      | No `.github/workflows/` — no automated CI                                                                               |

---

## c) NOT STARTED

| Item                                       | Priority        | Notes                                                |
| ------------------------------------------ | --------------- | ---------------------------------------------------- |
| Fix broken Pipeline benchmarks             | HIGH            | 10/13 fail — regression from single-use guard        |
| GitHub Actions CI                          | HIGH            | No `.github/workflows/` exists                       |
| GoReleaser multi-module config             | MEDIUM          | No `.goreleaser.yml`                                 |
| API stability review for v1.0.0            | MEDIUM          | All exported symbols need audit                      |
| API stability guarantee document           | MEDIUM          | Go compat promise style                              |
| v1.0.0 release criteria                    | MEDIUM          | No documented minimum bar                            |
| `go:generate stringer` for 4 enum types    | LOW             | Severity, FixStrategy, Category, SuppressionKind     |
| `iter.Seq[Finding]` on `Report.All()`      | LOW             | Go 1.26 feature — `All()` already returns `iter.Seq` |
| Spatial index for `Correlate`              | LOW             | Currently O(n²), capped at 10K                       |
| Streaming merge                            | LOW             | Process one finding at a time                        |
| Plugin architecture for external detectors | LOW             | Dynamic loading                                      |
| Pipeline middleware/interceptor pattern    | LOW             | Custom stage injection                               |
| IDE plugin stubs                           | OUT OF SCOPE v1 |                                                      |
| Web UI / Styled CLI / TUI                  | OUT OF SCOPE v1 |                                                      |

---

## d) TOTALLY FUCKED UP

### 1. Pipeline Benchmarks — REGRESSION (P0)

**10 of 13 pipeline benchmarks FAIL** with `"pipeline: Run already called; create a new Pipeline for each invocation"`.

**Root cause:** `pipeline_bench_test.go:72-76` — `runBenchPipeline` creates one `Pipeline` via `New()`, then calls `p.Run()` inside `for range b.N`. The single-use `ran` bool guard (introduced in `pipeline.go` to prevent accidental double-`Run()`) rejects the second iteration.

**Fix:** Move `New(pipelineCfg, b.TempDir(), detectors...)` inside the `for range b.N` loop (before `ResetTimer`, then reset timer, then loop).

**Only FixEngine benchmarks pass** because they don't go through `Pipeline.Run()`.

### 2. CLI Test Coverage at 69.8%

The `cmd/go-finding` package has the lowest coverage. Many code paths in `config.go` and `main.go` are only covered by integration tests that depend on external tools (`go vet`, `staticcheck`).

### 3. No CI — Zero Automation

There are no `.github/workflows/`. Every quality gate (test, lint, bench) is manual via `nix run .#*`. This means:

- No PR checks
- No benchmark regression tracking
- No automated release validation
- Fuzz corpus not persisted

### 4. `.golangci.yml` Auto-Format War

The config file has an ongoing fight between 2-space and 4-space indentation caused by auto-configure tools. Listed in TODO as MEDIUM but persists across sessions.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture & Design

1. **Centralize triage logic** — `HasFix()` is the canonical "is fixable?" check, but pipeline triage and FixEngine use different logic paths. Consolidate.
2. **Consistent structured errors in pipeline** — `fix_applier.go` uses `FindingError`; `pipeline.go` uses raw `fmt.Errorf`. Should be unified.
3. **FixApplier lifecycle** — Each iteration creates a new `FixApplier` with a temp backup dir. If iteration N+1 starts, iteration N's backup dir is orphaned. Lift to Pipeline constructor.
4. **`Report.Merge()` mutates receiver** — Should return a new `*Report` instead of mutating.
5. **`maxIterations: 0` inconsistency** — CLI defaults to 1, config defaults to 5. Different defaults in different places.

### Code Quality

6. **Decompose `findingFromSarResult`** — Cognitive complexity exceeds gocognit threshold.
7. **Decompose `applySarifProperties`** — Complexity 26 exceeds threshold of 25.
8. **Extract `Equal()` 9-condition boolean** into readable helper with early returns.
9. **Consolidate SARIF write methods** — 4 methods with DRY violations.
10. **Modernize to Go 1.21+ stdlib** — `slices.Contains`, `slices.Delete`, `maps.Keys` etc.
11. **Inline `lock()`/`unlock()` wrappers** in `report.go` — just use `r.mu.Lock()`/`Unlock()` directly.
12. **Extract `findingKey` to shared utility** — duplicated in `verify.go` and `merge.go`.

### Testing

13. **Fix the pipeline benchmarks** — P0 regression.
14. **Add concurrent Report read-write race test** — `report_test.go` doesn't have one.
15. **Add SARIF parser fuzz test** — For untrusted/malformed input.
16. **Verify `ParseID` with Windows paths** — `C:\Users\...` edge case.
17. **CLI integration test coverage** — Currently 69.8%, should target 85%+.

### Documentation

18. **Complete `doc.go`** — Currently ~40% complete per TODO_LIST.
19. **Update USAGE_GUIDE.md** — Not updated for v0.4.x features.
20. **Document provider chain** in user-facing docs.
21. **Add godoc examples** for key APIs (`ExampleBuilder`, `ExampleFilter`, `ExamplePipeline`).

### DevOps

22. **GitHub Actions CI** — test, lint, bench, fuzz on every PR.
23. **GoReleaser config** — Automated multi-platform releases.
24. **Persist fuzz corpus** — 17 fuzz targets with no checked-in seed corpus.

---

## f) Top #25 Things to Do Next

| #  | Item                                                               | Priority | Impact        | Effort     |
| -- | ------------------------------------------------------------------ | -------- | ------------- | ---------- |
| 1  | **Fix pipeline benchmarks** — move `New()` inside loop             | P0       | Correctness   | S (10 min) |
| 2  | **GitHub Actions CI** — test + lint + race on push/PR              | P0       | Reliability   | M (2-4 hr) |
| 3  | **Centralize triage logic** — make `HasFix()` canonical            | P1       | Architecture  | M (2 hr)   |
| 4  | **Consistent structured errors in pipeline**                       | P1       | Consistency   | M (2 hr)   |
| 5  | **FixApplier lifecycle** — lift to Pipeline constructor            | P1       | Correctness   | M (3 hr)   |
| 6  | **CLI test coverage → 85%+**                                       | P1       | Quality       | M (4 hr)   |
| 7  | **API stability review** — audit all exported symbols              | P1       | Release       | M (4 hr)   |
| 8  | **Decompose `findingFromSarResult`** — reduce cognitive complexity | P2       | Readability   | S (1 hr)   |
| 9  | **Decompose `applySarifProperties`** — reduce complexity below 25  | P2       | Readability   | S (30 min) |
| 10 | **Update USAGE_GUIDE.md for v0.4.x**                               | P2       | Docs          | M (2 hr)   |
| 11 | **Complete `doc.go`** — currently ~40%                             | P2       | Docs          | M (3 hr)   |
| 12 | **GoReleaser config**                                              | P2       | DevOps        | S (1 hr)   |
| 13 | **Concurrent Report race test**                                    | P2       | Correctness   | S (30 min) |
| 14 | **`Report.Merge()` returns new `*Report`**                         | P2       | API           | S (1 hr)   |
| 15 | **Modernize to Go 1.21+ stdlib** — `slices`, `maps`                | P2       | Modernization | M (3 hr)   |
| 16 | **SARIF parser fuzz test**                                         | P2       | Security      | S (1 hr)   |
| 17 | **Fix `maxIterations` default inconsistency** — CLI vs config      | P2       | Correctness   | S (15 min) |
| 18 | **Extract `Equal()` into readable helper**                         | P3       | Readability   | S (30 min) |
| 19 | **Document provider chain** in user-facing docs                    | P3       | Docs          | S (1 hr)   |
| 20 | **`go:generate stringer`** for 4 enum types                        | P3       | DevEx         | S (1 hr)   |
| 21 | **FixEngine: line-offset tracking** for cumulative shifts          | P3       | Correctness   | L (1 day)  |
| 22 | **`FixStrategySuggest` SARIF round-trip** — preserve suggestion    | P3       | Fidelity      | S (1 hr)   |
| 23 | **SARIF schema validation test** against JSON schema               | P3       | Correctness   | M (2 hr)   |
| 24 | **Define v1.0.0 release criteria**                                 | P3       | Planning      | S (30 min) |
| 25 | **Persist fuzz corpus** — seed corpus for 17 targets               | P3       | Reliability   | M (3 hr)   |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the Pipeline single-use guard be removed in favor of a `Reset()` method, or is the correct fix to change only the benchmarks?**

The single-use `ran` bool guard was added intentionally — `Pipeline` is stateful (metrics, iterations, results). Reusing it silently could produce confusing bugs. But the benchmark pattern of `New() → for { Run() }` is idiomatic Go benchmarking. Options:

1. **Fix benchmarks only** — move `New()` inside loop (simplest, keeps safety guard)
2. **Add `Reset()` method** — clears state, allows reuse (more flexible, but more surface area)
3. **Remove single-use guard** — document that Run() is not idempotent (risky)

This is an **owner decision** because it's an API design question with breaking-change implications.

---

## Project Health Dashboard

| Metric              | Value                           | Trend                              |
| ------------------- | ------------------------------- | ---------------------------------- |
| Version             | v0.4.2                          | Stable                             |
| Production Go lines | 7,905                           | Growing                            |
| Test lines          | 19,048                          | Growing (2.4:1 ratio)              |
| Go files            | 114                             | Growing                            |
| Test pass rate      | 100% (tests) / 23% (benchmarks) | Tests stable, benchmarks regressed |
| Root coverage       | 98.7%                           | Excellent                          |
| Pipeline coverage   | 96.0%                           | Excellent                          |
| CLI coverage        | 69.8%                           | Needs work                         |
| Detectors coverage  | 96.1%                           | Excellent                          |
| Lint warnings       | 0                               | Clean                              |
| Race detector       | Clean                           | Clean                              |
| Open TODO items     | ~97                             | Slowly shrinking                   |
| Dependencies        | 6 direct, minimal transitive    | Healthy                            |
| Git tags            | 8 (v0.1.0 → v0.4.2)             | Regular releases                   |
| Unpushed commits    | 0                               | Up to date                         |

---

_Assisted-by: Crush <crush@charm.land>_
