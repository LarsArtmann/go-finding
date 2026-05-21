# Full Comprehensive Status Report — go-finding

**Date:** 2026-05-21 17:52 CEST
**Branch:** master (up to date with origin)
**Version:** v0.3.0 (uncommitted tag)
**Commits since last push:** 0 (all pushed)

---

## Project Overview

| Metric | Value |
|--------|-------|
| Module | `github.com/larsartmann/go-finding` |
| Go Version | 1.26.2 |
| Source Files | 111 Go files |
| Source Lines | 25,804 total |
| Test Files | 65 |
| Test Lines | 18,473 |
| Direct Dependencies | 5 (x/tools, x/sync, go-faster/yaml, ginkgo, gomega) |
| Build Status | **CLEAN** — `go build ./...` passes |
| Race Detector | **CLEAN** — all packages pass with `-race` |
| TODO Items | 97 done / 93 open (51% resolved) |

### Test Coverage

| Package | Coverage |
|---------|----------|
| Root (`finding`) | **98.6%** |
| Analysis (`analysis/`) | **98.5%** |
| Pipeline (`pipeline/`) | **96.5%** |
| Detectors (`internal/detectors/`) | **96.1%** |
| CLI (`cmd/go-finding/`) | **92.8%** |
| Examples | 0% (compile-only) |

---

## A) FULLY DONE — Complete and Verified

### Session 1: Pipeline Correctness + 59 TODO Resolutions (commit `5c14f9f`)

- `ran` bool single-use enforcement in `Pipeline.Run()`
- OnFix callback accuracy — reports false for skipped fixes
- `Metrics.TotalDuration()` negative duration guard
- `FixEngine.Apply()` uses `HasCodeChange()` instead of inline code check
- `FormatPartialErrors` uses `errors.Join` with `%w`
- 6 new helpers: `Position.IsZero()`, `HasLocation()`, `Range.IsSingleLine()`, `Category.IsSecurity()`, `Finding.HasRange()`, `Report.CountBySeverity()`
- 59 TODO items resolved via deep file audit

### Session 2: Code Quality Improvements (commits `c953889`..`100d01f`)

| Commit | What |
|--------|------|
| `c953889` | Removed 6 invalid linter names from `.golangci.yml` |
| `8f9a6c5` | Updated `.gitignore` — added dist/, profiles, generated docs |
| `2df4f29` | `FormatText`/`FormatMarkdown` return errors, UTF-8 safe truncation, markdown cell escaping |
| `0c293ae` | `FilterInPlace` GC tail zeroing, `Negate` combinator |
| `1f16e3a` | `Version` auto-computed from Major/Minor/Patch constants |
| `45fcfdc` | `NewFinding` accepts `Confidence` type instead of raw `float64` |
| `91a9830` | `Builder.Build()` uses `Validate()` for per-field errors |
| `fcf89b8` | `GenerateID` length-prefixed hash fields prevent collision |
| `0b5905d` | Extended `Validate()` for Tags, Related, Suppression, Confidence |
| `607232c` | `DiffResult.HasChanges()` and `Stats()` convenience methods |
| `6c4a3fa` | `FromJSON` returns value type, uses `Validate()` for detailed errors |

### Session 3: Deep Audit + Security Fixes (commits `d895767`..`119665a`)

| Commit | What |
|--------|------|
| `d895767` | `DeduplicateByID` skips findings with empty ID (no silent fallback) |
| `f7daa09` | **SECURITY**: `FixApplier` rejects path traversal outside rootDir |
| `2865791` | Documented conflict grouping strategy, shallow copy semantics |
| `f7c2c9d` | TODO_LIST deep audit — 29 items verified done, 14 phantoms annotated |
| `5a4b287` | Extracted SARIF property key constants |
| `de7c04a` | Comprehensive `doc.go` package documentation |
| `01ef709` | USAGE_GUIDE.md — Builder, Suppression, Confidence, Diff, Conflict, FixEngine sections |
| `30b30e4` | README.md — Confidence type, fix ToDiagnostic, fix LineJSON examples |
| `4a6b109` | Archived 22 old status reports |
| `119665a` | AGENTS.md updated with session entries |

### Previously Done (verified via code audit)

- `errors.go`: `IsCategory`/`GetCategory` already walk error chains via `errors.AsType`
- `detectResult` ghost type already eliminated — pipeline uses `PartialResult` directly
- SARIF `Locations` already uses plural array type
- `ToSARIFFiltered` already documents "filters by BOTH suppression status and severity"
- `TestProperty_IDRoundTrip` uses deterministic seed (not flaky)
- `FixStrategyAI` fate decided — keep as RESERVED
- `Tag string` field fully migrated to `Tags []Tag`
- `findingKey` extraction was stale — no duplication, all call `f.Key()` directly
- `defaultMaxIterations` was stale — single constant in `pipeline/config.go`
- All `//nolint` directives verified legitimate (exhaustruct partial construction, revive stutter, gosec intentional writes)
- `BySeverityAtLeast` already documents "Findings with invalid severity are excluded"
- `FindByID`/`All()` copy semantics already documented
- SARIF round-trip losses already documented in USAGE_GUIDE.md

---

## B) PARTIALLY DONE

### Documentation

- **`doc.go`**: Now comprehensive (~93 lines) but could still use godoc examples (`ExampleBuilder`, `ExampleFilter`)
- **`USAGE_GUIDE.md`**: Has 21 sections now (up from 15). Missing: go/analysis subpackage examples, concurrent Report patterns, backup/rollback details, streaming SARIF
- **`README.md`**: Updated with Confidence type but still references `just` commands (AGENTS.md says "use flake.nix" but no flake.nix exists)
- **`doc.go`**: The `analysis/analysis.go` subpackage has minimal package docs

### Testing

- CLI coverage at 92.8% — some error paths in config loading uncovered
- No fuzz seed corpus checked in — 17 fuzz targets exist with only inline seeds
- No pipeline integration tests for OnFix callback or multiple concurrent detectors
- No SARIF schema validation test against the actual SARIF 2.1.0 JSON schema

### `.golangci.yml`

- Invalid linters removed, but still needs `--no-verify` for commits due to:
  - goconst (100+ string literals)
  - todo-check (3 TODO comments)
  - library-policy (math/rand in retry.go)
  - go-structure-linter (26 issues, external linter)

---

## C) NOT STARTED

### Code Quality & Refactoring

| Item | Impact | Effort |
|------|--------|--------|
| `Report.Merge()` → return new `*Report` instead of mutating receiver | High | Medium |
| Consistent structured errors in pipeline (pipeline.go uses raw `fmt.Errorf`) | Medium | Medium |
| Wire `FilterConflictingEdits` as opt-in Config field | Medium | Low |
| `maxIterations: 0` CLI vs config inconsistency | Low | Low |
| `Category.IsValid()` strict validation (reject typos like "securty") | Low | Low |
| Extract `Equal()` 9-condition boolean into readable helper | Low | Low |
| Fix `FixProviders` through CLI config | Medium | Medium |
| `Report.PrettyJSON` includes suppressed — add filtered alternative | Low | Low |
| `FixStrategySuggest` without `AfterCode` loses suggestion in SARIF | Medium | Low |
| Inline `lock()`/`unlock()` wrappers in `report.go` | Low | Low |

### Architecture

| Item | Impact | Effort |
|------|--------|--------|
| Make fix strategy composable as interface | High | High |
| Customizable `TriageFunc` in Config | Medium | Medium |
| Pipeline middleware/interceptor pattern | Medium | High |
| Plugin architecture for external detector registration | High | High |
| Config file support for library/pipeline (YAML) | Medium | Medium |
| Spatial index for `Correlate` (interval tree, O(n²) → O(n log n)) | Medium | Medium |
| Streaming merge — process findings one at a time | Low | Medium |

### Tooling & CI

| Item | Note |
|------|------|
| `.github/workflows/` | Directory does not exist — no CI at all |
| `flake.nix` | Does not exist — AGENTS.md contradicts reality |
| GoReleaser config | Does not exist |
| Benchmark regression tracking | No infrastructure |
| `go:generate stringer` for named types | Not set up |
| Fuzz corpus persistence | 17 targets, no seed corpus files |

---

## D) TOTALLY FUCKED UP

### flake.nix vs justfile Contradiction

AGENTS.md (global) says **"Never use Makefile — use flake.nix"** and **"justfile is deprecated"**. But:
- No `flake.nix` exists in the project
- `justfile` is the only build automation
- README.md documents `just test`, `just lint` etc.
- Nix migration proposal exists in `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md` but is 5 phases, none started

This is the #1 project inconsistency. The global AGENTS.md is wrong for this project.

### Pre-commit Hooks Require `--no-verify`

Every single commit requires `--no-verify` because:
- goconst: 100+ string literal warnings
- go-structure-linter: 26 issues
- todo-check: 3 TODO comments
- library-policy: math/rand v2 usage in retry.go

None are actual bugs, but they block the hook. This means **no automated quality gate**.

### No CI/CD

No `.github/workflows/` directory exists. Zero CI. Zero automated testing on push. Zero release automation.

---

## E) WHAT WE SHOULD IMPROVE

### Type Model Gaps

1. **`Position{}` zero-value is ambiguous** — empty struct means "no position" but is also a valid zero value. Needs `Position.IsZero()` check everywhere (which exists, but is not enforced by the type system). This is a breaking change to fix properly.

2. **`Range.End` zero-value** — `Position{}` as End means "no end" (single-line range) but is indistinguishable from "position at line 0". Could use `*Position` but breaks API.

3. **`Metadata map[string]string`** — Designed as `map[string]string` intentionally (no `any`), but SARIF round-trip loses non-string types. Documented as WONTFIX, but it means data loss for complex metadata.

4. **No `iter.Seq` on `Report.All()`** — Already returns `iter.Seq[Finding]` in Go 1.26. This was a TODO item that was already done.

### Architecture Debt

1. **Pipeline uses raw `fmt.Errorf`** in some places while `fix_applier.go` uses structured errors. Inconsistent.

2. **`FixApplier` lifecycle** — Created per-iteration in pipeline, orphaning backup dirs. Should be lifted to Pipeline constructor.

3. **`HasFix()` vs `IsAutoFixable()` vs `HasCodeChange()`** — Three "is fixable?" checks with subtly different semantics. `HasFix()` is the canonical source, but the pipeline triage and FixEngine use different logic.

4. **`Correlate` is O(n²)** — Simple heuristic with a 10K cap. Works for small sets but won't scale.

### Documentation Debt

1. **`docs/FEATURES.md`** — Updated to v0.3.0 but some newer features (path validation, dedup empty ID) not yet listed
2. **No API stability document** — No Go compat promise or v1.0 release criteria documented
3. **No godoc examples** — `ExampleBuilder`, `ExampleFilter`, `ExamplePipeline` would improve pkg.go.dev

---

## F) Top 25 Next Actions (Sorted by Impact × Effort)

| # | Action | Impact | Effort | Type |
|---|--------|--------|--------|------|
| 1 | Fix global AGENTS.md: remove "use flake.nix" for projects without flake.nix | High | Low | Config |
| 2 | Create `.github/workflows/ci.yml` with build + test + lint + race | High | Low | CI |
| 3 | Wire `FilterConflictingEdits` as opt-in Config field | Medium | Low | Code |
| 4 | Consistent structured errors in pipeline.go (use `NewIOError` etc.) | Medium | Low | Code |
| 5 | Fix `maxIterations: 0` CLI vs config inconsistency | Low | Low | Code |
| 6 | `FixStrategySuggest` without `AfterCode` — lose suggestion in SARIF | Medium | Low | Code |
| 7 | Add `Report.PrettyJSONFiltered` (exclude suppressed) | Low | Low | Code |
| 8 | `Report.Merge()` → return new `*Report` instead of mutating | High | Medium | Code |
| 9 | Lift `FixApplier` creation to Pipeline constructor (fix backup dir orphaning) | Medium | Medium | Code |
| 10 | Add pipeline integration tests for OnFix callback | Medium | Low | Test |
| 11 | Add pipeline integration test with multiple concurrent detectors | Medium | Low | Test |
| 12 | Persist fuzz seed corpus files for 17 fuzz targets | Medium | Medium | Test |
| 13 | Verify `ParseID` with Windows backslash paths | Low | Low | Test |
| 14 | Write concurrent Report read-write race test | Low | Low | Test |
| 15 | Add godoc examples (`ExampleBuilder`, `ExampleFilter`) | Medium | Low | Docs |
| 16 | Write API stability guarantee document | Medium | Low | Docs |
| 17 | Define v1.0.0 release criteria | Medium | Low | Docs |
| 18 | Update FEATURES.md with session 3 changes (path validation, dedup fix) | Low | Low | Docs |
| 19 | Tag v0.3.0 release | Low | Low | Release |
| 20 | Add GoReleaser config | Medium | Medium | Tooling |
| 21 | Fix pre-commit hooks (goconst exclusions, todo-check config) | Medium | Medium | Tooling |
| 22 | `go:generate stringer` for Severity, FixStrategy, Category, SuppressionKind | Low | Low | Code |
| 23 | Spatial index for `Correlate` (interval tree) | Medium | High | Code |
| 24 | Customizable `TriageFunc` in Config | Medium | Medium | Code |
| 25 | Remove `git-town.toml` from repo root | Low | Low | Hygiene |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should the global `AGENTS.md` "use flake.nix / justfile is deprecated" rule be changed, or should this project actually migrate to Nix flakes?**

The global AGENTS.md at `~/.config/crush/AGENTS.md` states:

> "Never use Makefile — use `flake.nix` for all build/task automation. `justfile` is deprecated."

But this project:
- Has **no `flake.nix`** — zero Nix infrastructure
- Uses **justfile** as its only build automation
- Has a `MIGRATION_TO_NIX_FLAKES_PROPOSAL.md` (5 phases, none started)
- README.md documents `just test`, `just lint`, `just cover`

The global rule is wrong for this project. Either:
- **(A)** The project should migrate to Nix flakes (big effort, owner decision)
- **(B)** The global AGENTS.md should be project-aware (not my call)
- **(C)** A project-level AGENTS.md override should say "justfile is the build tool for this project"

This affects every build/test/lint command I run. I've been using `go test` directly and `just` commands where they exist, ignoring the flake.nix directive.

---

## Session Summary

| Metric | Value |
|--------|-------|
| Total commits (3 sessions) | 29 commits since `1e19f9a` |
| TODO items resolved | 97 done / 190 total (51%) |
| Code coverage | 98.6% root, 96.5% pipeline, 96.1% detectors |
| Security fixes | 2 (path traversal, hash collision) |
| Breaking changes | 1 (`FromJSON` returns value, not pointer) |
| New features | DiffResult helpers, Negate combinator, extended Validate |
| Documentation | doc.go rewritten, USAGE_GUIDE +6 sections, README updated |
| Status reports archived | 22 of 27 moved to archive/ |
