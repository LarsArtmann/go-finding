# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. Seven tools detect issues; zero route them to remediation. This library solves that with:

1. **Unified Finding type** — Common representation for all tools
2. **Pipeline** — Automated detect → triage → fix → verify loop
3. **SARIF output** — Standard interchange format
4. **LSP integration** — IDE support

## Key Files

| Area                | Files                                                                                                                                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Core types**      | `finding.go`, `finding_methods.go`, `finding_validate.go`, `finding_equal.go`, `position.go`, `range.go`, `report.go`, `filter.go`, `merge.go`, `diff.go`, `format.go`, `json.go`, `id.go`, `errors.go` |
| **Named types**     | `severity.go`, `confidence.go`, `category.go`, `category_linter.go`, `tag.go`, `fix_strategy.go`, `suppression.go`, `branded_types.go`                                                                       |
| **SARIF**           | `sarif_types.go`, `sarif_export.go`, `sarif_import.go` (hand-rolled, not go-sarif — see ADR #9)                                                                                                         |
| **LSP**             | `lsp.go`                                                                                                                                                                                                |
| **Extensibility**   | `detector.go`, `adapter.go` (ToolAdapter[O]), `registry.go` (DetectorRegistry), `interval_tree.go` (IntervalIndex[T])                                                                                   |
| **Pipeline**        | `pipeline/pipeline.go` (Run), `pipeline/pipeline_detect.go` (detect/triage/apply), `pipeline/pipeline_iteration.go`, `pipeline/config.go`, `pipeline/config_file.go`                                    |
| **Fix engine**      | `pipeline/fix_engine.go`, `pipeline/fix_provider.go`, `pipeline/fix_applier.go`, `pipeline/fix_edit.go`, `pipeline/conflict.go`, `pipeline/goast/provider.go`                                           |
| **Pipeline extras** | `pipeline/stage_hook.go`, `pipeline/line_shift.go`, `pipeline/metrics.go`, `pipeline/retry.go`, `pipeline/partial.go`, `pipeline/generated_filter.go`                                                   |
| **Analysis**        | `analysis/analysis.go` (go/analysis ↔ Finding)                                                                                                                                                          |
| **Detectors**       | `internal/detectors/govet.go`, `internal/detectors/staticcheck.go`, `internal/detectors/helpers.go`                                                                                                     |
| **CLI**             | `cmd/go-finding/main.go`, `config.go`, `registry.go`, `fix_provider_registry.go`, `generated_filter.go`, `output_adapter.go` (go-output adapter)                                                        |

## Testing & Build

```bash
nix run .#test                              # Run tests
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector
golangci-lint run ./...                     # Lint
bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 25  # Benchmark regression check
```

## Dependencies

- `golang.org/x/tools` — go/analysis framework (analysis/ subpackage only)
- `golang.org/x/sync` — errgroup for parallel detection
- `github.com/go-faster/yaml` — YAML config (CLI only)
- `github.com/onsi/ginkgo/v2` + `gomega` — BDD testing
- ~~`github.com/stretchr/testify`~~ — REMOVED (was banned dep, confirmed unused via `go mod why`)
- `github.com/LarsArtmann/gogenfilter/v3` — Auto-generated Go file detection
- `github.com/larsartmann/go-output` — CLI output formatting: markdown tables, CSV, TSV (CLI only; root `finding` package stays dependency-free)

## Design Principles

1. **Minimal dependencies** — core types depend only on stdlib
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags
4. **One extensibility field** — `Finding.Metadata` is `map[string]string`. NO `Properties map[string]any`
5. **Compatible** — Works with existing Go analysis tools
6. **Resilient** — Retry logic, partial success, nil-safe metrics
7. **Root package dependency-free** — `golang.org/x/tools` only in `analysis/`

## Important Behaviors (Gotchas)

- **Report{} zero-value safe** — Uses value `sync.Mutex`, safe for concurrent use without initialization
- **Report.findings is unexported** — Use `FindingsSnapshot()` for a deep copy, `All()` for iteration, or `FindByID()` for single lookups
- **Pipeline.Run() is single-use** — Returns `errAlreadyRan` on second call
- **NewFixApplier returns error** — Propagates backup dir creation failures
- **Confidence is a named type** — `type Confidence float64` with `IsValid()`/`Clamp()`
- **NewFinding accepts Confidence** — Not raw `float64`
- **FixStrategyAI is reserved** — No backend; kept as marker for future AI remediation
- **GenerateID is length-prefixed** — Uses `writeLenField` (uint32 big-endian) to prevent hash collisions when field values contain colons
- **Position.Offset uses -1 sentinel** — `Position{}` (zero value) has Offset=0 meaning "byte 0". Constructors (Pos, NewRange, FromLSP, SARIF import) set Offset=-1 for "unset". Use `HasOffset()` (>= 0) to check.
- **FixStrategy normalized** — `NormalizeFixStrategy()` converts "" to "none". Called by Builder.Build(), SARIF import, and Equal().
- **HasFix() requires code for Direct** — `FixStrategyDirect` needs BeforeCode or AfterCode for HasFix()=true, aligning with Validate().
- **math/rand v1/v2 split** — Production uses `math/rand/v2`; tests use `math/rand` (v1) due to `testing/quick` API constraint
- **FixEngine descending-offset** — All edits resolve against the same original content snapshot; multi-edit correctness proven by tests
- **LineShiftMap shifts Position + Range** — `ShiftedPosition` shifts line + column (single-line edits); `ShiftedRange` shifts both endpoints
- **SubstringProvider column-aware** — Disambiguates multiple occurrences by line + column distance
- **context.Context on I/O** — `WriteSARIF`, `FindingsFromSARIF`, etc. accept context as first arg
- **StageHooks replace OnStage** — Use `Config.StageHooks` with `StageHook`/`StageHookFunc` for before/after events with abort capability
- **Branded types prevent mixups** — `ID`, `RuleName`, `ToolName`, `FilePath` are distinct string types. Use `finding.ID("x")` not raw `"x"` for fields. JSON marshals identically to string.
- **Validate() decomposed** — `finding_validate.go` delegates to 6 per-field validators (`validateIdentity`, `validateClassification`, `validateFix`, `validateReferences`, `validateSpatial`, `validateSuppression`). Complexity per validator < 10.
- **SeverityAliases removed** — Use `RegisterSeverityAlias()` / `LookupSeverityAlias()`. Global map guarded by `sync.RWMutex`.
- **CategoryOf is canonical** — `CategoryOf(err)` returns the category (old `GetCategory` removed)

## CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, CSV, TSV, JSON, SARIF output (markdown/CSV/TSV via go-output adapter)
- YAML/JSON config (`-config`), severity filter, profiling
- `-filter-generated` — removes findings from auto-generated files (sqlc, protobuf, etc.)
- `-fix-provider go-ast` — enables AST-aware fix provider
- `-byte-level-conflict` — precise overlap detection
- Dynamic detector registry (`RegisterDetector`)

## Architecture Decisions

- **SARIF hand-rolled** — Not go-sarif. Zero extra deps, custom property bag, streaming + context. See ADR #9.
- **Pipeline split** — `pipeline.go` + `pipeline_detect.go`, both under 350 lines
- **Byte-level FixEngine** — `[]byte` edit ops with descending-offset application, O(F+R) single-pass
- **FixProvider chain** — OffsetProvider → LineProvider → SubstringProvider (fallback); custom providers prepended
- **lineIndexAware lazy caching** — Line offset index built once per file, only when a LineProvider/SubstringProvider handles a finding
- **GoASTProvider** — AST-aware provider in `pipeline/goast/` (opt-in `go/parser` dependency)
- **IntervalIndex[T]** — Generic O(log n + k) overlap queries; used by Correlate
- **DetectorRegistry** — Thread-safe plugin architecture with `Register`/`Build`/`BuildAll`
- **MergeIter** — Streaming `iter.Seq[Finding]` merge with dedup
- **ConfigFile** — JSON config loading with `ResolveDetectors`/`ResolveProviders`
- **go-output CLI adapter** — `cmd/go-finding/output_adapter.go` adapts `[]Finding` → `output.TableData` for markdown/CSV/TSV. Root `finding` package stays dependency-free. See `docs/PRO_CONTRA_go-output-integration.md`.
- **Branded primitive types** — `type ID/RuleName/ToolName/FilePath string` in `branded_types.go`. Compile-time type safety preventing ID/Rule/Tool/File mixups. JSON marshals as string. Named `ID` not `FindingID` to avoid revive stutter (`finding.FindingID`).
- **Validate decomposition** — Monolithic `Validate()` split into 6 per-field validators. Each returns `[]error`, aggregated by `Validate()`.
- **SeverityAliases thread-safe** — `sync.RWMutex` guarded global map; `RegisterSeverityAlias()` / `LookupSeverityAlias()` API.
- **FindingTransformer** — Renamed from `FindingProcessor`/`Process()`. Pipeline uses `Config.Processors []FindingTransformer`.
- **Conflict** — Renamed from `ConflictInfo`. `AnalyzeConflicts() []Conflict`.
- **LSPRelated** — Renamed from `LSPRelatedInfo`.

## Removed APIs (v1.0.0)

All deprecated APIs from v0.6.0–v0.9.0 have been removed. No deprecated APIs remain.

See `docs/MIGRATION_v1.0.md` for migration details.

---

_Assisted-by: Crush <crush@charm.land>_
