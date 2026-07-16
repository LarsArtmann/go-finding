# Status Report: Test Defragmentation

> **📦 RESOLUTION STATUS (updated 2026-07-16)**
>
> Test defragmentation committed in v1.2.0 (`f48b0fa`). Zero `_extra`/`_bugfix`/`coverage` test
> files remain. The naming convention is now enforced via the AGENTS.md "Test Organization" section.
> The remaining multi-file splits (SARIF 7 files, Pipeline 7 files, etc.) are considered legitimate
> decomposition by concern, not fragmentation.

**Date:** 2026-07-06 06:41  
**Session goal:** Diagnose and fix the "36.5K lines feels like a lot" problem  
**Verifier:** `go test -race -count=1 ./...` — all 4 modules green

---

## a) FULLY DONE

### Root cause diagnosed

The 36.5K line count was **not** a production code problem. Production is lean: 11.5K lines across 78 files (~148 lines/file), largest file 369 lines, zero external deps in core. The bloat was **test fragmentation** — 125 test files for 78 production files, with a 2.1:1 test:production ratio inflated by disposable `_extra`/`_bugfix`/`coverage` suffix files.

### Test files defragmented (17 files eliminated)

| #   | File deleted                            | Merged into                    | Module   |
| --- | --------------------------------------- | ------------------------------ | -------- |
| 1   | `finding_extra_test.go`                 | `finding_test.go`              | core     |
| 2   | `position_extra_test.go`                | `position_test.go`             | core     |
| 3   | `coverage_extra_test.go`                | `coverage_test.go`             | core     |
| 4   | `example_extra_test.go`                 | `example_test.go`              | core     |
| 5   | `json_extra_test.go`                    | `json_test.go`                 | core     |
| 6   | `report_extra_test.go`                  | `report_test.go`               | core     |
| 7   | `bdd_extra_test.go`                     | `bdd_test.go`                  | core     |
| 8   | `assert_extra_test.go`                  | `testutil_test.go`             | core     |
| 9   | `example_basic_test.go`                 | `example_test.go`              | core     |
| 10  | `example_cli_test.go`                   | `example_test.go`              | core     |
| 11  | `pipeline/conflict_extra_test.go`       | `pipeline/conflict_test.go`    | pipeline |
| 12  | `pipeline/fix_applier_extra_test.go`    | `pipeline/fix_applier_test.go` | pipeline |
| 13  | `pipeline/fix_applier_bugfix_test.go`   | `pipeline/fix_applier_test.go` | pipeline |
| 14  | `pipeline/pipeline_bugfix_test.go`      | `pipeline/pipeline_test.go`    | pipeline |
| 15  | `pipeline/example_extra_test.go`        | `pipeline/example_test.go`     | pipeline |
| 16  | `cmd/go-finding/main_extra_test.go`     | `cmd/go-finding/main_test.go`  | CLI      |
| 17  | `cmd/go-finding/coverage_extra_test.go` | `cmd/go-finding/main_test.go`  | CLI      |

### Files renamed for honesty

| Old name                                | New name                       | Why                                                           |
| --------------------------------------- | ------------------------------ | ------------------------------------------------------------- |
| `coverage_test.go`                      | `validate_test.go`             | Name signals what it tests, not the metric it was written for |
| `pipeline/coverage_config_file_test.go` | `pipeline/config_file_test.go` | Same reason                                                   |

### Multi-subject file split

`pipeline/coverage_config_file_test.go` tested ConfigFile, FixApplier, and pickNearestOccurrence in one file. Split correctly:

- ConfigFile tests → `pipeline/config_file_test.go`
- FixApplier test → `pipeline/fix_applier_test.go`
- pickNearestOccurrence → `pipeline/fix_provider_integration_test.go`

### Import merging handled

Every merge where the `_extra` file had additional imports (`context`, `errors`, `os`, `math`, `log`, `path/filepath`, `runtime/pprof`, `flag`, `gogenfilter`, `time`) was handled by merging import blocks before appending.

### Duplicate symbols resolved

- `standardTestFinding` — only defined in `finding_test.go` (correct)
- `type failingWriter` — existed in both `main_test.go` and `main_extra_test.go`; kept parent definition, merged the method receiver from `_extra`
- `TestBDD` runner — only in `bdd_test.go`; `_extra` had no runner

### Quality gates passed

| Check                                          | Result   |
| ---------------------------------------------- | -------- |
| `go test -race -count=1 ./...` (all 4 modules) | PASS     |
| `gofmt -l .`                                   | 0 issues |
| `go vet ./...`                                 | 0 issues |
| `golangci-lint run ./...`                      | 0 issues |
| `go build ./...` (all 4 modules)               | PASS     |
| FDescribe/FIt committed focus scan             | 0 found  |
| `_extra`/`_bugfix`/`coverage` files remaining  | **0**    |

### AGENTS.md updated

Added "Test Organization" section with explicit rules: no `_extra` suffix files, no `_bugfix` suffix files, no `coverage_test.go`, shared helpers go in `testutil_test.go`, one example file per package.

### Metrics

```
Test files:    125 → 108  (17 eliminated, -13.6%)
Total lines:  36,516 → 36,372  (net -144 lines after merge overhead)
_extra/_bugfix/coverage fragment files:  15+ → 0
```

---

## b) PARTIALLY DONE

### Linter coverage incomplete

Ran `golangci-lint run ./...` on core module only (0 issues). Did NOT run linter on pipeline, CLI, or analysis modules. The `nix run .#lint` command from AGENTS.md was not attempted.

### GOWORK=off per-module isolation not verified

AGENTS.md documents `GOWORK=off go test ./...` as a per-module isolation test. Not run this session. The replace directives in each sub-module's `go.mod` are untested.

---

## c) NOT STARTED

### Remaining test file consolidation opportunities

The defragmentation eliminated `_extra`/`_bugfix` files, but many subjects still span multiple test files that could potentially merge:

| Subject   | File count | Candidates for merge                                                                                                                                                                              |
| --------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| SARIF     | 7 files    | `sarif_test.go` + `sarif_export_test.go` + `sarif_import_test.go` + `sarif_properties_test.go` + `sarif_roundtrip_test.go` + `sarif_suppression_test.go` + `sarif_fuzz_test.go`                   |
| Pipeline  | 7 files    | `pipeline_test.go` + `pipeline_detector_test.go` + `pipeline_fixapplier_test.go` + `pipeline_logging_test.go` + `pipeline_metrics_test.go` + `pipeline_triage_test.go` + `pipeline_bench_test.go` |
| FixEngine | 4 files    | `fix_engine_test.go` + `fix_engine_edit_test.go` + `fix_engine_fuzz_test.go` + `fix_engine_bench_test.go`                                                                                         |
| Merge     | 4 files    | `merge_test.go` + `merge_correlate_test.go` + `merge_dedup_test.go` + `merge_fuzz_test.go`                                                                                                        |
| LSP       | 4 files    | `lsp_test.go` + `lsp_to_test.go` + `lsp_roundtrip_test.go` + `lsp_fuzz_test.go`                                                                                                                   |
| Finding   | 3 files    | `finding_test.go` + `finding_builder_test.go` + `finding_valid_test.go`                                                                                                                           |
| Validate  | 2 files    | `validate_test.go` + `finding_valid_test.go` (both test IsValid/HasFix — possible overlap)                                                                                                        |

**Note:** Some of these splits are legitimate (fuzz/bench files are conventionally separate in Go). Merging them needs judgment, not blind consolidation.

### Testutil duplication across modules

Root `testutil_test.go` (24 helpers) and `pipeline/testutil_test.go` (31 helpers) have overlapping assertion patterns (`AssertErrIs`, `MakeFinding`, etc.). Go test helpers cannot be imported across modules, so some duplication is structural. Not addressed.

### Actual trivial/coverage-chasing test evaluation

The initial analysis said coverage tests were "chasing metrics." On closer look, they were genuine validation tests — the naming was wrong, not the content. But I did NOT do a deep audit of whether any individual test cases are trivial (testing getters/setters, zero-value paths, etc.). That audit remains open.

### `splitbrain_test.go` not investigated

There's a file called `splitbrain_test.go` in the root. The name suggests it's testing for a known anti-pattern. Not investigated whether it's a real test or a development artifact.

---

## d) TOTALLY FUCKED UP

### Sloppy merge technique — 3 compilation failures

Used `tail -n +N` to append file contents, which **cut inside import blocks** and left stray `)` and import fragments. This caused **3 separate compilation failures** that required manual fixes:

1. `coverage_test.go:321` — stray `. "github.com/onsi/gomega"` + `)` fragment
2. `position_test.go:308` — stray `)` fragment
3. `pipeline/conflict_test.go:141` — stray `)` fragment
4. `pipeline/fix_applier_test.go:210` — stray `"github.com/larsartmann/go-finding"` + import block fragment
5. `pipeline/fix_applier_test.go:461` — stray `. "github.com/onsi/gomega"` fragment
6. `pipeline/pipeline_test.go:262` — stray import block fragment
7. `pipeline/example_test.go:77` — stray `)` fragment

**7 manual fixes** that should have been 0. The correct approach would have been: read the file, identify the exact line after the import block, use a script that strips the entire `package` + `import` block from the source before appending. Or better: use `goimports` post-merge to auto-resolve imports.

### Used `trash` instead of `git rm`

Used `trash` to delete files per AGENTS.md safety rules, but `git mv` for renames. The `trash` approach works (git sees the deletion) but it's inconsistent — `git rm` would have staged the deletion immediately. Not a real problem since the working tree state is correct, but it means `git status` shows unstaged deletions.

---

## e) WHAT WE SHOULD IMPROVE

### Merge technique needs a reusable approach

The `tail -n +N` approach was fragile and wrong for files with multi-line import blocks. A proper merge script would:

1. Parse the Go file with `go/parser` to find the import block boundaries
2. Union the import sets
3. Append only the declarations (not package/import)
4. Run `goimports` to clean up

This should be documented as a procedure if we ever do this again.

### BDD test quality not audited

Loaded the `bdd-testing` skill per instructions but never actually evaluated whether the existing BDD tests follow its conventions (black-box package, fresh subject per spec, behavior assertions vs implementation). The merge was purely mechanical.

### Test coverage percentage not measured

No `go test -cover` was run. We don't know if the merges changed coverage (they shouldn't have — same tests, different files — but we didn't verify).

---

## f) Up to 25 Things We Should Get Done Next

| #   | Task                                                                                                                                       | Impact | Effort  |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------- |
| 1   | Run `nix run .#lint` across all modules to verify linter passes                                                                            | High   | Low     |
| 2   | Run `GOWORK=off go test ./...` in each module dir to verify replace directives                                                             | High   | Low     |
| 3   | Run `go test -cover ./...` and compare coverage before/after merge                                                                         | Medium | Low     |
| 4   | Audit `validate_test.go` vs `finding_valid_test.go` for test duplication                                                                   | Medium | Low     |
| 5   | Audit SARIF 7-file split — can sarif_fuzz + sarif_roundtrip merge into sarif_test?                                                         | Medium | Medium  |
| 6   | Audit pipeline 7-file split — are pipeline_detector/triage/logging/metrics separate concerns or fragments?                                 | Medium | Medium  |
| 7   | Evaluate whether `_fuzz_test.go` and `_bench_test.go` files should merge into parent (Go convention says no, but worth conscious decision) | Low    | Low     |
| 8   | Audit LSP 4-file split for consolidation                                                                                                   | Low    | Medium  |
| 9   | Audit merge 4-file split for consolidation                                                                                                 | Low    | Medium  |
| 10  | Check `testutil_test.go` vs `pipeline/testutil_test.go` for extractable shared patterns                                                    | Low    | Medium  |
| 11  | Investigate `splitbrain_test.go` — is it a real test or a dev artifact?                                                                    | Low    | Low     |
| 12  | Audit individual test cases for triviality (getter/setter tests, zero-value paths)                                                         | Medium | High    |
| 13  | Run `ginkgo unfocus` to verify no focus specs exist (manual grep found none, but tooling is authoritative)                                 | Low    | Trivial |
| 14  | Evaluate whether `example_test.go` (now ~750 lines) is too large after merging 4 files                                                     | Low    | Low     |
| 15  | Consider splitting `example_test.go` by domain (basic examples vs advanced examples) if >400 lines                                         | Low    | Low     |
| 16  | Document the merge technique as a reusable procedure in AGENTS.md or a script                                                              | Low    | Low     |
| 17  | Check if `finding_builder_test.go` should merge into `finding_test.go`                                                                     | Low    | Low     |
| 18  | Check if `errors_test.go` + `errors_valid_test.go` should merge                                                                            | Low    | Low     |
| 19  | Run `nix run .#test` (the documented full-suite command) to verify it still works                                                          | High   | Trivial |
| 20  | Run `nix run .#bench` to verify benchmarks still pass                                                                                      | Medium | Low     |
| 21  | Consider adding CI check that rejects `_extra_test.go` / `_bugfix_test.go` filenames                                                       | Medium | Low     |
| 22  | Audit `pipeline/byte_conflict_test.go` — does it belong in `conflict_test.go`?                                                             | Low    | Low     |
| 23  | Audit `position_overlap_test.go` — does it belong in `position_test.go`?                                                                   | Low    | Low     |
| 24  | Evaluate `report_iter_test.go` + `report_validate_test.go` — should they merge into `report_test.go`?                                      | Low    | Low     |
| 25  | Commit the changes (user has not asked for this yet)                                                                                       | High   | Trivial |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the SARIF/lsp/merge/pipeline/fix_engine multi-file splits be consolidated further, or are they legitimate test organization?**

The `_extra`/`_bugfix` pattern was clearly wrong — those were fragments created reactively. But the current splits (e.g., `sarif_export_test.go` + `sarif_import_test.go` + `sarif_roundtrip_test.go` + `sarif_properties_test.go` + `sarif_suppression_test.go` + `sarif_fuzz_test.go` + `sarif_test.go` = 7 files for one `sarif_export.go` + `sarif_import.go` + `sarif_types.go` production area) could be either:

- **Legitimate decomposition** — each file tests a distinct concern (export, import, round-trip, properties, suppression, fuzz) and the split aids navigation
- **Over-fragmentation** — the same root cause as `_extra` files, just with better names

I cannot determine this without understanding the team's navigation preferences. A 7-file split for ~800 lines of production SARIF code could be either perfectly organized or wildly over-fragmented. **What's the target file count per production file — 1:1, or is a bounded N:1 acceptable when the test concerns are genuinely distinct?**

---

## Summary

The session successfully eliminated the test fragmentation pattern (17 files, 0 remaining). All quality gates pass. The merge technique was sloppy (7 manual fixes from import-block truncation) but the end state is correct. The main open question is whether further consolidation of the remaining multi-file subjects is warranted or would be over-engineering.
