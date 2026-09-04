# Post-Audit Comprehensive Execution Plan

> **Status: EXECUTED.** 63/68 tasks completed, 5 BLOCKED (SARIF schema vendoring — needs user decision). 5 real bugs found and fixed. Full details in `docs/status/2026-07-24_22-16_post-audit-execution-full-run.md`. Follow-up remediation plan: `docs/planning/2026-07-24_22-20_post-session-remediation-and-deep-verification.md`.

**Date:** 2026-07-24 22:04
**Source:** Status report `2026-07-24_21-58_todo-list-freshness-audit.md`
**Scope:** All 50 TODOs from the session self-review, split into max-12-minute tasks

---

## Pareto Breakdown

### The 1% that delivers 51%

**Fix the harm I caused this session.** I made 3 mistakes (wrongly removed SARIF item, skipped tests/lint, deleted completed items without CHANGELOG). Fixing these is non-negotiable — they represent actual damage, not just missing work.

### The 4% that delivers 64%

**Verify my changes are correct and record completed work.** Re-read files I edited, verify cross-file consistency, reconcile CHANGELOG. This catches any secondary damage from my rebuild.

### The 20% that delivers 80%

**Quick verification wins across the repo.** Fast checks (5-10 min each) that catch broad categories of drift: broken links, ghost files, wrong counts, stale claims. High signal per minute.

### The remaining 80% (last 20%)

**Deep subsystem audits and blocked work.** Thorough verification of individual subsystems (pipeline, LSP, SARIF, consumer APIs). Important for confidence but lower urgency.

---

## Full Task Table — Sorted by Impact, then Effort (quick wins first)

### TIER 1: CRITICAL — Fix My Mistakes (7 tasks, ~55 min total)

> These represent actual harm. Must complete before anything else.

| # | Task                                                                                                                  | Impact   | Effort | Source      | Deps |
| - | --------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ----------- | ---- |
| 1 | Restore SARIF schema validation to TODO_LIST.md as BLOCKED (was wrongly removed — it's BLOCKED not DEFERRED)          | Critical | 5m     | Mistake #1  | —    |
| 2 | Remove SARIF schema validation from ROADMAP.md "Hardening" section (avoid split brain — keep only in TODO as BLOCKED) | Critical | 5m     | Mistake #1  | #1   |
| 3 | Run `export GOEXPERIMENT=jsonv2 && go test -race -count=1 ./...` (full test suite — I only ran build)                 | Critical | 10m    | Mistake #2  | —    |
| 4 | Analyze and fix any test failures found                                                                               | Critical | 12m    | Mistake #2  | #3   |
| 5 | Run `golangci-lint run ./...` (linter — I skipped this)                                                               | Critical | 8m     | Mistake #2  | —    |
| 6 | Fix any lint issues found                                                                                             | Critical | 12m    | Mistake #2  | #5   |
| 7 | Run `nix flake check` (Nix quality gate — never ran)                                                                  | Critical | 8m     | Not started | —    |

### TIER 2: HIGH — Correctness & CHANGELOG Reconciliation (8 tasks, ~60 min total)

> Verify my rebuild is correct. Record completed work properly.

| #  | Task                                                                                              | Impact | Effort | Source         | Deps  |
| -- | ------------------------------------------------------------------------------------------------- | ------ | ------ | -------------- | ----- |
| 8  | Re-read TODO_LIST.md end-to-end — verify it's coherent after rebuild                              | High   | 5m     | Self-check     | #1,#2 |
| 9  | Re-read ROADMAP.md end-to-end — verify "Hardening" section flows after enrichment                 | High   | 8m     | Improve #6     | #2    |
| 10 | Review all 18 post-v1.3.0 commits — classify as user-facing vs internal                           | High   | 10m    | Mistake #3     | —     |
| 11 | Draft CHANGELOG `[Unreleased]` entries for user-facing commits                                    | High   | 10m    | Mistake #3     | #10   |
| 12 | Add `[Unreleased]` section to CHANGELOG.md                                                        | High   | 5m     | Mistake #3     | #11   |
| 13 | Verify FEATURES.md ↔ TODO_LIST.md consistency (no PLANNED-in-TODO + FULLY_FUNCTIONAL-in-FEATURES) | High   | 10m    | Partially done | #8    |
| 14 | Verify FEATURES.md ↔ ROADMAP.md consistency (no duplicated deferred items)                        | High   | 8m     | Cross-file     | #9    |
| 15 | Decide: does post-v1.3.0 work warrant a v1.3.1 tag? (depends on user answer to Q2)                | High   | 5m     | #21            | #12   |

### TIER 3: MEDIUM-HIGH — Quick Verification Wins (11 tasks, ~70 min total)

> Fast checks (5-8 min each) that catch broad drift categories. Maximum signal per minute.

| #  | Task                                                                                              | Impact   | Effort | Source      | Deps |
| -- | ------------------------------------------------------------------------------------------------- | -------- | ------ | ----------- | ---- |
| 16 | Check all internal markdown links resolve across entire repo (`rg -roE '\]\([^)]+\)' *.md docs/`) | Med-High | 5m     | #35         | —    |
| 17 | Verify all AGENTS.md "Key Files" paths exist (every file in the table)                            | Med-High | 8m     | #15         | —    |
| 18 | Verify Core module `go.mod` has zero production deps (ginkgo/gomega are test-only)                | Med-High | 5m     | #44         | —    |
| 19 | Check `go.work` includes all 4 modules and versions are synced                                    | Med-High | 5m     | #46         | —    |
| 20 | Verify replace directives in pipeline/, analysis/, cmd/go-finding/ point to correct core path     | Med-High | 8m     | #45         | —    |
| 21 | Check for exhaustruct `//nolint` leftovers (33 were removed — verify none missed)                 | Med-High | 5m     | #41         | —    |
| 22 | Fix AGENTS.md json/v2 file count: claims "9 files" but actual count is **24**                     | Med-High | 5m     | New finding | —    |
| 23 | Verify FixStrategyAI has no backend (confirm FEATURES "PLANNED" is honest)                        | Med-High | 5m     | #10         | —    |
| 24 | Run `bash scripts/version-check.sh` — verify version.go matches git tag                           | Med-High | 5m     | #24         | —    |
| 25 | Verify makezero `always: false` is still the correct setting                                      | Med      | 5m     | #43         | —    |
| 26 | Verify CHANGELOG version/compare links match repo URL pattern                                     | Med      | 5m     | #20         | —    |

### TIER 4: MEDIUM — GOWORK=off Per-Module Isolation (4 tasks, ~20 min total)

> Each module must build independently. AGENTS.md documents this as a required CI check.

| #  | Task                                                                                 | Impact | Effort | Source | Deps |
| -- | ------------------------------------------------------------------------------------ | ------ | ------ | ------ | ---- |
| 27 | Run `GOWORK=off GOEXPERIMENT=jsonv2 go test ./...` in Core module (`.`)              | Med    | 5m     | #22    | —    |
| 28 | Run `GOWORK=off GOEXPERIMENT=jsonv2 go test ./...` in Pipeline module (`pipeline/`)  | Med    | 5m     | #22    | —    |
| 29 | Run `GOWORK=off GOEXPERIMENT=jsonv2 go test ./...` in Analysis module (`analysis/`)  | Med    | 5m     | #22    | —    |
| 30 | Run `GOWORK=off GOEXPERIMENT=jsonv2 go test ./...` in CLI module (`cmd/go-finding/`) | Med    | 5m     | #22    | —    |

### TIER 5: MEDIUM — AGENTS.md Verification (3 tasks, ~26 min total)

> AGENTS.md is the primary context file for AI sessions. Drift here misleads every future session.

| #  | Task                                                                                               | Impact | Effort | Source | Deps |
| -- | -------------------------------------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 31 | Verify AGENTS.md module table (paths, dependency lists, "Depends on" column)                       | Med    | 8m     | #14    | —    |
| 32 | Test AGENTS.md build commands actually work: `nix run .#test`, `nix run .#bench`, `nix run .#lint` | Med    | 10m    | #16    | —    |
| 33 | Verify AGENTS.md "Removed APIs (v1.0.0)" — grep for any deprecated API still in code               | Med    | 8m     | #18    | —    |

### TIER 6: MEDIUM — FEATURES.md Section-by-Section (8 tasks, ~70 min total)

> FEATURES.md is 49KB. Verify each section against code.

| #  | Task                                                                                | Impact | Effort | Source | Deps |
| -- | ----------------------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 34 | Verify FEATURES "Finding Type" table — every field name + type matches `finding.go` | Med    | 8m     | #9     | —    |
| 35 | Verify FEATURES Builder section — every `With*` method exists in builder code       | Med    | 8m     | #9     | —    |
| 36 | Verify FEATURES Position & Range section — methods match `position.go`/`range.go`   | Med    | 8m     | #9     | —    |
| 37 | Verify FEATURES Pipeline section — stages, config, hooks match `pipeline/` code     | Med    | 10m    | #9     | —    |
| 38 | Verify FEATURES SARIF section — export/import claims match `sarif_*.go`             | Med    | 8m     | #9     | —    |
| 39 | Verify FEATURES LSP section — `ToLSP`/`FromLSP` claims match `lsp.go`               | Med    | 8m     | #9     | —    |
| 40 | Spot-check 5 API signatures in FEATURES against actual Go signatures                | Med    | 10m    | #13    | —    |
| 41 | Check for post-v1.3.0 shipped features missing from FEATURES.md (grep 18 commits)   | Med    | 8m     | #11    | #10  |

### TIER 7: MEDIUM — Consumer Claim Verification (3 tasks, ~23 min total)

> The "22 consumers, 14 with Go code" claim is unverified. Could be stale.

| #  | Task                                                                                         | Impact | Effort | Source | Deps |
| -- | -------------------------------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 42 | Read `docs/reviews/2026-07-05_20-55_consumer-audit.html` — verify "22 consumers, 14 with Go" | Med    | 10m    | #5     | —    |
| 43 | Update consumer count claim in TODO_LIST.md if stale                                         | Med    | 5m     | #5     | #42  |
| 44 | Assess whether BuildFlow BLOCKED item is still relevant or should be closed                  | Med    | 8m     | #6     | —    |

### TIER 8: MEDIUM — Documentation Verification (5 tasks, ~40 min total)

| #  | Task                                                                                       | Impact  | Effort | Source    | Deps |
| -- | ------------------------------------------------------------------------------------------ | ------- | ------ | --------- | ---- |
| 45 | Verify `docs/guides/fix-engine.md` accuracy against current FixEngine API                  | Med     | 10m    | #36       | —    |
| 46 | Assess `docs/MIGRATION_v1.0.md` — still needed, or archive?                                | Low     | 5m     | #37       | —    |
| 47 | Verify `docs/release-procedure.md` matches actual release process                          | Med     | 8m     | #38       | —    |
| 48 | Audit `docs/reviews/` for stale references (13 HTML reports — check for deleted code refs) | Low-Med | 12m    | #39       | —    |
| 49 | Verify `docs/DOMAIN_LANGUAGE.md` exists and is current (AGENTS.md references it)           | Med     | 5m     | Doc model | —    |

### TIER 9: MEDIUM — Pipeline Verification (4 tasks, ~42 min total)

| #  | Task                                                                                | Impact | Effort | Source | Deps |
| -- | ----------------------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 50 | Audit FixProvider chain: OffsetProvider → LineProvider → SubstringProvider ordering | Med    | 10m    | #29    | —    |
| 51 | Verify line-shift map correctness in multi-edit scenarios (read tests)              | Med    | 10m    | #30    | —    |
| 52 | Check conflict detection edge cases — byte-level overlap precision                  | Med    | 12m    | #31    | —    |
| 53 | Verify StageHooks contract — before/after events, abort behavior                    | Med    | 10m    | #32    | —    |

### TIER 10: MEDIUM — LSP Verification (2 tasks, ~15 min total)

| #  | Task                                                                        | Impact | Effort | Source | Deps |
| -- | --------------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 54 | Verify LSPDiagnosticData round-trip tests exist and cover all fields        | Med    | 10m    | #33    | —    |
| 55 | Verify SeverityCritical LSP round-trip test exists (LSP collapses to Error) | Med    | 5m     | #34    | —    |

### TIER 11: LOW-MEDIUM — Benchmarks & Code Quality (3 tasks, ~25 min total)

| #  | Task                                                                                    | Impact  | Effort | Source | Deps |
| -- | --------------------------------------------------------------------------------------- | ------- | ------ | ------ | ---- |
| 56 | Run `nix run .#bench` — performance regression check                                    | Low-Med | 10m    | #40    | —    |
| 57 | Compare benchmarks to `benchmarks/baseline.txt` via `scripts/bench-check.sh`            | Low-Med | 5m     | #23    | #56  |
| 58 | Audit `.golangci.yml` exclusions — are they all still needed after exhaustruct cleanup? | Low-Med | 10m    | #42    | #21  |

### TIER 12: LOW — SARIF Schema (BLOCKED — 5 tasks, ~52 min total)

> These are blocked on a user decision (vendoring 7K-line schema). Tracked but may not be actionable yet.

| #  | Task                                                              | Impact | Effort | Source | Deps    |
| -- | ----------------------------------------------------------------- | ------ | ------ | ------ | ------- |
| 59 | Research SARIF 2.1.0 schema vendoring approach (embed? testdata?) | Low    | 10m    | #26    | User OK |
| 60 | Download SARIF 2.1.0 JSON schema from official source             | Low    | 8m     | #26    | #59     |
| 61 | Vendor schema into `testdata/sarif/`                              | Low    | 10m    | #26    | #60     |
| 62 | Write SARIF schema validation test using vendored schema          | Low    | 12m    | #27    | #61     |
| 63 | Verify existing SARIF round-trip tests cover all properties       | Low    | 8m     | #28    | —       |

### TIER 13: LOW — Consumer Ecosystem Verification (5 tasks, ~42 min total)

| #  | Task                                                                    | Impact | Effort | Source | Deps |
| -- | ----------------------------------------------------------------------- | ------ | ------ | ------ | ---- |
| 64 | Test `BuildOrDefault()` returns zero `Finding{}` on validation error    | Low    | 8m     | #47    | —    |
| 65 | Test `Template.Build()` creates correctly stamped findings              | Low    | 8m     | #47    | —    |
| 66 | Verify `ApplySimpleFixes()` works on sample BeforeCode→AfterCode data   | Low    | 8m     | #48    | —    |
| 67 | Verify `CheckBinary`/`RunCmd` produce correct `NewIOError` on failure   | Low    | 8m     | #49    | —    |
| 68 | Review `ToolAdapter[O]` API — ready for consumer use or still internal? | Low    | 10m    | #50    | —    |

---

## Summary Statistics

| Metric                                          | Value                       |
| ----------------------------------------------- | --------------------------- |
| Total tasks                                     | 68                          |
| Total estimated time                            | ~535 min (~9 hours)         |
| Tier 1 (Critical — fix my mistakes)             | 7 tasks / ~55 min           |
| Tier 2 (High — correctness & CHANGELOG)         | 8 tasks / ~60 min           |
| Tier 3 (Med-High — quick wins)                  | 11 tasks / ~70 min          |
| Tier 4-6 (Medium — module/doc verification)     | 15 tasks / ~116 min         |
| Tier 7-10 (Medium — subsystem verification)     | 14 tasks / ~120 min         |
| Tier 11-13 (Low — benchmarks, SARIF, consumers) | 13 tasks / ~119 min         |
| Blocked on user decisions                       | 6 tasks (TIER 12 + #15)     |
| New findings this session                       | 1 (#22: json/v2 count 9→24) |

---

## Execution Order Recommendation

```
Phase 1 (1% → 51%): Tasks #1-7     — Fix my mistakes NOW
Phase 2 (4% → 64%): Tasks #8-15    — Verify correctness, reconcile CHANGELOG
Phase 3 (20% → 80%): Tasks #16-26  — Quick verification sweep
Phase 4 (80% → 100%): Tasks #27-68 — Deep audits (parallelizable by subsystem)
```

**Phases 1-3 should be done before any commit.** Phase 4 can be spread across sessions.

---

_Assisted-by: Crush <crush@charm.land>_
