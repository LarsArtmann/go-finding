# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).
> Owner questions (ApplySimpleFixes philosophy, brew/nix, report policy)
> live in [ROADMAP.md](ROADMAP.md) "Open questions" — not here.

> **Rebuilt 2026-09-23 (evening)** after the whole-list execution round: the
> v1.13.0 train was COMPLETED (all 5 tags + GitHub Releases, Latest = v1.13.0,
> proxy + consumer-compat verified), issue #36 closed, the release-integrity
> gates landed (changelog-drift, README version sweep, CI nix job,
> consumer-compat job, one-tag-per-push procedure, 5-tag preflight), the
> testing bundle landed (99% root coverage, fuzz campaigns, ordering/conflict
> pins), GOEXPERIMENT=jsonv2 was dropped (json/v2 GA in Go 1.27), and the four
> planned consumer filings resolved without filing (two claims verified stale,
> one fixed directly in go-business-rules 431d8cd, art-dupl GAP-2 already
> filed as #1). Execution record: the annotation log in
> `docs/planning/2026-09-23_15-01_SUPERB-pareto-plan-release-truth-and-gates.md`.

---

## 🔴 HIGH Priority — next train (v1.14.0) & gates that need a tag

| Task                                                                                                    | Impact | Effort | Notes                                                                                                                                                                                         |
| ------------------------------------------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cut the v1.14.0 train for the multi-edit feature (Edits, EditListProvider, ResolveFlightRecorderConfig) | Crit   | L      | Everything on master since 450824e. Includes: CHANGELOG cuts ([Unreleased] is primed), `features "(unreleased)"` labels flip to v1.14.0, preflight `--bench --stress`, 5 tags ONE PER PUSH.   |
| CLI: delegate `flightRecorderFileConfig.resolve()` to `pipeline.ResolveFlightRecorderConfig`            | Med    | S      | Prepared design in place; blocked on pipeline/v1.14.0 existing (no-replace policy). Kills the duration-parse split brain; wording family then unified via the pinned pipeline messages.       |
| Fix `.goreleaser.yml` nix-pipe template bug (`{{binary}}` is not a valid template field)                | Med    | S      | The v1.13.0 recovery release worked around it with `--skip=nix` (CI's skip_upload path never renders it). Fix before the next goreleaser config change; local dry-runs otherwise always fail. |
| Restore sigstore signing on the next workflow-run release train                                         | Med    | S      | v1.13.0 shipped WITHOUT `.sigstore.json` bundles (recovery release ran GoReleaser locally where cosign keyless has no OIDC). Checksums.txt is unsigned too.                                   |
| erraudit T13/T14 migration onto tagged analysis/pipeline APIs 🔒(after v1.14.0)                         | High   | M      | Blocked on the Edits bridge being TAGGED (not in v1.13.0). Also file the upstream erraudit issue: `nolint-audit` false-reports type-aware directives as stale (verified 2026-09-23, 24/24).   |
| Build `scripts/release-train.sh` + post-release verification (T11)                                      | Med    | L      | The prep→verify→tag→serialized-push→rerun→resync→postflight sequence ran manually three times now; encode it (proxy ×5, clean-dir `go get` ×5, assets, pkg.go.dev backoff).                   |

## 🟠 HIGH-MED Priority — follow-ups & ecosystem

| Task                                                       | Impact | Effort | Notes                                                                                                                                                                      |
| ---------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Opportunistic consumer bumps beyond the leads              | Low    | Med    | go-linter-sdk + golangci-lint-auto-configure are on v1.13.0 (done); the remaining v1.8.0-era rows bump per consumer. gomend + licenseforge stay blocked upstream (#1/#46). |
| Watch erraudit upstream for the nolint-audit staleness fix | Low    | Low    | Once fixed, the advisory listing in `scripts/error-audit.sh --nolint-audit` becomes usable as a removal oracle again.                                                      |
| Re-check library-policy after hook issue #74 resolves      | Low    | Low    | #74 still OPEN (verified 2026-09-23).                                                                                                                                      |

## 🟡 MEDIUM Priority — testing & code quality

| Task                                                                                 | Impact | Effort | Notes                                                                                                                                        |
| ------------------------------------------------------------------------------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Dependabot watch: gomega bumps merged; pin-vs-dependabot decision for actions/* SHAs | Low    | Low    | #25 merged; sbom bump was manual review.                                                                                                     |
| Periodic `/mnt/buildcache` hygiene                                                   | Low    | Low    | The shared GOCACHE hit 100% (167G) and broke every build; `go clean -cache` recovered. Consider a cron/prune or size alert before it recurs. |

## 🟢 LOW Priority — launch & product

| Task                                                                      | Impact | Effort | Notes                                                                                                                                 |
| ------------------------------------------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| Submit Awesome Go entry (LAUNCH3)                                         | Med    | Low    | Entry text ready in `docs/brainstorming/launch-announcement-draft.md` (finalized 2026-09-23); needs an awesome-go fork + PR. 🔒 owner |
| Create `LarsArtmann/homebrew-tap` + `HOMEBREW_TAP_GITHUB_TOKEN` (LAUNCH4) | Med    | S      | goreleaser brews section renders but ships `skip_upload: true` until the tap repo + secret exist. ROADMAP open question. 🔒 owner     |
| CONTRIBUTING / release-runbook page for public audience                   | Low    | Med    | Release procedure is internal-facing.                                                                                                 |
| v2.0 design spike session                                                 | Low    | High   | Parked in [ROADMAP.md](ROADMAP.md) "Hardening": Position sentinel, FixStrategy union, TagSet, sub-structs.                            |
| Sampling NO-GO review date (e.g. 2027-01 calendar entry)                  | Low    | Low    | ROADMAP has a revisit trigger; consider a date.                                                                                       |

---

_Assisted-by: Crush <crush@charm.land>_
