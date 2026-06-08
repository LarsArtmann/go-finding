# Deep Review: go-finding v0.4.2

**Date:** 2026-06-08
**Reviewer:** Crush (assisted)
**Scope:** 68 test files, 52 source files, ~19,400 lines of test code, 95%+ coverage across all packages.

---

## PRO

### Architecture & Design

| #   | Strength                        | Detail                                                                                                                                                                                                                                                                  |
| --- | ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Crystal-clear domain model**  | `Finding`, `Report`, `Position`, `Range`, `Severity`, `Confidence` — each type is small, named with domain vocabulary, and has a single purpose. This is textbook DDD data modeling.                                                                                    |
| 2   | **Zero-dependency core**        | Root `finding` package imports only stdlib. `golang.org/x/tools` isolated to `analysis/` subpackage. This is a deliberate, well-documented decision that makes the library trivially adoptable.                                                                         |
| 3   | **Named types everywhere**      | `Severity`, `Confidence`, `Category`, `Tag`, `FixStrategy`, `LSPSeverity`, `LSPDiagnosticTag`, `ErrorCategory`, `CompletionReason`, `CorrelationScore` — no raw strings/ints in the API. Makes invalid states hard to construct.                                        |
| 4   | **Builder API is excellent**    | `NewBuilder(rule, toolName, msg, sev, pos).WithX(...).Build()` — required params in constructor, optional via fluent chain, `Build()` calls `Validate()`. Clean, type-safe, ergonomic.                                                                                  |
| 5   | **SARIF round-trip fidelity**   | Property bag strategy (`go-finding/*` prefix constants) preserves all non-standard fields. Streaming via `json.Encoder`. Context-aware `WriteSARIF`/`FindingsFromReader`. This is production-grade interchange.                                                         |
| 6   | **Pipeline design is sound**    | `detect → process → triage → fix → verify` loop with single-use guard, context propagation, structured logging (`slog`), per-detector timeouts, partial success, and metrics. The `FindingProcessor` chain between detection and triage is a clean extensibility point. |
| 7   | **Byte-level FixEngine**        | Descending-offset application with frontier boundary is correct. `FixProvider` interface (Offset → Line → Substring chain) is composable for domain-specific (AST-aware) providers. Conflict detection with `ConflictInfo.ConflictsWith` gives actionable diagnostics.  |
| 8   | **Thread-safe Report**          | `sync.RWMutex` with value semantics on `Report{}`. `Merge` copies under read lock. `MergeInto` returns new report without mutating. `iter.Seq[Finding]` on `All()` is modern Go.                                                                                        |
| 9   | **Error design**                | Category-based `FindingError` with `errors.Is`/`errors.As` support. `WithFinding()`/`WithPosition()` return copies (not mutate). `IsFindingError`, `GetCategory`, `IsCategory` helpers.                                                                                 |
| 10  | **Filter system is composable** | `FilterFunc` predicates with `Negate`, `AnyOf`, `FilterInPlace` (GC-safe tail zeroing). `BySeverityAtLeast`, `ByConfidenceAtLeast` for range queries. Clean, functional, no magic.                                                                                      |

### Testing & Quality

| #   | Strength                                | Detail                                                                                                                                                                                     |
| --- | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 11  | **97.2% core coverage, 93.7% pipeline** | Across 68 test files with ~19,400 lines. Race detector clean. Stress test job (`-count=20`) in CI.                                                                                         |
| 12  | **17 fuzz targets**                     | `FuzzParseID`, `FuzzMergeRandom`, `FuzzToSARIF`, `FuzzRoundTripID`, etc. Covers parsing, serialization, and round-trip paths.                                                              |
| 13  | **BDD tests with ginkgo/gomega**        | Pipeline behavior described in `pipeline/bdd_test.go`. Migration from testify is complete (zero source imports).                                                                           |
| 14  | **CI is thorough**                      | 4-job matrix (test on ubuntu+macOS, coverage enforcement, lint with golangci-lint v2, govulncheck). Per-package coverage thresholds via `scripts/coverage-check.sh`.                       |
| 15  | **Zero lint warnings**                  | All `err113`, `errcheck`, `gosec`, `goconst`, `staticcheck`, `exhaustruct`, `golines`, `paralleltest`, `nolintlint` issues resolved. 80 `//nolint` directives — all audited as legitimate. |

### Documentation & DX

| #   | Strength                     | Detail                                                                                                                                           |
| --- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| 16  | **Outstanding `doc.go`**     | Comprehensive package documentation with examples, cross-references, and honest round-trip loss documentation. ~200 lines of godoc.              |
| 17  | **Honest FEATURES.md**       | Every feature has a status. `FixStrategyAI` is explicitly marked as PLANNED placeholder. Known limitations are documented, not hidden.           |
| 18  | **Domain language defined**  | `CONTEXT.md` / `docs/DOMAIN_LANGUAGE.md` — every concept defined with purpose and invariants.                                                    |
| 19  | **API stability commitment** | `docs/API_STABILITY.md` follows Go compatibility promise style. JSON schemas in `docs/schemas/`. Release criteria in `docs/RELEASE_CRITERIA.md`. |
| 20  | **Examples compile**         | `examples/basic/`, `examples/builder/`, `examples/pipeline/` with `example_compile_test.go`.                                                     |

---

## CONTRA

### Critical Issues

| #   | Issue                                                   | Impact                                                                                                                                                                                                            | Location        |
| --- | ------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------- |
| 1   | **`flake.nix` has infinite recursion: `goPkg = goPkg`** | **Build is broken.** `nix build` will infinite-loop. The `let goPkg = goPkg;` should be `let goPkg = pkgs.go_1_26;`. Every `mkApp` and `devShell` references this, so the entire Nix toolchain is non-functional. | `flake.nix:68`  |
| 2   | **`Correlate` is O(n²) with no spatial index**          | For 10K+ findings, correlation is quadratic. Capped at 10,000 to prevent hangs, but the cap silently drops correlations. No interval tree or spatial index.                                                       | `merge.go:15`   |
| 3   | **5 commits ahead of origin, no git tag for v0.4.2**    | `Version = "0.4.2"` in code but no corresponding git tag. CI badges point to GitHub Actions but remote is behind. Consumers can't `go get` at a tagged version.                                                   | `version.go:15` |

### Design Concerns

| #   | Issue                                                            | Impact                                                                                                                                                                                                                                                                                                                 | Severity |
| --- | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------ | ------ |
| 4   | **`Metadata map[string]string` is the only extensibility field** | Documented as intentional, but: (a) no namespacing convention enforced — two tools using `"group"` collide silently, (b) no type safety — consumers must parse JSON strings, (c) no validation helper for metadata keys. The `Properties map[string]any` rejection was correct, but the alternative could be stronger. | Medium   |
| 5   | **`Position.Offset` is -1 sentinel**                             | `Offset` uses `-1` for "not set" while `Line`/`Column` use `0`. This inconsistency is documented but creates a subtle trap: `Position{}` has `Offset=0` which means "byte 0 of file", not "unset". Callers must check `HasOffset()` or compare against `-1`.                                                           | Medium   |
| 6   | **`Range.End` zero-value ambiguity**                             | `Range{Start: pos}` with no End means "single point" — but `End` is not a pointer, so you can't distinguish "no end set" from "end at zero position". `HasEnd()` checks `Line > 0                                                                                                                                      |          | Offset >= 0` but this conflates "intentionally zero" with "unset". | Medium |
| 7   | **`SubstringProvider` is fragile**                               | Uses `strings.Index` for matching `BeforeCode` in content. Ambiguous when the same text appears multiple times in a file. No nearest-position heuristic. This is the default fallback provider.                                                                                                                        | Medium   |
| 8   | **`Report.Merge` mutates receiver**                              | `Merge(other)` appends to `r.Findings` in-place. `MergeInto(other)` returns a new report. Two very similar names, very different semantics. Easy to pick the wrong one.                                                                                                                                                | Low      |
| 9   | **No `GroupID` concept**                                         | Deferred from art-dupl integration. Clone groups, cluster IDs, or any N-way relationship requires encoding in `Metadata`. `Related[clone-of]` creates O(N²) links for a group of N findings.                                                                                                                           | Low      |
| 10  | **`GenerateID` collision window**                                | Hash-based IDs use only 8 hex chars (first 4 bytes of SHA-256). For repositories with many positionless findings from the same tool+rule, birthday-paradox collision starts at ~65K findings.                                                                                                                          | Low      |

### Infrastructure Gaps

| #   | Issue                                 | Impact                                                                                                                                   |
| --- | ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 11  | **`flake.nix` missing `maintainers`** | Global AGENTS.md specifies `maintainers = [ maintainers.larsartmann ];` in meta block. This flake has no `maintainers` field.            | Low    |
| 12  | **No release workflow triggered**     | `.github/workflows/release.yml` exists but no `v*` tag has been pushed. GoReleaser config exists but is unused.                          | Medium |
| 13  | **No fuzz seed corpus**               | 17 fuzz targets exist but `testdata/fuzz/` only has auto-generated corpora from previous runs. No intentionally crafted edge-case seeds. | Low    |
| 14  | **Examples have 0% coverage**         | `examples/basic`, `examples/builder`, `examples/pipeline` show 0.0% coverage. They're compiled but not executed in CI.                   | Low    |

### Minor Issues

| #   | Issue                                                                                                                                                                                                                                                                          |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 15  | `doc.go` line 135 references `FindingsFromSARIF(ctx, data)` with context parameter — this is correct but the function signature change (added `context.Context`) happened recently; verify all examples are consistent.                                                        |
| 16  | `DeduplicateBy` is an untyped `int` const (iota). Should be a named type like the other enums. Not a bug, but inconsistent with the rest of the codebase's named-type discipline.                                                                                              |
| 17  | `confidence.go` defines named constants (`ConfidenceNone` through `ConfidenceFull`) but `NewFinding` accepts raw `Confidence` — callers can construct `Confidence(0.37)` without using named constants. The type safety is there but the ergonomics push toward raw values.    |
| 18  | `pipeline/config.go:79-82`: `DefaultMaxIterations = 5` and `DefaultTimeout = 10 * time.Minute` are generous. For a CLI tool, 10 minutes is fine. For a library embedded in a server, the caller must remember to override. Documented in `DefaultConfig()` but could surprise. |

---

## Overall Assessment

| Dimension          | Rating   | Notes                                                                                                                                                                                                  |
| ------------------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Architecture**   | **9/10** | Domain model is clean, separation of concerns is strong, extensibility points (providers, processors, triage func) are well-designed. One point off for `Metadata` being the only escape hatch.        |
| **Code Quality**   | **9/10** | Named types everywhere, zero lint, modern Go idioms (`iter.Seq`, `maps.Clone`, `slices.SortFunc`, `errors.AsType`). One point off for `SubstringProvider` fragility.                                   |
| **Testing**        | **9/10** | 95%+ coverage, 17 fuzz targets, BDD tests, stress tests, race-clean. One point off for no fuzz seed corpus and example coverage.                                                                       |
| **Documentation**  | **8/10** | Excellent godoc, honest FEATURES.md, domain language, API stability doc, schemas. Two points off: USAGE_GUIDE partially stale, no real-world integration guide.                                        |
| **Infrastructure** | **6/10** | CI is thorough when it runs, but the flake.nix is broken (`goPkg = goPkg`), no release has been tagged, and 5 commits are unpushed. This is the weakest dimension.                                     |
| **API Stability**  | **8/10** | Strong commitment documented, intentional decisions recorded, `Properties map[string]any` rejected on principle. Two points off for `DeduplicateBy` being untyped and `Report.Merge` naming confusion. |

---

## Verdict

This is a well-crafted, production-quality library with excellent domain modeling and testing. The main risks are infrastructure-level (broken flake, no release tag, unpushed commits) rather than design-level. The art-dupl integration evaluation was thorough and the deferred items (GroupID, per-relationship metadata) are correctly scoped. The go-output integration PRO/CONTRA analysis is sound — CLI-only dependency with adapter pattern is the right call.

---

## Top 3 Actions

1. **Fix `flake.nix:68`** — `let goPkg = pkgs.go_1_26;` (currently infinite recursion)
2. **Push commits + tag `v0.4.2`**
3. **Add spatial index for `Correlate`** or document the O(n²) ceiling prominently

---

_Assisted-by: Crush <crush@charm.land>_
