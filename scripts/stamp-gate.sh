#!/usr/bin/env bash
# Stamp gate: asserts version.go matches the tag you are about to create.
# Born from the twice-run incident class (v1.14.0, then v1.15.1): the version
# stamp landed in a commit AFTER the tag was pushed, so released artifacts
# self-reported the previous version and master burned red for days. This gate
# makes the stale-stamp state impossible to tag through silently.
#
# Usage:
#   ./scripts/stamp-gate.sh <tag-name>
#
# <tag-name> is the core tag (vX.Y.Z) or any sub-module variant
# (pipeline/vX.Y.Z, analysis/vX.Y.Z, toolsdk/vX.Y.Z, cmd/go-finding/vX.Y.Z);
# the prefix is stripped and the same X.Y.Z is asserted for all of them.
#
# Run as the LAST step before `git tag` (release-procedure "Tag creation").
# The wire-in point is: ./scripts/release-preflight.sh --tag vX.Y.Z
#
# Exit codes: 0 = OK (stamp matches, tag does not exist yet), 1 = FAIL,
# 2 = usage error.
#
# Verdict discipline: prints an explicit OK/FAIL line on every path
# (dead-gate lesson, 2026-09-08).
set -euo pipefail

cd "$(dirname "$0")/.."

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <tag-name>   (e.g. v1.16.0 or pipeline/v1.16.0)"
	exit 2
fi

TAG="$1"
VERSION="${TAG##*/}"

if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "FAIL: stamp-gate: '$TAG' is not a vX.Y.Z tag name"
	exit 2
fi

MAJOR=$(sed -n 's/^const VersionMajor = //p' version.go)
MINOR=$(sed -n 's/^const VersionMinor = //p' version.go)
PATCH=$(sed -n 's/^const VersionPatch = //p' version.go)
STAMP="v${MAJOR}.${MINOR}.${PATCH}"

if [ -z "$MAJOR" ] || [ -z "$MINOR" ] || [ -z "$PATCH" ]; then
	echo "FAIL: stamp-gate: version.go did not parse"
	exit 1
fi

if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null; then
	echo "FAIL: stamp-gate: tag $TAG already exists"
	exit 1
fi

if [ "$STAMP" = "$VERSION" ]; then
	echo "OK: stamp-gate: version.go ($STAMP) == intended tag $TAG"
	exit 0
fi

echo "FAIL: stamp-gate: version.go ($STAMP) != intended tag $VERSION. Bump version.go (or fix the tag name) BEFORE creating the tag"
exit 1
