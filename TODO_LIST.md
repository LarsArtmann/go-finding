# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

> **Harvested 2026-09-08** from the Pareto master plan
> (`docs/planning/2026-09-08_04-12_pareto-master-plan-ci-billing-releases-v1.7.0.md`)
> with per-item verification. Completed plan items were moved to this file's
> ✅ rows or removed; survivors are listed below with evidence.

---

## 🔴 HIGH Priority

### Post-Release: v1.6.0 → v1.7.0 Verification & Distribution

v1.6.0 tags exist (all 4 modules) but produced **no release run** — the
workflows were disabled at push time. v1.5.0's core release failed at the
cosign signing step (root cause fixed); `analysis/v1.5.0` succeeded. Strategy
decision needed: backfill v1.5.0/v1.6.0 releases vs forward-only v1.7.0.

> **Billing/CI (2026-09-08):** the user decided to **ignore the GitHub Actions
> billing failure** — the account is being switched soon. CI stays dark until
> then; local verification (race ×4, lint ×4, structural scripts, arch-lint,
> nix flake check) is the active quality gate. After the switch: re-run CI via
> `gh workflow run ci.yml --ref master` and verify all jobs.

| Task                                          | Status    | Impact | Effort | Notes                                                                                                                     |
| --------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| Decide release strategy (backfill vs forward) | ⬜ `TODO` | High   | Low    | D6 user decision. Backfill = `gh workflow run release.yml --ref v1.6.0` (workflow_dispatch ready).                        |
| Re-run Release workflow per decided tags      | ⬜ `TODO` | High   | Low    | Blocked on billing + D6.                                                                                                  |
| Verify release assets (CLI binary, notes)     | ⬜ `TODO` | Med    | Low    | After first successful run. Check `HOMEBREW_TAP_GITHUB_TOKEN` secret exists (`gh secret list`).                           |
| Ship v1.7.0 (GroupID, outcomes, rollback)     | ⬜ `TODO` | High   | Med    | Requires billing green + D1 (rollback default sign-off). Pre-release work: tests, benchmarks, docs all done this session. |
| Bump consumers to v1.7.0                      | ⬜ `TODO` | High   | Med    | go-humanize-linter + go-linter-sdk; migration guide ready at `docs/guides/consumer-migration-v1.7.md`.                    |
| Comment + close issues #27/#28                | ⬜ `TODO` | Med    | Low    | D2 decision: now vs at release. Fixes verified green locally.                                                             |

### Make Repo Public — Phase 2/3 (unchanged)

| Task                                             | Status    | Impact | Effort  | Notes                                                     |
| ------------------------------------------------ | --------- | ------ | ------- | --------------------------------------------------------- |
| Verify pkg.go.dev renders after first public tag | ⬜ `TODO` | Med    | Low     | Triggered by first `go get` after visibility flip         |
| Track Go json/v2 stabilization (Go 1.27+)        | ⬜ `TODO` | Low    | Ongoing | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes |
| Verify GoReleaser + Homebrew tap on public tag   | ⬜ `TODO` | Med    | Low     | Requires the Homebrew secret (see release assets row)     |
| Write announcement (blog/r/golang/Slack/Twitter) | ⬜ `TODO` | High   | Medium  | Drive adoption                                            |
| Submit to Awesome Go                             | ⬜ `TODO` | Low    | Low     | Discoverability                                           |

## 🟡 MEDIUM Priority

| Task                                   | Status       | Impact | Effort | Notes                                                                                                                                                                                                                                                                                       |
| -------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Full FEATURES.md vs code walk          | ✅ `DONE`    | Med    | High   | Completed 2026-09-08: 4 parallel verification passes over all 22 sections; ~26 discrepancies found and fixed (signatures, defaults, complexity claims, CLI formats). See git history + DONE table.                                                                                          |
| Stress tests as mandatory release gate | ✅ `DONE`    | Low    | Low    | Decided 2026-09-08: step 4 of the release procedure is a **mandatory** gate (repeat=20 race passed 2026-09-08). Recorded in `docs/release-procedure.md`.                                                                                                                                    |
| Fix BuildFlow auto-configure loop      | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact. No longer the commit gate (replaced by dprint hook, see ✅ below).                                                                                 |
| Ship unblocked Phase D follow-ups      | ✅ `DONE`    | Med    | Med    | Metrics.RecordOutcome + CLI fix summary, deterministic FixOutcome/FixApplyResult JSON + FindingError-typed outcome errors, rolled-back paths in error text, examples/outcomes, SARIF/LSP grouping guide. Remaining follow-ups (GroupID validation D7, OnFix D3, ApplyDryRun D4) stay gated. |

## ✅ DONE (verified this cycle, 2026-09-08)

| Task                                       | Evidence                                                                                                                                                                                                                                                                                                                      |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Re-enable CI workflow                      | Was `disabled_manually`; `gh workflow enable ci.yml` (commit `722b0d9`)                                                                                                                                                                                                                                                       |
| Fix release signing (cosign v3)            | `.goreleaser.yml` bundle-mode config; root cause of v1.5.0 4h failure                                                                                                                                                                                                                                                         |
| Add workflow_dispatch to ci/release        | Enables manual runs without re-tagging                                                                                                                                                                                                                                                                                        |
| Fix 6+18 pipeline/CLI lint issues          | Already resolved upstream; verified 0 issues × 4 modules with golangci-lint (config linters confirmed enabled)                                                                                                                                                                                                                |
| CI lint matrix (4 modules)                 | `ci.yml` lint job now matrices over `.`, `pipeline`, `analysis`, `cmd/go-finding`                                                                                                                                                                                                                                             |
| dprint commit gate restored                | `pkgs.dprint` in devShell (0.56.1); tracked hook `scripts/hooks/pre-commit`; installed via shellHook; BuildFlow hook replaced                                                                                                                                                                                                 |
| Benchmark regression check + new benches   | `BenchmarkApplyWithOutcomes` ×8; legacy-path allocation regression eliminated (`apply(wantOutcomes)`); B/op parity vs pre-rework                                                                                                                                                                                              |
| Benchmark baseline covers pipeline         | `benchmarks/baseline.txt` now includes pipeline module; CI benchmark job extended to match                                                                                                                                                                                                                                    |
| Golden wire tests                          | `groupId` JSON byte-exact, SARIF `go-finding/groupId`, LSP tags metadata key (finding/json/lsp/sarif test files)                                                                                                                                                                                                              |
| `-fix-rollback-all` CLI flag + E2E         | Flag + config-file E2E tests; precedence: flag or config enables (OR)                                                                                                                                                                                                                                                         |
| pipeline.New rollback-policy wiring test   | `TestPipelineRun_FixRollbackAllFiles_Wiring` (default keeps fixes; AllFiles rolls back)                                                                                                                                                                                                                                       |
| Cancel/backup-failure policy tests         | ApplyWithReport under both policies; backup hygiene (no .bak leaks, Close cleans up)                                                                                                                                                                                                                                          |
| RolledBack reporting bug fix               | Backup-failure and cancel paths now record `report.RolledBack` truthfully                                                                                                                                                                                                                                                     |
| Fuzz + error chains                        | `FuzzApplyWithOutcomes` (635K execs), `FuzzParseLSPDiagnosticTags` (4.4M execs) — found+fixed negative-tag parsing                                                                                                                                                                                                            |
| `errors.Is` chain assertion                | Failed outcome wraps `ErrPositionUnresolvable` (test pinned)                                                                                                                                                                                                                                                                  |
| Indirect dep bumps confirmed               | x/mod .38→.40, x/net .57→.58, x/text .40→.41, x/tools .48→.49 kept; documented in CHANGELOG; go.sum holes fixed                                                                                                                                                                                                               |
| `nix flake check` green                    | treefmt drift fixed (3 files), vendorHash updated, all checks passed                                                                                                                                                                                                                                                          |
| Stress tests (repeat=20, race)             | Core + pipeline suites passed (2026-09-08)                                                                                                                                                                                                                                                                                    |
| API_STABILITY.md v1.5.0+/v1.7.0 symbols    | ParseConfidence, ResolveSafePath, alias registry, outcomes/rollback APIs, deterministic-output guarantee                                                                                                                                                                                                                      |
| CONTRIBUTING.md project tree               | Verified against `git ls-files`; added `fix_outcome.go` (only missing file)                                                                                                                                                                                                                                                   |
| Consumer migration guide                   | `docs/guides/consumer-migration-v1.7.md`, linked from README                                                                                                                                                                                                                                                                  |
| doc.go/README/USAGE_GUIDE overview updates | GroupID + outcomes + rollback in all three; DOMAIN_LANGUAGE cross-linked from godoc                                                                                                                                                                                                                                           |
| DOMAIN_LANGUAGE.md new terms               | Fix Outcome (6 statuses), Rollback Policy, Shift Map, Group, Clone Group                                                                                                                                                                                                                                                      |
| Archive 3 resolved status reports          | `git mv` to `docs/status/archived/` (2026-09-07 status, 2× 2026-08-08 pareto execution)                                                                                                                                                                                                                                       |
| docs-freshness warnings resolved           | Point-in-time assessment docs excluded from scan (were permanent false positives)                                                                                                                                                                                                                                             |
| Run `go-arch-lint check` locally           | Passed ("OK - No warnings found")                                                                                                                                                                                                                                                                                             |
| Phase D: outcome metrics + JSON + tests    | `Metrics.RecordOutcome`/`OutcomeCounts` + CLI `Fix outcomes:` summary; deterministic `FixOutcome`/`FixApplyResult` JSON; `FindingError`-typed failed outcomes; rolled-back paths in error text; provider-precedence/shift-map/LSP-wire golden tests; `pipeline/examples/outcomes`; test stubs deduped into `testutil_test.go` |
| FEATURES.md full per-signature walk        | 4 parallel verification passes, all 22 sections; ~26 stale claims fixed (Stable() method, WriteSARIF ctx, provider conditions, 6 CLI formats, O(n+k) complexity, etc.)                                                                                                                                                        |
| SARIF/LSP grouping guide                   | `docs/guides/finding-groups.md` (spec-verified: correlationGuid/relatedLocations/codeFlows analysis + LSP recipe); linked from USAGE_GUIDE                                                                                                                                                                                    |
| Stress-gate decision recorded              | `docs/release-procedure.md` step 4 marked mandatory (2026-09-08)                                                                                                                                                                                                                                                              |
| Billing decision recorded                  | User: ignore billing failure, switching accounts; CI re-verify deferred to post-switch                                                                                                                                                                                                                                        |

## 🟢 LOW Priority

| Task                        | Status       | Impact | Effort | Notes                                                                                |
| --------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| Consumer compatibility test | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
