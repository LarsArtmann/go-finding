#!/usr/bin/env bash
# Verifies all sub-module replace directives point at correct relative paths.
# Usage: ./scripts/replace-audit.sh
# Fails with exit code 1 if any replace directive is missing or points to the wrong path.

set -euo pipefail

ERRORS=0

check_replace() {
	local modfile="$1"
	local module="$2"
	local expected_path="$3"

	# Match both single-line "replace module => path" and block-style
	# "replace (\n module => path\n)" directives.
	if ! grep -qE "${module} => ${expected_path}" "$modfile"; then
		echo "ERROR: $modfile missing 'replace ${module} => ${expected_path}'"
		ERRORS=$((ERRORS + 1))
	fi
}

echo "Checking replace directives..."

check_replace "analysis/go.mod" "github.com/larsartmann/go-finding" "../"
check_replace "pipeline/go.mod" "github.com/larsartmann/go-finding" "../"
check_replace "cmd/go-finding/go.mod" "github.com/larsartmann/go-finding" "../.."
check_replace "cmd/go-finding/go.mod" "github.com/larsartmann/go-finding/pipeline" "../../pipeline"

if [ "$ERRORS" -gt 0 ]; then
	echo "FAIL: $ERRORS replace directive issue(s) found."
	exit 1
fi

echo "OK: all replace directives correct."
