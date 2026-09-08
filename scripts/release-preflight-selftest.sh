#!/usr/bin/env bash
# Self-test for release-preflight.sh — proves the gate actually bites.
#
# Runs the real preflight inside a disposable git worktree with an injected
# failure and asserts (a) preflight exits 1 and (b) the expected FAIL line
# appears in the output. Two failure classes are injected:
#
#   1. Tag collision — version.go claims a version whose tag already exists.
#   2. Version drift — version.go disagrees with sub-module go.mod requires.
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

FAILURES=0

# expect_fail <label> <expected FAIL line>
expect_fail() {
	local label="$1" expected="$2" output
	if ! output=$(cd "$WORKTREE" && bash scripts/release-preflight.sh 2>&1); then
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

echo "== self-test 1/2: tag collision =="
sed -i -e 's/^const VersionMinor = .*/const VersionMinor = 8/' \
	-e 's/^const VersionPatch = .*/const VersionPatch = 0/' "$WORKTREE/version.go"
expect_fail "tag collision detected" "FAIL: tag v1.8.0 already exists"
git -C "$WORKTREE" restore version.go

echo "== self-test 2/2: version drift =="
sed -i 's/^const VersionMinor = .*/const VersionMinor = 9/' "$WORKTREE/version.go"
expect_fail "version drift detected" "FAIL: version drift"

if [ "$FAILURES" -gt 0 ]; then
	echo "SELF-TEST FAIL: $FAILURES injection(s) not caught. Do NOT trust the preflight gate."
	exit 1
fi

echo "SELF-TEST PASS: preflight catches injected tag collision and version drift."
