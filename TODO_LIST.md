# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).
> Owner questions (release backfill, brew/nix distribution, stale clones) live in
> [ROADMAP.md](ROADMAP.md) "Open questions" — not here.

> **Harvested 2026-09-10** from `docs/status/2026-09-09_03-24_late-night-session-self-review.md`
> §f (50 items), each verified against code before inclusion: CI is fully green
> (21/21 jobs incl. benchmark), Dependabot PRs #23/#24/#25 merged and #29 closed,
> v1.9.0–v1.9.2 released, v1.10.0 + `toolsdk/v1.10.0` released with 34-asset
> GitHub Releases, pkg.go.dev renders (core + toolsdk verified), the stale
> `/tmp` worktrees are removed, and the tag-batching/dispatch rules are now in
> `docs/release-procedure.md`. The former 🔴 HIGH "Release & CI integrity"
> section (concurrency group, unpushed-commits gate, version-stamp guard,
> post-tag self-test, worktree-hygiene gate, `nix flake check` follow-up) is
> DONE — see CHANGELOG.md [Unreleased]. Everything below is what actually remains.

---

## 🟠 HIGH-MED Priority — Consumer ecosystem

| Task                                                             | Impact | Effort | Notes                                                                                                                                                                                                                       |
| ---------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| File BuildFlow issue (`TestNoLintPathExclusions` policy failure) | Med    | Low    | Evidence in the `docs/ecosystem.md` sweep table (report items 23). Diagnose-and-file, per the Q2 decision.                                                                                                                  |
| File branching-flow issue (pkg/errors + pkg/fs build failures)   | Med    | Low    | Pre-existing at v1.4.1; capture fresh output, file in that repo (report item 24).                                                                                                                                           |
| File go-business-rules issue (pre-existing test failures)        | Med    | Low    | Capture fresh output first (report item 25).                                                                                                                                                                                |
| art-dupl: file the GroupID integration intent in art-dupl        | Med    | Low    | GAP-2 consumer; feature request with the GAP-2 context — no issue carries it yet (report item 27; feedback doc `docs/feedback/2026-06-05_art-dupl-integration-evaluation.md`).                                              |
| Opportunistic consumer bumps to v1.10.0                          | Low    | Med    | Additive since v1.8.0; no consumer is past v1.8.0 yet. Leads: go-linter-sdk, golangci-lint-auto-configure (report items 26/29). gomend + licenseforge stay blocked on their broken BuildFlow replace paths (issues #1/#46). |
| Re-check library-policy after hook issue #74 resolves            | Low    | Low    | Verify the devShell hook healing stuck (report item 30).                                                                                                                                                                    |
| Consumer compatibility matrix as CI job                          | Med    | Med    | Unblocked (repo public); validate the public promise mechanically (report item 42).                                                                                                                                         |
| Tag missing v1.10.0 sub-module releases                          | High   | Low    | **Local tags created 2026-09-11** (`pipeline/v1.10.0`, `analysis/v1.10.0`, `cmd/go-finding/v1.10.0` at master, annotated). Remaining: push `master`, then the 3 tags in ONE push (≤3-per-push rule), then confirm Release workflow + proxy resolution. Local `cmd/go-finding` require already bumped to `pipeline v1.10.0` (version-drift gate green). |

## 🟡 MEDIUM Priority — Testing & code quality

| Task                                                         | Impact | Effort | Notes                                                                                                                                              |
| ------------------------------------------------------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Same-second modtime-tie case in the concurrent-rotation test | Low    | Low    | `writeMu` fix papers over modtime-tie nondeterminism; assert count, not identity (report item 32).                                                 |
| Real fuzz campaign for `FuzzPruneSnapshotsHostileDir`        | Low    | Low    | Seeds green only; run `-fuzz` 30s+ (report item 33).                                                                                               |
| Formatter partial-write matrix: `FormatTextRich` failAt=1    | Low    | Low    | Verify the main-write failure case is covered; complete the table (report item 34).                                                                |
| IntervalIndex/correlate uncovered blocks (24 stmts at 98.3%) | Low    | Med    | Push toward 99% or consciously accept (report item 35).                                                                                            |
| Dependabot watch: confirm gomega bumps stay merged           | Low    | Low    | #25 merged; watch the next cycle for re-opening (report item 19). go.mod sweep verified: ginkgo v2.32.1 consistent (core+pipeline; others stdlib). |
| `actions/*` SHA-drift monitoring policy                      | Low    | Low    | sbom bump was manual review; decide pin-vs-dependabot for actions (report item 21).                                                                |
| json/v2 stabilization watch (Go 1.27)                        | Low    | —      | Ongoing; tracked in [ROADMAP.md](ROADMAP.md) "json/v2 stabilization watch".                                                                        |

## 🟢 LOW Priority — Launch & product

| Task                                                    | Impact | Effort | Notes                                                                                                                         |
| ------------------------------------------------------- | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------- |
| Publish announcement + submit to Awesome Go             | Med    | Low    | Drafts exist (`docs/brainstorming/launch-announcement-draft.md`); CI green + pkg.go.dev live unblocked them (report item 43). |
| CONTRIBUTING / release-runbook page for public audience | Low    | Med    | Release procedure is internal-facing; consider a public runbook page (report item 44).                                        |
| v2.0 design spike session                               | Low    | High   | Parked in [ROADMAP.md](ROADMAP.md) "Hardening": Position sentinel, FixStrategy union, TagSet, sub-structs (report item 39).   |
| Sampling NO-GO review date                              | Low    | Low    | ROADMAP has a revisit trigger; consider a calendar date (e.g. 2027-01) instead (report item 41).                              |

---

_FlightRecorder future ideas (pprof capture, OTel bridge, trace diff) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions" (scored table; pprof/OTel/trace-diff rejected with reasons)._

_The deferred breaking changes with concrete designs are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

_Assisted-by: Crush <crush@charm.land>_
