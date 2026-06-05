package analysis

import (
	"go/ast"
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

func TestResolvePos_ColumnGT1(t *testing.T) {
	t.Parallel()

	src := "package main\n\nfunc main() {}\n"

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "col.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	// line 3, column 6 should be 'm' in 'main'
	pos := fset.Position(f.Name.Pos())

	p := finding.Position{
		File:   "col.go",
		Line:   pos.Line,
		Column: pos.Column,
	}

	got := resolvePos(p, fset)
	if got == token.NoPos {
		t.Error("expected valid Pos for column > 1")
	}

	gotPos := fset.Position(got)
	if gotPos.Column != pos.Column {
		t.Errorf("Column = %d, want %d", gotPos.Column, pos.Column)
	}
}

func TestResolvePos_LineBeyondFile(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	p := finding.Position{File: testFilename, Line: 9999}

	got := resolvePos(p, fset)
	if got != token.NoPos {
		t.Errorf("expected NoPos for line beyond file, got %v", got)
	}
}

func TestResolvePos_EmptyFile(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	p := finding.Position{File: "", Line: 1}

	got := resolvePos(p, fset)
	if got != token.NoPos {
		t.Errorf("expected NoPos for empty file, got %v", got)
	}

	p2 := finding.Position{File: "x.go", Line: 0}

	got2 := resolvePos(p2, fset)
	if got2 != token.NoPos {
		t.Errorf("expected NoPos for line 0, got %v", got2)
	}
}

func TestToDiagnostic_WithRange(t *testing.T) {
	t.Parallel()

	src := "package main\n\nfunc old() {\n\treturn\n}\n"

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, "range.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	f := finding.Finding{
		Message:     "replace func",
		BeforeCode:  "func old()",
		AfterCode:   "func new()",
		FixStrategy: finding.FixStrategyDirect,
		Position:    finding.Position{File: "range.go", Line: 3, Column: 1},
		Range: &finding.Range{
			Start: finding.Position{File: "range.go", Line: 3, Column: 1},
			End:   finding.Position{File: "range.go", Line: 3, Column: 11},
		},
	}

	diag := ToDiagnostic(f, fset)

	if len(diag.SuggestedFixes) != 1 {
		t.Fatalf("SuggestedFixes = %d, want 1", len(diag.SuggestedFixes))
	}

	edit := diag.SuggestedFixes[0].TextEdits[0]
	if edit.End <= edit.Pos {
		t.Errorf("End (%v) should be > Pos (%v) for range-based fix", edit.End, edit.Pos)
	}
}

func TestToDiagnostic_InsertionOnly(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	f := finding.Finding{
		Message:     "add import",
		AfterCode:   "import \"fmt\"",
		FixStrategy: finding.FixStrategySuggest,
		Position:    finding.Position{File: testFilename, Line: 1},
	}

	diag := ToDiagnostic(f, fset)

	if len(diag.SuggestedFixes) != 1 {
		t.Fatalf("SuggestedFixes = %d, want 1", len(diag.SuggestedFixes))
	}

	edit := diag.SuggestedFixes[0].TextEdits[0]
	if edit.Pos != edit.End {
		t.Errorf("insertion should have Pos == End, got %v != %v", edit.Pos, edit.End)
	}
}

func TestFromDiagnostic_IDGeneration(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "test",
	}

	got := FromDiagnostic(d, fset, "govet", "printf")

	if got.ID == "" {
		t.Error("ID should not be empty")
	}

	// ID should contain tool name and rule
	if !strings.Contains(got.ID, "govet") || !strings.Contains(got.ID, "printf") {
		t.Errorf("ID = %q, should contain tool and rule", got.ID)
	}
}

func TestNodePosition_FullAST(t *testing.T) {
	t.Parallel()

	src := `package p; func foo() { var x int = 1; _ = x }`
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	// Find the function declaration
	var fn *ast.FuncDecl

	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			fn = fd

			break
		}
	}

	if fn == nil {
		t.Fatal("no function found")
	}

	got := NodePosition(fset, fn.Name)

	if got.File != "p.go" {
		t.Errorf("File = %q, want %q", got.File, "p.go")
	}

	if got.Line == 0 {
		t.Error("Line should not be 0")
	}
}
