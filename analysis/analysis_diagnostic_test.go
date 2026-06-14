package analysis

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
)

func TestToDiagnostic_WithSuggestedFix(t *testing.T) {
	t.Parallel()

	src := "package main\n\nfunc main() {\n\told()\n}\n"

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, "fixme.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	f := finding.Finding{
		Rule:        "R1",
		ToolName:    "tool",
		Message:     "replace old with new",
		BeforeCode:  "old()",
		AfterCode:   "new()",
		FixStrategy: finding.FixStrategyDirect,
		Position:    finding.Position{File: "fixme.go", Line: 4, Column: 2},
		Suggestion:  "use new()",
	}

	diag := ToDiagnostic(f, fset)

	if len(diag.SuggestedFixes) != 1 {
		t.Fatalf("SuggestedFixes = %d, want 1", len(diag.SuggestedFixes))
	}

	if diag.SuggestedFixes[0].Message != "use new()" {
		t.Errorf("SuggestedFix.Message = %q, want %q", diag.SuggestedFixes[0].Message, "use new()")
	}

	if len(diag.SuggestedFixes[0].TextEdits) != 1 {
		t.Fatalf("TextEdits = %d, want 1", len(diag.SuggestedFixes[0].TextEdits))
	}

	edit := diag.SuggestedFixes[0].TextEdits[0]
	if string(edit.NewText) != "new()" {
		t.Errorf("NewText = %q, want %q", string(edit.NewText), "new()")
	}
}

func TestToDiagnostic_FileNotInFset(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f := finding.Finding{
		Message: "orphan finding",
		Position: finding.Position{
			File: "nonexistent.go",
			Line: 10,
		},
	}

	diag := ToDiagnostic(f, fset)

	if diag.Message != "orphan finding" {
		t.Errorf("Message = %q, want %q", diag.Message, "orphan finding")
	}

	if diag.Pos != token.NoPos {
		t.Errorf("Pos should be NoPos for file not in fset, got %v", diag.Pos)
	}
}

func TestToDiagnostic_RoundTrip(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	orig := &analysis.Diagnostic{
		Pos:            f.Name.Pos(),
		Message:        "round trip test",
		Category:       "test",
		SuggestedFixes: suggestedFix("fix it"),
	}

	// Forward: Diagnostic -> Finding
	got := FromDiagnostic(orig, fset, "tool", "R1")

	// Reverse: Finding -> Diagnostic
	result := ToDiagnostic(got, fset)

	if result.Message != orig.Message {
		t.Errorf("Message: got %q, want %q", result.Message, orig.Message)
	}

	if result.Category != orig.Category {
		t.Errorf("Category: got %q, want %q", result.Category, orig.Category)
	}

	// Position should round-trip exactly
	origPos := fset.Position(orig.Pos)

	resultPos := fset.Position(result.Pos)
	if origPos.Line != resultPos.Line || origPos.Column != resultPos.Column {
		t.Errorf("Position: got %d:%d, want %d:%d",
			resultPos.Line, resultPos.Column, origPos.Line, origPos.Column)
	}
}

func TestToDiagnostic_WithRelated(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	pos := fset.Position(f.Name.Pos())

	finding := finding.Finding{
		Message:  "main issue",
		Position: FromTokenPosition(pos),
		Related: []finding.RelatedRef{
			{
				Relation: "causes",
				Position: FromTokenPosition(pos),
			},
		},
	}

	diag := ToDiagnostic(finding, fset)

	if len(diag.Related) != 1 {
		t.Fatalf("Related = %d, want 1", len(diag.Related))
	}

	if diag.Related[0].Message != "causes" {
		t.Errorf("Related[0].Message = %q, want %q", diag.Related[0].Message, "causes")
	}
}
