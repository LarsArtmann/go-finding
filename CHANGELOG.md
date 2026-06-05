# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`RelatedRef.Range *Range`** — Span-based related locations with deep-copy in `Clone()`, validation in `Validate()` (rejects inverted ranges), and value equality via `equalRelated()`. Enables full geometric relationships between related findings.
- **`LSPDiagnosticTag` type and constants** — `Unnecessary = 1` and `Deprecated = 2` per LSP spec; `Tags []LSPDiagnosticTag` field on `LSPDiagnostic`.
- **LSP diagnostic tag round-trip** — `FromLSP()` preserves `DiagnosticTag` values in `Metadata["go-finding/lsp-diagnostic-tags"]` as comma-separated integers and reconstructs them on `ToLSP()`.
- **LSP related information range support** — `ToLSP()` emits proper `LSPRange` end positions from `rel.Range`; `FromLSP()` reconstructs `RelatedRef.Range` from LSP related info end positions.
- **SARIF `region.snippet` support** — `SarifRegion.Snippet` field enables native SARIF snippet round-trip. `findingRegion()` populates it on export; `applySarifPosition()` reads it back on import. The property bag `go-finding/snippet` remains as fallback.
- **SARIF related location range round-trip** — `sarifRelatedLocs()` writes `RelatedRef.Range` end coordinates into related location regions; `findingFromSarResult()` reconstructs `RelatedRef.Range` from `physicalLocation.region` end coordinates.
- **JSON Schema `relatedRef.range`** — `docs/schemas/finding.schema.json` now includes `range` property on the `relatedRef` definition referencing `#/$defs/range`.
- **Comprehensive `doc.go`** — Expanded from ~40% to full package documentation covering Diff, SARIF, LSP, Formatting, JSON, ID Generation, Pipeline, Fix Providers, Filtering, Merging, Correlation, Suppression, Error Handling, and Known Limitations.
- **`TriageFunc` config option** — Customizable triage function in `pipeline.Config`. `DefaultTriageFunc` preserves existing behavior; `TriageResult` moved to `config.go` for discovery.
- **`ByteLevelConflictDetection` config option** — Opt-in byte-level edit conflict detection that groups fixes by file, reads content, runs `FilterConflictingEdits` per file, and gracefully degrades on read errors.

### Fixed

- **`Equal()` for `RelatedRef`** — Previously used `slices.Equal()` which compared `*Range` pointers by identity. Now uses `equalRelated()` for proper deep value comparison.
- **`Clone()` for `RelatedRef`** — Previously did shallow `copy()` of the `Related` slice, sharing `*Range` pointers between original and clone. Now deep-copies each `RelatedRef` and its `Range` pointer.

### Testing

- Coverage: root 97.1%, analysis 98.5%, pipeline 94.0%, internal/detectors 95.9%, cmd/go-finding 70.0%. Total 91.3%.
- 9 new tests: `TestToLSP_RelatedWithRange`, `TestFromLSP_RelatedWithRange`, `TestFromLSP_PreservesDiagnosticTags`, `TestSARIF_RoundTrip_RelatedRefRange`, `TestSARIF_RegionSnippet`, `TestFinding_Equal_RelatedRefRange`, `TestFinding_Validate_InvertedRelatedRange`, plus `TestClone` updated and schema round-trip extended.
- `go test -race -count=1 ./...` passes; `golangci-lint run ./...` reports 0 issues.

## [0.4.2] - 2026-06-01

### Added

- **`Summary.FilesScanned`** — New field on `Summary` counting total files scanned (including clean files with no findings). Complements existing `FilesAffected` which only counts files with findings. Tools can now display "X files scanned" in success paths where `FilesAffected` is 0.
- **`Category.IsValid()` rune-level validation** — Enforces `^[a-z][a-z0-9-]*$` format at the rune level, rejecting invalid categories like `"Security"`, `"UPPERCASE"`, or `"has space"`. Uses De Morgan's law for clarity (per staticcheck QF1001).

### Fixed

- **`ErrorCategory.IsValid()` format validation** — Now enforces the same lowercase-hyphenated convention as `Category.IsValid()`. Rejects typos like `"Validation"`, `"HAS_SPACE"`, `"has space"`.
- **SARIF import FixStrategy assignment** — Results with actual code replacements (`InsertedText`) were incorrectly assigned `FixStrategySuggest`, preventing pipeline auto-application. Now: replacements → `FixStrategyDirect`, description-only → `FixStrategySuggest`.
- **`FileBackup` permission preservation** — `Restore()` now uses the original file's `Mode()` instead of hardcoded `0o600`. Previously, restoring an executable (`0755`) would strip its execute bits.
- **CLI config duration parse errors** — `toPipelineConfig()` now returns `(Config, error)` and reports malformed duration strings (e.g., `"abc"`) instead of silently falling back to defaults.
- **`Config.DetectorTimeouts` negative validation** — `Config.Validate()` now rejects negative per-detector timeouts, which would previously panic at runtime via `context.WithTimeout`.
- **CLI redundant timeout wrapping** — Removed double `context.WithTimeout` application. Pipeline already wraps `ctx` internally; the CLI's second wrap with the same value was redundant and misleading.

### Changed

- **Upgraded `gogenfilter/v3`** — Multiple dependency bumps to pre-release versions with critical fixes for generated-file detection edge cases.
- **Nix flake overhaul** — Full `nix-review` improvements including `buildGoModule` with fileset source filtering, E2E sandbox fix, and added `lake.nix` configuration.
- **Code quality improvements** — `byFindingID` uses `cmp.Compare` instead of manual three-way comparison; `FixApplier` control flow made more explicit; `gci` formatting fixed in `pipeline/config.go`.
- **Deprecated `justfile` removed** — Build automation fully migrated to Nix flakes. `AGENTS.md` and `CONTRIBUTING.md` updated to reference `nix` commands.

## [0.4.1] - 2026-05-27

### Added

- **`GeneratedFileFilter` pipeline processor** — `pipeline.GeneratedFileFilter` wraps `gogenfilter/v3` to automatically remove findings from auto-generated Go source files (sqlc, protobuf, mockgen, templ, wire, stringer, deepcopy-gen, oapi-codegen, moq, go-enum, and generic `// Code generated by` comments). Configurable per-generator type with include/exclude glob patterns. Gracefully keeps findings when source files can't be read.
- **`gogenfilter/v3` dependency** — `github.com/LarsArtmann/gogenfilter/v3 v3.0.2` for two-phase generated file detection (filename-first zero-I/O, then content-based).
- **CLI `-filter-generated` flag** — Enables generated file filtering in the pipeline.
- **CLI `-filter-generated-types` flag** — Comma-separated generator types to filter (default: `all`). Supported: `all`, `sqlc`, `templ`, `go-enum`, `protobuf`, `oapi-codegen`, `deepcopy-gen`, `wire`, `moq`, `mockgen`, `stringer`, `generic`.
- **CLI `-generated-exclude` flag** — Comma-separated glob patterns for files to always exclude.
- **CLI `-generated-include` flag** — Comma-separated glob patterns restricting filter scope.
- **Config file `filterGenerated` field** — YAML/JSON config file support for generated file filtering.
- **Config file `filterGenTypes` field** — Generator types in config file (default: `"all"`).
- **Config file `generatedExclude` field** — Exclude patterns in config file.
- **Config file `generatedInclude` field** — Include patterns in config file.

### Changed

- **CLI `run()` refactored** — Extracted `cliFlags` struct, `parseFlags()`, and `writeResults()` functions. Resolves pre-existing `funlen` lint violation (148 → <120 lines).
- **`.golangci.yml` depguard** — Added `github.com/LarsArtmann/gogenfilter` to allow-lists.

## [0.4.0] - 2026-05-27

### Fixed

- **golangci-lint issues resolved (7 → 0)** — Fixed contextcheck, exhaustruct, revive, and other lint warnings across pipeline, analysis, and CLI packages.
- **`FormatPartialErrors` error wrapping** — Changed from `%v` with `fmt.Sprintf` to `%s` with `strings.Join` for cleaner error messages.
- **`.golangci.yml` config** — Replaced non-existent `gomodguard_v2` with `gomodguard`.

### Changed

- **Named constants for magic strings** — Extracted SARIF property keys (`go-finding/edit/*`), LSP severity key, conflict reason strings, merge tool names, analysis relation constant, and parseSeverity to named constants. All previously hardcoded strings are now exported package-level constants.
- **`ErrPositionUnresolvable` sentinel error** — Replaced 4 `//nolint:nilerr` directives in `fix_provider.go` with an explicit sentinel error. Callers already skip on error, so behavior is unchanged but now type-safe and lint-clean.
- **`parseSeverity` refactored** — Switch/case replaced with map-based lookup using `Severity.String()`.

### Added

- **Per-detector timeouts** — `DetectorTimeouts` config for individual detector timeout control.
- **`FormatText` / `FormatMarkdown`** — Human-readable and markdown table output formatters.
- **`Finding.ToDiagnostic()`** — Reverse conversion from Finding to `analysis.Diagnostic`.
- **`slog` logging integration** — Pipeline now accepts `*slog.Logger` for structured logging.
- **`OnStage` callback** — Pipeline lifecycle callback for monitoring stage transitions.
- **`FixApplier` lifecycle** — `Start()` and `Stop()` methods for provider lifecycle management.
- **Godoc examples** — Added `ExampleDiff`, `ExampleFormatText`, `ExampleFormatMarkdown`.

## [0.2.1] - 2026-05-01

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
- **`finding.Version` constant** — Programmatic version checking via `finding.Version` (`"0.2.1"`).
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
