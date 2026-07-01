package analysis

import (
	"context"
	"fmt"
	"go/token"
	"go/types"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

// AnalyzerDetector wraps a go/analysis.Analyzer to implement finding.Detector.
// It runs the analyzer on Go packages specified by the provided patterns.
type AnalyzerDetector struct {
	analyzer *analysis.Analyzer
	patterns []string
	fset     *token.FileSet
	sev      finding.Severity
}

// AnalyzerOption configures an AnalyzerDetector.
type AnalyzerOption func(*AnalyzerDetector)

// WithSeverity sets the default severity for findings (defaults to SeverityWarning).
func WithSeverity(sev finding.Severity) AnalyzerOption {
	return func(d *AnalyzerDetector) { d.sev = sev }
}

// WithFileSet sets a shared token.FileSet for position resolution.
// If not set, a new FileSet is created per Detect call.
func WithFileSet(fset *token.FileSet) AnalyzerOption {
	return func(d *AnalyzerDetector) { d.fset = fset }
}

// NewAnalyzerDetector creates a Detector that runs the given go/analysis.Analyzer
// on Go packages matching the provided patterns (e.g., "./...", "./pkg/...").
func NewAnalyzerDetector(
	a *analysis.Analyzer,
	patterns []string,
	opts ...AnalyzerOption,
) *AnalyzerDetector {
	d := &AnalyzerDetector{ //nolint:exhaustruct
		analyzer: a,
		patterns: patterns,
		sev:      finding.SeverityWarning,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Name returns the analyzer's name.
func (d *AnalyzerDetector) Name() string {
	return d.analyzer.Name
}

// Compile-time interface assertion — catches missing methods at build time.
var _ finding.Detector = (*AnalyzerDetector)(nil)

// Detect runs the analyzer on the configured packages and converts
// diagnostics to findings.
func (d *AnalyzerDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	fset := d.fset
	if fset == nil {
		fset = token.NewFileSet()
	}

	cfg := &packages.Config{ //nolint:exhaustruct
		Context: ctx,
		Fset:    fset,
		Tests:   false,
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedTypesSizes,
	}

	pkgs, err := packages.Load(cfg, d.patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages %v: %w", d.patterns, err)
	}

	var allFindings []finding.Finding

	for _, pkg := range pkgs {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		findings, err := d.analyzePackage(pkg, fset)
		if err != nil {
			return nil, fmt.Errorf("analyze package %s: %w", pkg.PkgPath, err)
		}

		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

func (d *AnalyzerDetector) analyzePackage(
	pkg *packages.Package,
	fset *token.FileSet,
) ([]finding.Finding, error) {
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package %s has errors: %w", pkg.PkgPath, pkg.Errors[0])
	}

	diagnostics, err := d.runAnalyzer(pkg, fset)
	if err != nil {
		return nil, err
	}

	findings := make([]finding.Finding, 0, len(diagnostics))

	for _, diag := range diagnostics {
		ruleCode := d.analyzer.Name
		if diag.Category != "" {
			ruleCode = diag.Category
		}

		f := FromDiagnostic(diag, fset, d.analyzer.Name, ruleCode, d.sev)
		findings = append(findings, f)
	}

	return findings, nil
}

func (d *AnalyzerDetector) runAnalyzer(
	pkg *packages.Package,
	fset *token.FileSet,
) ([]*analysis.Diagnostic, error) {
	var diagnostics []*analysis.Diagnostic

	pass := &analysis.Pass{ //nolint:exhaustruct
		Analyzer:   d.analyzer,
		Fset:       fset,
		Files:      pkg.Syntax,
		OtherFiles: pkg.OtherFiles,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf:   make(map[*analysis.Analyzer]any),
		Report: func(diag analysis.Diagnostic) {
			diagnostics = append(diagnostics, &diag)
		},
		ImportObjectFact:  nilObjectFactImporter,
		ImportPackageFact: nilPackageFactImporter,
		ExportObjectFact:  nilFactExporter,
		ExportPackageFact: nilPackageFactExporter,
		AllObjectFacts:    func() []analysis.ObjectFact { return nil },
		AllPackageFacts:   func() []analysis.PackageFact { return nil },
	}

	_, err := d.analyzer.Run(pass)

	return diagnostics, err //nolint:wrapcheck // thin adapter, caller wraps
}

// Nil fact stubs for analyzers that don't use facts.

func nilObjectFactImporter(_ types.Object, _ analysis.Fact) bool { return false }

func nilPackageFactImporter(_ *types.Package, _ analysis.Fact) bool { return false }

func nilFactExporter(_ types.Object, _ analysis.Fact) {}

func nilPackageFactExporter(_ analysis.Fact) {}
