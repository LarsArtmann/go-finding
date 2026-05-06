# Status Report — Byte-Level FixEngine Redesign

**Date:** 2026-05-06 05:32 CEST
**Branch:** master
**Commits since last status:** 0 (changes uncommitted)

---

## What Changed

### Core Redesign: String-Based → Byte-Based FixEngine with Plugin Architecture

The FixEngine was completely rewritten from a fragile `[]string`/substring-matching
approach to a byte-level `[]byte` edit engine with a composable FixProvider interface.

**Problem:** The old engine used `strings.Split`/`strings.Join` roundtrips and
`strings.Replace` for applying fixes — fragile because:
- Line indices shift after each edit, requiring careful descending-order application
- Substring matching is ambiguous when the same text appears multiple times
- No extension point for domain-specific (AST-aware) providers
- Deterministic ordering of multiple fixes per file was heuristic-based

**Solution:** Byte-level edits (`FixEdit{Offset, Length, Replacement}`) applied
in descending offset order with a frontier boundary to prevent overlaps.

---

## New Files

| File | Lines | Purpose |
|------|-------|---------|
| `pipeline/fix_edit.go` | 77 | `FixEdit` type — byte-level edit operations with `Overlaps`, `Validate`, `IsInsert`, `IsDelete` |
| `pipeline/fix_provider.go` | 322 | `FixProvider` interface + 3 default providers (OffsetProvider, LineProvider, SubstringProvider) |

## Rewritten Files

| File | Lines | Change |
|------|-------|--------|
| `pipeline/fix_engine.go` | 126 | Engine works on `[]byte`, delegates to providers, sorts edits descending by offset |
| `pipeline/fix_applier.go` | 156 | Removed `strings.Split`/`Join`, reads/writes `[]byte` directly. Added `NewFixApplierWithProviders` |
| `pipeline/pipeline.go` | 620 | Added `Config.FixProviders []FixProvider`, wires providers in `applyDirectFixes` |
| `pipeline/fix_engine_test.go` | 435 | Complete rewrite for byte-based API with FixEdit, FixProvider, and helper tests |

## Unchanged Files

All other files untouched: `conflict.go`, `file_backup.go`, `metrics.go`,
`partial.go`, `retry.go`, `verify.go`, `result.go`, `bdd_test.go`,
all pipeline tests except `fix_engine_test.go`, root package, CLI, detectors, examples.

---

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test -race -count=1 ./...` | PASS — **490 tests**, 0 failures |
| `golangci-lint run ./...` | PASS — **0 issues** |
| FixApplier tests (26 tests) | ALL PASS |
| FixEngine tests (41 tests) | ALL PASS |

---

## Architecture: FixProvider Interface

```go
type FixProvider interface {
    Name() string
    CanHandle(f finding.Finding) bool
    Edits(content []byte, f finding.Finding) ([]FixEdit, error)
}
```

**Default provider chain (tried in order):**

1. **OffsetProvider** — findings with `Range.Start.Offset >= 0 && Range.End.Offset >= 0`
   (most accurate — bypasses line/column entirely)
2. **LineProvider** — findings with `Position.Line > 0`
   (converts line/column → byte offset, supports range, insertion, replacement)
3. **SubstringProvider** — fallback for any finding with `BeforeCode`
   (fragile substring matching — documented as last resort)

**Domain-specific providers** can be registered via:
- `NewFixEngineWithProviders(providers...)` — for standalone engine use
- `Config.FixProviders` — for pipeline integration
- `NewFixApplierWithProviders(rootDir, providers...)` — for direct applier use

---

## Status by Category

### a) FULLY DONE

- [x] FixEdit type with byte-level operations
- [x] FixProvider interface with full documentation
- [x] OffsetProvider — byte-offset based edits
- [x] LineProvider — line/column → byte offset conversion
- [x] SubstringProvider — fallback substring matching
- [x] FixEngine rewrite to []byte with provider delegation
- [x] FixApplier rewrite to []byte (no more string roundtrips)
- [x] Pipeline Config.FixProviders wiring
- [x] NewFixApplierWithProviders constructor
- [x] Complete test rewrite (41 fix-related tests)
- [x] All existing tests still pass (490 total)
- [x] Zero lint issues

### b) PARTIALLY DONE

Nothing partially done — the core redesign is complete.

### c) NOT STARTED

- [ ] Domain-specific FixProvider implementations (Go AST, Rust syn, .gitignore, etc.)
- [ ] Finding type model improvements for byte-offset support
- [ ] Pipeline concurrency improvements (Run multiple pipelines concurrently)
- [ ] FixEdit → Finding.RoundTrip() or similar for SARIF round-tripping of edits
- [ ] CLI flag for custom FixProvider registration
- [ ] BDD tests for the new FixProvider architecture
- [ ] Documentation updates (AGENTS.md, FEATURES.md, TODO_LIST.md)
- [ ] Example showing custom FixProvider usage

### d) TOTALLY FUCKED UP

- [ ] **Nothing broken** — all 490 tests pass, zero lint issues
- [ ] **Known risk:** The `LineProvider.replacementEdit` uses `Position.Column` as the
  byte offset for BeforeCode matching. If Column is 0 (not set), it defaults to 0 which
  means the start of the line — this works but is fragile for findings without column info.
  The SubstringProvider catches these cases.

### e) WHAT WE SHOULD IMPROVE

1. **Finding type should have explicit `ByteOffset`/`ByteLength` fields** — Currently
   `Position.Offset` exists but is rarely populated by detectors. The `Range` type has
   `Start.Offset`/`End.Offset` but no detector populates them. We should add explicit
   byte-offset fields that detectors are encouraged to fill.

2. **FixEdit should be serializable** — For SARIF round-tripping, FixEdit should have
   JSON tags and conversion to/from Finding fields.

3. **BeforeCode/AfterCode should be []byte** — The Finding type uses `string` for
   BeforeCode/AfterCode. For byte-level accuracy (especially with non-UTF8 content),
   these should be `[]byte` or at least have byte-offset alternatives.

4. **ConflictDetector should work with FixEdit** — Currently conflict detection uses
   `Range.Overlaps()` on Findings. After providers resolve to FixEdits, we should detect
   conflicts at the edit level (FixEdit.Overlaps already exists but isn't wired in).

5. **lineColToOffset is O(n) per call** — For files with many fixes, we re-scan from
   the start for each fix. Should build a line-offset index once per file.

6. **FixProvider.Edits receives full file content** — For large files this is wasteful.
   Consider passing a `ContentProvider` interface that lazily reads only needed regions.

7. **Insertion behavior changed subtly** — Old engine inserted at `lineIdx` (0-based),
   new engine inserts at `lineColToOffset(line, col)` which for col=1 is the start of
   the line. This matches the old behavior but the newline appending is different.

8. **AGENTS.md / FEATURES.md / TODO_LIST.md are stale** — They don't reflect the
   new FixProvider architecture.

### f) Top 25 Things to Get Done Next

**Priority 1 — Correctness & Type Safety (High Impact, Low Effort)**

1. Update `AGENTS.md` with new FixProvider architecture documentation
2. Update `FEATURES.md` to reflect byte-level engine and provider system
3. Update `TODO_LIST.md` with new items from this redesign
4. Add `FixEdit` to SARIF round-trip support (JSON tags + conversion)
5. Wire `FixEdit.Overlaps` into conflict detection pipeline

**Priority 2 — Domain-Specific Providers (High Impact, Medium Effort)**

6. Create `pipeline/fix/provider_go_ast.go` — Go AST-based FixProvider skeleton
7. Create `pipeline/fix/provider_rust.go` — Rust syn-based provider interface
8. Create `pipeline/fix/provider_text.go` — consolidate Offset/Line/Substring into sub-package
9. Add provider auto-detection: register providers by file extension
10. Add `FixProviderRegistry` for central provider management

**Priority 3 — Performance (Medium Impact, Low Effort)**

11. Build line-offset index once per file (O(1) lookup instead of O(n) per fix)
12. Add benchmark comparing old string-based vs new byte-based engine
13. Profile allocation patterns in `applyEdits` — reduce []byte copies
14. Add `ContentProvider` interface for lazy content access

**Priority 4 — Pipeline Concurrency (High Impact, Medium Effort)**

15. Make `Pipeline.Run()` safe for concurrent use (or document how to create multiple)
16. Add `PipelinePool` for running independent pipelines concurrently
17. Add concurrent fix application across files (within a single pipeline)
18. Add `PipelineResult` merge for combining results from parallel pipelines

**Priority 5 — Testing & Quality (Medium Impact, Medium Effort)**

19. Add BDD/Ginkgo tests for FixProvider interface contract
20. Add fuzz tests for `lineColToOffset` edge cases
21. Add integration test with custom FixProvider end-to-end
22. Add property-based test: byte-level edits should preserve file invariant

**Priority 6 — Type Model Improvements (High Impact, High Effort)**

23. Add `ByteRange` type to Finding with explicit Offset/Length
24. Consider `BeforeCode []byte` / `AfterCode []byte` on Finding (breaking change)
25. Add `Finding.ToEdits() []FixEdit` convenience method

### g) Top #1 Question I Cannot Figure Out Myself

**Should the domain-specific FixProvider implementations (Go AST, Rust syn, etc.)
live INSIDE this library as `pipeline/fix/` sub-packages, or should they be
SEPARATE modules/repositories that import go-finding as a dependency?**

Arguments for inside:
- Easier discovery, single dependency for users
- Can share internal helpers (line-offset indexing, etc.)

Arguments for outside:
- go-finding stays minimal-dependency (core types depend only on stdlib)
- No heavy AST parser dependencies in the core module
- Each provider can version independently
- Clean separation of concerns

This is a product/architecture decision that affects the module structure permanently.

---

## Files Changed

```
pipeline/fix_edit.go          |  77 NEW
pipeline/fix_provider.go      | 322 NEW
pipeline/fix_engine.go        | 126 (was 203, -77 lines, cleaner)
pipeline/fix_applier.go       | 156 (was 152, +4 lines)
pipeline/pipeline.go          | 620 (was 611, +9 lines for FixProviders)
pipeline/fix_engine_test.go   | 435 (was 265, +170 lines, much more coverage)
```

**Net change:** +730 lines of new code, -317 lines removed = +413 lines net.
490 tests pass, 0 lint issues.
