# Status Report: Public Release Assessment Session

> **Date:** 2026-07-24 23:11
> **Session scope:** PRO/CONTRA analysis for making go-finding public + TODO list
> **Assessed by:** Crush

> **Update 2026-07-28:** the 3-phase public-release plan from this session was executed.
> SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates, and support policy were created
> (commits `f325122`, `63a6639`). The `GOPRIVATE` warning was removed from README. Dependabot
> was expanded to all 4 sub-modules. v1.4.0 was tagged (2026-07-26). The repo itself is still
> private pending the Phase 3 visibility flip — see TODO_LIST "Make Repo Public — Phase 3".
> Open question Q3 (Homebrew tap secret) is still unresolved.

---

## What This Session Did

The user asked for a PRO/CONTRA analysis + actionable TODO list for making the repo public. I:

1. **Researched the project** — README, all 4 `go.mod` files, CI workflows, release workflow, dependabot config, `.gitignore`, version.go, FEATURES.md, TODO_LIST.md, CHANGELOG.md, CONTRIBUTING.md, CONTEXT.md, PUBLIC_OR_PRIVATE.md, docs/READINESS_REPORT.md, GitHub API (repo visibility, stars, license)
2. **Scanned for blockers** — secrets (none), committed binaries (none), large files (none), hardcoded personal paths (none), email leaks (only AUTHORS file, intentional)
3. **Discovered a split brain** — `PUBLIC_OR_PRIVATE.md` claimed the repo was already public; GitHub API confirmed it's `PRIVATE`
4. **Created `docs/PRO_CONTRA_make-public.md`** — Fresh assessment with 8 PROs, 7 CONTRAs, 3-phase TODO
5. **Fixed `PUBLIC_OR_PRIVATE.md`** — Corrected the false "public" banner
6. **Updated `TODO_LIST.md`** — Added 16 public-release tasks across 3 phases
7. **Pivoted on json/v2** — Initial draft framed `encoding/json/v2` as a "BLOCKER" recommending migration. User said they want to keep json/v2. Rewrote the analysis to treat it as a conscious cutting-edge choice, not a defect.

---

## a) FULLY DONE

| Task                                                 | Evidence                                                                                    |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| GitHub API verification (private vs public)          | `gh repo view --json visibility` → `PRIVATE`                                                |
| Secrets scan (passwords, keys, tokens, PII)          | `rg` scan across all non-doc files — clean                                                  |
| Committed artifacts scan (binaries, large files)     | `git ls-files` + size check — clean                                                         |
| Personal paths scan (`/home/lars`, hardcoded emails) | `rg` scan — only AUTHORS file (intentional)                                                 |
| Read all 4 `go.mod` files                            | Confirmed deps, replace directives, go 1.26.4                                               |
| Read CI workflow (8 jobs)                            | test, coverage, lint, version-check, govulncheck, module-isolation, stress, dupl, benchmark |
| Read release workflow + dependabot                   | GoReleaser + cosign + SBOM; 4 modules + GH actions tracked                                  |
| Created `docs/PRO_CONTRA_make-public.md`             | Full PRO/CONTRA + 3-phase TODO (16 tasks)                                                   |
| Fixed `PUBLIC_OR_PRIVATE.md` split brain             | Removed false "public" banner, added correction                                             |
| Updated `TODO_LIST.md` with public-release tasks     | 3 phases, 16 tasks in HIGH Priority section                                                 |
| Pivoted json/v2 framing after user decision          | Rewrote CONTRA #1, Executive Summary, Assessment Matrix, TODO                               |

---

## b) PARTIALLY DONE

| Task                               | What's done                                              | What's missing                                                                  |
| ---------------------------------- | -------------------------------------------------------- | ------------------------------------------------------------------------------- |
| json/v2 file inventory             | Identified 4 core production files + 20 test/other files | Didn't audit pipeline/analysis/CLI modules for json/v2 usage in production code |
| README review for public readiness | Identified GOPRIVATE warning + buried json/v2 note       | Didn't draft the actual replacement text                                        |
| Internal docs noise assessment     | Counted 91 internal docs (status, planning, reviews)     | Didn't categorize which are safe to keep vs should be moved/archived            |

---

## c) NOT STARTED

- No `SECURITY.md`, `CODE_OF_CONDUCT.md`, issue/PR templates created (only listed as TODO)
- No GitHub repo description/topics set (only listed as TODO)
- No README edits made (GOPRIVATE warning still present, json/v2 not prominent)
- No verification that existing tags produce working GitHub Releases
- No check of GitHub repo settings (issues enabled? discussions? wiki?)
- No pkg.go.dev rendering check (impossible while private)
- No Homebrew tap repo verification
- No announcement draft

---

## d) TOTALLY FUCKED UP / Mistakes Made

### 1. Biased Initial json/v2 Recommendation

**What:** My first draft of `PRO_CONTRA_make-public.md` framed `encoding/json/v2` as a "BLOCKER" and explicitly recommended migrating to `encoding/json` (v1) as "Option A — Most pragmatic."

**Why it was wrong:** I presented my opinion as the default before asking the user. The user had to correct me. I should have presented json/v2 neutrally — as a decision point with tradeoffs — and asked the user which direction they preferred before writing the recommendation.

**Lesson:** When a technical choice has legitimate arguments on both sides (cutting-edge vs broad compatibility), present it as a decision for the user, not a foregone conclusion.

### 2. Garbled Text in Initial Draft

**What:** The original "If You Choose to Stay Private" section contained `窗口窗口` (Chinese characters meaning "window") in the phrase "The first-mover advantage窗口窗口 is closing."

**Why it happened:** Token-level corruption during generation. Caught and fixed in the json/v2 pivot edit, but it should never have been written.

### 3. Didn't Verify Consumer Count Independently

**What:** I stated "22 known consumers" and "14 with Go code" based on existing project docs (TODO_LIST.md, FEATURES.md). I did not independently verify this by checking the consumer repos or running a `gh` query.

**Impact:** These numbers may be stale or inflated. They feed directly into the PRO arguments.

### 4. Didn't Check Module Proxy Readiness

**What:** I didn't verify whether the Go module proxy can actually resolve the tags once the repo goes public. The multi-module tagging strategy (`v*`, `pipeline/v*`, `analysis/v*`, `cmd/go-finding/v*`) is complex. A stale or missing tag on any sub-module would cause `go get` failures.

**Impact:** First impression matters. If `go get github.com/larsartmann/go-finding/pipeline` fails on day one, adopters leave.

---

## e) WHAT WE SHOULD IMPROVE (Self-Critique of This Session)

### Process Improvements

1. **Ask before recommending on contested decisions.** The json/v2 framing cost a round-trip. I should have asked "json/v2 is required — do you want to keep it or migrate?" before writing the analysis.

2. **Verify numbers, don't parrot docs.** "22 consumers" came from existing docs. I should have verified via `gh search repos` or consumer repo checks.

3. **Test claims, don't just assert.** I claimed "go get fails without GOEXPERIMENT" but didn't actually run `GOWORK=off GOEXPERIMENT=off go build ./...` to prove it. The claim is likely correct (based on AGENTS.md), but verification > assumption.

4. **Audit all modules, not just core.** I focused the json/v2 analysis on the 4 core files but didn't check whether pipeline/analysis/CLI also have production json/v2 imports that would affect consumers.

### Content Improvements for the Analysis

5. **Missing: GitHub repo settings audit.** Are issues enabled? Wiki? Discussions? Projects? These affect community experience.

6. **Missing: Existing GitHub Releases check.** Do the v1.0.0–v1.3.0 tags already have GitHub Releases, or just tags? The release workflow suggests they should.

7. **Missing: License file verification.** I saw the MIT badge but didn't verify `LICENSE` file content matches MIT text.

8. **Missing: `doc.go` / pkg.go.dev readiness.** I didn't check whether `doc.go` has proper package-level documentation that renders well on pkg.go.dev.

9. **Missing: Dependency audit for public consumption.** The pipeline depends on `gogenfilter`, the CLI depends on `go-output` — both are LarsArtmann repos. Are THEY public? If private, consumers can't resolve them.

10. **Missing: `go.sum` completeness for GOWORK=off.** The module-isolation CI job tests this, but I didn't verify the go.sum files have all needed hashes for public proxy resolution.

---

## f) Up to 50 Things We Should Get Done Next

### Phase 1: Critical (before `gh repo edit --visibility public`)

| #      | Task                                                                                                                                                         |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~1~~  | ~~Make `GOEXPERIMENT=jsonv2` prominent in README Installation section (move to top, add `go env -w` one-liner)~~ done — 2026-07-24_23-25 session             |
| ~~2~~  | ~~Remove `GOPRIVATE` warning from README (lines 46-47)~~ done — 2026-07-24_23-25 session                                                                     |
| ~~3~~  | ~~Add GitHub repo description: `gh repo edit --description "Unified data model and pipeline for Go static analysis tools"`~~ done — 2026-07-24_23-25 session |
| ~~4~~  | ~~Add GitHub topics: `gh repo edit --add-topic go,static-analysis,sarif,lsp,code-quality,golang,static-analysis-tools`~~ done — 2026-07-24_23-25 session     |
| 5      | ~~Verify `go-output` is public~~ ✅ `DONE`                                                                                                                   |
| 6      | ~~Verify `gogenfilter` is public~~ ✅ `DONE`                                                                                                                 |
| 7      | Run `GOWORK=off GOEXPERIMENT=off go build ./...` in each module to confirm json/v2 failure mode                                                              |
| 8      | Audit pipeline/analysis/CLI for json/v2 production imports (not just core)                                                                                   |
| 9      | Check GitHub repo settings (issues, wiki, discussions enabled?)                                                                                              |
| 10     | Verify existing tags have GitHub Releases (v1.0.0 through v1.3.0)                                                                                            |
| ~~11~~ | ~~Verify `LICENSE` file is standard MIT text~~ done — MIT verified 23-43                                                                                     |

### Phase 2: Community Health (before announcing)

| #      | Task                                                                                                                |
| ------ | ------------------------------------------------------------------------------------------------------------------- |
| ~~12~~ | ~~Create `SECURITY.md` (vulnerability reporting policy)~~ done — SECURITY.md shipped v1.4.0                         |
| ~~13~~ | ~~Create `CODE_OF_CONDUCT.md` (Contributor Covenant v2.1)~~ done — CODE_OF_CONDUCT.md shipped v1.4.0                |
| ~~14~~ | ~~Create `.github/ISSUE_TEMPLATE/bug_report.yml`~~ done — templates shipped 23-25                                   |
| ~~15~~ | ~~Create `.github/ISSUE_TEMPLATE/feature_request.yml`~~ done — templates shipped 23-25                              |
| ~~16~~ | ~~Create `.github/PULL_REQUEST_TEMPLATE.md`~~ done — templates shipped 23-25                                        |
| ~~17~~ | ~~Add support policy section to README ("MIT license, best-effort support, no SLA")~~ done — README Support section |
| ~~18~~ | ~~Decide on 91 internal docs~~ — **No action (user decision: keep as-is)**                                          |
| ~~19~~ | ~~Add "Internal docs" note to README if keeping them~~ **Won't implement — keep as-is per Q2 resolution.**          |
| 20     | Verify `doc.go` has package-level comment that renders on pkg.go.dev                                                |
| 21     | Check all exported types/functions have GoDoc comments                                                              |
| ~~22~~ | ~~Verify `examples/` directory compiles and is referenced in README~~ done — examples compile-tested                |
| ~~23~~ | ~~Add `FUNDING.yml` (optional — GitHub Sponsors)~~ **Won't implement — no FUNDING.yml planned.**                    |

### Phase 3: Launch

| #      | Task                                                                              |
| ------ | --------------------------------------------------------------------------------- |
| ~~24~~ | ~~Tag v1.4.0 as the first public version anchor~~ done — v1.4.0 tagged 2026-07-26 |
| 25     | Verify GoReleaser produces correct binaries on public tag                         |
| 26     | Verify `HOMEBREW_TAP_GITHUB_TOKEN` secret exists and tap repo is configured       |
| 27     | Verify `brew install go-finding` works after release                              |
| 28     | Verify pkg.go.dev renders documentation after first public `go get`               |
| 29     | Write blog post: "Seven tools detect issues. Zero route them to remediation."     |
| 30     | Prepare r/golang post (follow their self-promotion rules)                         |
| 31     | Post in Go Slack #showcase channel                                                |
| 32     | Post on Twitter/X                                                                 |
| 33     | Submit to [Awesome Go](https://github.com/avelino/awesome-go) via PR              |
| 34     | Submit to Awesome SARIF                                                           |
| 35     | Add "Used by" section to README once stars/forks appear                           |

### Ongoing / Post-Launch

| #      | Task                                                                                                                          |
| ------ | ----------------------------------------------------------------------------------------------------------------------------- |
| 36     | Track Go 1.27 release — json/v2 stabilization removes `GOEXPERIMENT` requirement                                              |
| 37     | When json/v2 stabilizes: remove all `GOEXPERIMENT=jsonv2` from docs, CI, flake.nix                                            |
| 38     | Set up GitHub Discussions for Q&A (separate from issues)                                                                      |
| 39     | Add `CONTRIBUTING.md` section on "Adding new detectors"                                                                       |
| ~~40~~ | ~~Create a logo/visual identity for the project~~ **Won't implement — no logo planned.**                                      |
| ~~41~~ | ~~Create a project website (Astro + Starlight pattern)~~ **Won't implement — no website planned, ROADMAP non-goal.**          |
| 42     | Add badge for Go Reference (pkg.go.dev) once verified rendering                                                               |
| 43     | Monitor first week of issues/PRs — respond within 48h                                                                         |
| ~~44~~ | ~~Write a "v2.0 vision" roadmap entry (what would justify a major bump)~~ done — v2.0 in ROADMAP Hardening                    |
| 45     | Consider a `CHANGELOG.md` entry for "Repo made public"                                                                        |
| 46     | Update `AGENTS.md` with the new `docs/PRO_CONTRA_make-public.md` reference                                                    |
| 47     | Run a final full-code-review skill before going public                                                                        |
| ~~48~~ | ~~Run brutal-self-review skill before going public~~ done — this self-critique series                                         |
| 49     | Verify all 4 module tags resolve on `proxy.golang.org` after visibility flip                                                  |
| ~~50~~ | ~~Create a "Migration guide for consumers" (removing GOPRIVATE, adding GOEXPERIMENT)~~ done — consumer-migration guides exist |

---

## g) Questions I CANNOT Answer Myself

### ~~Q1: Are `go-output` and `gogenfilter` public?~~

**RESOLVED (2026-07-24):** User confirmed both repos are public. Transitive
dependencies will resolve for consumers after go-finding goes public. Not a blocker.

### ~~Q2: Do you want the 91 internal docs to remain public, be moved, or be gitignored?~~

**RESOLVED (2026-07-24):** User said "no" — no action. All internal docs stay as-is.

### Q3: What is the Homebrew tap repo and is the `HOMEBREW_TAP_GITHUB_TOKEN` secret configured?

The release workflow references `HOMEBREW_TAP_GITHUB_TOKEN` in the GoReleaser step. I cannot verify whether this secret exists in GitHub or whether the tap repo (`homebrew-tap` or similar) is set up. If it's missing, the first public release will fail at the GoReleaser step. I cannot check repo secrets without admin access.

---

## Session Files Modified

| File                             | Change                                                                        |
| -------------------------------- | ----------------------------------------------------------------------------- |
| `docs/PRO_CONTRA_make-public.md` | **Created** — Full PRO/CONTRA analysis (supersedes `PUBLIC_OR_PRIVATE.md`)    |
| `PUBLIC_OR_PRIVATE.md`           | **Edited** — Corrected false "public" banner → correction pointing to new doc |
| `TODO_LIST.md`                   | **Edited** — Added 16 public-release tasks across 3 phases in HIGH Priority   |

---

_Assisted-by: Crush <crush@charm.land>_
