package finding

import (
	"testing"
	"time"
)

const suppTestIntentional = "intentional"

func TestSuppression_IsExpired_NilSuppression(t *testing.T) {
	t.Parallel()

	now := time.Now()

	var s *Suppression
	if s.IsExpired(now) {
		t.Error("nil suppression should not be expired")
	}
}

func TestSuppression_IsExpired_NilExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Now()

	s := &Suppression{Kind: SuppressionInSource, Reason: suppTestIntentional}
	if s.IsExpired(now) {
		t.Error("suppression without ExpiresAt should not be expired")
	}
}

func TestSuppression_IsExpired_PastTime(t *testing.T) {
	t.Parallel()

	now := time.Now()
	past := now.Add(-1 * time.Hour)

	s := &Suppression{Kind: SuppressionInSource, ExpiresAt: &past}
	if !s.IsExpired(now) {
		t.Error("suppression with past expiry should be expired")
	}
}

func TestSuppression_IsExpired_FutureTime(t *testing.T) {
	t.Parallel()

	now := time.Now()
	future := now.Add(24 * time.Hour)

	s := &Suppression{Kind: SuppressionInSource, ExpiresAt: &future}
	if s.IsExpired(now) {
		t.Error("suppression with future expiry should not be expired")
	}
}

func TestSuppressionKind_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind SuppressionKind
		want string
	}{
		{SuppressionInSource, "in-source"},
		{SuppressionInConfig, "in-config"},
		{SuppressionInReview, "in-review"},
	}

	for _, tt := range tests {
		if string(tt.kind) != tt.want {
			t.Errorf("SuppressionKind = %q, want %q", tt.kind, tt.want)
		}
	}
}

func TestSuppression_IsActive(t *testing.T) {
	t.Parallel()

	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name string
		s    *Suppression
		now  time.Time
		want bool
	}{
		{"nil suppression", nil, now, false},
		{
			"valid and not expired",
			&Suppression{Kind: SuppressionInSource, Rule: benchRule},
			now,
			true,
		},
		{
			"valid with future expiry",
			&Suppression{Kind: SuppressionInSource, Rule: benchRule, ExpiresAt: &future},
			now,
			true,
		},
		{
			"valid but expired",
			&Suppression{Kind: SuppressionInSource, Rule: benchRule, ExpiresAt: &past},
			now,
			false,
		},
		{"invalid missing kind", &Suppression{Rule: benchRule}, now, false},
		{"invalid missing rule", &Suppression{Kind: SuppressionInSource}, now, false},
		{"invalid empty", &Suppression{}, now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.s.IsActive(tt.now); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuppression_Fields(t *testing.T) {
	t.Parallel()

	s := Suppression{
		Kind:   SuppressionInConfig,
		Rule:   benchRule,
		Reason: "accepted false positive",
	}

	if s.Kind != SuppressionInConfig {
		t.Errorf("Kind = %q, want %q", s.Kind, SuppressionInConfig)
	}

	if s.Rule != benchRule {
		t.Errorf("Rule = %q, want %q", s.Rule, benchRule)
	}

	if s.Reason != "accepted false positive" {
		t.Errorf("Reason = %q, want %q", s.Reason, "accepted false positive")
	}
}
