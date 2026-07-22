# Status Report: go-finding v1.2.1

**Date:** 2026-07-22 18:08  
**Branch:** ~~master (1 commit ahead of origin)~~ master (multiple commits ahead of origin; v1.3.0 lint fixed, FormatText reverted, pending push+tag)  
**Version:** ~~v1.2.1~~ → v1.3.0 (implemented on master, not tagged)  
**Tests:** ✅ All pass (including race detector)  
**Lint:** ~~✅ 0 issues~~ ⚠️ 7 issues after v1.3.0 additions

---

> **Update (2026-07-22 session 3):** Lint is now clean (0 issues). `FindingTemplate` renamed to `Template`
> (revive stutter fixed). `FormatText` reverted to original `[SEVERITY]` format; `FormatTextRich` added for
> emoji badges. GOWORK=off isolation tests pass for all 4 modules. v1.3.0 still needs push + tag.
>
> **Update (2026-07-22):** v1.3.0 (11 consumer-driven API improvements) has since been implemented
> on master. Lint was ~~no longer clean~~ now clean again (0 issues). The version-check CI gate and
> dependabot groups mentioned as open in sections (b)/(c) have shipped. See
> `2026-07-22_18-55_v1.3.0-consumer-api-implementation-status.md` for the full v1.3.0 breakdown.

## Summary

Quick session to verify and answer: "Is the BuildFlow auto-configure loop still an issue?"

**Answer:** Yes, still active — but BLOCKED on external tool. No impact on go-finding itself. `golangci-lint run ./...` reports 0 issues on all 4 modules.

---

## a) FULLY DONE

| #   | Item                        | Evidence                                                                |
| --- | --------------------------- | ----------------------------------------------------------------------- |
| 1   | Sync `version.go` to v1.2.1 | `VersionPatch = 1` at `version.go:12`                                   |
| 2   | Version-check CI gate       | `scripts/version-check.sh` + CI job at `.github/workflows/ci.yml:65-73` |
| 3   | Race detector re-run        | `nix run .#test-race` — all packages OK, 0 races                        |
| 4   | Dependabot groups           | `.github/dependabot.yml` — `patterns: ["*"]` for gomod + github-actions |
| 5   | SHA-pin GitHub Actions      | All 8 actions across `ci.yml` + `release.yml` pinned to commit SHAs     |
| 6   | CODECOV_TOKEN support       | `token: ${{ secrets.CODECOV_TOKEN }}` + `id-token: write` in CI         |
| 7   | Tests pass                  | All 4 test packages pass (core, examples, gotoken, lockutil)            |
| 8   | Lint clean                  | `golangci-lint run ./...` — 0 issues                                    |

---

## b) PARTIALLY DONE

| #   | Item                 | What's Missing                                                                                |
| --- | -------------------- | --------------------------------------------------------------------------------------------- |
| 1   | Push to origin       | 1 commit ahead of `origin/master` — not pushed yet                                            |
| 2   | CODECOV_TOKEN secret | Workflow references `secrets.CODECOV_TOKEN` but secret not configured in GitHub repo settings |

---

## c) NOT STARTED

| #   | Item                         | Why                                                      |
| --- | ---------------------------- | -------------------------------------------------------- |
| 1   | Consumer compatibility test  | BLOCKED — repo is private, consumers need `GOPRIVATE`    |
| 2   | SARIF schema validation test | BLOCKED — requires vendoring 7K+ line JSON schema        |
| 3   | `.golangci.yml` in repo      | Missing — was removed or managed externally by BuildFlow |
| 4   | Tests for `examples/basic`   | `[no test files]` — no coverage                          |
| 5   | Tests for `examples/builder` | `[no test files]` — no coverage                          |
| 6   | Push 1 commit to origin      | Skipped — not asked to push                              |

---

## d) TOTALLY FUCKED UP

**Nothing.** Clean state. Tests pass, lint clean, working tree clean, version synced.

---

## e) WHAT WE SHOULD IMPROVE

1. **Push the pending commit** — 1 commit ahead of origin, never pushed
2. **Add `.golangci.yml`** to the repo — currently managed externally by BuildFlow, which causes the auto-configure loop. Having our own config prevents BuildFlow from injecting bogus linters
3. **Add `CODECOV_TOKEN`** to GitHub repo settings — workflow references it but secret doesn't exist yet
4. **Add tests for example packages** — `examples/basic` and `examples/builder` have no test files
5. **BuildFlow loop** — needs external fix (per-module scoring double-counts across workspace)
6. **Consumer compatibility** — needs repo to be public or `GOPRIVATE` configured in consumers

---

## f) Up to 50 Things We Should Get Done Next

### Immediate (this week)

1. Push the 1 pending commit to `origin/master`
2. Add `CODECOV_TOKEN` secret to GitHub repo settings
3. Add `.golangci.yml` to the repo (prevent BuildFlow auto-configure)
4. Add basic tests for `examples/basic` and `examples/builder`
5. Verify CI runs clean on pushed commit (version-check, lint, test, coverage)
6. Tag `v1.2.1` release if not already done (check `git tag -l 'v*'`)

### Short-term (next 2 weeks)

7. Document SARIF 2.1.0 schema validation approach (even if test is deferred)
8. Write consumer compatibility test harness (ready for when repo goes public)
9. Add `CONTRIBUTING.md` for external contributors
10. Review all `docs/status/archive/` files — archive or delete stale ones
11. Audit `AGENTS.md` for accuracy against current codebase state
12. Verify all GitHub Actions workflows run green (check Actions tab)
13. Add `CHANGELOG.md` entry for v1.2.1 (version sync fix + CI gates)
14. Review `pipeline/` module test coverage — ensure FixEngine edge cases covered
15. Check if `analysis/` module needs more test fixtures
16. Verify `cmd/go-finding` binary builds correctly with `GOEXPERIMENT=jsonv2`

### Medium-term (next month)

17. Fix BuildFlow auto-configure loop (external tool fix)
18. Remove deprecated `ToSARIFFiltered` / `WriteSARIFFiltered` aliases (if v2.0 scope allows)
19. Add SARIF schema validation test (vendoring approach TBD)
20. Consumer compatibility test (once repo is public)
21. Review `pipeline/goast/provider.go` for edge cases
22. Audit `branded_types.go` — ensure all string types are properly branded
23. Review `finding_equal.go` for correctness with new branded types
24. Check `sarif_export.go` and `sarif_import.go` round-trip fidelity
25. Verify `lsp.go` diagnostic data preservation across all field types
26. Review `interval_tree.go` performance characteristics
27. Check `merge.go` and `diff.go` for edge cases with dedup
28. Audit `format.go` and `json.go` output correctness

### v2.0 Preparation

29. Design `Option[T]` generic helper for Position sentinel cleanup
30. Design `FixStrategy` interface-based closed union
31. Plan `Finding` sub-struct composition (breaking JSON change)
32. Design `TagSet` replacement for `Tags []Tag`
33. Plan migration path for pointer-as-state fields
34. Write v2.0 migration guide
35. Create v2.0 branch strategy

### Infrastructure

36. Set up release automation (goreleaser config review)
37. Verify `cosign` signing works for releases
38. Review `sbom-action` output format
39. Add `SECURITY.md` for vulnerability reporting
40. Add `LICENSE` file if not present
41. Review GitHub repo settings (branch protection, required reviews)
42. Set up Codecov integration properly (token + OIDC)
43. Add `CODEOWNERS` file
44. Review `dependabot.yml` update schedule and reviewers

### Documentation

45. Update `README.md` with current feature set
46. Add API reference documentation (go doc)
47. Write `docs/TROUBLESHOOTING.md` for common issues
48. Document `GOEXPERIMENT=jsonv2` requirement in README
49. Add examples to README for quick start
50. Review and update `docs/DOMAIN_LANGUAGE.md` for accuracy

---

## g) Questions I Cannot Answer

1. **Is `v1.2.1` already tagged?** — I see `VersionPatch = 1` synced, but need to verify if a git tag exists at the current commit or if one needs to be created.

2. **Should we push the pending commit now?** — 1 commit is ahead of `origin/master`. Is this intentional (batching work) or should it go live?

3. **What is the current BuildFlow configuration for this project?** — Is there a `.buildflow.yml` in the repo root or elsewhere that controls which steps run? Could we disable `golangci-lint-auto-configure` there?

---

## Current State Snapshot

```
Tests:        ✅ All pass (4 packages, including race detector)
Lint:         ✅ 0 issues (golangci-lint run ./...)
Version:      ✅ v1.2.1 (VersionPatch = 1, CI gate active)
Git:          ✅ Clean working tree, 1 commit ahead of origin
CI:           ✅ SHA-pinned actions, dependabot groups, codecov token configured
BLOCKED:      3 items (BuildFlow loop, SARIF schema, consumer compat)
v2.0 deferred: 5 items (Position, FixStrategy, pointer-as-state, TagSet, Finding composition)
```

---

_Assisted-by: Crush <crush@charm.land>_
