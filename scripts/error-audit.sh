#!/usr/bin/env bash
# Gates the codebase on hierarchical error-handling violations via erraudit.
#
# Runs erraudit (github.com/larsartmann/erraudit) with --type-aware in every
# module of the workspace and fails on any violation. Intentional patterns are
# suppressed in code via //nolint:erraudit directives WITH a reason; audit the
# directives for staleness with: erraudit nolint-audit ./...
#
# Do NOT add --enforce-samber-oops, --enforce-go-error-family, or
# --enforce-generic-return here: those flags contradict this project's
# documented error model (sentinel errors + FindingError + stdlib wrapping).
# See docs/research/erraudit-violation-analysis.md.
#
# Note: erraudit's repository is private, so this gate runs locally only.
# CI wiring is blocked until the tool is public or credentials are configured.
#
# Usage: ./scripts/error-audit.sh [module-dir ...]
#        (default: all five workspace modules)

set -euo pipefail

if ! command -v erraudit >/dev/null 2>&1; then
	echo "ERROR: erraudit not found on PATH."
	echo "Fix: go install github.com/larsartmann/erraudit/cmd/erraudit@latest"
	exit 1
fi

export GOEXPERIMENT=jsonv2

if [ "$#" -gt 0 ]; then
	MODULES=("$@")
else
	MODULES=(. pipeline analysis toolsdk cmd/go-finding)
fi

FAILED=0
for mod in "${MODULES[@]}"; do
	if [ ! -f "$mod/go.mod" ]; then
		echo "ERROR: missing go.mod in module dir: $mod"
		exit 1
	fi

	if ! output=$(cd "$mod" && erraudit ./... --type-aware 2>&1); then
		FAILED=1
		echo "--- erraudit output for $mod ---"
		echo "$output"
		echo "--- end output for $mod ---"
	else
		echo "$output" | grep "^Total Violations" | sed "s|^|$mod: |"
	fi
done

if [ "$FAILED" -ne 0 ]; then
	echo "FAIL: erraudit found error-handling violations (full output above)."
	echo "Fix the flagged sites, or add //nolint:erraudit with a reason for intentional patterns."
	exit 1
fi

echo "OK: erraudit clean across ${#MODULES[@]} module(s) (no error-handling violations)."
