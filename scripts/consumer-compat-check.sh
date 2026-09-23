#!/usr/bin/env bash
# Consumer compatibility check: simulates an external consumer resolving each
# published go-finding module from the public module proxy — no replace
# directives, no workspace, no GOPRIVATE. Turns the "plain go get works"
# promise into a mechanical gate.
#
# Usage: bash scripts/consumer-compat-check.sh [version]
#   version defaults to "latest"; pass e.g. v1.10.0 to pin.
#
# No GOEXPERIMENT needed: encoding/json/v2 is GA since Go 1.27, which the
# go.mod floor already requires.
# Fails with exit code 1 if any module fails to resolve, build, or run.

set -euo pipefail

VERSION="${1:-latest}"

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
	mkdir -p "$dir"
	(
		cd "$dir"
		go mod init compat.local >/dev/null 2>&1
		go get "${module}@${VERSION}" >/dev/null
		cp "$probe_file" main.go
		echo "ok" >expected.txt
		go run . >actual.txt 2>err.txt
		diff -q expected.txt actual.txt >/dev/null
	) >/dev/null 2>&1
	verdict $? "$name resolves from proxy, builds, and runs (${module}@${VERSION})"
}

check_cli_module() {
	local name="$1" module="$2"
	local dir="$WORK/$name"
	mkdir -p "$dir"
	(
		cd "$dir"
		export GOBIN="$dir/bin"
		mkdir -p "$GOBIN"
		go install "${module}@${VERSION}" >/dev/null
		"$GOBIN/go-finding" -version | grep -Eq '[0-9]+\.[0-9]+\.[0-9]+'
	) >/dev/null 2>&1
	verdict $? "$name installs from proxy and -version prints semver (${module}@${VERSION})"
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

echo "Consumer compatibility check (proxy resolution, no replaces/workspace/GOPRIVATE)..."

check_import_module core "github.com/larsartmann/go-finding" "$PROBES/core.go"
check_import_module pipeline "github.com/larsartmann/go-finding/pipeline" "$PROBES/pipeline.go"
check_import_module analysis "github.com/larsartmann/go-finding/analysis" "$PROBES/analysis.go"
check_import_module toolsdk "github.com/larsartmann/go-finding/toolsdk" "$PROBES/toolsdk.go"
check_cli_module cli "github.com/larsartmann/go-finding/cmd/go-finding"

if [ "$FAILURES" -gt 0 ]; then
	echo "FAIL: $FAILURES module(s) failed the consumer compatibility check."
	exit 1
fi
echo "PASS: all 5 modules resolve, build, and run as an external consumer."
