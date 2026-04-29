# Comprehensive Execution Plan: Hardening & API Stability

**Date:** 2026-04-30 00:05  
**Project:** go-finding  
**Status:** Post-Hardening Deepening — v0.x → v1.0 Readiness  

---

## Brutally Honest Assessment

### What Did We Forget?
1. **Stale binary artifacts:** Two 3.7MB binaries (`basic`, `builder`) are sitting in the repo root, untracked and unignored.
2. **Test stability verification:** We never ran `go test -count=100` to confirm flaky tests were truly fixed.
3. **`FixStrategyAI` split brain:** `triage()` groups it with `FixStrategySuggest` (no auto-apply), but `HasFix()` groups it with `FixStrategyDirect` (claims it has a fix). This is a **silent contract violation**.
4. **`run()` error paths untested:** `pipeline.New` failure, `p.Run` failure, and metrics output branches in the CLI are all uncovered (49% function coverage).
5. **Stale TODOs:** `TODO_LIST.md` claims `applyTriage` is at 11.1% coverage (it's 88.9%), `Correlate()` is unwired (it IS wired), and `intersectionByOffset` + `HasOffset` are at 0% (they are tested).
6. **`.gitignore` gaps:** No rule for example binaries or `govet` binary.
7. **`hasLineRange` line 146:** The `r.Start.Line == 0 || p.Line == 0` early return is not directly tested (85.7% coverage).
8. **Pipeline error paths:** `detectSequential` context cancellation, `detectParallel` goroutine errors, and `applyTriage` all-conflicts + apply-error branches are untested.
9. **SARIF parser fuzz:** We fuzz `ToSARIF` but do not fuzz `FindingsFromSARIF` with truly malformed/trusted input.
10. **Security in CI:** `govulncheck` exists in `justfile` but is NOT in the GitHub Actions CI workflow.

### What's Stupid That We Do Anyway?
1. **`FixStrategyAI` contradicts itself.** Users who set `FixStrategyAI` get `Suggest`-equivalent triage but `HasFix()` promises a fix exists. This is dishonest API design.
2. **We celebrate 99.1% root-package coverage while ignoring 81.1% CLI coverage.** The CLI is the user-facing surface. It matters more.
3. **We document phantoms instead of deciding.** `FixStrategyAI` has been "documented as placeholder" for multiple sessions. Decision paralysis is worse than a wrong decision.
4. **We leave 3.7MB binaries in the repo.** This bloats clones and looks unprofessional.
5. **We have TODO items that are already done.** This trains us to ignore the TODO list.

### What Could We Have Done Better?
1. **Run `go clean` + `git status` before every commit** to catch artifacts.
2. **Test CLI `run()` with dependency injection from the start** instead of fighting global `flag.CommandLine`.
3. **Make `FixStrategyAI` decision immediately** when it was first flagged as phantom, not defer it across 5+ sessions.
4. **Update `TODO_LIST.md` in the same commit** that completes the work.

### What Could Still Improve?
1. **`cmd/go-finding` coverage: 81.1% → 90%+** — biggest customer-facing gap.
2. **`pipeline` coverage: 96.4% → 98%+** — error paths in detection and triage.
3. **Root package coverage: 99.1% → 99.5%+** — edge cases in `hasLineRange`, `compareOp`, `WriteSARIF`.
4. **Security posture:** Add `govulncheck` to CI, add SARIF parser fuzz for malformed input.
5. **Performance:** Profile allocations on hot paths, add benchmark regression tracking.
6. **API cleanliness:** Resolve `FixStrategyAI`, remove or document all phantom behavior.
7. **Developer experience:** Add `go:generate stringer`, modernize loops to `slices.Contains`.

### Did We Lie?
- **Yes, by stale data.** We reported 9 commits ahead of origin; `git status` shows 6. The summary was from an intermediate rebase state.
- **Yes, by omission.** We claimed `applyTriage` had 11.1% coverage. It has 88.9%. The TODO was copied from an old report without verification.
- **Yes, by omission.** We said `Correlate()` was a ghost system. It was wired into Pipeline in a previous session via `CorrelateFindings` config field. The TODO was stale.
- **Yes, by stale LSP cache.** The typecheck warning on `examples/builder/main.go` is a gopls cache issue — the file is correct.

### How Can We Be Less Stupid?
1. **`git status && git clean -fdxn` before every planning session** — know the true state.
2. **Update TODOs in the same PR as the fix** — never let them rot.
3. **Run `go test -count=100` before declaring flaky tests fixed**.
4. **Make architectural decisions in one session, not five** — implement, remove, or deprecate. No more placeholders.
5. **Treat CLI coverage as the most important metric** — it's the user interface.

### Ghost Systems Found
| Ghost | Location | Value | Action |
|-------|----------|-------|--------|
| `FixStrategyAI` contradiction | `finding.go:126`, `pipeline.go:461` | Constant exists but behavior is split-brained | **Decide and fix in one session** |
| Stale TODOs | `TODO_LIST.md` | Completed items still marked open | **Audit and close** |
| `govulncheck` in justfile only | `justfile` | Security check not enforced in CI | **Wire into CI workflow** |

### Scope Creep Check
- **NO web frameworks** (gin, templ, htmx) — this is a static analysis library.
- **NO SQL, auth, email** (sqlc, casbin, resend) — no runtime database or HTTP surface.
- **NO AI backend** — out of scope for v0.x. `FixStrategyAI` is a type decision, not an implementation.
- **OpenTelemetry** — defer to v1.x. Not needed for core library stability.

### Split Brains Detected
- **`FixStrategyAI`**: `triage()` → `Suggest` bucket; `HasFix()` → grouped with `Direct`. These must agree.
- **`examples/builder/main.go` LSP warning**: gopls shows a typecheck error on correct code. This is a tooling split-brain, not a code split-brain.

---

## 24-Task Plan (30–100 min each)

Sorted by: **Impact × Customer Value / Effort**

| # | Task | Package | Effort | Impact | Customer Value | Est |
|---|------|---------|--------|--------|----------------|-----|
| 1 | Delete stale `basic`/`builder` binaries + update `.gitignore` | repo | Low | **High** | **High** — professionalism | 30m |
| 2 | Fix `FixStrategyAI` split brain: decide behavior, fix `HasFix()` or `triage()` | `finding` / `pipeline` | Low | **High** | **High** — API honesty | 45m |
| 3 | Add `run()` test: `pipeline.New` error path (invalid config) | `cmd` | Medium | **High** | **High** — CLI reliability | 60m |
| 4 | Add `run()` test: `p.Run` error path (detector failure) | `cmd` | Medium | **High** | **High** — CLI reliability | 60m |
| 5 | Add `run()` test: metrics output branch | `cmd` | Low | **Medium** | **Medium** — CLI completeness | 30m |
| 6 | Add `applyTriage` edge cases: all-conflicts early return + apply error | `pipeline` | Medium | **Medium** | **Medium** — coverage | 45m |
| 7 | Add `detectSequential` error path: context cancellation + detector error | `pipeline` | Medium | **Medium** | **Medium** — coverage | 45m |
| 8 | Add `detectParallel` error path: goroutine detector failure | `pipeline` | Medium | **Medium** | **Medium** — coverage | 45m |
| 9 | Add `hasLineRange` edge case test (Start.Line == 0) | `finding` | Low | **Low** | **Low** — coverage | 30m |
| 10 | Add `severityToSARIFLevel` / `severityToLSP` default-case tests | `finding` | Low | **Low** | **Low** — coverage | 30m |
| 11 | Add `Range.Contains` edge case: inverted ranges, nil receiver | `finding` | Medium | **Low** | **Low** — correctness | 45m |
| 12 | Add SARIF parser fuzz test: `FindingsFromSARIF` with malformed input | `finding` | Medium | **Medium** | **Medium** — security | 60m |
| 13 | Add `FindingsFromSARIF` schema validation test | `finding` | Medium | **Low** | **Medium** — correctness | 45m |
| 14 | Add `govulncheck` step to GitHub Actions CI | `.github` | Low | **Medium** | **Medium** — security | 30m |
| 15 | Add `RetryConfig.Validate` edge-case tests | `pipeline` | Low | **Low** | **Low** — coverage | 30m |
| 16 | Profile memory allocations: run benchmarks with `-benchmem` | repo | Low | **Low** | **Low** — performance | 30m |
| 17 | Add benchmark regression tracking script | repo | Medium | **Low** | **Low** — performance | 45m |
| 18 | Modernize remaining loops to `slices.Contains` / `maps.Keys` | `finding` | Low | **Low** | **Low** — maintenance | 30m |
| 19 | Add `go:generate stringer` for `Severity`, `FixStrategy`, `Category` | `finding` | Low | **Low** | **Medium** — DX | 45m |
| 20 | Remove stale TODO items and update `TODO_LIST.md` | repo | Low | **Low** | **Medium** — hygiene | 30m |
| 21 | Remove unused `//nolint` directives (audit all 60 matches) | repo | Low | **Low** | **Low** — hygiene | 30m |
| 22 | Document SARIF round-trip losses (`RelatedRef.FindingID`, `BeforeCode`) | `finding` | Medium | **Low** | **Low** — docs | 30m |
| 23 | Add per-package coverage thresholds in CI | `.github` | Medium | **Low** | **Medium** — quality gate | 45m |
| 24 | Add `go.work` for local development | repo | Low | **Low** | **Low** — DX | 30m |

---

## 60-Task Breakdown (max 12 min each)

### Tier 1: Cleanup & Trust (P0)

| # | Task | Est | File |
|---|------|-----|------|
| 1.1 | Delete `basic` binary from repo root | 2m | terminal |
| 1.2 | Delete `builder` binary from repo root | 2m | terminal |
| 1.3 | Add `basic`, `builder` to `.gitignore` | 3m | `.gitignore` |
| 1.4 | Verify `git status` is clean | 2m | terminal |
| 2.1 | Decide `FixStrategyAI` behavior: Suggest-like or remove from `HasFix()` | 5m | `finding.go` |
| 2.2 | Implement decision: fix `HasFix()` or `triage()` | 5m | `finding.go` / `pipeline.go` |
| 2.3 | Update tests to match new behavior | 5m | `finding_test.go`, `fix_strategy_test.go` |
| 2.4 | Update docs (`USAGE_GUIDE.md`, `README.md`) | 5m | docs |
| 20.1 | Audit `TODO_LIST.md` for completed items | 5m | `TODO_LIST.md` |
| 20.2 | Mark completed items, remove stale ones | 5m | `TODO_LIST.md` |
| 20.3 | Verify all remaining TODOs are actually open | 5m | terminal |

### Tier 2: CLI Reliability (P1)

| # | Task | Est | File |
| 2.4 | Add `run()` test: invalid `pipeline.New` config triggers error | 10m | `cmd/go-finding/integration_test.go` |
| 3.1 | Add `run()` test: `p.Run` returns error (mock failing detector) | 10m | `cmd/go-finding/integration_test.go` |
| 3.2 | Add `run()` test: empty detector list path | 8m | `cmd/go-finding/integration_test.go` |
| 3.3 | Add `run()` test: metrics output branch (`result.Metrics.TotalDuration > 0`) | 8m | `cmd/go-finding/integration_test.go` |
| 3.4 | Add `run()` test: outputResults write error path via `run()` | 8m | `cmd/go-finding/integration_test.go` |
| 3.5 | Refactor `run()` to be more testable (extract pipeline creation) | 12m | `cmd/go-finding/main.go` |

### Tier 3: Pipeline Edge Cases (P1)

| # | Task | Est | File |
| 6.1 | Add `applyTriage` test: all fixes conflict → `len(safeFixes) == 0` | 8m | `pipeline/pipeline_test.go` |
| 6.2 | Add `applyTriage` test: `applyDirectFixes` returns error | 8m | `pipeline/pipeline_test.go` |
| 7.1 | Add `detectSequential` test: context cancelled mid-detection | 8m | `pipeline/pipeline_test.go` |
| 7.2 | Add `detectSequential` test: detector error stops sequential run | 8m | `pipeline/pipeline_test.go` |
| 8.1 | Add `detectParallel` test: one detector fails, others succeed | 8m | `pipeline/pipeline_test.go` |
| 8.2 | Add `detectParallel` test: context cancelled during parallel detection | 8m | `pipeline/pipeline_test.go` |

### Tier 4: Root Package Edge Cases (P2)

| # | Task | Est | File |
| 9.1 | Add `hasLineRange` test: `r.Start.Line == 0` returns false | 5m | `position_extra_test.go` |
| 9.2 | Add `hasLineRange` test: `p.Line == 0` returns false | 5m | `position_extra_test.go` |
| 10.1 | Add `severityToSARIFLevel` test: empty severity → `"warning"` | 5m | `sarif_test.go` |
| 10.2 | Add `severityToLSP` test: invalid severity → `LSPSeverityWarning` | 5m | `lsp_test.go` |
| 11.1 | Add `Range.Contains` test: inverted range behavior | 8m | `position_extra_test.go` |
| 11.2 | Add `Range.Contains` test: zero-value range | 8m | `position_extra_test.go` |
| 11.3 | Add `toZeroBased` test: `Line = 0` → `0` | 5m | `lsp_test.go` |
| 15.1 | Add `RetryConfig.Validate` test: `BaseDelay > 0 && MaxDelay == 0` | 5m | `pipeline/retry_test.go` |
| 15.2 | Add `RetryConfig.Validate` test: `BaseDelay > MaxDelay` | 5m | `pipeline/retry_test.go` |
| 15.3 | Add `RetryConfig.Validate` test: all negative values | 5m | `pipeline/retry_test.go` |

### Tier 5: Security & Fuzz (P2)

| # | Task | Est | File |
| 12.1 | Create `FuzzFindingsFromSARIF` with random byte input | 10m | `sarif_fuzz_test.go` |
| 12.2 | Add malformed JSON corpus cases to fuzz seed | 5m | `sarif_fuzz_test.go` |
| 12.3 | Add missing required fields corpus cases | 5m | `sarif_fuzz_test.go` |
| 13.1 | Add `FindingsFromSARIF` schema version validation test | 8m | `sarif_test.go` |
| 13.2 | Add `FindingsFromSARIF` unknown tool/empty run test | 8m | `sarif_test.go` |
| 14.1 | Add `govulncheck` step to `.github/workflows/ci.yml` | 8m | `.github/workflows/ci.yml` |
| 14.2 | Verify `govulncheck` passes in CI simulation | 5m | terminal |

### Tier 6: Performance & Tooling (P3)

| # | Task | Est | File |
| 16.1 | Run `go test -bench=. -benchmem` and capture baseline | 5m | terminal |
| 16.2 | Analyze top 3 allocation hotspots from benchmark output | 10m | terminal |
| 17.1 | Create `scripts/bench-compare.sh` for regression tracking | 10m | `scripts/bench-compare.sh` |
| 18.1 | Replace manual loop in `merge_test.go:208` with `slices.Contains` | 5m | `merge_test.go` |
| 18.2 | Replace manual loops in `filter.go` if any remain | 5m | `filter.go` |
| 19.1 | Add `//go:generate go run golang.org/x/tools/cmd/stringer` to `severity.go` | 3m | `severity.go` |
| 19.2 | Generate `severity_string.go` | 5m | terminal |
| 19.3 | Generate `fixstrategy_string.go` | 5m | terminal |
| 19.4 | Generate `category_string.go` | 5m | terminal |
| 21.1 | Audit all `//nolint` directives for staleness | 8m | all `.go` files |
| 21.2 | Remove confirmed stale directives | 5m | various |
| 22.1 | Document `RelatedRef.FindingID` round-trip loss in SARIF | 5m | `sarif.go` comments |
| 22.2 | Document `BeforeCode` round-trip loss in SARIF | 5m | `sarif.go` comments |
| 23.1 | Add per-package coverage threshold script | 8m | `scripts/cov-check.sh` |
| 23.2 | Wire threshold script into CI workflow | 5m | `.github/workflows/ci.yml` |
| 24.1 | Create `go.work` with root module | 5m | `go.work` |

---

## Execution Graph

```mermaid
flowchart TD
    subgraph P0["Tier 0: Cleanup & Trust"]
        A1[1.1-1.4 Delete stale binaries]
        A2[2.1-2.4 Fix FixStrategyAI split brain]
        A3[20.1-20.3 Audit & update TODOs]
    end

    subgraph P1["Tier 1: CLI Reliability"]
        B1[3.1-3.5 run() error path tests]
    end

    subgraph P2["Tier 2: Pipeline Edge Cases"]
        C1[6.1-6.2 applyTriage tests]
        C2[7.1-7.2 detectSequential tests]
        C3[8.1-8.2 detectParallel tests]
    end

    subgraph P3["Tier 3: Root Package Edge Cases"]
        D1[9.1-9.2 hasLineRange tests]
        D2[10.1-10.2 severity default tests]
        D3[11.1-11.3 Range.Contains tests]
        D4[15.1-15.3 RetryConfig tests]
    end

    subgraph P4["Tier 4: Security & Fuzz"]
        E1[12.1-12.3 SARIF fuzz tests]
        E2[13.1-13.2 SARIF schema tests]
        E3[14.1-14.2 govulncheck in CI]
    end

    subgraph P5["Tier 5: Performance & Tooling"]
        F1[16.1-16.2 Memory profiling]
        F2[17.1 Benchmark regression script]
        F3[18.1-18.2 Modernize loops]
        F4[19.1-19.4 go:generate stringer]
        F5[21.1-21.2 Remove stale nolint]
        F6[22.1-22.2 SARIF round-trip docs]
        F7[23.1-23.2 Coverage thresholds]
        F8[24.1 go.work]
    end

    A1 --> A2
    A2 --> A3
    A3 --> B1
    B1 --> C1
    C1 --> C2
    C2 --> C3
    C3 --> D1
    D1 --> D2
    D2 --> D3
    D3 --> D4
    D4 --> E1
    E1 --> E2
    E2 --> E3
    E3 --> F1
    F1 --> F2
    F2 --> F3
    F3 --> F4
    F4 --> F5
    F5 --> F6
    F6 --> F7
    F7 --> F8
```

---

## How This Contributes to Customer Value

1. **Trust:** Working examples, clean repo, and honest API contracts mean users trust the library.
2. **Reliability:** CLI error-path coverage means users get clear error messages, not silent failures.
3. **Security:** SARIF fuzz tests and `govulncheck` in CI protect against untrusted input and known vulnerabilities.
4. **Performance:** Memory profiling and benchmark regression tracking keep the library fast on large codebases.
5. **Maintainability:** Stale TODO cleanup, stringer generation, and modernized loops reduce future friction.
6. **API Stability:** Resolving `FixStrategyAI` split brain is a prerequisite for tagging v1.0.0.

---

## Execution Notes

- **Never break the build.** Every commit must pass `go test ./...` and `just lint`.
- **Commit after each smallest self-contained change.** Use detailed commit messages.
- **Push when a tier is complete**, not necessarily after every commit.
- **If a task reveals a larger problem**, stop and reassess the plan before proceeding.
- **Target:** 93.4% total coverage → 95%+, `cmd/go-finding` 81.1% → 90%+.
