# Status: Session-Start Snapshot — go-finding + project-discovery-sdk (2026-11-17 18:46)

**Scope of this report:** This session started at 18:46 with this status request as its
FIRST instruction — **zero work was performed in this session before the report.**
Sections (a)–(d) therefore report **STATE AS FOUND**, reconstructed with live commands
this session: `git status` / `git log` / `git show --stat` in BOTH repos; read the latest
status reports (go-finding 2026-11-08 ×2, project-discovery-sdk 2026-11-11 15-03);
read both TODO_LIST.md heads; diff-inspected every daemon commit since each repo's last
report; checked root + toolsdk CHANGELOGs.

**Verified live this session:** repo cleanliness, branch sync, commit contents 2026-11-08→11-17
(go-finding) and 2026-11-11→11-17 (SDK), CHANGELOG state, `RaceEnabled` in-repo references.

**NOT verified (and why):** no test/lint/build run anywhere (nothing this session authored
to gate); GitHub CI states (no `gh` calls — out of scope per instruction); proxy/pkg.go.dev
indexing; whether the nix devshell golangci-lint binary actually matches the new ci.yml
v2.15.0 pin (claim comes from the committing session's CHANGELOG text, not re-checked);
everything asserted only by the prior reports (labeled "per report" below).

---

## a) FULLY DONE

| # | Item                                                                                                                                                                                                          | Evidence                                                      |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| 1 | **go-finding: clean tree, master == origin/master**                                                                                                                                                           | `git rev-list --left-right --count` → 0/0, `git status` clean |
| 2 | **Lint stack bumped v2.14.0 → v2.15.0** (ci.yml action pin + rationale comment updated to go1.27.4) — landed 11-14 via daemon commit `d4fa74d`                                                                | `git show d4fa74d` (this session)                             |
| 3 | **errorfamily `wrapcheck` exemption** added to `.golangci.yml` (fleet convention import-path sig) — `db58fb0` 11-14                                                                                           | `git show db58fb0`                                            |
| 4 | **`toolsdk.RaceEnabled` build-tag pair** (`race_on.go`/`race_off.go`, fleet-wide copy of the go-cqrs-lite idiom, for BuildFlow providers) — `8562d6b` 11-14                                                   | `git show 8562d6b`                                            |
| 5 | toolsdk `doc.go` +16 lines (11-12) and three flake.lock bumps (11-10, 11-12, 11-15) — all committed                                                                                                           | `git show 2a44be1 d8bea5d 5319d0a c04d884`                    |
| 6 | **SDK: v0.24.0 fully released 2026-11-11** — uniform 20-module bump, vendorHash recomputed, 20 tags, verify-external 20/20 PASS, GitHub Release `--latest`, issues #16/#19 closed, docs bookkeeping `bf9045a` | per 15-03 report (its scope line lists the live commands)     |
| 7 | **SDK: clean tree, master == origin/master**; since the release only flake.lock bumps + the 15-03 report itself landed (`095f419`, `0b79ed1`)                                                                 | `git status`, `git log` (this session)                        |
| 8 | go-finding prior-session tail (per 11-08 reports, not re-verified): docs diet pass, v1.15.1 train, 23-consumer audit all closed                                                                               | docs/status/2026-11-08 ×2                                     |

## b) PARTIALLY DONE

1. **SDK proxy/pkg.go.dev propagation** — 404 on both post-push checks; never re-checked.
   Six days cold now. Consumability proven (GOPROXY=direct matrix); visibility unconfirmed.
2. **SDK tag-CI event evidence** — the ≤3-batch tag-push fix has still never been proven
   end-to-end (billing frozen; no `gh run list` snapshot taken).
3. **SDK coverage** — AGENTS table still "Measured 2026-11-02"; the v0.24.0 tests
   (4 new edge tests) are unmeasured additions.
4. **go-finding lint-pin documentation** — ci.yml says v2.15.0 but **AGENTS.md still says
   "(2.14.0)"** in two places; memory drift introduced by whichever session bumped the pin.
5. **SDK README / release-process.md stale lines** — `Project.Remotes` unmentioned in
   README; release-process.md still says "now on v0.23.0".
6. **go-finding erraudit gate** — 0 violations locally, still no CI home (private upstream
   repo blocks runner installs; months-old gap, TODO HIGH).

## c) NOT STARTED

**go-finding:** corpus drain batch 1 (35 April-era archived reports; audit_max ratchet
340→down); erraudit CI-wirable (public repo or vendored binary); any v1.16.0-train-shaped
work; re-verification of the 11-10→11-15 daemon-committed work (no gate battery has
visibly run since 11-08); toolsdk next-tag decision for `RaceEnabled`.

**SDK** (from the 15-03 report, all still open): watch-mode multi-remote parity test;
Q3 fork-visibility semantics decision; cross-repo `Remotes` adoption (daemon, overview
fork-aware links, cvagent/pdg `go get` sweep); `scripts/post-release.sh`; proxy warm-up
runbook step; benchstat of the git-enricher subprocess cost; `go.work` 1.27.1→1.27
alignment; B1 (v0.19.2 GitHub Release backfill) and B2 (Go proposal filing) — user-gated.

## d) TOTALLY FUCKED UP

1. **go-finding root `CHANGELOG.md` [Unreleased] is MANGLED** (commit `d4fa74d`, 11-14):
   `### Changed- CI lint stack raised to v2.15.0 … sub-modules.### Fixed` — heading,
   full paragraph, and the NEXT heading are glued into ONE line. Renders as a single
   giant h3; the entry is effectively unreadable. Same daemon-edit mangle class as the
   2026-11-08 strike-marker incident (AGENTS has the pattern).
2. **Changelog drift on new public API** — `toolsdk.RaceEnabled` (11-14) appears in
   NEITHER changelog: root says "Added — Nothing yet.", toolsdk [Unreleased] is EMPTY.
   The next `toolsdk/v*` tag ships undocumented API — the exact changelog-drift gate
   class from v1.15.1.
3. **`RaceEnabled` has zero in-repo consumers or tests** — only its own two files
   reference it (`rg -l` this session). Intended consumers are external (BuildFlow
   providers), but nothing here exercises it; the toolsdk/doc.go additions from 11-12
   ride the same unverified tail.
4. **Carried, user-side — SDK B4 billing freeze:** third consecutive release shipped with
   local-gates-stand-in; the CI-truth layer (govulncheck, OS matrix, clean-runner
   vendorHash, network tag-require-check) has been dark since v0.23.0.

## e) WHAT WE SHOULD IMPROVE

1. **Two-minute fixes first:** repair the CHANGELOG mangle; add the `RaceEnabled` entries
   to both changelogs. Neither was done this session (report-only by your instruction) —
   they are the cheapest real defects on the board.
2. **Daemon heuristic commits are eating history:** all substantive 11-10→11-15 work
   landed as `chore: auto-commit N changed file(s)` — reconstructing what happened
   required per-commit `git show` archaeology (this session proved the cost: 7 commits).
   Sessions should author their own commits for real work when permitted, or at minimum
   leave a status-report/TODO trace; three go-finding work items have NO report at all.
3. **AGENTS.md went stale the same day the pin changed** — the 2.14.0→2.15.0 bump was
   committed 11-14 but the memory file (which explicitly owns that number) wasn't touched.
   The session that changes a documented number must update the doc in the same breath.
4. **Status-report cadence lags work:** go-finding's newest report is 11-08; work landed
   through 11-15. TODO_LIST "Last verified" also 11-08. Reports exist to be fresh.
5. **SDK post-release follow-through still has no automation** — proxy warm-up, tag-CI
   evidence, release-body read-back remain manual-and-skipped (15-03 report said this;
   one week later nothing changed). `scripts/post-release.sh` is the named fix.
6. **go.work vs go.mod version split-brain (SDK)** — go.mods uniform `1.27`, go.work still
   `1.27.1`; will drift again at the next toolchain bump.

## f) UP TO 50 THINGS TO GET DONE NEXT

Brainstorm + queue fuel, roughly impact-ordered within groups. GF = go-finding, SDK = project-discovery-sdk.

**Defect repair (GF — do first):**

1. Repair root CHANGELOG [Unreleased] mangle (split heading/paragraph/`### Fixed`).
2. Add `RaceEnabled` entry to root + toolsdk CHANGELOGs ([Unreleased] / Added).
3. Verify the nix devshell golangci-lint binary == ci.yml v2.15.0 pin, then fix the two stale "(2.14.0)" claims in AGENTS.md.
4. Add a toolsdk test exercising `RaceEnabled` under both build arms (`-race` on/off).
5. Review the 11-12 toolsdk/doc.go +16 additions (doc-matches-names rule) — content never reviewed in a report.

**Verification debt (GF):**
6. Full gate battery once on master (`nix run .#test`, per-module lint, `GOWORK=off` builds) — nothing has visibly gated the 11-10→11-15 work.
7. `gh run list --workflow=ci.yml` — confirm a green run exists for `c04d884` (2026-10-01 dropped-event class).
8. Corpus drain batch 1: annotate-resolve 10 April-era archived reports, then ratchet audit_max down.
9. Make erraudit CI-wirable: publish the erraudit repo or vendor the binary.
10. Decide toolsdk/v1.17.0 timing — 14 consumers; `RaceEnabled` + doc.go are sitting unreleased.

**Verification debt (SDK):**
11. B4 billing fix (user-side) — unblocks 12–16.
12. Post-billing: release-gate / verify-external / tag-require-check green on the v0.24.0 tag.
13. Confirm proxy.golang.org indexed all 20 v0.24.0 paths (`@v/list`); trigger pkg.go.dev docs for root/domain/enrichment-git.
14. Confirm tag-push CI events fired per batch (proves the ≤3 batching fix).
15. Read back the v0.24.0 GitHub Release body rendering.
16. govulncheck run (new indirect floors landed in `be50596`).
17. Coverage re-measurement, all 20 modules + AGENTS table refresh.
18. Benchstat git enricher before/after vs docs/benchmarks baseline (#19 subprocess cost).
19. `nix flake check --no-build` back into the ritual.
20. `scripts/build.sh . v0.24.0` ldflags → `sdk.Version()` smoke test.

**Test gaps (SDK, Remotes/safe.directory follow-through):**
21. Watch-mode multi-remote parity test (`handleFileEvent` path populates `Remotes`).
22. Stale `remote.pushDefault` fall-through test.
23. Detached-HEAD + multi-remote derivation test.
24. insteadOf-rewritten remote URL — decide + test the contract.
25. Token-embedded HTTPS remote userinfo-strip test at enricher level.
26. `worktreeConfig` per-worktree remote override probe + test.
27. Symlinked repo path + scoped `safe.directory` documentation/test.
28. Extend bare/worktree tests to CommitCount/LastCommitDate assertions.
29. Fuzz/property test: `parseRemoteLine` vs weird `git remote -v` shapes.
30. MR×git interplay: repoinfo-consumer test reading the `Remote` override.
31. Retro-run check-draft on the two shipped issue-close comments (learning only).

**Cross-repo consumers (SDK):**
32. Adopt `Project.Remotes` in project-discovery-daemon enrichment.
33. Overview: fork-aware link rendering (14:05 Q2 — unblocked by v0.24.0).
34. cvagent/pdg: `go get` sweep to v0.24.0 (go-ecosystem-upgrade flow).
35. External-consumer grep for widened-filter surprises.

**Decisions (owner calls):**
36. SDK Q3: per-remote vs primary visibility semantics for forks.
37. SDK: shorthand URL's implicit github.com host — enrichment surface or parse-only.
38. SDK: uniform 20-module vs changed-modules-only tagging (ADR).
39. Verify `BrandedProject`/`AsBranded` carries `Remotes`; wire if dropped.

**Docs & hygiene:**
40. SDK README: `Remotes` in the Project field table + filter quickstart line.
41. SDK: fix release-process.md "now on v0.23.0" drift; add proxy warm-up + release-body read-back steps.
42. SDK: annotate the 14:05 report (Q1 answered; successor pointer) — docs-health ANNOTATE.
43. SDK: align go.work to `1.27` (kill the two-place version definition).
44. SDK: AGENTS #120 record (v0.24.0 stripped the enrichment/git→domain replace); design-decision entry for the hostless-URL rationale; FEATURES row for refusal-logging.
45. SDK: TODO_LIST harvest of the 15-03 (f) list.
46. GF: TODO_LIST "Last verified" refresh + harvest this report's (f).

**Tooling/infra (SDK):**
47. `scripts/post-release.sh` automating checklist steps 2–9.
48. release.sh guard: verify tags point at the bump commit (`--points-at`).
49. Perf: batch the two git config probes into one subprocess per project; consider `for-each-ref` for version enrichment.
50. SBOM / go-licenses re-run after the indirect drift; Dependabot scope review beyond `/mr`.

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **SDK B4 — billing:** has the GitHub Actions payment/spending-limit issue been fixed
   since 2026-11-11? If yes, I'll re-run the tag-CI/release-gate verification next session;
   if deliberately frozen, tell me whether "local-gates-stand-in" is now the accepted
   permanent policy or still a temporary disposition. (Intent/policy, not discoverable.)
2. **go-finding — the undocumented 11-10→11-15 work:** `toolsdk.RaceEnabled`, the doc.go
   additions, and the lint-stack bump landed with no report, no TODO row, and no changelog
   entries. Was that a COMPLETE fleet rollout (just needs bookkeeping), or does it have an
   unfinished tail (tests, consumers, a toolsdk/v1.17.0 tag)? I can read the code but not
   the intent — should I verify + release it, or leave it?
3. **Priority between the two queues:** next session budget — go-finding (corpus drain /
   erraudit CI / toolsdk release decision) or SDK (post-v0.24.0 follow-through)? Both
   boards are healthy; the ordering is an owner call.

---

**THEN WAIT FOR INSTRUCTIONS.**
