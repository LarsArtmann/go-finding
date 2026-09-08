# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. Seven tools detect issues; zero route them to remediation. This library solves that with:

1. **Unified Finding type** — Common representation for all tools
2. **Pipeline** — Automated detect → triage → fix → verify loop
3. **SARIF output** — Standard interchange format
4. **LSP integration** — IDE support

## Module Structure (Multi-Module Go Workspace)

Unix-style decomposition — each module does one thing well, composes via replace directives.

| Module       | Path                                               | External Deps                | Depends on     |
| ------------ | -------------------------------------------------- | ---------------------------- | -------------- |
| **Core**     | `github.com/larsartmann/go-finding`                | go-error-family              | —              |
| **Pipeline** | `github.com/larsartmann/go-finding/pipeline`       | x/sync, gogenfilter          | Core           |
| **Analysis** | `github.com/larsartmann/go-finding/analysis`       | x/tools                      | Core           |
| **CLI**      | `github.com/larsartmann/go-finding/cmd/go-finding` | yaml, go-output, gogenfilter | Core, Pipeline |

`go.work` coordinates all 4 modules for development. Each sub-module has `replace` directives for `GOWORK=off` CI/consumer builds.

## Key Files

| Area                | Files                                                                                                                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Core types**      | `finding.go`, `finding_methods.go`, `finding_validate.go`, `finding_equal.go`, `position.go`, `range.go`, `report.go`, `filter.go`, `merge.go`, `diff.go`, `format.go`, `json.go`, `id.go`, `errors.go`, `simple_fix.go` |
| **Named types**     | `severity.go`, `confidence.go`, `category.go`, `category_linter.go`, `tag.go`, `fix_strategy.go`, `suppression.go`, `branded_types.go`                                                                                   |
| **SARIF**           | `sarif_types.go`, `sarif_export.go`, `sarif_import.go` (hand-rolled, not go-sarif — see ADR #9)                                                                                                                          |
| **LSP**             | `lsp.go`                                                                                                                                                                                                                 |
| **Extensibility**   | `detector.go`, `adapter.go` (ToolAdapter[O]), `registry.go` (DetectorRegistry), `interval_index.go` (IntervalIndex[T])                                                                                                   |
| **gotoken**         | `gotoken/gotoken.go` (shared go/token utilities, public package)                                                                                                                                                         |
| **lockutil**        | `lockutil/lockutil.go` (shared sync.Locker helpers — `Locked`, `RLocked` — for generic mutex-guarded critical sections)                                                                                                  |
| **Pipeline**        | `pipeline/pipeline.go` (Run), `pipeline/pipeline_detect.go`, `pipeline/pipeline_iteration.go`, `pipeline/config.go`, `pipeline/config_file.go`, `pipeline/flight_recorder.go`                                            |
| **Fix engine**      | `pipeline/fix_engine.go`, `pipeline/fix_provider.go`, `pipeline/fix_applier.go`, `pipeline/fix_outcome.go`, `pipeline/fix_edit.go`, `pipeline/conflict.go`, `pipeline/goast/provider.go`                                 |
| **Pipeline extras** | `pipeline/stage_hook.go`, `pipeline/flight_recorder.go`, `pipeline/line_shift.go`, `pipeline/metrics.go`, `pipeline/retry.go`, `pipeline/partial.go`, `pipeline/generated_filter.go`                                     |
| **Analysis**        | `analysis/analysis.go` (go/analysis ↔ Finding)                                                                                                                                                                           |
| **Detectors**       | `cmd/go-finding/internal/detectors/govet.go`, `staticcheck.go`, `helpers.go`                                                                                                                                             |
| **CLI**             | `cmd/go-finding/main.go`, `config.go`, `registry.go`, `fix_provider_registry.go`, `generated_filter.go`, `output_adapter.go`                                                                                             |

## Testing & Build

```bash
nix run .#test                              # Run tests (all modules via go.work)
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector (workspace)
GOWORK=off go test ./...                    # Per-module isolation test (run in each module dir)
golangci-lint run ./...                     # Lint
bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 25  # Benchmark regression check
bash scripts/version-check.sh                                    # Verify version.go matches git tag
bash scripts/release-preflight.sh                                # Pre-tag structural gate (see below)
```

> **GOEXPERIMENT=jsonv2 required.** The project imports `encoding/json/v2`.
> All `nix run .#*` apps and devShells set this env var automatically. Direct `go`
> commands (outside `nix develop`) require `export GOEXPERIMENT=jsonv2` first — otherwise you get
> "build constraints exclude all Go files" errors.

### Release preflight (mandatory before any tag)

`scripts/release-preflight.sh` runs the structural gates as code (version drift,
replace audit, tag collisions, clean tree, GOWORK=off builds, docs checks).
Born from the v1.7.0 incident where sub-module go.mod files were tagged with a
stale core reference because the checklist lived in the operator's head.
`--bench`/`--stress` flags add the heavy gates.

### Dead-gate lesson (2026-09-08: three in one day)

A check that "passes" by not actually running is worse than no check. Found:
bench-check.sh awk pattern never matched benchstat output; pre-commit hook
dead via stale local `core.hooksPath`; version-drift.sh aborted silently under
`set -euo pipefail`. Rules: (1) every check script prints an explicit OK/FAIL
verdict line, (2) verify a new gate's FAIL path once by intentionally breaking
something, (3) never pipe a check script's output through filters that can cut
the verdict line.

## Module Dependencies (per go.mod)

| Module                      | Production Deps              | Test Deps         |
| --------------------------- | ---------------------------- | ----------------- |
| **Core** (`.`)              | go-error-family              | ginkgo/v2, gomega |
| **Pipeline** (`pipeline/`)  | x/sync, gogenfilter          | ginkgo/v2, gomega |
| **Analysis** (`analysis/`)  | x/tools                      | (stdlib testing)  |
| **CLI** (`cmd/go-finding/`) | yaml, go-output, gogenfilter | gomega            |

### Old Dependencies (now isolated to sub-modules)

- `golang.org/x/tools` — go/analysis framework (**analysis module only**)
- `golang.org/x/sync` — errgroup (**pipeline module only**)
- `github.com/go-faster/yaml` — YAML config (**CLI module only**)
- `github.com/onsi/ginkgo/v2` + `gomega` — BDD testing (core + pipeline test deps)
- `github.com/LarsArtmann/gogenfilter/v3` — Auto-generated Go file detection (pipeline + CLI)
- `github.com/larsartmann/go-output` — CLI output formatting (CLI module only)

## Design Principles

1. **Minimal dependencies** — core module keeps a small, deliberate dependency surface
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — Conversions (SARIF, LSP) preserve all data via Metadata/Tags/Data fields
4. **One extensibility field** — `Finding.Metadata` is `map[string]string`. NO `Properties map[string]any`
5. **Compatible** — Works with existing Go analysis tools
6. **Resilient** — Retry logic, partial success, nil-safe metrics
7. **Multi-module** — Unix-style decomposition: each module has a single purpose and composes independently. Dual go.work + replace strategy.

## Important Behaviors (Gotchas)

_Updated 2026-09-08 diet pass; pre-diet text archived in `docs/planning/archived/2026-09-08_agents-gotchas-diet.md`._

### Environment & workflow

- **GOEXPERIMENT=jsonv2 required** — project uses `encoding/json/v2` (Go 1.26 experimental). All `nix run .#*` apps and devShells export it. Direct `go build`/`go test` outside nix need `export GOEXPERIMENT=jsonv2`; the per-module path needs BOTH `GOWORK=off` and `GOEXPERIMENT=jsonv2`. Drop when json/v2 stabilizes (Go 1.27+, tracked in ROADMAP).
- **Run `go test` invocations SEQUENTIALLY on this machine** — two concurrent `go test`/`go build` runs sharing `GOCACHE=/mnt/buildcache/go-build` produce transient `[build failed]` errors. A plain retry succeeds; don't misdiagnose as code/toolchain problems.
- **Formatting has three coordinated signals** — treefmt is not on the devShell PATH; use `nix fmt` (canonical gate, also in the pre-commit hook via `--fail-on-change`), `golangci-lint fmt` (local autofix; `.golangci.yml` formatters mirror treefmt rules: gofumpt, goimports, golines@120), and dprint (markdown/JSON/YAML, pre-commit). Do NOT let the two Go formatters' rules diverge. If the pre-commit hook seems dead, check `git config core.hooksPath` — a stale `.githooks` value once silently disabled it.
- **CI billing superseded (2026-09-08 evening)** — the account-switch gate was bypassed by making the repo PUBLIC: Actions on public repos are free/unlimited. Post-flip `ci.yml` dispatched (run 34274674104) — verify every job; the last private-repo run had ALL 20 jobs fail in 43s from billing (no logs). Until the first green public run, local gates remain the quality bar (race x4, lint x4, structural scripts, go-arch-lint, dprint, `nix flake check`, stress `ginkgo --repeat=20 --race` as MANDATORY release gate).
- **Repo is PUBLIC (2026-09-08)** — visibility flipped via `gh repo edit`; module proxy resolution verified: plain `go get github.com/larsartmann/go-finding@v1.8.0` works with NO `GOPRIVATE`. Public-repo Actions are free, which dissolves the CI billing blocker.
- **Multi-module release tagging** — sub-modules need directory-prefixed tags (`pipeline/v*`, `analysis/v*`, `cmd/go-finding/v*`); core uses unprefixed `v*`; sub-modules have no version.go. See `docs/release-procedure.md`.
- **version-check.sh needs `--match 'v[0-9]*'`** — plain `git describe` picks sub-module tags alphabetically first. Any script resolving the core version from tags must use this flag.
- **CI scripts guard the 4-module structure** — `replace-audit.sh`, `version-drift.sh`, `test-naming.sh`, `go-work-sync.sh`, `docs-freshness.sh` (backtick spans + links only), `docs-api-check.sh` (documented identifiers must exist in code), `json-deterministic-check.sh`, plus `go-arch-lint` (`.go-arch-lint.yml`, 11 components, one-directional flow cli->pipeline->core, analysis->core). All wired into ci.yml.

### Core type rules

- **Branded types prevent mixups** — `ID`, `RuleName`, `ToolName`, `FilePath`, `GroupID` are distinct string types (`branded_types.go`). Use `finding.ID("x")` / `finding.FilePath("p")`; string literals auto-convert. JSON marshals as plain string. `GroupByFile` returns `map[FilePath][]Finding`.
- **Confidence is a named type** — `type Confidence float64` with `IsValid()`/`Clamp()`; `NewFinding`/`Builder` accept it (not raw float64), defaulting to `ConfidenceFull`; `ParseConfidence("high")` inverts `String()`, errors match `ErrInvalidConfidence` via `errors.Is`.
- **Position.Offset uses -1 sentinel** — `Position{}` zero value means byte 0; constructors set Offset=-1 for "unset". Check with `HasOffset()` (>= 0). Line=0 file-only positions are valid: `validateIdentity()` uses `HasFile()`; `Position.IsValid()` still requires Line>0 (backward compat).
- **GenerateID is length-prefixed** — `writeLenField` (uint32 big-endian) prevents hash collisions when field values contain colons.
- **Range end conventions** — `Range.EndOrStart` / `EndOffsetOrStart` give the effective end for single-point ranges (`End.Line == 0` or `End.Offset < 0` means "same as Start"); overlap/intersection use them.
- **FixStrategy normalized** — `NormalizeFixStrategy()` converts "" to "none" (Builder.Build, SARIF import, Equal short-circuit). `FixStrategyAI` is reserved (no backend). `HasFix()` requires BeforeCode/AfterCode for Direct.
- **tagsEqual fast path** — `Equal()` checks `slices.Equal` before clone+sort; same-order tags cost 0 allocations.
- **Validate() decomposed** — 6 per-field validators in `finding_validate.go`; add new rules there.
- **Severity aliases are API** — `RegisterSeverityAlias()`/`LookupSeverityAlias()` (RWMutex-guarded global map); `SeverityFromLevel(level, fallback)` maps strings incl. aliases; `PriorityString()` uses the `severityPriorities` map; `Badge()` derives from `Emoji()` (update only `Emoji()` for new mappings).
- **Deterministic JSON is mandatory** — all production marshal calls use `marshalOpts` / `prettyMarshalOpts` (`json.Deterministic(true)`); `marshalJSONString` wraps marshal-to-string. `encoding/json/v2` serializes map keys in unspecified order otherwise. Enforced by `json-deterministic-check.sh`. Wire types are separate structs where tags differ (`fixEditJSON`, `fixOutcomeJSON`, `fixApplyResultJSON`); errors serialize as message strings.
- **FindingError implements go-error-family** — `ErrorCode()` = `"finding.<category>"`, `ErrorFamily()` maps to errorfamily families; `errorfamily.Classify(err)` works on go-finding errors. See ADR #15.
- **`must[T]`** — `errors.go`; any new Must-constructor delegates to it.
- **doc.go must match current names** — after ANY rename, grep `doc.go` for the old symbol (godoc prose misleads otherwise).
- **Consumer count grows** — last audit (2026-07-22) counted 22 consumers (14 with Go code); never assert a fixed number without checking latest data.

### Core API surface

- **Report** — zero-value safe (value mutex); `findings` unexported: use `FindingsSnapshot()` (deep copy), `All()`, `FindByID()`; `NewReportFromFindings(tool, findings)` is the one-step creator.
- **Builder/Template** — `BuildOrDefault()` returns zero `Finding{}` on invalid input; `Template` stamps common fields (`NewTemplate` + `With*`), `Template.Builder()` returns a chainable `*Builder` for per-finding overrides. Convenience APIs: `ApplySimpleFixes` (core string replace), `CheckBinary`/`RunCmd` (external tool helpers -> `NewIOError`), `FormatTextRich` (emoji; `FormatText` keeps `[SEVERITY]`), `FormatTable`. Full docs: doc.go, README, `docs/guides/`.
- **Suppression** — `IsSuppressedAt` uses `Suppression.IsActive(now)` (valid Kind + Rule AND not expired); invalid suppressions are inactive.
- **GroupID groups findings** — optional `Finding.GroupID` (branded); JSON `groupId` (omitempty); in `Equal()`; builder `WithGroupID`; SARIF property `go-finding/groupId`; LSP via `LSPDiagnosticData.GroupID`; `Report.GroupFindings()` returns active grouped findings (map — order unspecified). See `docs/guides/finding-groups.md`.
- **LSP fidelity** — `LSPDiagnosticData` on `diag.Data` preserves ID/Severity/FixStrategy/Confidence/Category/Tags/code/Snippet/Suppression/Metadata/RelatedFindingIDs; round-trip lossless incl. SeverityCritical (LSP collapses to Error). `ToLSP()` re-emits `Tags` from `Metadata[LSPDiagnosticTagsKey]` (malformed entries skipped; negatives rejected).
- **SARIF options** — `ToSARIFWithOpts(WithIncludeSuppressed(), WithMinSeverity(sev))`; deprecated `ToSARIFFiltered` retained as alias. SARIF is hand-rolled (ADR #9).
- **context.Context on I/O** — `WriteSARIF`, `FindingsFromSARIF`, etc. take context first.

### Pipeline behavior

- **Pipeline.Run() is single-use** — second call returns `errAlreadyRan`.
- **StageHooks** — `Config.StageHooks` with `StageHook`/`StageHookFunc`; both StageBefore and StageAfter errors abort.
- **StageTiming closure must run exactly once** — `Metrics.RecordStage` uses `+=` (`metrics.go:50`); invoking the done-closure on both success and error paths double-records (past bug in `pipeline_iteration.go`).
- **RetryConfig validation uses named sentinels** — `pipeline/retry.go`; NEVER inline `errors.New` in validation returns (breaks `errors.Is`).
- **ResolveSafePath is the path traversal security boundary** — `pipeline/path_safety.go` resolves symlinks and verifies containment within root before ANY filesystem op on Finding paths. `ResolveRoot` + `ResolveSafePathFrom` batch-cache (resolve root once); `ResolveSafePath` is the single-call wrapper.
- **FixEngine** — byte-level, descending-offset application; all edits resolve against one original content snapshot; O(F+R) single pass. Provider chain: Offset -> Line -> Substring (fallback), custom providers prepended; `lineIndexAware` caches the line index lazily per file. `SubstringProvider` is column-aware.
- **Fix outcomes (v1.7.0)** — `ApplyWithOutcomes` returns `FixApplyResult` with one `FixOutcome` per input finding (input order): applied / no-change / refused / conflict / invalid / failed. `Apply`/`ApplyWithConflicts` delegate (unchanged shapes). `OutcomeFor`/`OutcomeCounts`/`HasErrors` query. Failed `Err` values are typed `*finding.FindingError` with position; `errors.Is`/`As` reach the provider cause. Guide: `docs/guides/outcomes.md`.
- **Rollback is per-file by default (ADR-016)** — `RollbackPolicyFailingFile` restores only the failing file; opt into all-or-nothing via `SetRollbackPolicy(RollbackPolicyAllFiles)` / `Config.FixRollbackAllFiles` / config `fixRollbackAllFiles` / CLI `-fix-rollback-all`. Soft per-finding failures never abort: applied edits stay, errors surface in outcomes + joined error. `ApplyWithReport` returns `ApplyReport`; `FailedOutcomes()` isolates failures; `Metrics.RecordOutcome`/`OutcomeCounts` aggregate; CLI prints `Fix outcomes:` summary.
- **`RolledBack` lists backed-up files, not modified files** — soft-failed files are still backed up; restoring them is a content no-op but they appear in `RolledBack` and the `(rolled back: ...)` error text. Tests must expect them.
- **NewFixApplier returns error** — propagates backup dir creation failures.
- **LineShiftMap** — `ShiftedPosition` shifts line+column; `ShiftedRange` shifts both endpoints.
- **FlightRecorderHook** — wraps Go 1.25 `runtime/trace.FlightRecorder`; snapshots on `SlowStageThreshold` or `Snapshot(ctx, reason)`; `OnStageEvent` never errors; degraded mode when Go's singleton recorder is taken (check `Degraded()`); `Close()` waits for in-flight snapshots. Config via `FlightRecorderFileConfig` + `ResolveFlightRecorder()` (5 string-encoded fields) with full CLI parity (`flightRecorder` config section vs `-trace` flags) — keep both structs in sync. CLI flags: `-trace`, `-trace-dir`, `-trace-slow`.
- **Analysis BeforeCode** — `analysis.FromDiagnostic` reads source files to extract `BeforeCode` from TextEdits.

### Testing & tooling

- **math/rand split** — production `math/rand/v2`; tests `math/rand` (v1, `testing/quick` constraint).
- **testify is transitive only** — `// indirect` via ginkgo/slim-sprig; not used directly (banned but unavoidable).
- **NewParallelGomega** — per-module `testutil_test.go` helper (Helper + Parallel + NewWithT in one). `paralleltest` linter disabled because it can't trace `t.Parallel()` through the helper; re-enabling means inlining or nolint-ing 100+ tests.
- **makezero `always: false` (intentional)** — `always: true` flags idiomatic `make+copy` (23 FPs); `false` still catches real `make+append` over-allocation.
- **Stress gate is mandatory** — `ginkgo -r --race --repeat=20 --skip-package=examples` (core, pipeline), `go test -race -count=20` (analysis, CLI) before any tag (release-procedure step 4).
- **`pipeline/examples/`** — one runnable program per subdir + `example_compile_test.go` (`go build -o /dev/null`), mirroring root `examples/`.

## CLI Features

- Built-in govet and staticcheck detectors
- Text, markdown, CSV, TSV, JSON, SARIF output (markdown/CSV/TSV via go-output adapter)
- YAML/JSON config (`-config`), severity filter (`-min-severity`), profiling
- `-filter-generated` — removes findings from auto-generated files (sqlc, protobuf, etc.)
- `-fix-provider go-ast` — enables AST-aware fix provider
- `-fix-rollback-all` — opt into all-or-nothing rollback (default: per-file)
- `-byte-level-conflict` — precise overlap detection
- `-trace` — enable Go execution trace flight recorder for diagnostics (`-trace-dir`, `-trace-slow`, `-trace-max-files`, `-trace-gzip` for config)
- `Fix outcomes:` stderr summary after fix stages (canonical order, nonzero counts only)
- Config-file `flightRecorder` section — alternative to `-trace` flags; supports `enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes` (full parity with pipeline `FlightRecorderFileConfig`)
- Dynamic detector registry (`RegisterDetector`)

## Architecture Decisions

ADR log: `docs/architecture-decisions.md` (#16 = per-file rollback default, v1.7.0). Key structural calls:

- **SARIF hand-rolled, not go-sarif** — custom property bag, streaming + context. ADR #9.
- **go-error-family integration** — core dependency for unified error classification. ADR #15.
- **Pipeline split** — `pipeline.go` + `pipeline_detect.go`, both under 350 lines.
- **FixProvider chain** — Offset -> Line -> Substring (fallback), custom prepended; `lineIndexAware` lazy line-index caching; GoASTProvider in `pipeline/goast/` (opt-in go/parser).
- **IntervalIndex[T]** — generic sorted-slice O(n + k) overlap queries, used by Correlate (interval-tree NO-GO note in ROADMAP).
- **DetectorRegistry** — thread-safe plugin architecture (`Register`/`Build`/`BuildAll`); **MergeIter** — streaming `iter.Seq[Finding]` merge with dedup.
- **ConfigFile** — JSON config loading with `ResolveDetectors`/`ResolveProviders`.
- **go-output CLI adapter** — markdown/CSV/TSV live in the CLI module only; root `finding` stays without it. See `docs/PRO_CONTRA_go-output-integration.md`.
- **Naming history** — `FindingTransformer` (was FindingProcessor), `Conflict` (was ConflictInfo), `-min-severity` (was `-severity`, alias retained), `CategoryOf` (was GetCategory).
- **Pipeline convenience functions** — `pipeline.Detect(ctx, detectors...)`, `pipeline.ApplyToContent(content, fixes)`.
- **FixEngine guide** — `docs/guides/fix-engine.md`; outcomes guide `docs/guides/outcomes.md`.

## Test Organization

**One test file per production file.** All tests for a subject live in `<subject>_test.go` — no `_extra`/`_bugfix`/`coverage` suffixes (split by subject, put regression tests in the parent file, never name files after metrics). Shared helpers in `testutil_test.go`; one `example_test.go` per package (don't fragment).

## Removed APIs (v1.0.0)

All deprecated APIs from v0.6.0–v0.9.0 have been removed. No deprecated APIs remain.

See `docs/MIGRATION_v1.0.md` for migration details.

---

_Assisted-by: Crush <crush@charm.land>_
