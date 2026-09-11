# Session Status: erraudit Gate, Error-Management Hardening & v1.10.0 Release Repair

**Date:** 2026-09-11 06:43 CEST
**Repo:** go-finding @ master (`bbf748c` + uncommitted doc edits absorbed by auto-daemon)
**Trigger:** User pasted an erraudit report showing 79 violations (58 ERROR / 21 WARNING) and demanded "SUPERB ERROR MANAGEMENT #DDD", then "fix what makes sense", then this status report.
**Format note:** User explicitly requested `.md`; the status-report skill's canonical format is HTML — override honored per user instruction.

---

## TL;DR

The pasted 79-violation report was an **artifact of wrong tool flags** (`--enforce-samber-oops --enforce-generic-return`; oops was rejected 3/10 in a documented assessment). Chasing it properly anyway surfaced **5 real error-handling defects and 2 release-integrity breaks** — all fixed. The project now has an executable error-management gate (`scripts/error-audit.sh`, 0 violations × 5 modules, FAIL path proven). Remaining: one push (user-gated) and erraudit CI wiring (blocked on private repo).

---

## Session Arc (what actually happened)

1. Reproduced the report with fresh runs — baseline (correct flags) = 3 false positives; user's flag combo = 79. Report verified, not trusted.
2. Found the prior session's verdict doc (`docs/research/erraudit-violation-analysis.md`, 2026-08-09): oops REJECTED at 3/10 fit. A month later the same confusion recurred → root cause was **policy-as-prose, not policy-as-code**.
3. Read the full error model (`errors.go`: sentinels + `FindingError` implementing go-error-family `Coded`/`Classified`) and confirmed it is sound.
4. Audited **all 5 modules** (the user's run only covered root): found 3 CRITICAL context-loss + 2 legacy `errors.As` + ~25 intentional-but-undocumented patterns in pipeline/CLI.
5. Fixed the real defects, suppressed the intentional patterns with reasons, built the gate, then fell down the release-integrity rabbit hole (version-drift red at HEAD; missing v1.10.0 sub-module tags).

---

## a) FULLY DONE

| #  | Item                                                                                                                 | Evidence                                                                                                                                                   |
| -- | -------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | erraudit findings reproduced & verified (not trusted from paste)                                                     | baseline 3 FP; flag combo 79 — exact match with user's report                                                                                              |
| 2  | 3 baseline false positives suppressed with reasons                                                                   | `id.go:202-203` (hash.Write never errors), `errors.go:164` (AsType ok-pattern) — `//nolint:erraudit // <reason>`                                           |
| 3  | 3 CRITICAL context-loss errors FIXED                                                                                 | `pipeline/file_backup.go:91` (backup dir via `ioErrorAt`), `pipeline/pipeline.go:63,74` (`rootDir=%q`)                                                     |
| 4  | 2 genuine `errors.As` → `errors.AsType` migrations                                                                   | `cmd/go-finding/internal/detectors/{govet,staticcheck}.go`                                                                                                 |
| 5  | 24 intentional patterns documented with reasons                                                                      | 12 pipeline + 13 CLI… net 24; `nolint-audit`: **27 needed, 0 stale**                                                                                       |
| 6  | erraudit gate script with explicit verdict                                                                           | `scripts/error-audit.sh`; **FAIL path proven** (injected violation → exit 1; removed → exit 0)                                                             |
| 7  | 0 error-handling violations across all 5 modules                                                                     | root, pipeline, analysis, toolsdk, cmd/go-finding                                                                                                          |
| 8  | Pre-existing version-drift gate failure FIXED                                                                        | `cmd/go-finding/go.mod` pipeline require `v1.9.2 → v1.10.0`; GOWORK=off build OK                                                                           |
| 9  | Missing v1.10.0 sub-module alignment tags CREATED (local)                                                            | `pipeline/v1.10.0`, `analysis/v1.10.0`, `cmd/go-finding/v1.10.0` — annotated, at `bbf748c`, format matches `pipeline v1.9.2 — version alignment` precedent |
| 10 | Docs updated: research addendum §9, AGENTS.md (gate section + command), CHANGELOG [Unreleased], TODO_LIST (tag item) | all committed by daemon                                                                                                                                    |
| 11 | Full verification matrix                                                                                             | race tests ×5 modules OK · golangci-lint 0 issues ×3 touched modules · 7 structural scripts OK · `nix flake check` OK · `nix fmt` 0 changed                |
| 12 | Tool quirks discovered & documented                                                                                  | `nolint-audit ./...` silently scans nothing (use `.`); preflight exit code must not be read after `                                                        |

## b) PARTIALLY DONE

| Item                                  | Done                                                                                | Missing                                                                                                                                                             |
| ------------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| v1.10.0 sub-module release completion | go.mod aligned; 3 alignment tags created locally; preflight structural checks green | **Push** (user-gated): `master` then 3 tags in one push; then Release-workflow confirmation + proxy `go get` verification + pkg.go.dev                              |
| erraudit CI integration               | Local gate exists and is proven                                                     | `ci.yml` wiring **blocked**: erraudit repo is PRIVATE (no `go install` from public runners without credentials); token plumbing deliberately not reintroduced       |
| Error-management gate maturity        | Gates violations, prints verdicts                                                   | No `nolint-audit` phase (stale directives only caught manually); `[feature:logger]` stderr noise printed on FAIL; no version pin (`erraudit version` reports `dev`) |

## c) NOT STARTED

- Pushing anything (explicitly user-gated; never push without request).
- Consumer compatibility sweep for the 3 changed error-message formats (see f-14).
- Upstreaming the 2 erraudit detector/tooling bugs found this session (f-10, f-11).
- Stress gate (`ginkgo --repeat=20 --race`) — mandatory before next _tag push_, not required for today's changes; **not run**.
- `go-arch-lint` (arch-check job) — **not run this session**; low risk (no package-boundary changes) but honest gap in the "all gates green" claim.
- DDD error-model v2 discussions (typed `Detect` error, branded error codes) — deliberately not started (would be API churn without a driver).

## d) TOTALLY FUCKED UP

Repo-level (found & fixed this session):

1. **`version-drift.sh` was RED at HEAD since the v1.10.0 release (2026-09-10)** — `cmd/go-finding` still required `pipeline v1.9.2`. Yesterday's "all gates green" claims missed it. Fixed today.
2. **v1.10.0 shipped broken for sub-module consumers** — `pipeline/v1.10.0`, `analysis/v1.10.0`, `cmd/go-finding/v1.10.0` never tagged; proxy consumers cannot resolve those modules at v1.10.0. Mitigated locally (tags created); publish still pending.

Process-level (my own failures this session, all caught but all real):
3. **Fell into the exact pipeline-masking trap AGENTS.md warns about**: read `$?` after `… preflight | tail -20` and briefly believed preflight passed (exit 0 was `tail`'s). Caught on the next command with `> file; echo $?`.
4. **Wrote an unverified number into a doc**: addendum claimed "16 directives"; actual count 27/24. Caught by my own `rg` recount minutes later — but the discipline is count-THEN-write, not write-then-correct.
5. **Recommended a broken invocation in 3 places before testing it**: `erraudit nolint-audit ./...` silently scans nothing; I had documented that exact form in the script comment, AGENTS.md, and the research doc, then discovered the quirk when it reported "No directives found" against 27 known directives. All 3 fixed, but docs-then-verify is backwards.

## e) WHAT WE SHOULD IMPROVE

1. **Verify command forms before they touch docs** — today's `nolint-audit ./...` episode is the third instance of this class (see dead-gate lesson). Rule: any command cited in docs must appear in a transcript with its expected output first.
2. **Never read exit codes after a pipe** — even for "quick looks". `cmd > f; echo $?` or `PIPESTATUS` always.
3. **Auditing only the module you were handed is not auditing** — the 3 CRITICAL context-loss defects lived in `pipeline/`, invisible to the root-module report everyone was looking at. Multi-module repos need per-module audits as the default motion.
4. **"No action required" verdicts need an executable artifact** — the 2026-08-09 research doc was correct and still failed to prevent this month's recurrence. A doc + a gate script would have.
5. **Error messages are API** — 3 message formats changed today with zero consumer verification (f-14). Either treat messages as stable contract (golden tests) or document loudly that they are not.
6. **Tag creation should be part of release-preflight's happy path** — preflight's "tag already exists" check produced 4 false alarms the moment tags were legitimately created locally; it should distinguish local vs remote (`git ls-remote`).

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm — ROADMAP fuel, not commitments)

_Release & push (blocked on user):_

1. Push `master` (docs/CHANGELOG/TODO updates land).
2. Push the 3 alignment tags in ONE push (≤3-per-push rule) — Release workflow auto-triggers.
3. Confirm Release workflow run(s) green after tag push.
4. Verify proxy resolution: `go get github.com/larsartmann/go-finding/pipeline@v1.10.0` (GOWORK=off, no GOPRIVATE).
5. Verify pkg.go.dev renders the new module versions.
6. Decide: do sub-module tags need GitHub Releases, or tag-only is fine (procedure currently ambiguous).

_erraudit gate evolution:_
7. Make erraudit repo public (or provide fine-grained PAT) → then wire `error-audit.sh` into `ci.yml`.
8. Add a `nolint-audit` phase to `error-audit.sh` so stale directives fail the gate.
9. Filter `[feature:logger]` noise from gate FAIL output for readability.
10. File erraudit issue: `AsType` ok-pattern (`_, ok := errors.AsType[..]`) flagged as `ignored` — first return is a value, not an error (verified at `errors.go:164`).
11. File erraudit issue: `nolint-audit ./...` silently scans nothing while `.` works.
12. Pin a minimum erraudit version in `error-audit.sh` (binary reports `dev`; gates shouldn't float).
13. Investigate why the `context_loss` detector found nothing in the root module — detector gap or genuinely clean? (run `--no-suppress` root audit manually).
14. Add `--no-suppress` report-only mode to `error-audit.sh` (audit view without weakening the gate).
15. Integrate error-audit into `release-preflight.sh`'s heavy gates so releases re-prove it.
16. Teach preflight's tag-collision check to distinguish local-only tags from remote tags (`git ls-remote`) — today's 4 false FAILs.
17. Cache/speed: gate runs erraudit 5×; explore single workspace-mode invocation if the tool gains one.

_Error model / DDD:_
18. Consumer compatibility sweep: grep all Go consumers for string-matches on the 3 changed message formats ("init fix applier", "validate config", "create backup dir").
19. Decide whether `ToolAdapter.Detect` should return `*FindingError` (Parse/IO category) instead of bare wraps — the one `generic_return` site with real domain weight.
20. Manual context-loss audit of the ROOT module (the detector's CRITICALs all came from pipeline; root was audited by eye only).
21. Author `docs/guides/error-handling.md` — sentinel + `FindingError` + wrapping rules + when `errors.Is` vs `AsType` vs `CategoryOf`.
22. Add `docs/DOMAIN_LANGUAGE.md` entries for error categories/families mapping.
23. Golden tests for error message formats (catch accidental message churn; pairs with f-18).
24. Consider typed `ErrorCode` constants instead of stringly `"finding." + category`.
25. Property/fuzz test: `FindingError` round-trip through `errorfamily.Classify` for every category.
26. BDD (Ginkgo) suite for the error contract: Is/Unwrap/AsType/CategoryOf invariants.

_CI & gates hygiene:_
27. Run `go-arch-lint` (arch-check) to close today's honest gap.
28. Run the mandatory stress gate before the next tag push (`ginkgo -r --race --repeat=20` + `go test -race -count=20` split).
29. Silence golangci-lint's "unknown linters in //nolint: erraudit" warning (config allowlist) — cosmetic.
30. Verify pre-commit hook alive (`git config core.hooksPath` gotcha) — not checked this session.
31. Add `govulncheck` + arch-check to the local preflight routine docs (currently CI-only in muscle memory).

_Consumers & ecosystem:_
32. Opportunistic consumer bumps to v1.10.0 (go-linter-sdk, golangci-lint-auto-configure) — unblocks on push.
33. Re-check gomend/licenseforge BuildFlow replace-path issues (#1/#46) — still blocked upstream.
34. Update `docs/ecosystem.md` consumer table after erraudit-related changes ship.
35. Notify consumers of the error-message format changes in release notes if f-18 finds matches.

_Docs health:_
36. docs-health ANNOTATE pass: older status reports claiming "all gates green" on 2026-09-10 (version-drift was red) — annotate, don't rewrite.
37. Cross-link the erraudit policy from `docs/research/` §9 into README's development section (one line).
38. Add `error-audit.sh` to the flake devShell documentation (install hint for fresh machines).
39. Harvest this report's f-list into TODO_LIST/ROADMAP via docs-health HARVEST (this section is the input, not the tomb).

_Type-model & code (DDD v2 candidates, all API-affecting — needs a driver):_
40. Branded `ErrorCode` type (compile-time distinction from arbitrary strings).
41. Unify `ioErrorAt`-style helpers (pipeline has the pattern; root module doesn't) into a shared, documented constructor set.
42. Review `simple_fix.go` / `context.go` error paths for the same context-loss class pipeline had.
43. Audit `gotoken/` + `lockutil/` for raw panics vs `must()` discipline.
44. toolsdk: confirm Spec/provider error paths use `FindingError` (untouched this session, unverified).
45. Consider `Position` on every `FindingError` construction site (currently optional; DDD question: should it be?).

_Bigger swings:_
46. erraudit JSON output → CI artifact → violation-count trend over time.
47. Wire erraudit's LSP server into the editor setup for inline feedback.
48. Scoreboard: extend `scripts/` verdict convention (OK/FAIL line) audit to every script (dead-gate rule as code).
49. Evaluate `errors.AsType` adoption sweep repo-wide (only 2 sites migrated; are there more `errors.As` in tests?).
50. Revisit `--enforce-generic-return` for NEW API only (opt-in lint on changed lines) — captures the DDD value without churning existing API.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Push?** `master` + the 3 alignment tags are ready; pushing the tags auto-triggers the Release workflow. Do you want me to push now and monitor the run, or do you want to bundle this into the next versioned release instead?
2. **Error messages as contract:** do any of the 14 Go consumers (or your private projects) string-match go-finding error messages? I changed 3 formats today ("init fix applier (rootDir=%q)…", "validate config (rootDir=%q…)…", backup-dir error text) and cannot see consumer test suites from here. If yes, I need the list to patch compatibility or revert message shapes.
3. **erraudit CI path:** to wire the gate into GitHub Actions, the erraudit repo either goes public or a fine-grained PAT lands as an Actions secret. Which way do you want it — or should the gate stay local-only?

---

## Verification Matrix (end of session)

| Check                               | Result                                                                       |
| ----------------------------------- | ---------------------------------------------------------------------------- |
| erraudit `--type-aware` × 5 modules | 0 violations (was 3 FP + 14 + 13)                                            |
| `error-audit.sh` FAIL path          | proven (exit 1 injected / exit 0 clean)                                      |
| `nolint-audit .`                    | 27 needed, 0 stale                                                           |
| race tests × 5 modules              | all `ok`                                                                     |
| golangci-lint × 3 touched modules   | 0 issues                                                                     |
| 7 structural scripts                | all OK (version-drift fixed)                                                 |
| `nix flake check`                   | passed (after go.mod change)                                                 |
| `nix fmt`                           | 0 changed                                                                    |
| release-preflight                   | fails ONLY on expected pre-push state (unpushed commits; tags exist locally) |

_Assisted-by: Crush_
