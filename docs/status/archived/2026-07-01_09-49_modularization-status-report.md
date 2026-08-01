# Modularization Status — Unix-Style Multi-Module Decomposition

> **📦 RESOLUTION STATUS (updated 2026-07-16)**
>
> SUPERSEDED by the final-status report (same day, 10:10). Branch merged to master, all 4 modules
> verified with GOWORK=off isolation, nix build passes. v1.1.0+ shipped with this structure.

**Date:** 2026-07-01 09:49
**Branch:** `modularize/unix-style`
**Base:** `master` (v1.0.0)
**Commits:** 17 on branch (16 pushed + 1 uncommitted go.work.sum drift)

---

## a) FULLY DONE ✅

| #   | Task                                                    | Commit    | Verification                                       |
| --- | ------------------------------------------------------- | --------- | -------------------------------------------------- |
| 1   | Move gotoken from internal/ to public package           | `9a9394a` | go build + analysis/goast tests pass               |
| 2   | Move detectors into cmd/go-finding/internal/ tree       | `92adedf` | go build + cmd tests pass                          |
| 3   | Move benchutil into pipeline/internal/ tree             | `85aec26` | go build + pipeline bench tests pass               |
| 4   | Relocate pipeline-dependent examples to pipeline module | `ba0ee49` | Root no longer imports pipeline                    |
| 5   | Create pipeline/go.mod (extract pipeline module)        | `ad0892f` | GOWORK=off build+test in pipeline/                 |
| 6   | Create analysis/go.mod (extract analysis module)        | `035ebcc` | GOWORK=off build+test in analysis/                 |
| 7   | Create cmd/go-finding/go.mod (extract CLI module)       | `035ebcc` | GOWORK=off build+test in cmd/                      |
| 8   | Create go.work + tidy all modules                       | `035ebcc` | All modules build standalone                       |
| 9   | Update flake.nix for multi-module (modRoot, vendorHash) | `e127e53` | nix build .#default passes                         |
| 10  | Update AGENTS.md for multi-module structure             | `70502da` | Reviewed manually                                  |
| 11  | Write modularization proposal HTML                      | `d6454d6` | docs/modularization/2026-07-01_PROPOSAL.html       |
| 12  | Write execution plan HTML                               | `2671d29` | docs/modularization/2026-07-01_EXECUTION_PLAN.html |
| 13  | Track go.work/go.work.sum (remove from .gitignore)      | `556525e` | Both files committed                               |
| 14  | Fix coverage-check.sh stale internal/detectors path     | `3f9dfb5` | Path updated to cmd/go-finding/internal/detectors  |
| 15  | Fix integration-guide.md stale detector path            | `3f9dfb5` | Path updated                                       |
| 16  | Add per-module GOWORK=off CI isolation job              | `954d681` | .github/workflows/ci.yml module-isolation job      |
| 17  | Update README.md per-module install instructions        | `e3fbded` | Core/Pipeline/CLI install sections                 |
| 18  | Fix CONTRIBUTING.md stale directory tree                | `e3fbded` | detectors path corrected                           |
| 19  | Update ADR #7 with actual GoAST resolution              | `7731ecc` | docs/architecture-decisions.md                     |
| 20  | Update FEATURES.md goast dependency description         | `7731ecc` | pipeline/goast described correctly                 |
| 21  | Fix goreleaser hook (go mod tidy → go mod download)     | `7731ecc` | goreleaser snapshot build passes                   |
| 22  | Document testify as transitive dep in AGENTS.md         | `7731ecc` | Gotcha section updated                             |
| 23  | Add compile-time interface assertions                   | `70f1029` | goast.Provider + analysis.AnalyzerDetector         |
| 24  | Workspace race tests                                    | verified  | go test -race ./... — all pass                     |
| 25  | GOWORK=off per-module isolation tests                   | verified  | All 4 modules build+test standalone                |
| 26  | Nix build verification                                  | verified  | nix build .#default — binary works                 |
| 27  | Goreleaser snapshot verification                        | verified  | goreleaser build --snapshot — succeeds             |
| 28  | Push to remote                                          | verified  | origin/modularize/unix-style up to date            |

### Final Module Structure

| Module       | Path                                | Production Deps              | Test Deps      | Lines (prod) |
| ------------ | ----------------------------------- | ---------------------------- | -------------- | ------------ |
| **Core**     | `github.com/larsartmann/go-finding` | **Zero** — stdlib only       | ginkgo, gomega | 5,425        |
| **Pipeline** | `.../pipeline`                      | x/sync, gogenfilter          | ginkgo, gomega | 4,034        |
| **Analysis** | `.../analysis`                      | x/tools                      | stdlib testing | 393          |
| **CLI**      | `.../cmd/go-finding`                | yaml, go-output, gogenfilter | gomega         | 1,183        |

### Key Achievement

A consumer who does `go get github.com/larsartmann/go-finding` now gets **zero external dependencies**. Previously they got x/tools, x/sync, gogenfilter, yaml, and go-output transitively. Zero breaking changes — all public import paths unchanged.

---

## b) PARTIALLY DONE 🟡

| Item                     | Status                         | Details                                                                                               |
| ------------------------ | ------------------------------ | ----------------------------------------------------------------------------------------------------- |
| go.work.sum drift        | Uncommitted                    | `go work sync` added 8 new hashes; needs commit                                                       |
| doc.go module references | Stale                          | References "analysis subpackage" and "pipeline subpackage" as if same-module; now separate Go modules |
| CI module-isolation job  | Written but untested on GitHub | The YAML is correct but hasn't run in actual CI yet                                                   |

---

## c) NOT STARTED ⬜

| Item                                                                            | Impact | Notes                                                                                                                           |
| ------------------------------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------- |
| Archive stale root proposals (PROPOSAL.md, MIGRATION_TO_NIX_FLAKES_PROPOSAL.md) | Low    | These are historical proposal docs from April 2026; should move to docs/archive/ or delete                                      |
| Per-module golangci-lint config                                                 | Medium | Current `.golangci.yml` is root-only; each module may need its own config for per-module linting in CI                          |
| Pipeline go.mod requires core at v1.0.0, CLI requires pipeline at v0.0.0        | Low    | Version drift in internal requires. Functionally correct (replace overrides) but inconsistent with v0.0.0 normalization pattern |
| Go workspace vendor mode (`go work vendor`)                                     | Low    | Not currently used; documented as option in skill references                                                                    |
| Per-module coverage thresholds                                                  | Medium | coverage-check.sh works but thresholds may need adjustment now that packages moved between modules                              |

---

## d) TOTALLY FUCKED UP 💥

| Item                                                  | Severity          | What Happened                                                                                                             | Fix                                                                |
| ----------------------------------------------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| **go.work was originally in .gitignore**              | Critical (fixed)  | The initial modularization committed go.work but .gitignore still excluded it. No collaborator could build the workspace. | Fixed in `556525e` — removed from .gitignore, force-added          |
| **Root go.mod tidy failed after pipeline extraction** | Delay (recovered) | Created pipeline/go.mod first, but root `go mod tidy` failed because cmd still imported pipeline from root module.        | Worked around by creating all sub-modules before tidying root.     |
| **flake.nix vendorHash broke**                        | Expected (fixed)  | Multi-module changed dependency graph; old vendorHash invalid.                                                            | Fixed with `lib.fakeHash` → error reveals correct hash → `e127e53` |
| **No issues remain unfucked**                         | —                 | All critical issues have been resolved and committed                                                                      | —                                                                  |

---

## e) WHAT WE SHOULD IMPROVE 🔄

### Architecture

1. **Version normalization** — Internal `require` entries use mixed v0.0.0 and v1.0.0. Normalize all to v0.0.0 per real-world-patterns.md, or accept v1.0.0 consistently since the tag exists.
2. **doc.go references** — Root package doc mentions "analysis subpackage" and "pipeline subpackage" as if they're in the same module. Should clarify they're separate importable modules.
3. **Test dependency isolation** — Core go.mod still has ginkgo/gomega as direct deps for BDD tests. Could split BDD tests into a companion `finding_test` module to achieve truly zero-dep core go.mod (real-world-patterns §Test Infrastructure).
4. **Per-module golangci-lint** — Each module should have or inherit lint config for independent CI linting.

### Process

5. **go.work.sum should be committed immediately after go work sync** — We left it dirty across multiple commits. Should be atomic.
6. **The skill's Phase 5 (execution plan) was written retroactively** — Should be written before execution, not after. The proposal was written correctly but the execution plan was backfilled.

### Developer Experience

7. **flake.nix devShell no longer sets GOWORK=off** — Correct for development, but CI shell also dropped it. CI builds rely on replace directives now, which is correct but should be documented.
8. **goreleaser `go mod download` hook** — Changed from `go mod tidy` to `go mod download` to avoid cross-module resolution issues. Should verify this works for release builds (snapshot passes but a real release hasn't been cut).

---

## f) TOP 25 THINGS TO DO NEXT

| #   | Priority | Task                                                                           | Impact                    | Effort   |
| --- | -------- | ------------------------------------------------------------------------------ | ------------------------- | -------- |
| 1   | 🔴 P0    | Commit go.work.sum drift (8 new hashes from go work sync)                      | Blocks clean working tree | 1 min    |
| 2   | 🔴 P0    | Update doc.go to clarify analysis/pipeline are separate modules                | Misleading godoc          | 5 min    |
| 3   | 🟠 P1    | Normalize internal require versions (all v0.0.0 or all v1.0.0)                 | Consistency               | 5 min    |
| 4   | 🟠 P1    | Move stale root proposals (PROPOSAL.md, MIGRATION_TO_NIX) to docs/archive/     | Repo cleanliness          | 3 min    |
| 5   | 🟠 P1    | Per-module golangci-lint config or inherit strategy                            | Per-module CI linting     | 15 min   |
| 6   | 🟠 P1    | Add go.work sync idempotency CI check (`go work sync && git diff --exit-code`) | Catches drift             | 5 min    |
| 7   | 🟡 P2    | Verify coverage-check.sh thresholds match new package paths                    | CI coverage correctness   | 10 min   |
| 8   | 🟡 P2    | Add replace directive audit CI check (no absolute paths)                       | Portability               | 5 min    |
| 9   | 🟡 P2    | Add version drift detection CI check                                           | Catch version mismatches  | 10 min   |
| 10  | 🟡 P2    | Consider splitting BDD tests into finding_test companion module                | Zero-dep core go.mod      | 30 min   |
| 11  | 🟡 P2    | Update docs/integration-guide.md for multi-module import patterns              | User guidance             | 10 min   |
| 12  | 🟡 P2    | Update docs/USAGE_GUIDE.md if it references old structure                      | Docs accuracy             | 5 min    |
| 13  | 🟡 P2    | Consider `go work vendor` for reproducible builds                              | Build reproducibility     | 15 min   |
| 14  | 🟢 P3    | Add module boundary diagram to README.md                                       | Onboarding                | 10 min   |
| 15  | 🟢 P3    | Add "choosing the right module" section to docs                                | User guidance             | 10 min   |
| 16  | 🟢 P3    | Create docs/MIGRATION_multi-module.md for consumers                            | Migration guide           | 15 min   |
| 17  | 🟢 P3    | Tag v1.1.0 with multi-module structure                                         | Release                   | 5 min    |
| 18  | 🟢 P3    | Consider independent module versioning (per-module semver tags)                | Consumer flexibility      | Decision |
| 19  | 🟢 P3    | Add module dependency graph to AGENTS.md (D2 or ASCII)                         | Developer orientation     | 10 min   |
| 20  | 🟢 P3    | Verify govulncheck works across all modules                                    | Security scanning         | 5 min    |
| 21  | 🟢 P3    | Consider go-arch-lint for layer enforcement                                    | Architecture guard        | 20 min   |
| 22  | 🟢 P3    | Benchmark multi-module vs monolith build times                                 | CI parallelization data   | 15 min   |
| 23  | 🟢 P3    | Add per-module CHANGELOG entries                                               | Release communication     | 10 min   |
| 24  | 🟢 P3    | Review if gotoken should be `internal/` within core module                     | API surface minimization  | Decision |
| 25  | 🟢 P3    | Consider merging examples into a single examples module                        | Simplicity                | Decision |

---

## g) TOP QUESTION ❓

**Should the internal require versions be normalized to v0.0.0 or left at v1.0.0?**

The real-world-patterns.md skill reference recommends v0.0.0 to eliminate pseudo-version churn. However:

- The root module has a real `v1.0.0` git tag, so `go mod tidy` resolves internal requires to `v1.0.0` naturally
- Replace directives make the version irrelevant for resolution
- Using v1.0.0 is arguably more honest (it IS the real version) but causes `go mod tidy` to keep resolving it
- Using v0.0.0 requires manually overriding what `go mod tidy` produces, which fights the toolchain

I cannot determine the correct answer because both work functionally. The skill says v0.0.0 but Go's tooling prefers real versions. **Which approach should we use?**

---

## Verification Summary

```
Module           Build   Test    Vet     GOWORK=off Build   GOWORK=off Test
Core (.)         ✅      ✅      ✅      ✅                  ✅
Pipeline         ✅      ✅      ✅      ✅                  ✅
Analysis         ✅      ✅      ✅      ✅                  ✅
CLI              ✅      ✅      ✅      ✅                  ✅
Nix build        ✅
Goreleaser       ✅
Lint             ✅ (0 issues)
```

All green. Zero breaking changes. Ready for review and merge.

---

_Assisted-by: Crush <crush@charm.land>_
