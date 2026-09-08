#!/usr/bin/env bash
# Pre-tag release gate. Runs every structural check that must pass BEFORE a
# release tag exists. Born from the v1.7.0 incident: sub-module go.mod files
# were tagged with a stale core reference because the checklist lived in the
# operator's head, not in a script (see docs/status/2026-09-08_18-48 report d/1).
#
# Usage:
#   ./scripts/release-preflight.sh                 # structural gates (fast)
#   ./scripts/release-preflight.sh --bench         # also run bench-check vs baseline
#   ./scripts/release-preflight.sh --stress        # also run repeat=20 race stress
#   ./scripts/release-preflight.sh --post-tag      # post-tag verification mode
#
# Always run from the repo root (relative paths).
# Heavy gates (bench, stress) are opt-in: preflight's job is to make the
# cheap-to-run, easy-to-forget checks impossible to skip.
#
# --post-tag inverts the tag-collision guard (all 4 tags must now EXIST)
# and additionally runs version-check.sh (version.go == latest tag). Run it
# right after tagging, before `git push origin --tags`.
#
# Self-test: scripts/release-preflight-selftest.sh injects synthetic drift
# and a tag collision into a disposable git worktree and asserts preflight
# FAILS on each — proving the gate bites.
#
# Verdict discipline: every check prints an explicit OK/FAIL line and the
# script prints a final verdict. A check that prints nothing is a bug in this
# script (dead-gate lesson, 2026-09-08).

set -euo pipefail

cd "$(dirname "$0")/.."

RUN_BENCH=0
RUN_STRESS=0
POST_TAG=0
for arg in "$@"; do
	case "$arg" in
	--bench) RUN_BENCH=1 ;;
	--stress) RUN_STRESS=1 ;;
	--post-tag) POST_TAG=1 ;;
	*)
		echo "usage: $0 [--bench] [--stress] [--post-tag]"
		exit 2
		;;
	esac
done

FAILURES=0

step() { printf '\n== %s ==\n' "$1"; }

run_check() {
	local name="$1"
	shift
	step "$name"
	if "$@"; then
		echo "OK: $name"
	else
		echo "FAIL: $name"
		FAILURES=$((FAILURES + 1))
	fi
}

echo "Release preflight $(date '+%Y-%m-%d %H:%M:%S')"

# --- Cheap, structural, always ---

step "working tree clean"
if [ -z "$(git status --porcelain)" ]; then
	echo "OK: working tree clean"
else
	echo "FAIL: working tree has uncommitted changes — commit or stash first:"
	git status --porcelain
	FAILURES=$((FAILURES + 1))
fi

step "version.go parses"
MAJOR=$(sed -n 's/^const VersionMajor = //p' version.go)
MINOR=$(sed -n 's/^const VersionMinor = //p' version.go)
PATCH=$(sed -n 's/^const VersionPatch = //p' version.go)
NEXT_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
if [[ "$MAJOR" =~ ^[0-9]+$ && "$MINOR" =~ ^[0-9]+$ && "$PATCH" =~ ^[0-9]+$ ]]; then
	echo "OK: version.go says $NEXT_VERSION"
else
	echo "FAIL: version.go did not parse (got '$NEXT_VERSION')"
	FAILURES=$((FAILURES + 1))
fi

step "target tags"
for t in "v${MAJOR}.${MINOR}.${PATCH}" "pipeline/v${MAJOR}.${MINOR}.${PATCH}" "analysis/v${MAJOR}.${MINOR}.${PATCH}" "cmd/go-finding/v${MAJOR}.${MINOR}.${PATCH}"; do
	if [ "$POST_TAG" -eq 1 ]; then
		if git rev-parse -q --verify "refs/tags/$t" >/dev/null; then
			echo "OK: tag $t exists (post-tag mode)"
		else
			echo "FAIL: tag $t missing (post-tag mode expects it to exist)"
			FAILURES=$((FAILURES + 1))
		fi
	elif git rev-parse -q --verify "refs/tags/$t" >/dev/null; then
		echo "FAIL: tag $t already exists"
		FAILURES=$((FAILURES + 1))
	else
		echo "OK: tag $t is new"
	fi
done

if [ "$POST_TAG" -eq 1 ]; then
	run_check "version.go matches latest tag" bash scripts/version-check.sh
fi

run_check "version drift (go.mod references match version.go)" bash scripts/version-drift.sh
run_check "replace directives point at ../" bash scripts/replace-audit.sh
run_check "go.work matches directory structure" bash scripts/go-work-sync.sh
run_check "test file naming" bash scripts/test-naming.sh
run_check "JSON deterministic marshal" bash scripts/json-deterministic-check.sh
run_check "docs API references" bash scripts/docs-api-check.sh
run_check "docs freshness" bash scripts/docs-freshness.sh

step "GOWORK=off build per module"
for dir in . pipeline analysis cmd/go-finding; do
	if (cd "$dir" && GOWORK=off GOEXPERIMENT=jsonv2 go build ./...); then
		echo "OK: GOWORK=off build in $dir"
	else
		echo "FAIL: GOWORK=off build in $dir"
		FAILURES=$((FAILURES + 1))
	fi
done

step "go mod tidy -diff per module (tidy-ness)"
for dir in . pipeline analysis cmd/go-finding; do
	if (cd "$dir" && GOWORK=off GOEXPERIMENT=jsonv2 GOPRIVATE='github.com/larsartmann/*' go mod tidy -diff); then
		echo "OK: go.mod tidy in $dir"
	else
		echo "FAIL: $dir/go.mod needs 'go mod tidy' (run it with GOWORK=off)"
		FAILURES=$((FAILURES + 1))
	fi
done

# --- Opt-in heavy gates ---

if [ "$RUN_BENCH" -eq 1 ]; then
	step "benchmark capture (core + pipeline, count=10)"
	if {
		GOEXPERIMENT=jsonv2 go test -run='^$' -bench=. -benchmem -count=10 ./... > /tmp/preflight-bench.txt
		(cd pipeline && GOEXPERIMENT=jsonv2 go test -run='^$' -bench=. -benchmem -count=10 ./... >> /tmp/preflight-bench.txt)
	}; then
		echo "OK: benchmark capture"
	else
		echo "FAIL: benchmark capture"
		FAILURES=$((FAILURES + 1))
	fi
	step "benchmark vs baseline"
	if bash scripts/bench-check.sh benchmarks/baseline.txt /tmp/preflight-bench.txt 25; then
		echo "OK: bench within threshold"
	else
		echo "FAIL: bench regression (or capture error)"
		FAILURES=$((FAILURES + 1))
	fi
fi

if [ "$RUN_STRESS" -eq 1 ]; then
	step "stress: ginkgo repeat=20 (core)"
	if (GOEXPERIMENT=jsonv2 ginkgo -r --race --repeat=20 --skip-package=examples); then
		echo "OK: core stress"
	else
		echo "FAIL: core stress"
		FAILURES=$((FAILURES + 1))
	fi
	step "stress: ginkgo repeat=20 (pipeline)"
	if (cd pipeline && GOEXPERIMENT=jsonv2 ginkgo -r --race --repeat=20 --skip-package=examples); then
		echo "OK: pipeline stress"
	else
		echo "FAIL: pipeline stress"
		FAILURES=$((FAILURES + 1))
	fi
	for dir in analysis cmd/go-finding; do
		step "stress: go test -race -count=20 ($dir)"
		if (cd "$dir" && GOEXPERIMENT=jsonv2 go test -race -count=20 ./...); then
			echo "OK: $dir stress"
		else
			echo "FAIL: $dir stress"
			FAILURES=$((FAILURES + 1))
		fi
	done
fi

# --- Verdict ---

printf '\n================================\n'
if [ "$FAILURES" -gt 0 ]; then
	echo "PREFLIGHT FAIL: $FAILURES check(s) failed. Do NOT tag."
	exit 1
fi
echo "PREFLIGHT PASS: all checks green. Safe to tag $NEXT_VERSION (+ pipeline/ analysis/ cmd/go-finding/ prefixed variants)."
if [ "$POST_TAG" -eq 1 ]; then
	echo "Post-tag mode: all 4 tags exist and version.go matches. Push with: git push origin --tags"
else
	echo "Reminder: stress gate is mandatory before tagging (release-procedure step 4)."
fi
exit 0
