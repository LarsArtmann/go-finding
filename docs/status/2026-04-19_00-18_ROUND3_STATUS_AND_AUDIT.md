# Status Report — 2026-04-19 00:18

## Executive Summary

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools.
Three rounds of self-audit (38 tasks in R1+R2, 13 of 16 in R3) have been completed.
**13 commits are unpushed** (12 committed + tests pass, working tree clean).

---

## A) FULLY DONE ✅

### Round 3 Completed (13 tasks, all committed):

| ID    | Commit    | Description                                                           |
| ----- | --------- | --------------------------------------------------------------------- |
| R3.1  | `a7553c7` | Clamp Range.LineCount/Length negative returns for malformed ranges    |
| R3.2  | `13081cc` | Deep-clone findings in Merge to prevent shared pointer aliasing       |
| R3.3  | `31661b9` | Add len(out)>0 guard to staticcheck exit error handling               |
| R3.4  | `5165562` | Remove collectAllFindings fallback that bypassed deduplication        |
| R3.5  | `f9f0411` | Include rule in DeduplicateByPosition key to prevent cross-tool dedup |
| R3.6  | `52997a0` | Guard Range.Adjacent against Column==0 false positives                |
| R3.7  | `1938fba` | Extract SuggestedFixes content in FromDiagnostic                      |
| R3.8  | `514ef62` | Remove trailing newline from FormatDiagnostic                         |
| R3.9  | `f898aee` | Remove dead fixToGroup map in AnalyzeConflicts                        |
| R3.10 | `342390d` | Use composite key fallback in DiffFindings for empty IDs              |
| R3.11 | `19b0639` | Guard severity comparison methods against invalid values              |
| R3.12 | `324d994` | Align HasFix and SARIF fix generation definitions                     |
| R3.13 | `9d40d22` | Convert ParseID to return ParsedID struct                             |

### Prior Rounds (38 tasks, all committed in prior sessions):

- Round 1: 21 tasks (P0 bugs, P1 quality, P2 API, P3 tests)
- Round 2: 11 tasks (deep re-read findings)
- Round 3.16: Removed redundant cmd.Stderr=nil (done alongside R3.3)

**Total: 51 audit tasks completed across 3 rounds.**

---

## B) PARTIALLY DONE ⚠️

| Task                                  | Status          | Notes                                       |
| ------------------------------------- | --------------- | ------------------------------------------- |
| R3.14 — Correlation JSON tags         | **Not started** | merge.go:151-155 still lacks json tags      |
| R3.15 — Remove FindingsFromSARIF stub | **Not started** | sarif.go:281-283 still returns always-error |

---

## C) NOT STARTED ❌

### Known Issues from Round 3 Audit (16 remaining):

| #   | Issue                                                  | Severity         | Location                              |
| --- | ------------------------------------------------------ | ---------------- | ------------------------------------- |
| 1   | `OnFix` callback never called for successful fixes     | **P0**           | pipeline/pipeline.go:466-481          |
| 2   | Partial detection missing metrics recording            | **P0**           | pipeline/partial.go:63-117            |
| 3   | `Config` has no validation                             | P1               | pipeline/pipeline.go:63-86            |
| 4   | `RetryConfig` has no validation                        | P1               | pipeline/retry.go:19-23               |
| 5   | `FixApplier` no rollback on partial failure            | P1               | pipeline/pipeline.go:522-553          |
| 6   | `Pipeline` not safe for concurrent Run()               | P1               | pipeline/pipeline.go:99-107           |
| 7   | `Metrics` exported map fields (thread-unsafe)          | P1               | pipeline/metrics.go:12-14             |
| 8   | `TotalDuration()` returns negative when EndTime unset  | P1               | pipeline/metrics.go:52-57             |
| 9   | SARIF `Location` should be `Locations []SarifLocation` | P2               | sarif.go:43                           |
| 10  | `Report.AddFinding` not goroutine-safe                 | P1               | report.go:37-39                       |
| 11  | Profiling FD leak in CLI                               | **P0**           | cmd/go-finding/main.go:133-149        |
| 12  | `DeduplicateByPosition` == `DeduplicateByRule` keys    | **P0 (NEW BUG)** | merge.go:129-144                      |
| 13  | `Correlation` no JSON tags                             | P2               | merge.go:151-155                      |
| 14  | `FindingsFromSARIF` always-error stub                  | P2               | sarif.go:281-283                      |
| 15  | `HasFix` true for Direct even without AfterCode        | P2               | finding.go:90-99                      |
| 16  | gopls hints (~12 non-critical)                         | P3               | bench_test.go, coverage_test.go, etc. |

---

## D) TOTALLY FUCKED UP 💥

### 1. NEW BUG INTRODUCED: `DeduplicateByPosition` == `DeduplicateByRule`

**R3.5 changed `DeduplicateByPosition` key from `File:Line:Column` to `Rule:File:Line:Column`** —
which is now **identical** to `DeduplicateByRule` key format (`Rule:File:Line:Column`).
These two strategies are now functionally equivalent. The `Position` strategy should NOT include Rule.

**This is the highest priority fix.** It silently changed dedup semantics.

### 2. ParseID Migration Risk

The R3.13 change to return `ParsedID` struct is a **breaking API change** for any external callers.
This is fine for an unreleased library, but needs documentation.

---

## E) WHAT WE SHOULD IMPROVE

### Honest Self-Critique:

1. **Introduced a regression in R3.5** — Changed `DeduplicateByPosition` to include Rule, but didn't check if `DeduplicateByRule` already used that format. The two are now identical. Should have diffed the key formats.

2. **No integration tests for dedup strategies** — We have unit tests for individual functions, but no test that verifies `DeduplicateByPosition` actually differs from `DeduplicateByRule` in behavior.

3. **`OnFix` callback is dead code** — The `Config.OnFix` field exists but is never called for the success path. This means nobody using the pipeline can react to successful fixes. This is a fundamental pipeline feature gap.

4. **Metrics gap in partial detection** — When `GracefulDegradation` is enabled and a detector fails, the partial detection path doesn't record timing metrics. This means the metrics dashboard is incomplete.

5. **Pipeline concurrency story is weak** — `Pipeline.Run()` mutates shared state without synchronization. The library documentation doesn't mention this limitation. Either add a mutex or document that `Run()` must not be called concurrently.

6. **Exported map fields in Metrics** — `StageDurations`, `DetectorTimes`, `FindingsFound` are exported maps. Even though methods use a mutex, external code can read/write the maps directly. This is a data race waiting to happen.

7. **No `Config.Validate()`** — The pipeline accepts any config, including negative `MaxRetries`, zero `MaxIterations`, etc. This leads to silent misbehavior rather than clear errors.

8. **Split brain: `HasFix()` vs `HasSuggestion()`** — Two methods that overlap. `HasFix` now returns true for `Suggest` when `AfterCode` is set, but `HasSuggestion` also checks `AfterCode`. The boundary between "fix" and "suggestion" is unclear.

9. **SARIF `Location` vs `Locations`** — SARIF 2.1.0 spec uses `locations` (array), but our type uses `Location` (singular). This is a spec compliance issue.

10. **No examples directory content** — The `examples/` directory structure exists but the `ls` shows 0 files. The example code is referenced in AGENTS.md but may have been moved or removed.

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by importance/impact/effort:

| #   | Task                                                                    | Severity | Effort | Impact   |
| --- | ----------------------------------------------------------------------- | -------- | ------ | -------- |
| 1   | **FIX: DeduplicateByPosition key should NOT include Rule** (regression) | P0       | 5min   | Critical |
| 2   | Add test verifying DeduplicateByPosition ≠ DeduplicateByRule behavior   | P0       | 10min  | Critical |
| 3   | **FIX: OnFix callback never called for successful fixes**               | P0       | 15min  | High     |
| 4   | **FIX: Profiling FD leak in CLI**                                       | P0       | 10min  | High     |
| 5   | **FIX: Partial detection missing metrics recording**                    | P0       | 15min  | High     |
| 6   | Add Config.Validate() method                                            | P1       | 15min  | High     |
| 7   | Add RetryConfig.Validate() method                                       | P1       | 10min  | Medium   |
| 8   | Guard TotalDuration() against negative return                           | P1       | 5min   | Medium   |
| 9   | Make Metrics map fields unexported with accessor methods                | P1       | 20min  | High     |
| 10  | Add sync.Mutex to Pipeline for concurrent Run() safety                  | P1       | 15min  | High     |
| 11  | Add sync.Mutex to Report.AddFinding                                     | P1       | 10min  | Medium   |
| 12  | FixApplier rollback all files on partial failure                        | P1       | 30min  | High     |
| 13  | Fix SARIF Location → Locations (plural, array)                          | P2       | 20min  | Medium   |
| 14  | Add JSON tags to Correlation struct                                     | P2       | 2min   | Low      |
| 15  | Remove FindingsFromSARIF always-error stub                              | P2       | 5min   | Low      |
| 16  | Clarify HasFix vs HasSuggestion boundary                                | P2       | 15min  | Medium   |
| 17  | Clean up gopls hints (rangeint, newexpr, mapsloop, stringsseq)          | P3       | 15min  | Low      |
| 18  | Add examples back to examples/ directory                                | P2       | 20min  | Medium   |
| 19  | Add Pipeline integration test for OnFix callback                        | P1       | 10min  | High     |
| 20  | Document Pipeline is NOT concurrent-safe (or make it so)                | P1       | 5min   | Medium   |
| 21  | Add benchmark for Merge with large finding sets                         | P2       | 10min  | Low      |
| 22  | Review Correlate O(n²) performance                                      | P2       | 20min  | Medium   |
| 23  | Add gosec/staticcheck to CI linting                                     | P2       | 15min  | Medium   |
| 24  | Push all commits to origin                                              | —        | 1min   | —        |
| 25  | Write comprehensive planning doc for Round 4                            | —        | 15min  | —        |

---

## G) TOP #1 QUESTION

**The `DeduplicateByPosition` key format.** R3.5 added Rule to prevent cross-tool dedup.
But now `DeduplicateByPosition` and `DeduplicateByRule` are identical.
The question: **Should `DeduplicateByPosition` include Rule or not?**

- Including Rule: Prevents cross-tool dedup of same-location findings (useful when multiple tools report same issue)
- NOT including Rule: Deduplicates across tools at same position (useful for deduping the same issue reported by multiple tools)

These are opposite semantics. The original code (before R3.5) did NOT include Rule — that was the "position-only" dedup. The `DeduplicateByRule` was supposed to dedup by rule+position (semantically same issue from same tool).

**My recommendation**: Revert `DeduplicateByPosition` back to `File:Line:Column` (no Rule). Keep `DeduplicateByRule` as `Rule:File:Line:Column`. They serve different purposes.

---

## Build/Test State

- `go build ./...` — ✅ Passes
- `go vet ./...` — ✅ Passes
- `go test -race -count=1 ./...` — ✅ All 4 packages pass
- Branch: `master`, 13 commits ahead of origin (NOT PUSHED)
- Working tree: Clean

## Project Stats

| Metric              | Count  |
| ------------------- | ------ |
| Production Go files | 26     |
| Test files          | 39     |
| Total lines         | 13,299 |
| Production lines    | 4,490  |
| Test lines          | 8,809  |
| Test/Code ratio     | 1.96:1 |
| Unpushed commits    | 13     |

---

_Assisted-by: Crush <crush@charm.land>_
