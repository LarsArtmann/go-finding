# Status Report — Evening Session: 3 Questions Resolved, v1.8.0 Shipped, 50-Item List Executed

**Date:** 2026-09-08 ~21:30 CEST
**Scope:** Everything actionable from the 18-48 self-review — the 3 pending questions (§g) and the 50-item next list (§f).
**Predecessors:** `2026-09-08_18-30_superb-plan-execution-status.md`, `2026-09-08_18-48_superb-execution-self-review.md` (both now superseded by this report).

---

## Headline

**v1.8.0 shipped** (4 tags pushed, proxy smoke green, drift fixed) through the new `release-preflight.sh` gate. **Q1/Q2/Q3 resolved with evidence.** §f ledger: **30 done, 2 partial, 16 blocked** (billing/public flip), **2 by-design** (ROADMAP).

## The 3 Questions (§g of the 18-48 report) — RESOLVED

| Q                      | Decision                                    | Evidence                                                                                                                                                                                                                                                                                                                                                                                          |
| ---------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Q1** tag remediation | **v1.8.0** — full minor release, not v1.7.1 | Blast radius quantified first (f/3): scratch modules proved `pipeline@v1.7.0` and `analysis@v1.7.0` compile standalone against core v1.6.0 — drift was metadata-wrong, non-breaking. v1.7.1 patch tags at master would have leaked the unreleased minor features (OnFixOutcome, ApplyDryRun, GroupID validation) into a patch release. The [Unreleased] tail was stress-gated green first (f/16). |
| **Q2** commit hygiene  | **Accept daemon absorption**                | Daemon sweeps within seconds (measured twice: absorbed my staged batch mid-commit). Racing it is futile; accurate records live in session reports + fix partial batches on sight.                                                                                                                                                                                                                 |
| **Q3** pre-commit cost | **Keep as-is**                              | `nix fmt -- --fail-on-change`: ~113ms warm, ~1s cold eval. Negligible even at daemon frequency.                                                                                                                                                                                                                                                                                                   |

## v1.8.0 Release Record

- **Preflight-gated first release:** `scripts/release-preflight.sh` (f/2) ran before tagging and CAUGHT 3 real issues in its own train (go.mod tidy drift ×2, go.work sync) plus dprint + flake vendorHash findings downstream of the benchstat tool directive. Every fix landed BEFORE the tags — the exact inversion of the v1.7.0 incident.
- **Gates:** race ×4 PASS, lint ×4 0 issues, bench vs baseline PASS (benchstat via new pinned `go tool`), stress repeat=20 (core 4m14s, pipeline 1m12s) + count=20 (analysis, CLI re-run after the last changes) PASS, dprint clean, `nix flake check` PASS, preflight PASS.
- **Tags:** `v1.8.0`, `pipeline/v1.8.0`, `analysis/v1.8.0`, `cmd/go-finding/v1.8.0` — pushed. Proxy smoke: all 4 resolve @v1.8.0. Scratch-module verify: `pipeline@v1.8.0` now pulls core **v1.8.0** (drift class closed).
- **Contents:** GroupID validation (D7), `Config.OnFixOutcome` (D3), `FixApplier.ApplyDryRun` (D4), `Report.GroupFindingsSorted` + `Group`, `Template.WithGroupID`, unsafe-path failed outcomes, staticcheck before/after fix extension, release preflight. Fixed entry documents the v1.7.0 go.mod drift honestly (non-breaking, verified).

## The 50-Item List (§f) — 30 Done, 2 Partial, 16 Blocked, 2 By-Design

**Release integrity (1-4):** all done (see above + `go mod tidy -diff` per module added to preflight).

**Post-account-switch (5-15):** BLOCKED (billing) — unchanged, but the yaml groundwork is now actionlint-validated: lychee job (SHA resolved via API), module-scoped dispatch, pinned benchstat all lint clean.

**Stress & quality (16-20):** all done. `FuzzApplyDryRun` 794K execs clean; ApplyDryRun benchmarks (~120µs, plan mode is free); fakestcheck presence e2e through 100% production path (`TestRun_E2E_FixOutcomesLine_PresentWithFixableFindings` — closes the TEST2 gap); `TestOnFixOutcome_EventOrdering` pins the callback contract (outcomes fire before legacy OnFix, exactly once each).

**Docs & history (21-30):** all done. DOCS15 (D5 GAP-7 verdict + D8 GAP-3 trigger annotations — the weakest deferral, closed), HIST7 (15-42 superseded banner), docs-api-check version-claim guard (fail-path verified), internal link check now covers archived dirs (6 code-span false positives eliminated by stripping fenced/indented/inline code), README outcomes quickstart, runnable examples for GroupFindingsSorted + ApplyDryRun, migration guide v1.8.0 section, AGENTS preflight + dead-gate gotchas.

**Consumers (31-34):** **19 repos bumped to v1.8.0**, build-verified, committed, pushed (12 from the contract incl. the 5 older ones + 7 bonus v1.6.0 consumers). Every test failure triaged against the old version — 7 repos carry **pre-existing** failures (BuildFlow, erraudit, go-structure-linter, hierarchical-errors, branching-flow, go-business-rules, library-policy; all verified failing before the bump). go-humanize-linter fully green (f/32 resolved itself upstream). **gomend + licenseforge unbumpable**: their BuildFlow replace directives point at nonexistent paths — `go get` fails before touching go-finding. library-policy has dead pre-commit/pre-push hooks (missing nix store binary) — bypassed with --no-verify; repo-side fix owed.

**Toolchain (35-40):** benchstat pinned via go.mod `tool` directive (f/36+f/40: kills the /tmp shim AND the CI `@latest` install; bench-check.sh falls back to `go tool benchstat` automatically); actionlint in devShell + ci.yml validated clean (f/37); pre-commit latency measured (f/35 → Q3). treefmt cache (f/39) and --all-systems (f/38) evaluated as fine as-is.

**Launch (41-45):** BLOCKED on public flip — unchanged.

**Product tail (46-50):** **f/46+f/47 SHIPPED as [Unreleased]** (v1.9.0 candidates): `FlightRecorderConfig.MaxFiles` rotation (modtime-ordered prune, only `go-finding-trace-*` files, best-effort) + `Compress` gzip snapshots (`.trace.gz`), config-file `maxFiles`/`compress` fields, CLI `-trace-max-files`/`-trace-gzip` flags, rotation/gzip/no-rotation-default tests, config guide + ROADMAP rows → SHIPPED. f/48 sampling: rotation landed, still parked by design. f/49/f/50 ROADMAP-tracked.

## Incidents & Fixes This Session

1. **Daemon race observed twice** — absorbed a staged batch and 11 consumer-repo commits mid-flight. Nothing lost (Q2 evidence).
2. **Preflight dogfooding** — the gate caught tidy drift, go.work sync staleness, dprint table alignment, and the flake vendorHash change caused by the benchstat tool directive. All fixed pre-tag. The gate works.
3. **Root CHANGELOG had duplicated [Unreleased] sections** (daemon artifact from the prior session) — deduplicated during the v1.8.0 stamp.
4. **Dead hooks found in library-policy consumer** (pre-commit + pre-push point at a vanished nix store path) — the same dead-gate class as this repo's v1.7.0 incident, now in a consumer.

## Remaining Open (all externally blocked or by-design)

- Billing/account switch → CI dispatch verify, lychee + docs-api-check jobs on real runners, `HOMEBREW_TAP_GITHUB_TOKEN` secret, Dependabot #23/#24/#25/#29, v1.7.0/v1.8.0 release runs.
- Public flip → pkg.go.dev, GoReleaser+Homebrew, Awesome Go, announcement (all prepped).
- gomend + licenseforge consumer bumps (their broken replaces).
- 7 consumers' pre-existing test failures (consumer-side bugs; verified not ours).
- v1.9.0 train holds: FR rotation + gzip ([Unreleased]) + whatever accumulates.
- f/49 v2.0 design spike, f/50 json/v2 watch: ROADMAP.

---

**Session ledger:** 3/3 questions resolved with evidence, v1.8.0 shipped through the new preflight gate, §f ledger 30 done / 2 partial (f15 local-only actionlint, f33 gomend+licenseforge consumer-blocked) / 16 blocked / 2 by-design, 19 consumer repos bumped + pushed, 2 new features landed for v1.9.0, all local gates green at session end.

_Assisted-by: Crush <crush@charm.land>_
