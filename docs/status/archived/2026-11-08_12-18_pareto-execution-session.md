# ARCHIVED 2026-11-08 — superseded same-day by the 12:48 self-critique

> **Live remainder:** the open tail is re-listed (with honest dispositions) in
> docs/status/2026-11-08_12-48_pareto-execution-self-critique.md §f and routed
> into TODO_LIST.md. This report is kept for the done/partial-done evidence
> tables and the session gotchas.

# Status — Pareto Plan Execution Session (2026-11-08 12:18)

**Session scope:** execute `docs/planning/2026-11-08_08-50_SUPERB-pareto-plan-stamp-truth-and-the-docs-long-tail.md` — the owner directive was the WHOLE list. 39 of 42 tasks executed to verified completion (T38 corpus drain recorded NOT-STARTED with reason; T25 partially train-gated by nature; T39 prep recorded, external comms stay owner-gated). 3 cross-repo changes shipped (erraudit bump; clean-wizard + md-go-validator bumps). Two real upstream-class defects found and fixed in the process.

## a) Done — release truth hardening (the 1% tier)

| Task                          | Verdict                                                                       | Evidence                                                                                                                                                                                         |
| ----------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| T1 pre-tag stamp gate         | done — `scripts/stamp-gate.sh`, `--tag` preflight wiring, 6-scenario selftest | c753540; FAIL paths proven live (v1.15.2 mismatch + already-exists + usage)                                                                                                                      |
| T2 cosign "failure" diagnosis | done — ROOT CAUSE was NOT cosign                                              | goreleaser nixpkgs pipe 404 on `LarsArtmann/nur-packages` AFTER publishing (run 41863470480 log); installer + all 17 signatures succeeded. Fixed: `nix.skip_upload: true`; recorded in CHANGELOG |
| T24 goreleaser deprecations   | done — `goreleaser check` clean                                               | `brews` → `homebrew_casks` (2.19 schema, `binaries` list), `format_overrides.formats`; snapshot render proof with throwaway local tag                                                            |

## b) Done — consumer-facing truth (4% tier)

| Task                       | Verdict                                                                                                                        |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| T3 CI bench numbers        | PASS verdict recorded in benchmarks/README (run 42456449428: no regression beyond time +250% / alloc +10%)                     |
| T4 pkg.go.dev              | all 5 pages render (cmd v1.15.1 crawled after re-trigger; was 404)                                                             |
| T5 CI release gate         | `release-gate` job in release.yml; PROVEN to bite: `go install cmd@v1.15.1` prints 1.15.0 ≠ 1.15.1                             |
| T6 detector-coverage guide | `docs/guides/detector-coverage.md`, claims verified against code (fixed ComputeSummary signature + JSON names)                 |
| T7 README sell + ADR #20   | Why-bullet + DetectorV2 snippet + ADR #20 written; FEATURES/README linked                                                      |
| T8 DetectorV2 examples     | root + pipeline examples, both wired into compile/run tests, lint-clean both modules; pipeline DefaultConfig gotcha documented |
| T16 api-stability sweep    | DetectorV2/DetectResult/RunDetect/AddDetectResult + Result.FilesScanned + FR config identifiers added                          |

## c) Done — release-integrity automation (20% tier)

| Task                      | Verdict                                                                                                                                                                                          |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| T9 compare-links verifier | 76/76 pairs resolve via authenticated `gh api`; bogus pair FAILs; preflight `--links` flag                                                                                                       |
| T10 lychee token          | GITHUB_TOKEN env added to the step (first live run still 502/504-flaked — rerun path exercised, see §g)                                                                                          |
| T11 release-train.sh      | 8 stages; dry prints plan; postflight against v1.15.1 FAILS on the shipped stamp defect (rc=1) — the tooling detects the defect end-to-end                                                       |
| T12 erraudit migration    | analysis/pipeline pins → v1.15.1 (commit c78b8db5 on erraudit's cmdguard-v5, pushed; 350 BDD specs green)                                                                                        |
| T13 upstream issue        | NOT FILED — the 2026-09-23 false-stale premise does NOT reproduce (27 needed/0 stale in all 3 modes, erraudit 5c841cf); AGENTS re-verified instead. Verify-before-filing gate worked as designed |
| T14 CLI FR delegation     | resolve() delegates to pipeline.ResolveFlightRecorderConfig; wording pinned by new test; orphaned nolint removed                                                                                 |
| T17 watchdog              | ci-watchdog.sh (3 paths proven live) + every-30-min watchdog.yml workflow                                                                                                                        |
| T18 procedure truth pass  | phased-CI-red table, d/1 stress rule, files-to-gates map, GoReleaser recovery recipe, interrupted-train rule                                                                                     |
| T15 compare-ref sets      | analysis + cmd CHANGELOGs: [Unreleased] + full ref sets (append-only-safe: all pairs reference existing tags)                                                                                    |

## d) Done — decisions recorded (owner-gated tier)

All six 🔒 items DECIDED 2026-11-08 and implemented (TODO_LIST decisions ledger): tag-gating = documented exception (no gate weakened); toolsdk = independent (+ preflight loops now include toolsdk, closing the documented blind spot); dependabot collapsed to `/` (failed run 36348692826 was GitHub updater infra, green since Nov 1); branch-protection contexts fixed via test/lint wrapper jobs (dependabot PRs no longer structurally unmergeable); codecov branch DELETED (fix on master, PR #2 superseded); launch-comms = quiet-maintenance until v1.16.0.

## e) Done — tail tier

T27 CLI `Coverage:` stderr line (e2e pins the v1-silence contract); T29 SARIF round-trip quick property (1000 iters, seed 42; surfaced + pinned the import edit-synthesis); T30 shellcheck clean (19 findings fixed) + gate wired into preflight; T31 boundary READMEs; T32 seed convention + measured cost (~20 ms for 192 files); T33 go-worksum-check.sh (3 legs, FAIL path proven); T34 consumer-compat `--head` (3 real script bugs died: subshell-abort dead-gate, double-slash replace, tidy-before-probe); T35 version-lattice.sh (selftest via git plumbing; toolsdk lag correctly WARNs); T36 lint-matrix job summary; T40 CONTRIBUTING truth pass + release runbook; T41 v2.0 spike agenda; T42 hygiene batch (buildcache no-cron, SA1012 accepted after testing the alternative, provenance PARKED, standing-watches table); T26 ecosystem sweep (23 consumers, all on core v1.15.1 after same-day bumps; licenseforge unblocked); T28 bumps (clean-wizard d2efd2c, md-go-validator c7fc05a, both suite-green and pushed).

## f) Up to 50 things we should get done next

| # | Item                                                                                                                                    | Size |
| - | --------------------------------------------------------------------------------------------------------------------------------------- | ---- |
~~| 1 | Corpus drain batch 1: 10 April-era archived reports through the emit-keys flow; lower the audit_max ratchet (327 items; 35 April files) | XL   |~~ routed — 2026-11-08 12:48 self-critique §f 1 (unchanged, XL)
~~| 2 | LAUNCH4 finish: verify `HOMEBREW_TAP_GITHUB_TOKEN` secret exists, flip `homebrew_casks.skip_upload` off at the v1.16.0 train            | S    |~~ routed — 2026-11-08 12:48 self-critique §f 2
~~| 3 | LAUNCH3 Awesome Go PR: owner go required (external comms)                                                                               | S    |~~ routed — 2026-11-08 12:48 self-critique §f 3
~~| 4 | T25: LIVE runner render of the nix pipe (snapshot proof exists; live is train-gated)                                                    | S    |~~ routed — 2026-11-08 12:48 self-critique §f 4
~~| 5 | v2.0 spike SESSION (agenda written, decisions CHANGE/KEEP/REJECT per item)                                                              | 60m  |~~ routed — 2026-11-08 12:48 self-critique §f 5
~~| 6 | Verify the rerun of run 42473455406 goes green (lychee 502/504 flake class; API proof 76/76 already in hand)                            | S    |~~ verified — rerun re-flaked on different links; superseded by the lychee-exclusion fix 0b722ee, CI green run 42475326942
~~| 7 | Detached long-gate runner (only remaining hygiene idea)                                                                                 | S    |~~ routed — 2026-11-08 12:48 self-critique §f 7

## g) Gotchas discovered this session

1. **`| tail` eats exit codes** — `cmd | tail; echo $?` reports tail's rc. Bit three times; now a named dead-gate rule.
2. **Local git config forces annotated tags** — `git tag <name> <sha>` needs an editor; always `git tag -a -m` in scripts (bit release-train + lattice selftest).
3. **`go mod tidy` drops unresolvable local-replace requires** — the consumer-compat wrapper-module approach died; the copy-and-retarget approach survived.
4. **The daemon absorbs edited files within seconds** — commit units immediately; `git status` before every authored message (d/2 rule, re-learned ~8 times).
5. **`GOWORK=off` CLI builds resolve PUBLISHED core** — the at-HEAD CLI stamp probe needs an appended replace in the copy, or it faithfully reproduces the published stamp defect (which is exactly how the probe proved itself).

## h) Verification state

docs-gates battery 5/5 PASS; docs-api-check, docs-freshness, link-check, changelog-drift green; preflight selftest 6/6; actionlint clean on all three workflows; shellcheck 0 findings across scripts/ + hooks/; goreleaser check clean; flake check passed (devShell change); erraudit full suite green in erraudit; clean-wizard + md-go-validator suites green. Master CI: last code-carrying run had markdown-link-check red on 49 anonymous 502/504 gateway errors against API-valid links (76/76 authenticated proof); rerun in flight per the documented verdict path — all other jobs green including the new wrapper jobs, at-HEAD consumer-compat, and watchdog.
