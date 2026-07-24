# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> Completed items live in [CHANGELOG.md](CHANGELOG.md).
> Long-term ideas and deferred breaking changes live in [ROADMAP.md](ROADMAP.md).

---

## 🔴 HIGH Priority

_No high-priority items. v1.3.0 is released; see [CHANGELOG.md](CHANGELOG.md)._

## 🟡 MEDIUM Priority

| Task                       | Status       | Impact | Effort | Evidence                                                                                    |
| -------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------- |
| Fix BuildFlow auto-configure loop | 🔵 `BLOCKED` | Med    | —      | External tool. BuildFlow's detect→repair cycle re-triggers golangci-lint per-module, reporting "2 findings" that are a scoring artifact |

## 🟢 LOW Priority

| Task                        | Status       | Impact | Effort | Evidence                                                                             |
| --------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------ |
| SARIF schema validation test | 🔵 `BLOCKED` | Low    | —      | Requires vendoring 7K+ line SARIF 2.1.0 JSON schema                                 |
| Consumer compatibility test | 🔵 `BLOCKED` | Low    | —      | Repo is private; consumers need `GOPRIVATE` set. 22 known consumers, 14 with Go code |

---

_The following deferred breaking changes have concrete designs and are tracked in [ROADMAP.md](ROADMAP.md) under "Hardening (owner decisions pending)": Position sentinel redesign, FixStrategy closed union, pointer-as-state cleanup, Tags→TagSet, Finding sub-struct composition._

---

_Assisted-by: Crush <crush@charm.land>_
