# go-finding SDK Readiness Report

**Date:** 2026-04-24
**Assessed by:** Crush

## Verdict: Conditionally Ready

Strong core library with real bugs that needed fixing. After applying 5 blocking fixes, the SDK is viable for early adopters, but several medium-priority issues remain before it's production-grade.

---

## Blocking Issues Found & Fixed

| #   | Issue                                                                                                                                                                                                                               | Severity | Fix                                                     |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------- |
| 1   | **Pipeline tests didn't compile** — unused `"errors"` import in `verify_test.go`                                                                                                                                                    | Blocker  | Removed import                                          |
| 2   | **Data races in CLI tests** — `TestRun_BadSeverity`, `TestRun_MissingConfig`, `TestRun_NoDetectors`, `TestRun_BadProfilingPath` all called `t.Parallel()` while modifying global state (`flag.CommandLine`, `os.Args`, `os.Stderr`) | Blocker  | Removed `t.Parallel()` from tests touching global state |
| 3   | **`TestFatalf` race** — replaced `os.Stderr` globally while other parallel tests wrote to it                                                                                                                                        | Blocker  | Removed `t.Parallel()`                                  |
| 4   | **Stale `AGENTS.md`** — referenced `pipeline/astfix.go` which doesn't exist                                                                                                                                                         | Medium   | Updated to `pipeline/fix_applier.go`                    |
| 5   | **Outdated README stats** — showed 81.1% total coverage, actual is 94.8%                                                                                                                                                            | Medium   | Updated to real numbers                                 |

---

## Test & Coverage Status

| Package            | Coverage  | Status                    |
| ------------------ | --------- | ------------------------- |
| **Root (finding)** | **98.7%** | Excellent                 |
| **Pipeline**       | **94.5%** | Very good                 |
| **Detectors**      | **95.9%** | Very good                 |
| **CLI**            | **77.3%** | Adequate                  |
| **Total**          | **94.8%** | Passes CI threshold (75%) |

- All tests pass with `-race`
- Fuzz tests exist for ID generation, merge, and SARIF parsing
- Property-based tests via `testing/quick` for filter/group/merge
- Benchmarks exist for hot paths (ID gen, filter, merge, SARIF, parallel detection)

---

## Core Library Assessment: Strong

### What works well

- **Clean API design** — `Finding`, `Report`, `Severity`, `Position` types are well-structured with proper methods (`IsValid`, `Clone`, `Equal`, `Compare`)
- **String-based enums** — Human-readable JSON, no int↔string confusion
- **Lossless SARIF round-trip** — Non-standard fields preserved in `Properties["go-finding/*"]`
- **Composable filtering** — `Filter(findings, BySeverity(...), NotSuppressed, HasFix)` is elegant
- **Zero stdlib deps for core** — Only `golang.org/x/sync` and `golang.org/x/tools` for pipeline/analysis
- **Structured errors** — `FindingError` with categories, `errors.Is()` support, positional context
- **Go 1.26 idioms** — `iter.Seq[Finding]` on `Report.All()`, `slices.SortFunc`, `maps.Clone`
- **Thread-safe Metrics** — `sync.Mutex` throughout, `Snapshot()` for point-in-time copies

### Known limitations (documented)

- **SARIF critical round-trip** — `SeverityCritical → "error" → SeverityError` is lossy (SARIF 2.1.0 has no "critical" level). Documented and preserved in properties.

---

## Pipeline Assessment: Good, with caveats

| Feature                       | Status                                                                                             |
| ----------------------------- | -------------------------------------------------------------------------------------------------- |
| Parallel detection (errgroup) | Working, benchmarks show 5x speedup                                                                |
| Conflict detection            | Working, overlapping fixes filtered                                                                |
| Fix application (line-based)  | Working with backup/rollback                                                                       |
| Verification                  | Working, re-runs detectors                                                                         |
| Retry with backoff            | Working, validated config                                                                          |
| Partial success               | Working, `PartialErrors` surfaced                                                                  |
| Metrics                       | Working, snapshot in result                                                                        |
| Dry run                       | Working                                                                                            |
| **OnFix callback**            | **Bug: reports success for ALL safeFixes, even those skipped by FixApplier (empty Position.File)** |
| **FixStrategyAI**             | **Phantom constant — no backend, triage treats it as Suggest**                                     |

---

## Remaining Issues (Not Fixed)

### Medium Priority

| Issue                                        | Details                                                                                                                                                                         |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **OnFix callback inaccuracy**                | `pipeline.go:505` calls `OnFix(f, true)` for all `safeFixes`, but some may not have been applied (e.g. empty `Position.File`). Should only callback for actually-applied fixes. |
| **FixStrategyAI is misleading**              | No AI backend exists. Users setting `FixStrategy: "ai"` get Suggest-equivalent behavior silently. Should either implement, warn, or remove.                                     |
| **FixApplier path traversal (gosec G703)**   | 3 `os.WriteFile` calls flagged. False positive for a tool designed to write fixes, but should add validation that paths stay within `rootDir`.                                  |
| **CLI `run()` uses global flag state**       | Not testable in parallel. Should accept `io.Writer` + `*flag.FlagSet` as parameters.                                                                                            |
| **`Report.AddFinding` not goroutine-safe**   | Documented but not enforced. Could cause subtle bugs in concurrent pipelines.                                                                                                   |
| **Pipeline not safe for concurrent `Run()`** | Documented but `Pipeline.findings` slice is shared state.                                                                                                                       |

### Low Priority

| Issue                        | Details                                                                                                                       |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| 15 remaining lint issues     | `paralleltest` (5, intentional), `gosec` (3, false positive), `gochecknoglobals` (2), `goconst` (2), `golines` (2 formatting) |
| `TODO_LIST.md` has ~80 items | Many are stale (e.g. "Fix DeduplicateByPosition" is already correct)                                                          |
| `Correlate()` O(n²)          | Noted in docs; fine for small-medium finding sets                                                                             |

---

## TODO_LIST.md Audit

Several items in the TODO list are **stale or incorrect**:

| TODO Item                                               | Actual Status                                                                                                                     |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| "Fix DeduplicateByPosition key should NOT include Rule" | Already correct — `DeduplicateByPosition` key is `file:line:col`, `DeduplicateByRule` is `rule:file:line:col`. They are distinct. |
| "Fix Metrics.StageTiming value receiver bug"            | Already correct — `StageTiming` uses pointer receiver `(m *Metrics)`. No mutex copy.                                              |
| "Fix profiling FD leak"                                 | Already correct — cleanup function closes FDs in reverse order, error paths also close.                                           |
| "Fix CI Go version matrix"                              | Already correct — `.github/workflows/ci.yml` uses Go 1.26.                                                                        |
| "Fix release workflow Go version"                       | Already correct — `release.yml` uses Go 1.26.                                                                                     |

---

## Documentation Assessment

| Document            | Status                                                                  |
| ------------------- | ----------------------------------------------------------------------- |
| README.md           | Good — quick start, API overview, badges, coverage stats (now accurate) |
| CHANGELOG.md        | Excellent — follows Keep a Changelog, 3 releases documented             |
| CONTRIBUTING.md     | Good — dev setup, coding standards, PR process                          |
| docs/USAGE_GUIDE.md | Comprehensive — all API surfaces documented with code examples          |
| doc.go              | Good — GoDoc with examples                                              |
| AGENTS.md           | Fixed — now references correct files                                    |
| config.example.yaml | Present                                                                 |

---

## CI/CD Assessment

| Component           | Status                                                                        |
| ------------------- | ----------------------------------------------------------------------------- |
| GitHub Actions CI   | Multi-OS (ubuntu + macos), Go 1.26, race detector, coverage enforcement (75%) |
| Codecov integration | Present                                                                       |
| golangci-lint       | 80+ linters enabled, strict config                                            |
| GoReleaser          | Configured for linux/darwin/windows, amd64/arm64                              |
| Release workflow    | Tag-triggered, runs tests + vet before release                                |

---

## Benchmark Results

| Benchmark            | ns/op      | B/op      | allocs/op |
| -------------------- | ---------- | --------- | --------- |
| GenerateID (hash)    | 169        | 120       | 5         |
| ParseID              | 224        | 312       | 4         |
| Filter               | 38,605     | 294,913   | 1         |
| FilterMultiple       | 39,500     | 294,913   | 1         |
| GroupByFile          | 75,506     | 425,225   | 125       |
| Merge                | 306,279    | 1,185,010 | 42        |
| Correlate            | 51,613     | 214,640   | 36        |
| ToSARIF              | 268,487    | 272,530   | 1,508     |
| FromSARIF            | 446,734    | 203,713   | 2,233     |
| Parallel detection   | 5,445,206  | 85,051    | 230       |
| Sequential detection | 26,492,439 | 80,784    | 142       |

Parallel detection shows **~5x speedup** over sequential.

---

## Summary Ratings

| Dimension                | Rating | Notes                                                                                                               |
| ------------------------ | ------ | ------------------------------------------------------------------------------------------------------------------- |
| **API Design**           | ★★★★☆  | Clean, idiomatic, composable. `FixStrategyAI` is misleading.                                                        |
| **Test Quality**         | ★★★★★  | 94.8% coverage, fuzz + property tests, race-clean                                                                   |
| **Code Quality**         | ★★★★☆  | Good overall. Lint close to zero (15 remaining, mostly intentional).                                                |
| **Documentation**        | ★★★★☆  | Thorough. Some stale references (now fixed).                                                                        |
| **Production Readiness** | ★★★☆☆  | Core types are solid. Pipeline OnFix bug and path-traversal flags need attention. CLI testability is weak.          |
| **Ecosystem Fit**        | ★★★★★  | Solves a real problem. 5 projects already listed as consumers. SARIF + LSP + go/analysis integrations are complete. |

---

## Recommendation

**The core `finding` package is ready for v0.x consumption.** The `pipeline` package needs the OnFix callback fix and a FixStrategyAI decision before it can be trusted in production automation. The CLI is functional but needs a testability refactor (dependency injection for `os.Stdout`/`flag.CommandLine`) to reach adequate test coverage.

### Must-fix before v1.0.0

1. Fix `OnFix` callback to only report actually-applied fixes
2. Decide on `FixStrategyAI` — implement, deprecate, or document as placeholder
3. Add path validation in `FixApplier` to prevent writes outside `rootDir`
4. Refactor CLI `run()` to accept dependencies for testability
5. Clean stale items from `TODO_LIST.md`

---

_Assisted-by: Crush <crush@charm.land>_
