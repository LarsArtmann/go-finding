# Release Procedure

## Pre-release Checklist

1. Verify all tests pass: `go test -race -count=1 ./...`
2. Verify lint passes: `golangci-lint run ./...`
3. Run stress test: `go test -race -count=20 ./...`
4. Update `CHANGELOG.md` with release notes
5. Update `version.go` if major/minor bump

## Release Steps

### For a new version (e.g., v0.2.0):

```bash
# 1. Ensure clean working tree
git status

# 2. Create release branch
git checkout -b release/v0.2.0

# 3. Update CHANGELOG.md
# Move items from "Unreleased" to new version section

# 4. Update version.go if needed
# Edit Version variable

# 5. Commit
git add CHANGELOG.md version.go
git commit -m "chore: prepare release v0.2.0"

# 6. Tag
git tag -a v0.2.0 -m "Release v0.2.0"

# 7. Push
git push origin release/v0.2.0 --tags

# 8. Create GitHub Release from tag
# Use the CHANGELOG entries as release notes
```

### GoReleaser (when configured):

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
# GoReleaser workflow triggers automatically
```

## Version Scheme

- Pre-v1: `v0.MINOR.PATCH` — breaking changes allowed in minor
- Post-v1: `vMAJOR.MINOR.PATCH` — semantic versioning strictly
