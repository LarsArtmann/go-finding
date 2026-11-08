# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).
> Owner questions (ApplySimpleFixes philosophy, brew/nix, report policy,
> tag-gating policy, launch comms mode) live in [ROADMAP.md](ROADMAP.md)
> "Open questions" — not here.

> **Rebuilt 2026-11-08** (docs-health AUDIT: all non-archived 2026-0* reports
> annotated + archived; harvested from the 2026-09-23 reports, the SUPERB
> pareto plan log, and the 2026-10-12 v1.15.0 train report §f). The v1.15.1
> train was COMPLETED same-day (version.go stamp fixed post-tag, sibling
> tags pipeline/analysis/cmd at v1.15.1 with green Release runs, cmd
> resynced at 720eb6d); sigstore signing is RESTORED (17 `.sigstore.json`
> bundles on v1.15.0 and v1.15.1 — the 09-27 row closed); toolsdk/v1.16.0
> shipped complete with its GitHub Release (its TODO section closed);
> master CI is green again after 4 days red (incomplete go.sums + stamp).

---

## 🔴 HIGH Priority — release integrity & next-train blockers

| Task                                                                                                                                              | Impact | Effort | Notes                                                                                                                                                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Corpus drain batch 1: sample 10 April-era archived reports; annotate-resolve via the docs-health emit-keys flow, then lower the audit_max ratchet | Medium | XL     | 35 April files in docs/status/archived/; 327 pre-canon items total (audit_max ratchet 340 only lets the count fall). NOT STARTED 2026-11-08 (session budget; a half-annotated file breaks the annotate gate, so partial work is worse than none). |
| Make erraudit CI-wirable: make the erraudit repo public or vendor the binary                                                                      | Med    | M      | The gate has had no CI home for months (private repo blocks `go install` on runners). (10-12 f35.)                                                                                                                                                |

## 🔒 Owner-gated decisions (block gates or releases)

| Decision                                                                            | Why it blocks                                                                                                                               | Notes |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ----- |
| Tag-gating policy: DECIDED 2026-11-08 - documented exception, NOT phase-aware gates | Gates stay strict; the phased-CI-red interpretation table lives in docs/release-procedure.md. Gate code untouched.                          |       |
| Toolsdk versioning: DECIDED 2026-11-08 - INDEPENDENT, keep                          | Preflight false positives documented with the sub-module-only recipe (release-procedure.md); toolsdk joined the preflight build/tidy loops. |       |
| Dependabot gomod config: DECIDED 2026-11-08 - collapsed to `/`                      | Root PRs cover all modules; per-module labels dropped.                                                                                      |       |
| Branch protection contexts: DECIDED 2026-11-08 - wrapper jobs fix it in-repo        | test/lint aggregates satisfy the bare contexts; PRs are no longer structurally unmergeable. Admin-merge remains the direct-push path.       |       |
| codecov branch: DECIDED 2026-11-08 - DELETED                                        | Fix already on master; PR #2 was closed/superseded; the branch carried only daemon noise.                                                   |       |
| Launch comms mode: DECIDED 2026-11-08 - quiet-maintenance until the v1.16.0 train   | Announcement rewrite rides the train with stamp-correct artifacts + LAUNCH3/4 (see ROADMAP).                                                |       |

## 🟠 HIGH-MED Priority — ecosystem & docs

| Task                                                                                                                 | Impact | Effort | Notes                                                                                                                                                                                                          |
| -------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Render-verify the nix pipe LIVE on the next train, then retire --skip=nix                                            | Low    | S      | SNAPSHOT render proof exists (2026-11-08: goreleaser snapshot wrote dist/nix/pkgs/go-finding/default.nix; config now skip_upload=true after the nur-packages 404 fix). LIVE runner render remains train-gated. |
| Consumer bumps: remaining 21 repos already on core v1.15.1 (2026-11-08 sweep); next batch is opportunistic per train | Low    | Med    | clean-wizard + md-go-validator bumped to v1.15.1 this session; ecosystem.md carries the fresh pin table.                                                                                                       |

## 🟡 MEDIUM Priority — docs & testing quality

| Task                                                       | Impact | Effort | Notes                                                                                      |
| ---------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------ |
| Watch erraudit upstream for the nolint-audit staleness fix | Low    | Low    | Once fixed, `scripts/error-audit.sh --nolint-audit` becomes a usable removal oracle again. |
| Re-check library-policy after hook issue #74 resolves      | Low    | Low    | #74 still OPEN (verified 2026-09-23).                                                      |
| Dependabot watch: actions/* SHA pin-vs-bot decision        | Low    | Low    | Standing watch.                                                                            |

## 🟢 LOW Priority — launch & hygiene

| Task                                                                                                                        | Impact | Effort | Notes                                                                                                      |
| --------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------- |
| Submit Awesome Go entry (LAUNCH3) — owner fork + PR; needs owner go on external comms — owner fork + PR                     | Med    | Low    | Entry text ready in `docs/brainstorming/launch-announcement-draft.md`. 🔒 owner.                           |
| LAUNCH4: tap repo EXISTS (public); verify `HOMEBREW_TAP_GITHUB_TOKEN` secret exists + flip `homebrew_casks.skip_upload` off | Med    | S      | goreleaser brews section renders but ships `skip_upload: true`. ROADMAP open question. 🔒 owner.           |
| v2.0 design spike SESSION (agenda written: docs/planning/2026-11-08_v2-design-spike-agenda.md)                              | Low    | High   | Parked in [ROADMAP.md](ROADMAP.md) "Hardening": Position sentinel, FixStrategy union, TagSet, sub-structs. |
| `scripts/pre-push-verify.sh`: include shellcheck-gate + version-lattice in its structural list | Low | S | New gates from 2026-11-08 are preflight-wired but not yet in the pre-push battery. (2026-11-08 12:48 self-critique f/42.) |
| Detached long-gate runner (nohup + log + verdict line) for local stress/bench                                               | Low    | S      | An interrupted poll killed a stress gate once (10-12 d4/e4).                                               |

## Standing watches (checked each session, not actionable tasks)

| Watch                              | Where it bites                                                            | Last checked                   |
| ---------------------------------- | ------------------------------------------------------------------------- | ------------------------------ |
| library-policy hooks issue #74     | consumer build breakage; blocks that repo's bumps                         | 2026-11-08                     |
| Dependabot Updates scheduled run   | updater-infra failures (the 2026-09-27 GITHUB_REGISTRIES_PROXY class)     | 2026-11-08 (green since Nov 1) |
| erraudit upstream behavior changes | the error-audit gate's verdicts (see AGENTS nolint-audit re-verification) | 2026-11-08                     |
| CI-event watchdog scheduled run    | master red / silent CI drop (runs every 30 min; red X = the alert)        | live                           |

---

_Assisted-by: Crush <crush@charm.land>_
