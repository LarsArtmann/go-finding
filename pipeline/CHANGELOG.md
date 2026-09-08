# Pipeline Module Changelog

All notable changes to the `pipeline` module (`github.com/larsartmann/go-finding/pipeline`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

No pipeline-module changes yet.

## [1.9.1] - 2026-09-08

Version-alignment release with core v1.9.1. No pipeline-module code changes.

## [1.9.0] - 2026-09-08

### Added

- **`FlightRecorderConfig.MaxFiles` + trace rotation** — After each successful snapshot, the oldest `go-finding-trace-*` files beyond the cap are deleted (best-effort; deletion errors are logged, never surfaced). Default 0 = unlimited. Pruning matches only this hook's snapshot files and sorts by modification time.
- **`FlightRecorderConfig.Compress`** — Wraps snapshot output in gzip (`.trace.gz` suffix). Default false preserves plain `.trace` files that open directly with `go tool trace`; compressed files gunzip losslessly — verified end-to-end with `go tool trace` and pinned by a header-equality test. Config-file fields `maxFiles` and `compress` added to `FlightRecorderFileConfig`.

### Fixed

- **Snapshot rotation is serialized** — `pruneSnapshots` now holds the same mutex as `WriteTo`, so two concurrent snapshots can never list and delete overlapping file sets (previously possible: double-remove ENOENT warnings and nondeterministic pruning on same-second modtime ties). Pinned by a concurrent-rotation test under `-race` and a hostile-directory fuzz target.

## [1.8.0] - 2026-09-08

### Added

- **`FixApplier.ApplyDryRun` (D4)** — Plan/apply UX: resolves every fix against current file contents and returns the `ApplyReport` a real run would produce (would-apply counts, per-finding outcomes, shift maps) without writing, backing up, or rolling back anything. `RolledBack` is always empty; unsafe paths and provider failures surface exactly as in a real run.
- **`Config.OnFixOutcome` (D3)** — Outcome-status callback fired once per fixable finding with the exact `FixOutcomeStatus` and the typed error for failed resolutions. Additive successor to the boolean `Config.OnFix` (now deprecated, still functional; setting both fires both).

### Changed

- **Unsafe-path findings now surface as failed outcomes instead of being silently dropped** — Findings whose `Position.File` fails the path-containment check (traversal outside the root, e.g. `../../etc/passwd`) previously vanished during grouping with no signal. They are now reported as `FixOutcomeFailed` entries (validation-category `*finding.FindingError` with the finding's position) and included in the joined error return of `ApplyWithReport` (and the `Apply`/`ApplyWithDetails`/`ApplyWithShiftMap` delegators). Safe findings in the same run still apply; nothing about the security boundary itself changed.

## [1.7.0] - 2026-09-08

### Added

- **`FixEngine.ApplyWithOutcomes` + `FixApplyResult`** — Per-finding fix outcomes (issue #27). Each input finding gets one `FixOutcome` in input order distinguishing `applied`, `no-change`, `refused` (matched provider produced zero edits without error), `conflict`, `invalid` (edit dropped as invalid/out of bounds), and `failed` (provider error, carried in `Err`). Helpers: `FixApplyResult.OutcomeFor(id)`, `OutcomeCounts()`, `HasErrors()`. `Apply` and `ApplyWithConflicts` delegate to it, so their legacy return shapes are unchanged.
- **`FixApplier.ApplyWithReport` + `ApplyReport`** — Disk-level fix run reporting: applied findings, per-finding `Outcomes`, shift maps, and `RolledBack` file list. `FailedOutcomes()` isolates provider failures. Soft per-finding failures are returned as a joined error alongside the applied fixes instead of aborting the run.
- **`RollbackPolicy`** — Configurable failure semantics for multi-file fix runs (issue #28). `RollbackPolicyFailingFile` (new default) restores only the failing file; earlier files keep their fixes. `RollbackPolicyAllFiles` preserves the legacy all-or-nothing rollback. Set via `FixApplier.SetRollbackPolicy`, `Config.FixRollbackAllFiles`, or the config-file field `fixRollbackAllFiles`.

### Changed

- **`applyToFile` no longer fails on provider resolve errors when edits applied** — A file with both applied edits and unresolvable findings is written and counted as applied; the resolve errors surface in `ApplyReport.Outcomes` and the joined error return instead of triggering a full rollback of all previously fixed files.

### Fixed

- **`ApplyWithReport` now reports rolled-back files on backup failure and cancellation** — When a file's backup could not be created, or the context was cancelled mid-run, the run aborted but `ApplyReport.RolledBack` silently omitted the files that had actually been restored, and a failing rollback's error was swallowed. Both paths now record every modified file in `RolledBack` (resolved absolute paths) and wrap rollback failures so `errors.Is`/`errors.As` still reach the original cause. Covered by cancel and backup-failure tests for both rollback policies plus an end-to-end pipeline rollback wiring test.

### Added

- **`Metrics.RecordOutcome(status)` + `OutcomeCounts()`** — Aggregate per-finding fix outcomes into printable counts. `MetricsSnapshot.OutcomeCounts` carries a point-in-time copy; the pipeline's fix stage records one outcome per finding automatically, and the CLI prints a `Fix outcomes:` summary line.
- **Deterministic JSON for `FixOutcome` and `FixApplyResult`** — Custom `MarshalJSON`/`UnmarshalJSON` using `json.Deterministic(true)` (per the repo determinism rule). Errors serialize as message strings and unmarshal as plain errors; finding/status/content round-trip losslessly.
- **`pipeline/examples/outcomes`** — Runnable demo of per-finding outcomes (`ApplyWithOutcomes`: applied / refused / failed) and the `RollbackPolicy` knob, with a compile check in `pipeline/examples/example_compile_test.go`.

### Changed

- **Failed outcomes carry typed `*finding.FindingError`** — Provider resolution failures are now wrapped as parse-category `FindingError` with the finding's position attached (and `errorfamily` classification), instead of a bare `fmt.Errorf` wrapper. `errors.Is`/`As` chains to the original provider cause are preserved.
- **Rollback errors now list the restored paths** — `ApplyWithReport` error text embeds `(rolled back: <path1>, <path2>)` on cancellation, backup failure, and hard file failure, alongside the existing `rollback also failed` wrapper. Note `RolledBack` lists every restored file, including files that were backed up but never modified (those restores are content no-ops).

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
