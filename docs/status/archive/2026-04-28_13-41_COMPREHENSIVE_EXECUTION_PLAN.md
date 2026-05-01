# Comprehensive Execution Plan — go-finding + go-structure-linter

**Date:** 2026-04-28 13:41 UTC  
**Scope:** Every known TODO, bug, improvement, and debt item across both repos  
**Constraint:** Each task ≤ 12 minutes (break down large items)  
**Sorted by:** Importance × Impact / Effort × Customer Value

---

## Scoring Legend

| Dimension          | Scale                                            |
| ------------------ | ------------------------------------------------ |
| **Importance**     | 1-5 (5 = blocks release / migration / trust)     |
| **Impact**         | 1-5 (5 = affects every consumer / every run)     |
| **Effort**         | 1-5 (1 = 1-3 min, 3 = 5-8 min, 5 = 10-12 min)    |
| **Customer Value** | 1-5 (5 = direct consumer-visible improvement)    |
| **Score**          | `(Importance × Impact × CustomerValue) / Effort` |

---

## Phase P0: BLOCKING — Fix Before Any Release

| #   | ID   | Task                                                                         | Repo       | File(s)                          | I   | Im  | E   | CV  | Score | Est |
| --- | ---- | ---------------------------------------------------------------------------- | ---------- | -------------------------------- | --- | --- | --- | --- | ----- | --- |
| 1   | M-13 | Fix `ToSARIFFiltered` name → document it filters by severity AND suppression | go-finding | `sarif.go:159`                   | 5   | 4   | 2   | 5   | 50.0  | 5m  |
| 2   | M-4  | SARIF round-trip: preserve non-string metadata via JSON type assertion       | go-finding | `sarif.go:450-453`               | 5   | 3   | 3   | 4   | 20.0  | 8m  |
| 3   | M-8  | `Range.LineCount()` returns 0 for inverted ranges — document or fix          | go-finding | `position.go:90-95`              | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 4   | M-15 | `FixStrategySuggest` without `AfterCode` loses suggestion in SARIF           | go-finding | `finding.go:110-111`, `sarif.go` | 4   | 3   | 3   | 4   | 16.0  | 8m  |
| 5   | M-16 | `Report.PrettyJSON` includes suppressed — add filtered alternative           | go-finding | `json.go:24-31`                  | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 6   | M-17 | `maxIterations: 0` CLI vs config inconsistency                               | go-finding | `cmd/go-finding/main.go:391-392` | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 7   | M-18 | `BySeverityAtLeast` excludes invalid severities — document behavior          | go-finding | `filter.go:42-44`                | 3   | 3   | 1   | 3   | 27.0  | 3m  |
| 8   | M-6  | Document `Report.All()` yields copies — modifications lost                   | go-finding | `report.go:137-145`              | 3   | 3   | 1   | 3   | 27.0  | 3m  |
| 9   | M-7  | Document `FindByID` returns copy — modifications don't affect report         | go-finding | `report.go:113-123`              | 3   | 3   | 1   | 3   | 27.0  | 3m  |
| 10  | M-12 | `Correlation` JSON tags snake_case → camelCase                               | go-finding | `merge.go:154-158`               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 11  | M-19 | Clamp `Confidence` in `NewFinding` constructor                               | go-finding | `finding.go:43-53`               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 12  | M-20 | Clamp `Confidence` in `Builder.WithConfidence`                               | go-finding | `finding_builder.go:82-84`       | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 13  | M-5  | Verify `ParseID` with backslash Windows paths (`C:\Users\...`)               | go-finding | `id.go`, `id_test.go`            | 3   | 2   | 2   | 2   | 6.0   | 5m  |

---

## Phase P1: QUICK WINS — High Impact, Low Effort (≤ 5 min each)

| #   | ID     | Task                                                                 | Repo       | File(s)                             | I   | Im  | E   | CV  | Score | Est |
| --- | ------ | -------------------------------------------------------------------- | ---------- | ----------------------------------- | --- | --- | --- | --- | ----- | --- |
| 14  | PUSH   | Push all 6 commits to origin/master                                  | go-finding | —                                   | 5   | 5   | 1   | 5   | 125.0 | 2m  |
| 15  | TAG    | Tag v0.1.3 release after push                                        | go-finding | git                                 | 4   | 4   | 1   | 4   | 64.0  | 2m  |
| 16  | H-12   | Add test for `RegisterDetector` concurrent safety                    | go-finding | `cmd/go-finding/main_test.go`       | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 17  | TEST1  | Add `Finding.IsValid()` test (0% coverage)                           | go-finding | `finding_test.go`                   | 4   | 2   | 2   | 2   | 8.0   | 5m  |
| 18  | TEST2  | Add `Suppression.IsValid()` test (0% coverage)                       | go-finding | `suppression_test.go`               | 4   | 2   | 2   | 2   | 8.0   | 5m  |
| 19  | TEST3  | Add `ErrorCategory.IsValid()` test (0% coverage)                     | go-finding | `errors_test.go`                    | 4   | 2   | 2   | 2   | 8.0   | 5m  |
| 20  | TEST4  | Add `FilterConflictingFixes` test (0% coverage)                      | go-finding | `pipeline/conflict_test.go`         | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 21  | TEST5  | Add `AnalyzeConflicts` test (0% coverage)                            | go-finding | `pipeline/conflict_test.go`         | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 22  | TEST6  | Add `Severity.LessThan` invalid-input test (66.7% branch)            | go-finding | `severity_test.go`                  | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 23  | TEST7  | Add `equalTimePtr` both-nil test                                     | go-finding | `finding_test.go`                   | 3   | 1   | 1   | 1   | 3.0   | 3m  |
| 24  | TEST8  | Add `RetryConfig.Validate` edge-case tests                           | go-finding | `pipeline/retry_test.go`            | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 25  | LINT1  | Remove 3 unused `//nolint` directives                                | go-finding | 3 files                             | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 26  | LINT2  | Fix `tagliatelle` — `finding_ids` → `findingIds` in merge.go         | go-finding | `merge.go`                          | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 27  | LINT3  | Replace `slices.Contains` at `merge_test.go:208`                     | go-finding | `merge_test.go`                     | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 28  | LINT4  | Preallocate `all` slice in `pipeline_test.go:332`                    | go-finding | `pipeline_test.go`                  | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 29  | LINT5  | Extract `"changed"` string to constant at `finding_extra_test.go:46` | go-finding | `finding_extra_test.go`             | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 30  | DOC1   | Add godoc to exported `Pos()` function                               | go-finding | `position.go:376`                   | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 31  | CLEAN1 | Delete stale `report/jscpd-report.json`                              | go-finding | root                                | 2   | 1   | 1   | 1   | 2.0   | 2m  |
| 32  | CLEAN2 | Delete binary artifacts (`go-finding`, `govet`) from repo root       | go-finding | root                                | 2   | 1   | 1   | 1   | 2.0   | 2m  |
| 33  | CLEAN3 | Delete stale coverage files (`cover.out`, `coverage.out`)            | go-finding | root                                | 2   | 1   | 1   | 1   | 2.0   | 2m  |
| 34  | CLEAN4 | Add `govet` binary to `.gitignore`                                   | go-finding | `.gitignore`                        | 2   | 1   | 1   | 1   | 2.0   | 2m  |
| 35  | GSL-1  | Remove `replace` directive from go-structure-linter go.mod           | gsl        | `go.mod:5`                          | 5   | 5   | 2   | 5   | 62.5  | 5m  |
| 36  | GSL-2  | Delete `internal/events/events.go` and all Publish calls             | gsl        | multiple                            | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 37  | GSL-3  | Remove `RuleChecker` type alias from `interface.go:125`              | gsl        | `internal/rules/interface.go`       | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 38  | GSL-4  | Remove `SeverityProvider` type alias from `interface.go:129`         | gsl        | `internal/rules/interface.go`       | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 39  | GSL-5  | Remove `RunAllRules` pass-through from `run_all.go`                  | gsl        | `internal/rules/run_all.go`         | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 40  | GSL-6  | Remove `ContainsSubstringsValidator` from fixable_helpers            | gsl        | `internal/rules/fixable_helpers.go` | 3   | 1   | 1   | 1   | 3.0   | 3m  |

---

## Phase P2: HIGH CUSTOMER VALUE — Real Improvements Consumers Notice

| #   | ID     | Task                                                                   | Repo       | File(s)                                         | I   | Im  | E   | CV  | Score | Est |
| --- | ------ | ---------------------------------------------------------------------- | ---------- | ----------------------------------------------- | --- | --- | --- | --- | ----- | --- |
| 41  | SARIF1 | Fix SARIF `Location` → `Locations` (plural, array) for spec compliance | go-finding | `sarif.go:43`                                   | 4   | 4   | 3   | 5   | 26.7  | 8m  |
| 42  | SARIF2 | Add SARIF extension to preserve `SeverityCritical` on round-trip       | go-finding | `sarif.go`                                      | 4   | 3   | 4   | 4   | 12.0  | 10m |
| 43  | SARIF3 | Document SARIF round-trip losses in README                             | go-finding | `README.md`                                     | 3   | 3   | 2   | 3   | 13.5  | 5m  |
| 44  | CLI1   | Add CLI end-to-end test with mock detector                             | go-finding | `cmd/go-finding/integration_test.go`            | 4   | 4   | 3   | 5   | 26.7  | 8m  |
| 45  | CLI2   | Add `io.WriterTo` for SARIF output — direct writing                    | go-finding | `sarif.go`                                      | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 46  | CLI3   | Fix profiling FD leak in `setupProfiling`                              | go-finding | `cmd/go-finding/main.go:133-149`                | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 47  | CLI4   | Fix `cmd/go-finding/main.go` wrapcheck errors                          | go-finding | `cmd/go-finding/main.go`                        | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 48  | CLI5   | Fix 26 lint warnings in `cmd/go-finding/main.go`                       | go-finding | `cmd/go-finding/main.go`                        | 3   | 3   | 4   | 3   | 6.8   | 10m |
| 49  | FIX1   | Fix `Metrics.StageTiming` value receiver — copies mutex                | go-finding | `pipeline/metrics.go`                           | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 50  | FIX2   | Guard `TotalDuration()` against negative when `EndTime` unset          | go-finding | `pipeline/metrics.go:52-57`                     | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 51  | FIX3   | Make Metrics exported map fields unexported with accessors             | go-finding | `pipeline/metrics.go:12-14`                     | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 52  | FIX4   | Add sync.Mutex to `Report.AddFinding` for goroutine safety             | go-finding | `report.go:37-39`                               | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 53  | FIX5   | Wire `Correlate()` into Pipeline as optional stage                     | go-finding | `pipeline/pipeline.go`                          | 4   | 3   | 4   | 4   | 12.0  | 10m |
| 54  | FIX6   | Document Pipeline is NOT safe for concurrent `Run()`                   | go-finding | `pipeline/pipeline.go:99-107`                   | 3   | 2   | 1   | 2   | 12.0  | 3m  |
| 55  | ID1    | Fix `ParseID` Windows backslash paths                                  | go-finding | `id.go`                                         | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 56  | TEST9  | Add `Finding.Equal` field-mismatch test                                | go-finding | `finding_test.go`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 57  | TEST10 | Add `cloneFindings` edge-case test                                     | go-finding | `merge_test.go`                                 | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 58  | TEST11 | Add `Range.Contains` edge-case tests (80% → 100%)                      | go-finding | `position_test.go`                              | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 59  | TEST12 | Add `checkColumnRange` + `hasLineRange` tests                          | go-finding | `position_test.go`                              | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 60  | TEST13 | Add `Verifier.Verify` error-path tests                                 | go-finding | `pipeline/verify_test.go`                       | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 61  | TEST14 | Add `severityToSARIFLevel` edge-case test                              | go-finding | `sarif_test.go`                                 | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 62  | TEST15 | Add `findingFromSarResult` SARIF import path tests                     | go-finding | `sarif_test.go`                                 | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 63  | TEST16 | Add `PrettyJSON` / `LineJSON` error-path tests                         | go-finding | `json_test.go`                                  | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 64  | TEST17 | Add `DeduplicateByPosition` ≠ `DeduplicateByRule` behavior test        | go-finding | `merge_test.go`                                 | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 65  | TEST18 | Add `LSP toZeroBased` test for 0 Line case                             | go-finding | `lsp_test.go`                                   | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 66  | TEST19 | Convert 4 `errors.New()` in `retry.go` to sentinel errors              | go-finding | `pipeline/retry.go`                             | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 67  | TEST20 | Add `applyTriage` tests with `FixStrategyDirect` findings              | go-finding | `pipeline/pipeline_test.go`                     | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 68  | TEST21 | Add FixApplier error-path unit tests (87% → 90%+)                      | go-finding | `pipeline/fix_applier_test.go`                  | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 69  | TEST22 | Add SARIF parser fuzz test for untrusted input                         | go-finding | `sarif_fuzz_test.go`                            | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 70  | TEST23 | Add SARIF schema validation test                                       | go-finding | `sarif_test.go`                                 | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 71  | GSL-7  | Remove `samber/do/v2` — replace DI container with plain struct         | gsl        | `container.go`                                  | 5   | 5   | 4   | 5   | 31.3  | 10m |
| 72  | GSL-8  | Remove `samber/mo` — replace `LinterResult[T]` with `(T, error)`       | gsl        | `errors/errors_result.go`                       | 5   | 5   | 4   | 5   | 31.3  | 10m |
| 73  | GSL-9  | Fix 225 lint warnings batch 1: revive                                  | gsl        | project-wide                                    | 4   | 4   | 4   | 3   | 12.0  | 10m |
| 74  | GSL-10 | Fix 225 lint warnings batch 2: exhaustruct + errcheck                  | gsl        | project-wide                                    | 4   | 4   | 4   | 3   | 12.0  | 10m |
| 75  | GSL-11 | Fix 225 lint warnings batch 3: wsl_v5 + formatting                     | gsl        | project-wide                                    | 4   | 4   | 3   | 3   | 16.0  | 8m  |
| 76  | GSL-12 | Fix global state: inject GitAuthorProvider into rules                  | gsl        | `rule_service.go`, `utils/git_author.go`        | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 77  | GSL-13 | Consolidate gitignore matching into single utility                     | gsl        | `sensitive_file_rule.go`, `checks/gitignore.go` | 4   | 3   | 3   | 3   | 12.0  | 8m  |
| 78  | GSL-14 | Remove dead `CheckContext` from Rule interface                         | gsl        | `interface.go:27`, `rule_service.go:183`        | 4   | 3   | 2   | 3   | 18.0  | 5m  |
| 79  | GSL-15 | Add tests for `agent_config_rule.go`                                   | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 80  | GSL-16 | Add tests for `benchmark_naming_rule.go`                               | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 81  | GSL-17 | Add tests for `fuzz_test_structure_rule.go`                            | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 82  | GSL-18 | Add tests for `golden_file_rule.go`                                    | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 83  | GSL-19 | Add tests for `pkg_errors_structure_rule.go`                           | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 84  | GSL-20 | Add tests for `race_condition_test_rule.go`                            | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 85  | GSL-21 | Add tests for `test_isolation_rule.go`                                 | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 86  | GSL-22 | Add tests for `testdata_directory_rule.go`                             | gsl        | `internal/rules/`                               | 3   | 2   | 2   | 2   | 6.0   | 5m  |

---

## Phase P3: POLISH — Medium Impact, Low Effort

| #   | ID     | Task                                                                   | Repo       | File(s)                            | I   | Im  | E   | CV  | Score | Est |
| --- | ------ | ---------------------------------------------------------------------- | ---------- | ---------------------------------- | --- | --- | --- | --- | ----- | --- |
| 87  | DOC2   | Update `CHANGELOG.md` with v0.1.0 release content                      | go-finding | `CHANGELOG.md`                     | 3   | 3   | 2   | 3   | 13.5  | 5m  |
| 88  | DOC3   | Update `AGENTS.md` with module split decision context                  | go-finding | `AGENTS.md`                        | 2   | 2   | 1   | 2   | 8.0   | 3m  |
| 89  | DOC4   | Improve `README.md` — add badges, pipeline examples, API overview      | go-finding | `README.md`                        | 3   | 4   | 3   | 4   | 16.0  | 8m  |
| 90  | DOC5   | Fix `.gitignore` line 43 corruption                                    | go-finding | `.gitignore`                       | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 91  | DOC6   | Create `examples/` directory with standalone examples                  | go-finding | `examples/`                        | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 92  | DOC7   | Add godoc examples for key APIs                                        | go-finding | multiple                           | 3   | 3   | 3   | 3   | 9.0   | 8m  |
| 93  | DOC8   | Add `version.go` with semver constants                                 | go-finding | `version.go`                       | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 94  | DOC9   | Set up pkg.go.dev documentation                                        | go-finding | —                                  | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 95  | DOC10  | Write migration guide for module split                                 | go-finding | `docs/`                            | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 96  | CODE1  | Extract `findingKey` to shared utility (dupe in verify.go + merge.go)  | go-finding | 2 files                            | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 97  | CODE2  | Extract SARIF constants (10 property keys) to `sarif_constants.go`     | go-finding | `sarif.go`                         | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 98  | CODE3  | Add `go:generate stringer` for enums                                   | go-finding | 4 files                            | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 99  | CODE4  | Modernize to Go 1.21+: `slices.Contains`, `slices.Delete`, `maps.Keys` | go-finding | multiple                           | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 100 | CODE5  | Clarify `HasFix` vs `HasSuggestion` boundary                           | go-finding | `finding.go:90-99`                 | 2   | 2   | 1   | 2   | 8.0   | 3m  |
| 101 | CODE6  | Add `iter.Seq[Finding]` on `Report.All()` for Go 1.26                  | go-finding | `report.go`                        | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 102 | CODE7  | Replace hardcoded temp dir in `pipeline.go`                            | go-finding | `pipeline/pipeline.go`             | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 103 | CODE8  | Suppress gosec G401 false positive for SHA-256 backup hash             | go-finding | `pipeline/fix_applier.go`          | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 104 | CODE9  | Disable `wsl_v5` + `nlreturn` in `.golangci.yml`                       | go-finding | `.golangci.yml`                    | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 105 | CODE10 | Fix `pipeline.go` err113 dynamic errors                                | go-finding | `pipeline/pipeline.go`             | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 106 | CODE11 | Fix `pipeline.go` ST1005 capitalized error strings                     | go-finding | `pipeline/pipeline.go`             | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 107 | CODE12 | Fix `json.go` err113 dynamic errors                                    | go-finding | `json.go`                          | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 108 | CODE13 | Update tests for sentinel error string changes                         | go-finding | `json_test.go`, `pipeline_test.go` | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 109 | CODE14 | Fix revive — rename unused receivers/params to `_` (4 locations)       | go-finding | multiple                           | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 110 | CODE15 | Clean up `pipeline/astfix.go` — evaluate if it belongs in `analysis/`  | go-finding | `pipeline/astfix.go`               | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 111 | CODE16 | Add `Range.Contains(p Position) bool`                                  | go-finding | `position.go`                      | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 112 | GSL-23 | Add `CONTRIBUTING.md` review/update                                    | gsl        | `CONTRIBUTING.md`                  | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 113 | GSL-24 | Fix `.gitignore` corruption in go-structure-linter                     | gsl        | `.gitignore`                       | 2   | 1   | 1   | 1   | 2.0   | 3m  |
| 114 | GSL-25 | Add `go.work` for local development                                    | gsl        | `go.work`                          | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 115 | GSL-26 | Add GitHub release workflow                                            | gsl        | `.github/workflows/`               | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 116 | GSL-27 | Add GoReleaser multi-module config                                     | gsl        | `.goreleaser.yml`                  | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 117 | GSL-28 | Add per-package coverage thresholds in CI                              | gsl        | `.github/workflows/ci.yml`         | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 118 | GSL-29 | Add `govulncheck` to CI                                                | gsl        | `.github/workflows/`               | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 119 | GSL-30 | Add gosec/staticcheck to CI linting                                    | gsl        | `.github/workflows/`               | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 120 | GSL-31 | Clear Go build cache, verify `-race` tests pass                        | gsl        | —                                  | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 121 | GSL-32 | Profile memory allocation hotspots                                     | gsl        | —                                  | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 122 | GSL-33 | Add benchmark regression tracking in CI                                | gsl        | `.github/workflows/`               | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 123 | GSL-34 | Add property-based tests for core operations                           | gsl        | `property_test.go`                 | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 124 | GSL-35 | Add benchmarks for hot paths                                           | gsl        | `bench_test.go`                    | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 125 | GSL-36 | Add fuzz tests for core types                                          | gsl        | `fuzz_test.go`                     | 2   | 1   | 3   | 1   | 0.7   | 8m  |

---

## Phase P4: ARCHITECTURE — Breaking Changes for v0.2.0

| #   | ID     | Task                                                                                   | Repo       | File(s)              | I   | Im  | E   | CV  | Score | Est |
| --- | ------ | -------------------------------------------------------------------------------------- | ---------- | -------------------- | --- | --- | --- | --- | ----- | --- |
| 126 | ARCH1  | Group `Finding` into embedded sub-structs (Identity, Position, Fix, Context, Metadata) | go-finding | `finding.go`         | 4   | 4   | 5   | 4   | 12.8  | 12m |
| 127 | ARCH2  | Migrate test helpers to `finding.NewBuilder`                                           | go-finding | `*_test.go`          | 3   | 3   | 5   | 3   | 5.4   | 12m |
| 128 | ARCH3  | Reduce `FindingsFromSARIF` cognitive complexity (90 → 35)                              | go-finding | `sarif.go`           | 3   | 3   | 4   | 3   | 6.8   | 10m |
| 129 | ARCH4  | Flatten `Range.Overlaps` nesting at `position.go:337`                                  | go-finding | `position.go`        | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 130 | ARCH5  | Error wrapping audit — ensure all internal errors use `%w`                             | go-finding | project-wide         | 3   | 3   | 4   | 3   | 6.8   | 10m |
| 131 | ARCH6  | Add `go.work` for go-finding local development                                         | go-finding | `go.work`            | 2   | 1   | 2   | 1   | 1.0   | 5m  |
| 132 | ARCH7  | Create `Project` abstraction in go-structure-linter                                    | gsl        | new file             | 4   | 4   | 5   | 4   | 12.8  | 12m |
| 133 | ARCH8  | Create `internal/adapters/finding_adapter.go` — Issue↔Finding conversion               | gsl        | new file             | 5   | 5   | 5   | 5   | 25.0  | 12m |
| 134 | ARCH9  | Add SARIF output via adapter in go-structure-linter                                    | gsl        | `internal/adapters/` | 4   | 4   | 5   | 5   | 16.0  | 12m |
| 135 | ARCH10 | Add LSP output via adapter in go-structure-linter                                      | gsl        | `internal/adapters/` | 4   | 4   | 5   | 4   | 12.8  | 12m |
| 136 | ARCH11 | Wrap rules as `pipeline.Detector` instances                                            | gsl        | `internal/rules/`    | 5   | 5   | 5   | 5   | 25.0  | 12m |
| 137 | ARCH12 | Replace `RuleService.RunAllRulesParallel` with `pipeline.New()`                        | gsl        | `internal/services/` | 5   | 5   | 5   | 5   | 25.0  | 12m |
| 138 | ARCH13 | Wire pipeline conflict detection, retry, metrics                                       | gsl        | `internal/services/` | 4   | 4   | 5   | 4   | 12.8  | 12m |
| 139 | ARCH14 | Remove `internal/cache/` (replaced by pipeline)                                        | gsl        | `internal/cache/`    | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 140 | ARCH15 | Remove `internal/progress/` (replaced by pipeline callbacks)                           | gsl        | `internal/progress/` | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 141 | ARCH16 | Add go-finding as dependency: `go get github.com/larsartmann/go-finding`               | gsl        | `go.mod`             | 5   | 5   | 1   | 5   | 125.0 | 2m  |
| 142 | ARCH17 | Full integration test: go-structure-linter → go-finding pipeline                       | gsl        | `tests/integration/` | 5   | 5   | 5   | 5   | 25.0  | 12m |

---

## Phase P5: NICE TO HAVE — Low Impact / Long-term

| #   | ID     | Task                                                                        | Repo       | File(s)                          | I   | Im  | E   | CV  | Score | Est |
| --- | ------ | --------------------------------------------------------------------------- | ---------- | -------------------------------- | --- | --- | --- | --- | ----- | --- |
| 143 | NICE1  | Decide: implement or remove `FixStrategyAI` phantom constant                | go-finding | `fix_strategy.go:19`             | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 144 | NICE2  | Implement AI backend for `FixStrategyAI`                                    | go-finding | new file                         | 2   | 1   | 5   | 1   | 0.4   | 12m |
| 145 | NICE3  | Decide on repository name: `finding` vs `finding-sdk` vs `go-finding`       | go-finding | —                                | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 146 | NICE4  | Decide on stable ID format: hashes vs readable strings                      | go-finding | —                                | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 147 | NICE5  | Decide on BuildFlow: replace or add `Finding` alongside                     | go-finding | —                                | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 148 | NICE6  | Decide on suppression expiry enforcement                                    | go-finding | —                                | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 149 | NICE7  | Decide on go-business-rules `Severity` sharing                              | go-finding | —                                | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 150 | NICE8  | API stability review — lock exported API before v1.0.0                      | go-finding | —                                | 3   | 3   | 3   | 2   | 6.0   | 8m  |
| 151 | NICE9  | Consider re-tagging v0.1.0 or tag v0.1.1                                    | go-finding | git                              | 2   | 2   | 1   | 1   | 4.0   | 3m  |
| 152 | NICE10 | Measure `golang.org/x/tools` transitive dep size                            | go-finding | —                                | 1   | 1   | 2   | 1   | 0.5   | 5m  |
| 153 | NICE11 | Watch mode for continuous analysis with fsnotify                            | go-finding | new file                         | 2   | 2   | 5   | 2   | 1.6   | 12m |
| 154 | NICE12 | IDE plugin stubs — VS Code                                                  | go-finding | `ide/vscode/`                    | 2   | 1   | 5   | 2   | 0.8   | 12m |
| 155 | NICE13 | Web UI prototype for pipeline monitoring                                    | go-finding | `web/`                           | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 156 | NICE14 | Distributed detection support                                               | go-finding | —                                | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 157 | NICE15 | Nix migration Phase 0: Install Nix with flakes                              | go-finding | —                                | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 158 | NICE16 | Nix migration Phase 1: Create `flake.nix`                                   | go-finding | `flake.nix`                      | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 159 | NICE17 | Nix migration Phase 2: Nix-based GitHub Actions                             | go-finding | `.github/workflows/`             | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 160 | NICE18 | Nix migration Phase 3: direnv `.envrc`, update docs                         | go-finding | multiple                         | 1   | 1   | 3   | 1   | 0.3   | 8m  |
| 161 | NICE19 | Nix migration Phase 5: Pin nixpkgs, add overlay                             | go-finding | `nix/`                           | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 162 | NICE20 | Verify Go 1.26 availability in nixpkgs                                      | go-finding | —                                | 1   | 1   | 2   | 1   | 0.5   | 5m  |
| 163 | NICE21 | Verify golangci-lint v2 in nixpkgs                                          | go-finding | —                                | 1   | 1   | 2   | 1   | 0.5   | 5m  |
| 164 | NICE22 | Add build tags for `goexperiment.*` in Nix derivation                       | go-finding | `flake.nix`                      | 1   | 1   | 3   | 1   | 0.3   | 8m  |
| 165 | NICE23 | Add `nix/overlay.nix` for consumers                                         | go-finding | `nix/overlay.nix`                | 1   | 1   | 3   | 1   | 0.3   | 8m  |
| 166 | NICE24 | Fix Nix Go 1.26.0 stdlib corruption                                         | go-finding | `/nix/store/...`                 | 1   | 1   | 5   | 1   | 0.2   | 12m |
| 167 | NICE25 | Fix flaky `TestProperty_IDRoundTrip` — Unicode seed control                 | go-finding | `property_test.go`               | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 168 | NICE26 | Clean stale files: archive old status docs, delete superseded planning docs | go-finding | `docs/`                          | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 169 | NICE27 | Fix `DeduplicateByPosition` key should NOT include Rule                     | go-finding | `merge.go:129-144`               | 3   | 2   | 2   | 2   | 6.0   | 5m  |
| 170 | NICE28 | Resolve module split `example_test.go` cross-module dependency              | go-finding | `example_test.go:254-288`        | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 171 | NICE29 | Add pipeline example with config file                                       | go-finding | `examples/`                      | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 172 | NICE30 | Add `go-finding` CLI end-to-end test with REAL detector                     | go-finding | `cmd/go-finding/`                | 3   | 3   | 4   | 3   | 6.8   | 10m |
| 173 | NICE31 | Build at least 1 real detector integration for CLI                          | go-finding | `cmd/go-finding/`                | 3   | 3   | 4   | 3   | 6.8   | 10m |
| 174 | NICE32 | Fix partial detection missing metrics recording                             | go-finding | `pipeline/partial.go:63-117`     | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 175 | NICE33 | Replace string-based fix application with position-aware approach           | go-finding | `pipeline/pipeline.go`           | 4   | 3   | 5   | 3   | 7.2   | 12m |
| 176 | NICE34 | Fix `Metrics.StageTiming` value receiver — copies mutex                     | go-finding | `pipeline/metrics.go`            | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 177 | NICE35 | Clean up gopls hints (~12 non-critical)                                     | go-finding | multiple                         | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 178 | NICE36 | Investigate FixApplier backup persistence across iterations                 | go-finding | `pipeline/fix_applier.go`        | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 179 | NICE37 | Fix profiling FD leak in CLI                                                | go-finding | `cmd/go-finding/main.go:133-149` | 3   | 2   | 3   | 2   | 4.0   | 8m  |
| 180 | NICE38 | Improve `setupProfiling` test coverage                                      | go-finding | `cmd/go-finding/main_test.go`    | 2   | 1   | 3   | 1   | 0.7   | 8m  |
| 181 | NICE39 | Fix CI Go version matrix → 1.26.0                                           | go-finding | `.github/workflows/ci.yml`       | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 182 | NICE40 | Fix release workflow Go version → 1.26.0                                    | go-finding | `.github/workflows/release.yml`  | 2   | 2   | 2   | 2   | 4.0   | 5m  |
| 183 | NICE41 | Prove pipeline concept with throwaway script                                | go-finding | `scripts/`                       | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 184 | NICE42 | Revise `MODULE_SPLIT_PLAN.md` addressing 9 gaps                             | go-finding | `docs/planning/`                 | 2   | 2   | 3   | 2   | 2.7   | 8m  |
| 185 | NICE43 | Add `NewReport()` constructor (already exists — verify)                     | go-finding | `report.go`                      | 1   | 1   | 1   | 1   | 1.0   | 3m  |
| 186 | NICE44 | Review `Correlate` O(n²) performance (already capped at 10K)                | go-finding | `merge.go`                       | 2   | 2   | 2   | 1   | 2.0   | 5m  |
| 187 | NICE45 | Add `Range.Contains(p Position) bool` (already exists — verify)             | go-finding | `position.go`                    | 1   | 1   | 1   | 1   | 1.0   | 3m  |

---

## Summary Statistics

| Phase            | Tasks   | Est. Time       | Cumulative |
| ---------------- | ------- | --------------- | ---------- |
| P0: BLOCKING     | 13      | ~85 min         | 85 min     |
| P1: QUICK WINS   | 27      | ~135 min        | 220 min    |
| P2: HIGH VALUE   | 46      | ~380 min        | 600 min    |
| P3: POLISH       | 39      | ~260 min        | 860 min    |
| P4: ARCHITECTURE | 17      | ~195 min        | 1055 min   |
| P5: NICE TO HAVE | 45      | ~470 min        | 1525 min   |
| **TOTAL**        | **187** | **~25.4 hours** | —          |

### By Repository

| Repo                | Tasks | Est. Time   |
| ------------------- | ----- | ----------- |
| go-finding          | 147   | ~20.5 hours |
| go-structure-linter | 40    | ~4.9 hours  |

### By Type

| Type                     | Count |
| ------------------------ | ----- |
| Bug fixes                | 45    |
| Tests                    | 32    |
| Documentation            | 15    |
| Cleanup/Refactoring      | 38    |
| Architecture/Integration | 30    |
| Research/Decisions       | 14    |
| CI/Infrastructure        | 13    |

---

_Generated: 2026-04-28 13:41 UTC_
