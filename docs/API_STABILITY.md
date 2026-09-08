# API Stability — go-finding

**Last audited:** 2026-09-08
**Version:** v1.7.0

go-finding follows the [Go 1 Compatibility Promise](https://go.dev/doc/go1compat) philosophy.

## Commitment

Since v1.0.0:

1. **No breaking changes** in any minor or patch release (v1.x.y)
2. Breaking changes require a **major version bump** (v2.0.0)
3. All exported symbols are part of the public API

All exported symbols are classified as:

- **stable** — API is locked. Will not change without a major version bump.
- **unstable** — API may change between minor versions.
- **deprecated** — Scheduled for removal. Use the documented alternative.
- **reserved** — Published constant with no runtime implementation.

## Root Package (`finding`) — Stable

### Core Types

| Type       | Status | Notes                                                                                                   |
| ---------- | ------ | ------------------------------------------------------------------------------------------------------- |
| `Finding`  | stable | Core data type. Fields may be added but never removed or renamed.                                       |
| `Report`   | stable | Thread-safe container. `findings` slice is unexported; use `FindingsSnapshot()`, `All()`, `FindByID()`. |
| `Position` | stable | File/Line/Column/Offset. Zero-value is valid ("unpositioned").                                          |
| `Range`    | stable | Start/End Position. `IsValid()` checks End >= Start.                                                    |
| `Builder`  | stable | Fluent Finding builder. `BuildOrDefault()` returns zero-value on invalid.                               |
| `Template` | stable | Pre-configured builder factory (v1.3.0). Stamp common fields, build many findings.                      |

### Named Types

| Type               | Status | Notes                                             |
| ------------------ | ------ | ------------------------------------------------- |
| `Severity`         | stable | Info, Warning, Error, Critical                    |
| `Confidence`       | stable | float64 with `IsValid()`/`Clamp()`/`Compare()`    |
| `Category`         | stable | Standard categories (Correctness, Security, etc.) |
| `Tag`              | stable | Classification tags                               |
| `FixStrategy`      | stable | Direct, Suggest, AI (reserved)                    |
| `ErrorCategory`    | stable | Validation, Parse, IO, Conflict, Internal         |
| `SuppressionKind`  | stable | FalsePositive, WonFix, etc.                       |
| `CorrelationScore` | stable | float64 [0.0, 1.0]                                |
| `DeduplicateBy`    | stable | Enum: ID, Position, Rule                          |
| `RelationKind`     | stable | Related, Causes, IsCausedBy                       |
| `LSPSeverity`      | stable | Named int: Error/Warning/Info/Hint                |
| `LSPDiagnosticTag` | stable | Named int: Unnecessary(1), Deprecated(2)          |
| `ID`               | stable | Branded string type (v1.2.0). Prevents ID mixups. |
| `RuleName`         | stable | Branded string type (v1.2.0)                      |
| `ToolName`         | stable | Branded string type (v1.2.0)                      |
| `FilePath`         | stable | Branded string type (v1.2.0)                      |

### Supporting Types

| Type               | Status | Notes                                             |
| ------------------ | ------ | ------------------------------------------------- |
| `ToolInfo`         | stable | Name + Version                                    |
| `Summary`          | stable | Aggregated statistics. New fields may be added.   |
| `Suppression`      | stable | Kind/Reason/ExpiresAt                             |
| `Correlation`      | stable | Score + finding pair                              |
| `ModifiedPair`     | stable | Before/After from Diff                            |
| `ParsedID`         | stable | ID parse result                                   |
| `RelatedRef`       | stable | Related finding reference                         |
| `DiffResult`       | stable | Diff comparison result                            |
| `SimpleFixResult`  | stable | Result of ApplySimpleFixes (v1.3.0)               |
| `Interval[T]`      | stable | Generic interval for overlap queries              |
| `IntervalIndex[T]` | stable | Generic O(n+k) interval index for overlap queries |
| `DetectorRegistry` | stable | Thread-safe plugin registry for detectors         |
| `LinterRegistry`   | stable | Linter→category mapping registry (84 linters)     |
| `SARIFOption`      | stable | Functional option for SARIF export                |

### Interfaces and Functions

| Symbol              | Status | Notes                                                |
| ------------------- | ------ | ---------------------------------------------------- |
| `Detector`          | stable | Interface: `Name()` + `Detect(ctx)`                  |
| `DetectorFunc`      | stable | `func(ctx) ([]Finding, error)` implementing Detector |
| `NamedDetectorFunc` | stable | Wrap DetectorFunc with name                          |
| `FilterFunc`        | stable | `func(Finding) bool`                                 |
| `ToolAdapter[O]`    | stable | Generic tool→Finding converter                       |
| `MergeOption`       | stable | Functional option                                    |
| `MergeOptions`      | stable | Deduplicate + DeduplicateBy                          |

### LSP Wire Types

| Type            | Status | Notes               |
| --------------- | ------ | ------------------- |
| `LSPDiagnostic` | stable | Full LSP diagnostic |
| `LSPPosition`   | stable |                     |
| `LSPRange`      | stable |                     |
| `LSPLocation`   | stable |                     |
| `LSPRelated`    | stable |                     |

### Error Types

| Symbol               | Status | Notes                                         |
| -------------------- | ------ | --------------------------------------------- |
| `FindingError`       | stable | Structured error with Category/Cause          |
| `NewValidationError` | stable | Constructor                                   |
| `NewParseError`      | stable | Constructor                                   |
| `NewIOError`         | stable | Constructor                                   |
| `NewConflictError`   | stable | Constructor                                   |
| `NewInternalError`   | stable | Constructor                                   |
| `CategoryOf`         | stable | Extract ErrorCategory from error              |
| `IsCategory`         | stable | Check error category                          |
| `IsFindingError`     | stable | Type assertion                                |
| `ErrorCode()`        | stable | go-error-family Coded interface (v1.4.0)      |
| `ErrorFamily()`      | stable | go-error-family Classified interface (v1.4.0) |

### Key Functions

| Function                                                  | Status | Notes                                                        |
| --------------------------------------------------------- | ------ | ------------------------------------------------------------ |
| `Combine`                                                 | stable | Merge reports with dedup                                     |
| `Correlate`                                               | stable | Find related findings                                        |
| `Diff`                                                    | stable | Compare finding sets by ID                                   |
| `Filter` / `FilterInPlace`                                | stable | Filter by predicates                                         |
| `GenerateID`                                              | stable | Deterministic ID                                             |
| `ParseID`                                                 | stable | Parse generated ID                                           |
| `FindingsFromSARIF`                                       | stable | SARIF import                                                 |
| `FindingsFromReader`                                      | stable | Streaming SARIF import                                       |
| `FindingsFromJSON`                                        | stable | JSON import                                                  |
| `ReportFromJSON`                                          | stable | JSON → Report                                                |
| `FromJSON`                                                | stable | JSON → Finding                                               |
| `FromLSP`                                                 | stable | LSP → Finding                                                |
| `FormatText` / `FormatMarkdown`                           | stable | Human-readable output                                        |
| `FormatTextRich` / `FormatTable`                          | stable | Emoji-badged / table output (v1.3.0)                         |
| `ApplySimpleFixes`                                        | stable | BeforeCode→AfterCode replacement (v1.3.0)                    |
| `CheckBinary` / `RunCmd`                                  | stable | External tool helpers (v1.3.0)                               |
| `SeverityFromLevel`                                       | stable | String→Severity with aliases (v1.3.0)                        |
| `NewReportFromFindings`                                   | stable | One-step report creation (v1.3.0)                            |
| `FilePos`                                                 | stable | File-level Position constructor (v1.3.0)                     |
| `ParseSeverity` / `MustParseSeverity`                     | stable | String → Severity                                            |
| `ParseConfidence`                                         | stable | String → Confidence; inverse of Confidence.String() (v1.5.0) |
| `RegisterSeverityAlias` / `LookupSeverityAlias`           | stable | Thread-safe severity alias registry (v1.5.0)                 |
| `ResolveSafePath` / `ResolveSafePathFrom` / `ResolveRoot` | stable | Path-traversal-safe path resolution (v1.5.0)                 |
| `ValidateAll`                                             | stable | Batch-validate findings (returns map[int]error)              |
| `ParseCategory` / `MustParseCategory`                     | stable | String → Category                                            |
| `CategoryForLinter`                                       | stable | Linter→category lookup                                       |
| `RegisterLinterCategory`                                  | stable | Register linter mapping                                      |
| `NewToolAdapter`                                          | stable | Generic adapter constructor                                  |
| `FromSARIFLevel`                                          | stable | SARIF level → Severity                                       |

### Report Methods

| Method                              | Status | Notes                           |
| ----------------------------------- | ------ | ------------------------------- |
| `Report.MergeInto`                  | stable | Returns new Report, no mutation |
| `Report.AddFinding` / `AddFindings` | stable | Thread-safe                     |
| `Report.All`                        | stable | `iter.Seq[Finding]`             |
| `Report.ActiveFindings`             | stable | Non-suppressed                  |
| `Report.ComputeSummary` / `At`      | stable |                                 |
| `Report.FindByID` / `FindByRule`    | stable |                                 |
| `Report.FindingsSnapshot`           | stable | Deep-cloned                     |
| `Report.Validate`                   | stable |                                 |
| `Report.ToSARIF` / `WriteSARIF`     | stable |                                 |
| `Report.PrettyJSON` / `Filtered`    | stable |                                 |
| `Report.Filter` / `Map`             | stable | Transform                       |
| `Report.Len` / `CountBySeverity`    | stable | Counts                          |

### Filter Constructors

| Constructor                                                                                                                           | Status |
| ------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| `ByCategory`, `ByConfidence`, `ByConfidenceAtLeast`, `ByFile`, `ByFixStrategy`, `ByRule`, `BySeverity`, `BySeverityAtLeast`, `ByTool` | stable |
| `AnyOf`, `Negate`, `FilterInvalid`, `NotSuppressed`, `WithFix`, `WithSuggestion`                                                      | stable |

### Constants

| Constant                                    | Status       | Notes                                       |
| ------------------------------------------- | ------------ | ------------------------------------------- |
| `Severity*` (Info/Warning/Error/Critical)   | stable       |                                             |
| `Category*` (Correctness/Security/etc.)     | stable       |                                             |
| `FixStrategyDirect/Suggest`                 | stable       |                                             |
| `FixStrategyAI`                             | **reserved** | No backend. Won't auto-apply.               |
| `RelationKind*`, `SuppressionKind*`, `Tag*` | stable       |                                             |
| `VersionMajor/Minor/Patch`, `Version`       | stable       |                                             |
| `OffsetUnknown`                             | stable       | Sentinel for unset Position offset (v1.2.0) |
| `Severity.Badge()` / `PriorityString()`     | stable       | Methods on Severity (v1.3.0)                |
| `LSPSeverity*`, `LSPDiagnosticTag*`         | stable       |                                             |
| `LSPSeverityKey`, `LSPDiagnosticTagsKey`    | stable       | Metadata keys                               |

### Functions

| Function                | Status | Notes                                        |
| ----------------------- | ------ | -------------------------------------------- |
| `RegisterSeverityAlias` | stable | Add custom severity alias                    |
| `LookupSeverityAlias`   | stable | Look up canonical Severity                   |
| `CategoryOf`            | stable | Category from error (replaces `GetCategory`) |

---

## Pipeline Package (`pipeline`) — Stable

### Types

| Type                                                                     | Status                               |
| ------------------------------------------------------------------------ | ------------------------------------ |
| `Pipeline`, `Config`, `CompletionReason`, `Stage`                        | stable                               |
| `Iteration`, `PipelineResult`, `PartialResult`                           | stable                               |
| `FixApplier`, `FixEngine`, `FixEdit`, `FixProvider`, `FixGroup`          | stable                               |
| `Conflict`, `VerifyResult`, `TriageResult`                               | stable                               |
| `Metrics`, `MetricsSnapshot`                                             | stable                               |
| `FileBackup`                                                             | stable                               |
| `RetryDetector`, `RetryConfig`                                           | stable                               |
| `GeneratedFileFilter`                                                    | stable                               |
| `FindingTransformer`, `TransformerFunc`                                  | stable                               |
| `Detector`, `DetectorFunc` (type aliases)                                | stable                               |
| `OffsetProvider`, `LineProvider`, `SubstringProvider`                    | stable                               |
| `StageHook`, `StageHookFunc`, `StageEvent`, `StageTiming`                | stable                               |
| `LineShiftMap`, `LineShiftEntry`                                         | stable                               |
| `ConfigFile`                                                             | stable                               |
| `FlightRecorderHook`, `FlightRecorderConfig`                             | stable                               |
| `FixOutcome`, `FixOutcomeStatus`, `FixApplyResult`                       | stable (v1.7.0) |
| `RollbackPolicy` (`RollbackPolicyFailingFile`, `RollbackPolicyAllFiles`) | stable (v1.7.0) |

### Functions

All exported functions in `pipeline` are **stable**.

Notable additions:

| Function                                     | Status | Notes                                                                              |
| -------------------------------------------- | ------ | ---------------------------------------------------------------------------------- |
| `NewFlightRecorderHook`                      | stable | Added v1.5.0. Constructor for FlightRecorderHook.                                    |
| `DefaultFlightRecorderConfig`                | stable | Added v1.5.0. Returns default config.                                                |
| `Detect`                                     | stable | One-shot detection convenience function (v1.3.0)                                   |
| `ApplyToContent`                             | stable | Content-level fix application without FS (v1.3.0)                                  |
| `ConfigFromFile` / `ConfigFromReader`        | stable | JSON/YAML config file loading                                                      |
| `FixEngine.ApplyWithOutcomes`                | stable | Added v1.7.0. Per-finding outcomes; Apply/ApplyWithConflicts delegate to it |
| `FixApplier.ApplyWithReport` / `ApplyReport` | stable | Added v1.7.0. Run report with outcomes, shift maps, RolledBack files        |
| `FixApplier.SetRollbackPolicy`               | stable | Added v1.7.0. Per-file rollback default; AllFiles opt-in                    |
| `ApplyReport.FailedOutcomes`                 | stable | Added v1.7.0. Isolates failed outcomes                                      |

---

## Analysis Package (`analysis`) — Stable

All exported symbols in `analysis` are **stable**.

---

## What May Change (Unstable)

These are not yet locked:

- `Config` fields — new fields may be added
- `Summary` fields — new fields may be added
- `FixEdit` internal representation (JSON format may evolve)
- SARIF internal types (unexported) — not part of public API

## Reserved Placeholders

- `FixStrategyAI` and `NeedsAI()` are published with no backend. Pipeline triage treats `FixStrategyAI` the same as `FixStrategySuggest` (no auto-apply).

## Removed APIs (v1.0.0)

All deprecated APIs have been removed. See `docs/MIGRATION_v1.0.md` for migration details.

- `Report.Findings` field → unexported (`findings`). Use `FindingsSnapshot()`, `All()`, `FindByID()`.
- `Report.Merge()` → Use `MergeInto()`.
- `Config.OnStage` → Use `StageHooks`.
- `Metrics.RecordFix()` → Use `RecordFixes(1)`.
- `CountBySeverity()` free function → Use `Report.CountBySeverity()`.
- `SeverityAliases()` → Use `LookupSeverityAlias()`.
- `GetCategory()` → Use `CategoryOf()`.
- `HasFix` / `HasSuggestion` free functions → Use `WithFix` / `WithSuggestion`.

## Deprecation Policy

1. Deprecated symbols get a `// Deprecated:` comment
2. Deprecated symbols remain for at least one major version
3. Removal only in the next major version

## Version Scheme

- `v0.x.y` — pre-release, API may change without notice
- `v1.x.y` — stable, backward-compatible
- `v2.x.y` — breaking changes from v1

## Current Status

**v1.7.0** — `Finding.GroupID` + `Report.GroupFindings()` + SARIF/LSP round-trip; `FixEngine.ApplyWithOutcomes` + `FixApplyResult` (issue #27); `RollbackPolicy` per-file default + `ApplyWithReport` (issue #28); CLI `-fix-rollback-all` flag. Additive except the documented rollback default change (see CHANGELOG `[1.7.0]`).

**Deterministic output guarantee** — All production JSON marshaling uses `encoding/json/v2` with `json.Deterministic(true)` (`marshalOpts`/`prettyMarshalOpts` in `json.go`), enforced by `scripts/json-deterministic-check.sh` in CI: byte-identical output for identical input across runs.

**v1.6.0** — FlightRecorder config-file section (`FlightRecorderFileConfig`, `ResolveFlightRecorder`), CLI `flightRecorder` config section, 4 CI structural-check scripts, `docs-freshness.sh`, per-module CHANGELOGs, go-arch-lint boundary enforcement, LSP serialization benchmarks, flight recorder + path safety edge case tests.

**v1.5.0** — `ParseConfidence`, `Template.Builder`, `ResolveSafePath` exports, severity alias registry (`RegisterSeverityAlias`/`LookupSeverityAlias`), `FixStrategy` normalization, `HasFix` Direct-code requirement, `Range.EndOrStart`/`EndOffsetOrStart`.

**v1.4.1** — Zero new public APIs. Internal refactoring: extracted `must[T]`, `marshalJSONString`, `fixEditJSON`; consolidated test setup (`NewParallelGomega`); resolved 7 pipeline lint issues; zero code duplication.

**v1.4.0** — 2 additive APIs: `FindingError.ErrorCode()`, `FindingError.ErrorFamily()` for go-error-family integration.

**v1.3.0** — 11 additive APIs: Template, FormatTextRich, ApplySimpleFixes, BuildOrDefault, NewReportFromFindings, FilePos, SeverityFromLevel, PriorityString, CheckBinary, RunCmd, FormatTable.
