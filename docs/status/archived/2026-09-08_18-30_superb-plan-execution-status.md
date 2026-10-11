# SUPERB Pareto Plan — Execution Status

> **📦 RESOLUTION STATUS (updated 2026-09-08 21:30)**
>
> SUPERSEDED by the evening-session report
> (`2026-09-08_21-30_evening-session-v1.8.0-questions-resolved.md`).
> All 3 pending questions resolved; v1.8.0 shipped; 19 consumers bumped;
> FR rotation + gzip landed for v1.9.0.
> Open items are externally blocked (billing/public flip) or ROADMAP-tracked.

**Date:** 2026-09-08 (evening session)
**Input:** `docs/planning/2026-09-08_16-55-SUPERB-pareto-plan-v1.7.0-release-and-tail.html` (24 comprehensive + 85 micro tasks)
**Result:** Executed end to end, same day. Per-task verdicts with evidence are annotated inline in the plan HTML (strikethrough + ✅/🔵); row-level statuses in TODO_LIST.md.

## Headline

**v1.7.0 SHIPPED.** All four module tags pushed (`v1.7.0`, `pipeline/v1.7.0`, `analysis/v1.7.0`, `cmd/go-finding/v1.7.0`), all local gates green, proxy smoke resolves all four modules at v1.7.0. Issues #27/#28 closed with fix summaries and labels. Both lead consumers bumped and verified (go-humanize-linter @v1.7.0; go-linter-sdk @v1.7.0, tagged v0.3.0 since v0.2.0 already existed).

## Decisions executed (user delegated via blanket directive)

- **D1** — ship per-file rollback default (ADR-016 written)
- **D6** — forward-only release, push at tag time
- **D2** — close #27/#28 now
- **D7** — GroupID validation: permissive machine-identifier format (no whitespace/control chars, ≤128B); strict charset rejected as consumer-hostile
- **D3** — additive `Config.OnFixOutcome` over breaking OnFix (deprecated, retained)
- **D4** — `FixApplier.ApplyDryRun`: full ApplyReport semantics, zero writes

## Blocked externally (11 micro tasks, 2 comprehensive)

- Billing: REL18 (release run; dispatch failed instantly, logs unavailable; **`HOMEBREW_TAP_GITHUB_TOKEN` secret does not exist** — must be created before the post-switch release run), CI1–5
- Public flip: LAUNCH1/3/4 (repo verified PRIVATE; LAUNCH2 announcement drafted and parked in `docs/brainstorming/launch-announcement-draft.md`)
- By design: HIST7 (next session), DOCS15 (feedback-doc annotations deferred)

## Bugs found and fixed beyond the plan (the dead-gate class)

1. **bench-check.sh regression check was a silent no-op** — awk matched `/^[Bb]enchmark/` but benchstat strips the prefix; the gate had never detected anything. Fixed; it immediately flagged real deltas (GroupID struct cost, identical allocations), baseline regenerated same-session with rationale.
2. **The dprint pre-commit hook was dead** — stale local `core.hooksPath=.githooks` pointed at a nonexistent dir, disabling the gate since it was set. Revived; extended with `nix fmt -- --fail-on-change` (DOCS17).
3. **version-drift.sh aborted silently on drift** — `set -euo pipefail` + a version-pinned grep exited before printing diagnostics. This is why the release train initially missed that all sub-module go.mod files still required v1.6.0; caught by the final sweep, fixed (go.mod refs + script).
4. **flake vendorHash drift** — automated 05:45 go.sum housekeeping paired with a stale vendorHash; `nix flake check` failed until synced.
5. **Docs drift caught by the new guard** — USAGE_GUIDE claimed 5 nonexistent `Tag*` constants and used one in a compiling example; SPLIT-BRAIN archived doc had a move-broken relative link.

## Verification (final sweep, all green)

race tests ×4 modules, golangci-lint 2.13.1 ×4 (0 issues), replace-audit, version-drift, test-naming, go-work-sync, docs-freshness (0/0), docs-api-check (297 identifiers), json-deterministic, go-arch-lint, dprint, `nix flake check`. Post-tag additions also passed the mandatory stress gate before the release train (ginkgo repeat=20 race + go test count=20 race).

## Shipped beyond the release (now in `[Unreleased]`)

`GroupID` validation (D7), `Config.OnFixOutcome` (D3), `FixApplier.ApplyDryRun` (D4), `Report.GroupFindingsSorted` + `Group`, `Template.WithGroupID`, unsafe-path failed-outcome surfacing, ADR-017/018, docs-api-check drift guard, outcomes guide, AGENTS diet (37→22 KB), consumer sweep table, June-HTML per-finding annotations, master-plan disposition, CI dispatch inputs + stress-scope decision, lint toolchain alignment (exhaustruct_v5, CI 2.13.1).

## Next session

- HIST7: annotate `2026-09-08_15-42` once superseded by this report.
- DOCS15: feedback-doc GAP-3/D5/D8 annotations.
- Post-account-switch: re-run CI (module-scoped dispatch now available), verify all jobs, triage Dependabot #23/#24/#25/#29, create the HOMEBREW secret, watch the v1.7.0 release run, then decide on backfilling v1.5/v1.6 releases.
- Consumer upgrades: 8 pipeline consumers flagged for the rollback migration (table in `docs/ecosystem.md`).
