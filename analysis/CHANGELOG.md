# Analysis Module Changelog

All notable changes to the `analysis` module (`github.com/larsartmann/go-finding/analysis`)
will be documented in this file.

For root-level changes, see the [root CHANGELOG.md](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

No changes yet. The analysis module is stable since v1.5.0.

## [1.5.0] - 2026-08-06

### Changed

- Go module dependencies synchronized across all sub-modules.

## [1.4.1] - 2026-07-22

### Added

- **`FromDiagnostic` BeforeCode extraction** — Reads source files to extract `BeforeCode` from `TextEdit` ranges, enabling full fix data on go/analysis findings.
