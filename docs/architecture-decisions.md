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
