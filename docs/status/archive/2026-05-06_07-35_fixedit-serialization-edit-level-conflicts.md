# Status Report — FixEdit Serialization & Edit-Level Conflict Detection

**Date:** 2026-05-06 07:35 CEST
**Branch:** master
**Commits since last status:** 1 (changes uncommitted)

---

## What Changed

### FixEdit Serialization + Edit-Level Conflict Detection

The byte-level FixEngine redesign (previous session) introduced `FixEdit`, `FixProvider`, and
byte-level edit application. This session closes two gaps identified in the previous status report:

1. **FixEdit was not serializable** — no JSON tags, no SARIF round-tripping. This meant edits
   couldn't be persisted, transmitted, or stored in SARIF property bags.

2. **FixEdit.Overlaps was unwired** — `ConflictDetector` still operated at the `Range.Overlaps()`
   level (line/column-based), missing byte-precision overlap detection that the engine already
   tracked internally via its frontier boundary.

**Solution:** Added JSON marshaling/unmarshaling to `FixEdit`, SARIF property bag round-tripping,
`ApplyWithConflicts` to surface edit-level conflicts, and `FilterConflictingEdits` for byte-precision
conflict filtering.

Also: updated all three documentation files (AGENTS.md, FEATURES.md, TODO_LIST.md) which were stale
after the byte-level redesign.

---

## New Code

| File                          | Lines         | Change                                                                                                  |
| ----------------------------- | ------------- | ------------------------------------------------------------------------------------------------------- |
| `pipeline/fix_edit.go`        | 165 (was 77)  | +88 lines: JSON tags, `MarshalJSON`/`UnmarshalJSON`, `ToSARIFProperties`, `FixEditFromSARIFProperties`  |
| `pipeline/fix_engine.go`      | 143 (was 126) | +17 lines: `ApplyWithConflicts` method, `applyEditsWithConflicts` with conflict tracking                |
| `pipeline/conflict.go`        | 253 (was 224) | +29 lines: `FilterConflictingEdits` function                                                            |
| `pipeline/fix_engine_test.go` | 631 (was 435) | +196 lines: tests for `ApplyWithConflicts`, `FilterConflictingEdits`, JSON round-trip, SARIF properties |

## Modified Code

| File                      | Change                                                                                                                              |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `pipeline/fix_applier.go` | `applyToFile` now uses `ApplyWithConflicts` instead of `Apply`                                                                      |
| `AGENTS.md`               | Added fix_edit.go, fix_provider.go, fix_engine.go to pipeline table; added FixProvider features                                     |
| `FEATURES.md`             | Rewrote section 16.7 for byte-level architecture; added FixEdit/FixProvider to summary matrix; added `FixProviders` to Config table |
| `TODO_LIST.md`            | Added 5 new FixProvider Architecture items; fixed stale references; marked column-shift item as resolved                            |

---

## Verification

| Check                          | Result                           |
| ------------------------------ | -------------------------------- |
| `go build ./...`               | PASS                             |
| `go test -race -count=1 ./...` | PASS — **498 tests**, 0 failures |
| `golangci-lint run ./...`      | PASS — **0 issues**              |
| FixEdit JSON round-trip tests  | ALL PASS (3 tests)               |
| FixEdit SARIF property tests   | ALL PASS (3 tests)               |
| ApplyWithConflicts tests       | ALL PASS (2 tests)               |
| FilterConflictingEdits tests   | ALL PASS (2 tests)               |

---

## Architecture: New APIs

### FixEdit Serialization

```go
// JSON round-trip (Source omitted from output)
edit := FixEdit{Offset: 10, Length: 5, Replacement: []byte("hello")}
data, _ := json.Marshal(edit)
var got FixEdit
json.Unmarshal(data, &got)

// SARIF property bag round-trip
props := edit.ToSARIFProperties()           // map[string]string
reconstructed := FixEditFromSARIFProperties(props)  // *FixEdit
```

### Edit-Level Conflict Detection

```go
// FixEngine.ApplyWithConflicts returns applied, conflicts, result
applied, conflicts, newContent := engine.ApplyWithConflicts(content, fixes)

// FilterConflictingEdits uses engine for byte-precision filtering
safeFixes := FilterConflictingEdits(content, fixes, engine)
```

The `ConflictDetector` still operates at `Range.Overlaps()` level for pre-filtering
(before file content is available). `FilterConflictingEdits` provides byte-precision
filtering when content is available. Both are used in the pipeline:

1. `FilterConflictingFixes` in `applyTriage` (Range-level, no content needed)
2. `FixEngine.applyEditsWithConflicts` in `applyToFile` (byte-level, content available)

---

## Status by Category

### a) FULLY DONE

- [x] FixEdit JSON serialization — `MarshalJSON`/`UnmarshalJSON` with Source exclusion
- [x] FixEdit SARIF property bag round-tripping — `ToSARIFProperties`/`FixEditFromSARIFProperties`
- [x] `FixEngine.ApplyWithConflicts` — surfaces overlapping edit conflicts
- [x] `FilterConflictingEdits` — byte-precision conflict filtering with engine
- [x] FixApplier uses `ApplyWithConflicts` for edit-level conflict surfacing
- [x] AGENTS.md updated with FixEdit/FixProvider/FixEngine table entries
- [x] FEATURES.md section 16.7 rewritten for byte-level architecture
- [x] TODO_LIST.md updated with new items, stale references fixed
- [x] All 498 tests pass, zero lint issues
- [x] FixEdit JSON tags (`json:"offset"`, `json:"length"`, `json:"replacement,omitempty"`, `json:"-"` for Source)
- [x] FixEdit `Validate()` error sentinels (`errNegativeOffset`, `errNegativeLength`)
- [x] FixEdit `Overlaps()` with adjacent and zero-length insertion handling
- [x] FixEdit `IsInsert()`/`IsDelete()`/`EndOffset()` helper methods
- [x] FixProvider interface with OffsetProvider, LineProvider, SubstringProvider
- [x] Custom provider registration via engine, applier, and pipeline config
- [x] TODO_LIST.md: column-shift handling marked as resolved by byte-level redesign
- [x] TODO_LIST.md: stale `FixEngine.partitionFixes()` reference corrected

### b) PARTIALLY DONE

Nothing partially done — all started items are complete.

### c) NOT STARTED

- [ ] Build line-offset index — `lineColToOffset` is O(n) per call; build `[]int` index once per file
- [ ] BDD/Ginkgo tests for FixProvider interface contract
- [ ] Domain-specific FixProvider implementations (Go AST, Rust syn, etc.)
- [ ] Finding type model improvements for byte-offset support (`ByteRange` type)
- [ ] Pipeline concurrency improvements (concurrent pipelines, per-file parallel fix application)
- [ ] CLI flag for custom FixProvider registration
- [ ] Example showing custom FixProvider usage
- [ ] `ContentProvider` interface for lazy content access (avoid passing full file to providers)
- [ ] FixEdit `ConflictsWith` field population in `applyEditsWithConflicts` (currently nil)
- [ ] Wire `FilterConflictingEdits` into `applyTriage` as optional second-pass filtering
- [ ] Benchmark comparing old string-based vs new byte-based engine
- [ ] Profile allocation patterns in `applyEdits` — reduce `[]byte` copies
- [ ] Extract `diagnostic.go` to `finding/analysis` subpackage
- [ ] Split `sarif.go` into 3 files
- [ ] Split `pipeline/pipeline.go`
- [ ] Fix `Pipeline.Run()` mutability
- [ ] Fix `FixApplier` cross-iteration persistence (temp dir leaks)

### d) TOTALLY FUCKED UP

- [ ] **Nothing broken** — all 498 tests pass, zero lint issues
- [ ] **Known risk:** `FilterConflictingEdits` uses `Finding.ID` to match conflicts back to
      the input slice. Findings without IDs (empty string) cannot be distinguished — if two
      findings have the same empty ID, one being a conflict would filter both. In practice,
      all pipeline findings have generated IDs via `GenerateID`, but direct construction without
      `NewFinding` could hit this.
- [ ] **Known risk:** `FixEditFromSARIFProperties` stores replacement as `string(props["go-finding/edit/replacement"])`
      which means non-UTF-8 bytes in replacements would be corrupted. The SARIF property bag is
      `map[string]string`, so this is inherently lossy for binary content. Acceptable for source code.
- [ ] **Known risk:** `applyEditsWithConflicts` doesn't populate `ConflictInfo.ConflictsWith` —
      it only sets the `Reason`. The Range-based `AnalyzeConflicts` provides richer info about
      which finding conflicts with which. The edit-level path could be enhanced later.

### e) WHAT WE SHOULD IMPROVE

1. **`lineColToOffset` should use a line-offset index** — Currently O(n) per call, re-scanning
   from the start for each fix. For files with many fixes, build a `[]int` line-start offset
   index once and use O(1) lookups. This is the single biggest performance win available.

2. **FixEdit.ConflictsWith should be populated** — Currently `ConflictInfo.ConflictsWith` is
   nil in edit-level conflicts. We know which edit was skipped (the conflict) and which edit
   caused the skip (the one at the frontier boundary). Track and populate this for richer
   conflict analysis.

3. **FilterConflictingEdits should handle empty-ID findings** — Use index-based matching
   or `Key()` method instead of `ID` for cases where findings don't have stable IDs.

4. **FixApplier.Close() is never called in pipeline** — `applyDirectFixes` creates a new
   `FixApplier` per call but never `defer close`s it. Temp backup directories leak. This is
   a pre-existing bug but critical to fix.

5. **Pipeline.Run() mutates internal state** — `p.findings` and `p.iterations` are mutated
   during `Run()`. Pipeline should be reusable or documented as single-use. Pre-existing.

6. **Finding type should have explicit ByteRange** — `Position.Offset` exists but detectors
   rarely populate it. The `Range` type has `Start.Offset`/`End.Offset` but no detector fills
   them. Add explicit byte-offset encouragement or a dedicated `ByteRange` type.

7. **BeforeCode/AfterCode should be []byte** — The Finding type uses `string` for these.
   For byte-level accuracy (non-UTF8 content), these should be `[]byte` or have byte-offset
   alternatives. Breaking change — deferred to v2.

8. **Domain-specific provider location decision** — Should Go AST, Rust syn providers live
   inside `pipeline/fix/` or as separate modules? Inside = easy discovery, single dep.
   Outside = minimal dependencies, independent versioning. This is a product decision.

9. **BDD tests for FixProvider** — No ginkgo specs cover the FixProvider contract. Should
   test provider ordering, CanHandle precedence, error handling, and custom provider behavior.

10. **FixEdit benchmark** — No benchmark comparing byte-level vs old string-based approach.
    Add `BenchmarkFixEngine_Apply` to measure improvement and catch regressions.

### f) Top 25 Things to Get Done Next

**Priority 1 — Correctness & Bug Fixes (High Impact, Low Effort)**

1. Fix `FixApplier.Close()` leak in `applyDirectFixes` — defer close the applier
2. Document `Pipeline.Run()` single-use contract in godoc
3. Build line-offset index in `LineProvider` — O(1) lookup instead of O(n) per fix
4. Populate `ConflictInfo.ConflictsWith` in edit-level conflict detection
5. Handle empty-ID findings in `FilterConflictingEdits` — use `Key()` fallback

**Priority 2 — Architecture Decisions (High Impact, Low Effort)**

6. Decide domain-specific provider location (inside vs outside library)
7. Extract `diagnostic.go` to `finding/analysis` subpackage (removes 12MB dep)
8. API stability review — audit all exported symbols for v1.0.0 lock
9. Decide `NewFinding` API pattern — functional options vs builder-only vs current
10. Centralize triage logic — `HasFix()`, `triage()`, `Apply()` filtering overlap

**Priority 3 — Testing & Quality (Medium Impact, Medium Effort)**

11. Add BDD/Ginkgo tests for FixProvider interface contract
12. Add `BenchmarkFixEngine_Apply` comparing byte-level performance
13. Add fuzz tests for `lineColToOffset` edge cases
14. Add integration test with custom FixProvider end-to-end
15. Add property-based test: byte-level edits preserve file invariants

**Priority 4 — Code Quality (Medium Impact, Low Effort)**

16. Convert `ConflictDetector`, `Verifier` to package-level functions (stateless structs)
17. Split `sarif.go` into 3 files (570 lines, above 350 threshold)
18. Split `pipeline/pipeline.go` (611 lines, extract adapters/config)
19. Split `cmd/go-finding/main.go` (457 lines, extract config/output)
20. Make `Report` always thread-safe (`sync.Once` for lazy mutex init)

**Priority 5 — Performance (Medium Impact, Low Effort)**

21. Profile allocation patterns in `applyEdits` — reduce `[]byte` copies
22. Add `ContentProvider` interface for lazy content access in `FixProvider.Edits`
23. Add concurrent fix application across files in pipeline

**Priority 6 — Type Model (High Impact, High Effort)**

24. Add `ByteRange` type to Finding with explicit Offset/Length
25. Consider `BeforeCode []byte` / `AfterCode []byte` on Finding (breaking change)

### g) Top #1 Question I Cannot Figure Out Myself

**Should the domain-specific FixProvider implementations (Go AST, Rust syn, etc.)
live INSIDE this library as `pipeline/fix/` sub-packages, or should they be
SEPARATE modules/repositories that import go-finding as a dependency?**

This was the same question from the previous status report. It remains unanswered
because it's a product/architecture decision with permanent module-structure implications:

- **Inside:** Easier discovery, single dependency, shared internal helpers.
  Risk: core module grows heavy AST parser dependencies.

- **Outside:** go-finding stays minimal-dependency (core types = stdlib only),
  each provider versions independently, clean separation.
  Risk: fragmentation, harder discovery, multiple small modules to maintain.

The decision affects whether we add `go/ast` as a dependency to this module. Currently
the design principle is "core types depend only on stdlib" — `go/ast` IS stdlib, but
`golang.org/x/tools` (needed for `go/analysis` integration) is not. An AST provider
could work with just `go/ast` + `go/token` + `go/parser` (all stdlib). But a Rust
provider would need an external parser. This asymmetry suggests the "outside" pattern
may be more consistent.

---

## Files Changed

```
pipeline/fix_edit.go          | 165 (was 77, +88 lines)
pipeline/fix_engine.go        | 143 (was 126, +17 lines)
pipeline/conflict.go          | 253 (was 224, +29 lines)
pipeline/fix_applier.go       | 156 (1 line changed)
pipeline/fix_engine_test.go   | 631 (was 435, +196 lines)
AGENTS.md                     | +9 lines
FEATURES.md                   | +63 lines
TODO_LIST.md                  | +14 lines
```

**Net change:** +419 lines added, -27 lines removed = +392 lines net.
498 tests pass, 0 lint issues.
