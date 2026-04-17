# Module Split Plan — go-finding SDK

**Status:** Revised Proposal v2 | **Date:** 2026-04-16
**Identity:** SDK (library-first). CLI is a demo binary, not the product.

---

## Context

go-finding is an **SDK** providing a unified data model and pipeline for static analysis tools. Per the founding principle in PROPOSAL.md:

> _"Converters live in each tool, not in the SDK. Tools depend on the SDK; the SDK depends on nothing."_

Currently a single Go module (`github.com/larsartmann/go-finding`) with one sub-package (`pipeline/`), a demo CLI (`cmd/go-finding/`), and example programs (`examples/`). The root `go.mod` declares three external dependencies:

| Dependency                   | Used by                                       | Disk Size                | Transitive Deps                                    |
| ---------------------------- | --------------------------------------------- | ------------------------ | -------------------------------------------------- |
| `golang.org/x/tools` v0.44.0 | `diagnostic.go` only                          | **33 MB** (6 transitive) | go-cmp, goldmark, x/mod, x/net, x/telemetry, x/sys |
| `golang.org/x/sync` v0.20.0  | `pipeline/pipeline.go`, `pipeline/partial.go` | 104 KB (0 transitive)    | —                                                  |
| `gopkg.in/yaml.v3` v3.0.1    | `cmd/go-finding/main.go` only                 | 504 KB (0 transitive)    | —                                                  |

**Key insight:** The root package (everything except `pipeline/` and `cmd/`) is **stdlib-only**. The only file that needs `golang.org/x/tools` is `diagnostic.go` — a converter for `go/analysis.Diagnostic`. This creates a **33 MB unnecessary download** for SDK consumers who only want the data model.

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
  Integration ──►  │  diagnostic.go  (→ finding, id)       │  ← 33 MB golang.org/x/tools
                    └──────────────┬──────────────────────┘
                                   │ imports
                    ┌──────────────▼──────────────────────┐
                    │     Pipeline Package                  │
                    │  (pipeline/)                          │
                    │  pipeline.go, conflict.go, verify.go  │
                    │  retry.go, metrics.go, partial.go     │  ← 104 KB golang.org/x/sync
                    └──────────────────────────────────────┘
                                   │ imports
                    ┌──────────────▼──────────────────────┐
                    │     CLI (cmd/go-finding/)             │
                    │  main.go                              │  ← 504 KB gopkg.in/yaml.v3
                    └──────────────────────────────────────┘
```

### Method-on-Type Coupling (Critical Constraint)

Go requires methods on a type to be defined in the **same package** as the type. The following methods on core types exist in format/integration files:

| File            | Method                         | On Type   | Impact if moved          |
| --------------- | ------------------------------ | --------- | ------------------------ |
| `sarif.go`      | `Report.ToSARIF()`             | `*Report` | Must convert to function |
| `sarif.go`      | `Report.ToSARIFFiltered()`     | `*Report` | Must convert to function |
| `lsp.go`        | `Finding.ToLSP()`              | `Finding` | Must convert to function |
| `json.go`       | `Report.PrettyJSON()`          | `*Report` | Must convert to function |
| `json.go`       | `Finding.LineJSON()`           | `Finding` | Must convert to function |
| `diagnostic.go` | `Finding.AnalysisDiagnostic()` | `Finding` | Must convert to function |

### Test Cross-Module Dependency (Blocker — Resolved)

`example_test.go:254-288` (`ExamplePipeline`) imports `pipeline` from root package tests. If pipeline becomes a separate module, root's `go.mod` would need pipeline as a dependency — breaking the "zero dep" claim.

**Resolution:** Move `ExamplePipeline` (and any other pipeline-dependent examples) to `pipeline/example_test.go`. Root tests must not import pipeline. This is the correct home anyway — the example demonstrates pipeline usage, not core type usage.

---

## Recommendation: Approach B — SDK Module Split (4 modules)

Split by dependency boundaries. Each module only declares what it actually needs.

```
go-finding/
├── go.mod                           # Module 1: SDK Core
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
│   │   require: golang.org/x/tools  # Only consumer of this 33 MB dep
│   └── diagnostic.go
│
├── pipeline/
│   ├── go.mod                       # Module 3: Pipeline orchestration
│   │   require: go-finding          # (root)
│   │   require: golang.org/x/sync
│   └── *.go
│
├── cmd/go-finding/
│   ├── go.mod                       # Module 4: Demo CLI binary
│   │   require: go-finding          # (root)
│   │   require: go-finding/pipeline # (pipeline module)
│   │   require: gopkg.in/yaml.v3
│   └── main.go
│
└── go.work                          # Workspace for local development
```

**Dependency flow (clean DAG, no cycles):**

```
  cmd/go-finding ──► pipeline ──► core (go-finding)     ← ZERO deps
       │                                ▲
       └────────────────────────────────┘
       │
       └──► analysis ──► core (go-finding)               ← ZERO deps
```

### Why This Is the Right Split for an SDK

1. **PROPOSAL.md principle alignment:** _"Tools depend on the SDK; the SDK depends on nothing."_ Moving `diagnostic.go` (a converter for `go/analysis.Diagnostic`) out of root **enforces** this principle at the module level.

2. **33 MB dep isolation:** SDK consumers who use core types + SARIF/LSP/JSON output (the majority) get **zero transitive dependencies**. Only consumers who need `go/analysis` integration opt into the 33 MB `golang.org/x/tools` download.

3. **`analysis/` is not "too thin":** It's a focused integration adapter for `go/analysis.Diagnostic`, just like each downstream tool (art-dupl, branching-flow) will have its own adapter. The difference is this one ships with the SDK because `go/analysis` is ecosystem-standard.

4. **Format adapters stay in root:** SARIF, LSP, JSON are stdlib-only. Methods on types (`Report.ToSARIF()`, `Finding.ToLSP()`) are idiomatic Go. No dependency to isolate, no reason to break the API.

---

## API Changes Required

### Breaking Changes (6 — all in go/analysis integration)

| Current (root package)                        | New (analysis package)                         | Migration                         |
| --------------------------------------------- | ---------------------------------------------- | --------------------------------- |
| `finding.FromDiagnostic(d, fset, tool, rule)` | `analysis.FromDiagnostic(d, fset, tool, rule)` | Change import                     |
| `finding.FromTokenPosition(pos)`              | `analysis.FromTokenPosition(pos)`              | Change import                     |
| `finding.NodePosition(fset, node)`            | `analysis.NodePosition(fset, node)`            | Change import                     |
| `finding.NodeRange(node, fset)`               | `analysis.NodeRange(node, fset)`               | Change import                     |
| `finding.FormatDiagnostic(d, fset, name)`     | `analysis.FormatDiagnostic(d, fset, name)`     | Change import                     |
| `f.AnalysisDiagnostic()`                      | `analysis.ToDiagnostic(f)`                     | Change import + method → function |

### Migration Guide for Downstream Consumers

**Before:**

```go
import "github.com/larsartmann/go-finding"

func main() {
    f := finding.FromDiagnostic(diag, fset, "mytool", "RULE001")
    d := f.AnalysisDiagnostic()
    pos := finding.NodePosition(fset, node)
}
```

**After:**

```go
import (
    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-finding/analysis"
)

func main() {
    f := analysis.FromDiagnostic(diag, fset, "mytool", "RULE001")
    d := analysis.ToDiagnostic(f)
    pos := analysis.NodePosition(fset, node)
}
```

Everything else (Finding, Report, SARIF, LSP, JSON, Filter, Merge, Pipeline) — **no changes**.

### Who Is Affected

- Only users of the 6 go/analysis integration functions
- These are Go tool authors who already depend on `golang.org/x/tools`
- Migration is mechanical: add one import, change prefix on 1-6 calls

### Who Is NOT Affected

- Users of core types (Finding, Report, Severity, etc.)
- Users of format output (SARIF, LSP, JSON)
- Users of filtering, merging, grouping
- Users of the pipeline package
- Downstream tools (art-dupl, branching-flow, etc.) — they write their own converters

---

## Comparison Matrix

| Criterion                             | A: Minimal (2 modules) | **B: SDK Split (4 modules)**     | C: Full Modular (5+ modules) |
| ------------------------------------- | ---------------------- | -------------------------------- | ---------------------------- |
| External deps in root module          | 3 (unchanged)          | **0**                            | 0                            |
| `golang.org/x/tools` isolated (33 MB) | No                     | **Yes**                          | Yes                          |
| Breaking API changes                  | 0                      | **6** (go/analysis only)         | 14+                          |
| Method → function conversions         | 0                      | **1**                            | 5                            |
| Modules to maintain                   | 2                      | **4**                            | 5+                           |
| User migration effort                 | None                   | **Low** (go/analysis users only) | High (all users)             |
| Dependency download for core users    | ~34 MB                 | **0 MB**                         | 0 MB                         |
| Conceptual clarity                    | Low                    | **High**                         | Very high                    |
| Risk of over-splitting                | None                   | **Low**                          | High                         |
| Effort                                | ~1h                    | **~4-6h**                        | ~12-16h                      |
| SDK principle compliance              | Violates               | **Enforces**                     | Over-engineered              |

---

## Implementation Plan (Approach B)

### Step 1: Resolve the test blocker

Move `ExamplePipeline` from root `example_test.go` to `pipeline/example_test.go`.

**Before** (`example_test.go:254-288`):

```go
import "github.com/larsartmann/go-finding/pipeline"

func ExamplePipeline() {
    detector := pipeline.NamedDetectorFunc(...)
    p := pipeline.New(cfg, ".", detector)
    ...
}
```

**After** (`pipeline/example_test.go` — new file):

```go
package pipeline_test

import (
    "github.com/larsartmann/go-finding"
    "github.com/larsartmann/go-finding/pipeline"
)

func ExamplePipeline() { /* same code */ }
```

Verify root tests pass with `go test .` (no pipeline import).

### Step 2: Create `analysis/` module

1. Create `analysis/` directory
2. Move `diagnostic.go` → `analysis/diagnostic.go`
3. Change package from `finding` to `analysis`
4. Convert `Finding.AnalysisDiagnostic()` → `analysis.ToDiagnostic(f finding.Finding) analysis.Diagnostic`
5. Move `diagnostic_test.go` → `analysis/diagnostic_test.go`, update package + imports
6. Create `analysis/go.mod`:

   ```
   module github.com/larsartmann/go-finding/analysis

   go 1.26.0

   require (
       github.com/larsartmann/go-finding v0.0.0
       golang.org/x/tools v0.44.0
   )
   ```

### Step 3: Create `pipeline/` module

1. Add `pipeline/go.mod`:

   ```
   module github.com/larsartmann/go-finding/pipeline

   go 1.26.0

   require (
       github.com/larsartmann/go-finding v0.0.0
       golang.org/x/sync v0.20.0
   )
   ```

### Step 4: Create `cmd/go-finding/` module

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

### Step 5: Update root `go.mod`

Remove all three external dependencies. Root becomes:

```
module github.com/larsartmann/go-finding

go 1.26.0
```

### Step 6: Create `go.work`

```go
go 1.26.0

use (
    .
    ./analysis
    ./pipeline
    ./cmd/go-finding
)
```

### Step 7: Update examples

- `examples/govet/main.go` — uses `finding` types only, no change needed
- `examples/detectorutil/util.go` — no go-finding dependency, no change
- `examples/artdupl/main.go` — uses `finding` types + `detectorutil`, no change
- `examples/branching/main.go` — uses `finding` types only, no change
- `examples/staticcheck/main.go` — uses `finding` types + `detectorutil`, no change

None of the examples use `FromDiagnostic` or other `analysis/` functions — they construct `Finding` structs directly. No example changes needed.

### Step 8: Update CI

Fix the Go version matrix (currently broken — tests Go 1.21/1.22/1.23 but go.mod requires 1.26.0):

```yaml
strategy:
  matrix:
    go-version: ["1.26"]

steps:
  - name: Test root
    run: cd . && go test -v -race -coverprofile=coverage.out ./...

  - name: Test analysis
    run: cd analysis && go test -v -race ./...

  - name: Test pipeline
    run: cd pipeline && go test -v -race ./...

  - name: Build CLI
    run: cd cmd/go-finding && go build -v ./...
```

Note: Each `go test` runs within its module directory. `go.work` makes cross-module resolution work. CI should also run `GOWORK=off` tests to verify published dependency resolution works.

### Step 9: Update release workflow

Fix Go version + update for multi-module tags:

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: "1.26"
```

Update `.goreleaser.yml` to build from the CLI module:

```yaml
builds:
  - id: go-finding
    main: ./cmd/go-finding
    dir: ./cmd/go-finding # Build from CLI module root
```

Release tags use subdirectory prefix format:

- Root: `git tag v1.1.0`
- Analysis: `git tag analysis/v0.1.0`
- Pipeline: `git tag pipeline/v0.1.0`
- CLI: `git tag cmd/go-finding/v0.1.0`

### Step 10: First release procedure (bootstrap)

The `v0.0.0` versions in `go.mod` only work with `go.work` active. For the first published release:

```bash
# 1. Ensure all modules pass tests locally with go.work
go test ./...

# 2. Verify each module independently (simulate published resolution)
GOWORK=off bash -c 'cd analysis && go get github.com/larsartmann/go-finding@latest'
GOWORK=off bash -c 'cd pipeline && go get github.com/larsartmann/go-finding@latest'

# 3. Tag and push root first (downstream modules need a published version)
git tag v1.1.0
git push origin v1.1.0

# 4. Update analysis/go.mod to reference the real published version
cd analysis && go get github.com/larsartmann/go-finding@v1.1.0

# 5. Tag and push analysis
cd .. && git tag analysis/v0.1.0 && git push origin analysis/v0.1.0

# 6. Repeat for pipeline and CLI
```

### Step 11: Update documentation

- Update AGENTS.md with new module structure
- Add `analysis/doc.go` with package documentation
- Update PROPOSAL.md "File Structure" section
- Update README.md import examples

### Step 12: Rollback plan

If the split causes issues:

1. Revert the commit(s)
2. All code returns to single-module state
3. `go.work` is deleted
4. `go mod tidy` restores single `go.mod`
5. No data loss — files move back to original locations

---

## Dependency Size Comparison

### Before (single module — current state)

```
go-finding  ──►  golang.org/x/tools v0.44.0  (12 MB + 21 MB transitive = 33 MB total)
            ──►  golang.org/x/sync v0.20.0    (104 KB, 0 transitive)
            ──►  gopkg.in/yaml.v3 v3.0.1      (504 KB, 0 transitive)

Total download for any consumer: ~34 MB, 3 direct + 6 transitive = 9 modules
```

### After (Approach B — SDK split)

| Consumer type                          | What they import          | Download   | Transitive deps |
| -------------------------------------- | ------------------------- | ---------- | --------------- |
| **Core only** (types + SARIF/LSP/JSON) | `go-finding`              | **0 MB**   | **0**           |
| **Pipeline user**                      | `go-finding` + `pipeline` | **104 KB** | **0**           |
| **go/analysis user**                   | `go-finding` + `analysis` | **33 MB**  | **6**           |
| **CLI user**                           | everything                | **34 MB**  | **6**           |

**Result:** The majority of SDK consumers (downstream tools writing `ToFindings()` converters) get **zero dependencies**. They import the SDK, build `Finding` structs, output SARIF — and download nothing.

---

## What NOT to Split

| Candidate                      | Why not split                                                                                                                                              |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| SARIF/LSP/JSON into `format/`  | Stdlib-only — no dependency to isolate. Breaking method API (`report.ToSARIF()` → function) is a massive ergonomic regression with zero dependency payoff. |
| Filter/Merge into `query/`     | Tightly coupled to `Report` type. No dependency boundary.                                                                                                  |
| Errors/ID into separate module | Too granular. No dependency to isolate. Both stdlib-only.                                                                                                  |
| Examples into modules          | Not importable packages. Demo code only. Use `go.work` for local dev.                                                                                      |
| `detectorutil/` into module    | Zero dependency on go-finding. Not part of the SDK API.                                                                                                    |

---

## Open Questions (Resolved)

1. **Should `analysis/` also contain `astfix.go`?**
   The pipeline has `astfix.go` that uses `go/ast`. Keep it in pipeline for now — `go/ast` is stdlib, not `golang.org/x/tools`. If pipeline later needs `golang.org/x/tools/go/analysis`, it adds `analysis` as a dependency.

2. **Should `examples/` be a module?**
   No. Examples are demonstration programs, not importable packages. They use `go.work` for local development.

3. **go.work in version control?**
   Yes, commit `go.work`. This is a monorepo workspace. The Go team's recommendation against committing `go.work` applies to published libraries consumed externally, not monorepos with multiple modules.

4. **Initial version for sub-modules?**
   Start at `v0.1.0` for `analysis/` and `pipeline/` to signal they're new extractions. Root stays at `v1.x.x`.

5. **Replace directives during development?**
   `go.work` handles this. No manual `replace` directives in `go.mod`. The `v0.0.0` versions are ignored when `go.work` is active.
