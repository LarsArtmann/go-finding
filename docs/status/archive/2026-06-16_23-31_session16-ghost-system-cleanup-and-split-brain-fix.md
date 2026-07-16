# Session 16 — Ghost System Cleanup & Split-Brain Fix

**Date:** 2026-06-16 23:31 CEST
**Branch:** master (pushed)
**Commits this session:** 4 (`ca958f1` → `3ebbb30`)
**Since last tag (v0.6.1):** 20 commits

---

## Executive Summary

This session started with a brutal self-review triggered by "What's something stupid we do?" The answer: **two ghost systems** (exported APIs with zero consumers, zero wiring, zero tests) and a **split-brain notification pattern** where every pipeline stage boundary called two separate notification mechanisms for the same event.

All three issues were fixed, committed, and pushed. Zero lint issues. All tests pass with race detector. Pre-commit BuildFlow passes all 34 checks.

---

## a) FULLY DONE ✅

### This Session's Work

| Commit    | Title                                                             | Impact                                                                   |
| --------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------ |
| `ca958f1` | Remove two ghost systems: FixStrategyResolver and MiddlewareFunc  | Deleted 52 lines of dead code that looked like features                  |
| `3985edf` | Fix lint warnings in goast provider and update stale TODO entries | 2 pre-existing lint issues resolved; 2 stale TODO entries marked REMOVED |
| `f687179` | Eliminate OnStage/StageHooks split brain: unify notification path | Removed `notifyStage()`, folded OnStage into `fireStageHook()`           |
| `3ebbb30` | Document OnStage deprecation and split-brain elimination in docs  | AGENTS.md + CHANGELOG.md updated                                         |

### Codebase Health Metrics (as of this session)

| Metric                  | Value                                    | Status                   |
| ----------------------- | ---------------------------------------- | ------------------------ |
| **Production LOC**      | 10,492 (181 files, excl. tests/examples) | —                        |
| **Test LOC**            | 22,534                                   | 2.15× test-to-code ratio |
| **Test functions**      | 631                                      | Comprehensive            |
| **Benchmarks**          | 48                                       | Comprehensive            |
| **Example functions**   | 29                                       | Good                     |
| **Fuzz targets**        | 22 (all with seed corpus)                | Comprehensive            |
| **Direct dependencies** | 6                                        | Minimal ✅               |
| **Lint issues**         | 0                                        | ✅                       |
| **Build**               | Passes                                   | ✅                       |
| **Race detector**       | Passes (stable across 30+ runs)          | ✅                       |

### Test Coverage by Package

| Package                  | Coverage | Trend                                   |
| ------------------------ | -------- | --------------------------------------- |
| `analysis/`              | 94.1%    | ↑ from 79.5% (Session 14 fixes)         |
| `internal/detectors/`    | 96.1%    | Stable                                  |
| `internal/gotoken/`      | 92.7%    | Stable                                  |
| Root package (`finding`) | 92.3%    | ↑ from 95.7% v0.6.1 (new code dilution) |
| `pipeline/`              | 90.4%    | ↓ from 92.8% (Session 11 new code)      |
| `cmd/go-finding/`        | 83.5%    | ↓ from 90.7% (new features)             |
| `pipeline/goast/`        | 80.8%    | New package, needs more tests           |

### Core Architecture (All Complete)

- ✅ **Unified Finding type** — 50+ fields covering all tool outputs
- ✅ **Pipeline** — detect → process → triage → apply → verify loop
- ✅ **Byte-level FixEngine** — 452× faster (Session 12 optimization)
- ✅ **SARIF round-trip** — Export + import with full property bag fidelity
- ✅ **LSP conversion** — Diagnostic ↔ Finding bidirectional
- ✅ **GoASTProvider** — AST-aware fix provider for .go files
- ✅ **DetectorRegistry** — Thread-safe plugin architecture
- ✅ **ConfigFile** — YAML/JSON config loading
- ✅ **IntervalIndex[T]** — Generic overlap queries (wired into Correlate)
- ✅ **LineShiftMap** — Byte-offset-aware line shift tracking
- ✅ **MergeIter()** — Streaming iter.Seq merge
- ✅ **StageHook** — Per-stage before/after hooks with abort capability
- ✅ **FindingProcessor** — Composable transforms between detect and triage
- ✅ **GeneratedFileFilter** — Auto-generated file detection (sqlc, protobuf, etc.)
- ✅ **TriageFunc** — Customizable triage logic
- ✅ **ByteLevelConflictDetection** — Precise overlap detection
- ✅ **Correlate** — Cross-tool finding correlation with IntervalIndex

---

## b) PARTIALLY DONE 🟡

### Deprecated APIs (Scheduled for v1.0.0 Removal)

| API                               | Replacement                       | Status                                                                     |
| --------------------------------- | --------------------------------- | -------------------------------------------------------------------------- |
| `Report.Findings` (public field)  | `Report.FindingsSnapshot()`       | Deprecated, internal migration done (`findingsLocked()`, `readFindings()`) |
| `Report.Merge()`                  | `Report.MergeInto()`              | Deprecated                                                                 |
| `OnStage` callback                | `StageHooks`                      | Deprecated this session — now fires inside `fireStageHook`                 |
| `Metrics.RecordFix()`             | `Metrics.RecordFixes(1)`          | Deprecated                                                                 |
| `CountBySeverity()` free function | `Report.CountBySeverity()` method | Deprecated                                                                 |

These are all tracked in `docs/architecture-decisions.md` ADR #11 for the v1.0.0 breaking change batch.

### CLI Features (Partially Wired)

- 🟡 **FixProviders through CLI config** — Provider name registry exists (`cmd/go-finding/fix_provider_registry.go`), `-fix-provider` flag exists, but ConfigFile→provider resolution is incomplete (TODO #18a-d)
- 🟡 **GoReleaser** — `.goreleaser.yml` fully configured but no release has been cut since v0.6.1

### Documentation Gaps

- 🟡 **FEATURES.md** — Missing v0.7.0 features (IntervalIndex, MergeIter, LineShiftMap, DetectorRegistry, StageHook)
- 🟡 **README.md** — Last updated for v0.6.1; missing GoASTProvider, CategoryForLinter updates
- 🟡 **doc.go** — Package docs don't mention several new features
- 🟡 **Godoc examples** — IntervalIndex, MergeIter, LineShiftMap, ConfigFile, DetectorRegistry lack runnable examples
- 🟡 **v1.0 Migration Guide** — Not yet written (consumers need guidance)

---

## c) NOT STARTED ⬜

### Integration Tests (High Value, Never Done)

| #   | What                                         | Why It Matters                                 |
| --- | -------------------------------------------- | ---------------------------------------------- |
| 14  | ConfigFile → ResolveDetectors → Pipeline.Run | Pieces exist but never assembled end-to-end    |
| 15  | DetectorRegistry → Build → Pipeline.Run      | Registry tested in isolation only              |
| 17  | Type alias backward compat                   | Verify `pipeline.Detector == finding.Detector` |

### v1.0.0 Release Preparation

- ⬜ Release criteria checklist with concrete pass/fail thresholds
- ⬜ Migration guide for consumers
- ⬜ v1.0.0 tag and GitHub release

### Code Quality Polish

- ⬜ `Category.Compare()` method (matching Severity/Confidence pattern)
- ⬜ `slices.Collect` modernization pass (12 candidates)
- ⬜ LineShiftMap Range/Column shifting (currently only shifts `Position.Line`)
- ⬜ SubstringProvider nearest-position heuristic improvement

---

## d) TOTALLY FUCKED UP 💥 (Honest Assessment)

### Things We Shipped That Were Stupid

1. **FixStrategyResolver** — We shipped an interface whose only method already existed on the type. `DefaultResolver.CanAutoApply()` literally called `FixStrategy.CanAutoApply()`. Zero consumers. Zero tests. It was documented as a feature in AGENTS.md, CHANGELOG.md, and the comprehensive TODO plan. **Fixed this session.**

2. **MiddlewareFunc/ComposeMiddleware** — We shipped a middleware pattern that was never wired into `Config` or `Pipeline.Run`. No `Middlewares` field existed. No test file existed (the plan doc falsely claimed "tested standalone"). The plan doc even had a TODO to "integration test: full pipeline with middleware" — but there was nothing to integrate. **Fixed this session.**

3. **OnStage vs StageHooks split brain** — We shipped TWO notification mechanisms for the same event. Every stage boundary called both `notifyStage()` (legacy OnStage) and `fireStageHook()` (new StageHooks). If both were configured, the "after stage" event fired twice. **Fixed this session.**

4. **AGENTS.md is 399 lines** — This file should be concise, enduring context. Instead it's a running changelog of every session's decisions. It violates its own documented purpose: "NOT for change logs, task lists, feature status." It needs a major trim.

5. **go-structure-linter reports 40 issues** — The external linter reports 38 ERROR + 2 WARNING issues. Most are "root-package-files" (suggesting files be in `/internal/` or `/pkg/`), which is an intentional design decision for this library (root package is the public API). The linter is configured in BuildFlow but these are accepted noise.

6. **Coverage regression** — Root package dropped from 95.7% → 92.3%, pipeline from 92.8% → 90.4%, CLI from 90.7% → 83.5%. New Session 11-14 features were added faster than tests were written for them.

### Pre-existing Issues (Not Fixed This Session)

- `Position{}` zero-value ambiguity (`Offset=0` means both "byte 0" and "unset") — **OWNER DECISION required**
- `FixStrategy ""` vs `FixStrategyNone` — two valid "no fix" states — **OWNER DECISION required**
- `go-finding` and `result` binaries in working directory (not git-tracked, but trigger go-structure-linter warnings)

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process Improvements

1. **Stop shipping ghost systems** — Before adding any new exported API, verify it has at least one consumer wired end-to-end. The "build it and they will come" approach produced two phantom features this session.
2. **Trim AGENTS.md aggressively** — 399 lines is too long. Move session-specific decisions to `CHANGELOG.md` or `docs/architecture-decisions.md`. Keep AGENTS.md under 200 lines of enduring context.
3. **Write integration tests before declaring features done** — The Session 11 features were marked `[x]` in TODO_LIST.md but never assembled end-to-end. Unit tests are necessary but not sufficient.
4. **Run self-reviews proactively** — This session's findings came from a user-triggered self-review. These should happen before commits, not after.

### Architecture Improvements

5. **Resolve Position zero-value semantics** — This is the #1 blocking decision for v1.0.0. It affects every consumer and every type that embeds or references Position.
6. **Add named string types** — `ToolName`, `RuleName`, `FindingID` instead of raw `string`. Prevents mixing up parameters at call sites.
7. **LineShiftMap is incomplete** — Only shifts `Position.Line`. Doesn't shift `Range.Start.Line`, `Range.End.Line`, or `Position.Column`. Post-fix finding positions can be wrong for multi-line ranges.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

Sorted by **impact / effort ratio** (highest first).

### Tier A: High Impact, Low Effort (Do First)

| #   | Task                                                                                     | Impact   | Effort   | Why                                                                                  |
| --- | ---------------------------------------------------------------------------------------- | -------- | -------- | ------------------------------------------------------------------------------------ |
| 1   | **Resolve Position zero-value semantics** (OWNER DECISION)                               | Critical | Decision | Blocks v1.0.0; affects every type. Pick `-1` sentinel or `*int` and commit.          |
| 2   | **Integration test: ConfigFile → ResolveDetectors → Pipeline.Run**                       | High     | 1hr      | Prove the ConfigFile feature actually works end-to-end. Currently zero proof.        |
| 3   | **Integration test: DetectorRegistry → Build → Pipeline.Run**                            | High     | 1hr      | Registry tested in isolation only. Needs end-to-end proof.                           |
| 4   | **v1.0 Migration Guide**                                                                 | High     | 1hr      | Consumers need guidance for FindingsSnapshot, unexported types, deprecated APIs.     |
| 5   | **Fix CLI coverage: 83.5% → 90%+**                                                       | Medium   | 1hr      | Biggest coverage gap. Target new features (fix_provider_registry, generated_filter). |
| 6   | **Fix pipeline coverage: 90.4% → 93%+**                                                  | Medium   | 1hr      | Test ApplyWithShiftMap, groupFindingsBySafePath, recordShiftMap.                     |
| 7   | **Trim AGENTS.md to <200 lines**                                                         | Medium   | 30min    | 399 lines violates its own purpose. Move session logs to CHANGELOG.                  |
| 8   | **Update FEATURES.md**                                                                   | Low      | 30min    | Missing v0.7.0 features. One-file update.                                            |
| 9   | **Godoc examples: IntervalIndex, MergeIter, LineShiftMap, DetectorRegistry, ConfigFile** | Medium   | 30min    | Zero discoverable examples for 5 features.                                           |
| 10  | **Add `Category.Compare()` method**                                                      | Low      | 15min    | Follows established Severity/Confidence pattern.                                     |

### Tier B: Medium Impact, Medium Effort

| #   | Task                                                       | Impact | Effort | Why                                                                 |
| --- | ---------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------- |
| 11  | **Complete FixProviders through CLI config** (TODO #18a-d) | Medium | 2hr    | Users can't specify custom providers without writing Go code.       |
| 12  | **v1.0.0 release criteria checklist**                      | High   | 1hr    | `docs/RELEASE_CRITERIA.md` needs concrete pass/fail thresholds.     |
| 13  | **LineShiftMap: extend to Range + Column shifting**        | Medium | 30min  | Post-fix finding positions can be wrong for multi-line ranges.      |
| 14  | **SubstringProvider nearest-position heuristic**           | Medium | 30min  | First-match `strings.Index` is ambiguous with multiple occurrences. |
| 15  | **Integration test: type alias backward compat**           | Low    | 15min  | Verify pipeline type aliases compile-match root package.            |
| 16  | **Update README.md for v0.7.0+**                           | Medium | 30min  | Missing GoASTProvider, updated stats, new feature sections.         |
| 17  | **Update doc.go with new feature examples**                | Medium | 30min  | Package docs miss several features.                                 |
| 18  | **Audit deprecated APIs for v1.0.0 removal timeline**      | Medium | 1hr    | 5 deprecated APIs need concrete removal dates.                      |
| 19  | **Benchmark regression thresholds in CI**                  | Medium | 30min  | Benchmark job exists but never fails on regression.                 |
| 20  | **Fuzz CategoryForLinter**                                 | Low    | 15min  | Case-insensitive lookup edge cases.                                 |

### Tier C: Lower Priority / Blocked

| #   | Task                                                                     | Impact | Effort   | Why                                                                 |
| --- | ------------------------------------------------------------------------ | ------ | -------- | ------------------------------------------------------------------- |
| 21  | **GoReleaser release with latest tag**                                   | Medium | 30min    | 20 commits ahead of last tag (v0.6.1). Should cut v0.7.0 or v0.8.0. |
| 22  | **slices.Collect modernization pass**                                    | Low    | 30min    | 12 candidates. Code polish, no behavior change.                     |
| 23  | **Add .github/dependabot.yml**                                           | Low    | 15min    | Auto-dependency updates for golang.org/x.                           |
| 24  | **Fix FixStrategy "" vs FixStrategyNone normalization** (OWNER DECISION) | Medium | Decision | Two valid "no fix" states is a type smell.                          |
| 25  | **Decide Report.Findings unexport timing** (OWNER DECISION)              | High   | Decision | Field is deprecated but still public. When to pull the trigger?     |

---

## g) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF 🤔

### **What are the v1.0.0 Position zero-value semantics?**

`Position{}` has a fundamental ambiguity that I cannot resolve without your input because every option is a breaking change with different tradeoffs:

**The problem:** `Position.Offset = 0` means both:

- "byte offset 0" (valid — the first byte of the file)
- "unset" (the zero value of the struct)

This means `Position{}.HasOffset()` returns `true` (0 is set), but `Position{}.IsZero()` also returns `true` (it's the zero value). The struct lies about its own state.

**Three options, all breaking:**

| Option                                 | Pro                                            | Con                                                                                  |
| -------------------------------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------ |
| **A: Use `-1` sentinel**               | Simple, no API change to field types           | Negative offset is a footgun; `uint` can't represent it; needs validation everywhere |
| **B: Use `*int` pointers**             | Makes zero-value truly "unset"; null is honest | Nil dereference risk; API ergonomics worse; breaks JSON serialization                |
| **C: Separate `HasOffset bool` field** | Explicit; no magic values                      | Redundant with the existence of a value; struct grows                                |

This decision cascades to `Range.End` (same ambiguity: `Line == 0` means "unset" but 0 is a valid line in some systems), `FixStrategy ""` vs `FixStrategyNone`, and the v1.0.0 `Report.Findings` unexport.

**I need you to pick a direction.** I recommend **Option A (`-1` sentinel)** because it's the least invasive, but I won't make this call alone — it affects every consumer.

---

## Session Commits

```
3ebbb30 Document OnStage deprecation and split-brain elimination in docs
f687179 Eliminate OnStage/StageHooks split brain: unify notification path
3985edf Fix lint warnings in goast provider and update stale TODO entries
ca958f1 Remove two ghost systems: FixStrategyResolver and MiddlewareFunc
```

---

## Test & Lint Status

```
go build ./...          — PASS
go vet ./...            — PASS
go test -race ./...     — PASS (all packages)
nix run .#lint          — 0 issues
BuildFlow pre-commit    — 34/34 checks PASS
```

---

_Assisted-by: Crush:glm-5.2_
