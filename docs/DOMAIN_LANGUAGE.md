# Domain Language

A **Unified Language** for go-finding — shared across Contributors, Consumers, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a consumer, define it here.

## Glossary

| Term       | Definition                                                 | Context                              |
| ---------- | ---------------------------------------------------------- | ------------------------------------ |
| go-finding | The project/library name                                   | A unified finding model and pipeline |
| Finding    | A single issue detected by a static analysis tool          | Core domain object                   |
| Report     | A thread-safe container for findings from one tool run     | Aggregation root                     |
| Detector   | A component that runs analysis and produces Findings       | Pipeline input                       |
| Pipeline   | The detect → triage → fix → verify orchestration loop      | Pipeline module                      |
| FixEngine  | Pure byte-level edit application without filesystem access | In-memory fix application            |
| FixApplier | Filesystem-aware fix application with backup/rollback      | Disk-based fix application           |

## Entities

Objects with identity and lifecycle.

| Term     | Definition                                                            | Context                |
| -------- | --------------------------------------------------------------------- | ---------------------- |
| Finding  | A single issue: what's wrong, where, how severe, and how to fix it    | Identified by ID       |
| Report   | A collection of Findings from one analysis run, with summary stats    | Identified by ToolInfo |
| Pipeline | A stateful orchestrator that runs iterations until stable or max      | Single-use (Run once)  |
| Detector | A named analysis tool (govet, staticcheck, custom) producing Findings | Plugin architecture    |

## Value Objects

Immutable objects defined by attributes.

| Term            | Definition                                                         | Context                       |
| --------------- | ------------------------------------------------------------------ | ----------------------------- |
| Position        | A location in source: File (FilePath), Line, Column, Offset        | 1-based Line/Column           |
| Range           | A span from Start to End Position (same file)                      | For region-based findings     |
| Severity        | Urgency level: info, warning, error, critical                      | Ordered, comparable           |
| Confidence      | Certainty of a Finding on a 0.0–1.0 scale                          | Named float64 type            |
| FixStrategy     | How a Finding can be remediated: none, suggest, direct, ai         | Determines auto-apply         |
| Category        | Domain classification: security, style, correctness, etc. (16 std) | Extensible                    |
| Tag             | Multi-label classification for richer filtering                    | Extensible                    |
| Suppression     | An expiring mark that a Finding is intentionally ignored           | Kind + Rule + optional TTL    |
| FixEdit         | A byte-level edit: Offset, Length, Replacement                     | Descending-offset apply       |
| Snippet         | Surrounding source code context around a Finding's position        | For display and SARIF         |
| RelatedRef      | A cross-reference from one Finding to another with a relation kind | Spatial or logical link       |
| Correlation     | A scored relationship between Findings from different tools        | Spatial + proximity heuristic |
| Template        | Pre-configured builder factory for batch finding creation (v1.3.0) | Stamp common fields once      |
| SimpleFixResult | Outcome of a BeforeCode→AfterCode fix (Applied, Reason)            | Core package fix application  |
| ID              | Branded string type for Finding identity                           | Distinct from RuleName etc.   |
| RuleName        | Branded string type for rule identifiers                           | Distinct from ID etc.         |
| ToolName        | Branded string type for tool names                                 | Distinct from FilePath etc.   |
| FilePath        | Branded string type for file paths                                 | Distinct from ToolName etc.   |

## Bounded Contexts

Subsystems with distinct vocabulary.

| Context        | Description                                                                             |
| -------------- | --------------------------------------------------------------------------------------- |
| Detection      | Running Detectors to produce raw Findings                                               |
| Transformation | Applying FindingTransformer chain (filter, enrich, normalize) between detect and triage |
| Triage         | Categorizing Findings by FixStrategy (direct, suggest, none)                            |
| Fix            | Resolving Findings to FixEdits via FixProviders, applying byte-level edits              |
| Verification   | Re-running Detectors post-fix to confirm issues are resolved                            |
| Interchange    | Lossless serialization: SARIF 2.1.0, LSP diagnostics, JSON                              |

## Commands

Actions the system performs.

| Command   | Description                                                       |
| --------- | ----------------------------------------------------------------- |
| Detect    | Run registered Detectors and collect Findings                     |
| Transform | Apply FindingTransformer chain to raw Findings                    |
| Triage    | Group Findings by FixStrategy for the apply phase                 |
| Apply     | Resolve Findings to FixEdits and apply them (with conflict check) |
| Verify    | Re-run Detectors to confirm fixes resolved original Findings      |
| Correlate | Find related Findings across tools (spatial + proximity)          |

### v1.3.0 Convenience Commands

| Command               | Description                                                         |
| --------------------- | ------------------------------------------------------------------- |
| BuildOrDefault        | Build finding with zero-value fallback on validation error          |
| NewReportFromFindings | One-step report creation (NewReport + AddFindings + ComputeSummary) |
| ApplySimpleFixes      | BeforeCode→AfterCode string replacement without pipeline FixEngine  |
| FilePos               | Create a file-level Position (Line=0, Offset=-1)                    |
| SeverityFromLevel     | Map severity string (canonical + aliases) to Severity with fallback |
| CheckBinary           | Verify a binary exists in system PATH                               |
| RunCmd                | Execute external command and capture stdout                         |
| FormatTextRich        | Rich text output with emoji severity badges and category display    |

## Events

Things that happen in the domain.

| Event              | Description                                                       |
| ------------------ | ----------------------------------------------------------------- |
| Finding detected   | A Detector produced a new Finding                                 |
| Finding fixed      | A direct fix was successfully applied                             |
| Finding remaining  | A previously detected Finding is still present after verification |
| New finding        | A Finding introduced by a fix that wasn't there before            |
| Conflict detected  | Two or more fixes overlap in the same file region                 |
| Iteration complete | One full detect → triage → apply pass finished                    |
| Pipeline stable    | No new Findings detected — the loop can stop                      |
| Stage begin        | A StageHook fires before a pipeline stage executes                |
| Stage end          | A StageHook fires after a pipeline stage completes                |
| Slow stage         | A pipeline stage exceeded the FlightRecorder slow-stage threshold |
| Snapshot captured  | The FlightRecorder wrote a trace snapshot to disk                 |

---

## Observability

| Term               | Definition                                                     | Context                       |
| ------------------ | -------------------------------------------------------------- | ----------------------------- |
| StageHook          | Interface for before/after stage boundary notifications        | Abort on error                |
| FlightRecorderHook | Wraps `runtime/trace.FlightRecorder`; snapshots on slow stages | Diagnostic-only, never aborts |
| Metrics            | Timing and count data: stage durations, detector times, fixes  | Collected during pipeline run |
| MetricsSnapshot    | Immutable point-in-time copy of Metrics                        | Returned in PipelineResult    |

---

## Pipeline Configuration

| Term                       | Definition                                                         | Context                        |
| -------------------------- | ------------------------------------------------------------------ | ------------------------------ |
| Iteration                  | One complete detect → triage → apply cycle                         | Repeats until stable or max    |
| CompletionReason           | Why the pipeline finished: stable, max-iterations, cancelled, etc. | Returned in PipelineResult     |
| DryRun                     | Run detect+triage but skip fix application                         | "What would happen?" analysis  |
| GracefulDegradation        | Continue on detector failures, collecting partial results          | Partial errors in result       |
| ByteLevelConflictDetection | Precise byte-overlap detection for conflicting fixes               | More accurate than range-based |
| VerifyAfterFix             | Re-run detectors after fixes to confirm resolution                 | Final verification pass        |

---

## Finding Identity

**Finding ID** — The stable unique identifier for a finding, produced by `GenerateID(tool, rule, pos)`. Format: `tool:rule:file:line:col` (human-readable). Two findings are identical if and only if their IDs are equal (ADR #12).

**Key** — A fallback identifier for findings without an ID, built from `ToolName + File + Rule + Message`. Includes Message (unlike GenerateID), so two findings with different messages have different Keys even at the same position. Prefer ID as canonical identity.

**Dedup Key** — A merge-time identifier used by `MergeIter` and `Combine`. Three strategies: ByID (exact), ByPosition (tool+file+line+col), ByRule (rule+file+line+col). These are deliberate relaxations of canonical identity, not competing definitions.

---

## Fix Application (pipeline)

| Term            | Definition                                                                                                           | Context                        |
| --------------- | -------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| Fix Outcome     | Per-finding result of a fix attempt: one of applied, no-change, refused, conflict, invalid, failed                   | `FixApplyResult.Outcomes`      |
| Applied         | The finding's edits were resolved and written into the content                                                       | `FixOutcomeApplied`            |
| No-change       | The finding carries no code change (`BeforeCode`/`AfterCode` absent)                                                 | `FixOutcomeNoChange`           |
| Refused         | Every matching provider returned zero edits without error — the provider saw the finding and declined                | `FixOutcomeRefused`            |
| Conflict        | The finding's edits overlapped an earlier finding's edits and were skipped                                           | `FixOutcomeConflict`           |
| Invalid         | The finding's edits were resolved but dropped as out-of-bounds during application                                    | `FixOutcomeInvalid`            |
| Failed          | A provider error occurred; the cause is carried in `FixOutcome.Err` and matchable via `errors.Is`                    | `FixOutcomeFailed`             |
| Rollback Policy | Scope of restoration after a hard file error: default restores only the failing file; `AllFiles` restores everything | `FixApplier.SetRollbackPolicy` |
| Shift Map       | Line-offset mapping from before-fix to after-fix content, keeping later findings resolvable after earlier edits      | `ApplyReport.ShiftMaps`        |

---

## Groups

| Term        | Definition                                                                                                                    | Context                    |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------- |
| Group       | A named set of findings belonging to one logical unit, identified by `Finding.GroupID` (e.g. N blocks of one cloned function) | `GroupID` branded type     |
| Clone Group | A group produced by a duplicate-code detector (e.g. art-dupl): each member is one occurrence of the same cloned code          | Canonical GroupID use case |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
