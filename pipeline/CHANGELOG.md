# Pipeline Module Changelog

All notable changes to the `pipeline` module (`github.com/larsartmann/go-finding/pipeline`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **`FlightRecorderFileConfig` + `ConfigFile.ResolveFlightRecorder()`** — Flight recorder configuration via JSON config files. Exposes 5 fields (`Enabled`, `OutputDir`, `SlowStageThreshold`, `MinAge`, `MaxBytes`) as string-encoded durations matching the ConfigFile convention. `ResolveFlightRecorder()` constructs a `*FlightRecorderHook` from the config section, returning `(nil, nil)` when disabled.
- **resolveSafePath edge case tests** — Circular symlinks (self-referential and mutual), dangling symlinks pointing outside root, and root-is-symlink path traversal prevention.
- **FlightRecorder edge case tests** — Snapshot write error (permission denied), slow last-stage snapshot during pipeline run, concurrent stage events safety with race detector.

## [1.5.0] - 2026-08-06

### Added

- **`FlightRecorderHook`** — Go runtime execution trace flight recorder wrapping `runtime/trace.FlightRecorder`. Snapshots on demand or automatically when a stage exceeds `SlowStageThreshold`.
- **Deterministic JSON output** — All production marshal calls pass `json.Deterministic(true)`.
- **`Finding.Equal()` tag-order fix** — Tags now compared order-insensitively.

### Changed

- Go module dependencies synchronized across all sub-modules.
