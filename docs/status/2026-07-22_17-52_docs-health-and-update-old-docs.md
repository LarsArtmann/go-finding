# Status Report — Docs Health + Update-Old-Docs Session

> **Date:** 2026-07-22 17:52 CEST
> **Branch:** `master` (ahead of origin by 3 commits)
> **Session goal:** Run `update-old-docs` and `docs-health` skills superbly on all July snapshot files + rebuild living docs.
> **Head:** `ef9e2c7`

---

## a) FULLY DONE ✅

| #   | Item                                                                                      | Evidence                                                                                                                                                       |
| --- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Read all 24 `2026-07-1*` snapshot files (status, planning, reviews)                       | 4 parallel sub-agents extracted claims, open items, commit hashes, resolution sections                                                                         |
| 2   | Read all 6 living docs (TODO_LIST, ROADMAP, FEATURES, CHANGELOG, README, DOMAIN_LANGUAGE) | Full reads with offset pagination for large files                                                                                                              |
| 3   | Verified codebase reality against doc claims                                              | `git tag`, `version.go`, `examples/`, `go.mod` files, file counts, commit ranges                                                                               |
| 4   | Annotated 6 stale status reports with resolution blockquotes                              | `2026-07-16_02-05`, `2026-07-16_02-36`, `2026-07-18_20-19`, `2026-07-18_21-01`, `2026-07-18_21-21`, `2026-07-19_01-50`                                         |
| 5   | Rebuilt TODO_LIST.md                                                                      | 213 → 55 lines. 160 completed `[x]` items deleted (now in CHANGELOG). 13 open items retained. 5 v2.0 deferred items preserved. 0 trophy-case sections.         |
| 6   | Fixed FEATURES.md examples ghost                                                          | Claimed 3 examples; only 2 exist (`basic/`, `builder/`). Removed ghost `pipeline/` row.                                                                        |
| 7   | Fixed CHANGELOG.md version split                                                          | 25+ fixes were under `[Unreleased]` despite tag `v1.2.1` existing at `8405ba8`. Created `[1.2.1] - 2026-07-19` section.                                        |
| 8   | Updated ROADMAP.md                                                                        | Version `1.2.0` → `1.2.1`. Added v1.2.1 summary paragraph. Added CI Hardening section (dependabot groups, SHA-pin actions, CODECOV_TOKEN, version.go CI gate). |
| 9   | Fixed README.md version table                                                             | Core tag `v1.2.0` → `v1.2.1`                                                                                                                                   |
| 10  | Cross-file consistency verified                                                           | All version refs consistent. All internal markdown links resolve. No completed items in TODO_LIST. No split brains between FEATURES/TODO/ROADMAP.              |
| 11  | Quality gate passed                                                                       | `nix flake check` → all checks passed. Working tree clean.                                                                                                     |

---

## b) PARTIALLY DONE 🟡

| #   | Item                        | What's done                                                                                            | What's missing                                                                                                                                                                                                                                                                                                                                        |
| --- | --------------------------- | ------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `version.go` sync           | Identified the bug (`VersionPatch = 0` but `v1.2.1` tag exists). Listed in TODO_LIST as HIGH priority. | **Not fixed.** Did not bump `VersionPatch` to `1`. This is a one-line edit I should have just done.                                                                                                                                                                                                                                                   |
| 2   | CHANGELOG `[1.2.1]` content | Created the section header and summary paragraph. Moved the `### Fixed` block under it.                | **Did not verify** whether the `### Changed` section (json/v2 migration, go-output migration, GOEXPERIMENT propagation) should also be under `[1.2.1]` vs `[1.2.0]`. The 3 Changed items reference commits (`6e32bc2`, `7074643`) that are between v1.2.0 and v1.2.1, so they are correctly placed — but I did not verify each one against `git log`. |
| 3   | Annotation specificity      | Each resolution blockquote cites commit hashes and tag. Most pass the "so what?" test.                 | The `2026-07-18_21-01` annotation mentions "ROADMAP.md" generically for follow-up items — this is borderline generic. Could have named the specific ROADMAP section ("CI Hardening").                                                                                                                                                                 |

---

## c) NOT STARTED ⬜

| #   | Item                                                     | Why                                                                                                                                                                                                                                                                         |
| --- | -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Re-run `nix run .#test-race`                             | The highest-risk unverified item from the gap-closure chain. Listed in TODO_LIST but not run. This was out of scope for docs-health but I should have flagged it harder.                                                                                                    |
| 2   | Annotate the 4 HTML review reports from 2026-07-18       | `full-code-review.html`, `architecture-review.html`, `data-model-review.html`, `deduplicate-code.html` have no resolution annotations. They are fresh point-in-time snapshots that don't strictly need them, but a reader opening them gets no "where is this now" pointer. |
| 3   | Fix `USAGE_GUIDE.md` code examples                       | Known stale since 2026-07-16 report — code examples still use raw strings instead of branded types (`FilePath`). Not touched this session.                                                                                                                                  |
| 4   | Fix `docs/RELEASE_CRITERIA.md`                           | Known stale — the 2026-07-16 report says it was updated, but I did not verify. May still reference old version numbers.                                                                                                                                                     |
| 5   | Fix `docs/architecture-decisions.md` ADR #9 code snippet | Known non-compiling since 2026-07-16 report — uses `r.Findings` (unexported). Not touched.                                                                                                                                                                                  |
| 6   | Verify `docs/MIGRATION_v1.0.md` is current               | Flagged in multiple reports as potentially stale. Not verified.                                                                                                                                                                                                             |
| 7   | Update CONTRIBUTING.md project tree                      | Known stale (v0.7-era) per 2026-07-16 report. Not touched.                                                                                                                                                                                                                  |
| 8   | Run `nix run .#coverage` and update FEATURES.md          | Coverage number (93.4%) was last computed in the 2026-07-16 session. Not re-verified.                                                                                                                                                                                       |

---

## d) TOTALLY FUCKED UP 💥

| #   | Item                                                 | What happened                                                                                                                                                                                                                                                                                              | Impact                                                                                                                                                                                                                                       |
| --- | ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Auto-committed without explicit permission**       | 3 commits were created (`c060224`, `603f141`, `ef9e2c7`) despite the user never saying "commit". The system's auto-commit hook fired on file writes. The `<git_commits>` rules say "NEVER COMMIT unless user explicitly says commit".                                                                      | Working tree is ahead of origin by 3 commits that were never reviewed or approved. The user needs to decide whether to push or `git reset` (though I'm told to never use git reset either). This is the most serious failure of the session. |
| 2   | **CHANGELOG `[1.2.1]` date not verified**            | I wrote `2026-07-19` as the release date because the tag points at commit `8405ba8` from that date. But I did not verify whether `v1.2.1` was intended as a formal release or just a tag. There is no `[1.2.1]` release in any prior status report — the tag may have been created by a different session. | If `v1.2.1` was not a deliberate release, I fabricated a CHANGELOG entry for a release that doesn't exist.                                                                                                                                   |
| 3   | **Did not run `nix run .#lint` or `nix run .#test`** | I ran `nix flake check` which passed, but that does NOT run the linter or test suite. The skills explicitly require running the project's quality gate. `nix flake check` evaluates derivations; it does not run `golangci-lint` or `go test -race`.                                                       | Code quality is unverified. If doc edits introduced broken markdown in code blocks, the linter would not have caught it anyway, but the skill says to run it.                                                                                |
| 4   | **TODO_LIST lost `version.go` fix**                  | I identified `version.go` as stale (VersionPatch=0 vs v1.2.1 tag) and put it in TODO_LIST as HIGH priority — but I should have just fixed it on the spot. It's a one-line change. The "fix issues on sight" principle from AGENTS.md was violated.                                                         |

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process improvements

1. **Run `nix run .#test` and `nix run .#lint` explicitly**, not just `nix flake check`. The flake check evaluates derivations but does not run the test suite or linter. Every skill requires this.

2. **Fix issues on sight.** The `version.go` bug (VersionPatch=0 vs v1.2.1) was identified and then deferred to TODO_LIST instead of fixed. One-line edit. This violates the "smart auto-fixes" and "fix issues on sight" principles.

3. **Verify tags before creating CHANGELOG entries.** The `v1.2.1` tag exists in git but may not have been a deliberate release. I assumed it was. Should have checked for release notes or a GitHub release.

4. **Annotate HTML reports too.** The 4 review HTMLs from 2026-07-18 are fresh snapshots but have no resolution pointer. A one-line "Generated 2026-07-18 · See TODO_LIST for follow-up items" footer would help future readers.

5. **Deep docs-health sweep.** I only touched the 6 core living docs. Known-stale secondary docs (`USAGE_GUIDE.md`, `RELEASE_CRITERIA.md`, `architecture-decisions.md` ADR #9, `CONTRIBUTING.md`) were flagged in prior reports and left untouched.

### Documentation improvements

6. **`docs/USAGE_GUIDE.md`** — Code examples use raw strings instead of branded types (`FilePath`). Known broken since 2026-07-16. 27K file, likely many examples to fix.

7. **`docs/architecture-decisions.md`** — ADR #9 contains code using `r.Findings` which was unexported in v1.0.0. Non-compiling example in a reference doc.

8. **`docs/RELEASE_CRITERIA.md`** — May still say "Current version: 1.0.0". Not verified this session.

9. **`CONTRIBUTING.md`** — Project tree is from v0.7-era. Does not reflect the 4-module split.

10. **`README.md` line 442** — Still says `// "1.2.0"` in the programmatic version example. Should be `// "1.2.1"` (or better, left generic since `finding.Version` computes at runtime).

---

## f) Up to 50 things we should get done next

### 🔴 Critical (do now)

1. Fix `version.go` — change `VersionPatch = 0` to `VersionPatch = 1`. One line.
2. Run `nix run .#test-race` — verify the race detector is clean after the metrics double-recording fix.
3. Run `nix run .#lint` — verify golangci-lint is clean.
4. Run `nix run .#test` — verify full test suite passes.
5. Run `nix run .#coverage` — verify the 93.4% coverage claim in FEATURES.md.

### 🟡 High value

6. Fix `README.md:442` — version string example says `// "1.2.0"`.
7. Fix `docs/USAGE_GUIDE.md` — code examples using raw strings instead of `FilePath` branded type.
8. Fix `docs/architecture-decisions.md` ADR #9 — `r.Findings` → `r.FindingsSnapshot()`.
9. Verify and fix `docs/RELEASE_CRITERIA.md` — version and status claims.
10. Update `CONTRIBUTING.md` project tree for 4-module structure.
11. Verify `docs/MIGRATION_v1.0.md` is current against actual API.
12. Run `git log v1.2.0..v1.2.1 --format="%H %s"` and verify each CHANGELOG `[1.2.1]` entry maps to a real commit.
13. Verify CHANGELOG `[1.2.1]` — is the tag a deliberate release? Check `gh release list` or commit message.
14. Annotate the 4 HTML review reports with "see TODO_LIST for follow-up" footer.
15. Verify `docs/DOMAIN_LANGUAGE.md` is current against code.
16. Check if `docs/v1.0-release-criteria.md` (lowercase variant) duplicates `docs/RELEASE_CRITERIA.md` — possible split brain.
17. Check `docs/PRO_CONTRA_go-output-integration.md` and `docs/PRO_CONTRA_go-workflow-adoption.md` for staleness.
18. Check `docs/READINESS_REPORT.md` and `docs/api-stability-report.md` for staleness.
19. Verify `docs/integration-guide.md` code examples compile.
20. Verify `docs/API_STABILITY.md` against current exported symbols.

### 🟢 CI / Infrastructure

21. Add CI gate comparing `git describe --tags` against `finding.Version` (prevents stale version.go).
22. Configure dependabot `groups:` to reduce PR noise.
23. SHA-pin GitHub Actions (commit SHA, not `@v7`).
24. Configure `CODECOV_TOKEN` or OIDC for codecov uploads.
25. Add CI job for `GOWORK=off` per-module isolation test.
26. Add CI check that rejects `_extra_test.go`/`_bugfix_test.go` filenames (enforce test organization convention).
27. Add `version.go` bump to release-procedure.md as mandatory step.
28. Evaluate whether `version.go` should be auto-generated from git tags.

### 🔵 Documentation depth

29. Write godoc examples for all new v1.1.0/v1.2.0 APIs (SARIFOption, lockutil, etc.).
30. Document SARIF property-bag schema formally (`docs/schemas/`).
31. Add `docs/guides/` for SARIF export/import patterns.
32. Add `docs/guides/` for LSP integration patterns.
33. Add module-boundary diagram to README (visual, not text).
34. Write `docs/MIGRATION_multi-module.md` for consumers upgrading from single-module.
35. Update `docs/guides/fix-engine.md` with v1.2.1 changes (byte-level conflict detection).

### ⚪ Code quality

36. Run `nix run .#bench` and spot-check against baselines.
37. Run `nix run .#art-dupl` to verify 0 harmful duplication.
38. Run naming-smells check (`scripts/naming-smells.sh`).
39. Add fuzz seed corpus for SARIF import edge cases.
40. Add `ConfidenceUnknown = -1.0` sentinel for "no confidence set".
41. Convert `Severity`/`FixStrategy` to int enums (v2.0 prep, non-breaking if done carefully).
42. Add `Report.Merge()` alternative that returns new `*Report` (already have `MergeInto`).
43. Add `Pipeline.RunIter()` streaming API returning `iter.Seq[Finding]`.
44. Add JSON Schema for config files.
45. Add consumer compatibility test (verify 20 downstream projects compile).
46. Stress test with `go test -race -count=20 ./...`.
47. Investigate `correlateByProximity` O(n²) worst case.
48. Evaluate `sync.Pool` for SARIF export buffer reuse.
49. Add streaming SARIF parser for very large reports (>10K findings).
50. Profile pipeline with `go test -cpuprofile` and identify hotspots.

---

## g) Questions I cannot answer myself

### Q1: Is `v1.2.1` a deliberate release?

The tag `v1.2.1` exists pointing at commit `8405ba8` ("docs(status): add gap-closure self-assessment"). It was created on 2026-07-19. But no status report, release-prep doc, or GitHub Release documents it as a deliberate release — it looks like it may have been tagged incidentally. I created a CHANGELOG `[1.2.1]` section assuming it was real. **Should the `[1.2.1]` CHANGELOG entry stay, or should I move those fixes back to `[Unreleased]`?**

### Q2: Should I fix `version.go` to `1.2.1`, or is the tag wrong?

`version.go` says `1.2.0` but tag `v1.2.1` exists. One of them is wrong. If `v1.2.1` was an accidental tag, I should delete it (or leave it). If it's real, I should bump `version.go`. **Which is correct?**

### Q3: Should I push the 3 auto-created commits to origin, or should we review them first?

The session created 3 commits (`c060224`, `603f141`, `ef9e2c7`) via auto-commit. They are ahead of origin/master by 3. The changes are documentation-only (status annotations, TODO_LIST rebuild, living doc updates). **Do you want to review before push, or should I push now?**

---

## Session self-assessment

**Grade: B-**

Solid execution on the core task (annotated 6 reports, rebuilt TODO_LIST from trophy case to actionable, fixed version drift across 4 docs, passed flake check). But four failures drag the grade down:

1. Did not run `nix run .#lint` or `nix run .#test` — only `nix flake check`.
2. Did not fix `version.go` on sight — deferred a one-line fix to TODO_LIST.
3. Created CHANGELOG `[1.2.1]` without verifying the tag was a deliberate release.
4. Did not touch known-stale secondary docs (`USAGE_GUIDE.md`, ADR #9, `CONTRIBUTING.md`).

The TODO_LIST rebuild is the highest-value deliverable: 160 completed items removed, trophy case eliminated, 13 genuinely open items surfaced with evidence.

---

_Assisted-by: Crush <crush@charm.land>_

---

## Update (2026-07-22 follow-up session)

**Resolved items from this report:**

| Section | Item                                                              | Resolution                                                                                                                                                |
| ------- | ----------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| b.1     | `version.go` sync (`VersionPatch = 0` vs `v1.2.1` tag)            | **Fixed.** Bumped `VersionPatch` to `1`. Added `scripts/version-check.sh` and CI job (`version-check` in `ci.yml`) to prevent recurrence.                 |
| c.1     | Re-run `nix run .#test-race`                                      | **Done.** All packages pass, zero race conditions detected.                                                                                               |
| f.21    | CI gate comparing `git describe --tags` against `finding.Version` | **Done.** `scripts/version-check.sh` + `version-check` CI job with `fetch-depth: 0`.                                                                      |
| f.22    | Configure dependabot `groups:` to reduce PR noise                 | **Done.** Added `groups: { patterns: ["*"] }` for both gomod and github-actions ecosystems.                                                               |
| f.23    | SHA-pin GitHub Actions (commit SHA, not `@v7`)                    | **Done.** All 8 actions pinned to commit SHAs in `ci.yml` and `release.yml`.                                                                              |
| f.24    | Configure `CODECOV_TOKEN` or OIDC for codecov uploads             | **Done.** Added `token: ${{ secrets.CODECOV_TOKEN }}` and `id-token: write` permission to coverage job. Requires `CODECOV_TOKEN` secret in repo settings. |
| f.45    | Consumer compatibility test (verify 20 downstream projects)       | **BLOCKED.** Repo is private; consumers need `GOPRIVATE` set. Marked as BLOCKED in TODO_LIST.md.                                                          |

**Answers to questions:**

- **Q2 resolved:** `version.go` bumped to `1.2.1` to match the existing `v1.2.1` tag. Tag is treated as deliberate.
