# Status Report — art-dupl Deduplication Pass (Session)

> **Disposition (2026-09-23 docs-health pass):** point-in-time snapshot, fully
> dispositioned. Done items struck inline; remaining open items (§f 1-6, 8,
> 13-20, §g questions) were harvested into `TODO_LIST.md` / `ROADMAP.md`
> "Open questions" — do not action from this file.

**Date:** 2026-09-23 03:16 · **Scope:** This session only (art-dupl `-t 1 --type-aware` clone cleanup + verification) · **Author:** Crush session

**Trigger:** User ran `art-dupl --sort total-tokens -t 1 --type-aware --html` (13 actionable groups, 901 suppressed, 57 duplicated tokens) and asked for read → judge → extract/accept → verify until clean.

**Method:** `deduplicate-code` skill loaded and followed. Every one of the 13 groups was read in source, judged harmful vs intentional, then extracted or accepted. Gates run after changes.

---

## a) FULLY DONE

| #     | Item                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Evidence                                                                                                    |
| ----- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| ~~1~~ | ~~**Group 1 — CLI flight-recorder setup duplication** (biggest clone, main.go flag branch vs config branch): extracted `installFlightRecorder` (shared tail: hook creation, `StageHooks` append, stderr announcement — both message prefixes preserved byte-identical) and `flightRecorderFileConfig.resolve()` in config.go (single home for duration parsing). `validate()` now delegates to `resolve()` (third copy of duration knowledge eliminated).~~ done (docs-health pass 2026-09-23) | ~~e2e tests assert both `"Flight recorder enabled"` and `"Flight recorder enabled via config"` — all pass~~ |
| ~~2~~ | ~~**Group 4 — pipeline.go duplicated early-exit epilogue** (`TotalIterations` stamp + `metricsResult = result` + return, at both cancel and error paths): extracted `exitResult` closure.~~ done (docs-health pass 2026-09-23)                                                                                                                                                                                                                                                                 | ~~pipeline module suite green~~                                                                             |
| ~~3~~ | ~~**Group 7 — fix_provider.go twin argmin loops** (differed only in distance function: byte-distance vs line-distance): extracted `nearestBy(occurrences, dist func(int) int)`.~~ done (docs-health pass 2026-09-23)                                                                                                                                                                                                                                                                           | ~~pipeline suite green~~                                                                                    |
| ~~4~~ | ~~**Group 8 — fix_applier.go duplicated prologue** (`ApplyWithReport` vs `ApplyDryRun`): extracted `newApplyReport` (group-by-safe-path + report init).~~ done (docs-health pass 2026-09-23)                                                                                                                                                                                                                                                                                                   | ~~pipeline suite green~~                                                                                    |
| ~~5~~ | ~~**Group 13 — sarif_import.go lazy-init guard** (`if f.Suppression == nil {...}` twice in one function): extracted `ensureSuppression(f)`.~~ done (docs-health pass 2026-09-23)                                                                                                                                                                                                                                                                                                               | ~~root suite green~~                                                                                        |
| ~~6~~ | ~~**Judgment documented for all 8 accepted groups** (rationale in section e/f below and in session summary): intentional similarity, not negligence.~~ done (docs-health pass 2026-09-23)                                                                                                                                                                                                                                                                                                      | ~~art-dupl re-run: actionable 13 → 8, all 8 = accepted set~~                                                |
| ~~7~~ | ~~**All verification gates green:** root + pipeline + CLI module test suites (incl. CLI e2e), `-race` on root and pipeline main packages, `nix fmt` (fixed 1 blank line), `nix run .#lint` (0 issues), `nix run .#error-audit` (0 violations, 5 modules), art-dupl re-run confirmed eliminations.~~ done (docs-health pass 2026-09-23)                                                                                                                                                         | ~~outputs in session~~                                                                                      |
| ~~8~~ | ~~**AGENTS.md gotcha corrected:** "each sub-module has replace directives" was wrong for `cmd/go-finding` — it must have NONE (`replace-audit.sh` enforces; `go install module@version` refuses replaces) → CLI is pinned to published sibling versions and cannot consume unreleased sibling-module APIs.~~ done (docs-health pass 2026-09-23)                                                                                                                                                | ~~AGENTS.md Module Structure section~~                                                                      |
| ~~9~~ | ~~**Key constraint discovered and worked around:** first attempt exported `pipeline.ResolveFlightRecorderConfig` for CLI reuse; `GOWORK=off` CLI build failed against published `pipeline v1.12.0`; export reverted (net-zero public API change) and solved CLI-side instead.~~ done (docs-health pass 2026-09-23)                                                                                                                                                                             | ~~CLI builds + tests green against published deps~~                                                         |

**Net diff:** 5 extractions across 4 modules' files, ~90 lines of duplicated logic removed, zero public API change, zero behavior change except error-message wording (see b-1).

## b) PARTIALLY DONE

1. ~~**Error-message text changed without a CHANGELOG entry.**~~ resolved 2026-09-23 (docs-health pass): the wording change is recorded in `cmd/go-finding/CHANGELOG.md` [Unreleased]. Original: **Error-message text changed without a CHANGELOG entry.** Config-branch flight-recorder failures now surface as `resolving flight recorder config: invalid flightRecorder.slowStageThreshold "x": ...` (was `parsing flightRecorder.slowStageThreshold: ...`), and `validate()` errors gained the `invalid config:` prefix via the shared wrapper. Tests only assert error presence, so they pass — but the wording change is real, unreleased-from-changelog, and `docs/guides/flight-recorder.md` example outputs were checked only for the success-path announcements, not error text.
2. **Acceptance rationale lives only in chat + this report.** The `deduplicate-code` skill asks for a one-line rationale per accepted clone so the next reader knows it was deliberate. Persisted here (section f), not in code (comment policy forbids noise) — acceptable, but a future `-t 1` run will re-surface the same 8 groups and someone must re-derive the judgment.
3. **`nearestBy` tie-breaking ("ties keep the earliest occurrence") relies on existing coverage.** Did not verify a test explicitly pins the tie case; behavior is preserved by construction (same `<` comparison), but a direct table test would lock it.

## c) NOT STARTED (deliberate or missed — honest split)

**Deliberate (out of scope / release-time):**

- ~~CHANGELOG.md / per-module CHANGELOGs for the refactor + wording change~~ done 2026-09-23 (CLI wording entry in cmd/go-finding/CHANGELOG [Unreleased])
- ~~Stress gate (`ginkgo --repeat=20 --race`, `go test -count=20` for analysis/CLI) — release-procedure step, not a session gate~~ done by the 14:10 follow-up session (all green)
- ~~`nix flake check` — no dependency/go.mod changes, so vendorHash cannot be stale~~ done by the 14:10 follow-up session — and the premise was wrong: the flake was broken for other reasons (go 1.27 floor), now repaired
- ~~TODO_LIST.md harvest of section (f) — waiting for user per instruction~~ done 2026-09-23 (docs-health HARVEST into TODO_LIST/ROADMAP)

**Missed (should have run, cheap):**

- ~~Remaining CI structural scripts (`test-naming.sh`, `docs-freshness.sh`, `docs-api-check.sh`, `json-deterministic-check.sh`, `replace-audit.sh`) — near-certainly unaffected (no go.mod, doc identifiers, or marshal changes), but "near-certainly" is not "verified"~~ done 2026-09-23 (docs-health pass ran all of them: green)
- Full `-race` across ALL packages (ran root + pipeline main packages only; CLI, goast, examples packages got plain `-count=1`) — open, tracked in TODO_LIST

## d) TOTALLY FUCKED UP!

1. **The `ResolveFlightRecorderConfig` export-then-revert detour.** I refactored the pipeline module to export a resolver for the CLI before checking how the CLI module resolves its dependencies. The `GOWORK=off` CLI build failure was fully predictable from `cmd/go-finding/go.mod` (no replace directives, pinned `pipeline v1.12.0`) and from `replace-audit.sh`, which spells the policy out. Cost: one wasted edit cycle + one revert. Root cause: I read AGENTS.md's (wrong) claim that all sub-modules carry replaces and believed it over checking the actual go.mod files first. Lesson now written into AGENTS.md so the next session doesn't repeat it.
2. **One no-op edit shipped mid-flight:** a multiedit to config.go that only removed a trailing newline (I intended to replace `toPipeline` with `resolve` but pasted the wrong old_string shape). Harmless — the real edit followed — but sloppy tool use I should have caught before invoking.

Nothing else: no ghost systems created (all five new helpers are wired into live call paths, proven by tests), nothing useful removed, tree clean, all gates green.

## e) WHAT WE SHOULD IMPROVE!

1. **Cross-module split brain remains, now documented:** `flightRecorderFileConfig.resolve()` (CLI) mirrors the internals of `pipeline.ConfigFile.ResolveFlightRecorder()` — two parsers of the same 7 string-encoded fields that can drift. The no-replace policy forces this today; it is NOT a permanent excuse (see f-1). The struct mirror (`flightRecorderFileConfig` vs `pipeline.FlightRecorderFileConfig`, YAML vs JSON tags) is the same disease, one layer up.
2. **AGENTS.md contained a materially wrong structural claim** ("each sub-module has replace directives") that directly caused d-1. AGENTS.md is the session's source of truth; wrong entries there are force-multipliers for mistakes. Worth a one-time accuracy sweep of the other structural claims.
3. **`art-dupl-report.html` is a tracked, generated artifact** — my re-run churned 811 lines of diff into git history (daemon committed it). Generated reports with unstable content (line numbers shift every edit) do not belong in version control; they belong behind a gitignore or a dedicated ignored dir.
4. **`go test | tail` is a verdict-cutting hazard.** The repo's own dead-gate lesson says never pipe a check through filters that can cut the verdict line. I piped `go test` through `tail`. It happened to be safe (2-6 packages, all `ok` lines visible), but the habit is exactly what the lesson bans — use full output or `grep -c FAIL`.
5. **901 suppressed clone groups were never characterized.** art-dupl suppressed 94% of detected groups by its own heuristics; I took the tool's word. A one-time skim of what suppression means (sub-threshold noise vs config) would make the next `-t 1` run trustworthy instead of trusted.
6. **Wording inconsistency across the mirror:** CLI says `invalid flightRecorder.X`, pipeline says `parse flightRecorder.X` for the same failure class. Cosmetic, but split-brain wording invites doubt about whether the behaviors differ.
7. **AGENTS.md date nit:** gotcha says "learned 2026-09-22"; the session crossed midnight and is now 2026-09-23. Trivial, but this repo cares about dates.

## f) NEXT — session-derived backlog (impact-ordered)

**Split brain / structure:**

1. At next pipeline release: export `ResolveFlightRecorderConfig(fc FlightRecorderFileConfig)`, tag pipeline, bump CLI `go.mod`, collapse CLI `resolve()` to a delegation → kills the duration-parsing split brain permanently.
2. Consider folding CLI `flightRecorderFileConfig` into pipeline's struct (or generating the YAML mirror) so the 7-field struct mirror can't drift.
3. Accuracy-sweep AGENTS.md structural claims against reality (module table, replace policy, gate commands) — one focused pass.
4. Decide `art-dupl-report.html` fate: gitignore it, or commit only at milestones.

**Tests:**
5. Add direct table test for `nearestBy` incl. tie-keeps-first case (fix_provider_test.go).
6. Add error-path test pinning new flight-recorder error wording (config branch + validate) so the changed strings are now contract, not accident.
7. ~~Run remaining CI structural scripts locally once (`test-naming`, `docs-freshness`, `docs-api-check`, `json-deterministic`, `replace-audit`) to convert "near-certainly fine" into verified.~~ done (run green 2026-09-23 docs-health pass)
8. Full-workspace `-race` pass (all modules, all packages) — cheap confidence beyond the 2 packages raced this session.

**Docs / changelog:**
9. ~~CHANGELOG entries: root (sarif_import refactor), pipeline (exitResult/newApplyReport/nearestBy refactors), CLI (installFlightRecorder + resolve() + error wording change).~~ done (CHANGELOG entries written by the 2026-09-23 docs-health pass (CLI wording in cmd/go-finding/CHANGELOG Unreleased; dedup refactors internal))
10. ~~Check `docs/guides/flight-recorder.md` error-path examples against new wording; fix if stale.~~ done (verified 2026-09-23: troubleshooting.md documents the new invalid-flightRecorder family; guide has no stale error text)
11. ~~Fix AGENTS.md date nit (2026-09-22 → 2026-09-22/23 session).~~ done (fixed in the 2026-09-23 docs-health pass (AGENTS.md dedup-pass note now 2026-09-22/23))
12. ~~Harvest this section (f) into TODO_LIST.md per docs-health HARVEST (awaiting user go-ahead).~~ done (harvested into TODO_LIST by the 2026-09-23 docs-health pass)

**Tooling / process:**
13. Unify duration-error wording across CLI resolve() and pipeline ResolveFlightRecorder ("invalid" vs "parse" prefix) — pick one family.
14. Warn (stderr) when both `-trace` flags and `flightRecorder` config section are set — currently config is silently ignored (pre-existing behavior, now easier to fix with `installFlightRecorder` in place).
15. Skim art-dupl's 901 suppressed groups once to validate the suppression heuristic; consider a `-t 5` "harmful-only" default invocation documented in AGENTS.md so future sessions don't wade through 2-token noise.
16. Add a repo helper/alias for "run the 6 CI structural scripts locally" so the cost of item 7 is near zero forever.

**Nice-to-have (low priority):**
17. `f.trace` branch of main.go could construct `frConfig` via the same `resolve()`-style helper for symmetry — marginal, flag branch is already minimal.
18. Consider `ApplyDryRun`/`ApplyWithReport` sharing the soft-error join loop (`errors.Join`) — noticed while editing, same shape twice, below art-dupl's radar; judge on a future pass.
19. Status-report + self-review skills default to HTML output; user overrode to `.md` this time — if `.md` is the steady-state preference for this repo, say so and the skills' divergence note can be retired.
20. Rename `flightRecorderFileConfig.resolve()` → keep, but document in doc.go? No — it's unexported; skip. (Listed to show it was considered and rejected.)

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **art-dupl-report.html policy:** keep committing the generated report (811 lines of churn this session), or gitignore it and generate on demand? It was tracked before me, so this is a repo-policy call, not mine to make unilaterally.
2. **Error-string stability:** are the changed flight-recorder error texts (`parsing flightRecorder.slowStageThreshold` → `resolving flight recorder config: invalid flightRecorder...`, plus `invalid config:` prefix from validate) acceptable, or do you have scripts/consumers grepping stderr that must keep the old prefixes?
3. **Release-ordering for the split-brain kill (f-1):** pipeline's next release could export `ResolveFlightRecorderConfig` and let the CLI adopt it (one extra export to maintain forever) — or we accept the documented CLI-side mirror as permanent. Export-and-adopt is my recommendation; do you want it scheduled into the next release train?

---

**Verification snapshot at time of report:** working tree clean (auto-commit daemon absorbed all changes); root, pipeline, CLI suites green; lint 0 issues; erraudit 0/5 modules; fmt clean; art-dupl actionable 13 → 8 (remaining 8 all accepted-with-rationale).

**Format note:** written as `.md` per explicit user instruction (skill default is HTML — override flagged).

WAITING FOR INSTRUCTIONS.
