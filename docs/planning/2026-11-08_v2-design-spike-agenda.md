# v2.0 Design Spike Agenda

**Purpose:** one session, whiteboard-first, no code. Each item is a
known-friction point where v1's type system compensates for a modeling
decision made under v1 constraints. The spike DECIDES whether v2 changes
the model, keeps it, or explicitly rejects the change (a written no is a
result).

**Input basis:** ROADMAP Hardening section + the v1.15.x train lessons +
ADR #20's costs. Order = expected contention (hardest decisions first,
while the session is fresh).

## 1. Position.Offset sentinel (-1)

- Today: `Position{}` zero value means byte 0; constructors set Offset=-1
  for "unset"; `HasOffset()` disambiguates. Three states encoded in one
  field.
- Spike: split `Offset *int` (nil = unset) vs a dedicated
  `PositionKind` enum vs keeping the sentinel. Constraint: SARIF
  byte-offset round-trip must stay lossless (offset 0 survives omitempty
  via pointer fields today — check the SARIF wire cost of each option).
- Decide: which representation, and does `Position` itself split into
  `FilePosition` / `BytePosition` variants.

## 2. FixStrategy union

- Today: `FixStrategy` is a string type with `NormalizeFixStrategy()`
  ("" → "none"), `FixStrategyAI` reserved-but-unimplemented, and
  `HasFix()` branching on the strategy. The string union invites
  unrepresentable combinations (AI + direct edits).
- Spike: sealed-sum emulation (struct with Kind + per-kind payload) vs
  typed constants + behavior methods vs status quo. Constraint: JSON
  compatibility for v1 consumers (wire shape stability across the major
  bump), SARIF property bag encoding.

## 3. TagSet (dedup + ordering by construction)

- Today: `Tags []Tag` with `tagsEqual` fast path + clone+sort fallback in
  `Equal()`; order-independent equality means every comparison pays a
  sort eventually.
- Spike: normalized-on-construction TagSet (sorted, deduped at
  `WithTags`), cost model for construction-heavy vs comparison-heavy
  workloads. Decide whether GroupBy/Correlate benchmarks justify it.

## 4. Sub-struct extraction (Finding is 20+ fields)

- Today: identity (ID/Rule/Tool), placement (Position/Range/Snippet),
  remediation (FixStrategy/Before/After/Edits/Suggestion), provenance
  (Category/Tags/Metadata/GroupID) live flat on one struct. Every field
  addition re-opens padding/benchmark questions (the v1.7.0 and issue-#36
  regenerations).
- Spike: does extraction actually reduce friction, or does it just move
  the flat-field problem into accessor naming? Constraint: JSON field
  stability, the builder chain ergonomics, and Equal's 16-field cascade.

## 5. Explicit "unknown" for coverage (ADR #20 residue)

- Today: v1 detectors yield FilesScanned=0, which JSON cannot distinguish
  from a verified empty scan. ADR #20 documented the ambiguity and
  rejected the schema break for v1.
- Spike: `FilesScanned *int` (nil = unknown) or a `Coverage` sub-struct.
  This is the one item where v2's major-version break buys a real
  semantic fix.

## Anti-agenda

- NO discussion of new features (pipeline stages, providers, output
  formats) — v2.0's scope is the type system debt only.
- NO migration-tooling design until items 1-5 have verdicts; the
  migration cost shapes which verdicts are affordable.
- Every item ends with one of: CHANGE (with wire-compat plan), KEEP
  (with the friction accepted and documented), REJECT (with the reason
  written down). A spike that ends "we'll think about it again" is a
  failed spike.
