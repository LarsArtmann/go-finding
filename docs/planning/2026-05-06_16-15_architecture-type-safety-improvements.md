# Comprehensive Execution Plan — go-finding

**Date:** 2026-05-06 16:15 CET
**Session:** Post-Review Architecture & Type Safety Improvements
**Status:** Planning → Execution

---

## Reflection: What I Forgot / Could Improve

### What I Got Wrong

1. **The goconst "test.go" warning is a FALSE POSITIVE** — All 13 references already use `testFilename`. The LSP is caching stale state from before the fix. No work needed.
2. **`lock()`/`unlock()` wrapper methods are dead weight** — When `mu` was `*sync.Mutex`, these needed nil checks. Now that `mu` is `sync.Mutex` (value), they just call `r.mu.Lock()`/`r.mu.Unlock()`. Should be inlined.
3. **Triage split brain should have been fixed FIRST** — I removed deprecated types and split the CLI before fixing the actual correctness issue. Priority inversion.
4. **`Correlation.Confidence` in `merge.go` is ALSO a bare float64** — Not just `Finding.Confidence`. Both need the strong type.
5. **`WithTag` deprecation is already fully resolved** — Only `WithTags` exists in code. That TODO item is stale.

### What Could Be Better

1. **FixEngine.Apply() has its own "has code" filter** at `fix_engine.go:64-66` that differs from `Finding.HasFix()`. The engine skips findings with empty BeforeCode+AfterCode, but `HasFix()` considers FixStrategyDirect as always having a fix. This means a Direct fix without code is silently dropped.
2. **`Pipeline.triage()` doesn't call `HasFix()`** — It only categorizes by FixStrategy. A finding with `FixStrategySuggest` but no `AfterCode` goes into `Suggest` even though `HasFix()` returns false. The pipeline then tries to apply it and FixEngine silently drops it.
3. **No `TriageCategory` enum** — Pipeline returns `TriageResult{Direct, Suggest, None}` but these are just slice containers with no type safety.

### What Existing Code Fits Requirements

- **Confidence type**: `clampConfidence()` already exists — becomes a method on the type
- **SARIF round-trip**: `SarifResult.Properties map[string]any` already exists — just needs to be on Finding too
- **Triage centralization**: `Finding.HasFix()` already exists — just needs to be used everywhere
- **Error types**: `FindingError` with categories already exists — no new error types needed

### Library Considerations

- **`go-faster/yaml`** — REQUIRED to replace banned `go.yaml.in/yaml/v3` (per how-to-golang policy). Only affects CLI config parsing.
- **`encoding/json/v2`** — Available in Go 1.26.2 but major migration. DEFER to separate session.
- **No new libraries needed** — Existing patterns (Builder, functional options, FilterFunc) cover all needs.

---

## Task List (sorted by Impact × Effort, max 12 min each)

### Tier 1: Correctness (1% → 51% impact)

| # | Task | Files | Effort | Impact |
|---|------|-------|--------|--------|
| 1 | Centralize triage: make `HasFix()` the canonical "is fixable?" source | `finding.go`, `pipeline/pipeline.go`, `pipeline/fix_engine.go` | 10m | HIGH |
| 2 | Fix FixEngine.Apply() to use `HasFix()` instead of its own filter | `pipeline/fix_engine.go` | 5m | HIGH |

### Tier 2: Type Safety (4% → 64% impact)

| # | Task | Files | Effort | Impact |
|---|------|-------|--------|--------|
| 3 | Define `type Confidence float64` with `IsValid()`, `String()`, constants | NEW: `confidence.go` | 5m | MEDIUM |
| 4 | Migrate `Finding.Confidence` and `Correlation.Confidence` to `Confidence` type | `finding.go`, `merge.go`, + 15 files | 10m | MEDIUM |
| 5 | Migrate all Confidence call sites (28 occurrences across 15 files) | All production + test files | 10m | MEDIUM |

### Tier 3: Policy & Cleanup (20% → 80% impact)

| # | Task | Files | Effort | Impact |
|---|------|-------|--------|--------|
| 6 | Replace `go.yaml.in/yaml/v3` with `go-faster/yaml` | `go.mod`, `cmd/go-finding/config.go` | 10m | MEDIUM |
| 7 | Remove deprecated `diagnostic.go` wrappers from root package | `diagnostic.go` | 5m | MEDIUM |
| 8 | Inline `lock()`/`unlock()` wrappers — use `r.mu.Lock()`/`Unlock()` directly | `report.go` | 5m | LOW |
| 9 | Add `WriteSARIF` error-path test with failingWriter | NEW in `sarif_test.go` | 10m | MEDIUM |
| 10 | Add context-cancel tests for partial detection | `pipeline/partial_test.go` | 10m | MEDIUM |

### Tier 4: Documentation & Verification

| # | Task | Files | Effort | Impact |
|---|------|-------|--------|--------|
| 11 | Update `AGENTS.md` with new Confidence type and triage centralization | `AGENTS.md` | 5m | LOW |
| 12 | Update `TODO_LIST.md` with all completed items | `TODO_LIST.md` | 5m | LOW |
| 13 | Full build + test + lint verification | All | 5m | HIGH |

---

## What's NOT in This Plan (Deferred with Reason)

| Item | Reason |
|------|--------|
| `Properties map[string]any` on Finding | Medium effort, needs design decision on Metadata vs Properties coexistence |
| `encoding/json/v2` migration | Major migration touching every marshal/unmarshal call site |
| Decompose `FindingsFromSARIF` (CC 90) | Code works correctly, refactoring risk without test coverage for edge cases |
| CLI `run()` testability | Injecting io.Writer/FlagSet is good but non-blocking |
| `Category.IsValid()` strict validation | Breaking change — custom categories like "go-vet" are valid today |
| API stability review | Requires user decision on v1.0 criteria |
| `NewFinding` API pattern decision | Requires user decision (functional options vs builder-only vs current) |
| Domain-specific provider location | Architecture decision — separate module vs internal |
| Nix flake migration | Entirely separate workstream |
| Structured logging (`slog`) | Non-blocking improvement |

---

## Execution Graph

```
[1: Centralize triage] ──→ [2: FixEngine filter]
                                       │
[3: Confidence type def] ──→ [4: Migrate fields] ──→ [5: Migrate call sites]
                                       │
[6: Replace yaml lib] ──→ [7: Remove diagnostic.go]
                                       │
[8: Inline lock/unlock] ──→ [9: SARIF error test] ──→ [10: Cancel tests]
                                       │
                  [11: Update AGENTS.md] ──→ [12: Update TODO_LIST.md]
                                       │
                              [13: Full verification]
```

---

## v1.0 Release Criteria Question

**What is the minimum bar for shipping v1.0?**

Options:
- A) API lock only — decide `NewFinding` pattern, Confidence type, lock API surface
- B) API lock + triage fix — A + centralize triage logic
- C) Full quality gate — B + 90%+ coverage, no banned deps, all P0 done
- D) Ship now — current state is good enough, iterate post-v1

**My recommendation:** Option B. API lock + triage centralization. The Confidence type and yaml replacement are blocking API stability. Everything else can ship in v1.1.
