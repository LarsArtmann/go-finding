// Package finding provides a unified data model and pipeline for static analysis tools.
//
// The finding package solves the fragmentation problem in Go's static analysis
// ecosystem where each tool invents its own types for findings. It provides:
//   - A common Finding type that all tools can use
//   - Standard severity levels (info, warning, error, critical)
//   - Fix strategies (none, suggest, direct, ai)
//   - Position tracking with range support
//   - SARIF 2.1.0 output generation
//   - LSP Diagnostic conversion
//   - go/analysis integration
//   - Report merging and filtering
//
// # Quick Start
//
// Create a finding:
//
//	f := finding.Finding{
//	    ID:       finding.GenerateID("my-tool", "unused-var", finding.Position{File: "main.go", Line: 5}),
//	    Rule:     "unused-var",
//	    ToolName: "my-tool",
//	    Message:  "variable x is unused",
//	    Severity: finding.SeverityWarning,
//	    Position: finding.Position{File: "main.go", Line: 5, Column: 2},
//	}
//
// Create a report:
//
//	report := finding.NewReport(finding.ToolInfo{Name: "my-tool"})
//	report.AddFinding(f)
//	report.ComputeSummary()
//
// Output as SARIF:
//
//	sarifJSON, err := report.ToSARIF()
//
// # Core Types
//
// The main types are Finding, Report, and the supporting types:
//
//   - Finding: A single issue detected by a tool
//   - Report: Container for all findings from a tool run
//   - Severity: info, warning, error, critical
//   - FixStrategy: none, suggest, direct, ai
//   - Position: File, line, column location
//   - Range: Start and end positions
//
// # Filtering
//
// Filter findings using predicates:
//
//	errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))
//	autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))
//	byFile := finding.GroupByFile(findings)
//
// # Converting from go/analysis
//
// Convert from the standard Go analysis framework:
//
//	finding := finding.FromDiagnostic(diag, pass.Fset, "my-analyzer", "RULE001")
//
// # Pipeline
//
// The package includes a pipeline for automated fixing:
//
//  1. Detect: Run tools and collect findings
//  2. Triage: Route by fix strategy
//  3. Fix: Apply direct fixes, route AI fixes
//  4. Verify: Re-run and validate
//
// See the pipeline subpackage for details.
//
// # Cross-Tool Correlation
//
// Correlate finds related findings across different tools using simple heuristics
// (same file, nearby lines). It is a standalone utility, not wired into the pipeline:
//
//	correlations := finding.Correlate(allFindings)
//	for _, c := range correlations {
//	    fmt.Printf("%v are related: %s (%.1f)\n", c.FindingIDs, c.Reason, c.Confidence)
//	}
//
// # Known Limitations
//
// SeverityCritical maps to SARIF level "error" (SARIF 2.1.0 has no "critical" level).
// The original severity is preserved in Properties["go-finding/severity"] for round-trip fidelity.
//
// # Related Projects
//
//   - go/analysis: The standard Go analysis framework
//   - SARIF 2.1.0: Static Analysis Results Interchange Format
//   - LSP: Language Server Protocol
package finding
