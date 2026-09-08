# Session 17 — TODO List Completion Status

**Date:** 2026-06-17 00:49 CEST
**Branch:** master
**Commits this session:** pending (this report)
**Since last tag (v0.6.1):** 20 commits + this session's work

---

## Executive Summary

This session executed the entire Top 25 TODO list from `docs/status/2026-06-16_23-31_session16-ghost-system-cleanup-and-split-brain-fix.md`.
All low-hanging items are complete. The three owner-decision blockers for v1.0.0
(Position zero-value semantics, FixStrategy normalization, Report.Findings
unexport timing) are documented in `docs/RELEASE_CRITERIA.md` and require
explicit owner input before the v1.0.0 tag.

All tests pass with the race detector, lint is at zero issues, and both CLI and
pipeline coverage targets were exceeded.

---

## a) FULLY DONE ✅

### This Session's Committed Work

| #       | Task                                                                                 | Files                                                                                                                                              | Status        |
| ------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| 10      | Add `Category.Compare()` method                                                      | `category.go`, `category_test.go`                                                                                                                  | ✅            |
| 15      | Integration test: type alias backward compat                                         | `pipeline/integration_endtoend_test.go`                                                                                                            | ✅            |
| 20      | Fuzz `CategoryForLinter`                                                             | `category_linter_test.go`, `testdata/fuzz/FuzzCategoryForLinter/gosec`                                                                             | ✅            |
| 23      | Add `.github/dependabot.yml`                                                         | `.github/dependabot.yml`                                                                                                                           | ✅            |
| 22      | `slices.Collect` modernization pass                                                  | `correlate.go`, `pipeline/partial.go`, `pipeline/fix_applier.go`, `cmd/go-finding/*.go`                                                            | ✅            |
| 14      | SubstringProvider nearest-position heuristic                                         | `pipeline/fix_provider.go`, `pipeline/fix_engine_test.go`                                                                                          | ✅            |
| 13      | LineShiftMap: extend to Range + Column shifting                                      | `pipeline/line_shift.go`, `pipeline/line_shift_test.go`, `pipeline/pipeline_detect.go`                                                             | ✅            |
| 2       | Integration test: ConfigFile → ResolveDetectors → Pipeline.Run                       | `pipeline/integration_endtoend_test.go`                                                                                                            | ✅            |
| 3       | Integration test: DetectorRegistry → Build → Pipeline.Run                            | `pipeline/integration_endtoend_test.go`                                                                                                            | ✅            |
| 11      | Complete FixProviders through CLI config                                             | `cmd/go-finding/fix_provider_registry.go`, `cmd/go-finding/main.go`, `cmd/go-finding/config.example.yaml`, `cmd/go-finding/coverage_extra_test.go` | ✅            |
| 5       | Fix CLI coverage 83.5% → 90%+                                                        | `cmd/go-finding/coverage_extra_test.go`                                                                                                            | ✅ 91.2%      |
| 6       | Fix pipeline coverage 90.4% → 93%+                                                   | `pipeline/coverage_config_file_test.go`                                                                                                            | ✅ 95.2%      |
| 9       | Godoc examples: IntervalIndex, MergeIter, LineShiftMap, DetectorRegistry, ConfigFile | `example_extra_test.go`, `pipeline/example_test.go`                                                                                                | ✅            |
| 17      | Update `doc.go` with new feature examples                                            | `doc.go`                                                                                                                                           | ✅            |
| 8       | Update `FEATURES.md`                                                                 | `FEATURES.md`                                                                                                                                      | ✅            |
| 16      | Update `README.md` for v0.7.0+                                                       | `README.md`                                                                                                                                        | ✅            |
| 4       | v1.0 Migration Guide                                                                 | `docs/MIGRATION_v1.0.md`                                                                                                                           | ✅            |
| 12      | v1.0.0 release criteria checklist                                                    | `docs/RELEASE_CRITERIA.md`                                                                                                                         | ✅            |
| 18      | Audit deprecated APIs for v1.0.0 removal timeline                                    | `docs/RELEASE_CRITERIA.md`, `AGENTS.md`                                                                                                            | ✅            |
| 19      | Benchmark regression thresholds in CI                                                | `.github/workflows/ci.yml`, `scripts/bench-check.sh`                                                                                               | ✅            |
| 7       | Trim `AGENTS.md` to <200 lines                                                       | `AGENTS.md`                                                                                                                                        | ✅ 115 lines  |
| 1/24/25 | Document owner-decision items in release criteria                                    | `docs/RELEASE_CRITERIA.md` Blocker 1/2/3                                                                                                           | ✅ documented |
| 21      | GoReleaser release prep (document, no tag/push)                                      | `CHANGELOG.md`, version already at 0.7.0                                                                                                           | ✅ ready      |

### Codebase Health Metrics (as of this session)

| Metric                  | Value                                  | Status                   |
| ----------------------- | -------------------------------------- | ------------------------ |
| **Production LOC**      | 10,666 (excl. tests/examples/testdata) | —                        |
| **Test LOC**            | 23,343                                 | 2.19× test-to-code ratio |
| **Test functions**      | 660                                    | Comprehensive            |
| **Benchmarks**          | 48                                     | Comprehensive            |
| **Example functions**   | 35                                     | Comprehensive            |
| **Fuzz targets**        | 23                                     | Comprehensive            |
| **Direct dependencies** | 6                                      | Minimal ✅               |
| **Lint issues**         | 0                                      | ✅                       |
| **Build**               | Passes                                 | ✅                       |
| **Race detector**       | Passes (all packages)                  | ✅                       |

### Test Coverage by Package

| Package                  | Coverage | Trend                                            |
| ------------------------ | -------- | ------------------------------------------------ |
| `analysis/`              | 94.1%    | Stable                                           |
| `internal/detectors/`    | 96.1%    | Stable                                           |
| `internal/gotoken/`      | 92.7%    | Stable                                           |
| Root package (`finding`) | 93.7%    | ↓ from 95.7% v0.6.1 (new examples/code dilution) |
| `pipeline/`              | 95.2%    | ↑ from 90.4%                                     |
| `cmd/go-finding/`        | 91.2%    | ↑ from 83.5%                                     |
| `pipeline/goast/`        | 80.8%    | Stable                                           |

### Core Architecture (All Complete)

- ✅ Unified Finding type — 50+ fields covering all tool outputs
- ✅ Pipeline — detect → process → triage → apply → verify loop
- ✅ Byte-level FixEngine — 452× faster (Session 12 optimization)
- ✅ SARIF round-trip — Export + import with full property bag fidelity
- ✅ LSP conversion — Diagnostic ↔ Finding bidirectional
- ✅ GoASTProvider — AST-aware fix provider for .go files
- ✅ DetectorRegistry — Thread-safe plugin architecture
- ✅ ConfigFile — YAML/JSON config loading
- ✅ IntervalIndex[T] — Generic overlap queries (wired into Correlate)
- ✅ LineShiftMap — Byte-offset-aware line + column + range shift tracking
- ✅ MergeIter() — Streaming iter.Seq merge
- ✅ StageHook — Per-stage before/after hooks with abort capability
- ✅ FindingProcessor — Composable transforms between detect and triage
- ✅ GeneratedFileFilter — Auto-generated file detection (sqlc, protobuf, etc.)
- ✅ TriageFunc — Customizable triage logic
- ✅ ByteLevelConflictDetection — Precise overlap detection
- ✅ Correlate — Cross-tool finding correlation with IntervalIndex

---

## b) PARTIALLY DONE 🟡

### Deprecated APIs (Scheduled for v1.0.0 Removal)

| API                               | Replacement                 | Status                                                                  |
| --------------------------------- | --------------------------- | ----------------------------------------------------------------------- |
| `Report.Findings` (public field)  | `Report.FindingsSnapshot()` | Deprecated, internal migration done; owner decision needed for unexport |
| `Report.Merge()`                  | `Report.MergeInto()`        | Deprecated                                                              |
| `OnStage` callback                | `StageHooks`                | Deprecated this session                                                 |
| `Metrics.RecordFix()`             | `Metrics.RecordFixes(1)`    | Deprecated                                                              |
| `CountBySeverity()` free function | `Report.CountBySeverity()`  | Deprecated                                                              |

### v1.0.0 Release Preparation

- 🟡 Release criteria checklist written, but owner decisions pending on 3 blockers
- 🟡 Migration guide written, but depends on final Position/FixStrategy decisions
- 🟡 GoReleaser ready, but no tag/release cut since v0.6.1

### Documentation Gaps

- 🟡 `USAGE_GUIDE.md` — Needs v0.7.0+ feature additions
- 🟡 `docs/guides/fix-providers.md` — Good, but could add GoASTProvider-specific section

---

## c) NOT STARTED ⬜

### Integration Tests (High Value, Now Partially Done)

| #  | What                                         | Why It Matters       |
| -- | -------------------------------------------- | -------------------- |
| 14 | ConfigFile → ResolveDetectors → Pipeline.Run | ✅ Done this session |
| 15 | DetectorRegistry → Build → Pipeline.Run      | ✅ Done this session |
| 17 | Type alias backward compat                   | ✅ Done this session |

### v1.0.0 Release Preparation

- ⬜ Resolve Position zero-value semantics (owner decision)
- ⬜ Resolve FixStrategy "" vs FixStrategyNone (owner decision)
- ⬜ Unexport Report.Findings (owner decision)
- ⬜ Cut v0.8.0 or v1.0.0 tag and GitHub release

---

## d) TOTALLY FUCKED UP 💥 (Honest Assessment)

### Things We Shipped That Were Stupid

1. **FixStrategyResolver / MiddlewareFunc** — Already fixed in Session 16. Ghost systems shipped with zero consumers.

2. **OnStage vs StageHooks split brain** — Already fixed in Session 16. Two notification mechanisms for the same event.

3. **AGENTS.md was 399 lines** — Fixed this session. It is now 115 lines of enduring context; session changelogs live in CHANGELOG.md and git history.

### Pre-existing Issues (Still Not Fixed — Owner Decisions)

- `Position{}` zero-value ambiguity (`Offset=0` means both "byte 0" and "unset") — **OWNER DECISION required**
- `FixStrategy ""` vs `FixStrategyNone` — two valid "no fix" states — **OWNER DECISION required**
- `Report.Findings` public field — deprecated but still public — **OWNER DECISION required**

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process Improvements

1. **Stop shipping ghost systems** — Before adding any new exported API, verify it has at least one consumer wired end-to-end.
2. **Write integration tests before declaring features done** — Session 11 features had unit tests but no end-to-end assembly; this session fixed that for ConfigFile and DetectorRegistry.
3. **Run self-reviews proactively** — Session 16's findings came from a user-triggered review. These should happen before commits.
4. **Keep AGENTS.md under 150 lines** — The 115-line result is good. Reject attempts to turn it back into a changelog.

### Architecture Improvements

5. **Resolve Position zero-value semantics** — This is the #1 blocking decision for v1.0.0. Pick `-1` sentinel (recommended) or another option and commit.
6. **Normalize FixStrategy "" → "none"** — Two valid "no fix" states is a type smell.
7. **Unexport Report.Findings in v1.0.0** — Internal migration is complete; pull the trigger.
8. **Add more pipeline/goast tests** — Coverage is 80.8%, the lowest package. AST-aware provider deserves deeper edge-case testing.
9. **USAGE_GUIDE.md refresh** — v0.7.0+ features (DetectorRegistry, IntervalIndex, MergeIter, LineShiftMap, StageHook) are not well represented.

---

## f) TOP 25 THINGS TO GET DONE NEXT 🎯

Sorted by **impact / effort ratio** (highest first).

### Tier A: High Impact, Low Effort (Do First)

| #  | Task                                                            | Impact   | Effort   | Why                                                               |
| -- | --------------------------------------------------------------- | -------- | -------- | ----------------------------------------------------------------- |
| 1  | **Resolve Position zero-value semantics** (OWNER DECISION)      | Critical | Decision | Blocks v1.0.0; affects every type. Pick `-1` sentinel and commit. |
| 2  | **Normalize FixStrategy "" → FixStrategyNone** (OWNER DECISION) | High     | Decision | Two "no fix" states is a type smell; trivial once decided.        |
| 3  | **Unexport Report.Findings** (OWNER DECISION)                   | High     | Decision | Migration path done; encapsulates the mutex.                      |
| 4  | **Cut v0.8.0 or v1.0.0 release**                                | High     | 30min    | 20+ commits since v0.6.1; CHANGELOG and GoReleaser ready.         |
| 5  | **Refresh USAGE_GUIDE.md for v0.7.0+**                          | Medium   | 30min    | Missing new features.                                             |
| 6  | **Improve pipeline/goast test coverage 80.8% → 90%+**           | Medium   | 1hr      | Lowest coverage package.                                          |
| 7  | **Run a full brutal self-review before v1.0.0**                 | High     | 1hr      | Catch remaining ghost systems / split brains.                     |
| 8  | **Add more SubstringProvider property tests**                   | Low      | 30min    | Column disambiguation needs randomized coverage.                  |
| 9  | **Fuzz LineShiftMap**                                           | Low      | 30min    | Random edits + positions to catch shift bugs.                     |
| 10 | **Add E2E CLI test for `-fix-provider go-ast`**                 | Medium   | 1hr      | Verify CLI fix provider wiring end-to-end.                        |

### Tier B: Medium Impact, Medium Effort

| #  | Task                                                             | Impact | Effort   | Why                                                                              |
| -- | ---------------------------------------------------------------- | ------ | -------- | -------------------------------------------------------------------------------- |
| 11 | **Add named string types** (`ToolName`, `RuleName`, `FindingID`) | Medium | 2hr      | Prevents parameter mixing at call sites; breaking change.                        |
| 12 | **Add `FindingID` typed constructor + validation**               | Medium | 1hr      | Stronger ID semantics.                                                           |
| 13 | **Correlate spatial-index optimization**                         | Medium | 2hr      | Replace O(k²) worst case with proper index.                                      |
| 14 | **Add JSON schema validation CI check**                          | Medium | 30min    | Ensure schemas stay in sync with code.                                           |
| 15 | **Benchmark baseline update + commit**                           | Low    | 15min    | `benchmarks/baseline.txt` currently absent; needed for CI gate to be meaningful. |
| 16 | **Add `-count=5` stress to pre-commit BuildFlow**                | Medium | 15min    | Catch flaky races earlier.                                                       |
| 17 | **Document fix provider authoring with GoASTProvider example**   | Medium | 30min    | Improve `docs/guides/fix-providers.md`.                                          |
| 18 | **Add property test for MergeIter deduplication**                | Low    | 30min    | Streaming correctness under random inputs.                                       |
| 19 | **Add fuzz target for ConfigFile parsing**                       | Low    | 30min    | Random JSON configs.                                                             |
| 20 | **Re-evaluate `FixStrategyAI`** — keep reserved or remove?       | Low    | Decision | Already decided KEEP; document this once more in ADR.                            |

### Tier C: Lower Priority / Blocked

| #  | Task                                        | Impact | Effort   | Why                                             |
| -- | ------------------------------------------- | ------ | -------- | ----------------------------------------------- |
| 21 | **SARIF 2.2 support evaluation**            | Low    | 2hr      | Not needed now; revisit if consumers demand it. |
| 22 | **GoReleaser cosign signing verification**  | Low    | 1hr      | Nice-to-have for supply chain.                  |
| 23 | **Add .github/FUNDING.yml**                 | Low    | 15min    | GitHub sponsorship metadata.                    |
| 24 | **Add issue templates**                     | Low    | 30min    | Standardize bug reports / feature requests.     |
| 25 | **Evaluate nix flake migration completion** | Low    | Decision | Current flake.nix works; revisit if needed.     |

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

## Session Verification

```
go build ./...          — PASS
go vet ./...            — PASS
go test -race -count=1 ./...     — PASS (all packages)
golangci-lint run ./... — 0 issues
BuildFlow pre-commit    — not run this session
```

---

_Assisted-by: Crush:glm-5.2_
