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

> v1.4.1 was tagged on 2026-07-28 (see [CHANGELOG](CHANGELOG.md)). All four module tags exist: `v1.4.1`, `pipeline/v1.4.1`, `analysis/v1.4.1`, `cmd/go-finding/v1.4.1`. Unreleased FlightRecorder work is in `[Unreleased]`.

## 🟡 MEDIUM Priority

| Task                              | Status       | Impact | Effort | Notes                                                                                                                                   |
| --------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |
| Commit benchmark baseline         | ⬜ `TODO`   | Med    | Low    | `/benchmarks/` is gitignored (`.gitignore:53`); `scripts/bench-check.sh` compares against `benchmarks/baseline.txt` which doesn't exist in a fresh clone. CI benchmark regression check has never worked. |

## 🟢 LOW Priority

| Task                         | Status       | Impact | Effort | Evidence                                                                             |
| ---------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                  |
| Consumer compatibility test  | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
