# Architecture Decisions

**Status:** Open for review — no action required unless you disagree.

These decisions need product input. They are documented here for visibility.

---

## 1. FixStrategyAI Semantics

**Current behavior:** `FixStrategyAI` behaves identically to `FixStrategySuggest`:

- `HasFix()` returns true only if `AfterCode != ""`
- Pipeline triage treats it as a suggestion (no auto-apply)
- `NeedsAI()` returns true (distinguishing method)

**Question:** Should `FixStrategyAI` mean:

- **(A)** "An AI could fix this" (capability) → `HasFix()` should always return true
- **(B)** "An AI has generated a suggested fix" (artifact) → current behavior is correct

**Recommendation:** Keep (B). It matches the existing SARIF round-trip behavior and doesn't over-promise.

---

## 2. Stable ID Format

**Current:** `tool:rule:file:line:col` (human-readable)

**Options:**

- (A) Keep readable strings (current)
- (B) SHA-256 hash (collision-resistant but opaque)
- (C) Both: readable format with optional hash suffix

**Recommendation:** Keep (A) for v1. Add hash in v2 if collisions become a problem.

---

## 3. Repository Name

**Current:** `go-finding` (GitHub) / `finding` (Go module path: `github.com/larsartmann/go-finding`)

**Options:** `finding`, `finding-sdk`, `go-finding`

**Recommendation:** Keep `go-finding`. It's clear, unique, and follows Go naming conventions.

---

## 4. Suppression Expiry

**Current:** `Suppression.ExpiresAt` field exists but is not enforced.

**Question:** Should the pipeline skip expired suppressions automatically?

**Recommendation:** Defer to v1.1. Add `IsActive()` method that checks expiry, but don't auto-filter.

---

## 5. API Stability for v1.0.0

**Current:** v0.2.1 — API-stable beta. All three conditions below are resolved.

**Status:** All blocking decisions resolved:

- [x] FixStrategyAI decision (#1 above) — Resolved: keep as artifact (B)
- [x] Builder.Build() error return is stable — Returns `(Finding, error)` since v0.2.0
- [x] ID format is finalized (#2 above) — Resolved: keep readable format (A) for v1

**Next target:** v1.0.0 = production release with API stability guarantee.

---

## 6. NewFinding API Pattern

**Decision:** Keep the current 6-parameter constructor + Builder pattern.

**Rationale:**

- `NewFinding(rule, toolName, message, severity, pos, confidence)` provides required fields with auto-generated ID and clamped confidence.
- `NewBuilder(rule, toolName, message, severity, pos)` provides a fluent API for optional fields.
- Functional options were considered but rejected — they add allocation and complexity for no ergonomics gain over the builder.
- The builder pattern is idiomatic Go and already in use.

**Status:** Resolved — no changes needed.

---

## 7. Domain-Specific FixProvider Location

**Decision:** Domain-specific providers (Go AST, Rust syn, etc.) should live in **separate modules** (`pipeline/fix/` is not appropriate for multi-language providers).

**Rationale:**

- Providers are domain-specific and would import heavy dependencies (go/ast, Rust parser, etc.).
- Separate modules allow consumers to import only what they need.
- The `FixProvider` interface in `pipeline/` is the contract; implementations live outside.
- Example: `github.com/larsartmann/go-finding-provider-goast` as a separate repo.

**Status:** Resolved — providers live in separate modules.

---

## 8. Properties map[string]any — Rejected

**Decision:** NOT adding `Properties map[string]any` to the `Finding` struct. `Metadata map[string]string` is the sole extensibility field.

**Rationale:**

- `map[string]string` is fully typed, lossless for interchange (SARIF, JSON, CLI, env vars).
- `map[string]any` requires type assertions at every read site, violating type safety.
- Complex values can be JSON-serialized into string values.
- This decision preserves the simplicity and type safety of the struct.
- See commit `024b6a3` for the detailed rationale.

**Status:** WONTFIX — intentionally rejected.

---

## 9. go-sarif Library vs Hand-Rolled SARIF

**Question:** Should we replace the hand-rolled SARIF implementation with `github.com/owenrumney/go-sarif/v3`?

**Current state:** Hand-rolled SARIF across 3 files (~470 LOC):
- `sarif_types.go` — 12 struct types, constants, severity conversion
- `sarif_export.go` — Export: `ToSARIF`, `WriteSARIF`, `WriteSARIFFiltered`, `WriteTo`
- `sarif_import.go` — Import: `FindingsFromSARIF`, `FindingsFromReader`

**go-sarif v3 assessment:**
- 83 stars, actively maintained (v3.3.0, Oct 2025)
- Full SARIF 2.1.0 + 2.2.0 spec support with validation
- Builder API (`NewRun`, `AddRule`, `CreateResultForRule`, etc.)
- `Open` / `FromBytes` / `FromString` parsing with strict validation option

**Comparison:**

| Dimension | Hand-rolled | go-sarif v3 |
|-----------|-------------|-------------|
| Extra dependencies | 0 | +1 module |
| Lines we maintain | ~470 | 0 (but adapter layer ~200-300) |
| Spec coverage | Subset we use | Full 2.1.0 + 2.2.0 |
| Schema validation | None | Built-in `Validate()` |
| go-finding round-trip | Native (property bag designed for it) | Requires adapter |
| Streaming I/O | Native (`json.Encoder`/`Decoder`) | Unknown |
| Context cancellation | First-class | Unknown |
| API shape | Finding-centric | SARIF-centric |
| Test coverage | 100% (unit + fuzz, 1.6M execs) | External |

**Why hand-rolled wins:**

1. **Dependency minimalism** — The project design principle #1 is "core types depend only on stdlib." Adding a SARIF dependency for ~470 LOC violates this.
2. **No adapter tax** — go-sarif's API is SARIF-centric (build runs, add rules, create results). Our API is Finding-centric (`report.ToSARIF()`). A clean adapter would add ~200-300 LOC of mapping code, negating most of the "maintenance savings."
3. **Round-trip fidelity is custom** — Our property bag namespace (`go-finding/*`) is purpose-built for Finding round-trip. Reimplementing this on go-sarif's generic `PropertyBag` is no simpler.
4. **Streaming + context** — Our `WriteSARIF(ctx, w)` and `FindingsFromReader(ctx, r)` are first-class streaming with context cancellation. go-sarif's API shape would require wrapping.
5. **We don't need the extra spec coverage** — We use: `Run`, `Tool`, `Driver`, `Result`, `Location`, `PhysicalLocation`, `ArtifactLocation`, `Region`, `Message`, `Fix`, `ArtifactChange`, `Replacement`, `RelatedLocation`, `PropertyBag`. That's 15 types. go-sarif exposes 100+ types we would never touch.

**When to reconsider:**

- If we need **SARIF 2.2.0** features (currently no demand)
- If we need **schema validation** in production (currently blocked by schema size, not implementation)
- If we need **code flows, graphs, or taxonomies** (not on roadmap)
- If the hand-rolled implementation grows beyond ~1000 LOC (suggests we're reimplementing too much)

**Recommendation:** Keep hand-rolled. The evaluation confirms the original decision was correct. Close the TODO.

**Status:** Resolved — keep hand-rolled.
