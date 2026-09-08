# Fix Outcomes Guide

Per-finding fix results for content-level (`FixEngine`) and disk-level (`FixApplier`) fix runs. Covers the six outcome statuses, rollback semantics, error typing, and metrics aggregation introduced in v1.7.0.

## Table of Contents

- [Why outcomes?](#why-outcomes)
- [The six statuses](#the-six-statuses)
- [Content-level: ApplyWithOutcomes](#content-level-applywithoutcomes)
- [Querying results: OutcomeFor / OutcomeCounts / HasErrors](#querying-results-outcomefor--outcomecounts--haserrors)
- [Disk-level: ApplyWithReport](#disk-level-applywithreport)
- [Rollback semantics](#rollback-semantics)
- [Typed outcome errors](#typed-outcome-errors)
- [Metrics aggregation](#metrics-aggregation)
- [JSON serialization](#json-serialization)

---

## Why outcomes?

Before v1.7.0, a fix run answered two questions: how many fixes applied, and which findings were skipped as conflicts. Everything else was invisible: a provider that matched a finding but produced zero edits (a *refusal*) was indistinguishable from success; a provider error on one finding could roll back every clean file in the run (issue #28).

Outcomes make every finding's fate explicit. One `FixOutcome` per input finding, in input order — no silence, no aggregation before you see it.

## The six statuses

| Status      | Meaning                                                                      | `Err` set? |
| ----------- | ---------------------------------------------------------------------------- | ---------- |
| `applied`   | At least one edit for the finding was applied                                 | no         |
| `no-change` | The finding carries no code change (`HasCodeChange()` false)                   | no         |
| `refused`   | Every matching provider returned zero edits, without error — it saw the finding and declined | no |
| `conflict`  | The finding's edits overlapped an earlier finding's edits and were skipped     | no         |
| `invalid`   | Edits were resolved but dropped as invalid or out of bounds                    | no         |
| `failed`    | A provider returned an error while resolving the finding to edits              | **yes**    |

`refused` is the status that used to be invisible: it is not an error. A provider declining to edit (e.g. the pattern no longer matches after an earlier fix) is a normal, reportable outcome.

## Content-level: ApplyWithOutcomes

`FixEngine.ApplyWithOutcomes` applies findings to in-memory content and returns the complete result:

```go
engine := pipeline.NewFixEngine()

result := engine.ApplyWithOutcomes(content, findings)

for _, outcome := range result.Outcomes {
    fmt.Printf("%s: %s\n", outcome.Finding.ID, outcome.Status)
}
```

`result.Outcomes` has exactly one entry per input finding, in input order. The legacy entry points (`Apply`, `ApplyWithConflicts`) delegate to the same implementation and keep their old return shapes — they skip the outcome bookkeeping, so their allocation profile is unchanged from pre-v1.7.0.

```go
content, applied, n := engine.Apply(content, fixes)               // simple
applied, edits, conflicts, content, errs :=                       // detailed
    engine.ApplyWithConflicts(content, fixes)
result := engine.ApplyWithOutcomes(content, fixes)                // complete
```

## Querying results: OutcomeFor / OutcomeCounts / HasErrors

```go
result := engine.ApplyWithOutcomes(content, findings)

if result.HasErrors() {
    for _, err := range result.Errors {
        log.Printf("provider error: %v", err)
    }
}

if outcome := result.OutcomeFor(finding.ID("abc123")); outcome != nil {
    fmt.Println("status:", outcome.Status)
}

for status, count := range result.OutcomeCounts() {
    fmt.Printf("%s=%d\n", status, count) // e.g. applied=3, refused=1
}
```

`OutcomeFor` returns `nil` when the ID was not part of the run (not an error — check your inputs). Map iteration order over `OutcomeCounts` is unspecified; sort keys if you need stable output (the CLI's `Fix outcomes:` summary prints in canonical status order).

## Disk-level: ApplyWithReport

`FixApplier.ApplyWithReport` is the disk-level equivalent: it writes fixes, creates backups, and reports what happened per file and per finding.

```go
applier, err := pipeline.NewFixApplier(rootDir)
if err != nil { // backup dir creation failed
    return err
}
defer applier.Close()

report, err := applier.ApplyWithReport(ctx, findings)

fmt.Printf("applied:   %d\n", report.Applied)
for _, outcome := range report.Outcomes {
    if outcome.Status == pipeline.FixOutcomeFailed {
        log.Printf("%s failed: %v", outcome.Finding.ID, outcome.Err)
    }
}
for _, path := range report.RolledBack {
    log.Printf("rolled back: %s", path)
}
```

`ApplyReport` fields:

| Field         | Contents                                                                 |
| ------------- | ------------------------------------------------------------------------ |
| `Applied`     | Number of findings successfully written to disk                          |
| `AppliedFixes`| Applied findings in application order                                    |
| `ShiftMaps`   | `map[string]*LineShiftMap` per file with applied edits                    |
| `Outcomes`    | One `FixOutcome` per *fixable* input finding, in processing order         |
| `RolledBack`  | File paths restored from backup (see [rollback semantics](#rollback-semantics)) |

`report.FailedOutcomes()` isolates the `failed` entries.

**Soft vs. hard failures.** A provider resolve error on one finding is *soft*: the run continues, other findings in the same file still apply, and the errors come back both in `Outcomes` and as a joined error return. A hard file failure (write error, backup failure) stops the run and triggers rollback.

## Rollback semantics

Since v1.7.0 the default rollback policy is **per-file** (`RollbackPolicyFailingFile`, ADR-016): a hard failure on one file restores *that file only*; fixes already written to earlier files stay on disk. The previous all-or-nothing behavior is available opt-in:

```go
applier.SetRollbackPolicy(pipeline.RollbackPolicyAllFiles)
```

or via `pipeline.Config.FixRollbackAllFiles`, the config-file field `fixRollbackAllFiles`, or the CLI flag `-fix-rollback-all`.

Two nuances worth knowing:

- **`RolledBack` lists backed-up files, not modified files.** A file whose findings all soft-fail is still backed up before processing; if a later hard failure triggers rollback under `AllFiles`, restoring that file is a content no-op but the path still appears in `RolledBack` (and in the `(rolled back: ...)` error text). Assert on the paths you expect to have been *restored*, including unchanged ones.
- **Per-finding soft failures never roll back anything.** Only hard file failures (write, backup, cancellation) do.

Migration notes for consumers upgrading from ≤ v1.6.x: see [consumer-migration-v1.7.md](consumer-migration-v1.7.md).

## Typed outcome errors

`FixOutcome.Err` for `failed` outcomes is a `*finding.FindingError` (parse category) with the finding's position attached. The original provider cause is preserved in the chain:

```go
var ferr *finding.FindingError
if errors.As(outcome.Err, &ferr) {
    fmt.Println("at", ferr.File, ferr.Position)
}

if errors.Is(outcome.Err, myProviderErr) { // chains to the original cause
    // ...
}
```

`errorfamily.Classify(err)` maps these to `Rejection` (parse category) for routing decisions.

## Metrics aggregation

The pipeline records one outcome per finding automatically during the fix stage:

```go
m := pipeline.NewMetrics()
m.RecordOutcome(pipeline.FixOutcomeRefused)

counts := m.OutcomeCounts() // map[FixOutcomeStatus]int
```

`MetricsSnapshot.OutcomeCounts` carries a point-in-time copy in `PipelineResult`. The go-finding CLI prints a `Fix outcomes:` stderr summary from these counts (canonical status order, nonzero counts only).

## JSON serialization

`FixOutcome` and `FixApplyResult` marshal deterministically (`json.Deterministic(true)`, per the repo-wide determinism rule). Errors serialize as message strings; on unmarshal they come back as plain `errors.New` values — error *identity* does not survive serialization, so match on status, not on the error value, after a round-trip. `FixApplyResult.Content` serializes as base64.

---

_Related: [fix-engine.md](fix-engine.md) for the engine and providers, [finding-groups.md](finding-groups.md) for `GroupID`, [consumer-migration-v1.7.md](consumer-migration-v1.7.md) for upgrade notes._
