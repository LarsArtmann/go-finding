<!-- proposed title:
RepairResult has no structured facts channel — provenance degrades into Description prose
labels: enhancement
-->

## Why

`RepairResult` is `struct { Description string }` today. The anti-lie rule behind it is right — BuildFlow re-runs Detect and measures the finding delta instead of trusting self-reported fix counts (`toolsdk/spec.go:130-135`). But the rule conflates two things: fix COUNTS (claims about effect, correctly re-measured) and FACTS (what the tool did: which files it touched, what the pre-state was). Facts have no channel, so they degrade into prose.

Concrete fallout from the first real toolsdk conversion — go-auto-upgrade's provider (`larsartmann/go-auto-upgrade` `pkg/provider`, @ d6c1158):

- The compile gate's baseline-failed flag is buried in a sentence: `"; tree was already broken before migration (baseline build failed), changes kept"` (`pkg/provider/provider.go:290-292`). A host that wants to surface "this repair ran on an already-broken tree" would have to string-match its own tool's Description.
- BuildFlow's old hand-written glue carried structured telemetry (`RunRecord.FailureCacheSuppressed`). The toolsdk Spec has no equivalent field, so the host loses it on switch-over (documented as a deliberate delta in go-auto-upgrade's AGENTS.md, 2026-10-11).

## Source-level cause

```go
// toolsdk/spec.go:138-143 (v1.15.0, go-finding @ 91a216f)
// RepairResult describes the outcome of a repair run. Only Description is
// meaningful to BuildFlow; fix counts are measured by re-detection, not
// self-reported (structural anti-lie design).
type RepairResult struct {
	Description string
}
```

## Proposal (additive, non-breaking)

- [ ] `RepairResult.ModifiedFiles []string` — files the repair wrote (or would write under dry-run)
- [ ] `RepairResult.Facts []Fact{Key, Value string}` (or `map[string]string`) — tool-defined provenance pairs, e.g. `baseline-failed=true`, `failure-cache-suppressed=true`
- [ ] Doc contract: facts are non-gating provenance. BuildFlow keeps measuring fix counts by re-detection; nothing downstream gates on a fact. The anti-lie design stays intact — facts describe the run, they don't claim its effect.

## To verify

- Existing specs compile unchanged (additive fields only).
- go-auto-upgrade's `describe()` string surgery disappears; `BaselineFailed` moves to a fact.
- BuildFlow's conversion (`ToolFromSpec`) maps or explicitly ignores the new fields, updated in the same change per the Spec↔domain field-parity rule.

💘 Generated with Crush
