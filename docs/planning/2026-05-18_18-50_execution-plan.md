# Execution Plan — Full Codebase Hardening

**Generated:** 2026-05-18 18:50
**Source:** Comprehensive full audit (docs/status/2026-05-18_18-45_comprehensive-full-audit.md) + TODO_LIST.md (42 open items)
**Rule:** Every task ≤ 12 minutes. All items from audit + TODO list included.

---

## Priority Scoring

| Axis               | Weight | Description                             |
| ------------------ | ------ | --------------------------------------- |
| **Impact**         | 40%    | Correctness > API stability > UX > docs |
| **Effort**         | 30%    | Lower effort first (Pareto)             |
| **Customer Value** | 20%    | External-facing > internal              |
| **Risk**           | 10%    | Breaking change penalty                 |

---

## Execution Plan — All Tasks

### Legend

- `P0` = Correctness bug (MUST fix now)
- `P1` = Design/API (before v1.0 API lock)
- `P2` = Quality improvement
- `P3` = Future/deferred
- `DOC` = Documentation only
- `WONTFIX` = Intentionally rejected, close ticket

---

| #   | Priority | Task                                                                                              | File(s)                                  | Effort | Impact   | Depends On |
| --- | -------- | ------------------------------------------------------------------------------------------------- | ---------------------------------------- | ------ | -------- | ---------- |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 1: P0 CORRECTNESS BUGS**                                                                  |                                          |        |          |            |
| 1   | P0       | Add `RLock`/`RUnlock` to `Report.ActiveFindings()`                                                | `report.go`                              | 5m     | Critical | —          |
| 2   | P0       | Add `RLock`/`RUnlock` to `Report.BySeverity()`                                                    | `report.go`                              | 3m     | Critical | —          |
| 3   | P0       | Add `RLock`/`RUnlock` to `Report.ByCategory()`                                                    | `report.go`                              | 3m     | Critical | —          |
| 4   | P0       | Add `RLock`/`RUnlock` to `Report.ByFixStrategy()`                                                 | `report.go`                              | 3m     | Critical | —          |
| 5   | P0       | Add `RLock`/`RUnlock` to `Report.FindByID()`                                                      | `report.go`                              | 5m     | Critical | —          |
| 6   | P0       | Add `RLock`/`RUnlock` to `Report.FindByRule()`                                                    | `report.go`                              | 3m     | Critical | —          |
| 7   | P0       | Add `RLock`/`RUnlock` to `Report.Len()`                                                           | `report.go`                              | 3m     | Critical | —          |
| 8   | P0       | Add `RLock`/`RUnlock` to `Report.Filter()`                                                        | `report.go`                              | 5m     | Critical | —          |
| 9   | P0       | Add `RLock`/`RUnlock` to `Report.Map()`                                                           | `report.go`                              | 5m     | Critical | —          |
| 10  | P0       | Add `RLock`/`RUnlock` to `Report.All()` — document iterator lock caveat                           | `report.go`                              | 8m     | Critical | —          |
| 11  | P0       | Write concurrent Report read-write race test                                                      | `report_test.go` (new)                   | 10m    | High     | #1-10      |
| 12  | P0       | Run tests with `-race` to verify Report fix                                                       | CLI                                      | 2m     | High     | #11        |
|     |          |                                                                                                   |                                          |        |          |            |
| 13  | P0       | Analyze `IsAutoFixable()` vs `Validate()` — decide canonical behavior                             | Design                                   | 5m     | High     | —          |
| 14  | P0       | Fix `IsAutoFixable()` to require `BeforeCode != ""` OR fix `Validate()` to allow `AfterCode`-only | `finding.go`                             | 8m     | High     | #13        |
| 15  | P0       | Write test proving `IsAutoFixable`/`Validate` agreement for all FixStrategy+code combos           | `finding_valid_test.go`                  | 10m    | High     | #14        |
|     |          |                                                                                                   |                                          |        |          |            |
| 16  | P0       | Fix `DeduplicateByID` — skip findings with empty ID instead of colliding                          | `merge.go`                               | 5m     | Medium   | —          |
| 17  | P0       | Write test: two findings with empty IDs are NOT deduplicated                                      | `merge_test.go`                          | 8m     | Medium   | #16        |
|     |          |                                                                                                   |                                          |        |          |            |
| 18  | P0       | Fix conflict detection: change `Overlaps` to `Adjacent` check in group extension                  | `conflict.go`                            | 10m    | Medium   | —          |
| 19  | P0       | Write test: 3 fixes where A→B overlap but not A→C, verify C is applied                            | `conflict_test.go`                       | 10m    | Medium   | #18        |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 2: P1 DESIGN / API FIXES**                                                                |                                          |        |          |            |
| 20  | P1       | Add `context.Context` param to `FindingProcessor.Process`                                         | `adapters.go`                            | 5m     | High     | —          |
| 21  | P1       | Add `error` return to `FindingProcessor.Process`                                                  | `adapters.go`                            | 3m     | High     | #20        |
| 22  | P1       | Update `ProcessorFunc` adapter to accept ctx + return error                                       | `adapters.go`                            | 5m     | High     | #21        |
| 23  | P1       | Update `NamedProcessorFunc` adapter                                                               | `adapters.go`                            | 3m     | High     | #22        |
| 24  | P1       | Update `namedProcessor` struct + Process method                                                   | `adapters.go`                            | 5m     | High     | #23        |
| 25  | P1       | Update pipeline.go processor loop to pass ctx + handle error                                      | `pipeline.go`                            | 8m     | High     | #24        |
| 26  | P1       | Update all FindingProcessor test mocks                                                            | `*_test.go`                              | 10m    | High     | #24        |
| 27  | P1       | Run tests — fix compilation errors from interface change                                          | CLI                                      | 10m    | High     | #26        |
|     |          |                                                                                                   |                                          |        |          |            |
| 28  | P1       | Make `ComputeSummary` deterministic — add `now time.Time` param or use `IsSuppressedAt`           | `report.go`                              | 8m     | Medium   | —          |
| 29  | P1       | Update all `ComputeSummary` callers                                                               | `report.go`, `merge.go`                  | 5m     | Medium   | #28        |
| 30  | P1       | Write test: ComputeSummary produces same result at different times                                | `report_test.go`                         | 8m     | Medium   | #29        |
|     |          |                                                                                                   |                                          |        |          |            |
| 31  | P1       | Deduplicate `defaultMaxIterations` — export from pipeline, import in CLI                          | `config.go`, `main.go`                   | 5m     | Low      | —          |
|     |          |                                                                                                   |                                          |        |          |            |
| 32  | P1       | Decide: unexport `Confidence` field or validate at read sites                                     | Design                                   | 5m     | Medium   | —          |
| 33  | P1       | Implement Confidence protection (whichever design from #32)                                       | `finding.go` or `confidence.go`          | 10m    | Medium   | #32        |
| 34  | P1       | Write test: direct struct construction cannot produce invalid Confidence                          | `confidence_test.go`                     | 8m     | Medium   | #33        |
|     |          |                                                                                                   |                                          |        |          |            |
| 35  | P1       | Reuse FixApplier across pipeline iterations instead of per-iteration creation                     | `pipeline.go`                            | 10m    | Medium   | —          |
| 36  | P1       | Run pipeline tests to verify FixApplier reuse                                                     | CLI                                      | 5m     | Medium   | #35        |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 3: OPEN TODO LIST ITEMS**                                                                 |                                          |        |          |            |
| 37  | DOC      | Close `Properties map[string]any` TODO as `[WONTFIX]` with rationale                              | `TODO_LIST.md`                           | 3m     | Low      | —          |
| 38  | DOC      | Record `NewFinding` API pattern decision in `docs/architecture-decisions.md`                      | `architecture-decisions.md`              | 8m     | Low      | —          |
| 39  | DOC      | Record domain-specific provider location decision in ADR                                          | `architecture-decisions.md`              | 8m     | Low      | —          |
| 40  | DOC      | Mark `Decide NewFinding API pattern` as done in TODO_LIST.md                                      | `TODO_LIST.md`                           | 2m     | Low      | #38        |
| 41  | DOC      | Mark `Decide domain-specific provider location` as done in TODO_LIST.md                           | `TODO_LIST.md`                           | 2m     | Low      | #39        |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 4: V1.0 API AUDIT**                                                                       |                                          |        |          |            |
| 42  | P1       | Audit root package exported types — list all symbols                                              | Spreadsheet/doc                          | 10m    | High     | —          |
| 43  | P1       | Audit root package exported funcs — list all symbols                                              | Spreadsheet/doc                          | 10m    | High     | —          |
| 44  | P1       | Audit pipeline package exported types — list all symbols                                          | Spreadsheet/doc                          | 10m    | High     | —          |
| 45  | P1       | Audit pipeline package exported funcs — list all symbols                                          | Spreadsheet/doc                          | 10m    | High     | —          |
| 46  | P1       | Audit CLI package exported symbols                                                                | Spreadsheet/doc                          | 5m     | Medium   | —          |
| 47  | P1       | Audit analysis package exported symbols                                                           | Spreadsheet/doc                          | 5m     | Medium   | —          |
| 48  | P1       | Cross-reference symbols with FEATURES.md — flag gaps                                              | Doc                                      | 10m    | High     | #42-47     |
| 49  | P1       | Write API stability report — mark each symbol STABLE/EXPERIMENTAL/DEPRECATED                      | New doc                                  | 10m    | High     | #48        |
| 50  | DOC      | Update TODO_LIST.md: mark `API stability review` as done                                          | `TODO_LIST.md`                           | 2m     | Low      | #49        |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 5: P2 QUALITY IMPROVEMENTS**                                                              |                                          |        |          |            |
| 51  | P2       | Surface SubstringProvider ambiguity — add ambiguity count to FixEdit or log                       | `fix_provider.go`                        | 10m    | Low      | —          |
| 52  | P2       | Wire `FilterConflictingEdits` as opt-in pipeline option in Config                                 | `pipeline.go`, `config.go`               | 10m    | Medium   | —          |
| 53  | P2       | Write test for `FilterConflictingEdits` integration with pipeline                                 | `pipeline_test.go`                       | 10m    | Medium   | #52        |
| 54  | P2       | Add `finding.Diff()` utility — compare two finding sets                                           | New `diff.go`                            | 10m    | Low      | —          |
| 55  | P2       | Write tests for `finding.Diff()`                                                                  | `diff_test.go`                           | 10m    | Low      | #54        |
| 56  | P2       | Add `finding.FormatText()` — human-readable single-line format                                    | New `format.go`                          | 10m    | Low      | —          |
| 57  | P2       | Add `finding.FormatMarkdown()` — markdown table format                                            | `format.go`                              | 10m    | Low      | #56        |
| 58  | P2       | Write tests for `FormatText` + `FormatMarkdown`                                                   | `format_test.go`                         | 10m    | Low      | #57        |
| 59  | P2       | Add per-detector timeout to `Config` and `RetryDetector`                                          | `config.go`, `retry.go`                  | 10m    | Medium   | —          |
| 60  | P2       | Write test for per-detector timeout                                                               | `retry_test.go`                          | 8m     | Medium   | #59        |
| 61  | P2       | Add structured logging (`slog`) to pipeline stages                                                | `pipeline.go`                            | 10m    | Medium   | —          |
| 62  | P2       | Add structured logging (`slog`) to CLI `run()`                                                    | `main.go`                                | 10m    | Medium   | #61        |
| 63  | P2       | Add progress reporting callback to Config                                                         | `config.go`                              | 8m     | Low      | —          |
| 64  | P2       | Wire progress callback into pipeline loop                                                         | `pipeline.go`                            | 8m     | Low      | #63        |
| 65  | P2       | Write test for progress callback                                                                  | `pipeline_test.go`                       | 8m     | Low      | #64        |
| 66  | P2       | Add `go/analysis` reverse conversion: `ToDiagnostic()`                                            | `analysis/analysis.go`                   | 10m    | Medium   | —          |
| 67  | P2       | Write tests for `ToDiagnostic()`                                                                  | `analysis/analysis_test.go`              | 10m    | Medium   | #66        |
| 68  | P2       | Replace hardcoded `knownDetectorBuilders` with registry lookup in CLI                             | `main.go`                                | 10m    | Low      | —          |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 6: P3 FUTURE / DEFERRED**                                                                 |                                          |        |          |            |
| 69  | P3       | Evaluate `go-sarif` vs hand-rolled: write comparison doc                                          | New doc                                  | 10m    | Low      | —          |
| 70  | DOC      | Mark `Evaluate go-sarif` TODO with decision + link to doc                                         | `TODO_LIST.md`                           | 3m     | Low      | #69        |
| 71  | P3       | Design pipeline middleware/interceptor interface                                                  | Design doc                               | 10m    | Low      | —          |
| 72  | P3       | Implement middleware pattern: `PipelineOption` func chain                                         | `pipeline.go`                            | 10m    | Low      | #71        |
| 73  | P3       | Write BDD tests for middleware pattern                                                            | `pipeline/bdd_test.go`                   | 10m    | Low      | #72        |
| 74  | P3       | Design watch mode API — `Watch(ctx, dir, callback)`                                               | Design doc                               | 10m    | Low      | —          |
| 75  | P3       | Implement watch mode with fsnotify                                                                | New `watch.go`                           | 10m    | Low      | #74        |
| 76  | P3       | Write tests for watch mode                                                                        | `watch_test.go`                          | 10m    | Low      | #75        |
| 77  | P3       | Design semantic merge for conflicts — AST-aware conflict resolution                               | Design doc                               | 10m    | Low      | —          |
| 78  | P3       | Add styled CLI output with lipgloss — text format only                                            | `config.go`                              | 10m    | Low      | —          |
| 79  | P3       | Add interactive TUI for fix review — bubbletea prototype                                          | New `cmd/go-finding/tui.go`              | 10m    | Low      | —          |
| 80  | P3       | Add golangci-lint detector integration                                                            | New `internal/detectors/golangcilint.go` | 10m    | Low      | —          |
| 81  | P3       | Add errcheck detector integration                                                                 | New `internal/detectors/errcheck.go`     | 10m    | Low      | —          |
| 82  | P3       | Design LSP CodeAction support                                                                     | Design doc                               | 10m    | Low      | —          |
| 83  | P3       | Mark `Finding struct sub-grouping` as deferred-to-v2 in TODO                                      | `TODO_LIST.md`                           | 2m     | Low      | —          |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 7: NIX MIGRATION**                                                                        |                                          |        |          |            |
| 84  | P3       | Create `flake.nix` with devShell + build + test + lint                                            | `flake.nix`                              | 10m    | Medium   | —          |
| 85  | P3       | Verify `nix develop` works                                                                        | CLI                                      | 5m     | Medium   | #84        |
| 86  | P3       | Verify `nix build` works                                                                          | CLI                                      | 5m     | Medium   | #84        |
| 87  | P3       | Replace CI workflows with Nix-based versions                                                      | `.github/workflows/ci.yml`               | 10m    | Medium   | #85        |
| 88  | P3       | Add `.envrc` for direnv integration                                                               | `.envrc`                                 | 5m     | Low      | #85        |
| 89  | P3       | Update CONTRIBUTING.md with Nix instructions                                                      | `CONTRIBUTING.md`                        | 5m     | Low      | #88        |
| 90  | P3       | Add `nix flake check` integration                                                                 | `flake.nix`                              | 8m     | Medium   | #84        |
| 91  | P3       | Pin nixpkgs with `flake.lock`                                                                     | `flake.lock`                             | 3m     | Medium   | #84        |
| 92  | P3       | Test cross-platform (Linux + macOS)                                                               | CI                                       | 10m    | Medium   | #91        |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 8: EXTERNAL / OUT OF SCOPE — CLOSE TICKETS**                                              |                                          |        |          |            |
| 93  | DOC      | Mark `Decide on BuildFlow integration` as deferred (external dep)                                 | `TODO_LIST.md`                           | 2m     | Low      | —          |
| 94  | DOC      | Mark `Decide on go-business-rules Severity sharing` as deferred (external dep)                    | `TODO_LIST.md`                           | 2m     | Low      | —          |
| 95  | DOC      | Mark `Build a full language server` as rejected-per-ADR                                           | `TODO_LIST.md`                           | 2m     | Low      | —          |
| 96  | DOC      | Mark `Code fixes via LSP` as deferred                                                             | `TODO_LIST.md`                           | 2m     | Low      | —          |
| 97  | DOC      | Mark `AI backend for FixStrategyAI` as out-of-scope-v1                                            | `TODO_LIST.md`                           | 2m     | Low      | —          |
| 98  | DOC      | Mark remaining P3 out-of-scope items as `OUT-OF-SCOPE-V1`                                         | `TODO_LIST.md`                           | 5m     | Low      | —          |
|     |          |                                                                                                   |                                          |        |          |            |
|     |          | **PHASE 9: FINAL VERIFICATION**                                                                   |                                          |        |          |            |
| 99  | P0       | Run `go build ./...`                                                                              | CLI                                      | 1m     | Critical | All        |
| 100 | P0       | Run `go test -race -count=1 ./...`                                                                | CLI                                      | 5m     | Critical | All        |
| 101 | P0       | Run `golangci-lint run ./...`                                                                     | CLI                                      | 5m     | Critical | All        |
| 102 | P0       | Run `go vet ./...`                                                                                | CLI                                      | 1m     | Critical | All        |
| 103 | DOC      | Update TODO_LIST.md with all completed items                                                      | `TODO_LIST.md`                           | 5m     | Low      | All        |
| 104 | DOC      | Update FEATURES.md with any new status changes                                                    | `FEATURES.md`                            | 5m     | Low      | All        |
| 105 | DOC      | Update AGENTS.md with session learnings                                                           | `AGENTS.md`                              | 8m     | Low      | All        |

---

## Summary Statistics

| Metric                     | Value      |
| -------------------------- | ---------- |
| **Total tasks**            | 105        |
| **P0 (correctness)**       | 19 tasks   |
| **P1 (design/API)**        | 22 tasks   |
| **P2 (quality)**           | 18 tasks   |
| **P3 (future)**            | 26 tasks   |
| **DOC (documentation)**    | 16 tasks   |
| **WONTFIX closures**       | 4 tasks    |
| **Estimated total time**   | ~15 hours  |
| **Phase 1 (P0) time**      | ~2.5 hours |
| **Phase 1+2 (P0+P1) time** | ~7 hours   |

---

## Pareto Analysis: 20% Effort → 80% Value

If you only do **21 tasks** (Phase 1 + top P1s), you get **80% of the value**:

1. Tasks #1-12: Report thread-safety (fix + verify)
2. Tasks #13-15: IsAutoFixable/Validate agreement
3. Tasks #16-17: DeduplicateByID empty-ID fix
4. Tasks #18-19: Conflict detection fix
5. Tasks #20-27: FindingProcessor ctx+error
6. Tasks #28-30: ComputeSummary determinism
7. Task #31: Deduplicate defaultMaxIterations
8. Tasks #35-36: FixApplier reuse

**Time: ~3.5 hours. Value: All P0 bugs fixed + highest-impact P1 done.**

---

_Assisted-by: Crush_
