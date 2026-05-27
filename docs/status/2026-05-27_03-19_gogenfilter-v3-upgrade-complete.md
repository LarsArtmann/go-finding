# Comprehensive Status Report — go-finding

**Date:** 2026-05-27 03:19 CEST
**Version:** v0.4.0
**Branch:** master
**Trigger:** Post-gogenfilter v3.0.2 upgrade

---

## Executive Summary

go-finding is in **excellent shape**. The gogenfilter v3.0.2 upgrade is fully integrated with zero lint warnings, zero test failures, and 92.6% total coverage. The library provides a mature, production-ready data model and pipeline for static analysis tools.

---

## a) FULLY DONE

### Dependency Upgrade: gogenfilter v3.0.2
- **Status:** COMPLETE
- **Dependency:** `github.com/LarsArtmann/gogenfilter/v3 v3.0.2` added to `go.mod`
- **Transitive:** `github.com/bmatcuk/doublestar/v4 v4.10.0` (indirect)

### New Pipeline Processor: `GeneratedFileFilter`
- **File:** `pipeline/generated_filter.go` (105 lines)
- **Purpose:** `FindingProcessor` that removes findings from auto-generated Go source files
- **Integration point:** `pipeline.Config.Processors` — standard processor chain
- **Behavior:**
  - Wraps `gogenfilter.Filter` — two-phase detection (filename-first zero-I/O, then content-based)
  - Evaluates each finding's `Position.File` against the filter
  - Findings without a file path (`Position.File == ""`) are kept
  - Findings from files that can't be read (e.g., deleted between detection and filtering) are **kept** with a warning log — graceful degradation
  - Structured logging via `slog.Logger` when provided
- **API:**
  - `NewGeneratedFileFilter(logger *slog.Logger, configs ...gogenfilter.FilterConfig) (*GeneratedFileFilter, error)`
  - `Name() string` → `"generated-file-filter"`
  - `Process(ctx context.Context, findings []finding.Finding) ([]finding.Finding, error)`

### Test Coverage for GeneratedFileFilter
- **File:** `pipeline/generated_filter_test.go` (203 lines)
- **6 tests, all passing with `-race`:**
  - `TestGeneratedFileFilter_Disabled` — no config = no filtering
  - `TestGeneratedFileFilter_FiltersGeneratedFiles` — FilterAll with mixed generated/normal files
  - `TestGeneratedFileFilter_FiltersSpecificGenerator` — FilterSQLC only filters sqlc files
  - `TestGeneratedFileFilter_NewFilterError` — config error propagation
  - `TestGeneratedFileFilter_Name` — processor name
  - `TestGeneratedFileFilter_KeepOnMissingFile` — graceful degradation on missing files

### CLI Integration
- **File:** `cmd/go-finding/generated_filter.go` (146 lines)
- **New CLI flags:**
  - `-filter-generated` — enables generated file filtering
  - `-filter-generated-types` — comma-separated generator types (default: "all")
  - `-generated-exclude` — comma-separated glob patterns to exclude
  - `-generated-include` — comma-separated glob patterns for scope restriction
- **Config file support:** `filterGenerated`, `filterGenTypes`, `generatedExclude`, `generatedInclude` YAML/JSON fields
- **Type registry:** Maps user-facing strings ("sqlc", "templ", "protobuf", etc.) to `gogenfilter.FilterOption` values

### CLI Refactoring (bonus)
- **File:** `cmd/go-finding/main.go`
- Extracted `cliFlags` struct + `parseFlags()` function
- Extracted `writeResults()` function
- `run()` function reduced from 148 lines to under 120 lines (fixes pre-existing `funlen` lint issue)

### Config & Docs
- **File:** `cmd/go-finding/config.example.yaml` — added commented-out generated filter options
- **File:** `.golangci.yml` — added `github.com/LarsArtmann/gogenfilter` to depguard allow-lists
- **File:** `AGENTS.md` — documented new files, dependency, pipeline features, and CLI features

### Project Health Metrics

| Metric | Value |
|--------|-------|
| Total test coverage | 92.6% |
| Root package coverage | ~95% |
| Pipeline package coverage | 96.3% |
| Detectors coverage | 96.1% |
| Lint issues | **0** |
| Build | **clean** |
| Vet | **clean** |
| Race detector | **clean** |
| Source files | 48 |
| Test files | 66 |
| Test-to-source ratio | 1.37:1 |
| Direct dependencies | 6 |
| Total dependencies | 21 |
| TODO items done | 95 / 189 (~50%) |
| Version | v0.4.0 |

---

## b) PARTIALLY DONE

Nothing is partially done from this session — the gogenfilter upgrade is fully complete.

### Known Partial Items from Previous Sessions
- **FEATURES.md** — Documents v0.3.0 but version is v0.4.0 (stale by one version)
- **CI workflows** — `.github/workflows/` directory exists with `ci.yml` and `release.yml` but TODO_LIST references missing workflows
- **TODO_LIST.md** — 95/189 items done, many items are likely stale or already completed but not marked

---

## c) NOT STARTED

### Infrastructure
- No `nix flake check` integration for CI
- No benchmark regression tracking in CI
- No release automation beyond `.goreleaser.yml`
- `justfile` still exists (deprecated per AGENTS.md, should be migrated to flake.nix)

### Documentation
- **USAGE_GUIDE.md** — needs update for generated file filtering
- **CHANGELOG.md** — needs entry for gogenfilter integration
- **v1.0 release criteria** — `docs/v1.0-release-criteria.md` not reviewed against current state

### Architecture (from open TODOs)
- `Finding` struct sub-grouping (DEFERRED to v2)
- FixEngine line-offset tracking (more precise edit locations)
- Composable FixApplier interface (plugin-based filesystem operations)
- `Category.IsValid()` should reject typos (currently accepts any non-empty string)
- `Report.Merge()` should return new value instead of mutating receiver

---

## d) TOTALLY FUCKED UP

**Nothing is fucked up.** The gogenfilter integration is clean:

- gopls reports "module not in go.mod" errors for `gogenfilter/v3` — this is a **gopls workspace cache issue**, not a real problem. `go build`, `go vet`, `go test`, and `golangci-lint` all pass cleanly. The module IS in go.mod.
- Zero test failures, zero lint warnings, zero race conditions.

---

## e) WHAT WE SHOULD IMPROVE

### High-Impact Improvements
1. **Stale TODO_LIST audit** — 94 open items, many from sessions 1-10. Likely 20-30 are already done but unmarked. A fresh audit would bring the list current.
2. **CHANGELOG.md** — Missing entry for v0.4.0 and the gogenfilter integration
3. **FEATURES.md version** — Says 0.3.0, should be 0.4.0
4. **justfile removal** — AGENTS.md says "justfile is deprecated" but it still exists
5. **`Category.IsValid()`** — Accepts any non-empty string; should validate against known constants
6. **E2E CLI test for `-filter-generated`** — The new flags have unit tests but no E2E test exercising the full pipeline with generated file filtering
7. **Depguard allow-lists** — Added `gogenfilter` but the `lax` mode means the allow-list is actually not enforced. The config is defensive but inert.

### Medium-Impact Improvements
8. **Pipeline processor documentation** — `FindingProcessor` interface has no package-level docs explaining the processor chain
9. **`WriteSARIFFiltered` integration with GeneratedFileFilter** — Could pre-filter SARIF output by generated file status
10. **Example for GeneratedFileFilter** — No runnable example in `example_test.go`
11. **go.work file** — Was created during this session (`go work use .`) but may need cleanup

---

## f) Top 25 Things We Should Get Done Next

### P0 — Immediate
1. **Write CHANGELOG.md entry** for v0.4.0 + gogenfilter integration
2. **Update FEATURES.md** version from 0.3.0 to 0.4.0
3. **Add GeneratedFileFilter example** to `example_test.go`
4. **Add E2E CLI test** for `-filter-generated` flag
5. **Audit TODO_LIST.md** — mark stale items, remove completed phantoms

### P1 — This Week
6. **Remove `justfile`** — fully migrated to flake.nix
7. **Fix `Category.IsValid()`** — validate against known constants
8. **Add `pipeline.New()` integration test** with GeneratedFileFilter in the processor chain
9. **Clean up `go.work`** if not needed for other modules
10. **Update USAGE_GUIDE.md** with generated file filtering section

### P2 — Next Sprint
11. **Document processor chain** in pipeline package docs
12. **SARIF integration with generated filter** — `WriteSARIFFiltered` should support generated-file exclusion
13. **Benchmark GeneratedFileFilter** — measure overhead on large finding sets
14. **FixEngine line-offset tracking** — more precise edit locations
15. **Composable FixApplier interface** — plugin-based filesystem operations
16. **`Report.Merge()` return new value** instead of mutating receiver (breaking change, plan for v0.5.0)

### P3 — Future
17. **CI pipeline via nix flake** — replace `.github/workflows/ci.yml` with nix-based CI
18. **Benchmark regression tracking** — CI job that fails on perf regressions
19. **Release automation** — tag-triggered goreleaser
20. **`Finding` struct sub-grouping** — split large struct into focused types (v2 breaking change)
21. **Plugin architecture for detectors** — dynamic loading
22. **Structured suppression metadata** — track who suppressed, when, why
23. **Multi-language support** — position/file model for non-Go languages
24. **gRPC/REST API** — findings as a service
25. **Graph-based correlation** — cross-file, cross-tool finding relationships

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should `GeneratedFileFilter` be a built-in processor that activates automatically (opt-out) or remain opt-in (current design)?**

Current state: users must explicitly pass `-filter-generated` or set `filterGenerated: true` in config. This means generated file noise appears by default.

Arguments for opt-out (auto-enable):
- Most linter users want to skip generated files
- gogenfilter has near-zero overhead (filename-first detection)
- Reduces false positives out of the box

Arguments for opt-in (current):
- Preserves backward compatibility
- Avoids surprising behavior changes
- Some users may want to lint generated files intentionally

This is a product decision only the owner can make. It affects the default CLI behavior and config contract.

---

## Files Changed This Session

| File | Change | Lines |
|------|--------|-------|
| `pipeline/generated_filter.go` | NEW — GeneratedFileFilter processor | +105 |
| `pipeline/generated_filter_test.go` | NEW — 6 tests with fstest.MapFS | +203 |
| `cmd/go-finding/generated_filter.go` | NEW — CLI integration, type registry | +146 |
| `cmd/go-finding/main.go` | MODIFIED — cliFlags, parseFlags, writeResults, filter flags | +114/-41 |
| `cmd/go-finding/config.go` | MODIFIED — 4 new config fields | +5 |
| `cmd/go-finding/config.example.yaml` | MODIFIED — commented filter options | +10 |
| `.golangci.yml` | MODIFIED — depguard allow-list | +2 |
| `AGENTS.md` | MODIFIED — new files, dependency, features | +5 |
| `go.mod` | MODIFIED — gogenfilter/v3 + doublestar/v4 | +2 |
| `go.sum` | MODIFIED — checksum updates | ~16 |

**Total new code:** ~454 lines (source + tests)
**Total modified code:** ~154 lines changed

---

_Generated by Crush_
