# API Stability Guarantee

go-finding follows the [Go 1 Compatibility Promise](https://go.dev/doc/go1compat) philosophy.

## Commitment

Once we tag v1.0.0:

1. **No breaking changes** in any minor or patch release (v1.x.y)
2. Breaking changes require a **major version bump** (v2.0.0)
3. All exported symbols are part of the public API

## What Is Stable

These types and functions are committed to their current signatures:

- `Finding` struct — fields may be added but never removed or renamed
- `Report` struct and all methods
- `Position`, `Range` types and methods
- `Severity`, `Confidence`, `Category`, `Tag`, `FixStrategy`, `SuppressionKind` named types
- `Diff()`, `Combine()`, `Correlate()`, `DeduplicateByID/Position/Rule`
- `GenerateID()`, `ParseID()`
- SARIF functions: `ToSARIF()`, `WriteSARIF()`, `FindingsFromSARIF()`, `FindingsFromReader()`
- LSP conversion: `ToLSPDiagnostics()`
- Pipeline: `Config`, `Pipeline.Run()`, `Detector`, `FindingProcessor`, `FixProvider`
- Builder API: `NewBuilder()`, `Build()`, all `With*` methods
- Error types: `FindingError`, sentinel errors (`ErrValidation`, etc.)
- Filter combinators: `Filter`, `FilterInPlace`, `AnyOf`, `Negate`

## What May Change

These are not yet locked:

- `Pipeline` internals (unexported fields)
- `Config` fields — new fields may be added
- `FixEdit` internal representation
- `SarifLog` and related SARIF types (internal representation)
- `Summary` fields — new fields may be added

## Deprecation Policy

1. Deprecated symbols get a `// Deprecated:` comment
2. Deprecated symbols remain for at least one major version
3. Removal only in the next major version

## Version Scheme

- `v0.x.y` — pre-release, API may change without notice
- `v1.x.y` — stable, backward-compatible
- `v2.x.y` — breaking changes from v1

## Current Status

**Pre-release** (v0.3.x). API is stabilizing but not yet locked.
