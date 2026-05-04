package finding

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis"
)

const testSrcMain = `package main; func main() {}`

func TestFromDiagnostic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := testSrcMain

	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:      f.Name.Pos(),
		Message:  "test diagnostic",
		Category: "test",
	}

	fd := FromDiagnostic(d, fset, "testtool", "R001")
	if fd.ToolName != "testtool" {
		t.Errorf("expected toolName 'testtool', got %q", fd.ToolName)
	}

	if fd.Rule != "R001" {
		t.Errorf("expected rule 'R001', got %q", fd.Rule)
	}

	if fd.Message != "test diagnostic" {
		t.Errorf("expected message 'test diagnostic', got %q", fd.Message)
	}

	if fd.Severity != SeverityWarning {
		t.Errorf("expected severity warning, got %v", fd.Severity)
	}

	if fd.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestFromDiagnostic_WithSuggestedFixes(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := testSrcMain

	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "has fix",
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   "fix it",
				TextEdits: []analysis.TextEdit{{NewText: []byte("fixed")}},
			},
		},
	}

	fd := FromDiagnostic(d, fset, "tool", "R001")
	if fd.FixStrategy != FixStrategyDirect {
		t.Errorf("expected FixStrategyDirect, got %v", fd.FixStrategy)
	}

	if fd.Suggestion != "fix it" {
		t.Errorf("expected suggestion 'fix it', got %q", fd.Suggestion)
	}

	if fd.AfterCode != "fixed" {
		t.Errorf("expected afterCode 'fixed', got %q", fd.AfterCode)
	}
}

func TestFromDiagnostic_WithRelated(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := `package main; func main() { println() }`

	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "test",
		Related: []analysis.RelatedInformation{
			{
				Pos:     f.Decls[0].Pos(),
				Message: "related info",
			},
		},
	}

	fd := FromDiagnostic(d, fset, "tool", "R001")
	if len(fd.Related) != 1 {
		t.Fatalf("expected 1 related, got %d", len(fd.Related))
	}

	if fd.Related[0].Relation != "related" {
		t.Errorf("expected relation 'related', got %q", fd.Related[0].Relation)
	}

	if fd.Related[0].FindingID == fd.ID {
		t.Errorf("related ref should have unique ID, got same as parent %q", fd.ID)
	}

	if fd.Related[0].FindingID == "" {
		t.Error("related ref should have non-empty FindingID")
	}
}

func TestFromTokenPosition(t *testing.T) {
	t.Parallel()

	pos := token.Position{
		Filename: "test.go",
		Line:     10,
		Column:   5,
		Offset:   100,
	}

	p := FromTokenPosition(pos)
	if p != (Position{File: "test.go", Line: 10, Column: 5, Offset: 100}) {
		t.Errorf("FromTokenPosition() = %v, want match", p)
	}
}

func TestNodePosition(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()

		fset := token.NewFileSet()

		p := NodePosition(fset, nil)
		if p != (Position{}) {
			t.Errorf("expected empty position for nil node, got %v", p)
		}
	})

	t.Run("valid node", func(t *testing.T) {
		t.Parallel()

		fset := token.NewFileSet()
		src := testSrcMain

		f, err := parser.ParseFile(fset, "test.go", src, 0)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		p := NodePosition(fset, f.Name)
		if p.File == "" {
			t.Error("expected non-empty file")
		}

		if p.Line == 0 {
			t.Error("expected non-zero line")
		}
	})
}

func TestNodeRange(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()

		fset := token.NewFileSet()

		r := NodeRange(fset, nil)
		if r != (Range{}) {
			t.Errorf("expected empty range for nil node, got %v", r)
		}
	})

	t.Run("valid node", func(t *testing.T) {
		t.Parallel()

		fset := token.NewFileSet()
		src := testSrcMain

		f, err := parser.ParseFile(fset, "test.go", src, 0)
		if err != nil {
			t.Fatalf("parse: %v", err)
		}

		var decl *ast.FuncDecl

		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				decl = fd

				break
			}
		}

		if decl == nil {
			t.Fatal("no func decl found")
		}

		r := NodeRange(fset, decl)
		if r.Start.File == "" {
			t.Error("expected non-empty start file")
		}

		if r.End.Line == 0 {
			t.Error("expected non-zero end line")
		}
	})
}

func TestFormatDiagnostic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := testSrcMain

	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "test message",
	}

	out := FormatDiagnostic(d, fset, "myanalyzer")
	if out == "" {
		t.Error("expected non-empty output")
	}

	if !contains(out, "test message") {
		t.Errorf("expected output to contain 'test message', got %q", out)
	}

	if !contains(out, "myanalyzer") {
		t.Errorf("expected output to contain 'myanalyzer', got %q", out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
