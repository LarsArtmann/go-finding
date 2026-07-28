# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

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

> v1.4.0 was tagged on 2026-07-26 (see [CHANGELOG](CHANGELOG.md)). All four module tags exist: `v1.4.0`, `pipeline/v1.4.0`, `analysis/v1.4.0`, `cmd/go-finding/v1.4.0`.

## 🟡 MEDIUM Priority

| Task                              | Status       | Impact | Effort | Notes                                                                                                                                   |
| --------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |

### Pipeline pre-existing lint issues (7 findings)

All in the `pipeline/` module, pre-existing before the dedup-to-zero sweep. None in changed files.

| Task                                                         | Status    | Impact | Effort | Evidence                                        |
| ------------------------------------------------------------ | --------- | ------ | ------ | ----------------------------------------------- |
| Decompose `Pipeline.Run()` (137 lines, exceeds funlen 120)   | ⬜ `TODO` | Med    | Medium | `pipeline/pipeline.go`                          |
| Decompose `Pipeline.runIteration()` (124 lines, exceeds 120) | ⬜ `TODO` | Med    | Medium | `pipeline/pipeline_iteration.go`                |
| Reduce `applyTriage` cognitive complexity (36, limit 35)     | ⬜ `TODO` | Low    | Low    | `pipeline/pipeline_detect.go`                   |
| Fix `goast/provider.go` exhaustruct (`result{ok: false}`)    | ⬜ `TODO` | Low    | Low    | `pipeline/goast/provider.go:129`                |
| Wrap `errgroup.Wait()` error in `convenience.go`             | ⬜ `TODO` | Low    | Low    | `pipeline/convenience.go:67` (wrapcheck)        |
| Add `//nolint:gosec` to `fix_applier_test.go` path traversal | ⬜ `TODO` | Low    | Low    | `pipeline/fix_applier_test.go:578` (gosec G703) |
| Rename unused `s` receivers in `saboteurProvider` test mock  | ⬜ `TODO` | Low    | Low    | `pipeline/fix_applier_test.go` (revive)         |

## 🟢 LOW Priority

| Task                         | Status       | Impact | Effort | Evidence                                                                             |
| ---------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                  |
| Consumer compatibility test  | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
