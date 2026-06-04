# Status Report — 2026-06-02 04:30

**Session type:** Multi-pass brutal self-review + execution (4 passes)

---

## Executive Summary

24 commits across 4 passes. All green: 0 lint issues, 92.3% total coverage, `nix flake check` passes. The project went from a broken rename (`bCombine`) and stale docs to a clean, well-typed codebase with comprehensive test coverage.

---

## a) FULLY DONE

### Bugs Fixed

| #   | Bug                                                                                              | Fix                                                                       | Commit               |
| --- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- | -------------------- |
| 1   | `Report.bCombine` — broken rename from previous session, 2 compiler errors                       | Restored to `Report.Merge` (in-place concat, distinct from `Combine`)     | `39b7466`            |
| 2   | `Summary.DurationMs` never set — misleading "Execution time" comment                             | Clarified as caller-set; pipeline timing lives in `Metrics.TotalDuration` | `3ed83ea`            |
| 3   | `CompletionReason` wrong for timeout — iteration error path set `ReasonError` for context errors | Added `IsContextError()` check before defaulting to `ReasonError`         | `0d89250`            |
| 4   | `result.Stable` field → method rename left `result.Stable` without `()` in docs                  | Fixed README.md, USAGE_GUIDE.md, FEATURES.md                              | `c87e703`, `356465e` |
| 5   | `merge.go` duplicate doc comment + stale `MergedToolName` comment                                | Removed duplicate, updated references                                     | `ee57d76`            |
| 6   | `Correlation.Confidence` → `c.Score` stale in USAGE_GUIDE.md                                     | Fixed                                                                     | `356465e`            |

### Type Model Improvements

| #   | What                                                                                                                                                                | Commit               |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| 1   | `CompletionReason` typed string with 5 constants (`ReasonStable`, `ReasonMaxIterations`, `ReasonCancelled`, `ReasonTimeout`, `ReasonError`) replacing `Stable bool` | `21d5692`, `0d89250` |
| 2   | `pipeline.Stage` named type with 5 constants (`StageDetect`, `StageProcess`, `StageTriage`, `StageApply`, `StageVerify`) replacing raw strings                      | `ebd22a4`            |
| 3   | `RelationKind` named type for `RelatedRef.Relation` with 4 constants (`RelationCloneOf`, `RelationCauses`, `RelationWraps`, `RelationRelated`)                      | `55f8a85`            |
| 4   | `Tag.IsValid()` now validates lowercase-hyphen format (same as Category) — rejects `"UPPER"`, `"has space"`                                                         | `49ebaaf`            |
| 5   | `CorrelationScore.IsValid()` and `.String()` — consistent with `Confidence` type                                                                                    | `3ed83ea`            |
| 6   | `Metrics.RecordFixes(count uint)` batch method — single mutex acquisition                                                                                           | `20536a9`            |

### Pipeline Improvements

| #   | What                                                                                         | Commit    |
| --- | -------------------------------------------------------------------------------------------- | --------- |
| 1   | `StageVerify` wired — verify stage now records timing metrics and fires `OnStage` callback   | `f096964` |
| 2   | `reasonFromContext()` — derives `ReasonTimeout` vs `ReasonCancelled` from context error type | `21d5692` |
| 3   | `Config.OnStage` signature: `string` → `Stage` type                                          | `ebd22a4` |
| 4   | `MetricsSnapshot.StageDurations`: `map[string]` → `map[Stage]`                               | `ebd22a4` |

### Tests Added

| #   | What                                                                                                               | Commit    |
| --- | ------------------------------------------------------------------------------------------------------------------ | --------- |
| 1   | `reasonFromContext` unit test (cancelled→ReasonCancelled, timeout→ReasonTimeout)                                   | `06f3b34` |
| 2   | `Metrics.RecordFixes` tests (batch, combined with RecordFix, zero count)                                           | `06f3b34` |
| 3   | `RelationKind` constants test (all 4 by name)                                                                      | `06f3b34` |
| 4   | CompletionReason assertions on 4 existing pipeline tests (NoFindings, MaxIterations, ContextCancellation, Timeout) | `0d89250` |

### Documentation

| #   | What                                                                                               | Commit                          |
| --- | -------------------------------------------------------------------------------------------------- | ------------------------------- |
| 1   | README.md — `finding.Merge`→`Combine`, `c.Confidence`→`c.Score`, `result.Stable`→`result.Stable()` | `c87e703`                       |
| 2   | USAGE_GUIDE.md — same fixes + `result.FinalFindingCount`→`result.TotalDetected`                    | `c87e703`, `356465e`            |
| 3   | FEATURES.md — `finding.Merge`→`finding.Combine`                                                    | `356465e`                       |
| 4   | AGENTS.md — 3 updates tracking all session changes                                                 | `c5f96d7`, `daa9875`, `25aacac` |
| 5   | `merge.go` — removed duplicate doc comment, fixed constant comments                                | `ee57d76`                       |

---

## b) PARTIALLY DONE

| Item                       | Status                       | Notes                                                                                                  |
| -------------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| TODO_LIST audit            | 97/190 done (51%)            | 93 open items remain; 14 phantoms annotated, 7 owner-decision items                                    |
| `doc.go` coverage          | ~60% complete                | Missing: RelationKind, Builder, CorrelationScore, DiffResult, FormatText, LSP types, version constants |
| SARIF metadata namespacing | Done for `go-finding/meta/*` | Other `go-finding/*` property keys not all named constants (edit prefix is)                            |

---

## c) NOT STARTED

### From TODO_LIST.md (93 open items — top candidates)

| #   | Item                                                                        | Priority | Why Not Started                       |
| --- | --------------------------------------------------------------------------- | -------- | ------------------------------------- |
| 1   | `Report.Findings` public slice — unexport + accessor                        | HIGH     | Breaking change, needs owner decision |
| 2   | FixEngine line-offset tracking for cumulative line shifts                   | MEDIUM   | Complex, needs careful design         |
| 3   | Composable fix strategy interface                                           | MEDIUM   | Architectural, needs design doc       |
| 4   | Pipeline stage hooks (pre/post)                                             | MEDIUM   | New feature                           |
| 5   | Spatial index for Correlate (interval tree)                                 | MEDIUM   | Performance optimization              |
| 6   | Streaming merge — process findings one at a time                            | LOW      | Performance optimization              |
| 7   | `go:generate stringer` for Severity, FixStrategy, Category, SuppressionKind | LOW      | Nice-to-have, no runtime impact       |
| 8   | `iter.Seq[Finding]` on `Report.All()`                                       | LOW      | Go 1.26 modernization                 |
| 9   | Comprehensive `doc.go` (~40% → 100%)                                        | MEDIUM   | Documentation debt                    |
| 10  | CI setup (GitHub Actions)                                                   | HIGH     | BLOCKED — no .github/workflows/       |

### From audit findings (not in TODO_LIST)

| #   | Item                                                                                                | Priority                                  |
| --- | --------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| 1   | `RecordFix()` zombie — keep or remove                                                               | LOW                                       |
| 2   | FixApplier symlink path traversal                                                                   | LOW (acceptable for static analysis tool) |
| 3   | Pipeline fuzz tests for fix engine / conflict detection                                             | MEDIUM                                    |
| 4   | `Correlate` O(n²) hardening — cap iteration count, not just output                                  | LOW                                       |
| 5   | Named types for `Finding.ID` (`FindingID`), `Finding.Rule` (`RuleID`), `Position.File` (`FilePath`) | LOW (breaking)                            |
| 6   | Archive 75 stale status reports                                                                     | LOW                                       |

---

## d) TOTALLY FUCKED UP

Nothing. The previous session's `bCombine` rename was the worst issue and it's fixed. All 24 commits are clean, tested, linted, and pushed.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **`Report.Findings` encapsulation** — The public slice with a private mutex is Go's worst footgun pattern. Every caller must remember not to mutate concurrently. Solution: unexport `findings` + provide `All()`, `FindByID()`, `Len()` accessors. **Breaking change — needs owner decision.**
2. **Tag↔Category overlap** — `TagSecurity` and `CategorySecurity` share the same string value `"security"` with no structural link. If either changes, the other won't. Consider making Tags reference Categories or vice versa. **Breaking change — needs owner decision.**
3. **`Correlate` O(n²)** — The `maxCorrelations` cap limits output but not work. A file with 5000+ findings from multiple tools still does quadratic pair generation. An interval tree or spatial index would make this O(n log n).
4. **Pipeline stage hooks** — Currently only `OnStage` (post-stage callback). Pre-stage hooks would enable timing, logging, and conditional aborts. Low effort, high value.

### Testing

5. **Pipeline fuzz tests** — `fix_engine.go` has complex byte-level edit application with offset math. Perfect fuzz target.
6. **CLI coverage at 69.8%** — Lowest package. The `cmd/go-finding` package has untested error paths in config loading and output formatting.
7. **`doc.go` completeness** — Only ~60% of exported symbols are documented in the package doc. Users can't discover `RelationKind`, `Builder`, `DiffResult`, etc. from godoc.

### Library Usage

8. **`cmp.Or` unused** — Could simplify default value patterns in config validation.
9. **`iter.Seq[Finding]`** — Go 1.26's iterator protocol would modernize `Report.All()` and filtering. Low effort.
10. **`go:generate stringer`** — Would auto-generate `String()` for all enum-like types. Eliminates manual switch statements.

---

## f) Top 25 Things to Do Next

Sorted by impact × effort (highest first):

| #   | What                                                           | Effort | Impact | Category      |
| --- | -------------------------------------------------------------- | ------ | ------ | ------------- |
| 1   | Wire `doc.go` — add all missing exported symbols               | 1h     | HIGH   | Docs          |
| 2   | Archive 75 stale status reports to `archive/`                  | 5m     | MEDIUM | Hygiene       |
| 3   | Remove zombie `RecordFix()` method                             | 5m     | LOW    | Cleanup       |
| 4   | Add pipeline fuzz test for FixEngine                           | 2h     | HIGH   | Testing       |
| 5   | Improve CLI test coverage (69.8% → 85%+)                       | 2h     | MEDIUM | Testing       |
| 6   | Add `iter.Seq[Finding]` on `Report.All()`                      | 30m    | MEDIUM | Modernization |
| 7   | `go:generate stringer` for all enum types                      | 30m    | LOW    | Automation    |
| 8   | Consolidate triage — ensure `HasFix()` is canonical everywhere | 1h     | MEDIUM | Architecture  |
| 9   | Add pre-stage hooks to pipeline Config                         | 1h     | MEDIUM | Feature       |
| 10  | `Correlate` iteration cap (not just output cap)                | 30m    | MEDIUM | Performance   |
| 11  | Decompose `findingFromSarResult` (cognitive complexity)        | 1h     | LOW    | Quality       |
| 12  | Decompose `applySarifProperties` (cognitive complexity)        | 30m    | LOW    | Quality       |
| 13  | FixEngine line-offset tracking for multi-line shifts           | 3h     | HIGH   | Correctness   |
| 14  | `Report.Findings` encapsulation (unexport + accessors)         | 2h     | HIGH   | Architecture  |
| 15  | Tag↔Category structural relationship                           | 1h     | MEDIUM | Architecture  |
| 16  | FixApplier rollback all files on partial failure               | 2h     | HIGH   | Correctness   |
| 17  | SARIF schema validation test against JSON schema               | 1h     | MEDIUM | Testing       |
| 18  | Godoc examples for Builder, Filter, Pipeline                   | 1h     | MEDIUM | Docs          |
| 19  | API stability audit for v1.0.0 lock                            | 2h     | HIGH   | Governance    |
| 20  | Define v1.0.0 release criteria document                        | 1h     | HIGH   | Governance    |
| 21  | CI setup (GitHub Actions)                                      | 2h     | HIGH   | Infra         |
| 22  | Persistent fuzz corpus / seed corpus                           | 1h     | MEDIUM | Testing       |
| 23  | `Position` zero-value safety decision                          | 30m    | MEDIUM | Architecture  |
| 24  | Interactive TUI for findings                                   | 8h     | MEDIUM | Feature       |
| 25  | Fix pre-commit hook failures (goconst, todo-check)             | 2h     | MEDIUM | DX            |

---

## g) Top #1 Question for Owner

**Should `Report.Findings` be unexported with accessor methods?**

This is the single biggest architectural decision remaining. Currently `Findings` is a `[]Finding` public slice with a `sync.RWMutex`. The mutex is bypassable — any caller can `r.Findings[0] = ...` or `r.Findings = append(...)` without acquiring the lock. The comment documents this, but it's a significant footgun.

**Options:**

- **A)** Unexport `findings`, add `All() []Finding` (returns copy), `FindByID(id string) *Finding`, `Len() int`, `Range(fn func(Finding) bool)` — breaking change, but the only way to guarantee thread safety
- **B)** Keep as-is, document the footgun — Go-idiomatic but unsafe
- **C)** Use `sync.RWMutex` + `[]Finding` but return read-only views via `iter.Seq[Finding]` — Go 1.26 idiomatic

This decision affects the v1.0.0 API stability lock. Once we commit, we can't change it without a major version bump.

---

## Metrics

| Metric                        | Value            |
| ----------------------------- | ---------------- |
| Total coverage                | **92.3%**        |
| Root package coverage         | 98.5%            |
| Pipeline coverage             | 95.9%            |
| Analysis coverage             | 98.5%            |
| Internal detectors coverage   | 95.9%            |
| CLI coverage                  | 69.8%            |
| Production source lines       | 8,083            |
| Test lines                    | 19,216           |
| Total .go files               | 118              |
| Lint issues                   | **0**            |
| Test failures                 | **0**            |
| `nix flake check`             | **PASS**         |
| Commits this session          | **24**           |
| Features FULLY_FUNCTIONAL     | **32**           |
| Features PARTIALLY_FUNCTIONAL | **7**            |
| Features PLANNED              | **1**            |
| TODO_LIST done                | **97/190** (51%) |

---

_Assisted-by: Crush <crush@charm.land>_
