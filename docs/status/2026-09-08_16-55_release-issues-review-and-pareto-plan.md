# Status Report — Release Readiness, Issue/PR Review & Pareto Plan v1.7.0

**Date:** 2026-09-08 16:55 CEST
**Session arc (this stretch):** docs-health audit (16-46 report) → user asked
"Time for a new version and release?" → release-readiness check → "Reviewed all
GitHub Issues?" → issue review → this report + a comprehensive Pareto plan.
**Predecessor:** `docs/status/2026-09-08_16-46_docs-health-annotation-archive-sweep.md`

---

## a) FULLY DONE (verified)

| # | Item | Evidence |
| - | ---- | -------- |
| 1 | **Release-readiness assessment** — local master 8 ahead / 0 behind origin, tree clean (daemon-committed through `f66b9cb`); `version.go` = 1.6.0; `[Unreleased]` sections present in root + all 3 sub-module CHANGELOGs; verdict: **v1.7.0 is fully loaded — every pre-release gate already ran green on this code (race ×4, lint ×4, structural scripts, arch-lint, stress repeat=20, flake check)**; only D1/D6/push gate the train | git/version.go/CHANGELOG checks; 15-42 report §verification |
| 2 | **GitHub issues review** — exactly 2 issues exist, both OPEN, both created 2026-09-07, both by the owner, **zero comments, no labels**: #27 (FixEngine zero-edit == success) and #28 (one resolveError rolls back ALL edits). Both fixes are on master, race/fuzz/golden-verified, changelogged in [Unreleased]; #28 is cited by name in the CHANGELOG Changed entry | `gh issue list --state all`; `gh issue view --json` ×2 |
| 3 | **PR-queue review (late — see d)** — 4 OPEN Dependabot PRs: #23 sbom-action 0.24.0→0.24.2, #24 pipeline gomod ×2, #25 CLI gomod ×2, #29 gomega 1.42.1→1.43.0 (CLI); #26 closed | `gh pr list --state all` |
| 4 | **Pareto plan created** — `docs/planning/2026-09-08_16-55-SUPERB-pareto-plan-v1.7.0-release-and-tail.html` (self-contained Bauhaus HTML, 97 KB, D2 graph inlined as SVG): Pareto tiers (1% = ship v1.7.0; 4% = +consumer bumps; 20% = +verification & quality tail), **24 comprehensive tasks (30–100 min)** and **85 micro tasks (≤ 12 min each — user cap)**, every task impact/effort/gate-badged; consumes TODO_LIST.md + audit findings + issues + PR queue completely | plan file; `d2` render |
| 5 | **Skills honored** — pareto-planning SKILL.md + html-report-kit guide loaded BEFORE planning; docs-health ANNOTATE convention respected (plans annotated, never rewritten) | session log |
| 6 | **Recommendations delivered** — D2 (comment+close #27/#28 now, alternative dismissed) and implicit D1 (ship per-file default; #28 is its motivation) | conversation |

## b) PARTIALLY DONE

| Item | Done | Remaining |
| ---- | ---- | --------- |
| v1.7.0 release train | All pre-release work complete (prior sessions); mechanical steps enumerated as REL1–REL19 in the plan | Execution — frozen on D1 + D6 + push word |
| Issues #27/#28 | Reviewed; fix summaries drafted in essence; recommendation given | Comments + closure + labels not posted (awaiting D2) |
| TODO_LIST ↔ plan linkage | Plan consumes TODO_LIST fully | Back-reference (plan link in TODO_LIST header) not yet added |

## c) NOT STARTED

- Execution of the plan: none of the 85 micro tasks started (this stretch was assessment + planning only).
- ADR-016 draft (awaits D1); Dependabot triage (awaits account switch); consumer bumps (awaits tag).

## d) TOTALLY FUCKED UP (honest failures)

1. **"Reviewed all GitHub Issues?" was answered literally.** I reviewed issues but NOT the PR queue — the 4 open Dependabot PRs only surfaced now, when the status-report prompt made me look. A tracker review means issues AND PRs AND discussions; one `gh` sweep each, every time.
2. **The release question was never actually answered.** After the readiness check I had the facts but delivered no verdict before the topic pivoted to issues; the answer (ready, gated on D1/D6/push) only appeared folded into the issues reply. Facts-first is fine; verdict-must-follow.
3. **Release procedure not re-read.** My REL1–REL19 steps mirror the reports' description of `docs/release-procedure.md`, not the file verbatim. Must verify against the actual procedure before executing the train.
4. **`gh issue view 27 --comments` returned empty output** and I silently retried with `--json` instead of diagnosing. It worked, but the silent-retry habit is how tool quirks become folklore.
5. **Standing self-critique from 16-46 still applies** (agent-trusted annotations, unilateral Won't-implement closures, pseudo-precise health scores) — none re-addressed in this stretch.

## e) WHAT WE SHOULD IMPROVE

1. **Tracker review checklist:** `gh issue list --state all` + `gh pr list --state all` + discussions — all three, one command each.
2. **Answer in the same breath as the facts:** deliver the verdict + needed decisions immediately; detail can follow.
3. **Read the procedure file before planning against it** — plans built from memory inherit memory's drift.
4. **Link every plan from TODO_LIST/AGENTS** so the next session finds it (execution step 0).

## f) Up to 50 things next (top slice — full 85-task plan: `docs/planning/2026-09-08_16-55-SUPERB-pareto-plan-v1.7.0-release-and-tail.html`)

1. REL1 D1 sign-off → ADR-016. 2. REL2 D6 decision record. 3. REL3–REL6 stamp 4 CHANGELOGs. 4. REL7 version.go 6→7. 5. REL8–REL9 README/ROADMAP version refs. 6. REL10 version-check. 7. REL11 bench-check. 8. REL12 GOWORK=off ×4. 9. REL13 stress gate. 10. REL14 nix flake check. 11. REL15 dprint+freshness. 12. REL16 tag ×4. 13. REL17 push (8 commits + tags). 14. REL18 release run + assets (billing-gated). 15. REL19 proxy smoke. 16. REL20–REL22 comment+close+label #27/#28 (D2). 17. CONS1 humanize bump. 18. CONS2 SDK v0.2.0. 19. CONS3–CONS4 consumer sweep. 20. CONS5 migration note. 21. CI1 full CI verify (switch). 22. CI2–CI5 Dependabot #23/#24/#25/#29. 23. CI6–CI8 CI enhancements. 24. TEST1–TEST7 Phase D test bundle. 25. TEST8 safePath discard fix. 26. TEST9–TEST11 GroupFindings/WithGroupID. 27. TEST12–TEST13 ADR-017/018. 28. DOCS1–DOCS3 outcomes guide. 29. DOCS4 DOMAIN_LANGUAGE pointer. 30. DOCS5–DOCS6 docs-api-check.sh. 31. DOCS7–DOCS9 AGENTS diet. 32. DOCS10–DOCS13 living-docs verification. 33. DOCS14 link-check. 34. DOCS15 feedback GAP annotations. 35. DOCS16 lint alignment. 36. DOCS17–DOCS21 quality signals. 37. HIST1–HIST8 historical docs completion. 38. GATE1–GATE3 D7/D3/D4 designs. 39. LAUNCH1–LAUNCH4 launch track. 40. FR1 FlightRecorder triage. (41–85: the per-task granularity of 24–40 — see the plan HTML.)

## g) Questions I CANNOT figure out myself

1. **D1 — rollback default:** confirm shipping v1.7.0 with `RollbackPolicyFailingFile` as the default? One word unlocks ADR-016 + the whole REL train (all gates already green).
2. **D6 + push — release strategy:** forward-only v1.7.0 (recommended; backfill v1.5/v1.6 later if ever) — and do I push the 8 local commits + new tags once stamped?
3. **D2 — issues:** close #27/#28 now with fix-summary comments (recommended), or at tag time?

---

**Verification snapshot:** readiness facts checked (git/version/CHANGELOGs) · issues + PRs enumerated via `gh` · plan HTML self-contained (SVG inlined, dprint-checked repo) · tree clean.

**Then wait for instructions.**
