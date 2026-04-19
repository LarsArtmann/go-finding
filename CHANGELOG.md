# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-04-19

### Changed

- **SARIF decomposition** — `FindingsFromSARIF` refactored from cognitive complexity 90 to thin loop with 4 extracted helpers
- **Sentinel error migration** — All validation errors use `errors.New` sentinels + `fmt.Errorf("%w")` wrapping (err113 compliance)
- **golangci-lint zero issues** — Full lint compliance across 80+ enabled linters
- **CI upgraded** — Multi-OS matrix (ubuntu + macos), dedicated coverage job with 75% enforcement, tag-triggered builds
- **Code formatting** — Applied golines across entire codebase

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
