# Status Report: Semantic Deduplication Session

**Date:** 2026-05-21 17:24
**Session Focus:** Eliminate code duplication detected by `art-dupl --semantic`
**Status:** ✅ COMPLETE — 15 → 1 clone groups (93% reduction)

---

## A) FULLY DONE

### Semantic Deduplication (15 → 1 clone groups)

All 15 clone groups identified by `art-dupl -t 40 --semantic` resolved across 10 files:

| #   | File(s)                                        | Technique                                                                          | Before   | After   |
| --- | ---------------------------------------------- | ---------------------------------------------------------------------------------- | -------- | ------- |
| 1   | `pipeline/fix_engine_test.go`                  | `FixEdit.Overlaps`: 3 subtests → table-driven                                      | 3 clones | 0       |
| 2   | `pipeline/pipeline_bugfix_test.go`             | Extracted `directFix()` helper                                                     | 3 clones | 0       |
| 3   | `equal_test.go`                                | Extracted `testFindingEqualCases()` + `testFindingEqualCase()`                     | 3 clones | 0       |
| 4   | `pipeline/fix_engine_test.go`                  | `lineColToOffset`: 4 subtests → table-driven                                       | 3 clones | 0       |
| 5   | `example_test.go`                              | Extracted `newExampleFinding()` helper                                             | 2 clones | 0       |
| 6   | `report_extra_test.go`                         | Extracted `suppressionFinding()` helper                                            | 2 clones | 0       |
| 7   | `pipeline/bdd_test.go`                         | Extracted `singleFindingDetector()` helper                                         | 2 clones | 0       |
| 8   | `cmd/go-finding/main_test.go`                  | Extracted `testOutputSerializationError()` helper                                  | 2 clones | 0       |
| 9   | `pipeline/fix_engine_test.go`                  | `NearestLineMatch`: 2 subtests → table-driven                                      | 2 clones | 0       |
| 10  | `pipeline/fix_engine_test.go`                  | `LineRange` SingleLine + OutOfBounds → merged table-driven                         | 2 clones | 0       |
| 11  | `sarif_test.go`                                | Extracted `testSARIFNaNError()` helper                                             | 2 clones | 0       |
| 12  | `pipeline/bdd_test.go`                         | 3 detector+config+pipeline tests → `DescribeTable` + `runSingleDetectorPipeline()` | 2 clones | 0       |
| 13  | `example_test.go` ↔ `examples/builder/main.go` | **ACCEPTED** — cross-package intentional duplication                               | 1 clone  | 1 clone |
| 14  | `pipeline/fix_engine_test.go`                  | SubstringReplace + SubstringNotFound → merged table-driven                         | 2 clones | 0       |
| 15  | `coverage_test.go` + `finding_valid_test.go`   | Extracted `runIsValidTests()` helper                                               | 2 clones | 0       |

### Net Result

- **10 files changed**: 276 insertions, 382 deletions (net -106 lines)
- **All tests pass**: `go test -race -count=1 ./...` ✅
- **Coverage**: 95.0% total (root: 98.6%, pipeline: 96.5%, analysis: 98.5%)
- **1 remaining clone**: `example_test.go` ↔ `examples/builder/main.go` — cross-package intentional duplication (testable example vs standalone demo program)

### Techniques Applied

1. **Table-driven tests** — Convert subtests with identical structure to table-driven patterns (groups 1, 4, 9, 10, 14)
2. **Helper extraction** — Extract repeated struct/function construction into named helpers (groups 2, 3, 5, 6, 7, 8, 11, 15)
3. **Ginkgo DescribeTable** — Convert repetitive `It` blocks to `DescribeTable` with `Entry` rows (group 12)
4. **Accept intentional duplication** — Cross-package examples that serve different audiences (group 13)

---

## B) PARTIALLY DONE

None. All identified work is either fully complete or accepted.

---

## C) NOT STARTED

These are **not part of this session's scope** but tracked from TODO_LIST.md:

1. FixEngine: line-offset tracking for cumulative line shifts across multi-fix
2. Make fix strategy composable as interface — register `FixApplier` implementations per strategy
3. Add pipeline stage hooks — pre/post hooks for detect, triage, fix, verify
4. Improve `README.md` — add badges, pipeline examples, API overview
5. Update `USAGE_GUIDE.md` for v0.3.0 — new features not documented
6. Document provider chain in user-facing docs
7. Write API stability guarantee document
8. Define v1.0.0 release criteria
9. Clean up gopls hints (~12 non-critical)
10. Fix pre-commit hook failures (goconst, todo-check, library-policy)
11. Fix `.golangci.yml` indentation loop
12. API stability review — audit all exported symbols for v1.0.0 lock

---

## D) TOTALLY FUCKED UP

**Nothing.** All changes compiled and passed tests on first attempt. No reverts needed.

Minor hiccup: One Python patching script had a typo (`\tt` instead of `\t\t\t\t`) which was caught immediately by the compiler and fixed in seconds.

---

## E) WHAT WE SHOULD IMPROVE

### Test Code Quality

1. **Table-driven test adoption** — Many test files still use individual `t.Run` blocks with identical structure. Project-wide table-driven audit would find more opportunities.
2. **golangci-lint goconst warnings** — 49 warnings about repeated string literals. Many already have constants defined but aren't used consistently.
3. **errcheck warnings** — 7 unchecked error returns, mostly in test/example code (`FormatText`, `FormatMarkdown`, `h.Write`).
4. **err113 warning** — `pipeline.go:91` defines a dynamic error with `errors.New()` instead of a static sentinel.

### Architecture

5. **Pipeline test helpers** — The `runSingleDetectorPipeline()` and `singleFindingDetector()` helpers could be promoted to a shared `pipeline/testhelpers` package for reuse across test files.
6. **Test file organization** — `coverage_test.go` and `finding_valid_test.go` have overlapping `TestFindingIsValid` tests that could potentially be consolidated into one file.

### Process

7. **Pre-commit hooks broken** — goconst, todo-check, and library-policy hooks all fail. Needs fixing before CI adoption.
8. **No CI/CD pipeline** — No `.github/workflows/` exists. All quality gates are manual.
9. **Nix flake migration incomplete** — `justfile` still exists but is deprecated per project policy.

---

## F) Top 25 Things We Should Get Done Next

### P0 — Quality Gate (blocks release)

1. **Fix pre-commit hook failures** — goconst, todo-check, library-policy
2. **Fix err113: static error in pipeline.go:91** — Replace `errors.New()` with sentinel
3. **Fix errcheck: 7 unchecked error returns** — Especially `h.Write` in `id.go`
4. **Define v1.0.0 release criteria** in `docs/v1.0-release-criteria.md`
5. **Write API stability guarantee document** — Go compat promise style

### P1 — Documentation

6. **Update USAGE_GUIDE.md for v0.3.0** — Pipeline, providers, formatting, diff
7. **Document provider chain** — OffsetProvider→LineProvider→SubstringProvider
8. **Improve README.md** — Badges, pipeline examples, API overview, quickstart
9. **Update architecture-decisions.md** — Deduplication session decisions

### P1 — Architecture

10. **FixEngine: line-offset tracking** — Cumulative line shifts across multi-fix
11. **Composable fix strategy interface** — Register `FixApplier` per strategy
12. **Pipeline stage hooks** — Pre/post hooks for detect, triage, fix, verify
13. **API stability audit** — Review all exported symbols for v1.0.0 lock
14. **`Finding` struct sub-grouping** — DEFERRED v2 but plan the layout

### P2 — Code Quality

15. **Resolve 49 goconst warnings** — Use existing constants consistently
16. **Clean up ~12 gopls hints** — rangeint, newexpr, mapsloop, stringsseq
17. **Table-driven test audit** — Find remaining subtests with identical structure
18. **Promote test helpers** — `runSingleDetectorPipeline`, `singleFindingDetector` to shared package
19. **Consolidate `TestFindingIsValid`** — Merge `coverage_test.go` and `finding_valid_test.go` versions

### P2 — Infrastructure

20. **Fix `.golangci.yml` indentation loop** — Auto-configure keeps reformatting
21. **Create CI/CD pipeline** — GitHub Actions workflow
22. **Migrate justfile → nix flake** — Per project policy
23. **Add `golines` to CI** — BLOCKED until CI exists
24. **Nix flake for examples/** — Reproducible builds for examples

### P3 — Future

25. **Position/Range zero-value decision** — OWNER_DECISION needed (breaking change)

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `Position` zero-value be "no location" (current: `Position{}` means empty/invalid) or should it mean "line 1, column 1, offset 0" (zero = start of file)?**

- Current behavior: `Position{}` → `IsZero() == true`, file/line/col all zero
- Problem: `Position{File: "a.go"}` is "valid" but `Position{}` is not — inconsistent
- Impact: Affects `Range.IsValid()`, `Finding.Validate()`, `Position.HasLocation()`
- Breaking: Changing this would affect every consumer's zero-value assumptions
- Marked as `OWNER_DECISION` in TODO_LIST.md because only you know the intended semantics

---

## Metrics Snapshot

| Metric                | Value                                                     |
| --------------------- | --------------------------------------------------------- |
| Clone groups (before) | 15                                                        |
| Clone groups (after)  | 1                                                         |
| Clone reduction       | 93%                                                       |
| Files changed         | 10                                                        |
| Net lines removed     | 106                                                       |
| Total Go LOC          | 25,638                                                    |
| Test LOC              | 18,377 (71.7% of codebase)                                |
| Test coverage         | 95.0%                                                     |
| Root package coverage | 98.6%                                                     |
| Pipeline coverage     | 96.5%                                                     |
| All tests pass        | ✅                                                        |
| Linter errors         | 0                                                         |
| Linter warnings       | ~49 (goconst), 7 (errcheck), 1 (err113), 4 (paralleltest) |

---

_Assisted-by: Crush <crush@charm.land>_
