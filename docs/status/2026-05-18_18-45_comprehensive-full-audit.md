# Comprehensive Full Audit & Status Report

**Date:** 2026-05-18 18:45
**Reviewer:** Crush (full codebase review — every production file read)
**Version:** v0.2.1
**Commit:** 119d895 (master)

---

## Executive Summary

**go-finding is in excellent shape.** Build is clean, all tests pass with race detector, zero lint issues, and test coverage averages 97%+ across packages. The architecture is sound: clean separation between core types (root package), pipeline orchestration, CLI, and detectors.

That said, this review identified **5 P0 correctness issues**, **5 P1 design concerns**, and **5 P2 minor items** that should be addressed before a v1.0 release. None are blocking for continued v0.2.x development.

---

## Build, Test & Lint Matrix

| Check | Result |
|-------|--------|
| `go build ./...` | Clean — zero errors |
| `go vet ./...` | Clean — zero warnings |
| `golangci-lint run ./...` | **0 issues** |
| `go test -race -count=1 ./...` | **All pass** (9 packages) |
| `go test -race -count=1 -coverprofile` | See below |

### Coverage

| Package | Coverage |
|---------|----------|
| Core (`finding`) | **99.9%** |
| Analysis (`analysis`) | **100.0%** |
| Pipeline (`pipeline`) | **96.4%** |
| CLI (`cmd/go-finding`) | **95.4%** |
| Detectors (`internal/detectors`) | **96.1%** |
| Examples | Compile-only |

### Lines of Code

| Category | Lines |
|----------|-------|
| Production code | **6,675** |
| Test code | **17,211** |
| Test:Production ratio | **2.58:1** |

---

## a) FULLY DONE — Verified in Code

These items are **verified complete** by reading the actual source files during this review:

### Core Data Model
- [x] **Finding struct** — 14 fields, well-documented, Metadata-only extensibility (no Properties map[string]any)
- [x] **Builder API** — Fluent construction with `Build() (Finding, error)` and `MustBuild() Finding`
- [x] **Position / Range** — Full spatial algebra: `Contains`, `Overlaps`, `Intersection`, `Adjacent`, `Compare`
- [x] **Severity** — 4 levels with ordered comparison operators (`GreaterThan`, `LessThan`, etc.)
- [x] **FixStrategy** — 4 strategies (none/suggest/direct/ai) with `CanAutoApply()`, `NeedsAI()`
- [x] **Category** — 14 standard categories + custom accepted, `IsStandard()`/`IsValid()`
- [x] **Tag** — 10 standard tags + custom, `IsStandard()`/`IsValid()`
- [x] **Confidence** — Named type `float64` with `IsValid()`/`Clamp()`/`String()`, 5 named constants
- [x] **Suppression** — With `IsActive(now)` convenience, `IsExpired(now)`, 3 kinds
- [x] **RelatedRef** — Finding linkage with relation types
- [x] **ID generation** — `tool:rule:file:line:col` or hash-based, with `ParseID()`, `IsHashID()`
- [x] **Structured errors** — 5 categories, `errors.Is`/`errors.As` support, `WithFinding()`/`WithPosition()`

### Report Container
- [x] **Thread-safe AddFinding/AddFindings** — mutex-protected
- [x] **ComputeSummary** — Thread-safe, computes BySeverity/ByCategory/ByFixStrategy/FilesAffected/Suppressed
- [x] **Merge** — 3 deduplication strategies (ByID, ByPosition, ByRule)
- [x] **Correlate** — Cross-tool correlation, O(n²) capped at 10,000
- [x] **Filter/Map/All** — Composable predicates, iterator support
- [x] **Validate** — `Report.Validate()` and `ToolInfo.Validate()` added

### JSON Serialization
- [x] **Streaming** — `WriteJSON(w)`, `WriteSARIF(w)`, `WriteSARIFFiltered(w, sev)`
- [x] **io.WriterTo** — `Report.WriteTo(w)` with byte counting
- [x] **Invalid finding handling** — `ReportFromJSON`/`FindingsFromJSON` drop invalid, return count

### SARIF 2.1.0
- [x] **Export decomposed** — `sarifLocations()`, `sarifFixes()`, `sarifRelatedLocs()`, `sarifProperties()`
- [x] **Import decomposed** — `findingFromSarResult()`, `applySarifPosition()`, `applySarifProperties()`
- [x] **Round-trip fidelity** — go-finding/* property bag preserves all fields
- [x] **Named constants** — `sarifVersion`, `sarifSchema`, `sarifConfidenceScale`
- [x] **File-level exhaustruct exclusions** — SARIF and LSP types excluded in `.golangci.yml`

### LSP Conversion
- [x] **Finding→LSP** — `ToLSP()` with 0-based conversion, severity mapping, related info
- [x] **LSP→Finding** — `FromLSP(fileURI, diag)` preserves end position, raw severity in Metadata

### go/analysis Integration
- [x] **Extracted to analysis/ subpackage** — Root package has zero `golang.org/x/tools` dependency
- [x] **FromDiagnostic** — Auto-detects suggested fixes, sets FixStrategyDirect
- [x] **NodePosition/NodeRange** — AST node → Position/Range conversion

### Pipeline
- [x] **Byte-level FixEngine** — Descending offset application with frontier boundary
- [x] **FixProvider chain** — OffsetProvider → LineProvider → SubstringProvider
- [x] **Custom provider registration** — `NewFixEngineWithProviders`, `Config.FixProviders`
- [x] **FixEdit serializable** — MarshalJSON/UnmarshalJSON, ToSARIFProperties/FixEditFromSARIFProperties
- [x] **Conflict detection** — `DetectConflicts()`, `AnalyzeConflicts()`, `FilterConflictingFixes()`
- [x] **FixApplier lifecycle** — `Close()` cleans temp dirs, `defer Close()` in `applyDirectFixes`
- [x] **Verification** — `Verify()`, `DiffFindings()` — categorizes fixed/remaining/new
- [x] **Metrics** — Thread-safe, snapshot support, auto-populated in PipelineResult
- [x] **Retry** — Exponential backoff with jitter (`math/rand/v2`)
- [x] **Partial success** — `DetectPartial()`, context error separation, `FormatPartialErrors()`
- [x] **Parallel detection** — errgroup-based, `callbackMu` for OnFinding
- [x] **FindingProcessor** — Composable transforms (no context — see P1 item below)
- [x] **Context cancellation** — `IsContextError()` canonical check, all paths propagate
- [x] **Config validation** — `pipeline.New()` rejects invalid configs
- [x] **File backup/rollback** — Backup before modification, rollback on failure

### CLI
- [x] **Text/JSON/SARIF output** — 3 formats
- [x] **YAML/JSON config** — `go-faster/yaml` for config, validation
- [x] **Plugin detector registry** — Thread-safe `RegisterDetector()`
- [x] **CPU/memory profiling** — `setupProfiling()` with proper cleanup
- [x] **Graceful degradation** — Enabled by default in CLI

### Testing
- [x] **99.9% core coverage** — Unit, integration, fuzz, property-based, BDD, benchmark
- [x] **Fuzz tests** — JSON, ID, merge, SARIF (1M+ execs, zero panics)
- [x] **BDD tests** — 44+ ginkgo specs across root and pipeline packages
- [x] **Bug regression tests** — `*_bugfix_test.go` files
- [x] **Benchmarks** — Pipeline at 100/1k/10k findings

### CI/CD
- [x] **GitHub Actions** — CI with race, stress, govulncheck, coverage
- [x] **GoReleaser** — Full config with signing, SBOM, Homebrew, Nix, Scoop
- [x] **Release workflow** — `.github/workflows/release.yml`

### Documentation
- [x] **AGENTS.md** — Comprehensive project context for AI agents
- [x] **CONTEXT.md** — Domain language
- [x] **FEATURES.md** — Feature inventory with status indicators
- [x] **TODO_LIST.md** — Comprehensive with verification status
- [x] **CONTRIBUTING.md** — Including Nix setup path
- [x] **USAGE_GUIDE.md** — SARIF round-trip, pipeline behavior
- [x] **Migration guide** — v0.1 → v0.2
- [x] **JSON schemas** — `docs/schemas/finding.schema.json`, `report.schema.json`

---

## b) PARTIALLY DONE

### Open TODO Items with Code Started

| Item | Status | Detail |
|------|--------|--------|
| **API stability review** | Partial | Every exported symbol documented in FEATURES.md, but no formal v1.0 API lock audit performed |
| **Decide `NewFinding` API pattern** | Open | Current 6-param approach works, builder exists as alternative, but no formal decision recorded |
| **Decide domain-specific provider location** | Open | Interface designed, but Go AST / Rust syn providers not placed in module structure yet |
| **Evaluate go-sarif vs hand-rolled** | Deferred | Hand-rolled works well, deferred to post-v1 per TODO_LIST.md |
| **go/analysis reverse conversion** | Noted | Converting back to `analysis.Diagnostic` not supported, noted in README |

---

## c) NOT STARTED

### TODO_LIST.md Items with Zero Code

**P1 — Should Do Before v1.0:**
- [ ] `Properties map[string]any` alongside Metadata — Intentionally rejected per `finding.go:44-50` and dedicated doc commit `024b6a3`, but TODO item still open
- [ ] `Finding` struct sub-grouping (Identity/Location/Fix/Context) — Breaking, deferred to v2

**P2 — Nice to Have:**
- [ ] `go/analysis` reverse conversion (Finding → analysis.Diagnostic)
- [ ] Nix migration (full proposal exists, no code)
- [ ] Plugin architecture for detectors (runtime registration works, but hardcoded defaults)
- [ ] Pipeline middleware/interceptor pattern
- [ ] Watch mode with fsnotify
- [ ] Structured logging (slog)
- [ ] `finding.Diff()` function
- [ ] `finding.FormatText()` / `finding.FormatMarkdown()`
- [ ] Per-detector timeouts
- [ ] Semantic merge for conflicts
- [ ] Progress reporting to Pipeline
- [ ] Styled CLI output (lipgloss)
- [ ] Interactive TUI for fix review (bubbletea)
- [ ] More detector integrations (golangci-lint, errcheck)
- [ ] LSP language server
- [ ] Code fixes via LSP (CodeAction)

**P3 — Out of Scope:**
- Web UI, distributed detection, IDE plugins, OTel, AI backend, streaming analysis, WebSocket API, cloud integration, ML classification, trend analysis, webhooks, compliance reporting

---

## d) TOTALLY FUCKED UP — Issues Found in This Review

### P0 — Correctness Bugs

#### 1. `Report` Read Methods Are NOT Thread-Safe
**Files:** `report.go:159-240`
**Impact:** DATA RACE — `ActiveFindings()`, `BySeverity()`, `ByCategory()`, `ByFixStrategy()`, `FindByID()`, `FindByRule()`, `Len()`, `Filter()`, `Map()`, `All()` all read `r.Findings` without acquiring `r.mu`. If another goroutine calls `AddFinding()` concurrently, this is a textbook data race.

The godoc says "zero value safe for concurrent use" but only write methods acquire the lock. **Every read method races.**

#### 2. `IsAutoFixable()` / `Validate()` Disagree
**Files:** `finding.go:151` vs `finding.go:258-261`
**Impact:** A Finding can pass `IsAutoFixable()` but fail `Validate()`:
- `IsAutoFixable()` returns `true` when `FixStrategyDirect && (BeforeCode == "" || AfterCode != "")`
- `Validate()` rejects when `FixStrategyDirect && BeforeCode == "" && AfterCode != ""`

The pipeline uses `IsAutoFixable()` for triage, so a finding can be triaged as "direct fix" but fail validation if `Validate()` is called later.

#### 3. `FindingProcessor.Process` Missing Context + Error
**File:** `adapters.go:55-60`
**Impact:** The `FindingProcessor` interface takes no `context.Context` and returns no error. Processors that need I/O (enrichment, config loading, network calls) cannot signal failure or respect cancellation. The pipeline silently discards processor failures.

#### 4. Conflict Detection Overgrouping
**File:** `conflict.go:80-93`
**Impact:** When fix A overlaps fix B, and fix B overlaps fix C, all three are grouped together even if A doesn't overlap C. Only the first fix survives. This is overly conservative — fix C might be safely applicable.

#### 5. `DeduplicateByID` Uses Empty IDs as Same Key
**File:** `merge.go:140`
**Impact:** Two findings with empty IDs (possible via direct `Finding{}` construction) will be incorrectly deduplicated as the same finding during merge.

### P1 — Design Concerns

#### 6. `Confidence` Bypass via Direct Struct Construction
**File:** `finding.go:40`
**Impact:** `Finding{Confidence: 1.5}` compiles. `Validate()` catches it, but only if called.

#### 7. `Severity` / `FixStrategy` Are String Types — No Compile-Time Safety
**Impact:** `Finding{Severity: "typo"}` compiles. Runtime `IsValid()` catches it.

#### 8. `ComputeSummary()` Non-Deterministic with Expiring Suppressions
**File:** `report.go:149`
**Impact:** `f.IsSuppressed()` uses `time.Now()`, so `Suppressed` count varies with when `ComputeSummary()` is called.

#### 9. FixApplier Created Per Iteration
**File:** `pipeline.go:410-424`
**Impact:** Each `applyDirectFixes` call creates a new temp directory via `os.MkdirTemp`. Across MaxIterations, this creates unnecessary filesystem churn.

#### 10. `defaultMaxIterations` Duplicated
**Files:** `cmd/go-finding/main.go:19`, `pipeline/config.go:47`
**Impact:** Same constant in two packages. If one changes, the other won't.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Add mutex reads to Report** — `RLock`/`RUnlock` for all read methods. This is the single highest-impact correctness fix.

2. **Resolve `IsAutoFixable`/`Validate` disagreement** — Either allow `Direct + AfterCode only` in Validate, or make `IsAutoFixable` require `BeforeCode`. Document the canonical source.

3. **Add context + error to FindingProcessor** — `Process(ctx context.Context, findings []Finding) ([]Finding, error)`. Breaking but essential before v1.0.

4. **Make Conflict Detection byte-level aware** — Replace line-based grouping with the existing `FixEngine.ApplyWithConflicts` for conflict resolution. `FilterConflictingEdits` exists but isn't used in the pipeline.

### Type Safety

5. **Enforce Confidence range at construction** — Either unexport the field (use WithConfidence builder only) or add a `check()` in every method that reads it.

6. **Consider sealed interfaces or constructor-only types** for Severity/FixStrategy — Prevents `"typo"` from compiling.

### Testing

7. **Add concurrent read-write Report test** — Race detector would catch the P0 issue immediately.

8. **Add `IsAutoFixable`/`Validate` contradiction test** — Verify the fix resolves the disagreement.

### Documentation

9. **Close the `Properties map[string]any` TODO** — It's intentionally rejected with documented rationale. Mark it as `[WONTFIX]` or remove from TODO.

10. **Record API decision for `NewFinding`** — Current 6-param is fine, just needs a formal ADR entry.

---

## f) Top 25 Things to Get Done Next (Prioritized)

### P0 — Correctness (Do Before Any Feature Work)

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 1 | Add `RLock`/`RUnlock` to all Report read methods | 1h | **Critical** — data race |
| 2 | Fix `IsAutoFixable`/`Validate` disagreement | 30m | **High** — triage inconsistency |
| 3 | Add concurrent Report read-write race test | 30m | **High** — verify fix #1 |
| 4 | Fix `DeduplicateByID` empty-ID collision | 30m | **Medium** — silent data loss |
| 5 | Use `FilterConflictingEdits` in pipeline instead of line-based conflict detection | 2h | **Medium** — byte-level accuracy |

### P1 — Design (Do Before v1.0 API Lock)

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 6 | Add `context.Context` + `error` to `FindingProcessor.Process` | 1h | **High** — processor failures are silent |
| 7 | Add pipeline-level FixApplier reuse across iterations | 1h | **Medium** — unnecessary temp dirs |
| 8 | Enforce Confidence range — unexport or validate at every read site | 1h | **Medium** — silent out-of-range |
| 9 | Deduplicate `defaultMaxIterations` constant | 15m | **Low** — DRY |
| 10 | Make `ComputeSummary` deterministic for expiring suppressions | 30m | **Medium** — flaky counts |
| 11 | Close `Properties map[string]any` TODO as WONTFIX | 15m | **Low** — already rejected |
| 12 | Record `NewFinding` API decision in ADR | 15m | **Low** — documentation |
| 13 | Formal v1.0 API audit — every exported symbol reviewed | 3h | **High** — API stability |

### P2 — Quality (Nice to Have)

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 14 | Add `go/analysis` reverse conversion (Finding → Diagnostic) | 3h | **Medium** — round-trip completeness |
| 15 | Add structured logging (slog) to pipeline + CLI | 2h | **Medium** — observability |
| 16 | Add per-detector timeout configuration | 1h | **Medium** — fine-grained control |
| 17 | Surface SubstringProvider ambiguity with metrics/logging | 30m | **Low** — silent ambiguity |
| 18 | Add `finding.Diff()` utility function | 1h | **Low** — convenience |
| 19 | Add `finding.FormatText()` / `finding.FormatMarkdown()` | 2h | **Low** — output flexibility |
| 20 | Wire `FilterConflictingEdits` as pipeline option | 1h | **Low** — opt-in byte-level conflicts |

### P3 — Future (Post v1.0)

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 21 | Nix flake migration | 2d | **Medium** — reproducibility |
| 22 | Watch mode with fsnotify | 2d | **Medium** — continuous analysis |
| 23 | More detector integrations (golangci-lint, errcheck) | 1d each | **Medium** — ecosystem coverage |
| 24 | Plugin architecture (runtime detector loading) | 3d | **High** — extensibility |
| 25 | Interactive TUI for fix review (bubbletea) | 3d | **Medium** — UX |

---

## g) Top #1 Question I Cannot Resolve Myself

**Should `Report` read methods acquire `RLock`, or should `Report` transition to an immutable/append-only design?**

The current `Report` is a mutable container with `sync.Mutex` on writes but no protection on reads. There are two architectural paths:

**Option A: Add `RLock`/`RUnlock` to reads.** Simple, preserves current API. But `All()` returns `iter.Seq[Finding]` — the lock would need to be held for the entire iteration, which is a footgun if the caller doesn't exhaust the iterator.

**Option B: Make Report append-only, snapshot reads.** `AddFinding` appends, all reads return snapshots. Eliminates the lock entirely. But changes the API contract (`FindByID` returns a copy already, so this is close to current behavior).

This is a design decision that affects the public API and should be made before v1.0. I cannot decide this autonomously — it requires the project owner's judgment on the tradeoff between simplicity and API surface change.

---

## Dependency Audit

| Dependency | Version | Used By | Status |
|------------|---------|---------|--------|
| `golang.org/x/tools` | v0.45.0 | `analysis/` only | **OK** — isolated |
| `golang.org/x/sync` | v0.20.0 | `pipeline/` (errgroup) | **OK** |
| `github.com/go-faster/yaml` | v0.4.6 | CLI only | **OK** — not in core |
| `github.com/onsi/ginkgo/v2` | v2.29.0 | BDD tests | **OK** — test only |
| `github.com/onsi/gomega` | v1.41.0 | BDD test matchers | **OK** — test only |
| `github.com/stretchr/testify` | (indirect) | None — transitive only | **OK** — zero source imports |

---

## File-by-File Review Summary

### Root Package (Core Types)

| File | Lines | Quality | Notes |
|------|-------|---------|-------|
| `finding.go` | 353 | Excellent | Clean struct, good methods, Clone deep-copies correctly |
| `position.go` | 402 | Excellent | Full spatial algebra, well-decomposed helpers |
| `report.go` | 241 | **P0 issue** | Read methods lack mutex — data race |
| `filter.go` | 184 | Excellent | Composable predicates, in-place optimization |
| `merge.go` | 221 | Good | Correlate O(n²) capped, dedup strategies clean |
| `json.go` | 116 | Excellent | Streaming, invalid finding handling |
| `sarif_export.go` | 261 | Excellent | Well-decomposed, streaming support |
| `sarif_import.go` | 201 | Excellent | `stringProp` helper, bounds-checked |
| `sarif_types.go` | 156 | Excellent | Named constants, clear types |
| `lsp.go` | 192 | Good | Lossy conversion honestly documented |
| `errors.go` | 156 | Excellent | Categories, sentinels, WithFinding/WithPosition |
| `finding_builder.go` | 136 | Excellent | Fluent API, MustBuild panic for programmer errors |
| `severity.go` | 119 | Excellent | Ordered comparison, total ordering for invalid |
| `fix_strategy.go` | 44 | Excellent | Clean, AI reserved |
| `confidence.go` | 42 | Good | Clamped, but direct construction bypasses |
| `id.go` | 188 | Excellent | Hash-based fallback, Windows path handling |
| `suppression.go` | 56 | Excellent | IsActive convenience, 3 kinds |
| `category.go` | — | Excellent | 14 standard + custom |
| `tag.go` | — | Excellent | 10 standard + custom |
| `version.go` | 10 | Clean | v0.2.1 |

### Pipeline Package

| File | Lines | Quality | Notes |
|------|-------|---------|-------|
| `pipeline.go` | 439 | Good | Clean orchestration, single-use documented |
| `adapters.go` | 131 | **P1 issue** | FindingProcessor missing ctx+error |
| `config.go` | 85 | Excellent | Sentinel errors, joined validation |
| `result.go` | 50 | Clean | Findings()/SuggestedFindings() return copies |
| `conflict.go` | 241 | **P0 issue** | Overgrouping in detectConflictsInFile |
| `fix_engine.go` | 153 | Excellent | Descending offset, frontier boundary |
| `fix_edit.go` | 167 | Excellent | Overlaps, Validate, JSON, SARIF round-trip |
| `fix_provider.go` | 347 | Good | 3 providers, line offset index O(1) |
| `fix_applier.go` | 159 | Good | Backup/rollback, context cancellation |
| `verify.go` | 100 | Excellent | DiffFindings standalone utility |
| `metrics.go` | 164 | Excellent | Thread-safe, snapshot |
| `retry.go` | 128 | Excellent | math/rand/v2 jitter, context propagation |
| `partial.go` | 153 | Good | Context error separation correct |
| `file_backup.go` | 134 | Good | Hash-based backup naming |

### CLI

| File | Lines | Quality | Notes |
|------|-------|---------|-------|
| `main.go` | 195 | Good | Profiling, flag parsing |
| `config.go` | 229 | Good | YAML/JSON config, severity parsing |
| `registry.go` | — | Clean | Thread-safe detector registration |

### Analysis Package

| File | Lines | Quality | Notes |
|------|-------|---------|-------|
| `analysis.go` | 135 | Excellent | Clean go/analysis integration |

---

_This report was generated by reading every production source file in the codebase, running build/test/lint/vet, and cross-referencing with AGENTS.md, FEATURES.md, TODO_LIST.md, and CONTEXT.md._

_Assisted-by: Crush_
