package analysis

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
)

const testSrcMain = `package main; func main() {}`

const (
	testFilename    = "test.go"
	testAnalysisCat = "test"
	testRangeGoFile = "range.go"
)

func suggestedFix(msg string) []analysis.SuggestedFix {
	return []analysis.SuggestedFix{
		{
			Message:   msg,
			TextEdits: []analysis.TextEdit{{NewText: []byte("fixed")}},
		},
	}
}

func TestFromDiagnostic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := testSrcMain

	f, err := parser.ParseFile(fset, testFilename, src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:      f.Name.Pos(),
		Message:  "test diagnostic",
		Category: "test",
	}

	got := FromDiagnostic(d, fset, "test-tool", "R001")

	if got.Rule != "R001" {
		t.Errorf("Rule = %q, want %q", got.Rule, "R001")
	}

	if got.ToolName != "test-tool" {
		t.Errorf("ToolName = %q, want %q", got.ToolName, "test-tool")
	}

	if got.Message != "test diagnostic" {
		t.Errorf("Message = %q, want %q", got.Message, "test diagnostic")
	}

	if got.Severity != finding.SeverityWarning {
		t.Errorf("Severity = %v, want %v", got.Severity, finding.SeverityWarning)
	}

	if got.Category != "test" {
		t.Errorf("Category = %q, want %q", got.Category, "test")
	}

	if got.Position.File != testFilename {
		t.Errorf("Position.File = %q, want %q", got.Position.File, testFilename)
	}

	if got.FixStrategy != finding.FixStrategyNone {
		t.Errorf("FixStrategy = %v, want %v", got.FixStrategy, finding.FixStrategyNone)
	}
}

func TestFromDiagnostic_WithCustomSeverity(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "critical issue",
	}

	got := FromDiagnostic(d, fset, "tool", "R1", finding.SeverityCritical)
	if got.Severity != finding.SeverityCritical {
		t.Errorf("Severity = %v, want %v", got.Severity, finding.SeverityCritical)
	}
}

func TestFromDiagnostic_WithSuggestedFix(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:            f.Name.Pos(),
		Message:        "fix me",
		SuggestedFixes: suggestedFix("apply this fix"),
	}

	got := FromDiagnostic(d, fset, "tool", "R1")

	if got.FixStrategy != finding.FixStrategyDirect {
		t.Errorf("FixStrategy = %v, want %v", got.FixStrategy, finding.FixStrategyDirect)
	}

	if got.Suggestion != "apply this fix" {
		t.Errorf("Suggestion = %q, want %q", got.Suggestion, "apply this fix")
	}

	if got.AfterCode != "fixed" {
		t.Errorf("AfterCode = %q, want %q", got.AfterCode, "fixed")
	}
}

func TestFromDiagnostic_WithRelated(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "main issue",
		Related: []analysis.RelatedInformation{
			{Pos: f.End() - 1, Message: "related info"},
		},
	}

	got := FromDiagnostic(d, fset, "tool", "R1")

	if len(got.Related) != 1 {
		t.Fatalf("Related length = %d, want 1", len(got.Related))
	}

	if got.Related[0].Relation != finding.RelationKind("related") {
		t.Errorf("Related[0].Relation = %q, want %q", got.Related[0].Relation, "related")
	}
}

func TestFromTokenPosition(t *testing.T) {
	t.Parallel()

	pos := token.Position{
		Filename: "main.go",
		Line:     10,
		Column:   5,
		Offset:   42,
	}

	got := FromTokenPosition(pos)

	if got.File != "main.go" {
		t.Errorf("File = %q, want %q", got.File, "main.go")
	}

	if got.Line != 10 {
		t.Errorf("Line = %d, want 10", got.Line)
	}

	if got.Column != 5 {
		t.Errorf("Column = %d, want 5", got.Column)
	}

	if got.Offset != 42 {
		t.Errorf("Offset = %d, want 42", got.Offset)
	}
}

func TestNodePosition(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	got := NodePosition(fset, f.Name)

	if got.File != testFilename {
		t.Errorf("File = %q, want %q", got.File, testFilename)
	}

	if got.Line == 0 {
		t.Error("Line should not be 0 for a valid node")
	}
}

func TestNodePosition_NilNode(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	got := NodePosition(fset, nil)

	if got.File != "" {
		t.Errorf("File = %q, want empty", got.File)
	}
}

func TestNodeRange(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	got := NodeRange(fset, f.Name)

	if got.Start.File != testFilename {
		t.Errorf("Start.File = %q, want %q", got.Start.File, testFilename)
	}
}

func TestNodeRange_NilNode(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	got := NodeRange(fset, nil)

	if got.Start.File != "" || got.End.File != "" {
		t.Errorf("expected zero range for nil node, got %v", got)
	}
}

func TestFormatDiagnostic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "main.go", testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "unused variable",
	}

	got := FormatDiagnostic(d, fset, "govet")

	if got == "" {
		t.Error("FormatDiagnostic returned empty string")
	}

	// Should contain the analyzer name and message
	if !strings.Contains(got, "govet") {
		t.Errorf("expected %q to contain %q", got, "govet")
	}

	if !strings.Contains(got, "unused variable") {
		t.Errorf("expected %q to contain %q", got, "unused variable")
	}
}

func TestToDiagnostic_Basic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	// First get a Finding from a known position, then convert back.
	origPos := fset.Position(f.Name.Pos())

	finding := finding.Finding{
		Rule:     "R1",
		ToolName: "test",
		Message:  "test message",
		Severity: finding.SeverityError,
		Position: finding.Position{
			File:   origPos.Filename,
			Line:   origPos.Line,
			Column: origPos.Column,
		},
		Category: finding.CategoryCorrectness,
	}

	diag := ToDiagnostic(finding, fset)

	if diag.Message != "test message" {
		t.Errorf("Message = %q, want %q", diag.Message, "test message")
	}

	if diag.Category != "correctness" {
		t.Errorf("Category = %q, want %q", diag.Category, "correctness")
	}

	if diag.Pos == token.NoPos {
		t.Error("Pos should not be NoPos for a valid file/line")
	}

	// The position should round-trip back to the same line
	gotPos := fset.Position(diag.Pos)
	if gotPos.Line != origPos.Line {
		t.Errorf("Line = %d, want %d", gotPos.Line, origPos.Line)
	}
}
