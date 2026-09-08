# Analysis Module Changelog

All notable changes to the `analysis` module (`github.com/larsartmann/go-finding/analysis`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

No analysis-module changes yet.

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
