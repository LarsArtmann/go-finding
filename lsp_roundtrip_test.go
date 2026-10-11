package finding

import (
	"testing"
)

func TestLSPRoundTripPreservesAllFields(t *testing.T) {
	t.Parallel()

	original := Finding{
		ID:          ID("test-tool:rule1:main.go:10:5"),
		Rule:        "rule1",
		ToolName:    "test-tool",
		Message:     "unused variable",
		Severity:    SeverityWarning,
		Position:    Position{File: FilePath("main.go"), Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect,
		Confidence:  Confidence(0.85),
		Category:    CategoryCorrectness,
		Tags:        []Tag{"unused", "cleanup"},
		BeforeCode:  "x := 42",
		AfterCode:   "_ = x",
		Suggestion:  "Use _ to suppress unused variable",
	}

	diag := original.ToLSP()

	// Verify Data is populated
	if diag.Data == nil {
		t.Fatal("expected LSPDiagnostic.Data to be non-nil")
	}

	if diag.Data.ID != original.ID {
		t.Errorf("ID: got %q, want %q", diag.Data.ID, original.ID)
	}

	if diag.Data.FixStrategy != original.FixStrategy {
		t.Errorf("FixStrategy: got %q, want %q", diag.Data.FixStrategy, original.FixStrategy)
	}

	if diag.Data.Confidence != original.Confidence {
		t.Errorf("Confidence: got %v, want %v", diag.Data.Confidence, original.Confidence)
	}

	if diag.Data.Category != original.Category {
		t.Errorf("Category: got %q, want %q", diag.Data.Category, original.Category)
	}

	if diag.Data.BeforeCode != original.BeforeCode {
		t.Errorf("BeforeCode: got %q, want %q", diag.Data.BeforeCode, original.BeforeCode)
	}

	if diag.Data.AfterCode != original.AfterCode {
		t.Errorf("AfterCode: got %q, want %q", diag.Data.AfterCode, original.AfterCode)
	}

	if diag.Data.Suggestion != original.Suggestion {
		t.Errorf("Suggestion: got %q, want %q", diag.Data.Suggestion, original.Suggestion)
	}

	// Round-trip back
	restored := FromLSP(FilePath("main.go"), diag)

	if restored.ID != original.ID {
		t.Errorf("restored ID: got %q, want %q", restored.ID, original.ID)
	}

	if restored.FixStrategy != original.FixStrategy {
		t.Errorf("restored FixStrategy: got %q, want %q", restored.FixStrategy, original.FixStrategy)
	}

	if restored.Confidence != original.Confidence {
		t.Errorf("restored Confidence: got %v, want %v", restored.Confidence, original.Confidence)
	}

	if restored.Category != original.Category {
		t.Errorf("restored Category: got %q, want %q", restored.Category, original.Category)
	}

	if restored.BeforeCode != original.BeforeCode {
		t.Errorf("restored BeforeCode: got %q, want %q", restored.BeforeCode, original.BeforeCode)
	}

	if restored.AfterCode != original.AfterCode {
		t.Errorf("restored AfterCode: got %q, want %q", restored.AfterCode, original.AfterCode)
	}

	if restored.Suggestion != original.Suggestion {
		t.Errorf("restored Suggestion: got %q, want %q", restored.Suggestion, original.Suggestion)
	}

	if len(restored.Tags) != len(original.Tags) {
		t.Errorf("restored Tags length: got %d, want %d", len(restored.Tags), len(original.Tags))
	}
}

func TestLSPRoundTripWithoutData(t *testing.T) {
	t.Parallel()

	// A plain LSPDiagnostic without Data should still work (backward compat).
	diag := LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: 9, Character: 4},
			End:   LSPPosition{Line: 9, Character: 9},
		},
		Severity: LSPSeverityWarning,
		Code:     "test-rule",
		Source:   "test-tool",
		Message:  "test message",
	}

	f := FromLSP(FilePath("test.go"), diag)

	if f.Rule != "test-rule" {
		t.Errorf("Rule: got %q, want %q", f.Rule, "test-rule")
	}

	if f.ToolName != "test-tool" {
		t.Errorf("ToolName: got %q, want %q", f.ToolName, "test-tool")
	}

	if f.Position.Line != 10 {
		t.Errorf("Line: got %d, want %d", f.Position.Line, 10)
	}

	if f.FixStrategy != FixStrategyNone {
		t.Errorf("FixStrategy: got %q, want %q", f.FixStrategy, FixStrategyNone)
	}
}
