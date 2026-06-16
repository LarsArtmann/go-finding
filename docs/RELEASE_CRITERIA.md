# v1.0.0 Release Criteria

**Last updated:** 2026-06-17
**Current version:** 0.7.0

---

## Must Have (All Complete)

| Criterion                                         | Status | Pass/Fail Threshold                       |
| ------------------------------------------------- | ------ | ----------------------------------------- |
| Core types stable                                 | ✅     | No breaking changes for 2+ minor versions |
| SARIF 2.1.0 full round-trip                       | ✅     | `sarif_roundtrip_test.go` green           |
| LSP Diagnostic conversion                         | ✅     | `lsp_test.go` green                       |
| go/analysis integration                           | ✅     | `analysis/*_test.go` green                |
| Pipeline: detect → triage → fix → verify          | ✅     | `pipeline/*_test.go` green                |
| Byte-level FixEngine with provider chain          | ✅     | Benchmarks within baseline                |
| Conflict detection (position + byte-level)        | ✅     | —                                         |
| Report merging + deduplication (ID/position/rule) | ✅     | —                                         |
| Cross-tool correlation                            | ✅     | IntervalIndex-backed                      |
| Diff (before/after comparison)                    | ✅     | —                                         |
| Test coverage ≥ 90% all packages                  | ✅     | Root 92.3%, Pipeline 95.2%, CLI 91.2%     |
| Zero lint warnings                                | ✅     | `nix run .#lint` exits 0                  |
| Race detector clean                               | ✅     | `go test -race -count=1 ./...` green      |
| Comprehensive godoc examples (29+)                | ✅     | `Example*` functions run                  |
| Structured errors with category classification    | ✅     | —                                         |
| Builder pattern for Finding construction          | ✅     | —                                         |
| Context cancellation throughout pipeline          | ✅     | —                                         |
| Per-detector timeouts                             | ✅     | —                                         |
| Graceful degradation on detector failures         | ✅     | —                                         |
| Generated file filtering (gogenfilter)            | ✅     | —                                         |
| Customizable TriageFunc                           | ✅     | —                                         |
| ByteLevelConflictDetection opt-in                 | ✅     | —                                         |
| CI/CD pipeline (GitHub Actions)                   | ✅     | `ci.yml` green on master                  |
| Finding JSON schema                               | ✅     | `docs/schemas/finding.schema.json`        |
| Streaming merge (MergeIter)                       | ✅     | —                                         |
| Plugin architecture (DetectorRegistry)            | ✅     | —                                         |
| LineShiftMap (byte-offset-aware shift tracking)   | ✅     | —                                         |
| GoASTProvider (AST-aware fix provider)            | ✅     | `pipeline/goast/`                         |

---

## Release Blockers (OWNER DECISIONS REQUIRED)

These three items require an explicit decision from the project owner. They are
irreversible breaking changes that affect every consumer. **Pick a direction
before tagging v1.0.0.**

### Blocker 1: Position zero-value semantics (#1)

**Problem:** `Position.Offset = 0` means both "byte offset 0" (valid) and "unset"
(the zero value). `Position{}.HasOffset()` returns `true` while
`Position{}.IsZero()` also returns `true`. The struct lies about its own state.

**Recommendation:** Option A — use `-1` sentinel for "unset" offset.

- Pro: least invasive, no field-type change
- Con: negative offset is a footgun; `uint` can't represent it; needs validation

**Alternatives:** Option B (`*int` pointers — nil-safe but nil-deref risk), Option C
(separate `HasOffset bool` — explicit but struct grows).

**Cascades to:** `Range.End` (same ambiguity), `Position.Line == 0`.

### Blocker 2: FixStrategy "" vs FixStrategyNone (#24)

**Problem:** Two valid "no fix" states: the empty string `""` (zero value) and
`FixStrategyNone` (`"none"`). `Finding.Validate()` accepts both. This is a type smell.

**Recommendation:** Normalize `""` to `FixStrategyNone` in `Validate()` or
normalize on construction in the Builder. Document that `""` is treated as `"none"`.

### Blocker 3: Report.Findings unexport timing (#25)

**Problem:** `Report.Findings` is a public slice. External code can bypass the
mutex (encapsulation risk). Internal migration is done (`findingsLocked()`,
`readFindings()`, `FindingsSnapshot()`).

**Recommendation:** Unexport to `findings` in v1.0.0. Consumers use
`FindingsSnapshot()`, `All()`, `FindByID()`, `AddFinding()`. This is the
documented migration path (ADR #11).

---

## Should Have

| Criterion                           | Status | Notes                             |
| ----------------------------------- | ------ | --------------------------------- |
| API stability guarantee document    | ✅     | `docs/API_STABILITY.md`           |
| Real-world tool integration guide   | ✅     | `docs/integration-guide.md`       |
| USAGE_GUIDE.md complete for v0.7.0+ | 🟡     | Needs v0.7.0 feature additions    |
| v1.0 Migration Guide                | ✅     | `docs/MIGRATION_v1.0.md`          |
| Benchmark regression CI gate        | ✅     | `benchmark` CI job with benchstat |
| Dependabot for dependency updates   | ✅     | `.github/dependabot.yml`          |

---

## Deprecated API Removal Timeline (v1.0.0)

These APIs are deprecated and scheduled for removal in v1.0.0:

| API                               | Replacement                       | Deprecated Since | v1.0 Action  |
| --------------------------------- | --------------------------------- | ---------------- | ------------ |
| `Report.Findings` (public field)  | `Report.FindingsSnapshot()`       | v0.7.0           | **Unexport** |
| `Report.Merge()`                  | `Report.MergeInto()`              | v0.6.0           | **Remove**   |
| `OnStage` callback                | `StageHooks`                      | v0.7.0           | **Remove**   |
| `Metrics.RecordFix()`             | `Metrics.RecordFixes(1)`          | v0.6.0           | **Remove**   |
| `CountBySeverity()` free function | `Report.CountBySeverity()` method | v0.6.0           | **Remove**   |

---

## Minimum Bar for v1.0.0

1. All "Must Have" items complete ✅
2. All three OWNER DECISION blockers resolved
3. All deprecated APIs removed
4. No breaking changes after release without major version bump
5. CI/CD running green (test, lint, race, benchmark)
6. API stability guarantee published
7. Migration guide available for consumers
