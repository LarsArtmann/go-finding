# Consumer Migration — v1.6.0 → v1.7.0

Adoption guide for the v1.7.0 release train: per-finding fix outcomes, the
rollback policy default change, finding groups, and the v1.5.0 quality-of-life
APIs most consumers have not picked up yet.

Everything is additive unless explicitly marked **behavior change**.

---

## Rollback default changed (behavior change, issue #28)

**Before (≤ v1.6.0):** one file failure restored *every* file modified earlier
in the run — one bad file nuked all clean fixes.

**After (v1.7.0):** only the failing file is restored; earlier files keep their
applied fixes. Soft per-finding failures (provider resolve errors, refused
findings) no longer abort the run at all — they are reported while all applied
fixes stay on disk.

**Who is affected:** anyone running `--fix` across multiple files who relied on
all-or-nothing semantics. If you want the old behavior back:

```go
applier.SetRollbackPolicy(pipeline.RollbackPolicyAllFiles)
```

```yaml
# config file
fixRollbackAllFiles: true
```

```bash
# CLI
go-finding -dir . -fix-rollback-all
```

## Per-finding outcomes (issue #27)

`Apply`/`ApplyWithConflicts` cannot tell you which findings silently did
nothing. `ApplyWithOutcomes` gives you one `FixOutcome` per input finding:

```go
result := engine.ApplyWithOutcomes(content, fixes)

for _, o := range result.Outcomes {
    switch o.Status {
    case pipeline.FixOutcomeApplied:
    case pipeline.FixOutcomeRefused: // provider declined — no error
    case pipeline.FixOutcomeFailed:  // o.Err is matchable via errors.Is
    case pipeline.FixOutcomeConflict, pipeline.FixOutcomeInvalid,
        pipeline.FixOutcomeNoChange:
    }
}
```

On the filesystem level, `ApplyWithReport` returns an `ApplyReport` with
outcomes, shift maps, and `RolledBack` files; `report.FailedOutcomes()` isolates
provider failures for surfacing in your CLI.

**Migrate:** replace `ApplyWithConflicts` calls where you manually reconciled
applied-vs-input with `ApplyWithOutcomes` + `OutcomeFor(id)`.

## Finding groups (GroupID)

Tie related findings to one logical issue — the canonical case is a clone
group of N duplicated blocks from art-dupl:

```go
f := tmpl.Builder("dup-1", "cloned block", finding.SeverityWarning, pos).
    WithGroupID("clone-group-1").
    MustBuild()
```

Groups round-trip through JSON (`groupId`), SARIF (`go-finding/groupId`
property), and LSP (`LSPDiagnosticData.GroupID`). `Report.GroupFindings()`
returns active findings grouped by `GroupID`.

## v1.5.0 APIs worth adopting while you are here

**`Template.Builder`** — template defaults + per-finding overrides without
factory functions:

```go
// Before
f := newMigrationFinding(rule, msg, sev, pos) // your hand-rolled helper

// After
f := finding.NewTemplate("my-tool").
    WithCategory(finding.CategoryStyle).
    WithFixStrategy(finding.FixStrategyDirect).
    Builder(rule, msg, sev, pos).
    WithConfidence(finding.ConfidenceHigh).
    MustBuild()
```

**`ParseConfidence`** — replace your `--min-confidence` flag switch:

```go
conf, err := finding.ParseConfidence(flagValue) // "high", "0.42", "" → low
```

**`ResolveSafePath`** — replace hand-rolled symlink/traversal checks for any
finding-derived path before touching the filesystem:

```go
safe, err := pipeline.ResolveSafePath(rootDir, findingPath)
```

**`RegisterSeverityAlias` / `LookupSeverityAlias`** — instead of private
`mapSeverity` copies.

---

## Version bump checklist for consumers

1. `go get github.com/larsartmann/go-finding@v1.7.0` (+ sub-modules as needed)
2. If you run multi-file fixes and need all-or-nothing: opt back in (see above)
3. Compile: `GOEXPERIMENT=jsonv2 go build ./...` (jsonv2 is required)
4. If you consumed `ApplyWithConflicts` reconciliation code — migrate to
   `ApplyWithOutcomes`
