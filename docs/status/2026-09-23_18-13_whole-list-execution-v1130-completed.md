# Status Report — 2026-09-23 18:13 · Whole-List Execution Round 2 (Release Completed, Gates, Follow-ups)

> **Format note:** `.md` per the standing user mandate — this is an override of the
> status-report skill's HTML canonical, flagged here as required. `date` taken at
> step 0 (18:08, re-stamped 18:13).

## Context

The session directive was "GET SHIT DONE — the WHOLE TODO LIST," which resolves the
three owner questions asked in the three prior status reports (T2 fork, #36 timing,
mid-cycle CI policy). Execution order followed the Pareto plan's dependency chain:
release truth first, then gates, then the long tail.

## a) FULLY DONE

1. **T2: the v1.13.0 train COMPLETED.** Decisive fact: proxy.golang.org already served
   core + toolsdk v1.13.0 (fetched 09-22), so folding into v1.14.0 was impossible
   without a retract — completion was the only coherent path. Executed:
   - Core GitHub Release published via **local GoReleaser 2.17.1** in a clean
     worktree at 450824e: 17 assets (5 platform archives, 8 deb/rpm/apk packages,
     5 SBOMs, checksums). `--skip=sign,nix` (cosign keyless is CI-only; the config's
     nix pipe has a latent `{{binary}}` template bug — see d2). Local runs needed
     CI parity: `GITHUB_TOKEN` present + syft from nixpkgs.
   - Sibling tags `pipeline/v1.13.0`, `analysis/v1.13.0`, `cmd/go-finding/v1.13.0`
     created **annotated + SSH-signed** at 450824e, pushed ONE PER PUSH; their
     triggered Release workflows failed fast on the frozen go-1.26 pin (expected),
     and the three library releases were created manually in the workflow's exact
     format (title/body/prerelease).
   - Resync: CLI pipeline require → v1.13.0, go.sum refreshed, flake vendorHash
     updated. **version-drift GREEN — the mid-cycle red is gone, which dissolves
     owner question #3 (no --mid-cycle mode needed).**
   - Verification: proxy serves all 5 modules @ v1.13.0 (`.info` checked);
     `scripts/consumer-compat-check.sh v1.13.0` resolves, builds, and RUNS all 5
     modules as an external consumer, including `go install cmd/go-finding@v1.13.0`
     printing `1.13.0`. GitHub Releases: Latest = v1.13.0 + 3 library pre-releases
     - toolsdk.
2. **T3: issue #36 CLOSED** as completed, closing comment drafted via github-voice
   (checker: 0 FAIL, 0 WARN), evidence: FromDiagnostic fills `Edits` with all
   TextEdits, ToDiagnostic re-emits, SARIF artifactChanges, per-finding accounting;
   ships v1.14.0.
3. **T5:** release-procedure.md rewritten — ONE tag per push with serialized run
   watching; the ≤3-batch rule documented as superseded (both incident classes);
   post-train resync step documented. AGENTS.md gotcha updated.
4. **T6:** release-preflight expects all **5** tags; selftest has a 4th injection
   class (incomplete tag set = the exact v1.13.0 failure) — **4/4 PASS**, and CI
   verified the fix after the first run exposed a missing-committer-identity bug
   (lightweight `update-ref` refs now).
5. **T7:** `scripts/changelog-drift.sh` — every release tag must have a module
   CHANGELOG section. It immediately caught **3 real missing v1.13.0 sections**
   (backfilled). FAIL path proven (renamed section → exit 1). Wired into ci.yml
   structural-checks + preflight. Ancient tags are an explicit frozen allowlist.
6. **T8:** docs-api-check now sweeps ALL `vX.Y.Z` literals in README — table-cell
   stamps must equal version.go, no future claims in prose, URLs exempt. Both
   failure modes injected and proven exit 1. Caught a live overclaim:
   README "Multi-edit fixes (v1.14.0)" → "(unreleased)".
7. **T9:** CI `nix` job (`nix flake check` + `nix run .#lint`, Determinate
   installer pinned v23, 45min budget). **First full run: SUCCESS.**
8. **T10:** error-audit.sh hardened — `erraudit version` printed per run,
   `[feature:logger]` noise filtered from FAIL output, FAIL path proven (exit 1,
   clean verdict), and **nolint-audit demoted to opt-in advisory** after the
   session's key tooling finding (see d3): stripping its 24 "safe to remove"
   suggestions turned the violation gate RED at every site. AGENTS gotcha amended.
9. **T13:** Edits ↔ BeforeCode/AfterCode consistency DECIDED: intentionally loose
   (display pair may be truncated; strict equality would false-reject legit
   producers). Pinned by a validate test, documented in ADR #19 + the field doc.
10. **T14:** FixEngine semantics PINNED by tests + documented in fix-engine.md:
    Applied follows application order (not input order), and a partially-conflicted
    multi-edit finding appears in Applied AND Conflicts with an Applied outcome.
    Test offsets deliberately tie-free (`sortEditsDescending` is unstable).
11. **T15:** `BenchmarkFixEngine_EditListProvider_{1,10,100,1000}` + LineCol
    variants (offset path vs shared line-index path); bench-check.sh's gawk
    "unknown escape" warning fixed (stray `\` before a multibyte box char); new
    benches noted as ungated in benchmarks/README (no baseline regen without a
    failed-gate justification).
12. **T16:** `FuzzTextEditValidate` (core) + `FuzzEditListProvider` (pipeline,
    resolution + apply invariants) — **45s campaigns each, 5.5M execs total, zero
    findings.**
13. **T18 — all four filings resolved WITHOUT filing junk** (verify-before-filing
    gates did their job): BuildFlow `TestNoLintPathExclusions` PASSES on current
    HEAD (claim stale); branching-flow pkg/errors+pkg/fs BUILD (claim stale);
    go-business-rules failure root-caused (go1.26 directive vs Go 1.27 stdversion
    vet) and **FIXED DIRECTLY** (go.mod → 1.27.0, tests green, pushed 431d8cd);
    art-dupl GAP-2 was ALREADY filed as art-dupl#1.
14. **T19:** `consumer-compat` CI job added (runs the proxy-resolution gate on
    every push) and added to branch protection's required checks.
15. **T20:** go-linter-sdk bumped v1.10.0 → v1.13.0 (go.work aligned to 1.27.1;
    5 packages race-green; pushed 6f3231b). golangci-lint-auto-configure was
    already at v1.13.0. ecosystem.md gained a dated delta-sweep section.
16. **T21:** launch hardening — master branch protection (17 required checks,
    linear history, no force-push/deletion, **enforce_admins=false** — see d6),
    `protect-release-tags` ruleset (deletion + non-fast-forward blocked on all 5
    tag patterns), **private vulnerability reporting ENABLED** (SECURITY.md's
    promised intake now actually exists), boundary READMEs for docs/status,
    docs/planning, docs/reviews.
17. **T22 (draft):** launch announcement finalized for current reality (status
    header, short version at v1.13.0, checklist items checked off).
18. **T23:** finding-groups.md refreshed (groupId vs artifactChanges orthogonality);
    DOMAIN_LANGUAGE gained TextEdit/Edits/Insertion/Span terms; fix-providers
    guide gained the EditListProvider row + chain order; CLI edit-list docs exist
    in the guide's chain section; `pipeline/examples/multi-edit` written (3-edit
    fix, offsets against the original snapshot) and wired into the compile gate;
    GenerateID documents that fix data is deliberately excluded from identity.
19. **T23.6:** `Preview()` multi-hunk support — edit-list findings render one
    -/+ pair per edit instead of empty output; display-pair findings unchanged;
    tests pin all three shapes.
20. **T23.7:** Conflict dedup — a multi-edit finding whose several edits all
    conflict now merges into ONE Conflict entry with accumulated ConflictsWith;
    test pins 2 conflicts → 1 entry with [late, early].
21. **T24:** testing bundle — root coverage **99.0%** (IntervalIndex/correlate:
    zero-line guard, same-tool guard, maxCorrelations cap pinned at exactly
    10000, empty-index + comparator contracts, direct overlapLength tests; two
    provably unreachable guards documented as accepted); rotation test stamps
    strictly increasing modtimes (modtime-tie nondeterminism dead); hostile-dir
    fuzz campaign 40s clean (258K execs); FormatTextRich failAt=1 case added;
    full-workspace `-race` green in all 5 modules; nearestBy tie-break pinned;
    CLI `-version` verified (prints the core version it was built against).
22. **T26: GOEXPERIMENT=jsonv2 DROPPED** — Go 1.27 release notes confirm
    encoding/json/v2 is GA and the default encoding/json backend. Verified by
    build+vet+race+lint+flake-check with the variable explicitly unset. Swept
    from flake.nix, both workflows, 4 scripts, README (prerequisite is now just
    "Go 1.27+"), CONTRIBUTING, templates, troubleshooting + migration guides,
    RELEASE_CRITERIA, AGENTS (2 gotchas), ROADMAP (watch CLOSED), CHANGELOG.
23. **F25.1:** `/go/` gitignored (the trash itself stays operator-only).
24. **f/26:** local/CI golangci-lint divergence root-caused (CI pinned v2.13.1,
    nix has v2.13.2) and aligned; AGENTS gotcha added so buildflow bumps keep
    the action pin in sync.
25. **f/27:** link-check logic extracted to `scripts/link-check.sh` (CI calls it;
    runs locally, exit 0 verified).
26. **f/28:** `scripts/pre-push-verify.sh` — fmt, per-module race tests, 11
    structural gates, lint; explicit verdict lines; PASS verified end-to-end.
27. Docs truth: ROADMAP open question (T2) resolved with evidence; plan
    annotation log carries the full execution record; TODO_LIST rebuilt 100%
    open-work (~17 rows, all verified).

## b) PARTIALLY DONE

1. **T2 residual: pkg.go.dev render check for v1.13.0** — the `/fetch` endpoint
   returned 404 (likely needs POST/indexing lag); consumer-compat's successful
   build+run of all 5 modules is strong proxy evidence, but the pkg.go.dev pages
   themselves were not visually confirmed this round.
2. **T17:** pipeline now exports `ResolveFlightRecorderConfig` (delegation inside
   pipeline, wording pinned); the CLI-side `resolve()` delegation is PREPARED but
   cannot compile until pipeline/v1.14.0 exists (no-replace policy) — correctly
   blocked, row updated in TODO_LIST.
3. **T22:** announcement draft is final; the Awesome Go submission itself needs an
   awesome-go fork + PR (owner-gated).
4. **T25:** F25.1 done; the stray `./go` trash remains an operator call (the
   directory did not exist on this pass — possibly already trashed).
5. **Final CI verdict:** the newest run (35886575285, on d994d75) was still
   in_progress at report time. Verified so far: run 35881376042 was all-SUCCESS
   including the FIRST `nix` green and structural-checks with the new drift gate
   (only `benchmark` cancelled by the newer run — normal cancel-in-progress).
   The consumer-compat job's first verdict lands in the pending run.

## c) NOT STARTED

1. **T11:** `scripts/release-train.sh` + post-release verification script
   (the manual sequence has now run three times — it is the top remaining
   automation debt).
2. **T27:** forward design — v2.0 spike agenda, sampling NO-GO calendar date,
   public CONTRIBUTING/release-runbook page, brew/nix distribution proposals,
   v1.5–v1.8 release backfill decision, F27.4b gomend/licenseforge watch note.
3. **T12:** erraudit T13/T14 code migration — blocked on v1.14.0 (Edits bridge
   is not in v1.13.0; completing v1.13.0 did NOT unblock this, contrary to what
   an earlier plan line implied).
4. Awesome Go PR (see b3) and the homebrew tap creation (owner call, ROADMAP).

## d) TOTALLY FUCKED UP — brutal self-critique (what I got wrong)

1. **Stripped 24 nolint directives on an unverified oracle.** nolint-audit said
   "safe to remove" and I removed all 24 in one sweep BEFORE running the
   violation gate — the gate went red at every site. The dead-gate rule says
   prove a change's effect; I proved it the expensive way. Recovery: git
   archaeology + rewrite of error-audit.sh around the (now documented) finding
   that nolint-audit's staleness analysis is go/parser-only and false-reports
   type-aware-suppressing directives. Filed as an upstream erraudit insight,
   not yet filed as an issue.
2. **Two restore illusions caused by the auto-commit daemon.** (a) After the
   strip, my `git restore` was a NO-OP — the daemon had already committed the
   stripped state, so "clean tree" lied and I tested erraudit behavior against
   the wrong tree for several steps, briefly concluding the violation run
   ignored nolint directives entirely (false). (b) Earlier, a sed one-liner
   CORRUPTED analysis/CHANGELOG.md (repeated garbage fragments), the daemon
   committed the corruption, and my restore-from-HEAD restored the corrupted
   version. Root lesson, now obvious: after any restore, VERIFY CONTENT (grep
   the expected lines), never trust `git status`; and never do multi-line text
   surgery through shell quoting — use structured edits.
3. **GoReleaser trial-and-error cost.** Three snapshot attempts before the
   failure set was understood: missing syft (CI injects it via an action), then
   the nix-pipe `{{binary}}` template bug that CI's skip_upload path happens to
   mask. I should have read the config's nix section against the local
   goreleaser version BEFORE the first run. The `{{binary}}` bug is now a
   TODO_LIST row (it will bite the next local dry-run or CI goreleaser bump).
4. **Regex-swept YAML.** Removing `GOEXPERIMENT:` lines with regex left empty
   `env:` blocks and wrong indentation in release.yml; actionlint caught it, but
   structured YAML editing was the correct first move, not the recovery.
5. **Self-lockout by my own branch protection.** I enabled `enforce_admins:
   true` with 16 required checks and the next push bounced ("16 of 16 required
   status checks are expected") — required checks + admin enforcement make
   direct pushes mathematically impossible on a fresh commit. Fixed to
   `enforce_admins: false` (matching this repo's direct-push reality) and
   documented in AGENTS, but I should have reasoned about the push model BEFORE
   applying.
6. **Daemon-adjacent commit blur.** Another session's toolsdk `ModuleFanOut`
   change (legitimate, tested) rode inside my pushed selftest-fix commit, so
   that commit's message under-describes its content. Discovered post-push;
   fixing would need a force-push (banned). The change itself is judged sound
   and kept; the plan log records the mismatch.
7. **Plan premise error I should have caught in review:** F2.6 said "verify
   pkg.go.dev renders TextEdit/Edits" for the v1.13.0 completion — impossible,
   because the Edits feature landed AFTER the v1.13.0 tag. I derived the tag
   contents early but did not re-validate every plan line against them before
   executing.
8. **Disk exhaustion ignored until it bit.** `/mnt/buildcache` hit 100% (167G
   build cache) mid-session and broke builds; `go clean -cache` recovered, but
   checking `df` before a fuzz+bench-heavy session should be a reflex now
   (noted in AGENTS).
9. **Scope honesty:** T27 forward-design items and T11 release-train automation
   were in the plan's tail and remain untouched — the "whole list" is done
   except the parts that are genuinely new multi-hour builds or owner-gated.
   TODO_LIST carries them; I am not pretending otherwise.

## e) WHAT WE SHOULD IMPROVE

1. **Verify-after-restore discipline** (d2): make "content grep" the mandatory
   step after every `git restore` in daemon-committing repos; `git status`
   clean is not proof of anything.
2. **Prove tool oracles before mass edits** (d1): any "safe to remove"
   suggestion gets a ONE-site trial (remove one, run the gate, observe, restore)
   before bulk action — now codified in AGENTS + error-audit.sh's header.
3. **Structured edits over shell text-surgery** for anything multi-line (d2/d4).
4. **Push-model check before protection changes** (d5): list the repo's actual
   merge paths (direct pushes vs PRs) and pick `enforce_admins` accordingly.
5. **Pre-flight `df` check** in pre-push-verify.sh (cheap; the cache failure
   mode produced misleading "no space left" errors far from the cause).
6. **Prefer fixing tiny root causes over filing issues** (go-business-rules):
   the planned issue became a 1-line fix + push; the verification gates saved
   four junk issues and produced two real fixes instead.

## f) Top #25 next (impact-sorted; first 7 are the v1.14.0 cluster)

| #  | Task                                                                                                            | Why                                                           | Effort |
| -- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | ------ |
| 1  | Watch pending CI run 35886575285 to verdict; confirm consumer-compat's first green                              | last unverified gate                                          | 0      |
| 2  | Decide + cut the v1.14.0 train (CHANGELOG cuts, labels flip, preflight `--bench --stress`, 5 tags one-per-push) | ships Edits, closes the #36 loop publicly                     | L      |
| 3  | Post-train resync: CLI requires → v1.14.0, go.sum, vendorHash, version-drift green                              | same procedure as today's resync                              | S      |
| 4  | CLI delegation of `ResolveFlightRecorderConfig` (delete the mirrored parser)                                    | prepared; unblocked the moment pipeline/v1.14.0 exists        | S      |
| 5  | erraudit T13/T14 migration onto tagged APIs + upstream `nolint-audit` false-stale issue                         | consumer adoption of Edits                                    | M      |
| 6  | Fix `.goreleaser.yml` nix-pipe `{{binary}}` template bug                                                        | unblocks local dry-runs; latent CI hazard on goreleaser bumps | S      |
| 7  | Restore cosign signing on the next workflow-run train (v1.13.0 shipped unsigned)                                | supply-chain parity with v1.12.0                              | S      |
| 8  | `scripts/release-train.sh` (T11) + post-release verifier                                                        | the manual train has run 3×                                   | L      |
| 9  | pkg.go.dev render check for v1.13.0 pages (and Edits on v1.14.0)                                                | close b1 properly                                             | S      |
| 10 | Awesome Go PR from the finalized entry (owner fork)                                                             | adoption                                                      | S      |
| 11 | Homebrew tap repo + `HOMEBREW_TAP_GITHUB_TOKEN` (LAUNCH4)                                                       | brews section currently renders-and-skips                     | S      |
| 12 | Add `/mnt/buildcache` df guard to pre-push-verify.sh                                                            | today's failure mode                                          | S      |
| 13 | CONTRIBUTING / public release-runbook page (F27.3)                                                              | the release procedure is internal-facing                      | M      |
| 14 | v2.0 design spike agenda (T27.1)                                                                                | Position sentinel, FixStrategy union, TagSet                  | M      |
| 15 | Sampling NO-GO calendar date in ROADMAP (T27.2)                                                                 | revisit trigger needs a date                                  | S      |
| 16 | Brew/nix distribution + v1.5–v1.8 backfill written proposals (T27.4)                                            | two ROADMAP open questions need owner-ready text              | M      |
| 17 | F27.4b watch note: gomend/licenseforge #1/#46                                                                   | blocked upstream; keep visible                                | S      |
| 18 | Stale `~/projects/hierarchical-errors` clone decision (ROADMAP OQ)                                              | one `git -C remote -v` + trash call                           | S      |
| 19 | `ApplySimpleFixes` philosophy decision (ROADMAP OQ)                                                             | API-contract clarity for producers                            | M      |
| 20 | `art-dupl-report.html` tracked-vs-ignored decision (ROADMAP OQ)                                                 | 800-line diff churn per run                                   | S      |
| 21 | Dependabot/actions-SHA pin-vs-bot decision                                                                      | supply-chain consistency                                      | S      |
| 22 | FuzzEditListProvider corpus promotion: commit the 59 "interesting" corpus entries as seeds                      | keeps the campaign's coverage                                 | S      |
| 23 | Extend consumer-compat to also `go vet` each resolved module (cheap signal)                                     | catches API drift beyond build                                | S      |
| 24 | Benchmark the CLI binary from the v1.13.0 release assets vs local build                                         | validates goreleaser output equivalence                       | S      |
| 25 | Consider `--skip=nix` removal after #6 (keep local/CI goreleaser paths identical)                               | removes a documented divergence                               | S      |

## g) Top #1 question I can NOT figure out myself

**Do you want the v1.14.0 train cut NOW** (I run the full documented procedure:
CHANGELOG cuts → preflight `--bench --stress` → 5 annotated signed tags one per
push → serialized Release runs → resync commit), **or should master batch more
work first?** Everything downstream that I could unblock is unblocked; the train
itself publishes, and with branch protection now active every future train also
contends with required checks — your call on timing. (Runner-up, same theme:
say the word on the Awesome Go fork/PR and the homebrew tap creation, both
ready-to-execute but external.)
