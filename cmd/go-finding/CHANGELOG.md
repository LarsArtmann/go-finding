# CLI Module Changelog

All notable changes to the `cmd/go-finding` module (`github.com/larsartmann/go-finding/cmd/go-finding`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

No changes yet.

## [1.6.0] - 2026-08-08

### Added

- **`flightRecorder` config section** — YAML/JSON config files now support a `flightRecorder` section as an alternative to `-trace`, `-trace-dir`, and `-trace-slow` flags.
- **`minAge` and `maxBytes` config fields** — CLI `flightRecorder` config section now supports all 5 pipeline fields (`enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes`), achieving full parity with `pipeline.FlightRecorderFileConfig`.
- **FlightRecorder validation tests** — Error-path test for invalid `slowStageThreshold` and `minAge` durations; happy-path test for all-fields config.
- **E2E flight recorder tests** — `TestRun_E2E_TraceFlag` and `TestRun_E2E_TraceViaConfigFile` verify `.trace` file output.

### Changed

- **`flightRecorderFileConfig` struct** — Added `MinAge` (string) and `MaxBytes` (uint64) fields for full parity with pipeline's `FlightRecorderFileConfig`. Validation extended to parse-check `minAge` as a duration string.

## [1.5.0] - 2026-08-06

### Added

- **`-trace`, `-trace-dir`, `-trace-slow` CLI flags** — Enable Go execution trace flight recorder from the command line.
