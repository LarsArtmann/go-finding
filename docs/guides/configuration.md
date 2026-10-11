# Configuration Guide

This guide covers all configuration options for go-finding: CLI flags, YAML/JSON config files, and the library-level `ConfigFile` API.

---

## Quick Reference

| Method                  | Best for                           | Format                               |
| ----------------------- | ---------------------------------- | ------------------------------------ |
| CLI flags               | Quick one-off runs, CI pipelines   | Command-line                         |
| Config file (`-config`) | Reproducible runs, shared settings | YAML or JSON                         |
| Library `ConfigFile`    | Embedded in Go applications        | JSON (via `pipeline.ConfigFromFile`) |

---

## CLI Flags

| Flag                      | Type     | Default | Description                                                                                     |
| ------------------------- | -------- | ------- | ----------------------------------------------------------------------------------------------- |
| `-dir`                    | string   | `.`     | Root directory to analyze                                                                       |
| `-format`                 | string   | `text`  | Output format: `text`, `markdown`, `csv`, `tsv`, `json`, `sarif`                                |
| `-min-severity`           | string   | `info`  | Minimum severity: `info`, `warning`, `error`, `critical`                                        |
| `-severity`               | string   | `info`  | **Deprecated** alias for `-min-severity`                                                        |
| `-max-iterations`         | int      | `1`     | Maximum pipeline iterations                                                                     |
| `-parallel`               | bool     | `true`  | Run detectors in parallel                                                                       |
| `-verify`                 | bool     | `false` | Verify fixes by re-running detectors                                                            |
| `-timeout`                | duration | `10m`   | Pipeline timeout                                                                                |
| `-config`                 | string   | `""`    | YAML/JSON config file path                                                                      |
| `-output`                 | string   | `""`    | Write output to file (default: stdout)                                                          |
| `-cpuprof`                | string   | `""`    | Write CPU profile to file                                                                       |
| `-memprof`                | string   | `""`    | Write memory profile to file                                                                    |
| `-version`                | bool     | `false` | Print version and exit                                                                          |
| `-filter-generated`       | bool     | `false` | Filter out findings from auto-generated files                                                   |
| `-filter-generated-types` | string   | `all`   | Generator types to filter (comma-separated: `all`, `sqlc`, `templ`, `mockgen`, `protobuf`, ...) |
| `-generated-exclude`      | string   | `""`    | Comma-separated glob patterns to exclude from generated filtering                               |
| `-generated-include`      | string   | `""`    | Comma-separated glob patterns restricting generated-filtering scope                             |
| `-byte-level-conflict`    | bool     | `false` | Enable precise byte-level conflict detection for overlapping fixes                              |
| `-fix-provider`           | string   | `""`    | Comma-separated fix provider names to enable (e.g., `go-ast`)                                   |
| `-fix-rollback-all`       | bool     | `false` | Roll back ALL files when any file fails during fix (default: only the failing file is restored) |
| `-include-suppressed`     | bool     | `true`  | Include suppressed findings in SARIF output                                                     |
| `-trace`                  | bool     | `false` | Enable Go execution trace flight recorder                                                       |
| `-trace-dir`              | string   | `""`    | Directory for trace snapshot files (default: temp dir)                                          |
| `-trace-slow`             | duration | `0`     | Auto-snapshot trace when a pipeline stage exceeds this duration                                 |
| `-trace-max-files`        | int      | `0`     | Keep at most N trace snapshots, pruning oldest (0 = unlimited)                                  |
| `-trace-gzip`             | bool     | `false` | Write gzip-compressed `.trace.gz` snapshots                                                     |

---

## Config File

Use a config file for reproducible runs. The file is loaded with `-config`:

```bash
go-finding -config pipeline.yaml ./...
go-finding -config pipeline.json ./...
```

### Full YAML Example

```yaml
maxIterations: 3
parallelDetectors: true
verifyAfterFix: true
timeout: "10m"
detectorTimeouts:
  govet: "30s"
  staticcheck: "2m"
detectors:
  - name: govet
  - name: staticcheck
filterGenerated: true
filterGenTypes: "sqlc,templ"
generatedExclude:
  - "**/mock_*.go"
generatedInclude:
  - "internal/**"
byteLevelConflictDetection: true
fixProviders:
  - go-ast
flightRecorder:
  enabled: true
  outputDir: "./traces"
  slowStageThreshold: "30s"
  minAge: "1m"
  maxBytes: 4194304 # 4 MiB
  maxFiles: 20
  compress: true
```

### Full JSON Example

```json
{
  "maxIterations": 3,
  "parallelDetectors": true,
  "verifyAfterFix": true,
  "timeout": "10m",
  "detectorTimeouts": {
    "govet": "30s",
    "staticcheck": "2m"
  },
  "detectors": [{ "name": "govet" }, { "name": "staticcheck" }],
  "filterGenerated": true,
  "filterGenTypes": "sqlc,templ",
  "generatedExclude": ["**/mock_*.go"],
  "generatedInclude": ["internal/**"],
  "byteLevelConflictDetection": true,
  "fixProviders": ["go-ast"],
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

### Config File Fields

| Key                          | Type              | Default                    | Description                                                                                                                 |
| ---------------------------- | ----------------- | -------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `maxIterations`              | int               | `0` (inherits default 5)   | Max pipeline iterations. Must be >= 0.                                                                                      |
| `parallelDetectors`          | bool              | `false`                    | Run detectors concurrently                                                                                                  |
| `verifyAfterFix`             | bool              | `false`                    | Re-run detectors after fixes to verify                                                                                      |
| `timeout`                    | string            | `"10m"`                    | Pipeline timeout as a duration string                                                                                       |
| `detectorTimeouts`           | map[string]string | —                          | Per-detector timeouts (duration strings)                                                                                    |
| `detectors`                  | []detectorSpec    | `[{govet}, {staticcheck}]` | Detectors to run. Each entry has a `name` field.                                                                            |
| `filterGenerated`            | bool              | `false`                    | Filter findings from auto-generated Go source files                                                                         |
| `filterGenTypes`             | string            | `"all"`                    | Generator types to filter (comma-separated)                                                                                 |
| `generatedExclude`           | []string          | —                          | Glob patterns for files to exclude from generated filtering                                                                 |
| `generatedInclude`           | []string          | —                          | Glob patterns restricting generated-filtering scope                                                                         |
| `byteLevelConflictDetection` | bool              | `false`                    | Enable precise byte-level conflict detection during triage                                                                  |
| `fixProviders`               | []string          | —                          | Named fix providers to enable (e.g., `go-ast`)                                                                              |
| `fixRollbackAllFiles`        | bool              | `false`                    | All-or-nothing fix rollback: on file failure, restore all files modified in the run. Default restores only the failing file |
| `flightRecorder`             | object            | —                          | Flight recorder settings (see below)                                                                                        |

### Flight Recorder Fields

| Key                  | Type   | Default           | Description                                                 |
| -------------------- | ------ | ----------------- | ----------------------------------------------------------- |
| `enabled`            | bool   | `false`           | Must be `true` to activate                                  |
| `outputDir`          | string | `os.TempDir()`    | Directory for `.trace` snapshot files                       |
| `slowStageThreshold` | string | `""` (disabled)   | Duration string (e.g., `"30s"`, `"2m"`) for auto-snapshot   |
| `minAge`             | string | `"30s"`           | How long trace data is reliably retained in the ring buffer |
| `maxBytes`           | uint64 | `4194304` (4 MiB) | Maximum in-memory buffer size                               |
| `maxFiles`           | int    | `0` (unlimited)   | Cap on retained snapshot files (oldest pruned)              |
| `compress`           | bool   | `false`           | gzip-compress snapshots (`.trace.gz`)                       |

See [FlightRecorder Guide](flight-recorder.md) for a complete walkthrough.

---

## CLI vs Config File Precedence

CLI flags and config file values are **merged**, not mutually exclusive:

| Setting               | Merge behavior                                                                                            |
| --------------------- | --------------------------------------------------------------------------------------------------------- |
| Fix providers         | Config file `fixProviders` **and** CLI `-fix-provider` are merged and deduplicated                        |
| Byte-level conflict   | Enabled if **either** CLI flag **or** config key is set                                                   |
| Generated filtering   | Enabled if **either** CLI flag **or** config key is set; CLI types/exclude/include override config values |
| `gracefulDegradation` | Always `true` at the CLI level (hardcoded), regardless of config                                          |
| Flight recorder       | CLI `-trace` flag takes precedence when both are set                                                      |

---

## Library ConfigFile (Programmatic API)

For applications embedding go-finding directly, use `pipeline.ConfigFromFile` or `pipeline.ConfigFromReader` to load JSON config directly into a `pipeline.Config`:

```go
import "github.com/larsartmann/go-finding/pipeline"

data, err := os.ReadFile("config.json")
if err != nil {
    return err
}

// ConfigFromFile parses JSON, validates duration strings, and returns
// a pipeline.Config ready to pass to pipeline.New().
config, err := pipeline.ConfigFromFile(data)
if err != nil {
    return err
}

// Add detectors, fix providers, stage hooks, etc.
config.Detectors = detectors
config.FixProviders = providers

p, err := pipeline.New(config)
```

The library `ConfigFile` has additional fields not available in the CLI config:

| Field                 | Type     | Description                                                              |
| --------------------- | -------- | ------------------------------------------------------------------------ |
| `gracefulDegradation` | bool     | Continue on detector failures, collecting partial results                |
| `dryRun`              | bool     | Run detect+triage but skip fix application                               |
| `correlateFindings`   | bool     | Run cross-tool correlation on findings                                   |
| `severity`            | string   | Minimum severity filter (`"info"`, `"warning"`, `"error"`, `"critical"`) |
| `detectorNames`       | []string | Detector names to resolve via `DetectorRegistry`                         |
| `providerNames`       | []string | Fix provider names to resolve                                            |

---

## Runtime Config Defaults

The `pipeline.Config` struct uses these defaults when fields are zero-valued:

| Setting               | Default | Source                          |
| --------------------- | ------- | ------------------------------- |
| `MaxIterations`       | 5       | `pipeline.DefaultMaxIterations` |
| `Timeout`             | 10m     | `pipeline.DefaultTimeout`       |
| `ParallelDetectors`   | true    | `pipeline.DefaultConfig()`      |
| `GracefulDegradation` | true    | CLI hardcodes this              |
