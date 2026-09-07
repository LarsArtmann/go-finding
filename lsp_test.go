package finding

import (
	"testing"
	"time"
)

func TestSeverityToLSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     LSPSeverity
	}{
		{"critical maps to error", SeverityCritical, LSPSeverityError},
		{"error maps to error", SeverityError, LSPSeverityError},
		{"warning maps to warning", SeverityWarning, LSPSeverityWarning},
		{"info maps to info", SeverityInfo, LSPSeverityInfo},
		{"unknown maps to warning", Severity("unknown"), LSPSeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := severityToLSP(tt.severity); got != tt.want {
				t.Errorf("severityToLSP(%v) = %d, want %d", tt.severity, int(got), int(tt.want))
			}
		})
	}
}

func TestSeverityFromLSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sev  LSPSeverity
		want Severity
	}{
		{"error(1) maps to SeverityError", LSPSeverityError, SeverityError},
		{"warning(2) maps to SeverityWarning", LSPSeverityWarning, SeverityWarning},
		{"info(3) maps to SeverityInfo", LSPSeverityInfo, SeverityInfo},
		{"hint(4) maps to SeverityInfo", LSPSeverityHint, SeverityInfo},
		{"unknown(0) maps to SeverityWarning", 0, SeverityWarning},
		{"unknown(99) maps to SeverityWarning", 99, SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := severityFromLSP(tt.sev); got != tt.want {
				t.Errorf("severityFromLSP(%d) = %v, want %v", int(tt.sev), got, tt.want)
			}
		})
	}
}

func lspDiag(line, startChar, endChar int) LSPDiagnostic {
	return LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: line, Character: startChar},
			End:   LSPPosition{Line: line, Character: endChar},
		},
	}
}

func TestFromLSP(t *testing.T) {
	t.Parallel()

	diag := LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: 4, Character: 9},
			End:   LSPPosition{Line: 4, Character: 15},
		},
		Severity: LSPSeverityError,
		Code:     "unused-var",
		Source:   "golangci-lint",
		Message:  "unused variable: x",
	}

	f := FromLSP("file:///test.go", diag)

	if got, want := f.ID, "golangci-lint:unused-var:file:///test.go:5:10"; string(got) != want {
		t.Errorf("FromLSP ID = %q, want %q", got, want)
	}

	if got, want := f.Rule, "unused-var"; string(got) != want {
		t.Errorf("FromLSP Rule = %q, want %q", got, want)
	}

	if got, want := f.ToolName, "golangci-lint"; string(got) != want {
		t.Errorf("FromLSP ToolName = %q, want %q", got, want)
	}

	if got, want := f.Message, "unused variable: x"; got != want {
		t.Errorf("FromLSP Message = %q, want %q", got, want)
	}

	if got, want := f.Severity, SeverityError; got != want {
		t.Errorf("FromLSP Severity = %v, want %v", got, want)
	}

	if got, want := f.Position.Line, 5; got != want {
		t.Errorf("FromLSP Position.Line = %d, want %d (0-based→1-based)", got, want)
	}

	if got, want := f.Position.Column, 10; got != want {
		t.Errorf("FromLSP Position.Column = %d, want %d (0-based→1-based)", got, want)
	}

	if got, want := string(f.Position.File), "file:///test.go"; got != want {
		t.Errorf("FromLSP Position.File = %q, want %q", got, want)
	}

	if got, want := f.FixStrategy, FixStrategyNone; got != want {
		t.Errorf("FromLSP FixStrategy = %v, want %v", got, want)
	}

	if f.Range == nil {
		t.Fatal("FromLSP Range = nil, want non-nil (end differs from start)")
	}

	if got, want := f.Range.End.Line, 5; got != want {
		t.Errorf("FromLSP Range.End.Line = %d, want %d", got, want)
	}

	if got, want := f.Range.End.Column, 16; got != want {
		t.Errorf("FromLSP Range.End.Column = %d, want %d", got, want)
	}

	if got, want := f.Metadata[LSPSeverityKey], "1"; got != want {
		t.Errorf("FromLSP Metadata[lsp-severity] = %q, want %q", got, want)
	}

	p := ParseID(f.ID)
	if !p.OK() {
		t.Fatalf("ParseID(%q) failed", f.ID)
	}

	if p.Tool != "golangci-lint" {
		t.Errorf("ParseID Tool = %q, want %q", p.Tool, "golangci-lint")
	}

	if p.Rule != "unused-var" {
		t.Errorf("ParseID Rule = %q, want %q", p.Rule, "unused-var")
	}

	if p.File != "file:///test.go" {
		t.Errorf("ParseID File = %q, want %q", p.File, "file:///test.go")
	}

	if p.Line != 5 {
		t.Errorf("ParseID Line = %d, want %d", p.Line, 5)
	}

	if p.Column != 10 {
		t.Errorf("ParseID Column = %d, want %d", p.Column, 10)
	}
}

func TestFromLSPNoEndRange(t *testing.T) {
	t.Parallel()

	diag := lspDiag(4, 9, 9)
	diag.Severity = LSPSeverityWarning
	diag.Code = "simplify"
	diag.Source = "gofmt"
	diag.Message = "can simplify"

	f := FromLSP("file:///test.go", diag)

	if f.Range != nil {
		t.Errorf("FromLSP Range = %v, want nil (end same as start)", f.Range)
	}
}

func TestFromLSPRelated(t *testing.T) {
	t.Parallel()

	diag := lspDiag(0, 0, 5)
	diag.Severity = LSPSeverityWarning
	diag.Code = "dupl"
	diag.Source = "dupl"
	diag.Message = "clone detected"
	diag.Related = []LSPRelated{
		{
			Location: LSPLocation{
				URI:   "file:///other.go",
				Range: LSPRange{Start: LSPPosition{Line: 9, Character: 0}},
			},
			Message: "clone-of",
		},
	}

	f := FromLSP("file:///test.go", diag)

	if got, want := len(f.Related), 1; got != want {
		t.Fatalf("FromLSP Related length = %d, want %d", got, want)
	}

	if got, want := f.Related[0].Relation, RelationKind("clone-of"); got != want {
		t.Errorf("FromLSP Related[0].Relation = %q, want %q", got, want)
	}

	if got, want := string(f.Related[0].Position.File), "file:///other.go"; got != want {
		t.Errorf("FromLSP Related[0].File = %q, want %q", got, want)
	}

	if got, want := f.Related[0].Position.Line, 10; got != want {
		t.Errorf("FromLSP Related[0].Line = %d, want %d", got, want)
	}
}

// TestLSPRoundTrip_SeverityCritical verifies that SeverityCritical survives an
// LSP round-trip. LSP collapses critical → error, so the exact severity must be
// carried in LSPDiagnosticData. Regression for lsp.go:51.
func TestLSPRoundTrip_SeverityCritical(t *testing.T) {
	t.Parallel()

	original := Finding{
		ID:         ID("crit-1"),
		Rule:       "critical-rule",
		ToolName:   "tool",
		Message:    "critical issue",
		Severity:   SeverityCritical,
		Position:   Position{File: FilePath("main.go"), Line: 5},
		Confidence: ConfidenceHigh,
	}

	diag := original.ToLSP()

	// LSP diagnostic severity must be Error (critical is not representable).
	if diag.Severity != LSPSeverityError {
		t.Errorf("LSP severity: got %d, want %d (LSPSeverityError)", int(diag.Severity), int(LSPSeverityError))
	}

	restored := FromLSP("file:///main.go", diag)

	if restored.Severity != SeverityCritical {
		t.Errorf("round-trip severity: got %q, want %q (SeverityCritical)",
			restored.Severity, SeverityCritical)
	}
}

// TestLSPRoundTrip_SnippetMetadataSuppression verifies that Snippet, Metadata,
// and Suppression survive the LSP round-trip via LSPDiagnosticData.
func TestLSPRoundTrip_SnippetMetadataSuppression(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	original := Finding{
		ID:       ID("full-1"),
		Rule:     "rule-x",
		ToolName: "tool",
		Message:  "test",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("main.go"), Line: 3},
		Snippet:  "var x = unused",
		Metadata: map[string]string{"custom-key": "custom-val"},
		Suppression: &Suppression{
			Kind:   SuppressionInSource,
			Rule:   "rule-x",
			Reason: "intentional",
		},
	}

	restored := FromLSP("file:///main.go", original.ToLSP())

	if restored.Snippet != "var x = unused" {
		t.Errorf("Snippet: got %q, want %q", restored.Snippet, "var x = unused")
	}

	if restored.Metadata["custom-key"] != "custom-val" {
		t.Errorf("Metadata[custom-key]: got %q, want %q", restored.Metadata["custom-key"], "custom-val")
	}

	if restored.Suppression == nil {
		t.Fatal("expected Suppression to survive round-trip")
	}

	if restored.Suppression.Kind != SuppressionInSource {
		t.Errorf("Suppression.Kind: got %q, want %q", restored.Suppression.Kind, SuppressionInSource)
	}

	if restored.Suppression.Rule != "rule-x" {
		t.Errorf("Suppression.Rule: got %q, want %q", restored.Suppression.Rule, "rule-x")
	}

	_ = now // expiry not tested here; IsExpired boundary tested separately
}

// TestLSPRoundTrip_RelatedFindingID verifies that RelatedRef.FindingID survives
// the LSP round-trip instead of being regenerated.
func TestLSPRoundTrip_RelatedFindingID(t *testing.T) {
	t.Parallel()

	original := Finding{
		ID:       ID("main-1"),
		Rule:     "rule-a",
		ToolName: "tool",
		Message:  "main",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("main.go"), Line: 1},
		Related: []RelatedRef{{
			FindingID: ID("original-related-id"),
			Relation:  "caused-by",
			Position:  Position{File: FilePath("other.go"), Line: 5},
		}},
	}

	restored := FromLSP("file:///main.go", original.ToLSP())

	if len(restored.Related) != 1 {
		t.Fatalf("Related length: got %d, want 1", len(restored.Related))
	}

	if restored.Related[0].FindingID != ID("original-related-id") {
		t.Errorf("Related[0].FindingID: got %q, want %q (should NOT be regenerated)",
			restored.Related[0].FindingID, ID("original-related-id"))
	}
}

// TestLSPRoundTrip_GroupID verifies that GroupID survives the LSP round-trip.
func TestLSPRoundTrip_GroupID(t *testing.T) {
	t.Parallel()

	original := Finding{
		ID:       ID("clone-1"),
		Rule:     "clone-detected",
		ToolName: "art-dupl",
		Message:  "duplicate code",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("a.go"), Line: 3},
		GroupID:  "clone-group-42",
	}

	restored := FromLSP("file:///a.go", original.ToLSP())

	if restored.GroupID != "clone-group-42" {
		t.Errorf("GroupID: got %q, want %q", restored.GroupID, "clone-group-42")
	}
}

// TestLSPRoundTrip_DiagnosticTags verifies that LSP diagnostic tags stored in
// metadata by FromLSP are re-emitted on the diagnostic by ToLSP, and that a
// full round-trip preserves them.
func TestLSPRoundTrip_DiagnosticTags(t *testing.T) {
	t.Parallel()

	original := Finding{
		ID:       ID("unused-1"),
		Rule:     "unused-code",
		ToolName: "art-dupl",
		Message:  "unnecessary code",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("dup.go"), Line: 7},
		Metadata: map[string]string{LSPDiagnosticTagsKey: "1"},
	}

	diag := original.ToLSP()

	if len(diag.Tags) != 1 || diag.Tags[0] != LSPDiagnosticTagUnnecessary {
		t.Fatalf("Tags: got %v, want [%d]", diag.Tags, LSPDiagnosticTagUnnecessary)
	}

	restored := FromLSP("file:///dup.go", diag)

	if got := restored.Metadata[LSPDiagnosticTagsKey]; got != "1" {
		t.Errorf("Metadata[%s]: got %q, want %q", LSPDiagnosticTagsKey, got, "1")
	}

	rediag := restored.ToLSP()

	if len(rediag.Tags) != 1 || rediag.Tags[0] != LSPDiagnosticTagUnnecessary {
		t.Errorf("re-emitted Tags: got %v, want [%d]", rediag.Tags, LSPDiagnosticTagUnnecessary)
	}
}

// TestToLSP_DiagnosticTags_IgnoreMalformed verifies that malformed metadata
// values do not produce tag entries.
func TestToLSP_DiagnosticTags_IgnoreMalformed(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       ID("x"),
		Rule:     "r",
		ToolName: "t",
		Message:  "m",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("a.go"), Line: 1},
		Metadata: map[string]string{LSPDiagnosticTagsKey: "1,bogus,2"},
	}

	diag := f.ToLSP()

	if len(diag.Tags) != 2 {
		t.Fatalf("Tags: got %v, want 2 entries", diag.Tags)
	}

	if diag.Tags[0] != LSPDiagnosticTagUnnecessary || diag.Tags[1] != LSPDiagnosticTagDeprecated {
		t.Errorf("Tags: got %v, want [%d %d]", diag.Tags, LSPDiagnosticTagUnnecessary, LSPDiagnosticTagDeprecated)
	}
}
