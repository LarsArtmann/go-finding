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
