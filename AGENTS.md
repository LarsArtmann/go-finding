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

| `position.go` | Position type with IsZero/HasOffset/HasLocation helpers |
|| `range.go` | Range type with Overlaps/Intersection/Adjacent/IsInverted/IsSingleLine |
|| `finding.go` | Finding struct definition and builder (NewFinding) |
|| `finding_methods.go` | Finding methods: Clone, Has\*, String, Preview, IsSuppressed |
|| `finding_validate.go` | Finding validation: Validate, IsValid, Key |
|| `finding_equal.go` | Finding equality: Equal, equalStringSlices helpers |
| `report.go` | Report container with summary |
| `filter.go` | Filtering and grouping utilities |
| `validate_helpers.go` | Shared isValidLowercaseHyphen validation for Category and Tag |
| `merge.go` | Report combining with deduplication + Correlate |
| `sarif_types.go` | SARIF struct types, constants, severity conversion helpers |
| `sarif_export.go` | Report→SARIF export (ToSARIF, WriteSARIF, findingToSARIF) |
| `sarif_import.go` | SARIF→Finding import (FindingsFromSARIF, FindingsFromReader, applySarifProperties) |
| `lsp.go` | LSP Diagnostic conversion |
| `errors.go` | Structured error types (FindingError with categories) |
| `tag.go` | Tag type with IsStandard/IsValid/String methods |
| `category.go` | Category constants, ParseCategory, MustParseCategory |
| `category_linter.go` | CategoryForLinter registry with 70+ linter mappings, RegisterLinterCategory |
| `detector.go` | Detector interface, DetectorFunc, NamedDetectorFunc (moved from pipeline) |
| `adapter.go` | ToolAdapter[O] generic tool→Finding converter |
| `id.go` | ID generation utilities |
| `json.go` | JSON marshaling/unmarshaling |
| `suppression.go` | Suppression handling with IsActive convenience |
| `diff.go` | Diff(before, after) compares finding sets by ID |
| `format.go` | FormatText/FormatMarkdown for human-readable output |
| `registry.go` | DetectorRegistry — thread-safe named detector constructor registry |
| `interval_tree.go` | IntervalIndex[T] generic sorted-scan overlap queries |
| `fix_strategy.go` | FixStrategy constants, CanAutoApply, NeedsAI |

#### Analysis Package

| File                   | Purpose                                                   |
| ---------------------- | --------------------------------------------------------- |
| `analysis/analysis.go` | Bidirectional go/analysis.Diagnostic ↔ Finding conversion |

#### Pipeline Package

| File                           | Purpose                                                                                                                                                   |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline/pipeline_detect.go`  | Detection, triage, and apply logic extracted from pipeline.go                                                                                             |
| `pipeline/result.go`           | PipelineResult, CompletionReason, Iteration types                                                                                                         |
| `pipeline/stage.go`            | Stage named type with constants for pipeline stages                                                                                                       |
| `pipeline/pipeline.go`         | Pipeline struct, Run, structured logging, OnStage                                                                                                         |
| `pipeline/adapters.go`         | Type aliases to root package (Detector, DetectorFunc, NamedDetectorFunc), FindingProcessor interfaces, ProcessorFunc, NamedProcessorFunc, context helpers |
| `pipeline/config.go`           | Config struct, DefaultConfig, Validate, DetectorTimeouts, Logger, OnStage                                                                                 |
| `pipeline/conflict.go`         | Fix conflict detection and analysis                                                                                                                       |
| `pipeline/fix_edit.go`         | FixEdit type — byte-level edit operations (Offset, Length, Replacement)                                                                                   |
| `pipeline/fix_provider.go`     | FixProvider interface + 3 default providers (Offset, Line, Substring)                                                                                     |
| `pipeline/fix_engine.go`       | Byte-level FixEngine with provider delegation, descending-offset apply                                                                                    |
| `pipeline/fix_applier.go`      | Filesystem fix application with backup/rollback, custom providers                                                                                         |
| `pipeline/verify.go`           | Verification stage: re-run detectors, diff findings                                                                                                       |
| `pipeline/metrics.go`          | Timing/count metrics collection with snapshots                                                                                                            |
| `pipeline/retry.go`            | Exponential backoff retry wrapper for detectors                                                                                                           |
| `pipeline/partial.go`          | Partial success: collect from failed detectors                                                                                                            |
| `pipeline/generated_filter.go` | GeneratedFileFilter processor — removes findings from auto-generated files via gogenfilter                                                                |
| `pipeline/stage_hook.go`       | StageHook interface, StageHookFunc adapter, StageEvent                                                                                                    |
| `pipeline/line_shift.go`       | LineShiftMap — byte-offset-aware line shift tracking after edits                                                                                          |
| `pipeline/config_file.go`      | ConfigFile struct, ConfigFromFile/ConfigFromReader                                                                                                        |

#### CLI

| File                                 | Purpose                                                                        |
| ------------------------------------ | ------------------------------------------------------------------------------ |
| `cmd/go-finding/main.go`             | Entry point, run(), cliFlags struct, parseFlags(), writeResults()              |
| `cmd/go-finding/config.go`           | Config loading, severity parsing, output formatting (text/markdown/json/sarif) |
| `cmd/go-finding/registry.go`         | Detector builder registry with concurrent access                               |
| `cmd/go-finding/generated_filter.go` | CLI generated-file-filter integration (type registry, addGeneratedFilter)      |

#### Internal Detectors

| File                            | Purpose                                                  |
| ------------------------------- | -------------------------------------------------------- |
| `internal/detectors/govet.go`   | Go vet JSON → Finding converter (Detector impl)          |
| `internal/detectors/helpers.go` | Shared constants (DetectorName\*) and resolvePath helper |

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
- **FixApplier lifecycle** — Created eagerly in `pipeline.New()`, not lazily; errors caught at construction time
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
- **context.Context on I/O** — `WriteSARIF(ctx, w)`, `WriteSARIFFiltered(ctx, w, sev)`, `FindingsFromSARIF(ctx, data)`, `FindingsFromReader(ctx, r)` accept `context.Context` as first arg; cancelled context returns wrapped `ctx.Err()` before I/O begins. `WriteTo` (io.WriterTo compat) delegates with `context.Background()`. `FindingsFromReader` streams via `json.Decoder` without buffering full input.
- **Report.MergeInto immutable** — `MergeInto(other) *Report` returns a new Report without modifying receiver or other; existing `Merge()` kept for backward compat
- **analysis.DefaultRelation consolidated** — References `finding.RelationRelated` instead of duplicating `"related"` string
- **RecordFix deprecated** — `RecordFix()` deprecated in favor of `RecordFixes(1)`; will be removed in v1.0.0
- **FixApplier eager construction** — Created in `pipeline.New()`, not lazily in `applyDirectFixes`; backup dir errors caught at construction time
- **lineIndexAware lazy caching** — `FixEngine.resolveEdits` accepts `*[]int` and lazily builds the line offset index on first access by a `lineIndexAware` provider. Zero overhead for OffsetProvider-handled findings (byte offsets bypass line indexing entirely). LineProvider/SubstringProvider get O(1) index reuse across all findings in a file
- **Benchmark offset bug** — Session 13 fix: `generateOffsetFixes` used Range width 4 for BeforeCode "old()" (5 chars); OffsetProvider always rejected, silently benchmarking SubstringProvider. Fixed to use actual occurrence positions via `findOldOccurrences`
- **Correlate pre-allocation** — `correlations` slice pre-sized with `min(len(findings), maxCorrelations)`; `withRange`/`withoutRange` pre-sized with `len(fileFindings)`
- **buildLineOffsetIndex SIMD** — First pass uses `bytes.Count` (SIMD for single-byte needle) instead of byte-by-byte loop
- **findAllOccurrences capacity** — `defaultOccurrenceCapacity = 32` starting capacity; `bytes.Count` exact pre-allocation evaluated and rejected (double-scan overhead not worth it for fallback provider)

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

### Architecture Decisions (2026-06-01 review)

- **Shared validation helper** — `isValidLowercaseHyphen(s)` in `validate_helpers.go` used by both `Category.IsValid()` and `ErrorCategory.IsValid()`
- **Detector name constants** — Single source of truth in `internal/detectors/helpers.go` (`DetectorNameGovet`, `DetectorNameStaticcheck`); CLI references these via `detectors.DetectorName*`
- **Shared path resolution** — `resolvePath(dir, file)` in `internal/detectors/helpers.go` used by both govet and staticcheck
- **Pipeline split** — `pipeline/pipeline.go` (core: Run, runIteration, collectAllFindings) + `pipeline/pipeline_detect.go` (detect, triage, apply) — both under 350 lines
- **Filter DRY** — `matchesAll(finding, predicates)` in `filter.go` shared by `Filter` and `FilterInPlace`
- **Retry delay clamped** — `delay()` clamps result to `MaxDelay` after jitter to prevent exceeding cap

- **Context error detection in iterations** — `runIteration` error path uses `IsContextError()` to distinguish `ReasonTimeout`/`ReasonCancelled` from `ReasonError`
- **StageVerify wired** — Verify stage records timing metrics and fires `OnStage` callback

- **CorrelationScore methods** — `IsValid()` checks [0.0, 1.0]; `String()` formats decimal; consistent with `Confidence`
- **DurationMs caller-set** — `Summary.DurationMs` is NOT computed by `ComputeSummary()`; caller must set manually. Pipeline timing lives in `Metrics.TotalDuration`.

### Known Open Items

- `Report.Findings` is a public slice — external code can bypass mutex (encapsulation risk)
- `FixStrategyAI` + `NeedsAI()` are published API with no backend — **DECISION: KEEP as reserved** for future AI-powered remediation. Zero implementation cost. Do not remove.
- `Tag` constants partially overlap `Category` constants (TagSecurity/CategorySecurity etc.) — no structural link
- `RecordFix()` superseded by `RecordFixes(uint)` — convenience method with no production callers
- FixApplier path validation resolves symlinks (`filepath.EvalSymlinks` + `filepath.Clean`) — prevents symlink-based path traversal
- **SARIF: hand-rolled, not go-sarif** — Evaluated `github.com/owenrumney/go-sarif/v3` (v3.3.0, 83 stars, actively maintained). Decision: keep hand-rolled. Reasons: (1) zero extra dependencies aligns with design principle #1, (2) adapter layer would be ~200-300 LOC negating maintenance savings, (3) round-trip property bag is custom-built, (4) streaming + context cancellation are first-class, (5) we use only 15 of 100+ SARIF types. Revisit if we need SARIF 2.2, schema validation, or code flows. See docs/architecture-decisions.md #9.
- **ByteLevelConflictDetection** — `Config.ByteLevelConflictDetection bool` enables byte-level conflict filtering via `filterByFileEdits`; reads file content and uses FixEngine for precise overlap detection (source: pipeline/config.go)
- **TriageFunc** — `Config.TriageFunc` allows custom triage logic; `DefaultTriageFunc` preserves existing behavior; `TriageResult` type in config.go (source: pipeline/config.go)
- **Equal refactored** — `Finding.Equal()` uses individual early returns per field instead of monolithic condition; `SortFindingsByID` exported as shared utility (source: finding.go, diff.go)
- **PrettyJSONFiltered** — `Report.PrettyJSONFiltered()` returns JSON excluding suppressed findings (source: json.go)
- **SARIF decomposition** — Both `findingFromSarResult` (9 decisions) and `applySarifProperties` (15 decisions) are well under gocognit threshold (35); no further decomposition needed (source: sarif_import.go)
- **GoReleaser** — `.goreleaser.yml` exists with full config: builds, archives, checksums, changelog, cosign, SBOMs, brew, nix, nfpm, scoop (source: .goreleaser.yml)
- **Codebase fully modernized** — uses `slices.SortFunc`, `slices.Contains`, `slices.Backward`, `maps.Keys`, `maps.Clone`, `maps.Equal`, `iter.Seq`, `errors.AsType`, `for i := range N` (source: project-wide)
- **SARIF types unexported** — All SARIF struct types (`sarifLog`, `sarifRun`, `sarifResult`, etc.) are unexported; only `FromSARIFLevel` remains exported as a utility. Pre-v1.0 breaking change to prevent external coupling to internal representation.
- **Internal constants unexported** — `keySeparator`, `mergedToolName`, `emptyToolName` are unexported; no external consumers existed. Pre-v1.0 cleanup.
- **FindingsSnapshot()** — `Report.FindingsSnapshot()` returns a deep-cloned slice of all findings; safe for concurrent use without holding the lock. v1.0 migration path from direct `Findings` slice access.
- **FixStrategyAI kept as reserved** — Decision: keep `FixStrategyAI` and `NeedsAI()` as reserved values. Zero implementation cost, documented as reserved for future AI-powered remediation. Removing would break consumers who've started using the constant.
- **Pipeline tests split** — `pipeline_test.go` (1546 lines) split into `pipeline_new_test.go` (construction/config), `pipeline_triage_test.go` (triage/fix), keeping run tests in `pipeline_test.go`
- **CLI test coverage 90.7%** — Up from 70%; added tests for `addGeneratedFilter`, `parseFilterGenTypes`, `mustKeys`, `splitCommaList`
- **Fuzz seed corpus** — All 20 fuzz targets now have `f.Add()` seed values; 13 missing corpus directories created under `testdata/fuzz/`

### Session 6 (2026-06-08)

- **flake.nix infinite recursion fixed** — `goPkg = goPkg` → `goPkg = pkgs.go_1_26`; also consolidated duplicate `checks.build` into single block
- **v0.5.0 released** — Tagged and pushed; 42+ commits since v0.4.3 justified minor bump
- **Correlate complexity documented** — Godoc explains O(n·k) normal, O(k²) worst case, 10K cap
- **DeduplicateBy.String()** — `String()` method added for consistency with all other named types
- **Position semantic trap documented** — `Offset=0` means both "byte 0" (valid) and passes `HasOffset()=true` while also being `IsZero()=true`; now explicit in doc comments
- **flake.nix maintainers** — Added `maintainers = [ lib.maintainers.larsartmann ]`
- **CHANGELOG.md updated** — Full v0.5.0 entry with all changes
- **TODO_LIST.md audited** — 18 stale items marked done
- **finding.go split** — `finding.go` (509 lines) → `finding.go` (95) + `finding_methods.go` (162) + `finding_validate.go` (119) + `finding_equal.go` (149)
- **position.go split** — `position.go` (444 lines) → `position.go` (78) + `range.go` (366)
- **FixApplier symlink resolution** — Added `filepath.EvalSymlinks` to path validation preventing symlink-based traversal
- **ADR 10** — Report.Findings encapsulation strategy written to `docs/architecture-decisions.md`
- **lsp_test.go** — Replaced raw `"go-finding/lsp-severity"` with `LSPSeverityKey` constant
- **Metadata namespacing** — Documented `"toolName.key"` convention and reserved `"go-finding/"` prefix in finding.go godoc
- **GoReleaser verified** — `.github/workflows/release.yml` triggers on `v*` tags; `main.version` var exists for ldflags

### Session 7 (2026-06-09) — Consumer Audit Adoption

- **Detector moved to root package** — `Detector` interface, `DetectorFunc`, `NamedDetectorFunc` moved from `pipeline/` to root package; pipeline re-exports as type aliases for zero-breaking-change backward compatibility
- **SeverityAliases map** — `SeverityAliases` exported map with 9 common aliases (warn, high, medium, low, fatal, critical, note, advice, suggestion); `ParseSeverity` checks aliases after canonical names; replaces 7 duplicated `mapSeverity()` switch statements across consumers
- **CategoryForLinter registry** — `category_linter.go` with 70+ linter→category mappings (golangci-lint, go analyzers); `CategoryForLinter(name)` with case-insensitive lookup; `RegisterLinterCategory(name, cat)` for runtime overrides; defaults to `CategoryCorrectness` for unknown linters
- **ParseCategory/MustParseCategory** — `ParseCategory(s)` and `MustParseCategory(s)` in `category.go` following same pattern as ParseSeverity; validates `^[a-z][a-z0-9-]*$` format
- **ToolAdapter[O] generic** — `adapter.go` with `ToolAdapter[O any]` implementing `Detector`; `NewToolAdapter(name, run, parse, convert)` wires exec→parse→convert pipeline; replaces ~30 files of duplicated adapter code across 5 consumer projects
- **Lint fixes** — `finding_validate.go` godoc (revive), `merge.go` goconst nolint, `pipeline/fix_applier.go` noinlineerr refactor, `pipeline/retry.go` wrapcheck fix, `.golangci.yml` exclusions for severity.go and category_linter.go

### Session 8 (2026-06-09) — Deduplication + Race Fix

- **art-dupl @ 50: 2 → 0 clone groups** — Industry-standard threshold now reports zero real duplication in the codebase
- **art-dupl @ 25: 16 → 11 clone groups** — Remaining 11 are all idiomatic Go patterns (struct literal test data, factory functions, separate deliverables like testable example + runnable example)
- **`TestRegisterLinterCategory` table-driven** — Unified `TestRegisterLinterCategory` + `TestRegisterLinterCategoryOverride` into one table-driven test with `preRegistered/override/postRegistered` fields; preserves the global registry restoration in the loop
- **`TestDeduplicateStrategies_EmptyFileNotDeduplicated` table-driven** — Unified `TestDeduplicateByPosition_EmptyFileNotDeduplicated` + `TestDeduplicateByRule_EmptyFileNotDeduplicated` into one table-driven test parameterized on `DeduplicateBy`
- **`TestMetrics_RecordFixes` table-driven** — Unified 4 separate `RecordFix`/`RecordFixes` tests into one table-driven test covering zero, single batch, mixed deprecated/batch, and three single recordings
- **`TestFinding_Validate` table-driven** — Converted 12 separate `t.Run` subtests into one table-driven test with `mutate func(*Finding)` + `wantErr` + `errContains`; preserves the `IsCategory(err, ErrCategoryValidation)` check for ALL error cases (was only on one case before, now stronger)
- **`ExampleFormatMarkdown` + `ExampleGeneratedFileFilter`** — Replaced inline `finding.NewFinding(...)` calls with the existing `newExampleFinding` helper
- **`linterCategories` thread-safe** — Added `sync.RWMutex` (`linterCategoriesMu`) guarding the global map; `CategoryForLinter` uses `RLock`, `RegisterLinterCategory` uses `Lock`. Doc updated to "Safe for concurrent use" (was "safe for concurrent use via init-time or early-program setup")
- **Pre-existing data race fixed** — The `linterCategories` map race between `TestCategoryForLinter` (reads) and `TestRegisterLinterCategory` (writes) under `-race` was a flaky 18/20 → 20/20. Root cause: global map with no synchronization. Fix: proper RWMutex. Now `go test -race -count=1` is stable across 30 consecutive runs

### Session 9 (2026-06-09) — TODO List Execution

- **Report.Merge deprecated** — `// Deprecated:` godoc added; `MergeInto` is the replacement; will be removed in v1.0.0
- **art-dupl flake app + CI job** — `nix run .#art-dupl` app and `dupl` CI job (installs via `go install`, runs with threshold 50)
- **Fuzz seed corpus persisted** — All 20 fuzz targets have file-based seed corpus in `testdata/fuzz/` (70+ files); `.gitignore` excludes fuzzer-generated hex-hash files
- **TestExamplesRun integration test** — New `TestExamplesRun` in `examples/example_compile_test.go` compiles and runs all 3 examples, verifying expected output strings
- **API stability audit** — Every exported symbol across 3 packages audited and classified (stable/unstable/deprecated/reserved); `docs/API_STABILITY.md` rewritten with per-symbol tables
- **README.md updated** — Added ToolAdapter section, CategoryForLinter section, updated project stats to v0.6.1 numbers (95.7% root, 98.5% analysis, 93.7% pipeline, 90.7% CLI)
- **pkg.go.dev badge** — Already present in README since earlier session
- **CI race detection** — Verified CI already runs `-race` in both `test` and `stress` jobs
- **gocyclo threshold** — Already configured at `min-complexity: 25` with 0 violations
- **Property tests determinism** — Already deterministic via `rand.New(rand.NewSource(42))` seed in `checkPropertyAny`

---

### Session 11 (2026-06-09) — TODO Tasks 11-20

- **Report.Findings deprecation (ADR 10)** — `// Deprecated:` godoc on `Findings` field; internal `readFindings()` helper; thread-safe SARIF/JSON/CLI access via `FindingsSnapshot()`; field will be unexported in v1.0
- **StageHook interface** — `pipeline/stage_hook.go`: `StageHook` interface with `StageHookFunc` adapter; `StageEvent` struct; wired into detect/process/triage/apply/verify stages; pre-hook errors abort pipeline
- **LineShiftMap** — `pipeline/line_shift.go`: byte-offset-aware line shift tracking after fix edits; `NewLineShiftMap(content, edits)` + `ShiftedLine(original)`; uses `buildLineOffsetIndex` for O(1) line→byte lookup
- **IntervalIndex[T]** — `interval_tree.go`: generic sorted-scan overlap queries with O(log n + k) complexity; `NewIntervalIndex[T]` + `Query(start, end)`; uses binary search cutoff + linear scan
- **MergeIter()** — `merge.go`: streaming `iter.Seq[Finding]` merge; reads each report under RLock via `readFindings()`; clones each finding; supports early termination
- **ConfigFile** — `pipeline/config_file.go`: `ConfigFile` struct with `ConfigFromFile(path)` / `ConfigFromReader(r)`; YAML/JSON config loading for library use
- **DetectorRegistry** — `registry.go`: thread-safe named detector constructor registry; `Register/Build/BuildAll/Names/Has`; `MustRegister` panics variant; sentinel errors with `%w` wrapping
- **Benchmark CI gate** — `.github/workflows/ci.yml`: separate `benchmark` job with `go test -bench`

### Session 12 (2026-06-14) — Performance Optimization

- **FixEngine applyEditsToContent** — `pipeline/fix_engine.go`: Rewrote `applyEditsWithConflicts` to separate conflict detection (Phase 1) from edit application (Phase 2). New `applyEditsToContent` function uses a pre-allocated `[]byte` with single-pass descending-offset iteration, reducing complexity from O(n×F) to O(F+R). Eliminates per-edit `append(append(append(...)))` triple-copy pattern
- **offsetLineDistance binary search** — `pipeline/fix_provider.go`: Rewrote `offsetLineDistance` from O(n) byte-by-byte newline count to O(log n) binary search on line offset index. New `offsetToLine` helper uses `slices.BinarySearch`. Signature changed from `(content []byte, ...)` to `(lineIndex []int, ...)`
- **SubstringProvider line index cache** — `pipeline/fix_provider.go`: `SubstringProvider.Edits` now builds `lineOffsetIndex` ONCE before iterating occurrences, instead of calling the old O(n) `offsetLineDistance` per occurrence. For 10k occurrences on a 170KB file: O(F×L) → O(F + L×log F)
- **Combined FixEngine impact** — 452× faster (291s → 644ms for 1000 edits). Memory: 443MB → 525MB (+18% from line index allocation per call, acceptable tradeoff). Pipeline benchmark suite: 350s → 31s total
- **dedupKey strings.Builder** — `merge.go`: Replaced `fmt.Sprintf("%s:%s:%d:%d", ...)` in `DeduplicateByPosition` and `DeduplicateByRule` with `strings.Builder` + `strconv.Itoa`, avoiding format parser overhead
- **GroupBy pre-allocation** — `filter.go`: Pre-sized `GroupBy`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory` result maps with `make(map, len(findings))`
- **Diff/DiffFindings pre-allocation** — `diff.go`, `pipeline/verify.go`: Pre-allocated `added`, `removed`, `modified`, `unchanged` result slices with `make([]T, 0, len(input))`
- **Performance analysis report** — `docs/research/performance-analysis.html`: Comprehensive HTML report covering CPU, RAM, Disk/IO, Network, Concurrency, GPU, and scaling characteristics with live benchmark data
- **Performance optimization plan** — `docs/planning/2026-06-14_16-25_PERFORMANCE-OPTIMIZATION.md`: Pareto-principle plan with 1%/4%/20% breakdown, 15 medium-granularity tasks, 55 fine-granularity tasks, mermaid.js execution graph

### Session 13 (2026-06-14) — Provider Index Caching + Correlate Pre-allocation

- **lineIndexAware interface** — `pipeline/fix_provider.go`: Optional interface `lineIndexAware` with `EditsWithLineIndex(content, lineIndex, f)`. `LineProvider` and `SubstringProvider` both implement it. `FixEngine.resolveEdits` accepts `*[]int` and lazily builds the line offset index on first access by a `lineIndexAware` provider — zero overhead for findings handled by `OffsetProvider` (byte offsets)
- **LineProvider index caching** — `pipeline/fix_provider.go`: `LineProvider.Edits` delegates to `EditsWithLineIndex` with a freshly built index (backward compat). `FixEngine.ApplyWithConflicts` passes a shared `*[]int` to `resolveEdits`, building the index once instead of per finding. Benchmark: LineProvider 1000 fixes 150ms → 812μs (**185×**), 86MB → 4MB allocations
- **SubstringProvider index caching** — `pipeline/fix_provider.go`: Same pattern. Benchmark: SubstringProvider 1000 fixes 644ms → 404ms (**1.6×**), 525MB → 361MB, 20040 → 16040 allocs
- **buildLineOffsetIndex SIMD** — `pipeline/fix_provider.go`: First pass (newline count for pre-allocation) replaced byte-by-byte loop with `bytes.Count(content, []byte{'\n'})` — true SIMD acceleration for single-byte needle on amd64/arm64
- **findAllOccurrences starting capacity** — `pipeline/fix_provider.go`: `defaultOccurrenceCapacity = 32` constant; eliminates first 5 reallocation growth phases. Evaluated `bytes.Count` for exact pre-allocation but reverted — double-scan overhead (12% time increase) outweighed allocation savings for this fallback provider
- **Lazy index build** — `pipeline/fix_engine.go`: Initial eager approach (always build if any provider implements `lineIndexAware`) caused 8× regression on OffsetProvider benchmarks (22μs → 170μs for 1 fix). Fixed with lazy `*[]int` pointer — index built only when a `lineIndexAware` provider actually handles a finding
- **Correlate pre-allocation** — `merge.go`: Pre-sized `correlations` with `min(len(findings), maxCorrelations)` capacity, and `withRange`/`withoutRange` with `len(fileFindings)`. Benchmark: 113μs → 83μs (27%), 75 → 51 allocs (32%)
- **Benchmark offset bug fixed** — `pipeline/fix_engine_bench_test.go`: `generateOffsetFixes` used Range width 4 but BeforeCode "old()" is 5 chars — OffsetProvider always rejected, silently testing SubstringProvider. Rewrote to find actual "old()" positions via `findOldOccurrences` helper. Added `generateLineFixes` and `generateSubstringFixes` helpers with dedicated `BenchmarkFixEngine_LineProvider_*` and `BenchmarkFixEngine_Substring_*` benchmarks
- **SARIF struct pooling evaluated** — Decision: **SKIP**. Profiling shows 312KB/100 findings where bytes are dominated by un-poolable JSON output buffer (~200KB) and per-finding strings. Pooling struct headers saves ~0.2% of total bytes. Complexity (nested slice/map reset, lifecycle management) and risk (use-after-Put) unjustified for marginal benefit

---

### Session 14 (2026-06-14) — Go AST Provider + Top 25 TODO Items

- **GoASTProvider** — `pipeline/goast/provider.go`: Domain-specific FixProvider for `.go` files using `go/parser` + `go/ast`. Uses AST node positions to disambiguate BeforeCode occurrences, eliminating SubstringProvider's line-proximity guessing. FNV-1a content-hash cache avoids re-parsing. Parse failures return `(nil, nil)` to let text-based providers handle the finding. Separate subpackage keeps `go/parser` as opt-in dependency.
- **CLI fix provider wiring** — `cmd/go-finding/fix_provider_registry.go`: Name→constructor registry (`"go-ast"`). `-fix-provider` CLI flag and `fixProviders` YAML/JSON config field. Providers are prepended to default text-based chain. Validation rejects unknown provider names.
- **Fuzz applyEditsToContent** — `pipeline/fix_engine_fuzz_test.go`: Two fuzz targets — `FuzzApplyEditsToContent` (non-overlapping, verifies fast==naive oracle) and `FuzzApplyEditsToContent_Overlapping` (no-panic). Found and fixed two bugs: (1) overlapping edits caused slice bounds panic, (2) negative `finalSize` caused `makeslice` panic.
- **applyEditsToContent defensive hardening** — `pipeline/fix_engine.go`: Skip overlapping edits (`edit.Offset < prevEnd`) instead of panicking. Clamp `finalSize` to `max(0, finalSize)` to prevent negative capacity.
- **Mixed-provider integration tests** — `pipeline/fix_provider_integration_test.go`: Tests Offset/Line/Substring interleaving, lazy index build (OffsetProvider-only findings don't build index), and cross-provider correctness.
- **MergeIter pre-allocate seen map** — `merge.go`: Pre-sized `seen` map with total findings count when dedup enabled.
- **Combine → MergeIter DRY** — `merge.go`: `Combine` now delegates to `MergeIter` for dedup+clone logic, eliminating duplicated code paths. Pre-allocates findings slice from total.
- **findingsLocked() accessor** — `report.go`, `report_query.go`, `json.go`: Internal `findingsLocked()` method returns `Findings` slice without lock acquisition (caller must hold lock). All internal methods in `report_query.go` and `json.go` now use `findingsLocked()` instead of direct field access. Prepares for v1.0 unexport of `Findings`.
- **Correlate range benchmark** — `correlate_bench_test.go`: Range-based and point-based Correlate benchmarks at 10/100/1000 scales.
- **MergeIter vs Combine benchmark** — `correlate_bench_test.go`: Direct comparison at 5-10 reports × 100-1000 findings each, with/without dedup.
- **IntervalIndex benchmarks** — `correlate_bench_test.go`: Build and Query benchmarks at 100/1K/10K intervals.
- **benchstat CI regression** — `.github/workflows/ci.yml`: `benchmark` job now installs benchstat, runs `-count=10`, compares against `benchmarks/baseline.txt`, uploads results artifact. `scripts/bench-baseline.sh` for local baseline capture/comparison.
- **FixProvider authoring guide** — `docs/guides/fix-providers.md`: Comprehensive guide covering interface, built-in providers, `lineIndexAware` optional interface (when/how to implement), error handling, caching patterns, testing.
- **v1.0.0 breaking changes ADR** — `docs/architecture-decisions.md` #11: Consolidated 6-item plan (Findings unexport, Merge/RecordFix removal, Position.Offset sentinel, SARIF types). Internal migration progress documented (findingsLocked, readFindings, deprecated notices).
- **sync.Pool evaluation** — `docs/architecture-decisions.md` #12: Evaluated sync.Pool for `[]int` line offset index. Decision: SKIP. Index allocation is <0.1% of FixEngine time. Pool complexity (reset, lifecycle) not justified.
- **Benchmark offset bug** — Session 13 fix: `generateOffsetFixes` used Range width 4 for BeforeCode "old()" (5 chars); OffsetProvider always rejected, silently benchmarking SubstringProvider. Fixed to use actual occurrence positions via `findOldOccurrences`

---

_Assisted-by: Crush <crush@charm.land>_
