# Flight Recorder Guide

The Flight Recorder captures Go execution traces during pipeline runs, giving you deep visibility into where the pipeline spends time. It wraps Go 1.25's `runtime/trace.FlightRecorder`, which continuously buffers trace data in a ring buffer so you can capture the most recent window on demand.

## Table of Contents

- [Quick Start: CLI Flags](#quick-start-cli-flags)
- [Config File Integration](#config-file-integration)
- [Programmatic API](#programmatic-api)
- [Analyzing Traces with go tool trace](#analyzing-traces-with-go-tool-trace)
- [Configuration Reference](#configuration-reference)
- [How It Works](#how-it-works)

---

## Quick Start: CLI Flags

Enable tracing from the command line:

```bash
go-finding -trace ./...
```

This starts the flight recorder and registers it as a pipeline `StageHook`. When the pipeline finishes (or errors), a trace snapshot is written to your OS temp directory.

### Controlling Output Location

```bash
go-finding -trace -trace-dir ./traces ./...
```

### Auto-Snapshot on Slow Stages

```bash
go-finding -trace -trace-slow 30s ./...
```

When any pipeline stage exceeds 30 seconds, a snapshot is captured automatically with a filename like `go-finding-trace-001-detect-iter1.trace`.

### Rotation: Cap Retained Snapshots

```bash
go-finding -trace -trace-slow 10s -trace-max-files 20 ./...
```

After each successful snapshot, the oldest `go-finding-trace-*` files beyond the cap are deleted (best-effort, only the hook's own files are ever touched). Default: unlimited. Set this for long `-trace-slow` runs to avoid unbounded disk growth.

### Compressed Snapshots

```bash
go-finding -trace -trace-gzip ./...
```

Snapshots are written as `.trace.gz` files (often 5-10x smaller). `go tool trace` does not read gzip directly — gunzip first (see [Analyzing Traces](#analyzing-traces-with-go-tool-trace)).

### Full Example

```bash
go-finding -trace -trace-dir ./traces -trace-slow 10s -trace-max-files 20 -trace-gzip ./...
```

Output:

```
Flight recorder enabled (output: ./traces, slow threshold: 10s)
go-finding v1.5.0: analyzing . with 2 detector(s)
```

---

## Config File Integration

For YAML or JSON configuration users, flight recording can be enabled without CLI flags:

### YAML

```yaml
flightRecorder:
  enabled: true
  outputDir: "./traces"
  slowStageThreshold: "30s"
  minAge: "1m"
  maxBytes: 4194304 # 4 MiB
  maxFiles: 20 # prune oldest snapshots beyond this cap (0 = unlimited)
  compress: true # .trace.gz snapshots (gunzip before `go tool trace`)
```

### JSON

```json
{
  "flightRecorder": {
    "enabled": true,
    "outputDir": "./traces",
    "slowStageThreshold": "30s",
    "minAge": "1m",
    "maxBytes": 4194304,
    "maxFiles": 20,
    "compress": true
  }
}
```

```bash
go-finding -config pipeline.yaml ./...
```

Output:

```
Flight recorder enabled via config (output: ./traces, slow threshold: 30s)
```

The `-trace` CLI flag takes precedence when both are set.

### Pipeline Package (Library API)

If you use the pipeline package directly, construct the hook via `ConfigFile.ResolveFlightRecorder`:

```go
config, err := pipeline.ConfigFromFile(data)
if err != nil { return err }

hook, err := config.ResolveFlightRecorder()
if err != nil { return err }

if hook != nil {
    cfg.StageHooks = append(cfg.StageHooks, hook)
    defer hook.Close()
}
```

---

## Programmatic API

For full control, use `pipeline.NewFlightRecorderHook` directly:

```go
import "github.com/larsartmann/go-finding/pipeline"

hook, err := pipeline.NewFlightRecorderHook(pipeline.FlightRecorderConfig{
    OutputDir:          "./traces",
    SlowStageThreshold: 10 * time.Second,
    MaxFiles:           20, // prune oldest snapshots beyond 20 (0 = unlimited)
    Compress:           true, // write .trace.gz snapshots
})
if err != nil { return err }
defer hook.Close()

cfg := pipeline.Config{
    MaxIterations: 3,
    StageHooks:    []pipeline.StageHook{hook},
}

p, err := pipeline.New(cfg, rootDir, detector)
if err != nil { return err }

result, err := p.Run(ctx)

// Capture a manual snapshot after the run.
if path, snapErr := hook.Snapshot("manual"); snapErr == nil {
    fmt.Println("Trace written to:", path)
}
```

### Manual Snapshots

Call `Snapshot(ctx, reason)` at any point to capture the current trace buffer:

```go
path, err := hook.Snapshot(ctx, "custom-checkpoint")
// path: /tmp/go-finding-trace-003-custom-checkpoint.trace
```

The `reason` string is sanitized for use in the filename (non-alphanumeric characters become hyphens).

### Checking Status

```go
if hook.Enabled() {
    // recorder is active
}
```

---

## Analyzing Traces with go tool trace

Trace files are standard Go execution traces. Open them with the built-in tool:

```bash
go tool trace traces/go-finding-trace-001-detect-iter1.trace
```

### Compressed Snapshots (.trace.gz)

`go tool trace` cannot read gzip directly. Decompress first:

```bash
gunzip -k traces/go-finding-trace-001-detect-iter1.trace.gz
go tool trace traces/go-finding-trace-001-detect-iter1.trace
```

The `-k` flag keeps the compressed original. The decompressed file is byte-for-byte the trace the recorder produced.

This opens a web browser with interactive views including:

- **View trace** — Timeline of goroutines, network, and syscall blocks
- **Goroutine analysis** — Per-goroutine time breakdown
- **Network blocking profile** — Time spent waiting on network
- **Scheduling latency profile** — Scheduler delays
- **User-defined tasks** — Pipeline stages annotated as tasks

### What to Look For

| Symptom         | Trace View                         |
| --------------- | ---------------------------------- |
| Pipeline hangs  | Goroutine analysis → blocking time |
| Slow detection  | View trace → detector goroutines   |
| GC pressure     | Heap view under trace timeline     |
| Lock contention | View trace → goroutine wait states |

---

## Configuration Reference

| Field                | Type            | Default           | Description                                                 |
| -------------------- | --------------- | ----------------- | ----------------------------------------------------------- |
| `OutputDir`          | `string`        | `os.TempDir()`    | Directory for `.trace` snapshot files                       |
| `SlowStageThreshold` | `time.Duration` | `0` (disabled)    | Auto-snapshot when a stage exceeds this duration            |
| `MinAge`             | `time.Duration` | `30s`             | How long trace data is reliably retained in the ring buffer |
| `MaxBytes`           | `uint64`        | `4 MiB` (`4<<20`) | Maximum in-memory buffer size                               |
| `MaxFiles`           | `int`           | `0` (unlimited)   | Cap on retained snapshots; oldest pruned after each write   |
| `Compress`           | `bool`          | `false`           | gzip snapshots (`.trace.gz`); gunzip before `go tool trace` |
| `Logger`             | `*slog.Logger`  | `nil`             | Receives snapshot lifecycle events                          |

### Config File Fields

| YAML Key             | Type     | Description                                                     |
| -------------------- | -------- | --------------------------------------------------------------- |
| `enabled`            | `bool`   | Must be `true` to activate                                      |
| `outputDir`          | `string` | Override the output directory                                   |
| `slowStageThreshold` | `string` | Duration string (e.g. `"30s"`, `"2m"`)                          |
| `minAge`             | `string` | Duration string for ring buffer retention                       |
| `maxBytes`           | `uint64` | Buffer size in bytes                                            |
| `maxFiles`           | `int`    | Retained-snapshot cap (prune oldest beyond it; `0` = unlimited) |
| `compress`           | `bool`   | gzip snapshot output (`.trace.gz`)                              |

---

## How It Works

The flight recorder wraps `runtime/trace.FlightRecorder`, which continuously records execution trace data into a ring buffer. Unlike start/stop tracing, the flight recorder always runs but only writes to disk when you explicitly request a snapshot.

### Lifecycle

1. **Construction** — `NewFlightRecorderHook` creates the output directory, initializes the `trace.FlightRecorder`, and calls `Start()`.
2. **Pipeline execution** — Registered as a `StageHook`, it receives `StageBefore` and `StageAfter` events. If `SlowStageThreshold` is set, stages exceeding it trigger an async snapshot.
3. **Snapshots** — `Snapshot(ctx, reason)` or `writeSnapshot(ctx, num, reason)` calls `fr.WriteTo(file)` to dump the ring buffer to a `.trace` (or `.trace.gz`) file. A `sync.Mutex` (`writeMu`) serializes concurrent writes and the post-write rotation pass. The context is checked for cancellation before writing.
4. **Rotation** — when `MaxFiles` is set, each successful snapshot prunes the oldest `go-finding-trace-*` files beyond the cap. Pruning holds `writeMu`, so concurrent snapshots cannot list/delete overlapping sets.
5. **Shutdown** — `Close()` is idempotent. It waits for in-flight snapshot goroutines to finish, then calls `fr.Stop()`.

### Thread Safety

- Only one flight recorder can be active globally (Go runtime constraint).
- `WriteTo` is not safe for concurrent use — `writeMu` serializes calls, including the rotation pass.
- `Close()` waits for in-flight snapshots before stopping to avoid a data race between `WriteTo` and `Stop`.

### Error Handling

The `OnStageEvent` method **never returns errors**. The flight recorder is diagnostic-only and must never abort the pipeline. If a snapshot fails (e.g., disk full), the error is logged to the configured `slog.Logger` but the pipeline continues.
