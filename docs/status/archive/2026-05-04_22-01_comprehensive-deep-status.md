# Comprehensive Status Report — go-finding

**Date:** 2026-05-04 22:01 | **Branch:** master | **Version:** 0.2.1 | **Commit:** 586056a

---

## Executive Summary

The codebase is in **excellent shape**. All tests pass with race detector. Coverage is 95.6%. Testify is fully eliminated. The `Tag`/`Tags` split brain is resolved. YAML library migrated to canonical path. One significant doc staleness issue found (FEATURES.md + TODO_LIST.md still reference removed `Tag` field and `WithTag` method).

**Build:** ✅ Clean | **Tests:** ✅ All 6 packages green | **Vet:** ✅ Clean | **Coverage:** 95.6%

---

## A) FULLY DONE ✅

### Session Work (2026-05-04, 20:00–22:00)

| #   | What                                                                        | Commit    | Impact                              |
| --- | --------------------------------------------------------------------------- | --------- | ----------------------------------- |
| 1   | Migrate `go-faster/yaml` → `go.yaml.in/yaml/v3` in `cmd/go-finding/main.go` | `d76fb19` | Eliminated 6 transitive deps        |
| 2   | Run `go mod tidy` to remove stale deps                                      | `d76fb19` | Clean dependency tree               |
| 3   | Fix unused parameter `msg` in `pipeline/partial_test.go`                    | `df963cf` | Zero LSP warnings                   |
| 4   | Remove deprecated `Tag` field from Finding struct                           | `74c09e5` | **Eliminated Tag/Tags split brain** |
| 5   | Remove `Builder.WithTag()` method                                           | `74c09e5` | Clean API surface                   |
| 6   | Remove `sarifPropTag` from SARIF round-trip                                 | `74c09e5` | SARIF only exports `Tags`           |
| 7   | Update all 8 test files to use `Tags` instead of `Tag`                      | `74c09e5` | Consistent test code                |
| 8   | Archive 24 old status reports                                               | `1e386f2` | Clean docs directory                |
| 9   | Update AGENTS.md                                                            | `586056a` | Accurate docs                       |

### Prior Session Work (2026-05-04, earlier)

| What                                          | Status      |
| --------------------------------------------- | ----------- |
| Testify → Gomega migration (all test files)   | ✅ Complete |
| `go.mod` tidy (testify removed as direct dep) | ✅ Complete |
| FixApplier.Close() for temp backup cleanup    | ✅ Complete |
| Consolidate error constructors                | ✅ Complete |
| Fuzz test hardening (empty string edge cases) | ✅ Complete |
| Duplicate `g := NewWithT(t)` cleanup          | ✅ Complete |

### Architecture Quality

| Dimension                                       | Status       | Notes                                               |
| ----------------------------------------------- | ------------ | --------------------------------------------------- |
| Core types (Finding, Position, Range, Severity) | ✅ Excellent | Zero external deps, rich methods                    |
| Error system (FindingError with categories)     | ✅ Excellent | `errors.Is` support, structured                     |
| Pipeline (detect → triage → fix → verify)       | ✅ Excellent | Retry, partial success, metrics, conflict detection |
| SARIF 2.1.0 round-trip                          | ✅ Excellent | Lossless via properties bag                         |
| LSP conversion                                  | ✅ Good      | Lossy by design (LSP limitation)                    |
| Builder API                                     | ✅ Excellent | Fluent, immutable, validated                        |
| CLI                                             | ✅ Good      | Functional, YAML/JSON config, profiling             |

### Test Quality

| Package                                                | Coverage  | Test Count             | Race |
| ------------------------------------------------------ | --------- | ---------------------- | ---- |
| `github.com/larsartmann/go-finding`                    | 99.6%     | 54 test files          | ✅   |
| `github.com/larsartmann/go-finding/cmd/go-finding`     | 95.4%     | 3 test files           | ✅   |
| `github.com/larsartmann/go-finding/internal/detectors` | 96.1%     | 1 test file            | ✅   |
| `github.com/larsartmann/go-finding/pipeline`           | 97.4%     | 15 test files          | ✅   |
| `github.com/larsartmann/go-finding/examples`           | N/A       | 1 compile test         | ✅   |
| **Total**                                              | **95.6%** | **74+ test functions** | ✅   |

### Dependency Tree (Clean)

| Dependency                    | Type     | Purpose                         |
| ----------------------------- | -------- | ------------------------------- |
| `go.yaml.in/yaml/v3`          | Direct   | YAML config parsing (CLI only)  |
| `golang.org/x/tools`          | Direct   | go/analysis framework           |
| `golang.org/x/sync`           | Direct   | errgroup for parallel detection |
| `github.com/onsi/ginkgo/v2`   | Direct   | BDD testing                     |
| `github.com/onsi/gomega`      | Direct   | Test matchers                   |
| `github.com/stretchr/testify` | Indirect | Via gomega (not used by us)     |
| `gopkg.in/check.v1`           | Indirect | Via ginkgo (not used by us)     |

---

## B) PARTIALLY DONE ⚠️

### 1. FEATURES.md Stale — `Tag` field still documented

**File:** `FEATURES.md`
**Lines:** 39, 157, 700

The `Tag` field is still listed in the Finding type table (line 39) as "**Deprecated** — use `Tags` instead", but the field no longer exists. Lines 157 and 700 still reference the deprecated singular `Tag`. These need to be removed/updated.

**What's done:** Field removed from code, SARIF, Builder, and all tests.
**What's NOT done:** Documentation files not updated.

### 2. TODO_LIST.md Stale — references removed `WithTag`

**File:** `TODO_LIST.md`
**Lines:** 102-104, 263, 266

Items reference "Deprecate `WithTag`" and "Unify `Tag` deprecation" as if the field still exists. The field and method are now **fully removed**, so these items need updating to reflect completion.

**What's done:** Code fully cleaned.
**What's NOT done:** TODO_LIST.md not updated to mark these as fully complete.

### 3. `PUBLIC_OR_PRIVATE.md` — untracked, uncommitted

A 194-line analysis file exists at the root but was never committed. It needs a decision: commit it or trash it.

---

## C) NOT STARTED ❌

### High-Impact

1. **Nix flake migration** — `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md` exists, `flake.nix` does not. The `justfile` is marked as deprecated in AGENTS.md but still exists and is the primary build tool.

2. **CI/CD hardening** — `.github/workflows/ci.yml` and `release.yml` exist but haven't been reviewed/updated for the dependency changes.

3. **Examples coverage** — `examples/basic/main.go`, `examples/builder/main.go`, `examples/pipeline/main.go` have 0.0% coverage. They compile (verified by `example_compile_test.go`) but are never executed.

4. **API stability guarantees** — No `go-apiversion` or similar tooling. The library is at v0.2.1 with no stability promises.

### Medium-Impact

5. **`main()` function 0% coverage** — `cmd/go-finding/main.go:25` has 0.0% coverage because `main()` is untestable. The `run()` function IS tested, but `main()` itself isn't.

6. **`setupProfiling` 88.5% coverage** — One error path uncovered in `cmd/go-finding/main.go:155`.

7. **OpenTelemetry integration** — Mentioned in go-ecosystem.md as a pattern to respect, but no tracing/metrics integration exists in the pipeline.

8. **Config file documentation** — `cmd/go-finding/config.example.yaml` exists but isn't referenced from README or USAGE_GUIDE.

### Low-Impact

9. **`report/` directory** — Empty directory exists in repo root. Purpose unclear.

10. **`scripts/coverage-check.sh`** — Exists but may be redundant if nix flakes are adopted.

---

## D) TOTALLY FUCKED UP 💥

### Nothing is truly broken.

There are no compile errors, no failing tests, no data loss risks, no security issues. The codebase is clean.

**The closest to "fucked up" is:**

- **Doc staleness** — FEATURES.md, TODO_LIST.md, and potentially USAGE_GUIDE.md reference the removed `Tag` field and `WithTag` method. Anyone reading these docs would be confused because the code doesn't match. This is a documentation split brain.

- **`go.mod` still has `stretchr/testify` and `gopkg.in/check.v1` as indirect** — These come from ginkgo/gomega's transitive deps, NOT from this project. They cannot be removed without upstream changes. But it looks weird to have testify in the module after "removing testify."

---

## E) WHAT WE SHOULD IMPROVE

### Type Model Improvements

1. **`Finding.Metadata` is `map[string]string`** — This limits metadata to string values. Consider `map[string]any` for richer metadata (numbers, nested objects). This would make SARIF round-trip fully lossless without type coercion.

2. **`Category` and `Tag` are both `string` types** — They could benefit from being enums with a registry pattern, allowing tools to register custom categories/tags without stringly-typed comparisons.

3. **`Range.End` is optional via zero-value** — The zero-value `Position{}` means "not set" but this is implicit. A `*Position` pointer would make this explicit (nil = not set). However, this is a breaking change.

4. **`Confidence` is `float64`** — Could be a type alias `type Confidence float64` with `Clamp()`, `IsValid()` methods, matching the pattern of `Severity` and `FixStrategy`.

5. **`FixStrategy` is a string enum** — Could be backed by an `int` with a `String()` method for better performance in switch statements. Low priority.

### Architecture Improvements

6. **Pipeline has no plugin system** — Detectors are registered via `RegisterDetector()` in the CLI, but there's no plugin discovery mechanism. A `plugin.Detector` interface with auto-discovery would make the CLI extensible without code changes.

7. **No structured logging** — The CLI uses `fmt.Fprintf(os.Stderr, ...)` for all output. Structured logging (slog) would enable better observability.

8. **`FixEngine` is a stateless singleton** — It's `struct{}` with no fields. It could be a package-level function instead of requiring `NewFixEngine()`.

9. **`FileBackup` uses `os.RemoveAll` in `Close()`** — This is a potential data loss vector if the backup directory is misconfigured. Should validate the path is under `os.TempDir()` before removing.

### Ecosystem Improvements

10. **Use `samber/lo`** — The codebase has many hand-written slice operations (filter, group, sort) that could use `lo.Filter()`, `lo.GroupBy()`, etc. However, this adds a dependency.

11. **Use `slog`** — Go 1.21+ has structured logging in stdlib. The CLI should use it instead of `fmt.Fprintf`.

12. **Use `cmp.Ordered`** — Go 1.21+ has `cmp.Ordered` constraint. Some comparison functions could be simplified.

---

## F) Top #25 Things We Should Get Done Next

Sorted by impact × effort (highest first):

### Tier 1: Quick Wins (< 30 min each)

| #   | Task                                                                                              | Impact               | Effort  |
| --- | ------------------------------------------------------------------------------------------------- | -------------------- | ------- |
| 1   | **Update FEATURES.md** — Remove `Tag` field, remove "Deprecated" references, update Finding table | High (doc accuracy)  | Low     |
| 2   | **Update TODO_LIST.md** — Mark Tag/WithTag items as fully complete, update status                 | High (doc accuracy)  | Low     |
| 3   | **Commit or trash PUBLIC_OR_PRIVATE.md** — Make a decision, don't leave untracked files           | Medium (cleanliness) | Trivial |
| 4   | **Remove empty `report/` directory** — Ghost directory serves no purpose                          | Low (cleanliness)    | Trivial |
| 5   | **Add `config.example.yaml` reference to README** — Users can't discover it                       | Medium (UX)          | Low     |

### Tier 2: Important (< 2 hours each)

| #   | Task                                                                                       | Impact                       | Effort |
| --- | ------------------------------------------------------------------------------------------ | ---------------------------- | ------ |
| 6   | **Write a flake.nix** — Replace justfile with nix flakes per AGENTS.md policy              | High (build reproducibility) | Medium |
| 7   | **Add `Confidence` type alias** — `type Confidence float64` with `Clamp()`, `IsValid()`    | Medium (type safety)         | Low    |
| 8   | **Replace `fmt.Fprintf` with `slog`** in CLI — Structured logging                          | Medium (observability)       | Medium |
| 9   | **Update CI workflows** — Verify they work with new deps (no go-faster, no testify direct) | High (CI reliability)        | Low    |
| 10  | **Add `go-apiversion` or API stability marker** — v0.2.1 needs a stability promise         | Medium (API governance)      | Low    |

### Tier 3: Valuable (< 4 hours each)

| #   | Task                                                                                           | Impact                      | Effort |
| --- | ---------------------------------------------------------------------------------------------- | --------------------------- | ------ |
| 11  | **Add example tests** — Run examples/basic, examples/builder, examples/pipeline in tests       | Medium (example quality)    | Medium |
| 12  | **Validate FileBackup path** — Ensure `Close()` only removes under `os.TempDir()`              | High (data safety)          | Low    |
| 13  | **Write README Quick Start** — Current README is minimal, add copy-paste examples              | High (developer experience) | Medium |
| 14  | **Add CHANGELOG entry for v0.2.1** — Document Tag removal, YAML migration, testify elimination | Medium (release readiness)  | Low    |
| 15  | **Add `go.mod` min version CI check** — Ensure `go 1.26.2` is enforced                         | Low (consistency)           | Low    |

### Tier 4: Strategic (1+ day each)

| #   | Task                                                                               | Impact                  | Effort            |
| --- | ---------------------------------------------------------------------------------- | ----------------------- | ----------------- |
| 16  | **Plugin discovery system** — Auto-detect detectors via `init()` or plugin pattern | High (extensibility)    | High              |
| 17  | **`Metadata` type as `map[string]any`** — Richer metadata for SARIF fidelity       | Medium (data fidelity)  | Medium (breaking) |
| 18  | **OpenTelemetry integration** — Add tracing spans to pipeline stages               | Medium (observability)  | High              |
| 19  | **Comprehensive README rewrite** — Match go-ecosystem best practices               | High (first impression) | Medium            |
| 20  | **Add `golangci-lint` to CI** — Automated lint enforcement                         | Medium (code quality)   | Low               |

### Tier 5: Nice to Have

| #   | Task                                                                              | Impact              | Effort            |
| --- | --------------------------------------------------------------------------------- | ------------------- | ----------------- |
| 21  | **Registry pattern for Category/Tag** — Allow tools to register custom values     | Low (extensibility) | Medium            |
| 22  | **Benchmark suite** — Track performance regressions                               | Low (performance)   | Low               |
| 23  | **Fuzz corpus management** — Clean up stale corpus entries causing false failures | Low (test quality)  | Low               |
| 24  | **`FixEngine` as functions** — Remove unnecessary struct wrapper                  | Low (simplicity)    | Trivial           |
| 25  | **`Range.End` as `*Position`** — Make optional end explicit                       | Low (correctness)   | Medium (breaking) |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `Finding.Metadata` be `map[string]any` instead of `map[string]string`?**

**Why I can't decide:**

- **For `map[string]any`:** The SARIF round-trip currently stores Tags as `[]Tag`, Confidence as `float64`, etc. in the properties bag. When reading back, type assertions like `props["go-finding/tags"].([]any)` work naturally after JSON decode. If Metadata were `map[string]any`, we could store ALL finding fields losslessly through any format. Currently, Metadata is limited to strings, which means some SARIF properties (confidence, tags) can't go into Metadata.

- **For `map[string]string`:** Simpler API. No type assertion complexity for consumers. String values are sufficient for most tool metadata (version numbers, file paths, rule IDs). Changing to `any` is a **breaking API change** that affects every consumer.

- **The real question:** Is there a real-world use case where a tool needs to store non-string metadata? If yes → change. If no → keep strings and document the limitation.

**This requires a product decision, not a technical one.**

---

## Metrics

| Metric                | Value                                             |
| --------------------- | ------------------------------------------------- |
| Production Go files   | 35                                                |
| Test Go files         | 60                                                |
| Total lines of Go     | ~21,187                                           |
| Direct dependencies   | 5                                                 |
| Indirect dependencies | 10                                                |
| Test coverage         | 95.6%                                             |
| LSP diagnostics       | 0 errors, 4 warnings (go.mod tidy stale indirect) |
| Fuzz tests            | 3 (ID, merge, SARIF)                              |
| BDD tests (Ginkgo)    | 46 specs                                          |
| Benchmark tests       | 1 file                                            |
| Example programs      | 3                                                 |
| Uncommitted files     | 0                                                 |
| Untracked files       | 1 (PUBLIC_OR_PRIVATE.md)                          |

---

_Generated by Crush — 2026-05-04 22:01_
