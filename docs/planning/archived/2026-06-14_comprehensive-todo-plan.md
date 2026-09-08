# Comprehensive TODO Plan — All Open Items

**Date:** 2026-06-14
**Source:** Synthesized from 22 docs files (`docs/*/2026-06-*`), `TODO_LIST.md`, `FEATURES.md`
**Current state:** v0.7.0, 137/152 TODOs done, 90.0% coverage, 0 lint, 0 race issues

---

## Summary

| Tier | Description                       | Count | Blocker         |
| ---- | --------------------------------- | ----- | --------------- |
| 1    | Owner Decisions (blocking v1.0.0) | 8     | Need user input |
| 2    | Quality & Coverage                | 7     | None            |
| 3    | Integration Tests                 | 4     | None            |
| 4    | Feature Gaps                      | 3     | None            |
| 5    | Documentation                     | 9     | None            |
| 6    | Infrastructure & CI               | 8     | None            |
| 7    | Code Quality Polish               | 4     | None            |
| 8    | Blocked / Deferred / v2           | 12    | External        |

**Total open: 55 items** (of which 8 need owner decisions, 35 are actionable now, 12 are blocked/deferred)

---

## TIER 1: Owner Decisions (BLOCKING v1.0.0 — Need User Input)

These are design decisions that cannot be made autonomously. They affect the core data model and determine whether v1.0.0 can ship.

| # | Task                                              | Why It Matters                                                                                                                                                                     | Impact   | Effort   | Breaking? |
| - | ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------- | --------- |
| 1 | **Decide Position zero-value semantics**          | `Position{}` is both `IsZero()=true` AND `HasOffset()=true` (Offset=0 means "byte 0" AND is the zero value). This semantic trap affects every consumer.                            | Critical | Decision | Yes       |
| 2 | **Decide Range.End zero-value semantics**         | `Range.End.Line == 0` means "unset" via `HasEnd()`, but some consumers might expect 0 to mean "end of line 0". Ambiguous.                                                          | Critical | Decision | Yes       |
| 3 | **Decide PositionOffset sentinel design**         | Should Offset use `-1` as "not set"? Or a `*int` pointer? Or a separate `HasOffset bool`? Currently undocumented inconsistency with Line/Column (which use 0).                     | Critical | Decision | Yes       |
| 4 | **Set v1.0.0 release date and criteria**          | `docs/RELEASE_CRITERIA.md` exists but has no concrete checklist with dates or pass/fail thresholds. Project is at 90% coverage with 0 issues — is v1.0 "now" or "after decisions"? | Critical | 1hr      | No        |
| 5 | **Decide Report.Findings unexport timing**        | Field is `// Deprecated:` but still public. `FindingsSnapshot()` migration path exists. When to actually unexport? v1.0? v1.1?                                                     | High     | Decision | Yes       |
| 6 | **Decide Report.Merge() removal timing**          | Deprecated in favor of `MergeInto()`. Currently still exists. When to remove? v1.0.0? v1.1.0?                                                                                      | Medium   | Decision | Yes       |
| 7 | **Decide FixStrategy empty string normalization** | `FixStrategy ""` passes validation but differs from `FixStrategyNone`. Two valid "no fix" states. Normalize or reject empty?                                                       | Medium   | Decision | Yes       |
| 8 | **Decide Named string types**                     | `ToolName`, `RuleName`, `FindingID` as named types instead of raw `string`. Would touch 7+ signatures. Compile-time safety vs migration cost.                                      | Medium   | Decision | Yes       |

**Recommendation from all sessions:** Lock v1.0.0 NOW with current semantics documented as-is. Address items 1-3, 5, 7 in v1.1.0 or v2.0.0. The risk of indefinite postponement outweighs the risk of imperfect-but-documented semantics.

---

## TIER 2: Quality & Coverage (Actionable Now)

| #  | Task                                                      | Why                                                                                                            | Impact | Effort | Files                                             |
| -- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------- |
| 9  | **Fix analysis/ coverage: 79.5% → 90%+**                  | Lowest coverage in project. `adapter.go` has stale test-only imports and undefined symbols in LSP diagnostics. | High   | 1hr    | `analysis/adapter.go`, `analysis/adapter_test.go` |
| 9a | Audit analysis/adapter.go for stale imports and dead code | Root cause of low coverage                                                                                     | —      | 15min  | `analysis/`                                       |
| 9b | Add tests for uncovered analysis/ code paths              | Bring coverage to 90%+                                                                                         | —      | 45min  | `analysis/adapter_test.go`                        |
| 10 | **Add tests for `ApplyWithShiftMap`**                     | New v0.7.0 code, untested                                                                                      | Medium | 20min  | `pipeline/fix_applier_test.go`                    |
| 11 | **Add tests for `groupFindingsBySafePath`**               | New v0.7.0 code, untested                                                                                      | Medium | 20min  | `pipeline/pipeline_detect_test.go`                |
| 12 | **Add tests for `recordShiftMap`**                        | New v0.7.0 code, untested                                                                                      | Medium | 15min  | `pipeline/pipeline_detect_test.go`                |
| 13 | **Pipeline coverage: 90.2% → 93%+**                       | Dropped from 92.8% after v0.7.0 new code. Needs targeted tests for new paths.                                  | Medium | 1hr    | `pipeline/`                                       |

---

## TIER 3: Integration Tests (Actionable Now)

| #  | Task                                                               | Why                                                                                     | Impact | Effort | Files                                |
| -- | ------------------------------------------------------------------ | --------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------ |
| 14 | **Integration test: ConfigFile → ResolveDetectors → Pipeline.Run** | The pieces exist but have never been assembled end-to-end. No proof they work together. | High   | 1hr    | `pipeline/config_file_test.go` (new) |
| 15 | **Integration test: DetectorRegistry → Build → Pipeline.Run**      | Registry is tested in isolation but never wired through to a full pipeline run.         | Medium | 1hr    | `registry_test.go`                   |
| 16 | **Integration test: full pipeline with middleware**                | `ComposeMiddleware` is tested standalone but not through a real pipeline.               | Medium | 1hr    | `pipeline/middleware_test.go`        |
| 17 | **Integration test: type alias backward compat**                   | Verify `pipeline.Detector == finding.Detector` type alias compatibility.                | Low    | 15min  | `pipeline/adapters_test.go`          |

---

## TIER 4: Feature Gaps (Actionable Now)

| #   | Task                                                | Why                                                                                                                            | Impact | Effort | Files                                                    |
| --- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | -------------------------------------------------------- |
| 18  | **FixProviders through CLI config**                 | Last actionable TODO item. Users can't specify custom providers without writing Go code.                                       | Medium | 2hr    | `cmd/go-finding/config.go`, `cmd/go-finding/main.go`     |
| 18a | Design provider name registry pattern               | Providers are function types — need name→constructor registry                                                                  | —      | 30min  | `pipeline/fix_provider.go`                               |
| 18b | Add provider name resolution to ConfigFile          | Parse `fixProviders: [offset, line, substring]` from YAML                                                                      | —      | 30min  | `pipeline/config_file.go`                                |
| 18c | Wire resolved providers into pipeline.New()         | Connect ConfigFile providers to Config.FixProviders                                                                            | —      | 30min  | `cmd/go-finding/config.go`                               |
| 18d | Add CLI flag `-fix-provider`                        | Allow `go-finding -fix-provider=offset,line`                                                                                   | —      | 30min  | `cmd/go-finding/main.go`                                 |
| 19  | **LineShiftMap Range shifting**                     | Only shifts `Position.Line` — does not shift `Range.Start.Line`/`Range.End.Line` or `Position.Column`.                         | Low    | 30min  | `pipeline/fix_applier.go`, `pipeline/pipeline_detect.go` |
| 19a | Extend shift to Range.Start.Line and Range.End.Line | Apply same shift logic to range fields                                                                                         | —      | 20min  | `pipeline/pipeline_detect.go`                            |
| 19b | Extend shift to Position.Column                     | Column may shift if edit is on same line                                                                                       | —      | 10min  | `pipeline/pipeline_detect.go`                            |
| 20  | **SubstringProvider nearest-position heuristic**    | Uses `strings.Index` (first match). Ambiguous when same text appears multiple times. Should match nearest to finding position. | Medium | 30min  | `pipeline/fix_provider.go`                               |

---

## TIER 5: Documentation (Actionable Now)

| #  | Task                                                                | Why                                                                                                                            | Impact | Effort | Files                                 |
| -- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ------------------------------------- |
| 21 | **Godoc examples: IntervalIndex, MergeIter, LineShiftMap**          | New v0.7.0 features have zero discoverable examples.                                                                           | Medium | 30min  | `example_test.go`                     |
| 22 | **Godoc examples: DetectorRegistry, MiddlewareFunc, ConfigFile**    | New v0.7.0 features have zero discoverable examples.                                                                           | Medium | 30min  | `example_test.go`                     |
| 23 | **Update FEATURES.md with v0.7.0 features**                         | Missing: IntervalIndex, MergeIter, LineShiftMap, DetectorRegistry, MiddlewareFunc, ConfigFile, FixStrategyResolver, StageHook. | Medium | 30min  | `FEATURES.md`                         |
| 24 | **Update README.md with v0.7.0 features**                           | README hasn't been updated since v0.6.1. Missing new feature sections.                                                         | Medium | 30min  | `README.md`                           |
| 25 | **CHANGELOG.md entry for v0.7.0 IntervalIndex/LineShiftMap wiring** | CHANGELOG covers sessions 10-11 but not the IntervalIndex→Correlate and LineShiftMap→pipeline wiring.                          | Low    | 15min  | `CHANGELOG.md`                        |
| 26 | **Write v1.0 migration guide**                                      | Consumers need guide for: FindingsSnapshot, unexported types, deprecated APIs.                                                 | High   | 1hr    | `docs/MIGRATION_GUIDE.md` (new)       |
| 27 | **Add CONTRIBUTING.md section on ConfigFile usage**                 | New ConfigFile feature needs usage docs for contributors.                                                                      | Low    | 30min  | `CONTRIBUTING.md`                     |
| 28 | **Update doc.go with new feature examples**                         | Package docs don't mention IntervalIndex, MergeIter, StageHook, MiddlewareFunc, etc.                                           | Medium | 30min  | `doc.go`                              |
| 29 | **Document Tag/Category semantic overlap**                          | `TagSecurity` and `CategorySecurity` have no structural link despite semantic overlap. Document they're separate axes.         | Low    | 15min  | `doc.go` or `docs/DOMAIN_LANGUAGE.md` |

---

## TIER 6: Infrastructure & CI (Actionable Now)

| #   | Task                                                    | Why                                                                                                        | Impact | Effort | Files                                |
| --- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------ |
| 30  | **Fix go.sum stale entries**                            | BuildFlow warns about 9 stale entries. `go mod tidy` was run but entries persist.                          | Low    | 5min   | `go.sum`                             |
| 31  | **Fix flake.nix BuildFlow failures**                    | 6 nix checks fail (missing tools, meta attributes). Pre-existing but annoying for pre-commit hooks.        | Medium | 1hr    | `flake.nix`                          |
| 31a | Identify the 6 failing checks                           | Run `nix flake check` and catalog failures                                                                 | —      | 15min  | —                                    |
| 31b | Fix missing tool references                             | Add missing nixpkgs to flake inputs/devShell                                                               | —      | 30min  | `flake.nix`                          |
| 31c | Fix meta attributes                                     | Add missing meta fields                                                                                    | —      | 15min  | `flake.nix`                          |
| 32  | **Benchmark: old Correlate vs IntervalIndex Correlate** | IntervalIndex is wired into Correlate but never benchmarked against old approach. No proof of improvement. | Low    | 30min  | `merge_bench_test.go` (new)          |
| 33  | **Add benchmark regression thresholds to CI**           | Benchmark CI job exists but has no failure threshold. Should fail if `ns/op` increases by >20%.            | Medium | 30min  | `.github/workflows/ci.yml`           |
| 34  | **Add .github/dependabot.yml**                          | Auto-dependency updates for golang.org/x packages.                                                         | Low    | 15min  | `.github/dependabot.yml` (new)       |
| 35  | **GoReleaser release with v0.7.0 tag**                  | 10 commits ahead of origin. v0.7.0 tagged but not pushed.                                                  | High   | 30min  | —                                    |
| 36  | **Create GitHub release for v0.7.0**                    | Release notes, changelog link.                                                                             | Medium | 15min  | GitHub UI / `gh release create`      |
| 37  | **Fuzz CategoryForLinter**                              | Add fuzz target for case-insensitive lookup edge cases.                                                    | Low    | 15min  | `category_linter_fuzz_test.go` (new) |

---

## TIER 7: Code Quality Polish (Actionable Now)

| #  | Task                                                      | Why                                                                                                   | Impact | Effort | Files                                                             |
| -- | --------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ------ | ------ | ----------------------------------------------------------------- |
| 38 | **slices.Collect modernization pass**                     | 12 candidates identified. Applied `cmp.Or` to 2 locations. Full pass deferred to avoid churn.         | Low    | 30min  | `filter.go`, `report.go`, `pipeline/verify.go`, `sarif_import.go` |
| 39 | **Add Compare() method to Category**                      | Matching `Severity.Compare()` and `Confidence.Compare()` pattern.                                     | Low    | 15min  | `category.go`                                                     |
| 40 | **Create v1.0.0 release checklist**                       | `docs/RELEASE_CRITERIA.md` exists but needs concrete checklist with pass/fail thresholds and dates.   | High   | 1hr    | `docs/RELEASE_CRITERIA.md`                                        |
| 41 | **Audit all deprecated APIs for v1.0.0 removal timeline** | `RecordFix()`, `Report.Merge()`, `Report.Findings`, `CountBySeverity()` free function. Need timeline. | Medium | 1hr    | `docs/API_STABILITY.md`                                           |

---

## TIER 8: Blocked / Deferred / v2 (Not Actionable Now)

| #  | Task                                                                        | Status          | Blocker                                             |
| -- | --------------------------------------------------------------------------- | --------------- | --------------------------------------------------- |
| 42 | Finding struct sub-grouping (`FindingCore` + `FixInfo` + `SuppressionInfo`) | DEFERRED v2     | Breaking change — wait until v1.0 lock              |
| 43 | Add `golines` to CI                                                         | BLOCKED         | treefmt-nix doesn't support golines                 |
| 44 | Fix BuildFlow auto-configure loop                                           | BLOCKED         | External tool generates broken Go code              |
| 45 | Interactive TUI                                                             | OUT OF SCOPE v1 | —                                                   |
| 46 | Create `.envrc`                                                             | BLOCKED         | No Nix setup                                        |
| 47 | SARIF schema validation against official 2.1.0                              | BLOCKED         | Requires vendoring 7K+ line JSON schema             |
| 48 | Wire into go-structure-linter                                               | DEFERRED        | External project                                    |
| 49 | Watch mode for continuous analysis                                          | DEFERRED        | —                                                   |
| 50 | IDE plugin stubs                                                            | OUT OF SCOPE v1 | —                                                   |
| 51 | Web UI                                                                      | OUT OF SCOPE v1 | —                                                   |
| 52 | Code generation for enum types                                              | Nice to have    | 7× enum pattern boilerplate; Go generics can't help |
| 53 | Domain event emission from pipeline stages                                  | Nice to have    | StageHook covers part of this                       |

---

## Recommended Execution Order

### Phase 1: Quick Wins (< 30 min each, high trust value)

| Order | Task                                     | Effort |
| ----- | ---------------------------------------- | ------ |
| 1     | #30 Fix go.sum stale entries             | 5min   |
| 2     | #25 CHANGELOG.md entry for v0.7.0 wiring | 15min  |
| 3     | #39 Add Compare() to Category            | 15min  |
| 4     | #37 Fuzz CategoryForLinter               | 15min  |
| 5     | #34 Add .github/dependabot.yml           | 15min  |
| 6     | #29 Document Tag/Category overlap        | 15min  |
| 7     | #17 Type alias backward compat test      | 15min  |

### Phase 2: Quality & Coverage (1-2 hr total)

| Order | Task                                                                        | Effort |
| ----- | --------------------------------------------------------------------------- | ------ |
| 8     | #9 Fix analysis/ coverage → 90%+                                            | 1hr    |
| 9     | #10-12 Tests for ApplyWithShiftMap, groupFindingsBySafePath, recordShiftMap | 55min  |
| 10    | #13 Pipeline coverage → 93%+                                                | 1hr    |

### Phase 3: Documentation (2-3 hr total)

| Order | Task                                      | Effort |
| ----- | ----------------------------------------- | ------ |
| 11    | #21-22 Godoc examples for v0.7.0 features | 1hr    |
| 12    | #23 Update FEATURES.md                    | 30min  |
| 13    | #24 Update README.md                      | 30min  |
| 14    | #28 Update doc.go                         | 30min  |

### Phase 4: Integration Tests (2-3 hr total)

| Order | Task                                             | Effort |
| ----- | ------------------------------------------------ | ------ |
| 15    | #14 ConfigFile → ResolveDetectors → Pipeline.Run | 1hr    |
| 16    | #15 DetectorRegistry → Build → Pipeline.Run      | 1hr    |
| 17    | #16 Full pipeline with middleware                | 1hr    |

### Phase 5: Feature Gaps (3-4 hr total)

| Order | Task                                             | Effort |
| ----- | ------------------------------------------------ | ------ |
| 18    | #20 SubstringProvider nearest-position heuristic | 30min  |
| 19    | #19 LineShiftMap Range shifting                  | 30min  |
| 20    | #18 FixProviders through CLI config              | 2hr    |

### Phase 6: Release Prep (2-3 hr total)

| Order | Task                                    | Effort |
| ----- | --------------------------------------- | ------ |
| 21    | #40 v1.0.0 release checklist            | 1hr    |
| 22    | #41 Audit deprecated APIs               | 1hr    |
| 23    | #26 v1.0 migration guide                | 1hr    |
| 24    | #35-36 Push v0.7.0 tag + GitHub release | 45min  |

### Phase 7: Infrastructure Polish (1-2 hr total)

| Order | Task                                      | Effort |
| ----- | ----------------------------------------- | ------ |
| 25    | #32 Benchmark old vs new Correlate        | 30min  |
| 26    | #33 Benchmark regression thresholds in CI | 30min  |
| 27    | #31 Fix flake.nix BuildFlow failures      | 1hr    |
| 28    | #38 slices.Collect modernization          | 30min  |

### Phase 8: Owner Decisions (BLOCKING — present to user)

| Order | Task                                      | Why                         |
| ----- | ----------------------------------------- | --------------------------- |
| 29    | #1-3 Position/Range/Offset semantics      | Blocks v1.0 API freeze      |
| 30    | #4 v1.0.0 release date                    | Blocks all release planning |
| 31    | #5-6 Report.Findings/Merge removal timing | Blocks API cleanup          |
| 32    | #7 FixStrategy normalization              | Blocks type safety          |
| 33    | #8 Named string types                     | Blocks type safety          |

---

## Cross-Reference: Recurring Themes Across All Sessions

These themes appear in 3+ status reports and represent the highest-signal priorities:

| Theme                               | Sessions mentioning      | Status                                      |
| ----------------------------------- | ------------------------ | ------------------------------------------- |
| Position/Range zero-value semantics | 6, 7, 8, 9, 10, 11       | OWNER_DECISION (unresolved since session 4) |
| v1.0.0 release timeline             | 4, 5, 6, 7, 8, 9, 10, 11 | OWNER_DECISION (unresolved since session 4) |
| Report.Findings encapsulation       | 4, 5, 6, 7, 8, 9, 10, 11 | Deprecated, not unexported                  |
| analysis/ coverage < 80%            | 8, 9, 10, 11             | Not addressed                               |
| Integration tests for new features  | 9, 10, 11                | Not started                                 |
| Godoc examples for new features     | 9, 10, 11                | Not started                                 |
| FixProviders CLI config             | 4, 5, 6, 7, 8, 9, 10, 11 | Not started (7 sessions!)                   |
| SubstringProvider fragility         | 4, 5, 6, 7               | Not addressed                               |

---

_Assisted-by: Crush <crush@charm.land>_
