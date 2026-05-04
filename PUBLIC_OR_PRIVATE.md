# PUBLIC OR PRIVATE? — go-finding

**Date:** 2026-05-04 | **Decision:** CONDITIONAL — Public with prerequisites

---

## Verdict

**Open-source it** — but only after addressing the pre-release checklist below.

This project is a well-architected, genuinely useful library that fills a real gap in the Go ecosystem. It is not a half-finished prototype. The code quality, test coverage, documentation, and CI/CD maturity all support public release. The question is not _whether_, but _when_ and _with what polish_.

---

## PRO — Arguments for Making Public

### 1. Fills a Real Ecosystem Gap

> "Seven tools detect issues. Zero tools route them to remediation."

This is not a solved problem. Go has `staticcheck`, `govet`, `golangci-lint`, etc. — all producing different output formats. No standard interchange type or automated fix pipeline exists. `go-finding` addresses both. This is genuinely novel in the Go ecosystem.

### 2. Production-Grade Code Quality

- **21,000+ lines of source** across root package, pipeline, CLI, and internal detectors
- **15,000+ lines of tests** — 58 test files including unit, BDD, fuzz, integration, E2E, benchmark, coverage enforcement, and property-based tests
- Clean separation: core types (zero stdlib deps), pipeline (minimal deps), CLI (user-facing)
- Consistent error handling with structured `FindingError` and sentinel `errors.Is` support
- Thread-safe `Report` with mutex protection, `iter.Seq` iterators, `maps.Clone` deep copies
- Comprehensive validation: `Validate()`, `IsValid()`, `Key()`, `Equal()`, `Compare()`

### 3. Well-Designed Type System

- `Finding` is immutable data — no state machines, no mutation
- `Severity`, `FixStrategy`, `Category`, `Tag` — all string-based enums with `IsValid()` guards
- `Position` / `Range` with geometric operations: `Contains`, `Overlaps`, `Intersection`, `Adjacent`
- `Compare` methods on `Position`, `Range`, `Severity` — total ordering, sort-friendly
- `Clone()` deep copies on `Finding` — no accidental sharing
- Lossless SARIF round-trip via `properties` bag (except suppression)

### 4. Complete Interchange Format Support

- SARIF 2.1.0 export/import with documented lossiness
- LSP Diagnostic conversion (both directions)
- `go/analysis.Diagnostic` integration
- JSON marshaling with validation and invalid-filtering
- Builder API for fluent construction

### 5. Pipeline Is Feature-Complete

- Detect → Triage → Fix → Verify loop with configurable max iterations
- Parallel detection via `errgroup`
- Conflict detection for overlapping fixes
- Line-based fix application with backup/rollback
- Exponential backoff retry with jitter
- Graceful degradation (partial success)
- Composable `FindingProcessor` chain
- Cross-tool correlation (`Correlate`)
- Metrics collection with thread-safe snapshots
- Dry-run mode

### 6. Excellent Infrastructure

- GitHub Actions CI: test (Ubuntu + macOS), lint, govulncheck, coverage enforcement, stress tests
- GoReleaser for cross-platform binary releases (Linux/macOS/Windows, amd64/arm64)
- golangci-lint v2.10 with strict config
- Codecov integration
- 443 commits — active, iterative development
- MIT License — permissive, ecosystem-friendly

### 7. Documentation Is Above Average

- README with quick start, builder API, core types, filtering, pipeline usage
- CONTRIBUTING.md with dev setup, test commands, PR process
- CONTEXT.md with domain language
- FEATURES.md with feature matrix
- CHANGELOG.md tracking changes
- Architecture decision records in `docs/`
- Usage guide, integration guide, migration guide
- Readiness report already exists (`docs/READINESS_REPORT.md`)

### 8. Go Ecosystem Fit

- Module path is `github.com/larsartmann/go-finding` — proper public module
- Go 1.26 — modern, no legacy constraints
- Minimal dependencies: `golang.org/x/tools`, `golang.org/x/sync`, `gopkg.in/yaml.v3`
- BDD with Ginkgo/Gomega — Go-native testing frameworks
- Works with `go/analysis` framework — familiar to Go tool authors

### 9. Potential for Community Adoption

Any Go developer building or consuming static analysis tools could benefit:

- Tool authors get a standard output format (SARIF, LSP, JSON) for free
- CI/CD pipelines get structured findings instead of parsing `grep` output
- IDE integrations get LSP diagnostics without per-tool adapters
- The `pipeline` package enables automated fix workflows

### 10. First-Mover Advantage

No comparable Go library exists. Publishing now establishes `go-finding` as the standard before a competitor fills the gap.

---

## CONTRA — Arguments for Staying Private

### 1. Single Author Risk

- One contributor (Lars Artmann). Bus factor = 1.
- Maintenance burden: issues, PRs, semver compatibility promises, breaking changes
- API stability guarantees are harder once external consumers exist

### 2. Pre-Release Gaps

| Gap                                                                 | Severity |
| ------------------------------------------------------------------- | -------- |
| No pkg.go.dev documentation (unpublished)                           | High     |
| No API stability guarantee / versioning policy                      | Medium   |
| `coverage.out` and `cover.out` committed to repo                    | Low      |
| Binary `go-finding` committed to repo root (5.8MB)                  | Medium   |
| `go.work` and `go.work.sum` committed (workspace files)             | Low      |
| `git-town.toml` in repo root (personal tool config)                 | Low      |
| `testdata/` appears empty/uncommitted                               | Low      |
| `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md` — internal doc in public repo | Low      |

### 3. No Known External Users Yet

No validation from real-world usage outside the author's own tools. The API may need iteration based on actual consumer feedback — easier while private.

### 4. Competitive Exposure

Publishing reveals the architecture and approach to potential competitors. Given the first-mover advantage argument, this cuts both ways — but if the project isn't ready for adoption, someone could ship a simpler competitor faster.

### 5. Support Expectations

Public repos create implicit support obligations. Without funding or a team, issue triage and PR review can become a time sink.

### 6. The Pipeline's Fix Applier Touches Files

`FixApplier` writes to disk with backup/rollback — this is powerful but dangerous. A public release needs thorough documentation of safety guarantees and a clear disclaimer. Edge cases in concurrent file modification or partial failures could damage user trust early.

### 7. SARIF Round-Trip Has Known Limitations

Suppression data is lost on export. `SeverityCritical` maps to SARIF "error" (lossy). These are documented but may confuse new users.

---

## Conditional Recommendation

### Make public IF the following are addressed:

- [ ] **Remove committed artifacts:** Delete `coverage.out`, `cover.out`, `go-finding` binary from git history (or `.gitignore` them)
- [ ] **Add `.gitignore` entries** for `coverage.out`, `cover.out`, `go-finding` binary, `go.work`, `go.work.sum`
- [ ] **Remove personal tool configs:** `.gitignore` or delete `git-town.toml`
- [ ] **Tag v0.1.0** (or v1.0.0 if confident in API stability) and push the tag
- [ ] **Add versioning policy** to README: "Until v1.0, minor versions may include breaking changes"
- [ ] **Ensure pkg.go.dev renders** — push a tag and verify documentation appears
- [ ] **Write a blog post or Twitter thread** explaining the "seven tools, zero routers" problem and how `go-finding` solves it
- [ ] **Consider moving internal proposal docs** (`MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`, `PROPOSAL.md`) to a `docs/internal/` directory or removing them

### Timeline

| When         | What                                                        |
| ------------ | ----------------------------------------------------------- |
| **Now**      | Fix git hygiene (remove artifacts, update .gitignore)       |
| **Now**      | Tag v0.1.0                                                  |
| **Week 1**   | Verify pkg.go.dev rendering, write announcement             |
| **Week 2-4** | Make repo public, announce on r/golang, Go Slack, Twitter/X |
| **Month 2+** | Iterate based on community feedback, aim for v1.0           |

### Staying Private Makes Sense IF:

- You need 3+ more months to stabilize the API based on your own consumption
- You plan to build commercial tooling on top and want to keep the foundation private
- You don't want maintenance burden right now

Given the current maturity (443 commits, 15K test lines, full CI/CD, comprehensive docs), **none of these reasons apply strongly**. The project is ready for public consumption with minor cleanup.

---

## Final Assessment

| Dimension                  | Rating | Notes                                                  |
| -------------------------- | ------ | ------------------------------------------------------ |
| Code quality               | **A**  | Clean, idiomatic, well-structured                      |
| Test coverage              | **A**  | Unit, BDD, fuzz, integration, E2E, benchmarks          |
| Documentation              | **A-** | Excellent for a private project; needs pkg.go.dev      |
| API design                 | **A**  | Composable, lossless, well-typed                       |
| CI/CD maturity             | **A**  | Full pipeline: test, lint, vulncheck, coverage, stress |
| Release readiness          | **B+** | Binary artifacts in repo, no tag yet                   |
| Ecosystem value            | **A**  | Fills a genuine gap                                    |
| Maintenance sustainability | **B**  | Single author, no funding                              |

**Open it. The Go ecosystem needs this.**
