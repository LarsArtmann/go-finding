# Comprehensive Status Update — go-finding

**Date:** 2026-06-17 11:21
**Session:** Post-Dedup Deep-Dive + Self-Review Fixes
**Branch:** master (pushed to origin)
**Version:** v0.8.0
**Commit Base:** `2bd2a22` `test: add comprehensive unit tests for internal/benchutil package`

---

## a) FULLY DONE

### This Session (2026-06-17 11:21)

| # | Item                                        | Details                                                                                                                                               |
| - | ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **art-dupl scan @ t=20**                    | Semantic deduplication analysis: 46 clone groups / 117 tokens / 109 clones identified                                                                 |
| 2 | **Clone Group #3 eliminated (8 tokens)**    | Extracted 4 verbatim-copied utility functions (`findOldOccurrences`, `offsetToLineNumber`, `columnOfOffset`, `pickEvenly`) into `internal/benchutil/` |
| 3 | **Clone Group #1 eliminated (7 tokens)**    | Extracted `beforeAfterFinding(file, line, col, before, after)` helper for 7 identical Finding struct literals in `pipeline/goast/provider_test.go`    |
| 4 | **Clone Group #2 eliminated (4 tokens)**    | Extracted `AssertPanics(t, msg, fn)` helper to `testutil_test.go` for 4 panic-test boilerplate sites                                                  |
| 5 | **Self-review fix: 2 more recover() sites** | Converted `severity_test.go` and `category_test.go` panic tests to `AssertPanics` — consolidating ALL 6 panic-test sites (I had missed 2 first round) |
| 6 | **Self-review fix: benchutil tests**        | Added comprehensive unit tests for the new `internal/benchutil` package — 4 table-driven functions + integration test (was 0 tests)                   |
| 7 | **Benchmark regression verified**           | Ran all bench targets — no regression from package extraction (helpers are pre-`ResetTimer` setup)                                                    |

**3 commits, all pushed:**

- `fc2a014` refactor: extract shared benchutil and AssertPanics helper
- `7c7503f` refactor: consolidate last 2 panic-test boilerplates into AssertPanics
- `2bd2a22` test: add comprehensive unit tests for internal/benchutil package

### Prior Sessions (Consolidated)

| Session               | Key Deliverables                                                                                                                                                                               | Status               |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| Session 12            | FixEngine 452× speedup (291s → 644ms), provider index caching, performance analysis report, benchmark CI gate                                                                                  | ✅ All in production |
| Session 13            | LineProvider 185× speedup (150ms → 812μs), lazy line index build, buildLineOffsetIndex SIMD, benchmark offset bug fix                                                                          | ✅ All in production |
| Session 14            | GoASTProvider (domain-specific AST-aware fix provider), fix provider registry + CLI wiring, fuzz targets with seed corpus, mixed-provider integration tests, benchstat CI regression detection | ✅ All in production |
| Session 15            | 3 clone group eliminations, lint cleanup, zero test regressions                                                                                                                                | ✅ All in production |
| Session 16 (this one) | 3 more clone group eliminations + 2 missed sites + benchutil tests                                                                                                                             | ✅ Complete          |

### Architecture & Quality Milestones

- **191 Go source files**, **117 test files** — comprehensive coverage
- **11 packages**, all passing tests with `-race`
- **Coverage by package:**
  - `github.com/larsartmann/go-finding` (root): **93.7%**
  - `github.com/larsartmann/go-finding/analysis`: **94.1%**
  - `github.com/larsartmann/go-finding/cmd/go-finding`: **91.2%**
  - `github.com/larsartmann/go-finding/internal/detectors`: **96.1%**
  - `github.com/larsartmann/go-finding/internal/gotoken`: **92.7%**
  - `github.com/larsartmann/go-finding/internal/benchutil`: **100.0%** ✨ NEW
  - `github.com/larsartmann/go-finding/pipeline`: **95.2%**
  - `github.com/larsartmann/go-finding/pipeline/goast`: **80.8%**
- **art-dupl @ t=20:** 43 clone groups / 98 tokens / 96 clones (down from 46 / 117 / 109)
- **Production tokens:** 11 (all accepted as Go idioms — function signatures, interface methods)
- **Zero lint warnings** in production code (golangci-lint: 0 issues)
- **Zero data races** under `-race`
- **v0.8.0 released** (2026-06-17)
- **Zero `recover()` boilerplate** in tests outside `AssertPanics` itself (was 7 occurrences)

---

## b) PARTIALLY DONE

| Item                            | Status        | Blocker / Next Step                                                                                                  |
| ------------------------------- | ------------- | -------------------------------------------------------------------------------------------------------------------- |
| **v1.0 Release**                | ~85% complete | Owner decisions on breaking changes (see Section g)                                                                  |
| **Deprecated API removal**      | 5 APIs tagged | Need removal before v1.0: `Report.Findings`, `Report.Merge()`, `OnStage`, `Metrics.RecordFix()`, `CountBySeverity()` |
| **GoASTProvider coverage**      | 80.8%         | Below 95% target — needs more edge case tests                                                                        |
| **TODO_LIST.md freshness**      | Stale (5/20)  | Still references "session 13" as last update — needs re-sync with current state                                      |
| **SARIF schema validation**     | Not done      | Blocked on 7K+ line schema vendoring decision                                                                        |
| **CLI FixProviders via config** | Broken        | `Fix`FixProviders`through CLI config` is open in TODO_LIST.md                                                        |

---

## c) NOT STARTED

| Item                                                | Priority | Notes                                            |
| --------------------------------------------------- | -------- | ------------------------------------------------ |
| Integration test with downstream consumer           | High     | BuildFlow / hierarchical-errors / branching-flow |
| `README.md` covers all exported types with examples | Medium   | Required for v1.0                                |
| CHANGELOG.md updated for v1.0                       | Medium   | Required for v1.0                                |
| API stability guarantee documented                  | Medium   | Required for v1.0                                |
| `io.WriterTo` for SARIF output                      | Low      | Nice-to-have per release criteria                |
| Structured logging (slog)                           | Low      | Nice-to-have per release criteria                |
| CLI `run()` testability                             | Low      | Nice-to-have per release criteria                |
| Watch mode                                          | Deferred | Out of scope v1                                  |
| IDE plugin stubs                                    | Deferred | Out of scope v1                                  |
| Web UI                                              | Deferred | Out of scope v1                                  |
| Interactive TUI                                     | Deferred | Out of scope v1                                  |
| `Finding` struct sub-grouping                       | Deferred | Breaking change, deferred to v2                  |

---

## d) TOTALLY FUCKED UP

| Issue                                                                                                                                                                                                                     | Impact | Mitigation                                                                                            |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ----------------------------------------------------------------------------------------------------- |
| **First-round dedup was incomplete** — I accepted 2 `recover()` sites as "idiom" when they were identical to the 4 I extracted. I trusted the tool's grouping instead of grepping ALL `recover()` patterns codebase-wide. | Low    | Fixed in second round: grepped all `recover()`, converted both. Now zero boilerplate.                 |
| **Created `internal/benchutil` with 0 tests** — introduced a new exported package without any test coverage. Violates project testing mandate.                                                                            | Low    | Fixed in second round: added 219-line test file with 4 table-driven functions + integration test.     |
| **First round had no benchmark regression verification** — extracted helpers across package boundary without verifying no perf regression.                                                                                | Low    | Fixed in second round: ran all bench targets, confirmed no regression (helpers are pre-`ResetTimer`). |

**Root cause of all three:** Rushed first pass without systematic verification. I trusted the tool's output literally instead of independently verifying my work.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (next session)

1. **GoASTProvider coverage gap (80.8%)** — Below the 95% project target. Needs targeted edge case tests: malformed AST nodes, cache invalidation paths, nil content paths.
2. **TODO_LIST.md is stale** — Last updated 2026-05-20 (session 13 reference). Doesn't reflect sessions 14-16 work. Needs regeneration.
3. **`NewReplacementFinding` constructor decision** — The `Finding{Position, BeforeCode, AfterCode}` pattern appears in tests, benchmarks, AND real pipeline code. Currently no public constructor captures this "text replacement" concept. Needs owner decision (see Section g).

### Medium-term (next 3 sessions)

4. **CLI `FixProviders` through config is broken** — Open in TODO_LIST.md. Users can't configure fix providers via YAML/JSON config.
5. **v1.0 deprecated API removal** — 5 APIs need removal. Each is a breaking change requiring migration guide updates.
6. **`Report.Findings` → `FindingsSnapshot()`** migration — The biggest deprecated API. All internal code already uses `readFindings()`, but the field is public.

### Long-term (v1.0 blockers)

7. **Owner decisions on zero-value ambiguity** — `Position.Offset=0` means both "byte 0" and "unset". `Range.End` zero-value ambiguity. These are blocking v1.0.
8. **FixStrategy "" vs "none"** — Two valid "no fix" states. Needs consolidation for v1.0.
9. **SARIF schema validation** — Blocked on vendoring 7K+ line JSON schema.

---

## f) Top #25 Things We Should Get Done Next

Sorted by **Impact × Customer Value ÷ Effort** (highest ROI first).

| #   | Task                                                              | Impact | Effort | Customer Value |
| --- | ----------------------------------------------------------------- | ------ | ------ | -------------- |
| 1   | **Decide `Position.Offset=0` sentinel** — unblocks v1.0           | 🔴🔴🔴 | S      | All consumers  |
| 2   | **Decide `Range.End` zero-value** — unblocks v1.0                 | 🔴🔴🔴 | S      | All consumers  |
| 3   | **Decide `FixStrategy ""` vs "none"** — unblocks v1.0             | 🔴🔴🔴 | S      | All consumers  |
| 4   | **Decide `NewReplacementFinding` constructor** — improves API     | 🔴🔴   | S      | All consumers  |
| 5   | **GoASTProvider coverage 80.8% → 95%** — meets project target     | 🔴🔴   | M      | Library users  |
| 6   | **Fix CLI `FixProviders` through config** — broken feature        | 🔴🔴   | M      | CLI users      |
| 6   | **Regenerate TODO_LIST.md** — reflects current state              | 🔴     | S      | Developers     |
| 7   | **Integration test with downstream consumer** — required for v1.0 | 🔴🔴   | L      | All consumers  |
| 8   | **Remove `Report.Findings` field** — biggest deprecated API       | 🔴🔴   | M      | All consumers  |
| 9   | **Remove `Report.Merge()`** — deprecated API                      | 🔴     | S      | All consumers  |
| 10  | **Remove `OnStage`** — deprecated API                             | 🔴     | S      | All consumers  |
| 11  | **Remove `Metrics.RecordFix()`** — deprecated API                 | 🔴     | S      | All consumers  |
| 12  | **Remove `CountBySeverity()` free func** — deprecated API         | 🔴     | S      | All consumers  |
| 13  | **README.md covers all exported types** — required for v1.0       | 🔴🔴   | M      | New users      |
| 14  | **CHANGELOG.md updated for v1.0** — required for v1.0             | 🔴     | S      | All consumers  |
| 15  | **API stability guarantee documented** — required for v1.0        | 🔴     | S      | All consumers  |
| .go | **Doc examples for all exported types** — required for v1.0       | 🔴     | M      | New users      |
| 17  | **`io.WriterTo` for SARIF output** — nice-to-have                 | 🟡     | S      | Performance    |
| 18  | **Structured logging (slog)** — nice-to-have                      | 🟡     | M      | Observability  |
| 19  | **CLI `run()` testability** — nice-to-have                        | 🟡     | S      | CLI users      |
| 20  | **SARIF schema validation test** — nice-to-have                   | 🟡     | L      | Library users  |
| 21  | **`examples/pipeline` coverage 0%** — example with no tests       | 🟡     | S      | Developers     |
| 2   | **Doc examples for IntervalIndex, MergeIter, etc.** — gap in docs | 🟡     | S      | New users      |
| 23  | **Wire go-structure-linter** — external, deferred                 | 🟢     | L      | Developers     |
| 24  | **Add `golines` to CI** — blocked on treefmt-nix                  | 🟢     | L      | Developers     |
| 25  | **`.envrc` / Nix direnv** — blocked on no Nix setup               | 🟢     | L      | Developers     |

### High-impact code architecture improvements

| Improvement                                      | Rationale                                                                                    |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| **NewReplacementFinding constructor**            | Gives "text replacement" concept first-class name in domain. Replaces test-only helper.      |
| **Properties map[string]any alongside Metadata** | Currently `Metadata map[string]string`. SARIF property bags can have typed values. Consider. |
| **Finding struct sub-grouping**                  | 16 fields in one struct. Sub-grouping would improve clarity but deferred to v2 (breaking).   |
| **Stronger FixStrategy type**                    | Currently `type FixStrategy string`. Could be struct with methods to eliminate "" vs "none". |
| **Position zero-value safety**                   | Offset=0 ambiguity is a known design flaw. Could use `*int` or sentinel type. Breaking.      |

---

## g) Top #1 Question I Cannot Figure Out Myself

### The `Position.Offset=0` Ambiguity — Needs Your Decision

**Current state:**

```go
type Position struct {
    File   string
    Line   int
    Column int
    Offset int  // ← PROBLEM: 0 means both "byte 0" and "unset"
}
```

**The problem:** `Position.Offset=0` is ambiguous. It means both:

- "Byte offset 0 in the file" (the very first byte)
- "Offset not set" (default zero-value)

`HasOffset()` currently returns `true` for both, which is wrong for the "unset" case. This propagates into `Range.Length()` calculations, FixEngine edit resolution, and SARIF round-trip fidelity.

**Why I can't decide this myself:** There are three valid solutions, each with major tradeoffs:

| Option                           | Pro                           | Con                                                                               | Breaking? |
| -------------------------------- | ----------------------------- | --------------------------------------------------------------------------------- | --------- |
| **A. `Offset *int` (nil=unset)** | Zero-value safe, idiomatic Go | Breaking change. Every struct literal needs `&x`. APIs using `Offset int` change. | YES       |
| **B. Sentinel `-1` (unset)**     | Non-breaking for most code    | Less idiomatic. `-1` checks everywhere. `HasOffset()` fix only.                   | Partially |
| **C. Separate `HasOffset bool`** | Explicit, clear intent        | Adds field. Redundant state to keep in sync.                                      | YES       |

**This blocks v1.0** (per RELEASE_CRITERIA.md). Your call shapes the public API forever.

**My recommendation:** Option B (sentinel `-1`) as the pragmatic non-breaking path for v1.0, with Option A (`*int`) targeted for v2.0 when we can make breaking changes cleanly. But you may prefer to "do it right" now with Option A since v1.0 is the moment to lock the API.

---

## Session Changes (This Commit)

**3 commits, all pushed to origin/master:**

1. `fc2a014` — Extracted `internal/benchutil` package + `AssertPanics` helper + `beforeAfterFinding` helper
2. `7c7503f` — Consolidated 2 remaining `recover()` sites into `AssertPanics`
3. `2bd2a22` — Added comprehensive unit tests for `internal/benchutil`

**Files changed:**

| File                                    | Change                                |
| --------------------------------------- | ------------------------------------- |
| `internal/benchutil/benchutil.go`       | **NEW** — 4 exported functions        |
| `internal/benchutil/benchutil_test.go`  | **NEW** — 219 lines of tests          |
| `pipeline/fix_engine_bench_test.go`     | Removed 4 dup funcs, import benchutil |
| `pipeline/goast/provider_bench_test.go` | Removed 4 dup funcs, import bench     |
| `pipeline/goast/provider_test.go`       | Added `beforeAfterFinding` helper     |
| `testutil_test.go`                      | Added `AssertPanics` helper           |
| `adapter_test.go`                       | 3 panic sites → `AssertPanics`        |
| `finding_builder_test.go`               | 1 panic site → `AssertPanics`         |
| `severity_test.go`                      | 1 panic site → `AssertPanics`         |
| `category_test.go`                      | 1 panic site → `AssertPanics`         |

---

## Metrics Summary

| Metric                           | Before Session | After Session | Delta          |
| -------------------------------- | -------------- | ------------- | -------------- |
| art-dupl clone groups (t=20)     | 46             | 43            | **−3**         |
| art-dupl total clones            | 109            | 96            | **−13**        |
| art-dupl total tokens            | 117            | 98            | **−19 (−16%)** |
| art-dupl test tokens             | 98             | 87            | **−11**        |
| `recover()` boilerplate sites    | 7              | 0             | **−7**         |
| Test packages green (with -race) | 11             | 11            | 0 regressions  |
| golangci-lint issues             | 0              | 0             | clean          |
| Source files                     | 188            | 191           | +3             |
| Test files                       | 116            | 117           | +1             |
| New test coverage (benchutil)    | N/A            | 100.0%        | ✨             |

---

_Assisted-by: Crush <crush@charm.land>_
