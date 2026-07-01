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
//   - go/analysis integration (see github.com/larsartmann/go-finding/analysis module)
//   - Report merging, deduplication, and cross-tool correlation
//   - Diff to compare finding sets
//   - Human-readable text and markdown formatting
//   - A pipeline for automated detect → triage → fix → verify loops (see github.com/larsartmann/go-finding/pipeline module)
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
//   - FixEdit: Byte-level edit operation (offset, length, replacement)
//
// # Validation
//
// Every Finding can be validated with Validate() which returns detailed per-field errors:
//
//	if err := f.Validate(); err != nil {
//	    // err contains joined errors for each invalid field
//	}
//
// Report and ToolInfo also have Validate methods. Builder calls Validate automatically on Build().
//
// # Filtering
//
// Filter findings using composable predicates:
//
//	errors := finding.Filter(findings, finding.BySeverity(finding.SeverityError))
//	autoFixable := finding.Filter(findings, finding.ByFixStrategy(finding.FixStrategyDirect))
//	byFile := finding.GroupByFile(findings)
//
// Combine with Negate for inverse filters, AnyOf for union:
//
//	nonAuto := finding.Filter(findings, finding.Negate(finding.WithFix))
//	warnOrErr := finding.Filter(findings, finding.AnyOf(
//	    finding.BySeverity(finding.SeverityWarning),
//	    finding.BySeverity(finding.SeverityError),
//	))
//
// FilterInPlace modifies the slice in place (zeroes tail for GC safety).
// ByConfidence and ByConfidenceAtLeast filter on confidence values.
//
// # Merging and Deduplication
//
// Merge reports from multiple tools with configurable deduplication:
//
//	merged := finding.Combine(reports,
//	    finding.WithDeduplication(true),
//	    finding.WithDeduplicateBy(finding.DeduplicateByPosition),
//	)
//
// Three deduplication strategies: ByID (exact match), ByPosition (file:line:col), ByRule (rule+position).
// Combine always deep-clones findings. Report.Merge merges in-place with shallow copy.
//
// # Cross-Tool Correlation
//
// Correlate finds related findings across different tools based on file proximity:
//
//	correlations := finding.Correlate(allFindings)
//	for _, c := range correlations {
//	    fmt.Printf("%s ↔ %s (score: %.2f): %s\n",
//	        c.FindingIDs[0], c.FindingIDs[1], float64(c.Score), c.Reason)
//	}
//
// # Diff
//
// Compare two finding sets to categorize changes:
//
//	result := finding.Diff(before, after)
//	fmt.Println(result.Stats()) // "+2 -1 ~0 =3"
//
// DiffResult contains Added, Removed, Modified (with before/after pairs), and Unchanged.
// Use HasChanges() for a quick check.
//
// # SARIF 2.1.0
//
// Export and import SARIF format for CI/CD integration:
//
//	data, err := report.ToSARIF()           // all findings
//	data, err := report.ToSARIFFiltered(sev) // filtered by severity + suppression
//
// Import back:
//
//	findings, err := finding.FindingsFromSARIF(ctx, data)
//	findings, err := finding.FindingsFromReader(ctx, reader) // streaming
//
// WriteSARIF/WriteSARIFFiltered stream directly to io.Writer.
// Report.WriteTo implements io.WriterTo for io.Copy compatibility.
//
// go-finding-specific properties are preserved in the SARIF property bag for full round-trip fidelity.
//
// # LSP Diagnostics
//
// Convert findings to LSP Diagnostics for IDE integration:
//
//	diags := f.ToLSP()
//
// Convert back from LSP:
//
//	f := finding.FromLSP(uri, lspDiag)
//
// LSP conversion is lossy: FixStrategy, Confidence, BeforeCode, AfterCode, Suppression,
// Metadata, Category, and Tags are not preserved through LSP round-trips.
// Diagnostic tags (unnecessary, deprecated) are preserved via Metadata.
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
// FindingError supports WithFinding and WithPosition for attaching context.
//
// # Suppression
//
// Findings can be suppressed with optional TTL:
//
//	f.Suppression = &finding.Suppression{
//	    Kind:      finding.SuppressionInSource,
//	    Rule:      "unused-var",
//	    Reason:    "intentionally unused in test",
//	    ExpiresAt: &expiry,
//	}
//	f.IsSuppressed()                // true
//	f.Suppression.IsActive(time.Now()) // true if not expired
//
// Use ActiveFindings() to get only non-suppressed findings from a Report.
//
// # Formatting
//
// Human-readable output formats:
//
//	finding.FormatText(os.Stdout, findings)    // single-line per finding
//	finding.FormatMarkdown(os.Stdout, findings) // markdown table
//
// # JSON
//
// JSON serialization with validation:
//
//	f, err := finding.FromJSON(data)       // single finding, validates
//	r, dropped, err := finding.ReportFromJSON(data) // report, drops invalid
//
// PrettyJSON includes all findings; PrettyJSONFiltered excludes suppressed.
//
// # ID Generation
//
// Stable, deterministic IDs:
//
//	id := finding.GenerateID("tool", "rule", finding.Position{File: "main.go", Line: 5})
//	parsed := finding.ParseID(id)
//
// IDs are colon-separated with length-prefixed fields to prevent collisions.
//
// # Pipeline
//
// The pipeline module (github.com/larsartmann/go-finding/pipeline) provides an
// automated detect → triage → fix → verify loop. Import it separately:
//
//	p, err := pipeline.New(pipeline.Config{
//	    MaxIterations:     3,
//	    ParallelDetectors: true,
//	    Timeout:           5 * time.Minute,
//	    VerifyAfterFix:    true,
//	}, rootDir, detector1, detector2)
//	result, err := p.Run(ctx)
//
// Pipeline features:
//   - Configurable iterations with early termination
//   - Parallel or sequential detector execution
//   - FindingTransformer chain between detection and triage
//   - Customizable TriageFunc for categorizing findings
//   - Byte-level FixEngine with composable FixProvider chain
//   - Conflict detection (position-based or byte-level)
//   - Post-fix verification by re-running detectors
//   - Retry with exponential backoff for flaky detectors
//   - Graceful degradation on detector failures
//   - Structured logging via slog
//   - Stage and iteration callbacks
//   - Metrics collection with snapshots
//
// # Fix Providers
//
// The FixEngine resolves findings to byte-level edits via a provider chain:
//
//   - OffsetProvider: Direct byte offset ranges
//   - LineProvider: Line/column positions converted to byte offsets
//   - SubstringProvider: BeforeCode text matching (fallback)
//
// Register custom providers for domain-specific transformations:
//
//	applier, err := pipeline.NewFixApplierWithProviders(rootDir, myASTProvider)
//
// A Go AST-aware provider (pipeline/goast.Provider) is available for .go files,
// using go/parser to disambiguate BeforeCode occurrences structurally.
//
// # Detector Registry
//
// Register named detector constructors for plugin-style extensibility:
//
//	registry := finding.NewDetectorRegistry()
//	registry.MustRegister("my-tool", func() finding.Detector { ... })
//	det, err := registry.Build("my-tool")
//	all, err := registry.BuildAll() // sorted by name
//
// Thread-safe. Use with ConfigFile.ResolveDetectors for config-driven pipelines.
//
// # Interval Index
//
// Efficient overlap queries over half-open ranges in O(log n + k):
//
//	idx := finding.NewIntervalIndex(intervals)
//	overlaps := idx.Query(start, end)
//
// Used internally by Correlate for spatial finding correlation.
//
// # Streaming Merge
//
// MergeIter yields findings from multiple reports as an iterator,
// avoiding intermediate slice allocation:
//
//	for f := range finding.MergeIter(reports, finding.WithDeduplication(true)) {
//	    process(f)
//	}
//
// # Converting from go/analysis
//
// Convert from the standard Go analysis framework using the analysis module
// (github.com/larsartmann/go-finding/analysis), imported separately:
//
//	f := analysis.FromDiagnostic(diag, pass.Fset, "my-analyzer", "RULE001")
//
// # Known Limitations
//
// SeverityCritical maps to SARIF level "error" (SARIF 2.1.0 has no "critical" level).
// The original severity is preserved in the SARIF property bag for round-trip fidelity.
//
// LSP conversion is lossy: FixStrategy, Confidence, BeforeCode, AfterCode, Suppression,
// Metadata, Category, and Tags are not preserved through LSP round-trips.
//
// Report.findings is unexported for thread safety. Use AddFinding/AddFindings
// for writes, FindingsSnapshot/All/FindByID for reads.
package finding
