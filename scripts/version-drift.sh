#!/usr/bin/env bash
# Cross-checks that all 4 modules reference the same go-finding version.
# Usage: ./scripts/version-drift.sh
# Fails with exit code 1 if any module's required version differs from version.go.

set -euo pipefail

# Extract the core version from version.go.
MAJOR=$(sed -n 's/^const VersionMajor = //p' version.go)
MINOR=$(sed -n 's/^const VersionMinor = //p' version.go)
PATCH=$(sed -n 's/^const VersionPatch = //p' version.go)
CORE_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

ERRORS=0

check_version() {
	local modfile="$1"
	local module="$2"
	local expected="$3"

	local actual
	actual=$(grep -E "^\s+${module} ${expected}" "$modfile" | grep -v '// indirect' | head -1 | awk '{print $2}')
	if [ "$actual" != "$expected" ]; then
		echo "ERROR: $modfile requires ${module} ${actual}, expected ${expected}"
		ERRORS=$((ERRORS + 1))
	fi
}

echo "Checking version drift..."

check_version "analysis/go.mod" "github.com/larsartmann/go-finding" "$CORE_VERSION"
check_version "pipeline/go.mod" "github.com/larsartmann/go-finding" "$CORE_VERSION"
check_version "cmd/go-finding/go.mod" "github.com/larsartmann/go-finding" "$CORE_VERSION"
check_version "cmd/go-finding/go.mod" "github.com/larsartmann/go-finding/pipeline" "$CORE_VERSION"

if [ "$ERRORS" -gt 0 ]; then
	echo "FAIL: $ERRORS version drift issue(s) found."
	exit 1
fi

echo "OK: all modules reference version $CORE_VERSION."
