# FEATURES.md — go-finding

> **Version:** 0.2.1 | **Updated:** 2026-05-02
>
> A unified data model and pipeline for Go static analysis tools.
> Seven tools detect issues. Zero tools route them to remediation. This library fixes that.

---

## Status Legend

| Status          | Meaning                                                     |
| --------------- | ----------------------------------------------------------- |
| STABLE          | Fully implemented, tested, production-ready                 |
| FUNCTIONAL      | Works but has known limitations or rough edges              |
| EXPERIMENTAL    | Implemented and tested, but API may change                  |
| RESERVED        | Type/constant exists, tested, but no backend implementation |
| NOT_IMPLEMENTED | Mentioned in docs/comments but no code exists               |

---

## 1. Core Data Model

### 1.1 Finding Type

**Status:** STABLE

The central type representing a single issue detected by a static analysis tool.

| Field       | Type                | Purpose                                              |
| ----------- | ------------------- | ---------------------------------------------------- |
| ID          | `string`            | Stable unique identifier (`tool:rule:file:line:col`) |
| Rule        | `string`            | Rule/check name (e.g., `STRONG_ID`)                  |
| ToolName    | `string`            | Source tool name (e.g., `govet`)                     |
| Message     | `string`            | Human-readable description                           |
| Severity    | `Severity`          | info / warning / error / critical                    |
| Position    | `Position`          | Where the issue is (file, line, column, offset)      |
| Category    | `Category`          | Domain classification (security, style, etc.)        |
| Tag         | `string`            | **Deprecated** — use `Tags` instead                  |
| Tags        | `[]Tag`             | Multiple classification labels                       |
| FixStrategy | `FixStrategy`       | none / suggest / direct / ai                         |
| Suggestion  | `string`            | Human-readable fix description                       |
| BeforeCode  | `string`            | Code before the fix                                  |
| AfterCode   | `string`            | Code after the fix                                   |
| Range       | `*Range`            | Span-based findings (start/end positions)            |
| Snippet     | `string`            | Surrounding code context                             |
| Confidence  | `float64`           | 0.0–1.0                                              |
| Related     | `[]RelatedRef`      | Related findings (clone-of, wraps, causes)           |
| Suppression | `*Suppression`      | If suppressed                                        |
| Metadata    | `map[string]string` | Tool-specific key-value pairs                        |

Key methods: `Validate()`, `IsValid()`, `Clone()`, `Key()`, `Equal()`, `String()`, `Preview()`, `HasFix()`, `HasSuggestion()`, `IsSuppressed()`, `NormalizedConfidence()`

### 1.2 Builder API

**Status:** STABLE

Fluent builder for constructing `Finding` values with validation.

```go
f, err := NewBuilder("nilcheck", "govet", "possible nil deref", SeverityError, Pos("main.go", 42, 5)).
    WithFixStrategy(FixStrategyDirect).
    WithBeforeCode("x.foo").
    WithAfterCode("x.foo()").
    Build()
```

- `NewBuilder(rule, toolName, message, severity, pos)` — required fields
- `WithID()`, `WithCategory()`, `WithTags()`, `WithFixStrategy()`, `WithSuggestion()`, `WithBeforeCode()`, `WithAfterCode()`, `WithRange()`, `WithSnippet()`, `WithConfidence()`, `WithRelated()`, `WithSuppression()`, `WithMetadata()` — optional
- `Build()` — returns validated `Finding` or error
- `MustBuild()` — panics on invalid state

---

## 2. Position & Range

**Status:** STABLE

### 2.1 Position

| Field  | Type     | Notes                             |
| ------ | -------- | --------------------------------- |
| File   | `string` | Required                          |
| Line   | `int`    | 1-based; 0 = not set              |
| Column | `int`    | 1-based; 0 = not set              |
| Offset | `int`    | 0-based byte offset; -1 = not set |

Methods: `IsValid()`, `Equal()`, `Compare()`, `String()`, `HasOffset()`
Constructor: `Pos(file, line, column)`

### 2.2 Range

Span from `Start` to `End` position.

Methods: `IsValid()`, `HasEnd()`, `LineCount()`, `Length()`, `Equal()`, `Compare()`, `Contains(Position)`, `Overlaps(Range)`, `Intersection(Range)`, `Adjacent(Range)`

Constructors: `NewRange(file, startLine, startCol, endLine, endCol)`, `NewRangePtr(...)`

---

## 3. Severity

**Status:** STABLE

Four levels, ordered by urgency:

| Level    | String       | SARIF Mapping                             |
| -------- | ------------ | ----------------------------------------- |
| Info     | `"info"`     | `note`                                    |
| Warning  | `"warning"`  | `warning`                                 |
| Error    | `"error"`    | `error`                                   |
| Critical | `"critical"` | `error` (lossy — preserved in Properties) |

Methods: `IsValid()`, `Compare()`, `GreaterThan()`, `LessThan()`, `GreaterThanOrEqual()`, `LessThanOrEqual()`

---

## 4. Fix Strategy

**Status:** STABLE (except AI — see below)

| Strategy | String      | Auto-apply | Description                     |
| -------- | ----------- | ---------- | ------------------------------- |
| None     | `"none"`    | No         | No fix available                |
| Suggest  | `"suggest"` | No         | Human-readable suggestion       |
| Direct   | `"direct"`  | Yes        | Automatically applicable        |
| AI       | `"ai"`      | No         | **RESERVED** — needs AI backend |

Methods: `IsValid()`, `CanAutoApply()`, `NeedsAI()`

> **Honest assessment:** `FixStrategyAI` is a placeholder. `NeedsAI()` returns `true` for it, but no AI backend exists. The pipeline triage groups `ai` with `suggest` (no auto-apply). Safe to use as a marker for future AI integration.

---

## 5. Category & Tags

### 5.1 Category

**Status:** STABLE

14 predefined domain categories:

`security`, `style`, `performance`, `correctness`, `complexity`, `duplication`, `error-handling`, `migration`, `type-safety`, `structure`, `configuration`, `documentation`, `testing`, `unused`

Plus arbitrary custom categories accepted. Methods: `IsStandard()`, `IsValid()`, `String()`

### 5.2 Tags

**Status:** STABLE

Multi-label classification for richer filtering:

`security`, `performance`, `style`, `correctness`, `bug`, `deprecated`, `documentation`, `complexity`, `test`, `build`

Methods: `IsStandard()`, `IsValid()`, `String()`

Plus arbitrary custom tags accepted. The `Tag` field (singular) is deprecated in favor of `Tags` (plural).

---

## 6. Suppression

**Status:** STABLE

Mark findings as suppressed with reason and optional expiry.

| Field     | Type              | Purpose                               |
| --------- | ----------------- | ------------------------------------- |
| Kind      | `SuppressionKind` | `in-source`, `in-config`, `in-review` |
| Rule      | `string`          | Which rule is suppressed              |
| Reason    | `string`          | Why                                   |
| ExpiresAt | `*time.Time`      | Optional TTL                          |

Methods: `IsExpired(now)`, `IsValid()`, `IsActive(now)`, `SuppressionKind.IsValid()`

> `IsActive(now)` is a convenience combining `IsValid() && !IsExpired(now)`.

> **Note:** The library stores and checks suppression data. It does NOT parse `//nolint` or `//lint:ignore` directives — that is the caller's responsibility.

---

## 7. Report Container

**Status:** STABLE

Top-level container for a tool run.

| Feature                 | Method                              | Thread-safe |
| ----------------------- | ----------------------------------- | ----------- |
| Create                  | `NewReport(toolInfo)`               | Yes         |
| Add single finding      | `AddFinding(f)`                     | Yes (mutex) |
| Add multiple findings   | `AddFindings([])`                   | Yes (mutex) |
| Recompute stats         | `ComputeSummary()`                  | No          |
| Filter by severity      | `BySeverity(sev)`                   | Read-only   |
| Filter by category      | `ByCategory(cat)`                   | Read-only   |
| Filter by fix strategy  | `ByFixStrategy(fs)`                 | Read-only   |
| Find by ID              | `FindByID(id)` → `*Finding` (copy)  | Read-only   |
| Find by rule            | `FindByRule(rule)`                  | Read-only   |
| Active (non-suppressed) | `ActiveFindings()`                  | Read-only   |
| Generic filter          | `Filter(predicates...)` → `*Report` | Read-only   |
| Transform               | `Map(func) → *Report`               | Read-only   |
| Iterate                 | `All()` → `iter.Seq[Finding]`       | Read-only   |
| Count                   | `Len()`                             | Read-only   |

Summary stats: `Total`, `BySeverity`, `ByCategory`, `ByFixStrategy`, `FilesAffected`, `DurationMs`, `Suppressed`

---

## 8. Filtering & Sorting

**Status:** STABLE

### 8.1 Predicates

Composable filter functions:

`BySeverity`, `BySeverityAtLeast`, `ByCategory`, `ByFixStrategy`, `ByTool`, `ByRule`, `ByFile`, `NotSuppressed`, `HasFix`, `HasSuggestion`

### 8.2 Core Operations

| Operation       | Function                                 |
| --------------- | ---------------------------------------- |
| Filter          | `Filter(findings, predicates...)`        |
| In-place filter | `FilterInPlace(findings, predicates...)` |

### 8.3 Grouping

`GroupBy(findings, keyFn)`, `GroupByFile`, `GroupBySeverity`, `GroupByCategory`

### 8.4 Sorting

`SortByPosition(findings)` — file, line, column
`SortBySeverity(findings)` — most severe first

---

## 9. Report Merging & Deduplication

**Status:** STABLE

### 9.1 Merge

Combines multiple reports into one with optional deduplication.

```go
merged := finding.Merge(reports,
    finding.WithDeduplication(true),
    finding.WithDeduplicateBy(finding.DeduplicateByID),
)
```

### 9.2 Deduplication Strategies

| Strategy                | Match By                    |
| ----------------------- | --------------------------- |
| `DeduplicateByID`       | Exact ID match (default)    |
| `DeduplicateByPosition` | tool + file + line + column |
| `DeduplicateByRule`     | rule + file + line + column |

### 9.3 Cross-Tool Correlation

**Status:** FUNCTIONAL

Finds related findings across tools using heuristics (same file, nearby lines within 5 lines).

```go
correlations := finding.Correlate(allFindings)
```

Returns `[]Correlation` with `FindingIDs`, `Reason`, and `Confidence` score. Capped at 10,000 correlations to prevent O(n²) hangs.

> **Limitation:** Simple heuristic only — no semantic analysis. Good for surface-level grouping.

---

## 10. ID Generation & Parsing

**Status:** STABLE

### Generation

`GenerateID(toolName, rule, pos)` produces:

- `"tool:rule:file:line:col"` when line > 0 (human-readable)
- `"tool:rule:<sha256hash>"` when line == 0 (hash-based)

### Parsing

`ParseID(id)` → `ParsedID{Tool, Rule, File, Line, Column}`

`IsHashID(id)` → checks if hash-based

Handles Windows paths with colons correctly.

---

## 11. JSON Serialization

**Status:** STABLE

| Operation        | Function                                                 | Notes                        |
| ---------------- | -------------------------------------------------------- | ---------------------------- |
| Report → JSON    | `report.PrettyJSON()`                                    | Pretty-printed               |
| Report → Writer  | `report.WriteJSON(w)`                                    | Streaming, avoids allocation |
| JSON → Report    | `ReportFromJSON(data)` → `(*Report, dropped, error)`     | Drops invalid findings       |
| Finding → JSON   | `finding.LineJSON()`                                     | Compact single-line          |
| Finding → Writer | `finding.WriteJSON(w)`                                   | Streaming                    |
| JSON → Finding   | `FromJSON(data)`                                         | Validates required fields    |
| JSON → []Finding | `FindingsFromJSON(data)` → `([]Finding, dropped, error)` | Drops invalid                |

---

## 12. SARIF 2.1.0 Interchange

**Status:** STABLE

### Export

| Method                                 | Description                                    |
| -------------------------------------- | ---------------------------------------------- |
| `report.ToSARIF()`                     | Full report → SARIF JSON (excludes suppressed) |
| `report.ToSARIFFiltered(minSev)`       | Severity-filtered SARIF JSON                   |
| `report.WriteSARIF(w)`                 | Streaming SARIF output                         |
| `report.WriteSARIFFiltered(w, minSev)` | Streaming filtered output                      |

### Import

`FindingsFromSARIF(data)` → `([]Finding, error)` — round-trip fidelity via property bag

### Round-Trip Fidelity

- All non-standard fields preserved in `properties` bag (`go-finding/*` prefix)
- Suppressed findings excluded from export (lossy)
- `SeverityCritical` maps to SARIF `"error"` (no `"critical"` level in SARIF 2.1.0); original preserved in Properties

---

## 13. LSP Diagnostic Conversion

**Status:** STABLE

### Finding → LSP

`finding.ToLSP()` → `LSPDiagnostic`

Handles: severity mapping, 0-based conversion, Range, related information

### LSP → Finding

`FromLSP(fileURI, diag)` → `Finding`

Preserves: end position as Range, related information, raw LSP severity in Metadata

> **Known limitation:** Lossy conversion — `FixStrategy`, `Confidence`, `BeforeCode`, `AfterCode`, `Suppression`, `Metadata`, `Category`, `Tags` are lost in LSP format. Only position, severity, rule, message, and related info survive.

---

## 14. go/analysis Integration

**Status:** STABLE

| Function                                                      | Purpose                              |
| ------------------------------------------------------------- | ------------------------------------ |
| `FromDiagnostic(diag, fset, toolName, ruleCode, ...severity)` | `go/analysis.Diagnostic` → `Finding` |
| `FromTokenPosition(pos)`                                      | `token.Position` → `Position`        |
| `NodePosition(fset, node)`                                    | `ast.Node` → `Position`              |
| `NodeRange(fset, node)`                                       | `ast.Node` → `Range`                 |
| `FormatDiagnostic(diag, fset, name)`                          | Go-vet-style formatted string        |

Auto-detects suggested fixes and sets `FixStrategyDirect` with `AfterCode`.

---

## 15. Structured Errors

**Status:** STABLE

Category-based error types with `errors.Is` / `errors.As` support:

| Error           | Category     | Factory                          |
| --------------- | ------------ | -------------------------------- |
| `ErrValidation` | `validation` | `NewValidationError(msg, cause)` |
| `ErrIO`         | `io`         | `NewIOError(msg, cause)`         |
| `ErrParse`      | `parse`      | `NewParseError(msg, cause)`      |
| `ErrConflict`   | `conflict`   | `NewConflictError(msg, cause)`   |
| `ErrInternal`   | `internal`   | `NewInternalError(msg, cause)`   |

`FindingError` supports: `WithFinding()`, `WithPosition()`, `Unwrap()`, `Is()` for sentinel matching

Helpers: `IsFindingError(err)`, `GetCategory(err)`, `IsCategory(err, cat)`

---

## 16. Pipeline

**Status:** STABLE

The pipeline orchestrates a detect → triage → fix → verify loop.

### 16.1 Detector Interface

```go
type Detector interface {
    Name() string
    Detect(ctx context.Context) ([]finding.Finding, error)
}
```

Adapters: `DetectorFunc`, `NamedDetectorFunc(name, fn)`

### 16.2 Configuration

| Option                | Type                   | Default | Description                     |
| --------------------- | ---------------------- | ------- | ------------------------------- |
| `MaxIterations`       | `int`                  | 5       | Prevents infinite loops         |
| `ParallelDetectors`   | `bool`                 | `true`  | Concurrent detector execution   |
| `Timeout`             | `time.Duration`        | 10min   | Pipeline timeout                |
| `VerifyAfterFix`      | `bool`                 | `false` | Re-run detectors post-fix       |
| `GracefulDegradation` | `bool`                 | `false` | Continue on detector failures   |
| `DryRun`              | `bool`                 | `false` | Detect + triage only (no fixes) |
| `Retry`               | `*RetryConfig`         | `nil`   | Exponential backoff retries     |
| `Metrics`             | `*Metrics`             | `nil`   | Timing/count collection         |
| `CorrelateFindings`   | `bool`                 | `false` | Cross-tool correlation          |
| `Processors`          | `[]FindingProcessor`   | `nil`   | Composable finding transforms   |
| `OnFinding`           | `func(Finding)`        | `nil`   | Per-finding callback            |
| `OnFix`               | `func(Finding, bool)`  | `nil`   | Per-fix callback                |
| `OnIteration`         | `func(int, []Finding)` | `nil`   | Per-iteration callback          |

Config validation: `config.Validate()` returns joined errors for invalid values. `pipeline.New()` rejects invalid configs.

### 16.3 Pipeline Loop

1. **Detect** — Run detectors (parallel or sequential)
2. **Process** — Run `FindingProcessor` chain on raw findings
3. **Triage** — Categorize by `FixStrategy` (direct / suggest / none)
4. **Apply** — Apply direct fixes with conflict detection
5. **Repeat** — Until stable (zero findings) or `MaxIterations`

### 16.4 Finding Processors

**Status:** EXPERIMENTAL

Composable transforms that run on findings between detection and triage.

```go
type FindingProcessor interface {
    Process(ctx context.Context, findings []finding.Finding) ([]finding.Finding, error)
    Name() string
}
```

Adapters:

- `ProcessorFunc(fn)` — wraps a function as a `FindingProcessor` (name: `"anonymous"`)
- `NamedProcessorFunc(name, fn)` — wraps with a custom name

Processors are executed in order from `Config.Processors`. Use cases: filtering, enrichment, normalization, severity adjustment, deduplication.

### 16.5 Pipeline Result

| Field             | Type               | Description                      |
| ----------------- | ------------------ | -------------------------------- |
| `Stable`          | `bool`             | Reached zero findings            |
| `TotalIterations` | `int`              | Iterations executed              |
| `Iterations`      | `[]Iteration`      | Per-iteration details            |
| `TotalDetected`   | `int`              | Total findings across iterations |
| `Verification`    | `*VerifyResult`    | Post-fix verification            |
| `PartialErrors`   | `map[string]error` | Per-detector failures            |
| `Correlations`    | `[]Correlation`    | Cross-tool correlations          |
| `Metrics`         | `MetricsSnapshot`  | Timing and counts                |

### 16.6 Conflict Detection

**Status:** STABLE

- Detects overlapping fixes in the same file
- `FilterConflictingFixes()` — returns only non-conflicting fixes
- `AnalyzeConflicts()` — detailed `ConflictInfo` with reasons
- When multiple fixes overlap in same group, keeps first, marks rest as conflicts

### 16.7 Fix Application

**Status:** FUNCTIONAL

Two-phase fix engine:

1. **FixEngine (in-memory)** — `NewFixEngine()` provides a pure `Apply(lines, fixes)` that transforms string lines without filesystem access. Supports range-based and string-based replacement. Useful for testing and previewing changes.
2. **FixApplier (filesystem)** — `NewFixApplier(rootDir)` applies fixes to actual files with backup/rollback support.

Both engines support:

- File backup before modification
- Rollback on failure (restores all modified files)
- Context cancellation support mid-application
- Deterministic application order (sorted by path, descending by position)

> **Known limitation:** String-based replacement uses `strings.Replace` for nearest-to-line matching. If `BeforeCode` appears multiple times near the target line, it may pick the wrong occurrence.

### 16.8 Verification

**Status:** STABLE

Re-runs all detectors after fixes and categorizes findings:

| Category      | Meaning                              |
| ------------- | ------------------------------------ |
| `Fixed`       | Original findings no longer detected |
| `Remaining`   | Original findings still present      |
| `NewFindings` | Fresh findings introduced by fixes   |

`DiffFindings(original, post)` — standalone utility for comparing finding sets.

### 16.9 Metrics

**Status:** STABLE

Thread-safe metrics collection:

| Metric                | Method                           |
| --------------------- | -------------------------------- |
| Stage durations       | `RecordStage()`, `StageTiming()` |
| Detector timing       | `RecordDetector()`               |
| Findings per detector | `RecordDetector()`               |
| Fixes applied         | `RecordFix()`                    |
| Total duration        | `TotalDuration()`                |
| Point-in-time copy    | `Snapshot()` → `MetricsSnapshot` |

Auto-populated on `Pipeline.Run()` via `PipelineResult.Metrics`.

### 16.10 Retry

**Status:** STABLE

Configurable exponential backoff with jitter for flaky detectors:

```go
config := pipeline.DefaultRetryConfig()
// MaxRetries: 3, BaseDelay: 100ms, MaxDelay: 5s
```

`NewRetryDetector(inner, config)` wraps any `Detector` with retry logic.

Validation: `RetryConfig.Validate()` checks constraints.

### 16.11 Partial Success

**Status:** STABLE

When `GracefulDegradation` is enabled:

- `DetectPartial()` collects findings from successful detectors
- Failed detector errors stored in `PartialResult.Errors` map
- `PipelineResult.PartialErrors` exposes per-detector failures
- `FormatPartialErrors()` formats collected errors

### 16.12 File Backup & Rollback

**Status:** STABLE

- Creates backup copies before file modification
- `Backup(path)`, `Restore(path)`, `RollbackAll(paths)`
- Thread-safe, can be enabled/disabled
- On fix application failure: restores current file, then rolls back all previously modified files

---

## 17. Built-in Detectors

### 17.1 Go Vet Detector

**Status:** FUNCTIONAL

Runs `go vet -json ./...` and converts JSON output to Findings.

- Category: `correctness`
- Severity: `warning`
- FixStrategy: `suggest`
- Handles non-zero exit codes (still parses output)

### 17.2 Staticcheck Detector

**Status:** FUNCTIONAL

Runs `staticcheck -f json ./...` and converts JSON output to Findings.

- Severity: `error` or `warning` based on staticcheck output
- FixStrategy: `suggest`
- Confidence: `0.8`
- Category mapping: S/Q→style, U→unused, P/R/F→performance, A→correctness

> **Note:** Both detectors require the respective tools to be installed and available in `$PATH`.

---

## 18. CLI Tool

**Status:** FUNCTIONAL

Binary: `go-finding`

### Flags

| Flag              | Default | Description                            |
| ----------------- | ------- | -------------------------------------- |
| `-dir`            | `.`     | Root directory to analyze              |
| `-format`         | `text`  | Output format: `text`, `json`, `sarif` |
| `-severity`       | `info`  | Minimum severity filter                |
| `-max-iterations` | `1`     | Pipeline iterations                    |
| `-parallel`       | `true`  | Run detectors in parallel              |
| `-verify`         | `false` | Re-run detectors after fixes           |
| `-timeout`        | `10m`   | Pipeline timeout                       |
| `-config`         | (none)  | YAML/JSON config file                  |
| `-cpuprof`        | (none)  | CPU profile output                     |
| `-memprof`        | (none)  | Memory profile output                  |
| `-output`         | stdout  | Output file path                       |
| `-version`        | `false` | Print version                          |

### Config File (YAML/JSON)

```yaml
maxIterations: 3
parallelDetectors: true
verifyAfterFix: false
timeout: "5m"
detectors:
  - name: govet
  - name: staticcheck
```

### Default Behavior

Without `-config`: uses govet + staticcheck with the flag values.

### Plugin Detectors

`RegisterDetector(name, builder)` — thread-safe, allows adding custom detectors at runtime.

### Output Formats

| Format  | Description                                               |
| ------- | --------------------------------------------------------- |
| `text`  | Human-readable: `file:line:col: [severity] rule: message` |
| `json`  | Full JSON report                                          |
| `sarif` | SARIF 2.1.0                                               |

Metrics summary printed to stderr when available.

---

## 19. Testing

**Status:** STABLE

| Package                          | Coverage |
| -------------------------------- | -------- |
| Core (`finding`)                 | 99.6%    |
| Pipeline                         | 98.0%    |
| CLI (`cmd/go-finding`)           | 95.4%    |
| Detectors (`internal/detectors`) | 96.1%    |

Test categories:

- Unit tests per source file
- Integration tests (`pipeline/integration_test.go`, `cmd/go-finding/integration_test.go`)
- E2E tests (`cmd/go-finding/e2e_test.go`)
- Fuzz tests (`fuzz_test.go`, `id_fuzz_test.go`, `merge_fuzz_test.go`, `sarif_fuzz_test.go`)
- Property-based tests (`property_test.go`)
- Benchmarks (`bench_test.go`)
- Bug-specific regression tests (`*_bugfix_test.go`)

---

## 20. Examples

**Status:** FUNCTIONAL

Three runnable examples in `examples/`:

| Example     | Description                   |
| ----------- | ----------------------------- |
| `basic/`    | Direct Finding construction   |
| `builder/`  | Builder API usage             |
| `pipeline/` | Pipeline with custom detector |

> **Note:** Examples have no test files (compile-only check via `example_compile_test.go`).

---

## Summary Matrix

| Feature                           | Status       | Notes                                                             |
| --------------------------------- | ------------ | ----------------------------------------------------------------- |
| Finding type                      | STABLE       | Core data model, 99.5% coverage                                   |
| Builder API                       | STABLE       | Fluent construction with validation                               |
| Position & Range                  | STABLE       | Full spatial algebra (Contains, Overlaps, Intersection, Adjacent) |
| Severity (4 levels)               | STABLE       | With comparison operators                                         |
| FixStrategy (none/suggest/direct) | STABLE       | Production auto-fix for `direct`                                  |
| FixStrategy (ai)                  | RESERVED     | Constant exists, no AI backend                                    |
| Category (15 standard + custom)   | STABLE       | Domain classification                                             |
| Tags (multi-label)                | STABLE       | Supersedes deprecated singular Tag                                |
| Suppression                       | STABLE       | With TTL/expiry support                                           |
| Report container                  | STABLE       | Thread-safe, with summary statistics                              |
| Filtering & sorting               | STABLE       | Composable predicates + grouping                                  |
| Report merging                    | STABLE       | 3 deduplication strategies                                        |
| Cross-tool correlation            | FUNCTIONAL   | Simple heuristic, capped at 10K                                   |
| ID generation & parsing           | STABLE       | Hash-based fallback, Windows path handling                        |
| JSON serialization                | STABLE       | Streaming support, drops invalid findings                         |
| SARIF 2.1.0 export/import         | STABLE       | Round-trip via property bag                                       |
| LSP conversion                    | STABLE       | Lossy — drops fix/suppression metadata                            |
| go/analysis integration           | STABLE       | Full diagnostic → Finding conversion                              |
| Structured errors                 | STABLE       | 5 categories, errors.Is support                                   |
| Pipeline (detect→fix→verify)      | STABLE       | Iterative loop with configurable behavior                         |
| Finding processors                | EXPERIMENTAL | Composable transforms between detect and triage                   |
| Conflict detection                | STABLE       | Overlapping fix detection                                         |
| Fix application                   | FUNCTIONAL   | In-memory FixEngine + filesystem FixApplier, backup/rollback      |
| Verification                      | STABLE       | Diff-based: fixed / remaining / new                               |
| Metrics                           | STABLE       | Thread-safe, snapshot support                                     |
| Retry (exponential backoff)       | STABLE       | With jitter                                                       |
| Partial success                   | STABLE       | Graceful degradation on detector failure                          |
| File backup & rollback            | STABLE       | Automatic on fix failure                                          |
| Go vet detector                   | FUNCTIONAL   | Requires `go vet` in PATH                                         |
| Staticcheck detector              | FUNCTIONAL   | Requires `staticcheck` in PATH                                    |
| CLI tool                          | FUNCTIONAL   | 3 output formats, config file, profiling                          |
| Plugin detector registry          | STABLE       | Thread-safe `RegisterDetector`                                    |
| Config validation                 | STABLE       | Both pipeline and CLI configs                                     |
| Examples                          | FUNCTIONAL   | 3 runnable examples, compile-tested                               |

---

_Assisted-by: Crush <crush@charm.land>_
