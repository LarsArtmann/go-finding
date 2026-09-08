# Session 19 — Fuzz Test Fix & v0.9.0 Release

**Date:** 2026-06-18 10:10 CEST
**Branch:** master (pushed)
**Commits this session:** 2
**Current tag:** v0.9.0

---

## Executive Summary

This session was triggered by a CI failure: `test-fuzz: [Failed] 20 fuzz tests failed (3 passed)`. Root cause identified and fixed — three fuzz target names were regex substrings of other names, causing Go's `-fuzz` flag (which uses regex matching) to match multiple targets and refuse to run.

---

## a) FULLY DONE ✅

### Fuzz Target Naming Collision Fix

| Old Name                              | New Name                    | Collision                                      | Fixed |
| ------------------------------------- | --------------------------- | ---------------------------------------------- | ----- |
| `FuzzFilter`                          | `FuzzFilterPredicates`      | Matched FuzzFilterByFile, FuzzFilterBySeverity | ✅    |
| `FuzzGroupBy`                         | `FuzzGroupByKey`            | Matched FuzzGroupByFile                        | ✅    |
| `FuzzApplyEditsToContent_Overlapping` | `FuzzApplyOverlappingEdits` | Matched by FuzzApplyEditsToContent             | ✅    |

**Verification:** All 23 fuzz targets now match exactly 1 function when used with `-fuzz=<name>`. Seed corpus directories renamed to match. All tests pass with race detector.

### v0.9.0 Release (from prior session, confirmed stable)

- Tag `v0.9.0` pushed
- 11 split-brain issues resolved
- 1,506 test functions pass
- 93.6% root coverage, 95.2% pipeline coverage

---

## b) PARTIALLY DONE 🟡

Nothing new this session. Prior partial items remain:

- Deprecated APIs scheduled for v1.0.0 removal (7 entries)
- Category/Tags full deprecation (ADR #13 plan)
- CLI FixProviders config resolution

---

## c) NOT STARTED ⬜

- v1.0.0 release (cut tag, remove deprecated APIs, unexport Report.Findings)
- Update FEATURES.md, README.md, USAGE_GUIDE.md for v0.9.0
- Named string types (ToolName, RuleName, FindingID)
- Godoc examples for new APIs (Normalized, WithFix, NormalizeFixStrategy)

---

## d) TOTALLY FUCKED UP 💥

### The Fuzz Naming Bug Itself

This bug has been latent since the fuzz tests were first written. `FuzzFilter` was always a substring of `FuzzFilterByFile`, but the CI either didn't have a fuzz step or didn't catch it until now. The fix was trivial (rename 3 functions + their seed corpus directories), but the diagnosis required understanding that Go's `-fuzz` flag uses **regex matching**, not exact matching.

**Lesson:** When naming fuzz test functions, ensure NO name is a substring of any other name in the same package. This is a Go-specific footgun that isn't documented anywhere obvious.

---

## e) WHAT WE SHOULD IMPROVE 🔧

1. **Add a lint check for fuzz name collisions** — A simple script that checks all `func Fuzz*` names against each other for substring matches. Could be a pre-commit hook.
2. **Document the `-fuzz` regex behavior** — Add a note to CONTRIBUTING.md or the testing section of AGENTS.md.
3. **Consider a CI fuzz step in ci.yml** — Currently there's no fuzz testing in GitHub Actions CI. Add a job that runs each fuzz target for 30s.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

| #  | Task                                                   | Impact   | Effort |
| -- | ------------------------------------------------------ | -------- | ------ |
| 1  | Update FEATURES.md with v0.9.0 changes                 | High     | 30min  |
| 2  | Update README.md for v0.9.0                            | High     | 30min  |
| 3  | Add CI fuzz step to ci.yml                             | Medium   | 30min  |
| 4  | Add fuzz name collision check script                   | Low      | 15min  |
| 5  | Add godoc examples: Normalized, WithFix                | Medium   | 30min  |
| 6  | Named string types: ToolName, RuleName, FindingID      | High     | 2hr    |
| 7  | v1.0.0: remove deprecated APIs                         | Critical | 2hr    |
| 8  | v1.0.0: unexport Report.Findings                       | High     | 1hr    |
| 9  | Cut v1.0.0 tag                                         | Critical | 1hr    |
| 10 | Complete CLI FixProviders config                       | Medium   | 2hr    |
| 11 | LineShiftMap Range/Column completeness                 | Medium   | 1hr    |
| 12 | SubstringProvider nearest-position heuristic           | Medium   | 1hr    |
| 13 | Update USAGE_GUIDE.md                                  | Medium   | 1hr    |
| 14 | Integration test: full pipeline with split-brain fixes | Medium   | 1hr    |
| 15 | Summary.ByTag map                                      | Medium   | 30min  |
| 16 | Benchmark regression check                             | Medium   | 1hr    |
| 17 | Fuzz: Position sentinel consistency                    | Medium   | 30min  |
| 18 | Fuzz: FixStrategy normalization idempotency            | Medium   | 30min  |
| 19 | Property test: HasFix ⊇ IsAutoFixable                  | Medium   | 15min  |
| 20 | slices.Collect modernization pass                      | Low      | 1hr    |
| 21 | Clean up working directory binaries                    | Low      | 5min   |
| 22 | Consider cmp.Ordered for Compare methods               | Low      | 1hr    |
| 23 | Add OpenAPI/JSON schema generation                     | Low      | 2hr    |
| 24 | Position as value object refactor                      | High     | 4hr    |
| 25 | Explore sync.Pool for line offset index                | Low      | 2hr    |

---

## g) TOP #1 QUESTION 🤔

**The BuildFlow `test-fuzz` step isn't in `.buildflow.yml` or `ci.yml`. Where is it configured, and what exact command does it run?**

Understanding the exact invocation would let me verify the fix works in CI before pushing. I've verified locally that all 23 targets run individually without ambiguity, but I can't confirm the BuildFlow step passes without seeing its implementation.

---

_Assisted-by: Crush <crush@charm.land>_
