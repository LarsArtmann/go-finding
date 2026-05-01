# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-01

### Breaking

- **`NewFinding` signature changed** — Now takes 6 parameters (added `confidence float64` as last parameter). Confidence is clamped to `[0.0, 1.0]`.
  - Before: `NewFinding(rule, toolName, message string, severity Severity, pos Position) Finding`
  - After: `NewFinding(rule, toolName, message string, severity Severity, pos Position, confidence float64) Finding`
- **`Builder.Build()` returns `(Finding, error)`** — Previously panicked on invalid state. Now returns `ErrInvalidBuilder` sentinel error. Callers handling the old panic must now check error.
  - Added `Builder.MustBuild() Finding` for panic-on-error use cases (tests, examples).
- **`WithTags` parameter type changed** — From `...string` to `...Tag`. Use the new `Tag` type constants (`TagSecurity`, `TagPerformance`, etc.) or `Tag("custom")`.
- **`Finding.Tag string` deprecated** — Use `Finding.Tags []Tag` instead. `Tag` field still exists for backward compatibility but will be removed in v1.0.
- **`FixApplier` applies fixes in deterministic order** — Files are now sorted alphabetically before applying fixes. Previously, map iteration order was non-deterministic.

### Added

- **`Tag` type with standard constants** — `type Tag string` with `TagSecurity`, `TagPerformance`, `TagStyle`, `TagCorrectness`, `TagBug`, `TagDeprecated`, `TagDocumentation`, `TagComplexity`, `TagTest`, `TagBuild`.
- **`Finding.Tags []Tag` field** — Multi-tag classification support (JSON: `"tags"`).
- **`Finding.Validate() error`** — Comprehensive structural validation (ID, Rule, ToolName, Message, Severity, Position, FixStrategy, Confidence range).
- **`Finding.Preview() string`** — Returns unified-diff-style preview of `BeforeCode`/`AfterCode` fix changes.
- **`Finding.HasCategory() bool`** — Convenience check for non-empty category.
- **`Report.Filter(predicates ...FilterFunc) *Report`** — Returns new filtered report.
- **`Report.Map(fn func(Finding) Finding) *Report`** — Returns new report with transformed findings.
- **`Report.WriteJSON(w io.Writer) error`** — Streaming JSON output without buffer allocation.
- **`Finding.WriteJSON(w io.Writer) error`** — Single-finding streaming JSON output.
- **`Builder.MustBuild() Finding`** — Panics on invalid builder; convenience for tests and examples.
- **CLI `-output` flag** — Write output to file instead of stdout.
- **CLI `-version` flag** — Print version and exit.
- **`Pipeline.CorrelateFindings` config** — Optional post-detection correlation stage.
- **`PipelineResult.Correlations`** — Correlation results when correlation is enabled.
- **`finding.Version` constant** — Programmatic version checking via `finding.Version` (`"0.2.0"`).
- **`examples/` directory** — Standalone examples (`basic/`, `builder/`, `pipeline/`) with compile checks.
- **`docs/integration-guide.md`** — Real-world tool integration guide.
- **`docs/release-procedure.md`** — Release process documentation.
- **`docs/architecture-decisions.md`** — 5 documented architectural decisions.
- **`FEATURES.md`** — Comprehensive, honest feature inventory.

### Changed

- **`Merge()` performance** — 2.4x faster with pre-allocated slice and batch appending (408K → 173K ns/op for 1000 findings).
- **`Report.AddFinding` / `Report.AddFindings`** — Now thread-safe via `*sync.Mutex`.
- **`FixStrategyAI` in `HasFix()`** — Now treated as Suggest-equivalent (requires `AfterCode`).
- **`Correlation` JSON tags** — Changed to camelCase (`findingIds`).
- **`DeduplicateByPosition` key** — Now includes `ToolName` for cross-tool deduplication.
- **`Severity.Compare`** — Total ordering for invalid severities via string comparison tiebreaker.
- **`Finding.Equal`** — Uses `floatEq` with 1e-9 epsilon for float comparison.
- **`Range.LineCount()`** — Returns absolute span for inverted ranges.
- **SARIF metadata round-trip** — Preserves non-string values via `fmt.Sprintf("%v", v)`.
- **SARIF `FindingsFromSARIF`** — Decomposed from cognitive complexity 90 to thin loop with 4 extracted helpers.
- **`math/rand` → `math/rand/v2`** — In retry jitter.
- **`detectResult` ghost type eliminated** — Pipeline uses `PartialResult` directly.
- **`findingKey` extracted** — Now `Finding.Key()` method instead of duplicate helpers.

### Fixed

- **`OnFix` callback accuracy** — Fires only for actually-applied fixes, not skipped ones.
- **`Pipeline.OnFinding` data race** — Added `callbackMu sync.Mutex` for concurrent callback safety.
- **`Report` copylocks on JSON marshal** — Uses `*sync.Mutex` instead of value `sync.Mutex`.
- **`Correlate()` O(n²) hang** — Hard-limited at `maxCorrelations = 10000`.
- **`FilterInvalid` global mutable state** — Changed from `var` to function.
- **`Position.HasEnd()`** — Checks `End.Line > 0 || End.Offset >= 0` (offset 0 is valid).
- **`Adjacent()` fallback** — No longer falls back to offset-based when line info present.
- **`govet.go parsePosn`** — Uses `strconv.Atoi` with error checking.
- **`FixApplier.replaceNearestToLine`** — Finds closest occurrence instead of first match.
- **Backup path collisions** — Include nanosecond timestamp suffix.
- **`Report.FindByRule`** — Uses `ActiveFindings()` for consistency.
- **Global `log.SetFlags(0)` init** — Removed unexpected global logger side effect.
- **SARIF suggestion-only fixes** — Export as fix descriptions without replacements.
- **SARIF `BeforeCode` and `RelatedRef.FindingID` round-trip** — Now preserved in properties.
- **`hasLineRange` dead code** — Removed unreachable branch.
- **`FixApplier` concurrent backup paths** — Uses unique temp dir with `os.MkdirTemp`.
- **Flaky `TestProperty_IDRoundTrip`** — Seeded with deterministic `rand.NewSource(42)`.

### Testing

- Coverage: **root 99.6%**, **pipeline 98.0%**, **detectors 96.1%**, **CLI 95.4%**
- Fuzz tests for SARIF import (1.6M execs, zero panics) and merge
- Property-based tests for ID generation round-trip
- Comprehensive edge-case tests for FixEngine, FileBackup, FixApplier, RetryConfig, Verifier
- CI stress test (`-count=20`) with govulncheck

## [0.1.3] - 2026-04-19

### Breaking

- **`pipeline.New()` now returns `(*Pipeline, error)`** — Previously returned `*Pipeline` with no error. Now validates config via `Config.Validate()` and rejects invalid configs (negative iterations/timeout, invalid retry settings).

### Added

- **`PipelineResult.PartialErrors`** — `map[string]error` surfacing per-detector errors when `GracefulDegradation` is enabled. Callers can now inspect which detectors failed during partial success.
- **`PipelineResult.Metrics`** — `MetricsSnapshot` populated automatically after `Run()` completes, including `TotalDuration`, `StageDurations`, `DetectorTimes`, and `FindingsFound`.
- **CLI metrics output** — Metrics summary printed to stderr after each run (detect/fix counts, duration).
- **Tests for config validation, partial errors, metrics snapshot** — `TestNew_RejectsInvalidConfig`, `TestNew_ValidConfig_NoError`, `TestPipelineRun_PartialErrorsSurfaced`, `TestPipelineRun_MetricsInResult`.

### Changed

- **Test helpers moved to `_test.go`** — `testutil.go` merged into `testutil_test.go`; `suppression_test_util.go` renamed to `suppression_test_util_test.go`. No test-only code ships in production builds.
- **`FixStrategyAI` documented** as phantom/placeholder — not wired to any implementation.
- **`map[string]bool` → `map[string]struct{}`** in `sarif_test.go` for idiomatic Go sets.
- **Removed dead `detectorSpec.Args` field** from CLI.
- **Removed unused `//nolint:goconst` directive** in `finding_extra_test.go`.
- **Extracted "changed" sentinel to const** in clone test for goconst compliance.

### Fixed

- **Metrics snapshot defer ordering bug** — `TotalDuration` was always zero because snapshot was taken before deferred `SetEnd()`. Moved snapshot into deferred cleanup so it captures correct end time.

### Testing

- Coverage: 93.1% root, 87.1% pipeline, 71.6% detectors, 57.3% CLI
- All tests pass with `-race`, `go vet` clean

## [0.1.2] - 2026-04-19

### Changed

- **SARIF property key constants** — Extracted 10 hardcoded property strings into named constants (`sarifPropID`, etc.)
- **SARIF nil safety** — Fixed nil `Region` dereference panics in `applySarifPosition` and `findingFromSarResult` for malformed SARIF input
- **SortByPosition refactor** — Replaced hand-rolled three-way comparison with `Position.Compare`
- **map[string]bool → map[string]struct{}** — Idiomatic Go set in `pipeline.collectAllFindings`
- **Retry MaxDelay validation** — `MaxDelay == 0` with `BaseDelay > 0` now returns an error instead of busy-looping
- **Merge nil-safety** — Passing nil `*Report` in the reports slice no longer panics

### Fixed

- Hand-rolled `HasPrefix` check replaced with `strings.HasPrefix` in SARIF metadata parsing
- SARIF filtered results capacity hint for reduced allocations
- Missing test coverage for `NewFinding`, `SuppressionKind.IsValid`, `Severity.GTE/LTE`, `Finding.String()`, `Report.AddFindings`, `Merge` with nil reports
- Missing `staticcheckCategory` F-prefix test case

## [0.1.1] - 2026-04-19

### Changed

- **SARIF decomposition** — `FindingsFromSARIF` refactored from cognitive complexity 90 to thin loop with 4 extracted helpers
- **Sentinel error migration** — All validation errors use `errors.New` sentinels + `fmt.Errorf("%w")` wrapping (err113 compliance)
- **golangci-lint zero issues** — Full lint compliance across 80+ enabled linters
- **CI upgraded** — Multi-OS matrix (ubuntu + macos), dedicated coverage job with 75% enforcement, tag-triggered builds
- **Code formatting** — Applied golines across entire codebase

- **CLI integration tests** — Coverage from 24% to 59% (loadConfig, validate, profiling, output)
- **SARIF parse benchmark** — `BenchmarkFromSARIF` for the `FindingsFromSARIF` hot path
- **CONTRIBUTING.md** — Fixed Go version (1.26), added golangci-lint commands, expanded pre-submit checklist

### Fixed

- Missing `FixStrategyNone` case in `HasFix()` switch (exhaustive linter, potential silent bug)
- Indentation bug in `findingToSARIF` (line 221)
- Error wrapping in CLI `setupProfiling` (wrapcheck compliance)
- Unused test helper extraction and table-driven test modernization

## [0.1.0] - 2026-04-11

### Added

- **Core types** — `Finding`, `Severity`, `FixStrategy`, `Position`, `Range`, `Category`, `Suppression`, `Report`
  - `NewFinding()` constructor with auto-generated ID
  - `IsValid()`, `HasFix()`, `HasSuggestion()`, `IsSuppressed()`, `Clone()`
  - `Position`/`Range` with `Overlaps()`, `Intersection()`, `Adjacent()` geometric operations
- **Filtering** — `Filter`, `BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFile`, `ByRule`, `ByTool`, `ByFixStrategy`
- **Grouping** — `GroupBy`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`
- **Merging** — `Merge` with deduplication by ID, position, or rule; `Correlate` for cross-tool correlation
- **SARIF 2.1.0** — Full round-trip: `FindingsToSARIF` + `FindingsFromSARIF` with `ToSARIFFiltered`
- **LSP Diagnostic conversion** — `FromLSP`, `ToLSP`, `FromLSPRelated`
- **go/analysis integration** — `FromDiagnostic`, `AnalysisDiagnostic`
- **JSON serialization** — `FromJSON`, `ReportFromJSON`, `PrettyJSON`, `LineJSON` with dropped-finding counts
- **Structured errors** — `FindingError` with categories (Validation, Conflict, Apply, Verify, Pipeline)
- **ID generation** — Stable `tool:rule:file:line:col` format with FNV-1a 128-bit hash
- **Pipeline package** (`pipeline/`) — detect → triage → fix → verify loop
  - Parallel and sequential detector execution via `errgroup`
  - Fix conflict detection and resolution
  - AST-aware fix application with text fallback
  - Post-fix verification by re-running detectors
  - Metrics collection with timing, counts, and snapshots
  - Exponential backoff retry wrapper for flaky detectors
  - Partial success — continue with findings from successful detectors
  - `Config.Validate()` + `RetryConfig.Validate()` with sentinel errors
- **CLI tool** (`cmd/go-finding`) — JSON/YAML/SARIF/text output, pprof profiling, version via ldflags
- **Built-in detectors** (`internal/detectors/`) — govet + staticcheck JSON wrappers
- **CI/CD** — GitHub Actions with multi-OS matrix, coverage enforcement

### Testing

- 81.1% total coverage (root: 91.7%, pipeline: 84.0%, detectors: 71.6%, CLI: 24.1%)
- Fuzz tests for ID generation/parsing, merge, filter, dedup, correlate
- Property-based tests using `testing/quick` for filter, group, merge, ID round-trip
- Integration tests for backup/restore, graceful degradation, retry, verify

### Documentation

- `README.md` — project overview and quick start
- `CONTRIBUTING.md` — contribution guidelines and development setup
- `AGENTS.md` — AI assistant context for development
- `docs/USAGE_GUIDE.md` — comprehensive usage guide
- `cmd/go-finding/config.example.yaml` — sample configuration

### Dependencies

- `gopkg.in/yaml.v3` — YAML config support in CLI
- `golang.org/x/sync` — errgroup for parallel detection
- `golang.org/x/tools` — go/analysis framework integration
