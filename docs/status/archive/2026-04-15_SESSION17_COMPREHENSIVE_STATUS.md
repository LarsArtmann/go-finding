# Session 17 — Comprehensive Status Report

**Date:** 2026-04-15 (session 17, 11th continuation)
**HEAD:** `6e656d6` on master, pushed to origin
**Status:** Working tree clean, all tests passing (`go test -race ./...`)

---

## A) FULLY DONE

### Sessions 15–17 completed work

| # | Work                                                                                                                 | Commit(s) |
| - | -------------------------------------------------------------------------------------------------------------------- | --------- |
| 1 | Property-based tests with deterministic seed                                                                         | `e2e14f9` |
| 2 | Remove stale `examples/` from `.golangci.yml`                                                                        | `2eb0d4b` |
| 3 | 15 FixApplier error-path unit tests (511 lines)                                                                      | `a0c130f` |
| 4 | CLI error-path tests: `fatalf`, `outputResults` write errors, `outputText` with summary, 4 `run()` integration tests | `9575a0b` |
| 5 | Detector inner-function tests: cancelled context, govet integration, parse edge cases, `staticcheckCategory('A')`    | `6e656d6` |

### Project-wide achievements (all sessions)

- Full pipeline: detect → triage → fix → verify loop
- SARIF 2.1.0 export/import
- LSP Diagnostic conversion
- go/analysis integration (`diagnostic.go`)
- Conflict detection (overlapping fixes)
- Retry with exponential backoff
- Partial success (continue from failed detectors)
- Parallel detection (errgroup)
- Metrics with snapshots
- CLI binary (`cmd/go-finding`)
- Two detectors: govet, staticcheck
- ~4,883 lines production code, ~10,989 lines test code (2.25:1 ratio)
- 267 total commits

---

## B) PARTIALLY DONE

| Item                      | Status           | What remains                                                                                 |
| ------------------------- | ---------------- | -------------------------------------------------------------------------------------------- |
| Deep audit codebase read  | Complete         | Status report (this file), prioritized plan, execution                                       |
| Coverage improvement      | Ongoing          | Several functions below 80% (see Section E)                                                  |
| SARIF round-trip fidelity | Known limitation | `SeverityCritical` → `SeverityError` on import, `RelatedRef.FindingID` and `BeforeCode` lost |

---

## C) NOT STARTED

1. Modernize to Go 1.21+ standard library (`slices.Contains`, `maps.Keys`, `maps.Values`)
2. Investigate `FixApplier` not persisting across pipeline iterations
3. Wire `Correlate()` into Pipeline as optional step
4. Consider `Finding` struct sub-grouping (breaking API change — needs discussion)
5. Config file support for pipeline
6. Watch mode for pipeline
7. go-sarif library evaluation (SARIF export currently hand-rolled)

---

## D) TOTALLY FUCKED UP

None identified. No regressions, no broken features, no data loss.

---

## E) WHAT WE SHOULD IMPROVE

### 1. `applyTriage` coverage at 11.1% — THE most impactful gap

Pipeline tests all use `FixStrategySuggest` findings, so `applyTriage` never receives direct-fix findings to process. This is core pipeline logic with near-zero test coverage.

### 2. SARIF round-trip loss

`findingFromSarResult` at 52.9%. `SeverityCritical` becomes `SeverityError` on import (SARIF has no "critical" level). `RelatedRef.FindingID` is lost — only `Relation` and `Position` survive. `BeforeCode` is lost. Design limitation, but gaps should be documented and test-covered.

### 3. Stale artifacts in repo

- `report/jscpd-report.json` (57KB copy-paste report) — build artifact, should be deleted and gitignored
- `EXECUTION_PLAN_V2.md` at root — superseded planning document
- 3 unarchived status reports in `docs/status/` from sessions 12–14

### 4. `Metrics.StageTiming` value receiver bug (0% coverage)

In `pipeline/metrics.go`, the `StageTiming` method on the `StageTiming` struct uses a value receiver but contains a `sync.Mutex`. A value receiver operates on a copy, so the mutex is meaningless. The `Metrics.StageTiming()` method returns a `func()` that captures the `StageTiming` value — this is almost certainly a bug.

### 5. `intersectionByOffset` and `HasOffset` at 0% coverage

Offset-based Position/Range logic completely untested.

### 6. Standard library modernization

Codebase predates Go 1.21. Should adopt `slices.Contains`, `slices.Delete`, `maps.Keys`, `maps.Values` where appropriate.

### 7. `Severity.LessThan` missing invalid-input test

66.7% coverage — the `!ok` branch for invalid severities is untested.

### 8. `findingFromSarResult` SARIF import paths

52.9% — several SARIF import branches untested: rule metadata, help URI, markdown descriptions, nested related locations.

### 9. `FixApplier` not persisting across pipeline iterations

`applyDirectFixes` in `pipeline.Run()` creates a new `FixApplier` each call. Backups from iteration N don't persist to iteration N+1. If iteration 2 needs to rollback iteration 1's changes, it can't.

### 10. `Correlate()` standalone, not wired into pipeline

Documented as standalone, but it means the correlation feature exists without integration. Could be an optional pipeline stage.

---

## F) Top 25 Things To Do Next

### High impact, low/medium work

| # | Task                                                                                    | Impact   | Work   |
| - | --------------------------------------------------------------------------------------- | -------- | ------ |
| 1 | Add `applyTriage` tests with `FixStrategyDirect` findings                               | Critical | Medium |
| 2 | Fix `Metrics.StageTiming` value receiver bug                                            | High     | Low    |
| 3 | Clean stale files: delete `report/`, archive status docs, handle `EXECUTION_PLAN_V2.md` | Medium   | Low    |
| 4 | Add `intersectionByOffset` + `HasOffset` tests (0% → high)                              | Medium   | Low    |
| 5 | Add `findingFromSarResult` SARIF import path tests (52.9% → higher)                     | Medium   | Medium |
| 6 | Add `Severity.LessThan` invalid-input test                                              | Low      | Low    |
| 7 | Add `RetryConfig.Validate` edge-case tests                                              | Low      | Low    |
| 8 | Add `Verifier.Verify` error-path tests                                                  | Low      | Low    |
| 9 | Add `Range.Contains` edge-case tests (80% → 100%)                                       | Low      | Low    |

### Medium impact, low work

| #  | Task                                                     | Impact | Work |
| -- | -------------------------------------------------------- | ------ | ---- |
| 10 | Modernize: `slices.Contains`, `maps.Keys`, `maps.Values` | Medium | Low  |
| 11 | Add `checkColumnRange` + `hasLineRange` tests            | Low    | Low  |
| 12 | Add `equalTimePtr` both-nil test                         | Low    | Low  |
| 13 | Add `severityToSARIFLevel` edge-case test                | Low    | Low  |
| 14 | Add LSP `toZeroBased` test for 0 Line case               | Low    | Low  |
| 15 | Add `cloneFindings` edge-case test                       | Low    | Low  |
| 16 | Add `Finding.Equal` field-mismatch test                  | Low    | Low  |
| 17 | Document SARIF round-trip losses in code comments        | Low    | Low  |

### Medium impact, medium work

| #  | Task                                                   | Impact | Work   |
| -- | ------------------------------------------------------ | ------ | ------ |
| 18 | Wire `Correlate()` as optional pipeline stage          | Medium | Medium |
| 19 | Investigate `FixApplier` persistence across iterations | Medium | Medium |
| 20 | Add `PrettyJSON` / `LineJSON` error-path tests         | Low    | Medium |
| 21 | Improve `setupProfiling` coverage in CLI               | Low    | Medium |

### Low impact, high work

| #  | Task                                                               | Impact | Work |
| -- | ------------------------------------------------------------------ | ------ | ---- |
| 22 | Refactor `Finding` into embedded sub-structs (breaking API change) | High   | High |
| 23 | Evaluate `go-sarif` library for SARIF generation                   | Medium | High |
| 24 | Add config file support for pipeline                               | Medium | High |
| 25 | Add watch mode for pipeline                                        | Low    | High |

---

## G) Top #1 Question

**Is `Metrics.StageTiming` a bug?**

In `pipeline/metrics.go`, there's a `StageTiming` struct with a `sync.Mutex` field. The `StageTiming()` method on this struct uses a value receiver:

```go
func (s StageTiming) StageTiming() func() { ... }
```

A value receiver copies the struct, including the mutex — which is undefined behavior per Go's sync docs. The `Metrics.StageTiming()` method calls `s.timing.StageTiming()` where `s.timing` is a value (not pointer). This means the returned closure captures a copy of the mutex, making the synchronization ineffective.

**Recommended fix:** Change `StageTiming` to use a pointer receiver, or change `Metrics.timing` from `StageTiming` to `*StageTiming`, or both.

---

_Assisted-by: Crush <crush@charm.land>_
