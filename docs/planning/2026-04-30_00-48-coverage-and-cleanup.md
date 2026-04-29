# Execution Plan: Coverage Completion & CI Hardening

**Date:** 2026-04-30 00:48
**Project:** go-finding
**Theme:** Remove dead code, fill real coverage gaps, add CI quality gates

---

## Brutally Honest Assessment

### What Did I Forget?
1. I wrote a stale status report and committed it by accident (fixed in same commit, but still sloppy)
2. I didn't verify that `-race` was ALREADY in CI before proposing to add it
3. I didn't verify `RetryConfig.Validate` coverage before listing it as a TODO — it's at 100%
4. I didn't verify `partial.go` metrics recording — `runOneDetector` already records metrics on failure
5. I didn't verify `Verifier.Verify` error paths — `TestVerifier_Verify_DetectorError` already exists
6. I didn't run `-count=100` before the previous session ended

### What's Stupid That We Do Anyway?
1. **`hasLineRange` has unreachable dead code** — `r.Start.Line == 0 || p.Line == 0` can never be true because `Contains()` only calls `hasLineRange()` when both lines are > 0
2. **`detectSequential` has a `ctx.Done()` case that's never hit in production** — `Run()` checks `isContextDone()` before calling `detect()`, so the inner check is redundant for the sequential path
3. **We keep copying stale TODOs** — `RetryConfig.Validate`, `partial.go` metrics, `Verifier.Verify` error paths were all listed as open but are already done
4. **CI enforces 75% total coverage but not per-package** — a regression in `cmd/go-finding` (81% → 70%) would not fail CI if total stays above 75%

### What Could Be Done Better?
1. Verify TODOs are actually open before adding them to plans
2. Run `-count=100` before declaring "stable"
3. Remove dead code when discovered, not defer it

### What Could Still Improve?
1. Remove `hasLineRange` dead code
2. Test `detectSequential` context cancellation directly (the path through `Run()` is blocked by `Run`'s own check)
3. Test `detectPartialSequential` context cancellation
4. Add per-package coverage thresholds to CI
5. Run `-count=100` stress test

---

## Execution Plan (sorted by impact/effort)

| # | Task | Effort | Impact | Package |
|---|------|--------|--------|---------|
| 1 | Remove `hasLineRange` dead code | 10min | **High** — code clarity | `finding` |
| 2 | Add `detectSequential` context-cancel test | 15min | **Medium** — coverage | `pipeline` |
| 3 | Add `detectPartialSequential` context-cancel test | 15min | **Medium** — coverage | `pipeline` |
| 4 | Add per-package coverage thresholds to CI | 20min | **High** — prevents regression | `.github` |
| 5 | Run `-count=100` stress test | 10min | **High** — confidence | all |
| 6 | Update TODO_LIST.md: close stale items | 10min | **Medium** — hygiene | docs |

---

## How This Contributes to Customer Value

1. **Dead code removal** = smaller binary, faster compilation, less cognitive load
2. **Coverage gates** = bugs caught before merge
3. **Stress testing** = confidence in stability
4. **Clean TODOs** = accurate project state
