# PRO/CONTRA: Making go-finding Public

> **Date:** 2026-07-24 | **Current state:** Repo is **PRIVATE** (confirmed via GitHub API)
>
> This document supersedes the resolution banner in `PUBLIC_OR_PRIVATE.md`, which
> incorrectly claims the repo is already public. It is not. This is a fresh assessment
> at v1.3.0 with 1,005 commits, 193 Go files, and 112 test files.

---

## Executive Summary

**Verdict: Make public.** The project is production-grade and fills an unfilled gap.

The core library depends on `encoding/json/v2` (Go 1.26 experimental). This is a
**conscious decision to stay on the cutting edge** — the project targets Go 1.26+
users and positions itself as a forward-looking library. The `GOEXPERIMENT=jsonv2`
requirement must be documented prominently so adopters aren't surprised, but it is
not a blocker. When Go stabilizes json/v2 (expected in 1.27+), this becomes a
competitive advantage rather than friction.

| Dimension                  | Rating | Key Finding                                                       |
| -------------------------- | ------ | ----------------------------------------------------------------- |
| Code quality               | **A**  | Clean architecture, branded types, decomposed validators          |
| Test coverage              | **A**  | 112 test files, fuzz, BDD, property, integration, E2E, stress     |
| API stability              | **A**  | v1.0.0 frozen (2026-06-24), v1.3.0 current, no deprecated APIs    |
| CI/CD maturity             | **A**  | 8 CI jobs: test, lint, vulncheck, coverage, stress, dupl, bench   |
| Documentation              | **B+** | Strong README/CONTRIBUTING/FEATURES, but 91 internal docs exposed |
| Ecosystem value            | **A**  | Fills a genuine gap — no comparable Go library exists             |
| Adoption friction          | **B**  | `GOEXPERIMENT=jsonv2` required — targets Go 1.26+ early adopters  |
| Community readiness        | **A-**  | SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates, support policy |
| Maintenance sustainability | **B**  | Single author, 22 known consumers, MIT license                    |

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

| Component         | Status                                                                   |
| ----------------- | ------------------------------------------------------------------------ |
| GitHub Actions CI | 8 jobs: multi-OS test, lint, govulncheck, coverage, stress, dupl, bench  |
| GoReleaser        | Cross-platform binaries (linux/darwin/windows, amd64/arm64), cosign SBOM |
| golangci-lint v2  | 80+ linters, strict config, zero outstanding findings                    |
| Codecov           | Integrated with per-package coverage thresholds                          |
| Dependabot        | All 4 sub-modules + GitHub Actions tracked                               |
| Multi-module      | 4 independent Go modules with replace directives for GOWORK=off CI       |
| Benchmark CI      | Regression detection vs baseline (25% threshold)                         |

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

### 1. GOEXPERIMENT=jsonv2 — Conscious Cutting-Edge Choice

**Severity: Adoption friction (not a blocker)**

The core library imports `encoding/json/v2` in 4 production files:
`json.go`, `sarif_types.go`, `sarif_import.go`, `sarif_export.go`.

This means **every consumer** must:

1. Use Go 1.26+
2. Set `GOEXPERIMENT=jsonv2` environment variable
3. Without it: `build constraints exclude all Go files` — the import fails

**Decision: Stay on json/v2.** This is a deliberate choice to adopt the future Go
JSON standard early. The project targets forward-looking Go 1.26+ users. When Go
stabilizes json/v2 (expected Go 1.27+), this becomes a competitive advantage:
better performance, cleaner API, first in the ecosystem.

**Mitigation required:**

- Document the requirement **prominently** in README (not buried at the bottom)
- Add a "Prerequisites" section at the very top of Installation
- Consider a `go env -w GOEXPERIMENT=jsonv2` one-liner in the quick start
- Track Go json/v2 stabilization timeline and remove the env var when it lands

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

### 4. ~~Missing Community Health Files~~ ✅ RESOLVED

All community health files have been added (2026-07-24):

| File                             | Status    |
| -------------------------------- | --------- |
| `SECURITY.md`                    | ✅ Created |
| `CODE_OF_CONDUCT.md`             | ✅ Created |
| `.github/ISSUE_TEMPLATE/`        | ✅ Created |
| `.github/PULL_REQUEST_TEMPLATE.md` | ✅ Created |
| `.github/FUNDING.yml`            | ⬜ Optional (not added) |
| GitHub Topics & Description      | ✅ Set (11 topics)    |

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

| #   | Task                                             | Status    | Why                                                                                     |
| --- | ------------------------------------------------ | --------- | --------------------------------------------------------------------------------------- |
| 1   | **Fix `PUBLIC_OR_PRIVATE.md` split brain**       | ✅ Done   | Corrected 2026-07-24                                                                    |
| 2   | **Make GOEXPERIMENT=jsonv2 prominent in README** | ✅ Done   | Prerequisites block at top of Installation (2026-07-24)                                 |
| 3   | **Remove `GOPRIVATE` warning from README**       | ✅ Done   | Removed 2026-07-24; no longer needed once public                                        |
| 4   | **Add GitHub repo description + topics**         | ✅ Done   | Description + 11 topics set via `gh repo edit` (2026-07-24)                            |

#### Phase 2: Community Readiness (should fix before announcing)

| #   | Task                                                             | Status    | Why                                                       |
| --- | ---------------------------------------------------------------- | --------- | --------------------------------------------------------- |
| 5   | Add `SECURITY.md`                                                | ✅ Done   | Vulnerability reporting policy (2026-07-24)               |
| 6   | Add `CODE_OF_CONDUCT.md`                                         | ✅ Done   | Contributor Covenant v2.1 (2026-07-24)                    |
| 7   | Add `.github/ISSUE_TEMPLATE/` (bug + feature)                    | ✅ Done   | bug + feature templates + config.yml (2026-07-24)         |
| 8   | Add `.github/PULL_REQUEST_TEMPLATE.md`                           | ✅ Done   | PR quality checklist (2026-07-24)                         |
| 9   | Add support policy to README                                     | ✅ Done   | "Support" section added before Versioning (2026-07-24)   |
| 10  | ~~Decide on internal docs (move/archive/keep)~~ ✅ **No action** | Done      | User: keep all as-is                                      |
| 11  | Verify pkg.go.dev renders after first public tag                 | ⬜ TODO   | Triggered by first `go get` after visibility flip          |
| 12  | Track Go json/v2 stabilization (Go 1.27+)                        | ⬜ TODO   | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes |

#### Phase 3: Launch (do after visibility flip)

| #   | Task                                                           | Effort | Why                   |
| --- | -------------------------------------------------------------- | ------ | --------------------- |
| 13  | Tag v1.4.0 (or next minor)                                     | Low    | Public version anchor |
| 14  | Verify GoReleaser + Homebrew tap works on public tag           | Low    | Binary distribution   |
| 15  | Write announcement (blog post / r/golang / Go Slack / Twitter) | Medium | Drive adoption        |
| 16  | Add to Awesome Go lists                                        | Low    | Discoverability       |

---

## Assessment Matrix

| Question                               | Answer                                                                |
| -------------------------------------- | --------------------------------------------------------------------- |
| Is the code production-ready?          | **Yes** — v1.3.0, 95%+ coverage, zero-dep core                        |
| Does it fill a real gap?               | **Yes** — no comparable Go library exists                             |
| Are there known consumers?             | **Yes** — 22 projects, 14 with Go code                                |
| Is the API stable?                     | **Yes** — frozen since v1.0.0, no deprecated APIs                     |
| Can users `go get` it today?           | **Yes, with `GOEXPERIMENT=jsonv2`** — targets Go 1.26+ early adopters |
| Is the community infrastructure ready? | **Yes** — SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates, support policy (2026-07-24) |
| Is the git history clean?              | **Yes** — no secrets, no binaries, no PII                             |
| Should you wait?                       | **No** — everything is ready; json/v2 is a feature, not a bug         |

---

## If You Choose to Stay Private

Valid reasons to delay:

- You're building commercial tooling on top and want to keep the foundation private
- You don't want maintenance burden right now

**None of these outweigh the ecosystem value.** The maintenance burden is
manageable with a clear support policy. The first-mover advantage window is closing.

---

_This document is a living assessment. Update after each phase completion._
