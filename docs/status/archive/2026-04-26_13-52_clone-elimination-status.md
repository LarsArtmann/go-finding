# Clone Elimination Status Report

**Date:** 2026-04-26 13:52
**Current Clone Groups:** 68
**Starting Clone Groups:** 76
**Reduction:** 8 groups (10.5%)

---

## Executive Summary

Clone elimination effort has reduced clone groups from **76 to 68** (10.5% reduction). However, reaching **ZERO clone groups** is not fully achievable without significant architectural changes or accepting limitations.

### Key Constraints

1. **Architectural Clone (22 clones)** - The `Detector` interface signature `func(ctx context.Context) ([]finding.Finding, error)` appears in 22 locations. This is an INHERENT architectural pattern that cannot be eliminated without breaking the public API.

2. **Cross-Package Clones (12+8 clones)** - Patterns in `Position{File: "..."}` and `NewReport(ToolInfo{...})` appear across `cmd/`, `examples/`, and `pipeline/` packages. These cannot be consolidated without polluting the production API.

---

## Clone Groups by Status

| Status                | Count | Groups                                |
| --------------------- | ----- | ------------------------------------- |
| **FULLY DONE**        | 2     | Eliminated (11-clone, 5-clone groups) |
| **PARTIALLY DONE**    | 0     | -                                     |
| **NOT STARTED**       | 64    | All remaining groups                  |
| **TOTALLY FUCKED UP** | 0     | -                                     |

---

## Clone Groups Inventory

### TIER 1: CANNOT FIX (Architectural/Inherent)

| Group              | Clones | Pattern                                                | Files                                     | Fixable?                     |
| ------------------ | ------ | ------------------------------------------------------ | ----------------------------------------- | ---------------------------- |
| Detector interface | 22     | `func(ctx context.Context) ([]finding.Finding, error)` | `pipeline/`, `internal/detectors/`, tests | **NO** - Breaking API change |

### TIER 2: HIGH IMPACT (10-12 clones)

| Group                   | Clones | Pattern                            | Files                                   | Fixable?                                   |
| ----------------------- | ------ | ---------------------------------- | --------------------------------------- | ------------------------------------------ |
| Position struct literal | 12     | `Position{File: "...", Line: N}`   | `cmd/`, `examples/`, `pipeline_test.go` | **HARD** - Requires production API helpers |
| ToolInfo struct literal | 8      | `NewReport(ToolInfo{Name: "..."})` | `cmd/`, `examples/`                     | **HARD** - Requires production API helpers |

### TIER 3: MEDIUM IMPACT (4-8 clones)

| Group                           | Clones | Pattern                                           | Files                    | Fixable?                |
| ------------------------------- | ------ | ------------------------------------------------- | ------------------------ | ----------------------- |
| position_extra_test.go Range    | 8      | `Range{Start: Position{...}, End: Position{...}}` | `position_extra_test.go` | **YES** - Refactor test |
| position_extra_test.go Position | 6      | `Position{...}` inline declarations               | `position_extra_test.go` | **YES** - Use helper    |
| cmd/+filter_test.go             | 6      | `Position{...}` in Finding struct                 | `cmd/`, `filter_test.go` | **YES** - Use testify   |
| report_test.go struct fields    | 6      | `if r.X != v { t.Errorf(...) }`                   | `report_test.go`         | **YES** - Use testify   |
| pipeline/retry_test.go          | 4      | `assertIntEq(t, len(...), N, "...")`              | `pipeline/retry_test.go` | **YES** - Use testify   |

### TIER 4: LOW IMPACT (3-4 clones)

| Group                      | Clones | Pattern                       | Files                                           | Fixable?                     |
| -------------------------- | ------ | ----------------------------- | ----------------------------------------------- | ---------------------------- |
| json_test.go               | 4      | JSON field comparisons        | `json_test.go`                                  | **YES** - Use testify        |
| position.go severity       | 4      | `SeverityXXX.String()` switch | `position.go`                                   | **YES** - Production code    |
| severity.go                | 4      | `SeverityFromString` switch   | `severity.go`                                   | **YES** - Production code    |
| export+severity_test.go    | 4      | Table test patterns           | `export_test.go`, `severity_test.go`            | **YES** - Use testify        |
| position.go                | 4      | `Position.HasOffset()` method | `position.go`                                   | **YES** - Production         |
| pipeline/testutil+testutil | 4      | `assertIntEq` helper calls    | `pipeline/testutil_test.go`, `testutil_test.go` | **YES** - Inline or remove   |
| verify_test.go             | 4      | Table test patterns           | `pipeline/verify_test.go`                       | **YES** - Use testify        |
| detectors_test.go + sarif  | 4      | Finding field assertions      | `internal/detectors/`, `sarif_test.go`          | **YES** - Use testify        |
| detectors_test.go table    | 3      | Table test patterns           | `internal/detectors/`                           | **YES** - Already refactored |

### TIER 5: MINIMAL (2-3 clones, 40+ groups)

All remaining 2-clone groups. **YES** - Most can be eliminated with testify, but tedious work.

---

## Clone Group Distribution

```
22 clones:  1 group  (Architectural - CANNOT FIX)
12 clones:  1 group  (HIGH - Hard to fix)
 8 clones:  3 groups (MEDIUM/HIGH)
 6 clones:  4 groups (MEDIUM)
 4 clones: 11 groups (LOW)
 3 clones: 11 groups (LOW)
 2 clones: 37 groups (MINIMAL - Tedious)
```

---

## Work Completed

### FULLY DONE

1. **11-clone Finding field comparison group** (export_test.go)
   - Replaced `if f.Field != val { t.Errorf(...) }` patterns with `assert.Equal(t, val, f.Field)`
   - Files: `export_test.go`

2. **5-clone unused helper function group**
   - Removed dead code: `assertFloatEq`, `assertStringEq`, `assertBoolEq`, `assertSevEq`
   - Files: `testutil_test.go`

### Additional Refactoring Done

- **detectors_test.go**: Refactored Finding field assertions to use testify
- **testutil_test.go**: Cleaned up, removed unused helpers

---

## What We Should Improve

1. **Systematic testify conversion** - Convert all remaining `if/Errorf` patterns to testify assertions
2. **Test helper consolidation** - Create shared helpers in `testutil_test.go` that are ACTUALLY used
3. **Example file refactoring** - The `examples/` directory has many clone patterns but examples should be readable
4. **Production API helpers** - Consider adding `finding.Position(file, line, col)` helpers to reduce struct literal clones (but this pollutes API)

---

## Top #25 Things To Get Done Next

| #  | Priority | Task                                      | Clones Eliminated | Effort |
| -- | -------- | ----------------------------------------- | ----------------- | ------ |
| 1  | HIGH     | Convert report_test.go to testify         | 6                 | LOW    |
| 2  | HIGH     | Convert json_test.go to testify           | 4                 | LOW    |
| 3  | HIGH     | Convert filter_test.go to testify         | 6 (partial)       | LOW    |
| 4  | MEDIUM   | Convert position_extra_test.go            | 18 (6+4+8)        | MEDIUM |
| 5  | MEDIUM   | Convert detectors_test.go remaining       | 4                 | LOW    |
| 6  | MEDIUM   | Convert sarif_test.go to testify          | ~15               | MEDIUM |
| 7  | MEDIUM   | Convert retry_test.go to testify          | 4                 | LOW    |
| 8  | MEDIUM   | Convert verify_test.go to testify         | 4                 | LOW    |
| 9  | MEDIUM   | Convert fix_applier_test.go to testify    | 3                 | LOW    |
| 10 | MEDIUM   | Convert metrics_test.go to testify        | 3                 | LOW    |
| 11 | MEDIUM   | Convert partial_test.go to testify        | 3                 | LOW    |
| 12 | MEDIUM   | Convert bench_test.go to testify          | 3                 | LOW    |
| 13 | MEDIUM   | Convert cmd/go-finding tests to testify   | 6 (partial)       | MEDIUM |
| 14 | MEDIUM   | Convert export_test.go severity tests     | 4                 | LOW    |
| 15 | LOW      | Convert merge_test.go to testify          | 3                 | LOW    |
| 16 | LOW      | Convert equal_test.go to testify          | 3                 | LOW    |
| 17 | LOW      | Convert diagnostic_test.go + lsp_test.go  | 3                 | LOW    |
| 18 | LOW      | Convert fuzz_test.go to testify           | 3                 | LOW    |
| 19 | LOW      | Inline/remove assertIntEq calls           | 4                 | LOW    |
| 20 | LOW      | Convert conflict_test.go to testify       | 2                 | LOW    |
| 21 | LOW      | Convert conflict_extra_test.go to testify | 2                 | LOW    |
| 22 | LOW      | Convert pipeline_test.go remaining        | ~10               | MEDIUM |
| 23 | LOW      | Convert example tests (optional)          | ~10               | HIGH   |
| 24 | LOW      | Remove cross-package clones               | ~6                | HARD   |
| 25 | LOW      | Final cleanup and verification            | -                 | LOW    |

---

## Top #1 Question I Cannot Figure Out

**How to eliminate the 12-clone `Position{File: "..."}` pattern WITHOUT either:**

1. Adding helper functions to the production `finding` package API (pollutes the library)
2. Refactoring all example files to use verbose multi-line position construction (hurts readability)
3. Accepting that struct literal patterns in examples are intentional and fine

The `art-dupl` semantic clone detection treats any `Position{File: "main.go", Line: 1}` as identical regardless of surrounding context. This means even if we add helper functions to the TEST package (`testutil_test.go`), example files (which are in `finding_test` package) cannot use internal test helpers.

**Possible approaches:**

1. Accept this as an inherent limitation of code examples
2. Create a "test helpers" subpackage that's exported
3. Lower the clone detection threshold (but then we miss real duplicates)
4. Accept the clones in example files as "intentional documentation"

---

## Files Modified (Session)

| File                                                       | Lines Changed | Purpose            |
| ---------------------------------------------------------- | ------------- | ------------------ |
| `export_test.go`                                           | +11, -31      | testify assertions |
| `testutil_test.go`                                         | +17, -100     | testify + cleanup  |
| `internal/detectors/detectors_test.go`                     | +17, -83      | testify assertions |
| `docs/status/2026-04-26_13-28_clone-elimination-status.md` | NEW           | Status report      |

---

## Git History (Clone Elimination)

```
8d0eee1 refactor: use testify assertions and remove dead code
70a9bee refactor: use testify for Finding field assertions in export_test
8a0ead2 docs: add clone elimination status report for 2026-04-26
bf31f23 refactor: replace manual test assertions with testify
35ffaf4 chore: add testify dependency for test assertions
c97b7c3 fix: add missing assertIntEq helper to pipeline testutil
```

---

## Test Status

```
ok  github.com/larsartmann/go-finding        (cached)
ok  github.com/larsartmann/go-finding/cmd/go-finding   (cached)
ok  github.com/larsartmann/go-finding/internal/detectors   (cached)
ok  github.com/larsartmann/go-finding/pipeline   (cached)
```

All tests pass. ✅

---

## Conclusion

- **Progress:** 76 → 68 clone groups (10.5% reduction)
- **Achievable target:** ~30-40 clone groups (if we convert all test files to testify)
- **Unavoidable:** ~1 group (22-clone architectural)
- **Questionable value:** Cross-package clones in examples (~20 clones)

**Recommendation:** Accept current state OR invest significant effort in converting all remaining test files to testify assertions. Zero clones is not practical without either breaking API changes or polluting the production package with test-only helpers.
