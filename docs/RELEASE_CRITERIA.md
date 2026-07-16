# v1.0.0 Release Criteria

**Last updated:** 2026-07-16
**Released as:** 1.0.0 (project now at 1.2.0)
**Status:** ✅ RELEASED — historical record, criteria met for v1.0.0

---

## Must Have (All Complete)

| Criterion                                         | Status | Pass/Fail Threshold                                      |
| ------------------------------------------------- | ------ | -------------------------------------------------------- |
| Core types stable                                 | ✅     | No breaking changes for 2+ minor versions                |
| SARIF 2.1.0 full round-trip                       | ✅     | `sarif_roundtrip_test.go` green                          |
| LSP Diagnostic conversion                         | ✅     | `lsp_test.go` green                                      |
| go/analysis integration                           | ✅     | `analysis/*_test.go` green                               |
| Pipeline: detect → triage → fix → verify          | ✅     | `pipeline/*_test.go` green                               |
| Byte-level FixEngine with provider chain          | ✅     | Benchmarks within baseline                               |
| Conflict detection (position + byte-level)        | ✅     | —                                                        |
| Report merging + deduplication (ID/position/rule) | ✅     | —                                                        |
| Cross-tool correlation                            | ✅     | IntervalIndex-backed                                     |
| Diff (before/after comparison)                    | ✅     | —                                                        |
| Test coverage ≥ 90% all packages                  | ✅     | Root 92.3%, Pipeline 95.2%, CLI 91.2%                    |
| Zero lint warnings                                | ✅     | `GOEXPERIMENT=jsonv2 nix run .#lint` exits 0             |
| Race detector clean                               | ✅     | `GOEXPERIMENT=jsonv2 go test -race -count=1 ./...` green |
| Comprehensive godoc examples (29+)                | ✅     | `Example*` functions run                                 |
| Structured errors with category classification    | ✅     | —                                                        |
| Builder pattern for Finding construction          | ✅     | —                                                        |
| Context cancellation throughout pipeline          | ✅     | —                                                        |
| Per-detector timeouts                             | ✅     | —                                                        |
| Graceful degradation on detector failures         | ✅     | —                                                        |
| Generated file filtering (gogenfilter)            | ✅     | —                                                        |
| Customizable TriageFunc                           | ✅     | —                                                        |
| ByteLevelConflictDetection opt-in                 | ✅     | —                                                        |
| CI/CD pipeline (GitHub Actions)                   | ✅     | `ci.yml` green on master                                 |
| Finding JSON schema                               | ✅     | `docs/schemas/finding.schema.json`                       |
| Streaming merge (MergeIter)                       | ✅     | —                                                        |
| Plugin architecture (DetectorRegistry)            | ✅     | —                                                        |
| LineShiftMap (byte-offset-aware shift tracking)   | ✅     | —                                                        |
| GoASTProvider (AST-aware fix provider)            | ✅     | `pipeline/goast/`                                        |

---

## Release Blockers — RESOLVED ✅

All three blockers have been resolved. The decisions and implementations are
documented below for reference.

### Blocker 1: Position zero-value semantics (#1) — RESOLVED ✅

**Decision:** Option A — use `-1` sentinel for "unset" offset.

**Implementation:**

- `Position.IsZero()` now checks `Offset < 0` instead of `Offset == 0`
- All constructors (Pos, NewRange, FromLSP, SARIF import) set `Offset: -1`
- `Position{}` (zero value) has `Offset=0` = "byte 0" (valid) = `IsZero()=false`
- `HasOffset()` unchanged (`>= 0`)
- No more contradiction between IsZero() and HasOffset()

**Alternatives:** Option B (`*int` pointers — nil-safe but nil-deref risk), Option C
(separate `HasOffset bool` — explicit but struct grows).

**Cascades to:** `Range.End` (same ambiguity), `Position.Line == 0`.

### Blocker 2: FixStrategy "" vs FixStrategyNone (#24) — RESOLVED ✅

**Decision:** Normalize `""` to `FixStrategyNone` via `NormalizeFixStrategy()`.

**Implementation:**

- `NormalizeFixStrategy()` helper added to fix_strategy.go
- Called by `Builder.Build()`, `findingFromSarResult()`, and `Finding.Equal()`
- `Validate()` treats `""` as valid but normalizes to `"none"`
- One canonical "no fix" state: `FixStrategyNone`

### Blocker 3: Report.Findings unexport timing (#25)

**Problem:** `Report.Findings` is a public slice. External code can bypass the
mutex (encapsulation risk). Internal migration is done (`findingsLocked()`,
`readFindings()`, `FindingsSnapshot()`).

**Recommendation:** Unexport to `findings` in v1.0.0. Consumers use
`FindingsSnapshot()`, `All()`, `FindByID()`, `AddFinding()`. This is the
documented migration path (ADR #11).

**Status:** Internal migration complete. Unexport scheduled for v1.0.0 tag.

---

## Should Have

| Criterion                         | Status | Notes                             |
| --------------------------------- | ------ | --------------------------------- |
| API stability guarantee document  | ✅     | `docs/API_STABILITY.md`           |
| Real-world tool integration guide | ✅     | `docs/integration-guide.md`       |
| USAGE_GUIDE.md complete           | ✅     | `docs/USAGE_GUIDE.md`             |
| v1.0 Migration Guide              | ✅     | `docs/MIGRATION_v1.0.md`          |
| Benchmark regression CI gate      | ✅     | `benchmark` CI job with benchstat |
| Dependabot for dependency updates | ✅     | `.github/dependabot.yml`          |

---

## Deprecated API Removal — COMPLETED ✅

All deprecated APIs have been removed in v1.0.0:

| API                               | Replacement                 | Action Taken   |
| --------------------------------- | --------------------------- | -------------- |
| `Report.Findings` (public field)  | `Report.FindingsSnapshot()` | **Unexported** |
| `Report.Merge()`                  | `Report.MergeInto()`        | **Removed**    |
| `Config.OnStage` callback         | `Config.StageHooks`         | **Removed**    |
| `Metrics.RecordFix()`             | `Metrics.RecordFixes(1)`    | **Removed**    |
| `CountBySeverity()` free function | `Report.CountBySeverity()`  | **Removed**    |
| `SeverityAliases()`               | `LookupSeverityAlias()`     | **Removed**    |
| `GetCategory()`                   | `CategoryOf()`              | **Removed**    |
| `HasFix` free function            | `WithFix`                   | **Removed**    |
| `HasSuggestion` free function     | `WithSuggestion`            | **Removed**    |

---

## Minimum Bar for v1.0.0 — ALL MET ✅

1. ✅ All "Must Have" items complete
2. ✅ All three OWNER DECISION blockers resolved
3. ✅ All deprecated APIs removed
4. ✅ No breaking changes after release without major version bump
5. ✅ CI/CD running green (test, lint, race, benchmark)
6. ✅ API stability guarantee published (`docs/API_STABILITY.md`)
7. ✅ Migration guide available (`docs/MIGRATION_v1.0.md`)
