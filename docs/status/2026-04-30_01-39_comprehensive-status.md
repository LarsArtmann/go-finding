# Comprehensive Session Status Report

**Date:** 2026-04-30 01:39
**Session Focus:** Coverage completion, architectural reflection, and incremental improvements

---

## 1. Fully Done

### Version Mismatch Fix
- **Fixed** `cmd/go-finding/main.go` hardcoding `"0.1.0"` while `version.go` declared `"0.1.3"`
- CLI now derives version from `finding.Version` automatically
- Single-line change, no test impact, immediate correctness improvement

### Coverage Gap Closure
Closed 5 remaining coverage gaps across the codebase:

| File | Function | Before | After | How |
|------|----------|--------|-------|-----|
| `pipeline/fix_applier.go:23` | `NewFixApplier` | 75.0% | 100% | Added `TestNewFixApplier_MkdirTempFallback` using `t.Setenv("TMPDIR", "/etc/passwd")` to trigger `os.MkdirTemp` error path |
| `pipeline/file_backup.go:65` | `Backup` MkdirAll error | 83.3% | 91.7% | Added `TestFileBackup_Backup_MkdirAllError` using a file-as-directory to trigger `MkdirAll` failure |
| `pipeline/file_backup.go:99` | `Restore` read backup error | 92.3% | 100% | Added `TestFileBackup_Restore_ReadBackupError` deleting backup file after creation |
| `pipeline/pipeline.go:393` | `detectSequential` ctx cancel | 88.9% | 100% | Added `TestDetectSequential_ContextCancellation` with cancelled context and slow detector |
| `sarif.go:172/187` | `WriteSARIF`/`WriteSARIFFiltered` writer error | 66.7% | 100% | Added `TestWriteSARIF_WriterError` and `TestWriteSARIFFiltered_WriterError` using `failWriter` stub |

### Lint Cleanup
- Fixed all pre-existing `golines`, `gci`, and `testifylint` issues in the working tree
- Reformatted struct literals across `finding_builder_test.go`, `finding_extra_test.go`, `merge_test.go`, `sarif_test.go`
- Fixed `revive` unused-receiver warning on `failWriter.Write`
- Replaced `assert.Equal(t, "", ...)` with `assert.Empty(...)` per `testifylint`

### Verification
- All tests pass: `go test -race ./...` ✅
- All coverage thresholds pass: `bash scripts/coverage-check.sh` ✅
- Lint: 0 issues ✅

### Final Coverage Metrics
| Package | Coverage | Threshold |
|---------|----------|-----------|
| `finding` (root) | 99.4% | 98.0% ✅ |
| `pipeline` | 98.0% | 95.0% ✅ |
| `cmd/go-finding` | 96.2% | 90.0% ✅ |
| `internal/detectors` | 96.1% | 90.0% ✅ |
| **Total** | **95.5%** | **93.0%** ✅ |

---

## 2. Partially Done

### Architectural Reflection & Planning
- **Completed deep analysis** of the codebase architecture, type model, and extension points
- **Identified 6 high-impact improvements** for type model and CLI
- **Researched well-established libraries** that could benefit the project:
  - `github.com/charmbracelet/lipgloss` — styled CLI output
  - `github.com/alecthomas/kong` — structured CLI framework
  - `github.com/fsnotify/fsnotify` — watch/daemon mode
  - `github.com/owenrumney/go-sarif` — standard SARIF library (evaluated: not worth switching, our impl is solid)
  - `github.com/sergi/go-diff` — textual diff for verification
- **Did NOT implement** the planned type model improvements (Tags, Key fix, Properties) or CLI enhancements — these require focused implementation sessions

### Pre-existing Lint Issues in Working Tree
- Fixed all lint issues that existed before this session started
- Some were in files untouched by this session (e.g., `finding_builder_test.go`, `merge_test.go`)

---

## 3. Not Started

### Type Model Improvements
1. **Add `Tags []string` field** to `Finding` for multi-tag support (keep `Tag string` for backward compat)
2. **Fix `Finding.Key()` cross-tool collision** — include `ToolName` in fallback key when `ID` is empty
3. **Add `Properties map[string]any`** alongside `Metadata map[string]string` for structured round-trip data
4. **Improve `Category.IsValid()`** — currently accepts any non-empty string; should validate against standard categories or be renamed

### CLI Enhancements
5. **Add `-output` flag** for file output (currently only stdout)
6. **Add version flag** (`-version` / `-v`) to CLI
7. **Styled text output** using `lipgloss` or similar

### Documentation & Clarity
8. **Document `FixStrategyAI` placeholder status** — it's been an open question for multiple sessions
9. **Document round-trip losses** in SARIF format (already partially documented in `sarif.go` comment)

### Major Features
10. **Plugin architecture** for detectors (`hashicorp/go-plugin`)
11. **Interactive TUI** for fix review (`charmbracelet/bubbletea`)
12. **Watch/daemon mode** (`fsnotify`)
13. **Structured logging** replacing `fmt.Printf`
14. **Telemetry/metrics export** (Prometheus/OpenTelemetry)

---

## 4. Totally Fucked Up

### The Lint Formatting Battle
Spent **too many iterations** fighting `golines` line-length violations across test files. The cycle was:
1. Write test with inline struct literal
2. Run linter → golines failure
3. Manually break lines
4. Run linter → another golines failure on a different line
5. Repeat 6+ times

**Root cause:** No automated formatter configured that matches `golines` rules. `gofumpt` was tried but didn't handle the same cases.

**Lesson:** Should add `golines` to the project's toolchain or configure `.golangci.yml` with an auto-fix step.

### gopls Stale Cache
The `examples/builder/main.go` shows a phantom typecheck error in gopls ("assignment mismatch") even though `go build` succeeds. The file is correct — `f, err := finding.NewBuilder(...).Build()` properly handles the 2 return values.

**Root cause:** gopls cache stale after prior file modifications.

**Lesson:** Need to document `gopls restart` procedure in project setup docs.

### detectParallel "Dead Code" Analysis (Previous Session)
The previous session's status report incorrectly flagged `detectParallel`'s `g.Wait()` error path as "dead code that can never execute." This was **wrong** — `TestPipelineRun_ParallelDetectorError` proves errors from `runOneDetector` ARE propagated through `errgroup`. The error path is very much alive.

**Lesson:** Don't assume goroutine error collection patterns are dead without reading the errgroup contract and existing tests.

---

## 5. What We Should Improve

### Tooling
1. **Add `golines` to CI or justfile** — Enforce consistent line breaking automatically
2. **Add `golangci-lint run --fix`** to justfile for auto-fixing format issues
3. **Document gopls restart** in AGENTS.md for stale cache issues

### Type Model
4. **`Tag string` → `Tags []string`** — Single-tag is a real limitation; findings often need multiple classifications (e.g., "security", "injection", "xss")
5. **`Metadata map[string]string` → support `any` values** — Forces serialization of structured data. SARIF properties are `map[string]any`; our string-only limitation loses type information on round-trip
6. **`Finding.Key()` should include `ToolName`** — Two different tools reporting the same rule+message+file will collide on the fallback key

### Architecture
7. **Unify `DetectResult` wrapper** — Normal detection returns `[]Finding`; partial detection returns `PartialResult{Findings, Errors}`. A unified result type would simplify the pipeline
8. **ConflictDetector is too conservative** — Overlapping fixes are grouped and only the first is kept, even when they don't semantically conflict
9. **FixEngine.applyStringFixes joins entire file** — Multiple non-conflicting string fixes in the same file can interact unpredictably

### CLI
10. **Output goes to stdout only** — No file output, no HTTP sink, no structured log streaming
11. **Version is not exposed via flag** — Users can't check `go-finding -version`
12. **Only 2 built-in detectors** — No plugin discovery mechanism

### Documentation
13. **`FixStrategyAI` fate** — Multiple sessions have flagged this. Need a decision: remove, document as reserved, or implement minimal interface
14. **SARIF round-trip losses** — `BeforeCode`, `Suppression`, `RelatedRef.FindingID`, `Tag` are lost on export→import. Documented in `sarif.go` but should also be in user-facing docs

---

## 6. Top 25 Things To Do Next

Sorted by **Impact / Work ratio** (Pareto principle: highest impact first):

| # | Task | Impact | Work | Package |
|---|------|--------|------|---------|
| 1 | Fix `Finding.Key()` cross-tool collision | High | 10 min | `finding` |
| 2 | Add `Tags []string` field to Finding | High | 20 min | `finding` |
| 3 | Add CLI `-output` flag for file output | High | 15 min | `cmd` |
| 4 | Document `FixStrategyAI` placeholder status | Medium | 5 min | `finding` |
| 5 | Add CLI `-version` flag | Medium | 5 min | `cmd` |
| 6 | Improve `Category.IsValid()` docs/behavior | Low | 5 min | `finding` |
| 7 | Add `Properties map[string]any` for structured metadata | Medium | 30 min | `finding` |
| 8 | Add file output sink to Report (WriteFile) | Medium | 20 min | `finding` |
| 9 | Add streaming report writers (io.Writer interface) | Medium | 25 min | `finding` |
| 10 | Document all SARIF round-trip losses in user docs | Low | 15 min | `docs` |
| 11 | Add `golines` to justfile/CI | Low | 10 min | `tooling` |
| 12 | Add detector timeout per-detector | Medium | 30 min | `pipeline` |
| 13 | Improve conflict detection (semantic merge) | High | 2h | `pipeline` |
| 14 | Add progress bars for long-running operations | Medium | 45 min | `cmd` |
| 15 | Add structured logging (slog) | Medium | 1h | `cmd` + `pipeline` |
| 16 | Plugin architecture for detectors | High | 4h | `pipeline` |
| 17 | Interactive TUI for fix review | High | 6h | `cmd` |
| 18 | Watch/daemon mode | Medium | 3h | `cmd` |
| 19 | Evaluate `github.com/owenrumney/go-sarif` | Low | 1h | `finding` |
| 20 | Add diff output to verification stage | Medium | 1h | `pipeline` |
| 21 | Add config file schema validation | Medium | 45 min | `cmd` |
| 22 | Add telemetry/metrics export | Low | 2h | `pipeline` |
| 23 | Add fuzz tests for FixEngine | Medium | 30 min | `pipeline` |
| 24 | Add property-based tests for Range operations | Low | 20 min | `finding` |
| 25 | Add integration tests for real govet/staticcheck execution | Medium | 1h | `internal/detectors` |

---

## 7. The Top 1 Question I Can't Figure Out

**Should `Metadata map[string]string` become `map[string]any`?**

This is a breaking API change with significant architectural implications:

**Arguments FOR changing to `map[string]any`:**
- SARIF properties are natively `map[string]any`; our string-only model loses type information on every round-trip
- JSON unmarshaling of `map[string]any` preserves numbers, booleans, nested objects
- The `Builder.WithMetadata` API already takes `map[string]string` which is awkward for structured data
- At v0.1.x, breaking changes are acceptable if documented

**Arguments AGAINST:**
- It's a breaking change for all consumers of the library
- `map[string]any` requires type assertions at every call site, making the API more cumbersome
- The current `map[string]string` is simple, predictable, and serializes cleanly
- Could instead add a SEPARATE `Properties map[string]any` field, but this adds complexity and confusion ("when do I use Metadata vs Properties?")

**The real question:** Is the type fidelity worth the API complexity? For a library that primarily moves data between tools, should the core type be strongly typed (`string`) or loosely typed (`any`)?

**Current leaning:** Keep `Metadata map[string]string` for the core type (simplicity, predictability) but add helper methods like `GetMetadataInt(key string) (int, bool)` and `GetMetadataBool(key string) (bool, bool)` for common type conversions. This preserves the simple API while providing type-safe accessors.

But I'm genuinely unsure if this is the right call or if we should just bite the bullet and change to `any`.

---

## Appendix: Session Artifacts

### Commits This Session
1. `84ec401` — fix(cli): use finding.Version instead of hardcoded version string

### Files Modified
- `cmd/go-finding/main.go` — version fix
- `pipeline/fix_applier_test.go` — NewFixApplier fallback test
- `pipeline/file_backup_test.go` — Backup/Restore error-path tests + `finding` import
- `pipeline/pipeline_test.go` — detectSequential cancellation test
- `sarif_test.go` — Writer error tests + `errors` import + formatting
- `finding_builder_test.go` — lint formatting
- `finding_extra_test.go` — lint formatting
- `merge_test.go` — lint formatting + dedup strategy behavior test

### Coverage Changes
| Package | Before | After | Delta |
|---------|--------|-------|-------|
| `finding` | 99.2% | 99.4% | +0.2% |
| `pipeline` | 97.3% | 98.0% | +0.7% |
| **Total** | **95.2%** | **95.5%** | **+0.3%** |
