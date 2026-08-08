# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

### FlightRecorder Bug Fixes (from self-critique)

| Task                                                           | Status    | Impact | Effort | Notes                                                                                                        |
| -------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------ |
| Serialize concurrent `WriteTo` calls with `writeMu sync.Mutex` | ✅ `DONE` | High   | Low    | Fixed: added `writeMu sync.Mutex` to FlightRecorderHook struct, serializes WriteTo calls.                    |
| Fix `sanitizeFilename("")` to return `"snapshot"`              | ✅ `DONE` | Med    | Low    | Fixed: empty/all-special-chars input now returns `"snapshot"`.                                               |
| Add concurrent-snapshot serialization test                     | ✅ `DONE` | High   | Low    | `TestFlightRecorderHook_ConcurrentSnapshotsDoNotCollide` — 10 concurrent snapshots, all succeed, all unique. |
| Add `MkdirAll` error path test                                 | ✅ `DONE` | Med    | Low    | `TestFlightRecorderHook_MkdirAllError` — unwritable dir returns error.                                       |

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

| Task                                              | Status       | Impact | Effort | Notes                                                                                                                                   |
| ------------------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop                 | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |
| Commit benchmark baseline                         | ✅ `DONE`    | Med    | Low    | Fixed in 2026-08-01 session. Removed `/benchmarks/` from `.gitignore`, committed `benchmarks/baseline.txt`.                             |
| FlightRecorder `ConfigFile` integration           | ✅ `DONE`    | Med    | Medium | Added `FlightRecorderFileConfig` to pipeline `config_file.go` + CLI YAML/JSON support + `ResolveFlightRecorder()` method. |
| Write `docs/guides/flight-recorder.md` user guide | ✅ `DONE`    | Med    | Low    | Full guide: CLI flags, config file, programmatic API, `go tool trace` workflow, config reference, internals. |
| FlightRecorder `example_test.go`                  | ✅ `DONE`    | Low    | Low    | Added `ExampleNewFlightRecorderHook` and `ExampleConfigFile_ResolveFlightRecorder` to `pipeline/example_test.go`. |
| Check `doc.go` for FlightRecorder API references  | ✅ `DONE`    | Low    | Low    | Added FlightRecorder to pipeline feature list + dedicated section in `doc.go`. |
| CLI integration test for `-trace` flag            | ✅ `DONE`    | Low    | Medium | `TestRun_E2E_TraceFlag` + `TestRun_E2E_TraceViaConfigFile` verify `.trace` file output. |

## 🟢 LOW Priority

| Task                                    | Status       | Impact | Effort | Evidence                                                                                    |
| --------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------- |
| SARIF schema validation test            | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                         |
| Consumer compatibility test             | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code        |
| Per-module golangci-lint configs        | ⬜ `TODO`    | Low    | Medium | Workspace-level lint suffices but loses per-module precision. Flagged since modularization. |
| go-arch-lint module boundary CI         | ⬜ `TODO`    | Low    | Medium | Enforce module dependency boundaries in CI. Flagged in modularization reports.              |
| `go.work sync` idempotency CI           | ✅ `DONE`    | Low    | Low    | `scripts/go-work-sync.sh` — runs `go work sync` twice, checks for changes.                  |
| Replace directive audit CI              | ✅ `DONE`    | Low    | Low    | `scripts/replace-audit.sh` — verifies all sub-module replace directives.                    |
| Version drift detection CI              | ✅ `DONE`    | Low    | Low    | `scripts/version-drift.sh` — cross-checks all 4 module versions against `version.go`.       |
| Test filename convention CI             | ✅ `DONE`    | Low    | Low    | `scripts/test-naming.sh` — rejects banned test file naming patterns.                       |
| Docs-freshness CI check                 | ⬜ `TODO`    | Low    | Medium | Flag docs older than N days without review.                                                 |
| Per-module CHANGELOG entries            | ⬜ `TODO`    | Low    | Medium | Each sub-module tracks its own changes. Flagged in modularization reports.                  |
| Multi-module vs monolith benchmark      | ⬜ `TODO`    | Low    | Medium | Measure overhead of workspace vs single-module.                                             |
| SARIF/LSP/FilePath round-trip benchmark | ✅ `DONE`    | Low    | Low    | Added `BenchmarkToLSP`, `BenchmarkFromLSP`, `BenchmarkLSPRoundTrip` to `bench_test.go`. SARIF benchmarks already existed. |
| TOCTOU symlink swap runtime test        | ✅ `DONE`    | Low    | Low    | `TestResolveSafePath_TOCOU_SymlinkSwap` + `TestResolveSafePath_SymlinkSwap_OutsideToInside` in `path_safety_test.go`. |

---

_FlightRecorder future ideas (trace file rotation, OTel bridge, trace diff, AI-assisted analysis) are tracked in [ROADMAP.md](ROADMAP.md) under "FlightRecorder future directions"._

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
