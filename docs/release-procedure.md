# Release Procedure

## Pre-release Checklist

> **Driver:** `scripts/release-train.sh` encodes the train as stages
> (`dry` | `prep` | `verify` | `stress` | `tag` | `push` | `resync` |
> `postflight`) with the ordering and guards below built in — start there;
> this document is the manual the stages implement. `dry` prints the plan
> for the current tree. The `postflight` stage is the post-release verifier:
> proxy ×5, clean-dir `go get` ×5, release assets, pkg.go.dev backoff, and
> the stamp-vs-tag probe (it fails red on the v1.15.1 shipped stamp defect
> when run against that tag — proven 2026-11-08).

0. **Run the preflight script** (structural gates as code — added after the
   v1.7.0 sub-module go.mod drift incident):

   ```bash
   bash scripts/release-preflight.sh
   ```

   It checks: clean tree, unpushed commits (nothing in `@{u}..HEAD`; the branch
   must have an upstream — tags and CI run against what GitHub has, not your
   working copy), worktree hygiene (no stray or prunable worktrees outside the
   repo), version.go parse, target tags not already taken, version drift
   (go.mod references), replace directives, go.work sync, test naming, JSON
   determinism, docs API references (including the README version stamp), docs
   freshness, and GOWORK=off builds per module. Add `--bench` / `--stress` to
   also run the benchmark and stress gates through it. Add `--tag vX.Y.Z` to
   also run the stamp gate (below). **Do not tag until it passes.**
1. Verify all tests pass: `go test -race -count=1 ./...`
2. Verify lint passes: `golangci-lint run ./...`
3. Verify the flake: `nix flake check` (treefmt, package build, module isolation)
4. Run stress test (or `bash scripts/release-preflight.sh --stress`):
   - stdlib modules (analysis, CLI): `go test -race -count=20 ./...`
   - Ginkgo modules (core, pipeline): `ginkgo -r --race --repeat=20 --skip-package=examples`
   - **Note:** Ginkgo rejects `go test -count=N` for N>1 (`Only -count=1 is allowed`). Use `ginkgo --repeat=N` instead for the core and pipeline suites.
   - **Gate status (decided 2026-09-08): this step is mandatory.** A release
     must not be tagged until the repeat=20 race run passes. It is the only
     check that sweeps for concurrency flake — a single-run CI pass routinely
     misses races that surface under repeated scheduling. Revisit only if a
     cheaper equivalent lands in CI.
   - **CI stress scope (revised 2026-09-08 evening): CI now mirrors this
     split** — `ginkgo -r --race --repeat=20 --skip-package=examples` for
     core/pipeline plus `go test -race -count=20 ./analysis/... ./cmd/...`.
     The earlier decision ("CI keeps count=20 only") was wrong: Ginkgo rejects
     `go test -count=N` for N>1 outright, so the CI stress job failed instantly
     on the core suite and never actually stressed anything (first observed
     publicly on run 34274674104 after the repo went public).
5. Update `CHANGELOG.md` with release notes
6. Update `version.go` (core module only) **and the README version stamp** —
   the `fmt.Println(finding.Version) // "X.Y.Z"` example in README.md must
   match version.go exactly; `docs-api-check.sh` (wired into preflight and CI)
   fails on a stale or missing stamp (v1.9.1/v1.9.2 shipped while it lagged).
   The bump commit must land BEFORE the tag is pushed, and the tag's tree must
   carry it — verify with `git show <tag>:version.go`. The v1.14.0 core tag
   self-reports "1.13.0" because its bump commit landed after tagging, and
   `protect-release-tags` blocks the retag (found 2026-10-11).

   **Stamp gate (mandatory, added 2026-11-08 after the v1.15.1 repeat of the
   v1.14.0 class):** immediately before creating ANY tag, run

   ```bash
   bash scripts/release-preflight.sh --tag vX.Y.Z
   # or standalone: bash scripts/stamp-gate.sh vX.Y.Z
   ```

   It asserts version.go equals the tag you are about to create and that the
   tag does not exist yet (works for the prefixed sub-module variants too).
   The selftest injects a stale stamp and proves the exact FAIL line
   (`scripts/release-preflight-selftest.sh`). A green preflight WITHOUT
   `--tag` is not stamp evidence — the check only runs when you pass the
   intended tag.

   Corollary (the v1.15.1 repeat): a green CI run on the PRE-tag commit
   proves nothing about the tag — that run compared the stamp against the
   PREVIOUS nearest tag. After pushing ANY core tag, run
   `bash scripts/version-check.sh` locally (it compares against the new
   nearest tag the moment the tag exists) and `git show <tag>:version.go`.

   If the CHANGELOG carries compare links for the new version, also run
   `bash scripts/release-preflight.sh --links` (authenticated `gh api` per
   pair — the authority on compare-link validity; lychee 429/502 flakes on
   valid pairs do not count as evidence).

   Compare-ref convention (the v1.15.0 lesson): `[Unreleased]` points at the
   last EXISTING tag (`vX.Y.Z...HEAD`); the new version's own ref is added in
   a POST-tag commit once the tag resolves — a pre-tag ref 404s in lychee.
   And when proving a consumer pin at an exact commit, `go get <mod>@<sha>`
   LAST and never `go mod tidy` afterward — tidy in a clean dir silently
   re-resolves the pin to `@latest`, validating the WRONG version.

   After the train's Release runs, the `release-gate` job in release.yml
   re-checks the stamp at the ARTIFACT level: it installs the just-pushed
   tag from the module proxy (core probe / `go-finding --version`) and fails
   the run when the printed version differs from the tag name. A red
   release-gate job is a shipped-defect signal, not CI noise.
7. **Full FEATURES.md walk** — read `FEATURES.md` end to end and verify every
   row's status against the code being released. The file is the honest feature
   inventory; a release must not ship with stale rows. Check in particular that
   nothing is still marked "(unreleased)" for the version being tagged.

## Multi-Module Tagging

This is a **multi-module repository** (see `go.work`). Each sub-module is an
independent Go module and must be tagged with a **directory-prefixed** tag so the
Go module proxy can resolve it:

| Module   | Import path                                        | Tag prefix          | Version source        |
| -------- | -------------------------------------------------- | ------------------- | --------------------- |
| Core     | `github.com/larsartmann/go-finding`                | `v*` (no prefix)    | `version.go`          |
| Pipeline | `github.com/larsartmann/go-finding/pipeline`       | `pipeline/v*`       | tag (no `version.go`) |
| Analysis | `github.com/larsartmann/go-finding/analysis`       | `analysis/v*`       | tag (no `version.go`) |
| Tool SDK | `github.com/larsartmann/go-finding/toolsdk`        | `toolsdk/v*`        | tag (no `version.go`) |
| CLI      | `github.com/larsartmann/go-finding/cmd/go-finding` | `cmd/go-finding/v*` | tag (no `version.go`) |

> Sub-modules do **not** have a `version.go`; the git tag is the single source of
> truth for their version.

## Release Steps

### Core module (e.g., v0.2.0)

All release work happens directly on `master` — no release branches.

```bash
# 1. Ensure clean working tree
git status

# 2. Update CHANGELOG.md (move "Unreleased" items into the new version)

# 3. Update version.go
#    Edit VersionMajor / VersionMinor / VersionPatch

# 4. Commit on master
git add CHANGELOG.md version.go
git commit -m "chore: prepare release v0.2.0"

# 5. Tag (NO prefix for core)
git tag -a v0.2.0 -m "Release v0.2.0"

# 6. Push master and tags
git push origin master --tags

# 7. The release.yml workflow triggers on v* and builds the CLI binary via GoReleaser.
```

### Sub-module (e.g., pipeline v0.1.0)

Sub-modules are versioned **independently** from core. Their first release starts
at `v0.1.0`. Bump according to semver based on changes _in that module_ since its
last tag.

```bash
# 1. Ensure clean working tree
git status

# 2. (Optional) add a CHANGELOG note for the sub-module.

# 3. Tag with the DIRECTORY PREFIX
git tag -a pipeline/v0.1.0 -m "pipeline v0.1.0"

# 4. Push the tag
git push origin pipeline/v0.1.0

# 5. The release.yml workflow triggers on pipeline/v*, runs tests, and creates a
#    lightweight GitHub Release (no binary — library modules need none).
```

Repeat for `analysis/v*`, `toolsdk/v*`, and `cmd/go-finding/v*` as needed. Only tag a
sub-module when it has meaningful changes since its previous tag.

### All modules at once (coordinated release)

To align all modules to the same version (e.g., after a breaking change across the
workspace):

```bash
git tag -a v1.3.0                 -m "Release v1.3.0"
git tag -a pipeline/v1.3.0        -m "pipeline v1.3.0"
git tag -a analysis/v1.3.0        -m "analysis v1.3.0"
git tag -a toolsdk/v1.3.0         -m "toolsdk v1.3.0"   # when the SDK changes too
git tag -a cmd/go-finding/v1.3.0  -m "cli v1.3.0"
git push origin master
git push origin v1.3.0                      # ONE tag per push (see below)
git push origin pipeline/v1.3.0             # wait for its Release run to
for t in analysis/v1.3.0 toolsdk/v1.3.0 cmd/go-finding/v1.3.0; do
  git push origin "$t"                      # complete before the next push
done
```

Tag order (v1.14.0 onward): push the LIBRARY tags first (`v*`, `pipeline/v*`,
`analysis/v*`, `toolsdk/v*` — one per push, each watched to a terminal
Release run), then land the resync commit on master: bump
`cmd/go-finding/go.mod` requires to the new versions, `go mod tidy`, refresh
the flake `vendorHash`, commit
(`chore(release): resync cmd/go-finding to published vX.Y.Z siblings`), push,
and wait for green CI. Only THEN push `cmd/go-finding/vX.Y.Z` pointing AT the
resync commit, so `go install cmd/go-finding@vX.Y.Z` builds the CLI with the
new sibling engine — tagging cmd before the resync ships a binary pinned to
the PREVIOUS engine (the v1.13.0 CLI pinned pipeline v1.12.0). GoReleaser
binaries still attach only to core `v*` releases; sub-module tags get
lightweight releases by design.

### Tag pushing: ONE tag per push, serialized runs

Two incidents shaped this rule:

1. **v1.9.0** — a single `git push` that updates **more than three tags
   creates no workflow events** at all (no run, no error; detected only by
   absence).
2. **v1.12.0** — even legal batches of 3 raced each other: GitHub's
   Release concurrency group uses latest-queued-wins, so pushing 3 tags
   near-simultaneously queued 3 runs and CANCELLED the older two. Whole
   modules silently lost their Release runs.

Therefore: push **one tag per push** and wait for that tag's Release run to
reach a terminal state before pushing the next:

```bash
git push origin "$TAG"
gh run list --workflow=release.yml --limit 1   # watch until completed
gh run rerun <id>                              # if cancelled — ONE at a time,
                                               # wait for completion, then next
```

Never queue several reruns at once (they cancel each other again). After
the train, **verify every tag has a run AND a GitHub Release before
assuming it does** (`gh run list --workflow=release.yml` +
`gh release list`); silence is the failure mode.

If a run truly never appeared (the v1.9.0 failure mode), dispatch it
manually — after checking no run is queued/in progress.

### Phased CI red: what is EXPECTED at each train phase

Not every red run during a train is a regression. These jobs are
STRUCTURALLY red at specific phases — the proof pattern is the v1.15.1 tag
tree (`pipeline` still required v1.13.0 while its tag built green):

| Phase                                          | Structurally red job                                   | Why it cannot be green yet                                                |
| ---------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------------------- |
| After the version bump, BEFORE the core tag     | `version-check`                                        | Compares version.go to the NEAREST EXISTING tag — the tag does not exist. |
| Before the resync commit                        | `version-drift`                                        | Sub-module requires still point at the previous core version.             |
| The core tag's own commit                       | `module-isolation`                                     | CLI go.sum cannot carry checksums for tags that do not exist yet; the resync lands on master immediately after. |
| Sibling tag pushed, resync not landed           | same as above                                          | The `replace => ../` directives make every GOWORK=off gate green locally while consumers of the PUBLISHED module would break. |

The release gate is the Release workflow ON THE TAG itself, plus the
`release-gate` consumer-proof job — not master CI color mid-train. Do not
"fix" a phased red by weakening the gate; land the phase that un-reds it.

**Decision (2026-11-08, T19): documented exception, NOT phase-aware gates.**
The gates stay strict in every phase; this table is the interpretation
guide for a human reading a train's red runs. A flag like
`preflight --phase` would weaken the gate by configuration, exactly the
"no gate weakened to make a task pass" anti-pattern.

**Sub-module-only releases (toolsdk, T20 decision: versioning stays
INDEPENDENT).** Preflight derives its target-tag list from core's
version.go, so a sub-module-only tag (e.g. `toolsdk/v1.16.0` while core
sits at v1.15.1) fails the "tag already exists" checks as false positives.
For a sub-module-only release: run every OTHER gate green (since 2026-11-08
the GOWORK=off build/tidy loops include toolsdk, closing the old blind
spot), then verify the actual target tag does not exist and the clean-dir
consumer proof (`go get <mod>@<tag>` in a scratch module + build) passes
before pushing. The toolsdk GOWORK=off manual line in older docs is
obsolete — the loop covers it now.

### Stress on sibling-tag trains (the d/1 rule, 2026-11-08)

The v1.15.1 sibling tags went out without the repeat=20 race stress on a
no-delta rationale. Rule: **a sibling-tag train carries the stress gate, or
the train log carries a WRITTEN no-op justification** (what changed in the
module since its last tag, why repetition cannot surface anything new).
Silent rationalizations do not count — the justification is written BEFORE
the tag is pushed, into the train notes or the tag's release notes.

### Files to gates: what validates what

| Change                                      | Gate that bites                                                                                |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `version.go`                                | `stamp-gate.sh --tag` (pre-tag), `version-check.sh` (post-tag), README stamp via `docs-api-check.sh` |
| any `go.mod` require of a sibling            | `version-drift.sh`, `replace-audit.sh` (replaces must point at `../`)                          |
| `go.work` / module set                      | `go-work-sync.sh`, `nix flake check` (vendorHash)                                              |
| `CHANGELOG.md` compare links                | `verify-compare-links.sh --links` (authority), lychee + token in CI (flake-prone first pass)   |
| CHANGELOG sections vs tags                  | `changelog-drift.sh`                                                                           |
| Deterministic JSON marshal                  | `json-deterministic-check.sh`                                                                  |
| Docs API claims (identifiers, versions)     | `docs-api-check.sh`, `docs-freshness.sh`, docs-gates battery                                   |
| Release publish correctness (nix/homebrew)  | `goreleaser check` locally; the Release run itself (nix pipe is skip_upload since v1.15.1's 404) |
| Benchmark regression claims                  | the CI benchmark job on a clean runner — LOCAL `--bench` numbers are invalid under heavy machine load (uniform 2-4x time inflation, flat allocs; observed load avg 186 during the v1.14.0 train) |
| go.sum completeness after dependency bumps   | `go-worksum-check.sh` + a cold-cache build (`module-isolation`); warm-cache local builds stay green on zip-hash-only entries |

### Local GoReleaser recovery recipe

When the Release run failed AFTER assets were published (the v1.15.1
pattern — red run, shipped release), do NOT re-dispatch blindly (duplicate
assets, 422 race). Options in order:

1. Diagnose from the failed run's log (`gh run view <id> --log-failed`).
2. Render-test locally without publishing:
   `git tag -a v0.0.0-snaptest -m tmp HEAD && goreleaser release --snapshot
   --clean --skip=sign,publish,sbom && git tag -d v0.0.0-snaptest`
   (snapshot needs a plain-semver tag at HEAD; the throwaway tag is local
   only and deleted in the same command).
3. If assets are complete and signed, the release itself is healthy — fix
   the workflow/config on master and record the incident; never retag
   release tags (`protect-release-tags` blocks it).

### Interrupted trains

If a train stops mid-way (tag pushed, run failed, operator unavailable):
write the state into `docs/status/` (what is tagged, what is missing, what
is expected-red) BEFORE leaving. The next session resumes from that report —
the v1.15.1 train sat abandoned for 4 days with nobody recording which
fails were structural and which were defects, and the defect hunt restarted
from zero.

### Queue, don't race, releases

Do not manually dispatch Release "to be sure" without first checking whether
the tag push already triggered it. At v1.9.2 a manual dispatch raced the
auto-triggered run: two concurrent GoReleaser runs, one failed with a 422
duplicate-asset error (the published release itself was fine — the red run was
self-inflicted noise). Check `gh run list` first; if a run is queued/in
progress, wait for it.

## Version Scheme

Semantic versioning strictly: `vMAJOR.MINOR.PATCH`. Current core version:
see `version.go`. Sub-modules track their own independent semver.

## Consumer Setup

The repository is **public** (since 2026-09-08). No `GOPRIVATE` configuration is
needed — modules resolve through `proxy.golang.org` like any other Go module:

```bash
go get github.com/larsartmann/go-finding@latest
```

(Historical: while the repo was private, consumers had to set
`GOPRIVATE=github.com/larsartmann/go-finding`. This is no longer required.)

## Verifying a Release

Confirm each module resolves after pushing its tag:

```bash
go list -m github.com/larsartmann/go-finding@latest
go list -m github.com/larsartmann/go-finding/pipeline@latest
go list -m github.com/larsartmann/go-finding/analysis@latest
go list -m github.com/larsartmann/go-finding/cmd/go-finding@latest
```
