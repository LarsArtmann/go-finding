# TODO List — go-finding

**Generated:** 2026-05-06
**Source:** Comprehensive audit of 56 .md files + full code review cross-referenced with actual code
**Project Version:** v0.2.1

---

## Files Processed

| #   | File                                                                          | Status  |
| --- | ----------------------------------------------------------------------------- | ------- |
| 1   | `AGENTS.md`                                                                   | ✅ Read |
| 2   | `CONTEXT.md`                                                                  | ✅ Read |
| 3   | `PROPOSAL.md`                                                                 | ✅ Read |
| 4   | `README.md`                                                                   | ✅ Read |
| 5   | `CHANGELOG.md`                                                                | ✅ Read |
| 6   | `CONTRIBUTING.md`                                                             | ✅ Read |
| 7   | `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`                                         | ✅ Read |
| 8   | `docs/USAGE_GUIDE.md`                                                         | ✅ Read |
| 9   | `docs/READINESS_REPORT.md`                                                    | ✅ Read |
| 10  | `docs/integration-guide.md`                                                   | ✅ Read |
| 11  | `docs/release-procedure.md`                                                   | ✅ Read |
| 12  | `docs/architecture-decisions.md`                                              | ✅ Read |
| 13  | `docs/planning/2026-04-30_00_05-COMPREHENSIVE_HARDENING_AND_API_STABILITY.md` | ✅ Read |
| 14  | `docs/planning/2026-04-30_00-50_execution-plan.md`                            | ✅ Read |
| 15  | `docs/planning/2026-04-30_00-48-coverage-and-cleanup.md`                      | ✅ Read |
| 16  | `docs/planning/2026-04-30_01-00_adr-go-lsp-integration.md`                    | ✅ Read |
| 17  | `docs/planning/2026-04-29_COMPREHENSIVE_DEEPENING_PLAN.md`                    | ✅ Read |
| 18  | `docs/planning/2026-04-29_23-01_deep-codebase-hardening.md`                   | ✅ Read |
| 19  | `docs/planning/2026-04-29_23-26-HARDENING_AND_INTEGRATION.md`                 | ✅ Read |
| 20  | `docs/status/2026-04-30_03-40_comprehensive-status.md`                        | ✅ Read |
| 21  | `docs/status/2026-04-30_03-33_final-session-status.md`                        | ✅ Read |
| 22  | `docs/status/2026-04-30_02-59_comprehensive-status.md`                        | ✅ Read |
| 23  | `docs/status/2026-04-30_02-42_comprehensive-status.md`                        | ✅ Read |
| 24  | `docs/status/2026-04-30_01-39_comprehensive-status.md`                        | ✅ Read |
| 25  | `docs/status/2026-04-30_01-39_session2-hardening-status.md`                   | ✅ Read |
| 26  | `docs/status/2026-04-30_00-55_comprehensive-status.md`                        | ✅ Read |
| 27  | `docs/status/2026-04-30_00-45_comprehensive-status.md`                        | ✅ Read |
| 28  | `docs/status/2026-04-30_00-39_comprehensive-status.md`                        | ✅ Read |
| 29  | `docs/status/2026-04-29_23-30_comprehensive-status.md`                        | ✅ Read |
| 30  | `docs/status/2026-04-29_EXECUTION_STATUS.md`                                  | ✅ Read |
| 31  | `docs/status/2026-04-28_17-14_COMPREHENSIVE_STATUS.md`                        | ✅ Read |
| 32  | `docs/status/2026-04-28_13-41_bug_fixes_and_builder_api.md`                   | ✅ Read |
| 33  | `docs/status/2026-04-26_19-06_clone-elimination-status.md`                    | ✅ Read |
| 34  | `docs/status/2026-04-26_17-39_clone-elimination-status.md`                    | ✅ Read |
| 35  | `docs/status/2026-04-26_14-28_comprehensive-status.md`                        | ✅ Read |
| 36  | `docs/status/2026-04-26_13-52_clone-elimination-status.md`                    | ✅ Read |
| 37  | `docs/status/2026-04-26_13-28_clone-elimination-status.md`                    | ✅ Read |
| 38  | `docs/status/2026-04-25_01-19_comprehensive-status.md`                        | ✅ Read |
| 39  | `docs/status/2026-04-21_20-45_session-21-status.md`                           | ✅ Read |
| 40  | `docs/status/2026-04-15_SESSION17_COMPREHENSIVE_STATUS.md`                    | ✅ Read |
| 41  | `docs/status/archive/2026-04-19_04-11_COMPREHENSIVE_FINAL_STATUS.md`          | ✅ Read |
| 42  | `docs/status/archive/2026-04-19_00-18_ROUND3_STATUS_AND_AUDIT.md`             | ✅ Read |
| 43  | `docs/status/archive/2026-04-19_07-04_SESSION12_EXECUTION_PROGRESS.md`        | ✅ Read |
| 44  | `docs/status/archive/2026-04-19_13-53_SESSION14_COMPREHENSIVE_STATUS.md`      | ✅ Read |
| 45  | `docs/status/archive/2026-04-19_13-43_SESSION13_COMPREHENSIVE_STATUS.md`      | ✅ Read |
| 46  | `docs/status/archive/2026-04-19_05-04_SESSION6_LINT_FIXES_IN_PROGRESS.md`     | ✅ Read |
| 47  | `docs/status/archive/2026-04-18_19-34_V5_COMPLETE_AND_LINT_HARDENED.md`       | ✅ Read |
| 48  | `docs/status/archive/2026-04-16_01-13_COMPREHENSIVE_STATUS.md`                | ✅ Read |
| 49  | `docs/status/archive/2026-04-15_18-25_COMPREHENSIVE_STATUS.md`                | ✅ Read |
| 50  | `docs/status/archive/2026-04-15_session-9-audit-tests.md`                     | ✅ Read |
| 51  | `docs/status/archive/2026-04-13_22-33_PIPELINE_COMPLETE.md`                   | ✅ Read |
| 52  | `docs/status/archive/2026-04-12_10-45_COMPREHENSIVE_STATUS.md`                | ✅ Read |
| 53  | `docs/status/archive/2026-04-28_13-41_COMPREHENSIVE_EXECUTION_PLAN.md`        | ✅ Read |
| 54  | `docs/status/archive/EXECUTION_PLAN_V2.md`                                    | ✅ Read |
| 55  | `docs/status/2026-04-26_13-28_clone-elimination-status.md`                    | ✅ Read |
| 56  | `TODO_LIST.md` (previous)                                                     | ✅ Read |

---

## 🔴 P0 — Must Do Before v0.2.0

### Version & Release

- [x] **Bump version to v0.2.0** — Done. Current version is v0.2.1. (`version.go`)
- [x] **Release `[Unreleased]` in CHANGELOG.md** — Done. v0.2.0 and v0.2.1 released.

### Architecture Decisions (blocking API lock)

- [ ] **Decide `NewFinding` API pattern** — Functional options vs builder-only vs current 6-param approach. Breaking changes have happened without a decision. (Reference: `docs/status/2026-04-30_02-59`)
- [x] **Extract `diagnostic.go` to `finding/analysis` subpackage** — Done. Deprecated wrappers removed from root. Root package no longer imports `golang.org/x/tools`. (`analysis/`)
- [ ] **API stability review** — Audit every exported symbol for v1.0.0 lock. Target: v0.2.0 = API-stable beta. (`docs/architecture-decisions.md` Decision #5)

### Correctness

- [x] **Fix `Pipeline.Run()` mutability** — Documented single-use contract in godoc. Run() resets internal state but is not safe for concurrent use. (`pipeline/pipeline.go:252-258`)
- [x] **Fix `FixApplier` cross-iteration persistence** — `defer func() { _ = applier.Close() }()` added in `applyDirectFixes`. Temp directories now cleaned up after each iteration. (`pipeline/pipeline.go:610`)
- [x] **Split `sarif.go` into 3 files** — Split into `sarif_types.go`, `sarif_export.go`, `sarif_import.go`. (`e97f62e`)
- [x] **Split `pipeline/pipeline.go`** — Extracted to `pipeline.go` + `adapters.go` + `config.go`. (`1bb1b1e`)
- [x] **Split `cmd/go-finding/main.go`** — Extracted to main.go + config.go + registry.go. (`cmd/go-finding/`)
- [x] **Document Pipeline single-use contract** — Added godoc on Run() explaining single-use and state-reset behavior. (`pipeline/pipeline.go:252-258`, `dc3ca37`)

### FixProvider Architecture (from byte-level redesign)

- [x] **Wire `FixEdit.Overlaps` into conflict detection** — `ConflictInfo.ConflictsWith` now populated by checking `edit.Overlaps(prev)` against all applied edits. (`pipeline/fix_engine.go:124-129`, `a3e4b58`)
- [x] **Make `FixEdit` serializable** — Added `MarshalJSON`/`UnmarshalJSON`, `ToSARIFProperties`/`FixEditFromSARIFProperties`. Edit properties preserved through SARIF round-trip. (`pipeline/fix_edit.go`, `e47c219`)
- [x] **Build line-offset index** — `buildLineOffsetIndex` returns `[]int` where `index[i]` is byte offset of line `i+1`. O(n) build, O(1) per lookup. (`pipeline/fix_provider.go:287`, `cba1b19`)
- [x] **Add BDD tests for FixProvider** — 15 BDD specs for OffsetProvider, LineProvider, SubstringProvider + chain precedence. (`pipeline/bdd_test.go`, `aa25f80`)
- [ ] **Decide domain-specific provider location** — Should Go AST, Rust syn, etc. providers live INSIDE `pipeline/fix/` or as SEPARATE modules? Affects module structure permanently. (`docs/architecture-decisions.md`)

### Architecture Deepening

- [x] **Centralize triage logic** — `HasFix()` is the canonical source; `IsAutoFixable()` added for pipeline auto-apply categorization. (`finding.go:135`, `pipeline/pipeline.go:346`)
- [x] **Convert stateless structs to functions** — Removed deprecated `ConflictDetector` and `Verifier` wrapper types. Package-level functions used directly. (`pipeline/conflict.go`, `pipeline/verify.go`)
- [x] **Make Report always thread-safe** — Changed `mu *sync.Mutex` to `mu sync.Mutex`. Zero-value `Report{}` is now safe for concurrent use. (`report.go:10`)

---

## 🟠 P1 — Should Do Before v1.0.0

### Code Quality

- [x] **Decompose `FindingsFromSARIF`** — Done. Split into `findingFromSarResult` (52 lines), `applySarifPosition` (32 lines), `applySarifProperties` (54 lines) with `stringProp` helper. Main function is 19 lines. (`sarif_import.go`) — **Still TODO:** decompose `findingToSARIF` export side (115 lines).
- [x] **Error wrapping consistency audit** — Verified: 100% `%w` wrapping across all 44 production files. Zero violations. Every `fmt.Errorf` wraps errors. `errors.New` used only for sentinels (correct pattern).
- [x] **Refactor CLI `run()` for testability** — `outputResults(w io.Writer, report, format)` already extracted at `config.go:164`. `writeOutput()` handles file creation. Summary/metrics printing in `run()` still uses `os.Stderr` directly but is acceptable for CLI entry point.

### Tag/Finding Cleanup

- [x] **Deprecate `WithTag` builder method** — Done. Added `// Deprecated: Use WithTags instead.` godoc marker. (`finding_builder.go`)
- [x] **Unify `Tag` deprecation** — Done. `WithTag` fully removed from builder. All tests use `WithTags` (plural). Zero `WithTag` usage in codebase.
- [x] **Add `Tag.IsStandard()` method** — Done. Exists at `tag.go:21`. Matches `Category.IsStandard()` pattern.

### Missing Features from Planning

- [ ] **Add `Properties map[string]any`** alongside `Metadata map[string]string` — Structured round-trip data for SARIF. Currently `Metadata` is string-only. (`finding.go`)
- [x] **Add `Suppression.IsActive()` method** — Done. Exists at `suppression.go:43`. Combines `IsValid() && !IsExpired(now)`.
- [x] **Add `Report.Merge(other *Report)` method** — Done in commit `df82906`. In-place merge for accumulating findings. (`report.go:80`)
- [ ] **Add `io.WriterTo` for SARIF** — Direct streaming without buffer allocation. `WriteSARIF` exists but is not `io.WriterTo`. (`sarif.go`)
- [x] **Confidence strong type** — Done. `type Confidence float64` with `IsValid()`/`Clamp()`/`String()` and 5 named constants. (`confidence.go`)

### Testing

- [x] **Add BDD tests for pipeline** — 44+ ginkgo BDD specs across root and pipeline packages including FixProvider contract. (`bdd_test.go`, `pipeline/bdd_test.go`)
- [x] **Add `WriteSARIF` error-path test** — Done. `TestWriteSARIF_WriterError` and `TestWriteSARIFFiltered_WriterError` use `failWriter` pattern. (`sarif_test.go:806-848`)
- [x] **Add `detectPartialSequential` context-cancel test** — Done. `TestDetectPartial_Sequential_CancelBeforeSecond` at `pipeline/partial_test.go:115`.
- [x] **Add `detectPartialParallel` context-cancel test** — Done. `TestDetectPartial_Parallel_CancelReturnsPartial` at `pipeline/partial_test.go:138`.

---

## 🟡 P2 — Nice to Have

### Performance & Tooling

- [ ] **Set up benchmark regression tracking** — Create `scripts/bench-compare.sh` or CI job. Baseline already captured in `bench_test.go`.
- [ ] **Performance benchmarks for 10k+ findings** — FixEngine benchmarks added (1/10/100/1000 fixes). Pipeline-level benchmarks still needed. (`pipeline/fix_engine_bench_test.go`, `c946025`)
- [ ] **Add `golines` to CI or justfile** — Enforce consistent line breaking automatically.
- [x] **Per-package coverage thresholds in CI** — Done. `scripts/coverage-check.sh` wired in CI via `coverage` job. Thresholds: root 98%, pipeline 95%, cmd 90%, detectors 90%, total 93%. **Note:** script has a bug — ignores its `coverage.out` argument (generates own data).

### SARIF & Output

- [ ] **SARIF schema validation test** — Verify output conforms to SARIF 2.1.0 JSON schema. Requires downloading schema. (`sarif_test.go`)
- [ ] **Evaluate `go-sarif` vs hand-rolled SARIF** — Assess migration cost for spec compliance. Deferred to post-v1.
- [x] **Document SARIF round-trip losses in user-facing docs** — Done. `docs/USAGE_GUIDE.md` now has SARIF Round-Trip Fidelity section with limitations and import docs.
- [ ] **`go/analysis` reverse conversion** — Converting back to `analysis.Diagnostic` is not yet supported. Noted in README.

### Documentation

- [x] **Add Nix setup path to `CONTRIBUTING.md`** — Done. Nix develop and direnv instructions added to Prerequisites section.
- [ ] **Document `FixStrategyAI` semantics in user-facing docs** — Currently only in code comments and architecture-decisions.md.
- [ ] **Add `Finding` JSON schema** — Formal JSON contract for API consumers.
- [x] **Create consumer migration guide** — Done. `docs/MIGRATION_v0.1-to-v0.2.md` covers all breaking changes and new features.

### Type Model

- [ ] **`Finding` struct sub-grouping** — Group fields into `Identity`, `Location`, `Fix`, `Context` embedded sub-structs. Breaking API change, deferred to v2.
- [x] **`Category.IsValid()` clarify semantics** — Godoc already documents the distinction clearly: `IsValid()` accepts custom categories, `IsStandard()` checks predefined constants. No change needed.
- [ ] **Protect `Confidence` in direct struct construction** — `Finding{Confidence: 1.5}` bypasses `NewFinding` clamping. Needs design decision.

---

## 🟢 P3 — Future / Deferred

### Nix Migration

- [ ] **Migrate from justfile to `flake.nix`** — Full proposal in `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md`. Phases 0–5.
  - [ ] Phase 0: Install Nix, create `flake.nix`
  - [ ] Phase 1: Verify `nix develop`, `nix build`
  - [ ] Phase 2: Replace CI workflows with Nix-based versions
  - [ ] Phase 3: direnv, `.envrc`, update docs
  - [ ] Phase 4: `nix flake check` integration
  - [ ] Phase 5: Pin nixpkgs, cross-platform testing

### Features

- [ ] **Plugin architecture for detectors** — Replace hardcoded `knownDetectorBuilders` with runtime registration. (`pipeline/pipeline.go`)
- [ ] **Pipeline middleware/interceptor pattern** — Allow custom stage injection between detect/triage/fix/verify.
- [ ] **Watch mode** with `fsnotify` for continuous analysis
- [ ] **Structured logging** (`slog`) — Replace `fmt.Fprintf` throughout `cmd` + `pipeline`
- [ ] **`finding.Diff()` function** — Compare finding sets (original vs fixed)
- [ ] **`finding.FormatText()` and `finding.FormatMarkdown()`** — Human-readable output formats
- [ ] **Detector timeout per-detector** — Configurable per-detector timeouts
- [x] **Column shift handling in `FixEngine`** — Resolved by byte-level redesign: edits now applied descending by offset with frontier boundary, eliminating column shift issues. (`pipeline/fix_engine.go`)
- [ ] **Semantic merge for conflicts** — In `pipeline` package
- [ ] **Progress reporting to Pipeline** — Callback for long-running operations
- [ ] **Styled CLI output** using `lipgloss`
- [ ] **Interactive TUI for fix review** (`bubbletea`)
- [ ] **`FuzzFindingsFromJSON` fuzzer** — JSON import is another attack surface beyond SARIF

### Integration

- [ ] **Decide on BuildFlow integration** — External project dependency, deferred
- [ ] **Decide on go-business-rules `Severity` sharing** — External project dependency, deferred
- [ ] **More detector integrations** — golangci-lint, errcheck, etc.

### LSP

- [ ] **Build a full language server** — If needed. Currently rejected per ADR (`docs/planning/2026-04-30_01-00_adr-go-lsp-integration.md`). Revisit if requirement changes.
- [ ] **Code fixes via LSP** — `CodeAction` support for direct fixes

### Out of Scope for v1

- [ ] Web UI prototype for pipeline monitoring
- [ ] Distributed detection
- [ ] IDE plugin stubs (VS Code)
- [ ] OpenTelemetry instrumentation
- [ ] AI backend for `FixStrategyAI`
- [ ] Streaming/incremental analysis
- [ ] WebSocket API for real-time finding streaming
- [ ] Cloud integration — send findings to external systems
- [ ] Finding classification ML
- [ ] Finding trend analysis over time
- [ ] Finding notification webhooks
- [ ] Compliance reporting

---

## ✅ Verified Completed (previously tracked, now confirmed via code audit)

These items were listed as TODOs across multiple planning/status docs but are **verified done** in the current codebase:

- [x] Fix `FixStrategyAI` split brain — `HasFix()` now treats AI like Suggest (requires `AfterCode`)
- [x] Add `Builder.Build()` error return — signature is `(Finding, error)`, not panic
- [x] Add `govulncheck` step to CI — `.github/workflows/ci.yml` has govulncheck job
- [ ] Add `go.work` for local development — Listed as done previously but file does NOT exist. Verify intent.
- [x] Add `FuzzFindingsFromSARIF` — `sarif_fuzz_test.go` has fuzz function (1.6M execs, zero panics)
- [x] Document SARIF round-trip losses in code — `sarif.go` godoc on `ToSARIF()`
- [x] Add benchmarks for hot paths — `bench_test.go` covers ID, Filter, Merge, SARIF
- [x] Profile memory allocation hotspots — Baseline captured
- [x] Add `version.go` with semver constants — now `v0.2.1`
- [x] Wire `Correlate()` into Pipeline — `CorrelateFindings` config field exists
- [x] Add `CONTRIBUTING.md` — Comprehensive guide exists
- [x] Add GitHub release workflow — `.github/workflows/release.yml` exists
- [x] Add GoReleaser config — `.goreleaser.yml` exists
- [x] Add `Report` goroutine safety — `sync.Mutex` on `AddFinding`/`AddFindings`
- [x] Add sync.Mutex for `OnFinding` callback — `callbackMu` in Pipeline
- [x] Convert retry errors to sentinels — `errMaxRetriesNegative` etc. at `retry.go:20-24`
- [x] Replace `math/rand` with `math/rand/v2` — Done in `retry.go`
- [x] Eliminate `detectResult` ghost type — Uses `PartialResult` directly
- [x] Extract `findingKey` to shared utility — `Finding.Key()` method
- [x] Add `FixEngine` unit tests — `fix_engine_test.go` exists
- [x] Add `FileBackup` direct tests — `file_backup_test.go` exists
- [x] Replace hardcoded `SeverityWarning` in `diagnostic.go` — Variadic `defaultSeverity` param
- [x] Add `Range.Contains` edge-case tests — Covered in `position_extra_test.go`
- [x] Add `DeduplicateByPosition` vs `DeduplicateByRule` behavior test — `TestDeduplicateStrategies_BehaviorDiff`
- [x] Add `Finding.Equal` field-mismatch test — `TestEqual_FieldMismatch` with 18 cases
- [x] Add `RetryConfig.Validate` edge-case tests — 7 comprehensive test functions
- [x] Add `FixApplier` error-path tests — 19 comprehensive test functions
- [x] Add `Verifier.Verify` error-path tests — `TestVerifier_Verify_DetectorError`
- [x] Add SARIF import tests — `TestFindingFromSarResult_WithFix`, `TestFindingFromSarResult_RankAsConfidence`
- [x] Add SARIF round-trip loss tests — `TestSARIF_RoundTripLosses`
- [x] Add `Builder.Build()` error-path test — `TestBuilder_Build_MissingFields`
- [x] Fix flaky `TestProperty_IDRoundTrip` — Seeded with `rand.New(rand.NewSource(42))`
- [x] Add `Correlation` JSON tags camelCase — Fixed
- [x] Remove `t.Parallel()` from CLI tests that mutate globals — All fixed with `//nolint:paralleltest`
- [x] Add `examples/` with compile checks — `example_compile_test.go` exists
- [x] Add `govet`/`builder`/`basic` to `.gitignore` — All present
- [x] Fix examples/builder compile error — Correctly uses `f, err := Build()`
- [x] `applyToFile` decomposed — 22-line function delegating to `FixEngine`
- [x] FixEngine unit tests — 262 lines in `fix_engine_test.go`
- [x] Add godoc examples for key APIs — `ExampleNewFinding`, `ExampleBuilder`, `ExampleFilter`
- [x] Add `go:generate stringer` — Not applicable: enums are `string` types, stringer only works with `int`
- [x] `Correlate` O(n²) limited to 10k — `maxCorrelations = 10000`
- [x] Delete stale binaries from repo root — No stale binaries found
- [x] CI stress test — `-count=20` job in `ci.yml`
- [x] `examples/example_compile_test.go` — Tests all 3 example dirs compile
- [x] `docs/integration-guide.md` — Real-world tool integration guide
- [x] `docs/release-procedure.md` — Release process documented
- [x] `docs/architecture-decisions.md` — 5 open decisions documented
- [x] `scripts/coverage-check.sh` — Per-package coverage thresholds
- [x] `Tag string` removed, replaced with `Tags []Tag` — Singular Tag field no longer exists on Finding struct
- [x] `NewFinding` accepts confidence parameter — 6-param signature with clamping
- [x] **Report.lock() mutex fix** — lock() was releasing mutex immediately via defer. Split into lock()/unlock(). `ComputeSummary()` now thread-safe. (`report.go`)
- [x] **WithTag Deprecated godoc marker** — Added `// Deprecated:` godoc marker for staticcheck detection. (`finding_builder.go`)
- [x] **FindingProcessor interface** — Added composable processor pattern (borrowed from golangci-lint) with `ProcessorFunc` adapter. (`pipeline/pipeline.go`)
- [x] **BDD test suite** — 29 ginkgo BDD specs across root and pipeline packages. (`bdd_test.go`, `pipeline/bdd_test.go`)
- [x] **Architecture diagrams** — Current + improved mermaid.js graphs in `docs/architecture-understanding/`

---

_This TODO list was generated by reading all 56 .md files in the repository, extracting every actionable item, cross-referencing with the actual codebase to verify completion status, and deduplicating across all sources._
