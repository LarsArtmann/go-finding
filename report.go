package finding

// Report is the top-level container for a tool run.
type Report struct {
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
	ByCategory    map[string]int      `json:"byCategory,omitempty"`    // Count by category
	ByFixStrategy map[FixStrategy]int `json:"byFixStrategy,omitempty"` // Count by fix strategy
	FilesAffected int                 `json:"filesAffected,omitempty"` // Unique files with findings
	DurationMs    int64               `json:"durationMs,omitempty"`  // Execution time
	Suppressed    int                 `json:"suppressed,omitempty"`    // Count of suppressed findings
}

// NewReport creates a new report with the given tool info.
func NewReport(tool ToolInfo) *Report {
	return &Report{
		Tool:     tool,
		Findings: make([]Finding, 0),
		Summary:  Summary{},
	}
}

// AddFinding adds a finding to the report.
func (r *Report) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
}

// AddFindings adds multiple findings to the report.
func (r *Report) AddFindings(findings []Finding) {
	r.Findings = append(r.Findings, findings...)
}

// ComputeSummary recalculates the summary from the current findings.
func (r *Report) ComputeSummary() {
	r.Summary.Total = len(r.Findings)
	r.Summary.BySeverity = make(map[Severity]int)
	r.Summary.ByCategory = make(map[string]int)
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
func (r Report) ActiveFindings() []Finding {
	var active []Finding
	for _, f := range r.Findings {
		if !f.IsSuppressed() {
			active = append(active, f)
		}
	}
	return active
}

// BySeverity returns findings filtered by severity.
func (r Report) BySeverity(sev Severity) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		if f.Severity == sev && !f.IsSuppressed() {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// ByCategory returns findings filtered by category.
func (r Report) ByCategory(cat string) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		if f.Category == cat && !f.IsSuppressed() {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// ByFixStrategy returns findings filtered by fix strategy.
func (r Report) ByFixStrategy(fs FixStrategy) []Finding {
	var filtered []Finding
	for _, f := range r.Findings {
		if f.FixStrategy == fs && !f.IsSuppressed() {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FindByID returns a finding by its ID, or nil if not found.
func (r Report) FindByID(id string) *Finding {
	for i, f := range r.Findings {
		if f.ID == id {
			return &r.Findings[i]
		}
	}
	return nil
}
