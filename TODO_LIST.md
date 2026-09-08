# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

> **Harvested 2026-09-08** from the latest status report
> (`docs/status/2026-09-08_15-42_phase-d-execution-features-walk-verification.md`,
> section f) plus the Pareto master plan
> (`docs/planning/2026-09-08_04-12_pareto-master-plan-ci-billing-releases-v1.7.0.md`).
> Every completed item was verified against code and moved to CHANGELOG; only
> open or decision-gated work remains.

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

| Task                                          | Status    | Impact | Effort | Notes                                                                                          |
| --------------------------------------------- | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------- |
| D1 sign-off: keep per-file rollback default   | ⬜ `TODO` | High   | Low    | Unlocks ADR-016 + the v1.7.0 release train (version.go, 4 tags, release run). Code/docs ready. |
| Decide release strategy (backfill vs forward) | ⬜ `TODO` | High   | Low    | D6 user decision. Backfill = `gh workflow run release.yml --ref v1.6.0` (workflow_dispatch ready). |
| Re-run Release workflow per decided tags      | ⬜ `TODO` | High   | Low    | Blocked on billing + D6. Proves the cosign v3 bundle-mode fix in anger.                        |
| Verify release assets (CLI binary, notes)     | ⬜ `TODO` | Med    | Low    | After first successful run. Check `HOMEBREW_TAP_GITHUB_TOKEN` secret exists (`gh secret list`). |
| Ship v1.7.0 (GroupID, outcomes, rollback)     | ⬜ `TODO` | High   | Med    | Requires billing green + D1. Pre-release work: tests, benchmarks, docs, migration guide — all done. |
| Bump consumers to v1.7.0                      | ⬜ `TODO` | High   | Med    | go-humanize-linter + go-linter-sdk, then sweep the remaining 12 Go consumers; guide at `docs/guides/consumer-migration-v1.7.md`. |
| Comment + close issues #27/#28                | ⬜ `TODO` | Med    | Low    | D2 decision: now vs at release. Fixes verified green locally, changelogged.                    |

### Make Repo Public — Phase 2/3

| Task                                             | Status    | Impact | Effort  | Notes                                                     |
| ------------------------------------------------ | --------- | ------ | ------- | --------------------------------------------------------- |
| Verify pkg.go.dev renders after first public tag | ⬜ `TODO` | Med    | Low     | Triggered by first `go get` after visibility flip         |
| Track Go json/v2 stabilization (Go 1.27+)        | ⬜ `TODO` | Low    | Ongoing | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes |
| Verify GoReleaser + Homebrew tap on public tag   | ⬜ `TODO` | Med    | Low     | Requires the Homebrew secret (see release assets row)     |
| Write announcement (blog/r/golang/Slack/X)       | ⬜ `TODO` | High   | Medium  | Content derivable from FEATURES.md (verified 2026-09-08) |
| Submit to Awesome Go                             | ⬜ `TODO` | Low    | Low     | Discoverability                                            |

## 🟡 MEDIUM Priority

### Phase D remainder (decision-gated)

| Task                                            | Status    | Impact | Effort | Notes                                                                        |
| ----------------------------------------------- | --------- | ------ | ------ | ---------------------------------------------------------------------------- |
| D7: GroupID validation (free-form vs `^[a-z0-9-]+$`) | ⬜ `TODO` | Med  | Low    | Implement in identity/spatial validator once D7 decided (`finding_validate.go`). |
| D3: `OnFix` outcome status                      | ⬜ `TODO` | Med    | Med    | Breaking callback change vs new `OnFixOutcome` — needs design pick.          |
| D4: `ApplyDryRun` (resolve outcomes, no writes) | ⬜ `TODO` | Med    | Med    | Plus continue-on-error shape; plan/apply UX.                                 |

### Ungated Phase D + test-quality follow-ups

| Task                                                       | Status    | Impact | Effort | Notes                                                                                          |
| ---------------------------------------------------------- | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------- |
| Deterministic `GroupFindings` option + `Template.WithGroupID` | ⬜ `TODO` | Med    | Low    | Ungated half of D7 (15-42 §f.11); add determinism note/test for map iteration order (§f.26).   |
| Metrics godoc snapshot/example text                        | ⬜ `TODO` | Low    | Low    | Finish L1-30 subtask: tests + wiring done, godoc text missing (15-42 §b).                      |
| CLI e2e test: `Fix outcomes:` stderr line in a real run     | ⬜ `TODO` | Low    | Low    | 15-42 §f.16.                                                                                    |
| Fold `cancelingProvider` + `upperProvider` into `pipeline/testutil_test.go` | ⬜ `TODO` | Low | Low | L1-33 leftover (15-42 §f.14).                                                                   |
| Fuzz `FixOutcome`/`FixApplyResult` `UnmarshalJSON`         | ⬜ `TODO` | Med    | Low    | Malformed wire data (15-42 §f.27).                                                              |
| Golden wire test pinning the `FixApplyResult` doc-snippet bytes | ⬜ `TODO` | Low | Low    | Round-trip exists; pin the documented example bytes (15-42 §f.21).                              |
| Benchmark `ApplyWithOutcomes` mixed outcomes at n=1000      | ⬜ `TODO` | Med    | Low    | Confirm legacy allocation parity survived outcome typing (15-42 §f.28).                        |
| Property test: `OutcomeCounts` sum == fixable findings      | ⬜ `TODO` | Low    | Low    | 15-42 §f.29.                                                                                    |
| ADR-017/018: outcome metrics + typed outcome errors        | ⬜ `TODO` | Low    | Low    | Small ADRs keep decision history complete (15-42 §f.49).                                       |

### Docs & quality (unblocked)

| Task                                                     | Status    | Impact | Effort | Notes                                                                                   |
| -------------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------------------------------- |
| Write `docs/guides/outcomes.md` consumer guide           | ⬜ `TODO` | Med    | Med    | ApplyWithOutcomes patterns, OutcomeFor/Counts/HasErrors, rolled-back semantics (15-42 §f.20). |
| DOMAIN_LANGUAGE: outcomes/rollback pointer + Metrics term cross-check | ⬜ `TODO` | Low | Low | 15-42 §f.17.                                                                            |
| `scripts/docs-api-check.sh`: backtick identifiers in FEATURES.md exist in code | ⬜ `TODO` | Med | Med | Mechanical drift guard; would have caught most of the ~26 stale claims (15-42 §f.18, §e.6). |
| Align golangci-lint local↔CI; migrate `exhaustruct` → `exhaustruct_v5` | ⬜ `TODO` | Med | Low | Local 2.13.1 vs CI 2.10.1; deprecation warning live (15-42 §f.22).                  |
| `nix fmt` as pre-commit/CI signal for Go treefmt drift   | ⬜ `TODO` | Low    | Low    | dprint hook covers md only (15-42 §f.23).                                               |
| Consider `.golangci.yml` gofumpt/golines autofix config  | ⬜ `TODO` | Low    | Low    | End hand-fixing loops (15-42 §f.24).                                                    |
| IntervalTree go/no-go research note in ROADMAP           | ⬜ `TODO` | Low    | Low    | Quantify O(log n + k) vs O(n + k) at consumer scale (15-42 §f.25).                      |

## 🟢 LOW Priority

| Task                                       | Status       | Impact | Effort | Notes                                                                                       |
| ------------------------------------------ | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------- |
| FlightRecorder idea triage → scored table  | ⬜ `TODO`    | Low    | Low    | Score rotation/gzip/pprof/OTel/diff/AI; graduate top 1-2 into ROADMAP (L1-47).              |
| Fix BuildFlow auto-configure loop          | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact. No longer the commit gate (dprint hook replaced it). |
| Consumer compatibility test                | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code.       |

### Post-account-switch CI work (gated on billing fix)

| Task                                              | Status    | Impact | Effort | Notes                                                             |
| ------------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------------- |
| Re-run + verify ALL CI jobs on master             | ⬜ `TODO` | High   | Low    | `gh workflow run ci.yml --ref master`; first real run since July. |
| Dependabot PR triage                              | ⬜ `TODO` | Med    | Med    | Queue moved while CI was dark (checkout 4→7 etc.).                |
| CI `workflow_dispatch` inputs for module-scoped runs | ⬜ `TODO` | Low  | Low    | Faster iteration once CI is alive (15-42 §f.46).                  |
| Gate CI benchmark job on `bench-check.sh` regression | ⬜ `TODO` | Low  | Low    | Verify current informational status, then gate (15-42 §f.47).    |
| Evaluate `ginkgo --repeat=N` in CI vs sampled stress matrix | ⬜ `TODO` | Low | Low | Stress is locally mandatory now; CI duplication may be wasteful (15-42 §f.48). |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
