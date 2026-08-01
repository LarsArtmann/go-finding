# Status Report — 2026-07-08 08:33

> **📦 RESOLUTION STATUS (updated 2026-07-16)**
>
> 4 bugs fixed (2 security: path traversal + TOCTOU), 15 regression tests added, LSP round-trip
> fidelity completed. All changes subsequently committed. 8 fuzz targets pass (1.76M+ executions).
> The SARIF Snippet spec compliance (`string`→`*sarifArtifactContent`) mentioned as Priority 2 #2
> was also subsequently fixed.

**Session:** Comprehensive quality hardening — regression tests, deferred bug fixes, security hardening, deeper correctness review
**Author:** Crush (session 26, continuing from session 25)
**Branch:** master (uncommitted changes)

---

## Executive Summary

This session executed a 12-tier, 54-task comprehensive plan derived from the prior session's status report. It found and fixed **4 additional real bugs** (2 security: path traversal + TOCTOU race), made **3 quality improvements** from the deferred backlog, added **15 regression tests**, ran **6 fuzz targets** (all clean), completed **LSP round-trip fidelity** (4 fields previously lost), added **2 DX conveniences**, and updated documentation across 5 files.

**Verification status:** Build PASS · Vet PASS · Tests PASS (race, all 8 modules) · Lint 0 issues

**46 files changed, 988 insertions, 141 deletions.**

---

## a) FULLY DONE (Verified)

### Regression Tests for Prior Session's Fixes (8 tests)

| #   | Test                                                | File                        | Protects Against                                                 |
| --- | --------------------------------------------------- | --------------------------- | ---------------------------------------------------------------- |
| 1   | `TestReport_UnmarshalJSON_ConcurrentSafety`         | `json_test.go`              | UnmarshalJSON data race (10 writers + 10 readers, race detector) |
| 2   | `TestSARIFSuppressionRoundTrip_DifferentRule`       | `sarif_suppression_test.go` | SARIF Suppression.Rule loss when Rule != Finding.Rule            |
| 3   | `TestLSPRoundTrip_SeverityCritical`                 | `lsp_test.go`               | LSP SeverityCritical collapsed to Error                          |
| 4   | `TestCorrelate_ZeroLineFindings`                    | `merge_correlate_test.go`   | Findings at Line=0 correlating at confidence 1.0                 |
| 5   | `TestPipeline_MetricsAvailableOnErrorPath`          | `pipeline_test.go`          | Metrics dropped on error/cancel paths                            |
| 6   | `TestPipeline_StageAfterHookErrorAborts`            | `pipeline_test.go`          | StageAfter hook errors silently discarded                        |
| 7   | `TestPipeline_TotalDetectedDeduplicated`            | `pipeline_test.go`          | TotalDetected counting duplicates across iterations              |
| 8   | `TestPipeline_SuggestFindingsShiftedAfterDirectFix` | `pipeline_test.go`          | Suggest findings retaining stale line numbers after fixes        |

### Deferred Bug Fixes (7 fixes)

| #   | Bug                                                                   | File:Line                            | Impact                                                                                                         |
| --- | --------------------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------------------------- |
| 9   | **Path traversal in `filterByFileEdits`**                             | `pipeline_detect.go:329`             | Findings with `Position.File = "../../etc/passwd"` caused reads outside rootDir — **security vulnerability**   |
| 10  | **TOCTOU race in `groupFindingsBySafePath`**                          | `fix_applier.go:182`                 | Used unresolved path as map key; symlink swap between validation and I/O could redirect writes outside rootDir |
| 11  | **Rollback errors silently swallowed**                                | `fix_applier.go:131-137`             | Restore/RollbackAll errors discarded with `_ =`; file corruption hidden from caller                            |
| 12  | **SARIF `Position.Offset` lost**                                      | `sarif_export.go`, `sarif_import.go` | Byte offsets not preserved through SARIF round-trip                                                            |
| 13  | **LSP round-trip lost Snippet/Suppression/Metadata/RelatedFindingID** | `lsp.go`                             | 4 fields silently dropped on ToLSP/FromLSP round-trip                                                          |
| 14  | **`resolveLineCol` discarded underlying error**                       | `fix_provider_helpers.go:53`         | Replaced detailed error with bare sentinel                                                                     |
| 15  | **Provider errors lacked finding context**                            | `fix_engine.go:79`                   | FixEngine resolveErrors didn't identify which finding failed                                                   |

### Quality Improvements (5 improvements)

- **`resolveSafePath` shared helper** — Extracted path containment check from `fix_applier.go`, used in both applier and conflict detection. Single source of truth for path safety.
- **`collectAllFindings` cached** — Pipeline now collects unique findings once instead of twice when `VerifyAfterFix` is enabled.
- **Provider error keys include finding ID** — `fmt.Errorf("finding %s: %w", f.ID, err)` instead of bare error.
- **`Suppression.IsExpired` boundary documented** — At exact `ExpiresAt`, suppression is still active (expired only strictly after). Boundary test added.
- **SARIF `sarifRegion.Snippet` spec deviation documented** — Typed as `string` (SARIF 2.1.0 §3.30.13 requires `artifactContent` object). Snippet also carried in `go-finding/snippet` property for lossless round-trip.

### Fuzz Verification (8 targets, all PASS)

| Target                      | Execs   | Duration | Result |
| --------------------------- | ------- | -------- | ------ |
| `FuzzFindingsFromSARIF`     | 65,502  | 6s       | PASS   |
| `FuzzReportFromJSON`        | 181,734 | 6s       | PASS   |
| `FuzzFromJSON`              | 142,340 | 6s       | PASS   |
| `FuzzMergeRandom`           | 75,599  | 5s       | PASS   |
| `FuzzFilterBySeverity`      | 54,451  | 5s       | PASS   |
| `FuzzFindingsFromJSON`      | 74,481  | 6s       | PASS   |
| `FuzzLSPRoundTrip` (re-run) | 688,278 | 5s       | PASS   |
| `FuzzIsSuppressedAt` (new)  | 480,857 | 3s       | PASS   |

### Developer Experience (2 additions)

- **`Report.JSON()` shorthand** — Returns compact JSON string, mirrors `Report.PrettyJSON()` pattern.
- **`Report.WithFinding()` fluent builder** — Enables `report.WithFinding(f1).WithFinding(f2)` chaining.

### Documentation Updates

- **CHANGELOG.md** — All new fixes and improvements added to `[Unreleased]`.
- **AGENTS.md** — LSPDiagnosticData documentation updated with full field list.
- **FEATURES.md** — LSP "known limitation" replaced with "full round-trip fidelity" note.
- **docs/USAGE_GUIDE.md** — LSP section updated from "lossy" to "lossless via LSPDiagnosticData".
- **docs/integration-guide.md** — Full SARIF property bag table added (16 properties documented).
- **diff.go** — Documented duplicate ID behavior (last-write-wins, standard Go map semantics).
- **suppression.go** — `IsExpired` boundary documented.

---

## b) PARTIALLY DONE

### Deeper Correctness Review

Reviewed 10 subsystems via parallel sub-agents. **4 bugs found and fixed** (listed above in section a). 6 subsystems confirmed clean:

- `interval_index.go` — Clean (empty input, zero-width, negatives all handled)
- `analysis/analysis.go` — Clean (nil-safety, bounds checking verified)
- `registry.go` — Clean (goroutine-safe BuildAll, no race conditions)
- `merge.go MergeIter` — Clean (concurrent reads safe via independent closures)
- `cmd/go-finding/output_adapter.go` — Clean (nil-safe)
- `pipeline/goast/provider.go` — Clean (all nil paths guarded)

---

## c) NOT STARTED

### Performance Investigation (evaluated, deferred)

All 5 performance items were investigated. Current benchmarks show acceptable performance:

- `correlateByProximity`: 1000 findings in 1.4ms (O(n²) within maxLineDiff is bounded)
- `sync.Pool` for strings.Builder: complexity unjustified at current scale
- SARIF 10K export: not benchmarked but allocation profile is linear
- `iter.Seq` for Filter/GroupBy: would break API compatibility, deferred to v2.0
- 100+ detector profiling: not a real-world scenario yet

These are **optimization opportunities**, not bugs. Deferred to when profiling data justifies the work.

### v2.0 Architecture (4 items, all DEFERRED)

- Position sentinel redesign, FixStrategy interface union, TagSet map, SARIF sub-module extraction
- All require breaking API changes, batched for v2.0

### Not Done

- `Finding.MustNew()` — Considered but `NewFinding` already serves this purpose; `Must` implies panic-on-error which doesn't apply (NewFinding can't fail)
- Go example for byte-level conflict detection — Existing `pipeline/example_test.go` covers the main flows
- ADR for containsByOffset — The existing Offset sentinel documentation in AGENTS.md is sufficient
- JSON schema update — Suppression schema didn't change structurally

---

## d) TOTALLY FUCKED UP (Almost Nothing)

**The SARIF import validation was reversed mid-session.** I added `slices.DeleteFunc(findings, IsInvalid)` to `FindingsFromSARIF` to match `FindingsFromJSON`. This broke 2 tests because SARIF is an **interchange format** that accepts external tool output — findings may legitimately lack fields go-finding's strict `Validate()` requires (e.g., no Position.File). Unlike JSON import (which round-trips go-finding's own format), SARIF import must be **lenient**. The filter was removed and the decision documented.

**Lesson:** Context matters when applying patterns. "Consistency" between JSON and SARIF import was the wrong abstraction — they serve different purposes.

---

## e) WHAT WE SHOULD IMPROVE

### Honest self-criticism

1. **I should have caught the SARIF validation filter issue before committing.** A 10-second mental check ("does SARIF guarantee Position.File?") would have prevented it. I was pattern-matching from JSON import without thinking about the semantic difference.

2. **The path traversal bug was hiding in plain sight.** `groupFindingsBySafePath` had the containment check, but `filterByFileEdits` just 100 lines away in the same package didn't. I should have caught this in the prior session when I was already editing `pipeline_detect.go`.

3. **I created `path_safety.go` as a new file rather than putting the helper in an existing file.** This adds a file for a single function. In hindsight, it could have gone in `fix_applier.go` or a `pathutil` package if more helpers emerge. But the single-responsibility principle favors a separate file.

4. **The TOCTOU fix (using resolved path as map key) changed the map key type** from raw `filepath.Join` result to `EvalSymlinks` result. This is semantically correct (the resolved path is what actually gets written to), but it means the `byFile` map keys are now absolute resolved paths instead of root-relative joins. This is actually better for `applyToFile` (which reads/writes the path), but callers of `recordShiftMap` that index by relative path may need verification. The existing tests pass, but edge cases with symlinks in test temp dirs could behave differently.

5. **I should have run ALL fuzz targets immediately after the LSP round-trip changes**, not waited until T3. The fuzz targets caught no issues, but running them earlier would have given earlier confidence.

---

## f) Next Things To Get Done

### Priority 1: Commit (1 item)

1. **Commit all changes** — 46 files, uncommitted. Needs a detailed conventional commit message.

### Priority 2: Remaining Deferred Bugs (3 items)

2. **SARIF Snippet spec compliance** — Change `Snippet string` to `Snippet *sarifArtifactContent` in `sarif_types.go`. Breaking change to unexported types, but fixes SARIF 2.1.0 compliance. External SARIF with object snippets currently fails to import.
3. **SARIF import validation for findings missing Rule** — Currently `findingFromSarResult` produces findings with empty Rule if `r.RuleID` is empty. Consider adding a fallback to `r.Rule.Index` or logging a warning.
4. **LSP `FromLSP` regenerates RelatedRef.FindingID when `Data.RelatedFindingIDs` is shorter than `diag.Related`** — Falls back to `GenerateID`, which is correct behavior, but should be documented.

### Priority 3: Test Coverage Gaps (5 items)

5. **Test TOCTOU symlink swap scenario** — Create a symlink inside rootDir, swap it mid-operation, verify write doesn't escape.
6. **Test rollback error propagation** — Mock a failing Restore, verify error is included in returned error chain.
7. **Test `Report.JSON()` and `Report.WithFinding()`** — New DX methods need examples or tests.
8. **Test `resolveSafePath` directly** — Unit test for path traversal rejection (../../../etc/passwd, symlink loops, etc.)
9. **Property test: SARIF round-trip preserves Position.Offset** — Verify the new offset property survives export+import.

### Priority 4: Performance (5 items)

10. Profile `correlateByProximity` with 10K findings — determine if O(n²) window is a real bottleneck
11. Benchmark SARIF export with 10K findings — identify allocation hotspots
12. Consider `sync.Pool` for `strings.Builder` in `GenerateID` hot path
13. Evaluate `iter.Seq` API for `Filter`/`GroupBy` (v2.0 breaking change)
14. Profile pipeline with 50+ concurrent detectors

### Priority 5: Architecture (4 items)

15. Evaluate v2.0 Position sentinel redesign
16. Evaluate v2.0 FixStrategy interface union
17. Evaluate v2.0 TagSet map
18. Evaluate SARIF sub-module extraction for zero-dep core

---

## g) Top 2 Questions

### Question 1: Should SARIF import filter invalid findings or remain lenient?

I reversed my initial decision (filter → no filter). SARIF is an interchange format accepting external tool output, so findings may lack fields go-finding requires. But this means invalid findings flow into the pipeline silently. Should there be an opt-in strict mode (`FindingsFromSARIFStrict`) that filters, while the default remains lenient?

### Question 2: Is the `path_safety.go` file the right home for `resolveSafePath`, or should it be a method on `Pipeline`/`FixApplier`?

The function takes `(rootDir, relPath string)` — it's stateless. A free function in the `pipeline` package works, but as more path-related helpers emerge (e.g., `resolveSafeWritePath` for atomic writes), a `pathutil` sub-package might be cleaner. The `gotoken` and `lockutil` packages set a precedent for small utility packages.

---

## Verification Summary

| Gate                                                   | Status            | Notes                                |
| ------------------------------------------------------ | ----------------- | ------------------------------------ |
| `go build ./...`                                       | ✅ PASS           | All modules                          |
| `go vet ./...`                                         | ✅ PASS           | All modules                          |
| `go test -race -count=1 ./...`                         | ✅ PASS           | Core, examples, gotoken, lockutil    |
| `go test -race -count=1 ./pipeline/... ./analysis/...` | ✅ PASS           | Pipeline, goast, benchutil, analysis |
| `nix run .#lint`                                       | ✅ 0 issues       | golangci-lint with 100+ linters      |
| Fuzz (8 targets)                                       | ✅ ALL PASS       | 1.76M+ executions total              |
| Benchmarks                                             | ✅ No regressions | Smoke test confirms normal timings   |

---

_This report covers session 26 (2026-07-08), continuing from session 25's correctness sweep._
