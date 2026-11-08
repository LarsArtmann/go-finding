#!/usr/bin/env bash
# Consumer compatibility check: simulates an external consumer resolving each
# published go-finding module from the public module proxy — no replace
# directives, no workspace, no GOPRIVATE. Turns the "plain go get works"
# promise into a mechanical gate.
#
# Usage:
#   bash scripts/consumer-compat-check.sh [version]   # proxy mode (default)
#   bash scripts/consumer-compat-check.sh --head      # at-HEAD mode
#
#   version defaults to "latest"; pass e.g. v1.10.0 to pin.
#
# At-HEAD mode (--head) swaps the proxy for `replace` directives against the
# WORKING TREE: a fresh consumer module compiles, vets, and runs the same
# probes against the current commit's API surface. This catches API drift
# BEFORE a train — proxy mode can only prove versions that are already
# published. The CLI probe builds the checked-out cmd module and asserts
# --version prints EXACTLY the version.go stamp (the loose semver grep of
# proxy mode cannot catch the v1.15.1 self-reports-1.15.0 class).
#
# No GOEXPERIMENT needed: encoding/json/v2 is GA since Go 1.27, which the
# go.mod floor already requires.
# Fails with exit code 1 if any module fails to resolve, build, vet, or run.

set -euo pipefail

MODE=proxy
VERSION=latest
for arg in "$@"; do
	case "$arg" in
	--head) MODE=head ;;
	latest | v*) VERSION="$arg" ;;
	*)
		echo "usage: $0 [--head] [version]"
		exit 2
		;;
	esac
done

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

FAILURES=0

verdict() {
	if [ "$1" -eq 0 ]; then
		echo "OK: $2"
	else
		echo "FAIL: $2"
		FAILURES=$((FAILURES + 1))
	fi
}

check_import_module() {
	local name="$1" module="$2" probe_file="$3"
	local dir="$WORK/$name"
	local rc=0
	mkdir -p "$dir"
	(
		cd "$dir"
		go mod init compat.local >/dev/null 2>&1
		cp "$probe_file" main.go
		if [ "$MODE" = head ]; then
			# Resolve the module (and core, its dependency root) from the
			# working tree — the at-HEAD API-drift check. The probe source is
			# copied FIRST: `go mod tidy` derives the require from actual
			# imports, and a replace alone adds nothing. The suffix keeps its
			# leading slash — appending another would double it.
			printf '\nreplace github.com/larsartmann/go-finding => %s\n' "$REPO_ROOT" >>go.mod
			if [ "$module" != "github.com/larsartmann/go-finding" ]; then
				printf 'replace %s => %s\n' "$module" "$REPO_ROOT${module#github.com/larsartmann/go-finding}" >>go.mod
			fi
			GOWORK=off go mod tidy >/dev/null 2>&1
		else
			go get "${module}@${VERSION}" >/dev/null
		fi
		echo "ok" >expected.txt
		go vet . >/dev/null 2>&1
		go run . >actual.txt 2>err.txt
		diff -q expected.txt actual.txt >/dev/null
	) >/dev/null 2>&1 || rc=$?
	if [ "$rc" -ne 0 ] && [ "${COMPAT_DEBUG:-0}" = 1 ]; then
		echo "--- $name go.mod:" >&2
		cat "$dir/go.mod" >&2
		echo "--- $name err.txt:" >&2
		cat "$dir/err.txt" 2>/dev/null >&2 || true
	fi
	if [ "$MODE" = head ]; then
		verdict "$rc" "$name builds, vets, and runs against the WORKING TREE at-HEAD ($module)"
	else
		verdict "$rc" "$name resolves from proxy, builds, vets, and runs (${module}@${VERSION})"
	fi
}

check_cli_module() {
	local name="$1" module="$2"
	local dir="$WORK/$name"
	local rc=0
	mkdir -p "$dir"
	(
		cd "$dir"
		export GOBIN="$dir/bin"
		mkdir -p "$GOBIN"
		if [ "$MODE" = head ]; then
			# Build the CLI as an at-HEAD consumer would see it: a wrapper
			# module replacing core AND cmd onto the working tree, so the
			# binary is stamped from the LOCAL version.go — not from the
			# published (possibly stamp-defective) core tag. Then assert the
			# EXACT stamp: this probe would have failed red on the v1.15.1
			# self-reports-1.15.0 class at build time.
			go mod init compat.local >/dev/null 2>&1
			printf 'require %s v0.0.0-local\n\nreplace %s => %s\nreplace github.com/larsartmann/go-finding => %s\n' \
				"$module" "$module" "$REPO_ROOT/cmd/go-finding" "$REPO_ROOT" >>go.mod
			GOWORK=off go build -o "$GOBIN/go-finding" "$module" >/dev/null 2>&1
			local stamp
			stamp="v$(sed -n 's/^const VersionMajor = //p' "$REPO_ROOT/version.go").$(sed -n 's/^const VersionMinor = //p' "$REPO_ROOT/version.go").$(sed -n 's/^const VersionPatch = //p' "$REPO_ROOT/version.go")"
			# finding.Version has no leading v; the stamp does.
			[ "$("$GOBIN/go-finding" -version)" = "${stamp#v}" ]
		else
			go install "${module}@${VERSION}" >/dev/null
			"$GOBIN/go-finding" -version | grep -Eq '[0-9]+\.[0-9]+\.[0-9]+'
		fi
	) >/dev/null 2>&1 || rc=$?
	if [ "$rc" -ne 0 ] && [ "${COMPAT_DEBUG:-0}" = 1 ] && [ "$MODE" = head ]; then
		echo "--- cli go.mod:" >&2; cat "$dir/go.mod" >&2
		echo "--- binary says:" >&2; "$dir/bin/go-finding" -version 2>&1 >&2 || true
		echo "--- expected stamp:" >&2
	fi
	if [ "$MODE" = head ]; then
		verdict "$rc" "$name builds from the working tree and -version prints the exact version.go stamp"
	else
		verdict "$rc" "$name installs from proxy and -version prints semver (${module}@${VERSION})"
	fi
}

PROBES="$(mktemp -d)"
trap 'rm -rf "$WORK" "$PROBES"' EXIT

cat >"$PROBES/core.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/larsartmann/go-finding"
)

func main() {
	if finding.Version == "" {
		fmt.Fprintln(os.Stderr, "empty version")
		os.Exit(1)
	}
	fmt.Println("ok")
}
EOF

cat >"$PROBES/pipeline.go" <<'EOF'
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/larsartmann/go-finding/pipeline"
)

func main() {
	findings, err := pipeline.Detect(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "detect:", err)
		os.Exit(1)
	}
	if len(findings) != 0 {
		fmt.Fprintln(os.Stderr, "expected zero findings without detectors")
		os.Exit(1)
	}
	fmt.Println("ok")
}
EOF

cat >"$PROBES/analysis.go" <<'EOF'
package main

import (
	"fmt"

	"github.com/larsartmann/go-finding/analysis"
)

var _ = analysis.FromDiagnostic

func main() {
	fmt.Println("ok")
}
EOF

cat >"$PROBES/toolsdk.go" <<'EOF'
package main

import (
	"fmt"

	"github.com/larsartmann/go-finding/toolsdk"
)

var _ toolsdk.Spec

func main() {
	fmt.Println("ok")
}
EOF

echo "Consumer compatibility check (proxy resolution + vet, no replaces/workspace/GOPRIVATE)..."

check_import_module core "github.com/larsartmann/go-finding" "$PROBES/core.go"
check_import_module pipeline "github.com/larsartmann/go-finding/pipeline" "$PROBES/pipeline.go"
check_import_module analysis "github.com/larsartmann/go-finding/analysis" "$PROBES/analysis.go"
check_import_module toolsdk "github.com/larsartmann/go-finding/toolsdk" "$PROBES/toolsdk.go"
check_cli_module cli "github.com/larsartmann/go-finding/cmd/go-finding"

if [ "$FAILURES" -gt 0 ]; then
	echo "FAIL: $FAILURES module(s) failed the consumer compatibility check."
	exit 1
fi
echo "PASS: all modules resolve, build, vet, and run as an external consumer."
