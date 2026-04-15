package finding

// Finding represents a single issue detected by a static analysis tool.
type Finding struct {
	// Identity
	ID       string `json:"id"`       // Stable unique identifier (e.g., "tool:rule:file:42:5")
	Rule     string `json:"rule"`     // Rule/check name (e.g., "STRONG_ID", "clone-detected")
	ToolName string `json:"toolName"` // Source tool name (e.g., "branching-flow", "art-dupl")

	// Core
	Message  string   `json:"message"`  // Human-readable description
	Severity Severity `json:"severity"` // info, warning, error, critical
	Position Position `json:"position"` // Where the issue is

	// Classification
	Category Category `json:"category,omitempty"` // Domain: "security", "style", "duplication", etc.
	Tag      string `json:"tag,omitempty"`      // Sub-classification: "phantom-type", "clone", etc.

	// Fix
	FixStrategy FixStrategy `json:"fixStrategy"`          // none, suggest, direct, ai
	Suggestion  string      `json:"suggestion,omitempty"` // Human-readable fix description
	BeforeCode  string      `json:"beforeCode,omitempty"` // Code before the fix
	AfterCode   string      `json:"afterCode,omitempty"`  // Code after the fix

	// Context
	Range       *Range       `json:"range,omitempty"`       // For span-based findings
	Snippet     string       `json:"snippet,omitempty"`     // Surrounding code context
	Confidence  float64      `json:"confidence,omitempty"`  // 0.0-1.0
	Related     []RelatedRef `json:"related,omitempty"`     // Related findings
	Suppression *Suppression `json:"suppression,omitempty"` // If suppressed

	// Extensibility
	Metadata map[string]string `json:"metadata,omitempty"` // Tool-specific key-value pairs
}

// RelatedRef links to another finding.
type RelatedRef struct {
	FindingID string   `json:"findingId"` // ID of the related finding
	Relation  string   `json:"relation"`  // e.g., "clone-of", "wraps", "causes"
	Position  Position `json:"position"`  // Quick access to related location
}

// IsValid returns true if the reference has a non-empty FindingID.
func (r RelatedRef) IsValid() bool {
	return r.FindingID != ""
}

// IsSuppressed returns true if this finding is suppressed.
func (f Finding) IsSuppressed() bool {
	return f.Suppression != nil && !f.Suppression.IsExpired()
}

// HasFix returns true if this finding has a fix available.
func (f Finding) HasFix() bool {
	return f.FixStrategy == FixStrategyDirect || f.FixStrategy == FixStrategyAI
}

// HasSuggestion returns true if this finding has a human-readable suggestion.
func (f Finding) HasSuggestion() bool {
	return f.Suggestion != "" || (f.BeforeCode != "" && f.AfterCode != "")
}

// IsValid returns true if the finding has required fields set.
func (f Finding) IsValid() bool {
	return f.ID != "" && f.Rule != "" && f.ToolName != "" &&
		f.Message != "" && f.Position.IsValid() && f.Severity.IsValid()
}
