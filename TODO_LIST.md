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

### Unblock CI (THE gate — everything frozen behind this)

GitHub Actions billing/spending limit must be fixed by the account owner. The
CI workflow was also found `disabled_manually` and has been re-enabled
(2026-09-08), and the release pipeline's cosign v3 signing failure was fixed
(commit `722b0d9`). Nothing further can be verified remotely until billing is
settled: a fresh run (34182192491) failed in seconds with "recent account
payments have failed or your spending limit needs to be increased".

| Task                                      | Status       | Impact | Effort | Notes                                                                                       |
| ----------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------- |
| Fix GitHub Actions billing/spending limit | 🔵 `BLOCKED` | High   | —      | **User action.** GitHub → Settings → Billing & plans. Then re-run CI via workflow_dispatch. |
| Verify all CI jobs green on master        | ⬜ `TODO`    | High   | Low    | `gh workflow run ci.yml --ref master` after billing fix; all jobs incl. stress, govulncheck |

### Post-Release: v1.6.0 → v1.7.0 Verification & Distribution

v1.6.0 tags exist (all 4 modules) but produced **no release run** — the
workflows were disabled at push time. v1.5.0's core release failed at the
cosign signing step (root cause fixed); `analysis/v1.5.0` succeeded. Strategy
decision needed: backfill v1.5.0/v1.6.0 releases vs forward-only v1.7.0.

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

| Task                                     | Status       | Impact | Effort | Notes                                                                                                                                                                                                                     |
| ---------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Full FEATURES.md vs code walk            | ⬜ `TODO`    | Med    | High   | Recurring gap since v1.3.0. Spot-verified 2026-09-08 (outcome/group sections match code); full per-signature walk still open.                                                                                             |
| Stress tests as mandatory release gate   | ⬜ `TODO`    | Low    | Low    | repeat=20 race run passed 2026-09-08 (core + pipeline). Decide: hard gate in release procedure, or drop as formal step.                                                                                                   |
| Fix BuildFlow auto-configure loop        | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact. No longer the commit gate (replaced by dprint hook, see ✅ below).               |
| Close v1.7.0 follow-ups from master plan | ⬜ `TODO`    | Med    | Med    | Deferred features: Metrics.RecordOutcome + CLI fix summary, outcome JSON deterministic marshaling, GroupID validation decision (D7), OnFix outcome status (D3), ApplyDryRun (D4), examples/outcomes. See plan §2 Phase D. |

## ✅ DONE (verified this cycle, 2026-09-08)

| Task                                       | Evidence                                                                                                                         |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| Re-enable CI workflow                      | Was `disabled_manually`; `gh workflow enable ci.yml` (commit `722b0d9`)                                                          |
| Fix release signing (cosign v3)            | `.goreleaser.yml` bundle-mode config; root cause of v1.5.0 4h failure                                                            |
| Add workflow_dispatch to ci/release        | Enables manual runs without re-tagging                                                                                           |
| Fix 6+18 pipeline/CLI lint issues          | Already resolved upstream; verified 0 issues × 4 modules with golangci-lint (config linters confirmed enabled)                   |
| CI lint matrix (4 modules)                 | `ci.yml` lint job now matrices over `.`, `pipeline`, `analysis`, `cmd/go-finding`                                                |
| dprint commit gate restored                | `pkgs.dprint` in devShell (0.56.1); tracked hook `scripts/hooks/pre-commit`; installed via shellHook; BuildFlow hook replaced    |
| Benchmark regression check + new benches   | `BenchmarkApplyWithOutcomes` ×8; legacy-path allocation regression eliminated (`apply(wantOutcomes)`); B/op parity vs pre-rework |
| Benchmark baseline covers pipeline         | `benchmarks/baseline.txt` now includes pipeline module; CI benchmark job extended to match                                       |
| Golden wire tests                          | `groupId` JSON byte-exact, SARIF `go-finding/groupId`, LSP tags metadata key (finding/json/lsp/sarif test files)                 |
| `-fix-rollback-all` CLI flag + E2E         | Flag + config-file E2E tests; precedence: flag or config enables (OR)                                                            |
| pipeline.New rollback-policy wiring test   | `TestPipelineRun_FixRollbackAllFiles_Wiring` (default keeps fixes; AllFiles rolls back)                                          |
| Cancel/backup-failure policy tests         | ApplyWithReport under both policies; backup hygiene (no .bak leaks, Close cleans up)                                             |
| RolledBack reporting bug fix               | Backup-failure and cancel paths now record `report.RolledBack` truthfully                                                        |
| Fuzz + error chains                        | `FuzzApplyWithOutcomes` (635K execs), `FuzzParseLSPDiagnosticTags` (4.4M execs) — found+fixed negative-tag parsing               |
| `errors.Is` chain assertion                | Failed outcome wraps `ErrPositionUnresolvable` (test pinned)                                                                     |
| Indirect dep bumps confirmed               | x/mod .38→.40, x/net .57→.58, x/text .40→.41, x/tools .48→.49 kept; documented in CHANGELOG; go.sum holes fixed                  |
| `nix flake check` green                    | treefmt drift fixed (3 files), vendorHash updated, all checks passed                                                             |
| Stress tests (repeat=20, race)             | Core + pipeline suites passed (2026-09-08)                                                                                       |
| API_STABILITY.md v1.5.0+/v1.7.0 symbols    | ParseConfidence, ResolveSafePath, alias registry, outcomes/rollback APIs, deterministic-output guarantee                         |
| CONTRIBUTING.md project tree               | Verified against `git ls-files`; added `fix_outcome.go` (only missing file)                                                      |
| Consumer migration guide                   | `docs/guides/consumer-migration-v1.7.md`, linked from README                                                                     |
| doc.go/README/USAGE_GUIDE overview updates | GroupID + outcomes + rollback in all three; DOMAIN_LANGUAGE cross-linked from godoc                                              |
| DOMAIN_LANGUAGE.md new terms               | Fix Outcome (6 statuses), Rollback Policy, Shift Map, Group, Clone Group                                                         |
| Archive 3 resolved status reports          | `git mv` to `docs/status/archived/` (2026-09-07 status, 2× 2026-08-08 pareto execution)                                          |
| docs-freshness warnings resolved           | Point-in-time assessment docs excluded from scan (were permanent false positives)                                                |
| Run `go-arch-lint check` locally           | Passed ("OK - No warnings found")                                                                                                |

## 🟢 LOW Priority

| Task                        | Status       | Impact | Effort | Notes                                                                                |
| --------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| Consumer compatibility test | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
