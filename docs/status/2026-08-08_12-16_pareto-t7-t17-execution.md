# Status Report: Pareto Plan Execution — T7-T17 Session

**Date:** 2026-08-08 12:16 CEST
**Session:** Continued Pareto plan execution — T7, T8, T9, T16, T17 + session debt fixes
**Base:** v1.5.0 (2026-08-06) → 24 commits ahead, unreleased
**Prior:** T1-T6 critical path completed in earlier session (see `docs/status/2026-08-08_11-47_pareto-critical-path-execution.md`)

---

## A. FULLY DONE (Verified, Tests Pass, Committed)

### Session Debt Fixes (D1, E1 from prior status report)

- **D1 fix: doc.go guide reference** — Restored a soft reference to the FlightRecorder guide ("see the FlightRecorder guide in the project documentation") instead of a filesystem path. Updated the field list from 3 fields to all 5 (`enabled`, `outputDir`, `slowStageThreshold`, `minAge`, `maxBytes`). Committed in `0eaa227`.
- **E1 fix: TODO_LIST.md CI script notes** — Updated all 4 CI script entries (lines 57-60) to mention their ci.yml job wiring. Changed notes from describing only the script to noting the job name. Committed in `0eaa227`.

### T7: FlightRecorderFileConfig Field Parity

- **Files:** `cmd/go-finding/config.go`, `cmd/go-finding/main.go`, `cmd/go-finding/integration_test.go`
- **What changed:**
  - Added `MinAge string` and `MaxBytes uint64` fields to CLI `flightRecorderFileConfig` struct (was 3 fields, now 5 — full parity with `pipeline.FlightRecorderFileConfig`)
  - Extended `validate()` to parse-check `minAge` as a duration string
  - Wired `minAge` and `maxBytes` through `main.go`'s config-file flight recorder setup path (parsing duration, setting on `frConfig`)
  - Added 2 test cases: `invalid flightRecorder minAge` (error path), `valid flightRecorder config with all fields` (happy path with all 5 fields)
- **Verified:** `go test -race -count=1 -run TestPipelineConfigFile_Validate ./cmd/go-finding/...` passes
- **Committed in:** `77eeb07` (code), `0eaa227` (docs)

### T16: resolveSafePath Edge Case Tests

- **Files:** `pipeline/path_safety_test.go` (+~110 lines)
- **Added 5 tests:**
  - `TestResolveSafePath_CircularSymlink` — Self-referential symlink (`loop.go → loop.go`) does not infinite-loop or panic
  - `TestResolveSafePath_MutualCircularSymlink` — Mutual circular (`a.go → b.go → a.go`) handled gracefully
  - `TestResolveSafePath_DanglingSymlink_OutsideRoot` — Broken symlink pointing to non-existent file outside root; unresolved path is inside root so returns safe
  - `TestResolveSafePath_RootIsSymlink` — rootDir is a symlink to another directory; file resolves to real path
  - `TestResolveSafePath_RootIsSymlink_PathTraversal` — Path traversal blocked even through symlinked root
- **Verified:** `go test -race -count=1 -run TestResolveSafePath ./pipeline/...` passes
- **Committed in:** `0bf35f6`

### T17: FlightRecorder Edge Case Tests

- **Files:** `pipeline/flight_recorder_test.go` (+~90 lines)
- **Added 3 tests:**
  - `TestFlightRecorderHook_SnapshotWriteError` — Snapshot to read-only dir returns error (simulates disk-full/permission denied). Skips when running as root (CI).
  - `TestFlightRecorderHook_SlowLastStageInPipeline` — SlowStageThreshold triggers snapshot on the verify (last) stage during a real pipeline run, not just isolated event calls
  - `TestFlightRecorderHook_ConcurrentStageEventsSafe` — 5 goroutines fire Before/After events for all 5 stages with 5 iterations each; verified with `-race` detector
- **Verified:** `go test -race -count=1 -run TestFlightRecorderHook ./pipeline/...` passes
- **Committed in:** `0bf35f6`

### T8: Docs-Freshness CI Script

- **Files:** `scripts/docs-freshness.sh` (new, ~65 lines), `.github/workflows/ci.yml` (+10 lines)
- **What was built:**
  - `docs-freshness.sh` — Two checks: (1) staleness (docs not modified in N days, default 180), (2) code-doc sync (docs whose referenced `.go` files were modified after the doc). Exits 0 always (informational warnings, not a gate). Uses `git log --format="%ct"` for timestamps.
  - ci.yml `docs-freshness` job — Runs the script with `fetch-depth: 0` for full git history
- **Verified:** Script runs locally, produces 7 out-of-sync warnings (pre-existing docs that reference source files modified since last doc update). Exits 0.
- **Committed in:** `afd71e0`

### T9: Per-Module CHANGELOGs

- **Files:** `pipeline/CHANGELOG.md`, `analysis/CHANGELOG.md`, `cmd/go-finding/CHANGELOG.md` (all new), `CHANGELOG.md` (updated)
- **What was created:**
  - 3 sub-module CHANGELOGs with `[Unreleased]` and `[1.5.0]` sections, cross-referenced to root
  - Root CHANGELOG.md updated with sub-module cross-reference block at top
  - Root CHANGELOG.md `[Unreleased]` updated: CLI flightRecorder entry now lists all 5 fields; added docs-freshness script, per-module CHANGELOGs, resolveSafePath edge case tests, FlightRecorder edge case tests entries; changed doc.go entry to reflect the updated field list
- **Verified:** All files exist; cross-reference links resolve
- **Committed in:** `afd71e0`, `0eaa227`

### AGENTS.md Updates

- Updated CLI config-file flight recorder gotcha: changed from "3 fields, intentionally simpler" to "5 fields, full parity, keep in sync"
- Updated CI scripts gotcha: changed from "Four scripts" to "Five scripts" (added docs-freshness.sh)
- **Committed in:** `0eaa227`

### TODO_LIST.md Updates

- Marked `Docs-freshness CI check` as DONE with ci.yml job note
- Marked `Per-module CHANGELOG entries` as DONE with file paths
- Updated all 4 CI script entries (go-work-sync, replace-audit, version-drift, test-naming) to mention ci.yml wiring

### Full Verification Suite

- **Build:** `go build ./...` passes (all 4 modules via go.work)
- **GOWORK=off build:** All 4 modules build standalone (consumer isolation)
- **Test:** `go test -race -count=1 ./...` passes for all 4 modules
- **Lint:** `golangci-lint run ./...` — 0 issues (all 4 modules)
- **CI scripts:** All 5 scripts pass locally (replace-audit, version-drift, test-naming, go-work-sync, docs-freshness)

---

## B. PARTIALLY DONE

### T7.5: docs/guides/flight-recorder.md examples NOT updated

The guide's **Configuration Reference table** (lines 199-207) correctly lists all 5 config file fields including `minAge` and `maxBytes`. But the **YAML and JSON examples** (lines 62-80) still show only 3 fields:

```yaml
flightRecorder:
  enabled: true
  outputDir: "./traces"
  slowStageThreshold: "30s"
```

These should be updated to include `minAge` and `maxBytes` examples now that the CLI has full parity. The table is correct; the examples are stale.

### No E2E test for CLI config file with minAge/maxBytes

T7 added validation tests (unit level) for the new `minAge`/`maxBytes` fields, and wired them through `main.go`. But there is no end-to-end test that loads a YAML/JSON config file with these fields and verifies they reach the pipeline `FlightRecorderConfig`. The existing `TestRun_E2E_TraceViaConfigFile` only tests with the old 3-field config.

### docs-freshness.sh produces noisy warnings

The script's `.go` file reference grep (`grep -oP '`?\K[a-zA-Z0-9_/]+\.go'`) is very broad. It matches:

- Code examples in markdown (e.g., `main.go` in a tutorial) that aren't real file references
- Filenames in prose that may not exist relative to the doc

This produces false-positive "out-of-sync" warnings. The script works (exit 0) but the warnings are noisier than ideal. A more precise parser would extract only paths in code spans or link syntax.

---

## C. NOT STARTED (From the Pareto Plan)

### Do Next (High Value)

1. **T10** — go-arch-lint module boundary CI enforcement
2. **T11** — Multi-module vs monolith benchmark comparison
3. **T12** — docs/guides/configuration.md (central config reference)
4. **T13** — docs/guides/troubleshooting.md
5. **T14** — README.md FlightRecorder mention in feature table
6. **T15** — docs/DOMAIN_LANGUAGE.md

### Testing Gaps

7. **T18** — sanitizeFilename property-based / fuzz test
8. **T19** — SARIF schema validation test

### Launch Track (Needs Repo Public)

9. **T20** — GoReleaser + Homebrew verification
10. **T21** — pkg.go.dev verification
11. **T22** — Launch announcement
12. **T23** — Submit to Awesome Go

### Strategic (T24-T35)

13-24. FlightRecorder v2 (rotation, gzip, pprof, OTel, trace diff), ecosystem expansion (consumer migration guide, ToolAdapter recipes, LSP code actions, language providers, AI remediation), v2.0 hardening (Position sentinel, FixStrategy union, TagSet, Finding sub-structs)

---

## D. TOTALLY FUCKED UP

### D1: Missed the flight-recorder.md YAML/JSON example update

T7 step 7.5 explicitly says "Update docs/guides/flight-recorder.md config table if needed." I checked the **Configuration Reference table** (correct, has all 5 fields) but did NOT check or update the **YAML/JSON examples** earlier in the same file (lines 62-80), which still show only 3 fields. This is a split-brain: the table says 5 fields, the example shows 3. A consumer copying the example gets an incomplete config.

### D2: docs-freshness.sh bash 4+ dependency (mapfile)

The script uses `mapfile` (bash 4+ feature). macOS ships bash 3.2 by default. The `docs-freshness` CI job runs on `ubuntu-latest` (bash 5+), so CI is fine. But if anyone runs this script locally on macOS without updated bash, it silently fails. Should either add a bash version check or use a POSIX alternative.

### D3: No test for docs-freshness.sh itself

The docs-freshness script has zero test coverage. It's a CI script that could break silently. The other 4 CI scripts also lack dedicated tests, but docs-freshness is the most complex (git timestamp parsing, grep extraction, threshold comparison). A simple smoke test that verifies exit code 0 on a known-good repo would prevent silent breakage.

### D4: Concurrent FlightRecorder test has timing sensitivity

`TestFlightRecorderHook_ConcurrentStageEventsSafe` uses `time.Sleep(2 * time.Millisecond)` inside goroutines to exceed the 1ms `SlowStageThreshold`. On a heavily loaded CI runner, goroutine scheduling delays could cause the Before/After pairing to race (Before from iteration N might pair with After from iteration N-1 if scheduling reorders them). The test verifies race-safety, not snapshot count accuracy, so this is unlikely to flake — but the `g.Expect(traceFiles).ToNot(gomega.BeEmpty())` assertion could fail if all 25 Before/After pairs somehow complete within the 1ms window on a very fast machine.

### D5: Didn't add CLI Features section update for minAge/maxBytes

The AGENTS.md "CLI Features" section mentions `-trace` flags but the config-file `flightRecorder` section description doesn't mention the now-available `minAge`/`maxBytes` fields. This is minor — the gotcha entry was updated — but the feature list is stale.

---

## E. WHAT WE SHOULD IMPROVE

### E1: YAML/JSON config examples are split-brain with the reference table

The flight-recorder.md guide has a Configuration Reference table (line 199) that lists all 5 config file fields. But the YAML example (line 62) and JSON example (line 72) only show 3. This is the most visible documentation gap from this session. Fix: add `minAge` and `maxBytes` to both examples.

### E2: docs-freshness.sh is a good start but needs refinement

The script's code-doc sync check is too broad — it greps all `.go` filenames from markdown, including code examples in prose. A better approach would be to only check paths inside code spans (` ``filename.go`` `) or fenced code blocks. This would dramatically reduce false-positive warnings.

### E3: No go-arch-lint module boundary enforcement

This was T10 in the Pareto plan and remains the highest-impact unfinished task. The multi-module architecture's #1 structural risk (cross-module coupling, e.g., pipeline importing from cmd/go-finding) is unenforced. All other module-level CI checks (replace directives, version drift, workspace sync) are wired; boundary enforcement is the missing piece.

### E4: CLI Features section in AGENTS.md doesn't mention config-file minAge/maxBytes

The section at line ~158 lists `-trace`, `-trace-dir`, `-trace-slow` flags and config-file `flightRecorder` section, but doesn't enumerate the 5 available fields. Should mention `minAge`/`maxBytes` for discoverability.

### E5: The 3 open questions from the prior status report are still unanswered

The prior status report (section G) asked:

- G1: Push now or batch?
- G2: FlightRecorderFileConfig parity direction? (now resolved — chose full parity)
- G3: version-drift.sh grep vs `go mod edit -json`?

G2 is resolved (we added the fields). G1 and G3 remain open.

### E6: Per-module CHANGELOGs need a maintenance plan

The 3 new CHANGELOGs are created but there's no CI check ensuring they're updated when sub-module changes are made (unlike the root CHANGELOG which has `changelog-check` job). Without enforcement, they'll go stale.

### E7: FlightRecorder concurrent test could use sync.AdvanceTime or mock clock

Instead of `time.Sleep(2 * time.Millisecond)`, a mock clock or `testing.T.Deadline()`-aware approach would eliminate timing sensitivity entirely. This is a minor improvement — the test works today.

---

## F. Next 50 Things to Get Done (Prioritized)

### Immediate (This Session's Debt)

1. Update flight-recorder.md YAML example to include `minAge` and `maxBytes`
2. Update flight-recorder.md JSON example to include `minAge` and `maxBytes`
3. Update AGENTS.md CLI Features section to enumerate config-file flightRecorder fields
4. Add E2E test: CLI config file with minAge/maxBytes reaching pipeline FlightRecorderConfig

### CI Quality Hardening (T10-T11)

5. **T10:** Research go-arch-lint capabilities and config format
6. **T10:** Write go-arch-lint YAML config defining module boundaries (no pipeline→cmd, no analysis→cmd, etc.)
7. **T10:** Run go-arch-lint locally, fix any violations
8. **T10:** Add go-arch-lint job to ci.yml
9. **T11:** Create single-module variant (temp flatten go.work)
10. **T11:** Run benchmark on multi-module workspace
11. **T11:** Run benchmark on single-module variant
12. **T11:** Compare with benchstat
13. **T11:** Write results to docs/reports/

### Documentation (T12-T15)

14. **T12:** Inventory all config-file options across all modules
15. **T12:** Write docs/guides/configuration.md (central config reference)
16. **T12:** Add YAML + JSON examples for each config section
17. **T12:** Cross-reference from README and other guides
18. **T13:** Collect common error messages from pipeline code
19. **T13:** Write docs/guides/troubleshooting.md (what/why/fix format)
20. **T14:** Add FlightRecorder row to README.md feature table
21. **T14:** Add config-file integration mention to README.md
22. **T15:** Write docs/DOMAIN_LANGUAGE.md (Finding, Report, Pipeline, Stage, etc.)

### Testing Gaps (T18-T19)

23. **T18:** Write fuzz test for sanitizeFilename
24. **T18:** Run fuzz for 1 minute to verify stability
25. **T19:** Evaluate SARIF schema vendoring vs lightweight JSON validator
26. **T19:** Implement SARIF schema validation test

### docs-freshness.sh Improvements

27. Add bash version check to docs-freshness.sh (fail fast on bash < 4)
28. Refine `.go` file reference grep to only match code spans/fenced blocks
29. Add smoke test for docs-freshness.sh (verify exit 0 on known repo)
30. Consider adding `--fail-on-stale` flag for opt-in gating

### Release Prep

31. Resolve G1: Push to remote and verify GitHub Actions runs all new jobs
32. Resolve G3: Decide on version-drift.sh grep vs `go mod edit -json`
33. Tag v1.5.1 (patch: FlightRecorder config-file integration + CI hardening)
34. Tag sub-module versions: pipeline/v1.5.1, analysis/v1.5.1, cmd/go-finding/v1.5.1
35. Run scripts/version-check.sh against new tags
36. Verify `go get github.com/larsartmann/go-finding@v1.5.1` resolves

### Per-Module CI Quality

37. Add changelog-check equivalent for sub-module CHANGELOGs
38. Add paths-ignore verification for docs-freshness job (confirm markdown changes trigger it)
39. Consider splitting structural-checks job into 3 parallel jobs for faster feedback
40. Add go-version matrix to docs-freshness (verify on multiple bash versions)

### Launch Track (T20-T23, Needs Repo Public)

41. **T20:** Verify GoReleaser + Homebrew tap works on public tag
42. **T21:** Verify pkg.go.dev renders after first public tag
43. **T22:** Write launch announcement (blog/r/golang/Slack)
44. **T23:** Submit to Awesome Go

### FlightRecorder v2 (T24-T26)

45. **T24a:** Design trace file rotation config (MaxFiles, MaxTotalBytes)
46. **T24b:** Add gzip compression to trace output
47. **T25a:** Add pprof capture alongside trace snapshots
48. **T26a:** Design OpenTelemetry span bridge from trace data

### v2.0 Hardening (T32-T35)

49. **T32:** Design Option[T] position type (eliminate -1 sentinel)
50. **T35:** Design Identity/Location/Classification/Fix sub-structs for Finding

---

## G. Questions (Cannot Resolve Without User Input)

### G1: Push to remote now?

All commits from both sessions (T1-T6 and T7-T17) are local-only. GitHub Actions has not validated any of the 7 new/updated CI jobs (`structural-checks`, `go-work-sync`, `docs-freshness`, and the existing jobs that now run with updated scripts). Should I push now to validate, or wait for more work to batch?

### G2: go-arch-lint (T10) — install as Go tool or use Docker image?

go-arch-lint can be installed via `go install` (adds a build step to CI) or run via Docker image (adds Docker layer pull). The existing CI pattern uses `go install` for tools like `govulncheck`, `art-dupl`, and `benchstat`. Should I follow the existing `go install` pattern, or is there a preference for the Docker approach?

### G3: version-drift.sh — keep the grep fix or upgrade to `go mod edit -json | jq`?

The prior status report raised this. The `grep -v '// indirect'` fix works but is fragile. `go mod edit -json | jq` would be robust but adds a `jq` dependency to CI (though `jq` is pre-installed on GitHub Actions runners). Should I upgrade the script for robustness, or keep the pragmatic grep approach?

---

_Assisted-by: Crush <crush@charm.land>_
