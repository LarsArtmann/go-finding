package finding

import "maps"

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

// Equal reports whether two findings are identical, including all nested fields.
func (f Finding) Equal(other Finding) bool {
	if f.ID != other.ID || f.Rule != other.Rule || f.ToolName != other.ToolName ||
		f.Message != other.Message || f.Severity != other.Severity ||
		!f.Position.Equal(other.Position) ||
		f.Category != other.Category || f.Tag != other.Tag ||
		f.FixStrategy != other.FixStrategy ||
		f.Suggestion != other.Suggestion ||
		f.BeforeCode != other.BeforeCode || f.AfterCode != other.AfterCode ||
		f.Snippet != other.Snippet || f.Confidence != other.Confidence {
		return false
	}

	if !f.equalRange(other) {
		return false
	}

	if len(f.Related) != len(other.Related) {
		return false
	}
	for i, r := range f.Related {
		if r != other.Related[i] {
			return false
		}
	}

	if !f.equalSuppression(other) {
		return false
	}

	return maps.Equal(f.Metadata, other.Metadata)
}

func (f Finding) equalRange(other Finding) bool {
	if f.Range == nil && other.Range == nil {
		return true
	}
	if f.Range == nil || other.Range == nil {
		return false
	}
	return f.Range.Equal(*other.Range)
}

func (f Finding) equalSuppression(other Finding) bool {
	if f.Suppression == nil && other.Suppression == nil {
		return true
	}
	if f.Suppression == nil || other.Suppression == nil {
		return false
	}
	return *f.Suppression == *other.Suppression
}
