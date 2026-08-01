# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

### FlightRecorder Bug Fixes (from self-critique)

| Task | Status | Impact | Effort | Notes |
| --- | --- | --- | --- | --- |
| Serialize concurrent `WriteTo` calls with `writeMu sync.Mutex` | ⬜ `TODO` | High | Low | `runtime/trace.FlightRecorder.WriteTo` is NOT safe for concurrent calls. Two slow stages in the same iteration cause "already in progress" error. Self-critique §d.1. |
| Fix `sanitizeFilename("")` to return `"snapshot"` | ⬜ `TODO` | Med | Low | Empty input produces malformed filename `go-finding-trace-000-.trace` (trailing hyphen). Self-critique §d.3. |
| Add concurrent-snapshot serialization test | ⬜ `TODO` | High | Low | Prove the mutex prevents collision. Self-critique §f.3. |
| Add `MkdirAll` error path test | ⬜ `TODO` | Med | Low | Pass unwritable directory to `NewFlightRecorderHook`. Self-critique §f.4. |

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
| Commit benchmark baseline         | ✅ `DONE`   | Med    | Low    | Fixed in 2026-08-01 session. Removed `/benchmarks/` from `.gitignore`, committed `benchmarks/baseline.txt`. |
| FlightRecorder `ConfigFile` integration | ⬜ `TODO` | Med | Medium | Add `FlightRecorder *FlightRecorderConfig` to `config_file.go` schema so YAML/JSON config users can enable tracing. Self-critique §b.3. |
| Write `docs/guides/flight-recorder.md` user guide | ⬜ `TODO` | Med | Low | Full user guide with `go tool trace` workflow. Self-critique §c.8. |
| FlightRecorder `example_test.go` | ⬜ `TODO` | Low | Low | Runnable godoc examples for pipeline package. Self-critique §c.1. |
| Check `doc.go` for FlightRecorder API references | ⬜ `TODO` | Low | Low | Ensure new symbols are documented in godoc prose. Self-critique §c.10. |
| CLI integration test for `-trace` flag | ⬜ `TODO` | Low | Medium | Automated test that `-trace` produces a `.trace` file. Self-critique §c.9. |

## 🟢 LOW Priority

| Task                         | Status       | Impact | Effort | Evidence                                                                             |
| ---------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                  |
| Consumer compatibility test  | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_FlightRecorder future ideas (trace file rotation, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
