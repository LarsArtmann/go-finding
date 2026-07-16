# Architecture Review — go-finding

**Date:** 2026-05-06

## Scalability Assessment

### Current Architecture

```
finding (root package)
├── Core Types: Finding, Position, Range, Severity, FixStrategy, Category, Tag, Suppression
├── Container: Report (mutex-protected)
├── Conversions: SARIF, LSP, go/analysis
├── Utilities: Filter, Merge, ID, JSON, Errors
└── Pipeline (subpackage)
    ├── Orchestrator: Pipeline (detect→triage→fix→verify loop)
    ├── Fix Application: FixEngine (in-memory), FixApplier (filesystem)
    ├── Support: ConflictDetector, Verifier, Metrics, Retry, PartialSuccess, FileBackup
    └── Adapters: DetectorFunc, NamedDetectorFunc, ProcessorFunc

cmd/go-finding (CLI)
└── Uses pipeline + internal/detectors (govet, staticcheck)
```

### Scalability Strengths

1. **Pipeline is detector-agnostic** — Any `Detector` interface impl works. Adding new tools = 1 function.
2. **Report is thread-safe** — Concurrent AddFinding with mutex protection.
3. **Graceful degradation** — Failed detectors don't block others (partial success).
4. **Configurable parallelism** — errgroup-based concurrent detection.
5. **Metrics built-in** — No external observability dependency.

### Scalability Bottlenecks

| Area      | Issue                               | Impact at Scale                                            |
| --------- | ----------------------------------- | ---------------------------------------------------------- |
| Correlate | O(n²) per file, capped at 10K       | Large codebases need spatial indexing                      |
| FixEngine | String-based replacement is fragile | Multiple fixes per file need deterministic ordering        |
| Merge     | Full clone on every merge           | 100K findings × merge = 100K allocations                   |
| Pipeline  | Single-pipeline-per-run             | Cannot run independent pipelines concurrently on same data |

### Modularity Score: 8/10

- **Strong:** Clean package boundaries, interface-based extensibility, adapter pattern for detectors/processors
- **Weak:** SARIF/LSP types mixed with conversion logic, pipeline.go is a god file (611 lines)

### Service Orientation Assessment

The library is correctly structured as a **library**, not a service. It has no:

- Network listeners
- Stateful servers
- External service dependencies
- Database connections

This is the right architecture for a Go library.

### Composability Assessment: 9/10

**Excellent composability:**

- `FilterFunc` predicates compose with `Filter(findings, p1, p2, p3)`
- `FindingProcessor` chain between detection and triage
- `MergeOption` functional options
- `Detector` / `DetectorFunc` / `NamedDetectorFunc` adapter chain
- `Report.Filter()` returns new Report for chaining

**Gap:** Fix application pipeline is not composable — you can't insert custom fix strategies.

## Recommendations

### Make More Composable

1. **Fix strategy as interface** — Instead of `FixStrategy` enum + switch in triage, allow callers to register `FixApplier` implementations per strategy
2. **Pipeline stages as hooks** — Allow pre/post hooks for each stage (detect, triage, fix, verify)
3. **Report as iterator** — `All()` returns `iter.Seq[Finding]` (already done!) — extend to `Filtered()`, `Grouped()`

### Make More Modular

4. **Split `pipeline.go`** — Extract adapters, config, context helpers into separate files
5. **Split `sarif.go`** — Types, export, import in separate files
6. **Extract `diagnostic.go`** — Move `golang.org/x/tools` dependency to opt-in subpackage

### Make More Scalable

7. **Spatial index for Correlate** — Sort by line, use interval tree instead of O(n²)
8. **Streaming merge** — Process findings one at a time instead of cloning all
9. **FixEngine: line-offset tracking** — Track cumulative line shifts for multi-fix application

---

_Assisted-by: Crush <crush@charm.land>_
