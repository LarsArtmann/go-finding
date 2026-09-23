package finding

// summary_coverage_test.go pins the Summary coverage contract (2026-09-22,
// erraudit TODO f2/f3): FilesScanned is ALWAYS serialized — "scanned 0" is
// load-bearing coverage evidence and must stay distinguishable from an
// omitted field — and SkippedModules serializes only when set, so
// module-agnostic tools produce unchanged JSON.

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSummaryJSON_FilesScannedAlwaysEmitted(t *testing.T) {
	data, err := json.Marshal(Summary{})
	if err != nil {
		t.Fatalf("marshal zero Summary: %v", err)
	}

	if !strings.Contains(string(data), `"filesScanned":0`) {
		t.Fatalf("filesScanned must be emitted even at 0 (coverage evidence), got: %s", data)
	}
}

func TestSummaryJSON_SkippedModulesRoundTrip(t *testing.T) {
	withModules, err := json.Marshal(Summary{FilesScanned: 9, SkippedModules: []string{"sub", "tools/x"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(withModules), `"skippedModules":["sub","tools/x"]`) {
		t.Fatalf("skippedModules must serialize when set, got: %s", withModules)
	}

	empty, err := json.Marshal(Summary{})
	if err != nil {
		t.Fatalf("marshal zero Summary: %v", err)
	}

	if strings.Contains(string(empty), "skippedModules") {
		t.Fatalf("skippedModules must be omitted when nil (module-agnostic tools), got: %s", empty)
	}
}
