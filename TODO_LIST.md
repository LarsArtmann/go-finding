# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## HIGH Priority

### Release: Publish v1.6.0 (ParseConfidence + Template.Builder)

Two new public APIs are implemented, tested, and documented in `[Unreleased]` but have no git tag. Consumers cannot `go get` them.

| Task                                          | Status    | Impact | Effort | Notes                                                                                              |
| --------------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------------------------------------------------- |
| Bump `version.go` to v1.6.0                   | TODO      | High   | Low    | Minor bump: new public API surface (`ParseConfidence`, `Template.Builder`).                        |
| Move `[Unreleased]` to `[1.6.0]` in CHANGELOG | TODO      | High   | Low    | Append-only. Retain empty `[Unreleased]` section.                                                  |
| Tag all 4 modules (`v1.6.0` + sub-module tags) | TODO      | High   | Low    | Follow `docs/release-procedure.md`. Verify go.mod has real versions at the tagged commit.          |
| Run `version-check.sh` after tagging           | TODO      | High   | Low    | `OK: version.go (v1.6.0) matches tag (v1.6.0)`.                                                    |
| Run GOWORK=off isolation tests for all modules | TODO      | High   | Low    | `GOWORK=off GOEXPERIMENT=jsonv2 go test -race -count=1 ./...` in each module dir.                  |

### Documentation Fixes

| Task                                                                   | Status | Impact | Effort | Notes                                                                                                              |
| ---------------------------------------------------------------------- | ------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------ |
| Fix `config.ToConfig()` reference in configuration guide               | TODO   | High   | Low    | `docs/guides/configuration.md:178` calls `config.ToConfig()` but the method is unexported (`toConfig`). Misleading. |
| Fix `dprint` missing-binary in devShell                                | TODO   | Med    | Low    | Pre-commit hook bypassed with `--no-verify` on every commit. Add `dprint` to `flake.nix` or make it optional.      |

### Make Repo Public - Phase 2: Community Readiness (remaining)

| Task                                             | Status    | Impact | Effort  | Notes                                                     |
| ------------------------------------------------ | --------- | ------ | ------- | --------------------------------------------------------- |
| Verify pkg.go.dev renders after first public tag | TODO      | Med    | Low     | Triggered by first `go get` after visibility flip         |
| Track Go json/v2 stabilization (Go 1.27+)        | TODO      | Low    | Ongoing | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes |

### Make Repo Public - Phase 3: Launch

| Task                                             | Status    | Impact | Effort | Notes                                         |
| ------------------------------------------------ | --------- | ------ | ------ | --------------------------------------------- |
| Verify GoReleaser + Homebrew tap on public tag   | TODO      | Med    | Low    | `HOMEBREW_TAP_GITHUB_TOKEN` secret must exist |
| Write announcement (blog/r/golang/Slack/Twitter) | TODO      | High   | Medium | Drive adoption                                |
| Submit to Awesome Go                             | TODO      | Low    | Low    | Discoverability                               |

> v1.5.0 was tagged on 2026-08-06 and pushed to remote. All four module tags exist: `v1.5.0`, `pipeline/v1.5.0`, `analysis/v1.5.0`, `cmd/go-finding/v1.5.0`. The `[Unreleased]` section in CHANGELOG.md contains work since v1.5.0.

## MEDIUM Priority

| Task                                              | Status       | Impact | Effort | Notes                                                                                                                                   |
| ------------------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop                 | BLOCKED      | Med    | —      | External tool. BuildFlow's detect->repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |
| Consumer repo go.mod bump (go-humanize-linter)    | BLOCKED      | Med    | Low    | Needs go-finding v1.6.0 + go-linter-sdk v0.2.0 published first. Replace directives must be removed before publish.                     |

## LOW Priority

| Task                                    | Status       | Impact | Effort | Evidence                                                                                                                                                                                                                               |
| --------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Consumer compatibility test             | BLOCKED      | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code                                                                                                                                                   |
| Per-module golangci-lint configs        | TODO         | Low    | Medium | Workspace-level lint suffices but loses per-module precision. Flagged since modularization.                                                                                                                                            |
| Multi-module vs monolith benchmark      | DONE         | Low    | Medium | Report at `docs/reports/2026-08-08_multi-module-vs-monolith.md`. Zero runtime overhead, negligible build overhead, significant dependency isolation.                                                                                   |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
