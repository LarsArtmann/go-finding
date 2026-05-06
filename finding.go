package finding

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
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
	Tags     []Tag    `json:"tags,omitempty"`     // Multiple tags for richer classification
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

// NewFinding creates a Finding with an auto-generated ID, default fix strategy,
// and clamped confidence.
func NewFinding(
	rule, toolName, message string,
	severity Severity,
	pos Position,
	confidence float64,
) Finding {
	return Finding{ //nolint:exhaustruct
		ID:          GenerateID(toolName, rule, pos),
		Rule:        rule,
		ToolName:    toolName,
		Message:     message,
		Severity:    severity,
		Position:    pos,
		FixStrategy: FixStrategyNone,
		Confidence:  clampConfidence(confidence),
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

	if len(f.Tags) > 0 {
		clone.Tags = make([]Tag, len(f.Tags))
		copy(clone.Tags, f.Tags)
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
	case FixStrategyDirect:
		return true
	case FixStrategySuggest, FixStrategyAI:
		return f.AfterCode != ""
	default:
		return false
	}
}

// IsAutoFixable returns true if this finding can be automatically applied
// by the pipeline. Unlike HasFix(), this also requires BeforeCode or AfterCode
// to be available for the FixEngine to produce byte-level edits.
func (f Finding) IsAutoFixable() bool {
	return f.FixStrategy == FixStrategyDirect && (f.BeforeCode != "" || f.AfterCode != "")
}

// HasSuggestion returns true if this finding has a human-readable suggestion.
func (f Finding) HasSuggestion() bool {
	return f.Suggestion != "" || (f.BeforeCode != "" && f.AfterCode != "")
}

// HasCategory returns true if this finding has a category set.
func (f Finding) HasCategory() bool {
	return f.Category != ""
}

// NormalizedConfidence returns the confidence clamped to [0.0, 1.0].
func (f Finding) NormalizedConfidence() float64 {
	return clampConfidence(f.Confidence)
}

// String returns a human-readable summary of the finding.
func (f Finding) String() string {
	var b strings.Builder
	b.WriteString(string(f.Severity))
	b.WriteByte(' ')
	b.WriteString(f.ToolName)
	b.WriteString(" [")
	b.WriteString(f.Rule)
	b.WriteString("] ")
	b.WriteString(f.Position.String())
	b.WriteString(": ")
	b.WriteString(f.Message)

	if f.Category != "" {
		b.WriteString(" (")
		b.WriteString(string(f.Category))
		b.WriteByte(')')
	}

	return b.String()
}

// Preview returns a unified-diff-style preview of the fix, or empty string if
// the finding has no fixable code change (BeforeCode and AfterCode both empty).
func (f Finding) Preview() string {
	if f.BeforeCode == "" && f.AfterCode == "" {
		return ""
	}

	var b strings.Builder

	if f.BeforeCode != "" {
		b.WriteString("- ")
		b.WriteString(f.BeforeCode)
		b.WriteByte('\n')
	}

	if f.AfterCode != "" {
		b.WriteString("+ ")
		b.WriteString(f.AfterCode)
		b.WriteByte('\n')
	}

	return b.String()
}

// Validate performs comprehensive validation and returns an error if the
// finding is invalid. It checks all fields that IsValid checks plus
// additional constraints: FixStrategy validity, Confidence range,
// and structural consistency.
func (f Finding) Validate() error {
	var errs []error

	if f.ID == "" {
		errs = append(errs, NewValidationError("finding.ID is required", nil))
	}

	if f.Rule == "" {
		errs = append(errs, NewValidationError("finding.Rule is required", nil))
	}

	if f.ToolName == "" {
		errs = append(errs, NewValidationError("finding.ToolName is required", nil))
	}

	if f.Message == "" {
		errs = append(errs, NewValidationError("finding.Message is required", nil))
	}

	if !f.Severity.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Severity %q is invalid", f.Severity), nil))
	}

	if !f.Position.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Position %+v is invalid", f.Position), nil))
	}

	if !f.FixStrategy.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.FixStrategy %q is invalid", f.FixStrategy), nil))
	}

	if f.Confidence < 0 || f.Confidence > 1 {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Confidence %f must be in [0.0, 1.0]", f.Confidence), nil))
	}

	if f.BeforeCode == "" && f.AfterCode != "" && f.FixStrategy == FixStrategyDirect {
		errs = append(errs, NewValidationError(
			"finding.FixStrategyDirect requires BeforeCode when AfterCode is set", nil))
	}

	return errors.Join(errs...)
}

// IsValid returns true if the finding has required fields set.
func (f Finding) IsValid() bool {
	return f.ID != "" && f.Rule != "" && f.ToolName != "" &&
		f.Message != "" && f.Position.IsValid() && f.Severity.IsValid()
}

// Key returns a stable identifier for the finding.
// If ID is set, it is returned; otherwise a deterministic key is built
// from ToolName, Position.File, Rule, and Message.
func (f Finding) Key() string {
	if f.ID != "" {
		return f.ID
	}

	return f.ToolName + "\x00" + f.Position.File + "\x00" + f.Rule + "\x00" + f.Message
}

// Equal reports whether two findings are identical, including all nested fields.
func (f Finding) Equal(other Finding) bool {
	if f.ID != other.ID || f.Rule != other.Rule || f.ToolName != other.ToolName ||
		f.Message != other.Message || f.Severity != other.Severity ||
		!f.Position.Equal(other.Position) ||
		f.Category != other.Category ||
		!slices.Equal(f.Tags, other.Tags) ||
		f.FixStrategy != other.FixStrategy ||
		f.Suggestion != other.Suggestion ||
		f.BeforeCode != other.BeforeCode || f.AfterCode != other.AfterCode ||
		f.Snippet != other.Snippet || !floatEq(f.Confidence, other.Confidence) {
		return false
	}

	if !f.equalRange(other) {
		return false
	}

	if !slices.Equal(f.Related, other.Related) {
		return false
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
