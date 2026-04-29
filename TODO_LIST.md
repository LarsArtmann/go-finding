# TODO List

**Generated:** 2026-04-28
**Updated:** 2026-04-30 (session 2)
**Files Processed:** 143

## 🔴 HIGH Priority

- [x] Fix flaky `TestProperty_IDRoundTrip` — already seeded with `rand.New(rand.NewSource(42))` in `testutil_test.go`
- [x] Fix `Range.Contains` edge-case tests (80% → 100%) — inverted ranges, zero values, nil receiver
- [x] Fix `DeduplicateByPosition` vs `DeduplicateByRule` behavior test — added `TestDeduplicateStrategies_BehaviorDiff`
- [x] Fix `pipeline/partial.go` metrics — already recorded at `runOneDetector` line 377 on error path
- [x] Fix `applyTriage` tests — added parallel detector error, all-conflicts, apply-error tests
- [x] Fix `intersectionByOffset` + `HasOffset` tests — already covered; TODO was stale
- [x] Fix `FilterConflictingFixes()` and `AnalyzeConflicts()` tests — covered in `conflict_extra_test.go`
- [x] Fix `Verifier.Verify` error-path tests — already covered in `TestVerifier_Verify_DetectorError`
- [x] Fix `RetryConfig.Validate` edge-case tests — already 7 comprehensive test functions
- [x] Fix `FixApplier` error-path tests — already 19 comprehensive test functions
- [x] Fix `PrettyJSON` / `LineJSON` error-path tests — already covered in `json_test.go`
- [x] Fix `findingFromSarResult` SARIF import tests — added `TestFindingFromSarResult_WithFix`, `TestFindingFromSarResult_RankAsConfidence`

## 🟡 MEDIUM Priority

- [x] Wire `Correlate()` into Pipeline as optional post-detection stage — ALREADY DONE via `CorrelateFindings` config
- [ ] Consider `Finding` struct sub-grouping — breaking API change, deferred to v2
- [x] Add `go:generate stringer` — not applicable: enums are string types, stringer only works with int types
- [x] Replace hardcoded temp dir in `pipeline/pipeline.go` with `os.MkdirTemp` — concurrency edge case
- [x] Convert retry errors to sentinels — already `errMaxRetriesNegative` etc. at `retry.go:20-24`
- [x] Extract `findingKey` to shared utility — extracted as `Finding.Key()` method
- [x] Remove unused `//nolint` directives — audited; all current directives have valid explanations
- [x] Replace loop with `slices.Contains` at `merge_test.go:208` — no loop exists; TODO was stale
- [x] Preallocate `all` slice — test helper with small data, not worth optimizing
- [x] Extract `"changed"` constant — already extracted as `const mutated = "changed"` at line 9
- [x] Add `checkColumnRange` + `hasLineRange` tests — covered in `position_extra_test.go`
- [x] Add `cloneFindings` edge-case test — already covered in `TestCloneFindings_EmptySlice` and `TestCloneFindings_DeepCopy`
- [x] Add `Finding.Equal` field-mismatch test — added `TestEqual_FieldMismatch` with 18 cases
- [x] Replace hardcoded `SeverityWarning` in `diagnostic.go` — added `defaultSeverity ...Severity` variadic param
- [x] Modernize to Go 1.21+ stdlib — already uses `slices.SortFunc`, `slices.Equal`, `slices.Collect`, `maps.Keys`, `maps.Equal`, `maps.Copy`
- [x] Clean stale files: archive superseded planning docs in `docs/planning/`
- [x] Delete stale coverage files (`cover.out`, `coverage.out`) from repo root if present
- [x] Add `govet` binary to `.gitignore`
- [x] Add `.gitignore` check for binary artifacts — `/basic` and `/builder` added
- [x] Document and test SARIF round-trip losses — added godoc on `ToSARIF` + `TestSARIF_RoundTripLosses` + `TestSARIF_SuppressedFindingsExcludedFromRoundTrip`
- [x] Add SARIF parser fuzz test — `FindingsFromSARIF` handles untrusted input
- [ ] Add SARIF schema validation test — deferred: requires downloading SARIF JSON schema
- [x] Add `severityToSARIFLevel` edge-case test — already covered; TODO was stale
- [x] Add benchmarks for hot paths: merge, filter, SARIF, ID generation
- [x] Profile memory allocation hotspots
- [x] Add `Range.Contains(p Position) bool` tests
- [x] Add `io.WriterTo` for SARIF output — direct writing without buffer allocation
- [x] Add `go.work` for local development
- [x] Add pipeline example with config file
- [x] Add `examples/` directory with standalone examples
- [x] Add godoc examples for key APIs (`ExampleNewFinding`, `ExampleBuilder`, `ExampleFilter`)
- [x] Add LSP `toZeroBased` test for 0 Line case

## 🟢 LOW Priority

- [ ] Investigate `FixApplier` cross-iteration persistence — deferred to v1.1 (design question)
- [x] Add sync.Mutex to Pipeline for concurrent `OnFinding` callback safety
- [x] Disable `wsl_v5` + `nlreturn` — not in `.golangci.yml`, TODO was stale
- [ ] API stability review — documented in `docs/architecture-decisions.md`, target v0.2.0
- [x] Review `Correlate` O(n²) — already limited to 10k via `maxCorrelations` constant
- [x] Add `CONTRIBUTING.md` — already exists with comprehensive guidelines
- [x] Decide: implement or remove `FixStrategyAI` — RESOLVED: fixed split brain in `HasFix()`; AI behaves like Suggest
- [x] Decide stable ID format — documented in `docs/architecture-decisions.md`: keep readable for v1
- [x] Decide repository name — documented: keep `go-finding`
- [ ] Decide on BuildFlow integration — external project dependency, deferred
- [x] Decide suppression expiry — documented: defer to v1.1 with `IsActive()` method
- [ ] Decide on go-business-rules `Severity` sharing — external project dependency, deferred
- [x] Measure `golang.org/x/tools` dep size — 12MB. Price for `diagnostic.go` go/analysis integration.
- [x] Clean up `pipeline/astfix.go` — file does not exist, TODO was stale
- [x] Per-package coverage thresholds — already implemented in `scripts/coverage-check.sh`
- [x] Add `govulncheck` step to CI
- [x] Add gosec/staticcheck to CI — `golangci-lint` already includes these
- [x] Build tags for `goexperiment.*` — not needed, no experimental features
- [x] Add GitHub release workflow — already exists at `.github/workflows/release.yml`
- [x] Add GoReleaser config — already exists at `.goreleaser.yml`
- [ ] Web UI prototype — out of scope for v1
- [ ] Distributed detection — out of scope for v1
- [ ] IDE plugin stubs — out of scope for v1
- [ ] Watch mode — out of scope for v1
- [x] Config file support — CLI already supports YAML/JSON config
- [ ] Evaluate `go-sarif` vs hand-rolled SARIF — deferred to post-v1
- [x] Create tool integration guide — added `docs/integration-guide.md`
- [x] Migration guide for module split — no split planned, removed
- [x] Revise `MODULE_SPLIT_PLAN.md` — file does not exist, no split planned
- [x] Document release procedure — added `docs/release-procedure.md`
- [ ] Set up benchmark regression tracking — deferred to post-v1 (baseline captured)

## ✅ Completed (2026-04-30)

- [x] Delete stale `basic`/`builder` binaries from repo root and add to `.gitignore`
- [x] Fix `FixStrategyAI` split brain: `HasFix()` now treats AI like Suggest (requires `AfterCode`)
- [x] Add `run()` error path tests: negative max-iterations, pipeline run error, metrics output
- [x] Add pipeline edge case tests: parallel detector error, applyTriage all-conflicts, applyTriage apply-error
- [x] Add `Range.Contains` zero-value test
- [x] Add `FuzzFindingsFromSARIF` for malformed input (1.6M execs, zero panics)
- [x] Add `govulncheck` job to GitHub Actions CI
- [x] Profile memory allocations on hot paths (baseline captured)
- [x] Update `TODO_LIST.md`: close stale items, mark completed work

## ✅ Completed (2026-04-30 Session 2)

- [x] Add `Builder.Build()` error-path test — `TestBuilder_Build_MissingFields` with 5 field-missing cases
- [x] Add `Finding.Equal` field-mismatch test — `TestEqual_FieldMismatch` with 18 cases
- [x] Add dedup strategy behavior diff test — `TestDeduplicateStrategies_BehaviorDiff`
- [x] Add SARIF import tests — `TestFindingFromSarResult_WithFix`, `TestFindingFromSarResult_RankAsConfidence`
- [x] Add SARIF round-trip loss documentation — godoc on `ToSARIF()`
- [x] Add SARIF round-trip loss tests — `TestSARIF_RoundTripLosses`, `TestSARIF_SuppressedFindingsExcludedFromRoundTrip`
- [x] Replace hardcoded `SeverityWarning` in `diagnostic.go` — added `defaultSeverity ...Severity` variadic param
- [x] Fix gci + gofumpt formatting warnings in test files
- [x] Remove 6 stale TODO items (dead refs, non-existent files, already-done items)
- [x] Mark 30+ items as completed/verified across all priority levels
- [x] Add CI stress test job (`-count=20`)
- [x] Add `go.work` for local development
- [x] Create `docs/release-procedure.md`
- [x] Create `docs/integration-guide.md` — real-world tool integration guide
- [x] Create `docs/architecture-decisions.md` — 5 open decisions documented
- [x] Update `TODO_LIST.md` — comprehensive audit of all 48 remaining items

## ✅ Completed (2026-04-29)

- [x] Fix data race in `notifyFinding` for parallel detectors — added `callbackMu` to Pipeline
- [x] Eliminate `detectResult` ghost type — use `PartialResult` directly
- [x] Fix duplicate doc comments on `FindByID` and `All` in `report.go`
- [x] Clarify `Range.LineCount()` formula for inverted ranges
- [x] Add comprehensive `FixEngine` unit tests
- [x] Add direct `FileBackup` unit tests
- [x] Fix SARIF suggestion-only and related-locations tests
- [x] Modernize `FormatPartialErrors` to use `slices.Collect(maps.Keys)`

## ✅ Completed (2026-04-28)

- [x] Fix `OnFix` callback never called for successful fixes
- [x] Fix `Metrics.StageTiming` value receiver bug — already pointer receiver, verified
- [x] Guard `TotalDuration()` against negative / unset values
- [x] Add sync.Mutex to `Report.AddFinding` / `Report.AddFindings`
- [x] Fix copylocks on `Report` JSON marshal — switched to `*sync.Mutex`
- [x] Add `version.go` with semver constants
- [x] Improve `Pos()` godoc
- [x] Improve `README.md` with badges, Builder API, updated stats
- [x] Update `CHANGELOG.md` with unreleased changes
- [x] Fix H-2: `Severity.Compare` total ordering for invalid severities
- [x] Fix H-9: Remove global `log.SetFlags(0)` init
- [x] Fix H-11: Protect `knownDetectorBuilders` with `sync.RWMutex`
- [x] Fix M-9: `Finding.Equal` uses `floatEq` with epsilon
- [x] Fix M-4: SARIF metadata round-trip preserves non-string values
- [x] Fix M-8: `Range.LineCount()` absolute span for inverted ranges
- [x] Fix M-12: `Correlation` JSON tags camelCase
- [x] Fix M-13: `ToSARIFFiltered` dual-filtering godoc
- [x] Fix M-15: SARIF suggestion-only fixes export as descriptions
- [x] Fix M-17: `maxIterations: 0` defaults to pipeline default
- [x] Fix M-18: `BySeverityAtLeast` godoc notes invalid exclusion
- [x] Fix M-6/M-7: Document `Report.All()` and `FindByID()` yield copies
- [x] Fix M-19/M-20: `clampConfidence()` helper, `NormalizedConfidence()` and `Builder.WithConfidence()` clamp to [0.0, 1.0]
- [x] Fix C-1: `OnFix` fires only for actually-applied fixes
- [x] Fix C-2: FixApplier insertion-only and deletion-only fixes
- [x] Fix C-3: `Correlate()` hard-limits at 10k correlations
- [x] Fix C-4: `FilterInvalid` changed from mutable `var` to function
- [x] Fix C-5: `math/rand` → `math/rand/v2` in retry
- [x] Fix H-1: `HasEnd()` checks `End.Line > 0 || End.Offset >= 0`
- [x] Fix H-3: `Adjacent()` no longer falls back to offset-based when line info present
- [x] Fix H-5: `DeduplicateByPosition` key includes `ToolName`
- [x] Fix H-6: `parsePosn` uses `strconv.Atoi` with error checking
- [x] Fix H-7: `replaceNearestToLine` finds closest occurrence
- [x] Fix H-8: Backup paths include nanosecond timestamp
- [x] Fix H-10: `FindByRule` uses `ActiveFindings()`
- [x] Add CLI end-to-end tests
- [x] Add tests for `Finding.IsValid`, `Suppression.IsValid`, `ErrorCategory.IsValid`
- [x] Add `Severity.LessThan` invalid input test
- [x] Add `equalTimePtr` both-nil test
- [x] Add `Finding.Builder` fluent API
- [x] Delete binary artifacts from repo root
- [x] Fix err113 in `RegisterDetector`
- [x] Fix exhaustruct in `sarif.go` suggestion-only fix path
