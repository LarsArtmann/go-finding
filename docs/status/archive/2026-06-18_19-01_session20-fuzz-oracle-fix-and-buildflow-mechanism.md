# Session 20 — Fuzz Oracle Fix, BuildFlow Mechanism Resolved, Cache Mirage Diagnosed

**Date:** 2026-06-18 19:01 CEST
**Branch:** master (up to date with origin)
**Current tag:** v0.9.0
**Commits this session:** 1 (authored via Crush, committed by user as `b2d6388`)
**Trigger:** Identical CI failure to Session 19 — `test-fuzz: [Failed] 20 fuzz tests failed (3 passed)`

---

## Executive Summary

Session 19 fixed fuzz **naming collisions** and believed the `test-fuzz` failure was resolved. It recurred this session with the **same error message but a different root cause**. This session performed a deeper diagnosis and uncovered that the "20 failures" was a **stale-cache mirage**: only **1 real invariant failure** existed (`FuzzMergeByPosition`), caused by a **wrong test oracle** (not a production bug).

Two outcomes:

1. **Fixed the real bug** — `FuzzMergeByPosition` asserted that equal dedup keys always merge to one finding. This is wrong: `DeduplicateByPosition` intentionally skips findings with empty `Position.File`. Corrected the oracle to match production semantics.
2. **Answered Session 19's Top #1 open question** — located BuildFlow's `test-fuzz` implementation and extracted the **exact command** it runs per target. Used it to verify all 23 targets pass at 30s.

**Verification:** 667 tests pass (`-race`), 23/23 fuzz targets pass with the exact BuildFlow invocation, full suite green.

---

## a) FULLY DONE ✅

### 1. Root-Caused the Recurring `test-fuzz` Failure (Definitively)

Session 19 stopped at "rename collisions". This session proved the failure has **three distinct failure modes** that all produce the identical `20 fuzz tests failed (3 passed)` message:

| Failure Mode                           | Real Failure Count      | Cause                                          | Fix                   |
| -------------------------------------- | ----------------------- | ---------------------------------------------- | --------------------- |
| Naming collision (Session 19)          | 23 (all refuse to run)  | `-fuzz` regex matches >1 target                | Rename 3 targets      |
| **Stale compile cache (this session)** | **0 real; 20 apparent** | Broken compile cached mid-refactor             | `go clean -testcache` |
| **Wrong fuzz oracle (this session)**   | **1 real**              | `FuzzMergeByPosition` asserted wrong invariant | Fix oracle            |

**Lesson:** The error message is structurally ambiguous. "20 failed (3 passed)" can mean _compile error_, _cache staleness_, OR _invariant violation_. BuildFlow's step does not distinguish them. See **Section d**.

### 2. Fixed `FuzzMergeByPosition` Oracle (`fuzz_test.go:255-268`)

**The bug was in the test, not production.** Production code is correct and intentional:

```go
// merge.go:145-148 — DeduplicateByPosition
case DeduplicateByPosition:
    if finding.Position.File == "" {
        return "", false   // a position without a file is not a meaningful dedup key
    }
```

This behavior is already documented and asserted by `FuzzDedupKey` (which expects `ok=false` for empty file). But `FuzzMergeByPosition` contradicted it — it assumed `key1 == key2 ⟹ merged len == 1` for ALL inputs, including empty-file positions.

**Fix** — assert the real two-case semantics:

```go
if key1 == key2 && file1 != "" && file2 != "" {
    g.Expect(merged.Findings).To(HaveLen(1))   // merged
} else {
    g.Expect(merged.Findings).To(HaveLen(2))   // both retained
}
```

**Why the seed corpus never caught it:** the fuzzer only recently generated an input where both `file1` and `file2` were empty strings with otherwise-equal positions. The persisted failing input: `file=""`, `line=-58`, `col=1` for both findings.

### 3. Added `pipeline/export_test.go` Bridge (committed `b2d6388`)

Standard Go `export_test.go` pattern: centralizes the `rangeFix`, `offsetFix`, `offsetFixWithID` test helpers in the internal `pipeline` package and exposes them to the external `pipeline_test` package via `ExportRangeFix` / `ExportOffsetFix` / `ExportOffsetFixWithID` wrappers. Eliminates helper duplication across `bdd_test.go` and `fix_engine_test.go`. This was the in-progress refactor whose broken intermediate state caused the cache mirage.

### 4. Located BuildFlow's `test-fuzz` Implementation (Session 19's open question)

**Source:** `/home/lars/projects/BuildFlow/tools/executor_testing.go:129` — `func (e *Executor) RunTestFuzz`

**Exact command per target:**

```bash
go test -run=<Name> -fuzz=<Name> -fuzztime=30s -fuzzminimizetime=30s -parallel=1 -shuffle=on -timeout=15m <pkg>
```

- Iterates each fuzz target individually (Go can't combine `-fuzz` with `./...`)
- Counts non-zero exits; failure message format: `"%d/%d fuzz tests failed: %v"`
- No `-fuzz` cache isolation between targets — a stale build cache poisons all subsequent targets in the same package

### 5. Verified All 23 Fuzz Targets Pass at Full BuildFlow Duration

Ran the exact BuildFlow command (`-fuzztime=30s -shuffle=on`) against all 23 targets from a clean cache:

```
30s BUILDFLOW-CMD SWEEP: PASS=23 FAIL=0   ALL 23 PASS
```

### Health Metrics (Current)

| Metric                        | Value |
| ----------------------------- | ----- |
| Test functions                | 667   |
| Fuzz targets                  | 23    |
| Benchmarks                    | 48    |
| Features FULLY_FUNCTIONAL     | 35    |
| Features PARTIALLY_FUNCTIONAL | 7     |
| Features PLANNED / BROKEN     | 0 / 0 |
| TODO/FIXME in production code | 0     |
| gofmt issues                  | 0     |
| Full `-race` suite            | green |

---

## b) PARTIALLY DONE 🟡

| Item                             | Status                                 | Note                                                                                                                                                                                                                                         |
| -------------------------------- | -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline/export_test.go` bridge | Committed but LSP shows stale warnings | `golangci-lint` reports 0 issues; `golangci_lint_ls` LSP cache shows phantom `gochecknoglobals`/`nlreturn` warnings on the `Export*` wrappers (they're `func`s now, not `var`s). LSP restart failed. Cosmetic.                               |
| Lint cleanliness                 | 3 pre-existing issues                  | `gocyclo` on `Validate` (complexity 32 > 25), 2× `varnamelen` (`fs`, `rt`). All pre-date this session; not blocking but block a "lint clean" claim.                                                                                          |
| go.sum stability                 | Drifting                               | 4 prior commits (`c59b7da`, `0d2e193`, `110269b`) cleaned transitive test deps; working tree shows go.sum modified again (removing `testify`, `go-spew`, `gkampitakis/*`, etc.). `go mod tidy` appears non-deterministic vs committed state. |

---

## c) NOT STARTED ⬜

(Carried forward; none newly started this session beyond the above.)

- v1.0.0 release — cut tag, remove 7 deprecated APIs, unexport `Report.Findings`
- Update `FEATURES.md`, `README.md`, `USAGE_GUIDE.md` for v0.9.0
- Named string types (`ToolName`, `RuleName`, `FindingID`)
- Godoc examples for `Normalized`, `WithFix`, `NormalizeFixStrategy`
- CI fuzz step in `ci.yml` (now unblocked — exact command known; see Section f #1)

---

## d) TOTALLY FUCKED UP 💥

### 1. The `test-fuzz` Error Message Is Structurally Ambiguous

This is the real fuckup. **Three completely different failure modes produce the identical error string** `N fuzz tests failed (M passed)`:

- **Compile error** (e.g. broken `export_test.go` mid-refactor) → every target in that package "fails"
- **Cache staleness** → same as above, but the code is actually fine; `go clean -testcache` fixes it
- **Real invariant violation** → an actual bug (rare; only `FuzzMergeByPosition` this session)

Session 19 saw "20 failed", assumed invariant violations, and fixed naming. Session 20 saw "20 failed" again, and it was **1 real bug + cache mirage**. **The same error twice, two different causes.** This will keep happening until the diagnostic is improved.

### 2. Session 19 Under-Diagnosed (Retrospective)

Session 19's fix was correct _for the collision case_ but it never verified the full suite with a clean cache at BuildFlow's real 30s duration, so the `FuzzMergeByPosition` oracle bug survived. The "Top #1 Question" Session 19 left open ("where is test-fuzz configured?") was the exact knowledge that would have caught this — if Session 19 had found the command, it would have reproduced the real failure. **This session closed that loop.**

### 3. Fuzz Oracle Soundness Is an Un-audited Surface

`FuzzMergeByPosition` asserted wrong behavior and passed for weeks because the fuzzer never generated the empty-file edge case. **Other fuzz oracles may harbour similar wrong assumptions about zero/empty values.** This is a class of bug distinct from "production bug found by fuzz" — it's "test bug hidden by fuzz". No audit exists.

### 4. Stale LSP Diagnostics

`golangci_lint_ls` reports phantom warnings on `export_test.go` (`gochecknoglobals`, `nlreturn`) that `golangci-lint run` does not reproduce. LSP restart failed (`lsp_restart` returned an error). Anyone trusting the IDE squiggles over the CLI will chase ghosts.

### 5. go.sum Ping-Pong

Four consecutive commits churned `go.sum`, and it is **still drifting** in the working tree. Either `go mod tidy` is non-deterministic for this dependency set, or BuildFlow's `go-mod-tidy` / `go-mod-ignore-check` steps are fighting the manual commits. Unresolved.

---

## e) WHAT WE SHOULD IMPROVE 🔧

1. **Add `scripts/fuzz-check.sh`** — runs each target from a clean cache with explicit per-target pass/fail output. Kills the ambiguity in Section d.1. Makes "20 failed" mean 20 real bugs or tells you it was a compile error. **Highest leverage.**
2. **Audit all 23 fuzz oracles** for zero/empty-value assumptions. `FuzzMergeByPosition` was wrong; check the rest. Especially `FuzzMerge_DedupByID` (empty ID handling) and `FuzzCorrelate` (empty file).
3. **Add a CI fuzz job to `ci.yml`** using the now-known exact command. 30s × 23 targets ≈ 12min — fits one CI job.
4. **Fix the 3 lint issues** — extract `Validate` sub-functions (gocyclo), rename `fs`→`strategy`, `rt`→`result`. Trivial; unblocks "lint clean" claims.
5. **Stabilize go.sum** — run `go mod tidy` once, commit once, stop the churn. Investigate whether BuildFlow's mod steps are the antagonist.
6. **Document the `-fuzz` regex + cache behavior** in `CONTRIBUTING.md` so future contributors don't repeat the misdiagnosis.
7. **Restart the golangci-lint LSP** (or file a config issue) — phantom warnings erode trust in IDE diagnostics.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

Sorted by impact. Items marked **NEW** or **unblocked** reflect this session's findings.

| #   | Task                                                                  | Impact   | Effort | Note                             |
| --- | --------------------------------------------------------------------- | -------- | ------ | -------------------------------- |
| 1   | **Add `scripts/fuzz-check.sh`** (clean-cache, per-target diagnostics) | Critical | 30min  | Kills the ambiguity fuckup (d.1) |
| 2   | **Audit all 23 fuzz oracles** for empty/zero assumptions              | Critical | 1hr    | Prevents next hidden oracle bug  |
| 3   | Add CI fuzz job to `ci.yml` (exact command now known)                 | High     | 30min  | **Unblocked this session**       |
| 4   | Fix 3 lint issues (gocyclo `Validate`, `varnamelen` `fs`/`rt`)        | High     | 20min  | Unblocks "lint clean"            |
| 5   | Stabilize go.sum (one tidy, one commit, investigate drift)            | High     | 30min  | Stops the ping-pong (d.5)        |
| 6   | Document `-fuzz` regex+cache behavior in CONTRIBUTING.md              | Medium   | 15min  | Prevents misdiagnosis recurrence |
| 7   | Update FEATURES.md with v0.9.0 changes                                | High     | 30min  | Carryover                        |
| 8   | Update README.md for v0.9.0                                           | High     | 30min  | Carryover                        |
| 9   | Named string types: `ToolName`, `RuleName`, `FindingID`               | High     | 2hr    | Carryover                        |
| 10  | v1.0.0: remove 7 deprecated APIs                                      | Critical | 2hr    | Carryover                        |
| 11  | v1.0.0: unexport `Report.Findings`                                    | High     | 1hr    | Carryover                        |
| 12  | Cut v1.0.0 tag                                                        | Critical | 1hr    | Carryover                        |
| 13  | Add godoc examples: `Normalized`, `WithFix`                           | Medium   | 30min  | Carryover                        |
| 14  | Complete CLI FixProviders config                                      | Medium   | 2hr    | Carryover                        |
| 15  | Integration test: full pipeline with split-brain fixes                | Medium   | 1hr    | Carryover                        |
| 16  | Restart/fix golangci-lint LSP (phantom warnings)                      | Low      | 15min  | (d.4)                            |
| 17  | Summary.ByTag map                                                     | Medium   | 30min  | Carryover                        |
| 18  | Benchmark regression check vs baseline                                | Medium   | 1hr    | Carryover                        |
| 19  | Property test: `HasFix ⊇ IsAutoFixable`                               | Medium   | 15min  | Carryover                        |
| 20  | LineShiftMap Range/Column completeness                                | Medium   | 1hr    | Carryover                        |
| 21  | SubstringProvider nearest-position heuristic                          | Medium   | 1hr    | Carryover                        |
| 22  | Update USAGE_GUIDE.md                                                 | Medium   | 1hr    | Carryover                        |
| 23  | slices.Collect modernization pass                                     | Low      | 1hr    | Carryover                        |
| 24  | Position as value object refactor                                     | High     | 4hr    | Carryover (v2)                   |
| 25  | Explore sync.Pool for line offset index                               | Low      | 2hr    | Carryover                        |

---

## g) TOP #1 QUESTION 🤔

**Why does `go.sum` keep drifting after four cleanup commits, and is BuildFlow's `go-mod-tidy` / `go-mod-ignore-check` step the antagonist?**

I cannot resolve this myself because it depends on the interaction between your local BuildFlow configuration, the `vendorHash`/`proxyVendor` settings in `flake.nix`, and however `go mod tidy` is being invoked across sessions. The working tree currently shows `go.sum` removing 9 transitive test dependencies (`testify`, `go-spew`, `gkampitakis/*`, `joshdk/go-junit`, `kr/*`, `maruel/natural`, `mfridman/tparse`, `pmezard/go-difflib`, `tidwall/gjson`) — but commits `c59b7da`, `0d2e193`, and `110269b` already tried to remove these. If I just commit the drift, it may recur next session. **I need to know: should I (a) commit the current `go.sum` as-is, (b) run `go mod tidy` fresh and commit that, or (c) leave it untouched because BuildFlow owns it?** I left it unstaged to avoid worsening the ping-pong.

---

## Files Changed This Session

| File                          | Change                                                      | Committed                       |
| ----------------------------- | ----------------------------------------------------------- | ------------------------------- |
| `fuzz_test.go`                | Fixed `FuzzMergeByPosition` oracle (empty-file semantics)   | ✅ `b2d6388`                    |
| `pipeline/export_test.go`     | NEW: `export_test.go` bridge (helpers + `Export*` wrappers) | ✅ `b2d6388`                    |
| `pipeline/bdd_test.go`        | Use `ExportRangeFix` / `ExportOffsetFix`                    | ✅ `b2d6388`                    |
| `pipeline/fix_engine_test.go` | Use `rangeFix` / `offsetFixWithID` directly                 | ✅ `b2d6388`                    |
| `go.sum`                      | Drift (9 transitive deps removed)                           | ⬜ Left unstaged (see Question) |

---

_Assisted-by: Crush <crush@charm.land>_
