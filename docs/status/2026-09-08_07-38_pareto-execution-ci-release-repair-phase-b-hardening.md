# Status Report — Pareto Plan Execution: CI/Release Repair, Phase-B Hardening, Docs Truthfulness

**Date:** 2026-09-08 07:38 CEST
**Session scope:** Executing `docs/planning/2026-09-08_04-12_pareto-master-plan-ci-billing-releases-v1.7.0.md` — everything not gated on user decisions/billing.
**Format note:** The status-report skill's canonical output is styled HTML; the user explicitly requested `.md` — honored, flagged as an override, not propagated back into the skill.

---

## 0. Session arc (what was attempted)

Execute the master plan end-to-end for all items NOT gated on user decisions (D1–D8) or the billing blocker. Result: **~24 plan items executed and verified**, 2 real code bugs found and fixed by new tests, 2 stale-TODO myths debunked with evidence, and the CI/release root causes fully diagnosed. **14 commits sit LOCAL and unpushed** (auto-commit daemon did not push this time); the only deliberately pushed commit was `722b0d9` (CI/release repairs, pushed to trigger CI — which then revealed billing is still broken).

---

## a) FULLY DONE (verified)

| #  | Item                                                                                                                                                                                                                                                                                                                                                                                                                                    | Evidence                                                                                              |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| 1  | **L1-03 release.yml root cause** — NOT a 4h hang: cosign v3 (installed by cosign-installer v4) rejects GoReleaser's cosign-v2 flags with `cannot specify service URLs and use signing config`; builds/SBOMs/checksums had already succeeded at 2m22s; the 4h08m total was queue/wait tail                                                                                                                                               | `gh run view 31125816682 --log-failed`; fix verified against GoReleaser's cosign-v3 migration note    |
| 2  | **Cosign signing fixed** — `.goreleaser.yml` now uses bundle mode (`sign-blob --bundle=${signature} --yes`, `signature: "${artifact}.sigstore.json"`)                                                                                                                                                                                                                                                                                   | commit `722b0d9` (pushed)                                                                             |
| 3  | **CI workflow was `disabled_manually`, not just billing-dead** — re-enabled via `gh workflow enable`; `workflow_dispatch` added to ci.yml + release.yml so stale tags (v1.6.0) can be released without re-tagging                                                                                                                                                                                                                       | API state check before/after; commits `722b0d9`                                                       |
| 4  | **Billing confirmed STILL BROKEN** (fresh run 34182192491 failed in seconds with the payments annotation). Important nuance vs the plan: billing _worked_ on 2026-08-06 (v1.5.0 `test` job passed) and broke again after                                                                                                                                                                                                                | run annotations; plan §0.1 annotated                                                                  |
| 5  | **L1-14 benchmarks** — added `BenchmarkApplyWithOutcomes` ×8 (offset+line × 1/10/100/1000); A/B vs pre-rework commit `ad612c8~1` via git worktree                                                                                                                                                                                                                                                                                       | `/tmp/bench_old.txt` vs `/tmp/bench_opt.txt`                                                          |
| 6  | **Allocation regression found & fixed** — legacy `Apply`/`ApplyWithConflicts` paid +10.6% B/op at n=1000 (outcome slice copying a full `Finding` per input, then discarded). Fix: internal `apply(content, fixes, wantOutcomes bool)`; legacy entry points skip outcome bookkeeping entirely. Result: allocs **bit-identical** to pre-rework (6/2037), ns/op at n≥10 improved up to −40%; public `ApplyWithOutcomes` behavior unchanged | before: 3990225→4414053 B/op; after: 3990224 B/op, allocs 6 vs 6, 2037 vs 2037                        |
| 7  | **CI benchmark coverage gap closed** — the benchmark job's `./...` from root only covers the ROOT module in workspace mode, so the FixEngine hot path was never regression-checked in CI. Job now also runs pipeline benches; `benchmarks/baseline.txt` regenerated (root+pipeline, 730 lines incl. new AWO benches)                                                                                                                    | ci.yml edit; baseline grep shows both pkgs                                                            |
| 8  | **L1-09/L1-10 debunked** — the "6 pipeline + 18 CLI lint issues" debt is ALREADY RESOLVED: 0 issues across all 4 modules with the config's linters confirmed enabled (contextcheck, dupl, err113, gocognit, goconst, nestif, nilnil, varnamelen, wrapcheck)                                                                                                                                                                             | golangci-lint × 4 modules                                                                             |
| 9  | **L1-08 lint matrix** — ci.yml lint job now matrices `[".", "pipeline", "analysis", "cmd/go-finding"]` (was root-only; debt was invisible in CI)                                                                                                                                                                                                                                                                                        | ci.yml; YAML validated via yq                                                                         |
| 10 | **L1-16 golden wire tests** — byte-exact Finding JSON (`groupId` present + absent), SARIF `go-finding/groupId` property + round-trip, LSP tags wire (`1,2`, `Data.Metadata` key spelling)                                                                                                                                                                                                                                               | `json_test.go`, `sarif_roundtrip_test.go`, `lsp_test.go`                                              |
| 11 | **L1-17 CLI flag + E2E** — `-fix-rollback-all` flag wired (OR with config, matching `byteLevelConflict` precedent); binary smoke E2E + config-file E2E; config-guide flag table row; **pipeline.New wiring test** proving Config→applier policy end-to-end (default keeps fixes / AllFiles rolls back)                                                                                                                                  | `TestPipelineRun_FixRollbackAllFiles_Wiring`, `TestRun_E2E_FixRollbackAll*`                           |
| 12 | **L1-18 policy tests + REAL BUG** — cancel (both policies) and backup-failure (both policies) tests; **bug found:** the backup-failure and cancel code paths called `RollbackAll` but never recorded `report.RolledBack` and swallowed rollback errors — the report lied about what was restored. Fixed truthfully (record + wrap). Plus backup-hygiene test: no `.bak` leaks into the target tree, `Close()` removes the backup dir    | `fix_applier.go` diff; `TestFixApplier_ApplyWithReport_BackupFailure`, `TestFixApplier_BackupHygiene` |
| 13 | **L1-19 fuzz + error chains + REAL BUG** — `FuzzApplyWithOutcomes` (635K execs clean; invariants: outcomes==inputs, non-nil content, valid statuses, counts==Applied) and `FuzzParseLSPDiagnosticTags` (4.4M execs) — **fuzz found `parseLSPDiagnosticTags("-1,abc,3")` accepted negative tags** (invalid per LSP spec); fixed (`n > 0`). `errors.Is` chain test pins failed outcome → `ErrPositionUnresolvable`                        | fuzz logs; `lsp.go` fix; `TestApplyWithOutcomes_FailedOutcomeWrapsErrPositionUnresolvable`            |
| 14 | **L1-38 dep bumps** — c04af34's bumps enumerated (x/mod .38→.40, x/net .57→.58, x/text .40→.41, x/tools .48→.49, all indirect; analysis' x/tools was a deliberate earlier bump): **kept** per no-churn guardrail, documented in root CHANGELOG `Dependencies` section; missing go.sum entries repaired via `GOWORK=off go mod tidy` ×4 modules                                                                                          | CHANGELOG; tidy diff (+13 lines)                                                                      |
| 15 | **L1-07 dprint gate** — dprint 0.56.1 in devShell; tracked hook `scripts/hooks/pre-commit` (`dprint check --staged`, actionable failure message); devShell shellHook auto-installs it (cmp-guarded); the BuildFlow hook (the reason `--no-verify` existed for 109 days) is replaced; 11 files formatted across two fmt passes                                                                                                           | flake.nix; `dprint check` clean                                                                       |
| 16 | **L1-12 nix flake check GREEN** — fixed: treefmt drift (3 files, incl. daemon-mangled `fix_outcome_test.go` imports), stale `vendorHash` (dep bumps), go.sum holes. Added `nix flake check` to the release-procedure checklist                                                                                                                                                                                                          | `nix flake check` → "all checks passed"                                                               |
| 17 | **L1-13 stress tests** — `ginkgo -r --race --repeat=20`: pipeline (3 suites) + core (9 suites) all passed                                                                                                                                                                                                                                                                                                                               | `/tmp/stress_*.txt`                                                                                   |
| 18 | **L1-23 API_STABILITY.md** — audited 2026-09-08; v1.5.0 symbols (ParseConfidence, alias registry, ResolveSafePath family), v1.7.0 unreleased (GroupID, outcomes, rollback), deterministic-output guarantee, per-version status section                                                                                                                                                                                                  | doc diff                                                                                              |
| 19 | **L1-24 CONTRIBUTING tree** — diffed against actual files; exactly one missing (`fix_outcome.go`), added                                                                                                                                                                                                                                                                                                                                | doc diff                                                                                              |
| 20 | **L1-26 DOMAIN_LANGUAGE** — Fix Outcome (6 statuses), Rollback Policy, Shift Map, Group, Clone Group; cross-linked from `fix_outcome.go`/`fix_applier.go` godoc                                                                                                                                                                                                                                                                         | doc + code comments                                                                                   |
| 21 | **L1-27 overviews** — doc.go bullet list, README features, USAGE_GUIDE sections (outcomes, rollback policy, groups)                                                                                                                                                                                                                                                                                                                     | 3 files                                                                                               |
| 22 | **L1-28 migration guide** — `docs/guides/consumer-migration-v1.7.md` (rollback behavior change, outcomes migration, groups, v1.5.0 adoption), linked from README docs table                                                                                                                                                                                                                                                             | new file                                                                                              |
| 23 | **L1-29 housekeeping + HARVEST** — 3 resolved reports `git mv`'d to `docs/status/archived/`; docs-freshness.sh now excludes point-in-time assessment docs (they were permanent false positives: 5 warnings → 1 transient); **TODO_LIST fully harvested**: 24 DONE rows with evidence, survivors re-scoped behind real gates                                                                                                             | TODO_LIST diff; freshness script                                                                      |
| 24 | **Full verification battery** — race tests ×4 modules (exit-code-verified, 0 FAIL lines — no pipe masking), lint 0 issues ×4, all 6 structural scripts green, `go-arch-lint check` OK, `dprint check` clean, `nix flake check` green                                                                                                                                                                                                    | job logs; scripts output                                                                              |

---

## b) PARTIALLY DONE

| Item                                      | State                                                                                                                    | What remains                                                                                                     |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| **L1-25 FEATURES walk**                   | Spot-verified: all outcome/group/rollback sections match code signatures                                                 | Full per-signature walk of all 16 sections (~100 min) — kept as TODO, honestly labeled                           |
| **L1-02 CI verification**                 | Workflow re-enabled, dispatch added, local equivalents all green                                                         | Actual GitHub-side green run — frozen on billing                                                                 |
| **L1-05/06/11 release repair + backfill** | Mechanical fixes shipped (cosign, dispatch, secret check pending)                                                        | Re-run release.yml on decided tags; asset verification; secret check `gh secret list`                            |
| **L1-21 v1.7.0 release train**            | ALL pre-release work done this session: golden tests, E2E, policy tests, fuzz, benchmarks+baseline, docs, CHANGELOGs     | D1 sign-off + billing; then version bump, tag ×4, push, watch                                                    |
| **Phase D deferred features**             | Harvested to TODO_LIST with scope (metrics outcomes, outcome JSON, GroupID validation D7, OnFix D3, DryRun D4, examples) | Implementation — deliberately deferred (features after benchmarks guardrail satisfied, but gated on D-decisions) |
| **Stress gate decision**                  | repeat=20 race run now proven green (evidence exists)                                                                    | Decide mandatory-gate vs drop from release procedure                                                             |
| **Master plan annotation**                | §0.1 execution findings added (L1-03 complete, billing still broken)                                                     | Final ANNOTATE pass when plan fully executed                                                                     |

## c) NOT STARTED (gated — by design, not neglect)

| Item                                                                        | Gate                                                                                                                               |
| --------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| L1-01 billing fix                                                           | **User-only action** (verified still failing 2026-09-08 ~05:15 CEST)                                                               |
| L1-04/D6 release strategy                                                   | User decision (recommendation ready: backfill via workflow_dispatch — releases are cosmetic to the Go proxy, tags already resolve) |
| D1 rollback default                                                         | User decision (per-file default shipped in code+docs; consumer comms depend on it)                                                 |
| L1-20 issues #27/#28 comment/close                                          | D2                                                                                                                                 |
| L1-22 consumer bumps                                                        | v1.7.0 release                                                                                                                     |
| L1-34 OnFix outcome                                                         | D3 (breaking vs new callback)                                                                                                      |
| L1-39 ApplyDryRun                                                           | D4                                                                                                                                 |
| L1-32 GroupID validation                                                    | D7                                                                                                                                 |
| L1-49 GAP-7/GAP-3                                                           | D5/D8                                                                                                                              |
| L1-35 art-dupl spike, L1-36 go-cqrs-lint adoption, L1-37 SARIF/LSP research | Post-release                                                                                                                       |
| L1-42..46 design spikes, L1-47 FlightRecorder triage, L1-48 Make-Public     | ROADMAP/post-release                                                                                                               |
| Dependabot PR triage                                                        | CI green (nothing should merge while gates are dark)                                                                               |

## d) TOTALLY FUCKED UP (own messes — all caught and cleaned)

1. **My fuzz harness panicked** (negative fuzz input → `content[offset:...]` slice bounds) — the engine was fine, my generator didn't normalize sign. Fixed; then fuzzed 635K execs clean. Lesson: normalize fuzz inputs FIRST.
2. **My error-chain test premise was wrong** — asserted a line-1000 fix would fail; the substring fall-through provider rescued it (by design). Rewrote the scenario (BeforeCode not present in content) so all providers legitimately fail.
3. **My own new tests introduced 4 lint issues** (golines >120 col, revive unused `t` ×2, tparallel) — caught by the pre-ship lint run, fixed, but should never have been written dirty.
4. **Bench-merge job ran from the wrong cwd** (`cd pipeline` earlier in the compound command) — baseline briefly stayed root-only; redone from repo root.
5. **Bash printf `%f` doesn't exist in mvdan/sh** — burned 3 round trips on benchmark median math before switching to awk.
6. **Two `edit` attempts failed on not-read / modified-file guards** and one Python heredoc failed on escape handling — pure tool-hygiene round trips.

Nothing in the PROJECT is fucked up: working tree is clean, all suites green, no regressions introduced. The remaining genuinely-broken thing is **GitHub billing (user-owned)**.

## e) WHAT WE SHOULD IMPROVE

1. **Check the actual blocker first.** The plan assumed billing died 2026-07-16 and stayed dead; reality: workflow was ALSO manually disabled, and billing had a working window on 2026-08-06. Two `gh api` calls at session start would have re-scoped Phase 0 immediately.
2. **Read failed-run logs before theorizing.** The "4h hang" hypothesis in the plan dissolved in minutes once `--log-failed` was read — the failure was a 3-minute cosign flag mismatch. Diagnosis before hypothesis.
3. **Audit gate COVERAGE, not gate existence.** The CI benchmark job existed and passed — but never measured the pipeline hot path. Same risk class as "CI lints root only". Every gate should carry a one-line "what does this actually measure" check.
4. **Write fuzz harnesses defensively** (normalize sign/range first, then derive) — my first corpus entry panicked in harness code, not engine code.
5. **Lint new test files immediately**, not at session end.
6. **Commit batches sooner.** 14 commits piled up local-only; the daemon didn't push and I hadn't been asked to. Work-at-risk assessment (unpushed = unverified on remote) should be part of the loop.
7. **`GOWORK=off go mod tidy` belongs in the dep-change workflow** — the daemon's bumps left go.sum holes that only isolated builds (flake/CI per-module) would catch; that's exactly how `nix flake check` broke next.
8. **The plan's cost model was directionally right but distribution-wrong** — more than half the effort went to root-causing Phase-0 infrastructure rather than the scheduled tasks; front-load a "state re-verification" step in every plan.
9. **CHANGELOG lag:** two behavior-relevant fixes from this session (RolledBack truthfulness, negative-tag parsing) are not yet in the CHANGELOG Unreleased sections — flagged as the first thing to do next session.

## f) Up to 50 things to get done next (brainstorm; harvest with rigor)

**Gate (blocking everything else):**

1. Fix GitHub Actions billing/spending limit (USER action).
2. `gh workflow run ci.yml --ref master` → verify ALL jobs green (incl. new lint matrix + extended benchmark job).
3. Triage whatever the fresh CI run exposes (first real run since July on current HEAD).

**Release train (v1.6.x backfill → v1.7.0):**
4. D6: decide backfill vs forward-only (recommendation: backfill via workflow_dispatch).
5. `gh workflow run release.yml --ref v1.6.0` → proves the cosign fix in anger.
6. `gh secret list` → confirm `HOMEBREW_TAP_GITHUB_TOKEN` exists.
7. Verify release assets (CLI binary, SBOMs, `.sigstore.json` bundles) on the new release.
8. D1: sign off per-file rollback default → write ADR-016.
9. Stamp CHANGELOGs `1.7.0` + date; bump `version.go`; `scripts/version-check.sh`.
10. Add this session's two fixes to CHANGELOG Unreleased: RolledBack truthfulness fix; negative LSP tag parsing fix.
11. Tag `v1.7.0` + 3 sub-module tags; push; watch Release workflow per tag.
12. `go get github.com/larsartmann/go-finding@v1.7.0` proxy smoke test.
13. D2: comment + close issues #27/#28 with fix summaries.
14. Post to 22 consumers: migration note (rollback default!) linking `docs/guides/consumer-migration-v1.7.md`.
15. Bump go-humanize-linter to v1.7.0; run its suite.
16. Bump go-linter-sdk to v1.7.0; run its suite.
17. Sweep remaining 12 Go consumers via the audit doc; note stragglers.
18. Post-release verification closure: CI green on all 4 new tags, releases exist, proxy resolution — record in TODO_LIST.
19. Push the 14 local commits (or confirm daemon policy) — remote is stale relative to local.
20. Dependabot triage: the PR queue kept moving while CI was dead (checkout 4→7 etc.); review/merge/rebase after CI is green.

**Phase D features (post-D-decisions):**
21. `Metrics.RecordOutcome(status)` + pipeline wiring + CLI fix summary print.
22. Deterministic JSON marshaling for `FixOutcome`/`FixApplyResult` (marshalOpts rule + json-deterministic script).
23. `FindingError`-typed outcome failures (errorfamily classification).
24. D7: GroupID validation decision → implement in identity/spatial validator.
25. `GroupFindings` deterministic-order option / `GroupFindingsSorted`.
26. `Template.WithGroupID` + test.
27. D3: `OnFix` outcome status (breaking change vs new `OnFixOutcome`).
28. D4: continue-on-hard-error option folded into `ApplyDryRun`.
29. Applier-level `DryRun` (resolve outcomes read-only, plan/apply UX).
30. `examples/outcomes` demo program (ApplyWithOutcomes + RollbackPolicy).
31. Dedupe erroring/static/saboteur provider stubs into `pipeline/testutil_test.go`.
32. Provider-precedence test: provider A refuses → provider B applies.
33. Shift-map test: file with applied+refused findings maps correctly.
34. `RolledBack`-in-error-text test (error message contains restored files).
35. `LSPDiagnosticData` full wire golden (all fields, not just tags).

**Deeper hardening / research:**
36. Consider tightening bench-check threshold for FixEngine hot path (25% may hide a 15% regression).
37. Evaluate `FixOutcome` memory shape (store index vs full Finding copy) — AWO at n=1000 is 4.4MB/op; v2 candidate, needs API tradeoff analysis.
38. art-dupl spike: `cloneGroupToFindings` adapter + SARIF round-trip of groups.
39. go-cqrs-lint: adopt `ApplyWithReport`/`FailedOutcomes`; re-verify the original #28 scenario end-to-end in that repo.
40. SARIF clone-group representation research (codeFlow/threadFlow vs properties).
41. LSP client grouping recipe doc (Data.GroupID) into docs/guides.
42. Full FEATURES.md per-signature walk (all 16 sections).
43. Stress-gate decision: mandatory in release procedure or remove (evidence now exists both ways).
44. Extend pre-commit gate: gofumpt/golangci-lint fmt on staged `.go` files (dprint covers md/json/yaml only).
45. Hook bootstrap for fresh clones (documented in CONTRIBUTING; currently only devShell installs it).
46. ROADMAP spikes L1-42..46 (Position sentinel, FixStrategy union, pointer-state, TagSet, sub-struct).
47. FlightRecorder triage: score rotation/gzip/pprof/OTel/diff/AI; graduate top 1–2.
48. GAP-7 final answer (D5) + GAP-3 revisit trigger (D8); annotate feedback doc.
49. Make-Public: pkg.go.dev render check on v1.7.0 (all modules); announcement draft; Awesome Go PR.
50. Watch json/v2 stabilization (Go 1.27) — remove `GOEXPERIMENT=jsonv2` when stable.

## g) Questions I can NOT figure out myself

1. **Billing:** Have you fixed the GitHub Actions billing/spending limit yet (or should I keep everything queued behind it)? I verified it still fails as of this morning; only the account owner can settle it.
2. **D1 (rollback default):** Do you confirm the per-file rollback default for v1.7.0? Code, docs, tests, and the migration guide all ship it — but flipping it later would churn consumers twice, so the release train is frozen on your sign-off.
3. **D6 (releases):** Backfill v1.5.0/v1.6.0 GitHub Releases now (workflow_dispatch is ready; cosign fix needs one real run to prove itself), or go forward-only with v1.7.0?

---

_Point-in-time snapshot. Next session: docs-health ANNOTATE this report once its items move, HARVEST section (f) survivors into TODO_LIST/ROADMAP._
