# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

| Task                                        | Status    | Impact | Effort | Evidence                                                                                                                                |
| ------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| Tag `v1.3.0` release                        | 🔴 `TODO` | High   | 5min   | `version.go` says 1.3.0 but no git tag exists (`git tag -l 'v1.3*'` is empty). Lint clean, tests pass, GOWORK=off verified.             |

## 🟡 MEDIUM Priority

| Task                                   | Status       | Impact | Effort | Evidence                                                                                                                                                                 |
| -------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Fix BuildFlow auto-configure loop      | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact                                  |

## 🟢 LOW Priority

| Task                         | Status       | Impact | Effort | Evidence                                                                             |
| ---------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                  |
| Consumer compatibility test  | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

## DEFERRED v2.0 (breaking changes)

Structural changes that must batch into v2.0. Tracked here, not in ROADMAP, because they have concrete designs. From [data-model review](docs/reviews/2026-07-18_21-10_data-model-review.html).

| Task                                                   | Status    | Evidence                                                                                                                             |
| ------------------------------------------------------ | --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| Redesign `Position` sentinel conventions               | 🔴 `TODO` | `position.go:23-29`: 0=unset for Line/Column, -1=unset for Offset, zero-value has Offset=0 = valid. Adopt `Option[T]` generic helper |
| Redesign `FixStrategy` as interface-based closed union | 🔴 `TODO` | `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved`                              |
| Cleanup pointer-as-state fields                        | 🔴 `TODO` | `Range *Range`, `Suppression *Suppression`, `ExpiresAt *time.Time`, `RelatedRef.Range *Range` all encode 3 states (nil/zero/valid)   |
| Convert `Tags []Tag` to `TagSet map[Tag]struct{}`      | 🔴 `TODO` | `finding.go:21`. Eliminates order-insensitive equality in `finding_equal.go`                                                         |
| Compose `Finding` from embedded sub-structs            | 🔴 `TODO` | `Identity{}`, `Location{}`, `Classification{}`, `Fix{}`. Changes JSON shape — must batch                                             |

## Completed This Session

| Task                                        | Resolution                                                                                     |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| Fix 7 lint issues from v1.3.0 session       | ✅ All 7 fixed: 4 exhaustruct nolints, 2 gosec nolints, FindingTemplate→Template rename         |
| Run `GOWORK=off` per-module isolation tests | ✅ All 4 modules pass with GOWORK=off GOEXPERIMENT=jsonv2                                       |
| Decide on FormatText behavioral change      | ✅ Option B: FormatText reverted to `[SEVERITY]` format, FormatTextRich added for emoji badges |
| Doc accuracy fixes                          | ✅ ADR #9 r.Findings, DOMAIN_LANGUAGE v1.3.0 terms, API_STABILITY v1.3.0, CONTRIBUTING tree    |

---

_Assisted-by: Crush <crush@charm.land>_
