# Status Report — go-finding Post-Merge

> **📦 RESOLUTION STATUS (updated 2026-07-16)**
>
> ALL NOT STARTED items subsequently completed. CHANGELOG written, v1.0.0/v1.1.0/v1.2.0 tagged,
> `docs/MIGRATION_v1.0.md` written, `modularize/unix-style` branch deleted, deprecated SARIF APIs
> removed in v1.1.0. The go.work .gitignore bomb was defused with negation overrides.

**Date:** 2026-07-06 02:32
**Branch:** `master` (merged from `modularize/unix-style`)
**Session scope:** M12-M22 completion → self-review fixes → merge to master → root cleanup
**Previous report:** `docs/status/2026-07-06_02-04_post-m12-m22-completion.md`

---

## a) FULLY DONE

### Merged to master ✅

- `--no-ff` merge commit with 121 files changed, +5883/-502 lines
- All 22 Pareto tasks from the improvement plan are on master
- All 9 test packages pass with `-race`
- `GOWORK=off` isolation passes for all 4 modules
- `golangci-lint`: 0 issues
- BuildFlow pre-commit hooks pass on every commit

### All 22 Pareto Tasks ✅

| Tier              | Tasks             | Status |
| ----------------- | ----------------- | ------ |
| Tier 0 (1%→51%)   | M01, M02, M03     | Done   |
| Tier 1 (4%→64%)   | M05-M09, M11      | Done   |
| Tier 2 (20%→80%)  | M04, M10, M12-M17 | Done   |
| Tier 3 (80%→100%) | M18-M22           | Done   |

### Root cleanup ✅

- Removed `benchmarks/.gitkeep` (empty placeholder)
- Gitignored benchmark artifacts, coverage, dist, reports, result dirs
- Fixed `go.work` tracking vs buildflow-managed `.gitignore` conflict (manual block overrides)

### Documentation ✅

- AGENTS.md: 6 new gotchas, 5 new architecture decisions
- `docs/guides/fix-engine.md`: Full FixEngine usage guide
- `docs/integration-guide.md`: 3 new sections (SARIF suppressions, LSP fidelity, FilePath)
- README.md: Updated flag references (`-min-severity`)
- Planning doc: Marked ALL 22 tasks complete

---

## b) PARTIALLY DONE

### go.work .gitignore conflict

- **Status:** `go.work` IS tracked in git AND listed in buildflow's `.gitignore` block
- **Functionally OK:** gitignore doesn't affect already-tracked files
- **Fragile:** If someone runs `git rm --cached go.work`, buildflow's block would prevent re-adding it
- **Root cause:** BuildFlow auto-regenerates its managed `.gitignore` block and includes `go.work`/`go.work.sum`. The project comment says "committed for collaborative development" but buildflow doesn't know that
- **Not fixed:** Needs buildflow config change, not .gitignore edit

### Deprecated SARIF APIs still present

- `ToSARIFFiltered` and `WriteSARIFFiltered` are marked deprecated but not removed
- Fully replaced by `ToSARIFWithOpts(WithMinSeverity(sev))`
- Previous status report asked if they should be removed before v1.0 — never answered

### modularize/unix-style branch

- Now 1 commit behind master (the benchmarks cleanup commit)
- Not deleted, not updated
- Should be deleted now that it's merged

---

## c) NOT STARTED

- **CHANGELOG.md** — No entries for the massive API changes (FilePath, SARIFOption, LSPDiagnosticData, CLI flag rename)
- **v1.0.0 git tag** — `v1.0.0` doesn't exist. Last tag is `v0.9.1`. Despite breaking changes being on master
- **Migration guide** — No `docs/MIGRATION_v1.0.md` for consumers facing `string` → `FilePath` change
- **Consumer compatibility test** — No verification that the 20 audited downstream projects still compile

---

## d) TOTALLY FUCKED UP

### go.work in .gitignore is a ticking bomb

BuildFlow's managed block still contains `go.work` and `go.work.sum` at lines 102-103. I tried to remove them from the managed block, but BuildFlow regenerated the block on the next pre-commit hook run, re-adding them. The manual comment at line 20 says "committed" but the tool actively fights this. **Any developer who trusts the .gitignore and runs `git clean` or `git rm --cached go.work` will lose the workspace file.**

### First commit was a 58-file mega-commit

`07aab19` lumped M12+M13+M14+M20+M22 into one commit because the changes were intertwined in the same files. M12 (FilePath) should have been committed alone first — it's a pure mechanical type change. The SARIF and LSP changes could have been separate commits on top.

### M12 FilePath migration used sed — error-prone

The sed-based bulk test fixing required 4+ manual corrections:

- `fuzz_test.go`: `file1,` vs `file1}` pattern mismatch
- `integration_test.go`: sed mangled a line into syntax error
- `testutil_test.go`: sed produced `finding.finding.NewRangePtr` double-prefix
- BDD tests needed separate `finding.FilePath(...)` prefix for external test package

A Go-aware tool (gopls rename, or AST-based script) would have been correct on the first pass.

### Analysis FromDiagnostic reads source from disk

`analysis.FromDiagnostic` now calls `os.ReadFile(startPos.Filename)` to extract `BeforeCode` from TextEdits. This is a **filesystem side effect in what was previously a pure conversion function**. It breaks for in-memory test fixtures and creates a hidden I/O dependency. Should have been `FromDiagnosticWithSource(d, fset, source, ...)` — but I took the shortcut.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **go.work .gitignore conflict** — Fix in BuildFlow config, not .gitignore. BuildFlow should have a whitelist for tracked files that its generator respects.
2. **Analysis BeforeCode extraction** — Extract a `FromDiagnosticWithSource` variant that accepts `[]byte` source content. The disk-reading version becomes a thin wrapper.
3. **Deprecated API accumulation** — Decide on a deprecation policy: remove at v1.1? Keep forever? Build-tag them?
4. **LSPDiagnosticData couples go-finding into LSP type** — Consider moving to a separate options type.
5. **FilePath migration uncovered missing validation** — `Range.Start.File` and `Range.End.File` can differ with no check. A `ValidateSameFile()` method or constructor validation would close this.

### Process

6. **No version tag despite breaking changes** — Master has `string` → `FilePath` API breakage with no tag. Consumers pinning `@v0.9.1` are safe but anyone using `@latest` will break.
7. **No CHANGELOG.md** — The biggest API change in the project's history has zero changelog entry.
8. **Commit granularity** — The 58-file mega-commit is unreviewable. Future work should commit per-task.
9. **modularize/unix-style branch is stale** — Should be deleted after merge to avoid confusion.

### Testing

10. **No CLI binary integration test** — We test packages but never run the compiled binary.
11. **No fuzz test for LSP round-trip** — SARIF has fuzz tests, LSP doesn't.
12. **No benchmark for SARIF/LSP/FilePath operations** — Performance is unmeasured.
13. **gopls still shows 3 `infertypeargs` infos** in adapter_test.go — never cleaned up.

---

## f) Up to 25 Things We Should Get Done Next

### Critical (do first)

1. **Fix go.work .gitignore bomb** — Configure BuildFlow to not ignore go.work, or add a post-generation hook that strips go.work from the managed block
2. **Write CHANGELOG.md** — Document all breaking changes, new APIs, and fixes
3. **Tag v1.0.0** — The code is stable, all tests pass, all 22 tasks done. Ship it.
4. **Delete `modularize/unix-style` branch** — It's merged and stale
5. **Write `docs/MIGRATION_v1.0.md`** — Show consumers how to migrate from `string` to `FilePath`

### High impact

6. **Extract `FromDiagnosticWithSource`** — Remove disk I/O from analysis conversion
7. **Remove deprecated `ToSARIFFiltered`/`WriteSARIFFiltered`** — Clean break for v1.0
8. **Add CLI integration test** — Run actual binary, verify SARIF/JSON/text output
9. **Add LSP fuzz test** — Mirror SARIF fuzz test pattern
10. **Add `--include-suppressed` CLI flag** — Let users control suppression visibility per-run
11. **Add `Range.ValidateSameFile()`** — Ensure Start.File == End.File
12. **Benchmark SARIF export/import** — Measure throughput on 10k findings

### Medium impact

13. **Add `ConfidenceUnknown = -1` sentinel** — Like `OffsetUnknown`, centralize "unset"
14. **Clean up 3 gopls `infertypeargs` infos** in adapter_test.go
15. **Add SARIF suppression expiry round-trip test** — Currently only kind/reason tested
16. **Document go-finding SARIF property-bag schema** — Formalize `go-finding/*` properties
17. **Add Go doc examples** — Runnable examples for top 5 APIs in `example_test.go`
18. **Add JSON Schema for config files** — Validate before parsing
19. **Add `Report.Merge()` method** — Combine multiple reports with dedup
20. **Review consumer compatibility** — Verify FilePath doesn't break the 20 downstream projects

### Lower impact

21. **Switch FixStrategy to int enum** — Compile-time exhaustiveness in switches
22. **Switch Severity to int enum** — Faster comparisons
23. **Add `Pipeline.RunIter()` streaming API** — iter.Seq for findings-as-detected
24. **Add `finding.FormatMarkdown()` to core** — Currently CLI-only via go-output
25. **Profile-guided optimization** — Run `-cpuprofile` on benchmarks, find hotspots

---

## g) Top #1 Question

**Should I tag v1.0.0 right now, or write the CHANGELOG + migration guide first?**

Master has breaking changes (`Position.File` is now `FilePath`, `-severity` renamed, `Suppression.Rule` is `RuleName`) with no version tag, no changelog, and no migration guide. Any consumer using `go get github.com/larsartmann/go-finding@latest` will get a compilation error.

I can't decide whether to:

- **(A)** Tag v1.0.0 now (code is stable, tests pass, ship it) and write CHANGELOG/migration as fast follow-up
- **(B)** Write CHANGELOG + migration guide first, then tag v1.0.0 as a clean release with docs
- **(C)** Tag v1.0.0-rc1 (release candidate) to signal "breaking changes, not final yet"

This is a release management decision I can't make alone — it depends on whether any consumers are actively pulling `@latest` and whether you want a formal release process.
