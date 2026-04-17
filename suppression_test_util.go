package finding

import (
	"testing"
	"time"
)

func expiringSuppression(t *testing.T, reason string) *Suppression {
	t.Helper()

	return &Suppression{
		Kind:   SuppressionInSource,
		Reason: reason,
		ExpiresAt: func() *time.Time {
			t := time.Now().Add(1 * time.Hour)

			return &t
		}(),
	}
}
