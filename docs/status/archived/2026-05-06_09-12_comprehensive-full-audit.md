# Status Report — Full Comprehensive Audit

**Date:** 2026-05-06 09:12 CET
**Session:** Multi-session status audit after byte-level FixEngine redesign + follow-up work
**Total commits:** 479 (25 since `0d00a20`)
**Unpushed:** 5 commits (`df82906`..`9107bfa`)
**Tests:** 985 passing, 0 failing (race-detector enabled)
**Lint:** 0 issues (golangci-lint), 9 LSP warnings (non-blocking)
**Build:** GREEN

---

## Project Health Dashboard

| Metric                | Value    | Trend                                               |
| --------------------- | -------- | --------------------------------------------------- |
| Production Go files   | 42       | +4 (analysis/ package, sarif split, pipeline split) |
| Test files            | 57       | +3                                                  |
| Production LOC        | 6,695    | +350 net                                            |
| Test LOC              | 16,553   | +1,100                                              |
| Test/Code ratio       | 2.47:1   | Improving                                           |
| Test count            | 985 PASS | +40 since last audit                                |
| Dependencies (direct) | 5        | Stable (ginkgo, gomega, yaml, x/sync, x/tools)      |
| Coverage — core       | 99.6%    | Stable                                              |
| Coverage — pipeline   | 98.0%    | Stable                                              |
| Coverage — CLI        | 95.4%    | Stable                                              |
| Coverage — analysis   | ~90%     | New package                                         |
| Coverage — detectors  | 96.1%    | Stable                                              |
| Lint issues           | 0        | Clean                                               |
| Go version            | 1.26.2   | Current                                             |

---

## a) FULLY DONE ✅

### Phase 1: Byte-Level FixEngine Redesign (5 commits)

| # | Commit    | Description                                                                 |
| - | --------- | --------------------------------------------------------------------------- |
| 1 | `52df52f` | Redesign FixEngine from string-based to byte-level with plugin architecture |
| 2 | `0d00a20` | FixEdit serialization, edit-level conflict detection, and doc updates       |
| 3 | `5081eae` | Fix temp dir leak — defer `FixApplier.Close()`                              |
| 4 | `dc3ca37` | Pipeline.Run() single-use contract in godoc                                 |
| 5 | `7eae850` | `strconv.Atoi` over `fmt.Sscanf`                                            |

### Phase 2: Performance & Safety (4 commits)

| # | Commit    | Description                                                      |
| - | --------- | ---------------------------------------------------------------- |
| 6 | `cba1b19` | O(1) line→byte offset index (`buildLineOffsetIndex`)             |
| 7 | `a3e4b58` | `ConflictInfo.ConflictsWith` populated via `edit.Overlaps(prev)` |
| 8 | `a91f948` | `Finding.Key()` for empty-ID conflict safety                     |
| 9 | `7eae850` | Performance: `strconv.Atoi` replaces `fmt.Sscanf`                |

### Phase 3: API Cleanup — Deprecation (2 commits)

| #  | Commit    | Description                                                               |
| -- | --------- | ------------------------------------------------------------------------- |
| 10 | `340f1f2` | Deprecate `ConflictDetector` → `DetectConflicts()` package-level function |
| 11 | `ad9c65a` | Deprecate `Verifier` → `Verify()` package-level function                  |

### Phase 4: File Organization (2 commits)

| #  | Commit    | Description                                                                             |
| -- | --------- | --------------------------------------------------------------------------------------- |
| 12 | `e97f62e` | Split `sarif.go` (570 lines) → `sarif_types.go` + `sarif_export.go` + `sarif_import.go` |
| 13 | `1bb1b1e` | Split `pipeline.go` (624 lines) → `pipeline.go` + `adapters.go` + `config.go`           |

### Phase 5: Testing & Benchmarks (2 commits)

| #  | Commit    | Description                                                           |
| -- | --------- | --------------------------------------------------------------------- |
| 14 | `c946025` | FixEngine benchmarks: 1/10/100/1000 fixes on 10k-line file            |
| 15 | `aa25f80` | 15 BDD specs for FixProvider contract (Offset/Line/Substring + chain) |

### Phase 6: SARIF Round-Trip (2 commits)

| #  | Commit    | Description                                                 |
| -- | --------- | ----------------------------------------------------------- |
| 16 | `e47c219` | Preserve `go-finding/edit/*` properties in SARIF round-trip |
| 17 | `9107bfa` | Fix byte offset test: offsets 29-34 (not 28-33)             |

### Phase 7: Code Formatting (1 commit)

| #  | Commit    | Description                                       |
| -- | --------- | ------------------------------------------------- |
| 18 | `48dfc8c` | Table alignment, parameter order, modernize loops |

### Phase 8: Documentation (2 commits)

| #  | Commit    | Description                                     |
| -- | --------- | ----------------------------------------------- |
| 19 | `b071665` | TODO_LIST.md — 11 items marked completed        |
| 20 | `686f803` | AGENTS.md — new file structure, deprecated APIs |

### Phase 9: Dependency Isolation (1 commit)

| #  | Commit    | Description                                                                          |
| -- | --------- | ------------------------------------------------------------------------------------ |
| 21 | `1d64c73` | Extract `diagnostic.go` → `analysis/` subpackage (12MB `x/tools` isolated from core) |

### Phase 10: Missing API (1 commit)

| #  | Commit    | Description                                         |
| -- | --------- | --------------------------------------------------- |
| 22 | `df82906` | `Report.Merge(other *Report)` in-place merge method |

### Previously Done (before byte-level redesign, verified)

- Version bump to v0.2.0, then v0.2.1
- FixStrategyAI split brain resolved
- Builder.Build() error return
- Tag → Tags []Tag migration
- Report.lock() mutex fix
- Correlate O(n²) capped at 10K
- FindingProcessor interface
- Architecture diagrams (D2 + mermaid)
- 44+ BDD specs across packages
- govulncheck in CI
- Fuzz tests (SARIF, JSON, ID, merge)
- Examples with compile checks
- CONTRIBUTING.md, release procedure, integration guide

---

## b) PARTIALLY DONE 🔧

### 1. CLI Refactoring (`cmd/go-finding/main.go` — 457 lines)

- **What exists:** Fully functional CLI with text/json/sarif output, YAML config, profiling
- **What's missing:** Split into `config.go` + `output.go`, refactor `run()` for testability (accept `io.Writer` + `*flag.FlagSet` instead of globals)
- **Impact:** P0 — blocking API stability review

### 2. Report Thread-Safety

- **What exists:** `Report` has `sync.Mutex` on `AddFinding`/`AddFindings`, `ComputeSummary`, `Merge`
- **What's missing:** Zero-value `Report{}` has nil mutex — silently non-thread-safe. Needs `sync.Once` lazy init or explicit documentation
- **Impact:** P0 — data race potential

### 3. Tag Deprecation Unification

- **What exists:** `Tag` field removed from Finding, `WithTag` builder method marked deprecated
- **What's missing:** Some test files may still use old patterns; inconsistent migration
- **Impact:** P1 — cleanup

---

## c) NOT STARTED 📋

### P0 — Must Do Before v1.0.0

| # | Task                              | Description                                                                                   |
| - | --------------------------------- | --------------------------------------------------------------------------------------------- |
| 1 | **Confidence strong type**        | `type Confidence float64` with validation. `Finding{Confidence: 1.5}` compiles but is invalid |
| 2 | **Centralize triage logic**       | `HasFix()`, `triage()`, `FixEngine.Apply()` have overlapping but different categorization     |
| 3 | **Properties map[string]any**     | `Metadata map[string]string` loses type info in SARIF round-trip                              |
| 4 | **API stability review**          | Audit every exported symbol for v1.0.0 lock                                                   |
| 5 | **Split cmd/go-finding/main.go**  | 457 lines → config.go + output.go                                                             |
| 6 | **Report thread-safety**          | sync.Once for nil mutex or clear documentation                                                |
| 7 | **Decide NewFinding API pattern** | Functional options vs builder-only vs current 6-param                                         |

### P1 — Should Do Before v1.0.0

| #  | Task                                     | Description                                             |
| -- | ---------------------------------------- | ------------------------------------------------------- |
| 8  | **Refactor CLI run()**                   | Accept `io.Writer` + `*flag.FlagSet`                    |
| 9  | **Wire FixProviders through CLI config** | Config.FixProviders in YAML/JSON                        |
| 10 | **io.WriterTo for SARIF**                | Direct streaming without buffer allocation              |
| 11 | **WriteSARIF error-path tests**          | failingWriter pattern, currently 75% coverage           |
| 12 | **Context-cancel tests**                 | detectPartial Sequential/Parallel cancel paths untested |
| 13 | **Error wrapping audit**                 | Ensure all internal errors use `%w`                     |
| 14 | **Unify Tag deprecation**                | Migrate remaining tests to `Tags []Tag`                 |
| 15 | **Protect Confidence construction**      | `Finding{Confidence: 1.5}` bypasses clamping            |
| 16 | **Decompose FindingsFromSARIF**          | Cognitive complexity 90 (threshold 35)                  |

### P2 — Nice to Have

| #  | Task                                 | Description                                   |
| -- | ------------------------------------ | --------------------------------------------- |
| 17 | **Document SARIF round-trip losses** | User-facing docs, not just code comments      |
| 18 | **Nix setup in CONTRIBUTING.md**     | Missing despite nix usage                     |
| 19 | **SARIF schema validation test**     | Verify output against SARIF 2.1.0 JSON schema |
| 20 | **Benchmark regression tracking**    | scripts/bench-compare.sh                      |
| 21 | **Pipeline benchmarks 10k+**         | Full pipeline scalability testing             |
| 22 | **Finding JSON schema**              | Formal contract for API consumers             |
| 23 | **Consumer migration guide**         | v0.1.3 → v0.2.0                               |
| 24 | **Category.IsValid() strict**        | Should validate against known categories      |
| 25 | **Evaluate go-sarif**                | Spec compliance assessment                    |

### P3 — Future / Deferred

| #  | Task                                  | Description                                       |
| -- | ------------------------------------- | ------------------------------------------------- |
| 26 | **Nix flake migration**               | Full justfile → flake.nix (6 phases)              |
| 27 | **Plugin architecture for detectors** | Runtime registration replacing hardcoded builders |
| 28 | **Pipeline middleware**               | Custom stage injection                            |
| 29 | **Structured logging (slog)**         | Replace fmt.Fprintf                               |
| 30 | **Styled CLI (lipgloss)**             | Terminal UI improvements                          |
| 31 | **Interactive TUI (bubbletea)**       | Fix review interface                              |
| 32 | **go/analysis reverse conversion**    | Finding → analysis.Diagnostic                     |
| 33 | **Finding.Diff()**                    | Compare finding sets                              |
| 34 | **Finding.FormatText/Markdown**       | Human-readable output formats                     |

---

## d) TOTALLY FUCKED UP 💥

### Nothing is broken. But here are the risks:

### 1. 🔴 LSP Shows Duplicate Declarations in Pipeline Package

```
pipeline/pipeline.go:17:6: Detector redeclared in this block
    pipeline/adapters.go:12:6: other declaration of Detector
```

**Actual state:** `pipeline.go` does NOT declare `Detector` — it only has `Pipeline` and `TriageResult` structs. The LSP (gopls) is showing stale diagnostics from an old cache. `go build ./...` passes clean. This is a **LSP cache issue**, not a code issue.

**Fix:** Restart gopls / LSP client to clear stale diagnostics.

### 2. 🟡 Orphaned `sarif.go` Ghost in LSP

```
sarif.go:1:9: No packages found for open file sarif.go
```

The file `sarif.go` was deleted in commit `e97f62e` and replaced by 3 split files. The LSP still has a ghost reference. Not a real issue — `go build` passes.

### 3. 🟡 gci and golines Formatters Unavailable

These formatters are enabled in `.golangci.yml` but broken in nixpkgs. `go install` is blocked by security policy. Currently `golangci-lint run` passes without them, but:

- `analysis/analysis.go` needs golines formatting
- `pipeline/fix_engine_bench_test.go` needs gci formatting

These are **cosmetic only** — no functional impact.

### 4. 🟡 Benchmark Performance Concern

FixEngine benchmarks show quadratic behavior:

| Fixes | Time  |
| ----- | ----- |
| 1     | 236ms |
| 10    | 2.4s  |
| 100   | 24.6s |
| 1000  | 286s  |

The substring matching in `SubstringProvider` is O(n\*m) where n=findings and m=file content. This is a known limitation — domain-specific providers (OffsetProvider, LineProvider) are O(n).

### 5. 🟡 Pre-commit Hook Not Executable

```
.git/hooks/pre-commit — not executable (warnings shown)
```

Not blocking — CI catches everything. But local pre-commit checks aren't running.

---

## e) WHAT WE SHOULD IMPROVE 🎯

### Architecture (High Impact)

1. **Confidence strong type** — The single biggest type-safety hole. `Finding{Confidence: 1.5}` compiles but is semantically invalid. A `type Confidence float64` with `NewConfidence(v float64) (Confidence, error)` and `Clamp()` would make invalid states unrepresentable. This is a direct violation of the project's own design principle: "make impossible states unrepresentable."

2. **Centralize triage logic** — Three different functions categorize fix readiness differently:
   - `Finding.HasFix()` checks `AfterCode != ""`
   - `Pipeline.triage()` switches on `FixStrategy`
   - `FixEngine.Apply()` filters on `BeforeCode != "" || AfterCode != ""`

   A single `CategorizeFix(f Finding) FixCategory` would eliminate inconsistency risk and make the logic testable in isolation.

3. **Properties map[string]any** — `Metadata map[string]string` forces structured data through `fmt.Sprintf("%v", v)`. SARIF properties is `map[string]any`. Adding a parallel `Properties` field preserves type information through round-trips.

### Testing Gaps (Medium Impact)

4. **Context cancellation** — `detectPartialSequential` and `detectPartialParallel` have cancel paths that are completely untested. These are critical for production reliability — a cancelled context should stop detection cleanly.

5. **SARIF schema validation** — No test verifies output conforms to SARIF 2.1.0 JSON schema. The hand-rolled types could silently drift from the spec. A single schema validation test would catch this.

### Tooling (Medium Impact)

6. **gci/golines in devShell** — Every contributor with the same nix setup will hit formatter issues. These should be in `flake.nix` devShell or the linter config should be relaxed.

7. **Benchmark regression tracking** — Baseline benchmarks exist but there's no CI job or script to detect regressions. A `scripts/bench-compare.sh` comparing against committed baselines would catch perf regressions.

### Documentation (Low Impact)

8. **SARIF round-trip losses documented** — Currently only in code comments. Users need to know what data survives SARIF export/import.

9. **Nix setup in CONTRIBUTING.md** — The project uses nix but the contribution guide doesn't mention it.

---

## f) Top #25 Things To Do Next 📋

Sorted by Impact × Effort (highest ROI first):

| #  | Task                                                   | Impact | Effort | ROI    | Blocker?   |
| -- | ------------------------------------------------------ | ------ | ------ | ------ | ---------- |
| 1  | **Push 5 unpushed commits**                            | HIGH   | LOW    | ⭐⭐⭐ | No         |
| 2  | **Confidence strong type** — `type Confidence float64` | MED    | MED    | ⭐⭐⭐ | No         |
| 3  | **Centralize triage logic** — single `CategorizeFix`   | MED    | MED    | ⭐⭐⭐ | No         |
| 4  | **Context-cancel tests** for detectPartial             | MED    | LOW    | ⭐⭐⭐ | No         |
| 5  | **Report thread-safety** — sync.Once for nil mutex     | MED    | MED    | ⭐⭐   | No         |
| 6  | **Split cmd/go-finding/main.go** — config + output     | MED    | MED    | ⭐⭐   | No         |
| 7  | **Properties map[string]any** on Finding               | MED    | MED    | ⭐⭐   | No         |
| 8  | **WriteSARIF error-path tests**                        | LOW    | LOW    | ⭐⭐   | No         |
| 9  | **io.WriterTo for SARIF** streaming                    | LOW    | LOW    | ⭐⭐   | No         |
| 10 | **Document SARIF round-trip losses**                   | MED    | LOW    | ⭐⭐   | No         |
| 11 | **Nix setup in CONTRIBUTING.md**                       | LOW    | LOW    | ⭐⭐   | No         |
| 12 | **Refactor CLI run()** for testability                 | MED    | MED    | ⭐     | No         |
| 13 | **Wire FixProviders through CLI config**               | MED    | MED    | ⭐     | No         |
| 14 | **Unify Tag deprecation** across tests                 | LOW    | LOW    | ⭐     | No         |
| 15 | **Protect Confidence construction**                    | MED    | MED    | ⭐     | No         |
| 16 | **Decide NewFinding API pattern**                      | HIGH   | N/A    | ⭐     | Needs user |
| 17 | **Decide domain provider location**                    | HIGH   | N/A    | ⭐     | Needs user |
| 18 | **API stability review** — audit all exports           | HIGH   | HIGH   | ⭐     | No         |
| 19 | **Error wrapping audit**                               | MED    | MED    | ⭐     | No         |
| 20 | **SARIF schema validation test**                       | MED    | MED    | ⭐     | No         |
| 21 | **Benchmark regression tracking**                      | LOW    | MED    | ⭐     | No         |
| 22 | **gci/golines in devShell**                            | MED    | LOW    | ⭐     | No         |
| 23 | **Consumer migration guide** v0.1→v0.2                 | MED    | MED    | ⭐     | No         |
| 24 | **Evaluate go-sarif vs hand-rolled**                   | MED    | MED    | ⭐     | Research   |
| 25 | **Finding JSON schema**                                | MED    | MED    | ⭐     | No         |

---

## g) Top #1 Question I Cannot Answer 🤔

**Should domain-specific FixProviders (Go AST, Rust syn, TypeScript compiler API) live INSIDE `pipeline/` as sub-packages, or as SEPARATE Go modules?**

This is a permanent module structure decision that cannot be reversed without a major version bump.

| Approach                                                                    | Pros                                                                   | Cons                                                                      |
| --------------------------------------------------------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| **Inside `pipeline/fix/goast/`**                                            | Easy discovery, shared test infra, single module                       | Couples domain deps (go/ast, rust syn) to pipeline module, bloated go.mod |
| **Separate modules** (`go-finding-provider-go`, `go-finding-provider-rust`) | Clean dep isolation, independent versioning, users pick what they need | Harder discovery, version skew risk, more repos to maintain               |
| **Hybrid: interface in core, impls as plugins**                             | Best of both: discoverable interfaces, swappable impls                 | Requires plugin registration mechanism, more complex                      |

**My recommendation:** Hybrid approach — `FixProvider` interface stays in `pipeline/`, domain-specific implementations live in separate modules (`go-finding-provider-goast`, etc.) that users import explicitly. This keeps the core lean while enabling a rich ecosystem.

**However:** This is YOUR architecture decision. It affects how third parties contribute, how dependencies flow, and how the project grows. It cannot be made without your input.

---

## Session Timeline

```
2026-04-12 ─── Project started
2026-04-13 ─── Pipeline complete
2026-04-15 ─── Session 17: comprehensive audit + cleanup
2026-04-18 ─── V5 complete, lint hardened
2026-04-19 ─── Sessions 12-14: hardening
2026-04-21 ─── Session 21 status
2026-04-25 ─── Comprehensive status
2026-04-26 ─── Clone elimination, architecture diagrams
2026-04-28 ─── Planning docs, execution plans
2026-04-29 ─── Deep codebase hardening, deepening plan
2026-04-30 ─── Comprehensive hardening + API stability plans, ADRs
2026-05-02 ─── Architecture diagrams (mermaid)
2026-05-04 ─── Dependency migration (go-faster/yaml → go.yaml.in/yaml), Tag removal
2026-05-05 ─── Byte-level FixEngine redesign + 20 follow-up commits
2026-05-06 ─── THIS REPORT — Full audit
```

---

## Build Verification

```
go build ./...           ✅ PASS (0 errors)
go vet ./...             ✅ PASS
go test -race -count=1   ✅ 985 PASS, 0 FAIL (6 packages)
golangci-lint run ./...  ✅ 0 issues
```

---

_Assisted-by: Crush <crush@charm.land>_
