package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverity_LessThan_InvalidInput(t *testing.T) {
	t.Parallel()

	// Both invalid: neither is less than the other.
	assert.False(t, Severity("unknown").LessThan(Severity("other")), "invalid < invalid")
	assert.False(t, SeverityInfo.LessThan(Severity("unknown")), "valid < invalid")
	assert.False(t, Severity("unknown").LessThan(SeverityInfo), "invalid < valid")
}
