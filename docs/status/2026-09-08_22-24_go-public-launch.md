# Status Report: go-finding Made Public — Launch & Verification

> **Date:** 2026-09-08 22:24–22:40 CEST | **Session:** visibility flip + verification
> **Trigger:** User instruction "just make it public" after the public/private
> assessment conversation (grounded in `docs/PRO_CONTRA_make-public.md`, 2026-07-24).

---

## TL;DR

`LarsArtmann/go-finding` is now **PUBLIC** (flipped 2026-09-08 22:24 CEST). Module
proxy resolution verified end-to-end with **no `GOPRIVATE`**. Post-flip CI is
actually running for the first time in weeks (billing had killed it): 13/20 jobs
green at snapshot, **5 failed, root causes not yet triaged**. Two significant
pre-existing discoveries: **GitHub Releases silently stopped at v1.4.0** (tags
v1.5.0–v1.8.0 exist but have no releases — GoReleaser presumably billing-dead)
and **`HOMEBREW_TAP_GITHUB_TOKEN` secret does not exist**. pkg.go.dev still
indexing (404 at last check, normal within the first hour).

---

## a) FULLY DONE

| Item | Evidence |
|---|---|
| Pre-flip readiness re-verified | v1.8.0 tagged; `version.go` matches; all Phase 1+2 blockers from `docs/PRO_CONTRA_make-public.md` confirmed done (community files, README `GOEXPERIMENT` prerequisite at line 29, repo description + topics, split-brain fix, MIT license, no secrets in history) |
| **Visibility flipped to PUBLIC** | `gh repo edit LarsArtmann/go-finding --visibility public --accept-visibility-change-consequences`; verified `isPrivate: false` |
| **Public proxy resolution verified** | From clean env, no `GOPRIVATE`: `GOWORK=off GOPROXY=https://proxy.golang.org go list -m github.com/larsartmann/go-finding@v1.8.0` → resolves; `@latest` → v1.8.0 |
| CI re-enabled by flip | `gh workflow run ci.yml --ref master` → run 34274674104 actually **executes** (public Actions are free). Contrast: last private-repo push run (34270639388) had **all 20 jobs fail in 43s with no logs** = billing block |
| Stale private-repo docs updated (6 files) | `AGENTS.md` (repo-public + CI-billing-superseded bullets), `docs/release-procedure.md` (consumer setup + verify commands de-GOPRIVATE'd), `docs/guides/troubleshooting.md` ("410 Gone" fix rewritten), `TODO_LIST.md` (consumer-compat test `BLOCKED` → `READY`), `PUBLIC_OR_PRIVATE.md` (RESOLVED banner), `docs/PRO_CONTRA_make-public.md` (flip update header) |

## b) PARTIALLY DONE

| Item | State | Gap |
|---|---|---|
| CI verification (post-flip) | 13 success, 5 failure, 1 skipped (changelog-check), 1 running at 4m5s snapshot | Failed: `coverage`, `markdown-link-check`, `arch-check`, `stress`, `docs-api-check`. Logs locked until run completes (`--log-failed`: "logs will be available when it is complete"). Root causes untriaged |
| pkg.go.dev rendering (TODO #11) | First proxy fetch triggered by verification; pkg.go.dev still 404 at ~15 min | Indexing typically minutes-to-an-hour; needs recheck + manual refresh request if stuck |
| TODO_LIST "Post-account-switch CI work" section | One row unblocked (consumer compat) | Section header still reads "gated on billing fix" — the gate is dissolved; whole section needs a rewrite pass, not done this session |
| Release pipeline | Tags v1.5.0–v1.8.0 pushed; only v1.4.0 has a GitHub Release | GoReleaser evidently silent-failing since ~v1.5.0. Not fixed — discovered this session |

## c) NOT STARTED (Phase 3 + known open items)

- Announcement (blog post / r/golang / Go Slack / X) — draft exists at
  `docs/brainstorming/launch-announcement-draft.md`, not reviewed or published
- Awesome Go submission
- json/v2 stabilization tracking (TODO #12, Go 1.27)
- GoReleaser + Homebrew tap verification on a real public tag
- pkg.go.dev render check for the 3 sub-modules
- FUNDING.yml (explicitly optional, never added)

## d) TOTALLY FUCKED UP (pre-existing, discovered today)

1. **README lost its `GOPRIVATE` warning on 2026-07-24 while the repo stayed
   private.** Item #3 in `docs/PRO_CONTRA_make-public.md` was marked ✅ "no longer
   needed once public" — but public never happened until today. ~6 weeks where
   the README had zero guidance for the actual (broken) consumer experience.
2. **GitHub Releases silently dead since v1.5.0.** Four core tags (v1.5.0–v1.8.0)
   exist in git with **no GitHub Release** — GoReleaser has presumably been
   failing (billing?) and nobody noticed because CI red had become background
   noise. `analysis/v1.5.0` is currently marked "Latest" release, which is wrong
   on its face.
3. **`HOMEBREW_TAP_GITHUB_TOKEN` secret does not exist** (`gh secret list` is
   empty). The GoReleaser homebrew step has never been able to work. This was
   flagged as "verify" in the July doc; it is now confirmed missing.
4. **CI as a dead gate (dead-gate lesson, 3rd instance).** Billing-blocked
   all-fail runs made red the normal state, which is exactly how #1 and #2 hid.
   A check that cannot run is worse than no check.
5. Nothing in *this session's* execution failed — flip was clean and verified.

## e) WHAT WE SHOULD IMPROVE

- **Triage the 5 failed CI jobs today.** A public launch with red CI is the first
  thing visitors see.
- **Public-repo hardening:** secret scanning + push protection are **disabled**
  (free on public repos); branch protection on master does not exist; no
  tag-protection rules. Going public without these invites supply-chain noise.
- **First-impression pass:** 39 files in `docs/status/` + `docs/reviews/` +
  `docs/planning/` are now world-readable per the "keep all as-is" decision.
  At minimum add one explanatory doc so the repo doesn't read as a personal
  journal (the July doc's own words).
- **`.buildflow.yml` is tracked** (mode 600, internal tool config) — decide:
  untrack or keep.
- **Release hygiene:** get v1.8.0 released properly (GoReleaser fixed, secret
  set or homebrew step removed), and re-point "Latest" to a core release.
- Update `docs/PRO_CONTRA_make-public.md` checkboxes (#11, Phase 3) as they
  complete — it is declared a living assessment.

## f) NEXT: 50 things to get done (prioritized)

**Today — stabilize the public state**

| # | Task | Why |
|---|---|---|
| 1 | Wait for run 34274674104 to finish; read `--log-failed`; triage the 5 failures | Red CI on a fresh public repo is the storefront |
| 2 | Fix `markdown-link-check` findings | Broken public docs links |
| 3 | Fix `docs-api-check` drift (documented identifiers vs code) | Public API doc accuracy |
| 4 | Fix `arch-check` (go-arch-lint) failure | Structural gate credibility |
| 5 | Investigate `coverage` failure (threshold vs Codecov upload/token) | Coverage is a headline metric |
| 6 | Investigate `stress` failure (flake vs real; race repeat=20) | It is the MANDATORY release gate |
| 7 | Recheck pkg.go.dev (all 4 module paths); force re-index if 404 persists >1h | TODO #11 |
| 8 | Verify proxy resolution for `pipeline/v1.8.0`, `analysis/v1.8.0` (or latest), `cmd/go-finding` tags | Core verified this session; sub-modules not |
| 9 | Create GitHub Release for v1.8.0 (at minimum manual notes + binaries) | 4 tags with no releases; "Latest" points at analysis/ |
| 10 | Diagnose GoReleaser: why v1.5.0–v1.8.0 produced no releases | Silent pipeline death |
| 11 | Create or consciously drop `HOMEBREW_TAP_GITHUB_TOKEN` (+ tap repo) | Secret confirmed missing |
| 12 | Enable secret scanning + push protection (free) | Public-repo supply-chain hygiene |
| 13 | Add branch protection on master (require green CI once achieved) | Now meaningful because CI runs |
| 14 | Rewrite TODO_LIST "Post-account-switch CI work" section (gate dissolved) | Doc truth |
| 15 | Decide `.buildflow.yml` (tracked, mode 600, internal) | Public hygiene |

**This week — launch proper**

| # | Task | Why |
|---|---|---|
| 16 | Decide announcement timing (see questions) | Strategy |
| 17 | Finalize announcement draft (`docs/brainstorming/launch-announcement-draft.md`) | Exists, unreviewed |
| 18 | r/golang post | Adoption |
| 19 | Go Slack (#golang / #tools) | Adoption |
| 20 | X/Twitter thread | Adoption |
| 21 | awesome-go PR | Discoverability |
| 22 | GitHub social preview image | First impression in link unfurls |
| 23 | Pin repo on LarsArtmann profile | Discoverability |
| 24 | Enable GitHub Discussions (support channel vs issue spam) | Community load management |
| 25 | Run consumer compatibility matrix (22 consumers, 14 with Go code) — now `READY` | Unblocked this session; validates the public promise |
| 26 | Verify all README badges resolve truthfully post-flip (codecov, pkg.go.dev, CI) | Badges were unread in private era |
| 27 | Verify Dependabot PRs still flow (public) | Dependency freshness |
| 28 | Add explanatory README for `docs/status|reviews|planning` (or move them) | Journal-look mitigation |
| 29 | Review ROADMAP/TODO_LIST for strategy-sensitive content you don't want public | One-time pass; keep-decision stands unless changed |
| 30 | SECURITY.md: enable GitHub private vulnerability reporting | Public repos need a working intake |

**Next 2–4 weeks — engineering tail**

| # | Task | Why |
|---|---|---|
| 31 | Track json/v2 stabilization (Go 1.27); drop `GOEXPERIMENT` when it lands | TODO #12; removes biggest adoption friction |
| 32 | Consider a compat shim or docs pattern for pre-1.26 users | Adoption friction |
| 33 | Tag protection rules (core `v*` vs sub-module prefixes) | Release integrity |
| 34 | Re-tag/patch to re-trigger GoReleaser end-to-end (v1.8.1 or v1.9.0) | Prove the pipeline, not just hope |
| 35 | Blog post on lars.software (website-launch pattern) | Durable launch content |
| 36 | pkg.go.dev example rendering pass (example_test.go is 22KB — confirm godoc shows well) | pkg.go.dev is the landing page |
| 37 | Add repo to profile README + ecosystem listing with sibling projects | Cross-discovery |
| 38 | CI: cache tuning so wall-clock of stress job doesn't dominate iteration | Free minutes ≠ free time |
| 39 | Consider Making "Latest" release semantics explicit in release-procedure.md | Prevents recurrence of #9 |
| 40 | Post-launch retro status report | Docs-health habit |

**Backlog / nice-to-have**

| # | Task | Why |
|---|---|---|
| 41 | FUNDING.yml (optional, consciously skipped in July) | Sustainability |
| 42 | `.github/FUNDING` vs GitHub Sponsors decision | Same |
| 43 | Issue label system + triage cadence doc | Public issue inflow |
| 44 | PR welcome policy in CONTRIBUTING (what gets merged) | Sets expectations |
| 45 | Changelog entry convention for repo-level changes (or explicitly exclude) | CHANGELOG purity |
| 46 | Consider `goreleaser` GitHub-Actions-only release (drop homebrew) if tap is not wanted | Simplification |
| 47 | Social proof: benchmark table vs golangci-lint output status quo | Sales page (README) ammo |
| 48 | Monitor stars/forks/issues weekly; first-responders SLA per README support policy | Maintenance reality |
| 49 | GOPRIVATE mentions in `scripts/release-preflight.sh` (defensive, harmless) — leave or clean | Hygiene |
| 50 | Archive `PUBLIC_OR_PRIVATE.md` into `docs/archive/` — question answered, doc now historical | Reduce top-level noise |

## g) QUESTIONS (cannot answer myself)

1. **Announce now or wait?** CI has 5 red jobs and pkg.go.dev hasn't rendered
   yet. Announce only after green + verified docs (my recommendation), or accept
   rough edges and post this week anyway?
2. **Internal docs stand?** The July decision was "keep all as-is" while the repo
   was still private. Now that it *is* public: keep everything (status reports,
   reviews, planning), add a boundary README, or move them to a private repo?
3. **Homebrew/binary distribution: wanted at all?** The tap secret never existed.
   Should I set up the tap repo + secret properly (needs your GitHub app/token
   input), or drop homebrew from GoReleaser and ship the Go module + release
   binaries only?

---

**Verification transcript (key commands):**

```bash
gh repo edit LarsArtmann/go-finding --visibility public --accept-visibility-change-consequences
gh repo view --json isPrivate            # → false
GOWORK=off GOPROXY=https://proxy.golang.org go list -m github.com/larsartmann/go-finding@v1.8.0   # resolves, no GOPRIVATE
gh workflow run ci.yml --ref master      # run 34274674104
gh secret list                           # → empty (HOMEBREW_TAP_GITHUB_TOKEN missing)
gh release list                          # → v1.4.0 is newest core release; v1.5.0–v1.8.0 missing
```
