# Status Report: brandid-lint BD003 Strong-ID Adoption

**Date:** 2026-11-17 19:38 CET
**Session scope:** `brandid-lint -rules all -fix .` reported 2 BD003 findings; this session researched the rule, fixed both sites with existing go-finding branded types, migrated all call sites, and verified every gate.

## Scope Line — What This Report VERIFIED vs NOT

**VERIFIED (fresh command output this session):**

| Gate                             | Command                                     | Result                                                 |
| -------------------------------- | ------------------------------------------- | ------------------------------------------------------ |
| brandid-lint (BD001+BD002+BD003) | `brandid-lint -rules all .` (run twice)     | exit 0, zero findings (includes BD001 clean)           |
| Root module tests + race         | `go test -race -count=1 .`                  | pass (43 ginkgo specs + full suite)                    |
| Root module tests                | `go test -count=1 .`                        | pass                                                   |
| go vet                           | `go vet .`                                  | clean                                                  |
| golangci-lint (root)             | `golangci-lint run ./...`                   | 0 issues (pre-existing `erraudit` nolint warning only) |
| JSON determinism                 | `bash scripts/json-deterministic-check.sh`  | OK                                                     |
| Workspace build (all 5 modules)  | `go build ./...` under go.work              | clean                                                  |
| Doc drift                        | `rg IsHashID doc.go README.md docs/guides/` | no signature references                                |

**NOT VERIFIED (and why):**

- Sub-module test suites (pipeline, analysis, toolsdk, CLI): only COMPILED (`go build ./...`), not run. `IsHashID` and the sarifResult wire struct have zero non-root consumers in-repo (rg verified), so risk is low, but the suites were not exercised.
- Formatting (`nix fmt` / gofumpt): NOT run. Edits preserved line shape, but the canonical formatter never executed.
- brandid-lint on `cmd/`: the tool EXCLUDES cmd/ paths by design (linter doc.go); the CLI module was never BD003-scanned.
- External consumer compile impact: the 23 consumer repos in docs/ecosystem.md were NOT swept; `IsHashID` is now source-breaking for callers passing `string` variables (string literals still compile).
- Stress gate, benchmarks, release preflight: not run (no release in this session).
- Daemon-commit contents: changes were verified by gates AFTER the daemon absorbed them, but the committed diffs were not audited against intent.

## a) FULLY DONE

1. **Research (read, not guessed):** BD003 semantics read from the linter source at `~/projects/go-branded-id/linter` (doc.go, rules.go, suggest_id.go): BD003 is DETECTION-ONLY (`-fix` repairs BD001 exclusively); suggested type names are mechanically derived from identifier names (that is why it suggested `IsHashID` and `RuleID`); suppression syntax is `//brandid-lint:ignore(BD003) <reason>` (BD002 polices directive integrity); the scan skips cmd/, ui/, output, test-helper, tooling paths plus `_test.go`.
2. **`IsHashID(id string)` → `IsHashID(id ID)`** — id.go:176, using the existing branded `finding.ID` (the parameter IS a finding identity; no new type invented).
3. **`sarifResult.RuleID string` → `RuleID RuleName`** — sarif_types.go:88. The field already carried `Finding.Rule` both directions; `RuleName` is the domain type. Wire JSON is byte-identical (named string types marshal as plain strings).
4. **Conversion simplification** — sarif_export.go:163 `RuleID: f.Rule` (was `string(f.Rule)`); sarif_import.go:81 `Rule: r.RuleID` (was `RuleName(r.RuleID)`).
5. **Ten call/test sites migrated:** bdd_test.go:519, example_test.go:75, id_fuzz_test.go:109, id_parse_test.go:53, id_test.go:128, relativize_test.go:172, sarif_test.go:157 + 195-197 (map key) + assertRulePresent signature (sarif_test.go:23).
6. **gomega typed-literal fix** — sarif_test.go:157 `gomega.Equal(RuleName("SA1000"))`; gomega/reflect.DeepEqual is type-sensitive.
7. **AGENTS.md updated** — new bullet under Core type rules: BD003 clean without the library, mechanical-name gotcha, minimal-dependency stance (never import go-branded-id into core), detection-only + suppression syntax, gomega typed-literal trap.
8. **All gates green** — see verified table above. Deliberate design decisions: NO duplicate `RuleID` branded type (split-brain avoidance), NO go-branded-id import into core (minimal-dependency principle, AGENTS.md).

## b) PARTIALLY DONE

1. **BD003 adoption is clean but ungated.** The repo scans clean under `-rules all` (BD001 included), but nothing PREVENTS regressions: no ci.yml job, no BuildFlow provider, no script wrapper, no version pin for the binary. The cleanliness is ad-hoc, one `brandid-lint` run deep.
2. **Type-safety sweep of ID-ish surfaces.** Root core is done; adjacent surfaces spotted in session output but NOT touched: toolsdk `spec_test.go` shows `FindingRuleIDs`/`ExcludedRuleIDs []string` (test file — skipped by the scanner; the production spec fields were not researched per the no-unrelated-research instruction).
3. **Documentation of the change.** AGENTS.md carries the durable pattern; CHANGELOG/MIGRATION entries are release-time work and were not started.
4. **Daemon hygiene.** id.go showed a mod-time of 2026-11-11 (before this session) when first edited — the auto-commit daemon had touched files between my read and edit (the tool's staleness guard caught it). Edits were verified by gates afterward, but the daemon's mid-session commit contents were not diff-audited (2026-11-08 strike-marker incident class).

## c) NOT STARTED

1. **Release train** for the `IsHashID(ID)` signature change: no tag, no CHANGELOG section, no release-preflight, no CLI require bump (CLI stays pinned to published siblings per the no-replace rule).
2. **Consumer impact sweep:** unknown how many of the 23 consumer repos call `IsHashID` with a `string` variable and would break on the next require bump.
3. **Wiring brandid-lint as a gate** (CI job / BuildFlow / script) — decided nothing, wired nothing.
4. **HARVEST of section (f)** into TODO_LIST.md / ROADMAP.md (docs-health) — explicitly deferred pending your instructions.
5. **BD003 policy for future findings** (adopt branded types vs suppress wire structs) — undecided.
6. **brandid-lint version-pin policy** — binary is nix-installed 0.5.1 while the library module in the cache is v0.7.0; no tracking rule exists.

## d) TOTALLY FUCKED UP (honest self-audit — nothing reached master broken; all caught and fixed in-session)

1. **Cross-file multiedit blunder:** I bundled the sarif_types.go edit INTO the id.go multiedit call. The tool's staleness guard rejected the whole call (`file has been modified since read`). Zero damage, pure process error: multiedit is one-file-per-call.
2. **Foreseen-but-not-prevented test failure:** in analysis I explicitly predicted `gomega.Equal("SA1000")` would fail (reflect.DeepEqual type sensitivity) — and still did NOT include the fix in the edit batch. The first test run failed on exactly that line. Predicting a failure and then waiting for the test to prove me right is wasted cycle time; the fix was one line I had already written in my head.
3. **Read-before-edit discipline:** 7 parallel edits were rejected (`you must read the file first`) because I had read those files via bash sed/rg instead of the view tool. One wasted round-trip batch, then re-viewed all 7 and re-applied.
4. **Misleading verification output:** the first `go test -race` tail showed ginkgo `SUCCESS!` and package `FAIL` in the same 5 lines — I needed a second targeted run to isolate TestToSARIF. Should have grepped the FAIL pattern in the first pass.

## e) WHAT WE SHOULD IMPROVE

1. **Pre-stage views, then batch edits:** view every target file first, edit in one parallel batch. Would have saved two wasted round trips this session.
2. **Predict-then-fix:** when a failure mode is predicted during analysis, apply the fix in the SAME batch — do not let the suite "confirm" a known issue.
3. **view tool for anything I intend to edit** — bash reads do not satisfy the edit precondition.
4. **Run tests immediately after type changes,** before lint: gopls catches assignment breaks but NOT DeepEqual type-sensitivity; only the suite caught that.
5. **Close the format loop:** run `nix fmt` on edited files as part of the change loop — skipped this session and got away with it because the edits preserved shape. Got away with it is not a policy.
6. **Read the linter's doc.go BEFORE acting on linter output:** the detection-only nature and suppression syntax were discoverable upfront instead of mid-session.
7. **Verify-then-claim:** every gate claim in this report cites the command that produced it (adopted via the status-report scope-line rule).

## f) Up to 50 Things To Get Done Next

_(brainstorm per status-report skill; most are ROADMAP fuel, HARVEST must apply routing rigor — owner instruction set N=50)_

**Release & API (1-10)**

1. Run per-module test suites: pipeline, analysis, toolsdk, CLI (this session only compiled them).
2. Decide version policy for `IsHashID(ID)`: ship in next core minor (v1.16.0) as documented source-break, OR add a deprecation shim first — needs owner input (see question 1).
3. Write the CHANGELOG entry for the BD003 adoption (signature + wire-struct typing).
4. Add a MIGRATION note if the shim path is chosen.
5. Sweep the 23 consumer repos (docs/ecosystem.md table) for `IsHashID(stringVar)` call sites.
6. Run `scripts/release-preflight.sh` when the train starts.
7. Tag core, then bump CLI require (no replace directives), then tag CLI — one tag per push, per release-procedure.
8. Post-tag checks: stamp-gate, pkg.go.dev render, verify-compare-links.sh.
9. Stress gate before any tag (ginkgo -r --race --repeat=20 core+pipeline; go test -race -count=20 analysis+CLI).
10. Re-run the docs/ecosystem.md consumer sweep after the release lands.

**Lint wiring (11-18)**

11. Wire `brandid-lint -rules all` into ci.yml with an explicit OK/FAIL verdict line (dead-gate rule 1).
12. Decide BuildFlow provider integration for fleet-wide BD003.
13. Pin the brandid-lint binary version in the nix flake for reproducibility.
14. Document the cmd/ exclusion and decide whether the CLI module needs its own scan strategy (the tool never scans cmd/ by design).
15. Add suppression-directive guidance (`//brandid-lint:ignore(BD003) reason`) to docs/guides/.
16. Decide a BD002-style staleness audit for ignore directives (remove-one-then-red-then-restore, like nolint-audit).
17. Add brandid-lint to the files-to-gates map in docs/release-procedure.md.
18. Define a tracking policy for binary-vs-module version skew (brandid-lint 0.5.1 vs library module v0.7.0).

**Tests & correctness (19-24)**

19. Pinned assertion that `sarifResult.RuleID` marshals as a plain JSON string (round-trip covers it; a one-line explicit test is cheap insurance).
20. Explicit table entry: `IsHashID(ID(""))` returns false (fuzz seed covers; make it visible).
21. Guard example_test.go golden output (already pinned by its `// Output:` block — keep it that way).
22. Optional manual audit for BD003-evading patterns (params NOT named `*id*` that are still string identities) in pipeline/analysis/toolsdk.
23. Domain decision: `RuleName` has no `IsValid()` while `GroupID` does — deliberate or gap?
24. Standard pre-release bench run (`nix run .#bench`) unaffected by this change, confirm once on the next train.

**Docs & memory (25-29)**

25. HARVEST this report's section (f) into TODO_LIST.md / ROADMAP.md via docs-health.
26. DOMAIN_LANGUAGE.md check concluded: no new term needed (RuleName exists; sarifResult is internal wire) — recorded, nothing to do.
27. Update docs/ecosystem.md notes as consumers migrate past the breaking change.
28. Cross-project lesson candidate for crush-config references/lessons.md: "a linter's suggestedTypeName is mechanical; fix with the existing domain type, not the suggestion" — only if it generalizes.
29. Annotate THIS report (docs-health ANNOTATE mode) when the API change ships.

**Hygiene (30-35)**

30. Run `nix fmt` (gofumpt/goimports) over the session's edited files — the one skipped loop step.
31. Audit this session's daemon commits: diff the auto-committed BD003 change against intent.
32. Investigate the pre-session mod-time drift on id.go (2026-11-11 stamp) — confirm no daemon fmt race damage (2026-11-08 strike-marker incident class).
33. Re-confirm gopls project diagnostics are 0 after all edits (last edit batch was clean; one explicit re-check pending).
34. Check `git status` to confirm report + code are fully absorbed by the daemon.
35. Confirm the golangci-lint `erraudit` unknown-nolint warning stays pre-existing-only (it is documented; verify no NEW unknown directives appeared).

**Fleet-wide BD003 readiness (36-42)**

36. BD003 candidates in toolsdk: `FindingRuleIDs` / `ExcludedRuleIDs []string` — decide branded rule-ID type vs suppression.
37. CLI module strategy given the cmd/ exclusion — manual audit or accept.
38. Prepare an adoption guide for other LarsArtmann repos before `-rules all` goes fleet-wide.
39. Ship the shared pattern doc: "prefer the existing domain branded type over the linter's mechanical suggestion."
40. Review the id-parse family boundary conversions (string parts in/out of ParseID) for branded-type friction.
41. Check go-output/cmdguard ecosystem interplay IF the fleet-wide gate is adopted.
42. BD001 parity: `-rules all` clean implies BD001 clean here — keep the default rule set green too (it is; verified).

**Process (43-45)**

43. Record the one-file-per-multiedit rule and the view-before-parallel-edits rule in project AGENTS.md gotchas.
44. Habit: re-run `brandid-lint -rules all .` after any branded-type-adjacent change (it is a seconds-fast syntactic scan).
45. Cite the linter source path (`~/projects/go-branded-id/linter/doc.go`) in AGENTS.md so future sessions research rule semantics in one hop.

## g) Questions I Cannot Figure Out Myself

1. **Release policy for the breaking signature:** `IsHashID` now takes `finding.ID`; callers passing `string` variables stop compiling on the next require bump (literals unaffected). Ship as-is in the next core minor with a CHANGELOG note, or add a deprecation shim (`IsHashIDString(string) bool`) to soften the 23-consumer upgrade? This sets the breaking-change tolerance for the whole fleet.
2. **Gate integration:** should `brandid-lint -rules all` become a CI gate in go-finding (and a BuildFlow provider fleet-wide), or stay opt-in/manual? If CI: acceptable runner budget for the job (it is a syntactic scan, seconds-fast)?
3. **BD003 adoption stance for future findings** (e.g., toolsdk's `FindingRuleIDs []string`): adopt branded types everywhere, or suppress-and-stay-primitive for wire/spec structs? Where exactly is the line between domain ID and wire format?

---

_WAITING FOR INSTRUCTIONS. Per the status-report skill, section (f) is the primary HARVEST input for TODO_LIST.md / ROADMAP.md — say the word and I run docs-health HARVEST. Per the Crush harness contract I did not commit this report; the auto-commit daemon picks it up._
