# AGENTS.md - go-finding

## Project Overview

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. Seven tools detect issues; zero route them to remediation. This library solves that with:

1. **Unified Finding type** — Common representation for all tools
2. **Pipeline** — Automated detect → triage → fix → verify loop
3. **SARIF output** — Standard interchange format
4. **LSP integration** — IDE support

## Module Structure (Multi-Module Go Workspace)

Five modules, one purpose each: **Core** `github.com/larsartmann/go-finding`;
**Pipeline** `.../pipeline`; **Analysis** `.../analysis`; **Tool SDK**
`.../toolsdk`; **CLI** `.../cmd/go-finding`. `go.work` coordinates all five;
cmd/go-finding must have NO replace directives (see Module Dependencies).

## Key Files

Moved to [docs/key-files.md](docs/key-files.md) (2026-11-08 diet pass; content unchanged).

## Testing & Build

```bash
nix run .#test                              # Run tests (ROOT module only — ./... does not cross module boundaries, even with go.work active; verified 2026-09-17. Test sub-modules per the GOWORK=off line below)
nix run .#bench                             # Run benchmarks
nix run .#lint                              # Run linter
go test -race -count=1 ./...                # Full suite with race detector (workspace)
GOWORK=off go test ./...                    # Per-module isolation test (run in each module dir)
golangci-lint run ./...                     # Lint
bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 250  # Benchmark regression check
bash scripts/version-check.sh   # version.go vs git tag (run after ANY tag push)
bash scripts/release-preflight.sh [--tag vX.Y.Z] [--links]  # see below
bash scripts/error-audit.sh     # erraudit gate, all 5 modules (see below)
```

### Release preflight (mandatory before any tag)

`scripts/release-preflight.sh` runs the structural gates as code (born from the
v1.7.0 stale-sub-module-require incident); `scripts/release-preflight-selftest.sh`
injects failure classes (tag collision, version drift, post-tag missing,
incomplete set, stale stamp, stamp wiring) into a disposable worktree and
asserts preflight FAILs on each. Flags: `--bench`, `--stress`, `--post-tag`,
`--tag vX.Y.Z` (stamp gate), `--links` (compare-link verification).

**Preflight is lockstep/core-centric for target tags** — it derives all five
target tags from core's `version.go`, so a sub-module-ONLY release (e.g.
toolsdk while core stays put) fails the tag-collision checks as false
positives; the sub-module-only recipe is in docs/release-procedure.md. The
GOWORK=off build/tidy loops cover all five modules since 2026-11-08 (the old
toolsdk-skip blind spot is closed).

The Release workflow (`release.yml`) runs under a queueing concurrency group
(`cancel-in-progress: false`), so a manual dispatch can no longer race the
tag-push auto-trigger (v1.9.2 produced two concurrent GoReleaser runs, one 422
duplicate-asset loser).

### erraudit gate (2026-09-11)

`bash scripts/error-audit.sh` runs `erraudit ./... --type-aware` in all 5
modules; 0 violations is green. Policy: sentinel errors + `*FindingError` +
stdlib `fmt.Errorf` wrapping (samber/oops REJECTED at 3/10 fit;
docs/research/erraudit-violation-analysis.md). NEVER interpret runs with
`--enforce-samber-oops`/`--enforce-generic-return` as defect lists (the
~79-violation reports are artifacts of flags that contradict the model).
`nolint-audit` staleness: re-verified 2026-11-08 (erraudit 5c841cf) — the
2026-09-23 false-stale does NOT reproduce (27 needed / 0 stale in all modes);
no upstream issue was filed. The only safe staleness test remains
remove-one-directive -> gate red -> restore. Blessed invocation:
`nix run .#error-audit`. CI wiring blocked: the erraudit repo is PRIVATE (no
`go install` from public runners).

### Dead-gate rules (born 2026-09-08, three dead gates in one day)

(1) every check script prints an explicit OK/FAIL verdict line; (2) prove a
gate's FAIL path once by breaking something; (3) never pipe a check's output
through a filter that can cut the verdict line (the `| tail` exit-code trap
bites twice now); (4) "verified by X" requires X FINISHED and its output read.

## Module Dependencies

Per-module deps live in each `go.mod` (source of truth); sub-modules carry
`replace github.com/larsartmann/go-finding => ../` for GOWORK=off builds while
**cmd/go-finding must have NO replace directives** (`scripts/replace-audit.sh`
enforces it). Consequence: the CLI is pinned to PUBLISHED siblings — no new
sibling API until that module is tagged and the CLI require bumped (dedupe
CLI-side instead; learned 2026-09-22/23). Old deps are isolated per module
(x/tools -> analysis, x/sync -> pipeline, yaml + go-output -> CLI).

## Design Principles

1. **Minimal dependencies** — core keeps a small, deliberate dependency surface
2. **Immutable** — Findings are data, not state machines
3. **Lossless** — SARIF/LSP round-trips preserve everything via Metadata/Tags/Data
4. **One extensibility field** — `Metadata map[string]string`; NO `Properties map[string]any`
5. **Compatible + resilient** — works with existing Go tools; retry, partial success, nil-safe metrics
6. **Multi-module** — one purpose per module, composed via go.work + published tags

## Important Behaviors (Gotchas)

_Diet passes 2026-09-08 + 2026-11-08; full earlier text lives in those dates'
archived status/planning docs; release war stories moved to docs/release-procedure.md._

### Environment & workflow

- **GOEXPERIMENT=jsonv2 DROPPED (2026-09-23)** — `encoding/json/v2` is GA since Go 1.27; anything telling you to export it is stale. The per-module path still needs `GOWORK=off`.
- **CI/Release `go-version` pins must track the toolchain floor** — a floor bump with a stale `go-version:` pin fails EVERY CI job while local gates stay green (it silently killed 6+ master runs + both v1.13.0 Release runs). Rule: a floor bump is a SAME-COMMIT edit of every `go-version:` in both workflows.
- **Run `go test` invocations SEQUENTIALLY on this machine** — concurrent go runs sharing `GOCACHE=/mnt/buildcache/go-build` produce transient `[build failed]` errors (retry succeeds; don't misdiagnose). Bump go.mod requires only on a quiet tree: editing requires mid-run produced phantom `could not import ... (invalid package name: "")` failures in go/packages tests (v1.12.0 stress run 1).
- **Formatting has three coordinated signals** — `nix fmt` (canonical gate), `golangci-lint fmt` (autofix; `.golangci.yml` mirrors treefmt: gofumpt, goimports, golines@120), dprint (markdown/JSON/YAML). If the pre-commit hook seems dead, check `git config core.hooksPath` — a stale `.githooks` value once silently disabled it.
- **`nix flake check --all-systems` DECLINED (2026-09-08)** — darwin derivations with no remote builders would fail it; plain `nix flake check` stays the gate. Revisit only if darwin builders are added.
- **Repo state (2026-09-08)** — PUBLIC (free Actions, no `GOPRIVATE` needed for any module; secret scanning + push protection on); pkg.go.dev renders for all five modules. Lychee excludes still-private sibling repos, this repo's compare/releases/issues links (they 500/502/504 randomly for anonymous checkers â two CI runs flaked on DIFFERENT valid pairs; scripts/verify-compare-links.sh owns them), and v0.1.x compares.
- **Tagged-commit CI `module-isolation` is EXPECTED red mid-train** (the CLI go.sum cannot carry checksums for tags that do not exist yet); the phased-red table in docs/release-procedure.md lists every structurally-red job per train phase.
- **v1.14.0 train lessons (2026-10-11, distilled)** — stamp-in-tagged-commit, cmd-tag-at-resync-commit, and the replace-masking consumer-breakage class are now GATES (stamp-gate.sh, version-lattice.sh, consumer-compat --head) and procedure sections. Local --bench numbers under heavy machine load are invalid; the CI benchmark job is authoritative. Full war story: docs/release-procedure.md + the 10-11 status report.
- **v1.15.0 train lessons (2026-10-12, distilled)** — compare-ref convention + lychee verdict path + phased-CI-red reality + the tidy-re-resolves-pins trap are all in docs/release-procedure.md now; the mechanical verdict is scripts/verify-compare-links.sh.
- **v1.15.1 stamp miss (2026-11-04, REPEAT of v1.14.0) + cold-cache go.sum rot** — a green PRE-tag CI run proves nothing about the tag; the post-push checks are in release-procedure.md (stamp-gate section). go.sum completeness is now GATED (`go-worksum-check.sh`). Restoration sequence for an incomplete train: complete go.sums → post-tag stamp fix → sibling requires → CHANGELOG section for the tag (changelog-drift) → sibling tags one-per-push → cmd resync last.
- **version-check.sh needs `--match 'v[0-9]*'`** — plain `git describe` picks sub-module tags alphabetically first. Any script resolving the core version from tags must use this flag.
- **Dependabot gomod PRs need a manual fix-up commit (2026-09-27 pattern)** — a `go` directive bump leaves `go.work` behind (every workspace go command dies until `go.work`'s `go` line is bumped + `go work sync` + per-module tidy), and the vendorHash goes stale (take the hash the CI nix job prints). Merge with `gh pr merge --admin` — admin bypass is the direct-push path; the required-context mismatch that forced it is fixed since 2026-11-08 (wrapper jobs) but branch protection still needs the expanded names for PRs. The gomod config is collapsed to `/` since 2026-11-08, so root PRs cover all modules.
- **CI scripts guard the 5-module structure** — all `scripts/*.sh` gates plus `go-arch-lint` (`.go-arch-lint.yml`, one-directional flow cli->pipeline->core) are wired into ci.yml; the files-to-gates map in docs/release-procedure.md says which gate bites on which change.
- **go.mod/go.sum bumps invalidate `flake.nix` `vendorHash`** — complete the bump with `go mod tidy` per module BEFORE updating the hash, then `nix flake check` (or iterate the bootstrap loop twice).
- **Branch protection is configured for DIRECT pushes** — required checks with `enforce_admins: false` (admins bypass merge gates; the checks bind future PRs; the bare-context mismatch is wrapper-fixed since 2026-11-08). Release tags cannot be deleted/force-updated (`protect-release-tags`). Private vulnerability reporting is ENABLED.
- **Keep the golangci-lint-action pin in sync with the nix binary** — CI pinned v2.13.1 while nixpkgs shipped v2.13.2; patch drift changed wsl_v5/varnamelen verdicts, so local "0 issues" and CI red disagreed (f/26, 2026-09-23). When buildflow bumps the nix golangci-lint, bump `version:` in ci.yml's lint job to match.
- **Buildcache prune policy (decided 2026-11-08)** — recovery stays `go clean -cache`; a periodic cron is deliberately NOT added (the pre-push-verify df guard FAILs below 5 GiB, which has caught every real event; a cron pruning a shared cache mid-build is a bigger hazard than the disk filling). Act on the guard, not on a schedule.
- **gopls SA1012 noise on nil-ctx contract tests (accepted 2026-11-08)** — toolsdk's `DryRunFromContext(nil)` tests carry `//nolint:staticcheck` which the golangci gate honors (0 issues) but gopls ignores, so the editor shows SA1012 on those lines. Switching to `//lint:ignore` was tested and REJECTED: golangci's staticcheck then reports the directive itself (1 finding). Accept the editor-only noise; the gate is the truth.
- **`/mnt/buildcache` can hit 100%** — the shared GOCACHE grew to 167G and made every write fail ("no space left on device" from go build). Recovery: `go clean -cache` (and `-fuzzcache` after fuzz campaigns). If builds fail with space errors, check `df /mnt/buildcache` first, not the code.
- **Push release tags ONE PER PUSH** — >3 tags per push creates NO events (v1.9.0); batched pushes race the Release queue and cancel runs (v1.12.0); queued reruns cancel each other. Watch each tag's run to a terminal state; full flow in docs/release-procedure.md "Tag pushing".
- **Master BRANCH pushes can drop the CI event too (2026-10-01, f394f54)** — the push updated origin/master but created NO ci.yml run (and no cancel, despite ci.yml's `cancel-in-progress: true` concurrency — proof the event never fired, same class as the v1.9.0 tag incident). After ANY master push, verify a CI run exists for the pushed SHA (`gh run list --workflow=ci.yml --json headSha`); if absent, check nothing is queued/in progress, then `gh workflow run ci.yml --ref master` and watch it green before tagging.
- **CI job runtime must fit its runner class** — before adding or retuning a CI job, state its expected runtime vs the runner (2-core public runners need ~45min for the count=10 benchmark job; a 15min cap silently cancelled it for a full session). A gate that times out is a dead gate.
- **`gh -R` for all cross-repo operations** — local directory names lie after renames (e.g. `~/projects/hierarchical-errors` is erraudit's pre-rename remote; `gh issue create` there filed into erraudit twice). Always `gh -R owner/repo ...` and `gh repo view` before the first cross-repo write.
- **Author commit messages AFTER `git status` confirms the batch** — the auto-commit daemon absorbs staged files seconds after staging; grand batch-sized messages routinely land over 1-3 file diffs. Check what is actually staged before writing the message.

### Core type rules

- **Branded types prevent mixups** — `ID`, `RuleName`, `ToolName`, `FilePath`, `GroupID` are distinct string types (`branded_types.go`). Use `finding.ID("x")` / `finding.FilePath("p")`; string literals auto-convert. JSON marshals as plain string. `GroupByFile` returns `map[FilePath][]Finding`.
- **Confidence is a named type** — `type Confidence float64` with `IsValid()`/`Clamp()`; `NewFinding`/`Builder` accept it (not raw float64), defaulting to `ConfidenceFull`; `ParseConfidence("high")` inverts `String()`, errors match `ErrInvalidConfidence` via `errors.Is`.
- **Position.Offset uses -1 sentinel** — `Position{}` zero value means byte 0; constructors set Offset=-1 for "unset". Check with `HasOffset()` (>= 0). Line=0 file-only positions are valid: `validateIdentity()` uses `HasFile()`; `Position.IsValid()` still requires Line>0 (backward compat).
- **GenerateID is length-prefixed** — `writeLenField` (uint32 big-endian) prevents hash collisions when field values contain colons.
- **Range end conventions** — `Range.EndOrStart` / `EndOffsetOrStart` give the effective end for single-point ranges (`End.Line == 0` or `End.Offset < 0` means "same as Start"); overlap/intersection use them.
- **FixStrategy normalized** — `NormalizeFixStrategy()` converts "" to "none" (Builder.Build, SARIF import, Equal short-circuit). `FixStrategyAI` is reserved (no backend). `HasFix()` for Direct requires BeforeCode/AfterCode OR an `Edits` list (`HasEditList()`); `HasCodeChange()` accepts any of the three.
- **Multi-edit fixes (issue #36, ADR #19)** — `TextEdit{Start, End Position; NewText}` (`text_edit.go`) + `Finding.Edits` (JSON `edits`) is the authoritative machine representation; BeforeCode/AfterCode stay the first-edit display summary. Insertion: `End` unset (`Offset` -1) or `End == Start`; `Position{}` is byte 0, NOT unset. Edit-list equality is order-independent (`editsEqual`, used by `Equal()`). Validate accepts edit-only Direct fixes; per-edit errors wrap as `finding.Edits[%d] is invalid`. SARIF round-trips edit lists incl. byte offsets (`sarifRegion.ByteOffset`/`ByteLength` are `*int` so offset 0 survives omitempty). `ApplySimpleFixes` REFUSES `len(Edits) > 1` (Reason points at pipeline.FixApplier) — never half-applies. Builder: `WithEdits(...)`. GenerateID excludes Edits deliberately.
- **tagsEqual fast path** — `Equal()` checks `slices.Equal` before clone+sort; same-order tags cost 0 allocations.
- **Validate() decomposed** — 6 per-field validators in `finding_validate.go`; add new rules there.
- **Severity aliases are API** — `RegisterSeverityAlias()`/`LookupSeverityAlias()` (RWMutex-guarded global map); `SeverityFromLevel(level, fallback)` maps strings incl. aliases; `PriorityString()` uses the `severityPriorities` map; `Badge()` derives from `Emoji()` (update only `Emoji()` for new mappings).
- **Deterministic JSON is mandatory** — all production marshal calls use `marshalOpts` / `prettyMarshalOpts` (`json.Deterministic(true)`); `marshalJSONString` wraps marshal-to-string. `encoding/json/v2` serializes map keys in unspecified order otherwise. Enforced by `json-deterministic-check.sh`. Wire types are separate structs where tags differ (`fixEditJSON`, `fixOutcomeJSON`, `fixApplyResultJSON`); errors serialize as message strings.
- **FindingError implements go-error-family** — `ErrorCode()` = `"finding.<category>"`, `ErrorFamily()` maps to errorfamily families; `errorfamily.Classify(err)` works on go-finding errors. See ADR #15.
- **`must[T]`** — `errors.go`; any new Must-constructor delegates to it.
- **doc.go must match current names** — after ANY rename, grep `doc.go` for the old symbol (godoc prose misleads otherwise).
- **Module-coverage evidence lives in Summary** — `Summary.FilesScanned` is always serialized ("scanned 0" is evidence, not an omission); `Summary.SkippedModules` carries nested go.mod modules a tool deliberately skipped; module-agnostic tools leave it nil and their JSON is unchanged.
- **Consumer count grows** — audited 2026-11-08 via code search: 23 consumer repos require go-finding on their default branch (21 on core v1.15.1, 2 on v1.15.0); submodule reach: toolsdk x14, pipeline x7, analysis x2. Table + notes in docs/ecosystem.md; never assert a fixed number without re-running the sweep.

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
- **FixEngine** — byte-level, descending-offset application; all edits resolve against one original content snapshot; O(F+R) single pass. Provider chain (v1.14.0): EditList -> Offset -> Line -> Substring (fallback), custom providers prepended; `lineIndexAware` caches the line index lazily per file. `SubstringProvider` is column-aware. Applied counts are PER FINDING (not per edit): `apply()` tracks each edit's owning input finding (`editOwner`, parallel to `allEdits`) and `applyEditsWithConflicts` appends a finding once — a multi-edit fix counts once everywhere, O(1), no post-hoc dedup (an earlier full-equality O(n²) dedup regressed GoAST_1000 by +706% and was replaced).
- **Fix outcomes (v1.7.0)** — `ApplyWithOutcomes` returns `FixApplyResult` with one `FixOutcome` per input finding (input order): applied / no-change / refused / conflict / invalid / failed. `Apply`/`ApplyWithConflicts` delegate (unchanged shapes). `OutcomeFor`/`OutcomeCounts`/`HasErrors` query. Failed `Err` values are typed `*finding.FindingError` with position; `errors.Is`/`As` reach the provider cause. Guide: `docs/guides/outcomes.md`.
- **Rollback is per-file by default (ADR-016)** — `RollbackPolicyFailingFile` restores only the failing file; opt into all-or-nothing via `SetRollbackPolicy(RollbackPolicyAllFiles)` / `Config.FixRollbackAllFiles` / config `fixRollbackAllFiles` / CLI `-fix-rollback-all`. Soft per-finding failures never abort: applied edits stay, errors surface in outcomes + joined error. `ApplyWithReport` returns `ApplyReport`; `FailedOutcomes()` isolates failures; `Metrics.RecordOutcome`/`OutcomeCounts` aggregate; CLI prints `Fix outcomes:` summary.
- **`RolledBack` lists backed-up files, not modified files** — soft-failed files are still backed up; restoring them is a content no-op but they appear in `RolledBack` and the `(rolled back: ...)` error text. Tests must expect them.
- **NewFixApplier returns error** — propagates backup dir creation failures.
- **LineShiftMap** — `ShiftedPosition` shifts line+column; `ShiftedRange` shifts both endpoints.
- **FlightRecorderHook** — wraps Go 1.25 `runtime/trace.FlightRecorder`; snapshots on `SlowStageThreshold` or `Snapshot(ctx, reason)`; `OnStageEvent` never errors; degraded mode when Go's singleton recorder is taken (check `Degraded()`); `Close()` waits for in-flight snapshots. Config via `FlightRecorderFileConfig` + `ResolveFlightRecorder()` (5 string-encoded fields) with full CLI parity (`flightRecorder` config section vs `-trace` flags) — keep both structs in sync; hosts holding a `FlightRecorderFileConfig` without a `ConfigFile` call the package-level `pipeline.ResolveFlightRecorderConfig(fc)` (v1.14.0). CLI flags: `-trace`, `-trace-dir`, `-trace-slow`.
- **EditListProvider (v1.14.0)** — first in the chain, handles `HasEditList()` findings: offsets direct, else line/col via the shared line index. Cross-file lists fail `ErrEditCrossFile`; stale/out-of-bounds offsets fail `ErrEditStale` (both `FixOutcomeFailed`, cause reachable via `errors.Is`). Loud failure over partial application, at every layer.
- **Analysis BeforeCode/Edits** — `analysis.FromDiagnostic` reads source files to extract `BeforeCode` from TextEdits; since v1.14.0 it fills `Finding.Edits` with ALL TextEdits of the first SuggestedFix (edits 2..N were previously dropped, issue #36) and `ToDiagnostic` re-emits the full list. SuggestedFixes 1..N are ALTERNATIVES — first wins, documented limitation.

### Testing & tooling

- **`nix run .#lint` covers the ROOT module only** — same module-boundary blind spot as `.#test`: its `golangci-lint run ./...` never reaches toolsdk/pipeline/analysis/CLI, while CI lints each module separately (2026-10-11: master's `lint (toolsdk)` job was red on 4 findings while local reported "0 issues"). Lint sub-modules directly: run `golangci-lint run ./...` inside each module dir; the devShell binary matches the ci.yml action pin (2.14.0), so local per-module runs have CI parity.
- **math/rand split** — production `math/rand/v2`; tests `math/rand` (v1, `testing/quick` constraint).
- **testify is transitive only** — `// indirect` via ginkgo/slim-sprig; not used directly (banned but unavoidable).
- **NewParallelGomega** — per-module `testutil_test.go` helper (Helper + Parallel + NewWithT in one). `paralleltest` linter disabled because it can't trace `t.Parallel()` through the helper; re-enabling means inlining or nolint-ing 100+ tests.
- **makezero `always: false` (intentional)** — `always: true` flags idiomatic `make+copy` (23 FPs); `false` still catches real `make+append` over-allocation.
- **Stress gate is mandatory** — `ginkgo -r --race --repeat=20 --skip-package=examples` (core, pipeline), `go test -race -count=20` (analysis, CLI) before any tag (release-procedure step 4). CI stress now mirrors this exact split (ginkgo CLI pinned v2.32.0, job timeout 30m): the old `go test -count=20 ./...` job failed instantly on ginkgo suites ("Only -count=1 is allowed") and stressed nothing — invisible while CI was billing-dead.
- **`pipeline/examples/`** — one runnable program per subdir + `example_compile_test.go` (`go build -o /dev/null`), mirroring root `examples/`.
- **Fuzz seeds: curated `seed_N` convention, negligible cost** — `testdata/fuzz/<FuzzName>/` commits only human-named entries (`seed_N`, `basic`, `empty`); fuzzer-generated hex-hash filenames are gitignored (that IS the promotion policy: a campaign hit worth keeping gets renamed `seed_N+1` and committed deliberately). Measured 2026-11-08: the full committed corpus (192 files across core+pipeline targets) executes in ~20 ms of test time inside the normal `go test` run — no trimming policy needed.

## CLI Features

Flag surface lives in README + FEATURES. Gotchas beyond the flag list:
markdown/CSV/TSV render via the go-output adapter (CLI module only); the
flightRecorder config section delegates to `pipeline.ResolveFlightRecorderConfig`
(one parser); `RegisterDetector` is the dynamic registry hook; the `Coverage:`
stderr line appears ONLY when a detector reports coverage evidence (DetectorV2;
v1-only runs print nothing — pinned by e2e test).

## Architecture Decisions

ADR log: `docs/architecture-decisions.md` (SARIF hand-rolled #9, go-error-family
#15, per-file rollback #16, typed edit lists #19, coverage channel #20). Facts
not recoverable from the ADR text alone: the FixProvider chain order is
EditList -> Offset -> Line -> Substring with custom providers prepended;
markdown/CSV/TSV render only in the CLI module (root `finding` stays without
go-output); naming history: `FindingTransformer` (was FindingProcessor),
`Conflict` (was ConflictInfo), `-min-severity` (was `-severity`, alias kept),
`CategoryOf` (was GetCategory). Guides: `docs/guides/fix-engine.md`,
`docs/guides/outcomes.md`, `docs/guides/detector-coverage.md`.

## Test Organization

**One test file per production file.** All tests for a subject live in `<subject>_test.go` — no `_extra`/`_bugfix`/`coverage` suffixes (split by subject, put regression tests in the parent file, never name files after metrics). Shared helpers in `testutil_test.go`; one `example_test.go` per package (don't fragment).

## Removed APIs
Deprecated APIs from v0.6.0–v0.9.0 are gone (see `docs/MIGRATION_v1.0.md`).

---

_Assisted-by: Crush <crush@charm.land>_
