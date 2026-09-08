# Ecosystem Architecture

go-finding is the hub of a constellation of libraries and tools that share a common data model for static analysis findings. This document maps the ecosystem, explains each component's role, and shows how they fit together.

## At a Glance

```
Tools that PRODUCE findings
┌──────────┐  ┌──────────┐  ┌──────────────────────┐
│ Linters  │  │ Auto-    │  │ BuildFlow checkers    │
│          │  │ config-  │  │ (gomod, nix, flake,   │
│          │  │ urers    │  │  todo, docfile...)    │
└────┬─────┘  └────┬─────┘  └──────────┬────────────┘
     │             │                   │
┌────▼────┐   ┌────▼──────────┐   ┌────▼──────────────┐
│ go-     │   │ linter-       │   │ go-checker-       │
│ linter- │   │ autoconfigure │   │ helpers           │
│ sdk     │   │ -sdk          │   │ (finding builders,│
│ (Rule+  │   │ (config I/O + │   │  fix pipeline,    │
│  Regist)│   │  ConfigIssue) │   │  safe I/O, tests) │
└────┬────┘   └────┬──────────┘   └────┬──────────────┘
     │             │                   │
     └─────────────┼───────────────────┘
                   │
           ┌───────▼────────┐
           │  go-finding    │
           │  (THE HUB)     │
           └───┬────────┬───┘
               │        │
      ┌────────▼──┐  ┌──▼──────────┐
      │BuildFlow/ │  │ pipeline/   │
      │finding    │  │ analysis/   │
      │(legacy    │  │ cmd/go-     │
      │  bridge)  │  │ finding/    │
      └───────────┘  └─────────────┘
```

## Components

### go-finding (this project)

The hub. Defines the unified `Finding` type, the detect-triage-fix-verify pipeline, SARIF/LSP/JSON output, and the `Detector`/`Report`/`Severity`/`Category` abstractions that every other component builds on.

Four Go modules with Unix-style decomposition:

| Module   | Path                                               | Role                                              |
| -------- | -------------------------------------------------- | ------------------------------------------------- |
| Core     | `github.com/larsartmann/go-finding`                | Types, filtering, merging, SARIF, LSP, formatting |
| Pipeline | `github.com/larsartmann/go-finding/pipeline`       | Detect-triage-fix-verify loop, fix engine         |
| Analysis | `github.com/larsartmann/go-finding/analysis`       | go/analysis diagnostic conversion                 |
| CLI      | `github.com/larsartmann/go-finding/cmd/go-finding` | Command-line tool with config, output formats     |

### go-linter-sdk

Shared scaffolding for **linters** (tools that find code issues). Eliminates the Violation-to-Finding converter layer by codifying the `go-structure-linter` pattern: rules emit `finding.Finding` directly.

- **Repo:** [github.com/larsartmann/go-linter-sdk](https://github.com/larsartmann/go-linter-sdk)
- **Depends on:** go-finding
- **Core types:** `Rule` interface, `RuleFunc` adapter, `RuleMeta`, `Registry`
- **Execution paths:**
  - `DetectorFromRegistry` — single opaque detector (simple CLI / BuildFlow DAG)
  - `DetectorsFromRegistry` — one detector per rule (pipeline parallelism, timeouts, error isolation)
- **Consumers (planned):** branching-flow, erraudit, go-structure-linter

### linter-autoconfigure-sdk

Shared foundation for **auto-configurers** (tools that fix linter config files like `.golangci.yml` or `.oxlintrc.json`). Owns config round-trip, finding emission for config issues, and a provider spec for BuildFlow integration.

- **Repo:** [github.com/larsartmann/linter-autoconfigure-sdk](https://github.com/larsartmann/linter-autoconfigure-sdk)
- **Depends on:** go-finding, go-atomic-write
- **Core types:** `ConfigIssue`, `ProviderSpec`, `ConfigError`
- **What it owns:**
  - Config I/O: `ReadConfig`, `LoadJSON[T]`, `SaveJSON` (idempotent, crash-durable atomic write)
  - Finding emission: `FindingFromIssue` / `FindingsFromIssues` (auto-attaches `FixStrategySuggest`)
  - Provider shape: `ProviderSpec{Analyze, Repair}` for BuildFlow integration
- **Deliberately not shared:** YAML parsing (tools use different libraries), `ProjectType` enum (incompatible domain concepts)
- **Consumers (planned):** golangci-lint-auto-configure, oxlint-auto-configure

### go-checker-helpers

Shared utilities for BuildFlow **checker modules** (gomod-checker, nix-checker, flake-meta-checker, todo-checker, etc.). The lowest-abstraction SDK: a bag of utility functions, not a framework.

- **Repo:** [github.com/larsartmann/go-checker-helpers](https://github.com/larsartmann/go-checker-helpers)
- **Depends on:** go-finding, go-error-family
- **What it owns:**
  - Finding construction: `NewFinding[T]`, `NewSuggestFinding`, `SafeBuildFinding`, `NewReport`
  - Direct-fix pipeline: `ApplyDirectFixes` (filter-group-iterate-aggregate scaffold), `ApplyTextFixes`
  - Safe filesystem I/O: `SafeJoin`, `SafeReadFile`, `SafeStat`, `SafeWriteFile` (path containment, eliminating gosec G304)
  - Test assertions: `AssertNoFindings`, `AssertExpectedFindingsCount`, `WriteTestFile`
- **Design principle:** Enablement is DAG membership, not runtime flags. Each checker is a DAG node in BuildFlow's pipeline.

### BuildFlow/finding

A thin adapter package inside the BuildFlow monorepo. Bridges BuildFlow's internal `PrioritizedViolation` model to go-finding's format. One-way conversion, no circular dependencies.

- **Path:** `github.com/larsartmann/buildflow/finding`
- **Depends on:** go-finding, BuildFlow internals (model, domain, constants)
- **What it owns:**
  - Type mapping: `MapCategory`, `MapSeverity`, `MapPosition`, `MapFixStrategy`
  - Report building: `BuildFindingsReport`, `BuildWorkflowReport` (merges step results with dedup)
  - Baseline comparison: `DiffReports`, `DiffSummary`, `LoadBaseline`, `SaveBaseline`
- **Why it exists:** BuildFlow predates go-finding and has its own violation domain model. This package is the cost of having a pre-existing type system.

## The "Legacy Bridge" Pattern

BuildFlow/finding is labeled "legacy bridge" because it exists **only** to convert BuildFlow's pre-go-finding `PrioritizedViolation` type into `finding.Finding`. Every mapping function (`MapCategory`, `MapSeverity`, etc.) is boilerplate that bridges two separate type systems.

The newer SDKs eliminate this pattern at the source. `go-structure-linter` proved the approach: alias `type Issue = finding.Finding` and emit findings directly, so there is no intermediate domain type and no converter layer. `go-linter-sdk` codifies this pattern; `linter-autoconfigure-sdk` and `go-checker-helpers` apply the same principle to their tool categories.

The `go-linter-sdk` README quantifies the cost of the bridge pattern:

| Tool                | Converter LOC                      |
| ------------------- | ---------------------------------- |
| branching-flow      | 1,871                              |
| erraudit            | 1,214                              |
| go-structure-linter | 0 (`type Issue = finding.Finding`) |

## SDK Comparison

All three SDKs serve the same meta-pattern: extract shared plumbing from N tools that target go-finding. They differ in abstraction level and tool category.

| Aspect                 | go-linter-sdk                  | linter-autoconfigure-sdk            | go-checker-helpers                          |
| ---------------------- | ------------------------------ | ----------------------------------- | ------------------------------------------- |
| **For**                | Linters (find code issues)     | Auto-configurers (fix config files) | BuildFlow checkers                          |
| **Abstraction level**  | High (full rule framework)     | Medium (config lifecycle)           | Low (utility functions)                     |
| **Core abstraction**   | `Rule` + `Registry`            | `ConfigIssue` + `ProviderSpec`      | `NewFinding` + `ApplyDirectFixes`           |
| **Owns execution?**    | Yes (`Registry.Run`)           | Partially (`ProviderSpec`)          | No (call what you need)                     |
| **Opinionated shape?** | Yes (`Rule` interface)         | Yes (`Analyze`/`Repair`)            | No                                          |
| **Extra deps**         | none                           | go-atomic-write                     | go-error-family, samber/lo                  |
| **Eliminates**         | Violation-to-Finding converter | Config I/O + priority mapping       | Finding construction + fix loop boilerplate |

## Dependency Graph

```
go-error-family ← go-finding ← ┬─ go-linter-sdk
                               ├─ linter-autoconfigure-sdk ← go-atomic-write
                               ├─ go-checker-helpers
                               └─ BuildFlow/finding ← BuildFlow internals
```

All roads lead to go-finding. No SDK depends on another SDK. Each SDK depends only on go-finding (plus its own utility deps), keeping the dependency graph flat and composable.

## Consumer Version Sweep (2026-09-08, post-v1.7.0)

Survey of local consumer repos. **Migration relevance:** only `pipeline`-importing
consumers are affected by the v1.7.0 rollback-default change (ADR-016);
`core-only` consumers are unaffected (no behavior change in the core package).
Upgrade guide: [docs/guides/consumer-migration-v1.7.md](guides/consumer-migration-v1.7.md).

| Consumer                     | go-finding | Pipeline? | Rollback migration needed | Last activity |
| ---------------------------- | ---------- | --------- | ------------------------- | ------------- |
| go-humanize-linter           | v1.7.0     | no        | no                        | 2026-09-08    |
| go-linter-sdk (v0.3.0)       | v1.7.0     | no        | no                        | 2026-09-08    |
| BuildFlow                    | v1.6.0     | yes       | **yes**                   | 2026-09-06    |
| Code-Quality-Agent           | v1.6.0     | yes       | **yes**                   | 2026-09-06    |
| erraudit                     | v1.6.0     | yes       | **yes**                   | 2026-09-04    |
| go-structure-linter          | v1.6.0     | yes       | **yes**                   | 2026-09-06    |
| hierarchical-errors          | v1.6.0     | yes       | **yes**                   | 2026-08-18    |
| oxlint-auto-configure        | v1.6.0     | yes       | **yes**                   | 2026-09-04    |
| template-AUTHORS             | v1.6.0     | yes       | **yes**                   | 2026-09-04    |
| template-SECURITY            | v1.6.0     | yes       | **yes**                   | 2026-09-04    |
| Polish-Customs               | v1.6.0     | no        | no                        | 2026-09-04    |
| go-auto-upgrade              | v1.6.0     | no        | no                        | 2026-09-06    |
| go-business-rules            | v1.6.0     | no        | no                        | 2026-09-04    |
| go-checker-helpers           | v1.6.0     | no        | no                        | 2026-09-04    |
| golangci-lint-auto-configure | v1.6.0     | no        | no                        | 2026-09-03    |
| library-policy               | v1.6.0     | no        | no                        | 2026-09-04    |
| linter-autoconfigure-sdk     | v1.6.0     | no        | no                        | 2026-09-04    |
| template-readme              | v1.6.0     | no        | no                        | 2026-09-04    |
| branching-flow               | v1.4.1     | no        | no                        | 2026-09-04    |
| licenseforge                 | v1.4.1     | no        | no                        | 2026-09-04    |
| md-go-validator              | v1.4.1     | no        | no                        | 2026-09-04    |
| gomend                       | v1.4.0     | no        | no                        | 2026-09-04    |
| template-CLI                 | v1.3.0     | no        | no                        | 2026-09-04    |

**Migration note for pipeline consumers (v1.7.0):** the fix rollback scope changed
from all-files to per-file by default. If your tool relied on all-or-nothing
semantics, opt back in with `FixApplier.SetRollbackPolicy(RollbackPolicyAllFiles)`,
`Config.FixRollbackAllFiles`, config-file `fixRollbackAllFiles`, or CLI
`-fix-rollback-all`. Soft per-finding failures no longer abort runs — check
`ApplyReport.Outcomes` instead of treating any error as total failure. Full guide:
[docs/guides/consumer-migration-v1.7.md](guides/consumer-migration-v1.7.md).

Two lead consumers are bumped and verified on v1.7.0 (suites green; one
pre-existing unrelated failure in go-humanize-linter's generated-file detector,
present on v1.6.0 as well). Remaining consumers upgrade opportunistically — the
guide covers the one migration-relevant change.
