# Status Report: Clone Elimination via testify Refactoring

**Date**: 2026-04-26\
**Time**: 19:06 UTC\
**Branch**: master (3 commits ahead of origin/master)\
**Goal**: Eliminate ALL code duplication detected by `art-dupl --semantic --sort total-tokens -t 15` to achieve ZERO clone groups.

---

## Executive Summary

**Progress**: Reduced clone groups from 76 → 70 (6 groups eliminated)\
**Status**: Build failing in `cmd/go-finding/integration_test.go` - need to apply testify assertions\
**Strategy**: Replace manual `if x != y { t.Errorf(...) }` assertions with testify calls (~9 tokens each, below 15-token threshold)

---

## A) WORK FULLY DONE ✅

### 1. Testify Dependency Added

- `go.mod` / `go.sum`: Added `github.com/stretchr/testify v1.11.1`

### 2. Root Package Test Files (Pipeline Package)

| File                              | Status      | Changes                                                                                                                                       |
| --------------------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline/fix_applier_test.go`    | ✅ Complete | Added testify import, replaced assertions with `assert.Equal`, `assert.Len`, `assert.ErrorIs`, `assert.NotNil`, `assert.True`                 |
| `pipeline/retry_test.go`          | ✅ Complete | Added testify import, replaced `errors.Is` checks with `assert.ErrorIs`, `if x != N` with `assert.Equal`                                      |
| `pipeline/partial_test.go`        | ✅ Complete | Added testify import, replaced `if len() != N`, boolean checks with `assert.Len`, `assert.True/False`, `assert.NotNil`, `assert.NoError`      |
| `pipeline/verify_test.go`         | ✅ Complete | Added testify import and require, replaced `assertDiffResult` helper and individual assertions with testify calls                             |
| `pipeline/conflict_test.go`       | ✅ Complete | Added testify import, replaced `if len() != N` and `if x != expected` with `assert.Len` and `assert.Equal`                                    |
| `pipeline/conflict_extra_test.go` | ✅ Complete | Added testify import, replaced `assertEqual` calls with `assert.Len`                                                                          |
| `pipeline/metrics_test.go`        | ✅ Complete | Added testify import, removed unused `assertEqual` helper, replaced all assertions with testify calls                                         |
| `pipeline/integration_test.go`    | ✅ Complete | Added testify import, replaced `if x != N`, boolean checks with `assert.Equal`, `assert.NotNil`, `assert.True/False`, `assert.GreaterOrEqual` |
| `pipeline/pipeline_test.go`       | ✅ Complete | Added testify import, replaced extensive assertions throughout (40+ replacements)                                                             |

### 3. cmd/go-finding Test Files

| File                          | Status      | Changes                                                                                                                                                     |
| ----------------------------- | ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cmd/go-finding/main_test.go` | ✅ Complete | Added testify import, replaced all string contains checks with `assert.Contains`, replaced `if x != y` with `assert.Equal`, removed unused `strings` import |

---

## B) WORK PARTIALLY DONE ⚠️

### 1. cmd/go-finding/integration_test.go ⚠️

- **Status**: Build failing - testify import added but not yet used
- **Files Modified**: 21 test files total modified but uncommitted
- **Issue**: `github.com/stretchr/testify/assert` imported but no assertions applied yet

---

## C) WORK NOT STARTED 🚫

### Remaining Clone Groups by Category

| Category                                     | Groups             | Files Affected                                                                          |
| -------------------------------------------- | ------------------ | --------------------------------------------------------------------------------------- |
| Finding struct literals with Position fields | 11                 | `export_test.go`, `detectors_test.go`, `sarif_test.go`                                  |
| Finding struct literals (12-clone)           | 12                 | `cmd/go-finding/main_test.go`, `example_*_test.go`, `pipeline/pipeline_test.go`         |
| Finding struct literals (8-clone)            | 8                  | `cmd/go-finding/main_test.go`, `example_*_test.go`                                      |
| `if x.Error() != "y"` pattern (10-clone)     | 10                 | `coverage_test.go`, `errors_test.go`, `id_test.go`, `merge_test.go`, `position_test.go` |
| `if x != N` in position tests (4/6/8-clone)  | 3 groups           | `position_extra_test.go`                                                                |
| Table-driven test assertions                 | 4 groups           | `json_test.go`, `detectors_test.go`                                                     |
| `if x != N` in filter/report tests           | 2 groups           | `filter_test.go`, `report_test.go`                                                      |
| Partial pipeline code (production)           | 1 group (3 clones) | `pipeline/partial.go`, `pipeline/pipeline.go`                                           |
| Metrics nil checks (production)              | 1 group (3 clones) | `pipeline/metrics_test.go`                                                              |
| Various test helpers                         | 10+ groups         | `severity_test.go`, `export_test.go`, `diagnostic_test.go`, etc.                        |

### Files Not Yet Modified

- `export_test.go` - Has 11-clone group for Finding struct literals
- `sarif_test.go` - Has 11-clone group for Finding struct literals
- `errors_test.go` - Has 10-clone group for `x.Error() != "y"` pattern
- `coverage_test.go` - Has 10-clone group
- `merge_test.go` - Has 10-clone group
- `report_test.go` - Has 6-clone and 4-clone groups
- `filter_test.go` - Has 6-clone group
- `json_test.go` - Has 4-clone and 3-clone groups
- `position_test.go` - Has 10-clone, 4-clone, 2-clone groups
- `position_extra_test.go` - Has 4/6/8-clone groups
- `severity_test.go` - Has 4-clone group
- `diagnostic_test.go` - Has 3-clone group
- `lsp_test.go` - Has 3-clone group
- `equal_test.go` - Has 3-clone group
- `example_test.go` - Has Finding struct literal clones
- `example_basic_test.go` - Has Finding struct literal clones
- `example_cli_test.go` - Has Finding struct literal clones
- `fuzz_test.go` - Has 3-clone and 2-clone groups
- `bench_test.go` - Has 3-clone and 2-clone groups
- `finding_test.go` - Has 2-clone group
- `id_test.go` - Has 10-clone and 2-clone groups
- `id_fuzz_test.go` - Has 2-clone groups
- `testutil_test.go` - Has 5-clone, 4-clone, 2-clone groups (root package helpers)
- `internal/detectors/detectors_test.go` - Has 11-clone, 6-clone, 3-clone groups

### Production Code Clones (Cannot Easily Eliminate)

- **22-clone**: `func(ctx context.Context) ([]finding.Finding, error)` - Detector interface signature spans production code (`pipeline.go`, `retry.go`, `internal/detectors`). Cannot eliminate without fundamental interface redesign.
- **4-clone**: `severity.go` String() method - Inherent to enum pattern
- **4-clone**: `position.go` range checks - Inherent to Position type methods
- **3-clone**: `pipeline/partial.go` error handling - Production code patterns
- **2-clone**: `pipeline/pipeline.go` and `pipeline/verify.go` - Production code
- **2-clone**: `internal/detectors/govet.go` and `staticcheck.go` - Inherent to detector pattern

---

## D) TOTALLY FUCKED UP! 🔴

### 1. Build Failure

- **File**: `cmd/go-finding/integration_test.go`
- **Error**: `imported and not used: "github.com/stretchr/testify/assert"`
- **Cause**: Testify import added but assertions not yet applied
- **Fix**: Apply testify assertions to replace manual assertions in this file

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Immediate Fixes

1. **Fix cmd/go-finding/integration_test.go build failure** - Apply testify assertions
2. **Commit current changes** - 21 files modified, 3 commits ahead of origin

### Clone Elimination Strategy Refinements

3. **Handle Finding struct literal clones** - These are 11-12 token patterns; need to either:
   - Create helper functions that return Finding with specific fields set
   - Or accept some Finding literal clones as "acceptable" design patterns
4. **Handle production code clones** - The 22-clone Detector interface is inherent to Go's type system
5. **Handle position.go/severity.go clones** - These are inherent to the enum/struct patterns

### Process Improvements

6. **Batch commit strategy** - Commit working changes immediately to avoid losing progress
7. **Test-after-every-file rule** - Never leave file with unused imports
8. **Track progress per-clone-group** - Know exactly which groups remain and which files contain them

---

## F) TOP #25 THINGS TO GET DONE NEXT 🎯

1. **Fix cmd/go-finding/integration_test.go build failure** (BLOCKING)
2. **Commit all working changes** - 21 modified files across 4 packages
3. **Replace assertions in export_test.go** - 11-clone group for Finding literals
4. **Replace assertions in sarif_test.go** - 11-clone group for Finding literals
5. **Replace assertions in errors_test.go** - 10-clone group
6. **Replace assertions in coverage_test.go** - 10-clone group
7. **Replace assertions in merge_test.go** - 10-clone group
8. **Replace assertions in position_test.go** - 10-clone, 4-clone, 2-clone groups
9. **Replace assertions in position_extra_test.go** - 4/6/8-clone groups
10. **Replace assertions in report_test.go** - 6-clone, 4-clone groups
11. **Replace assertions in filter_test.go** - 6-clone group
12. **Replace assertions in json_test.go** - 4-clone, 3-clone groups
13. **Replace assertions in severity_test.go** - 4-clone group
14. **Replace assertions in diagnostic_test.go** - 3-clone group
15. **Replace assertions in lsp_test.go** - 3-clone group
16. **Replace assertions in equal_test.go** - 3-clone group
17. **Replace assertions in example_test.go** - Finding struct literal clones
18. **Replace assertions in example_basic_test.go** - Finding struct literal clones
19. **Replace assertions in example_cli_test.go** - Finding struct literal clones
20. **Replace assertions in fuzz_test.go** - 3-clone, 2-clone groups
21. **Replace assertions in bench_test.go** - 3-clone, 2-clone groups
22. **Replace assertions in finding_test.go** - 2-clone group
23. **Replace assertions in id_test.go** - 10-clone, 2-clone groups
24. **Replace assertions in id_fuzz_test.go** - 2-clone groups
25. **Replace assertions in testutil_test.go (root)** - 5-clone, 4-clone, 2-clone groups

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT 🔮

**Question**: How should we handle the **inherent clone groups** that cannot be eliminated without fundamental redesign?

### The Problem

Three clone groups are **inherent to the codebase design**:

1. **22-clone Detector interface signature**: `func(ctx context.Context) ([]finding.Finding, error)` appears 22 times in:
   - Production code: `pipeline/pipeline.go`, `pipeline/retry.go`
   - Test code: Multiple test files
   - Detector implementations: `internal/detectors/govet.go`, `internal/detectors/staticcheck.go`

2. **4-clone severity.go String() methods**: Inherent to Go enum pattern

3. **4-clone position.go range checks**: Inherent to Position type methods

### The Options

| Option                           | Approach                                                            | Pros                       | Cons                                      |
| -------------------------------- | ------------------------------------------------------------------- | -------------------------- | ----------------------------------------- |
| **A. Accept inherent clones**    | Mark them as acceptable, focus on eliminating test assertion clones | Realistic, achievable goal | Doesn't meet "ZERO clones" goal literally |
| **B. Redesign interface**        | Create abstraction to hide the signature                            | Would eliminate 22-clone   | Major refactoring, risky                  |
| **C. Suppress via tool config**  | Configure art-dupl to ignore certain patterns                       | Quick fix                  | Not a real solution                       |
| **D. Use type aliases/wrappers** | Create wrapper types that encapsulate the signature                 | Moderate effort            | Adds complexity                           |

### My Recommendation

**Option A (Accept inherent clones) with explicit documentation**:

- Focus on eliminating all **test assertion clones** (~50 groups)
- Accept production code clones (~20 groups) as architectural patterns
- Document the acceptable clone categories in a `CLONES.md` file
- This achieves ~70% reduction (76 → ~22 groups) which is practical and maintainable

**What do you want us to do?**

---

## Clone Groups by Size (Remaining 70)

| Size | Count | Example Patterns                        |
| ---- | ----- | --------------------------------------- |
| 22   | 1     | Detector interface signature (INHERENT) |
| 12   | 1     | Finding struct literals with Position   |
| 11   | 1     | Finding struct literals                 |
| 10   | 1     | `x.Error() != "y"` patterns             |
| 8    | 2     | Finding struct literals, position.go    |
| 6    | 4     | Various position/filter patterns        |
| 5    | 1     | testutil helpers                        |
| 4    | 6     | Various severity/position patterns      |
| 3    | 12    | Various test helper patterns            |
| 2    | 40    | Various small patterns                  |

---

## Files Modified (21 total, uncommitted)

```
cmd/go-finding/integration_test.go  ⚠️ BUILD FAILING
cmd/go-finding/main_test.go         ✅
export_test.go                     ⚠️ Not yet modified
filter_test.go                     ⚠️ Not yet modified
go.mod                             ✅
go.sum                             ✅
json_test.go                       ⚠️ Not yet modified
merge_test.go                      ⚠️ Not yet modified
pipeline/conflict_extra_test.go    ✅
pipeline/conflict_test.go          ✅
pipeline/fix_applier_test.go       ✅
pipeline/integration_test.go        ✅
pipeline/metrics_test.go           ✅
pipeline/partial_test.go           ✅
pipeline/pipeline_test.go          ✅
pipeline/retry_test.go             ✅
pipeline/verify_test.go            ✅
report_test.go                     ⚠️ Not yet modified
sarif_test.go                     ⚠️ Not yet modified
testutil_test.go                   ⚠️ Not yet modified
```

---

## Test Status

| Package                                                | Status  | Notes                         |
| ------------------------------------------------------ | ------- | ----------------------------- |
| `github.com/larsartmann/go-finding`                    | ✅ OK   |                               |
| `github.com/larsartmann/go-finding/cmd/go-finding`     | 🔴 FAIL | Build failure - unused import |
| `github.com/larsartmann/go-finding/internal/detectors` | ✅ OK   |                               |
| `github.com/larsartmann/go-finding/pipeline`           | ✅ OK   |                               |

---

## Git Status

```
On branch master
Your branch is ahead of 'origin/master' by 3 commits.
Changes not staged for commit:
  modified:   21 files (as listed above)
```

---

_Generated: 2026-04-26 19:06 UTC_
