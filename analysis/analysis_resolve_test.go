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
