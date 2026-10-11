# Status Report: 2026-08-08 22:11 — Pareto T1-T7 TODO Execution & Self-Critique

## Session Context

Executed 7 TODO items from the `paste_1.txt` TODO list. All were `Low` priority items flagged across multiple reports. The auto-git daemon committed the work in 4 changesets.

---

## A) FULLY DONE

### 1. Extract marshalOpts Package-Level Constant

**Commit:** `a93b749`

- Created `marshalOpts = json.Deterministic(true)` and `prettyMarshalOpts` (deterministic + indent) in `json.go`
- Replaced all inline `json.Deterministic(true)` across `json.go` (6 sites), `sarif_export.go` (2 sites)
- Added missing `json.Deterministic(true)` to `pipeline/fix_edit.go` `MarshalJSON`
- Removed now-unused `encoding/json/jsontext` import from `sarif_export.go`
- Suppressed `gochecknoglobals` for the new vars in `.golangci.yml`
- **Single source of truth** — new call sites cannot forget the option

### 2. Export ResolveSafePath / ResolveSafePathFrom / ResolveRoot

**Commit:** `a93b749`

- Capitalized all three functions in `pipeline/path_safety.go`
- Updated all internal callers: `fix_applier.go`, `pipeline_detect.go`, `fix_applier_test.go`, `path_safety_test.go` (52 refs)
- Build + tests pass cleanly
- Consumers can now validate paths against a root directory themselves

### 3. FlightRecorder Context Propagation

**Commit:** `a93b749`

- `Snapshot(ctx, reason)` now accepts `context.Context` as first arg
- `writeSnapshot(ctx, num, reason)` checks `ctx.Err()` before writing
- `asyncSnapshot` uses `context.Background()` internally — critical fix discovered during testing: `Pipeline.Run()` wraps ctx with `WithTimeout` + `defer cancel()`, which fires on return, killing the async snapshot goroutine's context before it writes
- Updated all callers: 7 test sites, 1 CLI site (`main.go`)
- Added `TestFlightRecorderHook_SnapshotWithCancelledContext` test
- Updated doc.go and docs/guides/flight-recorder.md with new signatures

### 4. FlightRecorder Multiple Recorder Graceful Degradation

**Commit:** `a93b749`

- `NewFlightRecorderHook` detects "flight recorder already enabled" error and enters degraded mode
- Added `degraded bool` field to `FlightRecorderHook`
- Added `Degraded() bool` method
- All snapshot/stage operations silently skip when degraded
- Logger receives warning if configured
- Added `TestFlightRecorderHook_DegradedWhenConflict` test (verifies degraded mode, error on snapshot, no-op OnStageEvent)
- Updated docs/guides/troubleshooting.md with new behavior

### 5. CI Check for json.Marshal Without Deterministic

**Commit:** `d39872a`

- Created `scripts/json-deterministic-check.sh` with paren-depth-aware extraction
- Correctly handles multi-line `json.Marshal(...)` calls with nested function arguments
- Excludes test files, generated files, doc.go
- Wired into `.github/workflows/ci.yml` under `structural-checks` job
- Verified: passes on real codebase, correctly flags simulated violations

### 6. Refine docs-freshness.sh False-Positive Matching

**Commit:** `d39872a`

- Changed grep from `` `?\K[a-zA-Z0-9_/]+\.go `` (optional backtick, matches prose) to two patterns:
  - `` `\K[a-zA-Z0-9_/]+\.go(?=`) `` (backtick code spans only)
  - `\]\(\K[a-zA-Z0-9_/]+\.go(?=\))` (markdown links only)
- Eliminates false positives from prose mentions like "see finding.go"

### 7. Per-Module golangci-lint Coverage

**Commit:** `4809b49`

- **Decision:** Instead of creating 3 duplicated per-module configs, removed the path exclusions for `pipeline/` and `cmd/go-finding/` from root `.golangci.yml`
- All 4 modules now lint clean with a single root config (0 issues)
- Both lint and formatter exclusions removed
- Simpler, DRY, no config drift risk

### Documentation Updates

**Commit:** `4809b49` (unstaged: AGENTS.md, doc.go, guides)

- Updated AGENTS.md: marshalOpts pattern, exported path safety API, flight recorder context + degradation, CI script count (6→7), docs-freshness matching refinement
- Updated doc.go: `Snapshot(ctx, reason)` signature
- Updated docs/guides/flight-recorder.md: `Snapshot(ctx, reason)` in prose + code examples + lifecycle description
- Updated docs/guides/troubleshooting.md: graceful degradation troubleshooting advice

### Verification

- `go build ./...` — all 4 modules clean
- `go test -race -count=1 ./...` — all modules pass
- `golangci-lint run ./...` — 0 issues
- `golangci-lint fmt --diff ./...` — clean
- `GOWORK=off go build` — all 4 modules build standalone
- All 7 structural CI checks pass

---

## B) PARTIALLY DONE

Nothing is partially done. All 7 tasks were completed end-to-end.

---

## C) NOT STARTED

No items from the original TODO list were skipped. All 7 were executed.

---

## D) TOTALLY FUCKED UP

Nothing was fucked up. However, one significant issue was caught and fixed during execution:

### Bug Found & Fixed: asyncSnapshot Context Cancellation

- **Problem:** Initial implementation passed the pipeline's run context to `asyncSnapshot`. `Pipeline.Run()` wraps ctx with `context.WithTimeout` + `defer cancel()`. When Run returns, `cancel()` fires immediately, cancelling the async snapshot goroutine's context before it writes. Result: `TestFlightRecorderHook_SlowLastStageInPipeline` failed 3/3 times.
- **Root cause:** The pipeline's context lifecycle ends before diagnostic goroutines complete.
- **Fix:** `asyncSnapshot` uses `context.Background()` for the actual write. The public `Snapshot(ctx, reason)` API still accepts context for cancellation, but internal async snapshots are decoupled from pipeline lifecycle.
- **Lesson:** When adding context to diagnostic/observability code, verify the context outlives the diagnostic operation. Pipeline contexts are scoped to pipeline execution, not to post-execution diagnostics.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture / Design

1. **FixEdit.MarshalJSON now has inline `json.Deterministic(true)`** instead of using `marshalOpts`. This is because `fix_edit.go` is in the `pipeline` module, which can't import the core module's `marshalOpts`. Consider whether the pipeline module needs its own `marshalOpts` equivalent, or whether a shared internal package would help.
2. **`Degraded()` on FlightRecorderHook is not exported in the StageHook interface** — consumers checking `hook.Degraded()` need a type assertion from `StageHook` to `*FlightRecorderHook`. Consider whether degraded state should be surfaced through the hook interface or via the Logger only.
3. **The `strings.Contains(err.Error(), "flight recorder already enabled")` check is fragile** — it depends on the exact error string from `runtime/trace`. If Go changes the message, degraded mode silently breaks. No better alternative exists without Go stdlib cooperation.
4. **doc.go has stale API references risk** — The AGENTS.md notes that doc.go must be updated after renames. The `Snapshot(ctx, reason)` change was applied, but there may be other references to old APIs in doc.go's ~350 lines that weren't checked this session.

### Testing

5. **No test for the CI script itself** — `json-deterministic-check.sh` was manually verified but has no automated regression test. If someone modifies the script and breaks the awk logic, CI will silently pass everything.
6. **No fuzz test for degraded mode** — The degraded mode path is tested with one integration test but not fuzzed for edge cases (e.g., degraded hook in concurrent pipeline).
7. **`TestFlightRecorderHook_DegradedWhenConflict` depends on test execution order** — it creates a primary hook first, then a degraded one. If run in parallel with other flight recorder tests (which don't use `t.Parallel()`), it could interfere. The singleton nature is documented but fragile.

### Process

8. **The marshalOpts variables trigger `gochecknoglobals`** — suppressed via `.golangci.yml` exclusion for `json.go`. A package-level `var` is the Go-idiomatic way to define reusable options, but it's technically mutable. Could use a function returning a fresh slice, but that allocates per call. The tradeoff is acceptable.
9. **GOWORK=off path safety tests** — Not explicitly verified this session. The exported functions should work identically in GOWORK=off mode, but only the workspace build was tested for this specific change.

---

## F) Up to 50 Things to Get Done Next

### High Impact

1. ~~**Verify doc.go has no other stale API references** — Grep for all function/type names that changed in v1.3-v1.5 era~~ done (doc.go refs updated 2026-09-08)
2. **Add `ResolveSafePath` usage example to docs/guides/** — Now that it's exported, show consumers how to use it
3. ~~**Add `Degraded()` documentation to flight-recorder guide** — How to detect and handle degraded mode~~ done (Degraded() docs in troubleshooting)
4. ~~**Consider a `pipeline.MarshalOpts` equivalent** — So fix_edit.go doesn't inline `json.Deterministic(true)`~~ done (FixEdit uses marshalOpts, v1.6.0)
5. ~~**Add CHANGELOG.md entry for v1.6.0** — These are breaking API changes (Snapshot signature change)~~ done (CHANGELOG 1.6.0)
6. ~~**Version bump** — These changes warrant at least a minor version bump (new exported API + breaking signature change)~~ done (v1.6.0 tagged)
7. ~~**Run `go-arch-lint check`** — Verify the exported path safety functions don't violate boundary rules~~ done (go-arch-lint run locally, OK 2026-09-08)
8. **Run art-dupl** — Verify no new duplication was introduced
9. ~~**Stress test the flight recorder** — Run `-count=20` to verify degraded mode doesn't flake under stress~~ done (stress repeat=20 race passed, mandatory gate 2026-09-08)

### Medium Impact

10. **Add a `ResolveSafePathFrom` bench test** — Verify the caching doesn't regress performance
11. **Test `json-deterministic-check.sh` against edge cases** — Nested ternary, single-line calls, calls with no options at all
12. **Consider a Go-based json-deterministic linter** — The bash script works but a `go/analysis` pass would be more robust
13. **Audit all `context.Background()` in asyncSnapshot callers** — Verify no other diagnostic goroutines share pipeline context
14. **Add `ErrFlightRecorderDegraded` sentinel** — So consumers can `errors.Is` the snapshot failure reason
15. **Document the `marshalOpts` pattern in docs/guides/consumer-migration-v1.3.md** — For consumers writing their own MarshalJSON
16. ~~**Check if `FixEdit.MarshalJSON` determinism matters** — The struct has no maps, so ordering is already deterministic from struct fields. The added `Deterministic(true)` is belt-and-suspenders.~~ done (FixEdit determinism documented, v1.6.0)
17. ~~**Verify `flight_recorder_fuzz_test.go` still passes** — The fuzz test was not run this session~~ done (fuzz green 635K + 4.4M execs)
18. **Add integration test: degraded hook + pipeline run** — Verify a degraded hook doesn't break a real pipeline run
19. **Consider exporting `sanitizeFilename`** — Consumers building custom trace tooling might want it
20. **Review the `writeSnapshot` error messages** — Now includes context cancellation, should have consistent error wrapping

### Lower Priority

21. **Add `ResolveSafePath` to the examples/ directory** — Show the path validation pattern
22. **Document the `asyncSnapshot` context decision in an ADR** — Important architectural decision
23. **Consider a `FlightRecorderHook.DegradedReason() string` method** — More informative than just `bool`
24. **Add a test for `marshalOpts` immutability** — Verify nobody can accidentally mutate the package-level var
25. **Check `config_file.go` ResolveFlightRecorder with degraded mode** — Does it propagate the degraded state?
26. ~~**Audit `.golangci.yml` for other unnecessary exclusions** — Are there other paths that could be un-excluded?~~ done (exclusions audited in-session)
27. **Add `json-deterministic-check.sh` to flake.nix** — So it's available via `nix run`
28. **Consider a pre-commit hook for json-deterministic-check** — Catch violations before push
29. **Review all `slog.Logger.Warn` calls in flight_recorder.go** — Ensure consistent log levels
30. **Add flight recorder degraded mode to CLI output** — User should know their trace isn't being captured
31. **Consider `ResolveSafePathFrom` returning an error instead of bool** — More Go-idiomatic for the "why" case
32. ~~**Document path safety API in docs/api-stability-report.md** — New exported API needs stability commitment~~ done (path safety in API_STABILITY)
33. ~~**Verify the `docs-freshness.sh` change reduces warnings in CI** — Compare before/after warning counts~~ done (freshness warnings resolved 2026-09-08)
34. **Add a test for the paren-depth awk logic** — Edge case: mismatched parens, escaped parens in strings
35. **Consider `Snapshot` accepting `io.Writer` instead of always writing to file** — More flexible
36. **Review `ErrFlightRecorderNotEnabled` semantics** — Now also returned for degraded mode, which is slightly misleading
37. **Add benchmark: degraded mode overhead** — Verify the `degraded` check doesn't slow down the hot path
38. **Consider a `FlightRecorderHook.Status()` method** — Combining `Enabled()`, `Degraded()`, snapshot count
39. ~~**Update FEATURES.md with new exported APIs** — `ResolveSafePath`, `Degraded()`, context-aware `Snapshot`~~ done (FEATURES updated, walk 2026-09-08)
40. ~~**Update ROADMAP.md** — Mark these TODO items as done~~ done (ROADMAP marked)
41. ~~**Check if `TODO_LIST.md` needs updating** — Remove the completed items~~ done (TODO_LIST cleaned)
42. **Review the `_templ.go` exclusion in formatter** — Is it still needed?
43. **Consider CI job naming** — `structural-checks` now has 5 steps, might warrant splitting
44. **Add `--check-deterministic` flag to CLI** — Let users verify their own output determinism
45. ~~**Consider moving `marshalOpts` to a separate `jsonopts.go` file** — Better discoverability~~ **Won't implement — stays in json.go by design.**
46. **Review all exported function doc comments** — Ensure godoc renders correctly for new APIs
47. **Add `lint:check` Makefile target** — Even though Makefile is deprecated, some CI tools expect it
48. **Consider a Go workspace-level lint config** — For tools that support workspace linting
49. ~~**Verify `GOWORK=off go test` passes for pipeline module specifically** — The exports + rename~~ done (GOWORK=off pipeline test, 22-28 session)
50. **Add a session log entry** — Document the asyncSnapshot bug for future reference

---

## G) Questions (Cannot Answer Myself)

1. **Should the `Snapshot` signature change be considered a breaking change requiring v2.0.0?** The project is pre-v2.0.0 (currently v1.5.0), and `Snapshot` is used by the CLI and by consumers registering flight recorder hooks. The old signature `Snapshot(reason string)` is now `Snapshot(ctx context.Context, reason string)`. This is a source-level breaking change for any consumer calling `Snapshot` directly.

2. **Should `ResolveSafePath` / `ResolveSafePathFrom` / `ResolveRoot` be in a separate sub-package (e.g., `pipeline/safepath`)?** They're now in the `pipeline` package, which is a large package. A dedicated package would improve discoverability but adds import complexity for consumers.

3. **Should the `json-deterministic-check.sh` script be replaced with a proper `go/analysis`-based linter?** The bash script works but is inherently fragile (string matching, awk parsing of Go source). A real analyzer would be more robust but requires significantly more effort to build and maintain.

---

_Assisted-by: Crush <crush@charm.land>_
