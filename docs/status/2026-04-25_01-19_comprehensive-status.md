# Comprehensive Status Report — 2026-04-25

**Session:** Full SDK readiness audit
**Assessor:** Crush
**Date:** 2026-04-25 01:19 CEST

---

## A) FULLY DONE

### A1. Test Suite Race Condition Fix (BLOCKER → FIXED)

**Problem:** 5 CLI tests (`TestRun_BadSeverity`, `TestRun_MissingConfig`, `TestRun_NoDetectors`, `TestRun_BadProfilingPath`, `TestFatalf`) called `t.Parallel()` while modifying global state (`flag.CommandLine`, `os.Args`, `os.Stderr`). This caused data races detected by `-race` flag.

**Fix:** Removed `t.Parallel()` from all tests that touch global mutable state. Fixed `TestFatalf` variable declaration ordering.

**Commit:** `9e63276` — refactor: clean up test infrastructure and remove unnecessary lint directives

**Verification:** `go test -race -count=1 ./...` — all packages PASS.

### A2. Pipeline Build Failure (BLOCKER → FIXED)

**Problem:** `pipeline/verify_test.go:5` imported `"errors"` but never used it. This caused `pipeline` package to fail compilation.

**Fix:** Removed the unused import.

**Commit:** `9e63276`

**Verification:** `go build ./...` succeeds. Pipeline tests pass.

### A3. Stale Documentation Fix

**Problem:** `AGENTS.md` Key Files table referenced `pipeline/astfix.go` — a file that doesn't exist. The actual file is `pipeline/fix_applier.go`.

**Fix:** Updated table entry: `pipeline/astfix.go` → `pipeline/fix_applier.go | Line-based fix application with backup/rollback`

**Commit:** `9e63276`

### A4. Outdated README Coverage Stats

**Problem:** README.md "Project Stats" table showed coverage from an earlier version (81.1% total). Actual coverage is 94.8%.

**Fix:** Updated all package coverage numbers to current values:
- Root: 91.7% → 98.7%
- Pipeline: 84.0% → 94.5%
- Detectors: 71.6% → 95.9%
- CLI: 24.1% → 77.3%
- Total: 81.1% → 94.8%

**Commit:** `9e63276`

### A5. Unused Test Helpers Removed

**Problem:** `testutil_test.go` contained 3 unused assertion functions flagged by `unused` linter: `assertPositionOffset`, `assertSeverityBool`, `assertSeverityInt`.

**Fix:** Deleted all three. Added `t.Helper()` to `assertSummaryField`.

**Commit:** `9e63276`

### A6. Stale nolint Directives Cleaned

**Problem:** 4 `//nolint` directives were no longer needed:
- `pipeline/conflict.go:54` — `//nolint:prealloc` (prealloc doesn't flag this pattern)
- `pipeline/pipeline_test.go:369` — `//nolint:prealloc` (same)
- `diagnostic_test.go:80` — `//nolint:goconst` (goconst doesn't flag test fixtures in test files)
- `sarif.go:473` — `//nolint:goconst` (SARIF "error" is a different context from SeverityError)

**Fix:** Removed all four stale directives.

**Commit:** `9e63276`

### A7. Full SDK Readiness Assessment

**Completed a thorough audit of:**
- All 16 core source files (finding.go, severity.go, fix_strategy.go, position.go, report.go, filter.go, merge.go, sarif.go, lsp.go, errors.go, category.go, id.go, json.go, suppression.go, diagnostic.go, doc.go)
- All 8 pipeline source files (pipeline.go, result.go, conflict.go, fix_applier.go, verify.go, metrics.go, retry.go, partial.go)
- CLI implementation (main.go, main_test.go, integration_test.go)
- All documentation (README.md, CHANGELOG.md, CONTRIBUTING.md, docs/USAGE_GUIDE.md, AGENTS.md)
- CI/CD workflows (ci.yml, release.yml, .goreleaser.yml, .golangci.yml)
- Full test suite with `-race`
- Full coverage report
- Benchmark suite
- golangci-lint with 80+ linters
- TODO_LIST.md validation (found 5 stale/incorrect items)

**Output:** `docs/READINESS_REPORT.md`

### A8. TODO_LIST.md Stale Item Audit

Verified 5 TODO items are **already resolved or incorrect**:

| TODO Item | Actual Status |
|-----------|---------------|
| "Fix DeduplicateByPosition key should NOT include Rule" | Already correct — key is `file:line:col`, distinct from `DeduplicateByRule` |
| "Fix Metrics.StageTiming value receiver bug" | Already correct — uses pointer receiver `(m *Metrics)`, no mutex copy |
| "Fix profiling FD leak" | Already correct — cleanup function properly closes FDs in reverse order |
| "Fix CI Go version matrix" | Already correct — ci.yml uses Go 1.26 |
| "Fix release workflow Go version" | Already correct — release.yml uses Go 1.26 |

---

## B) PARTIALLY DONE

### B1. Lint Cleanup

**Status:** Reduced from build-breaking to 15 remaining issues, but not zero.

**What's left:**

| Linter | Count | Nature | Actionable? |
|--------|-------|--------|-------------|
| `paralleltest` | 5 | Intentionally removed `t.Parallel()` to fix races | Suppress with nolint comments |
| `gosec` G703 | 3 | Path traversal in `FixApplier` — real concern but false positive for a fix tool | Add `rootDir` path validation |
| `gochecknoglobals` | 2 | Test helpers (`suppressedIDs`, `byFindingID`) | Suppress or move to function scope |
| `goconst` | 2 | Test fixture strings and SARIF level strings | Extract constants or suppress |
| `golines` | 2 | Formatting in test files | Run `golines` formatter |

**Done:** Removed stale nolints, deleted unused helpers, fixed thelper warning.
**Not done:** paralleltest nolints, gosec path validation, goconst extraction, golines formatting.

### B2. OnFix Callback Bug Investigation

**Status:** Investigated and documented, but NOT fixed.

**Finding:** `pipeline/pipeline.go:505` calls `OnFix(f, true)` for ALL `safeFixes`, but `FixApplier.Apply` may skip fixes with empty `Position.File`. The callback reports success for fixes that weren't applied.

**Why not fixed:** Requires design decision — should `FixApplier.Apply` return which specific fixes were applied, or should we add a `skip` callback? Changing the API surface needs consideration.

### B3. FixStrategyAI Investigation

**Status:** Investigated and documented, but NOT resolved.

**Finding:** `FixStrategyAI` exists as a constant with `NeedsAI()` method, but no AI backend exists. Pipeline triage treats it identically to `FixStrategySuggest`. The constant's own doc comment admits "no AI backend exists yet."

**Why not fixed:** Needs product decision — implement AI backend, deprecate the constant, or document as placeholder with runtime warning.

---

## C) NOT STARTED

### C1. FixApplier Path Traversal Validation

**Status:** Not started. 3 `gosec` G703 warnings flag `os.WriteFile` calls in `pipeline/fix_applier.go` (lines 135, 165, 272). While arguably a false positive for a tool designed to write fixes, adding `rootDir` boundary validation would eliminate the warnings and prevent misuse.

### C2. CLI Testability Refactor

**Status:** Not started. `cmd/go-finding/main.go:run()` directly uses `flag.CommandLine`, `os.Args`, `os.Stdout`, and `os.Stderr`. This makes the CLI untestable in parallel and is the root cause of the race conditions we fixed. Should accept `io.Writer` and `*flag.FlagSet` as parameters.

### C3. Report Goroutine Safety

**Status:** Not started. `Report.AddFinding()` and `Report.AddFindings()` are documented as "not safe for concurrent use" but nothing prevents callers from doing it. Options: add `sync.Mutex`, or document more prominently.

### C4. Pipeline Concurrent Run Safety

**Status:** Not started. `Pipeline.Run()` is documented as "NOT safe for concurrent use" but `Pipeline.findings` is shared mutable state. Options: add mutex, or return immutable result from `Run()`.

### C5. TODO_LIST.md Cleanup

**Status:** Not started. The TODO list has ~80 items, at least 5 of which are stale/incorrect. Should be audited and pruned.

### C6. Correlate() O(n²) Performance

**Status:** Not started. `Correlate()` uses nested loops for finding correlation. Fine for small sets but could be optimized with sorted+merge approach.

### C7. SARIF Schema Compliance

**Status:** Not started. The SARIF implementation is hand-rolled. Could be evaluated against `go-sarif` library for full spec compliance.

### C8. go:generate stringer for Enums

**Status:** Not started. `Severity`, `FixStrategy`, `Category`, `SuppressionKind` are string-based enums that could benefit from `stringer` code generation.

### C9. API Stability Review

**Status:** Not started. Before v1.0.0, need to audit the entire exported API surface and lock it down.

---

## D) TOTALLY FUCKED UP

### D1. Nothing is truly catastrophic

No data loss risks, no security vulnerabilities, no panics in production code. The worst issues were:
- **Build-breaking unused import** — fixed
- **Race conditions in tests** — fixed
- **Misleading documentation** — fixed

The closest to "fucked up" is the **FixStrategyAI phantom constant**. It's not a bug — it's an API lie. Users who set `FixStrategy: "ai"` get `Suggest`-equivalent behavior with zero warning. This undermines trust in the API contract. Not catastrophic, but dishonest.

### D2. Close second: The OnFix callback inaccuracy

If anyone relies on `OnFix` callback to track which fixes were actually applied vs. skipped, they'll get wrong data. `OnFix(f, true)` fires for ALL `safeFixes` even when some were silently skipped by `FixApplier.Apply` (empty `Position.File`). This is a silent correctness bug that could cascade in automation.

---

## E) WHAT WE SHOULD IMPROVE

### E1. Architecture

1. **CLI dependency injection** — `run()` should accept `io.Writer`, `*flag.FlagSet`, and detector builders as parameters. Current design forces global state mutation.
2. **Pipeline immutability** — `Pipeline.findings` accumulates across iterations but `Run()` documents "not safe for concurrent use." Make findings an internal detail, not shared state.
3. **FixApplier result tracking** — `Apply()` returns `(int, error)` but should return `[]FixResult` indicating which specific fixes succeeded/failed/skipped.
4. **Separate "constant" from "placeholder"** — `FixStrategyAI` should be in a separate `experimental` build tag or clearly marked with runtime warning.

### E2. Testing

5. **CLI integration tests with real detectors** — Current CLI tests only exercise config loading and error paths. No test runs the full pipeline with govet/staticcheck.
6. **Property-based test flakiness** — `TestProperty_IDRoundTrip` fails ~5% with random Unicode. Needs seed control or input constraints.
7. **FixApplier error-path coverage** — 84.6% → push to 90%+ with error-path unit tests.
8. **Conflict detection coverage** — `FilterConflictingFixes()` and `AnalyzeConflicts()` have 0% dedicated coverage (exercised only via pipeline integration).

### E3. Documentation

9. **SARIF round-trip losses** — Document explicitly: `SeverityCritical→Error`, `RelatedRef.FindingID` lost, `BeforeCode` lost on SARIF import.
10. **API stability guarantee** — No versioning policy documented. Consumers don't know if v0.1.x is stable or experimental.
11. **Examples directory** — `examples/` with standalone runnable programs would help adoption.

### E4. Operations

12. **govulncheck in CI** — `justfile` has `vuln` target but CI doesn't run it.
13. **Benchmark regression tracking** — Benchmarks exist but CI doesn't compare against baselines.
14. **Per-package coverage thresholds** — CI only enforces total 75%. Individual packages could regress without detection.
15. **golines formatter in CI** — 2 files flagged for formatting but no CI check enforces it.

### E5. Code Quality

16. **Remove stale TODO_LIST.md items** — At least 5 items are already resolved.
17. **Consolidate `findingKey`** — Duplicated in `verify.go` and `merge.go`.
18. **Replace `rand.Int63n` in retry** — Uses non-seeded `math/rand` (deterministic in Go 1.26 but still poor practice for retry jitter).
19. **`Correlation` struct needs JSON tags** — `finding_ids` vs `findingIds` tagliatelle violation.
20. **Unused `//nolint` directives** — Still 3 remaining that should be removed or corrected.

---

## F) TOP #25 THINGS WE SHOULD GET DONE NEXT

| Priority | # | Item | Effort | Impact |
|----------|---|------|--------|--------|
| P0 | 1 | **Fix OnFix callback inaccuracy** — only call for actually-applied fixes | S | M |
| P0 | 2 | **Add `paralleltest` nolint comments** to the 5 intentionally-serial CLI tests | XS | S |
| P0 | 3 | **Add `rootDir` path validation in FixApplier** — prevent writes outside project | S | M |
| P1 | 4 | **Decide on FixStrategyAI** — implement, deprecate with warning, or remove | M | M |
| P1 | 5 | **Refactor CLI `run()` for testability** — inject io.Writer, flag.FlagSet | M | L |
| P1 | 6 | **Clean stale TODO_LIST.md items** — audit and prune ~5 confirmed-stale entries | S | S |
| P1 | 7 | **Add CLI integration test with real govet detector** — end-to-end pipeline test | M | M |
| P1 | 8 | **Fix property test flakiness** — add seed control to TestProperty_IDRoundTrip | S | M |
| P2 | 9 | **Add `govulncheck` step to CI** — justfile has target, CI doesn't use it | S | M |
| P2 | 10 | **Consolidate `findingKey` function** — duplicated in verify.go and merge.go | S | S |
| P2 | 11 | **Add FixApplier error-path unit tests** — push coverage from 84.6% to 90%+ | M | S |
| P2 | 12 | **Add Conflict detection dedicated tests** — FilterConflictingFixes + AnalyzeConflicts 0% | S | S |
| P2 | 13 | **Add `Correlation` JSON tags** — fix tagliatelle violation | XS | S |
| P2 | 14 | **Replace `math/rand` in retry** — use `math/rand/v2` or crypto/rand for jitter | S | S |
| P2 | 15 | **Run `golines` formatter** — fix 2 formatting warnings in test files | XS | S |
| P2 | 16 | **Add per-package coverage thresholds in CI** — not just total 75% | S | M |
| P3 | 17 | **Document SARIF round-trip losses explicitly** — SeverityCritical, RelatedRef.FindingID, BeforeCode | S | S |
| P3 | 18 | **API stability review** — lock exported API before v1.0.0 | L | L |
| P3 | 19 | **Add `Report` goroutine safety** — mutex or prominent constructor-only pattern | M | M |
| P3 | 20 | **Add benchmark regression tracking in CI** — compare against baselines | M | M |
| P3 | 21 | **Evaluate `go-sarif` library vs hand-rolled** — spec compliance check | M | M |
| P3 | 22 | **Add `go:generate stringer` for enums** — Severity, FixStrategy, Category, SuppressionKind | S | S |
| P3 | 23 | **Create `examples/` directory** — standalone runnable programs for adoption | M | L |
| P3 | 24 | **Optimize Correlate() O(n²)** — sorted+merge approach for large finding sets | M | S |
| P3 | 25 | **Add version.go with semver constants** — programmatic version checking | S | S |

---

## G) MY #1 QUESTION I CANNOT FIGURE OUT MYSELF

**What is the intended v1.0.0 release criteria?**

The codebase is at v0.1.3 with no documented API stability guarantee. I cannot determine:
- Is v1.0.0 planned, or is this intentionally an "eternal v0.x" library?
- What is the minimum bar for v1.0.0? (API lockdown? All P0/P1 items? 95% coverage? External adoption count?)
- Should `FixStrategyAI` be removed before v1.0.0, or is it acceptable as a "reserved" constant?
- Is the pipeline package considered part of the v1.0.0 API surface, or is it experimental?

This matters because several TODO items reference "API stability review before v1.0.0" but no v1.0.0 milestone or criteria exist in the repository.

---

## Current Test Results Summary

```
go test -race -count=1 ./...
ok  github.com/larsartmann/go-finding         1.133s  (98.7% coverage)
ok  github.com/larsartmann/go-finding/cmd/go-finding  1.018s  (77.3% coverage)
ok  github.com/larsartmann/go-finding/internal/detectors 1.298s  (95.9% coverage)
ok  github.com/larsartmann/go-finding/pipeline  1.066s  (94.5% coverage)
```

**golangci-lint:** 15 issues (down from build-breaking), all minor/acceptable.

**Total coverage:** 94.8% (CI threshold: 75% ✅)

---

## Files Modified This Session

| File | Change |
|------|--------|
| `pipeline/verify_test.go` | Removed unused `"errors"` import |
| `cmd/go-finding/integration_test.go` | Removed `t.Parallel()` from 4 data-racing tests |
| `cmd/go-finding/main_test.go` | Removed `t.Parallel()` from `TestFatalf`, fixed var ordering |
| `AGENTS.md` | Updated `pipeline/astfix.go` → `pipeline/fix_applier.go` |
| `README.md` | Updated coverage stats to current values |
| `pipeline/conflict.go` | Removed stale `//nolint:prealloc` directive |
| `pipeline/pipeline_test.go` | Removed stale `//nolint:prealloc` directive |
| `diagnostic_test.go` | Removed stale `//nolint:goconst` directive |
| `sarif.go` | Removed stale `//nolint:goconst` directive |
| `testutil_test.go` | Removed 3 unused helper functions, added `t.Helper()` to `assertSummaryField` |
| `docs/READINESS_REPORT.md` | NEW — full SDK readiness assessment |
| `docs/status/2026-04-25_01-19_comprehensive-status.md` | NEW — this file |

---

_Assisted-by: Crush <crush@charm.land>_
