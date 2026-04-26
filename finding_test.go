package finding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test helper functions

func standardTestFinding() Finding {
	return Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5, Offset: 0},
		Category:    "",
		Tag:         "",
		FixStrategy: FixStrategyNone,
		Suggestion:  "",
		BeforeCode:  "",
		AfterCode:   "",
		Range:       nil,
		Snippet:     "",
		Confidence:  100,
		Related:     []RelatedRef{},
		Suppression: nil,
		Metadata:    map[string]string{},
	}
}

func assertErrorCount(t *testing.T, findings []Finding, expected int) {
	t.Helper()

	if len(findings) != expected {
		t.Errorf("expected %d errors, got %d", expected, len(findings))
	}
}

func assertTotalCount(t *testing.T, r *Report, expected int) {
	t.Helper()

	if r.Summary.Total != expected {
		t.Errorf("expected %d total, got %d", expected, r.Summary.Total)
	}
}

// addFindingForTest is a helper to add a finding to a report with minimal boilerplate.
func addFindingForTest(t *testing.T, r *Report, id string, sev Severity, file string) {
	t.Helper()

	r.AddFinding(Finding{
		ID:          id,
		Severity:    sev,
		Position:    Position{File: file, Line: 0, Column: 0, Offset: 0},
		Category:    "",
		Tag:         "",
		FixStrategy: FixStrategyNone,
		Suggestion:  "",
		BeforeCode:  "",
		AfterCode:   "",
		Range:       nil,
		Snippet:     "",
		Confidence:  100,
		Related:     []RelatedRef{},
		Suppression: nil,
		Metadata:    map[string]string{},
	})
}

func TestSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		valid    bool
	}{
		{"info", SeverityInfo, true},
		{"warning", SeverityWarning, true},
		{"error", SeverityError, true},
		{"critical", SeverityCritical, true},
		{"invalid", Severity("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.severity.IsValid(); got != tt.valid {
				t.Errorf("IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestSeverityOrdering(t *testing.T) {
	t.Parallel()

	if SeverityInfo.GreaterThan(SeverityWarning) {
		t.Error("info should not be greater than warning")
	}

	if !SeverityError.GreaterThan(SeverityWarning) {
		t.Error("error should be greater than warning")
	}

	if !SeverityCritical.GreaterThan(SeverityError) {
		t.Error("critical should be greater than error")
	}
}

func TestFixStrategy(t *testing.T) {
	t.Parallel()

	if !FixStrategyNone.IsValid() {
		t.Error("none should be valid")
	}

	if !FixStrategyDirect.CanAutoApply() {
		t.Error("direct should be auto-applicable")
	}

	if FixStrategySuggest.CanAutoApply() {
		t.Error("suggest should not be auto-applicable")
	}

	if !FixStrategyAI.NeedsAI() {
		t.Error("ai should need AI")
	}
}

func TestPosition(t *testing.T) {
	t.Parallel()

	p := Position{File: "test.go", Line: 42, Column: 5}
	if !p.IsValid() {
		t.Error("position should be valid")
	}

	if p.String() != "test.go:42:5" {
		t.Errorf("String() = %s, want test.go:42:5", p.String())
	}
}

func TestPositionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pos  Position
		want string
	}{
		{Position{File: "test.go"}, "test.go"},
		{Position{File: "test.go", Line: 42}, "test.go:42"},
		{Position{File: "test.go", Line: 42, Column: 5}, "test.go:42:5"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.pos.String(); got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFinding(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect,
		BeforeCode:  "old",
		AfterCode:   "new",
	}

	if !f.IsValid() {
		t.Error("finding should be valid")
	}

	if !f.HasFix() {
		t.Error("finding should have fix")
	}

	if !f.HasSuggestion() {
		t.Error("finding should have suggestion")
	}

	if f.IsSuppressed() {
		t.Error("finding should not be suppressed")
	}
}

func TestFindingSuppressed(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityWarning,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		Suppression: &Suppression{Kind: SuppressionInSource, Reason: "intentional"},
	}

	if !f.IsSuppressed() {
		t.Error("finding should be suppressed")
	}
}

func TestReport(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test-tool", Version: "1.0.0"})

	addFindingForTest(t, r, "1", SeverityError, "a.go")
	addFindingForTest(t, r, "2", SeverityWarning, "b.go")
	addFindingForTest(t, r, "3", SeverityError, "a.go")

	r.ComputeSummary()

	assertTotalCount(t, r, 3)

	assertSummaryField(t, "FilesAffected", r.Summary.FilesAffected, 2)
	assertSummarySeverity(t, r, SeverityError, 2)

	errors := r.BySeverity(SeverityError)
	assertErrorCount(t, errors, 2)
}

func TestReportJSON(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{
		ID:       "test:rule1:file.go:10:5",
		Rule:     "rule1",
		ToolName: "test",
		Message:  "test message",
		Severity: SeverityError,
		Position: Position{File: "file.go", Line: 10, Column: 5},
	})
	r.ComputeSummary()

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Parse it back
	parsed, _, err := ReportFromJSON(data)
	if err != nil {
		t.Fatalf("ReportFromJSON failed: %v", err)
	}

	if parsed.Tool.Name != "test" {
		t.Errorf("expected tool name 'test', got '%s'", parsed.Tool.Name)
	}

	if len(parsed.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(parsed.Findings))
	}
}

func TestFindingJSON(t *testing.T) {
	t.Parallel()

	f := standardTestFinding()

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.ID != f.ID {
		t.Errorf("expected ID %s, got %s", f.ID, parsed.ID)
	}
}

func TestSARIFConversion(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect,
		AfterCode:   "fixed",
		Category:    CategoryStyle,
		Metadata:    map[string]string{"key": "value"},
	})
	r.ComputeSummary()

	sarif, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF failed: %v", err)
	}

	// Basic validation - should be valid JSON
	var log map[string]any

	err = json.Unmarshal(sarif, &log)
	if err != nil {
		t.Fatalf("SARIF is not valid JSON: %v", err)
	}

	if log["version"] != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %v", log["version"])
	}
}

func TestLSPConversion(t *testing.T) {
	t.Parallel()

	f := standardTestFinding()

	lsp := f.ToLSP()
	if lsp.Range.Start.Line != 9 { // 0-based
		t.Errorf("expected line 9 (0-based), got %d", lsp.Range.Start.Line)
	}

	if lsp.Severity != 1 { // Error
		t.Errorf("expected severity 1 (Error), got %d", lsp.Severity)
	}
}

func TestCategory(t *testing.T) {
	t.Parallel()

	if !CategorySecurity.IsValid() {
		t.Error("security should be a valid category")
	}

	if !Category("custom-category").IsValid() {
		t.Error("custom-category should be valid")
	}

	if Category("custom-category").IsStandard() {
		t.Error("custom-category should not be standard")
	}
}

func TestRangeContains(t *testing.T) {
	t.Parallel()

	r := NewRange("test.go", 10, 5, 20, 10)

	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"in range", Position{File: "test.go", Line: 15, Column: 7}, true},
		{"before range", Position{File: "test.go", Line: 5, Column: 1}, false},
		{"different file", Position{File: "other.go", Line: 15, Column: 7}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, r.Contains(tt.pos))
		})
	}
}
