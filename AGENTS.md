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

| File                           | Purpose                                                                                    |
| ------------------------------ | ------------------------------------------------------------------------------------------ |
| `pipeline/pipeline.go`         | Pipeline struct, Run, detect/triage/apply, structured logging, OnStage                     |
| `pipeline/adapters.go`         | Detector/FindingProcessor interfaces, adapter types, context helpers                       |
| `pipeline/config.go`           | Config struct, DefaultConfig, Validate, DetectorTimeouts, Logger, OnStage                  |
| `pipeline/conflict.go`         | Fix conflict detection and analysis                                                        |
| `pipeline/fix_edit.go`         | FixEdit type — byte-level edit operations (Offset, Length, Replacement)                    |
| `pipeline/fix_provider.go`     | FixProvider interface + 3 default providers (Offset, Line, Substring)                      |
| `pipeline/fix_engine.go`       | Byte-level FixEngine with provider delegation, descending-offset apply                     |
| `pipeline/fix_applier.go`      | Filesystem fix application with backup/rollback, custom providers                          |
| `pipeline/verify.go`           | Verification stage: re-run detectors, diff findings                                        |
| `pipeline/metrics.go`          | Timing/count metrics collection with snapshots                                             |
| `pipeline/retry.go`            | Exponential backoff retry wrapper for detectors                                            |
| `pipeline/partial.go`          | Partial success: collect from failed detectors                                             |
| `pipeline/generated_filter.go` | GeneratedFileFilter processor — removes findings from auto-generated files via gogenfilter |

#### CLI

| File                                 | Purpose                                                                        |
| ------------------------------------ | ------------------------------------------------------------------------------ |
| `cmd/go-finding/main.go`             | Entry point, run(), cliFlags struct, parseFlags(), writeResults()              |
| `cmd/go-finding/config.go`           | Config loading, severity parsing, output formatting (text/markdown/json/sarif) |
| `cmd/go-finding/registry.go`         | Detector builder registry with concurrent access                               |
| `cmd/go-finding/generated_filter.go` | CLI generated-file-filter integration (type registry, addGeneratedFilter)      |

#### Internal Detectors

| File                                | Purpose                                              |
| ----------------------------------- | ---------------------------------------------------- |
| `internal/detectors/govet.go`       | Go vet JSON → Finding converter (Detector impl)      |
| `internal/detectors/staticcheck.go` | Staticcheck JSON → Finding converter (Detector impl) |

### Testing

```bash
nix run .#test                              # Run tests
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector
golangci-lint run ./...                     # Lint
```

### Dependencies

- `golang.org/x/tools` - go/analysis framework (analysis/ subpackage only)
- `golang.org/x/sync` - errgroup for parallel detection
- `github.com/go-faster/yaml` - YAML config file parsing (CLI only)
- `github.com/onsi/ginkgo/v2` - BDD testing framework
- `github.com/onsi/gomega` - BDD test matchers
- `github.com/stretchr/testify` - **INDIRECT only** (transitive via go-faster/yaml). Zero source imports. Fully migrated to ginkgo/gomega.
- `github.com/LarsArtmann/gogenfilter/v3` - Auto-generated Go file detection (sqlc, protobuf, mockgen, etc.) — pipeline GeneratedFileFilter processor

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
- **Removed deprecated APIs** — `ConflictDetector` struct and `Verifier` struct removed; use `DetectConflicts()` and `Verify()` package-level functions
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
- **Pipeline single-use** — `Pipeline.Run()` enforces single invocation via `ran` bool guard; returns error on second call
- **FixEngine uses HasCodeChange** — `FixEngine.ApplyWithConflicts` uses `Finding.HasCodeChange()` instead of inline code check
- **OnFix accuracy** — `applyTriage` reports `false` for safeFixes not actually applied (engine couldn't resolve), not just for conflicts
- **TotalDuration guard** — `Metrics.TotalDuration()` and `Snapshot()` return 0 for negative durations (end < start)
- **FormatPartialErrors wrapping** — Uses `errors.Join` with `%w` per detector error; supports `errors.Is` for both `ErrPartialDetection` and inner errors
- **Position helpers** — `Position.IsZero()`, `Position.HasLocation()` for zero-value and location checks
- **Range helpers** — `Range.IsSingleLine()` for single-line range check
- **Finding.HasRange()** — Checks `f.Range != nil && f.Range.IsValid()`
- **Category.IsSecurity()** — Checks if category equals `CategorySecurity`
- **Report.CountBySeverity()** — Thread-safe count lookup using pre-computed summary
- **FormatText/FormatMarkdown return errors** — No longer silently swallow encoding errors; `escapeMarkdownCell` handles pipes/newlines; UTF-8 safe truncation via `utf8.RuneCountInString`
- **FilterInPlace GC safety** — Zeroes tail slice after compaction to prevent dangling pointer references
- **Negate combinator** — `Negate(FilterFunc)` inverts a filter; named `Negate` not `Not` to avoid gomega collision
- **Version auto-computed** — `Version` var computed via `fmt.Sprintf` from `Major`/`Minor`/`Patch` integer constants; eliminates manual sync risk
- **NewFinding Confidence type** — `NewFinding` accepts `Confidence` type instead of raw `float64`
- **Builder uses Validate()** — `Build()` calls `f.Validate()` for detailed per-field errors instead of `f.IsValid()` with generic error
- **GenerateID length-prefixed hash** — Uses `writeLenField` (uint32 big-endian length + bytes) to prevent hash collision when field values contain colons
- **Extended Validate()** — Checks Tags.IsValid(), Related.IsValid(), Suppression.IsValid(), Confidence range [0,1], non-empty FixStrategy validity; empty FixStrategy is valid (zero value)
- **FromJSON value return** — `FromJSON` returns `Finding` value not `*Finding`; uses `Validate()` for detailed errors; `FilterInvalid` stays with `IsValid()` for backward-compatible lossy filtering
- **DiffResult helpers** — `HasChanges()` reports additions/removals/modifications; `Stats()` returns `"+N -N ~N =N"` summary; `Modified` tracks same-ID content changes via `Equal()`
- **AnyOf filter combinator** — `AnyOf(FilterFunc...)` matches if ANY predicate matches; named `AnyOf` not `Or` to avoid gomega collision
- **ByConfidence filters** — `ByConfidence(Confidence)` and `ByConfidenceAtLeast(Confidence)` filter constructors
- **ParseSeverity** — `ParseSeverity(string)` and `MustParseSeverity(string)` for string→Severity conversion
- **Report.Merge data-race fix** — `Merge()` reads `other.Findings` under `other.mu.RLock()` and copies data
- **ActiveFindings deterministic** — Uses `IsSuppressedAt(now)` instead of `IsSuppressed()` for consistent suppression checks
- **Confidence validation consistency** — `Finding.Validate()` uses `Confidence.IsValid()` instead of direct float comparison
- **Dead code removed** — `Iteration.Failed` field removed (was declared but never written)
- **Zero lint warnings** — All err113, errcheck, gosec, goconst, staticcheck, exhaustruct, golines, paralleltest, nolintlint, prealloc issues resolved; `.golangci.yml` updated with appropriate exclusions
- **DeduplicateByID empty ID** — `dedupKey` returns `(string, bool)`; empty ID findings are never deduplicated (no silent fallback to position)
- **FixApplier path validation** — Rejects path traversal (e.g., `../../../etc/passwd`); checks `filepath.Clean(path)` is within `rootDir`
- **Shallow copy semantics documented** — `FindByID` and `All()` document that slice/map fields (Tags, Related, Metadata) are shallow copies; use `Clone()` for deep copy
- **Conflict grouping documented** — Transitive overlap grouping and conservative single-survivor resolution are documented as intentional design
- **SARIF property constants** — All `go-finding/*` property keys are named constants in `sarif_types.go`; added `sarifPropAfterCode`, `sarifPropEditPrefix`
- **TODO_LIST deep audit** — 97 done, 93 open; 29 items verified as already done, 14 phantoms annotated, 7 owner-decision items tagged
- **FixApplier preserves permissions** — `applyToFile` uses `os.Stat` + original `info.Mode()` instead of hardcoded `0o600`
- **resolveEdits surfaces provider errors** — Returns `([]FixEdit, error)` instead of silently swallowing; `ApplyWithConflicts` returns 4-tuple with `[]error` for provider errors
- **FilterConflictingEdits returns errors** — Signature changed to `([]finding.Finding, []error)` to surface provider errors
- **Diff.Modified tracks both versions** — `ModifiedPair{Before, After}` replaces flat `[]Finding`; enables inspecting what changed
- **VerifyResult.Modified** — `DiffFindings` now separates modified findings (same Key, different content) from unchanged
- **Confidence.Compare()** — `Compare(other Confidence) int` follows `Severity.Compare()` pattern
- **Confidence.String() named labels** — Returns `"none"/"low"/"medium"/"high"/"full"` for standard levels; decimal for custom
- **ErrInvalidBuilder removed** — Dead code eliminated; `Build()` returns `Validate()` errors directly
- **CLI delegates to finding.ParseSeverity** — `cmd/go-finding/config.go` no longer duplicates severity parsing logic
- **FixEngine multi-edit correctness** — Descending-offset application is correct; all edits resolve against same original content snapshot; tests prove it
- **SARIF AfterCode round-trip** — `sarifProperties` exports both `BeforeCode` and `AfterCode` as properties; import reads both for full round-trip fidelity
- **SARIF Rank omission** — `findingToSARIF` only sets `Rank` when `Confidence > 0`; NaN/zero confidence omitted per SARIF spec
- **Range.IsInverted()** — Detects End before Start (line or column level); used in `Finding.Validate()` to reject inverted ranges
- **DeduplicateBy edge cases** — `DeduplicateByPosition` and `DeduplicateByRule` skip findings with empty `Position.File` (return `false`), preventing false dedup of findings without file paths
- **LSPSeverity typed constant** — `type LSPSeverity int` with constants `LSPSeverityError/Warning/Info/Hint`; `LSPDiagnostic.Severity` uses named type instead of raw `int`
- **FixApplier resolveErrors surfaced** — `applyToFile` returns provider errors from `FixEngine.ApplyWithConflicts` via `errors.Join` instead of silently discarding
- **GeneratedFileFilter** — `pipeline.GeneratedFileFilter` FindingProcessor removes findings from auto-generated Go files (sqlc, protobuf, mockgen, templ, etc.) via `gogenfilter/v3`; configurable per-generator type, include/exclude patterns; CLI flags `-filter-generated`, `-filter-generated-types`, `-generated-exclude`, `-generated-include`
- **SARIF import FixStrategy** — `findingFromSarResult` sets `FixStrategyDirect` for replacements (auto-applicable), `FixStrategySuggest` for description-only fixes
- **FileBackup permission preservation** — `Backup()` captures original `os.FileInfo.Mode()`; `Restore()` uses stored mode instead of hardcoded `0o600`
- **Config.DetectorTimeouts validation** — `Config.Validate()` rejects negative per-detector timeouts (would panic at runtime)
- **CLI duration parse errors** — `toPipelineConfig()` returns `(Config, error)` and surfaces malformed duration strings instead of silently using defaults
- **ErrorCategory.IsValid() format validation** — Same `^[a-z][a-z0-9-]*$` rune validation as `Category.IsValid()`; rejects `"Validation"`, `"has space"`
- **Pipeline.Run() single-use** — Doc comment corrected: returns `errAlreadyRan` on second call (does not "reset internal state")
- **byFindingID unified** — Both `diff.go` and `pipeline/verify.go` use `cmp.Compare` instead of manual comparison
- **CLI timeout not double-wrapped** — Pipeline handles `context.WithTimeout` internally; CLI passes `context.Background()` directly

### CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, JSON, and SARIF output formats
- YAML/JSON config file support with validation (including per-detector timeouts)
- Severity filtering, timeout, max-iterations
- CPU/memory profiling
- Graceful degradation on detector failures
- Metrics summary output to stderr
- Dynamic detector registry — `RegisterDetector` for plugin detectors
- Generated file filtering — `-filter-generated` CLI flag + `filterGenerated` config; uses `gogenfilter/v3` to detect sqlc, protobuf, mockgen, etc. via `GeneratedFileFilter` FindingProcessor

---

_Assisted-by: Crush <crush@charm.land>_
