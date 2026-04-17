# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Pipeline package** (`pipeline/`) — detect → triage → fix → verify loop
  - Parallel and sequential detector execution via `errgroup`
  - Fix conflict detection and resolution
  - AST-aware fix application with text fallback
  - Post-fix verification by re-running detectors
  - Metrics collection with timing, counts, and snapshots
  - Exponential backoff retry wrapper for flaky detectors
  - Partial success — continue with findings from successful detectors
- **CLI tool** (`cmd/go-finding`) with flags:
  - `-dir`, `-format` (text/json/sarif), `-severity`, `-max-iterations`
  - `-parallel`, `-verify`, `-timeout`
  - `-config` for YAML/JSON config file support
  - `-cpuprof`/`-memprof` for `runtime/pprof` profiling
- **Built-in detectors** (`internal/detectors/`)
  - `govet` — Go vet JSON → Finding converter
  - `staticcheck` — staticcheck JSON → Finding converter
- **YAML config support** — config files accept `.yaml`/`.yml` and `.json`
- **Structured errors** — `FindingError` with categories (validation, IO, parse, conflict, internal) and position context
- **ID generation** — stable `tool:rule:file:line:col` format with SHA-256 hash-based fallback
- **Correlation** — heuristic cross-tool finding correlation (same file, nearby lines)
- **Retry** — `RetryConfig` with exponential backoff and jitter
- **Partial success** — `DetectPartial` for graceful degradation
- **Metrics** — `Metrics` type with stage/detector timing and `MetricsSnapshot`

### Testing

- Coverage tests for `IsValid`, `HasFix`, `HasSuggestion`, `IsSuppressed`
- Fuzz tests for ID generation/parsing, merge, filter, dedup, correlate
- Property-based tests using `testing/quick` for filter, group, merge, ID round-trip
- Benchmarks for ID generation, filtering, grouping, merge, correlate, SARIF
- GoDoc examples (`Example*()`) for all major public APIs

### Documentation

- `docs/USAGE_GUIDE.md` — comprehensive usage guide
- `CONTRIBUTING.md` — contribution guidelines and development setup
- `cmd/go-finding/config.example.yaml` — sample configuration

### Dependencies

- `gopkg.in/yaml.v3` — YAML config support in CLI

## [0.1.0] - 2026-04-11

### Added

- Core types: `Finding`, `Severity`, `FixStrategy`, `Position`, `Range`, `Category`, `Suppression`, `Report`
- Filtering: `Filter`, `BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFile`, `ByRule`, `ByTool`, `ByFixStrategy`
- Grouping: `GroupBy`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`
- Merging: `Merge` with deduplication by ID, position, or rule
- SARIF 2.1.0 output: `ToSARIF`, `ToSARIFFiltered`
- LSP Diagnostic conversion: `FromLSP`, `ToLSP`
- go/analysis integration: `FromDiagnostic`, `AnalysisDiagnostic`
- JSON serialization: `FromJSON`, `FindingsFromJSON`, `LineJSON`
- Initial test suite
