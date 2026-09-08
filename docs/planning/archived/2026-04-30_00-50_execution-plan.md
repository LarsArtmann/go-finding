# Execution Plan — go-finding

**Generated:** 2026-04-30 00:50
**Rule:** Every task ≤ 12 minutes. Sorted by importance → impact → effort → customer value.

---

## Scoring Key

| Axis       | Scale    | Meaning                                                                 |
| ---------- | -------- | ----------------------------------------------------------------------- |
| **Imp**    | P0–P3    | P0 = blocks users/CI, P1 = correctness, P2 = quality, P3 = nice-to-have |
| **Impact** | 1–5      | 5 = affects every user, 1 = affects 1 internal dev                      |
| **Effort** | 1–12     | Minutes. Every task ≤ 12 min                                            |
| **Value**  | 🔴🟡🟢⚪ | 🔴 = must-do, 🟡 = should-do, 🟢 = nice-to-have, ⚪ = deferred          |

---

## Plan

| #                                                             | Task                                                                                             | Imp | Impact | Effort | Value | File(s)                                      |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | --- | ------ | ------ | ----- | -------------------------------------------- |
| **PHASE 1: Correctness & Safety (P0)**                        |                                                                                                  |     |        |        |       |                                              |
| 1.1                                                           | Remove `t.Parallel()` from `TestFatalf` (mutates `os.Stderr` global)                             | P0  | 5      | 2      | 🔴    | `cmd/go-finding/main_test.go:194`            |
| 1.2                                                           | Remove `t.Parallel()` from `TestSetupProfiling_CPUProfileStartFailure` (pprof race)              | P0  | 5      | 2      | 🔴    | `cmd/go-finding/main_test.go:305`            |
| 1.3                                                           | Remove `t.Parallel()` from `TestSetupProfiling` subtests that start real CPU profiles            | P0  | 5      | 3      | 🔴    | `cmd/go-finding/integration_test.go:287-351` |
| 1.4                                                           | Add `Builder.Build()` error-path test — invalid builder returns `ErrInvalidBuilder`              | P0  | 4      | 8      | 🔴    | `finding_builder_test.go` (new)              |
| 1.5                                                           | Fix flaky `TestProperty_IDRoundTrip` — add `rand.New(rand.NewSeed(42))` or constrain Unicode     | P0  | 4      | 10     | 🔴    | `id_test.go`                                 |
| 1.6                                                           | Add `Verifier.Verify` error-path test: detector returns error on re-run                          | P0  | 4      | 10     | 🔴    | `pipeline/verify_test.go`                    |
| 1.7                                                           | Add `RetryConfig.Validate` edge-case test: zero MaxDelay, negative BaseDelay, BaseDelay>MaxDelay | P0  | 4      | 8      | 🔴    | `pipeline/retry_test.go`                     |
| 1.8                                                           | Add partial.go metrics recording during detection failures                                       | P0  | 3      | 10     | 🔴    | `pipeline/partial.go`                        |
| **PHASE 2: Test Coverage Gaps (P1)**                          |                                                                                                  |     |        |        |       |                                              |
| 2.1                                                           | Add `DeduplicateByPosition` vs `DeduplicateByRule` behavior diff test                            | P1  | 3      | 8      | 🟡    | `merge_test.go`                              |
| 2.2                                                           | Add `FixApplier` error-path test: permission denied, disk full simulation                        | P1  | 3      | 10     | 🟡    | `pipeline/fix_applier_test.go`               |
| 2.3                                                           | Add `findingFromSarResult` SARIF import test: rule metadata, helpURI, markdown                   | P1  | 3      | 10     | 🟡    | `sarif_test.go`                              |
| 2.4                                                           | Add `cloneFindings` edge-case test: nil slice, empty slice, large slice                          | P1  | 2      | 5      | 🟡    | `finding_test.go`                            |
| 2.5                                                           | Add `Finding.Equal` field-mismatch test: verify each field independently                         | P1  | 2      | 8      | 🟡    | `finding_test.go`                            |
| 2.6                                                           | Add SARIF round-trip loss documentation: `RelatedRef.FindingID`, `BeforeCode`                    | P1  | 3      | 10     | 🟡    | `sarif.go` godoc                             |
| 2.7                                                           | Add SARIF round-trip loss test: verify what's preserved and what's lost                          | P1  | 3      | 10     | 🟡    | `sarif_test.go`                              |
| 2.8                                                           | Add SARIF schema validation test against SARIF 2.1.0 JSON schema                                 | P1  | 3      | 12     | 🟡    | `sarif_test.go`                              |
| **PHASE 3: Code Hygiene (P2)**                                |                                                                                                  |     |        |        |       |                                              |
| 3.1                                                           | Audit `//nolint` directives: remove unused, add missing explanations                             | P2  | 2      | 12     | 🟡    | 60+ locations across all `.go`               |
| 3.2                                                           | Convert retry.go sentinel errors — already done! Verify & mark complete                          | P2  | 1      | 2      | 🟢    | `pipeline/retry.go:20-24`                    |
| 3.3                                                           | Remove TODO item "Convert 4 errors.New to sentinels" — already sentinel errors                   | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| 3.4                                                           | Replace hardcoded `SeverityWarning` in `diagnostic.go` with configurable param                   | P2  | 2      | 10     | 🟡    | `diagnostic.go:45`                           |
| 3.5                                                           | Preallocate `all` slice in `pipeline_test.go:315`                                                | P2  | 1      | 3      | 🟢    | `pipeline/pipeline_test.go`                  |
| 3.6                                                           | Extract `"changed"` string to constant in `finding_extra_test.go`                                | P2  | 1      | 2      | 🟢    | `finding_extra_test.go:9`                    |
| 3.7                                                           | Modernize `slices.Contains` in `merge.go` dedup key lookup                                       | P2  | 2      | 5      | 🟡    | `merge.go`                                   |
| 3.8                                                           | Modernize `slices.Delete` where manual slice manipulation exists                                 | P2  | 2      | 8      | 🟡    | pipeline + finding                           |
| 3.9                                                           | Modernize `maps.Keys`/`maps.Values` where manual extraction exists                               | P2  | 2      | 5      | 🟡    | Various                                      |
| 3.10                                                          | Fix gci formatting warning in `pipeline/fix_engine_test.go:29`                                   | P2  | 1      | 2      | 🟢    | `pipeline/fix_engine_test.go`                |
| 3.11                                                          | Fix gofumpt formatting warning in `sarif_test.go:322`                                            | P2  | 1      | 2      | 🟢    | `sarif_test.go`                              |
| 3.12                                                          | Remove stale TODO "Disable wsl_v5 + nlreturn" — not in config                                    | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| 3.13                                                          | Remove stale TODO "pipeline/astfix.go cleanup" — file doesn't exist                              | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| 3.14                                                          | Remove stale TODO "MODULE_SPLIT_PLAN.md 9 gaps" — file doesn't exist                             | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| 3.15                                                          | Remove stale TODO "Distributed detection support" — no plan, no scope                            | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| 3.16                                                          | Remove stale TODO "Write migration guide for module split" — no split planned                    | P2  | 1      | 1      | 🟢    | `TODO_LIST.md`                               |
| **PHASE 4: CI & Automation (P2)**                             |                                                                                                  |     |        |        |       |                                              |
| 4.1                                                           | Add `-count=5` stress test to CI as optional workflow                                            | P2  | 3      | 10     | 🟡    | `.github/workflows/ci.yml`                   |
| 4.2                                                           | Add per-package coverage thresholds (finding 95%, pipeline 90%, cmd 85%)                         | P2  | 3      | 10     | 🟡    | `.github/workflows/ci.yml`                   |
| 4.3                                                           | Add `gosec` to CI linting workflow                                                               | P2  | 2      | 8      | 🟡    | `.github/workflows/ci.yml`                   |
| 4.4                                                           | Add `staticcheck` to CI linting (separate from golangci-lint)                                    | P2  | 2      | 8      | 🟡    | `.github/workflows/ci.yml`                   |
| 4.5                                                           | Set up benchmark regression tracking in CI                                                       | P2  | 2      | 12     | 🟢    | `.github/workflows/bench.yml` (new)          |
| **PHASE 5: Documentation & DX (P2)**                          |                                                                                                  |     |        |        |       |                                              |
| 5.1                                                           | Document `FixStrategyAI` semantics: "AI-capable, requires AfterCode"                             | P2  | 4      | 8      | 🔴    | `fix_strategy.go` godoc                      |
| 5.2                                                           | Add `go.work` file for local multi-module development                                            | P2  | 2      | 5      | 🟢    | `go.work` (new)                              |
| 5.3                                                           | Create `CONTRIBUTING.md` with PR process, test requirements, lint config                         | P2  | 3      | 12     | 🟡    | `CONTRIBUTING.md` (new)                      |
| 5.4                                                           | Document first-release procedure (tag, changelog, GoReleaser)                                    | P2  | 2      | 8      | 🟡    | `docs/release-procedure.md` (new)            |
| 5.5                                                           | Create real-world tool integration guide with govet example                                      | P2  | 3      | 12     | 🟡    | `docs/integration-guide.md` (new)            |
| **PHASE 6: Architecture Decisions (P2–P3, needs user input)** |                                                                                                  |     |        |        |       |                                              |
| 6.1                                                           | Decide `FixStrategyAI` fate: keep as-is / implement / remove                                     | P2  | 5      | 5      | 🔴    | Product decision                             |
| 6.2                                                           | Decide stable ID format: deterministic hashes vs readable strings                                | P3  | 4      | 5      | 🟡    | Product decision                             |
| 6.3                                                           | Decide repository name: `finding` / `finding-sdk` / `go-finding`                                 | P3  | 3      | 3      | 🟡    | Product decision                             |
| 6.4                                                           | Decide on suppression expiry enforcement                                                         | P3  | 2      | 5      | ⚪    | Product decision                             |
| 6.5                                                           | Decide on go-business-rules `Severity` sharing                                                   | P3  | 2      | 5      | ⚪    | Product decision                             |
| 6.6                                                           | Decide BuildFlow: replace `PrioritizedViolation` or add `Finding` alongside                      | P3  | 2      | 5      | ⚪    | Product decision                             |
| **PHASE 7: Structural Improvements (P3)**                     |                                                                                                  |     |        |        |       |                                              |
| 7.1                                                           | Investigate `FixApplier` cross-iteration persistence: backups N→N+1                              | P3  | 3      | 12     | 🟡    | `pipeline/fix_applier.go`                    |
| 7.2                                                           | Consider `Finding` struct sub-grouping: `Location`, `Content`, `Fix`, `Metadata`                 | P3  | 4      | 5      | ⚪    | Breaking API change                          |
| 7.3                                                           | Review `Correlate` O(n²) performance (limited to 10k)                                            | P3  | 2      | 10     | 🟢    | `merge.go`                                   |
| 7.4                                                           | Measure `golang.org/x/tools` transitive dep size                                                 | P3  | 2      | 5      | 🟢    | `go.sum` analysis                            |
| 7.5                                                           | API stability review: audit every exported symbol for v1.0.0 lock                                | P3  | 5      | 12     | 🟡    | All exported APIs                            |
| 7.6                                                           | Add `go:generate stringer` for `Severity` (if worth it after audit)                              | P3  | 1      | 8      | 🟢    | `severity.go`                                |
| 7.7                                                           | Add `go:generate stringer` for `FixStrategy`                                                     | P3  | 1      | 5      | 🟢    | `fix_strategy.go`                            |
| 7.8                                                           | Add `go:generate stringer` for `Category`                                                        | P3  | 1      | 5      | 🟢    | `category.go`                                |
| 7.9                                                           | Add `go:generate stringer` for `SuppressionKind`                                                 | P3  | 1      | 5      | 🟢    | `suppression.go`                             |
| **PHASE 8: Release & Distribution (P3)**                      |                                                                                                  |     |        |        |       |                                              |
| 8.1                                                           | Add GitHub release workflow with goreleaser                                                      | P3  | 3      | 10     | 🟡    | `.github/workflows/release.yml`              |
| 8.2                                                           | Add GoReleaser config for cross-platform builds                                                  | P3  | 3      | 10     | 🟡    | `.goreleaser.yml` (new)                      |
| 8.3                                                           | Add build tags for `goexperiment.*` if needed                                                    | P3  | 1      | 5      | ⚪    | Build config                                 |
| **PHASE 9: Future Features (P3, deferred)**                   |                                                                                                  |     |        |        |       |                                              |
| 9.1                                                           | Evaluate `go-sarif` library vs hand-rolled SARIF for schema compliance                           | P3  | 3      | 12     | 🟢    | `sarif.go`                                   |
| 9.2                                                           | Config file support for library/pipeline (YAML)                                                  | P3  | 3      | 12     | ⚪    | `pipeline/config.go`                         |
| 9.3                                                           | Watch mode for continuous analysis with `fsnotify`                                               | P3  | 2      | 12     | ⚪    | `pipeline/watch.go` (new)                    |
| 9.4                                                           | IDE plugin stub — VS Code                                                                        | P3  | 2      | 12     | ⚪    | `ide/vscode/` (new)                          |
| 9.5                                                           | Web UI prototype for pipeline monitoring                                                         | P3  | 1      | 12     | ⚪    | `web/` (new)                                 |

---

## Summary Stats

| Phase                      | Tasks  | Total Effort | Avg Effort  |
| -------------------------- | ------ | ------------ | ----------- |
| 1. Correctness & Safety    | 8      | 53 min       | 6.6 min     |
| 2. Test Coverage Gaps      | 8      | 73 min       | 9.1 min     |
| 3. Code Hygiene            | 16     | 56 min       | 3.5 min     |
| 4. CI & Automation         | 5      | 48 min       | 9.6 min     |
| 5. Documentation & DX      | 5      | 45 min       | 9.0 min     |
| 6. Architecture Decisions  | 6      | 28 min       | 4.7 min     |
| 7. Structural Improvements | 9      | 74 min       | 8.2 min     |
| 8. Release & Distribution  | 3      | 25 min       | 8.3 min     |
| 9. Future Features         | 5      | 60 min       | 12.0 min    |
| **TOTAL**                  | **65** | **462 min**  | **7.1 min** |

### Stale TODOs to Remove (6 items)

These reference non-existent files, completed work, or impossible scope:

1. "Disable `wsl_v5` + `nlreturn`" — not in `.golangci.yml`
2. "Clean up `pipeline/astfix.go`" — file doesn't exist
3. "Revise `MODULE_SPLIT_PLAN.md`" — file doesn't exist
4. "Write migration guide for module split" — no split planned
5. "Distributed detection support" — no plan, no scope
6. "Convert 4 `errors.New()` to sentinels" — already sentinel errors in `retry.go`

### Already-Done TODOs to Verify & Close

1. Retry sentinels — already `errMaxRetriesNegative`, `errBaseDelayNegative`, etc. ✅

---

_Generated by Crush — 2026-04-30_
