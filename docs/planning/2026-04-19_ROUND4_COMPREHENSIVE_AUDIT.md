# Round 4 — Comprehensive Audit Execution Plan

**Date:** 2026-04-19
**Scope:** All remaining issues from deep re-audit (37 total, de-duplicated to 25 tasks)
**Constraint:** Each task ≤ 12 minutes; commit after each; push at end

---

## Priority Sorting Rationale

For a **general-purpose Go SDK/library**, priority order is:

1. **Correctness** — Bugs producing silently wrong results
2. **Thread safety** — Race conditions in exported types
3. **API contracts** — Missing validation, broken round-trips
4. **Robustness** — Error handling, edge cases
5. **API quality** — Consistency, discoverability
6. **Polish** — Docs, naming, lint hints

---

## Execution Table

Sorted by: **Importance → Impact → Effort → Customer Value**

| # | Task ID | Description | File(s) | Severity | Impact | Effort | Type |
|---|---------|-------------|---------|----------|--------|--------|------|
| 1 | R4.01 | **FIX REGRESSION**: DeduplicateByPosition key identical to DeduplicateByRule | `merge.go` | P0 | ALL users get wrong dedup | 5min | bug |
| 2 | R4.02 | Add test verifying DeduplicateByPosition ≠ DeduplicateByRule | `merge_test.go` | P0 | Prevents future regression | 5min | test |
| 3 | R4.03 | Fix OnFix callback never called for successful fixes | `pipeline/pipeline.go` | P0 | Breaks fix tracking | 8min | bug |
| 4 | R4.12 | Fix Offset==0 treated as "not set" in Range.Length/Contains/Overlaps | `position.go` | P0 | Silent wrong results at file start | 10min | bug |
| 5 | R4.04 | Fix profiling FD leak when StartCPUProfile fails | `cmd/go-finding/main.go` | P0 | Resource leak on error path | 5min | bug |
| 6 | R4.05 | Add metrics recording to partial detection | `pipeline/partial.go` | P0 | Silent metric gap | 8min | bug |
| 7 | R4.06 | Document Pipeline.Run() not concurrent-safe (or add mutex) | `pipeline/pipeline.go` | P0 | Data race if misused | 8min | safety |
| 8 | R4.07 | Make Metrics map fields unexported with accessor methods | `pipeline/metrics.go` | P0 | Thread-unsafe exported maps | 10min | safety |
| 9 | R4.10 | Guard TotalDuration() against negative return | `pipeline/metrics.go` | P0 | Negative duration in reports | 5min | bug |
| 10 | R4.11 | Document Report.AddFinding not goroutine-safe | `report.go` | P1 | Race if used concurrently | 5min | safety |
| 11 | R4.08 | Add Config.Validate() method | `pipeline/pipeline.go` | P1 | Invalid configs cause silent misbehavior | 8min | API |
| 12 | R4.09 | Add RetryConfig.Validate() method | `pipeline/retry.go` | P1 | Negative retries, zero delays | 8min | API |
| 13 | R4.13 | Fix SARIF SeverityCritical lost in round-trip | `sarif.go` | P1 | Data loss on export/import | 5min | bug |
| 14 | R4.14 | FindingsFromJSON returns count of dropped findings | `json.go` | P1 | Silent data loss | 8min | API |
| 15 | R4.15 | Fix SARIF Location → Locations (plural array per spec) | `sarif.go` | P1 | SARIF 2.1.0 spec violation | 10min | spec |
| 16 | R4.16 | FixApplier rollback all files on partial failure | `pipeline/pipeline.go` | P1 | Partial state on error | 10min | robustness |
| 17 | R4.17 | Iteration.Findings() returns copy of internal slice | `pipeline/pipeline.go` | P1 | Caller can mutate internals | 5min | safety |
| 18 | R4.18 | Handle g.Wait() errors in detectPartialParallel | `pipeline/partial.go` | P1 | goroutine leaks on panic | 5min | robustness |
| 19 | R4.19 | Add JSON tags to Correlation struct | `merge.go` | P2 | Serialization gap | 3min | API |
| 20 | R4.20 | Remove FindingsFromSARIF always-error stub | `sarif.go` | P2 | Confusing API surface | 5min | API |
| 21 | R4.21 | Move test helpers to testutil_test.go | `testutil.go` | P2 | Exported from production package | 8min | API |
| 22 | R4.22 | Add NewFinding() constructor | `finding.go` | P2 | Ergonomics | 10min | API |
| 23 | R4.23 | Fix fileHash() doc (FNV-1a not FNV-128) | `pipeline/pipeline.go` | P3 | Misleading doc | 2min | docs |
| 24 | R4.25 | Inject version via ldflags instead of hardcoding | `cmd/go-finding/main.go` | P3 | Release process friction | 8min | polish |
| 25 | R4.24 | Clean up gopls hints (rangeint, newexpr, mapsloop, stringsseq) | Various | P3 | Code quality | 10min | polish |

**Total estimated effort:** ~2.5 hours

---

## Dependency Graph

```
R4.01 ──→ R4.02  (test depends on fix)
R4.07 ──→ R4.10  (both touch metrics.go)
R4.08 ── standalone
R4.15 ── may affect R4.13 (both in sarif.go, but different concerns)
```

All other tasks are independent and can be executed in listed order.

---

## Verification Protocol

After **each** task:

```bash
go build ./... && go test -race -count=1 ./...
```

After **all** tasks:

```bash
go vet ./... && git push
```

---

## Scope Decisions

### Included (SDK quality bar)
- All P0/P1 items — correctness, safety, API contracts
- P2 items that affect library consumers (JSON tags, stubs, test helpers, constructors)

### Excluded (future work)
- R4.24 (gopls hints) — non-functional, can be done in a separate cleanup PR
- SARIF reader implementation — large feature, out of scope for audit fixes
- CLI watch mode, config files — v2.0 features

---

_Assisted-by: Crush <crush@charm.land>_
