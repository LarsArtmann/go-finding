# Status Report — 2026-06-15 Session 14 (Cont.)

## Executive Summary

Fixed the critical `nix flake check` failure, extracted shared `gotoken` utility
to eliminate code duplication, and benchmarked GoAST vs SubstringProvider.
Discovered and fixed a GoAST performance bottleneck (6× speedup via pointer-identity cache).

---

## A) FULLY DONE

### 1. Treefmt Config Mismatch Fixed

- **Root cause:** treefmt-nix golines defaults to `maxLength=100`; golangci-lint uses `max-len=120`
- **Fix:** Added `maxLength = 120` to `flake.nix` treefmt config
- **Result:** `nix flake check` now passes cleanly (was broken for weeks)
- **Commit:** `c316d10`

### 2. Shared gotoken Utility Extracted

- **Problem:** `pipeline/goast/provider.go` and `analysis/analysis.go` both independently implemented line/column→offset resolution using `go/token`
- **Solution:** Created `internal/gotoken/` package with 5 shared functions:
  - `LineColToOffset` — line/col → byte offset
  - `LineColToPos` — line/col → token.Pos
  - `FindFileByName` — filename lookup in FileSet
  - `FindInnermostNode` — deepest AST node containing byte offset
  - `NodeByteRange` — [start, end) byte offsets of AST node
- **Impact:** Eliminated ~35 lines of duplicated code; created clear home for future Go AST utilities
- **Commit:** `5626374`

### 3. GoAST Pointer-Identity Cache + Benchmarks

- **Discovery:** GoAST provider hashed entire content on EVERY `parse()` call — N findings = N O(n) hashes
- **Fix:** Added `unsafe.Pointer` comparison as fast path (same backing array = skip hash)
- **Impact:** 6× speedup on 100 findings in 10K-line file (6.7ms → 1.1ms)
- **Benchmark finding:** SubstringProvider is 15× faster than GoAST for simple unique strings (76μs vs 1.1ms). GoAST's value is disambiguation ACCURACY, not raw speed.
- **Commit:** `24c7483`

---

## B) PARTIALLY DONE

None — all 3 completed items are fully implemented, tested, and lint-clean.

---

## C) NOT STARTED

| Task                             | Reason                                              |
| -------------------------------- | --------------------------------------------------- |
| Make GoAST default in CLI        | Deferred — design decision needs user input         |
| Add FuzzGoASTProvider            | Would be valuable but lower priority than DRY fixes |
| Update performance analysis HTML | Session 14 numbers can be added later               |
| Coverage badge in README         | Cosmetic                                            |

---

## D) TOTALLY FUCKED UP

Nothing this session. Previous session's `golangci-lint --fix` corruption was already fixed in `fcb633b`.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **GoAST `FindInnermostNode` still traverses full AST per finding** — Could cache node lookup by offset or use a sorted interval index. Current: O(AST size) per finding. Could be O(log n) with pre-built index.
2. **Benchmark helpers duplicated** — `findOldOccurrences`, `offsetToLineNumber`, `columnOfOffset`, `pickEvenly` exist in both `pipeline/fix_engine_bench_test.go` and `pipeline/goast/provider_bench_test.go`. Should extract to shared bench test helper.
3. **`unsafe.Pointer` for cache identity** — Works correctly but is technically unsafe. Alternative: pass a content ID through the FixEngine. Low priority — the usage is contained and correct.

### Process

4. **`nix flake check` was broken for weeks** — The treefmt config mismatch was a pre-existing issue that blocked CI and pre-commit hooks. Should have been caught and fixed immediately.
5. **No `nix flake check` in pre-commit** — The BuildFlow hook runs `nix-flake-check` but it was failing. Need to verify the hook actually enforces this.

---

## F) TOP 25 THINGS TO DO NEXT

### High Impact (Architecture)

1. **Cache AST node lookup by offset** — Pre-build sorted interval index during parse; O(log n) per finding instead of O(AST size)
2. **Make GoAST default in CLI for .go files** — Design decision: should users opt-in or get it by default?
3. **Extract shared bench test helpers** — DRY between fix_engine_bench_test.go and goast/provider_bench_test.go
4. **Profile `FindInnermostNode` overhead** — Is `ast.Inspect` the bottleneck? Consider `ast.Walk` or direct field access
5. **Add GoAST fuzz target** — Property-based test for AST disambiguation correctness

### Medium Impact (Features)

6. **SARIF schema validation test** — Vendor lightweight validator
7. **Profile BenchmarkToSARIF** — Understand 1510 allocs/100 findings
8. **Add coverage badge to README** — Track coverage trends
9. **CLI integration test for `-fix-provider`** — End-to-end test
10. **Add `go-ast` to CLI help text** — Document available provider names
11. **Update performance analysis HTML** — Session 14 numbers
12. **Add `go-error-family` dependency** — Pre-commit hook flagged this
13. **Fix AGENTS.md length** — 402 lines, hook wants ≤377
14. **Remove committed binaries** — `go-finding` and `result` binaries in repo
15. **Add `.gitignore` for build outputs** — Prevent future binary commits

### Polish

16. **DeduplicateBy.String() edge cases** — More thorough enum testing
17. **golines → treefmt-nix maxLength auto-sync** — Prevent future config drift
18. **arena allocation** — When Go proposal lands
19. **SubstringProvider Boyer-Moore** — Only if non-Go files need it
20. **Add more fuzz targets** — FuzzProvider, FuzzCorrelate
21. **Benchmark findingsLocked vs direct access** — Quantify overhead
22. **Document `correlate.go` vs `merge.go` split** — ADR for file organization
23. **Add `IntervalIndex` to godocs** — Currently only tested via Correlate
24. **Consider `iter.Seq2` for MergeIter** — Yield (index, finding) pairs
25. **Fix go.mod sorting** — Pre-commit hook flagged unsorted require block

---

## G) TOP QUESTION

**Should we cache the AST node lookup results during parse time?**

Currently, `FindInnermostNode` calls `ast.Inspect` which traverses the entire AST tree for every finding. For 100 findings in a 10K-line file, that's 100 full-tree traversals.

Option A: Pre-build a sorted `[]struct{start, end int; node ast.Node}` during parse, then binary search per finding — O(log n) lookup.
Option B: Accept the traversal cost — `ast.Inspect` is fast enough for most files.
Option C: Lazy-build on first `FindInnermostNode` call, cache for subsequent calls.

The benchmark shows GoAST at 1.1ms for 100 findings. How much of that is `FindInnermostNode`? I can't tell without profiling. Should I profile first, or just implement Option A/C and measure?

---

_Generated: 2026-06-15 06:21_
_Session: 14 (Continued)_
_Commits: c316d10, 5626374, 24c7483_
_Tests: All passing (8 packages, 0 failures)_
_Lint: 0 golangci-lint issues_
_nix flake check: PASS_
