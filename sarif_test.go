package finding

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
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
	g := NewWithT(t)
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
	g.Expect(result.RuleID).To(Equal("SA1000"))
	g.Expect(string(SeverityError)).To(Equal(result.Level))
	g.Expect(result.Message.Text).To(Equal("bad code"))
	g.Expect(result.Locations[0].PhysicalLocation.ArtifactLocation.URI).To(Equal("main.go"))
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
	g := NewWithT(t)
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
	g.Expect(err).NotTo(HaveOccurred())

	log := unmarshalSARIF(t, data)
	require.Len(t, log.Runs[0].Results, 1)

	result := log.Runs[0].Results[0]
	require.Len(t, result.Fixes, 1)
	g.Expect(result.Fixes[0].Description.Text).To(Equal("do better"))
	g.Expect(result.Fixes[0].Changes).To(BeEmpty())
}

func TestToSARIF_WithRelated(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).NotTo(HaveOccurred())

	log := unmarshalSARIF(t, data)
	require.Len(t, log.Runs[0].Results, 1)

	result := log.Runs[0].Results[0]
	require.Len(t, result.Related, 1)
	g.Expect(result.Related[0].PhysicalLocation.ArtifactLocation.URI).To(Equal("b.go"))
	g.Expect(result.Related[0].PhysicalLocation.Region.StartLine).To(Equal(20))
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
	g := NewWithT(t)
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
	g.Expect(err).NotTo(HaveOccurred())

	data := buf.String()
	g.Expect(data).To(ContainSubstring(`"version": "2.1.0"`))
	g.Expect(data).To(ContainSubstring(`"tool"`))
}

func TestWriteSARIFFiltered(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(data).To(ContainSubstring("f1"))
	g.Expect(data).NotTo(ContainSubstring("f2"))
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
	g := NewWithT(t)
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

	g.Expect(got.ID).To(Equal(original.ID))
	g.Expect(got.Rule).To(Equal(original.Rule))
	g.Expect(got.ToolName).To(Equal(original.ToolName))
	g.Expect(got.Severity).To(Equal(original.Severity))
	g.Expect(got.Message).To(Equal(original.Message))
	g.Expect(got.Category).To(Equal(original.Category))
	g.Expect(got.Tag).To(Equal(original.Tag))
	g.Expect(got.FixStrategy).To(Equal(original.FixStrategy))
	g.Expect(got.Confidence).To(BeNumerically("~", original.Confidence, 1e-9))
	g.Expect(got.Suggestion).To(Equal(original.Suggestion))
	g.Expect(got.Snippet).To(Equal(original.Snippet))
	g.Expect(got.Position).To(Equal(original.Position))
	g.Expect(got.Range).NotTo(BeNil())
	g.Expect(got.Range.End).To(Equal(original.Range.End))
	g.Expect(got.Metadata["custom"]).To(Equal("value"))
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
	g := NewWithT(t)
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
	g.Expect(f.Confidence).To(BeNumerically("~", 0.75, 1e-9))
}

func TestFindingFromSarResult_FixesWithReplacements(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(f.Suggestion).To(Equal("fix it"))
	g.Expect(f.AfterCode).To(Equal("fixed code"))
	g.Expect(f.FixStrategy).To(Equal(FixStrategySuggest))
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
	g := NewWithT(t)
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

	g.Expect(f.ToolName).To(Equal("scanner"))
	g.Expect(f.Category).To(Equal(Category("security")))
	g.Expect(f.Tag).To(Equal("injection"))
	g.Expect(f.Confidence).To(BeNumerically("~", 0.85, 1e-9))
	g.Expect(f.Suggestion).To(Equal("fix it"))
	g.Expect(f.Snippet).To(Equal("code here"))
	g.Expect(f.Metadata["custom-key"]).To(Equal("custom-val"))
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
	g := NewWithT(t)
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
	g.Expect(err).To(MatchError(ContainSubstring("writing SARIF")))
}

func TestWriteSARIFFiltered_WriterError(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).To(MatchError(ContainSubstring("writing SARIF")))
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
	g := NewWithT(t)
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

	g.Expect(f.Rule).To(Equal("SA1000"))
	g.Expect(f.ToolName).To(Equal("staticcheck"))
	g.Expect(f.Message).To(Equal("unused variable"))
	g.Expect(f.Severity).To(Equal(SeverityWarning))
	g.Expect(f.Suggestion).To(Equal("remove unused variable"))
	g.Expect(f.AfterCode).To(Equal("fmt.Println()"))
	g.Expect(f.FixStrategy).To(Equal(FixStrategySuggest))
}

func TestFindingFromSarResult_RankAsConfidence(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(f.Confidence).To(BeNumerically("~", 0.75, 0.01))
}

func TestSARIF_RoundTripPreservesBeforeCodeAndFindingID(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	g.Expect(f.ID).To(Equal("test:R1:a.go:1:1"))
	g.Expect(f.Rule).To(Equal("R1"))
	g.Expect(f.Message).To(Equal("msg"))
	g.Expect(f.AfterCode).To(Equal("new code"))

	g.Expect(f.BeforeCode).To(Equal("old code"))

	require.Len(t, f.Related, 1)
	g.Expect(f.Related[0].FindingID).To(Equal("related-123"))
	g.Expect(f.Related[0].Relation).To(Equal("causes"))
}

func TestSARIF_TagsRoundTrip(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	g.Expect(findings[0].Tags).To(Equal([]Tag{TagSecurity, "injection", "xss"}))
}

func TestSARIF_SuppressedFindingsExcludedFromRoundTrip(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	g.Expect(findings).To(BeEmpty())
}
