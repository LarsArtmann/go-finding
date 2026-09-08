# Finding Groups: SARIF and LSP Representation

How `Finding.GroupID` maps to SARIF 2.1.0 and LSP, what the standards offer
for grouping, and the recipes for consumers that want to render groups.

## The domain problem

A clone detector (e.g., art-dupl) reports N findings that belong to one
logical set: "these N code blocks are duplicates of each other." Each member
is a standalone `Finding` with its own position, rule, and message;
`GroupID` is the string that ties them together. The group itself is not a
finding — it has no position of its own.

## Current wire representation (shipped)

| Layer | Representation | Lossless round-trip |
| ----- | -------------- | ------------------- |
| JSON  | `"groupId"` field on the finding object | yes |
| SARIF | property bag key `go-finding/groupId` on each result (`sarif_export.go`, `sarif_import.go`) | yes |
| LSP   | `LSPDiagnosticData.GroupID` inside the diagnostic's `data` property | yes |

The property-bag approach is deliberately boring: every SARIF consumer can
read the property bag, and unknown keys are ignored by tools that do not
care. `Report.GroupFindings()` reconstructs the groups in memory.

## What SARIF 2.1.0 offers for grouping

SARIF has **no first-class "clone group" object**. The candidate mechanisms,
verified against the OASIS spec:

1. **`result.correlationGuid`** — a GUID identifying an "equivalence class"
   of logically identical results. Closest standard fit, but it must be a
   GUID (our `GroupID` is a free-form string), and ecosystem support is thin
   (GitHub code scanning, for example, does not render equivalence classes).
2. **`result.relatedLocations`** — an array of location objects "relevant to
   the result other than the primary location". Display-only: a consumer can
   show "see also" links, but the related locations carry no rule, message,
   or per-member data of their own.
3. **`result.codeFlows` → `threadFlows` → `threadFlowLocation`** — execution
   path semantics ("how the program gets here"). Using it for clone
   membership is a semantic misuse and renders as call-stack-like UI.
4. **`result.partialFingerprints`** — stable identity for result matching
   across runs (dedup), not group membership. Could be abused as a group
   key, but consumers match on it rather than group by it.

### Recommendation

Keep the property bag as the canonical carrier (it round-trips losslessly
today). If a consumer needs standard-facing display links, post-process the
exported SARIF: pick one member as the primary result and attach the other
members' physical locations as `relatedLocations`. This is one-directional
(decoration for viewers); re-importing still relies on the property bag.
Do not overload `codeFlows`; treat `correlationGuid` as viable only if
`GroupID` semantics ever tighten to GUIDs (see ROADMAP, D7 follow-ups).

### Prototype: decorate a SARIF export with relatedLocations

```go
// after ToSARIFWithOpts(...), for each group with len(members) > 1:
// use the first member (sorted by position) as primary, link the rest.
primary := members[0]
for _, rel := range members[1:] {
    // sarifResult in this repo is hand-rolled (ADR #9); relatedLocations
    // would need to be added to sarifResult first — see roadmap note.
    _ = rel
}
```

This is intentionally not implemented in `sarif_export.go`: it adds
viewer-only structure that cannot round-trip, and every consumer's viewer
links groups differently. The property bag keeps the data; decoration is a
presentation concern.

## LSP recipe: grouping by `data.groupId`

LSP has no diagnostic grouping concept — diagnostics are flat. The recipe:

1. Convert findings with `Finding.ToLSP()`; each diagnostic carries its
   `GroupID` inside `Data` (`LSPDiagnosticData.GroupID`, JSON key
   `groupId`).
2. In the client (or language-server layer), bucket diagnostics by
   `data.groupId` where present:

```go
groups := make(map[finding.GroupID][]LSPDiagnostic)
for _, diag := range diagnostics {
    if diag.Data != nil && diag.Data.GroupID != "" {
        groups[diag.Data.GroupID] = append(groups[diag.Data.GroupID], diag)
    }
}
```

3. Render one "leader" per group (smallest position) and demote the other
   members: they already arrive with their severity intact, and clients that
   support diagnostic tags can fade them (tag `1` = Unnecessary) — see the
   `ToLSP` tag round-trip.
4. Actions (`CodeAction`) attach per member; for a "fix whole group"
   action, re-associate the members via the same `GroupID` bucket.

Because the group id survives JSON → SARIF → LSP → back, no additional
state is needed on the server side.
