# Status Report: Deterministic JSON/SARIF Output Fix — 2026-08-06 19:40

> **RESOLVED:** All items in this report were completed in v1.5.0 (2026-08-06). The version was bumped to v1.5.0 (not v1.4.2) because `ValidateAll` and `FlightRecorderHook` shipped in the same release. See `docs/status/2026-08-06_22-21_v1-5-0-release-completion-and-self-critique.md` for the full resolution. All 8 determinism regression tests proven to fail without the flag. Tags `v1.5.0`, `pipeline/v1.5.0`, `analysis/v1.5.0`, `cmd/go-finding/v1.5.0` pushed to remote.

## Session Goal

Fix the root cause of non-deterministic JSON/SARIF output: `encoding/json/v2` serializes Go map keys in unspecified order by default (unlike v1's alphabetical sort). Map fields (`Finding.Metadata`, `Summary.BySeverity`, `Summary.ByCategory`, SARIF `properties` bag) serialized in random order on every run, making every consumer that snapshots or diffs SARIF/JSON output flaky.

**User's original ask:** ~~Add `json.Deterministic(true)` to the ~3 marshal calls, tag v1.4.2, bump consumer go.mod, delete consumer-side `normalizeJSON` hack.~~ Tagged as v1.5.0 (minor bump, not patch) because new public API shipped alongside.

---

## a) FULLY DONE

| #  | Item                                                                                                             | Evidence                                         |
| -- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| 1  | Added `json.Deterministic(true)` to all 8 production marshal calls                                               | `json.go` (6 calls), `sarif_export.go` (2 calls) |
| 2  | Verified no marshal calls in `cmd/`, `analysis/` modules                                                         | grep returned no matches                         |
| 3  | Verified `pipeline/fix_edit.go` does NOT need the flag (`fixEditJSON` has no maps — only `int`, `int`, `[]byte`) | Manual inspection of struct definition           |
| 4  | Bumped version from v1.4.1 → v1.4.2                                                                              | `version.go:12`                                  |
| 5  | CHANGELOG.md updated with `[1.4.2]` section + version links at bottom                                            | Both edits applied                               |
| 6  | AGENTS.md updated with determinism requirement note (any new marshal call MUST pass `json.Deterministic(true)`)  | Added at line 146                                |
| 7  | Fixed golines formatting issue (lines >120 chars on PrettyJSON/PrettyJSONFiltered)                               | Wrapped to multi-line calls                      |
| 8  | Build passes                                                                                                     | `go build ./...` — clean                         |
| 9  | All tests pass with `-race -count=1` across all 4 modules                                                        | Core, Pipeline, Analysis, CLI                    |
| 10 | GOWORK=off isolation test passes for core module                                                                 | Verified per-module replace directives           |
| 11 | Linter passes — 0 issues                                                                                         | `golangci-lint run ./...`                        |
| 12 | Committed (auto-git daemon)                                                                                      | `dbb4d0e` (code), `03f7786` (docs)               |

### Files Changed

| File              | Change                                                                                                                                                       |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `json.go`         | 6 marshal calls: `MarshalJSON`, `PrettyJSON`, `PrettyJSONFiltered`, `LineJSON`, `Finding.WriteJSON`, `Report.WriteJSON` — all got `json.Deterministic(true)` |
| `sarif_export.go` | 2 marshal calls: `ToSARIFWithOpts` (Marshal), `WriteSARIFWithOpts` (MarshalWrite) — both got `json.Deterministic(true)`                                      |
| `version.go`      | `VersionPatch` 1 → 2                                                                                                                                         |
| `CHANGELOG.md`    | New `[1.4.2]` section + `[1.4.2]` / `[1.4.1]` compare links                                                                                                  |
| `AGENTS.md`       | New gotcha entry about deterministic JSON requirement                                                                                                        |

---

## b) PARTIALLY DONE

Nothing. Everything I started, I finished.

---

## c) NOT STARTED

| #     | Item                                       | Why                                                                                                                                               | Impact                             |
| ----- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| ~~1~~ | ~~**Git tag `v1.4.2` was NOT created**~~   | ~~I bumped `version.go` but never ran `git tag v1.4.2`.~~ **DONE** — Tagged as `v1.5.0` instead (minor bump). All 4 module tags pushed to remote. |                                    |
| 2     | **Consumer repo go.mod bump**              | This is in a different repo (not go-finding). The user mentioned it but it's out of scope for this repo.                                          | Deferred to consumer repo session. |
| 3     | **Consumer repo `normalizeJSON` deletion** | Same — different repo.                                                                                                                            | Deferred to consumer repo session. |
| 4     | **`scripts/version-check.sh` not run**     | AGENTS.md says to run it after version changes. Would verify `version.go` matches git tag — but since no tag exists, it would fail.               | Medium — should run after tagging. |

---

## d) TOTALLY FUCKED UP

| #     | What                                                | Severity                                                                                                                               | Why It Matters |
| ----- | --------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| ~~1~~ | ~~**No determinism regression test**~~              | ~~**HIGH**~~ **DONE** — 8 byte-identity regression tests written (`fe216d4`). All 8 proven to fail without `json.Deterministic(true)`. |                |
| ~~2~~ | ~~**Never verified Deterministic actually works**~~ | ~~Medium~~ **DONE** — Empirically verified by next session: ALL map types are non-deterministic without the flag. 8 tests prove it.    |                |
| ~~3~~ | ~~**Git tag omission**~~                            | ~~**HIGH**~~ **DONE** — Tagged as `v1.5.0` and pushed to remote.                                                                       |                |

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (blocks v1.4.2 release)

1. **Write a determinism regression test** — Marshal a `Finding` with 5+ `Metadata` keys and a `Report` with multiple severity/category entries. Assert `json.Marshal` output is byte-identical across 100 invocations. This test MUST compare raw `[]byte`, not parsed structures. Without this, the `json.Deterministic(true)` flag is an unverified claim.

2. **Create git tag `v1.4.2`** — `git tag v1.4.2` (after the regression test is committed).

3. **Run `scripts/version-check.sh`** — Verify `version.go` matches the new tag.

### Process / Habits

4. **"Fix root cause" claims need proof** — I claimed the root cause was fixed but never proved it. The proof should be: (a) a test that fails WITHOUT the fix and passes WITH it, or (b) a quick program showing non-deterministic output before / identical output after. Neither was done.

5. **Test byte-identity, not structural equality** — The entire bug class (non-deterministic serialization) is invisible to structural comparison. The test suite needs at least one test that compares raw serialized bytes. Consider a `TestDeterminism_ReportJSON_RawBytesIdentical` and `TestDeterminism_SARIF_RawBytesIdentical`.

6. **Semver tag discipline** — When bumping a version, the tag is part of the release, not optional. The version-check script exists for exactly this. Run it.

### Codebase-level

7. **Consider a package-level marshal options constant** — Instead of repeating `json.Deterministic(true)` at every call site, define a `var marshalOpts = json.Deterministic(true)` (or a `marshalOptions()` function if more options are needed). This makes it impossible to forget for new call sites — they all reference the shared options. Currently someone adding a new marshal method could easily forget the flag.

8. **AGENTS.md should mention `scripts/version-check.sh` in the release flow** — The gotcha about `--match 'v[0-9]*'` exists but the actual release procedure (bump version → tag → version-check) is only in `docs/release-procedure.md`, not summarized in AGENTS.md's Important Behaviors.

---

## f) Next 50 Things to Get Done

### Release blockers (do first)

1. Write `TestDeterminism_Finding_RawBytesIdentical` — marshal a Finding with 5+ Metadata keys 100x, assert byte-identical
2. Write `TestDeterminism_Report_RawBytesIdentical` — marshal a Report with multiple severity/category entries 100x, assert byte-identical
3. Write `TestDeterminism_SARIF_RawBytesIdentical` — `ToSARIF()` 100x, assert byte-identical
4. Verify those 3 tests FAIL when `json.Deterministic(true)` is temporarily removed (proves they guard the fix)
5. Commit regression tests
6. Create `git tag v1.4.2`
7. Run `scripts/version-check.sh` to verify tag matches version.go

### Determinism hardening

8. Extract `marshalOpts` package-level constant/variable — single source of truth for `json.Deterministic(true)` so new call sites can't forget it
9. Audit `example_test.go:520` — `json.Marshal(lintOutput{...})` in example output — should it be deterministic too?
10. Consider a linter rule or grep-based CI check that flags any `json.Marshal`/`json.MarshalWrite` call missing `json.Deterministic`
11. Add determinism note to `docs/API_STABILITY.md` — mark all JSON/SARIF output methods as "deterministic since v1.4.2"
12. Consider adding `json.Deterministic(true)` to `UnmarshalJSON` calls too (harmless, ignored, but documents intent for symmetry)

### Unreleased backlog (from CHANGELOG `[Unreleased]` section — verify if these should ship in v1.4.2 or a later minor)

13. `ValidateAll` — verify it's API-complete and tested, decide if it ships in next release
14. `FlightRecorderHook` — verify it's production-ready, decide release timing
15. `Finding.Equal()` tag-order fix — this is a behavioral change; verify no consumers break
16. `FlightRecorderHook` WriteTo race fix — verify the `writeMu` solution is correct
17. `sanitizeFilename("")` fix — verify edge case is handled
18. Decide: should `[Unreleased]` items move into `[1.4.2]` or wait for `[1.5.0]`?

### Consumer repo (out of scope for go-finding, but user mentioned)

19. Bump consumer `go.mod` to `github.com/larsartmann/go-finding@v1.4.2`
20. Delete `normalizeJSON` hack from consumer test code
21. Verify consumer tests pass without the normalize hack
22. Grep consumer repo for any other post-hoc JSON normalization hacks

### Documentation

23. Update `docs/API_STABILITY.md` with determinism guarantee row for all JSON/SARIF methods
24. Add determinism section to `docs/integration-guide.md` — "Output is deterministic since v1.4.2; snapshot/diff without normalization"
25. Update `docs/USAGE_GUIDE.md` example outputs if any show non-deterministic ordering
26. Summarize release procedure in AGENTS.md (currently only in `docs/release-procedure.md`)

### Testing improvements

27. Add `TestDeterminism_WriteSARIF_RawBytesIdentical` — streaming path, not just `ToSARIF`
28. Add `TestDeterminism_PrettyJSON_RawBytesIdentical`
29. Add `TestDeterminism_PrettyJSONFiltered_RawBytesIdentical`
30. Add `TestDeterminism_LineJSON_RawBytesIdentical`
31. Add `TestDeterminism_WriteJSON_Finding_RawBytesIdentical`
32. Add `TestDeterminism_WriteJSON_Report_RawBytesIdentical`
33. Add `TestDeterminism_MarshalJSON_RawBytesIdentical` — the Report.MarshalJSON path specifically
34. Consider a table-driven determinism test that covers all 8 methods in one function
35. Add a `testing/quick` property test: for any valid Finding with Metadata, serialized output is deterministic

### Code quality

36. Check if `encoding/json/v2` has a way to set deterministic mode globally (vs per-call) — reduces boilerplate and risk of omission
37. Consider wrapping all marshal calls behind a package-internal `marshalDeterministic(v any, opts ...) ([]byte, error)` helper
38. Verify `json.Deterministic` doesn't measurably impact performance (sorting map keys) — run benchmark comparison
39. Check if `Finding.Tags` (a `[]string`, not a map) benefits from Deterministic (it shouldn't — slices are ordered)

### Release process

40. Verify sub-modules don't need their own version bumps (pipeline, analysis, CLI) — they have separate `replace` directives but share core via go.work
41. Run `nix flake check` before final release
42. Run `nix run .#test` (the canonical test command, not raw `go test`)
43. Run `nix run .#lint` (the canonical lint command)
44. Run `nix run .#bench` and compare against baseline
45. Update README.md version badge if one exists
46. Check `docs/release-procedure.md` for any steps missed
47. Verify go.sum is clean after all changes
48. Run `GOWORK=off go test ./...` in each sub-module dir (pipeline, analysis, cmd/go-finding) to verify replace directives work for consumers
49. Consider a GitHub release with CHANGELOG excerpt
50. Push tags to remote (`git push --tags` — but only with explicit user approval per project rules)

---

## g) Questions

### Q1: Should I create the `git tag v1.4.2` now, or do you want to review/write the regression test first?

The tag currently does NOT exist. I bumped `version.go` and updated CHANGELOG, but forgot `git tag v1.4.2`. I can create it immediately, or I can write the determinism regression test first (which should arguably be part of the v1.4.2 release commit). Your call.

### Q2: Should the `[Unreleased]` items (ValidateAll, FlightRecorderHook, tag-order fix) ship in v1.4.2 or wait for a future minor release?

The CHANGELOG has a `[1.4.2]` section with ONLY the determinism fix. The `[Unreleased]` section still contains ValidateAll, FlightRecorderHook, and 3 other fixes. I need to know whether v1.4.2 is a quick patch (just determinism) or whether those items should move into `[1.4.2]`. I cannot determine this without knowing if those features are ready for release — I'd need to audit each one.

### Q3: Where is the consumer repo that has the `normalizeJSON` hack?

The original message references deleting `normalizeJSON` from a test, but this repo (go-finding) has no `normalizeJSON`. The consumer repo is presumably elsewhere on disk. I need its path to bump go.mod and delete the hack.

---

## Summary

The root cause fix is **correct in principle** — `json.Deterministic(true)` was added to all 8 production marshal calls, the build passes, tests pass, lint passes. But the session has **two critical gaps**: (1) no regression test proving the fix works, and (2) no git tag despite the user explicitly requesting `v1.4.2`. The fix is committed but the release is incomplete.

**Verdict: B+.** The fix is right, but the proof is missing and the tag is missing. A fix without a test is a claim; a version bump without a tag is a promise unkept.

---

_Assisted-by: Crush_
