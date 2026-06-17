package finding

import (
	"errors"
	"fmt"
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

// findingsLocked returns the findings slice without acquiring the lock.
// Caller MUST hold r.mu (RLock or Lock). Used internally to prepare for
// v1.0 unexport of the Findings field.
func (r *Report) findingsLocked() []Finding {
	return r.Findings
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
