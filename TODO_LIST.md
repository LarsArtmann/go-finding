# TODO List

**Updated:** 2026-07-22

Short-term, actionable work only. Completed items live in [CHANGELOG.md](CHANGELOG.md).
Long-term ideas live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

- [ ] **Sync `version.go` to v1.2.1** — `VersionPatch` still `0` but tag `v1.2.1` exists at `8405ba8`. Same class of bug as the v1.1.0 incident (shipped binary reported wrong version). Add CI check comparing `git describe --tags` against `finding.Version`.
- [ ] **Re-run `nix run .#test-race`** — The race detector was never re-run after the metrics double-recording fix (`0f39e10`). The gap-closure self-assessment (F01) explicitly flagged this as the highest-risk unverified item.

---

## 🟡 MEDIUM Priority

- [ ] **CI: dependabot `groups:` config** — Group gomod + github-actions updates to reduce PR noise. Currently each dep gets a separate PR.
- [ ] **CI: SHA-pin GitHub Actions** — Actions are pinned to major version tags (`@v7`), not commit SHAs. Supply-chain hardening.
- [ ] **CI: CODECOV_TOKEN or OIDC** — Codecov upload may be silently failing without a token configured.
- [ ] **Fix BuildFlow auto-configure loop** — **BLOCKED** (external tool). BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact. `golangci-lint run ./...` directly reports 0 issues on all 4 modules.

---

## 🟢 LOW Priority

- [ ] **SARIF schema validation test** — **BLOCKED** (requires vendoring 7K+ line SARIF 2.1.0 JSON schema).
- [ ] **Consumer compatibility test** — Verify downstream projects (20 known consumers) compile against latest release.

---

## DEFERRED v2.0 (breaking changes)

Structural changes that must batch into v2.0. Tracked here, not in ROADMAP, because they have concrete designs. From [data-model review](docs/reviews/2026-07-18_21-10_data-model-review.html).

- [ ] Redesign `Position` sentinel conventions — `position.go:23-29` mixes three conventions (0=unset for Line/Column, -1=unset for Offset, zero-value `Position{}` has Offset=0 = valid). Adopt `Option[T]` generic helper for one convention. Cascades into `Range`, `Finding`, `RelatedRef`.
- [ ] Redesign `FixStrategy` as interface-based closed union — `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved`. Eliminates `NormalizeFixStrategy` workaround; "Direct requires BeforeCode" becomes constructor invariant.
- [ ] Cleanup pointer-as-state fields — `Range *Range`, `Suppression *Suppression`, `Suppression.ExpiresAt *time.Time`, `RelatedRef.Range *Range`, `FindingError.Finding *Finding` all encode three states (nil/zero/valid). Replace with value+bool pairs.
- [ ] Convert `Tags []Tag` to `TagSet map[Tag]struct{}` — `finding.go:21`. Encodes set semantics at type level; eliminates order-insensitive equality in `finding_equal.go`.
- [ ] Compose `Finding` from embedded sub-structs — `Identity{ID,Rule,ToolName}`, `Location{Position,Range}`, `Classification{Category,Tags,Confidence}`, `Fix{FixStrategy,Suggestion,BeforeCode,AfterCode}`. Allows passing substructs to focused functions. Changes JSON shape — must batch with other v2.0 breaks.

---

_Assisted-by: Crush <crush@charm.land>_
