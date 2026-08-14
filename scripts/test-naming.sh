#!/usr/bin/env bash
# Rejects test files with banned naming conventions.
# Banned patterns: _extra_test.go, _bugfix_test.go, coverage_test.go
# Usage: ./scripts/test-naming.sh

set -euo pipefail

ERRORS=0

check_pattern() {
	local pattern="$1"
	local description="$2"

	local matches
	# shellcheck disable=SC2312
	matches=$(find . -name "$pattern" -not -path './vendor/*' -not -path './.git/*' 2>/dev/null || true)

	if [ -n "$matches" ]; then
		echo "ERROR: found $description:"
		echo "$matches"
		ERRORS=$((ERRORS + 1))
	fi
}

echo "Checking test file naming conventions..."

check_pattern "*_extra_test.go" "_extra_test.go files (split by subject instead)"
check_pattern "*_bugfix_test.go" "_bugfix_test.go files (regression tests belong in parent test file)"
check_pattern "*coverage_test.go" "coverage_test.go files (name by what they test, not a metric)"

if [ "$ERRORS" -gt 0 ]; then
	echo "FAIL: $ERRORS naming convention violation(s) found."
	echo "See AGENTS.md Test Organization section for conventions."
	exit 1
fi

echo "OK: all test files follow naming conventions."
