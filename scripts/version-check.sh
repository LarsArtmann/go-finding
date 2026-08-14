#!/usr/bin/env bash
# Verifies version.go matches the latest git tag.
# Usage: ./scripts/version-check.sh
# On non-tagged commits, compares the nearest tag prefix against version.go.
# Fails with exit code 1 on mismatch.

set -euo pipefail

GIT_TAG=$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null) || {
	echo "OK: no git tag found, skipping version check."
	exit 0
}

MAJOR=$(sed -n 's/^const VersionMajor = //p' version.go)
MINOR=$(sed -n 's/^const VersionMinor = //p' version.go)
PATCH=$(sed -n 's/^const VersionPatch = //p' version.go)
GO_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

if [ "$GIT_TAG" != "$GO_VERSION" ]; then
	echo "ERROR: version.go ($GO_VERSION) does not match git tag ($GIT_TAG)"
	exit 1
fi

echo "OK: version.go ($GO_VERSION) matches tag ($GIT_TAG)"
