# Session 21 — Post-v0.9.1 Release Status & go.sum War Analysis

**Date:** 2026-06-22 11:25 CEST
**Branch:** master (up to date with origin)
**Current tag:** v0.9.1 (tagged, signed, pushed)
**Working tree:** clean
**Commits this session:** 0 (report-only; v0.9.1 was released in the prior session and pushed)

---

## Executive Summary

v0.9.1 shipped — the broken fuzz suite from v0.9.0 is fixed, all 23 targets pass, the tag is pushed. The library is in its **healthiest state ever**: 667 tests green, 0 production TODOs, 35/42 features fully functional. But the release exposed a **systemic go.sum churn problem** that has now consumed **8 commits** across 3 sessions with no resolution in sight. This report documents the full state and lays out the path to v1.0.0.

---

## a) FULLY DONE ✅

### 1. v0.9.1 Released and Pushed

| Item                                                                     | Status |
| ------------------------------------------------------------------------ | ------ |
| `version.go` — `VersionPatch` 0→1                                        | ✅     |
| `CHANGELOG.md` — `[0.9.1]` section (Fixed/Changed/Added/Removed)         | ✅     |
| `README.md` — doc rot fixed (`"0.7.0"` → `"0.9.1"`)                      | ✅     |
| `docs/PRO_CONTRA_go-workflow-adoption.md` — maturity `v0.9.0` → `v0.9.1` | ✅     |
| Tag `v0.9.1` (annotated, SSH-signed)                                     | ✅     |
| Pushed to `origin/master`                                                | ✅     |
| Full `-race` suite green                                                 | ✅     |

### 2. v0.9.0 Fuzz Bug Definitively Fixed

Both root causes of the original `test-fuzz: [Failed] 20 fuzz tests failed (3 passed)` are resolved:

- **Naming collision** (Session 19): 3 fuzz targets renamed for regex uniqueness
- **Wrong oracle** (Session 20): `FuzzMergeByPosition` corrected for empty-file semantics

### 3. Health Metrics (Best Ever)

| Metric                         | Value                        |
| ------------------------------ | ---------------------------- |
| Test functions                 | 667                          |
| Fuzz targets                   | 23                           |
| Benchmarks                     | 48                           |
| Features FULLY_FUNCTIONAL      | 35                           |
| Features PARTIALLY_FUNCTIONAL  | 7                            |
| Features PLANNED / BROKEN      | 0 / 0                        |
| TODO/FIXME in production code  | 0                            |
| gofmt issues                   | 0                            |
| Full `-race` suite             | green                        |
| `golangci-lint run ./...` exit | 0 (3 warnings, non-blocking) |

---

## b) PARTIALLY DONE 🟡

| Item                      | Status                 | Detail                                                                                                                                                                                                |
| ------------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Lint cleanliness          | 3 warnings             | `gocyclo` on `Validate()` (complexity 32 > 25), `varnamelen` on `fs` (`finding_validate.go:44`) and `rt` (`splitbrain_test.go:101`). Pre-existing, non-blocking, but block a "lint clean" claim.      |
| Deprecated APIs           | 7 entries              | `Report.Findings`, `Report.Merge()`, `OnStage`, `Metrics.RecordFix()`, `CountBySeverity()`, `Finding.Tag`, `ConflictDetector`/`Verifier` structs. All scheduled for v1.0.0 removal.                   |
| CLI FixProviders config   | Incomplete             | ConfigFile → ResolveProviders path exists but error handling for unknown providers is rough.                                                                                                          |
| Category/Tags deprecation | ADR #13 planned        | Twin-method contract documented but migration path not yet enforced.                                                                                                                                  |
| LSP diagnostics           | Stale phantom warnings | `golangci_lint_ls` reports `gochecknoglobals` on `export_test.go` Export\* funcs that are actually `func`s (not `var`s). `golangci-lint` CLI does not reproduce. LSP restart failed in prior session. |

---

## c) NOT STARTED ⬜

### v1.0.0 Release (17 unchecked items in `docs/v1.0-release-criteria.md`)

| Category        | Open Items                                                                                                                                                  |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| API Stability   | Remove deprecated symbols, remove phantom `FixStrategyAI`, finalize `NewFinding` signature, audit naming consistency, decide on `Properties map[string]any` |
| Documentation   | README covers all types, CHANGELOG for v1.0, API stability guarantee documented, godoc comprehensive, examples compile                                      |
| Testing         | Coverage ≥ 95% (currently met), all fuzz tests pass (✅ met), property tests pass, integration test with downstream consumer                                |
| Pipeline Module | Provider location decision, FixApplier goroutine leak verified, Pipeline.Run() reusability confirmed                                                        |

### Other Not-Started

- Named string types (`ToolName`, `RuleName`, `FindingID`)
- Godoc examples for `Normalized`, `WithFix`, `NormalizeFixStrategy`
- CI fuzz job in `ci.yml`
- `scripts/fuzz-check.sh` (clean-cache per-target diagnostics)

---

## d) TOTALLY FUCKED UP 💥

### 1. The go.sum Ping-Pong (8 COMMITS, UNRESOLVED)

This is the single biggest embarrassment in the project's history. **Eight commits** across three sessions have churned `go.sum` back and forth with zero resolution:

```
9fbef16 chore: restore go.sum to go mod tidy canonical state
d7e05d8 chore: remove transitive test-only dependencies from go.sum
d909710 chore: canonicalize go.sum to match go mod tidy output
e5d5500 chore: canonicalize go.sum to match go mod tidy output
110269b chore: remove unused transitive test dependencies from go.sum
0d2e193 chore: add transitive test dependencies to go.sum
c59b7da chore: clean up go.sum by removing unused transitive dependencies
```

**Root cause:** BuildFlow's `go-mod-ignore-check` step strips transitive test dependencies (`testify`, `go-spew`, `gkampitakis/*`, etc.) that `go mod tidy` restores. These ARE legitimate transitive deps of `ginkgo`/`gomega` and `go-faster/yaml`. Every commit triggers BuildFlow's pre-commit hook, which strips them; the next `go mod tidy` restores them; the next commit strips them again. **Infinite loop.**

**Why it's fucked:** This pollutes git history, makes `go mod verify` fail for downstream consumers who run `go mod tidy`, and wastes review bandwidth. The v0.9.1 release alone has **3 go.sum commits** — nearly as many as the actual code changes.

**The fix is in BuildFlow, not go-finding:** `go-mod-ignore-check` needs to either (a) not strip transitive deps that `go mod tidy` legitimately includes, or (b) be disabled for this project. This cannot be fixed from within go-finding.

### 2. The `test-fuzz` Error Message Is Still Structurally Ambiguous

Documented in Session 20 but unfixed: "20 fuzz tests failed (3 passed)" can mean compile error, cache staleness, or real invariant violation. BuildFlow's step does not distinguish them. Two sessions were burned misdiagnosing this. `scripts/fuzz-check.sh` would kill the ambiguity but hasn't been written.

### 3. v0.9.0 Ships Broken

The v0.9.0 tag contains the fuzz naming collision bug. Anyone who pins v0.9.0 and runs `buildflow check test-fuzz` gets 20/23 failures. v0.9.1 fixes this, but v0.9.0 cannot be un-published. The CHANGELOG documents the fix but the tag itself is permanently broken.

---

## e) WHAT WE SHOULD IMPROVE 🔧

1. **Fix the go.sum ping-pong at the BuildFlow level** — This is Priority #1. Either disable `go-mod-ignore-check` for go-finding or fix its transitive-dep stripping logic. Until this is done, every commit will keep churning go.sum. **This is blocking v1.0.0 readiness.**
2. **Add `scripts/fuzz-check.sh`** — A 30-line script that runs each fuzz target from a clean cache with explicit per-target pass/fail. Kills the ambiguous error message fuckup permanently.
3. **Fix the 3 lint warnings** — Extract `Validate()` into sub-functions (gocyclo), rename `fs`→`strategy` and `rt`→`result`. 15 minutes of work.
4. **Add a CI fuzz job** — Now that the exact BuildFlow command is known (`go test -run=N -fuzz=N -fuzztime=30s -shuffle=on`), port it to `ci.yml`. ~30 min.
5. **Audit all 23 fuzz oracles** for empty/zero-value assumptions. `FuzzMergeByPosition` was wrong; others may be too.
6. **Start the v1.0.0 API cleanup** — 7 deprecated APIs need removal. This is the critical path to v1.0.0.
7. **Restart/fix the golangci-lint LSP** — Phantom warnings erode trust in IDE diagnostics.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

| #   | Task                                                                                 | Impact   | Effort | Category  |
| --- | ------------------------------------------------------------------------------------ | -------- | ------ | --------- |
| 1   | **Fix go.sum ping-pong in BuildFlow** (disable go-mod-ignore-check or fix stripping) | Critical | 1hr    | BuildFlow |
| 2   | **Remove 7 deprecated APIs** for v1.0.0                                              | Critical | 2hr    | v1.0.0    |
| 3   | **Unexport `Report.Findings`** (use `FindingsSnapshot()`)                            | Critical | 1hr    | v1.0.0    |
| 4   | **Cut v1.0.0 tag**                                                                   | Critical | 1hr    | Release   |
| 5   | Add `scripts/fuzz-check.sh` (kill ambiguous error)                                   | High     | 30min  | Testing   |
| 6   | Fix 3 lint warnings (gocyclo, varnamelen ×2)                                         | High     | 15min  | Quality   |
| 7   | Add CI fuzz job to `ci.yml`                                                          | High     | 30min  | CI        |
| 8   | Audit all 23 fuzz oracles for empty/zero assumptions                                 | High     | 1hr    | Testing   |
| 9   | Named string types: `ToolName`, `RuleName`, `FindingID`                              | High     | 2hr    | Types     |
| 10  | Integration test with downstream consumer (BuildFlow/hierarchical-errors)            | High     | 2hr    | Testing   |
| 11  | Update FEATURES.md for v0.9.1                                                        | Medium   | 30min  | Docs      |
| 12  | Document API stability guarantee (Go compat promise style)                           | Medium   | 1hr    | Docs      |
| 13  | Add godoc examples: `Normalized`, `WithFix`                                          | Medium   | 30min  | Docs      |
| 14  | Remove phantom `FixStrategyAI` constant                                              | Medium   | 15min  | Cleanup   |
| 15  | Decide `Properties map[string]any` vs `Metadata` for SARIF                           | Medium   | 2hr    | Design    |
| 16  | Document `-fuzz` regex+cache behavior in CONTRIBUTING.md                             | Low      | 15min  | Docs      |
| 17  | Complete CLI FixProviders config                                                     | Medium   | 2hr    | CLI       |
| 18  | Summary.ByTag map                                                                    | Medium   | 30min  | Feature   |
| 19  | Benchmark regression check vs baseline                                               | Medium   | 1hr    | Perf      |
| 20  | Property test: `HasFix ⊇ IsAutoFixable`                                              | Medium   | 15min  | Testing   |
| 21  | LineShiftMap Range/Column completeness                                               | Medium   | 1hr    | Pipeline  |
| 22  | SubstringProvider nearest-position heuristic                                         | Medium   | 1hr    | Pipeline  |
| 23  | Fix golangci-lint LSP phantom warnings                                               | Low      | 15min  | Tooling   |
| 24  | Position as value object refactor (v2)                                               | High     | 4hr    | v2.0      |
| 25  | Explore sync.Pool for line offset index                                              | Low      | 2hr    | Perf      |

---

## g) TOP #1 QUESTION 🤔

**Should `go-mod-ignore-check` be disabled for this project via `.buildflow.yml`, or should the stripping logic be fixed in BuildFlow itself?**

I cannot resolve this myself because it requires a decision about BuildFlow's design intent. The `go-mod-ignore-check` step strips transitive test dependencies from `go.sum`, but `go mod tidy` (the Go toolchain source of truth) restores them. This creates an infinite churn. The two options:

- **(a) Disable in go-finding:** Add `go-mod-ignore-check` to an exclude list in `.buildflow.yml`. Quick fix, but means the step won't run at all — losing its other checks.
- **(b) Fix in BuildFlow:** Make `go-mod-ignore-check` stop stripping deps that `go mod tidy` includes. Correct fix, but requires a BuildFlow release.

I need to know which approach you want before I touch `.buildflow.yml` or submit a BuildFlow PR. Until then, every commit will keep churning go.sum.

---

## Session History (Sessions 18-21)

| Session | Date             | Headline                              | Key Outcome                                    |
| ------- | ---------------- | ------------------------------------- | ---------------------------------------------- |
| 18      | 2026-06-18 08:30 | Split-brain resolution complete       | 11 issues fixed, v0.9.0 prepped                |
| 19      | 2026-06-18 10:10 | Fuzz naming collision fix             | 3 targets renamed, v0.9.0 released (broken)    |
| 20      | 2026-06-18 19:01 | Fuzz oracle fix & BuildFlow mechanism | `FuzzMergeByPosition` oracle fixed, 23/23 pass |
| 21      | 2026-06-22 11:25 | Post-v0.9.1 release status            | v0.9.1 shipped, go.sum war documented          |

---

_Assisted-by: Crush <crush@charm.land>_
