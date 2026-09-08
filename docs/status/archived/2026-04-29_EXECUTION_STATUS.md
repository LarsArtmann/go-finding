# Execution Status Report

**Date:** 2026-04-29 22:25 CEST
**Project:** go-finding
**Session Goal:** Architectural deepening via improve-codebase-architecture skill

---

## Build Health

| Check                   | Status   |
| ----------------------- | -------- |
| `go test ./...`         | PASS     |
| `just lint`             | 0 issues |
| `go build ./...`        | PASS     |
| Working tree            | Clean    |
| Commits ahead of origin | 3        |

## Coverage by Package

| Package                                                | Coverage | Trend |
| ------------------------------------------------------ | -------- | ----- |
| `github.com/larsartmann/go-finding`                    | 98.5%    | Flat  |
| `github.com/larsartmann/go-finding/pipeline`           | 94.9%    | Flat  |
| `github.com/larsartmann/go-finding/internal/detectors` | 96.1%    | Flat  |
| `github.com/larsartmann/go-finding/cmd/go-finding`     | 78.0%    | Flat  |

### Root-Package Coverage Gaps (sorted by severity)

| Function          | File              | Coverage | Gap                      |
| ----------------- | ----------------- | -------- | ------------------------ |
| `Key`             | `finding.go:160`  | 0.0%     | Newly added, needs tests |
| `PrettyJSON`      | `json.go:22`      | 75.0%    | Error path not tested    |
| `LineJSON`        | `json.go:84`      | 75.0%    | Error path not tested    |
| `ToSARIF`         | `sarif.go:149`    | 75.0%    | Error path not tested    |
| `ToSARIFFiltered` | `sarif.go:160`    | 75.0%    | Error path not tested    |
| `hasLineRange`    | `position.go:145` | 85.7%    | Edge cases               |
| `compareOp`       | `severity.go:58`  | 88.9%    | Invalid severity branch  |

### Pipeline-Package Coverage Gaps (sorted by severity)

| Function                  | File                | Coverage | Gap                     |
| ------------------------- | ------------------- | -------- | ----------------------- |
| `Backup`                  | `file_backup.go:55` | 83.3%    | Error paths             |
| `Restore`                 | `file_backup.go:85` | 84.6%    | Error paths             |
| `Apply`                   | `fix_applier.go:37` | 84.6%    | Rollback paths          |
| `detect`                  | `pipeline.go:330`   | 85.7%    | Branch combinations     |
| `detectParallel`          | `pipeline.go:424`   | 85.7%    | Error propagation       |
| `applyDirectFixes`        | `pipeline.go:529`   | 87.5%    | Metrics path            |
| `partitionFixes`          | `fix_engine.go:31`  | 88.9%    | Skip branch             |
| `detectSequential`        | `pipeline.go:402`   | 88.9%    | Context cancellation    |
| `applyTriage`             | `pipeline.go:488`   | 88.9%    | Dry-run branch          |
| `detectPartialSequential` | `partial.go:57`     | 90.0%    | Context cancellation    |
| `Run`                     | `pipeline.go:177`   | 93.8%    | Correlation branch      |
| `detectPartialParallel`   | `partial.go:81`     | 94.1%    | Context-done after wait |
| `detectConflictsInFile`   | `conflict.go:67`    | 96.0%    | Single-fix group        |

---

## Work Completed This Session

### Completed (7 items)

1. **Created CONTEXT.md** — Domain glossary for architecture reviews
2. **Extracted FileBackup from FixApplier** — Separated backup/restore I/O from fix logic; 100% test coverage
3. **Extracted FixEngine from FixApplier** — Pure in-memory fix application (`[]string, []Finding → []string, int`); fully testable without filesystem
4. **Consolidated detector execution into `runOneDetector`** — Eliminated 4× duplication across sequential/parallel × normal/partial paths; unified metrics recording, suppression filtering, and `OnFinding` notification
5. **Modernized `Finding.Equal` with `slices.Equal`** — Replaced hand-rolled `Related` loop with stdlib helper
6. **Extracted `findingKey` as `Finding.Key()`** — Moved stable-key logic to core type; reusable outside pipeline
7. **Created comprehensive deepening execution plan** — 21 tasks with dependency graph and risk register

### Attempted and Reverted (2 items)

8. **Embedded sub-structs for `Finding`** — Reverted because Go composite literals do not support promoted fields from embedded structs. Changing `Finding{ID: "x"}` to `Finding{Identity: Identity{ID: "x"}}` would break ~100+ call sites across tests and examples. Architectural win is real but breaking surface is too large for a single session.
9. **Auto-sync `Report.Summary` in `AddFinding`** — Reverted because `ComputeSummary()` is O(n) and calling it per `AddFinding` makes bulk addition O(n²). `Merge` calls `AddFinding` in a loop. Correct approach requires incremental summary updates with a `filesSeen` set, which needs a new private field on `Report`.

---

## Categorization of All Known Work

### Fully Done

- FileBackup extraction
- FixEngine extraction
- Detector execution consolidation
- `Finding.Equal` modernization
- `Finding.Key()` extraction
- CONTEXT.md and execution plan
- `FilterConflictingFixes` / `AnalyzeConflicts` tests (already 100% — TODO_LIST.md was stale)
- Retry sentinel errors (already done — TODO_LIST.md was stale)
- `findingKey` extraction
- `Correlate()` wired into Pipeline (already done)

### Partially Done / Design Needed

- **Auto-sync `Report.Summary`** — Needs incremental update strategy to avoid O(n²); blocked on adding `filesSeen` field or accepting full recomputation on `AddFindings` only
- **Embedded `Finding` sub-structs** — Architecturally desirable but breaking; needs v1.0.0 API stability decision

### Not Started — High Impact

- **Error-path tests for JSON (`PrettyJSON`, `LineJSON`)** — Low work, closes 75% → 100% gap
- **Error-path tests for SARIF (`ToSARIF`, `ToSARIFFiltered`)** — Low work, closes 75% → 100% gap
- **Error-path tests for `Verifier.Verify`** — Medium work, verifies detector failure during verification
- **Error-path tests for `FixApplier.Apply`** — Medium work, pushes 84.6% → 90%+
- **Hot-path benchmarks** — Medium work, establishes performance baseline
- **`Finding.Key()` tests** — Very low work, closes 0% → 100% gap

### Not Started — Medium Impact

- **`go-sarif` library evaluation** — Medium work, potential maintenance reduction
- **`hasLineRange` / `checkColumnRange` edge-case tests** — Low work
- **`intersectionByOffset` + `HasOffset` tests** — Low work
- **Range.Contains edge-case tests** — Low work
- **Finding constructor validation** — Low work, improves API safety
- **Modernize stdlib usage** (`slices.Contains`, `maps.Keys`) — Low work
- **Remove stale `//nolint` directives** — Very low work
- **Archive stale planning docs** — Very low work
- **Fix flaky `TestProperty_IDRoundTrip`** — Low work, add seed control

### Not Started — Low Impact

- **Profile memory allocations** — Low work, mostly informational
- **Add `go.work` for local development** — Low work
- **Replace hardcoded `SeverityWarning` in `diagnostic.go`** — Low work
- **Preallocate slices in tests** — Very low work, cosmetic
- **Extract string constants in tests** — Very low work, cosmetic

---

## Top 25 Next Steps (Ranked by Impact / Work)

| #  | Task                                                                 | Work   | Impact | Package                |
| -- | -------------------------------------------------------------------- | ------ | ------ | ---------------------- |
| 1  | Add `Finding.Key()` tests                                            | 5 min  | High   | `finding`              |
| 2  | Add `PrettyJSON` / `LineJSON` error-path tests                       | 15 min | High   | `finding`              |
| 3  | Add `ToSARIF` / `ToSARIFFiltered` error-path tests                   | 15 min | High   | `finding`              |
| 4  | Add `hasLineRange` + `checkColumnRange` edge-case tests              | 15 min | Medium | `finding`              |
| 5  | Add `intersectionByOffset` + `HasOffset` tests                       | 15 min | Medium | `finding`              |
| 6  | Add `Range.Contains` edge-case tests                                 | 15 min | Medium | `finding`              |
| 7  | Add `Verifier.Verify` error-path tests                               | 20 min | Medium | `pipeline`             |
| 8  | Add `FixApplier` error-path tests                                    | 30 min | Medium | `pipeline`             |
| 9  | Add `RetryConfig.Validate` edge-case tests                           | 15 min | Medium | `pipeline`             |
| 10 | Add hot-path benchmarks                                              | 30 min | Medium | `finding` / `pipeline` |
| 11 | Fix flaky `TestProperty_IDRoundTrip`                                 | 15 min | Medium | `finding`              |
| 12 | Add `findingFromSarResult` import-path tests                         | 20 min | Medium | `finding`              |
| 13 | Finding constructor validation                                       | 15 min | Medium | `finding`              |
| 14 | Remove stale `//nolint` directives                                   | 10 min | Low    | `finding` / `pipeline` |
| 15 | Modernize stdlib usage                                               | 20 min | Low    | various                |
| 16 | Archive stale planning docs                                          | 10 min | Low    | `docs/`                |
| 17 | Evaluate `go-sarif` library                                          | 45 min | Medium | `finding`              |
| 18 | Profile memory allocations                                           | 20 min | Low    | various                |
| 19 | Replace hardcoded `SeverityWarning` in `diagnostic.go`               | 15 min | Low    | `finding`              |
| 20 | Add `go.work` for local development                                  | 10 min | Low    | root                   |
| 21 | Preallocate slices in tests                                          | 10 min | Low    | `pipeline`             |
| 22 | Extract string constants in tests                                    | 10 min | Low    | `finding`              |
| 23 | Auto-sync `Report.Summary` (incremental)                             | 30 min | Medium | `finding`              |
| 24 | Add `DeduplicateByPosition` vs `DeduplicateByRule` differential test | 15 min | Medium | `finding`              |
| 25 | Add `cloneFindings` edge-case test                                   | 10 min | Low    | `finding`              |

---

## Blocking Questions

### #1 — How to evolve the `Finding` type without breaking every caller?

**Context:** The `Finding` struct has 20+ fields. Embedded sub-structs would group them logically (`Identity`, `Location`, `Fix`, `Context`) and collapse `Equal()` from ~30 lines to ~10 lines. However, Go composite literals do not support promoted fields — every `Finding{ID:
