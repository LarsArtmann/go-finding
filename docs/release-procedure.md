# Release Procedure

## Pre-release Checklist

0. **Run the preflight script** (structural gates as code — added after the
   v1.7.0 sub-module go.mod drift incident):

   ```bash
   bash scripts/release-preflight.sh
   ```

   It checks: clean tree, version.go parse, target tags not already taken,
   version drift (go.mod references), replace directives, go.work sync, test
   naming, JSON determinism, docs API references, docs freshness, and
   GOWORK=off builds per module. Add `--bench` / `--stress` to also run the
   benchmark and stress gates through it. **Do not tag until it passes.**
1. Verify all tests pass: `GOEXPERIMENT=jsonv2 go test -race -count=1 ./...`
2. Verify lint passes: `GOEXPERIMENT=jsonv2 golangci-lint run ./...`
3. Verify the flake: `nix flake check` (treefmt, package build, module isolation)
4. Run stress test (or `bash scripts/release-preflight.sh --stress`):
   - stdlib modules (analysis, CLI): `GOEXPERIMENT=jsonv2 go test -race -count=20 ./...`
   - Ginkgo modules (core, pipeline): `GOEXPERIMENT=jsonv2 ginkgo -r --race --repeat=20 --skip-package=examples`
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
6. Update `version.go` (core module only)
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
git push origin v1.3.0 pipeline/v1.3.0 analysis/v1.3.0   # batches of ≤3, see below
git push origin toolsdk/v1.3.0 cmd/go-finding/v1.3.0
```

### Tag pushing: batches of ≤3

A single `git push` that updates **more than three tags creates no workflow
events**: at v1.9.0, pushing master + 4 release tags in one push silently
skipped every Release trigger (no run, no error — detected only by absence).
Always split tag pushes into batches of ≤3, or dispatch Release manually:

```bash
gh workflow run Release --ref vX.Y.Z   # dispatch if no run appeared
```

After a release train, **verify the Release run exists before assuming it
does** (`gh run list --workflow=release.yml`); silence is the failure mode.

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
