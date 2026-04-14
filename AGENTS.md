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

| File              | Purpose                                     |
| ----------------- | ------------------------------------------- |
| `finding.go`      | Core Finding type                           |
| `severity.go`     | Severity enum (info/warning/error/critical) |
| `fix_strategy.go` | FixStrategy enum (none/suggest/direct/ai)   |
| `position.go`     | Position, Range types with Overlaps/Intersection/Adjacent |
| `report.go`       | Report container with summary               |
| `filter.go`       | Filtering and grouping utilities            |
| `merge.go`        | Report merging with deduplication + Correlate |
| `sarif.go`        | SARIF 2.1.0 output                          |
| `lsp.go`          | LSP Diagnostic conversion                   |
| `diagnostic.go`   | go/analysis integration                     |
| `errors.go`       | Structured error types (FindingError with categories) |
| `category.go`     | Category constants                          |
| `id.go`           | ID generation utilities                     |
| `json.go`         | JSON marshaling/unmarshaling                |
| `suppression.go`  | Suppression handling                        |

#### Pipeline Package

| File                       | Purpose                                          |
| -------------------------- | ------------------------------------------------ |
| `pipeline/pipeline.go`     | Pipeline orchestrator: detect → triage → fix → verify |
| `pipeline/conflict.go`     | Fix conflict detection and analysis              |
| `pipeline/astfix.go`       | AST-aware fix application (fallback to text)     |
| `pipeline/verify.go`       | Verification stage: re-run detectors, diff findings |
| `pipeline/metrics.go`      | Timing/count metrics collection with snapshots   |
| `pipeline/retry.go`        | Exponential backoff retry wrapper for detectors  |
| `pipeline/partial.go`      | Partial success: collect from failed detectors   |

#### Examples

| File                       | Purpose                                          |
| -------------------------- | ------------------------------------------------ |
| `examples/govet/main.go`   | Go vet JSON → Finding converter (Detector impl)  |

### Testing

```bash
just test        # Run tests
just bench       # Run benchmarks
just lint        # Run linter
```

### Dependencies

- `golang.org/x/tools` - go/analysis framework
- `golang.org/x/sync` - errgroup for parallel detection
- Standard library only (minimal deps)

### Design Principles

1. **Zero dependencies** for core types
2. **Immutable** - Findings are data, not state machines
3. **Lossless** - Conversions preserve all data
4. **Compatible** - Works with existing Go analysis tools
5. **Resilient** - Retry logic, partial success, nil-safe metrics

### Pipeline Features

- **Conflict detection** - Overlapping fixes are filtered before application
- **Verification** - Optional post-fix verification by re-running detectors
- **Metrics** - Optional timing/count collection with snapshot support
- **Retry** - Configurable exponential backoff for flaky detectors
- **Partial success** - Continue with findings from successful detectors
- **Parallel detection** - errgroup-based concurrent detector execution

### Future Work

See EXECUTION_PLAN_V2.md for roadmap. Remaining items: CLI tool (14), config files (15), watch mode (16), go-sarif evaluation (8).

---

_Assisted-by: Crush <crush@charm.land>_
