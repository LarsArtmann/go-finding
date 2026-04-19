# go-finding — Comprehensive Status Report

**Date:** 2026-04-19 05:04 CEST
**Session:** 6th round of self-audit (continuation of multi-session audit)
**Author:** Crush (AI assistant) + Lars Artmann (human)
**Branch:** `master`, 3 commits ahead of origin + uncommitted changes

---

## Executive Summary

**go-finding** is a Go library providing a unified data model and pipeline for static analysis tools. After 6 rounds of self-audit (~70+ tasks completed), the library is feature-complete at v0.1.0 with 81.1% test coverage, 4,839 production LOC, 9,039 test LOC across 4 packages. The primary blocker for CI green is ~25 remaining `golangci-lint` issues (down from ~45). Once lint is clean, remaining work is documentation polish and release preparation.

---

## A) FULLY DONE ✅

### Core Library (Root Package)

- [x] **Finding type** — Core data model with 20+ fields, constructor `NewFinding()`, auto-generated IDs
- [x] **Position/Range types** — `Overlaps()`, `Intersection()`, `Adjacent()` with full geometric correctness
- [x] **Severity enum** — Info/Warning/Error/Critical with comparison methods (`Greater`, `Less`, `AtLeast`, `AtMost`)
- [x] **FixStrategy enum** — None/Suggest/Direct/AI
- [x] **Report container** — `AddFinding`, `AddFindings`, summary stats, deduplication
- [x] **Filtering** — By severity, tool, rule, file, category; grouping utilities
- [x] **Merging** — Report merge with deduplication + `Correlate()` for cross-tool correlation
- [x] **SARIF 2.1.0** — Full round-trip: `FindingsToSARIF` + `FindingsFromSARIF` with decomposed helpers
- [x] **LSP integration** — `ToLSP()` converts Findings to LSP Diagnostics
- [x] **JSON marshaling** — `FromJSON`, `ReportFromJSON`, `PrettyJSON` with sentinel errors
- [x] **go/analysis integration** — `Diagnostic` type, conversion from `go/analysis.Diagnostic`
- [x] **Structured errors** — `FindingError` with categories (Validation/Conflict/Apply/Verify/Pipeline)
- [x] **Category constants** — Security/Performance/BugRisk/Style/Complexity/Redundancy/Dependency
- [x] **ID generation** — FNV-1a 128-bit hash, parseable format `tool:rule:file:line:col`
- [x] **Suppression** — Suppression type with reason tracking

### Pipeline Package

- [x] **Pipeline orchestrator** — Detect → Triage → Fix → Verify loop with configurable iterations
- [x] **Conflict detection** — Overlapping fix detection and analysis
- [x] **AST-aware fix application** — AST fallback with text-based fallback
- [x] **Verification stage** — Re-run detectors, diff findings pre/post fix
- [x] **Metrics** — Timing/count collection with snapshot support, thread-safe accessors
- [x] **Retry** — Exponential backoff with jitter for flaky detectors
- [x] **Partial success** — Graceful degradation, collect from failed detectors
- [x] **Parallel detection** — errgroup-based concurrent detector execution
- [x] **Config validation** — `Config.Validate()` + `RetryConfig.Validate()` with sentinel errors

### CLI Tool

- [x] **cmd/go-finding** — CLI with cobra, JSON/YAML output, pprof support, version via ldflags
- [x] **Detector specs** — Configurable detector selection (govet, staticcheck)
- [x] **Test suite** — Unit tests for buildDetectors, filter helpers, report helpers

### Infrastructure

- [x] **CI/CD** — GitHub Actions: multi-OS matrix (ubuntu + macos), coverage enforcement (75%), tag-triggered
- [x] **Tag v0.1.0** — Pushed to origin
- [x] **Justfile** — test, bench, cover, lint, fmt commands
- [x] **.golangci.yml** — Comprehensive config with 80+ linters, strict settings
- [x] **golines formatting** — All files formatted consistently

### Sentinel Error Migration (In-Progress → Committed)

- [x] `json.go` — `ErrInvalidFinding`, `ErrInvalidReport` (committed in working tree, NOT yet committed to git)
- [x] `pipeline/pipeline.go` — `errMaxIterations`, `errTimeout` (committed in working tree, NOT yet committed to git)

### Test Refactoring (Uncommitted)

- [x] `id_test.go` — Extracted `testParseIDCase` helper, reduced duplication across `TestParseID` and `TestParseID_WindowsPaths`
- [x] `cmd/go-finding/main_test.go` — Table-driven refactor of `TestBuildDetectors` with additional empty-specs case

---

## B) PARTIALLY DONE 🔧

### golangci-lint Cleanup (~25 issues remaining, down from ~45)

Current uncommitted progress:

- [~] **err113 (4 remaining in retry.go)** — Sentinel error pattern established in json.go + pipeline.go, but `pipeline/retry.go` still has 4 `errors.New()` calls that need sentinel treatment
- [~] **nolintlint (3 issues)** — Unused `//nolint` directives at `lsp_test.go:217`, `pipeline/conflict.go:137`, `pipeline/pipeline.go:678`
- [~] **wsl_v5 + nlreturn (9 issues)** — Pure style linters, should be disabled in `.golangci.yml`

### Tests That May Need Error String Updates

- [~] `json_test.go` — Tests checking for `"invalid finding: missing required fields (id, rule, ...)"` may need updating since error string changed with sentinel
- [~] `pipeline/pipeline_test.go` — Tests checking `"MaxIterations must be >= 0"` may need updating

---

## C) NOT STARTED ❌

### Documentation

- [ ] **CHANGELOG.md** — Exists (66 lines) but needs updating for v0.1.0 release content
- [ ] **README.md** — Exists (172 lines) but needs expansion: pipeline examples, API overview, badges
- [ ] **CONTRIBUTING.md** — Exists (173 lines), likely needs review
- [ ] **API docs (pkg.go.dev)** — Package godoc needs review for completeness
- [ ] **Examples directory** — Missing. Need standalone examples for common use cases

### Testing

- [ ] **CLI integration tests** — Test the actual binary with subprocess testing
- [ ] **Benchmarks** — Performance benchmarks for hot paths (merge, filter, SARIF)
- [ ] **Edge case tests** — Nil receivers, empty inputs, concurrent access patterns

### Release

- [ ] **Push uncommitted + unpushed commits** to origin
- [ ] **Release workflow** — Tag v1.0.0, GitHub Release, release notes

---

## D) TOTALLY FUCKED UP 💥

### Nothing is catastrophically broken. But these things suck:

1. **Go build cache corruption** — Required `go clean -cache` at session start. Nix + Go toolchain occasionally corrupts cache. This is an environmental issue, not a code issue.

2. **golangci-lint has 25 remaining issues** — We've been chipping away at this for 2+ sessions. The remaining issues are a mix of real code improvements (err113, exhaustive, nestif) and pedantic style (wsl_v5, nlreturn) that should just be disabled.

3. **CLI coverage is only 24.1%** — The `cmd/go-finding` package has minimal test coverage. Most functions are hard to test without subprocess testing.

4. **No examples directory** — The README promises "seven tools detect issues" but there's no standalone example showing how to use the library.

5. **3 unpushed commits + uncommitted changes** — Work is piling up. Need to commit and push.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture & Design

1. **Pipeline retry.go validation** — Should use same sentinel error pattern as pipeline.go and json.go
2. **Range.Overlaps nesting** — Complexity 10 (threshold 6). Can flatten with early returns
3. **FixStrategy exhaustive switch** — Missing `FixStrategyNone` case. Silent bug waiting to happen
4. **Error wrapping consistency** — Some paths wrap with `%w`, others don't. Standardize

### Testing

5. **CLI coverage 24.1% → target 60%+** — Need subprocess tests or refactored testability
6. **No benchmarks** — Can't reason about performance without data
7. **No fuzz tests** — ID parsing, JSON parsing, SARIF parsing are all fuzz-worthy

### Developer Experience

8. **No examples directory** — Critical for adoption. Need: basic usage, custom detector, pipeline usage
9. **README needs badges** — CI status, coverage, godoc, Go version
10. **pkg.go.dev documentation** — Every exported type needs godoc

### Release Readiness

11. **CHANGELOG needs v0.1.0 entry** — Currently generic
12. **Git history needs pushing** — 3 commits + uncommitted work sitting locally
13. **Tag strategy** — v0.1.0 exists but should we re-tag after lint cleanup?

---

## F) Top #25 Things to Do Next

| #   | Task                                                                                    | Priority | Effort | Category |
| --- | --------------------------------------------------------------------------------------- | -------- | ------ | -------- |
| 1   | **Fix `pipeline/retry.go`** — Convert 4 `errors.New()` to sentinel errors               | P0       | 5min   | Lint     |
| 2   | **Fix `finding.go:104`** — Add missing `FixStrategyNone` case in switch                 | P0       | 2min   | Lint     |
| 3   | **Fix `merge_test.go:208`** — Replace loop with `slices.Contains`                       | P0       | 2min   | Lint     |
| 4   | **Fix `position.go:337`** — Flatten `Range.Overlaps` nesting                            | P0       | 10min  | Lint     |
| 5   | **Fix `finding_extra_test.go:46`** — Extract `"changed"` to constant                    | P0       | 2min   | Lint     |
| 6   | **Fix nolintlint** — Remove 3 unused nolint directives                                  | P0       | 3min   | Lint     |
| 7   | **Fix revive** — Rename unused receivers/params to `_` (4 locations)                    | P0       | 3min   | Lint     |
| 8   | **Fix wrapcheck** — Wrap `os.Create` + `pprof.StartCPUProfile` errors in CLI            | P0       | 3min   | Lint     |
| 9   | **Update `.golangci.yml`** — Disable `wsl_v5` + `nlreturn` (9 pedantic issues)          | P0       | 2min   | Config   |
| 10  | **Fix tagliatelle** — `finding_ids` → `findingIds` in merge.go or configure tagliatelle | P1       | 3min   | Lint     |
| 11  | **Fix `position.go:376`** — Add godoc to exported `Pos()` function                      | P1       | 1min   | Lint     |
| 12  | **Fix prealloc** — Preallocate `all` slice in `pipeline_test.go:332`                    | P1       | 1min   | Lint     |
| 13  | **Run `golangci-lint run ./...`** — Verify ZERO issues                                  | P0       | 2min   | Verify   |
| 14  | **Commit all lint fixes** — Descriptive commit message                                  | P0       | 1min   | Git      |
| 15  | **Update CHANGELOG.md** — Document v0.1.0 changes comprehensively                       | P1       | 15min  | Docs     |
| 16  | **Improve README.md** — Add badges, pipeline examples, API overview                     | P1       | 20min  | Docs     |
| 17  | **Create examples/** — Standalone examples: basic, custom detector, pipeline            | P1       | 30min  | Docs     |
| 18  | **Add benchmarks** — Hot paths: merge, filter, SARIF, ID generation                     | P2       | 20min  | Testing  |
| 19  | **CLI integration tests** — Subprocess testing for cmd/go-finding                       | P1       | 30min  | Testing  |
| 20  | **Review CONTRIBUTING.md** — Ensure accuracy, add lint commands                         | P2       | 10min  | Docs     |
| 21  | **Review pkg.go.dev docs** — All exported types have godoc                              | P1       | 15min  | Docs     |
| 22  | **Push all commits** to origin                                                          | P0       | 1min   | Git      |
| 23  | **Update tests for sentinel errors** — Verify test assertions match new error strings   | P0       | 5min   | Testing  |
| 24  | **Consider re-tagging v0.1.0** after lint cleanup (or tag v0.1.1)                       | P2       | 5min   | Release  |
| 25  | **Add fuzz tests** — ID parsing, JSON, SARIF parsing                                    | P3       | 30min  | Testing  |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should `finding_ids` in the `Correlation` struct stay snake_case or change to camelCase?**

The `tagliatelle` linter flags `finding_ids` as non-camelCase. However:

- The SARIF spec uses snake_case for many fields
- JSON conventions vary: Go ecosystem often uses camelCase, but interchange formats like SARIF use snake_case
- Changing it could break any existing serialized data

**Options:**

1. Change `json:"finding_ids"` → `json:"findingIds"` (satisfy tagliatelle, breaks any existing serialized JSON)
2. Configure tagliatelle to allow snake_case for this specific field
3. Add `//nolint:tagliatelle` to the field
4. Disable tagliatelle entirely (it's the only issue)

I'd recommend option 3 (nolint directive) since this is a data interchange field and consistency with SARIF conventions matters more than Go JSON conventions.

---

## Metrics Dashboard

| Metric                      | Value                        |
| --------------------------- | ---------------------------- |
| Production LOC              | 4,839                        |
| Test LOC                    | 9,039                        |
| Test:Code ratio             | 1.87:1                       |
| Total coverage              | 81.1%                        |
| Root package coverage       | 91.7%                        |
| Pipeline coverage           | 84.0%                        |
| Internal/detectors coverage | 71.6%                        |
| CLI coverage                | 24.1%                        |
| Total commits               | 204                          |
| Lint issues remaining       | 25                           |
| Go files                    | 66                           |
| Packages                    | 4                            |
| Dependencies                | 3 (yaml.v3, x/sync, x/tools) |

## Current golangci-lint Issue Breakdown

| Linter      | Count         | Category      | Action                           |
| ----------- | ------------- | ------------- | -------------------------------- |
| err113      | 4             | Code quality  | Fix: sentinel errors in retry.go |
| wsl_v5      | 0 (disabled?) | Style         | Disable in config                |
| nlreturn    | 5             | Style         | Disable in config                |
| revive      | 6             | Code quality  | Fix real issues                  |
| nolintlint  | 3             | Cleanup       | Remove unused directives         |
| wrapcheck   | 2             | Code quality  | Wrap external errors             |
| exhaustive  | 1             | Bug risk      | Add missing switch case          |
| modernize   | 1             | Modernization | Use slices.Contains              |
| nestif      | 1             | Complexity    | Flatten nesting                  |
| goconst     | 1             | Code quality  | Extract constant                 |
| funlen      | 1             | Complexity    | Already has nolint               |
| tagliatelle | 1             | Style         | Decide on snake vs camel         |
| prealloc    | 1             | Performance   | Preallocate slice                |

---

## Uncommitted Changes Summary

```
 M cmd/go-finding/main_test.go    — Table-driven refactor of TestBuildDetectors
 M id_test.go                     — Extract testParseIDCase helper
 M json.go                        — Sentinel errors ErrInvalidFinding, ErrInvalidReport
 M pipeline/pipeline.go           — Sentinel errors errMaxIterations, errTimeout
```

**4 files changed, 77 insertions(+), 69 deletions(-)**

---

_Report generated by Crush (AI) on 2026-04-19 at 05:04 CEST._
