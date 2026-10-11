# Status Report: Flight Recorder Integration — Self-Critique

> **RESOLVED:** All P0 bugs fixed and shipped in v1.5.0. P1 feature gaps (ConfigFile integration, user guide, examples, CLI integration tests) all completed in subsequent sessions. See CHANGELOG `[1.5.0]` and `[Unreleased]` sections for details.

**Date:** 2026-08-01 19:40
**Session scope:** Leveraging Go 1.25 `runtime/trace.FlightRecorder` in go-finding
**Verdict:** Shipped a working feature. But I cut corners and left real bugs and gaps.

---

## a) FULLY DONE

1. **`pipeline/flight_recorder.go`** — `FlightRecorderHook` implementing `StageHook`. Wraps `runtime/trace.FlightRecorder`. Tracks per-stage durations via `StageBefore`/`StageAfter` events. Auto-snapshots when `SlowStageThreshold` exceeded. Manual `Snapshot(reason)` API. Thread-safe `Close()` (waits for in-flight snapshots before `fr.Stop()`). `OnStageEvent` never returns errors (diagnostic-only, never aborts pipeline). Stdlib-only, zero new dependencies.

2. **`pipeline/flight_recorder_test.go`** — 11 tests: lifecycle, idempotent close, never-errors contract, manual snapshot writes file, snapshot-after-close fails with sentinel, slow-stage auto-snapshot, fast-stage no-snapshot, pipeline integration, multiple snapshots, interface compliance, filename sanitization. All pass with `-race`.

3. **CLI flags in `cmd/go-finding/main.go`** — `-trace` (enable), `-trace-dir` (output directory), `-trace-slow` (auto-snapshot threshold). Hook registered before `pipeline.New()` (fixed the initial bug where it was added after construction). Error-path auto-snapshot. Proper cleanup on all exit paths.

4. **CLI smoke test** — Verified trace files are created (24KB valid `.trace` file), parseable by `go tool trace`.

5. **AGENTS.md updated** — Key Files table, Important Behaviors, CLI Features, Architecture Decisions.

6. **All tests pass** — Full workspace `-race -count=1` green across all 4 modules.

7. **Lint clean** — `golangci-lint` passes on all new/modified files. `gofmt` clean. `go vet` clean.

---

## b) PARTIALLY DONE

1. **Testing approach** — Tests are plain `testing` + `gomega`, NOT Ginkgo BDD. The project convention is Ginkgo for pipeline tests (the BDD suite has 27 specs). My tests work but don't follow the established pattern. Justification: flight recorder is a global singleton resource (only one active at a time), so `t.Parallel()` via `NewParallelGomega` is unsafe. But I could have used Ginkgo without parallel.

2. **Documentation** — AGENTS.md is updated, but no guide doc (`docs/guides/flight-recorder.md`), no `FEATURES.md` entry, no `CHANGELOG.md` entry, no `doc.go` API reference check.

3. **Config file integration** — Flight recorder is CLI-flag-only. The `ConfigFile` YAML/JSON system (`config_file.go`) doesn't know about it. Users who configure everything via `-config config.yaml` can't enable tracing from config.

---

## c) NOT STARTED

1. **example_test.go** — Project convention is one `example_test.go` per package. No flight recorder examples exist.
2. **Benchmark** — No benchmark for flight recorder overhead on pipeline execution.
3. **GOWORK=off verification** — Did not test per-module isolation (`GOWORK=off go test`) for the pipeline module.
4. **Nix flake check** — Did not run `nix flake check` or `nix run .#test` / `nix run .#lint`.
5. **CHANGELOG.md** — No entry.
6. **FEATURES.md** — No entry.
7. **ROADMAP.md** — No future enhancement noted.
8. **docs/guides/flight-recorder.md** — No user guide.
9. **Integration test for CLI `-trace` flag** — Tested manually, no automated test.
10. **doc.go reference check** — AGENTS.md says "After ANY rename, grep doc.go". Didn't check if doc.go needs flight recorder API references.

---

## d) TOTALLY FUCKED UP

1. ~~**CRITICAL BUG: Concurrent `WriteTo` race condition**~~ **FIXED** — Added `writeMu sync.Mutex` to serialize WriteTo calls. Shipped in v1.5.0. Test: `TestFlightRecorderHook_ConcurrentSnapshotsDoNotCollide`.

2. ~~**Initial CLI wiring was broken**~~ **FIXED** — Hook is now registered before `pipeline.New()`. Caught and fixed during original session smoke testing.

3. ~~**Empty `SanitizeFilename("")` edge case**~~ **FIXED** — Now returns `"snapshot"` as default. Shipped in v1.5.0. Test: `TestSanitizeFilename_AllSpecialChars`. Fuzz test: `FuzzSanitizeFilename`.

4. ~~**No test for `MkdirAll` failure**~~ **FIXED** — `TestFlightRecorderHook_MkdirAllError` tests unwritable directory. Shipped in v1.5.0.

---

## e) WHAT WE SHOULD IMPROVE

### Code Quality

1. **Serialize snapshot writes** — Add a `writeMu sync.Mutex` around `writeSnapshot` to prevent concurrent `WriteTo` calls. This is the most important fix.
2. **Sync.WaitGroup.Go compatibility** — Used `WaitGroup.Go` (Go 1.25+). Must verify this works in `GOWORK=off` per-module builds where the go.mod says `go 1.26.5`.
3. **FlightRecorderConfig should be in ConfigFile** — Add `FlightRecorder *FlightRecorderConfig` to the config file schema so YAML users can configure it.
4. **SanitizeFilename empty-string handling** — Return `"snapshot"` when input sanitizes to empty.
5. **Context propagation** — `writeSnapshot` doesn't accept a context. Long-running `WriteTo` calls can't be cancelled. Should pass the pipeline context through.

### Testing

6. **Convert to Ginkgo BDD** — Match the existing pipeline test suite convention.
7. **Add concurrent-snapshot test** — Explicitly test that two rapid snapshots don't collide.
8. **Add MkdirAll error path test** — Pass an unwritable directory.
9. **Add CLI integration test** — Automated test that `-trace` produces a `.trace` file.
10. **Add benchmark** — Measure overhead of continuous flight recording on pipeline execution time.

### Documentation

11. **docs/guides/flight-recorder.md** — Full user guide with `go tool trace` workflow.
12. **CHANGELOG.md entry** — Document the new feature.
13. **FEATURES.md entry** — Add to the feature inventory.
14. **example_test.go** — Runnable examples for godoc.
15. **doc.go check** — Verify API references are current.

### Architecture

16. **Core package trace helper** — The flight recorder concept could be generalized beyond the pipeline. A `finding.TraceSnapshot()` helper in core would let any consumer capture traces.
17. **Multiple recorder support** — Go's "one active recorder at a time" limit means the hook should detect if another recorder is already running and degrade gracefully (log a warning, not fail).
18. **Trace file rotation** — Current implementation writes numbered files. For long-running pipelines with `-trace-slow`, this could produce hundreds of files. Should add max-files or rotation.

---

## f) Up to 50 Things We Should Get Done Next

### P0 — Fix the bugs I shipped

1. ~~**Add `writeMu sync.Mutex` to serialize concurrent `WriteTo` calls** — Prevents the "already in progress" error~~ done (v1.5.0 CHANGELOG, writeMu race fix)
2. ~~**Fix `SanitizeFilename("")` to return `"snapshot"`** — Prevents malformed filenames~~ done (v1.5.0 CHANGELOG, sanitizeFilename fix)
3. ~~**Add test for concurrent snapshot serialization** — Prove the mutex works~~ done (test shipped, in-file §d.1 FIXED marker)
4. ~~**Add test for `MkdirAll` error path** — Pass `/dev/null/x` as output dir~~ done (test shipped, in-file §d.4 FIXED marker)
5. ~~**Run `GOWORK=off go test` in pipeline module** — Verify `WaitGroup.Go` doesn't break per-module builds~~ done (verified in 2026-08-06 v1.5.0 release session)

### P1 — Close the feature gap

6. ~~**Add `FlightRecorderConfig` to `ConfigFile` schema** — YAML/JSON config support~~ done (v1.6.0 CHANGELOG, FlightRecorderFileConfig)
7. ~~**Add `FlightRecorder` section to `config_file.go`** — Parse from config~~ done (v1.6.0 CHANGELOG, CLI flightRecorder section)
8. ~~**Write `docs/guides/flight-recorder.md`** — Full user guide~~ done (exists, docs/guides/flight-recorder.md)
9. ~~**Add `CHANGELOG.md` entry** — Under unreleased~~ done (v1.5.0 CHANGELOG)
10. ~~**Add `FEATURES.md` entry** — Mark as DONE~~ done (FEATURES updated, in-file Resolution appendix)
11. ~~**Write `example_test.go`** — Godoc runnable examples~~ done (example_test.go, 2026-08-08 session)
12. ~~**Check `doc.go` for API reference updates** — Ensure new symbols are documented~~ done (doc.go updated 2026-08-08)
13. ~~**Convert tests to Ginkgo BDD** — Match pipeline convention~~ **Won't implement — Ginkgo conversion deliberately skipped.**
14. ~~**Add CLI integration test** — Automated `-trace` flag test~~ done (CLI E2E -trace tests, 2026-08-08 session)
15. **Add benchmark** — `BenchmarkPipelineWithFlightRecorder`

### P2 — Polish and hardening

16. ~~**Context propagation through `writeSnapshot`** — Allow cancellation~~ done (v1.6.0 CHANGELOG, Snapshot ctx signature)
17. ~~**Detect pre-existing flight recorder** — Handle `Start()` error gracefully with a helpful message~~ done (v1.6.0 CHANGELOG, Degraded mode)
18. **Add trace file count limit** — Prevent disk fill on long runs with `-trace-slow`
19. **Add trace file size metric** — Log snapshot sizes to metrics
20. ~~**Add `-trace-min-age` CLI flag** — Currently hardcoded to 30s default~~ done (v1.6.0 CHANGELOG, config-file minAge full parity)
21. ~~**Add `-trace-max-bytes` CLI flag** — Currently hardcoded to 4 MiB~~ done (v1.6.0 CHANGELOG, config-file maxBytes full parity)
22. **Consider `io.Discard` mode** — Flight recorder enabled but snapshots discarded (overhead measurement only)
23. **Add `SnapshotN()` method** — Return last N snapshots as `[][]byte`
24. **Add `Snapshots() []string`** — Return all snapshot file paths
25. **Consider in-memory snapshot mode** — Return `[]byte` instead of writing to disk

### P3 — Broader integration

26. **Core package trace helper** — `finding.TraceSnapshot(ctx) ([]byte, error)` in core
27. **Integration with Metrics** — Add trace snapshot count to `MetricsSnapshot`
28. **Integration with StageHook event enrichment** — Include trace file path in `StageEvent`
29. **PipelineResult.TraceFiles field** — Expose trace file paths on the result
30. **LSP integration** — Trace file path in LSP diagnostic data
31. **SARIF integration** — Trace artifact in SARIF run metadata
32. **Detector-level tracing** — Per-detector trace regions with `runtime/trace.WithRegion`
33. **FixProvider tracing** — Trace regions around fix provider calls
34. **FixEngine tracing** — Trace regions around edit application
35. **Correlate tracing** — Trace region around correlation

### P4 — Operational

36. ~~**Nix flake check** — Verify the full build passes~~ done (nix flake check green 2026-09-08)
37. ~~**Version bump** — This is a feature, warrants at least minor version bump~~ done (v1.5.0 tagged 2026-08-06)
38. ~~**Release notes draft** — Document the flight recorder as a headline feature~~ done (v1.5.0 CHANGELOG release notes)
39. ~~**README.md update** — Mention flight recorder in the features list~~ done (README.md feature list)
40. ~~**docs/DOMAIN_LANGUAGE.md** — Add "Flight Recorder", "Trace Snapshot", "Slow Stage Threshold" terms~~ done (v1.6.0 CHANGELOG, DOMAIN_LANGUAGE)
41. ~~**CI pipeline** — Ensure `-trace` doesn't break CI runs~~ done (E2E trace-config tests in suite)
42. ~~**Performance regression check** — Run `bench-check.sh` to verify no regression~~ done (benchmarks/baseline.txt + bench-check.sh)
43. ~~**Consumer migration guide** — How consumers enable flight recording via library API~~ done (docs/guides/consumer-migration-v1.7.md)
44. ~~**go tool trace guide** — How to interpret the trace output for pipeline-specific patterns~~ done (docs/guides/flight-recorder.md)
45. ~~**Troubleshooting guide** — Common trace analysis scenarios (slow detector, fix contention, GC pause)~~ done (v1.6.0 CHANGELOG, troubleshooting guide)

### P5 — Future experiments

46. **Continuous trace sampling** — Sample 1% of pipeline runs with full tracing
47. **Trace diff tool** — Compare two trace snapshots to identify what changed between fast/slow runs
48. **OpenTelemetry integration** — Bridge trace snapshots to OTel spans
49. **Programmatic trace parsing** — Use the forthcoming `trace/flightrecorder` parse API (issue #62627)
50. **AI-assisted trace analysis** — Feed trace data to an LLM for anomaly detection

---

## g) Questions I Cannot Answer Myself

1. **Should the flight recorder be a pipeline-only feature, or should it live in the core module?** The core module has minimal dependencies by design. `runtime/trace` is stdlib, so it wouldn't add a dependency. But it would expand core's responsibility beyond "data model." I chose pipeline-only because that's where `StageHook` lives, but a case could be made for a `finding/tracing` sub-package.

2. **Should trace files contain pipeline metadata (findings count, stage info) or pure Go runtime trace?** Currently they're pure runtime traces — you need to correlate with logs to know which pipeline stage produced the slow region. Embedding pipeline metadata would require either a custom trace format (breaking `go tool trace` compatibility) or a sidecar JSON file. I chose pure runtime trace for tool compatibility.

3. **Should `-trace` be opt-in (current) or should the flight recorder always run with a tiny buffer?** Always-on with a 256KB buffer would add ~2-3% overhead but give every user "hindsight" capability for free. The Go blog example uses opt-in. I chose opt-in for safety, but the always-on approach would be more useful in practice.

---

## Resolution (2026-08-01)

**Items routed:**

- **P0 bugs** (concurrent WriteTo race, SanitizeFilename edge case, missing tests) → TODO_LIST HIGH priority
- **Feature gaps** (ConfigFile integration, flight-recorder guide, example_test.go, doc.go check, CLI integration test) → TODO_LIST MEDIUM priority
- **Future directions** (trace rotation, OTel bridge, trace diff, AI analysis, continuous sampling) → ROADMAP "FlightRecorder future directions"
- **Already resolved by docs-health session:** CHANGELOG.md `[Unreleased]` entry ✓, FEATURES.md §16.13 ✓, AGENTS.md gotchas ✓, README Pipeline Features table ✓

**Items NOT routed** (judged not actionable or out of scope):

- P2 items 20-25 (CLI flags for min-age/max-bytes, in-memory snapshot mode) — implementation details better decided when ConfigFile integration is done
- P3 items 26-35 (core trace helper, LSP/SARIF integration, per-detector tracing) — captured in ROADMAP broader ideas
- P4 items 40-45 (DOMAIN_LANGUAGE, troubleshooting guide) — will be addressed when writing the user guide (TODO_LIST)
