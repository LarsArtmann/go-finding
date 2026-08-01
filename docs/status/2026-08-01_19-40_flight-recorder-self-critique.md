# Status Report: Flight Recorder Integration — Self-Critique

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

1. **CRITICAL BUG: Concurrent `WriteTo` race condition** — During debugging I observed `"call to WriteTo for trace.FlightRecorder already in progress"`. The `runtime/trace.FlightRecorder.WriteTo` method is NOT safe for concurrent calls. If two stages exceed `SlowStageThreshold` in the same iteration (e.g., `StageDetect` and `StageApply` both slow), the second `asyncSnapshot` goroutine will fail because the first hasn't finished writing yet. **I saw this error in my own debug output and shipped anyway.** The fix is a dedicated write mutex or a serialized snapshot channel.

2. **Initial CLI wiring was broken** — I added the hook to `pipelineCfg.StageHooks` AFTER `pipeline.New()` had already copied the config into the `Pipeline` struct. The hook was silently ignored. Caught during smoke testing, but I should have thought about object ownership before writing the code.

3. **Empty `SanitizeFilename("")` edge case** — Returns `""`, which produces a filename like `go-finding-trace-000-.trace` (trailing hyphen). Not tested in the table-driven test — I actually have `{"", ""}` in the test but the resulting filename is still malformed.

4. **No test for `MkdirAll` failure** — I added `os.MkdirAll(config.OutputDir, 0o755)` but never tested the error path. If the directory can't be created, `NewFlightRecorderHook` returns an error, but there's no test proving this.

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

1. **Add `writeMu sync.Mutex` to serialize concurrent `WriteTo` calls** — Prevents the "already in progress" error
2. **Fix `SanitizeFilename("")` to return `"snapshot"`** — Prevents malformed filenames
3. **Add test for concurrent snapshot serialization** — Prove the mutex works
4. **Add test for `MkdirAll` error path** — Pass `/dev/null/x` as output dir
5. **Run `GOWORK=off go test` in pipeline module** — Verify `WaitGroup.Go` doesn't break per-module builds

### P1 — Close the feature gap

6. **Add `FlightRecorderConfig` to `ConfigFile` schema** — YAML/JSON config support
7. **Add `FlightRecorder` section to `config_file.go`** — Parse from config
8. **Write `docs/guides/flight-recorder.md`** — Full user guide
9. **Add `CHANGELOG.md` entry** — Under unreleased
10. **Add `FEATURES.md` entry** — Mark as DONE
11. **Write `example_test.go`** — Godoc runnable examples
12. **Check `doc.go` for API reference updates** — Ensure new symbols are documented
13. **Convert tests to Ginkgo BDD** — Match pipeline convention
14. **Add CLI integration test** — Automated `-trace` flag test
15. **Add benchmark** — `BenchmarkPipelineWithFlightRecorder`

### P2 — Polish and hardening

16. **Context propagation through `writeSnapshot`** — Allow cancellation
17. **Detect pre-existing flight recorder** — Handle `Start()` error gracefully with a helpful message
18. **Add trace file count limit** — Prevent disk fill on long runs with `-trace-slow`
19. **Add trace file size metric** — Log snapshot sizes to metrics
20. **Add `-trace-min-age` CLI flag** — Currently hardcoded to 30s default
21. **Add `-trace-max-bytes` CLI flag** — Currently hardcoded to 4 MiB
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

36. **Nix flake check** — Verify the full build passes
37. **Version bump** — This is a feature, warrants at least minor version bump
38. **Release notes draft** — Document the flight recorder as a headline feature
39. **README.md update** — Mention flight recorder in the features list
40. **docs/DOMAIN_LANGUAGE.md** — Add "Flight Recorder", "Trace Snapshot", "Slow Stage Threshold" terms
41. **CI pipeline** — Ensure `-trace` doesn't break CI runs
42. **Performance regression check** — Run `bench-check.sh` to verify no regression
43. **Consumer migration guide** — How consumers enable flight recording via library API
44. **go tool trace guide** — How to interpret the trace output for pipeline-specific patterns
45. **Troubleshooting guide** — Common trace analysis scenarios (slow detector, fix contention, GC pause)

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
