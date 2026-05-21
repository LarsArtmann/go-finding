// Package finding provides a unified data model and pipeline for static analysis tools.
//
// The finding package solves the fragmentation problem in Go's static analysis
// ecosystem where each tool invents its own types for findings. It provides:
//   - A common Finding type that all tools can use
//   - Standard severity levels (info, warning, error, critical)
//   - Named types for Confidence, Category, FixStrategy, Tag, SuppressionKind
//   - Position tracking with range support
//   - SARIF 2.1.0 output generation and import
//   - LSP Diagnostic conversion
//   - go/analysis integration (see analysis subpackage)
//   - Report merging, deduplication, and cross-tool correlation
//   - A pipeline for automated detect → triage → fix → verify loops
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
// Or use the Builder API for construction with validation:
//
//	f, err := finding.NewBuilder("unused-var", "my-tool", "variable x is unused",
//	    finding.SeverityWarning, finding.Pos("main.go", 5, 2)).
//	    WithCategory(finding.CategoryUnused).
//	    WithConfidence(finding.ConfidenceHigh).
//	    Build()
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
// The main types are Finding, Report, and supporting named types:
//
//   - Finding: A single issue detected by a tool
//   - Report: Thread-safe container for all findings from a tool run
//   - Severity: info, warning, error, critical (with comparison operators)
//   - Confidence: Named float64 type with IsValid/Clamp, range [0.0, 1.0]
//   - FixStrategy: none, suggest, direct, ai (ai is reserved)
//   - Category: 14 predefined + custom (security, style, performance, etc.)
//   - Tag: Multi-label classification (security, bug, deprecated, etc.)
//   - Position: File, line, column, offset location
//   - Range: Start and end positions with spatial operations (Contains, Overlaps, Adjacent)
//   - Suppression: Mark findings as suppressed with kind, reason, and optional expiry
//
// # Validation
//
// Every Finding can be validated with Validate() which returns detailed per-field errors:
//
//	if err := f.Validate(); err != nil {
//	    // err contains joined errors for each invalid field
//	}
//
// # Filtering
//
// Filter findings using composable predicates:
//
//	errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))
//	autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))
//	byFile := finding.GroupByFile(findings)
//
// Combine with Negate for inverse filters:
//
//	nonAuto := finding.Filter(findings, finding.Negate(finding.HasFix))
//
// # Merging and Deduplication
//
// Merge reports from multiple tools with configurable deduplication:
//
//	merged := finding.Merge(reports,
//	    finding.WithDeduplication(true),
//	    finding.WithDeduplicateBy(finding.DeduplicateByPosition),
//	)
//
// Three deduplication strategies: ByID (exact match), ByPosition, ByRule.
//
// # Cross-Tool Correlation
//
// Correlate finds related findings across different tools:
//
//	correlations := finding.Correlate(allFindings)
//
// # Diff
//
// Compare two finding sets:
//
//	result := finding.Diff(before, after)
//	fmt.Println(result.Stats()) // "+2 -1 =3"
//
// # Error Handling
//
// Structured error types with category-based classification:
//
//	err := finding.NewValidationError("missing field", nil)
//	errors.Is(err, finding.ErrValidation) // true
//
// Five error categories: Validation, IO, Parse, Conflict, Internal.
// Use IsFindingError, GetCategory, IsCategory for programmatic handling.
//
// # Suppression
//
// Findings can be suppressed with a TTL:
//
//	f.Suppression = &finding.Suppression{
//	    Kind:      finding.SuppressionKindInSource,
//	    Rule:      "unused-var",
//	    Reason:    "intentionally unused in test",
//	    ExpiresAt: &expiry,
//	}
//	f.IsSuppressed()   // true
//	f.Suppression.IsActive(time.Now()) // true if not expired
//
// # Converting from go/analysis
//
// Convert from the standard Go analysis framework using the analysis subpackage:
//
//	f := analysis.FromDiagnostic(diag, pass.Fset, "my-analyzer", "RULE001")
//
// # Pipeline
//
// The pipeline subpackage provides an automated detect → triage → fix → verify loop:
//
//	p, err := pipeline.New(pipeline.Config{
//	    Detectors: []pipeline.Detector{myDetector},
//	    Timeout:   5 * time.Minute,
//	})
//	result, err := p.Run(ctx)
//
// See the pipeline subpackage for configuration, custom fix providers, metrics,
// retry, partial success, and verification.
//
// # Known Limitations
//
// SeverityCritical maps to SARIF level "error" (SARIF 2.1.0 has no "critical" level).
// The original severity is preserved in Properties["go-finding/severity"] for round-trip fidelity.
//
// LSP conversion is lossy: FixStrategy, Confidence, BeforeCode, AfterCode, Suppression,
// Metadata, Category, and Tags are not preserved through LSP round-trips.
//
// # Related Projects
//
//   - go/analysis: The standard Go analysis framework
//   - SARIF 2.1.0: Static Analysis Results Interchange Format
//   - LSP: Language Server Protocol
package finding
