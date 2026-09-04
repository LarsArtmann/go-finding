# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

### Post-Release: v1.6.0 Verification

v1.6.0 tagged and pushed (all 4 modules). Remaining verification:

| Task                                       | Status    | Impact | Effort | Notes                                                                                               |
| ------------------------------------------ | --------- | ------ | ------ | --------------------------------------------------------------------------------------------------- |
| Verify CI passed on v1.6.0 tag push        | ⬜ `TODO` | High   | Low    | `gh run list` — check lint, test, govulncheck, arch-check, structural-checks jobs                   |
| Verify release.yml created GitHub Releases | ⬜ `TODO` | High   | Low    | `gh release list` — check all 4 tags have releases                                                  |
| Verify GoReleaser built CLI binary         | ⬜ `TODO` | Med    | Low    | Check v1.6.0 release assets include the binary                                                      |
| Bump consumer repos to go-finding v1.6.0   | ⬜ `TODO` | High   | Low    | go-humanize-linter + go-linter-sdk go.mod bumps. Replace directives must be removed before publish. |

### Lint Debt Cleanup (24 issues)

Commit `4809b49` removed lint path exclusions, exposing pre-existing issues. CI only lints root module (0 issues) so these don't block CI yet.

| Task                            | Status    | Impact | Effort | Notes                                                                                                                     |
| ------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| Fix 6 pipeline lint issues      | ⬜ `TODO` | High   | Med    | contextcheck (asyncSnapshot ctx), gosec x2 (test Chmod), nilnil (ResolveFlightRecorder), revive x2 (missing doc comments) |
| Fix 18 CLI lint issues          | ⬜ `TODO` | High   | Med    | dupl (e2e test), err113 x2, exhaustruct, gocognit (run=47), goconst x5, gosec x2, nestif x2, varnamelen x2, wrapcheck     |
| Add pipeline/CLI to CI lint job | ⬜ `TODO` | Med    | Low    | Currently CI only lints root module. Add `working-directory` matrix or separate steps.                                    |

### Environment

| Task                                    | Status    | Impact | Effort | Notes                                                                                                         |
| --------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------- |
| Fix `dprint` missing-binary in devShell | ⬜ `TODO` | Med    | Low    | Pre-commit hook bypassed with `--no-verify` on every commit. Add `dprint` to `flake.nix` or make it optional. |

### Make Repo Public — Phase 2: Community Readiness (remaining)

| Task                                             | Status    | Impact | Effort  | Notes                                                     |
| ------------------------------------------------ | --------- | ------ | ------- | --------------------------------------------------------- |
| Verify pkg.go.dev renders after first public tag | ⬜ `TODO` | Med    | Low     | Triggered by first `go get` after visibility flip         |
| Track Go json/v2 stabilization (Go 1.27+)        | ⬜ `TODO` | Low    | Ongoing | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes |

### Make Repo Public — Phase 3: Launch

| Task                                             | Status    | Impact | Effort | Notes                                         |
| ------------------------------------------------ | --------- | ------ | ------ | --------------------------------------------- |
| Verify GoReleaser + Homebrew tap on public tag   | ⬜ `TODO` | Med    | Low    | `HOMEBREW_TAP_GITHUB_TOKEN` secret must exist |
| Write announcement (blog/r/golang/Slack/Twitter) | ⬜ `TODO` | High   | Medium | Drive adoption                                |
| Submit to Awesome Go                             | ⬜ `TODO` | Low    | Low    | Discoverability                               |

> v1.6.0 was tagged on 2026-08-08 and pushed to remote. All four module tags exist: `v1.6.0`, `pipeline/v1.6.0`, `analysis/v1.6.0`, `cmd/go-finding/v1.6.0`. All modules resolve on the proxy.

## 🟡 MEDIUM Priority

| Task                                         | Status       | Impact | Effort | Notes                                                                                                                                                                  |
| -------------------------------------------- | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Run `nix flake check`                        | ⬜ `TODO`    | Med    | Low    | Documented quality gate, consistently skipped across sessions. Run and fix issues.                                                                                     |
| Run stress tests (ginkgo --repeat=20)        | ⬜ `TODO`    | Med    | Med    | Release procedure step 3, consistently skipped. Catches non-deterministic failures.                                                                                    |
| Fix BuildFlow auto-configure loop            | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact                                |
| Full FEATURES.md vs code walk                | ⬜ `TODO`    | Med    | High   | Recurring gap since v1.3.0. Verify every method signature, status label, and config default against source.                                                            |
| Update API_STABILITY.md with v1.5.0+ symbols | ⬜ `TODO`    | Med    | Med    | Missing FlightRecorderHook, ValidateAll, ParseConfidence, Template.Builder, ResolveSafePath, Degraded, deterministic output guarantee.                                 |
| Update CONTRIBUTING.md project tree          | ⬜ `TODO`    | Med    | Low    | ~10 files missing. Flagged since v1.3.0. Regenerate from `git ls-files`.                                                                                               |
| Consumer migration guide (docs/guides/)      | ⬜ `TODO`    | Med    | Med    | 14 Go consumers can simplify using Template.Builder, ParseConfidence, ResolveSafePath, etc. No guide exists yet.                                                       |
| Archive 3 remaining annotated status reports | ⬜ `TODO`    | Low    | Low    | Flight-recorder self-critique, pareto self-critique, comprehensive session status — all have RESOLVED banners but still in `docs/status/` not `docs/status/archived/`. |
| Add `nix flake check` to release procedure   | ⬜ `TODO`    | Low    | Low    | Step in `docs/release-procedure.md` pre-release checklist.                                                                                                             |
| Add stress tests as mandatory release gate   | ⬜ `TODO`    | Low    | Low    | Currently step 3 in release procedure but consistently skipped. Make it a hard gate or remove.                                                                         |

## 🟢 LOW Priority

| Task                                      | Status       | Impact | Effort | Notes                                                                                                              |
| ----------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------ |
| Consumer compatibility test               | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code                               |
| Per-module golangci-lint configs          | ✅ `DONE`    | Low    | Medium | Removed path exclusions from root `.golangci.yml` — all 4 modules use single config.                               |
| Run `go-arch-lint check` locally          | ⬜ `TODO`    | Low    | Low    | Wired into CI as `arch-check` job but never run locally. Verify it passes.                                         |
| Investigate 6 docs-freshness.sh warnings  | ⬜ `TODO`    | Low    | Low    | 6 out-of-sync doc references (json.go, fix_applier.go, doc.go, sarif_roundtrip_test.go). Pre-existing, EXIT 0.     |
| Add nolint directives with justifications | ⬜ `TODO`    | Low    | Low    | contextcheck on asyncSnapshot (intentional ctx.Background()), nilnil on ResolveFlightRecorder (idiomatic nil,nil). |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
