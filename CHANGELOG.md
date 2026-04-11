# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial implementation of core types:
  - `Finding` - unified static analysis finding representation
  - `Severity` - info/warning/error/critical levels with ordering
  - `FixStrategy` - none/suggest/direct/ai classification
  - `Position`/`Range` - source code location tracking
  - `Suppression` - unified suppression handling
  - `Report` - container for tool findings with summary statistics
- Filtering and grouping utilities for findings
- Report merging with deduplication support
- SARIF 2.1.0 output generation
- LSP Diagnostic conversion
- go/analysis.Diagnostic integration
- JSON serialization helpers
- Comprehensive test suite (25 tests)

### Changed

- N/A

### Deprecated

- N/A

### Removed

- N/A

### Fixed

- N/A

### Security

- N/A

## [0.1.0] - 2026-04-11

### Added

- Initial release with core types and basic functionality
- Support for unified finding representation across tools
