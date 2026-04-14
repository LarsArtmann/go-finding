package finding

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestFromDiagnostic(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := `package main; func main() {}`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "test diagnostic",
		Category: "test",
	}

	finding := FromDiagnostic(d, fset, "testtool", "R001")
	if finding.ToolName != "testtool" {
		t.Errorf("expected toolName 'testtool', got %q", finding.ToolName)
	}
	if finding.Rule != "R001" {
		t.Errorf("expected rule 'R001', got %q", finding.Rule)
	}
	if finding.Message != "test diagnostic" {
		t.Errorf("expected message 'test diagnostic', got %q", finding.Message)
	}
	if finding.Severity != SeverityWarning {
		t.Errorf("expected severity warning, got %v", finding.Severity)
	}
	if finding.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestFromDiagnostic_WithSuggestedFixes(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	src := `package main; func main() {}`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := &analysis.Diagnostic{
		Pos:     f.Name.Pos(),
		Message: "has fix",
		SuggestedFixes: []analysis.SuggestedFix{
			{Message: "fix it"},
		},
	}

	finding := FromDiagnostic(d, fset, "tool", "R001")
	if finding.FixStrategy != FixStrategyDirect {
		t.Errorf("expected FixStrategyDirect, got %v", finding.FixStrategy)
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

	finding := FromDiagnostic(d, fset, "tool", "R001")
	if len(finding.Related) != 1 {
		t.Fatalf("expected 1 related, got %d", len(finding.Related))
	}
	if finding.Related[0].Relation != "related" {
		t.Errorf("expected relation 'related', got %q", finding.Related[0].Relation)
	}
}

func TestAnalysisDiagnostic(t *testing.T) {
	t.Parallel()

	t.Run("with direct fix", func(t *testing.T) {
		t.Parallel()
		f := Finding{
			Message:     "test",
			FixStrategy: FixStrategyDirect,
			AfterCode:   "fixed",
			Suggestion:  "apply fix",
		}
		d := f.AnalysisDiagnostic()
		if d.Message != "test" {
			t.Errorf("expected message 'test', got %q", d.Message)
		}
		if len(d.SuggestedFixes) != 1 {
			t.Fatalf("expected 1 suggested fix, got %d", len(d.SuggestedFixes))
		}
		if string(d.SuggestedFixes[0].TextEdits[0].NewText) != "fixed" {
			t.Errorf("expected new text 'fixed'")
		}
	})

	t.Run("without fix", func(t *testing.T) {
		t.Parallel()
		f := Finding{
			Message:     "test",
			FixStrategy: FixStrategyNone,
		}
		d := f.AnalysisDiagnostic()
		if len(d.SuggestedFixes) != 0 {
			t.Errorf("expected 0 suggested fixes, got %d", len(d.SuggestedFixes))
		}
	})
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
	if p.File != "test.go" || p.Line != 10 || p.Column != 5 || p.Offset != 100 {
		t.Errorf("unexpected position: %+v", p)
	}
}

func TestNodePosition(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()
		fset := token.NewFileSet()
		p := NodePosition(fset, nil)
		if p.File != "" || p.Line != 0 || p.Column != 0 {
			t.Errorf("expected empty position for nil node, got %+v", p)
		}
	})

	t.Run("valid node", func(t *testing.T) {
		t.Parallel()
		fset := token.NewFileSet()
		src := `package main; func main() {}`
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
		r := NodeRange(nil, fset)
		if r.Start.File != "" || r.End.File != "" {
			t.Errorf("expected empty range for nil node, got %+v", r)
		}
	})

	t.Run("valid node", func(t *testing.T) {
		t.Parallel()
		fset := token.NewFileSet()
		src := `package main; func main() {}`
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

		r := NodeRange(decl, fset)
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
	src := `package main; func main() {}`
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
