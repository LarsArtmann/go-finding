# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools.

### Core Purpose

Seven tools detect issues. Zero tools route them to remediation. This library solves that by providing:

1. **Unified Finding type** - Common representation for all tools
2. **Pipeline** - Automated detect → triage → fix → verify loop
3. **SARIF output** - Standard interchange format
4. **LSP integration** - IDE support

### Key Files

#### Core Types (root package)

| File              | Purpose                                                   |
| ----------------- | --------------------------------------------------------- |
| `finding.go`      | Core Finding type                                         |
| `severity.go`     | Severity enum (info/warning/error/critical)               |
| `fix_strategy.go` | FixStrategy enum (none/suggest/direct/ai)                 |
| `position.go`     | Position, Range types with Overlaps/Intersection/Adjacent |
| `report.go`       | Report container with summary                             |
| `filter.go`       | Filtering and grouping utilities                          |
| `merge.go`        | Report merging with deduplication + Correlate             |
| `sarif.go`        | SARIF 2.1.0 output                                        |
| `lsp.go`          | LSP Diagnostic conversion                                 |
| `diagnostic.go`   | go/analysis integration                                   |
| `errors.go`       | Structured error types (FindingError with categories)     |
| `category.go`     | Category constants                                        |
| `id.go`           | ID generation utilities                                   |
| `json.go`         | JSON marshaling/unmarshaling                              |
| `suppression.go`  | Suppression handling                                      |

#### Pipeline Package

| File                   | Purpose                                               |
| ---------------------- | ----------------------------------------------------- |
| `pipeline/pipeline.go` | Pipeline orchestrator: detect → triage → fix → verify |
| `pipeline/conflict.go` | Fix conflict detection and analysis                   |
| `pipeline/fix_applier.go` | Line-based fix application with backup/rollback |
| `pipeline/verify.go`   | Verification stage: re-run detectors, diff findings   |
| `pipeline/metrics.go`  | Timing/count metrics collection with snapshots        |
| `pipeline/retry.go`    | Exponential backoff retry wrapper for detectors       |
| `pipeline/partial.go`  | Partial success: collect from failed detectors        |

#### CLI

| File                     | Purpose                                                                                                      |
| ------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `cmd/go-finding/main.go` | Functional CLI: govet+staticcheck detectors, pipeline integration, text/json/sarif output, config validation |

#### Internal Detectors

| File                                | Purpose                                              |
| ----------------------------------- | ---------------------------------------------------- |
| `internal/detectors/govet.go`       | Go vet JSON → Finding converter (Detector impl)      |
| `internal/detectors/staticcheck.go` | Staticcheck JSON → Finding converter (Detector impl) |

### Testing

```bash
just test        # Run tests
just bench       # Run benchmarks
just lint        # Run linter
```

### Dependencies

- `golang.org/x/tools` - go/analysis framework
- `golang.org/x/sync` - errgroup for parallel detection
- `gopkg.in/yaml.v3` - YAML config file parsing (CLI only)

### Design Principles

1. **Minimal dependencies** — core types depend only on stdlib
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata
4. **Compatible** — Works with existing Go analysis tools
5. **Resilient** — Retry logic, partial success, nil-safe metrics

### Pipeline Features

- **Conflict detection** — Overlapping fixes are filtered before application
- **Verification** — Optional post-fix verification by re-running detectors
- **Metrics** — Optional timing/count collection with snapshot support
- **Retry** — Configurable exponential backoff for flaky detectors
- **Partial success** — Continue with findings from successful detectors
- **Parallel detection** — errgroup-based concurrent detector execution
- **Config validation** — `pipeline.New()` rejects invalid configs, returns error
- **Partial error surfacing** — `PipelineResult.PartialErrors` exposes per-detector failures
- **Metrics snapshot in result** — `PipelineResult.Metrics` auto-populated after `Run()`
- **Line-based FixApplier** — Range-aware fixes target exact line spans; falls back to string replacement

### CLI Features

- Built-in govet and staticcheck detectors
- Text, JSON, and SARIF output formats
- YAML/JSON config file support with validation
- Severity filtering, timeout, max-iterations
- CPU/memory profiling
- Graceful degradation on detector failures
- Metrics summary output to stderr

---

_Assisted-by: Crush <crush@charm.land>_
