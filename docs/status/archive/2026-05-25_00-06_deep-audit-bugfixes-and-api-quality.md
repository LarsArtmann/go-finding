# Status Report — 2026-05-25 00:06

**Session Focus:** Deep codebase audit → find real bugs and quality gaps → fix them all

---

## A) FULLY DONE ✅

### Bugs Fixed (3)

| # | Issue                                                                                                                                                                                     | Files Changed                        | Impact                              |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ----------------------------------- |
| 1 | **SARIF AfterCode round-trip loss** — `AfterCode` was never exported as a SARIF property. If an intermediate tool stripped the fix structure, `AfterCode` was permanently lost on import. | `sarif_export.go`, `sarif_import.go` | Data loss on SARIF round-trip       |
| 2 | **DeduplicateByPosition/ByRule false dedup** — Findings with empty `Position.File` generated identical dedup keys (`":0:0"`), silently dropping distinct findings.                        | `merge.go`                           | Silent data loss in merge           |
| 3 | **Finding.Validate missed inverted ranges** — A Range with End before Start passed validation silently.                                                                                   | `finding.go`, `position.go`          | Invalid data accepted without error |

### API Quality Improvements (4)

| # | Change                                                                                                                                      | Files Changed                              | Impact                                                |
| - | ------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ | ----------------------------------------------------- |
| 4 | **LSPSeverity typed constant** — `type LSPSeverity int` replaces untyped `int` constants. `LSPDiagnostic.Severity` now uses the named type. | `lsp.go`, `lsp_test.go`, `finding_test.go` | Type safety prevents accidental invalid values        |
| 5 | **SARIF Rank omission** — `findingToSARIF` only sets `Rank` when `Confidence > 0`. NaN confidence no longer causes `json.Marshal` failures. | `sarif_export.go`                          | SARIF spec compliance; NaN resilience                 |
| 6 | **FixApplier resolveErrors surfaced** — `applyToFile` now returns provider errors via `errors.Join` instead of silently discarding them.    | `pipeline/fix_applier.go`                  | Previously invisible provider failures now propagated |
| 7 | **Range.IsInverted() helper** — New method detecting End before Start at line or column level.                                              | `position.go`                              | Reusable building block for validation                |

### Documentation Fixes (2)

| # | Issue                                                                                                              | Files Changed |
| - | ------------------------------------------------------------------------------------------------------------------ | ------------- |
| 8 | **doc.go pipeline.New example** — Would not compile. `Config{Detectors: ...}` → `New(cfg, ".", detector)`          | `doc.go`      |
| 9 | **README.md API examples** — `FromLSP` args were swapped; `FromDiagnostic` used wrong package and missing argument | `README.md`   |

### Tests Added (5 new test functions)

| Test                                                 | File                          | What It Validates                                               |
| ---------------------------------------------------- | ----------------------------- | --------------------------------------------------------------- |
| `TestRange_IsInverted`                               | `position_test.go` (8 cases)  | Inverted line, inverted column, same line/col, zero, start only |
| `TestFinding_Validate/nil_range_is_valid`            | `finding_valid_test.go`       | Nil range passes                                                |
| `TestFinding_Validate/valid_range_passes`            | `finding_valid_test.go`       | Normal range passes                                             |
| `TestFinding_Validate/inverted_range_is_invalid`     | `finding_valid_test.go`       | Inverted range rejected                                         |
| `TestDeduplicateByPosition_EmptyFileNotDeduplicated` | `merge_test.go`               | Empty file findings not falsely merged                          |
| `TestDeduplicateByRule_EmptyFileNotDeduplicated`     | `merge_test.go`               | Same for rule-based dedup                                       |
| `TestToSARIF_NaNConfidence`                          | `sarif_test.go`               | NaN confidence handled gracefully                               |
| `TestToSARIFFiltered_NaNConfidence`                  | `sarif_test.go`               | Same for filtered export                                        |
| `TestOutputResults_SARIFNaNConfidenceHandled`        | `cmd/go-finding/main_test.go` | CLI handles NaN SARIF gracefully                                |

### Quality Gates

| Metric               | Value                           |
| -------------------- | ------------------------------- |
| Tests                | **ALL PASS** (race detector on) |
| Lint                 | **0 issues**                    |
| Coverage (root)      | 98.6%                           |
| Coverage (analysis)  | 98.5%                           |
| Coverage (pipeline)  | 96.3%                           |
| Coverage (cmd)       | 92.3%                           |
| Coverage (detectors) | 96.1%                           |
| Coverage (total)     | ~95.1%                          |
| Lines of Go code     | 26,255                          |
| Files changed        | 18                              |
| Lines added/removed  | +198 / -34                      |

---

## B) PARTIALLY DONE ⚠️

Nothing partially done. All tasks were completed fully.

---

## C) NOT STARTED (from TODO_LIST.md — top actionable items)

### High Priority (unblocked)

| #  | Item                                                                               | Source                         |
| -- | ---------------------------------------------------------------------------------- | ------------------------------ |
| 1  | Extract `Equal()` 9-condition boolean into readable helper with early returns      | `finding.go:275-283`           |
| 2  | Decompose `findingFromSarResult` — cognitive complexity exceeds gocognit threshold | `sarif_import.go`              |
| 3  | Decompose `applySarifProperties` — cognitive complexity 26                         | `sarif_import.go:118`          |
| 4  | `Category.IsValid()` strict validation — accepts typos like `"securty"`            | `category.go`                  |
| 5  | FixApplier rollback all files on partial failure                                   | `pipeline/pipeline.go:522-553` |
| 6  | Consistent structured errors in pipeline — mixes `NewIOError` with `fmt.Errorf`    | `pipeline/`                    |
| 7  | Wire `FilterConflictingEdits` as opt-in pipeline Config field                      | `pipeline/conflict.go`         |
| 8  | Customizable `TriageFunc` in Config                                                | `pipeline/config.go`           |
| 9  | Report.Merge() → return new `*Report` instead of mutating receiver                 | `report.go`                    |
| 10 | Add `go:generate stringer` for Severity, FixStrategy, Category, SuppressionKind    | 4 files                        |

### Medium Priority (unblocked)

| #  | Item                                                                        | Source                       |
| -- | --------------------------------------------------------------------------- | ---------------------------- |
| 11 | Update USAGE_GUIDE.md for v0.3.0 — new features not documented              | `docs/`                      |
| 12 | Improve README.md — add badges, pipeline examples, API overview             | `README.md`                  |
| 13 | Fix `.golangci.yml` indentation — auto-configure keeps reformatting         | `.golangci.yml`              |
| 14 | Add pipeline integration test for OnFix callback                            | `pipeline/`                  |
| 15 | Add pipeline integration test with multiple concurrent detectors            | `pipeline/`                  |
| 16 | Lift `FixApplier` creation to Pipeline constructor — backup dirs orphaned   | `pipeline/pipeline.go`       |
| 17 | Centralize triage: `HasFix()` as canonical "is fixable?" check              | `pipeline/`                  |
| 18 | `FixStrategySuggest` without `AfterCode` loses strategy in SARIF round-trip | `sarif_export.go`            |
| 19 | Deduplicate `defaultMaxIterations` constant across cmd and pipeline         | `main.go:19`, `config.go:47` |
| 20 | Write API stability guarantee document (Go compat promise style)            | `docs/`                      |

---

## D) TOTALLY FUCKED UP 💥

Nothing. All changes are clean, tested, linted, and backward-compatible.

### Remaining Known Fragilities

| Issue                                                                                       | Severity                                                  | Location                              |
| ------------------------------------------------------------------------------------------- | --------------------------------------------------------- | ------------------------------------- |
| `fix_applier.go` rollback errors still silently discarded (`_ = a.backup.RollbackAll(...)`) | Low — rollback paths are error-recovery-of-error-recovery | `pipeline/fix_applier.go:113,121,133` |
| `pipeline.go:111` — `_ = p.applier.Close()` in defer discards cleanup error                 | Low — temp dir leak risk                                  | `pipeline/pipeline.go`                |
| `FixEngine.Apply` still discards conflicts/errors (documented, wrapper API)                 | Low — `ApplyWithConflicts` available                      | `pipeline/fix_engine.go:48`           |

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **SARIF types use `map[string]any` property bags** — Constrained by SARIF spec, but internal helpers should use typed accessors (`stringProp`) consistently
2. **`pipeline/pipeline.go` `runIteration` is 91 lines** — Extract detect/process/apply into distinct methods
3. **`cmd/go-finding/main.go` `run()` is 119 lines** — Extract flag parsing, config building, and output formatting
4. **`Report.Merge` mutates receiver** — Should return new `*Report` for immutability
5. **`FixApplier` lifecycle** — Created per-iteration in pipeline; backup dirs from iteration N orphaned when N+1 creates new applier. Should be created once at Pipeline construction.

### Type Safety

6. **`Category.IsValid()` accepts any non-empty string** — Should validate against known constants, at least with a warning
7. **`DeduplicateBy` int-based iota has no `String()`** — Invisible in logs/debug
8. **`SuppressionKind` and `ErrorCategory` have no `String()` method** — Zero value renders as `""`
9. **`ConflictInfo.Reason` is untyped `string`** — Should be `type ConflictReason string`

### Testing

10. **CLI coverage at 92.3%** — Lowest in the project. Missing: error paths in config loading, SARIF output edge cases
11. **No SARIF schema validation test** — Should validate output against SARIF 2.1.0 JSON schema
12. **No concurrent Report read-write race test** — Documented as TODO in `report_test.go`
13. **17 fuzz targets with no checked-in seed corpus** — Fuzzing starts from scratch each time

### Documentation

14. **USAGE_GUIDE.md not updated for v0.3.0** — Missing Builder, Suppression, Confidence, Diff, Conflict, FixEngine docs
15. **CONTRIBUTING.md pipeline file listing incomplete** — Missing `fix_edit.go`, `file_backup.go`, `result.go`
16. **CONTRIBUTING.md says "Never return bare fmt.Errorf" but shows it as correct** — Contradictory guidance

---

## F) TOP 25 THINGS TO GET DONE NEXT

Priority-ordered by impact × effort ratio:

| #  | Task                                                     | Impact | Effort | Type         |
| -- | -------------------------------------------------------- | ------ | ------ | ------------ |
| 1  | `Category.IsValid()` strict validation                   | High   | Low    | Bug          |
| 2  | Update USAGE_GUIDE.md for v0.3.0                         | High   | Medium | Docs         |
| 3  | Fix CONTRIBUTING.md contradictions                       | Medium | Low    | Docs         |
| 4  | Extract `Equal()` into early-return helper               | Medium | Low    | Refactor     |
| 5  | Decompose `findingFromSarResult` (complexity)            | Medium | Low    | Refactor     |
| 6  | Decompose `applySarifProperties` (complexity 26)         | Medium | Low    | Refactor     |
| 7  | Consistent structured errors in pipeline                 | Medium | Low    | Quality      |
| 8  | `DeduplicateBy.String()` method                          | Low    | Low    | Quality      |
| 9  | `SuppressionKind.String()` / `ErrorCategory.String()`    | Low    | Low    | Quality      |
| 10 | `ConflictReason` named type                              | Low    | Low    | Quality      |
| 11 | FixApplier rollback all files on partial failure         | High   | Medium | Bug          |
| 12 | Lift FixApplier creation to Pipeline constructor         | Medium | Medium | Architecture |
| 13 | Report.Merge() → return new `*Report`                    | Medium | Medium | Architecture |
| 14 | Customizable `TriageFunc` in Config                      | Medium | Medium | Feature      |
| 15 | Wire `FilterConflictingEdits` as opt-in Config field     | Medium | Medium | Feature      |
| 16 | Add pipeline integration test for OnFix callback         | Medium | Low    | Tests        |
| 17 | Add pipeline integration test with concurrent detectors  | Medium | Low    | Tests        |
| 18 | Write concurrent Report read-write race test             | Medium | Low    | Tests        |
| 19 | Add SARIF schema validation test                         | Medium | Medium | Tests        |
| 20 | `go:generate stringer` for 4 enum types                  | Low    | Low    | Quality      |
| 21 | `FixStrategySuggest` SARIF round-trip fix                | Medium | Medium | Bug          |
| 22 | Deduplicate `defaultMaxIterations` constant              | Low    | Low    | DRY          |
| 23 | Write API stability guarantee document                   | High   | Medium | Docs         |
| 24 | Persist fuzz corpus / seed corpus                        | Medium | Medium | Tests        |
| 25 | Fix CI: add Windows testing, caching, reduce duplication | Medium | Medium | CI           |

---

## G) TOP #1 QUESTION FOR OWNER

**Should `Category.IsValid()` be strict or lenient?**

Currently `Category.IsValid()` returns `true` for any non-empty string, including typos like `"securty"`. Options:

- **A) Strict** — Only accept `Category*` constants + registered custom categories via `RegisterCategory(string)`. Breaks consumers using arbitrary strings. Best type safety.
- **B) Lenient (current)** — Accept any non-empty string. Maximum compatibility. Risk of silent typos.
- **C) Warn-only** — Accept any string but log a warning for non-standard values. Middle ground.

This is an API design decision that affects the v1.0 compatibility promise. Once locked, changing from B→A is breaking.

---

## Session Metrics

| Metric             | Value   |
| ------------------ | ------- |
| Duration           | ~45 min |
| Files modified     | 18      |
| Lines added        | 198     |
| Lines removed      | 34      |
| New test cases     | ~20     |
| Bugs found & fixed | 3       |
| API improvements   | 4       |
| Doc fixes          | 2       |
| Zero regressions   | ✅      |

---

_Assisted-by: Crush <crush@charm.land>_
