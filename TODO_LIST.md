# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

### Release: Publish v1.6.0 (ParseConfidence + Template.Builder)

Two new public APIs are implemented, tested, and documented in `[Unreleased]` but have no git tag. Consumers cannot `go get` them.

| Task                                          | Status    | Impact | Effort | Notes                                                                                              |
| --------------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------------------------------------------------- |
| Bump `version.go` to v1.6.0                   | ⬜ `TODO` | High   | Low    | Minor bump: new public API surface (`ParseConfidence` at `confidence.go:86`, `Template.Builder` at `finding_builder.go:216`). |
| Move `[Unreleased]` to `[1.6.0]` in CHANGELOG | ⬜ `TODO` | High   | Low    | Append-only. Retain empty `[Unreleased]` section.                                                  |
| Tag all 4 modules (`v1.6.0` + sub-module tags) | ⬜ `TODO` | High   | Low    | Follow `docs/release-procedure.md`. Verify go.mod has real versions at the tagged commit (`git show <tag>:cmd/go-finding/go.mod`). |
| Run `version-check.sh` after tagging           | ⬜ `TODO` | High   | Low    | `OK: version.go (v1.6.0) matches tag (v1.6.0)`.                                                    |
| Run GOWORK=off isolation tests for all modules | ⬜ `TODO` | High   | Low    | `GOWORK=off GOEXPERIMENT=jsonv2 go test -race -count=1 ./...` in each module dir.                  |

### Documentation Fixes

| Task                                                                   | Status    | Impact | Effort | Notes                                                                                                              |
| ---------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------ |
| Fix `dprint` missing-binary in devShell                                | ⬜ `TODO` | Med    | Low    | Pre-commit hook bypassed with `--no-verify` on every commit. Add `dprint` to `flake.nix` or make it optional.      |

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

> v1.5.0 was tagged on 2026-08-06 and pushed to remote. All four module tags exist: `v1.5.0`, `pipeline/v1.5.0`, `analysis/v1.5.0`, `cmd/go-finding/v1.5.0`. The `[Unreleased]` section in CHANGELOG.md contains work since v1.5.0.

## 🟡 MEDIUM Priority

| Task                                              | Status       | Impact | Effort | Notes                                                                                                                                   |
| ------------------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop                 | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect->repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |
| Consumer repo go.mod bump (go-humanize-linter)    | 🔵 `BLOCKED` | Med    | Low    | Needs go-finding v1.6.0 + go-linter-sdk v0.2.0 published first. Replace directives must be removed before publish.                     |
| Full FEATURES.md vs code walk                     | ⬜ `TODO`    | Med    | High   | Recurring gap since v1.3.0. Verify every method signature, status label, and config default against source.                             |
| Update API_STABILITY.md with v1.5.0+ symbols      | ⬜ `TODO`    | Med    | Med    | Missing FlightRecorderHook, ValidateAll, ParseConfidence, Template.Builder, deterministic output guarantee. Flagged since v1.3.0.      |
| Update CONTRIBUTING.md project tree               | ⬜ `TODO`    | Med    | Low    | ~10 files missing. Flagged since v1.3.0. Regenerate from `git ls-files`.                                                               |
| Consumer migration guide (docs/guides/)            | ⬜ `TODO`    | Med    | Med    | 14 Go consumers can simplify using Template.Builder, ParseConfidence, etc. No guide exists yet.                                        |

## 🟢 LOW Priority

| Task                                              | Status       | Impact | Effort | Evidence                                                                                          |
| ------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------- |
| Consumer compatibility test                       | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code              |
| Per-module golangci-lint configs                  | ✅ `DONE`    | Low    | Medium | Removed path exclusions from root `.golangci.yml` — all 4 modules lint clean with single config. Simpler than duplicating. |
| Export ResolveSafePath/ResolveSafePathFrom        | ✅ `DONE`    | Low    | Low    | Exported in `pipeline/path_safety.go`. Consumers can now validate paths against root dir.         |
| FlightRecorder context propagation                | ✅ `DONE`    | Low    | Medium | `Snapshot(ctx, reason)` + `writeSnapshot(ctx, ...)` accept context. asyncSnapshot uses `context.Background()`. |
| FlightRecorder graceful degradation               | ✅ `DONE`    | Low    | Medium | `NewFlightRecorderHook` detects singleton conflict, enters degraded mode. `Degraded()` method.   |
| Refine docs-freshness.sh false-positive matching  | ✅ `DONE`    | Low    | Medium | Now matches only backtick code spans and markdown links, not prose mentions of `.go` files.       |
| Extract marshalOpts package-level constant        | ✅ `DONE`    | Low    | Low    | `marshalOpts` + `prettyMarshalOpts` in `json.go`. Single source of truth for deterministic JSON.  |
| Add CI check for json.Marshal without Deterministic | ✅ `DONE`  | Low    | Medium | `scripts/json-deterministic-check.sh` with paren-depth-aware extraction. Wired into CI.           |

---

_FlightRecorder future ideas (trace file rotation, gzip, pprof, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags->TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
