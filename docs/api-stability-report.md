# API Stability Report — v1.0.0 Audit

> **HISTORICAL SNAPSHOT (2026-05-18):** Point-in-time API audit at v1.0.0. Project is now at v1.2.0. For current API stability info, see `docs/API_STABILITY.md`.

**Date:** 2026-05-18
**Scope:** All exported symbols in root, pipeline, and analysis packages
**Total symbols:** 317 (root: 206, pipeline: 106, analysis: 5)

---

## Methodology

Every exported symbol was reviewed against FEATURES.md, cross-referenced with test coverage, and assessed for v1.0 API stability.

---

## Stability Classifications

### STABLE — Public API (v1.0 guarantee)

These symbols are production-ready, well-tested, and their signatures will not change in v1.x.

#### Root Package (finding)

**Core Types:**

- `Finding` struct and all methods (`Validate`, `IsValid`, `Clone`, `Equal`, `String`, `Key`, `HasFix`, `HasSuggestion`, `IsAutoFixable`, `IsSuppressed`, `IsSuppressedAt`, `HasCategory`, `NormalizedConfidence`, `Preview`, `ToLSP`)
- `Position` struct and all methods (`IsValid`, `Equal`, `Compare`, `String`, `HasOffset`)
- `Range` struct and all methods (`IsValid`, `HasEnd`, `LineCount`, `Length`, `Equal`, `Compare`, `Contains`, `Overlaps`, `Intersection`, `Adjacent`)
- `Severity` type and all constants/methods (`IsValid`, `Compare`, `String`, `GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`)
- `FixStrategy` type and all constants/methods (`IsValid`, `CanAutoApply`, `NeedsAI`, `String`)
- `Category` type and all constants/methods (`IsStandard`, `IsValid`, `String`)
- `Tag` type and all methods (`IsStandard`, `IsValid`, `String`)
- `Confidence` type and all constants/methods (`IsValid`, `Clamp`, `String`)
- `Suppression` struct and all methods (`IsExpired`, `IsValid`, `IsActive`)
- `SuppressionKind` type and all constants/methods
- `RelatedRef` struct and `IsValid` method
- `ErrorCategory` type and all constants/methods
- `FindingError` struct and all methods (`Error`, `Unwrap`, `Is`, `WithFinding`, `WithPosition`)
- `Correlation` struct
- `DeduplicateBy` type and all constants
- `FilterFunc` type
- `Builder` struct and all methods
- `Summary` struct
- `ToolInfo` struct and `Validate` method
- `Report` struct and all methods
- `LSPDiagnostic`, `LSPRange`, `LSPPosition`, `LSPRelatedInfo`, `LSPLocation` types

**Constructors:**

- `NewFinding`, `NewReport`, `NewBuilder`
- `Pos`, `NewRange`, `NewRangePtr`

**Functions:**

- `GenerateID`, `ParseID`, `IsHashID`
- `Filter`, `FilterInPlace`, `FilterInvalid`
- `BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFixStrategy`, `ByTool`, `ByRule`, `ByFile`, `NotSuppressed`, `HasFix`, `HasSuggestion`
- `GroupBy`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`
- `SortByPosition`, `SortBySeverity`
- `Merge`, `Correlate`, `RangeLinesEq`
- `WithDeduplication`, `WithDeduplicateBy`
- `FromJSON`, `FromLSP`, `FindingsFromJSON`, `FindingsFromSARIF`
- `NewValidationError`, `NewIOError`, `NewParseError`, `NewConflictError`, `NewInternalError`
- `IsFindingError`, `GetCategory`, `IsCategory`

**Sentinel Errors:**

- `ErrValidation`, `ErrIO`, `ErrParse`, `ErrConflict`, `ErrInternal`
- `ErrInvalidFinding`, `ErrInvalidReport`
- `ErrInvalidBuilder`

**Constants:**

- `KeySeparator`, `Version`, `VersionMajor`, `VersionMinor`, `VersionPatch`

**LSP Constants:**

- `LSPSeverityError`, `LSPSeverityWarning`, `LSPSeverityInfo`, `LSPSeverityHint`

#### Pipeline Package

**STABLE:**

- `Pipeline` struct, `New`, `Run`
- `Config` struct, `DefaultConfig`, `Validate`
- `Detector` interface, `DetectorFunc`, `NamedDetectorFunc`
- `FindingProcessor` interface, `ProcessorFunc`, `NamedProcessorFunc`
- `PipelineResult`, `Iteration` structs and methods
- `TriageResult` struct
- `FixEngine` struct, `NewFixEngine`, `NewFixEngineWithProviders`, `Apply`, `ApplyWithConflicts`
- `FixEdit` struct and all methods
- `FixProvider` interface
- `OffsetProvider`, `LineProvider`, `SubstringProvider` structs
- `FixApplier` struct, `NewFixApplier`, `NewFixApplierWithProviders`, `Close`, `Apply`, `ApplyWithDetails`
- `FileBackup` struct, `NewFileBackup`
- `ConflictInfo` struct
- `DetectConflicts`, `FilterConflictingFixes`, `FilterConflictingEdits`, `AnalyzeConflicts`
- `Verify`, `DiffFindings`, `VerifyResult` struct
- `Metrics` struct, `NewMetrics`, `MetricsSnapshot` struct, all methods
- `RetryConfig`, `DefaultRetryConfig`, `Validate`, `NewRetryDetector`
- `PartialResult` struct and methods
- `DetectPartial`, `FormatPartialErrors`, `ErrPartialDetection`
- `IsContextError`, `CheckCanceled`, `CheckCanceledWithMsg`, `WaitWithContext`
- `DefaultMaxIterations` constant

#### Analysis Package

**STABLE:**

- `FromDiagnostic`, `FromTokenPosition`, `NodePosition`, `NodeRange`, `FormatDiagnostic`

---

### EXPERIMENTAL — May change

These symbols work but may evolve:

| Symbol                                    | Package  | Reason                                         |
| ----------------------------------------- | -------- | ---------------------------------------------- |
| `FindingProcessor.Process(ctx, findings)` | pipeline | New ctx+error signature (changed this session) |
| `Config.CorrelateFindings`                | pipeline | Correlation heuristic may change               |
| `FixStrategyAI`                           | finding  | No backend, semantics may evolve               |
| `Correlate()` heuristic                   | finding  | Simple heuristic, may be replaced              |

---

### DEPRECATED — Do not use in new code

None currently. All previous deprecations have been removed.

---

## Breaking Changes from v0.2.1 → Current

1. **`Report.mu` changed from `sync.Mutex` to `sync.RWMutex`** — Binary compatible, but source that references the field type directly will need updating.
2. **`FindingProcessor.Process` signature changed** — Now takes `(context.Context, []Finding)` and returns `([]Finding, error)`. Custom implementations must be updated.
3. **`ComputeSummaryAt(now time.Time)` added** — Non-breaking addition. `ComputeSummary()` still works as before.
4. **`DefaultMaxIterations` exported** — Was unexported `defaultMaxIterations`. Non-breaking.

---

_Assisted-by: Crush_
