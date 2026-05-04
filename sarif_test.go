package finding

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	id, severity, fixStrategy, toolName, category, tag string,
	confidence float64,
	suggestion, snippet string,
) map[string]any {
	return map[string]any{
		"go-finding/id":          id,
		"go-finding/severity":    severity,
		"go-finding/fixStrategy": fixStrategy,
		"go-finding/toolName":    toolName,
		"go-finding/category":    category,
		"go-finding/tag":         tag,
		"go-finding/confidence":  confidence,
		"go-finding/suggestion":  suggestion,
		"go-finding/snippet":     snippet,
	}
}

func goFindingPropsWithCustom(
	id, severity, fixStrategy, toolName, category, tag string,
	confidence float64,
	suggestion, snippet, customKey, customVal string,
) map[string]any {
	props := goFindingProps(
		id,
		severity,
		fixStrategy,
		toolName,
		category,
		tag,
		confidence,
		suggestion,
		snippet,
	)
	props[customKey] = customVal
	return props
}

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

	require.Len(t, run.Results, 1, "Results")

	result := run.Results[0]
	assert.Equal(t, "SA1000", result.RuleID)
	assert.Equal(t, result.Level, string(SeverityError))
	assert.Equal(t, "bad code", result.Message.Text)
	assert.Equal(t, "main.go", result.Locations[0].PhysicalLocation.ArtifactLocation.URI)
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
	require.Len(t, results, 2, "filtered results")

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
	require.Len(t, results, 1, "suppressed excluded results")

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
	require.Len(t, result.Fixes, 1, "Fixes")

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

func TestToSARIF_SuggestionOnly(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
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
	require.NoError(t, err)

	log := unmarshalSARIF(t, data)
	require.Len(t, log.Runs[0].Results, 1)

	result := log.Runs[0].Results[0]
	require.Len(t, result.Fixes, 1)
	assert.Equal(t, "do better", result.Fixes[0].Description.Text)
	assert.Empty(t, result.Fixes[0].Changes)
}

func TestToSARIF_WithRelated(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID:       "f1",
				Rule:     "r1",
				Severity: SeverityError,
				Position: Position{File: "a.go", Line: 10},
				Related: []RelatedRef{
					{
						FindingID: "f2", Relation: "causes",
						Position: Position{File: "b.go", Line: 20},
					},
				},
			},
		},
	}

	data, err := r.ToSARIF()
	require.NoError(t, err)

	log := unmarshalSARIF(t, data)
	require.Len(t, log.Runs[0].Results, 1)

	result := log.Runs[0].Results[0]
	require.Len(t, result.Related, 1)
	assert.Equal(t, "b.go", result.Related[0].PhysicalLocation.ArtifactLocation.URI)
	assert.Equal(t, 20, result.Related[0].PhysicalLocation.Region.StartLine)
}

func TestToSARIF_EmptyReport(t *testing.T) {
	t.Parallel()

	r := MakeSimpleReport("tool")

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(log.Runs[0].Results))
	}
}

func TestToSARIF_ErrorPath(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m", Severity: SeverityError,
				Position: Position{File: "a.go"}, Confidence: math.NaN(),
			},
		},
	}

	_, err := r.ToSARIF()
	require.Error(t, err, "expected error for NaN confidence")
}

func TestToSARIFFiltered_ErrorPath(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m", Severity: SeverityError,
				Position: Position{File: "a.go"}, Confidence: math.NaN(),
			},
		},
	}

	_, err := r.ToSARIFFiltered(SeverityError)
	require.Error(t, err, "expected error for NaN confidence")
}

func TestWriteSARIF(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
		},
	}

	var buf strings.Builder
	err := r.WriteSARIF(&buf)
	require.NoError(t, err)

	data := buf.String()
	assert.Contains(t, data, `"version": "2.1.0"`)
	assert.Contains(t, data, `"tool"`)
}

func TestWriteSARIFFiltered(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
			{
				ID: "f2", Rule: "r2", Message: "m",
				Severity: SeverityInfo, Position: Position{File: "b.go"},
			},
		},
	}

	var buf strings.Builder
	err := r.WriteSARIFFiltered(&buf, SeverityWarning)
	require.NoError(t, err)

	data := buf.String()
	assert.Contains(t, data, "f1")
	assert.NotContains(t, data, "f2")
}

func TestFindingsFromSARIF_EmptyLog(t *testing.T) {
	t.Parallel()

	findings, err := FindingsFromSARIF([]byte(`{"version":"2.1.0","runs":[]}`))
	if err != nil {
		t.Fatalf("FindingsFromSARIF(): %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("FindingsFromSARIF() findings = %v, want empty", findings)
	}
}

func TestFindingsFromSARIF_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := FindingsFromSARIF([]byte(`not json`))
	if err == nil {
		t.Fatal("FindingsFromSARIF() expected error for invalid JSON, got nil")
	}
}

func TestFindingsFromSARIF_RoundTrip(t *testing.T) {
	t.Parallel()

	original := Finding{
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
		Range:       NewRangePtr("main.go", 10, 5, 10, 20),
	}

	report := NewReport(ToolInfo{Name: "govet"})
	report.AddFinding(original)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	findings, err := FindingsFromSARIF(sarif)
	if err != nil {
		t.Fatalf("FindingsFromSARIF(): %v", err)
	}
	require.Len(t, findings, 1, "round-trip findings")

	got := findings[0]

	assert.Equal(t, original.ID, got.ID)
	assert.Equal(t, original.Rule, got.Rule)
	assert.Equal(t, original.ToolName, got.ToolName)
	assert.Equal(t, original.Severity, got.Severity)
	assert.Equal(t, original.Message, got.Message)
	assert.Equal(t, original.Category, got.Category)
	assert.Equal(t, original.Tag, got.Tag)
	assert.Equal(t, original.FixStrategy, got.FixStrategy)
	assert.InDelta(t, original.Confidence, got.Confidence, 1e-9)
	assert.Equal(t, original.Suggestion, got.Suggestion)
	assert.Equal(t, original.Snippet, got.Snippet)
	assert.Equal(t, original.Position, got.Position)
	assert.NotNil(t, got.Range)
	assert.Equal(t, original.Range.End, got.Range.End)
	assert.Equal(t, "value", got.Metadata["custom"])
}

func TestSeverityToSARIFLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sev  Severity
		want string
	}{
		{"info", SeverityInfo, "note"},
		{"warning", SeverityWarning, "warning"},
		{"error", SeverityError, "error"},
		{"critical", SeverityCritical, "error"},
		{"unknown", Severity("unknown"), "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := severityToSARIFLevel(tt.sev)
			if got != tt.want {
				t.Errorf("severityToSARIFLevel(%v) = %q, want %q", tt.sev, got, tt.want)
			}
		})
	}
}

func TestFindingFromSarResult_Rank(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: SarifMessage{Text: "msg"},
		Rank:    75.0,
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
				Region:           &SarifRegion{StartLine: 10},
			},
		}},
	}

	f := findingFromSarResult(r, "tool")
	assert.InDelta(t, 0.75, f.Confidence, 1e-9)
}

func TestFindingFromSarResult_FixesWithReplacements(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: SarifMessage{Text: "msg"},
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
			},
		}},
		Fixes: []SarifFix{{
			Description: SarifMessage{Text: "fix it"},
			Changes: []SarifArtifactChange{{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
				Replacements: []SarifReplacement{{
					DeletedRegion: SarifRegion{StartLine: 5, StartColumn: 1},
					InsertedText:  SarifMessage{Text: "fixed code"},
				}},
			}},
		}},
	}

	f := findingFromSarResult(r, "tool")
	assert.Equal(t, "fix it", f.Suggestion)
	assert.Equal(t, "fixed code", f.AfterCode)
	assert.Equal(t, FixStrategySuggest, f.FixStrategy)
}

func TestFindingFromSarResult_RelatedLocations(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "r1",
		Level:   "error",
		Message: SarifMessage{Text: "main finding"},
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "main.go"},
			},
		}},
		Related: []SarifRelatedLoc{
			{
				PhysicalLocation: SarifPhysicalLocation{
					ArtifactLocation: SarifArtifactLocation{URI: "helper.go"},
					Region:           &SarifRegion{StartLine: 20, StartColumn: 3},
				},
				Message: SarifMessage{Text: "related call"},
			},
			{
				PhysicalLocation: SarifPhysicalLocation{
					ArtifactLocation: SarifArtifactLocation{URI: "util.go"},
				},
				Message: SarifMessage{Text: "no region"},
			},
		},
	}

	f := findingFromSarResult(r, "tool")
	require.Len(t, f.Related, 2, "Related")

	if f.Related[0].Relation != "related call" {
		t.Errorf("Related[0].Relation = %q, want %q", f.Related[0].Relation, "related call")
	}

	assertRelatedPosition(t, f.Related[0], "helper.go", 20)
	assertRelatedPosition(t, f.Related[1], "util.go", 0)
}

func TestFindingFromSarResult_NoLocations(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: SarifMessage{Text: "msg"},
	}

	f := findingFromSarResult(r, "tool")
	if f.Position.File != "" {
		t.Errorf("Position.File = %q, want empty (no locations)", f.Position.File)
	}
}

func TestFindingFromSarResult_Properties(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: SarifMessage{Text: "msg"},
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
			},
		}},
		Properties: goFindingPropsWithCustom(
			"test-id",
			"critical",
			"direct",
			"scanner",
			"security",
			"injection",
			0.85,
			"fix it",
			"code here",
			"custom-key",
			"custom-val",
		),
	}

	f := findingFromSarResult(r, "default-tool")
	if f.ID != "test-id" {
		t.Errorf("ID = %q, want %q", f.ID, "test-id")
	}

	if f.Severity != SeverityCritical {
		t.Errorf("Severity = %v, want %v", f.Severity, SeverityCritical)
	}

	if f.FixStrategy != FixStrategyDirect {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategyDirect)
	}

	assert.Equal(t, "scanner", f.ToolName)
	assert.Equal(t, Category("security"), f.Category)
	assert.Equal(t, "injection", f.Tag)
	assert.InDelta(t, 0.85, f.Confidence, 1e-9)
	assert.Equal(t, "fix it", f.Suggestion)
	assert.Equal(t, "code here", f.Snippet)
	assert.Equal(t, "custom-val", f.Metadata["custom-key"])
}

func TestApplySarifPosition_NilRegion(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
				Region:           nil,
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)
	assertFindingPosition(t, f, "a.go", 0)

	if f.Range != nil {
		t.Error("Range should be nil with nil region")
	}
}

func TestApplySarifPosition_WithEndPosition(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
				Region: &SarifRegion{
					StartLine:   10,
					StartColumn: 5,
					EndLine:     15,
					EndColumn:   20,
				},
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)
	if f.Range == nil {
		t.Fatal("Range should be set with end position")
	}

	if f.Range.End.Line != 15 {
		t.Errorf("Range.End.Line = %d, want 15", f.Range.End.Line)
	}

	if f.Range.End.Column != 20 {
		t.Errorf("Range.End.Column = %d, want 20", f.Range.End.Column)
	}
}

func TestApplySarifPosition_EndColumnOnly(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
				Region: &SarifRegion{
					StartLine:   10,
					StartColumn: 5,
					EndColumn:   20,
				},
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)
	if f.Range == nil {
		t.Fatal("Range should be set with EndColumn > 0")
	}
}

func TestWriteSARIF_WriterError(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
		},
	}

	err := r.WriteSARIF(&failWriter{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "writing SARIF")
}

func TestWriteSARIFFiltered_WriterError(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
		},
	}

	err := r.WriteSARIFFiltered(&failWriter{}, SeverityWarning)
	require.Error(t, err)
	assert.ErrorContains(t, err, "writing SARIF")
}

// failWriter is an io.Writer that always returns an error.
type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
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
	unmarshalJSON(t, sarif, &log)

	props := log.Runs[0].Results[0].Properties

	checks := goFindingPropsWithCustom(
		"govet:printf:main.go:10:5",
		"critical",
		"suggest",
		"govet",
		"correctness",
		"printf",
		0.9,
		"fix format string",
		`fmt.Sprintf("%d")`,
		"custom",
		"value",
	)

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

func TestFindingFromSarResult_WithFix(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "SA1000",
		Level:   "warning",
		Message: SarifMessage{Text: "unused variable"},
		Locations: []SarifLocation{
			{
				PhysicalLocation: SarifPhysicalLocation{
					ArtifactLocation: SarifArtifactLocation{URI: "main.go"},
					Region:           &SarifRegion{StartLine: 10, StartColumn: 5},
				},
			},
		},
		Fixes: []SarifFix{
			{
				Description: SarifMessage{Text: "remove unused variable"},
				Changes: []SarifArtifactChange{
					{
						ArtifactLocation: SarifArtifactLocation{URI: "main.go"},
						Replacements: []SarifReplacement{
							{
								DeletedRegion: SarifRegion{
									StartLine: 10, StartColumn: 5, EndLine: 10, EndColumn: 15,
								},
								InsertedText: SarifMessage{
									Text: "fmt.Println()",
								},
							},
						},
					},
				},
			},
		},
	}

	f := findingFromSarResult(r, "staticcheck")

	assert.Equal(t, "SA1000", f.Rule)
	assert.Equal(t, "staticcheck", f.ToolName)
	assert.Equal(t, "unused variable", f.Message)
	assert.Equal(t, SeverityWarning, f.Severity)
	assert.Equal(t, "remove unused variable", f.Suggestion)
	assert.Equal(t, "fmt.Println()", f.AfterCode)
	assert.Equal(t, FixStrategySuggest, f.FixStrategy)
}

func TestFindingFromSarResult_RankAsConfidence(t *testing.T) {
	t.Parallel()

	r := SarifResult{
		RuleID:  "R1",
		Level:   "error",
		Message: SarifMessage{Text: "msg"},
		Locations: []SarifLocation{
			{
				PhysicalLocation: SarifPhysicalLocation{
					ArtifactLocation: SarifArtifactLocation{URI: "a.go"},
					Region:           &SarifRegion{StartLine: 1},
				},
			},
		},
		Rank: 75.0,
	}

	f := findingFromSarResult(r, "tool")
	assert.InDelta(t, 0.75, f.Confidence, 0.01)
}

func TestSARIF_RoundTripPreservesBeforeCodeAndFindingID(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:          "test:R1:a.go:1:1",
		Rule:        "R1",
		ToolName:    "test",
		Message:     "msg",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		BeforeCode:  "old code",
		AfterCode:   "new code",
		FixStrategy: FixStrategySuggest,
		Related: []RelatedRef{{
			FindingID: "related-123", Relation: "causes", Position: Pos("b.go", 5, 1),
		}},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	require.NoError(t, err)

	findings, err := FindingsFromSARIF(data)
	require.NoError(t, err)
	require.Len(t, findings, 1)

	f := findings[0]

	assert.Equal(t, "test:R1:a.go:1:1", f.ID, "ID preserved via properties")
	assert.Equal(t, "R1", f.Rule, "Rule preserved")
	assert.Equal(t, "msg", f.Message, "Message preserved")
	assert.Equal(t, "new code", f.AfterCode, "AfterCode preserved via fix")

	assert.Equal(t, "old code", f.BeforeCode, "BeforeCode preserved via properties")

	require.Len(t, f.Related, 1)
	assert.Equal(
		t, "related-123", f.Related[0].FindingID,
		"RelatedRef.FindingID preserved via properties",
	)
	assert.Equal(t, "causes", f.Related[0].Relation, "RelatedRef.Relation preserved")
}

func TestSARIF_TagsRoundTrip(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:       "test:R1:a.go:1:1",
		Rule:     "R1",
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityError,
		Position: Pos("a.go", 1, 1),
		Tags:     []Tag{TagSecurity, "injection", "xss"},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	require.NoError(t, err)

	findings, err := FindingsFromSARIF(data)
	require.NoError(t, err)
	require.Len(t, findings, 1)

	assert.Equal(t, []Tag{TagSecurity, "injection", "xss"}, findings[0].Tags)
}

func TestSARIF_SuppressedFindingsExcludedFromRoundTrip(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:          "test:R1:a.go:1:1",
		Rule:        "R1",
		ToolName:    "test",
		Message:     "msg",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		Suppression: &Suppression{Kind: SuppressionInSource, Reason: "won't fix"},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	require.NoError(t, err)

	findings, err := FindingsFromSARIF(data)
	require.NoError(t, err)
	assert.Empty(t, findings, "suppressed findings are LOST in SARIF round-trip")
}
