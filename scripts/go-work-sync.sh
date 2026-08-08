#!/usr/bin/env bash
# Verifies that 'go work sync' is idempotent (running it twice produces no changes).
# Usage: ./scripts/go-work-sync.sh
# Requires GOEXPERIMENT=jsonv2 in the environment.

set -euo pipefail

if [ -z "${GOEXPERIMENT:-}" ]; then
  echo "WARNING: GOEXPERIMENT not set. Export GOEXPERIMENT=jsonv2 before running."
fi

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
