# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term, bounded work see [TODO_LIST.md](TODO_LIST.md).
> For shipped features and their status see [FEATURES.md](FEATURES.md).

---

## Current Phase: Consumer ecosystem growth

**Current version:** 1.3.0

v1.0.0 locked the API (2026-06-24). v1.1.0 added multi-module workspace, branded type safety, SARIF suppression round-trip, LSP data fidelity. v1.2.0 extracted `lockutil`, defragmented tests. v1.2.1 shipped 15+ correctness/security fixes, `encoding/json/v2` migration. v1.3.0 added 12 consumer-driven convenience APIs based on a full audit of 22 consumer projects.

The library is production-ready and API-stable. The focus now shifts to growing the consumer ecosystem and expanding language coverage.

---

## v1.0.0–v1.3.0 — API lock + consumer convenience ✅

**Status: Released.**

- ✅ API locked at v1.0.0. Breaking changes require major version bump.
- ✅ Multi-module workspace (core, pipeline, analysis, CLI) established.
- ✅ Branded types (`ID`, `RuleName`, `ToolName`, `FilePath`), hand-rolled SARIF, LSP round-trip fidelity.
- ✅ v1.3.0: Consumer-driven APIs (`BuildOrDefault`, `Template`, `NewReportFromFindings`, `FilePos`, `SeverityFromLevel`, `ApplySimpleFixes`, `CheckBinary`/`RunCmd`, `FormatTable`, `PriorityString`). File-level position validation relaxed.

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

Each would live in its own subpackage to keep language-specific dependencies out of the core.

### Tooling integrations

- **IDE plugins** — VS Code / Neovim consuming LSP diagnostics from `ToLSP()`
- **Watch mode** — re-run the pipeline on file change (`fsnotify`)
- **Interactive TUI** — triage and review findings before applying fixes
- **GitHub Actions action** — first-class SARIF upload with fix PR generation

### Consumer ecosystem

- **Consumer migration to v1.3.0 APIs** — 14 Go consumers can now simplify their codebases using `BuildOrDefault`, `Template`, `SeverityFromLevel`, `FilePos`, `NewReportFromFindings`, and `ApplySimpleFixes`. Each consumer independently reinvented these patterns.
- **More `ToolAdapter[O]` recipes** — Pre-built adapters for revive, ineffassign, errcheck, etc.

### Hardening (owner decisions pending)

These are known design tensions deferred because they require breaking changes:

- **Position zero-value** — `Position{}` has `Offset=0` (valid byte 0), not "unset". Resolved pragmatically in v0.9.0 with `-1` sentinel, but a type-safe redesign is still on the table for v2.0.
- **`Range.End` zero-value ambiguity** — same class of issue as Position.
- **SARIF schema validation** — blocked on vendoring the 7K-line SARIF 2.1.0 JSON schema for test-time validation.

---

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Web UI** — Not aligned with the library's core purpose (data model + pipeline, not presentation layer).
- **Hosted/SaaS offering** — Too costly relative to impact for a library project.
- **Non-Go language providers (pre-v2)** — Language expansion is roadmap, but not until the Go provider ecosystem is fully proven.
- **Generic `JSONToolDetector`** — Too opinionated; every tool's JSON shape differs. `CheckBinary`/`RunCmd` helpers suffice.
- **`Properties map[string]any`** — Explicitly banned. `Metadata map[string]string` stays for type safety and interchange simplicity.

---

_Assisted-by: Crush <crush@charm.land>_
