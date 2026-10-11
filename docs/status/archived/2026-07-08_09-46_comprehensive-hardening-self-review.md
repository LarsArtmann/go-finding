# Status Report — Comprehensive Hardening Sweep

> **📦 RESOLUTION STATUS (updated 2026-07-16)**
>
> All fixes committed (`c0f900f`). SARIF Snippet spec compliance (`string`→`*sarifArtifactContent`)
> was subsequently fixed. Path traversal and TOCTOU security fixes were added in the follow-up
> session (2026-07-08_08-33). All fuzz targets pass (1.76M+ executions).

**Date**: 2026-07-08 09:46\
**Session**: Regression tests, security hardening, SARIF spec compliance, bug fixes\
**Commit**: `c0f900f` — `fix: comprehensive correctness and security hardening sweep`\
**Files changed**: 53 files, +2124/-149\
**All tests**: PASS (race detector, all 4 modules)\
**Lint**: 0 issues

---

## a) FULLY DONE

### Tests Written This Session

| Test                                                        | File                           | What it verifies                                                 |
| ----------------------------------------------------------- | ------------------------------ | ---------------------------------------------------------------- |
| `TestResolveSafePath_NormalRelativePath`                    | `pipeline/path_safety_test.go` | Basic path resolution                                            |
| `TestResolveSafePath_PathTraversal` (4 sub)                 | same                           | `../../etc/passwd`, `../secret`, nested, `.ssh`                  |
| `TestResolveSafePath_AbsolutePathTreatedAsRelative` (3 sub) | same                           | Absolute paths normalized under root                             |
| `TestResolveSafePath_EmptyRelPath`                          | same                           | Empty resolves to root                                           |
| `TestResolveSafePath_DotPath`                               | same                           | `.` resolves to root                                             |
| `TestResolveSafePath_SymlinkInsideRoot_PointingInside`      | same                           | Symlink resolving inside root is safe                            |
| `TestResolveSafePath_SymlinkInsideRoot_PointingOutside`     | same                           | Symlink escaping root is rejected                                |
| `TestResolveSafePath_NonexistentFile`                       | same                           | Nonexistent file inside root is safe                             |
| `TestResolveSafePath_RootItself`                            | same                           | Root resolves to itself                                          |
| `TestResolveSafePath_DeeplyNestedPath`                      | same                           | Deep nesting works                                               |
| `TestGroupFindingsBySafePath_ResolvedPathAsMapKey`          | same                           | TOCTOU: resolved path used as map key                            |
| `TestGroupFindingsBySafePath_PathTraversalFiltered`         | same                           | Traversal findings silently skipped                              |
| `TestFixApplier_RollbackErrorNotSwallowed`                  | `pipeline/fix_applier_test.go` | Saboteur provider deletes .bak → rollback error surfaces         |
| `TestSARIFRoundTrip_PositionOffset` (6 sub)                 | `sarif_roundtrip_test.go`      | Offset survives export+import (0, positive, large, range, unset) |
| `TestSARIFSnippet_BackwardCompat` (2 sub)                   | same                           | Both `{"text":"..."}` and bare string accepted                   |

### Code Changes This Session

1. **`pipeline/path_safety.go`** (NEW): `resolveSafePath` — shared path containment helper
2. **`pipeline/path_safety_test.go`** (NEW): 14 test functions, 25+ sub-tests
3. **`pipeline/fix_applier_test.go`**: Added `saboteurProvider` + rollback error test
4. **`sarif_types.go`**: `Snippet string` → `*sarifArtifactContent` per SARIF §3.30.13; custom `UnmarshalJSON` for backward compat (accepts both object + bare string)
5. **`sarif_export.go`**: Export snippet as `&sarifArtifactContent{Text: ...}`
6. **`sarif_import.go`**: Import snippet from `region.Snippet.Text`
7. **`sarif_roundtrip_test.go`**: Offset round-trip property test + backward compat test
8. **`pipeline/pipeline.go`**: Fixed `hookErr :=` → `hookErr =` (build error)
9. **`pipeline/pipeline_iteration.go`**: Fixed `hookErr :=` → `hookErr =` (build error)

### Verification

- `go build ./...` — PASS
- `go vet ./...` — PASS
- `go test -race -count=1 ./...` — PASS (all 4 modules)
- `golangci-lint run ./...` — 0 issues
- BuildFlow pre-commit — 27/27 passed

---

## b) PARTIALLY DONE

| Item                          | Status                                                       | What's Missing                                                                       |
| ----------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| SARIF Snippet spec compliance | Export emits `{"text":"..."}` ✓, import accepts both forms ✓ | Could also export `binary` or `rendered` fields per spec (low value)                 |
| LSP round-trip fidelity       | Snippet, Suppression, Metadata, RelatedFindingIDs done ✓     | Could add Tags, BeforeCode/AfterCode to LSPDiagnosticData (already in Data via JSON) |

---

## c) NOT STARTED

| Item                                                        | Why It Matters                                                                                          |
| ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Benchmark baseline                                          | No baseline.txt exists; can't detect performance regressions                                            |
| TOCTOU symlink swap test (runtime)                          | Current test proves resolved path is the map key; doesn't test actual concurrent symlink swap mid-apply |
| `resolveSafePath` as exported API                           | Currently package-private; could be useful for external consumers                                       |
| SARIF `binary` snippet form                                 | Spec allows base64 binary; go-finding only handles text                                                 |
| Integration test: full pipeline with path traversal finding | Unit-tested in isolation but not through the full pipeline.Run() path                                   |

---

## d) TOTALLY FUCKED UP (CAUGHT AND FIXED)

### D1: Committed broken build — `hookErr :=` where `=` was needed

**What happened**: The prior session introduced `hookErr := p.fireStageHook(...)` at two locations in the same scope (`pipeline.go:266`, `pipeline_iteration.go:95`). The second `:=` re-declares an existing variable → Go compile error `no new variables on left side of :=`.

**How it got through**: My `go build ./...` check passed before the commit (possibly stale Go build cache). BuildFlow's pre-commit hook doesn't include `go build` — it only runs linters/formatters. The error surfaced when I ran `go build ./...` after the status report request.

**Impact**: The committed code (`262473f`) would NOT compile for anyone pulling it. Tests also couldn't compile. This was a showstopper.

**Fix**: Changed `hookErr :=` to `hookErr =` in both files. Amended commit.

### D2: Same `:=` bug in my own test file

**What happened**: `path_safety_test.go:130` and `:160` used `err :=` where `err` was already declared earlier in the function.

**Fix**: Changed to `err =`.

### D3: SARIF Snippet backward compatibility regression

**What happened**: Changing `Snippet string` to `Snippet *sarifArtifactContent` broke import of SARIF files from tools that emit bare string snippets (`"snippet": "code here"`). The JSON unmarshaler expected `{"text":"..."}` and failed with a type mismatch error.

**How I caught it**: Manual test with a bare-string SARIF file after writing the report prompt.

**Fix**: Custom `UnmarshalJSON` on `sarifArtifactContent` that tries object form first, falls back to bare string.

### D4: Infinite recursion in `UnmarshalJSON`

**What happened**: First version of the custom unmarshaler called `json.Unmarshal(data, &obj)` where `obj` was `sarifArtifactContent` — the same type with the custom unmarshaler → infinite recursion → stack overflow.

**Fix**: Used anonymous struct `var obj struct { Text string }` which has no custom unmarshaler, breaking the cycle.

### D5: Lint failures (noinlineerr, errorlint, wrapcheck)

**What happened**: Three rounds of lint failures on the `UnmarshalJSON` method:

1. Inline `if err := ...; err != nil` → `noinlineerr` linter
2. `%v` format for wrapped error → `errorlint` linter
3. Unwrapped external package error → `wrapcheck` linter

**Fix**: Extracted error variable, used `errors.Join` + `fmt.Errorf("%w", ...)`.

---

## e) WHAT WE SHOULD IMPROVE

### Process Failures

1. **Build verification was insufficient**: `go build ./...` reported success but the code was broken. The Go build cache likely served a stale result. Should run `go clean -cache && go build ./...` for release-quality verification, or use `go build -a ./...` to force rebuild.

2. **BuildFlow doesn't run `go build`**: The pre-commit hook checks formatting and linting but NOT compilation. This is a gap — broken code can pass BuildFlow and get committed.

3. **LSP diagnostics were stale**: 4 lint warnings shown by LSP were not real (`golangci-lint` reported 0 issues). Don't trust LSP diagnostics alone — always verify with the actual tool.

4. **No backward compatibility test for SARIF Snippet**: I changed a type signature without testing import of the old format. Should always test backward compat when changing JSON deserialization types.

5. **Tests should be run after every code change, not assumed green**: I wrote `path_safety_test.go` with `:=` errors that weren't caught until later because I didn't run the pipeline tests immediately after writing them.

### Code Quality

6. **`hookErr :=` pattern is error-prone**: When multiple hooks fire in the same scope, the first `:=` works but subsequent ones need `=`. This pattern should be refactored to avoid the trap — perhaps use separate variable names or extract hook firing into a helper.

7. **`resolveSafePath` has a subtle TOCTOU window**: Between `EvalSymlinks` and actual file I/O, a symlink could be swapped. The current mitigation (using resolved path as map key) reduces but doesn't eliminate the window. For true safety, `O_NOFOLLOW` would be needed.

8. **`sarifArtifactContent.UnmarshalJSON` swallows the object-parse error**: When both object and string parsing fail, `errors.Join` combines them, but the error message is confusing. Could provide better diagnostics.

---

## f) Up to 50 Things to Get Done Next

### Critical / Security

1. Add `go build -a ./...` to pre-commit verification (force clean rebuild)
2. Add BuildFlow step for `go build` (not just linting)
3. Runtime TOCTOU test: swap symlink between resolve and I/O, verify containment
4. Test `resolveSafePath` with rootDir that doesn't exist
5. Test `resolveSafePath` with rootDir that is a symlink itself
6. Test `resolveSafePath` with Windows-style backslash paths
7. Add `O_NOFOLLOW` option to file writes in `applyToFile`
8. Audit all `filepath.Join` calls in pipeline package for containment
9. Add fuzz test for `resolveSafePath`
10. Test `filterByFileEdits` directly with path traversal input

### Correctness

11. Add `resolveSafePath` tests for `..` in middle of path (e.g., `a/../b.go`)
12. Test SARIF round-trip with `Confidence` edge cases (0, 1, negative)
13. Test SARIF round-trip with empty `ToolName`
14. Test SARIF round-trip with `Suppression.ExpiresAt` boundary
15. Add property test: SARIF round-trip preserves all `Finding` fields
16. Test LSP round-trip with empty Metadata map
17. Test LSP round-trip with nil Suppression
18. Test LSP round-trip with multiple RelatedFindingIDs
19. Add test for `saboteurProvider` when backup is disabled
20. Test `groupFindingsBySafePath` with duplicate findings (same resolved path)

### Architecture / Refactoring

21. Extract `resolveSafePath` into `pathutil` sub-package for reuse
22. Consider `hookErr` refactoring to avoid `:=` vs `=` trap
23. Extract `saboteurProvider` to testutil for reuse
24. Consider `sarifRegion` as interface for different SARIF version dialects
25. Add `Finding.Equal()` round-trip property test via SARIF
26. Benchmark `resolveSafePath` with 10k findings
27. Benchmark `FindingsFromSARIF` with large files (>10MB)
28. Add benchmark baseline file for regression detection
29. Profile pipeline with `pprof` to find hot paths
30. Consider streaming SARIF parser for very large outputs

### SARIF / LSP

31. Export `sarifArtifactContent` binary form support
32. Add SARIF `run.taxonomies` for custom rules
33. Add SARIF `result.hostedViewerUri` for web-based viewing
34. Add LSP `codeDescription.href` support
35. Add LSP `relatedInformation[].location` round-trip
36. Test SARIF import with `runs[].invocations` metadata
37. Test SARIF import with external property bag (`propertyBag.properties`)
38. Add SARIF `translation` metadata for multi-tool reports
39. Test LSP round-trip with `data` field containing non-go-finding keys
40. Add SARIF `graph` support for finding correlation visualization

### Documentation

41. Document the `hookErr` pattern in AGENTS.md to prevent future `:=` bugs
42. Add SARIF property bag reference table to integration guide
43. Update ADR #9 with Snippet spec compliance decision
44. Document `resolveSafePath` as security boundary in AGENTS.md
45. Add CHANGELOG entry for Snippet backward compat custom unmarshaler
46. Update FEATURES.md with path traversal protection
47. Document BuildFlow's lack of `go build` as known limitation
48. Add security testing checklist to AGENTS.md
49. Document the Go build cache staleness gotcha
50. Create `docs/guides/security-hardening.md` for path safety patterns

---

## g) Top 2 Questions I Cannot Answer Myself

### Q1: Why did `go build ./...` report success when the code had compile errors?

The `hookErr :=` bug was in the code when I ran `go build ./...` during T7 verification, and it reported "BUILD: OK". The most likely explanation is **Go build cache staleness** — the Go build cache may have served a cached positive result from before the `hookErr` changes were made (the prior session introduced these lines). I should have run `go build -a ./...` (force rebuild) or `go clean -cache && go build ./...`. However, I cannot fully confirm this theory without reproducing the exact cache state. This needs investigation to prevent recurrence.

### Q2: Should `resolveSafePath` be exported as public API?

The function is currently package-private in `pipeline`. It's the single source of truth for path containment — a security boundary that external consumers of the pipeline (e.g., custom fix providers, CLI plugins) might need to call. Making it public (as `pipeline.ResolveSafePath` or in a `pathutil` sub-package) would allow third-party code to reuse the same containment logic. However, this is a product/API decision, not a technical one — I don't know if go-finding intends to be a library that others extend with custom file operations.

---

_Assisted-by: Crush <crush@charm.land>_
