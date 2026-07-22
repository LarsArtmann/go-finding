# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term, bounded work see [TODO_LIST.md](TODO_LIST.md).
> For shipped features and their status see [FEATURES.md](FEATURES.md).

---

## Current Phase: Post-v1 stabilization

**Current version:** 1.2.1

v1.0.0 locked the API (2026-06-24). v1.1.0 added the multi-module workspace, branded type safety (`FilePath`), SARIF suppression round-trip, and LSP data fidelity. v1.2.0 extracted the shared `lockutil` package and defragmented the test suite. v1.2.1 shipped 15+ correctness/security fixes (range geometry, path traversal, SARIF/LSP round-trip fidelity), migrated to `encoding/json/v2`, and completed a comprehensive skills audit sweep (8 review reports + 2 architecture diagrams). The library is production-ready and API-stable.

---

## v1.0.0 — API lock ✅

**Status: Released 2026-06-24.**

- ✅ **Remove deprecated APIs** — All deprecated APIs removed: `Report.Findings` (unexported), `Report.Merge()`, `OnStage`, `Metrics.RecordFix()`, `CountBySeverity()` free function, `SeverityAliases()`, `GetCategory()`, `HasFix`/`HasSuggestion` free functions. See [removed API table](docs/RELEASE_CRITERIA.md#deprecated-api-removal--completed-).
- ✅ **Position/Range zero-value redesign** — `-1` offset sentinel adopted (v0.9.0, confirmed in v1.0.0).
- ✅ **`FixStrategyAI` fate** — kept as reserved marker.

After v1.0.0, breaking changes require a major version bump per SemVer.

---

## Raw Ideas (not yet actionable)

These are directions worth exploring. They are **not** committed work — they exist to capture thinking before it is lost. When an idea becomes concrete enough to act on, it graduates to [TODO_LIST.md](TODO_LIST.md).

### AI-assisted remediation

`FixStrategyAI` is a reserved constant with no backend. A real implementation would need:

- A pluggable `AIProvider` interface (request → suggested diff)
- Guardrails: sandboxed apply, verification re-run, human approval gate
- Cost/rate-limit awareness in the pipeline

This is the largest open product direction and the original motivation for the library's fix pipeline.

### Language expansion

The core `Finding` model and SARIF/LSP interchange are language-agnostic. The fix engine has a Go AST provider (`pipeline/goast/`) but the provider architecture supports more:

- Rust (`syn`-based provider)
- TypeScript/JavaScript (tree-sitter)
- Python (ast / libcst)

Each would live in its own subpackage to keep `go/parser`-style opt-in dependencies out of the core.

### Tooling integrations

- **IDE plugins** — VS Code / Neovim consuming LSP diagnostics from `ToLSP()`
- **Watch mode** — re-run the pipeline on file change (`fsnotify`)
- **Interactive TUI** — triage and review findings before applying fixes
- **GitHub Actions action** — first-class SARIF upload with fix PR generation

### Ecosystem

- **`go-structure-linter` integration** — Wire go-finding as the finding model for LarsArtmann's structure linter
- **More `ToolAdapter[O]` recipes** — Pre-built adapters for revive, ineffassign, errcheck, etc.

### CI Hardening

- **Dependabot `groups:` config** — Group gomod + github-actions updates to reduce PR noise
- **SHA-pin GitHub Actions** — Pin actions to commit SHAs instead of major version tags
- **`CODECOV_TOKEN` or OIDC** — Ensure codecov uploads are authenticated, not silently failing
- **`version.go` CI gate** — Compare `git describe --tags` against `finding.Version` to prevent stale version constants (recurring incident: v1.1.0 binary reported "1.0.0")

### Hardening (owner decisions pending)

These are known design tensions deferred because they require breaking changes:

- **Position zero-value** — `Position{}` has `Offset=0` (valid byte 0), not "unset". Resolved pragmatically in v0.9.0 with `-1` sentinel, but a type-safe redesign is still on the table.
- **`Range.End` zero-value ambiguity** — same class of issue as Position.
- **SARIF schema validation** — blocked on vendoring the 7K-line SARIF 2.1.0 JSON schema for test-time validation.

---

## Out of Scope (v1)

Explicitly excluded from the v1.0.0 release:

- Web UI
- Hosted/SaaS offering
- Non-Go language providers (post-v1)

---

_Assisted-by: Crush <crush@charm.land>_
