# Comprehensive Status Report — 2026-04-30 00:55

**Session:** Coverage completion, dead code removal, CI hardening  
**Branch:** master (ahead of origin by 3 commits)  
**State:** All tests pass including `-count=100`. Lint clean.  
**Coverage:** 95.2% total

---

## A) FULLY DONE ✅

### This Session (3 commits)

| Commit    | What                                       | Impact                                                     |
| --------- | ------------------------------------------ | ---------------------------------------------------------- |
| `42e6035` | Remove unreachable check in `hasLineRange` | Code clarity — eliminated dead code, coverage 85.7% → 100% |
| `f4221ad` | Add per-package coverage threshold checker | CI — prevents regression with package-level gates          |
| `c7bdfe1` | Verified comprehensive status report       | Docs                                                       |

### Previous Session (15 commits, carried forward)

| Commit    | What                                         | Impact                         |
| --------- | -------------------------------------------- | ------------------------------ |
| `f30f791` | Delete stale binaries, `.gitignore`          | Cleanup                        |
| `07b8c3a` | Fix `FixStrategyAI` split brain              | Bug fix                        |
| `1de6bec` | `run()` error-path tests                     | Test — cmd 81% → 96%           |
| `d9bcc2a` | Pipeline edge case tests                     | Test — pipeline 96.4% → 97.3%  |
| `4e10e0f` | `FuzzFindingsFromSARIF`                      | Fuzz — 1.6M execs, zero panics |
| `c49665f` | `govulncheck` in CI                          | Security                       |
| `a043410` | Fix test flakiness under `-count=N`          | Stability                      |
| `914ddd3` | `WriteSARIF`/`WriteSARIFFiltered`            | Feature                        |
| `14745da` | `Builder.Build()` returns error              | Breaking API change            |
| `8cd5fef` | Godoc examples                               | Docs                           |
| `7d3bbcf` | Fix pipeline example crash                   | Bug fix                        |
| `6efbfd4` | `setupProfiling`/`outputResults` error tests | Test                           |
| `8eb6146` | Deterministic FixApplier order               | Bug fix                        |
| `61dc44e` | Zero-value `Range.Contains` test             | Test                           |
| `d3c5b2a` | Audit `TODO_LIST.md`                         | Hygiene                        |

### Cumulative Metrics

| Metric                 | Value                  |
| ---------------------- | ---------------------- |
| Total coverage         | **95.2%**              |
| `finding` package      | 99.2%                  |
| `pipeline` package     | 97.3%                  |
| `cmd/go-finding`       | 96.2%                  |
| `internal/detectors`   | 96.1%                  |
| Test functions         | ~445+                  |
| Lint issues            | 0                      |
| `-count=100` stability | ✅ PASS (all packages) |
| `-race`                | ✅ PASS                |

---

## B) PARTIALLY DONE 🔶

### `detectSequential` / `detectPartialSequential` Context Cancellation

**Status:** Uncovered defensive paths exist but are impractical to test.

- `Run()` checks `isContextDone()` before calling `detect()`, so the inner `ctx.Done()` checks in `detectSequential`/`detectPartialSequential` are only hit if cancellation happens between detectors
- Testing this requires precise timing that's inherently flaky
- These are defensive checks that duplicate the detector-level `ctx.Done()` checks inside `mockDetector.Detect()`

**Decision:** Accepted as acceptable uncovered defensive code. The detector-level cancellation is fully tested.

### `detectParallel` Goroutine Cancellation

**Status:** The `g.Wait()` error propagation path (line 433) is uncovered.

- `detectParallel` returns error from `g.Wait()` when a goroutine returns non-nil error
- But goroutines in `detectParallel` always return `nil` (errors are collected via mutex)
- So `g.Wait()` never returns an error in the current implementation
- The `fmt.Errorf("parallel detection: %w", err)` at line 434 is dead code

**Fix needed:** Either remove the dead error handling or test it by making a goroutine return non-nil.

---

## C) NOT STARTED ⬜

### Real Coverage Gaps ( represent reachable code )

1. **`detectParallel` `g.Wait()` error path** — dead code that should be removed
2. **`FixApplier` error paths** (87% → 90%+)
3. **`findingFromSarResult` SARIF import** — rule metadata, help URI, markdown descriptions
4. **`DeduplicateByPosition` vs `DeduplicateByRule`** behavior diff test
5. **`TestProperty_IDRoundTrip`** — ~5% failure claim needs verification

### Architecture

6. **`FixStrategyAI` fate** — awaiting user decision (remove / keep / implement)
7. **`Builder.Build()` validation** — error return is mostly theater
8. **`detectParallel` dead error handling** — remove or test
9. **`Finding` struct sub-grouping** — breaking change, defer to v1 planning

### CI & Tooling

10. **Benchmark regression tracking** — not started
11. **`gosec`/`staticcheck` in CI** — not started
12. **GitHub release workflow** — not started
13. **`-count=100` in CI** — too slow for CI (91s total), keep as manual check

### Documentation

14. SARIF round-trip losses doc
15. Real-world tool integration guide
16. Contribution guidelines review

---

## D) TOTALLY FUCKED UP 💀

### #1: `detectParallel` Has Untestable Dead Code

`detectParallel` at lines 433-435:

```go
if err := g.Wait(); err != nil {
    return nil, fmt.Errorf("parallel detection: %w", err)
}
```

This can NEVER execute because every goroutine returns `nil`. The error collection happens via mutex, not via errgroup error propagation. This is dead code that should be `_ = g.Wait()` or the error handling removed.

**Impact:** Low. Dead code, not a bug. But it inflates the function and misleads readers.

### #2: I Committed a Stale Status Report

Commit `a043410` accidentally included `docs/status/2026-04-30_00-35_comprehensive-status.md` which claimed 15 tests failed with `-race`. This was fixed in the same commit by replacing it, but the intermediate state was sloppy.

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Critical

1. **Remove `detectParallel` dead error handling** — `_ = g.Wait()` instead of `if err := g.Wait(); err != nil`
2. **Decide `FixStrategyAI` fate** — user input needed

### Process

3. **Verify TODOs before adding to plans** — `RetryConfig.Validate`, `partial.go` metrics, `Verifier.Verify` were all falsely listed as uncovered
4. **Run `-count=100` before declaring stable** — done this session, should be standard
5. **Don't commit status reports until verified** — the 00:35 report was committed with stale data

### Architecture

6. **`Builder.Build()` error return** — either add meaningful validation or simplify
7. **Remove `detectParallel` dead error handling** — one-line fix

---

## F) Top #15 Things To Do Next

| #   | Task                                                | Impact   | Effort | Blocked By    |
| --- | --------------------------------------------------- | -------- | ------ | ------------- |
| 1   | Remove `detectParallel` dead error handling         | **High** | 5min   | Nothing       |
| 2   | Decide `FixStrategyAI` fate                         | **High** | 30min  | User decision |
| 3   | Add `FixApplier` error-path tests                   | High     | 30min  | Nothing       |
| 4   | Add `findingFromSarResult` import tests             | Med      | 30min  | Nothing       |
| 5   | Fix flaky `TestProperty_IDRoundTrip`                | Med      | 30min  | Nothing       |
| 6   | `DeduplicateByPosition` vs `DeduplicateByRule` test | Low      | 15min  | Nothing       |
| 7   | Add `go:generate stringer` for enums                | Low      | 30min  | Nothing       |
| 8   | Modernize to Go 1.21+ stdlib                        | Low      | 45min  | Nothing       |
| 9   | Convert `retry.go` `errors.New()` to sentinels      | Low      | 15min  | Nothing       |
| 10  | Document SARIF round-trip losses                    | Low      | 30min  | Nothing       |
| 11  | Add benchmark regression tracking                   | Med      | 30min  | Nothing       |
| 12  | `gosec`/`staticcheck` in CI                         | Med      | 20min  | Nothing       |
| 13  | GitHub release workflow                             | Med      | 45min  | Nothing       |
| 14  | Real-world tool integration guide                   | Med      | 60min  | Nothing       |
| 15  | API stability review before v1                      | High     | 120min | Nothing       |

---

## G) Top #1 Question I Cannot Answer Myself

**Should `FixStrategyAI` be removed from the public API entirely?**

It has existed for 5+ sessions as a "placeholder." It now behaves like `Suggest` in both `triage()` and `HasFix()`. The only distinguishing methods are `NeedsAI()` (returns `true`) and `IsValid()` (returns `true`).

**Arguments for removal:**

- No AI backend exists. No timeline exists.
- It confuses users who set `FixStrategyAI` expecting AI-powered fixes.
- Every session has flagged it as a problem.
- Removing it before v1.0.0 avoids a deprecation cycle.

**Arguments for keeping:**

- Removing it is a breaking change for any consumers.
- It reserves the value for future implementation.
- `NeedsAI()` provides a clear semantic hook for future AI integration.

I need your decision: **remove it now**, **keep it with enhanced documentation**, or **build a minimal AI plugin interface** to justify its existence?

---

## Metrics

| Metric             | Value     | Change       |
| ------------------ | --------- | ------------ |
| Coverage           | **95.2%** | stable       |
| `finding` package  | 99.2%     | ↑ from 99.1% |
| `pipeline` package | 97.3%     | stable       |
| `cmd/go-finding`   | 96.2%     | stable       |
| `detectors`        | 96.1%     | stable       |
| Test functions     | ~445+     | stable       |
| `-count=100`       | ✅ PASS   | Verified     |
| `-race`            | ✅ PASS   | Verified     |
| Lint issues        | 0         | ✅           |
| Open TODOs         | 48        | stable       |
