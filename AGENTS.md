# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. Seven tools detect issues; zero route them to remediation. This library solves that with:

1. **Unified Finding type** — Common representation for all tools
2. **Pipeline** — Automated detect → triage → fix → verify loop
3. **SARIF output** — Standard interchange format
4. **LSP integration** — IDE support

## Module Structure (Multi-Module Go Workspace)

Unix-style decomposition — each module does one thing well, composes via replace directives.

| Module       | Path                                               | External Deps                | Depends on     |
| ------------ | -------------------------------------------------- | ---------------------------- | -------------- |
| **Core**     | `github.com/larsartmann/go-finding`                | go-error-family              | —              |
| **Pipeline** | `github.com/larsartmann/go-finding/pipeline`       | x/sync, gogenfilter          | Core           |
| **Analysis** | `github.com/larsartmann/go-finding/analysis`       | x/tools                      | Core           |
| **CLI**      | `github.com/larsartmann/go-finding/cmd/go-finding` | yaml, go-output, gogenfilter | Core, Pipeline |

`go.work` coordinates all 4 modules for development. Each sub-module has `replace` directives for `GOWORK=off` CI/consumer builds.

## Key Files

| Area                | Files                                                                                                                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Core types**      | `finding.go`, `finding_methods.go`, `finding_validate.go`, `finding_equal.go`, `position.go`, `range.go`, `report.go`, `filter.go`, `merge.go`, `diff.go`, `format.go`, `json.go`, `id.go`, `errors.go`, `simple_fix.go` |
| **Named types**     | `severity.go`, `confidence.go`, `category.go`, `category_linter.go`, `tag.go`, `fix_strategy.go`, `suppression.go`, `branded_types.go`                                                                                   |
| **SARIF**           | `sarif_types.go`, `sarif_export.go`, `sarif_import.go` (hand-rolled, not go-sarif — see ADR #9)                                                                                                                          |
| **LSP**             | `lsp.go`                                                                                                                                                                                                                 |
| **Extensibility**   | `detector.go`, `adapter.go` (ToolAdapter[O]), `registry.go` (DetectorRegistry), `interval_index.go` (IntervalIndex[T])                                                                                                   |
| **gotoken**         | `gotoken/gotoken.go` (shared go/token utilities, public package)                                                                                                                                                         |
| **lockutil**        | `lockutil/lockutil.go` (shared sync.Locker helpers — `Locked`, `RLocked` — for generic mutex-guarded critical sections)                                                                                                  |
| **Pipeline**        | `pipeline/pipeline.go` (Run), `pipeline/pipeline_detect.go`, `pipeline/pipeline_iteration.go`, `pipeline/config.go`, `pipeline/config_file.go`, `pipeline/flight_recorder.go`                                            |
| **Fix engine**      | `pipeline/fix_engine.go`, `pipeline/fix_provider.go`, `pipeline/fix_applier.go`, `pipeline/fix_outcome.go`, `pipeline/fix_edit.go`, `pipeline/conflict.go`, `pipeline/goast/provider.go`                                 |
| **Pipeline extras** | `pipeline/stage_hook.go`, `pipeline/flight_recorder.go`, `pipeline/line_shift.go`, `pipeline/metrics.go`, `pipeline/retry.go`, `pipeline/partial.go`, `pipeline/generated_filter.go`                                     |
| **Analysis**        | `analysis/analysis.go` (go/analysis ↔ Finding)                                                                                                                                                                           |
| **Detectors**       | `cmd/go-finding/internal/detectors/govet.go`, `staticcheck.go`, `helpers.go`                                                                                                                                             |
| **CLI**             | `cmd/go-finding/main.go`, `config.go`, `registry.go`, `fix_provider_registry.go`, `generated_filter.go`, `output_adapter.go`                                                                                             |

## Testing & Build

```bash
nix run .#test                              # Run tests (all modules via go.work)
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector (workspace)
GOWORK=off go test ./...                    # Per-module isolation test (run in each module dir)
golangci-lint run ./...                     # Lint
bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 25  # Benchmark regression check
bash scripts/version-check.sh                                    # Verify version.go matches git tag
```

> **GOEXPERIMENT=jsonv2 required.** The project imports `encoding/json/v2`.
> All `nix run .#*` apps and devShells set this env var automatically. Direct `go`
> commands (outside `nix develop`) require `export GOEXPERIMENT=jsonv2` first — otherwise you get
> "build constraints exclude all Go files" errors.

## Module Dependencies (per go.mod)

| Module                      | Production Deps              | Test Deps         |
| --------------------------- | ---------------------------- | ----------------- |
| **Core** (`.`)              | go-error-family              | ginkgo/v2, gomega |
| **Pipeline** (`pipeline/`)  | x/sync, gogenfilter          | ginkgo/v2, gomega |
| **Analysis** (`analysis/`)  | x/tools                      | (stdlib testing)  |
| **CLI** (`cmd/go-finding/`) | yaml, go-output, gogenfilter | gomega            |

### Old Dependencies (now isolated to sub-modules)

- `golang.org/x/tools` — go/analysis framework (**analysis module only**)
- `golang.org/x/sync` — errgroup (**pipeline module only**)
- `github.com/go-faster/yaml` — YAML config (**CLI module only**)
- `github.com/onsi/ginkgo/v2` + `gomega` — BDD testing (core + pipeline test deps)
- `github.com/LarsArtmann/gogenfilter/v3` — Auto-generated Go file detection (pipeline + CLI)
- `github.com/larsartmann/go-output` — CLI output formatting (CLI module only)

## Design Principles

1. **Minimal dependencies** — core module keeps a small, deliberate dependency surface
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags/Data fields
4. **One extensibility field** — `Finding.Metadata` is `map[string]string`. NO `Properties map[string]any`
5. **Compatible** — Works with existing Go analysis tools
6. **Resilient** — Retry logic, partial success, nil-safe metrics
7. **Multi-module** — Unix-style decomposition: each module has a single purpose and composes independently. Dual go.work + replace strategy.

## Important Behaviors (Gotchas)

- **GOEXPERIMENT=jsonv2 required** — The project uses `encoding/json/v2` (Go 1.26 experimental feature). All `nix run .#*` apps and devShells export `GOEXPERIMENT=jsonv2`. Direct `go build`/`go test` outside `nix develop` will fail with "build constraints exclude all Go files" unless you `export GOEXPERIMENT=jsonv2` first. The `GOWORK=off` per-module path needs BOTH `GOWORK=off` and `GOEXPERIMENT=jsonv2`.
- **Report{} zero-value safe** — Uses value `sync.Mutex`, safe for concurrent use without initialization
- **Report.findings is unexported** — Use `FindingsSnapshot()` for a deep copy, `All()` for iteration, or `FindByID()` for single lookups
- **Pipeline.Run() is single-use** — Returns `errAlreadyRan` on second call
- **NewFixApplier returns error** — Propagates backup dir creation failures
- **Confidence is a named type** — `type Confidence float64` with `IsValid()`/`Clamp()`
- **NewFinding accepts Confidence** — Not raw `float64`. Builder defaults to `ConfidenceFull` (1.0) since v1.3.0, appropriate for deterministic static analysis. Override with `.WithConfidence()`.
- **FixStrategyAI is reserved** — No backend; kept as marker for future AI remediation
- **GenerateID is length-prefixed** — Uses `writeLenField` (uint32 big-endian) to prevent hash collisions when field values contain colons
- **Position.Offset uses -1 sentinel** — `Position{}` (zero value) has Offset=0 meaning "byte 0". Constructors (Pos, NewRange, FromLSP, SARIF import) set Offset=-1 for "unset". Use `HasOffset()` (>= 0) to check.
- **File-level positions valid since v1.3.0** — `validateIdentity()` uses `Position.HasFile()` (File != ""), not `Position.IsValid()` (File != "" && Line > 0). Findings with `FilePos("config.yaml")` (Line=0) pass validation. `Position.IsValid()` still requires Line>0 for backward compat. Use `HasFile()` for file-only checks.
- **Builder.BuildOrDefault()** — Returns zero-value `Finding{}` on validation error, not panic. Eliminates the error-swallowing boilerplate (SafeBuildFinding / buildFinding) that consumers reinvent. Use `Build()` when you need validation errors.
- **Template** — Pre-configured builder factory (`NewTemplate(toolName)` + `WithCategory/WithFixStrategy/WithTags` + `Build(rule, msg, sev, pos)`). Stamp common fields once, build many findings. Eliminates `newMigrationFinding` / `buildFixableFinding` patterns.
- **Template.Builder** — Returns a pre-configured `*Builder` (not a final `Finding`) for per-finding chaining. Use when you need template-level defaults AND per-finding confidence/suggestion: `tmpl.Builder(rule, msg, sev, pos).WithConfidence(c).WithSuggestion(s).MustBuild()`. `Build` delegates to `Builder().BuildOrDefault()`.
- **ParseConfidence** — Inverse of `Confidence.String()`: `ParseConfidence("high")` returns `ConfidenceHigh`. Also accepts decimals (e.g. `"0.42"`). Empty string defaults to `ConfidenceLow`. Eliminates consumer-side switch statements for `--min-confidence` flags. Error is `ErrInvalidConfidence` sentinel, matchable via `errors.Is`.
- **NewReportFromFindings(tool, findings)** — One-step report creation: `NewReport` + `AddFindings` + `ComputeSummary`. Eliminates the 4-line boilerplate.
- **SeverityFromLevel(level, fallback)** — Maps severity strings (canonical + aliases) to `Severity`, returns fallback for unknown. Eliminates consumer-side `mapSeverity()` switches. Aliases expanded: "optional"→Info, "crit"→Critical added.
- **Severity.PriorityString()** — Reverse mapping: Critical→"critical", Error→"high", Warning→"medium", Info→"low".
- **FormatTextRich** — New rich text formatter with emoji severity badges, category display, and 💡 suggestion prefix. `FormatText` retains the original `[SEVERITY]` format for backward compatibility.
- **FormatTable(w, findings)** — Severity-badged table output (SEVERITY, LOCATION, RULE, MESSAGE columns).
- **ApplySimpleFixes(findings)** — BeforeCode→AfterCode string replacement in core package. 80% case for consumers that don't need the full pipeline FixEngine.
- **CheckBinary(name) / RunCmd(ctx, name, args)** — External tool helpers for the "run CLI tool → parse JSON" pattern. Returns `NewIOError` on failure.
- **DefaultLinterRegistry expanded** — Now includes gofumpt, nolintlint, depguard, nakedret, bidichk, tagliatelle, and 15+ more golangci-lint linters.
- **Range.EndOrStart / EndOffsetOrStart** — Effective end position for line-based (`End.Line == 0` → Start) and offset-based (`End.Offset < 0` → Start) single-point ranges. Used by overlap/intersection/extension to avoid duplicating the "unset means single point" convention.
- **FixStrategy normalized** — `NormalizeFixStrategy()` converts "" to "none". Called by Builder.Build(), SARIF import, and Equal(). In Equal(), normalization is short-circuited: raw values are compared first, and `NormalizeFixStrategy` is only called when they differ (handles "" vs "none" edge case).
- **tagsEqual fast path** — `finding_equal.go` checks `slices.Equal(a, b)` before clone+sort. When tags are in the same order (common when findings come from the same tool), `Equal()` makes 0 allocations. Only different-order tag sets trigger the clone+sort fallback (2 allocs).
- **ResolveSafePath batch caching** — `pipeline/path_safety.go` exports `ResolveRoot(rootDir)` (resolves symlinks once) and `ResolveSafePathFrom(resolvedRoot, relPath)` (per-path resolution). `groupFindingsBySafePath` and `filterByFileEdits` resolve root once and cache per-path results, eliminating redundant `EvalSymlinks` syscalls when many findings target the same file. The convenience wrapper `ResolveSafePath(rootDir, relPath)` remains for single-call use.
- **HasFix() requires code for Direct** — `FixStrategyDirect` needs BeforeCode or AfterCode for HasFix()=true, aligning with Validate().
- **math/rand v1/v2 split** — Production uses `math/rand/v2`; tests use `math/rand` (v1) due to `testing/quick` API constraint
- **FixEngine descending-offset** — All edits resolve against the same original content snapshot; multi-edit correctness proven by tests
- **LineShiftMap shifts Position + Range** — `ShiftedPosition` shifts line + column (single-line edits); `ShiftedRange` shifts both endpoints
- **SubstringProvider column-aware** — Disambiguates multiple occurrences by line + column distance
- **context.Context on I/O** — `WriteSARIF`, `FindingsFromSARIF`, etc. accept context as first arg
- **StageHooks replace OnStage** — Use `Config.StageHooks` with `StageHook`/`StageHookFunc` for before/after events with abort capability. Both StageBefore and StageAfter errors abort the pipeline.
- **IsSuppressedAt validates Suppression** — Uses `Suppression.IsActive(now)` which checks `IsValid()` (valid Kind + non-empty Rule) AND not expired. Invalid suppressions are treated as inactive.
- **Branded types prevent mixups** — `ID`, `RuleName`, `ToolName`, `FilePath` are distinct string types. Use `finding.ID("x")` not raw `"x"` for fields. JSON marshals identically to string.
- **Validate() decomposed** — `finding_validate.go` delegates to 6 per-field validators (`validateIdentity`, `validateClassification`, `validateFix`, `validateReferences`, `validateSpatial`, `validateSuppression`). Complexity per validator < 10.
- **SeverityAliases removed** — Use `RegisterSeverityAlias()` / `LookupSeverityAlias()`. Global map guarded by `sync.RWMutex`.
- **CategoryOf is canonical** — `CategoryOf(err)` returns the category (old `GetCategory` removed)
- **testify in go.mod is transitive** — `stretchr/testify` appears as `// indirect` in core go.mod because ginkgo/slim-sprig depends on it. It is NOT used directly. Banned per how-to-golang but unavoidable as a transitive dep of ginkgo.
- **Position.File is FilePath** — Changed from `string` to `FilePath` branded type. Use `finding.FilePath("path")` for string vars; string literals auto-convert. Constructors `Pos`, `NewRange`, `NewRangePtr` accept `FilePath`.
- **SARIFOption pattern** — Use `ToSARIFWithOpts(WithIncludeSuppressed(), WithMinSeverity(sev))` instead of deprecated `ToSARIFFiltered`. Suppressed findings can now be emitted with SARIF suppression arrays.
- **LSPDiagnosticData** — `ToLSP()` populates `diag.Data` with ID, Severity, FixStrategy, Confidence, Category, Tags, code data, Snippet, Suppression, Metadata, and RelatedFindingIDs. `FromLSP` restores them. Round-trip is lossless including SeverityCritical (which LSP collapses to Error) and RelatedRef.FindingID (which is preserved instead of regenerated).
- **Analysis BeforeCode** — `analysis.FromDiagnostic` now extracts `BeforeCode` from TextEdits by reading source file from disk.
- **GroupByFile returns map[FilePath][]Finding** — Updated to use branded type as map key.
- **lockutil.Locked/RLocked for mutex boilerplate** — Generic helpers `lockutil.Locked(sync.Locker, fn)` and `lockutil.RLocked(*sync.RWMutex, fn)` consolidate the m.mu.Lock()/defer m.mu.Unlock() pattern. Returns generic T; use `struct{}` for side-effect-only sections. Report/metrics/file_backup/registry/category_linter/etc. all use these.
- **makezero is `always: false` (intentional)** — `.golangci.yml` sets makezero `always: false`. The `always: true` mode flags the idiomatic `make([]T, len) + copy()` pattern as wrong (23 false positives). The `false` mode still catches the real bug: `make([]T, n) + append` (over-allocation). Idiomatic Go wins. One-line revert in `.golangci.yml` if append-only style is ever desired.
- **StageTiming closure must be called exactly once** — `Metrics.RecordStage` uses `+=` (`metrics.go:50`), so calling the closure returned by `stageTiming(stage)` more than once double-records the duration. When wrapping a stage that has success AND error/hook-abort paths, invoke the done-closure on exactly one path. Previous bug: `pipeline_iteration.go` called `applyDone()` on both the success path and the hook-error path.
- **RetryConfig validation uses named sentinels** — `pipeline/retry.go:20-25` defines 6 named sentinel errors (`errMaxRetriesNegative`, `errBaseDelayPositive`, etc.). Consumers can `errors.Is(err, errBaseDelayPositive)`. NEVER inline `errors.New("...")` in validation returns — it breaks `errors.Is()` matching. Any new validation rule must add a named sentinel var.
- **version-check.sh uses `--match 'v[0-9]*'`** — `git describe --tags --abbrev=0` without the match pattern picks up sub-module directory-prefixed tags (e.g. `analysis/v1.3.0`) alphabetically before the core `v*` tag. The `--match 'v[0-9]*'` filter ensures only core tags are matched. Any script that resolves the core version from git tags must use this flag.
- **doc.go API references must match current names** — `doc.go` contains godoc prose that references API symbols by name. After ANY rename (e.g. `GetCategory` → `CategoryOf`), grep `doc.go` for the old name. Stale references in godoc mislead consumers reading the package documentation.
- **FindingError implements go-error-family interfaces** — `ErrorCode()` returns `"finding.<category>"`; `ErrorFamily()` maps ErrorCategory to errorfamily.Family (Validation/Parse→Rejection, Conflict→Conflict, IO→Transient, Internal→Infrastructure). Consumers can call `errorfamily.Classify(err)` on go-finding errors. See ADR #15.
- **Consumer count is from 2026-07-22 audit** — The v1.0.0 audit (`docs/reviews/archived/2026-07-05_20-55_consumer-audit.html`, 2026-07-05) counted 20 consumers. The v1.3.0-era audit (`docs/planning/archived/2026-07-22_17-56_consumer-driven-api-improvements.md`, 2026-07-22) counted 22 consumers (14 with Go code). The count grows over time — do not assert a specific number without checking the latest data.
- **must[T] eliminates panic-on-error boilerplate** — `errors.go` defines `func must[T any](v T, err error) T` used by `MustParseCategory`, `MustParseSeverity`, and `Builder.MustBuild`. Any new Must-style constructor should delegate to `must(...)`, not repeat the `if err != nil { panic(err) }` pattern.
- **NewParallelGomega consolidates test setup** — Each module's `testutil_test.go` defines `NewParallelGomega(t *testing.T) *gomega.GomegaWithT` which calls `t.Helper()`, `t.Parallel()`, and `gomega.NewWithT(t)`. Test functions should use `g := NewParallelGomega(t)` instead of the two-line `t.Parallel(); g := NewWithT(t)` boilerplate. The `paralleltest` linter is disabled in `.golangci.yml` because it cannot trace `t.Parallel()` through the helper.
- **paralleltest linter disabled (intentional)** — `.golangci.yml` disables `paralleltest` because `t.Parallel()` is centralized inside `NewParallelGomega` helpers across all 4 modules. The linter cannot trace calls through wrapper functions. Re-enabling requires either inlining `t.Parallel()` everywhere or adding `//nolint:paralleltest` to 100+ test functions.
- **marshalJSONString centralizes marshal-to-string** — `json.go` defines `marshalJSONString(bytes []byte, err error) (string, error)` which wraps marshal errors and returns the string. Used by `Report.JSON()`, `PrettyJSON()`, `PrettyJSONFiltered()`, and `Finding.LineJSON()`.
- **marshalOpts / prettyMarshalOpts centralize deterministic JSON options** — `json.go` defines two package-level variables: `marshalOpts = json.Deterministic(true)` (compact) and `prettyMarshalOpts` (deterministic + 2-space indent). All production `MarshalJSON`, `PrettyJSON`, `PrettyJSONFiltered`, `WriteJSON`, `ToSARIFWithOpts`, `WriteSARIFWithOpts`, `LineJSON`, and `Finding.WriteJSON` use these. `encoding/json/v2` serializes Go map keys in unspecified order by default — without `Deterministic(true)`, output is non-reproducible. The CI script `json-deterministic-check.sh` enforces this at code level. Any new marshal call in production code MUST use `marshalOpts` or `prettyMarshalOpts`.
- **Badge() derives from Emoji()** — `Severity.Badge()` returns `Emoji() + " " + strings.ToUpper(string(s))` instead of a duplicate switch. Any new emoji-to-severity mapping only needs to update `Emoji()`.
- **severityPriorities map replaces switch** — `Severity.PriorityString()` uses a `map[Severity]string` lookup (`severityPriorities`) instead of a switch statement. Add new severity priorities to the map, not a new case.
- **fixEditJSON is the JSON wire type** — `pipeline/fix_edit.go` defines `fixEditJSON` at package level (not as inner types in Marshal/UnmarshalJSON) to control field tags independently of the `FixEdit` domain type.
- **FlightRecorderHook wraps Go 1.25 runtime/trace.FlightRecorder** — `pipeline/flight_recorder.go` provides `FlightRecorderHook` implementing `StageHook`. It continuously buffers Go execution trace data and captures snapshots when stages exceed `SlowStageThreshold` or on manual `Snapshot(ctx, reason)`. `OnStageEvent` never returns errors (diagnostic-only). If another flight recorder is already active (Go singleton limit), `NewFlightRecorderHook` enters degraded mode: returns `(hook, nil)` with `Degraded()` true, all snapshots silently skipped. Check `Degraded()` to detect this. `Snapshot` accepts `context.Context` — cancelled contexts skip the write. `asyncSnapshot` uses `context.Background()` internally (pipeline ctx may be cancelled before goroutine executes). `writeMu` serializes concurrent `WriteTo` calls. `Close()` waits for in-flight snapshots before `fr.Stop()`. CLI flags: `-trace`, `-trace-dir`, `-trace-slow`. See https://go.dev/blog/flight-recorder.
- **ResolveSafePath is the path traversal security boundary** — `pipeline/path_safety.go` exports `ResolveSafePath(rootDir, relPath)` which resolves symlinks and verifies the result stays within `rootDir`. This is the single source of truth for path containment, used by FixApplier and Pipeline to prevent path traversal attacks (e.g., `Position.File = "../../etc/passwd"`). All file paths from Finding data pass through this before any filesystem operation. Consumers can use `ResolveSafePath` / `ResolveSafePathFrom` / `ResolveRoot` for their own path validation.
- **FlightRecorderFileConfig / ResolveFlightRecorder** — `pipeline/config_file.go` defines `FlightRecorderFileConfig` (5 string-encoded fields: `Enabled`, `OutputDir`, `SlowStageThreshold`, `MinAge`, `MaxBytes`) as the JSON-friendly config representation. `ConfigFile.ResolveFlightRecorder()` constructs a `*FlightRecorderHook` from the config section, returning `(nil, nil)` when `FlightRecorder` is nil or `Enabled` is false. Duration fields use string syntax (e.g. `"30s"`) parsed via `time.ParseDuration`, matching the convention used elsewhere in `ConfigFile`.
- **CLI config-file flight recorder fallback** — `cmd/go-finding/config.go` has its own `flightRecorderFileConfig` (5 fields: `Enabled`, `OutputDir`, `SlowStageThreshold`, `MinAge`, `MaxBytes`) — now at full parity with the pipeline's `FlightRecorderFileConfig`. The CLI uses config-file flight recorder settings when `-trace` is not set on the command line. Both structs must be kept in sync when new fields are added.
- **CI scripts guard multi-module architecture** — Seven scripts in `scripts/` protect the 4-module structure: `replace-audit.sh` (verifies all replace directives point to `../`), `version-drift.sh` (cross-checks that all go.mod files reference the same core version), `test-naming.sh` (enforces one-test-file-per-production-file + no `_extra`/`_bugfix`/`coverage` suffixes), `go-work-sync.sh` (verifies go.work entries match directory structure), `docs-freshness.sh` (flags docs not modified in 180 days and docs whose referenced `.go` files changed since the doc was last updated; matches only backtick code spans and markdown links, not prose mentions), `json-deterministic-check.sh` (enforces `json.Deterministic(true)` on all production marshal calls). Additionally, `go-arch-lint` (via `.go-arch-lint.yml`) enforces one-directional package dependency flow across all 4 modules. All wired into `.github/workflows/ci.yml` as separate jobs.
- **Finding.GroupID groups related findings** — `GroupID GroupID` (branded type in `branded_types.go`) marks findings as members of a logical set (e.g., a clone group for art-dupl). Optional; JSON key `groupId` (omitempty); included in `Equal()`; builder `WithGroupID`; SARIF round-trips it as property `go-finding/groupId`; LSP round-trips it via `LSPDiagnosticData.GroupID`. `Report.GroupFindings()` returns `map[GroupID][]Finding` over active (non-suppressed) grouped findings, nil when none. art-dupl integration gaps are tracked with implementation status in `docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`.
- **ToLSP re-emits LSP diagnostic tags from metadata** — `LSPDiagnostic.Tags` is populated by `ToLSP()` from `Metadata[LSPDiagnosticTagsKey]` (comma-separated ints, as written by `FromLSP`). A FromLSP-created finding therefore keeps its tags on every subsequent ToLSP conversion, not just through one round-trip. Malformed entries in the metadata value are silently skipped.
- **FixEngine reports per-finding outcomes** — `FixEngine.ApplyWithOutcomes(content, fixes)` returns a `FixApplyResult` with one `FixOutcome` per input finding (input order): `applied` / `no-change` / `refused` (matched provider produced zero edits without error — previously indistinguishable from success, issue #27) / `conflict` / `invalid` (edit dropped as invalid or out of bounds) / `failed` (provider error in `Err`). `Apply` and `ApplyWithConflicts` delegate to it, so their return shapes and behavior are unchanged.
- **FixApplier rollback is per-file by default** — `RollbackPolicyFailingFile` (default, zero value) restores only the failing file; earlier files keep their applied fixes. Soft per-finding failures (provider resolve errors, refused findings) never abort the run: `applyToFile` writes valid applied edits and reports resolve errors via outcomes plus the joined error return. `RollbackPolicyAllFiles` preserves the legacy all-or-nothing rollback (set via `SetRollbackPolicy`, `Config.FixRollbackAllFiles`, or config-file `fixRollbackAllFiles`). `ApplyWithReport` returns an `ApplyReport` (applied fixes, outcomes, shift maps, rolled-back files). Before this change, one provider resolve error on the last file discarded all clean edits in all previous files (issue #28).

## CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, CSV, TSV, JSON, SARIF output (markdown/CSV/TSV via go-output adapter)
- YAML/JSON config (`-config`), severity filter (`-min-severity`), profiling
- `-filter-generated` — removes findings from auto-generated files (sqlc, protobuf, etc.)
- `-fix-provider go-ast` — enables AST-aware fix provider
- `-byte-level-conflict` — precise overlap detection
- `-trace` — enable Go execution trace flight recorder for diagnostics (`-trace-dir`, `-trace-slow` for config)
- Config-file `flightRecorder` section — alternative to `-trace` flags; supports `enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes` (full parity with pipeline `FlightRecorderFileConfig`)
- Dynamic detector registry (`RegisterDetector`)

## Architecture Decisions

- **SARIF hand-rolled** — Not go-sarif. No SARIF library dependency, custom property bag, streaming + context. See ADR #9.
- **go-error-family integration** — Core module depends on `go-error-family` for unified error classification. `FindingError` implements `Coded` + `Classified`. See ADR #15.
- **Pipeline split** — `pipeline.go` + `pipeline_detect.go`, both under 350 lines
- **Byte-level FixEngine** — `[]byte` edit ops with descending-offset application, O(F+R) single-pass
- **FixProvider chain** — OffsetProvider → LineProvider → SubstringProvider (fallback); custom providers prepended
- **lineIndexAware lazy caching** — Line offset index built once per file, only when a LineProvider/SubstringProvider handles a finding
- **ResolveSafePath batch caching** — Root symlink resolution cached once per batch; per-path results cached within `groupFindingsBySafePath` to avoid redundant `EvalSymlinks` when many findings target the same file. Functions now exported for consumer use: `ResolveSafePath`, `ResolveSafePathFrom`, `ResolveRoot`
- **GoASTProvider** — AST-aware provider in `pipeline/goast/` (opt-in `go/parser` dependency)
- **IntervalIndex[T]** — Generic O(n + k) overlap queries (sorted-slice impl); used by Correlate
- **DetectorRegistry** — Thread-safe plugin architecture with `Register`/`Build`/`BuildAll`
- **MergeIter** — Streaming `iter.Seq[Finding]` merge with dedup
- **ConfigFile** — JSON config loading with `ResolveDetectors`/`ResolveProviders`
- **go-output CLI adapter** — `cmd/go-finding/output_adapter.go` adapts `[]Finding` → `output.Table` for markdown/CSV/TSV. Root `finding` package stays dependency-free. See `docs/PRO_CONTRA_go-output-integration.md`.
- **Branded primitive types** — `type ID/RuleName/ToolName/FilePath string` in `branded_types.go`. Compile-time type safety preventing ID/Rule/Tool/File mixups. JSON marshals as string. Named `ID` not `FindingID` to avoid revive stutter (`finding.FindingID`).
- **Validate decomposition** — Monolithic `Validate()` split into 6 per-field validators. Each returns `[]error`, aggregated by `Validate()`.
- **SeverityAliases thread-safe** — `sync.RWMutex` guarded global map; `RegisterSeverityAlias()` / `LookupSeverityAlias()` API.
- **FindingTransformer** — Renamed from `FindingProcessor`/`Process()`. Pipeline uses `Config.Processors []FindingTransformer`.
- **Conflict** — Renamed from `ConflictInfo`. `AnalyzeConflicts() []Conflict`.
- **SARIFOption pattern** — Functional options (`WithIncludeSuppressed`, `WithMinSeverity`) for SARIF export. `ToSARIF`/`WriteSARIF` delegate to `ToSARIFWithOpts`/`WriteSARIFWithOpts`. Deprecated `ToSARIFFiltered`/`WriteSARIFFiltered` retained as aliases.
- **LSP data fidelity** — `LSPDiagnosticData` struct on `LSPDiagnostic.Data` preserves go-finding-specific fields (ID, FixStrategy, Confidence, Category, Tags, code data) through LSP round-trip.
- **CLI flag rename** — `-severity` renamed to `-min-severity` with deprecated alias retained for backward compat.
- **Analysis BeforeCode extraction** — `analysis.FromDiagnostic` reads source files to extract `BeforeCode` from TextEdit ranges, enabling full fix data on go/analysis findings.
- **Pipeline convenience functions** — `pipeline.Detect(ctx, detectors...)` for one-shot detection; `pipeline.ApplyToContent(content, fixes)` for content-level fix application without filesystem.
- **FixEngine guide** — `docs/guides/fix-engine.md` covers all FixEngine usage patterns.
- **lockutil package** — Generic `lockutil.Locked(sync.Locker, fn) T` and `lockutil.RLocked(*sync.RWMutex, fn) T` helpers eliminate m.mu.Lock()/defer m.mu.Unlock() boilerplate across Report, Metrics, FileBackup, registries, and AST provider. Stdlib only, follows `gotoken` precedent.
- **FlightRecorderHook** — Wraps Go 1.25 `runtime/trace.FlightRecorder` as a `StageHook`. Stdlib-only (no external dep). Continuously buffers execution trace; snapshots on slow-stage threshold breach or manual `Snapshot(reason)`. `OnStageEvent` never returns errors (diagnostic-only). `Close()` waits for in-flight snapshots before `fr.Stop()` to avoid `WriteTo`/`Stop` data race. See https://go.dev/blog/flight-recorder.
- **Multi-module release tagging** — Each sub-module needs a **directory-prefixed** git tag to resolve on the Go proxy: `pipeline/v*`, `analysis/v*`, `cmd/go-finding/v*`. Core uses unprefixed `v*`. Sub-modules have no `version.go`; the tag is the version source. See `docs/release-procedure.md`.
- **Repo is private** — Until made public, consumers MUST set `GOPRIVATE=github.com/larsartmann/go-finding` or module resolution 404s on the public proxy.
- **go-arch-lint boundary enforcement** — `.go-arch-lint.yml` (v3 format) defines 11 components (core, gotoken, lockutil, examples, pipeline, pipeline-goast, pipeline-internal, pipeline-examples, analysis, cli, cli-detectors). Dependency flow is one-directional: cli → pipeline → core, analysis → core. Test files excluded via `excludeFiles: ["_test\.go$"]`. Run locally with `go-arch-lint check`. Wired into ci.yml `arch-check` job.

## Test Organization

**One test file per production file.** No `_extra_test.go`, `_bugfix_test.go`, or `coverage_test.go` files. All tests for a subject live in `<subject>_test.go`.

- **No `_extra` suffix files** — If a test file gets too large, split by subject (e.g., `finding_test.go` + `finding_validate_test.go`), not by arbitrary "\_extra" suffix
- **No `_bugfix` suffix files** — Regression tests belong in the parent test file alongside the behavior they protect
- **No `coverage_test.go`** — Never name files after a metric. Name them after what they test (`validate_test.go`, `config_file_test.go`)
- **Shared helpers** go in `testutil_test.go` (or `assert_extra_test.go` folded into it)
- **Examples** — One `example_test.go` per package; don't fragment into `example_basic_test.go` + `example_cli_test.go` + `example_extra_test.go`

## Removed APIs (v1.0.0)

All deprecated APIs from v0.6.0–v0.9.0 have been removed. No deprecated APIs remain.

See `docs/MIGRATION_v1.0.md` for migration details.

---

_Assisted-by: Crush <crush@charm.land>_
