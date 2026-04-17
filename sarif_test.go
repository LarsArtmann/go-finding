package finding

import (
	"encoding/json"
	"strings"
	"testing"
)

// unmarshalSARIF unmarshals SARIF data and fails the test on error.
func unmarshalSARIF(t *testing.T, data []byte) *SarifLog {
	t.Helper()

	var log SarifLog

	err := json.Unmarshal(data, &log)
	if err != nil {
		t.Fatalf("unmarshal SARIF: %v", err)
	}

	return &log
}

func TestFromSARIFLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
		want  Severity
	}{
		{"note maps to info", "note", SeverityInfo},
		{"warning maps to warning", "warning", SeverityWarning},
		{"error maps to error", "error", SeverityError},
		{"unknown maps to warning", "unknown", SeverityWarning},
		{"empty maps to warning", "", SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := FromSARIFLevel(tt.level); got != tt.want {
				t.Errorf("FromSARIFLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestToSARIF(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "test-tool", Version: "1.0"},
		Findings: []Finding{
			{
				ID:       "f1",
				Rule:     "SA1000",
				Message:  "bad code",
				Severity: SeverityError,
				Position: Position{File: "main.go", Line: 10, Column: 5},
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	if log.Version != "2.1.0" {
		t.Errorf("Version = %q, want %q", log.Version, "2.1.0")
	}

	if len(log.Runs) != 1 {
		t.Fatalf("Runs length = %d, want 1", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "test-tool" {
		t.Errorf("Driver.Name = %q, want %q", run.Tool.Driver.Name, "test-tool")
	}

	if len(run.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(run.Results))
	}

	result := run.Results[0]
	if result.RuleID != "SA1000" {
		t.Errorf("RuleID = %q, want %q", result.RuleID, "SA1000")
	}

	if result.Level != "error" {
		t.Errorf("Level = %q, want %q", result.Level, "error")
	}

	if result.Message.Text != "bad code" {
		t.Errorf("Message.Text = %q, want %q", result.Message.Text, "bad code")
	}

	uri := result.Location.PhysicalLocation.ArtifactLocation.URI
	if uri != "main.go" {
		t.Errorf("URI = %q, want %q", uri, "main.go")
	}
}

func TestToSARIFFiltered(t *testing.T) {
	t.Parallel()

	suppression := expiringSuppression(t, "test")

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{ID: "f1", Severity: SeverityCritical, Rule: "r1", Position: Position{File: "a.go"}},
			{ID: "f2", Severity: SeverityError, Rule: "r2", Position: Position{File: "a.go"}},
			{ID: "f3", Severity: SeverityWarning, Rule: "r3", Position: Position{File: "a.go"}},
			{ID: "f4", Severity: SeverityInfo, Rule: "r4", Position: Position{File: "a.go"}},
			{
				ID:          "f5",
				Severity:    SeverityError,
				Rule:        "r5",
				Position:    Position{File: "a.go"},
				Suppression: suppression,
			},
		},
	}

	data, err := r.ToSARIFFiltered(SeverityError)
	if err != nil {
		t.Fatalf("ToSARIFFiltered: %v", err)
	}

	log := unmarshalSARIF(t, data)

	results := log.Runs[0].Results
	if len(results) != 2 {
		t.Fatalf(
			"Results length = %d, want 2 (critical + error, excluding warning/info/suppressed)",
			len(results),
		)
	}

	rules := make(map[string]bool)
	for _, res := range results {
		rules[res.RuleID] = true
	}

	if !rules["r1"] || !rules["r2"] {
		t.Errorf("expected r1 and r2 in results, got rules: %v", rules)
	}

	if rules["r3"] || rules["r4"] || rules["r5"] {
		t.Error("r3 (warning), r4 (info), r5 (suppressed) should be excluded")
	}
}

func TestToSARIF_SuppressedFindingsExcluded(t *testing.T) {
	t.Parallel()

	suppression := expiringSuppression(t, "won't fix")

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{ID: "f1", Rule: "r1", Severity: SeverityError, Position: Position{File: "a.go"}},
			{
				ID:          "f2",
				Rule:        "r2",
				Severity:    SeverityError,
				Position:    Position{File: "a.go"},
				Suppression: suppression,
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	results := log.Runs[0].Results
	if len(results) != 1 {
		t.Fatalf("Results length = %d, want 1 (suppressed excluded)", len(results))
	}

	if results[0].RuleID != "r1" {
		t.Errorf("RuleID = %q, want %q", results[0].RuleID, "r1")
	}
}

func TestToSARIF_WithFix(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID:          "f1",
				Rule:        "fix-rule",
				Severity:    SeverityWarning,
				Position:    Position{File: "a.go", Line: 5, Column: 1},
				FixStrategy: FixStrategyDirect,
				BeforeCode:  "old",
				AfterCode:   "new",
				Suggestion:  "replace old with new",
				Range:       MakeRangePtr("a.go", 5, 1, 5, 10),
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	result := log.Runs[0].Results[0]
	if len(result.Fixes) != 1 {
		t.Fatalf("Fixes length = %d, want 1", len(result.Fixes))
	}

	fix := result.Fixes[0]
	if fix.Description.Text != "replace old with new" {
		t.Errorf("Fix description = %q, want %q", fix.Description.Text, "replace old with new")
	}
}

func TestToSARIF_WithMetadata(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID:       "f1",
				Rule:     "r1",
				Severity: SeverityInfo,
				Position: Position{File: "a.go"},
				Metadata: map[string]string{"key1": "val1"},
				Category: "security",
				Tag:      "injection",
				ToolName: "scanner",
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, `"key1"`) || !strings.Contains(raw, `"val1"`) {
		t.Errorf("SARIF output should contain metadata, got: %s", raw)
	}

	if !strings.Contains(raw, `"go-finding/toolName"`) {
		t.Error("SARIF output should contain go-finding/toolName in properties")
	}
}

func TestToSARIF_EmptyReport(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool:     ToolInfo{Name: "tool"},
		Findings: []Finding{},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(log.Runs[0].Results))
	}
}

func TestToSARIF_RoundTripProperties(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "govet:printf:main.go:10:5",
		Rule:        "printf",
		ToolName:    "govet",
		Message:     "invalid format",
		Severity:    SeverityCritical,
		Position:    Position{File: "main.go", Line: 10, Column: 5},
		Category:    CategoryCorrectness,
		Tag:         "printf",
		FixStrategy: FixStrategySuggest,
		Confidence:  0.9,
		Suggestion:  "fix format string",
		Snippet:     "fmt.Sprintf(\"%d\")",
		Metadata:    map[string]string{"custom": "value"},
	}

	report := NewReport(ToolInfo{Name: "govet"})
	report.AddFinding(f)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	var log struct {
		Runs []struct {
			Results []struct {
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}
	if uErr := json.Unmarshal(sarif, &log); uErr != nil {
		t.Fatalf("unmarshal: %v", uErr)
	}

	props := log.Runs[0].Results[0].Properties

	checks := map[string]any{
		"go-finding/id":          "govet:printf:main.go:10:5",
		"go-finding/severity":    "critical",
		"go-finding/fixStrategy": "suggest",
		"go-finding/toolName":    "govet",
		"go-finding/category":    "correctness",
		"go-finding/tag":         "printf",
		"go-finding/confidence":  0.9,
		"go-finding/suggestion":  "fix format string",
		"go-finding/snippet":     "fmt.Sprintf(\"%d\")",
		"custom":                 "value",
	}

	for key, want := range checks {
		got, ok := props[key]
		if !ok {
			t.Errorf("missing property %q", key)

			continue
		}
		// JSON numbers unmarshal as float64
		if f64, ok := want.(float64); ok {
			if gotFloat, ok := got.(float64); !ok || gotFloat != f64 {
				t.Errorf("properties[%q] = %v, want %v", key, got, want)
			}
		} else if got != want {
			t.Errorf("properties[%q] = %v, want %v", key, got, want)
		}
	}
}
