package analysis

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
)

func TestNewAnalyzerDetector(t *testing.T) {
	t.Parallel()

	a := &analysis.Analyzer{
		Name: "test-analyzer",
		Doc:  "test",
		Run:  func(_ *analysis.Pass) (any, error) { return nil, nil }, //nolint:nilnil
	}

	d := NewAnalyzerDetector(a, []string{"."})

	if d.Name() != "test-analyzer" {
		t.Errorf("Name() = %q, want %q", d.Name(), "test-analyzer")
	}
}

func TestNewAnalyzerDetector_WithOptions(t *testing.T) {
	t.Parallel()

	a := &analysis.Analyzer{
		Name: "opt-test",
		Doc:  "test",
		Run:  func(_ *analysis.Pass) (any, error) { return nil, nil }, //nolint:nilnil
	}

	fset := token.NewFileSet()
	d := NewAnalyzerDetector(
		a, []string{"."},
		WithSeverity(finding.SeverityError),
		WithFileSet(fset),
	)

	if d.sev != finding.SeverityError {
		t.Errorf("severity = %v, want %v", d.sev, finding.SeverityError)
	}

	if d.fset != fset {
		t.Error("fset not set")
	}
}

func TestAnalyzerDetector_Detect_PackageErrors(t *testing.T) {
	t.Parallel()

	a := &analysis.Analyzer{
		Name: "err-test",
		Doc:  "test",
		Run:  func(_ *analysis.Pass) (any, error) { return nil, nil }, //nolint:nilnil
	}

	d := NewAnalyzerDetector(a, []string{"./nonexistent/..."})

	_, err := d.Detect(context.Background())
	if err == nil {
		t.Fatal("expected error for nonexistent package")
	}
}

func TestAnalyzerDetector_NilFactStubs(t *testing.T) {
	t.Parallel()

	obj := types.NewVar(token.NoPos, nil, "x", types.Typ[types.Int])

	var fact testFact

	if nilObjectFactImporter(obj, &fact) != false {
		t.Error("nilObjectFactImporter should return false")
	}

	if nilPackageFactImporter(nil, &fact) != false {
		t.Error("nilPackageFactImporter should return false")
	}

	nilFactExporter(obj, &fact)
	nilPackageFactExporter(&fact)
}

type testFact struct{}

func (testFact) AFact() {}

// stubAnalyzer is a minimal analyzer that reports a diagnostic for every function declaration.
//
//nolint:gochecknoglobals
var stubAnalyzer = &analysis.Analyzer{
	Name: "stub",
	Doc:  "reports every function declaration",
	Run: func(pass *analysis.Pass) (any, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				pass.Report(analysis.Diagnostic{
					Pos:     fn.Pos(),
					Message: "found function " + fn.Name.Name,
				})

				return true
			})
		}

		return nil, nil //nolint:nilnil
	},
}

func TestAnalyzerDetector_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	d := NewAnalyzerDetector(stubAnalyzer, []string{"."})

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("expected at least one finding from the analysis package")
	}

	for _, f := range findings {
		if f.ToolName != "stub" {
			t.Errorf("ToolName = %q, want %q", f.ToolName, "stub")
		}

		if f.Position.File == "" {
			t.Error("expected non-empty file position")
		}

		if f.Severity != finding.SeverityWarning {
			t.Errorf("Severity = %v, want %v", f.Severity, finding.SeverityWarning)
		}
	}
}

func TestAnalyzerDetector_Integration_WithSeverity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	d := NewAnalyzerDetector(
		stubAnalyzer, []string{"."},
		WithSeverity(finding.SeverityError),
	)

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("expected at least one finding")
	}

	for _, f := range findings {
		if f.Severity != finding.SeverityError {
			t.Errorf("Severity = %v, want %v", f.Severity, finding.SeverityError)
		}
	}
}
