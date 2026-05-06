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
| `tag.go`          | Tag type with IsStandard/IsValid/String methods           |
| `category.go`     | Category constants                                        |
| `id.go`           | ID generation utilities                                   |
| `json.go`         | JSON marshaling/unmarshaling                              |
| `suppression.go`  | Suppression handling with IsActive convenience            |

#### Pipeline Package

| File                       | Purpose                                                                 |
| -------------------------- | ----------------------------------------------------------------------- |
| `pipeline/pipeline.go`     | Pipeline orchestrator: detect → triage → fix → verify                   |
| `pipeline/conflict.go`     | Fix conflict detection and analysis                                     |
| `pipeline/fix_edit.go`     | FixEdit type — byte-level edit operations (Offset, Length, Replacement) |
| `pipeline/fix_provider.go` | FixProvider interface + 3 default providers (Offset, Line, Substring)   |
| `pipeline/fix_engine.go`   | Byte-level FixEngine with provider delegation, descending-offset apply  |
| `pipeline/fix_applier.go`  | Filesystem fix application with backup/rollback, custom providers       |
| `pipeline/verify.go`       | Verification stage: re-run detectors, diff findings                     |
| `pipeline/metrics.go`      | Timing/count metrics collection with snapshots                          |
| `pipeline/retry.go`        | Exponential backoff retry wrapper for detectors                         |
| `pipeline/partial.go`      | Partial success: collect from failed detectors                          |

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
go test -race -count=1 ./...   # Full suite with race detector
golangci-lint run ./...         # Lint
```

### Dependencies

- `golang.org/x/tools` - go/analysis framework
- `golang.org/x/sync` - errgroup for parallel detection
- `go.yaml.in/yaml/v3` - YAML config file parsing (CLI only, canonical library)
- `github.com/onsi/ginkgo/v2` - BDD testing framework
- `github.com/onsi/gomega` - BDD test matchers

### Design Principles

1. **Minimal dependencies** — core types depend only on stdlib
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags
4. **Compatible** — Works with existing Go analysis tools
5. **Resilient** — Retry logic, partial success, nil-safe metrics
6. **FixApplier.Close()** — Cleans up temporary backup directories; callers should defer close
7. **Config.FixProviders** — Custom fix providers for domain-specific (AST-aware) transformations

### Pipeline Features

- **Byte-level FixEngine** — `[]byte` edit operations with `FixEdit{Offset, Length, Replacement}`, applied descending by offset with frontier boundary
- **FixProvider interface** — Composable provider chain: OffsetProvider (byte offsets), LineProvider (line→byte), SubstringProvider (fallback)
- **Custom provider registration** — `NewFixEngineWithProviders`, `NewFixApplierWithProviders`, `Config.FixProviders`
- **Conflict detection** — Overlapping fixes are filtered before application
- **Verification** — Optional post-fix verification by re-running detectors
- **Metrics** — Optional timing/count collection with snapshot support
- **Retry** — Configurable exponential backoff for flaky detectors
- **Partial success** — Continue with findings from successful detectors
- **Parallel detection** — errgroup-based concurrent detector execution
- **Config validation** — `pipeline.New()` rejects invalid configs, returns error
- **Partial error surfacing** — `PipelineResult.PartialErrors` exposes per-detector failures
- **Metrics snapshot in result** — `PipelineResult.Metrics` auto-populated after `Run()`
- **FindingProcessor** — Composable transforms run between detection and triage (`ProcessorFunc`, `NamedProcessorFunc`)
- **FixApplier.Close()** — Cleans up temporary backup directories

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
