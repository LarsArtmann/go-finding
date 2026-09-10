# Changelog

All notable changes to the `toolsdk` sub-module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> The root [CHANGELOG.md](../CHANGELOG.md) records cross-module and core changes.
> Sub-modules have no `version.go`; the directory-prefixed git tag
> (`toolsdk/v*`) is the single source of truth for this module's version.

## [1.10.0] - 2026-09-10

First release. The BuildFlow provider plugin contract, migrated from
`github.com/larsartmann/buildflow/tool-sdk` so external tools can target a
stable, BuildFlow-independent SDK.

### Added

- `Spec` — declarative tool description (name, description, `Trigger`, `DependsOn`,
  `Inputs`, `Detect`/`Repair`/`HealthCheck`) expressed against canonical
  `finding` types; registered via `Register` (panics on malformed specs).
- `Trigger` constructors: `OnGoFiles`, `OnGoModule`, `OnFiles`, `AnyLanguage`.
- `Repairer`/`RepairerFunc`/`RepairResult` — repair contract; fix counts are
  measured by re-detection, never self-reported.
- Registry access: `All()`, plus test-isolation helpers
  (`SnapshotForTest`/`RestoreForTest`/`ResetForTest`).
- Dry-run context plumbing: `WithDryRun`/`DryRunFromContext`.
- The module depends only on the core `github.com/larsartmann/go-finding` module.

[1.10.0]: https://github.com/larsartmann/go-finding/releases/tag/toolsdk%2Fv1.10.0
