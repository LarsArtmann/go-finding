# Comprehensive Status Update — go-finding

**Date:** 2026-06-17 01:00 CEST  
**Branch:** master  
**Ahead of origin:** 0 (pushed)  
**Commit range since last sync:** `e9b5d83..9e67a92` (8 commits)

---

## Executive Summary

The Session 16 Top 25 TODO list is **complete and pushed**. Every planned item has been implemented, tested, linted, race-detected, and committed. Code quality is high, coverage exceeds targets, and the v0.7.0 release is now a documentation/decision problem, not an engineering problem.

Three **owner-decision blockers** remain unresolved because they require product/domain judgment, not code. They are documented in `docs/RELEASE_CRITERIA.md` and `docs/MIGRATION_v1.0.md`.

Two pre-existing status-file modifications were **left unstaged** because they were not authored in this session and should not be committed without explicit ownership.

---

## a) FULLY DONE

### Session 16 Top 25 TODO List — 100% Complete

All 23 tracked TODO items from `docs/status/2026-06-16_23-31_session16-ghost-system-cleanup-and-split-brain-fix.md` are done.

| #   | Item                                                          | Status | Evidence                                                                     |
| --- | ------------------------------------------------------------- | ------ | ---------------------------------------------------------------------------- |
| 1   | `Category.Compare()` method                                   | Done   | `category.go`, `category_test.go`                                            |
| 2   | Fuzz `CategoryForLinter`                                      | Done   | `category_linter_test.go` + seed corpus                                      |
| 3   | Dependabot config                                             | Done   | `.github/dependabot.yml`                                                     |
| 4   | `slices.Sorted` modernization                                 | Done   | `correlate.go`, `partial.go`, `fix_applier.go`, `cmd/go-finding/*.go`        |
| 5   | Godoc examples for IntervalIndex, MergeIter, DetectorRegistry | Done   | `example_extra_test.go`                                                      |
| 6   | Column-aware `SubstringProvider`                              | Done   | `pipeline/fix_provider.go`                                                   |
| 7   | `ShiftedPosition` / `ShiftedRange` / `ByteDelta`              | Done   | `pipeline/line_shift.go`, `pipeline/line_shift_test.go`                      |
| 8   | `pipeline_detect.go` uses shifted position/range              | Done   | `pipeline/pipeline_detect.go`                                                |
| 9   | ConfigFile integration tests                                  | Done   | `pipeline/integration_endtoend_test.go`                                      |
| 10  | DetectorRegistry integration tests                            | Done   | `pipeline/integration_endtoend_test.go`                                      |
| 11  | Pipeline type-alias coverage                                  | Done   | `pipeline/integration_endtoend_test.go`                                      |
| 12  | Coverage for ConfigFile methods                               | Done   | `pipeline/coverage_config_file_test.go`                                      |
| 13  | Coverage for `ApplyWithDetails`                               | Done   | `pipeline/coverage_config_file_test.go`                                      |
| 14  | Coverage for `pickNearestOccurrence` branches                 | Done   | `pipeline/coverage_config_file_test.go`                                      |
| 15  | Godoc examples for ConfigFile, LineShiftMap, DetectorRegistry | Done   | `pipeline/example_test.go`                                                   |
| 16  | CLI fix provider error surfacing                              | Done   | `cmd/go-finding/fix_provider_registry.go`, `cmd/go-finding/main.go`          |
| 17  | CLI coverage to 90%+                                          | Done   | `cmd/go-finding/coverage_extra_test.go`, `cmd/go-finding/main_extra_test.go` |
| 18  | `slices.Sorted` in CLI                                        | Done   | `cmd/go-finding/config.go`, `generated_filter.go`, `registry.go`             |
| 19  | `config.example.yaml` fix provider examples                   | Done   | `cmd/go-finding/config.example.yaml`                                         |
| 20  | CI benchmark regression gate                                  | Done   | `.github/workflows/ci.yml`, `scripts/bench-check.sh`                         |
| 21  | v0.7.0+ documentation refresh                                 | Done   | `doc.go`, `FEATURES.md`, `README.md`, `AGENTS.md`                            |
| 22  | Release criteria document                                     | Done   | `docs/RELEASE_CRITERIA.md`                                                   |
| 23  | v1.0 migration guide                                          | Done   | `docs/MIGRATION_v1.0.md`                                                     |

### Quality Gates — All Green

| Gate                                   | Result                                   |
| -------------------------------------- | ---------------------------------------- |
| `go test -race -count=1 ./...`         | PASS all packages                        |
| `golangci-lint run --timeout=5m ./...` | 0 issues                                 |
| `go vet ./...`                         | PASS                                     |
| `go build ./...`                       | PASS                                     |
| BuildFlow pre-commit (Go)              | 34/34 checks pass                        |
| BuildFlow pre-commit (Nix)             | 21/21 checks pass (after vendorHash fix) |

### Coverage — Targets Met or Exceeded

| Package                             | Coverage | Target                  |
| ----------------------------------- | -------- | ----------------------- |
| `github.com/larsartmann/go-finding` | 93.7%    | 90%+                    |
| `analysis`                          | 94.1%    | 90%+                    |
| `cmd/go-finding`                    | 91.2%    | 90%+                    |
| `internal/detectors`                | 96.1%    | 90%+                    |
| `internal/gotoken`                  | 92.7%    | 90%+                    |
| `pipeline`                          | 95.2%    | 93%+                    |
| `pipeline/goast`                    | 80.8%    | new package, acceptable |

### Commits Pushed to `origin/master`

```
9e67a92 docs: Session 17 status report
3840d4c nix: update vendorHash after dependency bump
ccb994e docs: v0.7.0+ documentation, release criteria, and migration guide
92c547a deps: update ginkgo, gomega, and pprof
c934894 ci: add dependabot and benchmark regression gate
7132d32 fix(cli): surface fix provider errors and raise CLI coverage to 91.2%
f5805c9 test: pipeline integration tests for ConfigFile, DetectorRegistry, and coverage gaps
af1332a feat: extend LineShiftMap to positions and ranges; column-aware SubstringProvider
563d3d4 feat: Category.Compare, fuzz CategoryForLinter, IntervalIndex/MergeIter/DetectorRegistry examples
```

---

## b) PARTIALLY DONE

### v0.7.0 Release Prep

- **Code:** ready.
- **Docs:** ready.
- **Release criteria:** drafted.
- **Blockers:** 3 owner-decision items documented but not resolved.

### `pipeline/goast` Coverage

- Package works and is integrated, but coverage is 80.8%. Parse-failure fallback paths and some edge cases are not fully exercised. This is acceptable for a new opt-in provider but should be raised before v1.0.

### Examples

- `examples/basic`, `examples/builder`, `examples/pipeline` report 0% coverage because they are runnable examples without tests. They compile and run via `TestExamplesRun`, but coverage tooling does not attribute this to the example packages.

---

## c) NOT STARTED

### v1.0.0 Breaking Changes

These are **intentionally not started** until v0.7.0 is tagged and a migration window has passed:

1. Unexport `Report.Findings` (replace with `FindingsSnapshot()` / accessors).
2. Remove deprecated `Report.Merge()` (use `MergeInto`).
3. Remove deprecated `RecordFix()` (use `RecordFixes`).
4. Resolve `Position.Offset` zero-value ambiguity (sentinel vs. byte 0).
5. Normalize `FixStrategy ""` vs. `FixStrategyNone`.
6. Keep SARIF types unexported (already done; maintain in v1.0).

### Additional Detectors

- staticcheck detector exists; govet exists.
- No additional first-party detectors (e.g., `go vet` shadow analyzer, custom linter) are planned for v0.7.0.

### IDE/LSP Server

- LSP diagnostic conversion exists.
- No standalone LSP server executable exists.

### Web UI / Reporting Dashboard

- Text, markdown, JSON, SARIF output exist.
- No HTML report UI or dashboard exists.

---

## d) TOTALLY FUCKED UP

### Nothing is totally fucked up.

The build is green, tests pass, lint is clean, coverage is high, and the branch is synced with origin. The closest thing to "fucked up" is the three long-standing semantic ambiguities that keep getting punted because they are **product decisions masquerading as code problems**:

1. `Position.Offset = 0` means both "byte 0" and "no offset provided" depending on context.
2. `FixStrategy ""` and `FixStrategyNone` coexist without canonical normalization.
3. `Report.Findings` is public for backward compat but breaks encapsulation.

These are not bugs; they are design debt that needs an owner decision.

---

## e) WHAT WE SHOULD IMPROVE

### 1. Close the v1.0 Blockers with Owner Decisions

The three blockers above are the only thing preventing a clean v1.0 roadmap. Pick a direction, document it, and implement the migration path.

### 2. Raise `pipeline/goast` Coverage to 90%+

Focus on:

- Parse failure fallback paths.
- Multi-occurrence disambiguation with identical BeforeCode.
- Cache hit/miss paths for the FNV content-hash cache.

### 3. Add Property-Based Tests for Fix Providers

The new `SubstringProvider` heuristic and `GoASTProvider` would benefit from `testing/quick` or fuzz targets that generate random source + findings and verify the provider returns valid edits.

### 4. Reduce BuildFlow Noise

`go-structure-linter` reports 39 issues, mostly false positives about root-package files and compiled binaries (`go-finding`, `result`). These are known and ignored, but they add noise to every commit. Either fix the structure linter config or move the files.

### 5. Add E2E CLI Test for Fix Provider Registration

We test `resolveFixProviders` in unit tests, but there is no end-to-end test that runs the CLI with `-fix-provider go-ast` and verifies the provider is actually used.

### 6. Document the Public API Surface More Explicitly

`docs/API_STABILITY.md` exists but should be regenerated after the v0.7.0 changes to ensure every new symbol (`Category.Compare`, `LineShiftMap` extensions, `GoASTProvider`) is classified.

### 7. Evaluate Whether `FixStrategyAI` Should Stay Reserved

It is currently documented as "reserved for future AI-powered remediation." Decide if this is still credible or if it should be removed before v1.0.

### 8. Commit or Revert the Pre-existing Status File Modifications

`docs/status/2026-06-15_08-48_SESSION-15-DEDUP-ELIMINATION-STATUS.md` and `docs/status/2026-06-16_23-31_session16-ghost-system-cleanup-and-split-brain-fix.md` are still modified in the working tree. They were not authored in this session and were intentionally left unstaged. An owner should review and commit or revert them.

---

## f) Top #25 Things We Should Get Done Next

### Release & Roadmap

1. **Tag v0.7.0** — code and docs are ready.
2. **Resolve Blocker 1:** Decide `Position.Offset` zero-value semantics.
3. **Resolve Blocker 2:** Decide `FixStrategy ""` vs. `FixStrategyNone` normalization.
4. **Resolve Blocker 3:** Decide `Report.Findings` unexport timing and accessor design.
5. **Update `API_STABILITY.md`** for v0.7.0 symbols.
6. **Write CHANGELOG v0.7.0 entry** if not already complete (verify).
7. **Draft v0.8.0 / v1.0.0 roadmap** based on migration guide.

### Testing & Quality

8. **Raise `pipeline/goast` coverage to 90%+** with targeted tests.
9. **Add fuzz target for `GoASTProvider.Edits`**.
10. **Add property tests for `pickNearestOccurrence`**.
11. **Add E2E CLI test for `-fix-provider go-ast`**.
12. **Run `go test -race -count=100` stress run** on pipeline package.
13. **Add benchmark baseline for `GoASTProvider`** and store in `benchmarks/`.

### Code & Architecture

14. **Implement v1.0 migration path** for `Report.Findings` (accessors, deprecate field).
15. **Remove `Report.Merge()` and `RecordFix()`** in v1.0 branch.
16. **Normalize `FixStrategy`** to a single zero-value representation.
17. **Resolve `Position.Offset` sentinel** (e.g., `-1` for unspecified).
18. **Extract root-package files** if `go-structure-linter` warnings are actionable.
19. **Review `FixStrategyAI`** — keep reserved or delete.
20. **Add `context.Context` to more pipeline I/O boundaries** if missing.

### Docs & Developer Experience

21. **Update `AGENTS.md`** with v0.7.0 decisions and command references.
22. **Clean up `docs/status/`** — commit or revert the two modified status files.
23. **Add a "Writing a Fix Provider" recipe** to `docs/guides/`.
24. **Add architecture decision record** for v0.7.0 release blockers.
25. **Improve README quickstart** with a real `go vet` + fix pipeline example.

---

## g) Top #1 Question I Cannot Figure Out Myself

> **When do we actually pull the trigger on unexporting `Report.Findings`?**

Everything else is either a clear code change or a clear product decision. But this one sits in the uncomfortable middle:

- **Unexport now (v0.7.x):** It is the "right" architecture — the field is already deprecated, `FindingsSnapshot()` exists, and internal code uses `findingsLocked()`. But it is a breaking change that could hurt early adopters who directly mutate `report.Findings`.
- **Wait for v1.0:** It honors semver and gives users a migration window. But it prolongs the encapsulation risk and means we keep carrying the deprecated API.
- **Hybrid:** Keep the field exported but panic on direct mutation via a setter trap. That is clever but probably worse than either clean option.

The engineering work is trivial either way. The decision is **risk tolerance for breaking current users vs. speed of paying down architectural debt**. I cannot make that call without knowing the actual consumer base and stability promises. That is the one question that needs an owner answer before v1.0 planning can proceed.

---

## Appendix: Current Working Tree

```
On branch master
Your branch is up to date with 'origin/master'.

Changes not staged for commit:
  modified:   docs/status/2026-06-15_08-48_SESSION-15-DEDUP-ELIMINATION-STATUS.md
  modified:   docs/status/2026-06-16_23-31_session16-ghost-system-cleanup-and-split-brain-fix.md

no changes added to commit
```

These modifications were present at session start and were intentionally not committed because they were not authored in this session.

---

_Assisted-by: Crush <crush@charm.land>_
