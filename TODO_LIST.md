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
- [ ] **Extract `diagnostic.go` to `finding/analysis` subpackage** — Removes 12MB `golang.org/x/tools` dep from core. Must happen before API stability lock. (`diagnostic.go`)
- [ ] **API stability review** — Audit every exported symbol for v1.0.0 lock. Target: v0.2.0 = API-stable beta. (`docs/architecture-decisions.md` Decision #5)

### Correctness

- [ ] **Fix `Pipeline.Run()` mutability** — `Run()` mutates `p.findings` and `p.iterations` internal state. Pipeline should be safe to reuse or document that each `Run()` needs a new Pipeline. (`pipeline/pipeline.go:253-256`)
- [ ] **Fix `FixApplier` cross-iteration persistence** — `applyDirectFixes` creates a new `FixApplier` each call, so backups from iteration N can't rollback iteration N+1. FixApplier.Close() is never called, leaking temp directories. (`pipeline/pipeline.go:597-610`, `pipeline/fix_applier.go:24`)
- [ ] **Split `sarif.go` into 3 files** — 570 lines, above 350 threshold. Split into `sarif_types.go`, `sarif_export.go`, `sarif_import.go`. (`sarif.go`)
- [ ] **Split `pipeline/pipeline.go`** — 611 lines, above 350 threshold. Extract Detector/Processor adapters to `pipeline/adapters.go`, Config+helpers to `pipeline/config.go`. (`pipeline/pipeline.go`)
- [ ] **Split `cmd/go-finding/main.go`** — 457 lines. Extract config parsing to `config.go`, output formatting to `output.go`. (`cmd/go-finding/main.go`)
- [ ] **Document Pipeline single-use contract** — Pipeline.Run() is not safe for concurrent or repeated use. Add godoc note. (`pipeline/pipeline.go:253`)

### FixProvider Architecture (from byte-level redesign)

- [ ] **Wire `FixEdit.Overlaps` into conflict detection** — `FixEdit.Overlaps()` exists but conflict detection still operates at `Range.Overlaps()` level. After providers resolve to `FixEdit`s, detect conflicts at the edit level for byte-precision. (`pipeline/conflict.go`, `pipeline/fix_edit.go:48`)
- [ ] **Make `FixEdit` serializable** — Add JSON tags and `ToSARIF`/`FromSARIF` conversion for SARIF round-tripping of edits. Currently `FixEdit` has no JSON representation. (`pipeline/fix_edit.go`)
- [ ] **Build line-offset index** — `lineColToOffset` is O(n) per call, re-scanning from the start for each fix. Build a `[]int` line-offset index once per file for O(1) lookup. (`pipeline/fix_provider.go:250`)
- [ ] **Add BDD tests for FixProvider** — No ginkgo BDD specs exist for the new FixProvider interface contract. (`pipeline/bdd_test.go`)
- [ ] **Decide domain-specific provider location** — Should Go AST, Rust syn, etc. providers live INSIDE `pipeline/fix/` or as SEPARATE modules? Affects module structure permanently. (`docs/architecture-decisions.md`)

### Architecture Deepening

- [ ] **Centralize triage logic** — Fix categorization appears in 3 places: `Finding.HasFix()`, `Pipeline.triage()`, and `FixEngine.Apply()` filtering. These overlap but aren't identical. Consolidate. (`finding.go:135`, `pipeline/pipeline.go:528`, `pipeline/fix_engine.go:54`)
- [ ] **Convert stateless structs to functions** — `ConflictDetector`, `Verifier` are zero-field structs. `FixEngine` now has state (providers). Convert ConflictDetector/Verifier to package-level functions for API honesty. (`pipeline/conflict.go:28`, `pipeline/verify.go:25`)
- [ ] **Make Report always thread-safe** — Zero-value `Report` has nil mutex (silently non-thread-safe). Use `sync.Once` for lazy init or document clearly. (`report.go:10-11`)

---

## 🟠 P1 — Should Do Before v1.0.0

### Code Quality

- [ ] **Decompose `FindingsFromSARIF`** — Cognitive complexity 90 (threshold 35). Already partially decomposed into helpers but main function may still be complex. Verify and further decompose if needed. (`sarif.go`)
- [ ] **Error wrapping consistency audit** — Ensure all internal errors use `%w` for unwrapping. `wrapcheck` linter catches gaps in CLI. (Various)
- [ ] **Refactor CLI `run()` for testability** — Uses global flag state (`flag.CommandLine`, `os.Args`, `os.Stderr`). Should accept `io.Writer` + `*flag.FlagSet` as parameters. (`cmd/go-finding/main.go`)

### Tag/Finding Cleanup

- [x] **Deprecate `WithTag` builder method** — Done. Added `// Deprecated: Use WithTags instead.` godoc marker. (`finding_builder.go`)
- [ ] **Unify `Tag` deprecation** — Either fully migrate tests to `Tags []Tag` or remove the deprecation. Current state is inconsistent. (Various test files)
- [x] **Add `Tag.IsStandard()` method** — Done. Exists at `tag.go:21`. Matches `Category.IsStandard()` pattern.

### Missing Features from Planning

- [ ] **Add `Properties map[string]any`** alongside `Metadata map[string]string` — Structured round-trip data for SARIF. Currently `Metadata` is string-only. (`finding.go`)
- [x] **Add `Suppression.IsActive()` method** — Done. Exists at `suppression.go:43`. Combines `IsValid() && !IsExpired(now)`.
- [ ] **Add `Report.Merge(other *Report)` method** — In-place merge for accumulating findings. Currently only a package-level `Merge()` function. (`report.go`)
- [ ] **Add `io.WriterTo` for SARIF** — Direct streaming without buffer allocation. `WriteSARIF` exists but is not `io.WriterTo`. (`sarif.go`)
- [ ] **Confidence strong type** — `type Confidence float64` with validation methods instead of bare `float64`. (`finding.go`)

### Testing

- [ ] **Add BDD tests for pipeline** — Done. 29 ginkgo BDD specs across root and pipeline packages. (`bdd_test.go`, `pipeline/bdd_test.go`)
- [ ] **Add `WriteSARIF` error-path test** — Use `failingWriter` pattern. Currently 75% coverage. (`sarif_test.go`)
- [ ] **Add `detectPartialSequential` context-cancel test** — 90% coverage, cancel path untested. (`pipeline/partial_test.go`)
- [ ] **Add `detectPartialParallel` context-cancel test** — 94.1% coverage, cancel path untested. (`pipeline/partial_test.go`)

---

## 🟡 P2 — Nice to Have

### Performance & Tooling

- [ ] **Set up benchmark regression tracking** — Create `scripts/bench-compare.sh` or CI job. Baseline already captured in `bench_test.go`.
- [ ] **Performance benchmarks for 10k+ findings** — Ensure pipeline scales to large codebases. (`bench_test.go`)
- [ ] **Add `golines` to CI or justfile** — Enforce consistent line breaking automatically.
- [ ] **Per-package coverage thresholds in CI** — Currently only total ≥75%. Individual package regressions (e.g., cmd 81%→70%) would go unnoticed. (Has `scripts/coverage-check.sh` but not wired to CI)

### SARIF & Output

- [ ] **SARIF schema validation test** — Verify output conforms to SARIF 2.1.0 JSON schema. Requires downloading schema. (`sarif_test.go`)
- [ ] **Evaluate `go-sarif` vs hand-rolled SARIF** — Assess migration cost for spec compliance. Deferred to post-v1.
- [ ] **Document SARIF round-trip losses in user-facing docs** — `sarif.go` has code comments but no user-facing documentation.
- [ ] **`go/analysis` reverse conversion** — Converting back to `analysis.Diagnostic` is not yet supported. Noted in README.

### Documentation

- [ ] **Add Nix setup path to `CONTRIBUTING.md`** — Currently missing despite migration proposal existing.
- [ ] **Document `FixStrategyAI` semantics in user-facing docs** — Currently only in code comments and architecture-decisions.md.
- [ ] **Add `Finding` JSON schema** — Formal JSON contract for API consumers.
- [ ] **Create consumer migration guide** — No migration guide for consumers upgrading from v0.1.3 to v0.2.0.

### Type Model

- [ ] **`Finding` struct sub-grouping** — Group fields into `Identity`, `Location`, `Fix`, `Context` embedded sub-structs. Breaking API change, deferred to v2.
- [ ] **`Category.IsValid()` strict validation** — Currently accepts any non-empty string. Should validate against known categories.
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
- [x] Add `go.work` for local development — `go.work` exists in repo root
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
