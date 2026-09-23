# Analysis Module Changelog

All notable changes to the `analysis` module (`github.com/larsartmann/go-finding/analysis`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Fixed

- **Multi-edit suggested fixes are no longer lossy (issue #36)** —
  `FromDiagnosticWithSource` carried only the first `TextEdit` of a
  diagnostic's suggested fix into `BeforeCode`/`AfterCode`, silently dropping
  edits 2..N. It now builds the full `Finding.Edits` list (new in core) from
  every text edit, keeping `BeforeCode`/`AfterCode` as the first-edit display
  summary. `ToDiagnostic` reconstructs a suggested fix with the complete edit
  list (previously it emitted a single synthetic edit), so go/analysis
  round-trips are lossless. Alternative fixes (`SuggestedFixes` 1..N) are
  separate fix proposals, not additional edits — the first still wins and the
  limitation is now documented on both functions.

## [1.12.0] - 2026-09-17

Module hygiene: dropped the stray `go-finding/pipeline` require from go.mod (the
analysis module does not depend on the pipeline module). No API changes.

## [1.11.0] - 2026-09-17

Version-alignment release with core v1.11.0. No analysis-module changes.

## [1.10.0] - 2026-09-10

Version-alignment release with core v1.10.0. No analysis-module changes.

## [1.9.2] - 2026-09-08

Version-alignment release with core v1.9.2 (release-pipeline fix; no module code changes).

## [1.9.1] - 2026-09-08

Version-alignment release with core v1.9.1. No analysis-module changes.

## [1.8.0] - 2026-09-08

Version-alignment release with core v1.8.0. No analysis-module API changes;
the tag corrects the module's `go-finding` core reference (the v1.7.0 tag
carried `v1.6.0`) and keeps all four modules on one version line.

## [1.7.0] - 2026-09-08

Version-alignment release with core v1.7.0 — no analysis-module changes since
v1.5.0. Tagged so all four modules share one version line.

No changes yet. The analysis module is stable since v1.5.0.

## [1.5.0] - 2026-08-06

### Changed

- Go module dependencies synchronized across all sub-modules.

## [1.4.1] - 2026-07-22

### Added

- **`FromDiagnostic` BeforeCode extraction** — Reads source files to extract `BeforeCode` from `TextEdit` ranges, enabling full fix data on go/analysis findings.
