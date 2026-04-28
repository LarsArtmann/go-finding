package finding

import (
	"iter"
	"sync"
)

// Report is the top-level container for a tool run.
// Use NewReport to create a thread-safe instance.
type Report struct {
	mu       *sync.Mutex // nil for zero-value Reports; initialized by NewReport
	Tool     ToolInfo    `json:"tool"`     // Tool metadata
	Findings []Finding   `json:"findings"` // All findings from this run
	Summary  Summary     `json:"summary"`  // Aggregated statistics
}

// ToolInfo contains metadata about the tool that generated the report.
type ToolInfo struct {
	Name    string `json:"name"`              // Tool name
	Version string `json:"version,omitempty"` // Tool version
}

// Summary contains aggregated statistics for a report.
type Summary struct {
	Total         int                 `json:"total"`                   // Total findings
	BySeverity    map[Severity]int    `json:"bySeverity"`              // Count by severity
	ByCategory    map[Category]int    `json:"byCategory,omitempty"`    // Count by category
	ByFixStrategy map[FixStrategy]int `json:"byFixStrategy,omitempty"` // Count by fix strategy
	FilesAffected int                 `json:"filesAffected,omitempty"` // Unique files with findings
	DurationMs    int64               `json:"durationMs,omitempty"`    // Execution time
	Suppressed    int                 `json:"suppressed,omitempty"`    // Count of suppressed findings
}

// NewReport creates a new report with the given tool info.
func NewReport(tool ToolInfo) *Report {
	r := &Report{
		mu:       &sync.Mutex{},
		Tool:     tool,
		Findings: make([]Finding, 0),
		Summary:  Summary{}, //nolint:exhaustruct
	}
	r.ComputeSummary()

	return r
}

// AddFinding adds a finding to the report.
// Safe for concurrent use.
func (r *Report) AddFinding(f Finding) {
	if r.mu != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	r.Findings = append(r.Findings, f)
}

// AddFindings adds multiple findings to the report.
// Safe for concurrent use.
func (r *Report) AddFindings(findings []Finding) {
	if r.mu != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	r.Findings = append(r.Findings, findings...)
}

// ComputeSummary recalculates the summary from the current findings.
func (r *Report) ComputeSummary() {
	r.Summary.Total = len(r.Findings)
	r.Summary.BySeverity = make(map[Severity]int)
	r.Summary.ByCategory = make(map[Category]int)
	r.Summary.ByFixStrategy = make(map[FixStrategy]int)

	files := make(map[string]struct{})
	suppressed := 0

	for _, f := range r.Findings {
		r.Summary.BySeverity[f.Severity]++

		r.Summary.ByFixStrategy[f.FixStrategy]++
		if f.Category != "" {
			r.Summary.ByCategory[f.Category]++
		}

		if f.Position.File != "" {
			files[f.Position.File] = struct{}{}
		}

		if f.IsSuppressed() {
			suppressed++
		}
	}

	r.Summary.FilesAffected = len(files)
	r.Summary.Suppressed = suppressed
}

// ActiveFindings returns all non-suppressed findings.
func (r *Report) ActiveFindings() []Finding {
	active := make([]Finding, 0, len(r.Findings))

	for _, f := range r.Findings {
		if !f.IsSuppressed() {
			active = append(active, f)
		}
	}

	return active
}

// BySeverity returns findings filtered by severity, excluding suppressed.
// For composable filtering, use filter.BySeverity with filter.NotSuppressed instead.
func (r *Report) BySeverity(sev Severity) []Finding {
	return Filter(r.ActiveFindings(), BySeverity(sev))
}

// ByCategory returns findings filtered by category, excluding suppressed.
// For composable filtering, use filter.ByCategory with filter.NotSuppressed instead.
func (r *Report) ByCategory(cat Category) []Finding {
	return Filter(r.ActiveFindings(), ByCategory(cat))
}

// ByFixStrategy returns findings filtered by fix strategy, excluding suppressed.
// For composable filtering, use filter.ByFixStrategy with filter.NotSuppressed instead.
func (r *Report) ByFixStrategy(fs FixStrategy) []Finding {
	return Filter(r.ActiveFindings(), ByFixStrategy(fs))
}

// FindByID returns a finding by its ID, or nil if not found.
// FindByID returns the finding with the given ID, or nil if not found.
// The returned Finding is a copy; modifications do not affect the report.
func (r *Report) FindByID(id string) *Finding {
	for _, f := range r.Findings {
		if f.ID == id {
			cp := f

			return &cp
		}
	}

	return nil
}

// FindByRule returns all non-suppressed findings matching the given rule name.
func (r *Report) FindByRule(rule string) []Finding {
	return Filter(r.ActiveFindings(), ByRule(rule))
}

// Len returns the number of findings in the report.
func (r *Report) Len() int {
	return len(r.Findings)
}

// All returns an iterator over all findings in the report.
// Supports break via yield returning false.
// All returns all findings in the report (including suppressed).
// The yielded Finding values are copies; modifications do not affect the report.
func (r *Report) All() iter.Seq[Finding] {
	return func(yield func(Finding) bool) {
		for _, f := range r.Findings {
			if !yield(f) {
				return
			}
		}
	}
}
