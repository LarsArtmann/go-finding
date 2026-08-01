# Status Report: Pipeline Lint Cleanup — 7 Issues Fixed

**Date:** 2026-07-28 14:01
**Session scope:** Fix the 7 pre-existing pipeline module lint findings from the dedup-to-zero sweep
**Status:** COMPLETED with caveats

---

## a) FULLY DONE

### 7 lint issues resolved (all verified by `golangci-lint run ./...`)

| #   | Linter        | Issue                                                               | Fix Applied                                                                                                                                          | File                                   |
| --- | ------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| 1   | `funlen`      | `Run()` 137 lines (>120)                                            | Extracted `runVerification()` method — fires before/after hooks around `Verify()` call, sets `result.Verification`                                   | `pipeline/pipeline.go`                 |
| 2   | `funlen`      | `runIteration()` 124 lines (>120)                                   | Extracted `applyStage()` method — DryRun guard + before/after apply hooks + `applyTriage` call                                                       | `pipeline/pipeline_iteration.go`       |
| 3   | `gocognit`    | `applyTriage` complexity 36 (>35)                                   | Extracted `shiftFindingsPositions()` + `shiftFindingSlice()` — eliminated duplicated position-shifting loops over `iter.findings` and `iter.suggest` | `pipeline/pipeline_detect.go`          |
| 4   | `exhaustruct` | `result{ok: false}` missing fields                                  | Explicit zero values: `result{fset: nil, file: nil, ok: false}`                                                                                      | `pipeline/goast/provider.go:129`       |
| 5   | `wrapcheck`   | `g.Wait()` error unwrapped                                          | Wrapped: `fmt.Errorf("detect: %w", err)`                                                                                                             | `pipeline/convenience.go:67`           |
| 6   | `gosec G703`  | Path traversal in test `os.Remove(bakPath)`                         | `//nolint:gosec` with justification (intentional test saboteur)                                                                                      | `pipeline/fix_applier_test.go:578`     |
| 7   | `revive`      | Unused `s` receiver on `saboteurProvider.Name()` and `.CanHandle()` | Changed `(s *saboteurProvider)` to `(*saboteurProvider)`                                                                                             | `pipeline/fix_applier_test.go:570-571` |

### Verification completed

- `golangci-lint run ./...` on pipeline module — all 7 target issues gone
- `go build ./...` — clean
- `go test -race -count=1 ./...` (workspace) — all pass
- `GOWORK=off go test ./...` (pipeline module isolation) — all pass

---

## b) PARTIALLY DONE

### Refactoring quality

The 3 extracted methods (`runVerification`, `applyStage`, `shiftFindingsPositions`) are correct and tested via existing integration tests, but:

- **No dedicated unit tests** were written for the new functions. They are exercised indirectly through existing pipeline tests, but a regression in `shiftFindingsPositions` or `applyStage` error paths would only surface in integration tests, not focused unit tests.
- **No benchmarks** were run to verify the refactoring introduced no performance regression. The extraction adds one function call indirection per iteration, which should be negligible, but this was not measured.

---

## c) NOT STARTED

1. **TODO_LIST.md update** — The 7 tasks from the paste should be marked done. Not touched.
2. **AGENTS.md update** — New helper functions (`runVerification`, `applyStage`, `shiftFindingsPositions`, `shiftFindingSlice`) are not documented in the "Key Files" or "Important Behaviors" sections.
3. **Unit tests for extracted functions** — Not written.
4. **Benchmark regression check** — Not run.
5. **CHANGELOG.md entry** — No entry for this cleanup.

---

## d) TOTALLY FUCKED UP

Nothing catastrophically broken. But one significant oversight:

### paralleltest config drift — DISCOVERED BUT NOT FIXED

**The AGENTS.md explicitly states:**

> `paralleltest` linter is disabled (intentional) — `.golangci.yml` disables `paralleltest` because `t.Parallel()` is centralized inside `NewParallelGomega` helpers across all 4 modules.

**Reality:** `.golangci.yml` line 93 has `- paralleltest` in the `enable` list. The linter is ENABLED, not disabled. This produces **100 findings** across the pipeline module alone.

This means either:

- The AGENTS.md documentation is lying (the linter was re-enabled but docs weren't updated), OR
- The `.golangci.yml` was accidentally changed to enable it

I flagged this in my final summary but did NOT investigate which is true, did NOT check git blame to find when it changed, and did NOT fix it. This is the kind of split-brain between docs and config that the AGENTS.md itself warns against.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (this session's work)

1. **Write focused unit tests** for `shiftFindingsPositions`, `shiftFindingSlice`, `runVerification`, and `applyStage` — especially the error paths in `runVerification` (before-hook failure, verify failure, after-hook failure) and `applyStage` (before-hook failure, applyTriage failure, after-hook failure).
2. **Run benchmark regression check** — `bash scripts/bench-check.sh` before/after to confirm the extraction didn't degrade performance.
3. **Fix the paralleltest config drift** — Either re-disable in `.golangci.yml` (matching AGENTS.md) or update AGENTS.md to reflect the new policy. Check `git log -p .golangci.yml` to find when it changed.
4. **Update TODO_LIST.md** — Mark the 7 tasks as done.
5. **Update AGENTS.md** — Add the new helper functions to the Key Files table.
6. **Add CHANGELOG.md entry** — Under "Fixed" or "Changed".

### Structural (broader observations)

7. **The `applyTriage` complexity fix was minimal** — I only shaved 1 point off (36→35 boundary). A more thorough decomposition would extract the `OnFix` callback logic (lines 248-262) into a `notifyFixCallbacks` helper, bringing complexity down further and making the function more readable.
8. **`shiftFindingSlice` could be generic** — The function takes `[]finding.Finding` but the pattern (iterate, check file match, shift position+range) could apply to any finding slice. Not worth over-engineering, but worth noting.
9. **`runVerification` returns error but callers must set `metricsResult`** — The pattern `if err != nil { metricsResult = result; return result, err }` is duplicated. The method could take `*PipelineResult` and handle the metrics assignment internally, but this would change the control flow semantics.

---

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

### Lint/Quality (high priority)

1. **Fix paralleltest config drift** — Investigate `git log -p .golangci.yml`, decide policy, align config + docs
2. **Run full workspace lint** (`golangci-lint run ./...` from root) to find issues in core, analysis, CLI modules
3. **Run `GOWORK=off go test ./...` in all 4 modules** to verify replace directives work
4. **Write unit tests for `shiftFindingsPositions`** — empty shiftMaps, single file, multiple files, no match
5. **Write unit tests for `shiftFindingSlice`** — empty slice, no file match, multiple matches
6. **Write unit tests for `runVerification`** — before-hook error, verify error, after-hook error, success
7. **Write unit tests for `applyStage`** — DryRun skip, before-hook error, applyTriage error, after-hook error, success
8. **Run benchmark regression check** before/after this refactoring
9. **Extract `OnFix` callback logic** from `applyTriage` to further reduce complexity
10. **Check all 4 modules for funlen violations** — not just pipeline
11. **Check all 4 modules for gocognit violations** — not just pipeline
12. **Run `nix flake check`** to verify the flake is healthy
13. **Run `nix run .#lint`** to use the project's canonical lint command

### Documentation

14. **Update TODO_LIST.md** — mark the 7 lint tasks as completed
15. **Update AGENTS.md Key Files table** — add `runVerification`, `applyStage`, `shiftFindingsPositions`
16. **Add CHANGELOG.md entry** for this lint cleanup
17. **Audit AGENTS.md for other config/code drift** — the paralleltest issue suggests other claims may be stale
18. **Check doc.go references** — verify no stale API references after refactoring (checked: clean)

### Testing

19. **Add table-driven tests for `runVerification` error paths**
20. **Add integration test for `applyStage` in DryRun mode** — verify hooks are NOT fired
21. **Add test for `shiftFindingsPositions` with overlapping files** in shiftMaps
22. **Run fuzz tests** (`go test -fuzz=FuzzApplyEditsToContent`) to verify refactoring didn't break edge cases
23. **Verify `StageTiming` closure called exactly once** in `applyStage` — the AGENTS.md warns about double-recording

### Architecture

24. **Consider making `runVerification` handle `metricsResult` assignment internally** to reduce caller boilerplate
25. **Review all `//nolint:gosec` suppressions** for validity and expiration
26. **Review all `//nolint:exhaustruct` suppressions** — are they still needed?
27. **Check if `applyStage` should be exported** for consumers who want custom iteration control
28. **Audit the pipeline module for other functions near funlen/gocognit limits** — prevent future violations

### CI/Build

29. **Verify `version-check.sh` passes** after changes
30. **Run `bash scripts/bench-check.sh`** with fresh baseline
31. **Verify `nix build` succeeds**
32. **Check if `golangci-lint` version matches what CI uses** — lint results may differ

### Broader codebase health

33. **Run lint on core module** — check for issues there
34. **Run lint on analysis module** — check for issues there
35. **Run lint on CLI module** — check for issues there
36. **Check for unused exports** across all modules
37. **Verify SARIF round-trip tests still pass** after any indirect changes
38. **Run `go vet ./...` across all modules**
39. **Check for `errcheck` issues** in all modules
40. **Review `dupl` findings** — any new duplication introduced?
41. **Review `gocritic` findings** across all modules
42. **Check `staticcheck` findings** — not just the CLI detector, but on the codebase itself
43. **Verify go.mod replace directives** are correct in all 4 modules
44. **Check `go.sum` is tidy** in all modules
45. **Run `golangci-lint run --fix ./...`** to auto-fix any safe issues
46. **Audit `fix_applier_test.go` for other unused receivers** beyond saboteurProvider
47. **Check if other test files have gosec G703 false positives**
48. **Review wrapcheck findings in other modules** — convenience.go pattern may repeat
49. **Consider adding `paralleltest` to CI gate or removing it entirely** — current state is ambiguous
50. **Full `nix run .#test` from root** to use the project's canonical test runner

---

## g) QUESTIONS (that I CANNOT figure out myself)

### 1. paralleltest policy: disable or fix?

The AGENTS.md says `paralleltest` is intentionally disabled because `t.Parallel()` is centralized in `NewParallelGomega`. But `.golangci.yml` has it enabled, producing 100 findings. **Which is the intended state?** If disabled: I remove it from `.golangci.yml`. If enabled: I need to either inline `t.Parallel()` in 100+ test functions or add `//nolint:paralleltest` everywhere, which contradicts the AGENTS.md rationale.

### 2. Should the 3 extracted methods get dedicated unit tests, or is integration coverage sufficient?

The existing pipeline tests exercise `runVerification`, `applyStage`, and `shiftFindingsPositions` indirectly. Writing focused unit tests for each would add ~100 lines of test code but provide better regression isolation. **Is this investment warranted given the integration tests already pass?**

### 3. Should I update TODO_LIST.md and CHANGELOG.md for this session's work, or is that handled by the auto-git daemon?

The AGENTS.md mentions "An auto-git commit daemon runs continuously and commits changes automatically." **Does this daemon also update TODO_LIST.md and CHANGELOG.md, or am I expected to do that manually for non-code documentation?**

---

## Session metrics

- **Files modified:** 5 (`pipeline.go`, `pipeline_iteration.go`, `pipeline_detect.go`, `goast/provider.go`, `convenience.go`, `fix_applier_test.go`)
- **Functions extracted:** 4 (`runVerification`, `applyStage`, `shiftFindingsPositions`, `shiftFindingSlice`)
- **Lint issues resolved:** 7/7
- **Tests:** All pass (workspace + GOWORK=off isolation)
- **Build:** Clean
- **Time:** ~15 minutes of active work

---

## Resolution (2026-08-01)

The 7 lint fixes and the `paralleltest` config drift discovery both shipped in v1.4.1 (CHANGELOG `[1.4.1]` documents all fixes). The open items from this report's self-critique — unit tests for extracted functions (`runVerification`, `applyStage`, `shiftFindingsPositions`), benchmark regression check, and AGENTS.md Key Files update for the new helpers — remain low-priority backlog. The extracted functions are covered by existing integration tests; dedicated unit tests would add depth but are not blocking.
