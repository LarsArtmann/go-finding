# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`finding.Builder`** — Fluent API for constructing Finding values with 13 chainable `With*` methods (`WithID`, `WithCategory`, `WithFixStrategy`, `WithBeforeCode`, `WithAfterCode`, `WithRange`, `WithSnippet`, `WithConfidence`, `WithRelated`, `WithSuppression`, `WithMetadata`, etc.). `WithConfidence` clamps to `[0.0, 1.0]`.
- **`version.go`** — Programmatic semver constants (`VersionMajor`, `VersionMinor`, `VersionPatch`, `Version`).
- **CLI end-to-end tests** — `TestRun_E2E_DefaultDetectors`, `TestRun_E2E_ConfigFile`, `TestRun_E2E_SARIFOutput` build the binary and exercise it as a subprocess.
- **Tests for uncovered functions** — `Finding.IsValid`, `Suppression.IsValid`, `ErrorCategory.IsValid`, `Severity.LessThan` invalid input, `equalTimePtr` both-nil, and more.
- **`RegisterDetector`** — Thread-safe detector registration with `sync.RWMutex` protection.

### Changed

- **`Report.AddFinding` / `Report.AddFindings`** — Now safe for concurrent use via `*sync.Mutex` (initialized in `NewReport`, nil-safe for zero-value Reports).
- **`Metrics.TotalDuration`** — Guards against both `endTime.IsZero()` and `startTime.IsZero()`.
- **`Pos()` godoc** — Improved to clarify it is a convenience constructor.
- **`maxIterations: 0`** in CLI config now defaults to `pipeline.DefaultConfig().MaxIterations` (5) instead of hardcoded 1.
- **`Correlation` JSON tags** — Changed from snake_case (`finding_ids`) to camelCase (`findingIds`).
- **`Range.LineCount()`** — Returns absolute span for inverted ranges instead of 0.
- **`Severity.Compare`** — Uses string comparison as tiebreaker for two different invalid severities, ensuring total ordering.
- **`Finding.Equal`** — Replaces direct `float64` equality with `floatEq` using 1e-9 epsilon.
- **`DeduplicateByPosition`** key now includes `ToolName` for cross-tool deduplication.
- **`ToSARIFFiltered`** godoc explicitly documents dual filtering (suppression + severity).
- **`SARIF metadata round-trip`** — Preserves non-string metadata values via `fmt.Sprintf("%v", v)`.
- **`Report.All()` and `Report.FindByID()`** godoc now explicitly states they yield copies.

### Fixed

- **`OnFix` callback** — Now fires only for actually-applied fixes, not for skipped/unapplied ones.
- **FixApplier insertion-only and deletion-only fixes** — Previously unsupported; now handled correctly.
- **`Correlate()` O(n²) hang** — Hard-limits at `maxCorrelations = 10000` to prevent unbounded execution.
- **`FilterInvalid`** — Changed from mutable package-level `var` to an exported function, eliminating a global mutable state bug.
- **`RetryConfig` jitter** — Switched from `math/rand` to `math/rand/v2` (`rand.Int64N`) for proper randomness.
- **`Position.HasEnd()`** — Now checks `End.Line > 0 || End.Offset >= 0` (offset 0 is valid).
- **`Adjacent()`** — No longer falls back to offset-based adjacency when line info is present.
- **`govet.go` `parsePosn`** — Uses `strconv.Atoi` with proper error checking instead of unchecked conversion.
- **`FixApplier.replaceNearestToLine`** — Finds occurrence closest to finding's line instead of first match.
- **Backup paths** — Include nanosecond timestamp suffix to prevent collisions.
- **`Report.FindByRule`** — Uses `ActiveFindings()` for consistency with other query methods.
- **Removed global `log.SetFlags(0)` `init()`** — Eliminated unexpected global logger side effect.
- **SARIF export** — Includes suggestion-only fixes as fix descriptions without replacements.
- **Copylocks on JSON marshal** — `Report` uses `*sync.Mutex` to avoid `sync.Mutex` copy when passed to `json.Marshal`.

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
