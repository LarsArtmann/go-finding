# Comprehensive Status Report — go-finding

**Date:** 2026-04-28 17:14
**Branch:** master (15 commits ahead of origin)
**Working Tree:** Clean
**Test Suite:** ALL PASS (with `-race`)
**go vet:** Clean (0 issues)

---

## Metrics Snapshot

| Metric                      | Value                      |
| --------------------------- | -------------------------- |
| Production Go LOC           | 5,242                      |
| Test Go LOC                 | 11,654                     |
| Test:Code Ratio             | 2.2:1                      |
| Total Coverage              | 94.5%                      |
| Root Package Coverage       | 98.5%                      |
| Pipeline Coverage           | 94.8%                      |
| Detectors Coverage          | 96.1%                      |
| CLI Coverage                | 74.2%                      |
| Lint Issues (golangci-lint) | 36 (0 errors, 36 warnings) |
| Commits Since Origin        | 15                         |
| Files Changed Since Origin  | 32                         |
| Lines Added Since Origin    | +1,598                     |
| Lines Removed Since Origin  | -200                       |
| Net Delta                   | +1,398                     |

### Coverage By Package

| Package          | Coverage | Functions <100%                    |
| ---------------- | -------- | ---------------------------------- |
| Root (`finding`) | 98.5%    | 9 functions below 100% (see below) |
| Pipeline         | 94.8%    | 10 functions below 100%            |
| Detectors        | 96.1%    | 2 functions below 100%             |
| CLI              | 74.2%    | 5 functions below 100%             |

### Functions Below 100% Coverage (Root Package)

| Function                          | Coverage   | Notes                                                   |
| --------------------------------- | ---------- | ------------------------------------------------------- |
| `main()`                          | 0.0%       | CLI entry point — tested via E2E subprocess             |
| `RegisterDetector()`              | 0.0%       | Public API, untested directly                           |
| `run()`                           | 49.0%      | CLI main logic — partially covered by integration tests |
| `setupProfiling()`                | 76.9%      | Hard to test pprof paths                                |
| `outputResults()`                 | 84.6%      | Missing SARIF error path                                |
| `clampConfidence()`               | 80.0%      | Negative input untested                                 |
| `Equal()`                         | 91.7%      | Some field-mismatch branches untested                   |
| `PrettyJSON()` / `LineJSON()`     | 75.0% each | Error paths untested                                    |
| `ToSARIF()` / `ToSARIFFiltered()` | 75.0% each | Marshal error paths untested                            |

---

## A) FULLY DONE

### Critical Bug Fixes (5/5)

| ID  | File                      | Fix                                                             | Status  |
| --- | ------------------------- | --------------------------------------------------------------- | ------- |
| C-1 | `pipeline/pipeline.go`    | OnFix callback fires only for actually-applied fixes            | ✅ DONE |
| C-2 | `pipeline/fix_applier.go` | Insertion-only and deletion-only fixes supported                | ✅ DONE |
| C-3 | `merge.go`                | `Correlate()` hard-limits at `maxCorrelations=10000`            | ✅ DONE |
| C-4 | `json.go`                 | `FilterInvalid` changed from mutable `var` to exported function | ✅ DONE |
| C-5 | `pipeline/retry.go`       | Switched `math/rand` → `math/rand/v2` (`rand.Int64N`)           | ✅ DONE |

### High Bug Fixes (11/11)

| ID   | File                          | Fix                                                                      | Status  |
| ---- | ----------------------------- | ------------------------------------------------------------------------ | ------- |
| H-1  | `position.go`                 | `HasEnd()` checks `End.Line > 0 \|\| End.Offset >= 0`                    | ✅ DONE |
| H-2  | `severity.go`                 | `Compare()` total ordering for invalid severities via string tiebreaker  | ✅ DONE |
| H-3  | `position.go`                 | `Adjacent()` no longer falls back to offset-based when line info present | ✅ DONE |
| H-5  | `merge.go`                    | `DeduplicateByPosition` key includes `ToolName`                          | ✅ DONE |
| H-6  | `internal/detectors/govet.go` | `parsePosn` uses `strconv.Atoi` with error checking                      | ✅ DONE |
| H-7  | `pipeline/fix_applier.go`     | `replaceNearestToLine` finds occurrence closest to finding's line        | ✅ DONE |
| H-8  | `pipeline/fix_applier.go`     | Backup paths include nanosecond timestamp suffix                         | ✅ DONE |
| H-9  | `cmd/go-finding/main.go`      | Removed global `log.SetFlags(0)` `init()` side effect                    | ✅ DONE |
| H-10 | `report.go`                   | `FindByRule` uses `ActiveFindings()` for consistency                     | ✅ DONE |
| H-11 | `cmd/go-finding/main.go`      | `knownDetectorBuilders` protected with `sync.RWMutex`                    | ✅ DONE |

### Medium Bug Fixes (14/14)

| ID        | Fix                                                                                                     | Status  |
| --------- | ------------------------------------------------------------------------------------------------------- | ------- |
| M-4       | SARIF metadata round-trip preserves non-string values via `fmt.Sprintf("%v", v)`                        | ✅ DONE |
| M-6/M-7   | Document `Report.All()` and `FindByID()` yield copies                                                   | ✅ DONE |
| M-8       | `Range.LineCount()` returns absolute span for inverted ranges                                           | ✅ DONE |
| M-9       | `Finding.Equal` uses `floatEq` with 1e-9 epsilon                                                        | ✅ DONE |
| M-10      | Renamed `FinalFindingCount` → `TotalDetected` with clear docs                                           | ✅ DONE |
| M-12      | `Correlation` JSON tags changed to camelCase                                                            | ✅ DONE |
| M-13      | `ToSARIFFiltered` godoc documents dual filtering                                                        | ✅ DONE |
| M-14      | `NewReport` auto-calls `ComputeSummary()`                                                               | ✅ DONE |
| M-15      | SARIF export includes suggestion-only fixes as descriptions                                             | ✅ DONE |
| M-17      | `maxIterations: 0` defaults to pipeline default (5)                                                     | ✅ DONE |
| M-18      | `BySeverityAtLeast` godoc notes invalid exclusion                                                       | ✅ DONE |
| M-19/M-20 | `clampConfidence()` helper; `NormalizedConfidence()` and `Builder.WithConfidence()` clamp to [0.0, 1.0] | ✅ DONE |

### Features Added

| Feature                                                                                        | Status  |
| ---------------------------------------------------------------------------------------------- | ------- |
| `finding.Builder` fluent API with 13 chainable methods                                         | ✅ DONE |
| `version.go` with semver constants (`VersionMajor`, `VersionMinor`, `VersionPatch`, `Version`) | ✅ DONE |
| `Config.CorrelateFindings` — wire `Correlate()` into Pipeline as optional stage                | ✅ DONE |
| `PipelineResult.Correlations` field                                                            | ✅ DONE |
| `RegisterDetector()` — thread-safe detector registration                                       | ✅ DONE |
| `Report.AddFinding/AddFindings` thread-safe via `*sync.Mutex`                                  | ✅ DONE |

### Safety & Lint Fixes

| Fix                                                                        | Status  |
| -------------------------------------------------------------------------- | ------- |
| `Report` copylocks: `sync.Mutex` → `*sync.Mutex` (nil-safe for zero-value) | ✅ DONE |
| `err113`: Added `errDetectorRegistered` sentinel in CLI                    | ✅ DONE |
| `exhaustruct`: Added nolint for suggestion-only `SarifFix`                 | ✅ DONE |
| `Metrics.TotalDuration` guards against `startTime.IsZero()`                | ✅ DONE |
| `Pos()` improved godoc                                                     | ✅ DONE |

### Test Improvements

| Test                                                                              | Status  |
| --------------------------------------------------------------------------------- | ------- |
| `Finding.IsValid()` tests                                                         | ✅ DONE |
| `Suppression.IsValid()` tests                                                     | ✅ DONE |
| `ErrorCategory.IsValid()` tests                                                   | ✅ DONE |
| `Severity.LessThan` invalid input test                                            | ✅ DONE |
| `equalTimePtr` both-nil test                                                      | ✅ DONE |
| CLI end-to-end tests (3: DefaultDetectors, ConfigFile, SARIFOutput)               | ✅ DONE |
| Pipeline correlation tests (enabled + disabled)                                   | ✅ DONE |
| Bug-specific test files (`pipeline_bugfix_test.go`, `fix_applier_bugfix_test.go`) | ✅ DONE |

### Documentation

| Doc                                                                                     | Status  |
| --------------------------------------------------------------------------------------- | ------- |
| `CHANGELOG.md` — comprehensive `[Unreleased]` section with all changes                  | ✅ DONE |
| `README.md` — Builder API section, updated coverage stats, fixed field reference        | ✅ DONE |
| `TODO_LIST.md` — cleaned from 159 → 119 items, organized by priority, completed section | ✅ DONE |
| `Correlate()` godoc updated to document Pipeline integration                            | ✅ DONE |

### go-structure-linter (Cross-Repo)

| Task                                                                                                                                       | Status  |
| ------------------------------------------------------------------------------------------------------------------------------------------ | ------- |
| Dead code removal: `file_cache.go` (215 lines), `health_service.go` (117 lines), `health_service_test.go`                                  | ✅ DONE |
| Removed references from `interfaces.go`, `container.go`, `types.go`, `container_test.go`, `severity_test.go`, `testutil/ginkgo_helpers.go` | ✅ DONE |

---

## B) PARTIALLY DONE

| Item                      | Status             | What's Left                                                                                                                                                                             |
| ------------------------- | ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Lint cleanup              | 36 warnings remain | testifylint (10), paralleltest (6), gosec (7), golines (3), gochecknoglobals (2), gci (1), gocognit (1), goconst (1), gofumpt (1), dupword (1), noctx (1), predeclared (1), thelper (1) |
| CLI coverage              | 74.2%              | `main()` at 0%, `RegisterDetector()` at 0%, `run()` at 49% — need more integration tests                                                                                                |
| `Finding.Equal` coverage  | 91.7%              | Field-mismatch branches untested                                                                                                                                                        |
| SARIF error-path coverage | 75%                | `ToSARIF`, `ToSARIFFiltered`, `PrettyJSON`, `LineJSON` marshal error paths untested                                                                                                     |

---

## C) NOT STARTED

| Item                                                                              | Priority        | Effort   | Impact                                   |
| --------------------------------------------------------------------------------- | --------------- | -------- | ---------------------------------------- |
| `Finding` struct sub-grouping into embedded sub-structs                           | P3 (breaking)   | Large    | Medium — cleaner API but breaking change |
| `go-finding` adapter in `go-structure-linter`                                     | P4 (cross-repo) | Large    | High — enables migration                 |
| `go:generate stringer` for Severity, FixStrategy, Category, SuppressionKind       | P2              | Small    | Low — nice-to-have                       |
| SARIF parser fuzz test                                                            | P1              | Medium   | High — handles untrusted input           |
| SARIF schema validation test                                                      | P2              | Small    | Medium                                   |
| Replace hardcoded temp dir in pipeline with `os.MkdirTemp`                        | P2              | Small    | Medium — concurrency safety              |
| Convert 4 `errors.New()` in `pipeline/retry.go` to sentinels                      | P2              | Small    | Low — consistency                        |
| Extract `findingKey` to shared utility (duplicated in `verify.go` and `merge.go`) | P2              | Small    | Low — DRY                                |
| Remove unused `//nolint` directives (3 locations)                                 | P2              | Tiny     | Low — cleanup                            |
| Modernize to Go 1.21+ stdlib (`slices.Contains`, `maps.Keys`)                     | P2              | Medium   | Low — idiom                              |
| Add `go.work` for local development                                               | P3              | Tiny     | Low — DX                                 |
| Add `examples/` directory with standalone examples                                | P2              | Medium   | Medium — adoptability                    |
| Add godoc examples for key APIs                                                   | P2              | Medium   | Medium — pkg.go.dev                      |
| Pipeline example with config file                                                 | P2              | Small    | Medium — adoptability                    |
| Add `Range.Contains` edge-case tests                                              | P2              | Small    | Low — already 85%+                       |
| Add `checkColumnRange` + `hasLineRange` tests                                     | P2              | Small    | Low — edge cases                         |
| Add `cloneFindings` edge-case test                                                | P2              | Tiny     | Low                                      |
| Add `Finding.Equal` field-mismatch test                                           | P2              | Small    | Low — coverage                           |
| Replace hardcoded `SeverityWarning` in `diagnostic.go`                            | P3              | Small    | Low                                      |
| Add `govulncheck` to CI                                                           | P3              | Small    | Medium — supply chain                    |
| Add GitHub release workflow                                                       | P3              | Medium   | Medium — distribution                    |
| Evaluate `go-sarif` library vs hand-rolled SARIF                                  | P3              | Medium   | High — spec compliance                   |
| Fix `TestProperty_IDRoundTrip` flakiness                                          | P1              | Small    | Medium — reliability                     |
| Fix `applyTriage` tests with `FixStrategyDirect` findings                         | P1              | Small    | Low — coverage                           |
| Add `FixApplier` error-path tests                                                 | P1              | Medium   | Medium — coverage                        |
| Add `FilterConflictingFixes` + `AnalyzeConflicts` tests                           | P1              | Small    | Low — coverage                           |
| Add `RetryConfig.Validate` edge-case tests                                        | P1              | Small    | Low                                      |
| Add `Verifier.Verify` error-path tests                                            | P1              | Small    | Low                                      |
| Add `PrettyJSON` / `LineJSON` error-path tests                                    | P1              | Small    | Low                                      |
| Add `findingFromSarResult` import path tests                                      | P2              | Medium   | Medium — SARIF fidelity                  |
| Add SARIF fuzz test                                                               | P1              | Medium   | High — security                          |
| Delete stale coverage files from repo root                                        | P2              | Tiny     | Low                                      |
| Add `govet` binary to `.gitignore`                                                | P2              | Tiny     | Low                                      |
| Document SARIF round-trip losses                                                  | P2              | Small    | Medium                                   |
| Fix `pipeline/partial.go` missing metrics recording                               | P1              | Small    | Medium                                   |
| Add `LSP toZeroBased` test for 0 Line case                                        | P2              | Tiny     | Low                                      |
| Replace loop with `slices.Contains` at `merge_test.go:208`                        | P2              | Tiny     | Low                                      |
| Preallocate `all` slice in `pipeline_test.go:332`                                 | P2              | Tiny     | Low                                      |
| Extract `"changed"` string to constant                                            | P2              | Tiny     | Low                                      |
| Add `intersectionByOffset` + `HasOffset` tests                                    | P1              | Small    | Low                                      |
| Add `severityToSARIFLevel` edge-case test                                         | P2              | Small    | Low                                      |
| Add benchmarks for hot paths                                                      | P2              | Medium   | Medium                                   |
| Profile memory allocation hotspots                                                | P3              | Medium   | Low                                      |
| Add `io.WriterTo` for SARIF output                                                | P3              | Small    | Low                                      |
| Decide on `FixStrategyAI` — implement or remove                                   | P3              | Decision | Medium                                   |
| Decide on repository name                                                         | P3              | Decision | Low                                      |
| Per-package coverage thresholds in CI                                             | P3              | Small    | Medium                                   |
| Nix migration (Phases 0-5)                                                        | P4              | Large    | Medium                                   |
| Web UI prototype for pipeline monitoring                                          | P4              | Large    | Low                                      |
| IDE plugin stubs — VS Code                                                        | P4              | Large    | Medium                                   |
| Watch mode with fsnotify                                                          | P4              | Large    | Medium                                   |
| Config file support for library/pipeline                                          | P3              | Medium   | Medium                                   |
| Create real-world tool integration guide                                          | P3              | Medium   | High                                     |

---

## D) TOTALLY FUCKED UP (Problems Found)

### 1. SARIF Critical Round-Trip Is Lossy

`SeverityCritical` → SARIF `"error"` → `SeverityError`. The `go-finding/severity` property preserves it, but only if `FindingsFromSARIF` reads the properties. Any downstream consumer that only reads the SARIF level will lose the distinction. **This is a known spec limitation**, not a bug, but it should be documented prominently.

### 2. `FixStrategyAI` Is a Phantom

Defined as a constant but never wired to any implementation. `NeedsAI()` returns true for it, but nothing handles the AI case. This misleads consumers into thinking AI fix support exists. Should either be removed or clearly documented as placeholder.

### 3. `applyToFile` Cognitive Complexity 42

`pipeline/fix_applier.go:177` has cognitive complexity 42 (threshold: 35). This is the line-based fix application logic with multiple fallback paths. Needs decomposition into smaller functions.

### 4. CLI `main()` at 0% Direct Coverage

`main()` is untestable in unit tests (it calls `os.Exit`). The E2E tests cover it as a subprocess but don't contribute to the coverage profile. This is a structural limitation of Go CLI testing.

### 5. `Finding` Struct Is a God Object

The `Finding` struct has 18 fields mixing identity, location, content, fix metadata, suppression, and correlation data. This makes the API overwhelming for new users. Grouping into embedded sub-structs (`Location`, `Fix`, `Meta`) would improve usability but is a breaking change.

### 6. No SARIF Schema Validation

`FindingsFromSARIF` parses arbitrary JSON without validating against the SARIF 2.1.0 schema. Malformed input is silently accepted, potentially producing nonsensical Findings. This is a security concern for tools that consume SARIF from untrusted sources.

### 7. Global Mutable State in `knownDetectorBuilders`

Even though we added `sync.RWMutex` protection, the global map is still mutable via `RegisterDetector()`. Multiple tests calling `RegisterDetector()` could interfere with each other. The mutex prevents data races but not logical interference.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **`Finding` struct decomposition** — Group 18 fields into 4-5 embedded sub-structs. Breaking but worth it for v0.2.0.
2. **SARIF compliance** — Evaluate `go-sarif` library or add schema validation. Hand-rolled SARIF is functional but may drift from spec.
3. **Pipeline immutability** — `Pipeline.Run()` mutates internal state. Should create fresh state per invocation or document the limitation clearly.
4. **Error wrapping audit** — Ensure all internal errors use `%w` for unwrapping. Some `pipeline/retry.go` errors still use dynamic `errors.New()`.

### Testing

5. **SARIF fuzz test** — `FindingsFromSARIF` handles untrusted input; needs fuzz coverage.
6. **Fix `TestProperty_IDRoundTrip` flakiness** — Fails ~5% with random Unicode; add seed control.
7. **CLI coverage to 80%+** — Add more integration tests for `run()` paths, profiling, outputResults.
8. **`FixApplier` error-path tests** — Push from 84% → 90%+ with backup failure, restore failure, concurrent access tests.
9. **`FilterConflictingFixes` + `AnalyzeConflicts` tests** — Currently 100% function coverage but driven only through pipeline tests. Add dedicated unit tests.

### Lint & Hygiene

10. **Reduce 36 lint warnings to <10** — Most are low-hanging: testifylint `require-error` (10), paralleltest (6), thelper (1).
11. **Decompose `applyToFile`** — Cognitive complexity 42 → under 35 via helper extraction.
12. **Add `govulncheck` to CI** — Supply chain security.
13. **Remove unused `//nolint` directives** — 3 stale directives from old code.

### Documentation

14. **Add `examples/` directory** — Standalone runnable examples for common use cases.
15. **Add godoc examples** — `ExampleNewFinding`, `ExampleBuilder`, `ExampleFilter`, `ExamplePipeline`.
16. **Document SARIF round-trip losses** — Critical→Error, RelatedRef.FindingID lost, BeforeCode lost.

### Cross-Repo

17. **`go-finding` adapter in `go-structure-linter`** — Map 60 rules to Finding types, design migration path.
18. **Decide on `FixStrategyAI`** — Remove the phantom or build a plugin interface.

---

## F) TOP 25 THINGS WE SHOULD GET DONE NEXT

| #   | Task                                                                            | Priority | Effort | Impact | Category      |
| --- | ------------------------------------------------------------------------------- | -------- | ------ | ------ | ------------- |
| 1   | Fix `TestProperty_IDRoundTrip` flakiness (seed control)                         | P1       | S      | M      | Testing       |
| 2   | Add SARIF parser fuzz test (`FindingsFromSARIF`)                                | P1       | M      | H      | Security      |
| 3   | Fix `pipeline/partial.go` missing metrics recording during partial detection    | P1       | S      | M      | Bug           |
| 4   | Add `FixApplier` error-path unit tests (backup/restore failure)                 | P1       | M      | M      | Testing       |
| 5   | Fix 10 testifylint `require-error` warnings                                     | P2       | S      | L      | Lint          |
| 6   | Fix 6 `paralleltest` warnings (add `t.Parallel()`)                              | P2       | S      | L      | Lint          |
| 7   | Decompose `applyToFile` (cognitive complexity 42 → <35)                         | P2       | M      | M      | Code quality  |
| 8   | Add `examples/` directory with 3-5 standalone examples                          | P2       | M      | M      | Docs          |
| 9   | Add godoc examples for key APIs (`ExampleBuilder`, `ExampleFilter`)             | P2       | M      | M      | Docs          |
| 10  | Fix 7 gosec warnings (G204 subprocess, G703 path traversal nolints)             | P2       | S      | M      | Lint/Security |
| 11  | Add `findingFromSarResult` import path tests (rule metadata, help URI)          | P2       | M      | M      | Testing       |
| 12  | Add `PrettyJSON` / `LineJSON` / `ToSARIF` error-path tests                      | P2       | S      | L      | Testing       |
| 13  | Document SARIF round-trip losses prominently                                    | P2       | S      | M      | Docs          |
| 14  | Replace hardcoded temp dir in `pipeline/pipeline.go` with `os.MkdirTemp`        | P2       | S      | M      | Bug           |
| 15  | Decide on `FixStrategyAI` — document as placeholder or remove                   | P2       | S      | M      | API           |
| 16  | Add `intersectionByOffset` + `HasOffset` tests (0% coverage)                    | P1       | S      | L      | Testing       |
| 17  | Add `FilterConflictingFixes` + `AnalyzeConflicts` dedicated tests               | P1       | S      | L      | Testing       |
| 18  | Evaluate `go-sarif` library vs hand-rolled SARIF for spec compliance            | P2       | M      | H      | Architecture  |
| 19  | Add `govulncheck` step to CI                                                    | P2       | S      | M      | Security      |
| 20  | Add `go:generate stringer` for Severity, FixStrategy, Category                  | P2       | S      | L      | DX            |
| 21  | Add `RegisterDetector()` test                                                   | P2       | S      | L      | Testing       |
| 22  | Fix `Finding.Equal` field-mismatch test (91.7% → 100%)                          | P2       | S      | L      | Testing       |
| 23  | Extract `findingKey` to shared utility (deduplicated from verify.go + merge.go) | P2       | S      | L      | DRY           |
| 24  | Add `clampConfidence()` negative input test (80% → 100%)                        | P2       | S      | L      | Testing       |
| 25  | Add `setupProfiling` error-path test (76.9% → 90%+)                             | P2       | S      | L      | Testing       |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should `Finding` be decomposed into embedded sub-structs (breaking change), or should we stabilize the current API as v0.2.0 and defer decomposition to v1.0?**

The current `Finding` struct has 18 fields, which is a lot. Grouping them into:

- `Location` (File, Line, Column, Range, Snippet)
- `Fix` (FixStrategy, Suggestion, BeforeCode, AfterCode)
- `Meta` (Confidence, Tag, Category, Metadata)
- `Identity` (ID, Rule, ToolName)

...would make the API cleaner and more navigable. But it's a **breaking change** that affects:

- Every `NewFinding()` call site
- Every struct literal `Finding{...}`
- Every field access `f.Position` → `f.Location.Position` (or similar)
- JSON serialization (unless we preserve flat JSON tags via custom marshal)
- The Builder API (already exists, mitigating some pain)

**The question is:** Is the API pain worth the structural improvement now, or should we ship v0.2.0 with the current flat struct and plan decomposition for a future major version?

My recommendation: **Stabilize current API for v0.2.0. Plan decomposition for v1.0.** The Builder API already provides a clean construction interface. Adding embedded sub-structs would break 5 downstream repos (art-dupl, branching-flow, hierarchical-errors, go-auto-upgrade, golangci-lint-auto-configure) that already consume `Finding`.

---

## Commit History Since Origin (15 Commits)

```
afc3eb0 feat(pipeline): wire Correlate into Pipeline as optional stage
ac6dba3 docs: clean up TODO_LIST — mark completed items, reorganize priorities
2408def docs(readme): add Builder API section and update coverage stats
69f18aa docs: update CHANGELOG with all unreleased changes
4ff116f fix(lint): resolve err113 and exhaustruct warnings
16f8976 fix(report): use pointer mutex to avoid copylocks on JSON marshal
8fe1ce9 fix(sdk): add Report mutex, guard TotalDuration, improve Pos godoc, add version constants
ccfa6b5 test(cli): add end-to-end tests for CLI
d50d3ae test(sdk): add tests for uncovered functions
31ed490 fix(sdk): resolve all P0 blocking bugs
109a91a docs(planning): add comprehensive 187-task execution plan
e4bec23 docs: add comprehensive status report for 2026-04-28
7044d29 fix(sdk): resolve H-9, H-11, and M-9 bugs
1bf9022 feat(builder): add fluent Finding builder API
36fa774 fix(sdk): resolve critical, high, and medium severity bugs
```

---

_Report generated by Crush at 2026-04-28 17:14_
