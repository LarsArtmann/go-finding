# Pipeline Module Changelog

All notable changes to the `pipeline` module (`github.com/larsartmann/go-finding/pipeline`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **`FixEngine.ApplyWithOutcomes` + `FixApplyResult`** — Per-finding fix outcomes (issue #27). Each input finding gets one `FixOutcome` in input order distinguishing `applied`, `no-change`, `refused` (matched provider produced zero edits without error), `conflict`, `invalid` (edit dropped as invalid/out of bounds), and `failed` (provider error, carried in `Err`). Helpers: `FixApplyResult.OutcomeFor(id)`, `OutcomeCounts()`, `HasErrors()`. `Apply` and `ApplyWithConflicts` delegate to it, so their legacy return shapes are unchanged.
- **`FixApplier.ApplyWithReport` + `ApplyReport`** — Disk-level fix run reporting: applied findings, per-finding `Outcomes`, shift maps, and `RolledBack` file list. `FailedOutcomes()` isolates provider failures. Soft per-finding failures are returned as a joined error alongside the applied fixes instead of aborting the run.
- **`RollbackPolicy`** — Configurable failure semantics for multi-file fix runs (issue #28). `RollbackPolicyFailingFile` (new default) restores only the failing file; earlier files keep their fixes. `RollbackPolicyAllFiles` preserves the legacy all-or-nothing rollback. Set via `FixApplier.SetRollbackPolicy`, `Config.FixRollbackAllFiles`, or the config-file field `fixRollbackAllFiles`.

### Changed

- **`applyToFile` no longer fails on provider resolve errors when edits applied** — A file with both applied edits and unresolvable findings is written and counted as applied; the resolve errors surface in `ApplyReport.Outcomes` and the joined error return instead of triggering a full rollback of all previously fixed files.

## [1.6.0] - 2026-08-08

### Added

- **`FlightRecorderFileConfig` + `ConfigFile.ResolveFlightRecorder()`** — Flight recorder configuration via JSON config files. Exposes 5 fields (`Enabled`, `OutputDir`, `SlowStageThreshold`, `MinAge`, `MaxBytes`) as string-encoded durations matching the ConfigFile convention. `ResolveFlightRecorder()` constructs a `*FlightRecorderHook` from the config section, returning `(nil, nil)` when disabled.
- **resolveSafePath edge case tests** — Circular symlinks (self-referential and mutual), dangling symlinks pointing outside root, and root-is-symlink path traversal prevention.
- **FlightRecorder edge case tests** — Snapshot write error (permission denied), slow last-stage snapshot during pipeline run, concurrent stage events safety with race detector.
- **`ResolveSafePath`, `ResolveSafePathFrom`, `ResolveRoot`** — Exported the path-traversal security boundary as public API. `ResolveSafePath(rootDir, relPath)` is the convenience wrapper; `ResolveRoot(rootDir)` + `ResolveSafePathFrom(resolvedRoot, relPath)` enable batch optimization (resolve symlinks once, cache per-path results).
- **`FlightRecorderHook.Degraded()`** — When a second flight recorder is created while one is already active (Go singleton limit), `NewFlightRecorderHook` returns a degraded hook instead of an error. `Degraded()` returns true; all snapshot operations silently no-op.

### Changed

- **`FlightRecorderHook.Snapshot` signature** — Now accepts `context.Context` as first parameter: `Snapshot(ctx, reason)`. Cancelled contexts skip the trace write and return `context.Cause(ctx)` wrapped in a descriptive error. **Breaking change** for callers using the v1.5.0 `Snapshot(reason)` signature.
- **`NewFlightRecorderHook` singleton behavior** — On "flight recorder already enabled" runtime error, now returns `(degradedHook, nil)` instead of `(nil, err)`.

### Fixed

- **`FixEdit.MarshalJSON` non-deterministic output** — Missing `json.Deterministic(true)` caused non-reproducible fix serialization. Now deterministic.

## [1.5.0] - 2026-08-06

### Added

- **`FlightRecorderHook`** — Go runtime execution trace flight recorder wrapping `runtime/trace.FlightRecorder`. Snapshots on demand or automatically when a stage exceeds `SlowStageThreshold`.
- **Deterministic JSON output** — All production marshal calls pass `json.Deterministic(true)`.
- **`Finding.Equal()` tag-order fix** — Tags now compared order-insensitively.

### Changed

- Go module dependencies synchronized across all sub-modules.
