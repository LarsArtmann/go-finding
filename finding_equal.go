package finding

import (
	"maps"
	"slices"
	"time"
)

// Equal reports whether two findings are identical, including all nested fields.
func (f Finding) Equal(other Finding) bool {
	if f.ID != other.ID {
		return false
	}

	if f.Rule != other.Rule {
		return false
	}

	if f.ToolName != other.ToolName {
		return false
	}

	if f.Message != other.Message {
		return false
	}

	if f.Severity != other.Severity {
		return false
	}

	if !f.Position.Equal(other.Position) {
		return false
	}

	if f.Category != other.Category {
		return false
	}

	if !slices.Equal(f.Tags, other.Tags) {
		return false
	}

	if NormalizeFixStrategy(f.FixStrategy) != NormalizeFixStrategy(other.FixStrategy) {
		return false
	}

	if f.Suggestion != other.Suggestion {
		return false
	}

	if f.BeforeCode != other.BeforeCode {
		return false
	}

	if f.AfterCode != other.AfterCode {
		return false
	}

	if f.Snippet != other.Snippet {
		return false
	}

	if f.Confidence != other.Confidence {
		return false
	}

	if !f.equalRange(other) {
		return false
	}

	if !f.equalRelated(other) {
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

func (f Finding) equalRelated(other Finding) bool {
	if len(f.Related) != len(other.Related) {
		return false
	}

	for i, a := range f.Related {
		b := other.Related[i]
		if a.FindingID != b.FindingID || a.Relation != b.Relation || !a.Position.Equal(b.Position) {
			return false
		}

		if a.Range == nil && b.Range == nil {
			continue
		}

		if a.Range == nil || b.Range == nil {
			return false
		}

		if !a.Range.Equal(*b.Range) {
			return false
		}
	}

	return true
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
