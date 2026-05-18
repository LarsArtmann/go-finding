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

| File | Purpose |
| ---- | ------- |

| `position.go` | Position, Range types with Overlaps/Intersection/Adjacent |
| `report.go` | Report container with summary |
| `filter.go` | Filtering and grouping utilities |
| `merge.go` | Report merging with deduplication + Correlate |
| `sarif_types.go` | SARIF struct types, constants, severity conversion helpers |
| `sarif_export.go` | Report→SARIF export (ToSARIF, WriteSARIF, findingToSARIF) |
| `sarif_import.go` | SARIF→Finding import (FindingsFromSARIF, applySarifProperties) |
| `lsp.go` | LSP Diagnostic conversion |
| `errors.go` | Structured error types (FindingError with categories) |
| `tag.go` | Tag type with IsStandard/IsValid/String methods |
| `category.go` | Category constants |
| `id.go` | ID generation utilities |
| `json.go` | JSON marshaling/unmarshaling |
| `suppression.go` | Suppression handling with IsActive convenience |
| `diff.go` | Diff(before, after) compares finding sets by ID |
| `format.go` | FormatText/FormatMarkdown for human-readable output |

#### Analysis Package

| File                   | Purpose                                                   |
| ---------------------- | --------------------------------------------------------- |
| `analysis/analysis.go` | Bidirectional go/analysis.Diagnostic ↔ Finding conversion |

#### Pipeline Package

| File                       | Purpose                                                                   |
| -------------------------- | ------------------------------------------------------------------------- |
| `pipeline/pipeline.go`     | Pipeline struct, Run, detect/triage/apply, structured logging, OnStage    |
| `pipeline/adapters.go`     | Detector/FindingProcessor interfaces, adapter types, context helpers      |
| `pipeline/config.go`       | Config struct, DefaultConfig, Validate, DetectorTimeouts, Logger, OnStage |
| `pipeline/conflict.go`     | Fix conflict detection and analysis                                       |
| `pipeline/fix_edit.go`     | FixEdit type — byte-level edit operations (Offset, Length, Replacement)   |
| `pipeline/fix_provider.go` | FixProvider interface + 3 default providers (Offset, Line, Substring)     |
| `pipeline/fix_engine.go`   | Byte-level FixEngine with provider delegation, descending-offset apply    |
| `pipeline/fix_applier.go`  | Filesystem fix application with backup/rollback, custom providers         |
| `pipeline/verify.go`       | Verification stage: re-run detectors, diff findings                       |
| `pipeline/metrics.go`      | Timing/count metrics collection with snapshots                            |
| `pipeline/retry.go`        | Exponential backoff retry wrapper for detectors                           |
| `pipeline/partial.go`      | Partial success: collect from failed detectors                            |

#### CLI

| File                         | Purpose                                                                        |
| ---------------------------- | ------------------------------------------------------------------------------ |
| `cmd/go-finding/main.go`     | Entry point, run(), profiling, flag parsing                                    |
| `cmd/go-finding/config.go`   | Config loading, severity parsing, output formatting (text/markdown/json/sarif) |
| `cmd/go-finding/registry.go` | Detector builder registry with concurrent access                               |

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

- `golang.org/x/tools` - go/analysis framework (analysis/ subpackage only)
- `golang.org/x/sync` - errgroup for parallel detection
- `github.com/go-faster/yaml` - YAML config file parsing (CLI only)
- `github.com/onsi/ginkgo/v2` - BDD testing framework
- `github.com/onsi/gomega` - BDD test matchers
- `github.com/stretchr/testify` - **INDIRECT only** (transitive via go-faster/yaml). Zero source imports. Fully migrated to ginkgo/gomega.

### Design Principles

1. **Minimal dependencies** — core types depend only on stdlib (golang.org/x/tools isolated to analysis/ subpackage)
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags
4. **One extensibility field** — `Finding.Metadata` is `map[string]string` intentionally. NO `Properties map[string]any`. Serialize complex values to JSON strings; the type-safety and interchange simplicity outweigh the convenience of `any`.
5. **Compatible** — Works with existing Go analysis tools
6. **Resilient** — Retry logic, partial success, nil-safe metrics
7. **FixApplier.Close()** — Cleans up temporary backup directories; callers should defer close
8. **Report zero-value safe** — `Report{}` uses value `sync.Mutex`, safe for concurrent use without initialization
9. **Config.FixProviders** — Custom fix providers for domain-specific (AST-aware) transformations
10. **Confidence named type** — `type Confidence float64` with `IsValid()`/`Clamp()`; prevents accidental out-of-range values
11. **Triage centralized** — `HasFix()` is canonical "is fixable?" source; `IsAutoFixable()` for pipeline auto-apply
12. **Root package dependency-free** — `golang.org/x/tools` only in `analysis/` subpackage
13. **NewFixApplier returns error** — `NewFixApplier(rootDir) (*FixApplier, error)` propagates backup dir creation failures. `NewFixApplier` delegates to `NewFixApplierWithProviders`.

### Pipeline Features

- **Byte-level FixEngine** — `[]byte` edit operations with `FixEdit{Offset, Length, Replacement}`, applied descending by offset with frontier boundary
- **FixProvider interface** — Composable provider chain: OffsetProvider (byte offsets), LineProvider (line→byte), SubstringProvider (fallback)
- **Custom provider registration** — `NewFixEngineWithProviders`, `NewFixApplierWithProviders`, `Config.FixProviders`
- **Conflict detection** — Overlapping fixes filtered before application; `ConflictInfo.ConflictsWith` populated with conflicting sources
- **Deprecated APIs** — `ConflictDetector` struct and `Verifier` struct deprecated; use `DetectConflicts()` and `Verify()` package-level functions
- **Verification** — Optional post-fix verification by re-running detectors
- **Metrics** — Optional timing/count collection with snapshot support
- **Retry** — Configurable exponential backoff for flaky detectors
- **Partial success** — Continue with findings from successful detectors
- **Parallel detection** — errgroup-based concurrent detector execution
- **Config validation** — `pipeline.New()` rejects invalid configs, returns error
- **Partial error surfacing** — `PipelineResult.PartialErrors` exposes per-detector failures
- **Metrics snapshot in result** — `PipelineResult.Metrics` auto-populated after `Run()`
- **FindingProcessor** — Composable transforms run between detection and triage (`ProcessorFunc`, `NamedProcessorFunc`)
- **FixApplier lifecycle** — `applyDirectFixes` defers `Close()` to prevent temp directory leaks
- **Line offset index** — `buildLineOffsetIndex` provides O(1) line→byte offset lookup
- **Context cancellation** — `IsContextError()` is the canonical check; all pipeline paths (retry, partial, verify) propagate `context.Canceled`/`context.DeadlineExceeded` immediately instead of silently swallowing them
- **Per-detector timeouts** — `Config.DetectorTimeouts map[string]time.Duration` overrides global timeout per detector
- **Structured logging** — `Config.Logger *slog.Logger` emits structured events for iteration start, triage, conflicts
- **Stage progress** — `Config.OnStage func(stage, iteration, count)` fires on detect/process/triage completion
- **KeySeparator** — `"\x00"` is the named constant for `Finding.Key()` composite key separator
- **SARIF import hardening** — `findingFromSarResult` generates IDs for non-go-finding SARIF; bounds-checked 3-level index access; `stringProp` helper reduces type-assertion boilerplate
- **SARIF export decomposed** — `findingToSARIF` decomposed into `sarifLocations()`, `sarifFixes()`, `sarifRelatedLocs()`, `sarifProperties()` helpers. SARIF version/schema are named constants.
- **SARIF result builder merged** — `sarifResultsFromFindings` accepts `minSeverity` parameter; `SeverityInfo` acts as "no filter"
- **Validate methods** — `Report.Validate()` validates Tool + all findings; `ToolInfo.Validate()` requires non-empty Name; follows same pattern as `Finding.Validate()`
- **Exhaustruct file-level exclusions** — SARIF types (`sarif_*.go`) and LSP types (`lsp.go`) excluded in `.golangci.yml` instead of 16 inline `//nolint:exhaustruct` directives
- **SARIF streaming** — `WriteSARIF`/`WriteSARIFFiltered` use `json.Encoder` for true streaming without intermediate `[]byte` allocation
- **FixApplier error propagation** — `NewFixApplier`/`NewFixApplierWithProviders` return `(*FixApplier, error)` instead of silently swallowing `MkdirTemp` errors
- **Partial error separation** — Context errors are propagated but excluded from `PartialResult.Errors` (they're not "partial" failures)
- **math/rand v1/v2 split** — Production code (`pipeline/retry.go`) uses `math/rand/v2` for jitter. Test code uses `math/rand` (v1) because `testing/quick.Config.Rand` requires `*math/rand.Rand` (stdlib API constraint, not removable).

### CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, JSON, and SARIF output formats
- YAML/JSON config file support with validation (including per-detector timeouts)
- Severity filtering, timeout, max-iterations
- CPU/memory profiling
- Graceful degradation on detector failures
- Metrics summary output to stderr
- Dynamic detector registry — `RegisterDetector` for plugin detectors

---

_Assisted-by: Crush <crush@charm.land>_
