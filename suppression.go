package finding

import "time"

// SuppressionKind indicates where a suppression was defined.
type SuppressionKind string

// Suppression kinds indicate where a suppression was defined.
const (
	SuppressionInSource SuppressionKind = "in-source" // e.g., //nolint, //lint:ignore
	SuppressionInConfig SuppressionKind = "in-config" // Config file rules
	SuppressionInReview SuppressionKind = "in-review" // Accepted as false positive
)

// Suppression represents a suppressed finding.
type Suppression struct {
	Kind      SuppressionKind `json:"kind"`                // Where the suppression is defined
	Rule      string          `json:"rule"`                // Which rule is suppressed
	Reason    string          `json:"reason"`              // Why it's suppressed
	ExpiresAt *time.Time      `json:"expiresAt,omitempty"` // Optional expiry
}

// IsExpired returns true if the suppression has expired relative to now.
func (s *Suppression) IsExpired(now time.Time) bool {
	if s == nil || s.ExpiresAt == nil {
		return false
	}

	return now.After(*s.ExpiresAt)
}

// IsValid returns true if the suppression has a kind and rule.
func (s *Suppression) IsValid() bool {
	if s == nil {
		return false
	}

	return s.Kind != "" && s.Rule != ""
}

// IsValid returns true if the suppression kind is a recognized value.
func (k SuppressionKind) IsValid() bool {
	switch k {
	case SuppressionInSource, SuppressionInConfig, SuppressionInReview:
		return true
	}

	return false
}
