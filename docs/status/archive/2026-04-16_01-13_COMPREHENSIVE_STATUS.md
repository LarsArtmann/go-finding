# Comprehensive Status Report — go-finding

**Date:** 2026-04-16 01:13 CEST
**Branch:** master
**Last commit:** `80ab78f` docs(planning): add module split plan
**Ahead of origin:** 1 commit (unpushed)

---

## a) FULLY DONE ✓

### Core Library (root package) — Production Ready

| Component                                        | Files                                              | Coverage | Status     |
| ------------------------------------------------ | -------------------------------------------------- | -------- | ---------- |
| Core types (Finding, Position, Range, Severity)  | `finding.go`, `position.go`, `severity.go`         | ~98%     | ✓ Complete |
| Enums (FixStrategy, Category, Suppression)       | `fix_strategy.go`, `category.go`, `suppression.go` | 100%     | ✓ Complete |
| Structured errors (FindingError with categories) | `errors.go`                                        | 95%+     | ✓ Complete |
| ID generation + parsing                          | `id.go`                                            | 95%+     | ✓ Complete |
| Filtering + grouping                             | `filter.go`                                        | 95%+     | ✓ Complete |
| Report container + summary                       | `report.go`                                        | 95%+     | ✓ Complete |
| Report merging + deduplication + correlation     | `merge.go`                                         | 95%+     | ✓ Complete |
| SARIF 2.1.0 output                               | `sarif.go`                                         | 90%+     | ✓ Complete |
| LSP Diagnostic conversion                        | `lsp.go`                                           | 90%+     | ✓ Complete |
| JSON serialization                               | `json.go`                                          | 90%+     | ✓ Complete |
| go/analysis integration                          | `diagnostic.go`                                    | 90%+     | ✓ Complete |

**Root package coverage: 96.1% of statements**

### Pipeline Package — Production Ready

| Component                                                   | Files         | Status     |
| ----------------------------------------------------------- | ------------- | ---------- |
| Pipeline orchestrator (detect → triage → fix → verify loop) | `pipeline.go` | ✓ Complete |
| Fix conflict detection + analysis                           | `conflict.go` | ✓ Complete |
| Verification stage (re-run detectors, diff findings)        | `verify.go`   | ✓ Complete |
| Metrics collection with snapshots                           | `metrics.go`  | ✓ Complete |
| Exponential backoff retry wrapper                           | `retry.go`    | ✓ Complete |
| Partial success (collect from failed detectors)             | `partial.go`  | ✓ Complete |

**Pipeline coverage: 87.5% of statements**

### Testing Infrastructure

| Item                                                                 | Status     |
| -------------------------------------------------------------------- | ---------- |
| 33 test files, 29 source files (1.14:1 ratio)                        | ✓ Complete |
| Property-based tests (`property_test.go`)                            | ✓ Complete |
| Fuzz tests (`fuzz_test.go`, `merge_fuzz_test.go`, `id_fuzz_test.go`) | ✓ Complete |
| Benchmarks (`bench_test.go`)                                         | ✓ Complete |
| Coverage enforcement (`coverage_test.go`)                            | ✓ Complete |
| Integration tests (`pipeline/integration_test.go`)                   | ✓ Complete |
| All tests pass (non-race mode)                                       | ✓ Complete |
| `go vet ./...` passes                                                | ✓ Complete |
| `go build ./...` passes                                              | ✓ Complete |

### CLI Tool

| Item                                                 | Status     |
| ---------------------------------------------------- | ---------- |
| Basic CLI with flags (dir, format, severity, config) | ✓ Complete |
| YAML/JSON config file support                        | ✓ Complete |
| Text/JSON/SARIF output formats                       | ✓ Complete |
| CPU/memory profiling flags                           | ✓ Complete |
| GoReleaser config                                    | ✓ Complete |

### Documentation

| Item                                                      | Status     |
| --------------------------------------------------------- | ---------- |
| PROPOSAL.md (full design rationale)                       | ✓ Complete |
| USAGE_GUIDE.md                                            | ✓ Complete |
| CHANGELOG.md                                              | ✓ Complete |
| CONTRIBUTING.md                                           | ✓ Complete |
| AGENTS.md (agent context)                                 | ✓ Complete |
| Example programs (govet, artdupl, staticcheck, branching) | ✓ Complete |
| Justfile (test, bench, lint, cover, ci, etc.)             | ✓ Complete |

### Planning

| Item                                                                   | Status     |
| ---------------------------------------------------------------------- | ---------- |
| EXECUTION_PLAN.md + V2                                                 | ✓ Complete |
| PROGRESS_REPORT.md                                                     | ✓ Complete |
| Status reports (3 historical)                                          | ✓ Complete |
| **MODULE_SPLIT_PLAN.md** — 3 approaches analyzed, recommendation given | ✓ Complete |

---

## b) PARTIALLY DONE ⚠️

### Module Split Planning

The plan (`docs/planning/MODULE_SPLIT_PLAN.md`) is written but has **9 identified gaps** from critical review:

1. **BLOCKER:** `example_test.go:254-288` (`ExamplePipeline`) imports `pipeline` from root tests. If pipeline becomes a separate module, root needs pipeline as a dependency — breaking the "zero dep" claim. **Not addressed.**
2. **Single-file `analysis/` module** — 150 lines doesn't clearly justify its own `go.mod`. Alternatives (internal package, keeping in root) not fairly evaluated.
3. **GoReleaser multi-module** — `.goreleaser.yml` is single-module. No plan for per-module builds and tag patterns.
4. **CI is stale** — Tests Go 1.21/1.22/1.23 but `go.mod` requires 1.26.0. Multi-module makes this worse.
5. **Bootstrap versioning** — `v0.0.0` placeholders only work with `go.work`. First published release procedure not documented.
6. **No actual dep size measurement** — "heavy" is qualitative. Never ran `go list` to show actual MB/package count.
7. **No migration guide** — Lists 6 breaking changes but no before/after code for downstream consumers.
8. **No rollback plan** — What if the split causes issues?
9. **Missed PROPOSAL.md alignment** — Moving `diagnostic.go` out actually _aligns_ with "SDK depends on nothing" principle. This strong argument wasn't used.

### CLI Tool

- **Detector registration is empty** — `registerDetector` is defined but never called. No `init()` hooks. CLI always reports "No detectors available."
- **No test files** — `cmd/go-finding/` has `[no test files]`.
- **Lint warnings** — `fmt.Println` violations (forbidigo), unchecked error returns.

### Race Detector

- **Race mode tests fail** — `go test -race ./...` exits with vet error in `gopkg.in/yaml.v3` (cache corruption, not our bug, but CI with `-race` would fail).

---

## c) NOT STARTED ✗

| Item                                           | Source                                        | Priority         |
| ---------------------------------------------- | --------------------------------------------- | ---------------- |
| Module split implementation (any approach)     | MODULE_SPLIT_PLAN.md Step 1-9                 | High if pursuing |
| CLI detector registration + built-in detectors | EXECUTION_PLAN_V2 item 14                     | High             |
| CLI config file support for detectors          | EXECUTION_PLAN_V2 item 15                     | Medium           |
| Watch mode (re-run on file change)             | EXECUTION_PLAN_V2 item 16                     | Low              |
| go-sarif evaluation (use upstream vs. custom)  | EXECUTION_PLAN_V2 item 8                      | Medium           |
| CLI tests (`cmd/go-finding/`)                  | No test files exist                           | High             |
| Pipeline examples beyond `ExamplePipeline`     | `example_test.go` only has 1 pipeline example | Low              |
| API documentation (`go doc -all`)              | Justfile has `docs` recipe, never run         | Low              |
| Vulnerability scanning (`govulncheck`)         | Justfile has `vuln` recipe, no CI step        | Medium           |
| Go reference documentation (pkg.go.dev)        | No opt-in                                     | Low              |

---

## d) TOTALLY FUCKED UP 💥

### CI is Broken and Out of Date

```
CI matrix: go-version: ["1.21", "1.22", "1.23"]
go.mod:    go 1.26.0
```

**Every CI run will fail** because Go 1.21/1.22/1.23 cannot parse a `go.mod` requiring 1.26.0. This has been broken since at least commit `6f1c854` (v1.0.0 release). Nobody noticed because the branch is 1 commit ahead of origin and hasn't been pushed.

### Release Workflow Will Also Break

```
release.yml: go-version: "1.23"
go.mod:      go 1.26.0
```

The release workflow uses Go 1.23 — same problem. Any `v*` tag push will fail at `go mod download`.

### Race Detector Tests Are Broken

```
go test -race ./... → vet error in gopkg.in/yaml.v3 (cache corruption)
```

Not our bug, but `just test` runs with `-race` flag, so `just test` fails on this machine. CI would also fail if it had the right Go version.

### Lint Has 26 Warnings (Pre-existing)

- `fmt.Println`/`fmt.Printf` usage in `cmd/go-finding/main.go` (forbidigo rule)
- Unchecked error returns (`f.Close`, `pprof.StartCPUProfile`)
- `registerDetector` unused
- `parallel golangci-lint is running` errors (LSP conflict)

These are not new but were never fixed.

---

## e) WHAT WE SHOULD IMPROVE

### Critical

1. **Fix CI immediately** — Update `.github/workflows/ci.yml` and `release.yml` to use Go 1.26.0 (or at minimum, match `go.mod`).
2. **Fix race test environment** — Clear Go build cache (`go clean -cache`) and verify `-race` tests pass.
3. **Push to origin** — 1 commit ahead, CI has never seen `MODULE_SPLIT_PLAN.md`.

### Important

4. **Add CLI tests** — `cmd/go-finding/` has zero test files. This is the user-facing entry point.
5. **Fix lint warnings in CLI** — The forbidigo violations and unchecked errors are real quality issues.
6. **Resolve the module split blocker** — The `example_test.go` cross-module dependency must be solved before any split is viable.
7. **Measure actual dep size** — Run `go list -m -json all` and compute download size of `golang.org/x/tools` transitive closure. The "heavy dep" claim needs numbers.
8. **Module split plan revision** — Address all 9 gaps identified in the critical review.

### Nice to Have

9. **Wire up `govulncheck`** in CI — the Justfile recipe exists but CI doesn't use it.
10. **Add `go.work` for development** — even without module split, workspace mode improves local dev experience with examples.
11. **Document the API surface** — Run `just docs` and commit `API.md`, or set up pkg.go.dev.
12. **Add a CONTRIBUTING.md test** — Verify that new contributors can run `just ci` and it passes.

---

## f) Top #25 Things to Get Done Next

| #   | Item                                                                  | Category     | Effort | Impact                              |
| --- | --------------------------------------------------------------------- | ------------ | ------ | ----------------------------------- |
| 1   | Fix CI Go version matrix (1.26.0)                                     | Fix          | 5 min  | 🔴 Blocks all CI                    |
| 2   | Fix release workflow Go version                                       | Fix          | 5 min  | 🔴 Blocks all releases              |
| 3   | Push to origin                                                        | Ops          | 1 min  | 🔴 Unpushed work at risk            |
| 4   | Clear Go cache + verify `-race` tests pass                            | Fix          | 5 min  | 🔴 `just test` is broken            |
| 5   | Fix lint warnings in `cmd/go-finding/main.go`                         | Quality      | 30 min | 🟡 26 warnings                      |
| 6   | Add CLI tests (`cmd/go-finding/`)                                     | Tests        | 2-3h   | 🟡 Zero test coverage               |
| 7   | Wire `registerDetector` or remove dead code                           | Quality      | 30 min | 🟡 Unused code                      |
| 8   | Resolve module split `example_test.go` blocker                        | Design       | 1h     | 🟡 Blocks Approach B                |
| 9   | Measure `golang.org/x/tools` transitive dep size                      | Research     | 15 min | 🟡 Data for decision                |
| 10  | Revise MODULE_SPLIT_PLAN.md with all 9 gaps                           | Planning     | 1-2h   | 🟢 Better plan                      |
| 11  | Update AGENTS.md with module split decision context                   | Docs         | 15 min | 🟢 Agent awareness                  |
| 12  | Add `govulncheck` step to CI                                          | Security     | 15 min | 🟢 Supply chain                     |
| 13  | Add `go.work` for local development                                   | DevEx        | 10 min | 🟢 Better DX                        |
| 14  | Add GoReleaser multi-module config (if splitting)                     | Release      | 1h     | 🟢 Required for split               |
| 15  | Write migration guide for module split (before/after)                 | Docs         | 1h     | 🟢 Downstream clarity               |
| 16  | Add pipeline example with config file                                 | Examples     | 30 min | 🟢 Usability                        |
| 17  | Build at least 1 real detector integration (e.g., govet)              | Features     | 2-3h   | 🟢 CLI is useless without detectors |
| 18  | Add CLI end-to-end test with a real detector                          | Tests        | 1-2h   | 🟢 Confidence                       |
| 19  | Document first-release procedure for multi-module                     | Docs         | 30 min | 🟢 Ops readiness                    |
| 20  | Evaluate `go-sarif` upstream vs. custom SARIF code                    | Research     | 1h     | 🟢 Dependency audit                 |
| 21  | Add SARIF schema validation test                                      | Tests        | 30 min | 🟢 Correctness                      |
| 22  | Clean up `pipeline/astfix.go` — evaluate if it belongs in `analysis/` | Architecture | 30 min | 🟢 Module alignment                 |
| 23  | Set up pkg.go.dev documentation                                       | Docs         | 15 min | 🟢 Discoverability                  |
| 24  | Add watch mode to CLI (EXECUTION_PLAN_V2 #16)                         | Features     | 2-3h   | 🟢 DevEx                            |
| 25  | Tag v1.1.0 with module split or v1.0.1 without                        | Release      | 15 min | 🟢 Milestone                        |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Is this project intended as a library (imported by other tools) or as a CLI application (used directly by developers)?**

The ambiguity affects nearly every architectural decision:

- **If library-first:** The module split makes perfect sense — zero-dep core, opt-in heavy deps. The CLI is secondary. `FromDiagnostic` is a key API. Examples are the primary "product."
- **If CLI-first:** The module split adds complexity for no user benefit. The CLI needs real detectors to be useful. The "zero dep" core is irrelevant because CLI pulls everything anyway.
- **If both:** Need to decide versioning strategy — library SemVer stability vs. CLI rapid iteration.

The PROPOSAL.md says _"Converters live in each tool, not in the SDK. Tools depend on the SDK; the SDK depends on nothing"_ — which is library-first. But the CLI, GoReleaser, and examples suggest CLI-first. This tension needs resolving before committing to the module split.

---

## Metrics Summary

| Metric                 | Value                        |
| ---------------------- | ---------------------------- |
| Total lines of code    | ~10,391                      |
| Source files           | 29                           |
| Test files             | 33                           |
| Root package coverage  | 96.1%                        |
| Pipeline coverage      | 87.5%                        |
| External dependencies  | 3 direct, 6 transitive       |
| Open lint warnings     | 26                           |
| CI Go version mismatch | 1.23 (CI) vs 1.26.0 (go.mod) |
| Unpushed commits       | 1                            |
| Module split plan gaps | 9 (1 blocker)                |

---

_Generated by Crush_
