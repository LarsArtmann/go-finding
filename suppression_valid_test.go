package finding

import (
	"testing"
)

func TestSuppression_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    Suppression
		want bool
	}{
		{"valid source", Suppression{Kind: SuppressionInSource, Rule: "R1"}, true},
		{"valid config", Suppression{Kind: SuppressionInConfig, Rule: "R1"}, true},
		{"valid review", Suppression{Kind: SuppressionInReview, Rule: "R1"}, true},
		{"empty kind", Suppression{Rule: "R1"}, false},
		{"empty rule", Suppression{Kind: SuppressionInSource}, false},
		{"nonempty kind+rule accepted", Suppression{Kind: "bogus", Rule: "R1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.s.IsValid() != tt.want {
				t.Errorf("IsValid() = %v, want %v", !tt.want, tt.want)
			}
		})
	}
}
