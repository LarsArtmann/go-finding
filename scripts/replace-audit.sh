#!/usr/bin/env bash
# Verifies sub-module replace directives:
#   - library sub-modules (analysis, pipeline, toolsdk) MUST point at the
#     correct relative core path (GOWORK=off builds resolve local source),
#   - cmd/go-finding must have NO replace directives: Go refuses
#     `go install module@version` when the target module's go.mod carries
#     replaces, and the README promises that install path.
# Usage: ./scripts/replace-audit.sh
# Fails with exit code 1 if any replace directive is missing, points to the
# wrong path, or appears in cmd/go-finding.

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

check_no_replace() {
	local modfile="$1"

	if grep -qE '^replace |^\treplace ' "$modfile"; then
		echo "ERROR: $modfile must not contain replace directives (breaks go install module@version)"
		ERRORS=$((ERRORS + 1))
	fi
}

echo "Checking replace directives..."

check_replace "analysis/go.mod" "github.com/larsartmann/go-finding" "../"
check_replace "pipeline/go.mod" "github.com/larsartmann/go-finding" "../"
check_replace "toolsdk/go.mod" "github.com/larsartmann/go-finding" "../"
check_no_replace "cmd/go-finding/go.mod"

if [ "$ERRORS" -gt 0 ]; then
	echo "FAIL: $ERRORS replace directive issue(s) found."
	exit 1
fi

echo "OK: all replace directives correct."
