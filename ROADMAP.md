# Roadmap

> Long-term direction and raw ideas not yet refined into actionable tasks.
> For short-term, bounded work see [TODO_LIST.md](TODO_LIST.md).
> For shipped features and their status see [FEATURES.md](FEATURES.md).

---

## Current Phase: Consumer ecosystem growth

**Current version:** 1.7.0 ([unreleased work](CHANGELOG.md#unreleased) will accumulate toward v1.8.0)

v1.0.0 locked the API (2026-06-24). v1.1.0 added multi-module workspace, branded type safety, SARIF suppression round-trip, LSP data fidelity. v1.2.0 extracted `lockutil`, defragmented tests. v1.2.1 shipped 15+ correctness/security fixes, `encoding/json/v2` migration. v1.3.0 added 12 consumer-driven convenience APIs based on a full audit of 22 consumer projects. v1.4.0 added `go-error-family` integration (unified error classification), community readiness infrastructure (SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates), and retired the "zero external deps" principle in favor of a small, deliberate dependency surface. v1.4.1 eliminated all code duplication (zero clones at `-t 1`), consolidated test setup across all 4 modules, and resolved 7 pipeline lint issues. v1.5.0 shipped deterministic JSON/SARIF output, `ValidateAll` batch validation, `FlightRecorderHook` pipeline observability, and the `tagsEqual` fix for order-insensitive equality. v1.6.0 shipped `ParseConfidence`, `Template.Builder`, flight-recorder config-file integration, exported path-safety APIs (`ResolveSafePath` family), and 7 CI structural-check scripts. v1.7.0 shipped `GroupID` finding groups, per-finding fix outcomes (`ApplyWithOutcomes`/`ApplyWithReport`), per-file rollback default (ADR-016), typed outcome errors, and outcome metrics.

The library is production-ready and API-stable. The focus now shifts to growing the consumer ecosystem, expanding language coverage, and completing the public launch.

---

## v1.0.0-v1.7.0 - API lock, consumer convenience, observability, community readiness

**Status: Released.**

- API locked at v1.0.0. Breaking changes require major version bump.
- Multi-module workspace (core, pipeline, analysis, CLI) established.
- Branded types (`ID`, `RuleName`, `ToolName`, `FilePath`), hand-rolled SARIF, LSP round-trip fidelity.
- v1.3.0: Consumer-driven APIs (`BuildOrDefault`, `Template`, `NewReportFromFindings`, `FilePos`, `SeverityFromLevel`, `ApplySimpleFixes`, `CheckBinary`/`RunCmd`, `FormatTable`, `PriorityString`). File-level position validation relaxed.
- v1.4.0: `go-error-family` integration (`FindingError.ErrorCode()` / `ErrorFamily()`), community infrastructure (SECURITY.md, CODE_OF_CONDUCT.md, issue/PR templates), documentation accuracy sweep. "Zero external deps" principle retired.
- v1.4.1: Zero code duplication (extracted `must[T]`, `marshalJSONString`, `decodeConfig`, `fixEditJSON`), test setup consolidated (`NewParallelGomega`), 7 pipeline lint issues resolved.
- v1.5.0: Deterministic JSON/SARIF output (`json.Deterministic(true)` on all 8 marshal call sites, 8 byte-identity regression tests), `ValidateAll` batch helper, `FlightRecorderHook` (Go runtime execution trace flight recorder for pipeline observability), `tagsEqual` fix (order-insensitive tag equality aligning code with documented contract).
- v1.6.0: `ParseConfidence` + `Template.Builder` consumer convenience APIs, flight-recorder config-file integration (`FlightRecorderFileConfig` + CLI `flightRecorder` section), exported path-safety boundary (`ResolveSafePath`/`ResolveSafePathFrom`/`ResolveRoot`), per-module CHANGELOGs, go-arch-lint module boundary enforcement, 7 CI structural-check scripts.
- v1.7.0: `GroupID` finding groups (JSON/SARIF/LSP round-trip, `Report.GroupFindings()`), per-finding fix outcomes (`FixEngine.ApplyWithOutcomes`, `FixApplier.ApplyWithReport`, `Metrics.RecordOutcome`/`OutcomeCounts`), per-file rollback default (`RollbackPolicyFailingFile`, ADR-016; issue #28), typed outcome errors (`*finding.FindingError`), negative-LSP-tag rejection (`FuzzParseLSPDiagnosticTags`), LSP tag re-emission.

---

## Raw Ideas (not yet actionable)

These are directions worth exploring. They are **not** committed work - they exist to capture thinking before it is lost. When an idea becomes concrete enough to act on, it graduates to [TODO_LIST.md](TODO_LIST.md).

### AI-assisted remediation

`FixStrategyAI` is a reserved constant with no backend. A real implementation would need:

- A pluggable `AIProvider` interface (request -> suggested diff)
- Guardrails: sandboxed apply, verification re-run, human approval gate
- Cost/rate-limit awareness in the pipeline

This is the largest open product direction and the original motivation for the library's fix pipeline.

### Language expansion

The core `Finding` model and SARIF/LSP interchange are language-agnostic. The fix engine has a Go AST provider (`pipeline/goast/`) but the provider architecture supports more:

- Rust (`syn`-based provider)
- TypeScript/JavaScript (tree-sitter)
- Python (ast / libcst)

Each would live in its own subpackage to keep language-specific dependencies out of the core.

### Tooling integrations

- **IDE plugins** - VS Code / Neovim consuming LSP diagnostics from `ToLSP()`
- **LSP code action support** - `ToLSP()` emits diagnostics but not code actions (auto-fix proposals). The data model has `FixStrategy`/`BeforeCode`/`AfterCode` but LSP code actions would require `LSPCodeAction` wire types.
- **Watch mode** - re-run the pipeline on file change (`fsnotify`)
- **Interactive TUI** - triage and review findings before applying fixes
- **GitHub Actions action** - first-class SARIF upload with fix PR generation
- **Profile-guided optimization** - PGO investigation. The pipeline hot path (fix engine, merge) could benefit from PGO profiles.

### Consumer ecosystem

- **Consumer migration to v1.3.0+ APIs** - 14 Go consumers can now simplify their codebases using `BuildOrDefault`, `Template`, `SeverityFromLevel`, `FilePos`, `NewReportFromFindings`, `ApplySimpleFixes`, `ParseConfidence`, and `Template.Builder`. Each consumer independently reinvented these patterns.
- **More `ToolAdapter[O]` recipes** - Pre-built adapters for revive, ineffassign, errcheck, etc.
- **go-linter-sdk integration** - The sibling `go-linter-sdk` repo now has `WithToolName`, `RuleFunc.NewFinding`, `FilterRules`, and `ExitCodeByConfidence` (implemented in the 2026-08-08 cross-repo refactor session). The pilot migration of `go-humanize-linter` eliminated 97 LOC. Next: release go-finding v1.7.0, publish go-linter-sdk v0.2.0, then port more linters.

### FlightRecorder future directions

The FlightRecorder feature (`pipeline/flight_recorder.go`) is shipped with config-file integration (`FlightRecorderFileConfig` + `ResolveFlightRecorder()`).

**Triage (2026-09-08, FR1):** each idea scored Impact × Effort from the perspective of pipeline consumers. Two shipped along the way (context propagation, multi-recorder degradation); two graduated to actionable ROADMAP items below; the rest stay parked with reasons.

| Idea                                   | Impact | Effort | Verdict                                                                                          |
| -------------------------------------- | ------ | ------ | ------------------------------------------------------------------------------------------------ |
| Trace file rotation (max-files)        | High   | Low    | **GRADUATED** — long `-trace-slow` runs produce hundreds of `.trace` files; cap + prune oldest     |
| Compressed trace output (gzip)         | Med    | Low    | **GRADUATED** — wrap `WriteTo` in `gzip.Writer`; opt-in config field                              |
| Automatic pprof capture                | Med    | Med    | Park — useful but duplicates what `runtime/pprof` flags already give operators                     |
| Continuous trace sampling (1% knob)    | Med    | Low    | Park — revisit once rotation lands (sampling without rotation is a disk hazard)                   |
| Core package trace helper (`finding/tracing`) | Low | Med | Rejected — speculative generalization; only one consumer pattern exists (pipeline)          |
| OpenTelemetry bridge                   | Low    | High   | Rejected — adds a heavy dependency for a rare use case in static-analysis tooling                 |
| Trace diff tool                        | Low    | High   | Rejected — niche debugging aid; belongs in consumer tooling, not the library                       |
| AI-assisted trace analysis             | Low    | High   | Rejected — LLM analysis is consumer territory; the library's job is capturing the trace           |
| Context propagation                    | —      | —      | **Shipped v1.5.0** — `Snapshot(ctx, reason)` accepts context; cancelled contexts skip the write   |
| Multiple recorder support              | —      | —      | **Shipped v1.6.0** — `Degraded()` mode degrades gracefully when Go's singleton recorder is taken  |

Graduated items (actionable when picked up):

- **Trace file rotation** - Add `MaxFiles` to `FlightRecorderConfig`/`FlightRecorderFileConfig`; on snapshot, delete oldest `.trace` files beyond the cap. Pairs naturally with the existing `MaxBytes` field.
- **Compressed trace output** - Add `Compress bool` (config `compressed`); wrap snapshot writes in `gzip.Writer` with `.trace.gz` suffix. Default off to preserve `go tool trace` compatibility expectations.

### Hardening (owner decisions pending)

These are known design tensions deferred because they require breaking changes. Concrete designs from the [data-model review](docs/reviews/archived/2026-07-18_21-10_data-model-review.html).

- **Position zero-value** - `Position{}` has `Offset=0` (valid byte 0), not "unset" (`position.go:33-38`: 0=unset for Line/Column, -1=unset for Offset). Resolved pragmatically in v0.9.0 with `-1` sentinel, but a type-safe redesign using `Option[T]` generic helpers is still on the table for v2.0.
- **`Range.End` zero-value ambiguity** - same class of issue as Position.
- **FixStrategy as closed union** - Current `type FixStrategy string` (`fix_strategy.go:4`) with string constants loses type safety. v2.0 design: `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved`.
- **Pointer-as-state fields** - `Range *Range` (`finding.go:29`), `Suppression *Suppression` (`finding.go:33`), `ExpiresAt *time.Time` (`suppression.go:20`), and `RelatedRef.Range *Range` (`finding.go:96`) all encode 3 states (nil/zero/valid) in a single pointer.
- **Tags to TagSet** - `Tags []Tag` (`finding.go:21`) forces order-insensitive equality in `finding_equal.go`. v2.0: `TagSet map[Tag]struct{}`.
- **Finding sub-struct composition** - Current flat struct (`finding.go:8-48`). v2.0: compose from `Identity{}`, `Location{}`, `Classification{}`, `Fix{}`. Changes JSON shape - must batch.

---

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Web UI** - Not aligned with the library's core purpose (data model + pipeline, not presentation layer).
- **Hosted/SaaS offering** - Too costly relative to impact for a library project.
- **Non-Go language providers (pre-v2)** - Language expansion is roadmap, but not until the Go provider ecosystem is fully proven.
- **Generic `JSONToolDetector`** - Too opinionated; every tool's JSON shape differs. `CheckBinary`/`RunCmd` helpers suffice.
- **`Properties map[string]any`** - Explicitly banned. `Metadata map[string]string` stays for type safety and interchange simplicity.

---

_Assisted-by: Crush <crush@charm.land>_
