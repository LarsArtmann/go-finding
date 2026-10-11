# Status Report — Phase D Execution, FEATURES Walk & Full Verification

> **📦 RESOLUTION STATUS (updated 2026-09-08 18:55)**
>
> SUPERSEDED twice: by the SUPERB plan execution status
> (`2026-09-08_18-30_superb-plan-execution-status.md`) and the brutal
> self-review (`2026-09-08_18-48_superb-execution-self-review.md`). The v1.7.0
> release this report prepared for shipped (tags pushed); its open tail
> (L1-49 D5/D8 feedback annotations = DOCS15) was executed 2026-09-08 evening
> session. Remaining open items live in the 18-48 report's §f list.

**Date:** 2026-09-08 15:42 CEST
**Session scope:** Continuation of the Pareto master plan
(`docs/planning/2026-09-08_04-12_pareto-master-plan-ci-billing-releases-v1.7.0.md`).
User dropped the billing gate ("switching accounts soon"), then directed
full execution of everything unblocked. Predecessor report:
`docs/status/2026-09-08_07-38_pareto-execution-ci-release-repair-phase-b-hardening.md`.

**Branch state at report time:** `master`, working tree **clean**, all changes
committed by the auto-commit daemon. **6 commits ahead of origin/master**
(this session only). Note: the previous session's ~16 unpushed commits are now
on origin — they were pushed between sessions **by someone else, not by me**;
I have not pushed anything.

---

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                                                                                                                                                                                                    | Evidence                                                                                                                                                                                   |
| - | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **L1-33 test-hardening bundle** — provider-precedence tests (refuse→fall-through, error→fall-through), shift-map mixed-outcome test, `RolledBack` paths embedded in error text at all 3 rollback failure sites (cancel / backup failure / hard file failure) with pinning tests, `LSPDiagnosticData` byte-exact wire golden (minimal + fully populated, determinism + round-trip asserts), broken `RollbackPolicy` godoc sentence fixed | `pipeline/fix_engine_test.go`, `pipeline/fix_applier_test.go`, `pipeline/fix_applier.go` (`rolledBackNote`), `lsp_test.go` (`TestLSPDiagnosticData_GoldenWire`)                            |
| 2 | **L1-30 outcome metrics + CLI summary** — `Metrics.RecordOutcome(status)`, `OutcomeCounts()`, `MetricsSnapshot.OutcomeCounts` (mutex-safe, snapshot is a copy), fix stage records one outcome per finding, CLI prints `Fix outcomes: applied=N, refused=N, ...` in canonical order                                                                                                                                                      | `pipeline/metrics.go`, `pipeline/pipeline_detect.go` (`applyDirectFixes` now uses `ApplyWithReport`), `cmd/go-finding/main.go`, tests incl. concurrent (race) + Run-level + CLI table test |
| 3 | **L1-31 deterministic outcome JSON + typed errors** — custom `MarshalJSON`/`UnmarshalJSON` for `FixOutcome` + `FixApplyResult` with `json.Deterministic(true)` (repo rule), errors as message strings; failed outcomes now carry `*finding.FindingError` (parse category) with the finding's position; `errors.Is/As` chains to the provider cause preserved and test-pinned                                                            | `pipeline/fix_outcome.go`, `pipeline/fix_engine.go`, `pipeline/fix_outcome_test.go` (round-trip + determinism), `fix_engine_test.go` (typed-error assertions)                              |
| 4 | **L1-41 examples + stub dedupe** — runnable `pipeline/examples/outcomes` demo (executed: applied/refused/failed all demonstrated correctly, including the substring-fallback conflict subtlety), compile check `pipeline/examples/example_compile_test.go`, 5 duplicated provider test stubs moved into `pipeline/testutil_test.go`                                                                                                     | `pipeline/examples/outcomes/main.go`, `pipeline/testutil_test.go`, three test files slimmed                                                                                                |
| 5 | **L1-37 SARIF/LSP grouping research** — `docs/guides/finding-groups.md`: verified against the SARIF 2.1.0 spec (correlationGuid, relatedLocations, codeFlows/threadFlows, partialFingerprints), recommendation = keep property bag canonical, relatedLocations decoration is viewer-only and intentionally NOT implemented; LSP client-side grouping recipe with code                                                                   | `docs/guides/finding-groups.md`, linked from `docs/USAGE_GUIDE.md`                                                                                                                         |
| 6 | **L1-13 stress-gate decision** — step 4 of the release procedure declared a **mandatory** gate with rationale and revisit condition                                                                                                                                                                                                                                                                                                     | `docs/release-procedure.md`                                                                                                                                                                |
| 7 | **L1-25 FEATURES.md full walk** — 4 parallel verification agents covered all 22 sections + Summary Matrix; **~26 stale claims found and fixed** (details in section d/e context below)                                                                                                                                                                                                                                                  | `FEATURES.md` (many edits), `correlate.go` (complexity comments corrected)                                                                                                                 |
| 8 | **Full verification battery** — race tests ×4 modules (11 packages, exit-code-verified), golangci-lint ×4 modules = 0 issues, all 6 structural scripts, go-arch-lint (after glob fix), dprint, `nix fmt`, `nix flake check` exit 0                                                                                                                                                                                                      | command outputs this session                                                                                                                                                               |
| 9 | **Docs housekeeping** — root/pipeline/CLI CHANGELOG Unreleased entries for all new work; TODO_LIST harvested (billing decision recorded, 3 MEDIUM rows closed, DONE table extended); AGENTS.md new gotchas (sequential `go test` on shared GOCACHE, `RolledBack`=backed-up semantics, treefmt not on devShell PATH, `pipeline/examples/` pattern, billing decision, typed outcome failures)                                             | `CHANGELOG.md`, `pipeline/CHANGELOG.md`, `cmd/go-finding/CHANGELOG.md`, `TODO_LIST.md`, `AGENTS.md`                                                                                        |

Fixes applied during the FEATURES walk (representative): `Stable` field →
`Stable()` method; `WriteSARIF(w)` → `WriteSARIF(ctx, w)` (and siblings);
provider `CanHandle` conditions corrected (all require `HasCodeChange()`);
`Apply` 3-value return documented; `VerifyResult.Modified`/`Resolved` added;
`Config.FixRollbackAllFiles` row added; `Snapshot(ctx, reason)`;
staticcheck `SA*`→correctness precedence; CLI flag table +`-fix-rollback-all`
+`-include-suppressed`; CLI formats corrected (6 formats, `table` is
library-only, no CSV/TSV footer); `StageHookFunc` missing `ctx` param fixed;
`NewGeneratedFileFilter` signature fixed; `DurationMs` → `FilesScanned`;
`ComputeSummary` thread-safe = yes; `FindByID` shallow-copy note;
`WithGroupID` added to builder list; hash-based file-level IDs documented;
`IntervalIndex` O(log n + k) → O(n + k) (doc + `correlate.go` comments);
`Correlation.Score` field name; test-naming claim reworded; examples table +
outcomes; Summary Matrix 7→6 formats.

## b) PARTIALLY DONE

| Item                                           | Done                                                                                                                                       | Remaining blocker                                                                           |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------- |
| v1.7.0 release train                           | All pre-release work: tests, benches (incl. pipeline in baseline), docs, migration guide, outcomes/metrics/rollback features in Unreleased | D1 (rollback default sign-off → ADR-016), D6 (release strategy), CI/account switch          |
| CI verification                                | Workflows repaired (dispatch triggers, cosign v3 bundle mode, 4-module lint matrix, extended benchmark job); local gates all green         | GitHub Actions billing — user switches accounts; then `gh workflow run ci.yml --ref master` |
| Issues #27/#28                                 | Fixes implemented, fuzz- and golden-tested, changelogged                                                                                   | D2: close now vs at release (comment + close)                                               |
| Phase D bundle (L1-30..41)                     | L1-30/31/33/37/41 shipped this session                                                                                                     | L1-32 (D7 GroupID validation), L1-34 (D3 OnFix), L1-39 (D4 ApplyDryRun) stay decision-gated |
| L1-30 subtask "docs snapshot in metrics godoc" | Tests + wiring done                                                                                                                        | Godoc example/snapshot text not yet written                                                 |
| Consumer migration to v1.7.0                   | Guide at `docs/guides/consumer-migration-v1.7.md`, README linked                                                                           | Actual consumer bumps (go-humanize-linter, go-linter-sdk) — needs release first             |
| TODO_LIST harvest                              | Unblocked rows closed with evidence                                                                                                        | D-gated rows intentionally remain open                                                      |

## c) NOT STARTED

- L1-35 art-dupl integration spike (`cloneGroupToFindings`) — depends on v1.7.0 release (L1-21).
- L1-36 go-cqrs-lint adoption (external repo, bump to v1.7.0).
- L1-42..46 ROADMAP design spikes (Position sentinel, FixStrategy closed union, pointer-as-state, Tags→TagSet, Finding sub-struct) — parked in ROADMAP by design.
- L1-47 FlightRecorder future triage (rotation, gzip, pprof, OTel, diff, AI).
- L1-49 D5/D8 answers (registry vs IsStandard split, GAP-3 revisit trigger).
- Make-public launch items: announcement post, Awesome Go, pkg.go.dev render check, Homebrew tap verification.
- Post-switch CI re-verification (`gh workflow run ci.yml --ref master`, all jobs incl. stress + govulncheck).
- Closing/commenting issues #27/#28 on GitHub (D2).

## d) TOTALLY FUCKED UP (honest failures this session)

1. **Piped exit-code masking — repeat offense.** First `nix flake check | tail`
   run printed a treefmt diff but my `echo "flake-exit=$?"` captured the
   _pipeline's_ exit code, reporting success while the check had FAILED.
   The follow-up run with output redirect + explicit `$?` caught it. This
   exact trap ("set -o pipefail / verify raw exit codes") is already in
   AGENTS.md from last session. Caught, but it should not have happened once,
   let alone twice.
2. **Stray `pipeline/examples/main.go`.** A duplicate main.go appeared at the
   examples root (content differing from `outcomes/main.go`). Origin unclear
   (write-tool path handling is the prime suspect). Removed via trash; the
   tree is clean now — but an unexplained file appearing mid-session deserves
   suspicion, not silence.
3. **Batch-then-build anti-pattern.** After moving 5 provider stubs into
   `testutil_test.go` I left an unused `"bytes"` import in
   `fix_engine_test.go`, producing a build failure discovered only at the
   next full test run. AGENTS.md's own rule ("build immediately after
   deleting/moving, before editing dependents") was violated.
4. **Two wrong first-draft test expectations** (caught by tests, fixed, but
   avoidable):
   - `RolledBack` test assumed only _written_ files are listed — reality:
     the `modified` list tracks _backed-up_ files, so soft-failed files
     appear too (restores are content no-ops).
   - Pipeline metrics Run test assumed refused findings reach the applier —
     reality: `FilterConflictingFixes` drops unresolvable candidates before
     `applyDirectFixes`, so only applier-level outcomes are counted.
     A closer read of `groupFindingsBySafePath`/`FilterConflictingFixes`
     beforehand would have saved two round trips.
5. **Concurrent go test runs on the shared GOCACHE** produced transient
   `[build failed]`/`[setup failed]` noise. I knew the risk and ran two race
   suites in parallel anyway; then briefly misread the failure as real.
   Re-ran sequentially — green. Now documented in AGENTS.md.
6. **Three sequential golines fix rounds** on `lsp_test.go` (split args,
   split composite literal, split second call) instead of running the
   formatter (`nix fmt`) up front. Wasted cycles; also discovered
   `treefmt` is not invokable via `nix develop -c treefmt` (not on PATH).
7. **Arch-lint glob gap.** Adding `pipeline/examples/outcomes/main.go` broke
   `go-arch-lint` because `pipeline-examples` didn't glob `**`. Caught in
   final verification (good), but I knew I was adding files under a
   component-globbed directory and could have checked the glob first.
8. **Systemic doc drift exposed, not caused, by this session:** the FEATURES
   walk found ~26 stale claims accumulated since v1.3.0 despite prior
   "spot-verified" passes. Spot checks are not verification. The
   docs-freshness CI script checks mtimes, not signature truth — nothing in
   CI would have caught any of these.

## e) WHAT WE SHOULD IMPROVE

1. **Never trust piped output for pass/fail** — always capture the raw exit
   code of the underlying command (redirect to file, `echo $?`). This is now
   written down; next step is _obeying_ it reflexively.
2. **Read the full call path before writing test expectations** — provider
   selection, pre-filters, and bookkeeping semantics (`modified` vs written)
   determine observable behavior.
3. **Run `nix fmt` before linting**, not after hand-fixing golines findings.
4. **Build after every structural move**, not after the batch.
5. **Make the FEATURES walk a recurring release gate** (docs-health VERIFY
   per release cycle). The 26-fix harvest proves per-signature verification
   must be periodic, not one-off.
6. **Consider a CI checker for doc-claimed API truth** — e.g. extract
   backtick identifiers from FEATURES.md and assert each exists via
   `go doc`/LSP. Would have caught most of the drift mechanically.
7. **Two golangci-lint versions in play** — local 2.13.1 vs CI-pinned
   v2.10.1 (plus deprecation warning: `exhaustruct` → `exhaustruct_v5`).
   Align them to avoid "green locally, different findings in CI".
8. **Reduce stale-modtime edit conflicts** with the auto-commit daemon —
   commit manually before long editing phases so the daemon doesn't touch
   files mid-edit (hit twice this session).
9. **Test-style consistency** — root module mixes plain `testing` goldens and
   gomega; pick one convention per file type to lower review noise.
10. **Stray-file vigilance** — if an unexplained file appears, diff it against
    its sibling and record the incident (done here) rather than silently
    deleting.

## f) NEXT 50 (prioritized, gated items marked)

**Release-critical (needs your decisions)**

1. D1 decision: keep per-file rollback default → ADR-016 + stamp v1.7.0 (gated).
2. D6 decision: backfill v1.6.0 releases (`gh workflow run release.yml --ref v1.6.0`) vs forward-only (gated).
3. Push the 6 unpushed local commits (on your word).
4. After account switch: `gh workflow run ci.yml --ref master`, verify ALL jobs (test, lint matrix, coverage, benchmark, govulncheck, stress, module-isolation, dupl, arch-check, docs-freshness).
5. Verify `gh secret list` has `HOMEBREW_TAP_GITHUB_TOKEN` before first successful release run.
6. Ship v1.7.0: version.go bump, 4-module tags, watch release run, verify assets.
7. Comment + close #27/#28 (D2 timing).
8. Bump consumers (go-humanize-linter, go-linter-sdk) to v1.7.0 using the migration guide; run their suites.
9. `bash scripts/version-check.sh` + `bash scripts/bench-check.sh benchmarks/baseline.txt current.txt 25` as pre-tag steps.

**Phase D remainder (decision-gated)**
10. D7: GroupID validation (free-form vs `^[a-z0-9-]+$`) → implement in validator.
11. D7 (ungated part worth pulling): deterministic `GroupFindings` option + `Template.WithGroupID` if we decouple them from the validation decision.
12. D3: `OnFix` outcome status (breaking callback vs new `OnFixOutcome`).
13. D4: `ApplyDryRun` (resolve outcomes without writing) + continue-on-error shape.
14. L1-33 leftover from plan wording: provider-stub dedupe follow-up — also fold `cancelingProvider` + `upperProvider` into `testutil_test.go`.
15. Metrics godoc snapshot/example text (finish L1-30 subtask 4).
16. CLI e2e test asserting the `Fix outcomes:` stderr line appears in a real run.

**Docs & quality (unblocked)**
17. Add outcomes/rollback section pointer to `docs/DOMAIN_LANGUAGE.md` Fix Application table (cross-check `Metrics` term coverage).
18. Extend `docs-freshness.sh` or add `scripts/docs-api-check.sh`: verify backtick identifiers in FEATURES.md exist in code (mechanical drift guard).
19. Add `finding-groups.md` link from README docs table.
20. Write `docs/guides/outcomes.md` consumer guide (ApplyWithOutcomes patterns, OutcomeFor/Counts/HasErrors, rolled-back semantics incl. backed-up-files nuance).
21. Golden wire test for `FixApplyResult`/`FixOutcome` JSON in the docs example (already have round-trip; pin bytes in doc snippet).
22. Align golangci-lint version local↔CI; migrate `exhaustruct` → `exhaustruct_v5` config before the linter removal lands.
23. Add `nix fmt` as a pre-commit/CI step signal so treefmt drift never reaches flake check (the dprint hook covers md only).
24. Consider `.golangci.yml` `gofumpt`/`golines` autofix pass config to end hand-fixing loops.
25. `IntervalIndex` follow-up research: quantify whether a real interval tree (O(log n + k)) is worth it at consumer scale; record go/no-go in ROADMAP.
26. `Report.GroupFindings()` determinism note/test (map iteration order) — document or add `GroupFindingsSorted` when D7 lands.
27. Fuzz `FixOutcome`/`FixApplyResult` UnmarshalJSON (malformed wire data).
28. Benchmark `ApplyWithOutcomes` with mixed outcomes (applied+failed+conflict) at n=1000 to confirm the legacy allocation parity survived the outcome-typing change.
29. Property test: metrics `OutcomeCounts` sum == number of fixable findings processed.
30. Sweep root-module tests for `t.Parallel()`-via-helper consistency after the golden tests.

**Launch / public presence (post account switch)**
31. pkg.go.dev render check after first public tag.
32. Announcement draft (blog/r/golang/Slack/X) — content mostly derivable from FEATURES.md now that it is verified.
33. Awesome Go submission.
34. GoReleaser + Homebrew tap verification on a public tag.
35. Consumer compatibility matrix (22 consumers, 14 with Go code) — still gated on GOPRIVATE/repo visibility.

** ROADMAP spikes (design docs, no code)**
36. L1-42 Position sentinel redesign go/no-go.
37. L1-43 FixStrategy closed union.
38. L1-44 pointer-as-state cleanup.
39. L1-45 Tags → TagSet.
40. L1-46 Finding sub-struct composition.

**FlightRecorder & bigger bets**
41. L1-47 FlightRecorder idea triage (rotation, gzip, pprof, OTel bridge, trace diff, AI analysis) → scored table into ROADMAP.
42. L1-49 D5/D8: GAP-7 final answer (registry vs IsStandard) + GAP-3 revisit trigger, annotate feedback doc.
43. art-dupl spike L1-35 (after v1.7.0).
44. go-cqrs-lint adoption L1-36 (after v1.7.0).
45. json/v2 stabilization tracking (drop GOEXPERIMENT when Go stdlib ships it) — subscribe/re-check on each toolchain bump.
46. CI: add `workflow_dispatch` inputs for module-scoped runs (faster iteration once CI is alive).
47. CI: benchmark job should fail on regression via `bench-check.sh` gate (currently informational?) — verify and gate.
48. Evaluate `ginkgo --repeat=N` cost vs a sampled stress matrix in CI (stress is local-mandatory now; CI duplication may be wasteful).
49. Write the missing ADRs: outcome metrics design (L1-30), typed outcome errors (L1-31) — small ADR-017/018 entries keep decision history complete.
50. Post-release retro: harvest this session's two false-expectation lessons into `docs/DOMAIN_LANGUAGE.md` (backed-up vs written files; pre-filter vs applier outcomes) so consumers don't rediscover them.

## g) THREE QUESTIONS I CANNOT ANSWER MYSELF

1. **D1 — rollback default:** Do you confirm shipping v1.7.0 with the per-file
   rollback default (`RollbackPolicyFailingFile`) as-is? A "yes" unlocks
   ADR-016 + the full v1.7.0 release train (version.go, 4 tags, release run).
2. **D6 — release strategy:** Backfill the dead v1.5.0/v1.6.0 releases via
   `gh workflow run release.yml --ref v1.6.0` (proves the cosign fix on an
   existing tag), or go forward-only from v1.7.0 and leave v1.5/v1.6
   un-released forever?
3. **Push policy while billing is broken:** The earlier 16 commits reached
   origin (pushed between sessions, not by me). Current local state is 6
   daemon commits ahead on a green tree. Should I push each verified batch
   as it completes, or keep everything local until you explicitly say push?

---

**Verification snapshot (this session, exit-code-verified):** race ×4 modules
(11 packages) ok · golangci-lint ×4 = 0 issues · test-naming / json-
deterministic-check / go-work-sync / replace-audit / version-drift /
docs-freshness all OK · go-arch-lint OK · dprint OK · `nix fmt` applied ·
`nix flake check` exit 0 · examples/outcomes executed with correct output.

**Then wait for instructions.**
