package finding

import (
	"encoding/json"
	"testing"
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
	if len(findings) != expected {
		t.Errorf("expected %d errors, got %d", expected, len(findings))
	}
}

func assertTotalCount(t *testing.T, r *Report, expected int) {
	if r.Summary.Total != expected {
		t.Errorf("expected %d total, got %d", expected, r.Summary.Total)
	}
}

// addFindingForTest is a helper to add a finding to a report with minimal boilerplate.
func addFindingForTest(r *Report, id string, sev Severity, file string) {
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

func TestGenerateID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		toolName string
		rule     string
		pos      Position
		want     string
	}{
		{
			"go-vet",
			"nilcheck",
			Position{File: "main.go", Line: 42, Column: 10},
			"go-vet:nilcheck:main.go:42:10",
		},
		{"go-vet", "nilcheck", Position{File: "main.go", Line: 42}, "go-vet:nilcheck:main.go:42"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			got := GenerateID(tt.toolName, tt.rule, tt.pos)
			if got != tt.want {
				t.Errorf("GenerateID() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id         string
		wantTool   string
		wantRule   string
		wantFile   string
		wantLine   int
		wantColumn int
		wantOK     bool
	}{
		{"go-vet:nilcheck:main.go:42:10", "go-vet", "nilcheck", "main.go", 42, 10, true},
		{"go-vet:nilcheck:main.go:42", "go-vet", "nilcheck", "main.go", 42, 0, true},
		{"invalid", "", "", "", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()

			tool, rule, file, line, col, ok := ParseID(tt.id)
			if ok != tt.wantOK {
				t.Errorf("ParseID() ok = %v, want %v", ok, tt.wantOK)

				return
			}

			if tool != tt.wantTool || rule != tt.wantRule || file != tt.wantFile ||
				line != tt.wantLine || col != tt.wantColumn {
				t.Errorf("ParseID() = (%s, %s, %s, %d, %d), want (%s, %s, %s, %d, %d)",
					tool, rule, file, line, col,
					tt.wantTool, tt.wantRule, tt.wantFile, tt.wantLine, tt.wantColumn)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError, ToolName: "tool1"},
		{ID: "2", Severity: SeverityWarning, ToolName: "tool2"},
		{ID: "3", Severity: SeverityError, ToolName: "tool2"},
	}

	// Filter by severity
	errors := Filter(findings, BySeverity(SeverityError))
	assertErrorCount(t, errors, 2)

	// Filter by tool
	tool2 := Filter(findings, ByTool("tool2"))
	if len(tool2) != 2 {
		t.Errorf("expected 2 tool2 findings, got %d", len(tool2))
	}

	// Filter by severity at least
	atLeastWarning := Filter(findings, BySeverityAtLeast(SeverityWarning))
	if len(atLeastWarning) != 3 {
		t.Errorf("expected 3 findings >= warning, got %d", len(atLeastWarning))
	}
}

func TestGroupBy(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityError},
	}

	bySeverity := GroupBySeverity(findings)
	if len(bySeverity[SeverityError]) != 2 {
		t.Errorf("expected 2 errors, got %d", len(bySeverity[SeverityError]))
	}

	if len(bySeverity[SeverityWarning]) != 1 {
		t.Errorf("expected 1 warning, got %d", len(bySeverity[SeverityWarning]))
	}
}

func TestReport(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test-tool", Version: "1.0.0"})

	addFindingForTest(r, "1", SeverityError, "a.go")
	addFindingForTest(r, "2", SeverityWarning, "b.go")
	addFindingForTest(r, "3", SeverityError, "a.go")

	r.ComputeSummary()

	assertTotalCount(t, r, 3)

	if r.Summary.FilesAffected != 2 {
		t.Errorf("expected 2 files affected, got %d", r.Summary.FilesAffected)
	}

	if r.Summary.BySeverity[SeverityError] != 2 {
		t.Errorf("expected 2 errors, got %d", r.Summary.BySeverity[SeverityError])
	}

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

	json, err := r.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Parse it back
	parsed, err := ReportFromJSON(json)
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

func TestMerge(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go"}})

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(Finding{ID: "2", Severity: SeverityWarning, Position: Position{File: "b.go"}})

	merged := Merge([]*Report{r1, r2})
	merged.ComputeSummary()

	assertTotalCount(t, merged, 2)

	if merged.Summary.FilesAffected != 2 {
		t.Errorf("expected 2 files, got %d", merged.Summary.FilesAffected)
	}
}

func TestMergeWithDeduplication(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool1"})
	r1.AddFinding(
		Finding{ID: "1", Severity: SeverityError, Position: Position{File: "a.go", Line: 10}},
	)

	r2 := NewReport(ToolInfo{Name: "tool2"})
	r2.AddFinding(
		Finding{ID: "1", Severity: SeverityWarning, Position: Position{File: "a.go", Line: 10}},
	) // Same ID

	merged := Merge([]*Report{r1, r2}, WithDeduplication(true))
	merged.ComputeSummary()

	assertTotalCount(t, merged, 1)
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

	if !IsStandardCategory(CategorySecurity) {
		t.Error("security should be standard category")
	}

	if IsStandardCategory("custom-category") {
		t.Error("custom-category should not be standard")
	}
}

func TestSuppression(t *testing.T) {
	t.Parallel()

	s := &Suppression{
		Kind:   SuppressionInSource,
		Rule:   "rule1",
		Reason: "intentional",
	}
	if s.IsExpired() {
		t.Error("suppression without expiry should not be expired")
	}
}

func TestRangeContains(t *testing.T) {
	t.Parallel()

	r := Range{
		Start: Position{File: "test.go", Line: 10, Column: 5},
		End:   Position{File: "test.go", Line: 20, Column: 10},
	}

	tests := []struct {
		name    string
		pos     Position
		want    bool
		message string
	}{
		{"in range", Position{File: "test.go", Line: 15, Column: 7}, true, ""},
		{
			"before range",
			Position{File: "test.go", Line: 5, Column: 1},
			false,
			"position before range should not be contained",
		},
		{
			"different file",
			Position{File: "other.go", Line: 15, Column: 7},
			false,
			"position in different file should not be contained",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Contains(tt.pos)
			if tt.want && !got {
				t.Error(tt.message)
			}

			if !tt.want && got {
				t.Error(tt.message)
			}
		})
	}
}
