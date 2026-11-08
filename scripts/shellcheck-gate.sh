#!/usr/bin/env bash
# Shellcheck gate for every repo shell script (scripts/ + scripts/hooks/).
# Clean is the only green: fixes went in 2026-11-08 (19 findings across 8
# scripts — dead variables, masked return values, glob-dash hazards, a
# sed case-pattern escape, an A && B || C trap, a word-splitting for-loop).
# Intentional single-quoted regexes carry inline `# shellcheck disable`
# directives with a reason comment.
#
# Usage: ./scripts/shellcheck-gate.sh
# Exit codes: 0 = clean, 1 = findings, 2 = shellcheck unavailable.
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v shellcheck >/dev/null; then
	echo "SKIP: shellcheck not installed (devShell provides it: nix develop)"
	exit 2
fi

if output=$(shellcheck scripts/*.sh scripts/hooks/* 2>&1); then
	echo "OK: shellcheck clean across scripts/*.sh and scripts/hooks/*"
	exit 0
fi
echo "FAIL: shellcheck findings:"
echo "$output"
exit 1
