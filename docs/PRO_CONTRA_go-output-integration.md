# PRO/CONTRA: Integrating go-output into go-finding

**Date:** 2026-05-01 | **Status:** Implemented 2026-06-23

---

## Executive Summary

| Dimension        | go-finding                                                   | go-output                                            |
| ---------------- | ------------------------------------------------------------ | ---------------------------------------------------- |
| **Domain**       | Static analysis data model + pipeline                        | Output formatting (12 formats)                       |
| **Core types**   | `Finding`, `Report`, `Severity`, `Position`                  | `TableData`, `TreeNode`, `Renderer`, `Format`        |
| **Dependencies** | 0 production (core is stdlib-only); test-only: ginkgo/gomega | 3 (lipgloss, go-faster/yaml, x/term) + 15 transitive |
| **Consumers**    | 22 tools (art-dupl, branching-flow, etc.)                    | 2 tools (project-meta, projects-management)          |
| **Maturity**     | v1.3.0, 93.6% coverage, API-stable since v1.0.0              | Production-ready, 91%+ coverage                      |
| **Philosophy**   | "Minimal dependencies — core types depend only on stdlib"    | Full-featured formatting with lipgloss styling       |

---

## Three Integration Models

### A. go-finding depends on go-output (add output formats)

go-finding's CLI gains markdown tables, terminal tables, CSV, etc. for rendering findings.

### B. go-output depends on go-finding (add Finding renderers)

go-output ships a `finding` subpackage with `FindingTableData()`, etc.

### C. Merge into one repo/monorepo

Combine both under a single module or workspace.

---

## PRO (all models)

| #   | Argument                                                                                                                                                                                          | Weight | Model |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ----- |
| 1   | **Richer CLI output** — go-finding's CLI currently has only `text`/`json`/`sarif`. go-output adds markdown tables, terminal tables (lipgloss), CSV, TSV, XML, YAML, DOT, Mermaid, D2 diagrams     | High   | A, C  |
| 2   | **Deduplication of format logic** — go-finding's `sarif.go` and `json.go` hand-roll serialization. go-output has tested JSON/CSV/TSV/XML/YAML/markdown renderers with escaping. Avoid reinventing | Medium | A, C  |
| 3   | **Single dependency for consumers** — Tools like branching-flow that already use both get one import instead of two                                                                               | Low    | C     |
| 4   | **Finding→Table adapter is natural** — `Report` → `TableData` is a trivial adapter: headers = `[File, Line, Severity, Rule, Message]`, rows from findings. This is the obvious integration point  | Medium | A, B  |
| 5   | **Graph visualization of correlations** — go-output's D2/Mermaid/DOT renderers could visualize `Correlation` and `RelatedRef` chains as graphs. Unique value-add                                  | Medium | A     |
| 6   | **Consistent CLI UX across your ecosystem** — All tools using go-output get the same `Format`/`SortBy`/`ColorMode` flags. go-finding benefits from this consistency                               | Medium | A     |
| 7   | **Co-versioning** — No version skew between "the output library" and "the finding library" when types change                                                                                      | Low    | C     |

---

## CONTRA (all models)

| #   | Argument                                                                                                                                                                                                                                                                                                     | Weight       | Model   |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ | ------- |
| 1   | **Breaks go-finding's core design principle** — AGENTS.md: "Minimal dependencies — core types depend only on stdlib." go-output brings lipgloss + 15 transitive deps. This is the strongest contra                                                                                                           | **Critical** | A, C    |
| 2   | **Domain mismatch** — go-finding is a _data model + pipeline_. go-output is a _presentation layer_. These are separate concerns. Merging violates SRP                                                                                                                                                        | High         | B, C    |
| 3   | **Consumer coupling** — 7 tools depend on go-finding (art-dupl, branching-flow, hierarchical-errors, etc.). Adding go-output as a dependency forces all of them to pull in lipgloss, even if they only want the data model. Library consumers who embed go-finding types don't need terminal table rendering | **Critical** | A, C    |
| 4   | **go-output has narrow adoption** — Only 2 consumers (project-meta, projects-management-automation). The formatting library isn't battle-tested across diverse domains yet                                                                                                                                   | Medium       | A, B, C |
| 5   | **The adapter is trivial** — `Report → TableData` is ~10 lines of code. Writing this once in go-finding's CLI package is simpler than importing a 12-format library                                                                                                                                          | Medium       | A       |
| 6   | **go-output's dependencies are heavy for a CLI** — lipgloss v2 brings 15 transitive dependencies (charmbracelet/x/\*, clipperhouse/\*, etc.). go-finding currently has 0 transitive deps in its core                                                                                                         | High         | A, C    |
| 7   | **Separate release cadences** — go-output is iterating on renderers and escaping. go-finding is stabilizing its API for v1.0. Coupling them creates coordination overhead                                                                                                                                    | Medium       | C       |
| 8   | **CLI-only need** — The only place go-finding needs formatting is `cmd/go-finding/main.go`. Core types, pipeline, and library consumers never need it. The dependency would be top-heavy                                                                                                                     | High         | A, C    |
| 9   | **SARIF is not a "format"** — SARIF is a standard interchange format with specific semantics. It shouldn't be "just another renderer" in go-output. It belongs in go-finding where the domain model lives                                                                                                    | Medium       | B, C    |
| 10  | **go-output's `enum` package overlaps** — go-finding has its own enum patterns for Severity, FixStrategy, etc. go-output's `enum` package is generic but adopting it would create a dependency for a trivial utility                                                                                         | Low          | A       |

---

## Analysis by Model

### Model A: go-finding → go-output (dependency)

**Verdict: NO (for core) / YES (for CLI only)**

The right approach is **optional CLI dependency**: go-finding's `cmd/go-finding` can import go-output, while the root `finding` package stays dependency-free. This is already the pattern — the CLI already depends on `gopkg.in/yaml.v3` while core doesn't.

Implementation:

```
go-finding/
├── finding/       # No change — zero deps
├── pipeline/      # No change
├── cmd/go-finding/
│   └── main.go    # Add go-output as CLI-only dep
```

### Model B: go-output → go-finding (dependency)

**Verdict: NO**

go-output is a general-purpose formatting library. Adding static analysis domain types as a dependency narrows its applicability and confuses its purpose.

### Model C: Monorepo merge

**Verdict: NO**

Two different domains, two different maturity levels, two different consumer bases. Merging creates a jack-of-all-trades package that's hard to reason about.

---

## Recommendation

**Model A (CLI-only dependency) with an adapter pattern — IMPLEMENTED.**

| Step | Action                                                                     | Status |
| ---- | -------------------------------------------------------------------------- | ------ |
| 1    | Add `go-output` (root + markdown + delimited) as CLI-only deps             | Done   |
| 2    | Write `findingToTableData(findings) *output.TableData` adapter in CLI      | Done   |
| 3    | Add `csv`, `tsv` formats; replace hand-rolled markdown with go-output      | Done   |
| 4    | Keep core `finding` package dependency-free                                | Done   |
| 5    | Ship `D2`/`Mermaid` visualization of `Correlation` chains as bonus feature | Future |

### What was implemented (2026-06-23)

- **Deps:** `go-output` v0.17.2, `go-output/markdown` v0.17.2, `go-output/delimited` v0.17.2 — all lightweight (root has only `x/term`, no lipgloss/bubbletea).
- **Adapter:** `cmd/go-finding/output_adapter.go` — `findingToTableData()` converts `[]Finding` → `*output.TableData` with columns: Location, Severity, Rule, Message.
- **New formats:** `csv` (streaming writer, auto-quoting), `tsv` (tab-separated). Both include a footer row with finding count.
- **Improved markdown:** go-output's `MarkdownTable` provides auto-aligned column widths, replacing the hand-rolled `finding.FormatMarkdown` in the CLI path.
- **Unchanged:** `text`, `json`, `sarif` remain hand-rolled (domain-specific). `finding.FormatText` and `finding.FormatMarkdown` stay in root package for library consumers.

### Original concerns, now mitigated

| Original concern              | Resolution                                                                                                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| lipgloss + 15 transitive deps | go-output restructured to multi-module; root has only `x/term`. Only `markdown` and `delimited` sub-modules imported — neither pulls lipgloss. |
| API stability                 | go-output v0.17.2, API frozen via ADR 006                                                                                                      |
| Narrow adoption               | Now at v0.17.2 with 16 formats, integration + BDD test suites                                                                                  |

---

## What NOT to do

- Do NOT add go-output as a dependency of the root `finding` package
- Do NOT merge the repositories
- Do NOT make go-output depend on go-finding
- Do NOT try to replace SARIF with go-output renderers — SARIF has domain-specific semantics that go-output's table/tree/graph model doesn't capture
