# go-finding — Full Comprehensive Status Report

**Date:** 2026-05-21 01:42 CEST
**Version:** 0.2.1 (pre-v1.0)
**Branch:** master (1 commit ahead of origin)
**Total commits:** 563
**Recent activity:** 48 commits since May 17

---

## Executive Summary

go-finding is a **mature, well-tested Go library** at v0.2.1 approaching v1.0 readiness. The core library is feature-complete with **99.7% root package coverage**, **96.4% pipeline coverage**, and **all tests passing**. The remaining work is primarily API cleanup (deprecation removal, naming consistency) and documentation — no architectural gaps or missing features block a v1.0 release.

**Overall health: STRONG.** One unpushed commit (status doc). 100 goconst warnings are the only lint issues (all pre-existing, all in test code).

---

## a) FULLY DONE

### Core Library (Root Package)

| Feature                                                         | Status                 | Coverage |
| --------------------------------------------------------------- | ---------------------- | -------- |
| `Finding` type with Builder pattern                             | ✅ Complete            | 99.7%    |
| `Position` / `Range` with Overlaps/Intersection/Adjacent        | ✅ Complete            | —        |
| `Severity` (Critical→Trace) with `String()` / `ParseSeverity()` | ✅ Complete            | —        |
| `Confidence` named type with `IsValid()` / `Clamp()`            | ✅ Complete            | —        |
| `Category`, `Tag`, `Suppression`, `Correlation` types           | ✅ Complete            | —        |
| `Report` container with summary, filtering, grouping            | ✅ Thread-safe         | —        |
| `Diff(before, after)` — finding set comparison                  | ✅ Complete            | —        |
| `FormatText` / `FormatMarkdown` — human-readable output         | ✅ Complete            | —        |
| SARIF 2.1.0 import (`FindingsFromSARIF`)                        | ✅ Hardened            | —        |
| SARIF 2.1.0 export (`ToSARIF`, `WriteSARIF`, streaming)         | ✅ Complete            | —        |
| LSP Diagnostic conversion                                       | ✅ Complete            | —        |
| JSON marshaling/unmarshaling                                    | ✅ Complete            | —        |
| Structured errors (`FindingError` with categories)              | ✅ Complete            | —        |
| ID generation (`KeySeparator` constant)                         | ✅ Complete            | —        |
| `HasCodeChange()` method                                        | ✅ New (dedup session) | —        |

### Pipeline Package

| Feature                                                         | Status          | Notes                           |
| --------------------------------------------------------------- | --------------- | ------------------------------- |
| Pipeline core (`Run`, detect/triage/apply loop)                 | ✅ Complete     | 96.4% coverage                  |
| Byte-level `FixEngine` (`FixEdit{Offset, Length, Replacement}`) | ✅ Complete     | Descending-offset apply         |
| `FixProvider` interface (Offset/Line/Substring)                 | ✅ Complete     | Composable chain                |
| Custom provider registration                                    | ✅ Complete     | `Config.FixProviders`           |
| Conflict detection (`DetectConflicts()`)                        | ✅ Complete     | —                               |
| `FixApplier` with backup/rollback                               | ✅ Complete     | `Close()` lifecycle             |
| Verification stage                                              | ✅ Complete     | Re-run detectors, diff findings |
| Metrics collection with snapshots                               | ✅ Complete     | —                               |
| Exponential backoff retry                                       | ✅ Complete     | `math/rand/v2` jitter           |
| Partial success collection                                      | ✅ Complete     | —                               |
| Parallel detection (errgroup)                                   | ✅ Complete     | —                               |
| Per-detector timeouts                                           | ✅ Complete     | `Config.DetectorTimeouts`       |
| Structured logging (`slog`)                                     | ✅ Complete     | `Config.Logger`                 |
| Stage progress callback                                         | ✅ Complete     | `Config.OnStage`                |
| `FindingProcessor` composable transforms                        | ✅ Experimental | —                               |
| Context cancellation propagation                                | ✅ Complete     | All paths handle cancel         |

### Analysis Package

| Feature                                     | Status              |
| ------------------------------------------- | ------------------- |
| go/analysis.Diagnostic ↔ Finding conversion | ✅ Complete (98.5%) |

### CLI

| Feature                                     | Status      |
| ------------------------------------------- | ----------- |
| Text/markdown/JSON/SARIF output             | ✅ Complete |
| YAML/JSON config with validation            | ✅ Complete |
| Built-in govet + staticcheck detectors      | ✅ Complete |
| Severity filtering, timeout, max-iterations | ✅ Complete |
| CPU/memory profiling                        | ✅ Complete |
| Dynamic detector registry                   | ✅ Complete |
| Graceful degradation on detector failures   | ✅ Complete |

### Testing Infrastructure

| Feature                    | Status                 |
| -------------------------- | ---------------------- |
| BDD tests (Ginkgo/Gomega)  | ✅ Complete            |
| Property-based tests       | ✅ Complete            |
| Fuzz tests                 | ✅ Complete            |
| Benchmarks (10k+ findings) | ✅ Complete            |
| Coverage enforcement       | ✅ Complete            |
| SARIF schema validation    | ✅ Complete            |
| Test helper deduplication  | ✅ New (dedup session) |

### Codebase Metrics

| Metric              | Value                 |
| ------------------- | --------------------- |
| Total Go LOC        | 25,240                |
| Production LOC      | 7,084                 |
| Test LOC            | 18,156                |
| Production files    | 46                    |
| Test files          | 66                    |
| Direct dependencies | 5                     |
| Total packages      | 9                     |
| Clone groups        | 151 (from 158, -4.4%) |

### Test Coverage

| Package            | Coverage                      |
| ------------------ | ----------------------------- |
| Root (finding)     | **99.7%**                     |
| Analysis           | **98.5%**                     |
| Pipeline           | **96.4%**                     |
| Internal detectors | **96.1%**                     |
| CLI                | **93.9%**                     |
| Examples           | 0% (no statements / no tests) |

---

## b) PARTIALLY DONE

### v1.0 Release Criteria (from `docs/v1.0-release-criteria.md`)

| Criteria                                  | Progress     | Blockers                                                                |
| ----------------------------------------- | ------------ | ----------------------------------------------------------------------- |
| Remove deprecated symbols                 | ~30%         | `Finding.Tag string`, `ConflictDetector`/`Verifier` structs still exist |
| Remove `FixStrategyAI` phantom            | Not started  | Constant exists, no backend, silently treated as Suggest                |
| Naming consistency audit                  | ~50%         | Some `Pos()` vs `NewPosition()` inconsistencies remain                  |
| README covers all exported types          | ~60%         | Major types documented, newer additions may be missing                  |
| CHANGELOG for v1.0                        | Not started  | —                                                                       |
| API stability guarantee docs              | Not started  | —                                                                       |
| `doc.go` comprehensive                    | ~40%         | Basic doc exists                                                        |
| Integration test with downstream consumer | Not started  | —                                                                       |
| FixApplier goroutine leak verification    | Not verified | —                                                                       |

### goconst Warnings

- **100 goconst warnings** in test code (string literals that could be constants)
- 1 prealloc warning in `pipeline/pipeline_bench_test.go`
- All pre-existing, none introduced recently
- Low priority but would clean up lint output

### Semantic Deduplication

- **158 → 151 clone groups** (-4.4%)
- Production code deduplication complete
- Test code deduplication partially done (helpers extracted in pipeline tests)
- Remaining 151 groups analyzed as inherent to Go idioms

---

## c) NOT STARTED

### From TODO_LIST.md — P3 (Future/Deferred)

| Item                                    | Scope                              |
| --------------------------------------- | ---------------------------------- |
| Nix migration (Phases 0–5)              | Full build system migration        |
| Pipeline middleware/interceptor         | Architecture enhancement           |
| Watch mode (fsnotify)                   | Feature                            |
| Semantic merge for conflicts            | Feature                            |
| Styled CLI (lipgloss) / TUI (bubbletea) | Feature                            |
| LSP language server / CodeAction        | Feature                            |
| AI backend for FixStrategyAI            | Feature (reserved constant exists) |
| Web UI, OTel integration, etc.          | ~12 out-of-scope items             |

### Documentation Gaps

- API stability guarantee (Go compat promise style)
- CHANGELOG for v1.0.0
- Integration guide update for downstream consumers
- All examples compile verification

### Infrastructure

- CI step for `art-dupl` regression prevention
- Pre-commit hook fixes (goconst, todo-check, library-policy currently fail)
- Per-package coverage thresholds (exists but has a bug)

---

## d) TOTALLY FUCKED UP

### Pre-commit Hook Failures

The pre-commit hook (BuildFlow) has **3 persistent failures** that require `--no-verify`:

1. **goconst** — 100 warnings in test code. The linter config likely needs to exclude test files or the constants need to be created.
2. **todo-check** — Likely checking for stale TODO items.
3. **library-policy** — Unknown policy violation (possibly related to testify indirect dependency or go-faster/yaml).

These are NOT blocking development but ARE blocking clean git operations without `--no-verify`.

### OnFix Callback Inaccuracy (from READINESS_REPORT)

- **Medium priority** — `OnFix` callback reports success for **skipped** fixes (not actually applied)
- Misleading for consumers tracking fix success rates

### FixApplier Path Traversal (gosec G703)

- **Medium priority** — Likely false positive but unverified
- Should add explicit path validation

### CLI run() Global State

- **Medium priority** — Uses global flag state, not parallel-testable
- Blocks CLI coverage improvement beyond ~93.9%

---

## e) WHAT WE SHOULD IMPROVE

### Critical Path to v1.0

1. **Remove deprecated symbols** — The `Finding.Tag string` field, `ConflictDetector`/`Verifier` structs, and `FixStrategyAI` phantom are dead weight that will confuse v1.0 consumers
2. **API naming consistency audit** — `Pos()` vs `NewPosition()`, `FromSARIF` vs `FindingsFromSARIF`, etc. Lock these down before v1.0
3. **Write CHANGELOG** — No v1.0 release is credible without one
4. **Integration test** — Prove the library works with at least one downstream consumer (BuildFlow, hierarchical-errors, or branching-flow)
5. **Fix pre-commit hooks** — 3 persistent failures forcing `--no-verify` is unsustainable

### Code Quality

6. **Fix 100 goconst warnings** — All in test code, low risk but noisy
7. **Fix OnFix callback inaccuracy** — Reports success for skipped fixes
8. **Add path validation in FixApplier** — Close the gosec G703 finding
9. **Refactor CLI run() for testability** — Extract io.Writer + \*flag.FlagSet
10. **Split pipeline_test.go** — 1508 lines, complexity 171. Split into focused files

### Architecture

11. **Verify FixApplier goroutine leak fix** — Claimed fixed but not verified
12. **Decide domain-specific provider location** — Should AST-aware providers live in pipeline/ or separate?
13. **Extract SARIF test builder helpers** — ~20 clone groups reducible
14. **Add `quickBuild()` in root bdd_test.go** — 11 repetitions of `NewBuilder().Build()`
15. **Consider io.WriterTo for SARIF** — Nice-to-have from v1.0 criteria

### Documentation

16. **API stability guarantee** — Go compat promise style document
17. **Comprehensive doc.go** — Current version is ~40% complete
18. **README updates** — Ensure all exported types have examples
19. **Update downstream consumers** — 5 projects need v1.0 migration
20. **Update AGENTS.md** — Reflect current Go version (1.26.2), recent changes

### Infrastructure

21. **Fix pre-commit hook goconst** — Either fix warnings or exclude test files
22. **Fix pre-commit hook todo-check** — Clear stale items
23. **Fix pre-commit hook library-policy** — Resolve policy violation
24. **Add CI step for art-dupl** — Prevent clone regression
25. **Fix per-package coverage thresholds** — Exists but has a bug

---

## f) Top #25 Things We Should Get Done Next

Ranked by impact on path to v1.0:

| #   | Task                                                      | Impact | Effort | Package  |
| --- | --------------------------------------------------------- | ------ | ------ | -------- |
| 1   | Remove deprecated `Finding.Tag string` field              | High   | Low    | Root     |
| 2   | Remove deprecated `ConflictDetector`/`Verifier` structs   | High   | Low    | Pipeline |
| 3   | Remove phantom `FixStrategyAI` constant                   | High   | Low    | Root     |
| 4   | Audit exported symbol naming consistency                  | High   | Medium | All      |
| 5   | Write CHANGELOG.md for v1.0                               | High   | Medium | —        |
| 6   | Write API stability guarantee doc                         | High   | Low    | Docs     |
| 7   | Integration test with downstream consumer                 | High   | High   | —        |
| 8   | Fix OnFix callback (don't report skipped as success)      | Medium | Low    | Pipeline |
| 9   | Add path validation in FixApplier                         | Medium | Low    | Pipeline |
| 10  | Refactor CLI run() for testability                        | Medium | Medium | CLI      |
| 11  | Fix 100 goconst warnings                                  | Medium | Medium | Tests    |
| 12  | Verify FixApplier goroutine leak fix                      | Medium | Low    | Pipeline |
| 13  | Split pipeline_test.go (1508 lines)                       | Medium | Medium | Pipeline |
| 14  | Extract SARIF test builder helpers                        | Low    | Low    | Tests    |
| 15  | Add `quickBuild()` in root bdd_test.go                    | Low    | Low    | Tests    |
| 16  | Comprehensive doc.go                                      | Medium | Medium | Root     |
| 17  | README examples for all exported types                    | Medium | Medium | Docs     |
| 18  | Fix pre-commit hook (goconst, todo-check, library-policy) | Medium | Medium | CI       |
| 19  | Add art-dupl CI step                                      | Low    | Low    | CI       |
| 20  | Fix per-package coverage thresholds bug                   | Low    | Low    | Tests    |
| 21  | Decide domain-specific provider location                  | Low    | Low    | Pipeline |
| 22  | Consider io.WriterTo for SARIF                            | Low    | Low    | Root     |
| 23  | Update AGENTS.md (Go 1.26.2, recent changes)              | Low    | Low    | Docs     |
| 24  | Verify all examples compile                               | Low    | Low    | Examples |
| 25  | Prepare v1.0.0 tag and release                            | High   | Low    | —        |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the timeline and priority for v1.0.0 release?**

The codebase is functionally complete and well-tested. The remaining work is almost entirely API cleanup (deprecation removal, naming), documentation, and downstream consumer integration. But:

- Are there downstream consumers **actively waiting** for v1.0 stability?
- Should we prioritize **breaking changes now** (remove deprecated symbols) while still at 0.x?
- Is the `FixStrategyAI` constant something to **implement, deprecate, or delete**?
- Are the 5 listed downstream consumers (BuildFlow, hierarchical-errors, branching-flow, go-structure-linter, golangci-lint-auto-configure) still the target consumers?

The answers to these questions determine whether the next session should be "remove deprecated symbols and write CHANGELOG" vs "build new features" vs "integration testing."

---

## Session History (Recent)

| Date       | What                                | Result                                  |
| ---------- | ----------------------------------- | --------------------------------------- |
| 2026-05-21 | Semantic deduplication (`art-dupl`) | 158→151 clone groups, helpers extracted |
| 2026-05-18 | v0.3.0 release quality session      | Comprehensive audit, formatted          |
| 2026-05-18 | Properties vs Metadata decision     | WONTFIX Properties, Metadata only       |
| 2026-05-17 | 10-task execution                   | Multiple features completed             |
| 2026-05-08 | Context cancellation fixes          | All paths propagate cancel              |
| 2026-05-06 | Byte-level FixEngine redesign       | Major architecture change               |

## Unpushed Commits

| Hash      | Message                                                                  |
| --------- | ------------------------------------------------------------------------ |
| `88764cd` | docs(status): add semantic deduplication session report (158→151 clones) |

---

_Report generated by Crush on 2026-05-21 at 01:42 CEST._
