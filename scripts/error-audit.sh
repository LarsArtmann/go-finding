#!/usr/bin/env bash
# Gates the codebase on hierarchical error-handling violations via erraudit.
#
# Runs erraudit (github.com/larsartmann/erraudit) with --type-aware in every
# module of the workspace and fails on any violation. Intentional patterns are
# suppressed in code via //nolint:erraudit directives WITH a reason.
#
# DO NOT trust `erraudit nolint-audit .` as a removal oracle. Verified
# 2026-09-23: it reported 24 directives "safe to remove"; stripping them made
# this gate go RED (the type-aware detector flags every one of those sites —
# they suppress real violations). nolint-audit analyzes with go/parser only
# and cannot see type-aware violations, so it false-reports such directives
# as stale. The only safe staleness test: remove ONE directive, run THIS
# gate, observe red, restore. nolint-audit remains available as an advisory
# listing via: ./scripts/error-audit.sh --nolint-audit
#
# Do NOT add --enforce-samber-oops, --enforce-go-error-family, or
# --enforce-generic-return here: those flags contradict this project's
# documented error model (sentinel errors + FindingError + stdlib wrapping).
# See docs/research/erraudit-violation-analysis.md.
#
# Note: erraudit's repository is private, so this gate runs locally only.
# CI wiring is blocked until the tool is public or credentials are configured.
#
# Usage: ./scripts/error-audit.sh [--nolint-audit] [module-dir ...]
#        (default: all five workspace modules)

set -euo pipefail

if ! command -v erraudit >/dev/null 2>&1; then
	echo "ERROR: erraudit not found on PATH."
	echo "Fix: go install github.com/larsartmann/erraudit/cmd/erraudit@latest"
	exit 1
fi

export GOEXPERIMENT=jsonv2

NOLINT_AUDIT=0
MODULES=()
for arg in "$@"; do
	case "$arg" in
	--nolint-audit) NOLINT_AUDIT=1 ;;
	*) MODULES+=("$arg") ;;
	esac
done

echo "erraudit $(erraudit version 2>/dev/null | awk '{print $2}') — record this commit when the gate verdict matters."

if [ "$NOLINT_AUDIT" -eq 1 ]; then
	# ADVISORY ONLY — see the header for why this cannot gate.
	echo "ADVISORY nolint-audit listing (not a gate; stale reports are unreliable for type-aware sites):"
	erraudit nolint-audit .
	exit 0
fi

if [ "${#MODULES[@]}" -eq 0 ]; then
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
		# Filter erraudit's internal `[feature:logger]` debug lines: they are
		# engine noise printed to stderr on every failing run and bury the
		# actual violation locations.
		echo "$output" | grep -v '^\[feature:'
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
