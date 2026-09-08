# Status Report: Feedback-Doc Gaps + Issues #27/#28 Implementation

**Date:** 2026-09-07 23:55 CEST
**Session scope:** Review of all `docs/feedback/` files and all GitHub issues; implementation of everything found open.
**Repo state at report time:** working tree clean (auto-commit daemon), branch `master`, all 4 modules green.

---

## 0. Executive Summary

The session reviewed 1 feedback document (art-dupl evaluation, 9 gaps) and 2 GitHub issues (#27, #28). Research showed **6 of 9 feedback gaps were already implemented** in prior releases. The remaining open work — GAP-2 (GroupID), a GAP-4 incompleteness (ToLSP tag emission), issue #27 (per-finding fix outcomes), and issue #28 (rollback scope) — was implemented end-to-end with tests, lint, docs, and changelogs. Nothing is broken; the notable debts are: no benchmark verification of the new allocations on the legacy engine path, a messy auto-commit history, unintended indirect-dependency bumps that rode along, and no release/tag yet for a change that alters documented default behavior.

---

## a) FULLY DONE

| #  | Item                                                                                                                                                                                                                                                                                                                                                | Evidence                                                                                               |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| 1  | Research pass: all 9 feedback gaps + 2 issues verified against current code (several doc claims were stale — 6 gaps already done)                                                                                                                                                                                                                   | Session research; findings recorded in feedback doc status table                                       |
| 2  | GAP-2: `Finding.GroupID` (branded type), `WithGroupID` builder, `Equal()` coverage, JSON `groupId` (omitempty)                                                                                                                                                                                                                                      | `branded_types.go`, `finding.go`, `finding_equal.go`, `finding_builder.go`                             |
| 3  | GroupID interchange: SARIF property `go-finding/groupId` (export+import), LSP `Data.GroupID` round-trip                                                                                                                                                                                                                                             | `sarif_types.go`, `sarif_export.go`, `sarif_import.go`, `lsp.go`                                       |
| 4  | `Report.GroupFindings()` (active findings only, nil when ungrouped)                                                                                                                                                                                                                                                                                 | `report_query.go`                                                                                      |
| 5  | GAP-4 completion: `ToLSP()` re-emits diagnostic tags from `Metadata[LSPDiagnosticTagsKey]` (malformed entries skipped)                                                                                                                                                                                                                              | `lsp.go`                                                                                               |
| 6  | Issue #27: `FixEngine.ApplyWithOutcomes` → `FixApplyResult` with one `FixOutcome` per finding (`applied`/`no-change`/`refused`/`conflict`/`invalid`/`failed`), `OutcomeFor`/`OutcomeCounts`/`HasErrors`; legacy `Apply`/`ApplyWithConflicts` delegate unchanged                                                                                     | `pipeline/fix_outcome.go`, `pipeline/fix_engine.go`                                                    |
| 7  | Issue #28: `RollbackPolicy` (`RollbackPolicyFailingFile` default, `RollbackPolicyAllFiles` legacy), `SetRollbackPolicy`, `Config.FixRollbackAllFiles`, config-file `fixRollbackAllFiles`, CLI config parity                                                                                                                                         | `pipeline/fix_applier.go`, `pipeline/config.go`, `pipeline/config_file.go`, `cmd/go-finding/config.go` |
| 8  | `FixApplier.ApplyWithReport` → `ApplyReport` (applied, outcomes, shift maps, `RolledBack`, `FailedOutcomes()`); soft failures no longer abort runs                                                                                                                                                                                                  | `pipeline/fix_applier.go`                                                                              |
| 9  | Tests: mixed-status outcomes, empty input, legacy parity, invalid-edit reporting, soft-error-keeps-files (#28 scenario), refused reporting, GroupID (JSON/Equal/builder/SARIF/LSP/report), LSP tag round-trip, config-file mapping, rewritten legacy rollback tests (default + AllFiles variants), saboteur test re-targeted to the hard-error path | `pipeline/fix_outcome_test.go`, `pipeline/fix_applier_test.go`, core test files                        |
| 10 | Lint clean (core, pipeline, CLI), `go vet` clean, `go-arch-lint` OK                                                                                                                                                                                                                                                                                 | session runs                                                                                           |
| 11 | Full suite green: 10 packages, `-race -count=1`, all 4 modules                                                                                                                                                                                                                                                                                      | session run                                                                                            |
| 12 | CI scripts: test-naming, json-deterministic, go-work-sync, replace-audit, version-drift, docs-freshness (0 stale)                                                                                                                                                                                                                                   | session runs                                                                                           |
| 13 | Docs: both CHANGELOGs (Unreleased sections), AGENTS.md (key files + 4 new gotchas), FEATURES.md (3 sections), fix-engine guide (outcomes + rollback policy + report), configuration guide (`fixRollbackAllFiles` row), feedback doc annotated with a 9-gap implementation-status table                                                              | files listed                                                                                           |

## b) PARTIALLY DONE

| # | Item                          | What exists                                                        | What is missing                                                                                                                                           |
| - | ----------------------------- | ------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Rollback policy test coverage | Hard-error path (write failure), AllFiles variant, soft-error path | No explicit test that earlier files keep fixes on **context cancel** or **backup failure** under the default policy (behavior implemented, untested)      |
| 2 | GAP-4 (diagnostic tags)       | Types, metadata preservation, ToLSP emission, round-trip test      | No golden JSON wire-format test protecting the `go-finding/lsp-diagnostic-tags` metadata spelling                                                         |
| 3 | GAP-2 (GroupID)               | Full core + interchange                                            | No golden JSON test protecting the `groupId` key spelling; no group-aware dedup/correlation integration; no usage-guide section or example in `examples/` |
| 4 | Issue follow-through          | Fixes implemented locally                                          | Issues #27/#28 not commented on or linked (no push/close without permission); fixes unreleased                                                            |
| 5 | Config-file parity            | Unit test for field mapping                                        | No E2E test proving `fixRollbackAllFiles` reaches the applier via a real config-file run (flight recorder got that treatment in v1.6.0)                   |
| 6 | New API documentation         | Guide + godoc                                                      | `doc.go` package overview not extended; `DOMAIN_LANGUAGE.md` lacks outcome/refused/rollback-policy/GroupID terms                                          |

## c) NOT STARTED

| #  | Item                                                                                                       | Note                                                                                                                     |
| -- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| 1  | Benchmark regression check (`scripts/bench-check.sh` / `nix run .#bench`) for the engine rework            | **Biggest gap.** `ApplyWithConflicts` now allocates outcome slices + reconciliation maps on the legacy path — unmeasured |
| 2  | Benchmark for `ApplyWithOutcomes` itself                                                                   | none exists                                                                                                              |
| 3  | Release: version bump, version.go, tag set (incl. `pipeline/v*`), release-date stamping in CHANGELOGs      | Blocking consumer uptake of both fixes                                                                                   |
| 4  | Semver/ADR decision for the rollback default change (behavior change of documented API default in a minor) | Needs an explicit ADR in `docs/architecture-decisions.md`                                                                |
| 5  | Fuzz coverage for `ApplyWithOutcomes`                                                                      | `fix_engine_fuzz_test.go` untouched                                                                                      |
| 6  | CLI: surface refused/failed counts in fix output; optional `-fix-rollback-all` flag                        | Config-file only today                                                                                                   |
| 7  | Metrics: record refused/failed/conflict counts (`RecordFixes` counts applied only)                         |                                                                                                                          |
| 8  | `OnFix` callback carrying outcome status                                                                   | Signature change — needs decision                                                                                        |
| 9  | JSON marshaling for `FixOutcome`/`FixApplyResult` (deterministic)                                          | Consumers will want to report outcomes                                                                                   |
| 10 | GAP-3 (per-relationship `RelatedRef` metadata)                                                             | Deferred per feedback doc — unchanged                                                                                    |
| 11 | GAP-7 final disposition (registered-categories registry vs. `IsStandard`/`IsValid` split)                  | Documented as "resolved differently", not decided with maintainer                                                        |
| 12 | HARVEST of this report's next-steps into `TODO_LIST.md`/`ROADMAP.md`                                       | Awaiting instructions                                                                                                    |
| 13 | errors.Is chain assertions for outcome errors (`ErrPositionUnresolvable` reachability)                     |                                                                                                                          |
| 14 | Structured `FindingError` for outcome failures (go-error-family integration)                               |                                                                                                                          |

## d) TOTALLY FUCKED UP

Nothing shipped broken — final state is green everywhere. Honest failures during/around the session:

1. **Git history quality: ~9 `chore: auto-commit N file(s) (heuristic)` commits** slice the feature work into arbitrary chunks (daemon + my no-commit rule collided). Before push this should ideally be squashed into coherent commits with real messages — needs force-with-lease approval or acceptance as-is.
2. **Unrequested dependency changes rode along in those auto-commits**: indirect bumps `x/mod 0.38→0.40`, `x/net 0.57→0.58`, `x/text 0.40→0.41`, `x/tools 0.48→0.49` (across root, pipeline, CLI go.mod/go.sum). Untested-as-a-change (tests pass), unreviewed, not what I intended to ship.
3. **`funlen` violation I introduced in `FromLSP`** mid-session (caught by lint, fixed by extracting `preserveLSPFidelity`). Process worked, but I should have kept the addition smaller.
4. **Two red test runs during development** (parallel-call panic in a new test; provisional `Applied` status bug in `ApplyWithReport` early returns). Both fixed; normal iteration, listed for honesty.

## e) WHAT WE SHOULD IMPROVE

1. **Measure, don't assume**: the engine refactor changed the allocation profile of a hot path and I shipped without a benchmark comparison. Benchmarks before/after should be part of the definition of done for engine changes.
2. **Behavior changes to documented defaults deserve an ADR + maintainer sign-off before implementation**, not just changelog lines afterwards.
3. **Commit intentionally**: either commit feature work properly in-session (with permission) or stage changes so the daemon's heuristic slicing at least lands per-topic.
4. **Guard interchange keys with golden tests**: property names (`go-finding/groupId`, `lsp-diagnostic-tags`) and JSON tags are wire contracts; they deserve explicit golden tests, not just round-trip tests.
5. **E2E for config plumbing**: every new config field should get the flight-recorder treatment (unit + E2E proving it reaches runtime), not just a mapping unit test.
6. **Close the loop on feedback docs proactively**: the art-dupl doc was 3 months stale (6 gaps done, undocumented). The status table added today fixes it; make "annotate feedback docs with implementation status" part of implementing any feedback item.
7. **Cancel/backup-failure paths were re-semantically changed without dedicated tests** — any policy branch should get its own test even when existing suites stay green.
8. **Consumer communication**: 22 known consumers; a default-behavior change needs a release-note blast (go-cqrs-lint first).

## f) 50 things to get done next (brainstorm, sorted roughly by impact)

1. Run `scripts/bench-check.sh` (baseline vs current) for FixEngine/Applier; investigate the `ApplyWithConflicts` allocation delta from outcome bookkeeping
2. Add `BenchmarkApplyWithOutcomes` (and grouped BenchmarkApply variants)
3. Cut the release: bump version, stamp CHANGELOGs, tag core + `pipeline/v*`, verify proxy/pkg.go.dev
4. Write ADR: rollback policy default change (semver rationale, migration note)
5. Comment on issues #27/#28 linking the fix; close after tag
6. HARVEST this list into `TODO_LIST.md`/`ROADMAP.md` (docs-health)
7. E2E CLI test: `fixRollbackAllFiles: true` config reaches applier (flight-recorder style)
8. Add `-fix-rollback-all` CLI flag for config parity
9. Tests: earlier files keep fixes on **context cancel** (default policy)
10. Tests: earlier files keep fixes on **backup failure** (default policy)
11. CLI fix summary: print refused/failed/conflict counts after fix stage
12. Golden JSON tests: `groupId` field spelling + `go-finding/groupId` + `go-finding/lsp-diagnostic-tags` keys
13. Fuzz `ApplyWithOutcomes` (extend `fix_engine_fuzz_test.go`)
14. Metrics: `RecordOutcome(status)`; surface in pipeline summary
15. `errors.Is` assertions: outcome.Err → `ErrPositionUnresolvable` chain
16. Confirm/tidy the unintended indirect dep bumps (or revert them deliberately)
17. Clean up git history (squash heuristic commits) before push — needs force-with-lease approval
18. `OnFix(finding, applied)` → carry outcome status (decide breaking-vs-new callback)
19. Deterministic JSON for `FixOutcome`/`FixApplyResult` (consumer reporting)
20. `FindingError`-typed outcome failures (position + go-error-family classification)
21. `DOMAIN_LANGUAGE.md`: outcome, refused, rollback policy, group, clone group
22. `doc.go`: mention GroupID + outcome API in package overview
23. `Template`-level `WithGroupID` (stamp group per template instance for clone detectors)
24. Usage guide: grouping + outcomes sections in `docs/USAGE_GUIDE.md`
25. `examples/`: an outcomes + rollback-policy example program
26. art-dupl integration spike: implement `cloneGroupToFindings` adapter from the feedback doc (unblocked by GAP-1/2/4)
27. go-cqrs-lint `--fix`: adopt `ApplyWithReport`/`FailedOutcomes` downstream; verify UX
28. Notify consumers (22 per audit) about the rollback default change in release notes
29. Decide: `ApplyWithReport` continue-on-hard-file-error option (currently stops)
30. `ApplyReport.RolledBack` file list also embedded in the error message text
31. Shift-map semantics test: files with mixed applied/refused findings
32. GroupID validation decision: free-form vs format; document either way
33. `Report.GroupFindings` deterministic ordering option (sorted keys/findings)
34. Group-aware dedup/correlation: `Correlate`/`MergeIter` respect GroupID?
35. SARIF research: represent clone groups via codeFlow/threadFlow or run-level metadata
36. LSP recipe doc: how editors group diagnostics by GroupID
37. Wire-format stability test for `LSPDiagnosticData` (new `groupId` tag)
38. Provider-precedence test: first provider refuses, second applies (document fall-through)
39. `handleFileError` bookkeeping: partial RollbackAll failure under AllFiles not fully listed in `RolledBack`
40. Expose engine-level `Conflicts` detail in `ApplyReport` (currently only outcome status)
41. GAP-7: decide registered-categories registry vs `IsStandard`/`IsValid` split; annotate feedback doc with the decision
42. GAP-3: define the trigger condition to revisit per-relationship metadata
43. Fuzz `parseLSPDiagnosticTags` (malformed metadata values)
44. `nix run .#test` + `nix run .#lint` full-parity run (session used go/golangci-lint directly)
45. dprint format check for the markdown files touched this session
46. README: mention GroupID + fix outcomes in feature list
47. `Report` summary: `ByGroup` counts in `Summary`? (decide)
48. Applier: optional `ApplyDryRun` (resolve outcomes without writing) for plan/apply UX
49. Engineering hygiene: deduplicate provider stubs across `fix_outcome_test.go`/`fix_applier_test.go` into `testutil_test.go`
50. Post-release: verify pkg.go.dev renders new APIs; close feedback-doc items fully (GAP-2/4 marked shipped, not just implemented)

## g) Questions I cannot figure out myself

1. **Release & semver intent for the rollback default change**: ship per-file rollback as the new default in v1.7.0 (my implementation), or do you want legacy all-or-nothing to stay the default with per-file as opt-in? This decides the ADR, the version number, and consumer comms.
2. **GitHub issue handling**: may I comment on #27/#28 now (linking the local fix, "pending release"), or should I wait until you push/tag? And do you want them closed by the release commit or manually?
3. **GAP-7 final answer**: is the `IsStandard()`/`IsValid()` split your permanent resolution for "Category.IsValid() accepts typos", or do you still want the registered-categories registry from the feedback doc?

---

_Point-in-time snapshot. Section (f) is brainstorm input for docs-health HARVEST, not a commitment list._
