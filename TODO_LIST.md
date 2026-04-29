# TODO List

**Generated:** 2026-04-28
**Updated:** 2026-04-30
**Files Processed:** 143

## 🔴 HIGH Priority

- [ ] Fix flaky `TestProperty_IDRoundTrip` — fails ~5% with random Unicode; add seed control or input constraints
- [x] Fix `Range.Contains` edge-case tests (80% → 100%) — inverted ranges, zero values, nil receiver
- [ ] Fix `DeduplicateByPosition` vs `DeduplicateByRule` behavior test — ensure they actually differ
- [ ] Fix `pipeline/partial.go` missing metrics recording during partial detection failures
- [x] Fix `applyTriage` tests — added parallel detector error, all-conflicts, apply-error tests
- [x] Fix `intersectionByOffset` + `HasOffset` tests — already covered; TODO was stale
- [x] Fix `FilterConflictingFixes()` and `AnalyzeConflicts()` tests — covered in `conflict_extra_test.go`
- [ ] Fix `Verifier.Verify` error-path tests
- [ ] Fix `RetryConfig.Validate` edge-case tests
- [ ] Fix `FixApplier` error-path unit tests — push coverage from 87% to 90%+
- [x] Fix `PrettyJSON` / `LineJSON` error-path tests — already covered in `json_test.go`
- [ ] Fix `findingFromSarResult` SARIF import path tests — rule metadata, help URI, markdown descriptions

## 🟡 MEDIUM Priority

- [x] Wire `Correlate()` into Pipeline as optional post-detection stage — ALREADY DONE via `CorrelateFindings` config
- [ ] Consider `Finding` struct sub-grouping into embedded sub-structs (`Location`, `Content`, `Fix`, `Metadata`) — breaking API change
- [ ] Add `go:generate stringer` for `Severity`, `FixStrategy`, `Category`, `SuppressionKind` — deferred: existing `String()` methods are sufficient
- [x] Replace hardcoded temp dir in `pipeline/pipeline.go` with `os.MkdirTemp` — concurrency edge case
- [ ] Convert 4 `errors.New()` calls in `pipeline/retry.go` to sentinel errors
- [x] Extract `findingKey` to shared utility — extracted as `Finding.Key()` method
- [ ] Remove 3 unused `//nolint` directives (lsp_test.go:217, conflict.go:137, pipeline.go:678)
- [x] Replace loop with `slices.Contains` at `merge_test.go:208` — no loop exists; TODO was stale
- [ ] Preallocate `all` slice in `pipeline_test.go:332`
- [ ] Extract `"changed"` string to constant at `finding_extra_test.go:46`
- [x] Add `checkColumnRange` + `hasLineRange` tests — covered in `position_extra_test.go`
- [ ] Add `cloneFindings` edge-case test
- [ ] Add `Finding.Equal` field-mismatch test
- [ ] Replace hardcoded `SeverityWarning` in `diagnostic.go` with configurable default
- [ ] Modernize to Go 1.21+ standard library: `slices.Contains`, `slices.Delete`, `maps.Keys`, `maps.Values`
- [x] Clean stale files: archive superseded planning docs in `docs/planning/`
- [x] Delete stale coverage files (`cover.out`, `coverage.out`) from repo root if present
- [x] Add `govet` binary to `.gitignore`
- [x] Add `.gitignore` check for binary artifacts — `/basic` and `/builder` added
- [ ] Document and test SARIF round-trip losses: `RelatedRef.FindingID` lost, `BeforeCode` lost
- [x] Add SARIF parser fuzz test — `FindingsFromSARIF` handles untrusted input
- [ ] Add SARIF schema validation test
- [x] Add `severityToSARIFLevel` edge-case test — already covered; TODO was stale
- [x] Add benchmarks for hot paths: merge, filter, SARIF, ID generation
- [x] Profile memory allocation hotspots
- [x] Add `Range.Contains(p Position) bool` tests
- [x] Add `io.WriterTo` for SARIF output — direct writing without buffer allocation
- [ ] Add `go.work` for local development
- [x] Add pipeline example with config file
- [x] Add `examples/` directory with standalone examples
- [x] Add godoc examples for key APIs (`ExampleNewFinding`, `ExampleBuilder`, `ExampleFilter`)
- [x] Add LSP `toZeroBased` test for 0 Line case

## 🟢 LOW Priority

- [ ] Investigate `FixApplier` not persisting across pipeline iterations — backups from iteration N don't persist to N+1
- [x] Add sync.Mutex to Pipeline for concurrent `OnFinding` callback safety
- [ ] Disable `wsl_v5` + `nlreturn` in `.golangci.yml` — 9 pedantic style issues
- [ ] API stability review — lock exported API before v1.0.0
- [ ] Review `Correlate` O(n²) performance (already limited to 10k correlations)
- [ ] Add contribution guidelines review (`CONTRIBUTING.md`)
- [x] Decide: implement or remove `FixStrategyAI` — RESOLVED: fixed split brain in `HasFix()`; AI behaves like Suggest
- [ ] Decide on stable ID format: deterministic hashes or readable strings
- [ ] Decide on repository name: `finding`, `finding-sdk`, or `go-finding`
- [ ] Decide on BuildFlow: replace `PrioritizedViolation` or add `Finding` alongside
- [ ] Decide on suppression expiry — include field but start without enforcement
- [ ] Decide on go-business-rules `Severity` sharing — extract to shared package
- [ ] Measure `golang.org/x/tools` transitive dep size
- [ ] Clean up `pipeline/astfix.go` — evaluate if it belongs in `analysis/`
- [ ] Per-package coverage thresholds in CI, not just total 75%
- [x] Add `govulncheck` step to CI
- [ ] Add gosec/staticcheck to CI linting
- [ ] Add build tags for `goexperiment.*` if needed
- [ ] Add GitHub release workflow
- [ ] Add GoReleaser multi-module config if splitting
- [ ] Web UI prototype for pipeline monitoring
- [ ] Distributed detection support
- [ ] IDE plugin stubs — VS Code
- [ ] Watch mode for continuous analysis with fsnotify
- [ ] Config file support for library/pipeline (YAML)
- [ ] Evaluate `go-sarif` library vs hand-rolled SARIF code for schema compliance
- [ ] Create real-world tool integration guide
- [ ] Write migration guide for module split with before/after code
- [ ] Revise `MODULE_SPLIT_PLAN.md` addressing all 9 identified gaps
- [ ] Document first-release procedure for multi-module
- [ ] Set up benchmark regression tracking in CI

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
