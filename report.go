package finding

import (
	"errors"
	"fmt"
	"iter"
	"sync"
	"time"
)

// Report is the top-level container for a tool run.
// The zero value is safe for concurrent use. Use [NewReport] to create
// a Report with pre-allocated findings.
// All methods are safe for concurrent use. Read methods (FindByID, Len,
// ActiveFindings, etc.) acquire a read lock; write methods (AddFinding,
// AddFindings, MergeInto) acquire a write lock.
type Report struct {
	mu   sync.RWMutex
	Tool ToolInfo `json:"tool"` // Tool metadata
	// Findings holds all findings from this run.
	//
	// Deprecated: Direct access is not thread-safe. Use [Report.FindingsSnapshot]
	// for a deep copy, [Report.All] for iteration, or [Report.FindByID] for single
	// lookups. This field will be unexported in v1.0.
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"` // Aggregated statistics
}

// Validate returns an error if the Report is invalid.
// It checks Tool info and validates each finding, returning joined errors.
// Safe for concurrent use.
func (r *Report) Validate() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var errs []error

	err := r.Tool.Validate()
	if err != nil {
		errs = append(errs, err)
	}

	for i, f := range r.Findings {
		err := f.Validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("findings[%d]: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// ToolInfo contains metadata about the tool that generated the report.
type ToolInfo struct {
	Name    string `json:"name"`              // Tool name
	Version string `json:"version,omitempty"` // Tool version
}

// Validate returns an error if the ToolInfo is invalid.
// A valid ToolInfo requires a non-empty Name.
func (t ToolInfo) Validate() error {
	if t.Name == "" {
		return NewValidationError("ToolInfo.Name is required", nil)
	}

	return nil
}

// Summary contains aggregated statistics for a report.
type Summary struct {
	Total         int                 `json:"total"`                   // Total findings
	BySeverity    map[Severity]int    `json:"bySeverity"`              // Count by severity
	ByCategory    map[Category]int    `json:"byCategory,omitempty"`    // Count by category
	ByFixStrategy map[FixStrategy]int `json:"byFixStrategy,omitempty"` // Count by fix strategy
	FilesAffected int                 `json:"filesAffected,omitempty"` // Unique files with findings
	FilesScanned  int                 `json:"filesScanned,omitempty"`  // Total files scanned (including clean files)
	DurationMs    int64               `json:"durationMs,omitempty"`    // Execution time (caller-set; not computed by ComputeSummary)
	Suppressed    int                 `json:"suppressed,omitempty"`    // Count of suppressed findings
}

// NewReport creates a new report with the given tool info.
func NewReport(tool ToolInfo) *Report {
	r := &Report{ //nolint:exhaustruct
		Tool:     tool,
		Findings: make([]Finding, 0),
		Summary:  Summary{}, //nolint:exhaustruct
	}
	r.ComputeSummary()

	return r
}

// newReportWithCapacity creates a new report with pre-allocated finding capacity.
func newReportWithCapacity(tool ToolInfo, capacity int) *Report {
	r := &Report{ //nolint:exhaustruct
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
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Findings = append(r.Findings, f)
}

// AddFindings adds multiple findings to the report.
// Safe for concurrent use.
func (r *Report) AddFindings(findings []Finding) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Findings = append(r.Findings, findings...)
}

// Merge merges another report's findings into this report in-place.
// The Tool info from other is ignored — this report retains its own.
// Summary is recomputed after merging.
// Safe for concurrent use.
//
// Deprecated: Use [Report.MergeInto] instead, which returns a new Report
// without modifying the receiver. Merge will be removed in v1.0.0.
func (r *Report) Merge(other *Report) {
	other.mu.RLock()
	cloned := make([]Finding, len(other.Findings))
	copy(cloned, other.Findings)
	other.mu.RUnlock()

	r.mu.Lock()
	r.Findings = append(r.Findings, cloned...)
	r.mu.Unlock()

	r.ComputeSummary()
}

// MergeInto returns a new Report containing findings from both r and other.
// Neither receiver nor other is modified. The new report uses r's ToolInfo.
func (r *Report) MergeInto(other *Report) *Report {
	r.mu.RLock()
	rFindings := make([]Finding, len(r.Findings))
	copy(rFindings, r.Findings)
	rTool := r.Tool
	r.mu.RUnlock()

	other.mu.RLock()
	oFindings := make([]Finding, len(other.Findings))
	copy(oFindings, other.Findings)
	other.mu.RUnlock()

	merged := &Report{ //nolint:exhaustruct
		Tool:     rTool,
		Findings: make([]Finding, 0, len(rFindings)+len(oFindings)),
	}
	merged.Findings = append(merged.Findings, rFindings...)
	merged.Findings = append(merged.Findings, oFindings...)
	merged.ComputeSummary()

	return merged
}

// addFindingUnchecked appends a finding without acquiring the mutex.
// Caller must hold the lock or guarantee single-goroutine access.
func (r *Report) addFindingUnchecked(f Finding) {
	r.Findings = append(r.Findings, f)
}

// readFindings returns a shallow copy of findings under RLock.
// Safe for concurrent use. The caller receives a snapshot that won't
// be affected by subsequent AddFinding/AddFindings calls.
func (r *Report) readFindings() []Finding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	findings := make([]Finding, len(r.Findings))
	copy(findings, r.Findings)

	return findings
}

// ComputeSummary recalculates the summary from the current findings.
// Uses time.Now() for suppression expiry checks. For deterministic results
// in tests, use ComputeSummaryAt.
// Safe for concurrent use with AddFinding/AddFindings.
func (r *Report) ComputeSummary() {
	r.computeSummaryAt(time.Now())
}

// ComputeSummaryAt recalculates the summary using the given time for
// suppression expiry checks. Use this in tests for deterministic results.
func (r *Report) ComputeSummaryAt(now time.Time) {
	r.computeSummaryAt(now)
}

func (r *Report) computeSummaryAt(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

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

		if f.IsSuppressedAt(now) {
			suppressed++
		}
	}

	r.Summary.FilesAffected = len(files)
	r.Summary.Suppressed = suppressed
}

// FindingsSnapshot returns a deep copy of all findings in the report.
// The returned slice is safe for concurrent use without holding any lock,
// making it suitable for v1.0 migration from direct Findings slice access.
// Each Finding is fully cloned via Clone(), so mutations are isolated.
func (r *Report) FindingsSnapshot() []Finding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.Findings) == 0 {
		return nil
	}

	snapshot := make([]Finding, len(r.Findings))
	for i, f := range r.Findings {
		snapshot[i] = f.Clone()
	}

	return snapshot
}

// ActiveFindings returns all non-suppressed findings.
// Uses time.Now() for suppression expiry checks. For deterministic results
// in tests, filter Findings directly with IsSuppressedAt.
// Safe for concurrent use.
func (r *Report) ActiveFindings() []Finding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	active := make([]Finding, 0, len(r.Findings))

	for _, f := range r.Findings {
		if !f.IsSuppressedAt(now) {
			active = append(active, f)
		}
	}

	return active
}

// BySeverity returns findings filtered by severity, excluding suppressed.
// For composable filtering, use filter.BySeverity with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) BySeverity(sev Severity) []Finding {
	return Filter(r.ActiveFindings(), BySeverity(sev))
}

// CountBySeverity returns the count of findings for the given severity,
// including suppressed findings. Uses the pre-computed summary.
func (r *Report) CountBySeverity(sev Severity) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Summary.BySeverity[sev]
}

// ByCategory returns findings filtered by category, excluding suppressed.
// For composable filtering, use filter.ByCategory with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) ByCategory(cat Category) []Finding {
	return Filter(r.ActiveFindings(), ByCategory(cat))
}

// ByFixStrategy returns findings filtered by fix strategy, excluding suppressed.
// For composable filtering, use filter.ByFixStrategy with filter.NotSuppressed instead.
// Safe for concurrent use.
func (r *Report) ByFixStrategy(fs FixStrategy) []Finding {
	return Filter(r.ActiveFindings(), ByFixStrategy(fs))
}

// FindByID returns the finding with the given ID, or nil if not found.
// The returned Finding is a shallow copy; modifications to value fields do not
// affect the report, but mutations to slice/map fields (Tags, Related, Metadata)
// will be shared. Use Clone() for a deep copy.
// Safe for concurrent use.
func (r *Report) FindByID(id string) *Finding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, f := range r.Findings {
		if f.ID == id {
			cp := f

			return &cp
		}
	}

	return nil
}

// FindByRule returns all non-suppressed findings matching the given rule name.
// Safe for concurrent use.
func (r *Report) FindByRule(rule string) []Finding {
	return Filter(r.ActiveFindings(), ByRule(rule))
}

// Len returns the number of findings in the report.
// Safe for concurrent use.
func (r *Report) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Findings)
}

// Filter returns a new report containing only findings that match all predicates.
// Safe for concurrent use.
func (r *Report) Filter(predicates ...FilterFunc) *Report {
	r.mu.RLock()
	filtered := Filter(r.Findings, predicates...)
	r.mu.RUnlock()

	result := NewReport(r.Tool)
	result.AddFindings(filtered)

	return result
}

// Map returns a new report with the given function applied to each finding.
// Safe for concurrent use.
func (r *Report) Map(fn func(Finding) Finding) *Report {
	r.mu.RLock()
	findings := make([]Finding, len(r.Findings))
	copy(findings, r.Findings)
	r.mu.RUnlock()

	result := NewReport(r.Tool)
	for _, f := range findings {
		result.AddFinding(fn(f))
	}

	return result
}

// All returns all findings in the report (including suppressed).
// The yielded Finding values are shallow copies; modifications to value fields
// do not affect the report, but mutations to slice/map fields (Tags, Related,
// Metadata) will be shared. Use Clone() for a deep copy.
//
// IMPORTANT: The returned iterator holds a read lock for the duration of
// iteration. You MUST exhaust the iterator (e.g., with a break or range)
// to release the lock. If you need a snapshot without holding the lock,
// call ActiveFindings() or use Filter.
func (r *Report) All() iter.Seq[Finding] {
	return func(yield func(Finding) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		for _, f := range r.Findings {
			if !yield(f) {
				return
			}
		}
	}
}
