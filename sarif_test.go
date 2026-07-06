package finding

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func simpleSARIFReport() *Report {
	return &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
		},
	}
}

func assertRulePresent(t *testing.T, rules map[string]struct{}, ruleID string, wantPresent bool) {
	t.Helper()

	_, ok := rules[ruleID]
	if ok != wantPresent {
		if wantPresent {
			t.Errorf("expected %s in results", ruleID)
		} else {
			t.Errorf("%s should be excluded", ruleID)
		}
	}
}

func goFindingProps(
	id, severity, fixStrategy, toolName, category string,
	tags []any,
	confidence float64,
	suggestion, snippet string,
) map[string]any {
	return map[string]any{
		sarifPropID:          id,
		sarifPropSeverity:    severity,
		sarifPropFixStrategy: fixStrategy,
		sarifPropToolName:    toolName,
		sarifPropCategory:    category,
		sarifPropTags:        tags,
		sarifPropConfidence:  confidence,
		sarifPropSuggestion:  suggestion,
		sarifPropSnippet:     snippet,
	}
}

func goFindingPropsWithCustom(
	id, severity, fixStrategy, toolName, category string,
	tags []any,
	confidence float64,
	suggestion, snippet, customKey, customVal string,
) map[string]any {
	props := goFindingProps(
		id,
		severity,
		fixStrategy,
		toolName,
		category,
		tags,
		confidence,
		suggestion,
		snippet,
	)
	props[customKey] = customVal

	return props
}

// unmarshalSARIF unmarshals SARIF data and fails the test on error.
func unmarshalSARIF(t *testing.T, data []byte) *sarifLog {
	t.Helper()

	var log sarifLog

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
	g := gomega.NewWithT(t)

	r := &Report{
		Tool: ToolInfo{Name: "test-tool", Version: "1.0"},
		findings: []Finding{
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

	g.Expect(run.Results).To(gomega.HaveLen(1))

	result := run.Results[0]
	g.Expect(result.RuleID).To(gomega.Equal("SA1000"))
	g.Expect(string(SeverityError)).To(gomega.Equal(result.Level))
	g.Expect(result.Message.Text).To(gomega.Equal("bad code"))
	g.Expect(result.Locations[0].PhysicalLocation.ArtifactLocation.URI).To(gomega.Equal("main.go"))
}

func TestToSARIFFiltered(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	suppression := expiringSuppression(t, "test")

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
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

	data, err := r.ToSARIFWithOpts(WithMinSeverity(SeverityError))
	if err != nil {
		t.Fatalf("ToSARIFWithOpts: %v", err)
	}

	log := unmarshalSARIF(t, data)

	results := log.Runs[0].Results
	g.Expect(results).To(gomega.HaveLen(2))

	rules := make(map[string]struct{})
	for _, res := range results {
		rules[res.RuleID] = struct{}{}
	}

	assertRulePresent(t, rules, "r1", true)
	assertRulePresent(t, rules, "r2", true)
	assertRulePresent(t, rules, "r3", false)
	assertRulePresent(t, rules, "r4", false)
	assertRulePresent(t, rules, "r5", false)
}

func TestToSARIF_SuppressedFindingsExcluded(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	suppression := expiringSuppression(t, "won't fix")

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
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
	g.Expect(results).To(gomega.HaveLen(1))

	if results[0].RuleID != "r1" {
		t.Errorf("RuleID = %q, want %q", results[0].RuleID, "r1")
	}
}

func TestToSARIF_WithFix(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID:          "f1",
				Rule:        "fix-rule",
				Severity:    SeverityWarning,
				Position:    Position{File: "a.go", Line: 5, Column: 1},
				FixStrategy: FixStrategyDirect,
				BeforeCode:  "old",
				AfterCode:   "new",
				Suggestion:  "replace old with new",
				Range:       NewRangePtr("a.go", 5, 1, 5, 10),
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	result := log.Runs[0].Results[0]
	g.Expect(result.Fixes).To(gomega.HaveLen(1))

	fix := result.Fixes[0]
	if fix.Description.Text != "replace old with new" {
		t.Errorf("Fix description = %q, want %q", fix.Description.Text, "replace old with new")
	}
}

func TestToSARIF_WithMetadata(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID:       "f1",
				Rule:     "r1",
				Severity: SeverityInfo,
				Position: Position{File: "a.go"},
				Metadata: map[string]string{"key1": "val1"},
				Category: "security",
				Tags:     []Tag{"injection"},
				ToolName: "scanner",
			},
		},
	}

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	raw := string(data)
	if !strings.Contains(raw, `"`+sarifPropMetaPrefix+"key1"+`"`) || !strings.Contains(raw, `"val1"`) {
		t.Errorf("SARIF output should contain metadata, got: %s", raw)
	}

	if !strings.Contains(raw, `"`+sarifPropToolName+`"`) {
		t.Error("SARIF output should contain go-finding/toolName in properties")
	}
}

func TestToSARIF_SuggestionOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID:          "f1",
				Rule:        "r1",
				Severity:    SeverityWarning,
				Position:    Position{File: "a.go", Line: 5},
				FixStrategy: FixStrategySuggest,
				Suggestion:  "do better",
			},
		},
	}

	data, err := r.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	log := unmarshalSARIF(t, data)
	g.Expect(log.Runs[0].Results).To(gomega.HaveLen(1))

	result := log.Runs[0].Results[0]
	g.Expect(result.Fixes).To(gomega.HaveLen(1))
	g.Expect(result.Fixes[0].Description.Text).To(gomega.Equal("do better"))
	g.Expect(result.Fixes[0].Changes).To(gomega.BeEmpty())
}
