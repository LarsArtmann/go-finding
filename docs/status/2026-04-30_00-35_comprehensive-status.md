# Comprehensive Status Report — 2026-04-30 00:35

**Session:** Post-hardening verification + new work
**Branch:** master (up to date with origin)
**State:** ⚠️ **NEW DATA RACE in `cmd/go-finding`** — 15 tests failing with `-race`
**Coverage:** 95.2% (up from 94.1%)

---

## A) FULLY DONE ✅

### Since Last Report (14 new commits from intervening session)

| Commit | What | Impact |
|--------|------|--------|
| `f30f791` | Delete stale binaries, add to `.gitignore` | Cleanup |
| `7d3bbcf` | Fix pipeline example runtime crash, add compile tests | Bug fix |
| `14745da` | `Builder.Build()` returns `(Finding, error)` instead of panicking | **Breaking API change** — `Build()` now validates |
| `8cd5fef` | Add godoc examples for `NewFinding` and Builder | Docs |
| `914ddd3` | Add `WriteSARIF`/`WriteSARIFFiltered` for direct `io.Writer` output | Feature — zero-alloc SARIF output |
| `8eb6146` | Deterministic fix application order in FixApplier | Bug fix — fixes were applied in random map iteration order |
| `6efbfd4` | Error-path tests for `setupProfiling` and `outputResults` | Test |
| `07b8c3a` | Fix `FixStrategyAI` split brain in `HasFix()` | Bug fix — AI now returns false from HasFix |
| `1de6bec` | Error-path tests for `run()` and metrics output | Test |
| `d9bcc2a` | Parallel detector error + applyTriage edge case tests | Test |
| `61dc44e` | Zero-value `Range.Contains` test | Test |
| `4e10e0f` | `FuzzFindingsFromSARIF` for malformed input | Fuzz test |
| `c49665f` | Add `govulncheck` job to CI workflow | CI |
| `d3c5b2a` | Audit and update TODO_LIST.md | Docs |

### Cumulative This Sprint (24 commits total)

| Category | Count |
|----------|-------|
| Bug fixes | 6 (race, concurrent backup, flaky test, example crash, fix order, AI split brain) |
| Ghost systems eliminated | 1 (detectResult) |
| New test functions | ~50+ |
| Coverage improvement | 94.1% → 95.2% |
| Breaking API changes | 1 (`Builder.Build()` now returns error) |

---

## B) PARTIALLY DONE 🔶

### Data Race in `cmd/go-finding` Tests — **INTRODUCED BY RECENT COMMITS**

**Status:** 🔴 **REGRESSION** — 15 tests fail with `-race`

**Root cause:** `TestFatalf` reassigns `os.Stderr` (a process-global variable) while running `t.Parallel()` alongside `TestSetupProfiling_*` tests that also read/write `os.Stderr` via `fmt.Fprintf(os.Stderr, ...)`.

The race trace shows:
```
Write at 0x000000bc82f0 by goroutine 41: TestFatalf() → os.Stderr = w
Read  at 0x000000bc82f0 by goroutine 44: setupProfiling() → fmt.Fprintf(os.Stderr, ...)
```

**Affected tests:** All 15 `cmd/go-finding` tests that run in parallel with `TestFatalf`.

**Fix needed:** `TestFatalf` must NOT use `t.Parallel()` since it mutates a global, OR redirect stderr through a different mechanism that doesn't require global mutation.

---

## C) NOT STARTED ⬜

### High-Value Items From Previous Plan (still valid)

1. Add `-race` to CI workflow default run — **the govulncheck was added but -race was not**
2. Add coverage gate to CI (min 90%)
3. `partial.go` missing metrics recording during partial detection failures
4. `Verifier.Verify` error-path tests
5. `RetryConfig.Validate` edge-case tests
6. `DeduplicateByPosition` vs `DeduplicateByRule` behavior diff test
7. Fix flaky `TestProperty_IDRoundTrip`
8. Profile memory allocation hotspots
9. API stability review before v1

---

## D) TOTALLY FUCKED UP 💀

### #1: We Shipped a New Data Race

The `TestFatalf` test and `TestSetupProfiling_*` tests were added in recent commits (`6efbfd4`, `1de6bec`). They both run `t.Parallel()` but mutate `os.Stderr` — a process-global variable. This means:

- **Every `-race` run of `cmd/go-finding` fails** — 15 out of ~30 tests
- This was not caught because the intervening session may not have run with `-race`
- The previous session's race fix (`1b82717`) was in the `pipeline` package — this new race is in `cmd/go-finding`

### #2: Builder.Build() Breaking Change Was Not Coordinated

Commit `14745da` changed `Build() Finding` to `Build() (Finding, error)`. This is a breaking API change for every consumer. While pre-v1 semver allows this, it should have been flagged in the commit message more prominently.

### #3: FixStrategyAI "Fix" Changed Semantics

Commit `07b8c3a` changed `HasFix()` to return `false` for `FixStrategyAI`. Previously it returned `true` (alongside `FixStrategyDirect`). The analysis in the previous status report concluded the original behavior was **correct and defensible** ("a fix exists in theory, but needs AI to generate it"). This commit reversed that decision without documenting the tradeoff. Now `FixStrategyAI` behaves identically to `FixStrategyNone` in `HasFix()`, which arguably makes it MORE of a phantom.

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Critical
1. **Fix the `cmd/go-finding` data race IMMEDIATELY** — this is a regression
2. **Add `-race` to CI** — both races (pipeline + cmd) survived because CI doesn't run with `-race`
3. **Revert or document the `FixStrategyAI` `HasFix()` change** — the original behavior was correct

### Process
4. **Breaking changes need explicit commit prefix** — `feat!:` or `refactor!:`
5. **Every commit should pass `-race`** — add to pre-push hook or CI gate
6. **Don't add `t.Parallel()` to tests that mutate globals** — lint rule or code review check

### Architecture
7. **`Builder.Build()` error return** — now returns error but never actually returns one (the validation is a placeholder). Either implement real validation or revert to panicking. A function that returns `(T, error)` but never returns error is misleading.
8. **Global `os.Stderr` mutation in tests** — use `t.TempDir()` + file-based capture or inject writers

---

## F) Top #25 Things To Do Next

| # | Task | Impact | Effort | Status |
|---|------|--------|--------|--------|
| 1 | **Fix `cmd/go-finding` data race** (`TestFatalf` + `setupProfiling` globals) | **Critical** | 30min | 🔴 Regression |
| 2 | Add `-race` to CI workflow + justfile | **Critical** | 15min | Not started |
| 3 | Revert `FixStrategyAI` HasFix change or document rationale | High | 15min | Needs decision |
| 4 | Verify `Builder.Build()` actually validates (or revert error return) | High | 30min | Needs review |
| 5 | Add coverage gate to CI (min 93%) | High | 20min | Not started |
| 6 | Fix flaky `TestProperty_IDRoundTrip` | High | 30min | Not started |
| 7 | Add `partial.go` metrics recording during failures | Med | 30min | Not started |
| 8 | Add `Verifier.Verify` error-path tests | Med | 20min | Not started |
| 9 | Add `RetryConfig.Validate` edge-case tests | Med | 20min | Not started |
| 10 | `DeduplicateByPosition` vs `DeduplicateByRule` diff test | Low | 15min | Not started |
| 11 | Add benchmarks for merge, filter, SARIF, ID generation | Med | 45min | Not started |
| 12 | Add JSON fuzz test for `FromJSON`/`ReportFromJSON` | Med | 30min | Not started |
| 13 | Modernize to Go 1.21+ stdlib throughout | Low | 45min | Not started |
| 14 | Convert `retry.go` `errors.New` to sentinels | Low | 15min | Not started |
| 15 | Add godoc examples for remaining key APIs | Low | 30min | Partially done |
| 16 | API stability review before v1 | High | 120min | Not started |
| 17 | Add `go.work` for local development | Low | 15min | Not started |
| 18 | Pipeline example with config file | Low | 30min | Not started |
| 19 | Profile memory allocation hotspots | Med | 60min | Not started |
| 20 | Set up benchmark regression tracking in CI | Med | 30min | Not started |
| 21 | Add pre-push hook for `-race` + lint | Med | 15min | Not started |
| 22 | Clean stale `cover.out`/`coverage.out` from repo root | Low | 5min | Not started |
| 23 | Document SARIF round-trip losses | Low | 30min | Not started |
| 24 | Evaluate `go-sarif` vs hand-rolled SARIF | Low | 60min | Not started |
| 25 | Contribution guidelines review | Low | 30min | Not started |

---

## G) Top #1 Question I Cannot Answer Myself

**Was the `FixStrategyAI` HasFix() change (`07b8c3a`) intentional and correct?**

Previous analysis concluded the original behavior (HasFix=true for AI) was defensible: "a fix exists in theory but needs AI to generate it." The new commit makes HasFix=false for AI, making AI behave identically to `None` in `HasFix()`. This means:
- `FixStrategyAI` now has **zero distinguishing behavior** from `FixStrategyNone` in any method except `NeedsAI()` and `IsValid()`
- Pipeline triage already routed AI to Suggest (correct), but now `HasFix()` also says "no fix" which contradicts the intent of "AI can fix this"

I need to know: should I revert to the original behavior, or keep the new one?

---

## Metrics

| Metric | Value | Change |
|--------|-------|--------|
| Coverage | **95.2%** | ↑ from 94.1% |
| finding package | 99.1% | ↓ from 99.6% (Builder.Build error paths) |
| pipeline package | 97.3% | ↑ from 96.4% |
| cmd/go-finding | 96.2% | ↑ from 78.0% |
| detectors | 96.1% | unchanged |
| Test functions | ~440+ | ↑ from 425 |
| Race failures | **15 in cmd/go-finding** | 🔴 REGRESSION |
| Lint issues | 0 | ✅ |
