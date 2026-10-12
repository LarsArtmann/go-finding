# Status Report — 2026-10-11 20:46 CEST

**Session scope**: "Merge/rebase all into master" → executed, then authorized-but-not-started v1.14.0 release train prep.
**Repo state at report time**: `master` = `origin/master` = `e37bf77`, working tree clean, 0 open PRs.
**Commits landed this session**: 11 (`9c675c7..e37bf77`).

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                      | Evidence                                |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| 1  | **`feat/detector-v2-coverage` merged to master** (5 commits, fast-forward, linear history) — DetectResult/DetectorV2 coverage-evidence channel + pipeline adoption of `RunDetect`                                         | `535f358` is ancestor of `e37bf77`      |
| 2  | **x/sync 0.23.0 → 0.24.0 landed** (cherry-pick of dependabot PR #45, authorship preserved) + `go work sync` verified idempotent                                                                                           | `f11ea8e`, 9 files across all 5 modules |
| 3  | **flake.nix vendorHash refreshed** for the bump (hash taken from the stale-FOD failure output; `nix build .` green after)                                                                                                 | `895ff56`                               |
| 4  | **Full local gate sweep**: all 5 modules `go test -race -count=1` green; `GOWORK=off` builds green; `go-work-sync.sh` OK; toolsdk lint 0 issues at **exact CI parity** (golangci-lint 2.14.0 locally = ci.yml pin)        | run sequentially per GOCACHE gotcha     |
| 5  | **dependabot PRs #45/#46 closed** with supersession comments; their remote branches auto-pruned afterwards (proxy confirmed satisfied)                                                                                    | `gh pr list --state open` → 0           |
| 6  | **codecov dependabot branch resolved as superseded** — master already pins `codecov-action@303a32d7… # v7.1.1` and uses `files:`; rebase conflict proved it; branch left in place (not merged, content already on master) | ci.yml:66-71                            |
| 7  | **4 pre-existing toolsdk lint findings fixed** (gochecknoglobals on `validLanguages`, SA1012 nil-ctx test nolint using the repo's established pattern, 2× wsl_v5 whitespace) — CI `lint (toolsdk)` went red→green         | `293f658`, `77687aa`; run 38163349658   |
| 8  | **README version stamp synced to v1.14.0** (module table + `finding.Version` example) — `docs-api-check.sh` verified green locally                                                                                        | `e37bf77`                               |
| 9  | **AGENTS.md gotcha recorded**: `nix run .#lint` covers the ROOT module only — the blind spot that hid the toolsdk findings behind a green local run                                                                       | `00c90df`                               |
| 10 | **Branch hygiene**: merged `feat/detector-v2-coverage` deleted (local + remote, safe `-d` since ff-merged)                                                                                                                | —                                       |
| 11 | **CI event-drop gotcha honored**: verified a ci.yml run exists for every pushed SHA (895ff56, 77687aa; docs-only pushes 00c90df/e37bf77 legitimately skip via `paths-ignore`)                                             | `gh run list` checks                    |

**CI accounting on the last code run (38163349658, commit 77687aa)**: 18/20 jobs green — incl. benchmark (20m26s), stress (7m26s), nix, coverage, govulncheck, consumer-compat, module-isolation, both OS test matrices. Remaining red: `structural-checks` (see b) and `docs-api-check` (fixed since, locally verified).

## b) PARTIALLY DONE

### 1. v1.14.0 release train — authorized, research 20% done, ZERO release actions taken

No CHANGELOG cuts, no tag pushes, no version bumps beyond what's listed in (a). Research completed before stopping:

- **Tag lattice (verified)**: `v1.14.0` (core, at `a3d22d1`, Oct 5) and `toolsdk/v1.14.0` exist. `pipeline/v1.13.0` is the latest pipeline tag (remote-verified). **Missing**: `pipeline/v1.14.0`, `analysis/v1.14.0`, `cmd/go-finding/v1.14.0`.
- **Why master is red**: `version-drift.sh` demands `cmd/go-finding/go.mod` pin `pipeline v1.14.0` (== core version.go). The pin is v1.13.0. Red since **Oct 5** — failures on 576c738, 70bfc77, 9c675c7, 895ff56, 77687aa. Pre-existing; the merge did not cause it. TODO_LIST.md:40 tracks this exact task ("Cut the v1.14.0 train … 5 tags ONE PER PUSH").
- **CRITICAL FINDING — the train cannot be cut at master HEAD**: pipeline@master calls `finding.RunDetect` in 5 files (`convenience.go`, `partial.go`, `pipeline_detect.go`, `retry.go`, `verify.go`), but `detect_result.go` is **NOT in the core v1.14.0 tag** (created by the detector-v2 merge, Oct 11). Tagging `pipeline/v1.14.0` at HEAD would publish a module that **cannot compile for consumers** (pipeline's `replace => ../` does not propagate to consumers; its go.mod would pin a core version lacking RunDetect).
- **Blind spot discovered**: `GOWORK=off go -C pipeline build` **passes** at HEAD and the CI `module-isolation` job is green — because pipeline/go.mod carries `replace github.com/larsartmann/go-finding => ../`, resolving against the local sibling tree. **Neither local GOWORK=off builds nor CI's module-isolation catch consumer-facing breakage from unpublished-API usage.** `consumer-compat` only tests published tags. This exact class of breakage is currently invisible to every gate.
- **Safe tag target candidate**: `9c675c7` (master immediately before the detector-v2 merge) — matches TODO's train definition ("everything on master since 450824e", multi-edit/Edits/EditListProvider/ResolveFlightRecorderConfig era) and contains no dependency on unpublished core APIs. **Not yet verified** that its pipeline content is complete/consistent for v1.14.0.

### 2. docs-api-check green locally, not yet re-proven in CI

Fixed by `e37bf77` and verified by running the actual gate script (`docs-api-check.sh` → "OK: README.md version stamp matches version.go"). But README pushes are `paths-ignore`d, so no CI run exists for `e37bf77`; the gate's next CI proof rides along with the next code push (e.g. the train's resync commit).

## c) NOT STARTED

1. **Reading `docs/release-procedure.md` end-to-end** — mandatory step 0 before any tag; interrupted mid-research.
2. **CHANGELOG cuts** (pipeline, analysis, root, cmd) — [Unreleased] sections still primed, untouched.
3. **FEATURES.md "(unreleased)" label flips** (5 mentions; must be 0 for _shipped_ features at tag time — release-procedure step 7).
4. **Tag pushes** (pipeline/analysis/cmd v1.14.0) — none.
5. **Resync commit** (cmd pin bump via `go get pipeline@v1.14.0`, go.sum, vendorHash) — none.
6. **Mandatory pre-tag gates**: stress (`ginkgo -r --race --repeat=20`, `go test -race -count=20`), `release-preflight.sh --bench --stress` (expect its documented false positives: 2 existing v1.14.0 tags + mid-train version-drift), manual `GOWORK=off` toolsdk build+vet (preflight skips it).
7. **Release-workflow watching** (one tag per push, each to terminal state — queueing concurrency gotcha).
8. **Post-release verification**: pkg.go.dev render, clean-dir `go get` ×5 (TODO T11 sequence).
9. **Unblocked-after-train work** (TODO_LIST): CLI flightRecorder delegate (line 41), erraudit T13/T14 migration (line 43), toolsdk/v1.16.0 train (line 46), sigstore signing restoration (line 42).
10. **Core v1.15.0 planning** — needed to ever publish RunDetect/DetectResult (detector-v2 core APIs currently unpublishable dependencies).

## d) TOTALLY FUCKED UP!

**Pre-existing (not caused this session, but discovered/named):**

1. **Master CI red for 6 days (Oct 5 → Oct 11) without an alarm** — 5 consecutive failing runs on the default branch of a public repo. Branch protection binds PRs only; admin direct pushes bypass. Nobody noticed until this session tripped over it.
2. **The v1.14.0 train was abandoned mid-cut on Oct 5** — core + toolsdk tagged, siblings + resync never done, CHANGELOGs never cut, README stamp never bumped. The repo's own gates have been screaming since.
3. **Consumer-facing compile breakage is invisible to every current gate** (the replace ⇒ ../ blind spot, see b.1). The next person to tag pipeline at HEAD ships a broken module to the public proxy — protected tags make that near-irreversible.

**My mistakes this session (fixed, but honest record):**

4. **Truncated CI verdict list → wrong report** — my `head -40` cut the job list and I reported "2 red jobs / 19 green"; `docs-api-check` was ALSO red on 895ff56. Caught one iteration later; my mid-session status claim to you was inaccurate (18/20, not 19/20). Rule broken: never truncate verdict lists.
5. **First CI-watch poll loop was broken** — jq expression error (`.conclusion` evaluated against the array) printed empty status for 30 minutes while the run failed. The failure state was only visible in the error-channel spillover. Should have smoke-tested the expression once before looping.
6. **First nolint comment exceeded golines@120** → introduced a NEW lint failure while fixing others; caught on re-lint. Should have copied the established short pattern (`options_test.go:99`) from the start.
7. **Two edit-tool mismatches** (em-dash wording drift; README edit before View) — corrected, but sloppy exact-match discipline cost round trips.
8. **Buildflow skill loaded late** — I manually ran lint and hand-pasted the vendorHash before checking buildflow coverage. (Repo turns out NOT covered — no `.buildflow.yml` — so the manual path was legitimate, but the check should have come first.)
9. **Auto-commit daemon raced me twice** — my authored commit messages for the lint fixes and flake.nix became generic "chore: auto-commit (heuristic)" commits (895ff56, 293f658, 77687aa). Content verified correct via `git show --stat` each time; history readability suffered.

## e) WHAT WE SHOULD IMPROVE!

1. **Close the consumer-compat gap at HEAD** — new CI check (or extend `consumer-compat`): temp module, `go get` each sibling at HEAD-commit pseudo-version, NO replaces, build. Would have caught the RunDetect breakage before it reached the train decision.
2. **Red-master alerting** — 6 days of default-branch red deserves a ping (GitHub notification routing on workflow failure, or required-check discipline on direct pushes).
3. **`nix run .#lint` should fan out per module** like CI does (or adopt the documented per-module invocation in a flake app). Same class as the known `.#test` root-only limitation — now documented in AGENTS.md, but the tooling still lies by omission.
4. **Encode the release train** (`scripts/release-train.sh`, TODO T11) — the prep→verify→tag→serialized-push→rerun→resync→postflight sequence has now been run manually repeatedly; this session nearly started a third manual run with a landmine in it.
5. **Version-lattice guard** — a gate that answers "if I tagged every module at HEAD right now, would consumers compile?" — cheap to script, prevents the exact trap found today.
6. **Daemon/agent commit choreography** — stage-and-commit immediately, or pre-author the message; the daemon absorbing staged batches is documented, and I still lost two messages to it.
7. **Smoke-test monitoring one-liners** (jq/awk in poll loops) with a single dry iteration before committing to a 30-minute loop.

## f) NEXT THINGS (prioritized, session-derived)

**The v1.14.0 train (decision + execution):**

1. Answer Q1 (tag target strategy, see g) — blocks everything below.
2. Read `docs/release-procedure.md` fully; reconcile with TODO_LIST:36-44.
3. Verify `analysis/*` and `cmd/go-finding/*` latest tags + diffs-since-tag (are v1.14.0 tags needed for content or lockstep only?).
4. Verify `9c675c7`'s pipeline tree is complete/consistent v1.14.0 content (TODO definition: everything since 450824e).
5. Check release.yml triggers cover `pipeline/v*`, `analysis/v*`, `cmd/go-finding/v*` (toolsdk was once missing — 4f37ec2 precedent).
6. Cut pipeline/CHANGELOG.md: multi-edit entries → `[1.14.0] - <date>`, detector-v2 stays `[Unreleased]`.
7. Cut analysis/CHANGELOG.md (if analysis ships anything).
8. Check root CHANGELOG.md state vs the already-tagged core v1.14.0 (was its section cut at tag time? backfill if missing).
9. Flip shipped-feature "(unreleased)" labels in FEATURES.md → v1.14.0 (detector-v2 entries stay unreleased).
10. Run mandatory stress gate (ginkgo --race --repeat=20 core+pipeline; go test -race -count=20 analysis+CLI).
11. Run `release-preflight.sh --bench --stress`; triage its documented false positives (existing v1.14.0 tags; mid-train version-drift).
12. Manual `GOWORK=off go -C toolsdk build ./... && vet` (preflight skips toolsdk).
13. Push tags ONE PER PUSH (pipeline/v1.14.0 first), watch each Release run to terminal state before the next.
14. Resync commit: cmd `go get pipeline@v1.14.0` + tidy + go.sum; expect `module-isolation` red on the tagged commit until resync (documented).
15. vendorHash refresh if cmd go.sum moved the FOD (nix build).
16. Push resync; watch CI — expect master **fully green** for the first time since Oct 5.
17. Postflight: pkg.go.dev renders pipeline v1.14.0; clean-dir `go get` ×5.
18. Update TODO_LIST: mark train cut done; unblock line 41 (CLI flightRecorder delegate).

**Lattice/quality debt surfaced this session:**
19. Consumer-compat-at-HEAD CI check (see e.1).
20. Version-lattice guard script (see e.5).
21. Red-master alerting (see e.2).
22. Per-module lint flake app (see e.3).
23. Plan core v1.15.0 + pipeline v1.15.0 train for detector-v2 (RunDetect/DetectResult publishability).
24. `scripts/release-train.sh` (TODO T11, see e.4).
25. erraudit T13/T14 migration (unblocks post-train; includes filing the upstream erraudit `nolint-audit` FP issue).
26. Sigstore signing restoration on next workflow-run release train (v1.13.0 shipped unsigned).
27. toolsdk/v1.16.0 train (TODO:46, fixes since 10429a5).
28. Delete the superseded local codecov branch (`git branch -D` — needs your OK, it is unmerged-but-superseded).
29. gopls SA1012 warnings persist on the two nolint'd nil-ctx test lines (editor-only noise; gopls ignores golangci nolint) — accept or switch to `//lint:ignore`.
30. Consider pinning `gh run watch`-based helpers over hand-rolled poll loops in future sessions (lesson d.5).
31. Docs: note in release-procedure that a train interrupted after core-tag leaves version-drift red by design (so the next session doesn't re-diagnose it from scratch).

## g) QUESTIONS I CANNOT ANSWER MYSELF

**Q1 (blocks the train):** pipeline/v1.14.0 tag target — **(a)** point it at `9c675c7` (pre-detector-v2; matches the TODO's "since 450824e" train definition; tag's own CHANGELOG stays one commit stale because the cut lands on master afterwards), or **(b)** bump core to v1.15.0 and ship detector-v2 in a combined bigger train (publishes RunDetect properly, but widens the release and delays master-green further)? My recommendation: (a) — smallest safe step to green, detector-v2 was never scoped into v1.14.0.

**Q2:** Should analysis and cmd get v1.14.0 tags even if their diffs since their last tags are trivial/empty (lockstep purity, preflight derives all five) — or only tag modules with real content? TODO says "5 tags"; the lattice convention is yours to call.

**Q3:** Master has bypassed-branch-protection direct pushes (admin) with required checks that only bind PRs — keep that workflow as-is (fast, operator-verifies), or tighten (e.g. alert on red master / self-merge discipline)? This session relied on the bypass three times; the 6-day-red incident is the counter-argument.

---

_Report generated by the 2026-10-11 merge session. No tags pushed, no CHANGELOGs cut, no release actions taken. `TODO_LIST.md` HARVEST from section (f) intentionally deferred per "wait for instructions"._
