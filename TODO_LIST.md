# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

### Make Repo Public — Phase 1: Critical Blockers

_See [`docs/PRO_CONTRA_make-public.md`](docs/PRO_CONTRA_make-public.md) for full analysis._

| Task                                             | Status    | Impact   | Effort  | Notes                                                         |
| ------------------------------------------------ | --------- | -------- | ------- | ------------------------------------------------------------- |
| Make `GOEXPERIMENT=jsonv2` prominent in README   | ✅ `DONE` | Critical | Trivial | Prerequisites block added at top of Installation (2026-07-24) |
| Fix `PUBLIC_OR_PRIVATE.md` false "public" banner | ✅ `DONE` | High     | Trivial | Corrected 2026-07-24                                          |
| Remove `GOPRIVATE` warning from README           | ✅ `DONE` | High     | Trivial | Removed 2026-07-24; no longer needed once public              |
| Add GitHub repo description + topics             | ✅ `DONE` | High     | Trivial | `gh repo edit` set description + 11 topics (2026-07-24)       |

### Make Repo Public — Phase 2: Community Readiness

| Task                                               | Status    | Impact    | Effort                                            | Notes                                                                   |
| -------------------------------------------------- | --------- | --------- | ------------------------------------------------- | ----------------------------------------------------------------------- |
| Add `SECURITY.md`                                  | ✅ `DONE` | Med       | Trivial                                           | Vulnerability reporting policy (GitHub private advisories) (2026-07-24) |
| Add `CODE_OF_CONDUCT.md`                           | ✅ `DONE` | Med       | Trivial                                           | Contributor Covenant v2.1 (2026-07-24)                                  |
| Add `.github/ISSUE_TEMPLATE/` (bug + feature)      | ✅ `DONE` | Med       | Trivial                                           | bug + feature templates + config.yml (2026-07-24)                       |
| Add `.github/PULL_REQUEST_TEMPLATE.md`             | ✅ `DONE` | Med       | Trivial                                           | PR quality checklist (2026-07-24)                                       |
| Add support policy to README                       | ✅ `DONE` | Med       | Trivial                                           | "Support" section added before Versioning (2026-07-24)                  |
| ~~Decide on 91 internal docs (move/archive/keep)~~ | ✅ `DONE` | No action | Keep all docs in place (user decision 2026-07-24) |
| Verify pkg.go.dev renders after first public tag   | ⬜ `TODO` | Med       | Low                                               | Triggered by first `go get` after visibility flip                       |
| Track Go json/v2 stabilization (Go 1.27+)          | ⬜ `TODO` | Low       | Ongoing                                           | Remove `GOEXPERIMENT` requirement when json/v2 stabilizes               |

### Make Repo Public — Phase 3: Launch

| Task                                             | Status    | Impact | Effort | Notes                                         |
| ------------------------------------------------ | --------- | ------ | ------ | --------------------------------------------- |
| Tag v1.4.0 (or next minor)                       | ⬜ `TODO` | Med    | Low    | Public version anchor                         |
| Verify GoReleaser + Homebrew tap on public tag   | ⬜ `TODO` | Med    | Low    | `HOMEBREW_TAP_GITHUB_TOKEN` secret must exist |
| Write announcement (blog/r/golang/Slack/Twitter) | ⬜ `TODO` | High   | Medium | Drive adoption                                |
| Submit to Awesome Go                             | ⬜ `TODO` | Low    | Low    | Discoverability                               |

## 🟡 MEDIUM Priority

| Task                              | Status       | Impact | Effort | Evidence                                                                                                                                |
| --------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |

## 🟢 LOW Priority

| Task                         | Status       | Impact | Effort | Evidence                                                                             |
| ---------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                  |
| Consumer compatibility test  | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
