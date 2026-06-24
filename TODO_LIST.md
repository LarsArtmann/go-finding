# TODO List

**Generated:** 2026-05-20
**Updated:** 2026-06-23 (session 24: all v1.0 API cleanup items completed — branded types, renames, validator decomposition)
**Files Processed:** 235

## 🔴 HIGH Priority

- [x] Add bounds checks in `findingFromSARIF` — bounds checks exist via `applySarifPosition`
- [x] Document SARIF critical round-trip loss — documented in doc.go, USAGE_GUIDE.md, FEATURES.md
- [ ] `Finding` struct sub-grouping — **DEFERRED v2** (breaking change)
- [x] **Fix `nix build .#` vendorHash mismatch** — `flake.nix:29` now declares correct hash `sha256-wISSq/...`. Commit `9d71627`.
- [x] **Decompose `Finding.Validate()` to satisfy gocyclo** — Split into 6 validators: `validateIdentity`, `validateClassification`, `validateFix`, `validateReferences`, `validateSpatial`, `validateSuppression`. Commit `ded53c2`.
- [x] Add `golines` to CI — treefmt-nix now supports golines; configured in `flake.nix` with maxLength=120
- [x] Clean up gopls hints (~12 non-critical) — production code: 0 rangeint, 0 newexpr, 2 mapsloop fixed
- [x] Implement art-dupl integration gaps (GAP-1, GAP-4, GAP-5, GAP-6, GAP-8) — fully implemented, tested, lint clean
- [x] Expand `doc.go` to comprehensive package documentation
- [x] Update `USAGE_GUIDE.md` for v0.4.x
- [x] Create `.github/workflows/ci.yml` — test, lint, race jobs on PR/push — `.github/workflows/ci.yml` exists with 5 jobs
- [x] Push 4 unpushed commits to origin — all commits pushed, v0.5.0 tagged
- [x] Fix `FixApplier` lifecycle — lifted to Pipeline constructor; `NewFixApplier` returns error
- [x] Restore `cmd/go-finding` test coverage from 70.0% toward 95% — now 90.7%
- [x] Update `README.md` with badges, pipeline diagram, API overview
- [x] Fix pre-commit hook failures — `goconst`, `todo-check`, `library-policy` → All passing. Fixed `err :=` redeclaration bugs in sarif_export.go/sarif_import.go (introduced by previous commit), removed duplicate `pipeline/pipeline_new_test.go`, fixed `FindingsSnapshot()` nil-for-empty behavior.

## 🟡 MEDIUM Priority — v1.0 API Cleanup (from data-model-review + naming-review, 2026-06-23)

These are breaking changes that should land before the v1.0 API lock. Each is small in isolation but becomes permanent post-1.0.

- [x] Add branded primitive types — `type ID string`, `type RuleName string`, `type ToolName string`, `type FilePath string` in `branded_types.go`. Applied to `Finding`, `RelatedRef`, `Correlation`, `ParsedID`. Type named `ID` (not `FindingID`) to avoid revive stutter. Commits `bce09ef`, `fbf0161`.
- [x] Wrap `SeverityAliases` in `sync.RWMutex` — `severity.go` now has `severityAliasesMu` + `RegisterSeverityAlias()` + `LookupSeverityAlias()`. Old `SeverityAliases()` deprecated as snapshot. Commit `fc3080b`.
- [x] Rename `FindingProcessor` → `FindingTransformer` and `Process()` → `Transform()` — `pipeline/adapters.go`. Commit `2811f1a`.
- [x] Rename `(*GeneratedFileFilter).Process()` → `Transform()` — `pipeline/generated_filter.go`. Implements `FindingTransformer`. Commit `2811f1a`.
- [x] Rename `GetCategory(err)` → `CategoryOf(err)` — `errors.go`. Old `GetCategory()` deprecated as wrapper. Commit `7577bd2`.
- [x] Rename `ConflictInfo` → `Conflict` — `pipeline/conflict.go`. `AnalyzeConflicts() []Conflict`. Commit `cd888a3`.
- [x] Rename `LSPRelatedInfo` → `LSPRelated` — `lsp.go`. Commit `cd888a3`.
- [x] Apply branded types to `RelatedRef.FindingID` and `Correlation.FindingIDs` — `finding.go:83`, `correlate.go:38`. Commit `bce09ef`.

## 🟡 MEDIUM Priority

- [x] Fix `pipeline/partial.go:107` — uses `IsContextError()`
- [x] Fix `pipeline/fix_applier.go` — `fmt` import actively used
- [x] Fix `cmd/go-finding/integration_test.go:108` — no duplicate, separate functions
- [x] Fix `examples/builder/main.go` compile error — correctly handles both values
- [x] Fix `.golangci.yml` to work without `--no-verify` / `--no-config` flag — removed 6 invalid linter names
- [x] Fix `.golangci.yml` indentation — normalized to 2-space by BuildFlow auto-configure
- [ ] Fix BuildFlow auto-configure loop — **BLOCKED** (external tool)
- [x] Fix `IsAutoFixable()`/`Validate()` agreement for Direct+AfterCode-only findings
- [x] Fix `Finding.Key()` cross-tool collision — `GenerateID` now length-prefixed hash
- [x] Fix `Pipeline.Run()` single-use contract — `ran` bool guard added
- [x] Fix OnFix callback inaccuracy — reports false for skipped fixes
- [x] Fix partial detection missing metrics
- [x] Fix `Metrics.StageTiming` value receiver bug — uses `MetricsSnapshot` value
- [x] Fix `FixEngine.Apply()` to use `HasCodeChange()`
- [x] Fix conflict detection overgrouping — documented as intentional design
- [x] Fix `DeduplicateByID` — skips findings with empty ID
- [x] Fix `%v` → proper error wrapping in `FormatPartialErrors`
- [x] Fix `Range.LineCount()` for inverted ranges
- [x] Fix `TestProperty_IDRoundTrip` flakiness
- [x] Refactor CLI `run()` for testability
- [x] Fix `NewFixApplier` error handling — propagates `MkdirTemp` error
- [x] Fix 10 testifylint `require-error` warnings — migrated to gomega
- [x] Fix 6 `paralleltest` warnings — all test functions have `t.Parallel()`
- [x] Update `FEATURES.md` for v0.4.x
- [x] Fix `.gitignore` line 43 corruption
- [x] FixEngine: line-offset tracking for cumulative line shifts across multi-fix — `LineShiftMap` in `pipeline/line_shift.go` with 7 tests
- [~] ~~Make fix strategy composable as interface~~ — REMOVED: ghost system with zero consumers. `FixStrategy.CanAutoApply()` is the single source of truth.
- [x] Add pipeline stage hooks — pre/post hooks for detect, triage, fix, verify — `StageHook` interface in `pipeline/stage_hook.go` with 2 tests
- [x] Add `DeduplicateBy.String()` method — consistent with all other named types
- [x] Mark 225 lint warnings phantom — `go-structure-linter` is external
- [x] Mark `GitAuthorProvider` phantom — `rule_service.go` does not exist
- [x] Mark `.git/hooks/pre-commit` executable — already executable
- [ ] Interactive TUI — **OUT OF SCOPE v1** (see [ROADMAP.md](ROADMAP.md#tooling-integrations))
- [ ] Phase 3: Create `.envrc` — **WONTFIX** (project uses Nix flakes, not direnv)

## 🟢 LOW Priority

- [x] API stability review — audit all exported symbols for v1.0.0 lock (docs/API_STABILITY.md)
- [x] Decide `FixStrategyAI` fate — keep as RESERVED placeholder
- [x] **Rename `fs` → `strategy` in `finding_validate.go`** — Done. Commit `d7dc292`.
- [x] **Rename `rt` → `result` in `splitbrain_test.go`** — Done. Commit `d7dc292`.
- [x] Document `BySeverityAtLeast` excludes invalid severities
- [x] Document `Report.All()` yields copies — with shallow copy caveat
- [x] Document `FindByID` returns copy — with shallow copy caveat
- [x] Make `Key()` NUL separator `"\x00"` a named constant
- [x] Document SARIF round-trip losses in user-facing docs
- [x] Document FixStrategyAI placeholder semantics
- [x] Document Correlate complexity — O(n·k) normal, O(k²) worst case, 10K cap
- [x] `Position` zero-value safety — **RESOLVED v0.9.0** (Offset uses -1 sentinel; constructors set it)
- [x] `Range.End` zero-value ambiguity — **RESOLVED v0.9.0** (cascaded from Position sentinel fix)
- [x] Add `finding.FormatText()`
- [x] Add `finding.FormatMarkdown()`
- [x] Write API stability guarantee document — `docs/API_STABILITY.md`
- [x] Decide stable ID format
- [x] Add `Finding` JSON schema — `docs/schemas/finding.json`
- [x] Protect `Confidence` in direct struct construction
- [x] Clamp `Confidence` in `Builder.WithConfidence`
- [x] Remove deprecated `Finding.Tag string` field — field does not exist
- [x] Remove deprecated `ConflictDetector`/`Verifier` structs
- [x] Remove phantom `FixStrategyAI` constant — decided KEEP as RESERVED
- [x] `Correlation` JSON tags camelCase
- [x] Make Metrics exported map fields unexported with accessor methods
- [x] Add `Config.Validate()` method with cross-field validation
- [x] Add `RetryConfig.Validate()` method
- [x] Wire `FilterConflictingEdits` as opt-in pipeline Config field + test
- [x] Handle empty-ID findings in `FilterConflictingEdits`
- [x] Populate `ConflictInfo.ConflictsWith` in edit-level conflict detection
- [x] Consistent structured errors in pipeline
- [x] Add per-detector timeout to `Config` and `RetryDetector`
- [x] Customizable `TriageFunc` in Config
- [x] Replace hardcoded detector builders with registry lookup in CLI
- [x] Add pipeline integration test for OnFix callback
- [x] Add pipeline integration test with multiple concurrent detectors
- [x] Add `io.WriterTo` for SARIF streaming output
- [ ] Add SARIF schema validation test against SARIF 2.1.0 JSON schema — **BLOCKED** (requires vendoring 7K+ line schema)
- [x] Mark GoReleaser as done — `.goreleaser.yml` exists
- [x] Mark Modernize as done — uses `slices.*`, `maps.*`, `iter.Seq`, `errors.AsType`
- [x] Merge duplicate SARIF result builders
- [x] Extract SARIF property key strings to named constants
- [x] Remove `FindingsFromSARIF` always-error stub
- [x] Split `sarif.go` into `sarif_types.go`, `sarif_export.go`, `sarif_import.go`
- [x] Add `Range.IsValid()` to check `End >= Start`
- [x] Add `Position.HasLocation() bool`
- [x] Add `DiffResult` convenience methods
- [x] Add `iter.Seq[Finding]` on `Report.All()`
- [x] Add `Position.IsZero()` helper
- [x] Add `Range.IsSingleLine()` helper
- [x] Add `Finding.HasRange()` helper
- [x] Add `Category.IsSecurity()` helper
- [x] Add `Tag.IsStandard()` method
- [x] Add `Report.CountBySeverity()` convenience method
- [x] Add `Range.Contains(p Position) bool`
- [x] Add `finding.Diff()`
- [x] Add `go/analysis` reverse conversion: `ToDiagnostic()`
- [x] Evaluate `go-sarif` vs hand-rolled — **OWNER_DECISION** (strategic) → Decision: keep hand-rolled. See docs/architecture-decisions.md #9.
- [x] Config file support for library/pipeline (YAML) — `ConfigFromFile`/`ConfigFromReader` in `pipeline/config_file.go`
- [x] Plugin architecture for external detector registration — `DetectorRegistry` in `registry.go` with Register/Build/BuildAll/Names/Has
- [~] ~~Pipeline middleware/interceptor pattern~~ — REMOVED: ghost system, never wired into Config/Pipeline.Run. `StageHook` covers the same use case at finer granularity.
- [x] Per-detector timeout configuration
- [x] Add structured logging (`slog`) to pipeline + CLI
- [x] Progress reporting callback for pipeline
- [x] Implement spatial index for `Correlate` — `IntervalIndex[T]` in `interval_tree.go` with O(log n + k) queries, 5 tests
- [x] Implement streaming merge — `MergeIter()` returning `iter.Seq[Finding]` in `merge.go` with 4 tests
- [x] Extract `findingKey` to shared utility
- [x] Modernize to Go 1.21+ stdlib
- [x] Reduce `FindingsFromSARIF` cognitive complexity
- [x] Consolidate SARIF write methods
- [x] Inline `lock()`/`unlock()` wrappers in `report.go`
- [x] Add GitHub release workflow — `.github/workflows/release.yml` exists with GoReleaser
- [x] Add GoReleaser multi-module config
- [x] Add gosec/staticcheck to CI — golangci-lint already includes both
- [x] Benchmark regression tracking — `benchmark` CI job in `.github/workflows/ci.yml`
- [x] Persist fuzz corpus / seed corpus — 20 fuzz targets with file-based corpus in testdata/fuzz/
- [x] Add LICENSE file
- [x] Add godoc examples for key APIs
- [x] Set up pkg.go.dev documentation — badge in README
- [x] Create `CONTRIBUTING.md`
- [x] Create real-world tool integration guide with govet example — `docs/integration-guide.md`
- [x] Add `Finding` JSON schema
- [x] Remove `samber/do/v2` — not in codebase
- [x] Remove `samber/mo` — not in codebase
- [x] Add tests for 15 untested rule files — `internal/rules/` doesn't exist
- [x] Remove `replace` directive from go.mod
- [x] Delete `internal/events/events.go` — not in codebase
- [ ] Wire into go-structure-linter — **DEFERRED** (external project, not our repo)
- [x] Remove stale `//nolint` directives — audited: all 80 are legitimate
- [x] Archive old status reports
- [x] Remove personal tool configs
- [x] Add `go.work` for local multi-module development — **WONTFIX** (single module, no need)
- [x] LineProvider/SubstringProvider line offset index caching — `lineIndexAware` interface in `pipeline/fix_provider.go`, lazy build via `*[]int` in `pipeline/fix_engine.go`. LineProvider 1000 fixes: 150ms → 859μs (175×), 86MB → 4MB allocs
- [x] Correlate pre-allocation — pre-sized `correlations` with `min(len(findings), maxCorrelations)` and `withRange`/`withoutRange` with `len(fileFindings)` in `merge.go`. 113μs → 83μs (27%), 75 → 51 allocs (32%)
- [x] FixEngine benchmark offset bug fix — `generateOffsetFixes` used Range width 4 but BeforeCode "old()" is 5 chars; OffsetProvider always rejected, silently testing SubstringProvider. Now uses correct offsets from `findOldOccurrences`
- [x] Dedicated LineProvider/SubstringProvider benchmarks — `BenchmarkFixEngine_LineProvider_*` and `BenchmarkFixEngine_Substring_*` in `pipeline/fix_engine_bench_test.go`
- [x] Evaluate SARIF struct pooling — **SKIP** (312KB/100 findings; bytes dominated by un-poolable JSON buffer + per-finding strings; struct headers are ~0.2% of total; sync.Pool complexity/risk unjustified)
- [ ] Watch mode — **DEFERRED** (see [ROADMAP.md](ROADMAP.md#tooling-integrations))
- [ ] IDE plugin stubs — **OUT OF SCOPE v1** (see [ROADMAP.md](ROADMAP.md#tooling-integrations))
- [ ] Web UI — **OUT OF SCOPE v1** (see [ROADMAP.md](ROADMAP.md#out-of-scope-v1))

## ⚪ Unknown / Owner Decision

- [x] Decide `NewFinding` API pattern — accepts `Confidence` type; Builder for complex cases
- [x] Define v1.0.0 release criteria — `docs/RELEASE_CRITERIA.md`
- [x] Decide domain-specific FixProvider module location — separate modules
- [x] Decide `Properties map[string]any` — WONTFIX, intentionally rejected
- [x] Repository structure — flat root package is intentional for a library (not an application)
- [x] `Correlation` JSON tags camelCase
- [x] `PositionOffset` sentinel design — **RESOLVED v0.9.0** (Offset=-1 sentinel; see CHANGELOG [0.9.0])
- [x] Add `ToolInfo.Validate()` method + tests
- [x] Add `RetryConfig.Validate()` method
- [x] `Report.Merge()` → removed in v1.0.0; use `MergeInto` (returns new `*Report`)
- [x] Fix `FixProviders` through CLI config — `-fix-provider` flag + `fixProviders` config field implemented (v0.8.0)

## ⚪ DEFERRED v2.0 (from data-model-review 2026-06-23)

Structural breaking changes that should batch into v2.0 alongside the `Finding` sub-grouping already deferred.

- [ ] Redesign `Position` sentinel conventions — `position.go:23-29` mixes three conventions (0=unset for Line/Column, -1=unset for Offset, zero-value `Position{}` has Offset=0 = valid). Adopt `Option[T]` generic helper for one convention. Cascades into `Range`, `Finding`, `RelatedRef`.
- [ ] Redesign `FixStrategy` as interface-based closed union — `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved`. Eliminates `NormalizeFixStrategy` workaround; "Direct requires BeforeCode" becomes constructor invariant.
- [ ] Cleanup pointer-as-state fields — `Range *Range`, `Suppression *Suppression`, `Suppression.ExpiresAt *time.Time`, `RelatedRef.Range *Range`, `FindingError.Finding *Finding` all encode three states (nil/zero/valid). Replace with value+bool pairs where possible.
- [ ] Convert `Tags []Tag` to `TagSet map[Tag]struct{}` — `finding.go:21`. Encodes set semantics at type level; eliminates need for order-insensitive equality in `finding_equal.go`.
- [ ] Compose `Finding` from embedded sub-structs — `Identity{ID,Rule,ToolName}`, `Location{Position,Range}`, `Classification{Category,Tags,Confidence}`, `Fix{FixStrategy,Suggestion,BeforeCode,AfterCode}`. Allows passing substructs to focused functions. Changes JSON shape — must batch with other v2.0 breaks.

## Recently Completed (2026-06-05)

- GAP-1: `RelatedRef.Range *Range` — deep-clone, validation, equality, SARIF round-trip
- GAP-4: `LSPDiagnosticTag` type + `Unnecessary(1)` / `Deprecated(2)` constants; `Tags` on `LSPDiagnostic`
- GAP-5: `Snippet` field on `SarifRegion`; `findingRegion()` populates; `applySarifPosition()` reads back
- GAP-6: `ToLSP()` emits proper `LSPRange` end from `rel.Range`; `sarifRelatedLocs()` uses `rel.Range`
- GAP-8: `FromLSP()` preserves `DiagnosticTag` values in metadata; reconstructs `RelatedRef.Range`
- 9 new tests added; JSON schema updated; 0 lint issues; race clean
- Comprehensive status report written to `docs/status/2026-06-05_02-20_comprehensive-status-update.md`
- `doc.go` expanded from ~40% to full package documentation
- `USAGE_GUIDE.md` refreshed for v0.4.x features
- `README.md` Fix Providers and Diff sections added

## art-dupl Integration Evaluation Summary

| Gap                                | Status          | Rationale                                                                      |
| ---------------------------------- | --------------- | ------------------------------------------------------------------------------ |
| GAP-1 `RelatedRef.Range`           | ✅ DONE         | Span-based related locations fully supported                                   |
| GAP-2 `GroupID`                    | ❌ DEFERRED     | One-consumer justification insufficient; use `Metadata["go-finding/group-id"]` |
| GAP-3 Per-relationship metadata    | ❌ DEFERRED     | Low value; `Finding.Metadata` workaround exists                                |
| GAP-4 `DiagnosticTag`              | ✅ DONE         | `Unnecessary`/`Deprecated` tags in LSP diagnostics                             |
| GAP-5 `Snippet` in SARIF           | ✅ DONE         | `region.snippet` round-trips properly                                          |
| GAP-6 `ToLSP` uses `rel.Range`     | ✅ DONE         | Proper LSP ranges for related info                                             |
| GAP-7 Strict `Category.IsValid()`  | ❌ BY DESIGN    | `IsValid()` checks format; `IsStandard()` checks membership                    |
| GAP-8 `FromLSP` preserves tags     | ✅ DONE         | Tags stored in metadata as comma-separated integers                            |
| GAP-9 `iter.Seq` on `Report.All()` | ✅ ALREADY DONE | Go 1.26 `iter.Seq` implemented                                                 |

---

_Assisted-by: Crush <crush@charm.land>_

<!--
Notes for maintainers:
- Use `[x]` for done, `[ ]` for open.
- Prefix WONTFIX/DEFERRED/BLOCKED items with the reason in bold.
- Keep items actionable and bounded.
- Archive completed items older than 30 days into docs/status/archive/todo-history.md.
-->
