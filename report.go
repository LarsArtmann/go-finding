package finding

import (
	"iter"
	"sync"
)

// Report is the top-level container for a tool run.
// The zero value is safe for concurrent use. Use [NewReport] to create
// a Report with pre-allocated findings.
type Report struct {
	mu       sync.Mutex
	Tool     ToolInfo  `json:"tool"`     // Tool metadata
	Findings []Finding `json:"findings"` // All findings from this run
	Summary  Summary   `json:"summary"`  // Aggregated statistics
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
		Tool:     tool,
		Findings: make([]Finding, 0),
		Summary:  Summary{}, //nolint:exhaustruct
	}
	r.ComputeSummary()

	return r
}

// newReportWithCapacity creates a new report with pre-allocated finding capacity.
func newReportWithCapacity(tool ToolInfo, capacity int) *Report {
	r := &Report{
		Tool:     tool,
		Findings: make([]Finding, 0, capacity),
		Summary:  Summary{}, //nolint:exhaustruct
	}
	r.ComputeSummary()

	return r
}

// AddFinding adds a finding to the report.
// Safe for concurrent use.
func (r *Report) AddFinding(f Finding) {
	r.lock()
	r.Findings = append(r.Findings, f)
	r.unlock()
}

// AddFindings adds multiple findings to the report.
// Safe for concurrent use.
func (r *Report) AddFindings(findings []Finding) {
	r.lock()
	r.Findings = append(r.Findings, findings...)
	r.unlock()
}

// Merge merges another report's findings into this report in-place.
// The Tool info from other is ignored — this report retains its own.
// Summary is recomputed after merging.
// Safe for concurrent use.
func (r *Report) Merge(other *Report) {
	r.lock()
	r.Findings = append(r.Findings, other.Findings...)
	r.unlock()

	r.ComputeSummary()
}

// addFindingUnchecked appends a finding without acquiring the mutex.
// Caller must hold the lock or guarantee single-goroutine access.
func (r *Report) addFindingUnchecked(f Finding) {
	r.Findings = append(r.Findings, f)
}

// lock acquires the report mutex.
func (r *Report) lock() {
	r.mu.Lock()
}

// unlock releases the report mutex.
func (r *Report) unlock() {
	r.mu.Unlock()
}

// ComputeSummary recalculates the summary from the current findings.
// Safe for concurrent use with AddFinding/AddFindings.
func (r *Report) ComputeSummary() {
	r.lock()
	defer r.unlock()

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

// Filter returns a new report containing only findings that match all predicates.
func (r *Report) Filter(predicates ...FilterFunc) *Report {
	filtered := Filter(r.Findings, predicates...)
	result := NewReport(r.Tool)
	result.AddFindings(filtered)
	return result
}

// Map returns a new report with the given function applied to each finding.
func (r *Report) Map(fn func(Finding) Finding) *Report {
	result := NewReport(r.Tool)
	for _, f := range r.Findings {
		result.AddFinding(fn(f))
	}
	return result
}

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
