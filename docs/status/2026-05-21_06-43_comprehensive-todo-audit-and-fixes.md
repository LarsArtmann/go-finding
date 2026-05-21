# Status Report — 2026-05-21 06:43

**Branch:** master | **Version:** v0.3.0 (tagged) | **Go:** 1.26.2

---

## Executive Summary

go-finding is in **strong shape**. Build is green, all tests pass with `-race`, coverage is 95.5%, and the core library is feature-complete for v0.3.0. This session resolved 59 TODO items (audited as already-done or newly fixed) and added 6 code fixes with 8 new test functions. The main gaps are: documentation freshness, lint hygiene (225 warnings, 43 exhaustruct nolints, 79 total nolints), missing ADRs, and CI maturity (govulncheck not tested, no flake.nix, no benchmark regression tracking).

---

## A) FULLY DONE ✅

### Resolved This Session (code changes, 17 files, +369/-120 lines)

| File                        | Change                                                                  | Impact                         |
| --------------------------- | ----------------------------------------------------------------------- | ------------------------------ |
| `pipeline/fix_engine.go:64` | `HasCodeChange()` replaces inline `BeforeCode == "" && AfterCode == ""` | Centralized filter logic       |
| `pipeline/metrics.go:93`    | `TotalDuration()` guards negative duration                              | Prevents UB when end < start   |
| `pipeline/metrics.go:149`   | `Snapshot()` same negative guard                                        | Consistent behavior            |
| `pipeline/pipeline.go:89`   | `ran` bool enforces single-use `Run()`                                  | Catches double-invocation bugs |
| `pipeline/pipeline.go:478`  | `OnFix` reports `false` for skipped safeFixes                           | Accurate callback semantics    |
| `pipeline/partial.go:138`   | `errors.Join` with `%w` per detector                                    | `errors.Is` chain support      |
| `position.go`               | `Position.IsZero()`, `Position.HasLocation()`, `Range.IsSingleLine()`   | API completeness               |
| `finding.go`                | `Finding.HasRange()`                                                    | API completeness               |
| `category.go`               | `Category.IsSecurity()`                                                 | API completeness               |
| `report.go`                 | `Report.CountBySeverity()`                                              | Convenience method             |

### Audit: Already Done Before This Session (58 items)

These were marked open in TODO_LIST.md but already resolved in prior work:

- **Build-blocking issues (4):** `isContextDone` undefined, unused `"fmt"` import, duplicate `NewWithT(t)`, `Build()` 2-value assignment — all false positives / already fixed
- **Features already implemented (30+):** `KeySeparator` constant, `Confidence` named type, `ToolInfo.Validate()`, `Config.Validate()`, `RetryConfig.Validate()`, `Report.Validate()` with `RLock`, `Suppression.IsActive()`, `FormatText/FormatMarkdown`, `Diff()`, `ToDiagnostic()`, `Tag.IsStandard()`, `Range.IsValid()`, `Range.Contains()`, `Range.LineCount()` inverted handling, `Config.DetectorTimeouts`, `Config.Logger` structured logging, `Config.OnStage`, `FindingProcessor.Process` with `context.Context` + `error`, detector registry, SARIF file split (`sarif_types.go`, `sarif_export.go`, `sarif_import.go`), `ErrPositionUnresolvable` sentinel, `NewFixApplier` error propagation, `ConflictInfo.ConflictsWith` populated, `Correlation` camelCase JSON tags
- **Non-existent items (10):** `container.go` / `samber/do/v2`, `errors/errors_result.go` / `samber/mo`, `internal/events/`, `internal/rules/`, `replace` directive in go.mod, deprecated `ConflictDetector`/`Verifier` structs
- **Infrastructure (6):** pre-commit hook executable, `LICENSE`, `CONTRIBUTING.md`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `.goreleaser.yml`

### Test Results

| Package              | Coverage  | Status |
| -------------------- | --------- | ------ |
| `go-finding` (root)  | 99.7%     | ✅     |
| `analysis`           | 98.5%     | ✅     |
| `cmd/go-finding`     | 93.9%     | ✅     |
| `internal/detectors` | 96.1%     | ✅     |
| `pipeline`           | 96.5%     | ✅     |
| **Total**            | **95.5%** | ✅     |

New tests added: `TestPosition_IsZero`, `TestPosition_HasLocation`, `TestRange_IsSingleLine`, `TestFinding_HasRange`, `TestCategory_IsSecurity`, `TestReport_CountBySeverity`, `TestPipelineRun_SingleUse`, `TestPipelineRun_SingleUse_ErrorMessage`, `TestMetrics_TotalDuration_NegativeGuard`, `TestMetrics_Snapshot_NegativeGuard`, `TestFormatPartialErrors_ErrorWrapping`.

---

## B) PARTIALLY DONE 🔶

| Area                   | Status                                                                 | Gap                                                                                                                                              |
| ---------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **TODO_LIST.md audit** | 59/252 items verified done                                             | 129 items remain open; 4 HIGH, 24 MEDIUM, 12 LOW, 89 UNKNOWN                                                                                     |
| **SARIF round-trip**   | Export/import decomposed into 3 files                                  | Still missing: `Location → Locations` (plural) spec compliance, non-string metadata preservation, cognitive complexity in `findingFromSarResult` |
| **Conflict detection** | `ConflictInfo.ConflictsWith` populated, `FilterConflictingEdits` works | Still groups by `Overlaps` instead of `Adjacent`; `FilterConflictingEdits` not wired as Config opt-in                                            |
| **CLI**                | Registry, profiling, config loading, text/markdown/json/sarif output   | `FixProviders` not wired through CLI config; `TriageFunc` not customizable; no `FixProviders` from YAML                                          |
| **Documentation**      | 21 godoc examples, USAGE_GUIDE, CHANGELOG, CONTRIBUTING, LICENSE       | FEATURES.md and USAGE_GUIDE.md outdated for v0.3.0; no ADRs for Properties/HasCodeChange/FixProvider decisions                                   |
| **CI/CD**              | `ci.yml` + `release.yml` exist, govulncheck wired                      | GoReleaser untested; no `-race` in CI; no benchmark regression; no per-package coverage thresholds                                               |
| **Lint**               | Down from hundreds to 79 nolints                                       | Still 43 exhaustruct nolints, 225 lint warnings (goconst, revive, wsl_v5), no `golines` in CI                                                    |
| **Metrics**            | Thread-safe with mutex, negative guard, snapshot                       | `MetricsSnapshot` has exported map fields (no accessor methods); partial detection doesn't record metrics                                        |

---

## C) NOT STARTED ⬜

### HIGH Priority (4 items)

1. Document SARIF critical round-trip loss — `SeverityCritical` → `"error"` → `SeverityError`
2. `Finding` struct sub-grouping into embedded sub-structs — deferred to v2, breaking change
3. Add `golines` to CI for consistent line breaking
4. Clean up gopls hints (~12 non-critical: rangeint, newexpr, mapsloop, stringsseq)

### MEDIUM Priority (24 items)

Notable unstarted items:

- Fix `.golangci.yml` to work without `--no-verify`
- Fix BuildFlow `gomodguard_v2` auto-configure loop
- Fix `Finding.Key()` cross-tool collision — include `ToolName` in fallback key
- Fix conflict detection overgrouping — `Overlaps` → `Adjacent`
- Fix `DeduplicateByID` — skip empty-ID findings
- Update USAGE_GUIDE.md and FEATURES.md for v0.3.0
- Fix `.gitignore` line 43 corruption
- Fix pre-commit hook failures (goconst, todo-check, library-policy)
- Fix 225 lint warnings in 3 batches
- FixEngine line-offset tracking for cumulative line shifts
- Make fix strategy composable as interface

### LOW Priority (12 items)

- API stability review for v1.0.0
- Decide `FixStrategyAI` fate
- Documentation gaps (BySeverityAtLeast, Report.All copies, FindByID copy, SARIF losses)
- Write API stability guarantee document
- Decide stable ID format

### UNKNOWN Priority (89 items)

The bulk of remaining items. Notable categories:

- **Missing tests** (~15 items): concurrent Report race test, SARIF fuzz test, Windows path ParseID, context-cancel tests, FixApplier error-path tests, multiple integration tests
- **Missing features** (~10 items): `io.WriterTo` for SARIF, streaming merge, spatial index for Correlate, TUI review, watch mode
- **Architecture decisions** (~8 items): Properties map, repository structure, NewFinding API, Confidence protection, FixProvider location
- **Code quality** (~10 items): Extract Equal helper, Category.IsValid strict, deprecated Tag cleanup, detectResult ghost type, detectParallel dead code
- **CI/Infra** (~8 items): flake.nix, go.work, benchmark regression, gosec, GoReleaser testing, pkg.go.dev
- **Documentation** (~8 items): ADRs, consumer migration guide, integration guide, doc.go completion, JSON schema
- **Cleanup** (~5 items): stale nolint directives, status reports archive, git-town.toml, coverage files, binary artifacts

---

## D) TOTALLY FUCKED UP 💥

| Issue                                   | Severity | Details                                                                                                                                               |
| --------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| **gopls unused constant warnings**      | Low      | 16+ `gopls unusedfunc` warnings in test files — test helper constants flagged but harmless                                                            |
| **No `flake.nix`**                      | Medium   | AGENTS.md says "never use Makefile, use flake.nix" but flake.nix doesn't exist. `justfile` is the actual build tool. Contradiction in project policy. |
| **TODO_LIST had 59 phantom items**      | Medium   | 59 items marked as open were already done or didn't exist. The TODO list was stale/inaccurate before this session.                                    |
| **89 items in "Unknown" priority**      | Process  | 69% of open items have no priority assigned. Makes triage impossible.                                                                                 |
| **No godoc examples in non-test files** | Low      | `func Example*` only in `*_test.go` files — they won't appear on pkg.go.dev unless in the same package (they are, so OK, but could be confusing)      |
| **`.golangci.yml` needs `--no-verify`** | Medium   | Config is fragile; auto-configure tools reformat it. Blocks clean CI.                                                                                 |

---

## E) WHAT WE SHOULD IMPROVE 📈

### Code Quality

1. **Reduce nolint count from 79 → <25** — audit every `//nolint` directive, fix the underlying issue or add file-level exclusions
2. **Eliminate exhaustruct nolints** — add constructors (`NewSummary()`, `NewPartialResult()`, etc.) for commonly-constructed structs
3. **Fix 225 lint warnings** — batch approach: revive (unused receivers/params), goconst (string literals → constants), wsl_v5 (formatting)
4. **Modernize to Go 1.21+ stdlib** — `slices.Contains`, `slices.Delete`, `maps.Keys` throughout

### Architecture

5. **Add ADRs** — Missing: Properties map type, HasCodeChange method, FixProvider plugin architecture, Pipeline single-use contract
6. **Consolidate `findingKey`** — duplicated in `verify.go` and `merge.go`
7. **Remove ghost types** — `detectResult` is identical to `PartialResult` with different field names
8. **Structured errors everywhere in pipeline** — `fix_applier.go` uses structured; `pipeline.go` uses raw `fmt.Errorf`

### Testing

9. **Add SARIF fuzz tests** — 20 fuzz targets exist but no seed corpus; SARIF import parses untrusted input
10. **Add concurrent Report race test** — `Report` has mutex but no dedicated race test
11. **CLI coverage 93.9% → 95%+** — missing error-path tests for profiling, config loading
12. **Add pipeline integration tests** — OnFix callback, multiple concurrent detectors

### Documentation

13. **Update FEATURES.md and USAGE_GUIDE.md for v0.3.0** — FormatText, FormatMarkdown, ToDiagnostic, Diff, HasRange, CountBySeverity all undocumented
14. **Complete doc.go** — currently ~40% complete at 93 lines
15. **Write consumer migration guide** — v0.1.3 → v0.2.0+ changes undocumented

### Infrastructure

16. **Fix `.golangci.yml`** — make it work without `--no-verify`; pin tool versions
17. **Add `-race` to CI** — only runs locally, not in GitHub Actions
18. **Test GoReleaser config** — exists but never validated
19. **Create `flake.nix`** or update AGENTS.md to remove the "use flake.nix" mandate
20. **Triage "Unknown" priority items** — 89 items need priority assignment

---

## F) Top 25 Things We Should Get Done Next

Ranked by impact × effort (Pareto principle):

| #   | Item                                                                 | Priority | Effort | Impact                       |
| --- | -------------------------------------------------------------------- | -------- | ------ | ---------------------------- |
| 1   | Fix `.golangci.yml` to work without `--no-verify`                    | HIGH     | S      | HIGH — unblocks clean CI     |
| 2   | Update FEATURES.md for v0.3.0                                        | MED      | S      | HIGH — discoverability       |
| 3   | Update USAGE_GUIDE.md for v0.3.0                                     | MED      | M      | HIGH — user-facing           |
| 4   | Triage 89 "Unknown" priority TODO items to correct priority          | PROCESS  | M      | HIGH — enables planning      |
| 5   | Add ADR for Properties map type decision                             | MED      | S      | MED — prevents relitigation  |
| 6   | Add ADR for FixProvider plugin architecture                          | MED      | S      | MED — architectural clarity  |
| 7   | Fix 10 testifylint `require-error` warnings                          | MED      | S      | MED — lint hygiene           |
| 8   | Fix 6 `paralleltest` warnings — add `t.Parallel()`                   | MED      | S      | MED — test quality           |
| 9   | Reduce exhaustruct nolints 43 → <25 via constructors                 | MED      | M      | MED — code quality           |
| 10  | Add `-race` flag to CI workflow                                      | MED      | S      | MED — catch races in CI      |
| 11  | Add SARIF fuzz test with seed corpus                                 | MED      | M      | HIGH — security              |
| 12  | Fix `Finding.Key()` cross-tool collision (add ToolName)              | MED      | S      | MED — correctness            |
| 13  | Fix conflict detection overgrouping (Overlaps → Adjacent)            | MED      | S      | MED — correctness            |
| 14  | Wire `FilterConflictingEdits` as opt-in Config field                 | MED      | S      | MED — API completeness       |
| 15  | Add concurrent Report race test                                      | MED      | S      | MED — concurrency safety     |
| 16  | Fix `DeduplicateByID` skip empty-ID findings                         | MED      | S      | MED — correctness            |
| 17  | Extract `findingKey` to shared utility (verify.go + merge.go)        | LOW      | S      | LOW — DRY                    |
| 18  | Fix 225 lint warnings batch 1: goconst (string literals → constants) | MED      | M      | MED — maintainability        |
| 19  | Fix `.gitignore` line 43 corruption                                  | MED      | S      | LOW — hygiene                |
| 20  | Complete doc.go (40% → 80%+)                                         | LOW      | M      | MED — discoverability        |
| 21  | Add `io.WriterTo` for SARIF streaming                                | LOW      | S      | LOW — API completeness       |
| 22  | Fix partial detection missing metrics recording                      | MED      | S      | MED — observability          |
| 23  | Add `golines` to CI for consistent line breaking                     | HIGH     | S      | MED — formatting consistency |
| 24  | Document SARIF critical round-trip loss                              | HIGH     | S      | MED — user awareness         |
| 25  | Test GoReleaser config                                               | MED      | S      | LOW — release readiness      |

---

## G) Top #1 Question I Cannot Figure Out Myself 🤔

**Should this project have a `flake.nix` or not?**

AGENTS.md (global) mandates: _"Never use Makefile — use `flake.nix` for all build/task automation in LarsArtmann projects"_ and _"justfile is deprecated"_. But this project has no `flake.nix` at all — it uses a `justfile` as its primary build tool. The TODO list has 5 phases of Nix migration as open items.

The contradiction means either:

1. **Migrate to `flake.nix`** — significant effort (5 phases), but aligns with project policy
2. **Update AGENTS.md to acknowledge `justfile` is acceptable** — quick fix, but goes against the standard
3. **Keep both** — `flake.nix` wraps `justfile` — common but adds complexity

This decision affects every subsequent CI/CD, build, and developer workflow choice. I cannot resolve it without knowing the project owner's intent.

---

## Metrics Dashboard

| Metric              | Value                            |
| ------------------- | -------------------------------- |
| Go files            | 111                              |
| Total lines         | 25,525                           |
| Test files          | 65                               |
| Test lines          | 18,382                           |
| Test:Code ratio     | 0.72                             |
| Coverage            | 95.5%                            |
| Fuzz targets        | 20                               |
| Benchmarks          | 30                               |
| Godoc examples      | 21                               |
| Direct dependencies | 5                                |
| Nolint directives   | 79                               |
| Exhaustruct nolints | 43                               |
| TODO items done     | 59 / 252 (23%) → now 59 resolved |
| TODO items open     | 129                              |
| HIGH open           | 4                                |
| MEDIUM open         | 24                               |
| LOW open            | 12                               |
| UNKNOWN open        | 89                               |
| Race detector       | ✅ Clean                         |
| Go version          | 1.26.2                           |
| Latest tag          | v0.3.0                           |
