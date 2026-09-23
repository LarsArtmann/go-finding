#!/usr/bin/env bash
# Self-test for release-preflight.sh — proves the gate actually bites.
#
# Runs the real preflight inside a disposable git worktree with an injected
# failure and asserts (a) preflight exits 1 and (b) the expected FAIL line
# appears in the output. Four failure classes are injected:
#
#   1. Tag collision — version.go claims a version whose tag already exists.
#   2. Version drift — version.go disagrees with sub-module go.mod requires.
#   3. Post-tag missing tag — --post-tag mode with no matching tag must FAIL
#      (only the happy path was proven before; report item 37).
#   4. Post-tag incomplete tag SET — 4 of 5 tags exist but toolsdk/v* is
#      missing; the incomplete v1.13.0 train (2 of 5 pushed) must be caught.
#
# The main repo is never modified (the worktree is disposable); the temp
# worktree is removed on exit. Returns non-zero if any injection was NOT
# caught. Run locally before releases and in CI (see ci.yml structural job).
#
# Usage: ./scripts/release-preflight-selftest.sh
set -euo pipefail

cd "$(dirname "$0")/.."

WORKTREE=$(mktemp -d /tmp/preflight-selftest-XXXXXX)
rmdir "$WORKTREE"
trap 'git worktree remove --force "$WORKTREE" >/dev/null 2>&1 || true' EXIT

git worktree add --quiet --detach "$WORKTREE" HEAD

# Version-relative injection targets: the latest existing tag (for the
# collision) and the NEXT minor (for the drift) are derived from the repo so
# the self-test keeps working at any version (hardcoded v1.8.0 stopped
# creating drift the moment version.go reached v1.9.0).
LATEST_TAG=$(git describe --tags --abbrev=0 --match 'v[0-9]*')
INJECT_COLLISION_MINOR=$(sed -n 's/^const VersionMinor = //p' "$WORKTREE/version.go")
INJECT_COLLISION_PATCH=$(sed -n 's/^const VersionPatch = //p' "$WORKTREE/version.go")
LATEST_MINOR=$(echo "$LATEST_TAG" | cut -d. -f2)
LATEST_PATCH=$(echo "$LATEST_TAG" | cut -d. -f3)
NEXT_MINOR=$((INJECT_COLLISION_MINOR + 1))

FAILURES=0

# expect_fail <label> <expected FAIL line> [preflight args...]
expect_fail() {
	local label="$1" expected="$2"
	shift 2
	local output
	if ! output=$(cd "$WORKTREE" && bash scripts/release-preflight.sh "$@" 2>&1); then
		if grep -qF "$expected" <<<"$output"; then
			echo "OK: $label"
			return
		fi
		echo "FAIL: $label — preflight failed but expected line missing: $expected"
		FAILURES=$((FAILURES + 1))
		return
	fi
	echo "FAIL: $label — preflight exited 0, expected exit 1 (gate is dead)"
	FAILURES=$((FAILURES + 1))
}

echo "== self-test 1/4: tag collision (injecting $LATEST_TAG) =="
sed -i -e "s/^const VersionMinor = .*/const VersionMinor = $LATEST_MINOR/" \
	-e "s/^const VersionPatch = .*/const VersionPatch = $LATEST_PATCH/" "$WORKTREE/version.go"
expect_fail "tag collision detected" "FAIL: tag $LATEST_TAG already exists"
git -C "$WORKTREE" restore version.go

echo "== self-test 2/4: version drift (injecting v-increment minor $NEXT_MINOR) =="
sed -i "s/^const VersionMinor = .*/const VersionMinor = $NEXT_MINOR/" "$WORKTREE/version.go"
expect_fail "version drift detected" "FAIL: version drift"

echo "== self-test 3/4: post-tag mode detects missing tag (still on v-injected version.go) =="
MISSING_PATCH=$(sed -n 's/^const VersionPatch = //p' "$WORKTREE/version.go")
MISSING_TAG="v$(sed -n 's/^const VersionMajor = //p' "$WORKTREE/version.go").$NEXT_MINOR.$MISSING_PATCH"
expect_fail "post-tag missing tag detected" "FAIL: tag $MISSING_TAG missing (post-tag mode expects it to exist)" --post-tag

echo "== self-test 4/4: post-tag mode detects INCOMPLETE tag set (toolsdk missing) =="
INJECTED_TAGS=()
for prefix in "" "pipeline/" "analysis/" "cmd/go-finding/"; do
	tag="${prefix}${MISSING_TAG}"
	# -m is required: the repo sets tag.gpgSign=true, so a bare `git tag`
	# creates a signed annotated tag and opens an editor.
	git -C "$WORKTREE" tag -m "selftest injected" "$tag" HEAD
	INJECTED_TAGS+=("$tag")
done
expect_fail "post-tag incomplete tag set detected" \
	"FAIL: tag toolsdk/${MISSING_TAG} missing (post-tag mode expects it to exist)" --post-tag
for tag in "${INJECTED_TAGS[@]}"; do
	git tag -d "$tag" >/dev/null
done

if [ "$FAILURES" -gt 0 ]; then
	echo "SELF-TEST FAIL: $FAILURES injection(s) not caught. Do NOT trust the preflight gate."
	exit 1
fi

echo "SELF-TEST PASS: preflight catches injected tag collision, version drift, post-tag missing tag, and incomplete tag set."
