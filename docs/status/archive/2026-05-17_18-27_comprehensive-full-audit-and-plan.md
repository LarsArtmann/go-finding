# Comprehensive Project Status Report — go-finding

**Date:** 2026-05-17 18:27
**Session:** Full audit, plan creation, and execution readiness
**Reporter:** Crush (AI assistant)
**Previous Report:** 2026-05-08_05-18_post-self-review-hardening.md

---

## 1. Project Snapshot

| Metric                   | Value                                                                     |
| ------------------------ | ------------------------------------------------------------------------- |
| Version                  | v0.2.1                                                                    |
| Total commits            | 519                                                                       |
| Production files         | 44 `.go` files                                                            |
| Production LOC           | 6,610                                                                     |
| Test files               | 60                                                                        |
| Test LOC                 | 16,708 (2.5:1 test ratio)                                                 |
| Total functions          | 1,110                                                                     |
| Test entry points        | 620 (Test/Fuzz/Benchmark/It/Describe)                                     |
| Overall coverage         | 95.3%                                                                     |
| Root package coverage    | 99.9%                                                                     |
| Pipeline coverage        | 95.9%                                                                     |
| CLI coverage             | 95.4%                                                                     |
| Analysis coverage        | 100.0%                                                                    |
| Detectors coverage       | 96.1%                                                                     |
| Direct dependencies      | 5 (golang.org/x/tools, golang.org/x/sync, go-faster/yaml, ginkgo, gomega) |
| Indirect dependencies    | 14                                                                        |
| Banned deps (transitive) | `go.yaml.in/yaml/v3` via `go-faster/yaml` — acceptable                    |
| `//nolint:exhaustruct`   | 56 directives                                                             |
| All tests passing        | YES (race detector enabled)                                               |
| Go vet clean             | YES                                                                       |
| Error wrapping           | 100% `%w` (zero violations)                                               |
| Benchmarks               | 11 (root + pipeline FixEngine + parallel)                                 |
| Fuzz tests               | 4 files, 571 LOC (1.6M+ execs, zero panics)                               |
| BDD tests                | 2 files, 1,082 LOC                                                        |
| Bug-fix regression tests | 2 files, 384 LOC                                                          |
| CI jobs                  | 5 (test+coverage, lint, govulncheck, stress)                              |
| Release workflow         | GoReleaser + cosign + SBOM                                                |
| Lint config              | 253-line `.golangci.yml` (strict)                                         |

---

## 2. What's FULLY DONE (Verified in Code)

Every item below was verified against actual source code in this session.

### 2.1 Core Types & Data Model

| Item                                 | Evidence                                                                                               | Verified   |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------ | ---------- |
| `Finding` struct with 17 fields      | `finding.go:17-46` — Identity, Core, Classification, Fix, Context, Extensibility groups                | 2026-05-17 |
| `Confidence` named type              | `confidence.go` — `type Confidence float64` with `IsValid()`, `Clamp()`, `String()`, 5 named constants | 2026-05-17 |
| `KeySeparator` constant              | `finding.go:14` — `"\x00"` NUL character                                                               | 2026-05-17 |
| `Tags []Tag` (plural only)           | `finding.go:30` — singular `Tag` field fully removed, `WithTag` builder method fully removed           | 2026-05-17 |
| `Severity` 4-level enum              | `severity.go:118` — info/warning/error/critical with comparison operators                              | 2026-05-17 |
| `Category` 14 standard constants     | `category.go` — with `IsStandard()` + `IsValid()` (non-strict)                                         | 2026-05-17 |
| `FixStrategy` 4 strategies           | `fix_strategy.go` — none/suggest/direct/ai with `CanAutoApply()`                                       | 2026-05-17 |
| `Position` + `Range` spatial algebra | `position.go:401` — Contains, Overlaps, Intersection, Adjacent                                         | 2026-05-17 |
| `Suppression` with TTL/expiry        | `suppression.go` — `IsActive(now)` convenience method                                                  | 2026-05-17 |
| `Builder` API with `Build() error`   | `finding_builder.go:29` — returns `(Finding, error)`, not panic                                        | 2026-05-17 |
| `NewFixApplier` returns error        | `pipeline/fix_applier.go:22` — `(*FixApplier, error)`                                                  | 2026-05-17 |

### 2.2 Pipeline

| Item                                   | Evidence                                                                | Verified   |
| -------------------------------------- | ----------------------------------------------------------------------- | ---------- |
| Detect → Triage → Fix → Verify loop    | `pipeline/pipeline.go:71-199` — 128-line `Run()` with iteration control | 2026-05-17 |
| Parallel detection via errgroup        | `pipeline/pipeline.go` — configurable `ParallelDetectors`               | 2026-05-17 |
| Byte-level FixEngine                   | `pipeline/fix_engine.go:153` — descending offset, frontier boundary     | 2026-05-17 |
| 3 default FixProviders                 | `pipeline/fix_provider.go:346` — Offset, Line, Substring                | 2026-05-17 |
| Conflict detection                     | `pipeline/conflict.go:240` — `DetectConflicts()`, `AnalyzeConflicts()`  | 2026-05-17 |
| Exponential backoff retry              | `pipeline/retry.go` — configurable `RetryConfig` with jitter            | 2026-05-17 |
| Partial success (graceful degradation) | `pipeline/partial.go:152` — collects from failed detectors              | 2026-05-17 |
| Metrics collection                     | `pipeline/metrics.go:163` — thread-safe, snapshot support               | 2026-05-17 |
| Finding processors (composable)        | `pipeline/adapters.go` — `ProcessorFunc`, `NamedProcessorFunc`          | 2026-05-17 |
| Context cancellation propagation       | `pipeline/adapters.go` — `IsContextError()`, all paths propagate        | 2026-05-17 |
| Config validation                      | `pipeline/config.go:66` — `Validate()` returns joined errors            | 2026-05-17 |
| FixEdit serializable                   | `pipeline/fix_edit.go:166` — JSON marshal/unmarshal + SARIF round-trip  | 2026-05-17 |

### 2.3 SARIF & Interchange

| Item                             | Evidence                                                                    | Verified   |
| -------------------------------- | --------------------------------------------------------------------------- | ---------- |
| SARIF 2.1.0 export               | `sarif_export.go:230` — `ToSARIF()`, `WriteSARIF()` (streaming)             | 2026-05-17 |
| SARIF 2.1.0 import               | `sarif_import.go:200` — `FindingsFromSARIF()` with property bag restoration | 2026-05-17 |
| SARIF streaming via json.Encoder | `sarif_export.go:69-92` — no intermediate `[]byte`                          | 2026-05-17 |
| LSP Diagnostic conversion        | `lsp.go:194` — bidirectional Finding ↔ LSPDiagnostic                        | 2026-05-17 |
| go/analysis integration          | `analysis/analysis.go` — isolated subpackage, root package dep-free         | 2026-05-17 |
| JSON serialization (streaming)   | `json.go` — `WriteJSON(w)`, `WriteSARIF(w)` use `json.Encoder`              | 2026-05-17 |

### 2.4 Report & Filtering

| Item                               | Evidence                                                        | Verified   |
| ---------------------------------- | --------------------------------------------------------------- | ---------- |
| Report thread-safe (zero-value)    | `report.go:12` — `sync.Mutex` value, not pointer                | 2026-05-17 |
| Report merge + dedup               | `merge.go:220` — 3 strategies (ID, Position, Rule)              | 2026-05-17 |
| Cross-tool correlation (heuristic) | `merge.go` — capped at 10K correlations                         | 2026-05-17 |
| Composable filter predicates       | `filter.go:183` — 10 predicates + `Filter()`, `FilterInPlace()` | 2026-05-17 |
| Grouping + sorting                 | `filter.go` — `GroupBy*`, `SortBy*`                             | 2026-05-17 |
| Iterator support (`iter.Seq`)      | `report.go:202` — `All()` returns `iter.Seq[Finding]`           | 2026-05-17 |

### 2.5 Testing Infrastructure

| Item                       | Evidence                                                                                            | Verified   |
| -------------------------- | --------------------------------------------------------------------------------------------------- | ---------- |
| Fuzz tests (4 targets)     | `fuzz_test.go`, `id_fuzz_test.go`, `sarif_fuzz_test.go`, `merge_fuzz_test.go`                       | 2026-05-17 |
| BDD tests (ginkgo)         | `bdd_test.go`, `pipeline/bdd_test.go` — 1,082 LOC                                                   | 2026-05-17 |
| Bug-fix regression tests   | `pipeline/pipeline_bugfix_test.go`, `pipeline/fix_applier_bugfix_test.go` — 384 LOC                 | 2026-05-17 |
| Property-based tests       | `property_test.go` — randomized round-trip verification                                             | 2026-05-17 |
| Benchmarks (11 targets)    | `bench_test.go`, `pipeline/fix_engine_bench_test.go` — FixEngine 1/10/100/1000 + parallel detection | 2026-05-17 |
| CI stress test (-count=20) | `.github/workflows/ci.yml` — stress job                                                             | 2026-05-17 |

### 2.6 Error Handling

| Item                                  | Evidence                                                      | Verified   |
| ------------------------------------- | ------------------------------------------------------------- | ---------- |
| 100% `%w` error wrapping              | All `fmt.Errorf` calls across 44 production files verified    | 2026-05-17 |
| Structured error types (5 categories) | `errors.go:155` — Validation, IO, Parse, Conflict, Internal   | 2026-05-17 |
| `errors.Is` / `errors.As` support     | `FindingError` implements `Is()` for sentinel matching        | 2026-05-17 |
| Sentinel errors                       | Pipeline uses named sentinels (`errMaxRetriesNegative`, etc.) | 2026-05-17 |

### 2.7 Documentation

| Item                           | Evidence                                          |
| ------------------------------ | ------------------------------------------------- |
| README.md                      | Comprehensive project overview                    |
| AGENTS.md                      | Detailed architecture reference for AI assistants |
| FEATURES.md                    | Full feature inventory with status indicators     |
| TODO_LIST.md                   | Prioritized TODO list cross-referenced with code  |
| CHANGELOG.md                   | Version history                                   |
| CONTRIBUTING.md                | Contribution guide                                |
| docs/USAGE_GUIDE.md            | User-facing usage documentation                   |
| docs/READINESS_REPORT.md       | Release readiness assessment                      |
| docs/integration-guide.md      | Real-world tool integration guide                 |
| docs/release-procedure.md      | Release process documentation                     |
| docs/architecture-decisions.md | 5 open architecture decisions                     |
| docs/v1.0-release-criteria.md  | Release criteria checklist                        |
| docs/planning/                 | 7 planning documents                              |
| docs/status/                   | 13 current + 39 archived status reports           |

---

## 3. What's PARTIALLY DONE

### 3.1 SARIF Import/Export — Functional but Complex

| Aspect                    | Status  | Details                                                                                              |
| ------------------------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `findingToSARIF`          | Works   | 115 lines, deep nesting, 6 `//nolint:exhaustruct` — needs decomposition                              |
| `applySarifProperties`    | Works   | 10 repetitive if-check blocks — should be table-driven                                               |
| Two result builders       | Works   | `sarifResultsFromFindings` + `sarifResultsFromFindingsFiltered` differ by 1 condition — should merge |
| SARIF constants           | Partial | `"2.1.0"` and schema URL are magic strings in `buildSarifLog`                                        |
| Metadata lossy round-trip | Known   | `sarifMetadataFromProps` uses `fmt.Sprintf("%v", v)` — loses type info                               |

### 3.2 CLI — Functional but Not Testable

| Aspect            | Status | Details                                                                       |
| ----------------- | ------ | ----------------------------------------------------------------------------- |
| CLI output        | Works  | Uses `os.Stdout`/`os.Stderr` directly — cannot test output without subprocess |
| Config loading    | Works  | `loadConfig()` is testable; `run()` is not                                    |
| Detector registry | Works  | Thread-safe `RegisterDetector()`                                              |

### 3.3 TODO_LIST.md — Has Stale Entries

| Item in TODO_LIST                  | Actual Code State                                                        | Fix Needed                                        |
| ---------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------- |
| "Decompose `FindingsFromSARIF`"    | Already 19 lines (simple); `findingFromSarResult` is 52 lines (moderate) | Re-scope or close                                 |
| "Error wrapping consistency audit" | 100% `%w` wrapping verified                                              | Mark as done                                      |
| "Unify `Tag` deprecation"          | Fully removed, tests migrated                                            | Mark as done                                      |
| "Confidence strong type"           | `confidence.go` exists fully                                             | Mark as done (it's checked but text says "- [ ]") |

---

## 4. What's NOT STARTED

### 4.1 P0 — Blocking Decisions (Need User Input)

| #   | Item                                     | Why Blocked                                                     |
| --- | ---------------------------------------- | --------------------------------------------------------------- |
| 1   | Decide `NewFinding` API pattern          | Functional options vs builder-only vs 6-param — breaking change |
| 2   | API stability review for v1.0            | Audit every exported symbol — needs user approval scope         |
| 3   | Decide domain-specific provider location | Inside `pipeline/` or separate modules — module structure       |

### 4.2 P1 — Code Quality (Actionable Now)

| #   | Item                                            | Effort | Impact |
| --- | ----------------------------------------------- | ------ | ------ |
| 4   | Add `Report.Validate()` method                  | 10min  | Medium |
| 5   | Add `ToolInfo.Validate()` method                | 8min   | Medium |
| 6   | Merge duplicate SARIF result builders           | 10min  | Medium |
| 7   | Decompose `findingToSARIF` into helpers         | 10min  | High   |
| 8   | Refactor `applySarifProperties` to table-driven | 10min  | Medium |
| 9   | Extract SARIF constants from magic strings      | 5min   | Low    |
| 10  | `WriteSARIF` error-path test                    | 10min  | Medium |
| 11  | `detectPartialSequential` cancel test           | 10min  | Medium |
| 12  | `detectPartialParallel` cancel test             | 10min  | Medium |
| 13  | Extract `outputResults()` from CLI `run()`      | 10min  | High   |

### 4.3 P1 — Type Model Improvements (Breaking)

| #   | Item                                       | Effort | Impact | Breaking? |
| --- | ------------------------------------------ | ------ | ------ | --------- |
| 14  | Add `Properties map[string]any` to Finding | 15min  | High   | YES       |
| 15  | `Category.IsValid()` strict validation     | 8min   | Medium | Minor     |

### 4.4 P2 — Nice to Have

| #   | Item                                         | Effort |
| --- | -------------------------------------------- | ------ |
| 16  | `io.WriterTo` for Report (SARIF)             | 8min   |
| 17  | Reduce `//nolint:exhaustruct` (10-15)        | 10min  |
| 18  | Pipeline benchmarks for 10k+ findings        | 10min  |
| 19  | Benchmark regression script                  | 10min  |
| 20  | Add Nix setup to CONTRIBUTING.md             | 8min   |
| 21  | Document SARIF round-trip losses (user docs) | 8min   |
| 22  | Document `FixStrategyAI` semantics           | 5min   |
| 23  | Consumer migration guide v0.1→v0.2           | 10min  |
| 24  | Per-package coverage thresholds in CI        | 15min  |
| 25  | Add `golines` to CI                          | 15min  |

### 4.5 P3 — Future / Deferred

Nix migration (Phases 0-5), plugin architecture, pipeline middleware, watch mode, `slog` structured logging, `finding.Diff()`, `FormatText()`/`FormatMarkdown()`, per-detector timeout, `go/analysis` reverse conversion, `go-sarif` evaluation, Finding JSON schema, SARIF schema validation test, styled CLI output (`lipgloss`), interactive TUI (`bubbletea`), `FuzzFindingsFromJSON` fuzzer, semantic merge for conflicts, progress reporting, BuildFlow integration, go-business-rules Severity sharing, full language server, CodeAction via LSP, and all "Out of Scope for v1" items (web UI, distributed detection, IDE plugins, OTEL, AI backend, streaming analysis, WebSocket, cloud integration, ML classification, trend analysis, webhooks, compliance reporting).

---

## 5. What's TOTALLY FUCKED UP

### 5.1 Nothing Is Actually Broken

All tests pass. No compilation errors. No data corruption bugs. No security vulnerabilities.

### 5.2 But Here's What's Wrong Structurally

| Problem                                        | Severity   | Details                                                                                                                                                                                                                                                                     |
| ---------------------------------------------- | ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Metadata map[string]string` lossy             | **High**   | `sarifMetadataFromProps` converts all values via `fmt.Sprintf("%v", v)`. Structured data (numbers, booleans, nested objects) from non-go-finding SARIF becomes `"map[a:1 b:2]"`. This is the single biggest type model flaw.                                                |
| `Category.IsValid()` accepts garbage           | **Medium** | `Category("securty")` passes `IsValid()`. `IsStandard()` is strict but `IsValid()` is the name people will reach for first.                                                                                                                                                 |
| 56 `//nolint:exhaustruct` directives           | **Medium** | Every nolint is a place where a struct has ~20 fields but only 3 are set. The real fix: zero-value safe structs or builder-only construction. Many are in SARIF types where the spec has many optional fields — acceptable for interchange types, but not for domain types. |
| TODO_LIST.md has 4 stale entries               | **Low**    | Confidence, error wrapping, Tag deprecation, and FindingsFromSARIF complexity are either done or mis-scoped.                                                                                                                                                                |
| No `Report.Validate()` / `ToolInfo.Validate()` | **Low**    | Every other key type (`Finding`, `Config`, `FixEdit`, `RetryConfig`) has `Validate()`. Inconsistency.                                                                                                                                                                       |
| `findingToSARIF` 115 lines, deep nesting       | **Low**    | Works correctly but hard to read. Should be decomposed into `sarifFixFromFinding()`, `sarifRelatedFromFinding()`, `sarifPropertiesFromFinding()`.                                                                                                                           |

---

## 6. What We Should IMPROVE

### 6.1 Architecture (Type Model)

1. **Add `Properties map[string]any` alongside `Metadata map[string]string`** — This is the #1 architecture improvement. `Metadata` is string-only and lossy. `Properties` would hold structured data for SARIF round-trip and future tool integration. `Metadata` stays for simple key-value, `Properties` for structured data.

2. **Make `Category.IsValid()` strict by default** — Currently `IsValid()` checks `!= ""` and `IsStandard()` checks the allowlist. Invert: `IsValid()` should validate against the known set AND accept custom categories via a registry or prefix convention. Or rename: `IsNotEmpty()` + `IsValid()`.

3. **Protect `Confidence` in direct struct construction** — `Finding{Confidence: 1.5}` bypasses `NewFinding` clamping. Options: make `Confidence` unexported with getter, or accept it as a known limitation of Go struct literals.

### 6.2 Code Quality

4. **Decompose `findingToSARIF`** — Extract 3 helper functions. Reduces cognitive complexity from ~30 to ~8 per function.

5. **Merge duplicate SARIF result builders** — One function with `*Severity` threshold parameter.

6. **Table-driven `applySarifProperties`** — Replace 10 if-blocks with a slice of property definitions.

7. **Extract SARIF constants** — `sarifVersion` and `sarifSchema` named constants.

8. **CLI `run()` testability** — Extract `outputResults(report, format, outputFile, stdout, stderr error)` function.

### 6.3 Testing Gaps

9. **WriteSARIF error path** — `failingWriter` pattern to test encoding errors.

10. **Partial detection cancel paths** — Sequential and parallel cancel tests.

### 6.4 Documentation

11. **User-facing SARIF round-trip docs** — Currently only in code comments.

12. **Consumer migration guide v0.1→v0.2** — No guide exists.

13. **Nix setup in CONTRIBUTING.md** — Migration proposal exists but setup path not documented.

---

## 7. Top #25 Things to Do Next

Sorted by Impact × Urgency ÷ Effort. Each estimated ≤ 12 minutes unless noted.

| Rank | Task                                            | Impact   | Effort     | Type        | Blocking? |
| ---- | ----------------------------------------------- | -------- | ---------- | ----------- | --------- |
| 1    | **Decide `NewFinding` API pattern**             | Critical | Discussion | P0 Decision | YES       |
| 2    | **Decide `Properties map[string]any` addition** | Critical | Discussion | P0 Decision | YES       |
| 3    | **API stability review (exported symbols)**     | Critical | Hours      | P0 Decision | YES       |
| 4    | Fix TODO_LIST.md stale entries (4 items)        | Medium   | 5min       | Cleanup     | No        |
| 5    | Add `Report.Validate()` + tests                 | Medium   | 10min      | Feature     | No        |
| 6    | Add `ToolInfo.Validate()` + tests               | Medium   | 8min       | Feature     | No        |
| 7    | Merge duplicate SARIF result builders           | Medium   | 10min      | Refactor    | No        |
| 8    | Decompose `findingToSARIF` into 3 helpers       | High     | 10min      | Refactor    | No        |
| 9    | Refactor `applySarifProperties` table-driven    | Medium   | 10min      | Refactor    | No        |
| 10   | Extract SARIF constants (version, schema)       | Low      | 5min       | Cleanup     | No        |
| 11   | Add `WriteSARIF` error-path test                | Medium   | 10min      | Test        | No        |
| 12   | Add `detectPartialSequential` cancel test       | Medium   | 10min      | Test        | No        |
| 13   | Add `detectPartialParallel` cancel test         | Medium   | 10min      | Test        | No        |
| 14   | Extract CLI `outputResults()` for testability   | High     | 10min      | Refactor    | No        |
| 15   | Make `Category.IsValid()` strict                | Medium   | 8min       | Type Model  | Minor     |
| 16   | Reduce 10-15 `//nolint:exhaustruct`             | Medium   | 10min      | Cleanup     | No        |
| 17   | Add `io.WriterTo` for Report (SARIF)            | Low      | 8min       | Feature     | No        |
| 18   | Pipeline benchmarks for 10k+ findings           | Medium   | 10min      | Test        | No        |
| 19   | Benchmark regression script                     | Low      | 10min      | Tooling     | No        |
| 20   | Document SARIF round-trip in USAGE_GUIDE        | Medium   | 8min       | Docs        | No        |
| 21   | Document `FixStrategyAI` in USAGE_GUIDE         | Low      | 5min       | Docs        | No        |
| 22   | Add Nix setup to CONTRIBUTING.md                | Low      | 8min       | Docs        | No        |
| 23   | Consumer migration guide v0.1→v0.2              | Medium   | 10min      | Docs        | No        |
| 24   | Decide domain-specific provider location        | High     | Discussion | P0 Decision | YES       |
| 25   | Add `golines` to CI                             | Low      | 15min      | Tooling     | No        |

**Executable now:** Tasks 4-23, 25 (21 tasks, ~3 hours total)
**Blocked on user:** Tasks 1, 2, 3, 24 (4 decisions)

---

## 8. Top #1 Question I Cannot Answer Myself

> **Should `NewFinding(rule, toolName, message, severity, pos, confidence)` stay as-is, or should we deprecate it in favor of `NewBuilder(...).Build()` being the ONLY construction path?**
>
> **Context:**
>
> - `NewFinding` is a convenience function with 6 positional parameters — easy to get wrong
> - `NewBuilder(...).Build()` is self-documenting and validates everything
> - Both exist today and produce valid Findings
> - `Finding{...}` struct literal bypasses ALL validation (no required fields enforced)
> - Making `NewFinding` the only path = breaking change for all consumers
> - Making builder the only path = cleaner but adds `.Build()` call overhead
>
> **My recommendation:** Keep `NewFinding` for backward compatibility, add `// Deprecated: Use NewBuilder instead` godoc, and encourage builder as the canonical path. Don't break consumers. Lock API for v1.0.
>
> **Why I can't decide:** This is a product/API philosophy question. Only you (the project owner) can make this call because it affects every consumer of the library.

---

## 9. Session Timeline

| Time        | Activity                                                                         |
| ----------- | -------------------------------------------------------------------------------- |
| 18:27       | Session started — full project audit                                             |
| 18:27-18:35 | Read TODO_LIST.md, FEATURES.md, verified all P0/P1 items against code            |
| 18:35-18:45 | Ran all tests (pass), benchmarks, coverage (95.3%), vet (clean)                  |
| 18:45-18:55 | Deep code analysis: error wrapping, SARIF complexity, CLI structure              |
| 18:55-19:00 | Agent sub-tasks: verify exhaustruct count, Validate() gaps, duplicated functions |
| 19:00-19:10 | Created comprehensive execution plan (22 executable + 4 blocked tasks)           |
| 19:10-19:15 | Writing this status report                                                       |

---

## 10. Decision Matrix — What Happens Next

| Scenario                        | Action                                               |
| ------------------------------- | ---------------------------------------------------- |
| User says "execute plan"        | Execute tasks 4-23 in order, commit after each       |
| User answers P0 question #1     | Implement the chosen API pattern, update all callers |
| User says "just the quick wins" | Execute tasks 4-10 (~55 minutes)                     |
| User says "focus on type model" | Execute tasks 14-15, answer P0 question #2 first     |
| User says "focus on tests"      | Execute tasks 11-13, 17-18 (~40 minutes)             |

---

_Assisted-by: Crush <crush@charm.land>_
