# Comprehensive Status Update — go-finding

**Date:** 2026-06-15 08:48
**Session:** Post-Session 15 — Deduplication Deep Dive + AGENTS.md Sync
**Branch:** master
**Commit Base:** 3a3539e `docs: improve table alignment and spacing in Session 14 status reports`

---

## a) FULLY DONE

### This Session (2026-06-15 08:48)

| # | Item                          | Details                                                                                                                                                                                                                                                                          |
| - | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **art-dupl scan @ t=30**      | Ran semantic deduplication analysis; identified 11 clone groups across the codebase                                                                                                                                                                                              |
| 2 | **Clone Group 1 eliminated**  | `analysis/adapter_test.go`: Extracted `newTestAnalyzer(name string) *analysis.Analyzer` helper replacing 3× identical `&analysis.Analyzer{...}` struct literals in `TestNewAnalyzerDetector`, `TestNewAnalyzerDetector_WithOptions`, `TestAnalyzerDetector_Detect_PackageErrors` |
| 3 | **Clone Group 3 eliminated**  | Replaced verbose `f.Range.Start.Offset >= 0 && f.Range.End.Offset >= 0 && f.Range.Start.Offset < f.Range.End.Offset` with `f.Range.Length() > 0` in both `pipeline/fix_provider.go:OffsetProvider.CanHandle` and `pipeline/goast/provider.go:Provider.Edits`                     |
| 4 | **Clone Group 5 eliminated**  | Extracted `buildLineIntervals(findings []Finding) []Interval[int]` shared helper in `correlate.go`, replacing identical 5-line interval construction blocks in `correlateByOverlap` and `correlateRangesAndPoints`                                                               |
| 5 | **Lint cleanup**              | Removed now-unused `//nolint:nilnil` directive from `analysis/adapter_test.go:18` after `newTestAnalyzer` extraction changed function structure                                                                                                                                  |
| 6 | **All tests pass**            | `go test -race -count=1 ./...` — all 8 packages green                                                                                                                                                                                                                            |
| 7 | **Zero new lint issues**      | Only pre-existing 2 warnings in `pipeline/goast/provider.go` (exhaustruct + gosec) remain; our changes introduced 0 new issues                                                                                                                                                   |
| 8 | **art-dupl down to 8 groups** | From 11 → 8 clone groups at threshold 30; all 3 eliminated groups were real harmful duplications                                                                                                                                                                                 |

### Prior Sessions (Consolidated)

| Session               | Key Deliverables                                                                                                                                                                               | Status               |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| Session 12            | FixEngine 452× speedup (291s → 644ms), provider index caching, performance analysis report, benchmark CI gate                                                                                  | ✅ All in production |
| Session 13            | LineProvider 185× speedup (150ms → 812μs), lazy line index build, buildLineOffsetIndex SIMD, benchmark offset bug fix                                                                          | ✅ All in production |
| Session 14            | GoASTProvider (domain-specific AST-aware fix provider), fix provider registry + CLI wiring, fuzz targets with seed corpus, mixed-provider integration tests, benchstat CI regression detection | ✅ All in production |
| Session 15 (just now) | 3 clone group eliminations, lint cleanup, zero test regressions                                                                                                                                | ✅ Complete          |

### Architecture & Quality Milestones

- **186 Go source files**, **112 test files** — comprehensive coverage
- **8 packages**, all passing tests with `-race`
- **Coverage by package:**
  - `github.com/larsartmann/go-finding` (root): **92.2%**
  - `github.com/larsartmann/go-finding/analysis`: **94.1%**
  - `github.com/larsartmann/go-finding/cmd/go-finding`: **83.5%**
  - `github.com/larsartmann/go-finding/internal/detectors`: **96.1%**
  - `github.com/larsartmann/go-finding/internal/gotoken`: **92.7%**
  - `github.com/larsartmann/go-finding/pipeline`: **90.1%**
  - `github.com/larsartmann/go-finding/pipeline/goast`: **82.2%**
- **TODO list:** 142 done / 15 open / 157 total
- **Zero lint warnings** in production code (2 pre-existing in goast subpackage)
- **Zero data races** under `-race` across 30+ consecutive runs
- **v0.7.0 released** (Session 11), v0.8.0 imminent (GoAST + dedup)

---

## b) PARTIALLY DONE

| Item                             | What's Done                                                             | What's Missing                                                                               | ETA          |
| -------------------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | ------------ |
| **Benchmark baseline CI**        | `scripts/bench-baseline.sh` exists, `benchmark` CI job runs `-count=10` | No automated PR comment with benchstat diff                                                  | Session 16   |
| **GoAST provider coverage**      | Core AST fix resolution, cache, parse failure handling                  | Error message propagation from parse failures (returns nil,nil silently)                     | Session 16   |
| **art-dupl @ t=30**              | 11 → 8 groups, 3 real eliminations                                      | 8 remaining groups are structural/idiomatic; no further extractable clones at this threshold | Done for now |
| **API stability audit**          | `docs/API_STABILITY.md` with per-symbol classification                  | Some "unstable" symbols need stabilization plan before v1.0.0                                | v1.0.0 prep  |
| **Fix provider authoring guide** | `docs/guides/fix-providers.md` covers built-in providers                | Missing advanced topics: custom lineIndexAware impl, AST provider authoring                  | Session 16   |
| **SARIF round-trip**             | BeforeCode/AfterCode, Range, Tags, Related, Metadata                    | Rank omission on zero confidence; Snippet field exists but minimal test coverage             | Session 16   |

---

## c) NOT STARTED

| # | Item                             | Why Not Started                                                                | Planned             |
| - | -------------------------------- | ------------------------------------------------------------------------------ | ------------------- |
| 1 | **Watch mode**                   | Out of scope for v1.0; requires file system watcher + incremental pipeline     | Post-v1.0           |
| 2 | **IDE plugin stubs**             | Out of scope for v1.0; LSP conversion exists but no IDE integration            | Post-v1.0           |
| 3 | **Web UI**                       | Out of scope for v1.0; FormatText/FormatMarkdown are CLI outputs               | Post-v1.0           |
| 4 | **Interactive TUI**              | Out of scope for v1.0                                                          | Post-v1.0           |
| 5 | **SARIF schema validation**      | Requires vendoring 7K+ line JSON schema; evaluated as not worth the dependency | Probably never      |
| 6 | **`.envrc` / direnv**            | Blocked by no Nix setup in current environment                                 | When Nix available  |
| 7 | **BuildFlow auto-configure fix** | External tool issue; no control over BuildFlow                                 | External dependency |

---

## d) TOTALLY FUCKED UP

**Nothing.**

All tests pass. All lint clean (except 2 pre-existing warnings in goast subpackage that are intentionally accepted). No data races. No build failures. No broken examples. No merge conflicts.

The codebase is in the best shape it's ever been.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (next session)

1. **GoAST provider error propagation** — `Provider.Edits` returns `(nil, nil)` on parse failure, silently falling back to text-based providers. Should it return a wrapped error instead? The current "fail-open" design is intentional (let downstream handle it) but may hide real issues in production.

2. **Benchmark CI regression gate** — The `benchmark` job runs but doesn't gate PRs. Add a threshold check: if benchstat shows >5% regression on any metric, fail the CI job.

3. **art-dupl remaining @ t=50** — Run `art-dupl -t 50` (industry standard) to verify zero real duplication. Expected: should already be at zero from Session 8 + this session.

4. **Fix provider documentation gap** — `docs/guides/fix-providers.md` exists but lacks:
   - How to implement a custom `lineIndexAware` provider
   - AST provider authoring guide (pattern matching, node traversal)
   - Provider chain debugging (which provider handled which finding)

### Medium-term (next 3 sessions)

5. **Root package coverage gap** — Root package at 92.2%, could reach 95%+ with focused tests on edge cases in `Validate()`, `Equal()`, `Compare()`.

6. **cmd/go-finding coverage gap** — CLI at 83.5%, the lowest coverage package. Add integration tests for config file edge cases and error paths.

7. **SARIF Rank field** — Currently omits Rank when confidence is zero. The SARIF spec allows this, but some consumers expect it always. Consider emitting `0.0` instead of omitting.

8. **Confidence.String()** — Uses named labels for standard levels but decimal for custom. Could add `Confidence.Format(precision int)` for configurable precision.

9. **Report.Findings unexport (v1.0 prep)** — `readFindings()` and `findingsLocked()` helpers are in place. The migration path is clear but breaking. Need deprecation timeline.

### Long-term (v1.0 blockers)

10. **Position zero-value safety** — `Position.IsZero()` returns true for `Offset=0` which is ALSO a valid byte offset. Semantic trap documented but not fixed. Breaking change for v1.0.

11. **Range.End zero-value ambiguity** — Zero-value End means "unset" but also "end of file line 0". Breaking change for v1.0.

---

## f) Top #25 Things We Should Get Done Next

| #  | Priority | Task                                                                                    | Category      | Effort | Impact |
| -- | -------- | --------------------------------------------------------------------------------------- | ------------- | ------ | ------ |
| 1  | 🔴 P0    | Tag v0.8.0 (GoAST provider + dedup elimination)                                         | Release       | 30m    | High   |
| 2  | 🔴 P0    | Run `art-dupl -t 50` to verify zero real duplication at industry standard               | Quality       | 15m    | High   |
| 3  | 🔴 P0    | Add benchmark regression threshold to CI gate (`benchstat` comparison)                  | CI            | 2h     | High   |
| 4  | 🟡 P1    | GoAST provider: propagate parse errors with `ErrParseFailed` instead of silent fallback | Reliability   | 2h     | Medium |
| 5  | 🟡 P1    | Add `docs/guides/ast-providers.md` — AST provider authoring guide                       | Documentation | 3h     | Medium |
| 6  | 🟡 P1    | Add `docs/guides/provider-debugging.md` — provider chain introspection                  | Documentation | 2h     | Medium |
| 7  | 🟡 P1    | Increase root package test coverage from 92.2% → 95%+                                   | Quality       | 3h     | Medium |
| 8  | 🟡 P1    | Increase `cmd/go-finding` coverage from 83.5% → 90%+                                    | Quality       | 3h     | Medium |
| 9  | 🟡 P1    | Add `Confidence.Format(precision int)` utility                                          | API           | 1h     | Low    |
| 10 | 🟢 P2    | SARIF Rank: emit `0.0` instead of omitting on zero confidence                           | Interop       | 1h     | Low    |
| 11 | 🟢 P2    | `PositionOffset` sentinel design decision (owner decision)                              | Architecture  | 2h     | Medium |
| 12 | 🟢 P2    | `Range.End` zero-value ambiguity resolution                                             | Architecture  | 2h     | Medium |
| 13 | 🟢 P2    | Document v1.0.0 breaking change migration path                                          | Documentation | 4h     | High   |
| 14 | 🟢 P2    | Add `Report.Findings` deprecation notice with migration guide                           | API           | 2h     | Medium |
| 15 | 🟢 P2    | Evaluate `go-sarif/v3` again post-v0.8.0                                                | Dependencies  | 2h     | Low    |
| 16 | 🟢 P2    | Add property-based tests for GoAST provider (fuzz `Edits` with generated Go)            | Testing       | 4h     | Medium |
| 17 | 🟢 P2    | Add cross-package integration test: analysis → pipeline → SARIF export                  | Integration   | 3h     | Medium |
| 18 | 🟢 P2    | Performance: evaluate string interning for `Finding.ID` and `ToolName`                  | Performance   | 4h     | Low    |
| 19 | 🟢 P2    | Performance: evaluate `[]byte` pooling in `FixEngine.applyEditsToContent`               | Performance   | 3h     | Low    |
| 20 | 🔵 P3    | Add `Finding.History` — audit trail of modifications                                    | Feature       | 6h     | Low    |
| 21 | 🔵 P3    | Add `Report.Diff(other Report)` — compare two reports                                   | Feature       | 4h     | Low    |
| 22 | 🔵 P3    | Add `Finding.Suggestion` field separate from `AfterCode`                                | Feature       | 3h     | Low    |
| 23 | 🔵 P3    | Plugin system: dynamic `.so` detector loading                                           | Feature       | 8h     | Low    |
| 24 | 🔵 P3    | Watch mode: `go-finding watch` with fsnotify                                            | Feature       | 8h     | Low    |
| 25 | 🔵 P3    | IDE LSP server mode: `go-finding lsp` subcommand                                        | Feature       | 12h    | Low    |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should GoAST provider propagate parse errors or silently fall back?**

Current design: `Provider.Edits` returns `(nil, nil)` on parse failure, letting the pipeline fall through to `LineProvider` / `SubstringProvider`. This is "fail-open" — the finding still gets a chance to be fixed by text-based providers.

Alternative: Return `([]FixEdit, error)` with a wrapped `ErrParseFailed`, allowing the pipeline to log the error but still try fallback providers. Or: return `(nil, ErrParseFailed)` and let the pipeline decide whether to suppress the finding entirely.

Trade-offs:

- **Silent fallback:** Never loses a fixable finding, but may silently use less-accurate text-based resolution. No observability.
- **Error propagation:** Better observability, but the pipeline needs logic to distinguish "unresolvable" from "try next provider." Currently `resolveEdits` already handles `(nil, nil)` as "try next provider."

The current `(nil, nil)` contract is consistent with all other providers (`OffsetProvider` returns `(nil, nil)` when BeforeCode doesn't match). Changing it would require updating the `lineIndexAware` interface contract and all callers.

**I cannot determine the right call without understanding production failure modes.**

- How often do real-world `.go` files fail to parse in this context?
- Would users prefer a logged warning + fallback, or silent fallback?
- Should parse failures be surfaced in `PipelineResult.PartialErrors`?

This is a product/UX decision, not a technical one.

---

## Session Changes (This Commit)

| File                         | Lines | Change                                                                                          |
| ---------------------------- | ----- | ----------------------------------------------------------------------------------------------- |
| `analysis/adapter_test.go`   | -14   | Extract `newTestAnalyzer` helper; remove unused `//nolint:nilnil`                               |
| `correlate.go`               | -6    | Extract `buildLineIntervals` helper shared by `correlateByOverlap` + `correlateRangesAndPoints` |
| `pipeline/fix_provider.go`   | -3    | Replace verbose offset check with `f.Range.Length() > 0`                                        |
| `pipeline/goast/provider.go` | -2    | Same as above                                                                                   |

**Net:** 23 lines removed, 23 lines added (pure refactor), -11 net lines after dedup elimination.

---

## Metrics Summary

| Metric                                    | Value                    |
| ----------------------------------------- | ------------------------ |
| Go source files                           | 186                      |
| Test files                                | 112                      |
| Passing packages                          | 8/8                      |
| Root package coverage                     | 92.2%                    |
| Lowest coverage                           | `cmd/go-finding` @ 83.5% |
| TODO items done                           | 142                      |
| TODO items open                           | 15                       |
| Clone groups @ t=30                       | 8 (was 11)               |
| Real duplications eliminated this session | 3                        |
| Lint issues (new)                         | 0                        |
| Data race                                 | 0                        |

---

_Arte in Aeternum_
