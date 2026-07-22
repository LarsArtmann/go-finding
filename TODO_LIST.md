# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

| Task | Status | Impact | Effort | Evidence |
|------|--------|--------|--------|----------|
| Fix 7 lint issues from v1.3.0 session | 🔴 `TODO` | High | 20min | `golangci-lint run ./...` fails: 4 exhaustruct, 2 gosec, 1 revive. Add `//nolint` directives and rename `FindingTemplate` to `Template` |
| Run `GOWORK=off` per-module isolation tests | 🔴 `TODO` | High | 15min | Plan F16.1-F16.4 never executed. Run `GOWORK=off GOEXPERIMENT=jsonv2 go test ./...` in each of the 4 module dirs |
| Commit remaining 5 uncommitted files | 🔴 `TODO` | High | 5min | `git status` shows `category_linter.go`, `format_test.go`, `report_test.go`, `simple_fix.go`, `simple_fix_test.go` unstaged |

## 🟡 MEDIUM Priority

| Task | Status | Impact | Effort | Evidence |
|------|--------|--------|--------|----------|
| Tag `v1.3.0` release | 🔴 `TODO` | Med | 5min | `version.go` says 1.3.0 but no git tag exists (`git tag -l 'v1.3*'` is empty) |
| Decide on FormatText behavioral change | 🔴 `TODO` | Med | — | `FormatText` output changed from `[ERROR]` to `🟠 ERROR`. Needs decision: keep, revert + add `FormatTextRich()`, or add options. See `docs/status/2026-07-22_18-55_*` Q1 |
| Fix BuildFlow auto-configure loop | 🔵 `BLOCKED` | Med | — | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |

## 🟢 LOW Priority

| Task | Status | Impact | Effort | Evidence |
|------|--------|--------|--------|----------|
| SARIF schema validation test | 🔵 `BLOCKED` | Low | — | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema |
| Consumer compatibility test | 🔵 `BLOCKED` | Low | — | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

## DEFERRED v2.0 (breaking changes)

Structural changes that must batch into v2.0. Tracked here, not in ROADMAP, because they have concrete designs. From [data-model review](docs/reviews/2026-07-18_21-10_data-model-review.html).

| Task | Status | Evidence |
|------|--------|----------|
| Redesign `Position` sentinel conventions | 🔴 `TODO` | `position.go:23-29`: 0=unset for Line/Column, -1=unset for Offset, zero-value has Offset=0 = valid. Adopt `Option[T]` generic helper |
| Redesign `FixStrategy` as interface-based closed union | 🔴 `TODO` | `type Fix interface { isFix() }` with `NoFix`, `Suggestion{Text}`, `Direct{Before,After}`, `AIReserved` |
| Cleanup pointer-as-state fields | 🔴 `TODO` | `Range *Range`, `Suppression *Suppression`, `ExpiresAt *time.Time`, `RelatedRef.Range *Range` all encode 3 states (nil/zero/valid) |
| Convert `Tags []Tag` to `TagSet map[Tag]struct{}` | 🔴 `TODO` | `finding.go:21`. Eliminates order-insensitive equality in `finding_equal.go` |
| Compose `Finding` from embedded sub-structs | 🔴 `TODO` | `Identity{}`, `Location{}`, `Classification{}`, `Fix{}`. Changes JSON shape — must batch |

---

_Assisted-by: Crush <crush@charm.land>_
