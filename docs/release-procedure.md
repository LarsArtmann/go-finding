# Release Procedure

## Pre-release Checklist

1. Verify all tests pass: `GOEXPERIMENT=jsonv2 go test -race -count=1 ./...`
2. Verify lint passes: `GOEXPERIMENT=jsonv2 golangci-lint run ./...`
3. Run stress test:
   - stdlib modules (analysis, CLI): `GOEXPERIMENT=jsonv2 go test -race -count=20 ./...`
   - Ginkgo modules (core, pipeline): `GOEXPERIMENT=jsonv2 ginkgo -r --race --repeat=20 --skip-package=examples`
   - > **Note:** Ginkgo rejects `go test -count=N` for N>1 (`Only -count=1 is allowed`). Use `ginkgo --repeat=N` instead for the core and pipeline suites.
4. Update `CHANGELOG.md` with release notes
5. Update `version.go` (core module only)

## Multi-Module Tagging

This is a **multi-module repository** (see `go.work`). Each sub-module is an
independent Go module and must be tagged with a **directory-prefixed** tag so the
Go module proxy can resolve it:

| Module   | Import path                                        | Tag prefix          | Version source        |
| -------- | -------------------------------------------------- | ------------------- | --------------------- |
| Core     | `github.com/larsartmann/go-finding`                | `v*` (no prefix)    | `version.go`          |
| Pipeline | `github.com/larsartmann/go-finding/pipeline`       | `pipeline/v*`       | tag (no `version.go`) |
| Analysis | `github.com/larsartmann/go-finding/analysis`       | `analysis/v*`       | tag (no `version.go`) |
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

Repeat for `analysis/v*` and `cmd/go-finding/v*` as needed. Only tag a
sub-module when it has meaningful changes since its previous tag.

### All modules at once (coordinated release)

To align all modules to the same version (e.g., after a breaking change across the
workspace):

```bash
git tag -a v1.3.0                 -m "Release v1.3.0"
git tag -a pipeline/v1.3.0        -m "pipeline v1.3.0"
git tag -a analysis/v1.3.0        -m "analysis v1.3.0"
git tag -a cmd/go-finding/v1.3.0  -m "cli v1.3.0"
git push origin master --tags
```

## Version Scheme

Semantic versioning strictly: `vMAJOR.MINOR.PATCH`. Current core version:
`1.6.0` (see `version.go`). Sub-modules track their own independent semver.

## Private-Repo Consumer Setup

This repository is currently **private**. Until it is made public, every consumer
must tell the Go toolchain not to use the public proxy/sumdb for it:

```bash
go env -w GOPRIVATE=github.com/larsartmann/go-finding
# or, for all repos in this org:
go env -w GOPRIVATE=github.com/larsartmann/*
```

Add this to your environment or CI before running `go mod tidy` / `go get`. Once
the repo is made public, this is no longer required and the modules resolve
through `proxy.golang.org` normally.

## Verifying a Release

Confirm each module resolves after pushing its tag:

```bash
GOPRIVATE=github.com/larsartmann/go-finding go list -m github.com/larsartmann/go-finding@latest
GOPRIVATE=github.com/larsartmann/go-finding go list -m github.com/larsartmann/go-finding/pipeline@latest
GOPRIVATE=github.com/larsartmann/go-finding go list -m github.com/larsartmann/go-finding/analysis@latest
GOPRIVATE=github.com/larsartmann/go-finding go list -m github.com/larsartmann/go-finding/cmd/go-finding@latest
```
