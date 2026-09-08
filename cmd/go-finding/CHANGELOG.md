# CLI Module Changelog

All notable changes to the `cmd/go-finding` module (`github.com/larsartmann/go-finding/cmd/go-finding`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.8.0] - 2026-09-08

### Added

- **staticcheck detector fix extension** — staticcheck-format JSON lines may carry `before`/`after` fields; when both are present the parsed finding becomes auto-fixable (`FixStrategyDirect` with literal replacement) and flows through the fix pipeline into `Fix outcomes:`. Real staticcheck output (without the fields) is unaffected. Enables auto-fix-capable external tools that emulate the staticcheck JSON format.
- **Presence e2e for `Fix outcomes:`** — `testdata/fakestcheck` fixture binary drives the full production path (detector → triage → fix → summary) in `TestRun_E2E_FixOutcomesLine_PresentWithFixableFindings`.

## [1.7.0] - 2026-09-08

### Added

- **`-fix-rollback-all` flag** — Opt into all-or-nothing rollback on hard file failures (pipeline default is per-file rollback since the rollback-policy change). OR-combined with the config-file `fixRollbackAllFiles` field.
- **`Fix outcomes:` stderr summary** — After a run with fixes, the CLI prints nonzero per-finding outcome counts (e.g. `applied=3, refused=1, failed=2`) in canonical order, sourced from the pipeline metrics.

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
