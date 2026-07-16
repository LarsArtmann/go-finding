# Release Procedure

## Pre-release Checklist

1. Verify all tests pass: `GOEXPERIMENT=jsonv2 go test -race -count=1 ./...`
2. Verify lint passes: `GOEXPERIMENT=jsonv2 golangci-lint run ./...`
3. Run stress test: `GOEXPERIMENT=jsonv2 go test -race -count=20 ./...`
4. Update `CHANGELOG.md` with release notes
5. Update `version.go`

## Release Steps

### For a new version (e.g., v0.2.0):

All release work happens directly on `master` — no release branches.

```bash
# 1. Ensure clean working tree
git status

# 2. Update CHANGELOG.md
# Move items from "Unreleased" to new version section

# 3. Update version.go if needed
# Edit Version variable

# 4. Commit on master
git add CHANGELOG.md version.go
git commit -m "chore: prepare release v0.2.0"

# 5. Tag
git tag -a v0.2.0 -m "Release v0.2.0"

# 6. Push master and tags
git push origin master --tags

# 7. Create GitHub Release from tag
# Use the CHANGELOG entries as release notes
```

### GoReleaser (when configured):

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
# GoReleaser workflow triggers automatically
```

## Version Scheme

Semantic versioning strictly: `vMAJOR.MINOR.PATCH`. Current version: `1.2.0` (see `version.go`).
