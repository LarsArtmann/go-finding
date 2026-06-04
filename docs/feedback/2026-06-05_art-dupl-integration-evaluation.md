# Feedback Report: go-finding as Output Target for art-dupl

**Date:** 2026-06-05
**Author:** Lars Artmann (via Crush)
**Context:** Evaluation of go-finding as the interchange/output layer for [art-dupl](https://github.com/LarsArtmann/art-dupl) — a Go code clone detection tool. art-dupl detects duplicated code via suffix tree + hash-based methods on ASTs and outputs results in text, HTML, JSON, SARIF, and plumbing formats.

---

## Executive Summary

go-finding is well-positioned to serve as art-dupl's output layer. The `Finding` type covers 80% of what art-dupl needs. The remaining 20% falls into two categories: **structural gaps** (things the type cannot express today) and **interchange fidelity gaps** (data lost in SARIF/LSP conversion). All gaps are small, targeted, and solvable without breaking changes.

**Verdict:** Adopt go-finding as art-dupl's output adapter. Implement 3 structural improvements in go-finding first.

---

## What Works Well

### 1. `Position` and `Range` are solid

- 1-based with `Offset` — matches most tooling conventions
- Full spatial algebra (`Contains`, `Overlaps`, `Intersection`, `Adjacent`, `Compare`)
- `IsZero()`, `HasLocation()`, `IsInverted()` — good ergonomics
- art-dupl's `syntax.Node` (Pos/End via `token.Pos`) maps cleanly: `token.Position.Line` → `Position.Line`, `token.Position.Column` → `Position.Column`

### 2. `Related` with `RelationKind` is the right idea

- `RelationCloneOf` exists as a standard constant — someone anticipated this use case
- Bidirectional links via `FindingID` + `Position` enable graph traversal
- `ToLSP()` already emits related information from `Related[]`

### 3. `CategoryDuplication` exists

- Predefined constant — art-dupl doesn't need a custom category
- Consistent with other categories (`CategorySecurity`, `CategoryPerformance`, etc.)

### 4. `Metadata map[string]string` is the right escape hatch

- art-dupl can encode clone-specific data (similarity %, group size, detection method) without go-finding needing to understand it
- String values are serializable by default — no `map[string]any` footguns

### 5. SARIF round-trip is mature

- Bidirectional `ToSARIF()` / `FindingsFromSARIF()` with property-bag preservation
- Streaming export via `json.Encoder`
- art-dupl already has SARIF output — migrating to go-finding's implementation reduces duplication

### 6. `Report` as a container is exactly right

- Thread-safe, zero-value usable
- `Summary` with `BySeverity`, `ByCategory`, `FilesAffected`, `FilesScanned`, `DurationMs` — all relevant for art-dupl's stats output
- `ActiveFindings()` handles suppression filtering

---

## Structural Gaps

These are things art-dupl needs that go-finding's type system cannot express today.

### GAP-1: `RelatedRef` has no `Range`

**Problem:** art-dupl clones have a span (start line → end line). `RelatedRef` only carries a `Position` (a point). When `ToLSP()` emits related information, it creates zero-length ranges.

**Current:**
```go
type RelatedRef struct {
    FindingID string       `json:"findingId"`
    Relation  RelationKind `json:"relation"`
    Position  Position     `json:"position"` // point only — no span
}
```

**Proposed:**
```go
type RelatedRef struct {
    FindingID string       `json:"findingId"`
    Relation  RelationKind `json:"relation"`
    Position  Position     `json:"position"`
    Range     *Range       `json:"range,omitempty"` // span of the related location
}
```

**Impact:**
- `ToLSP()` can emit proper ranges for related information
- art-dupl can represent each clone's full extent
- Backward-compatible: `*Range` is optional, zero value is `nil`

### GAP-2: No group/collection concept

**Problem:** art-dupl produces **clone groups** — N files containing identical code. A group is a first-class entity with its own metadata (number of clones, total lines duplicated, similarity percentage). go-finding has no way to say "these 5 findings belong to the same group."

**Workarounds and their costs:**

| Approach | Cost |
|----------|------|
| Link all findings via `Related[clone-of]` | O(N²) links for a group of N. No group-level metadata. SARIF doesn't reconstruct the group. |
| Encode group JSON in `Metadata["group"]` | Works but opaque. Can't filter/sort by group. No type safety. |
| Use one Finding per group + Related refs | Finding has one `Position` — which clone is "primary"? Arbitrary. |

**Proposed:** Add `GroupID string` to `Finding`:

```go
type Finding struct {
    // ... existing fields ...
    GroupID string `json:"groupId,omitempty"` // Groups findings into a logical set
}
```

Plus a `Report.GroupFindings() map[string][]Finding` accessor.

**Impact:**
- art-dupl sets `GroupID` to a deterministic hash of the clone group
- Downstream tools can aggregate, filter, and display by group
- SARIF exports `groupId` as a property (round-trips via property bag)
- LSP can group diagnostics in the problems panel

### GAP-3: No similarity/confidence metric per relationship

**Problem:** art-dupl knows the similarity percentage between two clones (e.g., 97.3% identical). This is per-relationship, not per-finding. `RelatedRef` has nowhere to put this.

**Proposed:** Use `Metadata` convention — but acknowledge the ergonomics are poor:

```go
// Convention: encode similarity in metadata
ref := RelatedRef{
    FindingID: cloneID,
    Relation:  RelationCloneOf,
    Position:  clonePos,
}
// No place for similarity % without Metadata on RelatedRef itself
```

**Alternative (breaking):** Add `Metadata map[string]string` to `RelatedRef`. This is consistent with `Finding.Metadata` but adds allocation overhead for the common case where `Related` has no metadata.

**Recommendation:** Defer. Use `Finding.Metadata["go-finding/clone-similarity"]` on the primary finding. Revisit if more consumers need per-relationship metadata.

---

## Interchange Fidelity Gaps

Data that exists in `Finding` but is lost during conversion to external formats.

### GAP-4: `LSPDiagnostic` missing `DiagnosticTag`

**Problem:** LSP 3.15+ defines `DiagnosticTag` with values `Unnecessary (1)` and `Deprecated (2)`. Code duplicates are "unnecessary" code. go-finding's `LSPDiagnostic` doesn't support this.

**Current:**
```go
type LSPDiagnostic struct {
    Range    LSPRange         `json:"range"`
    Severity LSPSeverity      `json:"severity,omitempty"`
    Code     string           `json:"code,omitempty"`
    Source   string           `json:"source,omitempty"`
    Message  string           `json:"message"`
    Related  []LSPRelatedInfo `json:"relatedInformation,omitempty"`
}
```

**Proposed:**
```go
type LSPDiagnosticTag int

const (
    LSPDiagnosticTagUnnecessary LSPDiagnosticTag = 1
    LSPDiagnosticTagDeprecated  LSPDiagnosticTag = 2
)

type LSPDiagnostic struct {
    Range    LSPRange           `json:"range"`
    Severity LSPSeverity        `json:"severity,omitempty"`
    Code     string             `json:"code,omitempty"`
    Source   string             `json:"source,omitempty"`
    Message  string             `json:"message"`
    Tags     []LSPDiagnosticTag `json:"tags,omitempty"`
    Related  []LSPRelatedInfo   `json:"relatedInformation,omitempty"`
}
```

**Impact:**
- art-dupl sets `Tags: []LSPDiagnosticTag{LSPDiagnosticTagUnnecessary}` — IDEs render duplicates as faded/struck-through
- Backward-compatible: `Tags` is `omitempty`
- `FromLSP` should preserve these in `Finding.Metadata`

### GAP-5: `Snippet` lost in SARIF export

**Problem:** `Finding.Snippet` exists but `findingToSARIF()` doesn't export it. SARIF 2.1.0 supports `location.contextRegion` for surrounding code context.

**Impact:** art-dupl sets `Snippet` to the duplicated code. After SARIF round-trip, it's gone.

**Proposed:** Add snippet to SARIF export via `contextRegion` or as a property.

### GAP-6: `ToLSP()` hard-codes zero-length ranges for `Related`

**Problem:** Even after fixing GAP-1 (adding `Range` to `RelatedRef`), `ToLSP()` currently hard-codes:

```go
Range: LSPRange{Start: lspPos, End: lspPos},
```

This should use `rel.Range` when available.

---

## Quality Improvements (Not art-dupl-specific)

### GAP-7: `Category.IsValid()` accepts typos

`Category("securty").IsValid()` returns `true` — any non-empty lowercase-hyphen string passes. This affects all consumers, not just art-dupl.

**Proposed:** Add a registered-categories set and check against it. Custom categories could be registered at init time or via a `RegisterCategory` function.

### GAP-8: `FromLSP` doesn't preserve `DiagnosticTag`

When `FromLSP` creates a `Finding` from an LSP diagnostic, it loses the `DiagnosticTag` information. If we add GAP-4, `FromLSP` should preserve these in metadata (e.g., `Metadata["go-finding/lsp-diagnostic-tags"] = "1,2"`).

### GAP-9: No `iter.Seq[Finding]` on `Report.All()`

Go 1.26's `iter.Seq` would make report iteration idiomatic. The TODO list already tracks this. Low priority but high ergonomics value.

---

## Priority Matrix

| Priority | Gap | Effort | art-dupl Value | General Value |
|----------|-----|--------|----------------|---------------|
| **P0** | GAP-1: `RelatedRef.Range` | S | Critical — clone spans | High — any multi-location finding |
| **P1** | GAP-2: `GroupID` | S | High — clone group identity | Medium — any batched finding set |
| **P1** | GAP-4: `DiagnosticTag` | S | High — IDE rendering | High — "unnecessary" code everywhere |
| **P2** | GAP-6: `ToLSP` use `RelatedRef.Range` | S | Blocked on GAP-1 | Same |
| **P2** | GAP-5: Snippet in SARIF | S | Medium — code context round-trip | Medium |
| **P3** | GAP-3: Per-relationship metadata | M | Low — can use Finding.Metadata | Low |
| **P3** | GAP-7: Strict `Category.IsValid()` | S | None | High — all consumers |
| **P3** | GAP-8: `FromLSP` preserve `DiagnosticTag` | S | None | Medium — bidirectional fidelity |
| **P3** | GAP-9: `iter.Seq` on Report | S | None | Low — Go 1.26 only |

Effort: S = small (≤1 day), M = medium (1-3 days)

---

## Implementation Order

1. **GAP-1** — `RelatedRef.Range` (unblocks GAP-2 and GAP-6)
2. **GAP-4** — `DiagnosticTag` (independent, high value)
3. **GAP-2** — `GroupID` (depends on use-case validation)
4. **GAP-6** — `ToLSP` uses `RelatedRef.Range` (trivial after GAP-1)
5. **GAP-5** — Snippet in SARIF (independent)

GAP-7, GAP-8, GAP-9 are independent and can be done in any order.

---

## What art-dupl Would Look Like

After GAP-1, GAP-2, GAP-4:

```go
func cloneGroupToFindings(group *artdupl.CloneGroup) []finding.Finding {
    groupID := generateGroupID(group)
    findings := make([]finding.Finding, 0, len(group.Clones))

    for _, clone := range group.Clones {
        f := finding.NewFinding(
            "clone-detected",
            "art-dupl",
            fmt.Sprintf("Duplicate of %s (%d lines, %d%% similar)",
                group.Clones[0].File, group.LineCount, group.Similarity),
            finding.SeverityWarning,
            finding.Position{
                File:   clone.File,
                Line:   clone.StartLine,
                Column: clone.StartCol,
            },
            finding.Confidence(group.Similarity / 100.0),
        )

        f.Category = finding.CategoryDuplication
        f.GroupID = groupID
        f.Range = &finding.Range{
            Start: f.Position,
            End: finding.Position{
                File:   clone.File,
                Line:   clone.EndLine,
                Column: clone.EndCol,
            },
        }
        f.Metadata = map[string]string{
            "go-finding/detection-method": group.Method,
            "go-finding/clone-line-count": strconv.Itoa(group.LineCount),
        }

        // Link to other clones in the group
        for _, other := range group.Clones {
            if other.File != clone.File || other.StartLine != clone.StartLine {
                f.Related = append(f.Related, finding.RelatedRef{
                    FindingID: finding.GenerateID("art-dupl", "clone-detected", finding.Position{
                        File: other.File, Line: other.StartLine, Column: other.StartCol,
                    }),
                    Relation: finding.RelationCloneOf,
                    Position: finding.Position{
                        File: other.File, Line: other.StartLine, Column: other.StartCol,
                    },
                    Range: &finding.Range{
                        Start: finding.Position{File: other.File, Line: other.StartLine},
                        End:   finding.Position{File: other.File, Line: other.EndLine},
                    },
                })
            }
        }

        findings = append(findings, f)
    }

    return findings
}
```

---

## Appendix: What We Evaluated But Rejected

| Idea | Why rejected |
|------|-------------|
| Make art-dupl depend on go-finding as internal representation | Domain mismatch: go-finding models *findings* (single issues), art-dupl models *clone groups* (multi-location relationships). Forcing clone groups through a finding-shaped pipe would fight the grain. |
| Use `Finding.Metadata` for everything | Works but loses type safety, filtering, sorting, and SARIF round-trip. Metadata is for tool-specific extension, not core domain data. |
| Build an LSP server in art-dupl | Overkill. art-dupl should produce diagnostics, not serve them. go-finding's `ToLSP()` is the right integration point. |
| `Properties map[string]any` on `Finding` | Already decided and documented in AGENTS.md — `map[string]string` is the right choice for interchange simplicity. |

---

## Conclusion

go-finding is 80% ready. Three small additions (`RelatedRef.Range`, `GroupID`, `DiagnosticTag`) close the gap to 95%. The remaining 5% (snippet in SARIF, per-relationship metadata) can be deferred without blocking art-dupl integration.

The key insight: **go-finding should remain a *consumer* library that art-dupl outputs to, not a shared internal representation.** This preserves both projects' domain clarity while enabling clean interchange via `Finding`, SARIF, and LSP.
