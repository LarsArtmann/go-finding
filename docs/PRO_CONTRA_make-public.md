# PRO/CONTRA: Making go-finding Public

> **Date:** 2026-07-24 | **Current state:** Repo is **PRIVATE** (confirmed via GitHub API)
>
> This document supersedes the resolution banner in `PUBLIC_OR_PRIVATE.md`, which
> incorrectly claims the repo is already public. It is not. This is a fresh assessment
> at v1.3.0 with 1,005 commits, 193 Go files, and 112 test files.

---

## Executive Summary

**Verdict: Make public — but fix the GOEXPERIMENT=jsonv2 barrier first.**

The project is production-grade: zero-dep core, 95%+ coverage, comprehensive CI,
lossless SARIF/LSP round-trip, clean multi-module architecture. The ecosystem gap
it fills is real and unfilled. The single biggest adoption blocker is the
`encoding/json/v2` dependency in the **core** library, which requires every consumer
to set `GOEXPERIMENT=jsonv2`. This should be resolved (migrate to `encoding/json`
or wait for Go 1.27 to stabilize json/v2) **before** going public.

| Dimension                      | Rating   | Key Finding                                                        |
| ------------------------------ | -------- | ------------------------------------------------------------------ |
| Code quality                   | **A**    | Clean architecture, branded types, decomposed validators           |
| Test coverage                  | **A**    | 112 test files, fuzz, BDD, property, integration, E2E, stress      |
| API stability                  | **A**    | v1.0.0 frozen (2026-06-24), v1.3.0 current, no deprecated APIs     |
| CI/CD maturity                 | **A**    | 8 CI jobs: test, lint, vulncheck, coverage, stress, dupl, bench    |
| Documentation                  | **B+**   | Strong README/CONTRIBUTING/FEATURES, but 91 internal docs exposed  |
| Ecosystem value                | **A**    | Fills a genuine gap — no comparable Go library exists              |
| **Adoption friction**          | **C**    | **GOEXPERIMENT=jsonv2 required — breaks `go get` for most users**  |
| Community readiness            | **C+**   | No SECURITY.md, no issue templates, no repo description on GitHub  |
| Maintenance sustainability     | **B**    | Single author, 22 known consumers, MIT license                     |

---

## PRO — Arguments for Making Public

### 1. Fills a Genuine, Unfilled Ecosystem Gap

> "Seven tools detect issues. Zero tools route them to remediation."

Go has `staticcheck`, `govet`, `golangci-lint`, `revive`, `nilness`, etc. — all with
different output formats. **No standard interchange type or automated fix pipeline
exists.** `go-finding` provides both. No comparable library exists in the Go ecosystem
as of 2026-07.

### 2. Production-Grade Code

- **193 Go source files**, **112 test files** — 58% test-to-source ratio
- **Zero external production dependencies** in core module (stdlib only)
- **95%+ test coverage** with race detector, fuzz, BDD (Ginkgo), property tests
- **Thread-safe** `Report` with mutex protection, `iter.Seq` iterators
- **Branded types** (`ID`, `RuleName`, `ToolName`, `FilePath`) — compile-time safety
- **Decomposed validation** — 6 per-field validators, each <10 complexity
- **Lossless SARIF/LSP round-trip** via property bags and `LSPDiagnosticData`

### 3. Proven by Real Consumers

22 known consumer projects (14 with Go code), including:
- `art-dupl` — code duplication detection
- `branching-flow` — Go code quality analyzer
- `hierarchical-errors` — error handling pattern detector
- `go-auto-upgrade` — dependency upgrade automation

The v1.3.0 release was explicitly **consumer-driven**: 11 additive APIs eliminating
boilerplate found by auditing all 22 consumers.

### 4. Excellent Infrastructure

| Component         | Status                                                                  |
| ----------------- | ----------------------------------------------------------------------- |
| GitHub Actions CI | 8 jobs: multi-OS test, lint, govulncheck, coverage, stress, dupl, bench |
| GoReleaser        | Cross-platform binaries (linux/darwin/windows, amd64/arm64), cosign SBOM |
| golangci-lint v2  | 80+ linters, strict config, zero outstanding findings                   |
| Codecov           | Integrated with per-package coverage thresholds                         |
| Dependabot        | All 4 sub-modules + GitHub Actions tracked                              |
| Multi-module      | 4 independent Go modules with replace directives for GOWORK=off CI      |
| Benchmark CI      | Regression detection vs baseline (25% threshold)                        |

### 5. API Stability Guaranteed

- API frozen since v1.0.0 (2026-06-24)
- Semantic versioning enforced
- `version.go` auto-checked against git tags in CI
- No deprecated APIs remain (all v0.x deprecations removed in v1.0.0)
- `docs/API_STABILITY.md` documents the contract

### 6. First-Mover Advantage

Publishing now establishes `go-finding` as the de facto standard before a competitor
fills the gap. Every month private is a month of lost mindshare and adoption.

### 7. Documentation Is Above Average

README with quick start, builder API examples, pipeline usage, CLI flags.
CONTRIBUTING.md with dev setup. FEATURES.md with honest feature inventory.
CHANGELOG.md following Keep a Changelog. Domain language glossary. Usage guide.
Architecture decision records. Migration guides.

### 8. Clean Git History

- No secrets found (scanned for passwords, API keys, tokens, private keys)
- No hardcoded personal paths (`/home/lars`, etc.)
- No committed binaries or large files
- No committed coverage artifacts (`.gitignore` covers `*.out`)

---

## CONTRA — Arguments Against Making Public (Right Now)

### 1. GOEXPERIMENT=jsonv2 — THE Critical Adoption Barrier

**Severity: BLOCKER**

The core library imports `encoding/json/v2` in 4 production files:
`json.go`, `sarif_types.go`, `sarif_import.go`, `sarif_export.go`.

This means **every consumer** must:
1. Use Go 1.26+ (not yet the default in most environments)
2. Set `GOEXPERIMENT=jsonv2` environment variable
3. Without it: `build constraints exclude all Go files` — the import fails

**Impact:** `go get github.com/larsartmann/go-finding` **does not work** out of the box.
A user following the README's installation instructions will hit an immediate wall.
This is the #1 reason most potential adopters will abandon the library.

**Options to resolve:**
- **A) Migrate back to `encoding/json`** (v1) — Most pragmatic. The json/v2 API is
  similar; the migration is mechanical. Removes the barrier entirely. Cost: lose
  json/v2 performance and some API ergonomics.
- **B) Wait for Go 1.27** — json/v2 is expected to be stabilized. But timeline is
  unknown (likely late 2026), and Go 1.26 adoption is still ramping.
- **C) Keep json/v2, document loudly** — Highest friction. Only viable if the target
  audience is exclusively Go 1.26+ early adopters.

**Recommendation: Option A.** Migrate to `encoding/json` before going public. The
adoption cost of json/v2 far outweighs its benefits for a library seeking broad
adoption.

### 2. Split Brain: PUBLIC_OR_PRIVATE.md Claims Public

`PUBLIC_OR_PRIVATE.md` has a resolution banner stating:
> "Decision made — project is open-source. v1.2.0 tagged and public."

**This is false.** The repo is `isPrivate: true` on GitHub. This document must be
corrected before going public to avoid confusion.

### 3. 91 Internal Documents Would Be Exposed

The `docs/` directory contains 227 files, of which ~91 are internal-facing:
- Status reports (`docs/status/` — 75+ files documenting every session)
- Self-critiques and brutal reviews
- Planning docs with internal task breakdowns
- AI-assisted development artifacts
- Architecture review HTML reports

**Impact:** These expose the internal development process, warts and all. While
transparency is good, the sheer volume (91 files) creates noise and makes the repo
look like a personal journal rather than a professional library. Competitors also
gain free insight into your architecture decisions and self-identified weaknesses.

**Options:**
- Move internal docs to `docs/internal/` with a README explaining they're process docs
- Or `.gitignore` the `docs/status/`, `docs/planning/`, `docs/reviews/` directories
- Or accept the transparency (some projects do this successfully)

### 4. Missing Community Health Files

| Missing File           | Purpose                                         |
| ---------------------- | ----------------------------------------------- |
| `SECURITY.md`          | Vulnerability reporting policy                  |
| `CODE_OF_CONDUCT.md`   | Community standards                             |
| `.github/ISSUE_TEMPLATE/` | Bug report and feature request templates     |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR checklist                        |
| `.github/FUNDING.yml`  | Sponsorship (optional)                          |
| GitHub Topics & Description | Discoverability (repo has no description)  |

### 5. Single-Author Maintenance Risk

Bus factor = 1. Going public creates implicit support obligations:
- Issue triage and response
- PR review
- Semver compatibility promises (harder to change once external consumers depend on APIs)
- Backward compatibility pressure on the 22 existing consumers

Without funding or a team, this can become a time sink. MIT license + "best effort"
support policy in README mitigates this but doesn't eliminate it.

### 6. No pkg.go.dev Documentation Verified

The README has a pkg.go.dev badge, but since the repo is private, the documentation
has never rendered publicly. First `go get` after going public will trigger indexing,
but the doc examples and package docs should be verified to render correctly.

### 7. GoReleaser Homebrew Tap Dependency

The release workflow references `HOMEBREW_TAP_GITHUB_TOKEN` secret, implying a
Homebrew tap repo. Verify this secret and tap repo exist and are configured correctly
before the first public release triggers GoReleaser.

---

## Conditional Recommendation

### Make public AFTER addressing these items (in priority order):

#### Phase 1: Critical Blockers (must fix before `gh repo edit --visibility public`)

| # | Task | Effort | Why |
|---|------|--------|-----|
| 1 | **Migrate `encoding/json/v2` → `encoding/json`** | Medium (4 files + tests) | Without this, `go get` fails for 99% of users |
| 2 | **Fix `PUBLIC_OR_PRIVATE.md` split brain** | Trivial | Remove false "already public" banner |
| 3 | **Remove `GOPRIVATE` warning from README** | Trivial | No longer needed once public |
| 4 | **Verify tests pass without `GOEXPERIMENT`** | Medium | Confirm migration is complete |
| 5 | **Add GitHub repo description + topics** | Trivial | Discoverability |

#### Phase 2: Community Readiness (should fix before announcing)

| #  | Task | Effort | Why |
|----|------|--------|-----|
| 6  | Add `SECURITY.md` | Trivial | Vulnerability reporting policy |
| 7  | Add `CODE_OF_CONDUCT.md` | Trivial | Community standards (Contributor Covenant) |
| 8  | Add `.github/ISSUE_TEMPLATE/` (bug + feature) | Trivial | Structured issue reporting |
| 9  | Add `.github/PULL_REQUEST_TEMPLATE.md` | Trivial | PR quality checklist |
| 10 | Add support policy to README | Trivial | Set expectations (MIT, best-effort) |
| 11 | Decide on internal docs (move/archive/keep) | Low-Med | Reduce noise from 91 internal files |
| 12 | Verify pkg.go.dev renders after first public tag | Low | Documentation discoverability |

#### Phase 3: Launch (do after visibility flip)

| #  | Task | Effort | Why |
|----|------|--------|-----|
| 13 | Tag v2.0.0 (if json/v2 removal is a breaking change) or v1.4.0 | Low | Public version anchor |
| 14 | Verify GoReleaser + Homebrew tap works on public tag | Low | Binary distribution |
| 15 | Write announcement (blog post / r/golang / Go Slack / Twitter) | Medium | Drive adoption |
| 16 | Add to Awesome Go lists | Low | Discoverability |

---

## Assessment Matrix

| Question | Answer |
|----------|--------|
| Is the code production-ready? | **Yes** — v1.3.0, 95%+ coverage, zero-dep core |
| Does it fill a real gap? | **Yes** — no comparable Go library exists |
| Are there known consumers? | **Yes** — 22 projects, 14 with Go code |
| Is the API stable? | **Yes** — frozen since v1.0.0, no deprecated APIs |
| Can users `go get` it today? | **No** — GOEXPERIMENT=jsonv2 blocks import |
| Is the community infrastructure ready? | **Partially** — missing 5 health files |
| Is the git history clean? | **Yes** — no secrets, no binaries, no PII |
| Should you wait? | **Only for json/v2 migration** — everything else is ready |

---

## If You Choose to Stay Private

Valid reasons to delay:
- You need the json/v2 API and don't want to migrate yet
- You're building commercial tooling on top and want to keep the foundation private
- You don't want maintenance burden right now

**None of these outweigh the ecosystem value.** The json/v2 migration is mechanical.
The maintenance burden is manageable with a clear support policy. The first-mover
advantage窗口窗口 is closing.

---

_This document is a living assessment. Update after each phase completion._
