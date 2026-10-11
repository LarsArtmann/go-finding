# Session 18 — Split-Brain Resolution & Post-Resolution Cleanup

**Date:** 2026-06-18 08:30 CEST
**Branch:** master (pushed)
**Commits this session:** 10 (`85fb10f` → `e94f0e3`)
**Since last tag (v0.8.0):** 10 commits

---

## Executive Summary

This session executed a **full split-brain audit and resolution** of the go-finding data model. The work was done in three phases:

1. **Audit** — Identified 11 split-brain issues in `docs/research/SPLIT-BRAIN.html` (80KB, every claim runtime-verified)
2. **Resolution** — Fixed all 11 issues in 5 atomic commits, each tier-tested independently
3. **Follow-up** — Caught and fixed self-review gaps (Validate dead code, missing ADRs, stale JSON schema, missing CHANGELOG)

**Result:** Both Critical v1.0.0 release blockers (#1 Position zero-value, #2 FixStrategy duality) are **resolved**. All 3 documented blockers in `RELEASE_CRITERIA.md` are marked RESOLVED. Build passes, race detector clean, 1506 test functions pass, coverage improved.

---

## a) FULLY DONE ✅

### Split-Brain Resolution — All 11 Issues

| #  | Issue                                                     | Severity | Commit    | Resolution                                                                                                                        |
| -- | --------------------------------------------------------- | -------- | --------- | --------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Position zero-value lies about its own state              | Critical | `b0eadc6` | Adopted `-1` sentinel for "unset" offset. `Position{}.IsZero()`=false (Offset=0 is valid byte 0). All constructors set Offset=-1. |
| 2  | Two valid "no fix" states: `""` vs `"none"`               | Critical | `5f0af9b` | `NormalizeFixStrategy()` helper. `Finding.Normalized()` method. Called by Builder, SARIF import, Equal().                         |
| 3  | Four fixability predicates disagree                       | High     | `a314808` | `HasFix()` now requires code for `FixStrategyDirect`. Documented fixability lattice: `IsAutoFixable() ⟹ HasFix()`.                |
| 4  | Category and Tags share vocabulary, can conflict          | High     | `a314808` | Validation invariant: rejects conflicting standard Category/Tags. ADR #13 for full deprecation plan.                              |
| 5  | Three duration fields, none authoritative                 | High     | `5f0af9b` | Removed `Summary.DurationMs`. `Metrics.TotalDuration()` is single source of truth.                                                |
| 6  | Three different finding identity definitions              | Medium   | `a314808` | ADR #12: GenerateID output is canonical. Key() documented as fallback.                                                            |
| 7  | SARIF exports two disagreeing regions per finding         | Medium   | `539151a` | `findingFixRegion()` now respects `Range.End`, matching location region.                                                          |
| 8  | SARIF property bag uses `map[string]any` against doctrine | Medium   | `5f0af9b` | Documented doctrine boundary in `sarifResult.Properties` doc.                                                                     |
| 9  | Suppression validation accepts unknown kinds              | Low      | `539151a` | `Suppression.IsValid()` now calls `Kind.IsValid()`.                                                                               |
| 10 | Category/Tag twin IsValid/IsStandard methods              | Low      | `539151a` | Documented clear "when to use which" contract.                                                                                    |
| 11 | HasFix/HasSuggestion free-function name shadowing         | Cosmetic | `539151a` | Added `WithFix`/`WithSuggestion`. Deprecated old functions.                                                                       |

### Documentation & Testing

| Deliverable                                       | Status | File                                                       |
| ------------------------------------------------- | ------ | ---------------------------------------------------------- |
| SPLIT-BRAIN.html audit report (80KB)              | ✅     | `docs/research/SPLIT-BRAIN.html`                           |
| Pareto execution plan with mermaid graph          | ✅     | `docs/planning/2026-06-17_22-59_SPLIT-BRAIN-RESOLUTION.md` |
| 6 integration tests for split-brain fixes         | ✅     | `splitbrain_test.go`                                       |
| ADR #12: Canonical Finding Identity               | ✅     | `docs/architecture-decisions.md`                           |
| ADR #13: Category/Tags Deprecation Plan           | ✅     | `docs/architecture-decisions.md`                           |
| CHANGELOG entries for all 11 fixes                | ✅     | `CHANGELOG.md`                                             |
| DOMAIN_LANGUAGE.md identity definitions           | ✅     | `docs/DOMAIN_LANGUAGE.md`                                  |
| RELEASE_CRITERIA.md blockers marked resolved      | ✅     | `docs/RELEASE_CRITERIA.md`                                 |
| AGENTS.md gotchas updated                         | ✅     | `AGENTS.md`                                                |
| JSON schema (durationMs removed)                  | ✅     | `docs/schemas/report.schema.json`                          |
| SPLIT-BRAIN.html status badges (44 "✅ Resolved") | ✅     | `docs/research/SPLIT-BRAIN.html`                           |

### Codebase Health Metrics

| Metric                  | Value              | Trend                         |
| ----------------------- | ------------------ | ----------------------------- |
| **Production LOC**      | 11,068 (192 files) | —                             |
| **Test LOC**            | 23,666             | 2.14× test-to-code ratio      |
| **Test functions**      | 1,506              | ↑ (added 6 split-brain tests) |
| **Direct dependencies** | 6                  | Minimal ✅                    |
| **Lint issues**         | 0                  | ✅                            |
| **Build**               | Passes             | ✅                            |
| **Race detector**       | Passes (stable)    | ✅                            |
| **go vet**              | Passes             | ✅                            |

### Test Coverage by Package

| Package                  | Coverage | Trend            |
| ------------------------ | -------- | ---------------- |
| `internal/benchutil/`    | 100.0%   | —                |
| `analysis/`              | 94.1%    | —                |
| `internal/detectors/`    | 96.1%    | —                |
| `internal/gotoken/`      | 92.7%    | —                |
| Root package (`finding`) | 93.6%    | ↑ (from 92.3%)   |
| `pipeline/`              | 95.2%    | —                |
| `cmd/go-finding/`        | 91.2%    | —                |
| `pipeline/goast/`        | 80.8%    | Needs more tests |

### v1.0.0 Release Blockers — ALL RESOLVED ✅

| Blocker                          | Decision                                        | Status                 |
| -------------------------------- | ----------------------------------------------- | ---------------------- |
| #1 Position zero-value semantics | Option A: `-1` sentinel                         | ✅ Resolved            |
| #2 FixStrategy "" vs "none"      | Normalize via `NormalizeFixStrategy()`          | ✅ Resolved            |
| #3 Report.Findings unexport      | Internal migration done; unexport at v1.0.0 tag | ✅ Resolved (internal) |

---

## b) PARTIALLY DONE 🟡

### Deprecated APIs (Scheduled for v1.0.0 Removal)

| API                               | Replacement                       | Status                              |
| --------------------------------- | --------------------------------- | ----------------------------------- |
| `Report.Findings` (public field)  | `Report.FindingsSnapshot()`       | Deprecated, internal migration done |
| `Report.Merge()`                  | `Report.MergeInto()`              | Deprecated                          |
| `OnStage` callback                | `StageHooks`                      | Deprecated                          |
| `Metrics.RecordFix()`             | `Metrics.RecordFixes(1)`          | Deprecated                          |
| `CountBySeverity()` free function | `Report.CountBySeverity()` method | Deprecated                          |
| `HasFix()` free function          | `WithFix()`                       | Deprecated this session             |
| `HasSuggestion()` free function   | `WithSuggestion()`                | Deprecated this session             |
| `Finding.Category` field          | `Finding.Tags` (ADR #13)          | Deprecation planned for v1.0.0      |

### Category/Tags Deprecation (ADR #13)

- ✅ Validation invariant active (prevents conflicts)
- ✅ ADR #13 written with migration plan
- 🟡 Full deprecation deferred to v1.0.0 batch
- 🟡 `Summary.ByCategory` still exists; `Summary.ByTag` not yet added

### CLI Features

- 🟡 **FixProviders through CLI config** — Provider name registry exists, `-fix-provider` flag exists, but ConfigFile→provider resolution is incomplete
- 🟡 **GoReleaser** — `.goreleaser.yml` fully configured but no release has been cut since v0.8.0

---

## c) NOT STARTED ⬜

### v1.0.0 Release Preparation

- ⬜ Cut v1.0.0 tag and GitHub release
- ⬜ Remove all deprecated APIs (breaking change batch)
- ⬜ Unexport `Report.Findings` field
- ⬜ Final v1.0.0 migration guide review

### Architecture Improvements Identified During Audit

- ⬜ **Named string types** — `ToolName`, `RuleName`, `FindingID` instead of raw `string` (prevents parameter mixing at call sites)
- ⬜ **LineShiftMap Range/Column completeness** — Only shifts `Position.Line` currently; doesn't fully handle multi-line range shifting
- ⬜ **SubstringProvider nearest-position heuristic** — Uses first-match `strings.Index`; ambiguous with multiple occurrences
- ⬜ **`go-structure-linter` noise** — 37 ERROR issues from root-package-files (intentional design decision, accepted noise)

### Documentation Gaps

- ⬜ **FEATURES.md** — Missing v0.8.0 features (split-brain fixes, Normalized method, ADRs)
- ⬜ **README.md** — Last updated for v0.7.0; missing Position sentinel changes, HasFix semantics change
- ⬜ **USAGE_GUIDE.md** — Needs update for new Normalized() method and WithFix/WithSuggestion
- ⬜ **Godoc examples** — `Normalized()`, `WithFix`, `WithSuggestion`, `NormalizeFixStrategy` lack runnable examples

---

## d) TOTALLY FUCKED UP 💥 (Honest Assessment)

### Mistakes Made and Fixed This Session

1. **Validate() dead code** — I wrote `f.FixStrategy = FixStrategyNone` inside `Validate()`, which has a value receiver. The assignment was dead code — mutations on value receivers don't persist. **Fixed in `dd307ca`** by adding `Finding.Normalized()` method and removing the dead assignment. This was caught during the self-review phase, not by tests — meaning my initial fix gave a false sense of correctness.

2. **JSON schema staleness** — I removed `Summary.DurationMs` from the struct but forgot to remove it from `docs/schemas/report.schema.json`. The schema lied about what fields existed. **Fixed in `e94f0e3`.** This was a classic "changed the code, forgot the contract" mistake.

3. **Missing promised deliverables** — The execution plan promised ADR #12, ADR #13, CHANGELOG entries, and DOMAIN_LANGUAGE definitions. I committed the code fixes and declared done without writing any of these. **Fixed in `e94f0e3`** after the self-review caught the gap.

4. **Category/Tags validation may be too strict** — The invariant I added rejects `Category="security", Tags=["performance"]` but allows `Category="security", Tags=["security", "performance"]`. This is correct but could surprise consumers who set Tags independently of Category without intending conflict. **Not yet a bug** (no consumer reports), but the invariant's semantics should be validated against real usage before v1.0.0.

### Pre-existing Issues (Not Fixed This Session)

- **`go-finding` and `result` binaries in working directory** — Not git-tracked, but trigger `go-structure-linter` warnings. Should be in `.gitignore` or cleaned up.
- **math/rand v1/v2 split** — Production uses `math/rand/v2`; tests use `math/rand` (v1) due to `testing/quick` API constraint. No clean fix until Go's `testing/quick` supports v2.

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process Improvements

1. **Always check the value vs pointer receiver** — The Validate() dead code bug is a recurring Go footgun. Consider a lint rule or convention: "if a method normalizes, it must have a pointer receiver or return a new value."

2. **Schema is a contract, not a byproduct** — When removing a struct field, grep ALL schemas, docs, and examples. The `durationMs` staleness was preventable with a simple `grep -r DurationMs docs/`.

3. **Write deliverables before declaring done** — The execution plan listed ADRs and CHANGELOG as tasks. I skipped them and declared done. The self-review caught it, but the pattern is: don't mark a plan "done" until every task is actually committed.

4. **Test with `nix flake check`** — The nix build caught a compile error (`no new variables on left side of :=`) that `go test` didn't catch in the same way. Use nix flake check as a pre-commit gate, not just BuildFlow.

### Architecture Improvements

5. **Named string types** — `type ToolName string`, `type RuleName string`, `type FindingID string`. Prevents mixing up parameters at call sites. Currently `GenerateID(toolName, rule string, pos Position)` — two `string` params that could be swapped. With named types, `GenerateID(tool ToolName, rule RuleName, pos Position)` is swap-proof.

6. **Consider `errors.Join` everywhere** — The project already uses it in `Validate()`. Extend to pipeline error collection and SARIF import for consistent error aggregation.

7. **Position as a value object** — Position is currently a mutable struct with public fields. Consider making it more value-object-like: unexported fields, constructor-only creation, `With*` methods for modifications. This would have prevented the zero-value split-brain entirely.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

Sorted by **impact / effort** ratio (highest first).

### Tier A: High Impact, Low Effort (Do First)

| # | Task                                                              | Impact | Effort | Why                                                                                  |
| - | ----------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------ |
| 1 | **Update FEATURES.md with v0.8.0 features**                       | High   | 30min  | Anyone reading FEATURES.md doesn't know about split-brain fixes, Normalized(), ADRs  |
| 2 | **Update README.md for v0.8.0**                                   | High   | 30min  | Position sentinel changes and HasFix semantics are consumer-visible breaking changes |
| 3 | **Add godoc examples: Normalized, WithFix, NormalizeFixStrategy** | Medium | 30min  | New API surface has zero discoverable examples                                       |
| 4 | **Clean up working directory binaries (`go-finding`, `result`)**  | Low    | 5min   | Triggers linter warnings; add to .gitignore                                          |
| 5 | **Add `Summary.ByTag` map**                                       | Medium | 30min  | Completes the Tags-as-primary-classification path from ADR #13                       |
| 6 | **Fuzz test for Position sentinel consistency**                   | Medium | 30min  | Property: `Position{}.HasOffset() && !Position{}.IsZero()` always holds              |
| 7 | **Fuzz test for FixStrategy normalization idempotency**           | Medium | 30min  | Property: `NormalizeFixStrategy(NormalizeFixStrategy(x)) == NormalizeFixStrategy(x)` |
| 8 | **Property test: HasFix ⊇ IsAutoFixable invariant**               | Medium | 15min  | Proves the lattice property for all FixStrategy + code combinations                  |

### Tier B: Medium Impact, Medium Effort

| #  | Task                                                             | Impact   | Effort | Why                                                              |
| -- | ---------------------------------------------------------------- | -------- | ------ | ---------------------------------------------------------------- |
| 9  | **Named string types: ToolName, RuleName, FindingID**            | High     | 2hr    | Prevents parameter mixing at call sites; type safety improvement |
| 10 | **v1.0.0 release: remove all deprecated APIs**                   | Critical | 2hr    | Breaking change batch; clears the deprecation debt               |
| 11 | **v1.0.0 release: unexport Report.Findings**                     | High     | 1hr    | Final step of internal migration                                 |
| 12 | **Cut v1.0.0 tag and GitHub release**                            | Critical | 1hr    | The finish line                                                  |
| 13 | **Complete FixProviders through CLI config**                     | Medium   | 2hr    | Users can't specify custom providers without writing Go code     |
| 14 | **LineShiftMap: extend to full Range + Column shifting**         | Medium   | 1hr    | Post-fix finding positions can be wrong for multi-line ranges    |
| 15 | **Integration test: full pipeline with all split-brain fixes**   | Medium   | 1hr    | End-to-end proof that the fixes work together                    |
| 16 | **SubstringProvider nearest-position heuristic**                 | Medium   | 1hr    | First-match is ambiguous with multiple occurrences               |
| 17 | **Update USAGE_GUIDE.md for v0.8.0+**                            | Medium   | 1hr    | Consumers need guidance for new APIs                             |
| 18 | **Add Category/Tags conflict resolution guide to migration doc** | Low      | 30min  | Consumers with existing Category+Tags data need migration path   |

### Tier C: Lower Impact or Higher Effort

| #  | Task                                                           | Impact | Effort | Why                                                            |
| -- | -------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------- |
| 19 | **Benchmark regression check after Position changes**          | Medium | 1hr    | Verify no performance regression from Offset=-1 default        |
| 20 | **Consider `sync.Pool` for line offset index**                 | Low    | 2hr    | Evaluated in ADR #11 as skip; revisit if profiling shows need  |
| 21 | **`slices.Collect` modernization pass**                        | Low    | 1hr    | 12 candidates identified; mechanical modernization             |
| 22 | **Add `FindingID` type and use in RelatedRef, Correlation**    | Medium | 2hr    | Part of named types effort (#9); separate if needed            |
| 23 | **Position as value object (unexport fields, With\* methods)** | High   | 4hr    | Would have prevented split-brain #1 entirely; large refactor   |
| 24 | **Explore `cmp.Ordered` for Compare methods**                  | Low    | 1hr    | Modernize Severity/Confidence/Category Compare implementations |
| 25 | **Add OpenAPI/JSON Schema generation from Go types**           | Low    | 2hr    | Currently hand-maintained schemas; could auto-generate         |

---

## g) TOP #1 QUESTION 🤔

**Should we cut v1.0.0 now, or do one more minor release (v0.9.0) to let consumers test the breaking changes?**

The split-brain fixes changed consumer-visible behavior:

- `Position{}.IsZero()` now returns `false` (was `true`)
- `HasFix()` returns `false` for `FixStrategyDirect` without code (was `true`)
- `Summary.DurationMs` field removed
- `Finding.Equal()` normalizes FixStrategy (previously unequal findings now compare equal)

These are all correct fixes, but they ARE behavioral changes. A v0.9.0 would let consumers discover these via a minor version bump (Go semver: no breaking changes within a minor). Alternatively, we could document these as "the v1.0.0 contract" and cut v1.0.0 directly — since we're pre-1.0, consumers should expect breaking changes.

**My recommendation:** Cut v1.0.0 directly. The project has been at v0.8.0 with "API-stable beta" status for a while. The split-brain fixes are the last structural blockers. Going to v0.9.0 just delays the inevitable and adds a version consumers have to skip.

**But I can't decide this myself** — it's a product/owner decision about release cadence and consumer communication.

---

## Session Statistics

| Metric           | Value                                |
| ---------------- | ------------------------------------ |
| Commits          | 10                                   |
| Files changed    | ~25                                  |
| Lines added      | ~2,500                               |
| Lines removed    | ~200                                 |
| Issues resolved  | 11 (all split-brain)                 |
| Tests added      | 6 integration + updated ~10 existing |
| ADRs written     | 2 (#12, #13)                         |
| BuildFlow checks | 34/34 passing                        |
| Race detector    | Clean                                |

---

_Assisted-by: Crush <crush@charm.land>_
