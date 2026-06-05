# TODO List

**Generated:** 2026-05-20
**Updated:** 2026-06-05 (art-dupl integration gaps closed, docs refreshed)
**Files Processed:** 235

## 🔴 HIGH Priority

- [x] Add bounds checks in `findingFromSARIF` — bounds checks exist via `applySarifPosition`
- [x] Document SARIF critical round-trip loss — documented in doc.go, USAGE_GUIDE.md, FEATURES.md
- [ ] `Finding` struct sub-grouping — **DEFERRED v2** (breaking change)
- [ ] Add `golines` to CI — **BLOCKED** (no .github/workflows/ exists)
- [x] Clean up gopls hints (~12 non-critical) — production code: 0 rangeint, 0 newexpr, 2 mapsloop fixed
- [x] Implement art-dupl integration gaps (GAP-1, GAP-4, GAP-5, GAP-6, GAP-8) — fully implemented, tested, lint clean
- [x] Expand `doc.go` to comprehensive package documentation
- [x] Update `USAGE_GUIDE.md` for v0.4.x
- [ ] Create `.github/workflows/ci.yml` — test, lint, race jobs on PR/push
- [ ] Push 4 unpushed commits to origin
- [ ] Fix `FixApplier` lifecycle — lift creation to Pipeline constructor; prevent backup dir leaks across iterations
- [ ] Restore `cmd/go-finding` test coverage from 70.0% toward 95%
- [ ] Update `README.md` with badges, pipeline diagram, API overview
- [x] Fix pre-commit hook failures — `goconst`, `todo-check`, `library-policy` → All passing. Fixed `err :=` redeclaration bugs in sarif_export.go/sarif_import.go (introduced by previous commit), removed duplicate `pipeline/pipeline_new_test.go`, fixed `FindingsSnapshot()` nil-for-empty behavior.

## 🟡 MEDIUM Priority

- [x] Fix `pipeline/partial.go:107` — uses `IsContextError()`
- [x] Fix `pipeline/fix_applier.go` — `fmt` import actively used
- [x] Fix `cmd/go-finding/integration_test.go:108` — no duplicate, separate functions
- [x] Fix `examples/builder/main.go` compile error — correctly handles both values
- [x] Fix `.golangci.yml` to work without `--no-verify` / `--no-config` flag — removed 6 invalid linter names
- [ ] Fix `.golangci.yml` indentation — **BLOCKED** (BuildFlow auto-configure reformats; current file is consistent 2-space)
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
- [ ] FixEngine: line-offset tracking for cumulative line shifts across multi-fix
- [ ] Make fix strategy composable as interface
- [ ] Add pipeline stage hooks — pre/post hooks for detect, triage, fix, verify
- [x] Update `AGENTS.md` with module split decision context + FixProvider architecture
- [x] Mark 225 lint warnings phantom — `go-structure-linter` is external
- [x] Mark `GitAuthorProvider` phantom — `rule_service.go` does not exist
- [x] Mark `.git/hooks/pre-commit` executable — already executable
- [ ] Interactive TUI — **OUT OF SCOPE v1**
- [ ] Phase 3: Create `.envrc` — **BLOCKED** (no Nix setup)

## 🟢 LOW Priority

- [ ] API stability review — audit all exported symbols for v1.0.0 lock
- [x] Decide `FixStrategyAI` fate — keep as RESERVED placeholder
- [x] Document `BySeverityAtLeast` excludes invalid severities
- [x] Document `Report.All()` yields copies — with shallow copy caveat
- [x] Document `FindByID` returns copy — with shallow copy caveat
- [x] Make `Key()` NUL separator `"\x00"` a named constant
- [x] Document SARIF round-trip losses in user-facing docs
- [x] Document FixStrategyAI placeholder semantics
- [ ] Document provider chain (OffsetProvider→LineProvider→SubstringProvider) in user-facing docs
- [ ] `Position` zero-value safety — **OWNER_DECISION** (breaking change)
- [ ] `Range.End` zero-value ambiguity — **OWNER_DECISION** (breaking change)
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
- [ ] Config file support for library/pipeline (YAML)
- [ ] Plugin architecture for external detector registration
- [ ] Pipeline middleware/interceptor pattern
- [x] Per-detector timeout configuration
- [x] Add structured logging (`slog`) to pipeline + CLI
- [x] Progress reporting callback for pipeline
- [ ] Implement spatial index for `Correlate` — interval tree instead of O(n²)
- [ ] Implement streaming merge
- [x] Extract `findingKey` to shared utility
- [x] Modernize to Go 1.21+ stdlib
- [x] Reduce `FindingsFromSARIF` cognitive complexity
- [x] Consolidate SARIF write methods
- [x] Inline `lock()`/`unlock()` wrappers in `report.go`
- [ ] Add GitHub release workflow — **BLOCKED** (no .github/workflows/)
- [x] Add GoReleaser multi-module config
- [ ] Add gosec/staticcheck to CI — **BLOCKED** (no .github/workflows/)
- [ ] Benchmark regression tracking — **BLOCKED** (no .github/workflows/)
- [ ] Persist fuzz corpus / seed corpus — 17 fuzz targets with no checked-in seed corpus
- [x] Add LICENSE file
- [x] Add godoc examples for key APIs
- [ ] Set up pkg.go.dev documentation
- [x] Create `CONTRIBUTING.md`
- [ ] Create real-world tool integration guide with govet example
- [x] Add `Finding` JSON schema
- [x] Remove `samber/do/v2` — not in codebase
- [x] Remove `samber/mo` — not in codebase
- [x] Add tests for 15 untested rule files — `internal/rules/` doesn't exist
- [x] Remove `replace` directive from go.mod
- [x] Delete `internal/events/events.go` — not in codebase
- [ ] Wire into go-structure-linter — **DEFERRED** (external project)
- [x] Remove stale `//nolint` directives — audited: all 80 are legitimate
- [x] Archive old status reports
- [x] Remove personal tool configs
- [ ] Add `go.work` for local multi-module development
- [ ] Watch mode — **DEFERRED**
- [ ] IDE plugin stubs — **OUT OF SCOPE v1**
- [ ] Web UI — **OUT OF SCOPE v1**

## ⚪ Unknown / Owner Decision

- [x] Decide `NewFinding` API pattern — accepts `Confidence` type; Builder for complex cases
- [x] Define v1.0.0 release criteria — `docs/RELEASE_CRITERIA.md`
- [x] Decide domain-specific FixProvider module location — separate modules
- [x] Decide `Properties map[string]any` — WONTFIX, intentionally rejected
- [ ] Repository structure — **OWNER_DECISION**
- [x] `Correlation` JSON tags camelCase
- [ ] `PositionOffset` sentinel design — **OWNER_DECISION** (breaking change)
- [x] Add `ToolInfo.Validate()` method + tests
- [x] Add `RetryConfig.Validate()` method
- [ ] `Report.Merge()` → return new `*Report` instead of mutating receiver
- [ ] Fix `FixProviders` through CLI config

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
