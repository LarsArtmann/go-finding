# Troubleshooting Guide

Common errors, their causes, and fixes.

---

## Build and Setup

### "build constraints exclude all Go files"

**Cause:** The project uses `encoding/json/v2` (experimental in Go 1.26). Without `GOEXPERIMENT=jsonv2`, the compiler excludes all files that import it.

**Fix:**

```bash
go env -w GOEXPERIMENT=jsonv2
```

Or for a single command:

```bash
GOEXPERIMENT=jsonv2 go build ./...
```

If using Nix, the devShell sets this automatically:

```bash
nix develop
```

> When Go stabilizes json/v2 (expected 1.27+), this step disappears.

### Per-module builds fail with "cannot find module"

**Cause:** Each sub-module has `replace` directives for `GOWORK=off` builds. Running from the wrong directory without the workspace can fail.

**Fix:** Always set `GOWORK=off` when building a single module:

```bash
cd pipeline && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...
```

### "could not read Username" or "410 Gone" from proxy.golang.org

**Cause:** The repo is private. The public Go proxy cannot resolve it.

**Fix:**

```bash
go env -w GOPRIVATE=github.com/larsartmann/go-finding
```

---

## Config File Errors

### "parsing YAML config: ..." or "parsing config: ..."

**Cause:** Malformed YAML or JSON in the config file.

**Fix:** Validate the syntax. Common issues:

- YAML: wrong indentation, missing quotes around duration strings (`timeout: 10m` should be `timeout: "10m"`)
- JSON: trailing commas, unquoted keys

See [Configuration Guide](configuration.md) for valid examples.

### "invalid timeout" or "parse timeout"

**Cause:** Duration string is not parseable by `time.ParseDuration`.

**Fix:** Use Go duration format: `"30s"`, `"5m"`, `"1h"`, `"1m30s"`. Plain numbers (`30`, `5`) are invalid.

### "unknown detector" / "unknown provider"

**Cause:** Config file references a detector or fix provider name that is not registered.

**Fix:** Check available names. Built-in detectors: `govet`, `staticcheck`. Fix providers: `go-ast`. The error message lists all available names.

### "invalid flightRecorder.slowStageThreshold" / "invalid flightRecorder.minAge"

**Cause:** Duration string in the `flightRecorder` config section is unparseable.

**Fix:** Use Go duration format (e.g., `"30s"`, `"2m"`).

### "unsupported format" (output)

**Cause:** Unknown value for `-format` flag.

**Fix:** Use one of: `text`, `markdown`, `csv`, `tsv`, `json`, `sarif`.

---

## Pipeline Runtime Errors

### "pipeline: Run already called; create a new Pipeline for each invocation"

**Cause:** `Pipeline.Run()` was called twice on the same instance. The pipeline is single-use.

**Fix:** Create a new `Pipeline` for each run:

```go
p, _ := pipeline.New(cfg, rootDir, detector)
result, _ := p.Run(ctx)

// For another run, create a new pipeline:
p2, _ := pipeline.New(cfg, rootDir, detector)
result2, _ := p2.Run(ctx)
```

### "pipeline cancelled"

**Cause:** The context passed to `Pipeline.Run()` was cancelled or timed out.

**Fix:** Increase the timeout (`-timeout` flag or `Config.Timeout`), or ensure the parent context isn't being cancelled prematurely.

### "pipeline: partial detection failures"

**Cause:** One or more detectors failed during graceful degradation mode. The pipeline continued with partial results.

**Fix:** Check the formatted error for per-detector details. Common causes:

- External tool not installed (e.g., `staticcheck` binary missing)
- Detector timeout exceeded (increase `detectorTimeouts`)
- Invalid analysis input (corrupt Go files)

### "max iterations must be >= 0" / "timeout must be >= 0"

**Cause:** `Config.Validate()` failed on negative values.

**Fix:** Set `MaxIterations >= 0` and `Timeout >= 0`. Use `0` for `MaxIterations` to inherit the default (5).

### Retry validation errors ("base delay must be > 0 when max retries > 0", etc.)

**Cause:** `RetryConfig` has inconsistent settings.

**Fix:** If `MaxRetries > 0`, then `BaseDelay` must be > 0. If `BaseDelay > 0`, then `MaxDelay` must be > 0 and >= `BaseDelay`. All values must be non-negative.

---

## Fix Application Errors

### "conflict: apply to <file>: ..."

**Cause:** Overlapping edits target the same byte range in a file.

**Fix:** Enable byte-level conflict detection (`-byte-level-conflict` or `byteLevelConflictDetection: true` in config) to get precise conflict reports. Review conflicting findings and suppress or fix them in priority order.

### "position unresolvable in content: line beyond end of file"

**Cause:** A finding's position references a line number that exceeds the file's actual line count. This can happen if the file was modified between detection and fix application.

**Fix:** Ensure the source files haven't changed between detection and fix application. If using a custom detector, verify the line numbers it produces.

### "position unresolvable in content: column beyond end of line"

**Cause:** A finding's column number exceeds the line width.

**Fix:** Verify column numbers in custom detectors. Column 0 or 1 is always safe; large columns will fail on short lines.

### Fix failed AND rollback also failed

**Cause:** A fix application failed and the automatic rollback also encountered errors. The file may be in an inconsistent state.

**Fix:** The backup directory (from `FixApplier`) contains original file copies. Restore manually from the backup directory shown in the error message. Investigate the root cause of both failures.

---

## Flight Recorder

### "flight recorder: not enabled or already closed"

**Cause:** `Snapshot()` was called on a disabled or already-closed `FlightRecorderHook`.

**Fix:** Check `hook.Enabled()` before calling `Snapshot()`. Ensure `Close()` hasn't been called yet.

### No .trace files produced despite flight recorder enabled

**Cause:** `SlowStageThreshold` is not set (or set too high), and no manual `Snapshot()` call was made.

**Fix:** Either set `slowStageThreshold` to a low value (e.g., `"1s"`) or call `hook.Snapshot(ctx, "reason")` explicitly after the pipeline run.

### Only one flight recorder can be active at a time

**Cause:** Go's `runtime/trace.FlightRecorder` is a global singleton. Creating a second one while the first is active causes `Start()` to fail.

**Fix:** Since v1.6.0, `NewFlightRecorderHook` handles this gracefully — if another recorder is active, it returns a degraded hook (`Degraded()` returns true) that silently skips all snapshots. Check `hook.Degraded()` if you need to detect this. Use `defer hook.Close()` before creating another.

---

## Finding Validation Errors

These errors come from `Finding.Validate()`. The `Builder` API catches these at construction time.

### "finding.ID is required" / "finding.Rule is required" / etc.

**Cause:** A required field on the `Finding` struct is empty.

**Fix:** Use the `Builder` API which validates at construction time:

```go
f, err := finding.NewBuilder(
    finding.RuleName("rule-name"),
    finding.ToolName("tool-name"),
    "descriptive message",
    finding.SeverityWarning,
    finding.Pos("file.go", 10, 5),
).Build()
```

### "finding.Severity is invalid"

**Cause:** Severity value is not one of: `SeverityInfo`, `SeverityWarning`, `SeverityError`, `SeverityCritical`.

**Fix:** Use the named constants. For parsing severity from strings, use `finding.SeverityFromLevel(level, fallback)`.

### "finding.Confidence must be in [0.0, 1.0]"

**Cause:** Confidence value is outside the valid range.

**Fix:** Use `finding.Confidence.Clamp()` to clamp to [0.0, 1.0], or use named levels: `ConfidenceNone` (0), `ConfidenceLow` (0.3), `ConfidenceMedium` (0.5), `ConfidenceHigh` (0.7), `ConfidenceFull` (1.0).

---

## Output Errors

### "creating output file: ..." (permission denied)

**Cause:** The `-output` flag points to a directory or path without write permission.

**Fix:** Verify the output path is writable and the parent directory exists.

### SARIF output is missing suppressed findings

**Cause:** Suppressed findings are excluded by default.

**Fix:** Set `-include-suppressed=true` (default) or use `WithIncludeSuppressed()` SARIF option.

---

## Getting Help

- Check the [Configuration Guide](configuration.md) for config file format
- Check the [Fix Engine Guide](fix-engine.md) for fix application details
- Check the [FlightRecorder Guide](flight-recorder.md) for trace recording
- Review [CHANGELOG.md](../../CHANGELOG.md) for recent changes
