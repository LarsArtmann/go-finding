package finding

import (
	"fmt"
	"maps"
	"math"
	"time"
)

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
	Tag      string   `json:"tag,omitempty"`      // Sub-classification: "phantom-type", "clone", etc.

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

// NewFinding creates a Finding with an auto-generated ID and default fix strategy.
func NewFinding(rule, toolName, message string, severity Severity, pos Position) Finding {
	return Finding{ //nolint:exhaustruct
		ID:          GenerateID(toolName, rule, pos),
		Rule:        rule,
		ToolName:    toolName,
		Message:     message,
		Severity:    severity,
		Position:    pos,
		FixStrategy: FixStrategyNone,
	}
}

func clampConfidence(c float64) float64 {
	if c < 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return c
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

// Clone returns a deep copy of the finding.
func (f Finding) Clone() Finding {
	clone := f

	if f.Range != nil {
		r := *f.Range
		clone.Range = &r
	}

	if len(f.Related) > 0 {
		clone.Related = make([]RelatedRef, len(f.Related))
		copy(clone.Related, f.Related)
	}

	if f.Suppression != nil {
		s := *f.Suppression
		if s.ExpiresAt != nil {
			t := *s.ExpiresAt
			s.ExpiresAt = &t
		}

		clone.Suppression = &s
	}

	if len(f.Metadata) > 0 {
		clone.Metadata = maps.Clone(f.Metadata)
	}

	return clone
}

// IsSuppressed returns true if this finding is suppressed at the current time.
func (f Finding) IsSuppressed() bool {
	return f.IsSuppressedAt(time.Now())
}

// IsSuppressedAt returns true if this finding is suppressed at the given time.
// Use this in tests for deterministic suppression checks.
func (f Finding) IsSuppressedAt(now time.Time) bool {
	return f.Suppression != nil && !f.Suppression.IsExpired(now)
}

// HasFix returns true if this finding has a fix available.
func (f Finding) HasFix() bool {
	switch f.FixStrategy {
	case FixStrategyNone:
		return false
	case FixStrategyDirect, FixStrategyAI:
		return true
	case FixStrategySuggest:
		return f.AfterCode != ""
	default:
		return false
	}
}

// HasSuggestion returns true if this finding has a human-readable suggestion.
func (f Finding) HasSuggestion() bool {
	return f.Suggestion != "" || (f.BeforeCode != "" && f.AfterCode != "")
}

// NormalizedConfidence returns the confidence clamped to [0.0, 1.0].
func (f Finding) NormalizedConfidence() float64 {
	return clampConfidence(f.Confidence)
}

// String returns a human-readable summary of the finding.
func (f Finding) String() string {
	return fmt.Sprintf("%s %s [%s] %s: %s",
		f.Severity, f.ToolName, f.Rule, f.Position, f.Message)
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
		f.Snippet != other.Snippet || !floatEq(f.Confidence, other.Confidence) {
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

	if f.Suppression.Kind != other.Suppression.Kind ||
		f.Suppression.Rule != other.Suppression.Rule ||
		f.Suppression.Reason != other.Suppression.Reason {
		return false
	}

	return equalTimePtr(f.Suppression.ExpiresAt, other.Suppression.ExpiresAt)
}

func equalTimePtr(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return a.Equal(*b)
}

// floatEq returns true if a and b are equal within a small epsilon.
func floatEq(a, b float64) bool {
	const epsilon = 1e-9

	return math.Abs(a-b) < epsilon
}
