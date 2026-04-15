# Module Split Plan — go-finding

**Status:** Proposal | **Date:** 2026-04-15

---

## Context

go-finding is currently a single Go module (`github.com/larsartmann/go-finding`) with one sub-package (`pipeline/`), a CLI (`cmd/go-finding/`), and example programs (`examples/`). The root `go.mod` declares three external dependencies:

| Dependency | Used by | Weight |
|---|---|---|
| `golang.org/x/tools` | `diagnostic.go` only | **Heavy** — pulls in go/analysis, go/packages, many transitive deps |
| `golang.org/x/sync` | `pipeline/pipeline.go`, `pipeline/partial.go` | Light |
| `gopkg.in/yaml.v3` | `cmd/go-finding/main.go` only | Light |

**Key insight:** The root package (everything except `pipeline/` and `cmd/`) is **stdlib-only**. The only file that needs `golang.org/x/tools` is `diagnostic.go`. This creates an unnecessary heavy dependency for users who only want the data model.

---

## Current Dependency Graph

```
                    ┌─────────────────────────────────────┐
                    │         Root Package                  │
                    │  (github.com/larsartmann/go-finding)  │
                    │                                       │
  Leaf types ──►   │  severity.go  (0 deps)                │
  (no intra-pkg)   │  fix_strategy.go  (0 deps)           │
                    │  category.go  (0 deps)                │
                    │  suppression.go  (0 deps)            │
                    │  position.go  (0 deps)                │
                    │                                       │
  Core type ──►    │  finding.go  (→ all leaf types)       │
                    │                                       │
  Utilities ──►    │  errors.go  (→ finding, position)     │
                    │  id.go  (→ position)                 │
                    │  filter.go  (→ finding + leaf types) │
                    │                                       │
  Aggregation ──►  │  report.go  (→ finding, filter)      │
                    │  merge.go  (→ report, filter)        │
                    │                                       │
  Format adapters  │  json.go  (→ finding, report)        │
  (methods on      │  sarif.go  (→ finding, report)       │  ← stdlib only
  core types)      │  lsp.go  (→ finding, position)       │
                    │                                       │
  Integration ──►  │  diagnostic.go  (→ finding, id)       │  ← golang.org/x/tools
                    └──────────────┬──────────────────────┘
                                   │ imports
                    ┌──────────────▼──────────────────────┐
                    │     Pipeline Package                  │
                    │  (pipeline/)                          │
                    │  pipeline.go, conflict.go, verify.go  │
                    │  retry.go, metrics.go, partial.go     │  ← golang.org/x/sync
                    └──────────────────────────────────────┘
                                   │ imports
                    ┌──────────────▼──────────────────────┐
                    │     CLI (cmd/go-finding/)             │
                    │  main.go                              │  ← gopkg.in/yaml.v3
                    └──────────────────────────────────────┘
```

### Method-on-Type Coupling (Critical Constraint)

Go requires methods on a type to be defined in the **same package** as the type. The following methods on core types exist in format/integration files:

| File | Method | On Type | Impact if moved |
|---|---|---|---|
| `sarif.go` | `Report.ToSARIF()` | `*Report` | Must convert to function |
| `sarif.go` | `Report.ToSARIFFiltered()` | `*Report` | Must convert to function |
| `lsp.go` | `Finding.ToLSP()` | `Finding` | Must convert to function |
| `json.go` | `Report.PrettyJSON()` | `*Report` | Must convert to function |
| `json.go` | `Finding.LineJSON()` | `Finding` | Must convert to function |
| `diagnostic.go` | `Finding.AnalysisDiagnostic()` | `Finding` | Must convert to function |

---

## Three Approaches

### Approach A: Minimal — Formalize Pipeline (2 modules)

Only give `pipeline/` its own `go.mod`. Root stays intact.

```
go-finding/
├── go.mod                           # Module 1: Core (all root files)
│   require: golang.org/x/tools      # Still heavy — unchanged
│   require: golang.org/x/sync       # Still pulls sync for nothing
│
├── pipeline/
│   ├── go.mod                       # Module 2: Pipeline
│   │   require: go-finding
│   │   require: golang.org/x/sync
│   └── *.go
```

**PROs:**
- Minimal change — only `pipeline/` gets a `go.mod`
- No API breakage at all
- Pipeline can be versioned independently
- `golang.org/x/sync` moves to pipeline's `go.mod`

**CONs:**
- Root still has `golang.org/x/tools` dependency (the heaviest one)
- Users who only want the data model still pull go/analysis transitive deps
- No real dependency boundary improvement for the root package
- `gopkg.in/yaml.v3` stays in root's `go.mod` even though only CLI uses it

**Effort:** ~1 hour

---

### Approach B: Dependency-Driven — Extract Heavy Deps (4 modules) ⭐ RECOMMENDED

Split by dependency boundaries. Each module only declares what it actually needs.

```
go-finding/
├── go.mod                           # Module 1: Core types + utilities
│   require: (none — stdlib only)    # ← ZERO external dependencies
│
│   Files: finding.go, severity.go, fix_strategy.go, position.go,
│          category.go, suppression.go, errors.go, id.go,
│          filter.go, report.go, merge.go,
│          json.go, sarif.go, lsp.go
│
├── analysis/
│   ├── go.mod                       # Module 2: go/analysis integration
│   │   require: go-finding          # (root)
│   │   require: golang.org/x/tools  # Only consumer of this heavy dep
│   └── diagnostic.go
│
├── pipeline/
│   ├── go.mod                       # Module 3: Pipeline orchestration
│   │   require: go-finding          # (root)
│   │   require: golang.org/x/sync
│   └── *.go
│
├── cmd/go-finding/
│   ├── go.mod                       # Module 4: CLI binary
│   │   require: go-finding          # (root)
│   │   require: go-finding/pipeline # (pipeline module)
│   │   require: gopkg.in/yaml.v3
│   └── main.go
│
└── go.work                          # Workspace for local development
```

**Dependency flow:**

```
  cmd/go-finding ──► pipeline ──► core (go-finding)
       │                                ▲
       └────────────────────────────────┘
       │
       └──► analysis ──► core (go-finding)
```

No circular dependencies. Clean DAG.

**API changes required:**

Only `analysis/diagnostic.go` needs changes — it moves from the root package to `analysis/`:

| Current | New | Breaking? |
|---|---|---|
| `finding.FromDiagnostic(...)` | `analysis.FromDiagnostic(...)` | Yes — import path change |
| `finding.FromTokenPosition(...)` | `analysis.FromTokenPosition(...)` | Yes — import path change |
| `finding.NodePosition(...)` | `analysis.NodePosition(...)` | Yes — import path change |
| `finding.NodeRange(...)` | `analysis.NodeRange(...)` | Yes — import path change |
| `finding.FormatDiagnostic(...)` | `analysis.FormatDiagnostic(...)` | Yes — import path change |
| `finding.Finding.AnalysisDiagnostic()` | `analysis.ToDiagnostic(finding.Finding)` | Yes — method → function |

All other API surface is preserved. `Finding.ToLSP()`, `Report.ToSARIF()`, `Report.PrettyJSON()`, etc. stay in root package.

**Who is affected by breaking changes:**
- Only users of `FromDiagnostic`, `AnalysisDiagnostic`, `NodePosition`, `NodeRange`, `FormatDiagnostic`
- These are go/analysis integration functions — a small, specific audience
- Migration is mechanical: change import path + convert method to function

**PROs:**
- **Root module becomes zero-dependency** — the biggest win
- Users who only need the data model + SARIF/LSP/JSON get zero transitive deps
- `golang.org/x/tools` (heavy) isolated to `analysis/` module — opt-in only
- `golang.org/x/sync` isolated to `pipeline/` module — opt-in only
- `gopkg.in/yaml.v3` isolated to CLI — not a library dep at all
- Pipeline already a separate package — natural boundary
- CLI can be versioned on its own schedule
- Each module's `go.mod` honestly reflects its actual dependencies
- All format methods (SARIF, LSP, JSON) preserved — no API break for most users
- Follows Go team's guidance: split by dependency boundaries

**CONs:**
- 6 functions change import paths (all go/analysis integration)
- `Finding.AnalysisDiagnostic()` method becomes `analysis.ToDiagnostic()` function
- More `go.mod` files to maintain
- Version tagging for subdirectory modules uses prefix format (`analysis/v1.2.3`)
- Slightly more complex CI setup (test each module independently)
- `analysis/` module is thin (1 file) — arguably too small to be its own module

**Effort:** ~4-6 hours (move file, update imports, create go.mod files, go.work, update CI, update examples)

---

### Approach C: Full Modularization — Extract Formats Too (5+ modules)

Same as Approach B, but also extract format adapters (SARIF, LSP, JSON) into a separate `format/` module.

```
go-finding/
├── go.mod                           # Module 1: Core types ONLY
│   require: (none)
│   Files: finding.go, severity.go, fix_strategy.go, position.go,
│          category.go, suppression.go, errors.go, id.go,
│          filter.go, report.go, merge.go
│
├── format/
│   ├── go.mod                       # Module 2: Output format adapters
│   │   require: go-finding           # (root)
│   └── sarif.go, lsp.go, json.go
│
├── analysis/
│   ├── go.mod                       # Module 3: go/analysis integration
│   │   require: go-finding
│   │   require: golang.org/x/tools
│   └── diagnostic.go
│
├── pipeline/
│   ├── go.mod                       # Module 4: Pipeline orchestration
│   │   require: go-finding
│   │   require: golang.org/x/sync
│   └── *.go
│
├── cmd/go-finding/
│   ├── go.mod                       # Module 5: CLI
│   │   require: go-finding, pipeline, format, gopkg.in/yaml.v3
│   └── main.go
│
└── go.work
```

**API changes required (in addition to Approach B):**

| Current | New | Breaking? |
|---|---|---|
| `report.ToSARIF()` | `sarif.Encode(report)` or `sarif.FromReport(report)` | Yes — method → function |
| `report.ToSARIFFiltered(sev)` | `sarif.EncodeFiltered(report, sev)` | Yes |
| `finding.ToLSP()` | `lsp.FromFinding(finding)` | Yes |
| `finding.FromLSP(...)` | `lsp.FromLSP(...)` | Yes — already a function |
| `report.PrettyJSON()` | `jsonfmt.Pretty(report)` | Yes — method → function |
| `finding.LineJSON()` | `jsonfmt.Line(finding)` | Yes |
| `finding.FromJSON(...)` | `jsonfmt.FromJSON(...)` | Yes — already a function |
| `finding.ReportFromJSON(...)` | `jsonfmt.ReportFromJSON(...)` | Yes |
| `finding.FindingsFromJSON(...)` | `jsonfmt.FindingsFromJSON(...)` | Yes |
| `finding.FromSARIFLevel(...)` | `sarif.FromLevel(...)` | Yes |
| All SARIF types (`SarifLog`, etc.) | `sarif.Log`, `sarif.Run`, etc. | Yes — type renames |
| All LSP types (`LSPDiagnostic`, etc.) | `lsp.Diagnostic`, `lsp.Range`, etc. | Yes — type renames |

**PROs:**
- Maximum dependency isolation
- Core module is the absolute minimum (types + utilities)
- Format adapters are independently versionable
- Cleanest conceptual separation: types vs. serialization vs. integration vs. orchestration

**CONs:**
- **14+ API breaking changes** — massive migration burden
- Methods → functions is a significant ergonomic regression (`report.ToSARIF()` → `sarif.Encode(report)`)
- SARIF/LSP/JSON types lose their prefix when moved to subpackage (`SarifLog` → either `sarif.SarifLog` redundant or `sarif.Log` ambiguous)
- Format module is still stdlib-only — no actual dependency savings vs. keeping in root
- Two of three format adapters (SARIF, JSON) only need stdlib `encoding/json` — no dependency to isolate
- Violates Go convention: format conversion methods on types are idiomatic Go
- Examples all need import path updates
- Very high migration cost for very low dependency payoff

**Effort:** ~12-16 hours (significant refactoring + test updates + documentation)

---

## Comparison Matrix

| Criterion | A: Minimal | B: Dependency-Driven ⭐ | C: Full Modularization |
|---|---|---|---|
| External deps in root module | 3 (unchanged) | **0** | **0** |
| `golang.org/x/tools` isolated | No | **Yes** | Yes |
| Breaking API changes | 0 | 6 (go/analysis only) | 14+ |
| Method → function conversions | 0 | 1 | 5 |
| Modules to maintain | 2 | 4 | 5+ |
| User migration effort | None | Low (only go/analysis users) | High (all users) |
| Dependency payoff | Low | **High** | Same as B for deps |
| Conceptual clarity | Low | **High** | Very high |
| Risk of over-splitting | None | Low | High |
| Effort | ~1h | ~4-6h | ~12-16h |
| Follows Go team guidance | Partially | **Yes** | Over-engineered |

---

## Recommendation: Approach B

**Why:** It delivers the highest value (zero-dependency core) at the lowest cost (6 breaking changes, all in a niche go/analysis integration API). The format adapters stay in root because:

1. They're stdlib-only — no dependency to isolate
2. Methods on types are idiomatic Go and convenient
3. Moving them would break the API for all users, not just go/analysis users

The `analysis/` module is thin (1 file) but justified because:
- `golang.org/x/tools` is the heaviest dependency in the project
- go/analysis users are a specific, identifiable audience
- It enforces the design principle from PROPOSAL.md: *"the SDK depends on nothing"*

---

## Implementation Plan (Approach B)

### Step 1: Create `analysis/` module

1. Create `analysis/` directory
2. Move `diagnostic.go` → `analysis/diagnostic.go`
3. Change package declaration from `finding` to `analysis`
4. Convert `Finding.AnalysisDiagnostic()` method to `analysis.ToDiagnostic(f Finding) analysis.Diagnostic`
5. Update all imports: `golang.org/x/tools/go/analysis`, `go/ast`, `go/token` stay; add `github.com/larsartmann/go-finding` for core types
6. Create `analysis/go.mod`:
   ```
   module github.com/larsartmann/go-finding/analysis

   go 1.26.0

   require (
       github.com/larsartmann/go-finding v0.0.0
       golang.org/x/tools v0.44.0
   )
   ```
7. Move `diagnostic_test.go` → `analysis/diagnostic_test.go` and update

### Step 2: Create `pipeline/` module

1. Add `pipeline/go.mod`:
   ```
   module github.com/larsartmann/go-finding/pipeline

   go 1.26.0

   require (
       github.com/larsartmann/go-finding v0.0.0
       golang.org/x/sync v0.20.0
   )
   ```
2. Update all pipeline imports to use the module path (already correct)

### Step 3: Create `cmd/go-finding/` module

1. Add `cmd/go-finding/go.mod`:
   ```
   module github.com/larsartmann/go-finding/cmd/go-finding

   go 1.26.0

   require (
       github.com/larsartmann/go-finding v0.0.0
       github.com/larsartmann/go-finding/pipeline v0.0.0
       gopkg.in/yaml.v3 v3.0.1
   )
   ```

### Step 4: Update root `go.mod`

1. Remove `golang.org/x/tools`, `golang.org/x/sync`, `gopkg.in/yaml.v3`
2. Root becomes:
   ```
   module github.com/larsartmann/go-finding

   go 1.26.0
   ```

### Step 5: Create `go.work`

```go
go 1.26.0

use (
    .
    ./analysis
    ./pipeline
    ./cmd/go-finding
)
```

### Step 6: Update examples

- `examples/govet/main.go` — if it uses `FromDiagnostic`, update import to `analysis` package
- `examples/detectorutil/util.go` — no change (doesn't use any finding types)
- Other examples — no change (they use root package types)

### Step 7: Update CI

- Test each module independently:
  ```yaml
  - cd / && go test github.com/larsartmann/go-finding/...  (root)
  - cd / && go test github.com/larsartmann/go-finding/analysis/...
  - cd / && go test github.com/larsartmann/go-finding/pipeline/...
  - cd / && go build github.com/larsartmann/go-finding/cmd/go-finding
  ```
- Or use `go.work` in CI: `GOWORK=off go test ./...` per module

### Step 8: Update documentation

- Update AGENTS.md, README.md, PROPOSAL.md with new module structure
- Add `analysis/README.md` explaining the go/analysis integration
- Update import examples in docs

### Step 9: Version tagging

- Root module: `git tag v1.1.0`
- Analysis module: `git tag analysis/v0.1.0`
- Pipeline module: `git tag pipeline/v0.1.0`
- CLI module: `git tag cmd/go-finding/v0.1.0`

---

## Open Questions

1. **Should `analysis/` also contain `astfix.go`?** — The pipeline has an `astfix.go` that uses `go/ast`. If pipeline needs go/analysis support, it should depend on the `analysis` module, not import `golang.org/x/tools` itself. Evaluate whether `astfix.go` belongs in `analysis/` or stays in `pipeline/` (pipeline would add `golang.org/x/tools` as a dependency only if AST fix is used).

2. **Should `examples/` be a module?** — No. Examples are demonstration programs, not importable packages. They can use `go.work` replace directives during development.

3. **go.work in version control?** — Yes, commit `go.work` for this repo. It's a monorepo workspace. The Go team's recommendation against committing `go.work` applies to libraries, not monorepos with multiple modules.

4. **Initial version for sub-modules?** — Start at `v0.1.0` for `analysis/` and `pipeline/` to signal they're new extractions. Root stays at `v1.x.x`.

5. **Replace directives during development?** — `go.work` handles this. No manual `replace` directives needed in `go.mod` files. The `v0.0.0` versions in `go.mod` are ignored when `go.work` is active.

---

## Dependency Size Comparison

### Before (single module)

```
go-finding  ──►  golang.org/x/tools  (heavy: ~50+ transitive deps)
            ──►  golang.org/x/sync    (light: 0 transitive deps)
            ──►  gopkg.in/yaml.v3     (light: 0 transitive deps)
```

Every user gets all three, even if they only need the data model.

### After (Approach B)

```
go-finding        ──►  (nothing)         ← data model users: ZERO deps
analysis/         ──►  go-finding
                  ──►  golang.org/x/tools ← go/analysis users: opt-in to heavy dep
pipeline/         ──►  go-finding
                  ──►  golang.org/x/sync  ← pipeline users: lightweight
cmd/go-finding/   ──►  go-finding
                  ──►  pipeline/
                  ──►  gopkg.in/yaml.v3   ← CLI only: not a library dep
```

**Result:** Most users (data model + SARIF/LSP/JSON) get zero transitive dependencies. Go/analysis users explicitly opt in. Pipeline users get a lightweight sync dependency. CLI is self-contained.

---

## What NOT to Split

These were considered and rejected:

| Candidate | Why not split |
|---|---|
| SARIF/LSP/JSON into `format/` | Stdlib-only — no dependency to isolate. Breaking method API not justified. |
| Filter/Merge into `query/` | Tightly coupled to `Report` type. No dependency boundary. |
| Errors/ID into separate module | Too granular. No dependency to isolate. Both stdlib-only. |
| Examples into modules | Not importable packages. Demo code only. |
| `detectorutil/` into module | Zero dependency on go-finding. Not part of the library API. |
