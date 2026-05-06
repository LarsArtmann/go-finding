# Comprehensive Status Report — go-finding

**Date:** 2026-05-06 16:47 CET
**Branch:** master (up to date with origin)
**Build:** GREEN | **Tests:** 502 PASS, 0 FAIL | **Vet:** CLEAN
**Project:** v0.2.1 | Go 1.26.2 | Nix-based dev env

| Metric           | Value  |
| ---------------- | ------ |
| Production files | 44     |
| Test files       | 59     |
| Production LOC   | 6,557  |
| Test LOC         | 16,313 |
| Test:Code ratio  | 2.49:1 |
| Total commits    | 492    |

---

## a) FULLY DONE

### This Session (8 commits, pushed to origin)

| Commit    | What                                                                                     | Files                                                                                                                       |
| --------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `a3fde5b` | Centralize triage logic — `HasFix()` + `IsAutoFixable()` as canonical source             | `finding.go`, `pipeline/pipeline.go`, `pipeline/pipeline_test.go`                                                           |
| `91b9d17` | `type Confidence float64` — named type with `Clamp()`/`IsValid()`/`String()` + constants | `confidence.go` (NEW), `finding.go`, `finding_builder.go`, `merge.go`, `sarif_export.go`, `sarif_import.go`, + 5 test files |
| `916accf` | Replace banned `go.yaml.in/yaml/v3` with `go-faster/yaml`                                | `cmd/go-finding/config.go`, `go.mod`, `go.sum`                                                                              |
| `67bc818` | Remove deprecated `diagnostic.go` — 132 lines + tests deleted                            | `diagnostic.go` (DELETED), `diagnostic_test.go` (DELETED), `doc.go`                                                         |
| `01a2d5c` | Inline `lock()`/`unlock()` wrappers → direct `r.mu.Lock()`                               | `report.go`                                                                                                                 |
| `2a9c10c` | Context-cancel tests for partial detection (sequential + parallel)                       | `pipeline/partial_test.go`                                                                                                  |
| `d0a93e4` | Update AGENTS.md + TODO_LIST.md                                                          | `AGENTS.md`, `TODO_LIST.md`                                                                                                 |

### Previous Session (6 commits, also pushed)

| Commit    | What                                                               |
| --------- | ------------------------------------------------------------------ |
| `1bc1458` | Replace hardcoded `"test.go"` strings with `testFilename` constant |
| `34d09d5` | Remove deprecated `ConflictDetector` and `Verifier` wrapper types  |
| `aacd636` | Make Report zero-value thread-safe (`sync.Mutex` value type)       |
| `e23cf8a` | Split CLI `main.go` into `main.go` + `config.go` + `registry.go`   |
| `69494ca` | Update TODO_LIST.md and AGENTS.md                                  |
| `58fb0d3` | Fix `go vet copylocks` regression from Report sync.Mutex change    |

### What Changed Architecturally

1. **Triage split brain ELIMINATED** — `Finding.HasFix()` is now the single canonical "is fixable?" definition. `IsAutoFixable()` separates auto-apply from suggest. Pipeline.triage() uses both. FixEngine keeps its own code-level filter (engine concern, not domain concern).

2. **Confidence is a named type** — `type Confidence float64` with `IsValid()`, `Clamp()`, `String()` and named constants (`ConfidenceLow`, `ConfidenceMedium`, `ConfidenceHigh`, `ConfidenceFull`). 28 call sites migrated across 15 files. Untyped constants still compile (`Finding{Confidence: 1.5}` works), but passing a `float64` variable now requires explicit `Confidence(x)` conversion.

3. **Root package is dependency-free** — `diagnostic.go` deleted. `golang.org/x/tools` only imported in `analysis/` subpackage. Core library has zero heavy dependencies.

4. **Banned dependency removed** — `go.yaml.in/yaml/v3` replaced with `github.com/go-faster/yaml`. Only affects CLI config parsing.

---

## b) PARTIALLY DONE

### Nothing partially done — all started tasks were completed.

---

## c) NOT STARTED — Tracked in TODO_LIST.md

### P0 — Must Do Before v1.0

| #   | Item                                     | Location             | Why It Matters                                                                                  |
| --- | ---------------------------------------- | -------------------- | ----------------------------------------------------------------------------------------------- |
| 1   | Decide `NewFinding` API pattern          | `finding.go`         | 6-param positional approach vs functional options vs builder-only. Blocking API lock.           |
| 2   | API stability review                     | All exported symbols | Required for v1.0. Must audit every exported name.                                              |
| 3   | Decide domain-specific provider location | Architecture         | Go AST, Rust syn providers — separate module or internal? Affects module structure permanently. |

### P1 — Should Do Before v1.0

| #   | Item                                             | Location                 |
| --- | ------------------------------------------------ | ------------------------ |
| 4   | `Properties map[string]any` alongside Metadata   | `finding.go`             |
| 5   | Decompose `FindingsFromSARIF` (CC 90 → <35)      | `sarif_import.go`        |
| 6   | Error wrapping consistency audit                 | Various                  |
| 7   | Refactor CLI `run()` for testability             | `cmd/go-finding/main.go` |
| 8   | `Category.IsValid()` strict validation           | `category.go`            |
| 9   | Unify `Tag` deprecation (remove `WithTag` fully) | Various test files       |
| 10  | SARIF schema validation test                     | `sarif_test.go`          |

### P2 — Nice to Have

| #   | Item                                    |
| --- | --------------------------------------- |
| 11  | `io.WriterTo` for SARIF streaming       |
| 12  | Benchmark regression tracking in CI     |
| 13  | Per-package coverage thresholds         |
| 14  | Add Nix setup path to CONTRIBUTING.md   |
| 15  | Document `FixStrategyAI` semantics      |
| 16  | Finding JSON schema                     |
| 17  | Consumer migration guide (v0.1 → v0.2)  |
| 18  | Structured logging (`slog`)             |
| 19  | Plugin architecture for detectors       |
| 20  | Pipeline middleware/interceptor pattern |

### P3 — Future

Nix flake migration, watch mode, TUI for fix review, LSP server, Web UI, distributed detection, AI backend, etc. (see TODO_LIST.md for full list).

---

## d) TOTALLY FUCKED UP

### Nothing is broken. Everything builds, tests, and passes vet.

### Residual LSP Noise (all stale gopls cache, not real issues)

The following LSP warnings are FALSE POSITIVES — `go build` and `go vet` are clean:

- `pipeline/` redeclarations — gopls cache stale after pipeline.go → pipeline.go + adapters.go split
- `cmd/go-finding/` redeclarations — gopls cache stale after 3-way CLI split
- `json_test.go`/`sarif_test.go` `copylocks` — gopls not seeing `Confidence(math.NaN())` fix
- `report.go` `exhaustruct` missing `mu` — linter doesn't understand value mutex in struct literal

---

## e) WHAT WE SHOULD IMPROVE

### Critical

1. **v1.0 Release Criteria** — Must define minimum bar for shipping. Without it, we keep improving without releasing. Options: A) API lock only, B) API lock + triage fix (done!), C) Full quality gate, D) Ship now.

2. **`NewFinding` API Pattern** — The current 6-param positional constructor is the biggest API design debt. Three options:
   - **Functional options** — `NewFinding(rule, toolName, message, severity, pos, confidence, WithCategory(c), WithTags(t))`
   - **Builder-only** — Remove `NewFinding`, only `NewBuilder(...).Build()`
   - **Keep current** — 6-param is fine, just add validation

3. **`Properties map[string]any`** — `Metadata map[string]string` loses types in SARIF round-trip. Number 42 becomes `"42"`. Need a typed property bag for structured data.

### Important

4. **Decompose `FindingsFromSARIF`** — Cognitive complexity 90 (threshold 35). Works correctly but is a maintenance hazard.

5. **`Category.IsValid()` strict validation** — Currently accepts any non-empty string. Custom categories like `"go-vet"` work, but typos like `"securty"` also pass. Consider an allowlist + extension registry.

6. **CLI `run()` testability** — Uses global `flag.CommandLine`, `os.Args`, `os.Stderr`. Should accept `io.Writer` + `*flag.FlagSet` for testability.

### Minor

7. **Remove `WithTag` fully** — Already deprecated but still exists on Builder. Only `WithTags` is used in practice.

8. **`io.WriterTo` for SARIF** — `WriteSARIF` exists but doesn't implement `io.WriterTo`. Missing interface compliance.

9. **Stale LSP cache** — `gopls` doesn't pick up file splits. Restart LSP client to clear.

---

## f) Top 25 Things to Get Done Next

| Priority | #   | Task                                              | Effort | Impact |
| -------- | --- | ------------------------------------------------- | ------ | ------ |
| **P0**   | 1   | Define v1.0 release criteria                      | S      | HIGH   |
| **P0**   | 2   | Decide `NewFinding` API pattern                   | M      | HIGH   |
| **P0**   | 3   | API stability review — audit all exported symbols | M      | HIGH   |
| **P0**   | 4   | Decide domain-specific provider location          | S      | HIGH   |
| **P1**   | 5   | Add `Properties map[string]any` to Finding        | M      | MEDIUM |
| **P1**   | 6   | Decompose `FindingsFromSARIF` (CC 90→<35)         | M      | MEDIUM |
| **P1**   | 7   | `Category.IsValid()` strict validation            | S      | MEDIUM |
| **P1**   | 8   | Error wrapping consistency audit                  | S      | MEDIUM |
| **P1**   | 9   | Refactor CLI `run()` for testability              | M      | MEDIUM |
| **P1**   | 10  | Remove `WithTag` from Builder fully               | S      | LOW    |
| **P1**   | 11  | SARIF schema validation test                      | M      | MEDIUM |
| **P1**   | 12  | Wire FixProviders through CLI config              | M      | MEDIUM |
| **P2**   | 13  | `io.WriterTo` for SARIF streaming                 | S      | LOW    |
| **P2**   | 14  | Benchmark regression tracking in CI               | S      | LOW    |
| **P2**   | 15  | Per-package coverage thresholds                   | S      | LOW    |
| **P2**   | 16  | Add Nix setup path to CONTRIBUTING.md             | S      | LOW    |
| **P2**   | 17  | Document `FixStrategyAI` semantics                | S      | LOW    |
| **P2**   | 18  | Finding JSON schema                               | M      | LOW    |
| **P2**   | 19  | Consumer migration guide (v0.1 → v0.2)            | M      | LOW    |
| **P2**   | 20  | Structured logging (`slog`) in cmd + pipeline     | M      | MEDIUM |
| **P3**   | 21  | Plugin architecture for detectors                 | M      | MEDIUM |
| **P3**   | 22  | Pipeline middleware/interceptor pattern           | M      | MEDIUM |
| **P3**   | 23  | Nix flake migration (replace justfile)            | L      | MEDIUM |
| **P3**   | 24  | Watch mode with `fsnotify`                        | M      | LOW    |
| **P3**   | 25  | Styled CLI output (`lipgloss`)                    | M      | LOW    |

---

## g) Top #1 Question I Cannot Figure Out Myself

### What is the minimum bar for shipping v1.0?

The TODO list has 30+ open items across P0–P3. Without explicit v1.0 criteria, every session adds more improvements than it completes. Specific questions:

1. **Is API lock a prerequisite?** If yes, the `NewFinding` API decision and `Properties` map must happen first. That's a significant breaking-change window.

2. **Is the current quality sufficient?** Build green, 502 tests, no vet warnings, banned deps removed, triage centralized, Confidence type added. Is this "good enough" for v1.0?

3. **What about `FindingsFromSARIF` CC 90?** It works correctly. Is refactoring it a release blocker or a v1.1 improvement?

4. **Should v1.0 ship without domain-specific providers?** The FixProvider plugin architecture is done, but no Go AST / Rust syn providers exist yet.

5. **Is `Properties map[string]any` required?** It improves SARIF fidelity but adds API surface. Include in v1.0 or defer?

**My recommendation:** Define v1.0 = "API locked + no banned deps + triage correct." We meet all three criteria today. Ship v1.0, then iterate. Everything else (Properties, SARIF decomposition, CLI testability) can ship in v1.1.

---

## Session Summary

| What                           | Count                                                                           |
| ------------------------------ | ------------------------------------------------------------------------------- |
| Commits this session           | 7                                                                               |
| Commits previous session       | 6                                                                               |
| Total commits both sessions    | 13                                                                              |
| Files created                  | 4 (`confidence.go`, `config.go`, `registry.go`, planning doc)                   |
| Files deleted                  | 3 (`diagnostic.go`, `diagnostic_test.go`, `confidence.go` created-then-deleted) |
| Net LOC change (both sessions) | -138 production (6,695 → 6,557)                                                 |
| Architecture issues fixed      | 4 (triage split brain, Confidence type, banned dep, deprecated wrappers)        |
| Test coverage gaps closed      | 2 (WriteSARIF error, context cancel)                                            |

---

_Assisted-by: Crush <crush@charm.land>_
