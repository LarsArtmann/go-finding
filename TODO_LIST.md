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

| Task                                          | Status       | Impact | Effort | Notes                                                                                                                                                                                                                                                                                     |
| --------------------------------------------- | ------------ | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D1 sign-off: keep per-file rollback default   | ✅ `DONE`    | High   | Low    | **Decided 2026-09-08: SHIP per-file rollback default.** ADR-016 written (`docs/architecture-decisions.md` #16). Unlocked the release train.                                                                                                                                               |
| Decide release strategy (backfill vs forward) | ✅ `DONE`    | High   | Low    | **Decided 2026-09-08 (D6): FORWARD-ONLY v1.7.0.** No v1.5/v1.6 backfill; revisit post-account-switch if desired. Push at tag time.                                                                                                                                                        |
| Re-run Release workflow per decided tags      | 🔵 `BLOCKED` | High   | Low    | Attempted 2026-09-08: dispatch run 34247887668 failed instantly (billing; logs unavailable). CI push run also all-red. Re-run after account switch.                                                                                                                                       |
| Verify release assets (CLI binary, notes)     | 🔵 `BLOCKED` | Med    | Low    | **Finding 2026-09-08: `HOMEBREW_TAP_GITHUB_TOKEN` secret does NOT exist** (`gh secret list` empty). Create it before the post-switch release run.                                                                                                                                         |
| Ship v1.7.0 (GroupID, outcomes, rollback)     | ✅ `DONE`    | High   | Med    | **Shipped 2026-09-08.** All gates green (bench vs regenerated baseline, GOWORK=off x4, stress repeat=20 race x4, flake check, dprint, docs-freshness). 4 tags pushed; proxy smoke resolves all 4 modules @v1.7.0. Bench gate awk bug fixed (was a silent no-op); flake vendorHash synced. |
| Bump consumers to v1.7.0                      | ✅ `DONE`    | High   | Med    | **Leads done 2026-09-08:** go-humanize-linter @v1.7.0 (suite green; 1 pre-existing unrelated failure), go-linter-sdk @v1.7.0 tagged v0.3.0, both pushed. 22-repo sweep table + migration note in `docs/ecosystem.md`; 8 pipeline consumers flagged for rollback migration.                |
| Comment + close issues #27/#28                | ✅ `DONE`    | Med    | Low    | **Decided 2026-09-08 (D2): close now** with fix-summary comments + `bug`/`fixed-in-v1.7.0` labels.                                                                                                                                                                                        |

### Make Repo Public — Phase 2/3

| Task                                             | Status       | Impact | Effort  | Notes                                                                                                                                                          |
| ------------------------------------------------ | ------------ | ------ | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Verify pkg.go.dev renders after first public tag | 🔵 `BLOCKED` | Med    | Low     | Repo is PRIVATE; checklist in `docs/brainstorming/launch-announcement-draft.md`.                                                                               |
| Track Go json/v2 stabilization (Go 1.27+)        | ⬜ `TODO`    | Low    | Ongoing | Watch section in ROADMAP (`json/v2 stabilization watch`); sweep GOEXPERIMENT when Go 1.27 ships it.                                                            |
| Verify GoReleaser + Homebrew tap on public tag   | 🔵 `BLOCKED` | Med    | Low     | Repo PRIVATE + `HOMEBREW_TAP_GITHUB_TOKEN` secret missing.                                                                                                     |
| Write announcement (blog/r/golang/Slack/X)       | ✅ `DONE`    | High   | Medium  | **Drafted 2026-09-08** (`docs/brainstorming/launch-announcement-draft.md`): short + blog-length versions, Awesome Go entry, launch checklist. Publish at flip. |
| Submit to Awesome Go                             | 🔵 `BLOCKED` | Low    | Low     | Repo PRIVATE; submission entry pre-drafted in the announcement doc.                                                                                            |

## 🟡 MEDIUM Priority

### Phase D remainder (decision-gated)

| Task                                                 | Status    | Impact | Effort | Notes                                                                                                                                                                                                                                        |
| ---------------------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D7: GroupID validation (free-form vs `^[a-z0-9-]+$`) | ✅ `DONE` | Med    | Low    | **Decided+shipped 2026-09-08:** permissive machine-identifier format (no whitespace/control chars, <=128B; `branded_types.go:GroupID.IsValid`, `finding_validate.go:validateGroupID`). Strict charset rejected - would break real ID styles. |
| D3: `OnFix` outcome status                           | ✅ `DONE` | Med    | Med    | **Decided+shipped 2026-09-08:** additive `Config.OnFixOutcome` (`pipeline/config.go`) fired from the same outcomes loop as metrics; boolean `OnFix` deprecated, retained.                                                                    |
| D4: `ApplyDryRun` (resolve outcomes, no writes)      | ✅ `DONE` | Med    | Med    | **Decided+shipped 2026-09-08:** `FixApplier.ApplyDryRun` (`pipeline/fix_applier.go`) - full ApplyReport shape, zero writes/backups; soft failures surface as in real runs.                                                                   |

### Ungated Phase D + test-quality follow-ups

| Task                                                                        | Status    | Impact | Effort | Notes                                                                                                                                                                                                         |
| --------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Deterministic `GroupFindings` option + `Template.WithGroupID`               | ✅ `DONE` | Med    | Low    | **Done 2026-09-08:** `Report.GroupFindingsSorted()` + `Group` (`report_query.go`), `Template.WithGroupID` (`finding_builder.go`); map-order note in godoc.                                                    |
| Metrics godoc snapshot/example text                                         | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** usage-oriented godoc + `ExampleMetrics` (`pipeline/metrics.go`, `pipeline/example_test.go`).                                                                                             |
| CLI e2e test: `Fix outcomes:` stderr line in a real run                     | ✅ `DONE` | Low    | Low    | **Done 2026-09-08 (as achievable):** formatter already table-tested; e2e pins the conditional (`cmd/go-finding/e2e_test.go`) - full presence-e2e needs a fix-emitting built-in detector (documented in test). |
| Fold `cancelingProvider` + `upperProvider` into `pipeline/testutil_test.go` | ✅ `DONE` | Low    | Low    | **Done 2026-09-08** (`pipeline/testutil_test.go`).                                                                                                                                                            |
| Fuzz `FixOutcome`/`FixApplyResult` `UnmarshalJSON`                          | ✅ `DONE` | Med    | Low    | **Done 2026-09-08:** `FuzzFixOutcomeUnmarshalJSON`/`FuzzFixApplyResultUnmarshalJSON` (`pipeline/fix_outcome_test.go`); 2.6M execs clean.                                                                      |
| Golden wire test pinning the `FixApplyResult` doc-snippet bytes             | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** `TestFixApplyResult_GoldenWireBytes` pins exact deterministic bytes (880B golden, `pipeline/fix_outcome_test.go`).                                                                       |
| Benchmark `ApplyWithOutcomes` mixed outcomes at n=1000                      | ✅ `DONE` | Med    | Low    | **Done 2026-09-08:** `BenchmarkFixEngine_ApplyWithOutcomes_Mixed_1000` - 1293 allocs/op at n=1000 (`pipeline/fix_engine_bench_test.go`).                                                                      |
| Property test: `OutcomeCounts` sum == fixable findings                      | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** 100 seeded corpora, one-outcome-per-finding invariant (`pipeline/fix_outcome_test.go`).                                                                                                  |
| Fix `safePath, _ =` error discard in `pipeline/fix_applier.go:305`          | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** unsafe-path findings surface as `FixOutcomeFailed` with validation error + joined error return; safe findings still apply.                                                               |
| ADR-017/018: outcome metrics + typed outcome errors                         | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** ADR-016/017/018 all in `docs/architecture-decisions.md`.                                                                                                                                 |

### Docs & quality (unblocked)

| Task                                                                           | Status    | Impact | Effort | Notes                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------ | --------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Write `docs/guides/outcomes.md` consumer guide                                 | ✅ `DONE` | Med    | Med    | **Done 2026-09-08:** full guide (statuses, rollback semantics incl. backed-up nuance, typed errors, metrics, JSON) + README link.                                                                         |
| DOMAIN_LANGUAGE: outcomes/rollback pointer + Metrics term cross-check          | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** Metrics/MetricsSnapshot outcome counts + cross-link (`docs/DOMAIN_LANGUAGE.md`).                                                                                                     |
| `scripts/docs-api-check.sh`: backtick identifiers in FEATURES.md exist in code | ✅ `DONE` | Med    | Med    | **Done 2026-09-08:** guard covers FEATURES/API_STABILITY/DOMAIN_LANGUAGE/USAGE_GUIDE (297 identifiers), wired into ci.yml. Caught GetCategory + 5 nonexistent Tag constants.                              |
| Align golangci-lint local↔CI; migrate `exhaustruct` → `exhaustruct_v5`         | ✅ `DONE` | Med    | Low    | **Done 2026-09-08:** CI pinned 2.13.1; exhaustruct_v5 with full-string ignore-patterns; FindingGroup→Group; 0 issues x4 modules.                                                                          |
| `nix fmt` as pre-commit/CI signal for Go treefmt drift                         | ✅ `DONE` | Low    | Low    | **Done 2026-09-08:** pre-commit runs `nix fmt -- --fail-on-change`; found+fixed a stale `core.hooksPath` that had silently disabled the whole hook.                                                       |
| Consider `.golangci.yml` gofumpt/golines autofix config                        | ✅ `DONE` | Low    | Low    | **Decided 2026-09-08:** formatters already configured (gci/goimports/gofumpt/golines@120, `.golangci.yml:368`) mirroring treefmt; `golangci-lint fmt` is the autofix. Rules must not diverge (AGENTS.md). |
| IntervalTree go/no-go research note in ROADMAP                                 | ✅ `DONE` | Low    | Low    | **Decided 2026-09-08: NO-GO.** Note in ROADMAP "Performance": measured µs-scale at consumer scale, cache-hostile trees, max-End augmentation as future middle path.                                       |

## 🟢 LOW Priority

| Task                                      | Status       | Impact | Effort | Notes                                                                                                                                                                                         |
| ----------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| FlightRecorder idea triage → scored table | ✅ `DONE`    | Low    | Low    | **Done 2026-09-08.** Scored table in ROADMAP "FlightRecorder future directions"; rotation + gzip graduated, 2 marked shipped, 4 rejected with reasons.                                        |
| Fix BuildFlow auto-configure loop         | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact. No longer the commit gate (dprint hook replaced it). |
| Consumer compatibility test               | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code.                                                                                                         |

### Post-account-switch CI work (gated on billing fix)

| Task                                                        | Status       | Impact | Effort | Notes                                                                                                                                                        |
| ----------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Re-run + verify ALL CI jobs on master                       | 🔵 `BLOCKED` | High   | Low    | Attempted 2026-09-08 (run 34247500002): all jobs red on billing, logs unavailable. Re-run after the account switch.                                          |
| Dependabot PR triage                                        | 🔵 `BLOCKED` | Med    | Med    | Needs green CI to validate #23/#24/#25/#29 (billing).                                                                                                        |
| CI `workflow_dispatch` inputs for module-scoped runs        | ✅ `DONE`    | Low    | Low    | **Done 2026-09-08:** `modules` input scopes the lint matrix on dispatch runs (`.github/workflows/ci.yml`).                                                   |
| Gate CI benchmark job on `bench-check.sh` regression        | ✅ `DONE`    | Low    | Low    | **Verified 2026-09-08:** job already gates on bench-check.sh - and the release train fixed the awk no-op that had silently disabled the check.               |
| Evaluate `ginkgo --repeat=N` in CI vs sampled stress matrix | ✅ `DONE`    | Low    | Low    | **Decided 2026-09-08:** CI keeps `go test -race -count=20` only; ginkgo repeat stays the local mandatory release gate (note in `docs/release-procedure.md`). |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
