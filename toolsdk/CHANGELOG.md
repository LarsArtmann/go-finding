# Changelog

All notable changes to the `toolsdk` sub-module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> The root [CHANGELOG.md](../CHANGELOG.md) records cross-module and core changes.
> Sub-modules have no `version.go`; the directory-prefixed git tag
> (`toolsdk/v*`) is the single source of truth for this module's version.

## [Unreleased]

### Added

- **`Trigger` validation at registration (issue #44)** — `Register` now
  panics on a malformed `Trigger`: `Language` must be `""` (language-agnostic)
  or one of the exported `ValidLanguages()` (`go`, `js`, `ts`, `python`,
  `nix`, `rust`), and `Files`/`Requires`/`NotRequires` must not contain
  empty patterns. A typo'd `Language` (e.g. `"golang"`) used to register
  cleanly and then silently never match downstream — exactly the startup
  failure mode the `Register` panic convention exists to prevent. The panic
  message names the field and the valid set; direct checks match the new
  `ErrInvalidTrigger` sentinel via `errors.Is`.
- **Typed option accessors (issue #43)** — `StringFromContext`,
  `IntFromContext`, and `BoolFromContext` read one declared option from the
  context and apply the tool's fallback on any miss (name unset, empty
  value set, nil context, nil value). A present value whose Go type does
  not match the accessor yields an error matching `ErrOptionKindMismatch`.
  Tools stop hand-rolling the lookup/assert/fallback plumbing; range and
  semantic validation stays tool-side, and declaration enforcement remains
  `Spec.ValidateOptions`' job.
- **`RepairResult` structured provenance (issue #42)** — additive,
  non-gating fields: `ModifiedFiles []string` (files the repair wrote, or
  would write under dry-run) and `Facts []Fact` (tool-defined key/value
  provenance, e.g. `baseline-failed` = `true`). Facts describe what the run
  did, never its effect: fix counts remain measured by re-detection
  (structural anti-lie design) and nothing downstream gates on a fact.
  Hosts surface provenance without string-matching `Description`. Existing
  specs compile unchanged.

## [1.15.0] - 2026-10-05

### Added

- **`Spec.SwitchCases` with `ExcludedRuleIDs`** — SDK-registered tools can now
  express conditional repair paths keyed on the diagnose output's finding rule
  IDs (R6 / gotcha-#194 parity): `Suffix` names the DAG node, `FindingRuleIDs`
  select which findings route to the case, `ExcludedRuleIDs` encode precedence
  so at most one mutating repair case runs per diagnosis (go-workflow runs
  every predicate-matching case in parallel; mutual exclusion is the spec
  author's responsibility). Consumers map the field verbatim — BuildFlow's
  `ToolFromSpec` mirrors it into `DAGTopology.SwitchCases`, with field parity
  pinned by test.

## [1.14.0] - 2026-10-01

### Added

- `Spec.Options []Option` — declares the tunable knobs a tool's Detect/Repair
  reads from the context — plus the per-run value channel `WithOptions` /
  `OptionsFromContext` and `Spec.ValidateOptions`. Kind-only validation at
  the SDK layer (int/string/bool); range and semantic validation stay with
  the tool implementation. Unknown option names fail loudly so a consumer
  config typo cannot silently run on defaults. Declarations are validated at
  registration (`Register` panics on malformed or duplicate options).
  Validation errors match the exported sentinels `ErrInvalidOption`,
  `ErrDuplicateOption`, `ErrUnknownOption`, and `ErrOptionKindMismatch` via
  `errors.Is`. `WithOptions` snapshots the values map (later caller mutation
  does not affect the run) and nil/empty values clear any options inherited
  from a parent context, so a run carries exactly the values it set.
  First consumer: art-dupl's `threshold` knob (its fixed 5-statement floor
  becomes settable per run without rebuilding art-dupl).

### Documented

- **golangci-lint `exhaustruct_v5` panic on Go 1.27 promoted-field keys** is an
  upstream pinning issue, not a toolsdk defect: golangci-lint (v2.13.2 and
  master) vendors `dev.gaijin.team/go/exhaustruct/v5` v5.0.3, whose analyzer
  predates `promotedKeysVersion = "go1.27"`; the fix shipped in v5.2.0
  (`analyzer/missing-fields-visitor.go`). Consumers gating on
  `exhaustruct_v5` (fleet repos triaging toolsdk findings with that linter)
  should ask golangci-lint to bump the dep rather than disable the linter —
  tracked upstream in golangci/golangci-lint#6780.

## [1.13.1] - 2026-09-23

`Spec` gains `ModuleFanOut bool`: declares that the tool runs once per Go
module in a multi-module workspace rather than once at the repo root.
Field-for-field parity with BuildFlow's `domain/tool.DAGTopology` — a spec
that omits the field silently loses per-module fan-out at conversion time.
First consumer: branching-flow (BuildFlow issue #17 migration).

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
