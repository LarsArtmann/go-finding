# Status Report — Release & CI Integrity Hardening Session

**Date:** 2026-09-10 08:03 CEST (Thursday)
**Session scope:** The 6-item 🔴 HIGH "Release & CI integrity" section of TODO_LIST.md (harvested 2026-09-10 from the 2026-09-09 self-review report §f), executed end to end: concurrency group, unpushed-commits gate, version-stamp guard, post-tag failure-path self-test, worktree-hygiene gate, and the `nix flake check` follow-up.
**Method:** READ → UNDERSTAND → RESEARCH → THINK → REFLECT → Execute per item; every new gate's FAIL path was intentionally triggered and verified once (dead-gate discipline).

---

## Executive Summary

All 6 planned items were implemented, and every new gate was proven to bite by injecting real failures — not just run once on the happy path. The session's headline event: **the `nix flake check` follow-up item was NOT a no-op box-tick — the flake was actually broken** (stale `vendorHash` + half-finished `go.sum` from dependency bumps), and the run exposed a second, larger problem: **the v1.10.0 release train shipped incompletely** (core tagged, sub-module go.mod requires left at v1.9.2, sub-module v1.10.0 tags never created). That second problem is discovered but NOT fixed — it needs an owner decision, because fixing it means tagging/pushing.

3 pre-existing FAILs remain in preflight, all outside agent authority (unpushed commits, mid-cycle tag state, version drift). Nothing new was broken; `go test ./...` exits 0; the self-test proves 3 failure classes are caught.

---

## a) FULLY DONE

Each item verified complete with evidence (commit = auto-daemon batch on master).

| # | What                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Evidence                                                                                                                                                        | Files                                   |
| - | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| 1 | **`release.yml` queueing concurrency group** — `group: release`, `cancel-in-progress: false`. A manual dispatch can no longer race the tag-push auto-trigger into two concurrent GoReleaser runs (the v1.9.2 422 duplicate-asset failure is now impossible by construction). One global group (not per-ref) was chosen deliberately: a dispatch on `master` racing a `refs/tags/v*` push has different refs, so only a single group dedupes that case.           | `e567d70`; `actionlint` on release.yml + ci.yml passes                                                                                                          | `.github/workflows/release.yml`         |
| 2 | **Unpushed-commits gate in preflight** — fails when `git log @{u}..HEAD` is non-empty ("tags and CI would run against code GitHub does not have") and when the branch has no upstream.                                                                                                                                                                                                                                                                           | `e567d70`; FAIL path proven in a disposable clone with an injected commit → correct FAIL line; detached-HEAD (no upstream) path exercised in every selftest run | `scripts/release-preflight.sh`          |
| 3 | **Worktree-hygiene gate in preflight** — fails on any registered worktree outside the repo (the /tmp strays class that went unnoticed for 2 days), exempts the main worktree and the worktree preflight itself runs in (so the selftest's disposable worktree doesn't false-positive), and fails on prunable registrations.                                                                                                                                      | `e567d70`; FAIL path proven: created real stray `/tmp/preflight-stray-probe` → `FAIL: stray worktree: …` → removed                                              | `scripts/release-preflight.sh`          |
| 4 | **Version-stamp completeness guard in `docs-api-check.sh`** — requires the README `finding.Version) // "X.Y.Z"` line to exist AND match `version.go`. Closes the v1.9.1/v1.9.2 lag class. FEATURES.md intentionally left to the existing overclaim guard (its `vX.Y.Z` mentions are per-feature introduction stamps, not a current-version line — adding one would be manufacturing a requirement).                                                              | `e567d70`; three exit codes verified by injection: stale stamp → 1, missing stamp → 1, clean → 0                                                                | `scripts/docs-api-check.sh`             |
| 5 | **Self-test covers `--post-tag` failure path** — selftest grew from 2 to 3 injections; #3 bumps version.go to a tag-less version, runs preflight `--post-tag`, asserts the exact `FAIL: tag … missing (post-tag mode expects it to exist)` line. Previously only the happy path was proven (report item 37).                                                                                                                                                     | `e567d70` + `1c72e47` (header fix 1/2→1/3); full selftest run: `SELF-TEST PASS`, 3/3 injections caught                                                          | `scripts/release-preflight-selftest.sh` |
| 6 | **`nix flake check` follow-up closed — and it caught a real regression.** The indirect dep bumps (x/net 0.58→0.59, x/text 0.41→0.42, x/tools 0.49→0.50, landed via daemon commit `10bbbe6`) left every module's go.sum half-tidied (missing `/go.mod` hash lines) and stale-dated flake `vendorHash`. Fixed with `GOWORK=off go mod tidy` in all 4 modules + `vendorHash` update (2-iteration bootstrap: J26q… → 8Tfq… → GHRD…). Final run: `all checks passed!` | `28b9cf8`; `go build ./...` OK; preflight tidy-ness gate now OK for all 4 modules; `nix flake check` green                                                      | `flake.nix`, 4× `go.sum`                |
| 7 | **Docs/memory kept truthful in the same session** — CHANGELOG `[Unreleased]` section (5 Added, 1 Fixed), TODO_LIST 🔴 section removed and marked DONE with pointer, AGENTS.md (preflight description incl. new gates + selftest, CI-scripts bullet incl. stamp guard, new vendorHash gotcha), release-procedure.md (step 0 gate list, step 6 README-stamp requirement).                                                                                          | `45100e1`; `nix fmt` clean (0 changed)                                                                                                                          | 4 doc files                             |

**Verification matrix (all run this session):** selftest 3/3 PASS · docs-api-check PASS (318 identifiers + stamp) · actionlint OK · `nix fmt` 0 changed · `nix flake check` all checks passed · `go test -count=1 ./...` exit 0 (4 packages ok) · full preflight: only the 3 known pre-existing FAILs remain.

---

## b) PARTIALLY DONE

| Item                                                                         | What works                                                                                                                                                                                                                                                                                                                                                                                            | What remains                                                                                                                                                                                                                             | Blocker                                                                                                                                                                                                                                                                                              | Effort           |
| ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| **v1.10.0 release train completion** (DISCOVERED this session, pre-existing) | Root cause fully mapped: at core tag `v1.10.0`, `pipeline/go.mod` and `cmd/go-finding/go.mod` still required core **v1.9.2** (`git show v1.10.0:…` proves it), and `pipeline/v1.10.0`, `analysis/v1.10.0`, `cmd/go-finding/v1.10.0` tags were **never created** (latest: all v1.9.2). At HEAD, `cmd/go-finding` requires core v1.10.0 but **pipeline still v1.9.2** → preflight `version drift` FAIL. | The actual fix: decide complete-the-train (fix requires → tag the 3 sub-modules at v1.10.0) vs fold-into-next-release; then push tags.                                                                                                   | **Owner decision + push rights.** I did NOT bump the pipeline require because `pipeline/v1.10.0` doesn't exist on the proxy — pointing a require at an unresolvable version would trade a visible gate failure for a silent consumer breakage. Tagging/pushing is explicitly out of agent authority. | S (once decided) |
| **Preflight green on master**                                                | All agent-side gates green; the new hygiene/unpushed/stamp gates print OK in a clean environment.                                                                                                                                                                                                                                                                                                     | 3 FAILs remain, all pre-existing: (1) ~7 unpushed commits (session work) — needs `git push origin master`; (2) `tag v1.10.0 already exists` — expected mid-cycle, self-resolves at the next version bump; (3) version drift (see above). | Push is operator-only by standing rule; (2)/(3) are release-train decisions.                                                                                                                                                                                                                         | S                |
| **CI validation of this session's changes**                                  | All changes validated locally: selftest, actionlint, exit-code probes, flake check, test suite.                                                                                                                                                                                                                                                                                                       | None of it has run in GitHub Actions yet — the local preflight selftest job in ci.yml (`preflight-selftest`) will be the first remote proof of the reworked scripts.                                                                     | Unpushed commits.                                                                                                                                                                                                                                                                                    | S (push)         |

---

## c) NOT STARTED

Nothing from the assigned 6-item scope was left unstarted. Adjacent work observed but untouched (all pre-existing, all still in TODO_LIST.md):

- **Consumer ecosystem filings** (BuildFlow `TestNoLintPathExclusions`, branching-flow build failures, go-business-rules test failures, art-dupl GroupID intent) — untouched; not in this session's scope.
- **v1.10.0 consumer bumps** (go-linter-sdk, golangci-lint-auto-configure leads) — untouched; note this interacts with the incomplete train in (b): consumers can't `go get pipeline@v1.10.0` until the train completes.
- **Testing-quality backlog** (modtime-tie test case, fuzz campaign, formatter partial-write matrix, IntervalIndex coverage) — untouched.
- **Launch items** (announcement draft, public release-runbook page, v2.0 design spike) — untouched.

This section is honest rather than empty: the session deliberately did not expand scope beyond the 6 assigned items plus the two discovered blockers.

---

## d) TOTALLY FUCKED UP

Radical honesty, most damaging first.

1. **The v1.10.0 core tag shipped with the exact defect the preflight was built to prevent.** `git show v1.10.0:cmd/go-finding/go.mod` shows requires `go-finding v1.9.2` + `pipeline v1.9.2` at the tagged commit, and `pipeline/go.mod` required core v1.9.2. This is the v1.7.0-incident class — the one `release-preflight.sh` exists for — meaning **the v1.10.0 core tag was cut without a green preflight** (or before these gates were wired into the operator's actual path). Severity: consumers resolving `cmd/go-finding@latest-by-tag` get pipeline v1.9.2. Mitigation: only core `v1.10.0` + `toolsdk/v1.10.0` were tagged, so no broken `cmd/go-finding/v1.10.0` module release exists on the proxy; the damage is contained to "train incomplete", not "broken artifacts published". Owner decision required (see b).

2. **The daemon's dependency bump committed a half-finished state to master.** Commit `10bbbe6` updated `h1:` lines in 4 go.sums but omitted the required `/go.mod` lines, leaving all 4 modules failing `go mod tidy -diff` — and the flake's vendorHash stale. Master was un-buildable-by-flake between that commit and this session's fix. Root cause unknown: something ran a partial tidy (writer unidentified — not this session, not manually). Mitigation landed: sums completed, vendorHash fixed, `nix flake check` green. Residual risk: **the writer is unidentified** — if it's an automated process, it can regress this again.

3. **Three releases in a row hit the same "manual dispatch races auto-trigger" wall before the fix landed** (v1.9.2's 422 duplicate-asset loser was documented in TODO_LIST while `.github/workflows/release.yml` demonstrably had no concurrency block). The gap was known, written down, and left unfixed through two more releases. Now fixed — but the pattern "known incident → TODO row → not fixed before next release" is the fucked-up part, not the YAML.

4. **Self-honesty item (agent): my first selftest edit contained a leftover draft line** (a redundant `MISSING_TAG` assignment) that I caught and cleaned only on re-read of the diff — the first `multiedit` shipped sloppy code that happened to be harmless because line 2 overwrote line 1. If the assignment order had been reversed, selftest 3 would have asserted the wrong tag string. Proof-by-running caught it (selftest 3/3 green) only after the cleanup.

---

## e) WHAT WE SHOULD IMPROVE

1. **Gate the tag push, not just the tag decision.** The v1.10.0 drift shipped because preflight is advisory to the operator. A server-side guard (CI job that runs `version-drift.sh` + tag-vs-go.mod check ON TAG PUSH and fails the Release workflow's test job) would have stopped the broken train mechanically. Preflight is necessary but not sufficient; the enforcement point must be where the tag lands, not where the operator sits.
2. **Identify the anonymous dependency-bump writer.** Something on this machine ran a partial `go mod tidy`/`go get` and the daemon committed the result. Until identified, `vendorHash` can silently regress again. Cheap first step: `go.mod`/`go.sum` mtime correlation with daemon commit times, or a shell-history check.
3. **Stop batching multi-topic daemon commits.** `10bbbe6` (dep bumps) and `e567d70` (my gate work) landed 12 seconds apart as sibling "heuristic" commits; blame for the go.sum breakage was only attributable by diffing. Content-scoped commit messages (or a daemon that refuses go.mod/go.sum without a `go mod tidy -diff` pre-check) would make regressions attributable in seconds.
4. **Make "flaky gate" verification a written step, not a virtue.** Every FAIL path this session was verified by _actually breaking something in a disposable context_ (clone probe, stray worktree, sed-injected stamps). The dead-gate lesson (2026-09-08) already mandates this; institutionalize it as a checklist row in release-procedure step 0 so it survives sessions.
5. **Preflight tag-exists check should distinguish mid-cycle from post-release.** The `FAIL: tag v1.10.0 already exists` line is noise 100% of the time except in the ~10-minute window before tagging. An explicit `--mid-cycle` acknowledgment (or auto-detect "tag exists AND CHANGELOG has no entry beyond it") would keep the signal-to-noise of the final verdict honest — 3 FAILs where 1 is actionable trains operators to skim.
6. **README version stamp is a manual mirror of version.go.** The new guard catches the lag, but the root cause is manual duplication. A tiny generator (or a `go generate` directive writing the README snippet from version.go) would make the divergence impossible rather than detected.
7. **Exit-code hygiene in shell probes.** Twice this session a pipeline (`cmd | tail; echo $?`) masked the real exit code (dead-gate lesson §3 again). Consider a tiny `scripts/lib/assert-exit.sh` helper so verification snippets can't repeat the mistake.

---

## f) Up to 50 things we should get done next

Ranked by impact; Effort: S <30min, M 30min–2hr, L >2hr. _(Brainstorm, not commitment — HARVEST should route: items 1–12 → TODO_LIST, the rest → ROADMAP unless promoted.)_

**Release integrity (direct follow-ups from this session)**

| #  | Task                                                                                                                                | Impact | Effort | Category      |
| -- | ----------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Decide v1.10.0 train completion: complete-the-train vs fold-into-next, then execute it                                              | High   | S      | Release       |
| 2  | Push master (~7 commits) so CI validates this session's gate work                                                                   | High   | S      | Release       |
| 3  | Watch the first CI run after push; verify `preflight-selftest` job passes with the 3-injection selftest                             | High   | S      | Release       |
| 4  | Add CI-side tag-push guard: Release workflow test job runs `version-drift.sh` so a drifted tag fails before GoReleaser              | High   | S      | Quality       |
| 5  | Identify the process that ran the partial go.sum bump (mtime/history correlation); neutralize or document it                        | High   | M      | Quality       |
| 6  | Bump `cmd/go-finding`'s pipeline require to match the decided train version and re-run preflight to green                           | High   | S      | Release       |
| 7  | Add `--mid-cycle` acknowledgment (or auto-detection) to preflight's tag-exists gate to cut verdict noise                            | Med    | S      | Quality       |
| 8  | Extend selftest with a 4th injection: stray worktree FAIL path (currently proven only by ad-hoc probe, not wired into the selftest) | Med    | S      | Quality       |
| 9  | Extend selftest with a 5th injection: unpushed-commits FAIL path (needs a fake upstream inside the disposable worktree)             | Med    | M      | Quality       |
| 10 | Add docs-api-check stamp-guard injections to the selftest family (lag + missing) so all five gates have permanent regression proof  | Med    | M      | Quality       |
| 11 | Generate the README `finding.Version` stamp from version.go (go:generate) so the guard's failure class becomes unrepresentable      | Med    | S      | Cleanup       |
| 12 | Document the concurrency-group behavior in release-procedure (what queues, what a manual dispatch now does)                         | Low    | S      | Documentation |

**Consumer ecosystem (pre-existing backlog, untouched this session)**

| #  | Task                                                                                            | Impact | Effort | Category |
| -- | ----------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 13 | File BuildFlow issue: `TestNoLintPathExclusions` policy failure (evidence in docs/ecosystem.md) | Med    | S      | Bug      |
| 14 | Capture fresh output and file branching-flow issue (pkg/errors + pkg/fs build failures)         | Med    | M      | Bug      |
| 15 | Capture fresh output and file go-business-rules issue (pre-existing test failures)              | Med    | M      | Bug      |
| 16 | File art-dupl GroupID integration intent (GAP-2 context, feedback doc exists)                   | Med    | S      | Feature  |
| 17 | Opportunistic consumer bumps to v1.10.x — blocked until item 1 resolves the train               | Med    | M      | Feature  |
| 18 | Consumer compatibility matrix as CI job (repo public, unblocked)                                | Med    | M      | Quality  |
| 19 | Re-check library-policy after hook issue #74 resolves                                           | Low    | S      | Quality  |

**Testing & code quality**

| #  | Task                                                                                                | Impact | Effort | Category |
| -- | --------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 20 | Fix same-second modtime-tie nondeterminism in concurrent-rotation test (assert count, not identity) | Low    | S      | Bug      |
| 21 | Run a real fuzz campaign (`-fuzz` 30s+) for `FuzzPruneSnapshotsHostileDir` beyond seeds             | Low    | S      | Quality  |
| 22 | Complete formatter partial-write matrix (`FormatTextRich` failAt=1 main-write case)                 | Low    | S      | Quality  |
| 23 | Cover IntervalIndex/correlate 24 uncovered stmts (98.3% → 99%) or consciously accept                | Low    | M      | Quality  |
| 24 | Confirm gomega/ginkgo Dependabot bumps stay merged next cycle                                       | Low    | S      | Quality  |
| 25 | Decide pin-vs-dependabot policy for pinned `actions/*` SHAs (sbom bump was manual)                  | Low    | S      | Quality  |

**Product & launch**

| #  | Task                                                                                              | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 26 | Publish launch announcement (draft exists) once CI green on public repo is confirmed              | Med    | S      | Feature       |
| 27 | Submit to Awesome Go                                                                              | Med    | S      | Feature       |
| 28 | Public CONTRIBUTING / release-runbook page (internal release-procedure is not public-facing)      | Low    | M      | Documentation |
| 29 | v2.0 design spike (Position sentinel, FixStrategy union, TagSet, sub-structs) — parked in ROADMAP | Low    | L      | Feature       |

**Session-specific process debts (from section e, made actionable)**

| #  | Task                                                                                                                                                                  | Impact | Effort | Category      |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 30 | Add "verify one FAIL path per new gate" as an explicit checklist row in release-procedure step 0                                                                      | Med    | S      | Documentation |
| 31 | Daemon policy: refuse/note commits touching go.mod/go.sum unless `go mod tidy -diff` is clean                                                                         | Med    | M      | Quality       |
| 32 | Post-tag mode: also assert all 4 tags exist (currently only checks the 4 target tags list — toolsdk excluded by design; document that exclusion in the script header) | Low    | S      | Documentation |
| 33 | Preflight: consider `git fetch` before the unpushed-commits check so `@{u}..HEAD` reflects GitHub truth, not a stale remote-tracking ref                              | Med    | S      | Quality       |
| 34 | Add a `docs/status/` README or index noting reports are point-in-time and ANNOTATE-mode rules apply                                                                   | Low    | S      | Documentation |
| 35 | Consider moving the 3 selftest injections + 2 planned ones into a table-driven loop inside the selftest script (less copy-paste as it grows)                          | Low    | S      | Cleanup       |
| 36 | `nix flake check` in CI as a scheduled (weekly) job so vendorHash rot is caught without a local run                                                                   | Med    | M      | Quality       |
| 37 | Version-drift script: also check `toolsdk/go.mod` references (currently unchecked because toolsdk has no version.go — but its core require can still drift)           | Med    | S      | Bug           |
| 38 | Add `actionlint` to preflight or CI so workflow YAML edits are gated like shell scripts                                                                               | Low    | S      | Quality       |
| 39 | Sweep docs/status/ for pre-public-repo reports still claiming "CI billing dead" and annotate them (ANNOTATE mode)                                                     | Low    | S      | Documentation |
| 40 | README install section: verify `go install ...@latest` instructions resolve against the actual latest published module tags after the train decision                  | Med    | S      | Documentation |

**Hardening / forward-looking**

| #  | Task                                                                                                                                                                        | Impact | Effort | Category      |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 41 | Preflight: add `git fetch --tags` + verify target tags don't exist on the REMOTE (local tag check misses a tag pushed from another machine)                                 | Med    | S      | Quality       |
| 42 | Release workflow: assert the tag being released matches `version.go` inside the workflow itself (defense-in-depth for item 4)                                               | Med    | S      | Quality       |
| 43 | Consider GoReleaser's `--fail-fast` + release asset pre-flight dry-run in a nightly job to catch config rot (422-class failures) before a real release                      | Low    | M      | Quality       |
| 44 | Add release-notes generation check: CHANGELOG `[Unreleased]` must be retitled to the version at tag time (currently advisory in release-procedure step 7 via FEATURES walk) | Low    | S      | Quality       |
| 45 | Selftest: run post-tag mode's version-check dependency explicitly (grep its OK line too, not just the FAIL) to prove mode-specific gates actually execute                   | Low    | S      | Quality       |
| 46 | Document in AGENTS.md that preflight assumes it runs from a normal clone (no upstream → FAIL) and why the selftest tolerates that                                           | Low    | S      | Documentation |
| 47 | json/v2 stabilization watch remains open (Go 1.27) — carry forward in ROADMAP                                                                                               | Low    | —      | Feature       |
| 48 | Evaluate `git worktree prune --dry-run` output check placement: currently only runs when zero strays found; make both branches independent                                  | Low    | S      | Cleanup       |
| 49 | Add shellcheck to preflight/CI for scripts/*.sh (7 scripts and growing)                                                                                                     | Med    | S      | Quality       |
| 50 | Consider signing release tags (cosign already in the pipeline for artifacts; tags themselves unsigned)                                                                      | Low    | M      | Quality       |

---

## g) Three questions I cannot figure out myself

1. **How should the v1.10.0 train be completed?** Complete-the-train (fix the two requires, tag `pipeline/v1.10.0` + `analysis/v1.10.0` + `cmd/go-finding/v1.10.0` retroactively against the already-pushed core tag) or fold-into-next-release (bump everything at v1.10.1/v1.11.0)? I mapped every fact needed for both options; the choice depends on whether you consider `cmd/go-finding`'s published v1.9.2 tag "good enough for consumers until next release" — a product call, not a code question. (What I tried: proxy state via tag listing, `git show <tag>:…` for all four go.mods, CHANGELOG claims "v1.10.0 + toolsdk/v1.10.0 released".)

2. **What ran the partial `go mod tidy`/`go get` that daemon-commit `10bbbe6` absorbed?** If it's a scheduled automation on this machine, the vendorHash regression WILL recur; if it was a human/other-agent one-off, it won't. I cannot see other sessions' processes or shell history. Options you could answer instantly: "that was Dependabot-on-my-machine / another Crush session / me, manually."

3. **Is preflight allowed to mutate remote state (run `git fetch`) as part of its checks?** The unpushed-commits gate compares against a possibly stale remote-tracking ref; making it truthful requires a fetch, which touches the network and takes seconds. I chose the conservative local-only check; if you consider fetch acceptable in a "preflight", item 33 becomes a 5-minute fix — otherwise it stays local-only by design.

---

## Appendix — Session timeline (evidence chain)

| Time (approx) | Event                                                                                                           |
| ------------- | --------------------------------------------------------------------------------------------------------------- |
| 07:39         | Read release.yml, release-preflight.sh, docs-api-check.sh, selftest, version-check.sh, TODO_LIST, ci.yml wiring |
| 07:44         | Implemented items 1–5 (release.yml, preflight ×2 gates, stamp guard, selftest 3rd injection)                    |
| 07:47         | Daemon `e567d70` absorbs gate work; `10bbbe6` (dep bumps, not mine) lands 12s later                             |
| 07:48         | First `nix flake check` → **vendorHash mismatch caught**; probe runs prove unpushed + stray-worktree FAIL paths |
| 07:51         | Full preflight run → surfaces drift + half-tidy (pre-existing); `go mod tidy` ×4; vendorHash bootstrap ×2       |
| 07:58         | `nix flake check` → all checks passed; stamp-guard exit codes verified (1/1/0); actionlint OK                   |
| 08:00         | Selftest 3/3 PASS; `go test ./...` exit 0; docs updated; `nix fmt` clean                                        |
| 08:03         | Final preflight: only the 3 known pre-existing FAILs remain → this report                                       |

**Commits this session (daemon-batched):** `e567d70` (gates), `10bbbe6` (dep bumps — not authored here), `28b9cf8` (tidy + vendorHash), `1c72e47` (selftest header fix), `45100e1` (docs).
