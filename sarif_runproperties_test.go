package finding

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestSetRunPropertyExportsRunLevelBag(t *testing.T) {
	report := NewReport(ToolInfo{Name: "io-surface", Version: "0.6.0"})
	report.AddFinding(Finding{
		ID:       "t1",
		ToolName: "io-surface",
		Rule:     "some-rule",
		Message:  "msg",
		Severity: SeverityWarning,
	})
	report.SetRunProperty("go-finding/meta/enforcement", "gate")

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	var log struct {
		Runs []struct {
			Properties map[string]any `json:"properties"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(sarif, &log); err != nil {
		t.Fatalf("unmarshal SARIF: %v", err)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(log.Runs))
	}
	if got := log.Runs[0].Properties["go-finding/meta/enforcement"]; got != "gate" {
		t.Fatalf("runs[0].properties[enforcement] = %v, want %q", got, "gate")
	}
}

func TestRunPropertiesOmittedWhenUnset(t *testing.T) {
	report := NewReport(ToolInfo{Name: "io-surface"})

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	if strings.Contains(string(sarif), `"properties":`) {
		// Per-result properties are legal; the zero-diff guarantee is about
		// the RUN level: no run object may carry a properties key when no
		// run property was recorded.
		var log struct {
			Runs []struct {
				Properties map[string]any `json:"properties"`
				Results    []struct {
					Properties map[string]any `json:"properties"`
				} `json:"results"`
			} `json:"runs"`
		}
		if err := json.Unmarshal(sarif, &log); err != nil {
			t.Fatalf("unmarshal SARIF: %v", err)
		}
		for i, run := range log.Runs {
			if run.Properties != nil {
				t.Fatalf("runs[%d].properties present without SetRunProperty (zero-diff contract)", i)
			}
		}
	}
}

func TestRunPropertiesRoundTripThroughSARIF(t *testing.T) {
	report := NewReport(ToolInfo{Name: "io-surface"})
	report.SetRunProperty("go-finding/meta/enforcement", "lint")
	report.SetRunProperty("scan-root", "/repo")

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	props, err := RunPropertiesFromSARIF(context.Background(), sarif)
	if err != nil {
		t.Fatalf("RunPropertiesFromSARIF(): %v", err)
	}
	if props["go-finding/meta/enforcement"] != "lint" {
		t.Fatalf("round-trip enforcement = %v, want %q", props["go-finding/meta/enforcement"], "lint")
	}
	if props["scan-root"] != "/repo" {
		t.Fatalf("round-trip scan-root = %v, want %q", props["scan-root"], "/repo")
	}
}

func TestRunPropertiesFromSARIFEmptyLog(t *testing.T) {
	props, err := RunPropertiesFromSARIF(context.Background(), []byte(`{"version":"2.1.0","runs":[]}`))
	if err != nil {
		t.Fatalf("RunPropertiesFromSARIF(): %v", err)
	}
	if props != nil {
		t.Fatalf("props = %v, want nil for a run-less log", props)
	}
}

func TestRunPropertiesCopyIsolation(t *testing.T) {
	report := NewReport(ToolInfo{Name: "io-surface"})
	report.SetRunProperty("k", "v1")

	got := report.RunProperties()
	got["k"] = "mutated"

	if report.RunProperties()["k"] != "v1" {
		t.Fatal("RunProperties() leaked the internal map: caller mutation visible")
	}
}
