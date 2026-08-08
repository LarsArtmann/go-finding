# Status Report: Pareto Plan Execution — T1-T6 Critical Path

**Date:** 2026-08-08 11:47 CEST
**Session:** Executed the "Do First" critical path (T1-T6) from `docs/planning/2026-08-08_10-55_ci-hardening-and-growth-pareto-plan.md`
**Base:** v1.5.0 (2026-08-06) → 19 commits ahead, unreleased

---

## A. FULLY DONE (Verified, Tests Pass, Committed)

### T1: AGENTS.md Gotchas (3 new entries + CLI Features update)
- **Files:** `AGENTS.md` (+4 lines)
- **Added:**
  - `FlightRecorderFileConfig / ResolveFlightRecorder` — Documents the pipeline's 5-field config struct and `ResolveFlightRecorder()` method
  - `CLI config-file flight recorder fallback` — Documents the intentional 3-field vs 5-field split between CLI and pipeline structs
  - `CI scripts guard multi-module architecture` — Documents the 4 scripts and their ci.yml jobs
  - CLI Features section: added config-file `flightRecorder` section bullet
- **Verified:** `go build ./...` passes

### T2: CLI Validation Error-Path Test
- **Files:** `cmd/go-finding/integration_test.go` (+25 lines)
- **Added 2 test cases** to the existing `TestPipelineConfigFile_Validate` table-driven test:
  - `invalid flightRecorder slowStageThreshold` — verifies `"not-a-duration"` is rejected
  - `valid flightRecorder config` — verifies `"30s"` + output dir is accepted
- **Verified:** `go test -race -count=1 -run TestPipelineConfigFile_Validate ./...` passes

### T3a: doc.go FlightRecorder Reference Fix
- **Files:** `doc.go` (2 lines changed)
- **Changed:** Removed `See the Flight Recorder Guide (docs/guides/flight-recorder.md)` reference — filesystem paths in godoc are not clickable and create stale-link risk. Replaced with inline field documentation.
- **Verified:** `go build ./...` passes

### T3b: version-drift.sh grep Fix
- **Files:** `scripts/version-drift.sh` (1 line changed)
- **Changed:** Added `| grep -v '// indirect'` to the grep pipeline in `check_version()`. Prevents false positives when a module appears as both direct and indirect dependency in the same go.mod.
- **Verified:** `bash scripts/version-drift.sh` passes

### T4: GOWORK=off Per-Module Isolation Verification
- **No files changed** — verification only
- **Tested all 4 modules** with `GOWORK=off GOEXPERIMENT=jsonv2`:
  - Core (.) — build + test PASS
  - Pipeline — build + test PASS
  - Analysis — build + test PASS
  - CLI (cmd/go-finding) — build + test PASS
- **Conclusion:** Replace directives are correct. Consumer builds work without go.work.

### T5: Wire 4 CI Scripts into ci.yml
- **Files:** `.github/workflows/ci.yml` (+23 lines)
- **Added 2 new jobs:**
  - `structural-checks` — runs `replace-audit.sh`, `version-drift.sh`, `test-naming.sh` (bash-only, no Go setup needed, 5-min timeout)
  - `go-work-sync` — runs `go-work-sync.sh` (needs Go setup, 5-min timeout)
- **Verified:** All 4 scripts pass locally. YAML structure matches existing job conventions (pinned checkout SHA, pinned setup-go SHA).

### T6: CHANGELOG [Unreleased] Entry
- **Files:** `CHANGELOG.md` (+11 lines)
- **Added 8 Added entries:** FlightRecorderFileConfig, CLI flightRecorder section, 4 CI scripts, TOCTOU tests, LSP benchmarks, FlightRecorder guide, ParseConfidence, Template.Builder
- **Added 2 Changed entries:** doc.go godoc fix, version-drift.sh indirect fix
- **Note:** ParseConfidence and Template.Builder entries were already present; preserved alongside new entries.

### Full Test Suite Verification
- `go test -race -count=1 ./...` — ALL 4 MODULES PASS
- `golangci-lint run ./...` — 0 issues (all 4 modules)
- All 4 CI scripts pass locally
- `go build ./...` — clean

---

## B. PARTIALLY DONE

### TODO_LIST.md status notes
The TODO_LIST.md marks the 4 CI scripts as DONE (lines 57-60) because the scripts existed. But the notes say things like "scripts/go-work-sync.sh — runs go work sync twice" without mentioning that they're **now wired into ci.yml**. The notes should be updated to reflect the CI integration. This is a doc-staleness gap I introduced.

### CHANGELOG completeness
The CHANGELOG [Unreleased] section documents the new work, but the `### Changed` section does not mention:
- AGENTS.md updates (T1) — arguably internal, not consumer-facing
- CLI integration_test.go new test cases (T2) — internal test, not changelog-worthy
- The ci.yml changes themselves (T5) — internal CI, but could be noted

---

## C. NOT STARTED (From the Pareto Plan, "Do Next" and Beyond)

### Do Next (High Value)
1. **T7** — FlightRecorderFileConfig parity (CLI has 3 fields, pipeline has 5)
2. **T10** — go-arch-lint module boundary CI
3. **T16** — resolveSafePath edge case tests (circular symlinks, dangling symlinks, root-is-symlink)
4. **T17** — FlightRecorder edge case tests (disk-full, last-stage threshold, concurrent detectors)
5. **T8** — Docs-freshness CI check script + ci.yml job
6. **T9** — Per-module CHANGELOG entries (pipeline/, analysis/, cmd/go-finding/)

### Do Later (Medium Value)
7. **T12** — docs/guides/configuration.md (central config reference)
8. **T11** — Multi-module vs monolith benchmark comparison
9. **T13** — docs/guides/troubleshooting.md
10. **T14** — README.md FlightRecorder mention in feature table
11. **T15** — docs/DOMAIN_LANGUAGE.md
12. **T18** — sanitizeFilename fuzz test

### Do When Ready (Launch Track — Needs Repo Public)
13. **T20** — GoReleaser + Homebrew verification
14. **T21** — pkg.go.dev verification
15. **T22** — Launch announcement
16. **T23** — Submit to Awesome Go
17. **T19** — SARIF schema validation test

### Do Eventually (Strategic — T24-T35)
18-35. FlightRecorder v2 features (rotation, gzip, pprof, OTel bridge, trace diff), ecosystem expansion (consumer migration guide, ToolAdapter recipes, LSP code actions, language providers, AI remediation), v2.0 hardening (Position sentinel, FixStrategy union, TagSet, Finding sub-structs)

---

## D. TOTALLY FUCKED UP (Nothing Critical — But Honest Gaps)

### D1: doc.go removed a useful cross-reference
I removed the `docs/guides/flight-recorder.md` path reference from doc.go because "filesystem paths in godoc are not clickable." But the guide **does exist** and **is useful**. I replaced it with inline field docs, but a developer reading godoc no longer knows the guide exists. I should have kept a softer reference like "See the project docs for a full workflow guide" rather than removing the pointer entirely.

### D2: version-drift.sh fix is a grep band-aid, not a real parser
The `grep -v '// indirect'` filter works but is fragile. If the go.mod format changes (e.g., comments on the same line as the require), this breaks silently. A proper fix would use `go mod edit -json` to parse the module file. But for a bash CI script, the grep approach is pragmatic and matches the existing style.

### D3: No paths-ignore on the new CI jobs
The `structural-checks` and `go-work-sync` jobs run on every push/PR, including markdown-only changes. The existing ci.yml has `paths-ignore` at the top level, which applies to all jobs, so this is actually fine — but I didn't verify this explicitly during the session. The top-level `paths-ignore` covers `**/*.md`, `docs/**`, etc., so markdown changes won't trigger these jobs.

### D4: Didn't verify ci.yml YAML syntax
I tried Python/Ruby/Perl YAML parsers — none were available. I visually verified the structure matches existing patterns. The YAML is likely valid, but I didn't get programmatic confirmation. GitHub Actions will validate on next push.

### D5: Didn't push to remote
The plan's T5 step 5.8 said "Push and verify CI runs the new jobs." I didn't push. The user didn't explicitly ask me to push. All work is committed locally but unverified on GitHub Actions.

---

## E. WHAT WE SHOULD IMPROVE

### E1: TODO_LIST.md is stale
The CI script entries (lines 57-60) say DONE but their notes don't mention they're now wired into ci.yml. Should update notes to: "Script created + wired into ci.yml as `structural-checks` job."

### E2: No per-module CHANGELOG
4 modules, 0 per-module changelogs. Sub-module consumers have no visibility into module-specific changes. This is T9 in the plan.

### E3: CI has no go-arch-lint boundary enforcement
The multi-module architecture's #1 structural risk (cross-module coupling) is unenforced. Scripts catch replace directive and version drift issues, but not architectural boundary violations like pipeline importing from cmd/go-finding.

### E4: 71 gopls stdversion warnings
The project uses `encoding/json/v2` (Go 1.26 experimental). gopls reports 71 warnings about `json.Unmarshal requires go1.27 or later`. These are noise — the project intentionally uses the experimental API with `GOEXPERIMENT=jsonv2`. Not a bug, but clutters diagnostics.

### E5: The Pareto plan itself has a numbering inconsistency
Step 3 lists 42 tasks in the header but only T1-T35 are numbered (with sub-tasks like T3a/T3b, T24a/T24b, etc. inflating the count). The total of "42" is correct but the numbering scheme (T1-T35 with letter suffixes) could confuse a reader.

### E6: AGENTS.md gotcha entry for CLI config struct is critical
The CLI `flightRecorderFileConfig` (3 fields) vs pipeline `FlightRecorderFileConfig` (5 fields) split is a **split-brain risk**. If someone adds a field to one and forgets the other, consumers get confused. This should be tracked as T7 (parity decision) with higher priority.

### E7: No "config reference" doc exists
Config-file options are documented piecemeal across `flight-recorder.md`, `fix-engine.md`, code comments, and godoc. There's no single `docs/guides/configuration.md` that lists all options. This is T12 in the plan.

---

## F. Next 50 Things to Get Done (Prioritized)

### Immediate (This Session's Debt)
1. Update TODO_LIST.md notes for CI script entries to mention ci.yml wiring
2. Push to remote and verify GitHub Actions runs the new jobs
3. Soft-reference the flight-recorder guide back in doc.go (not a filesystem path)

### Critical Path Continuation (T7-T11)
4. **T7:** Align CLI flightRecorderFileConfig fields with pipeline FlightRecorderFileConfig (add MinAge + MaxBytes or document intentional omission)
5. **T10:** Research go-arch-lint, write config, run locally, fix violations, add ci.yml job
6. **T16:** Write circular symlink test for resolveSafePath
7. **T16:** Write dangling/broken symlink test for resolveSafePath
8. **T16:** Write root-is-symlink test for resolveSafePath
9. **T17:** Write disk-full error path test for FlightRecorder
10. **T17:** Write last-stage SlowStageThreshold test for FlightRecorder
11. **T17:** Write concurrent detector + flight recorder test
12. **T8:** Design + write scripts/docs-freshness.sh
13. **T8:** Add docs-freshness job to ci.yml

### Module Documentation (T9)
14. Create pipeline/CHANGELOG.md skeleton
15. Create analysis/CHANGELOG.md skeleton
16. Create cmd/go-finding/CHANGELOG.md skeleton
17. Populate each with [Unreleased] entries from git log
18. Add cross-reference from root CHANGELOG.md to sub-module changelogs
19. Update AGENTS.md Project Documentation Files table with sub-module CHANGELOGs

### Documentation (T12-T15)
20. Write docs/guides/configuration.md (central config reference)
21. Add YAML + JSON examples for each config section
22. Write docs/guides/troubleshooting.md (common pipeline errors)
23. Add FlightRecorder to README.md feature table
24. Write docs/DOMAIN_LANGUAGE.md (pipeline domain terms glossary)

### Testing Gaps (T18-T19)
25. Write sanitizeFilename property-based / fuzz test
26. Evaluate SARIF schema vendoring vs lightweight JSON validator
27. Implement SARIF schema validation test

### Release Prep
28. Tag v1.5.1 (patch: FlightRecorder config-file integration + CI hardening)
29. Tag sub-module versions: pipeline/v1.5.1, analysis/v1.5.1, cmd/go-finding/v1.5.1
30. Run scripts/version-check.sh against new tags
31. Verify `go get github.com/larsartmann/go-finding@v1.5.1` resolves

### Multi-Module Quality
32. **T11:** Create single-module variant for benchmark comparison
33. **T11:** Run benchstat comparison (multi-module vs monolith)
34. **T11:** Write results to docs/reports/
35. Add `paths-ignore` verification for new CI jobs (confirm markdown changes don't trigger)
36. Consider splitting `structural-checks` job into 3 parallel jobs for faster feedback

### FlightRecorder v2 (T24-T26)
37. Design trace file rotation config (MaxFiles, MaxTotalBytes)
38. Implement rotation logic in flight_recorder.go
39. Add gzip compression to trace output
40. Add pprof capture alongside trace snapshots
41. Design OpenTelemetry span bridge from trace data
42. Design trace diff algorithm for comparing snapshots

### Ecosystem Expansion (T27-T31)
43. Write consumer migration guide (v1.3/v1.4 API simplifications)
44. Write ToolAdapter recipes for revive, errcheck, ineffassign
45. Design LSPCodeAction wire type + ToCodeActions() method
46. Design language provider interface for non-Go languages
47. Design AIProvider interface for AI-assisted remediation

### v2.0 Hardening (T32-T35)
48. Design Option[T] position type (eliminate -1 sentinel)
49. Design FixStrategy closed union (interface-based)
50. Implement TagSet type (map[Tag]struct{})

---

## G. Questions (Cannot Resolve Without User Input)

### G1: Push now or batch?
The 3 commits from this session (plus concurrent session commits) are local-only. Should I push now so GitHub Actions validates the new ci.yml jobs? Or wait for more work to batch?

### G2: FlightRecorderFileConfig parity direction
The pipeline struct has 5 fields (Enabled, OutputDir, SlowStageThreshold, MinAge, MaxBytes). The CLI struct has 3 (Enabled, OutputDir, SlowStageThreshold). Should I:
- **(a)** Add MinAge + MaxBytes to the CLI struct (full parity, more config surface), or
- **(b)** Document the intentional omission (CLI users don't need buffer tuning)?

### G3: version-drift.sh — keep grep or upgrade to `go mod edit -json`?
The current grep approach is pragmatic but fragile. `go mod edit -json | jq` would be robust but adds a jq dependency to CI. Keep the grep fix, or upgrade?

---

_Assisted-by: Crush <crush@charm.land>_
