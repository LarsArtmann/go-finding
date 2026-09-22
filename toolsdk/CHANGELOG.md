# Changelog

All notable changes to the `toolsdk` sub-module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> The root [CHANGELOG.md](../CHANGELOG.md) records cross-module and core changes.
> Sub-modules have no `version.go`; the directory-prefixed git tag
> (`toolsdk/v*`) is the single source of truth for this module's version.

## [Unreleased]

No toolsdk-module changes yet.

## [1.13.0] - 2026-09-22

`Trigger` gains `NotRequires []string`: disqualifying file patterns that
express ownership deference (a tool whose NotRequires pattern matches does
not run, e.g. a standalone formatter deferring to a repo's own treefmt
config). Field-for-field parity with BuildFlow's `domain/tool.Trigger`.

## [1.12.0] - 2026-09-17

Module hygiene: re-attached the package doc comment to its `package` statement
(revive `package-comments`), dropped the stray `go-finding/pipeline` require
from go.mod, and restored the `go` directive to `1.26.7` to match go.work.
No API changes.

## [1.11.0] - 2026-09-17

Version-alignment release with core v1.11.0: the go-error-family dependency
became an explicit indirect require, and CI wiring (arch graph, lint matrix)
now covers the module. No API changes.

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
