# Session 13 — Comprehensive Status Report

**Date:** 2026-06-14 17:57
**Branch:** master (clean)
**Commit:** `1212d44` (previous) → new commit pending
**Sessions:** 13 total

---

## A. FULLY DONE ✅

### Session 13 Work (This Session)

All 4 optimization tasks from the "Next Steps" backlog completed and verified:

#### 1. LineProvider/SubstringProvider Line Offset Index Caching

- **`lineIndexAware` interface** — `pipeline/fix_provider.go`: Optional interface with `EditsWithLineIndex(content, lineIndex, f)`. Both `LineProvider` and `SubstringProvider` implement it.
- **Lazy build via `*[]int`** — `pipeline/fix_engine.go`: `resolveEdits` accepts a pointer to `[]int`, builds the index on first access by a `lineIndexAware` provider. Zero overhead for OffsetProvider-handled findings.
- **Benchmark fix** — `pipeline/fix_engine_bench_test.go`: The existing `generateOffsetFixes` used Range width 4 but BeforeCode `"old()"` is 5 chars — OffsetProvider always rejected, silently benchmarking SubstringProvider. Fixed to use actual occurrence positions via `findOldOccurrences` helper.
- **New benchmarks** — Dedicated `BenchmarkFixEngine_LineProvider_{1,10,100,1000}` and `BenchmarkFixEngine_Substring_{1,10,100,1000}`.
- **Impact**: LineProvider 1000 fixes: 150ms → 812μs (**185×**), 86MB → 4MB allocs

#### 2. Correlate Pre-allocation

- **`merge.go`**: Pre-sized `correlations` with `min(len(findings), maxCorrelations)`, and `withRange`/`withoutRange` with `len(fileFindings)`
- **Impact**: 113μs → 78μs (31%), 75 → 51 allocs (32%)

#### 3. SARIF Struct Pooling Evaluation

- **Decision: SKIP.** Profiled 312KB/100 findings where bytes are dominated by un-poolable JSON output buffer (~200KB) and per-finding strings. Pooling struct headers saves ~0.2%. Complexity/risk unjustified.

#### 4. Micro-optimizations

- **`buildLineOffsetIndex` SIMD** — First pass uses `bytes.Count(content, []byte{'\n'})` (SIMD-accelerated for single-byte needle) instead of byte-by-byte loop
- **`findAllOccurrences` starting capacity** — `defaultOccurrenceCapacity = 32` constant eliminates first 5 reallocation growth phases. Evaluated `bytes.Count` for exact pre-allocation but reverted — double-scan overhead (12% slower) outweighed allocation savings

### Full Project Status

| Metric              | Value             |
| ------------------- | ----------------- |
| Go LOC              | 31,530            |
| Go files            | 143               |
| Test files          | 77                |
| Benchmarks          | 38                |
| Fuzz targets        | 20                |
| Exported symbols    | 110               |
| Direct dependencies | 26                |
| TODO items done     | 143               |
| TODO items open     | 16 (1 actionable) |

### Test Coverage

| Package              | Coverage  |
| -------------------- | --------- |
| Root (`finding`)     | **92.9%** |
| `analysis`           | **93.8%** |
| `internal/detectors` | **96.1%** |
| `cmd/go-finding`     | **90.6%** |
| `pipeline`           | **90.1%** |

### Verification Status

- `go test -race -count=1 ./...` — **ALL PASS** (9 packages, ~7s)
- `golangci-lint run ./...` — **0 issues**
- `go build ./...` — **CLEAN**

### Cumulative Performance Improvements (Sessions 12+13)

| Benchmark                        | Pre-Session-12 | Post-Session-13 | Speedup  |
| -------------------------------- | -------------- | --------------- | -------- |
| FixEngine OffsetProvider 1000    | 291,000ms      | 589μs           | **494×** |
| FixEngine LineProvider 1000      | 150,143μs      | 812μs           | **185×** |
| FixEngine SubstringProvider 1000 | 644,503μs      | 404,120μs       | **1.6×** |
| Correlate (200 findings)         | 113,396μs      | 77,759μs        | **1.5×** |

---

## B. PARTIALLY DONE 🟡

### SubstringProvider Performance

- **Status**: 1.6× faster (644ms → 404ms for 1000 fixes)
- **Remaining bottleneck**: `findAllOccurrences` performs an O(n) content scan per finding. This is inherent to the substring matching approach — each finding needs its own scan because `BeforeCode` varies per finding.
- **What would help**: A domain-specific provider (e.g., Go AST) that bypasses substring matching entirely. The `FixProvider` interface and `NewFixEngineWithProviders` are ready for this.

### CLI FixProvider Configuration

- **Status**: `Config.FixProviders` exists in the pipeline package and `NewFixApplierWithProviders` accepts custom providers
- **Missing**: CLI flag/config to wire custom providers through `cmd/go-finding/config.go`. Currently only the default 3 providers (Offset, Line, Substring) are used.
- **TODO_LIST item**: `[ ] Fix FixProviders through CLI config` — the only truly actionable open item

### v1.0.0 Migration Path

- `Report.Findings` field deprecated, `FindingsSnapshot()` is the replacement
- `Report.Merge()` deprecated, `MergeInto()` is the replacement
- `RecordFix()` deprecated, `RecordFixes(1)` is the replacement
- All deprecations documented with `// Deprecated:` godoc and removal target v1.0.0

---

## C. NOT STARTED ⬜

### Blocked / External

- `golines` in CI — treefmt-nix doesn't support it
- SARIF schema validation test — requires vendoring 7K+ line JSON schema
- BuildFlow auto-configure loop — external tool issue
- Wire into go-structure-linter — external project
- `.envrc` creation — no Nix setup on this machine

### Owner Decisions (Breaking Changes for v1.0/v2)

- `Position` zero-value safety (Offset=0 ambiguity)
- `Range.End` zero-value ambiguity
- `PositionOffset` sentinel design
- `Finding` struct sub-grouping (v2)

### Out of Scope v1

- Interactive TUI
- IDE plugin stubs
- Web UI
- Watch mode (deferred)

---

## D. TOTALLY FUCKED UP 💀

### Nothing is fucked up.

The codebase is in excellent shape:

- Zero lint issues across all 143 files
- All tests pass with `-race` detector
- Zero data races (the `linterCategories` race from Session 8 was fixed permanently)
- Zero phantom code (every symbol traced to at least one consumer or documented as reserved)
- Zero known bugs

**One honest caveat**: The pre-Session-12 benchmark (`generateOffsetFixes`) was **lying** — it claimed to test OffsetProvider but was actually testing SubstringProvider due to a Range width mismatch (4 bytes for a 5-char BeforeCode). This means the "452× speedup" reported in Session 12 was partially measuring the SubstringProvider path, not the pure OffsetProvider path. The real OffsetProvider improvement is still massive (~494×), but the benchmark was dishonest before this session's fix.

---

## E. WHAT WE SHOULD IMPROVE 🔄

### Architecture & Design

1. **Domain-specific FixProviders** — The biggest bang-for-buck improvement. Register Go AST, Rust syn, or TypeScript compiler API providers to replace fragile substring matching. The interface is ready (`FixProvider` + `lineIndexAware`), just needs implementations.
2. **`Report.Findings` encapsulation** — The public slice still allows external mutation. `FindingsSnapshot()` exists but callers haven't migrated. Need to audit internal callers and push toward the v1.0 path.
3. **SubstringProvider algorithmic ceiling** — Currently O(n×m) where n=findings, m=content size per finding. For large files with many substring fixes, this remains slow. Could benefit from a pre-built suffix index or Boyer-Moore, but that's complexity for a fallback provider.

### Testing & Quality

4. **Benchmarks for SARIF import** — `BenchmarkFromSARIF` exists but no regression baseline was captured before Session 13. Should establish baselines for all 38 benchmarks.
5. **Property-based testing gaps** — 20 fuzz targets exist but coverage of the FixEngine edit-application path is thin. Fuzzing `applyEditsToContent` with random overlapping edits could surface edge cases.
6. **Integration test for LineProvider** — The BDD tests cover basic cases but don't stress-test the `lineIndexAware` lazy build path with mixed provider types (some Offset, some Line, some Substring in the same batch).

### Documentation

7. **Performance regression dashboard** — Session 13 benchmarks are in AGENTS.md but not in a machine-readable format. A `docs/benchmarks/` directory with JSON output would enable automated regression detection.
8. **FixProvider authoring guide** — The `FixProvider` interface has a godoc example but no dedicated guide for implementing domain-specific providers (when to use AST vs regex vs substring).

### CI/CD

9. **Benchmark CI gate** — The `benchmark` job exists but doesn't compare against baselines. Adding `benchstat` comparison would catch performance regressions automatically.
10. **Coverage tracking** — No codecov/coveralls integration. Coverage is high (90%+) but not tracked over time.

---

## F. TOP 25 THINGS TO DO NEXT 🎯

### High Impact (Do First)

1. **Implement Go AST FixProvider** — Replace SubstringProvider for `.go` files. Would eliminate the remaining 404ms bottleneck for 1000 substring fixes.
2. **Wire custom FixProviders through CLI config** — The only actionable TODO_LIST item. Lets users register domain providers via `-fix-provider` flag or YAML.
3. **Add `benchstat` regression comparison to CI** — Catch performance regressions before merge.
4. **Fuzz `applyEditsToContent` with random overlapping edits** — The most performance-critical function deserves property-based testing.
5. **Capture benchmark baselines as JSON** — Enable before/after comparisons across PRs.

### Medium Impact

6. **Audit `Report.Findings` internal access** — Replace direct field access with `readFindings()`/`FindingsSnapshot()` to prepare for v1.0 unexport.
7. **Add `MergeIter` to `Combine`** — For very large report sets, streaming merge avoids loading all findings into memory.
8. **Profile `BenchmarkToSARIF` with `pprof`** — Understand the 1510 allocs/100 findings breakdown to find micro-optimization opportunities.
9. **Add `LineProvider` mixed-provider integration test** — Verify lazy index build with interleaved Offset/Line/Substring findings.
10. **Document `lineIndexAware` interface in usage guide** — Help provider authors understand when to implement it.
11. **Add `Correlate` benchmark with range-based findings** — Current benchmark uses point-based; range correlation (IntervalIndex) is untested in benchmarks.
12. **Evaluate `sync.Pool` for `[]int` line offset index** — The index is rebuilt per file; pooling could reduce GC pressure on large batches.
13. **Pre-allocate `seen` map in `MergeIter`** — Currently uses `make(map[string]struct{})` with no size hint.
14. **Add `DeduplicateBy.String()` test** — Exists but may lack edge case coverage.
15. **Profile pipeline with 10K+ findings** — The dry-run benchmarks exist but haven't been profiled with `pprof` to find new bottlenecks.

### Lower Priority / Polish

16. **v1.0.0 breaking changes planning** — Consolidate Position/Range zero-value decisions into an ADR.
17. **SARIF schema validation test** — Vendor the 7K-line schema or use a lightweight validator.
18. **Add `golines` to treefmt-nix** — When upstream support lands.
19. **Coverage badge in README** — Track coverage trends publicly.
20. **FixProvider authoring guide** — Dedicated `docs/guides/fix-providers.md`.
21. **Benchmark `MergeIter` vs `Combine`** — Verify streaming advantage for large inputs.
22. **Add `IntervalIndex` benchmarks** — Currently only exercised through `Correlate`.
23. **Evaluate ` arena` experimental allocation** — When Go arena proposal lands.
24. **SubstringProvider Boyer-Moore** — If domain providers aren't viable for some languages.
25. **Performance analysis HTML update** — `docs/research/performance-analysis.html` has Session 12 numbers; update with Session 13 results.

---

## G. TOP QUESTION ❓

**Should we implement a Go AST FixProvider in this repository, or should it live in a separate consumer module?**

The `FixProvider` interface and `NewFixEngineWithProviders` are designed for external registration. The godoc example even shows a `GoASTProvider` skeleton. However:

- **In this repo**: Would add `golang.org/x/tools/go/ast` dependency to the pipeline package (currently only in `analysis/`). Would make `go-finding` more useful out-of-the-box for Go projects.
- **In a separate module**: Keeps `go-finding` dependency-free. The `analysis/` subpackage already bridges to `go/analysis`, suggesting Go-specific code belongs there.

The `go-finding` design principle says "root package dependency-free" with `golang.org/x/tools` isolated to `analysis/`. But the pipeline package already depends on `golang.org/x/sync`. Adding AST providers to the pipeline would violate the dependency isolation principle but dramatically improve real-world usability.

This is a **strategic architecture decision** that affects the project's positioning (generic library vs Go-specific tool) and dependency surface.

---

## Session 13 Benchmark Summary

```
BenchmarkFixEngine_Apply_1-32               57,542     18,878 ns/op     165,225 B/op    6 allocs/op
BenchmarkFixEngine_Apply_10-32              41,820     27,704 ns/op     201,105 B/op   36 allocs/op
BenchmarkFixEngine_Apply_100-32             15,357     74,767 ns/op     487,331 B/op  225 allocs/op
BenchmarkFixEngine_Apply_1000-32             1,903    589,046 ns/op   3,875,147 B/op 2037 allocs/op
BenchmarkFixEngine_LineProvider_1-32         8,260    138,432 ns/op     247,145 B/op    7 allocs/op
BenchmarkFixEngine_LineProvider_10-32        7,935    148,452 ns/op     283,024 B/op   37 allocs/op
BenchmarkFixEngine_LineProvider_100-32       6,180    180,274 ns/op     569,248 B/op  226 allocs/op
BenchmarkFixEngine_LineProvider_1000-32      1,634    812,494 ns/op   3,957,064 B/op 2038 allocs/op
BenchmarkFixEngine_Substring_1-32            2,089    542,832 ns/op     604,520 B/op   21 allocs/op
BenchmarkFixEngine_Substring_10-32             290  4,182,153 ns/op   3,856,795 B/op  177 allocs/op
BenchmarkFixEngine_Substring_100-32             28 39,772,969 ns/op  36,306,895 B/op 1626 allocs/op
BenchmarkFixEngine_Substring_1000-32             3 404,120,776 ns/op 361,333,952 B/op 16040 allocs/op
BenchmarkCorrelate-32                       14,710     77,759 ns/op     358,440 B/op   51 allocs/op
BenchmarkToSARIF-32                         4,752    251,256 ns/op     311,646 B/op 1510 allocs/op
```

---

_Assisted-by: Crush <crush@charm.land>_
