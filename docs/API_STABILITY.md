# API Stability — go-finding

**Last audited:** 2026-06-09
**Version:** v0.6.1

go-finding follows the [Go 1 Compatibility Promise](https://go.dev/doc/go1compat) philosophy.

## Commitment

Once we tag v1.0.0:

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

| Type | Status | Notes |
|------|--------|-------|
| `Finding` | stable | Core data type. Fields may be added but never removed or renamed. |
| `Report` | stable | Thread-safe container. `Findings` slice is public; use accessors for concurrent safety. |
| `Position` | stable | File/Line/Column/Offset. Zero-value is valid ("unpositioned"). |
| `Range` | stable | Start/End Position. `IsValid()` checks End >= Start. |
| `Builder` | stable | Fluent Finding builder. |

### Named Types

| Type | Status | Notes |
|------|--------|-------|
| `Severity` | stable | Info, Warning, Error, Critical |
| `Confidence` | stable | float64 with `IsValid()`/`Clamp()`/`Compare()` |
| `Category` | stable | Standard categories (Correctness, Security, etc.) |
| `Tag` | stable | Classification tags |
| `FixStrategy` | stable | Direct, Suggest, AI (reserved) |
| `ErrorCategory` | stable | Validation, Parse, IO, Conflict, Internal |
| `SuppressionKind` | stable | FalsePositive, WonFix, etc. |
| `CorrelationScore` | stable | float64 [0.0, 1.0] |
| `DeduplicateBy` | stable | Enum: ID, Position, Rule |
| `RelationKind` | stable | Related, Causes, IsCausedBy |
| `LSPSeverity` | stable | Named int: Error/Warning/Info/Hint |
| `LSPDiagnosticTag` | stable | Named int: Unnecessary(1), Deprecated(2) |

### Supporting Types

| Type | Status | Notes |
|------|--------|-------|
| `ToolInfo` | stable | Name + Version |
| `Summary` | stable | Aggregated statistics. New fields may be added. |
| `Suppression` | stable | Kind/Reason/ExpiresAt |
| `Correlation` | stable | Score + finding pair |
| `ModifiedPair` | stable | Before/After from Diff |
| `ParsedID` | stable | ID parse result |
| `RelatedRef` | stable | Related finding reference |
| `DiffResult` | stable | Diff comparison result |

### Interfaces and Functions

| Symbol | Status | Notes |
|--------|--------|-------|
| `Detector` | stable | Interface: `Name()` + `Detect(ctx)` |
| `DetectorFunc` | stable | `func(ctx) ([]Finding, error)` implementing Detector |
| `NamedDetectorFunc` | stable | Wrap DetectorFunc with name |
| `FilterFunc` | stable | `func(Finding) bool` |
| `ToolAdapter[O]` | stable | Generic tool→Finding converter |
| `MergeOption` | stable | Functional option |
| `MergeOptions` | stable | Deduplicate + DeduplicateBy |

### LSP Wire Types

| Type | Status | Notes |
|------|--------|-------|
| `LSPDiagnostic` | stable | Full LSP diagnostic |
| `LSPPosition` | stable | |
| `LSPRange` | stable | |
| `LSPLocation` | stable | |
| `LSPRelatedInfo` | stable | |

### Error Types

| Symbol | Status | Notes |
|--------|--------|-------|
| `FindingError` | stable | Structured error with Category/Cause |
| `NewValidationError` | stable | Constructor |
| `NewParseError` | stable | Constructor |
| `NewIOError` | stable | Constructor |
| `NewConflictError` | stable | Constructor |
| `NewInternalError` | stable | Constructor |
| `GetCategory` | stable | Extract ErrorCategory from error |
| `IsCategory` | stable | Check error category |
| `IsFindingError` | stable | Type assertion |

### Key Functions

| Function | Status | Notes |
|----------|--------|-------|
| `Combine` | stable | Merge reports with dedup |
| `Correlate` | stable | Find related findings |
| `Diff` | stable | Compare finding sets by ID |
| `Filter` / `FilterInPlace` | stable | Filter by predicates |
| `GenerateID` | stable | Deterministic ID |
| `ParseID` | stable | Parse generated ID |
| `FindingsFromSARIF` | stable | SARIF import |
| `FindingsFromReader` | stable | Streaming SARIF import |
| `FindingsFromJSON` | stable | JSON import |
| `ReportFromJSON` | stable | JSON → Report |
| `FromJSON` | stable | JSON → Finding |
| `FromLSP` | stable | LSP → Finding |
| `FormatText` / `FormatMarkdown` | stable | Human-readable output |
| `ParseSeverity` / `MustParseSeverity` | stable | String → Severity |
| `ParseCategory` / `MustParseCategory` | stable | String → Category |
| `CategoryForLinter` | stable | Linter→category lookup |
| `RegisterLinterCategory` | stable | Register linter mapping |
| `NewToolAdapter` | stable | Generic adapter constructor |
| `FromSARIFLevel` | stable | SARIF level → Severity |

### Report Methods

| Method | Status | Notes |
|--------|--------|-------|
| `Report.Merge` | **deprecated** | Use `MergeInto`. Removed in v1.0.0. |
| `Report.MergeInto` | stable | Returns new Report, no mutation |
| `Report.AddFinding` / `AddFindings` | stable | Thread-safe |
| `Report.All` | stable | `iter.Seq[Finding]` |
| `Report.ActiveFindings` | stable | Non-suppressed |
| `Report.ComputeSummary` / `At` | stable | |
| `Report.FindByID` / `FindByRule` | stable | |
| `Report.FindingsSnapshot` | stable | Deep-cloned |
| `Report.Validate` | stable | |
| `Report.ToSARIF` / `WriteSARIF` | stable | |
| `Report.PrettyJSON` / `Filtered` | stable | |
| `Report.Filter` / `Map` | stable | Transform |
| `Report.Len` / `CountBySeverity` | stable | Counts |

### Filter Constructors

| Constructor | Status |
|------------|--------|
| `ByCategory`, `ByConfidence`, `ByConfidenceAtLeast`, `ByFile`, `ByFixStrategy`, `ByRule`, `BySeverity`, `BySeverityAtLeast`, `ByTool` | stable |
| `AnyOf`, `Negate`, `FilterInvalid`, `NotSuppressed`, `HasFix`, `HasSuggestion` | stable |

### Constants

| Constant | Status | Notes |
|----------|--------|-------|
| `Severity*` (Info/Warning/Error/Critical) | stable | |
| `Category*` (Correctness/Security/etc.) | stable | |
| `FixStrategyDirect/Suggest` | stable | |
| `FixStrategyAI` | **reserved** | No backend. Won't auto-apply. |
| `RelationKind*`, `SuppressionKind*`, `Tag*` | stable | |
| `VersionMajor/Minor/Patch`, `Version` | stable | |
| `LSPSeverity*`, `LSPDiagnosticTag*` | stable | |
| `LSPSeverityKey`, `LSPDiagnosticTagsKey` | stable | Metadata keys |

### Variables

| Variable | Status | Notes |
|----------|--------|-------|
| `SeverityAliases` | stable | Map of alias→Severity |

---

## Pipeline Package (`pipeline`) — Stable

### Types

| Type | Status |
|------|--------|
| `Pipeline`, `Config`, `CompletionReason`, `Stage` | stable |
| `Iteration`, `PipelineResult`, `PartialResult` | stable |
| `FixApplier`, `FixEngine`, `FixEdit`, `FixProvider`, `FixGroup` | stable |
| `ConflictInfo`, `VerifyResult`, `TriageResult` | stable |
| `Metrics`, `MetricsSnapshot` | stable |
| `FileBackup` | stable |
| `RetryDetector`, `RetryConfig` | stable |
| `GeneratedFileFilter` | stable |
| `FindingProcessor`, `ProcessorFunc` | stable |
| `Detector`, `DetectorFunc` (type aliases) | stable |
| `OffsetProvider`, `LineProvider`, `SubstringProvider` | stable |

### Functions

All exported functions in `pipeline` are **stable**.

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

## Known Pre-v1.0 Concerns

### Reserved Placeholders

- `FixStrategyAI` and `NeedsAI()` are published with no backend. Pipeline triage treats `FixStrategyAI` the same as `FixStrategySuggest` (no auto-apply).

### Deprecations

- `Report.Merge()` — deprecated in favor of `MergeInto()`. Will be removed in v1.0.0.
- `RecordFix()` — superseded by `RecordFixes(uint)`. Will be removed in v1.0.0.

### Encapsulation Risks

- `Report.Findings` is a public slice. Use `AddFinding`/`AddFindings`/`FindByID`/`All()` for thread-safe access. ADR 10 documents the migration path via `FindingsSnapshot()`.

### Owner Decisions (pre-v1.0)

1. `Position` zero-value safety — Is `Position{}` valid or an error?
2. `Range.End` zero-value ambiguity — Required vs optional?
3. `PositionOffset` sentinel design — Need `OffsetUndefined`?
4. `Report.Merge()` semantics — Should be removed after deprecation period.

## Deprecation Policy

1. Deprecated symbols get a `// Deprecated:` comment
2. Deprecated symbols remain for at least one major version
3. Removal only in the next major version

## Version Scheme

- `v0.x.y` — pre-release, API may change without notice
- `v1.x.y` — stable, backward-compatible
- `v2.x.y` — breaking changes from v1

## Current Status

**Pre-release** (v0.6.1). API is stabilizing but not yet locked. Target: v1.0.0 lock after resolving owner-decision items.
