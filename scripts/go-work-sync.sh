#!/usr/bin/env bash
# Verifies that 'go work sync' is idempotent (running it twice produces no changes).
# Usage: ./scripts/go-work-sync.sh
# encoding/json/v2 is GA since Go 1.27; no GOEXPERIMENT needed.

set -euo pipefail

echo "Running go work sync (first pass)..."
go work sync

echo "Running go work sync (second pass — should be a no-op)..."
go work sync

# Check if go.work or any go.mod changed after the second sync.
if ! git diff --exit-code go.work */go.mod >/dev/null 2>&1; then
	echo "ERROR: go work sync produced changes on second run (not idempotent)."
	echo "Diff:"
	git diff go.work */go.mod
	exit 1
fi

echo "OK: go work sync is idempotent."
