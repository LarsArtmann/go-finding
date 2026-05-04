package finding

import (
	"testing"
)

func TestSeverity_LessThan_InvalidInput(t *testing.T) {
	t.Parallel()

	if Severity("unknown").LessThan(Severity("other")) {
		t.Error("invalid < invalid should be false")
	}
	if SeverityInfo.LessThan(Severity("unknown")) {
		t.Error("valid < invalid should be false")
	}
	if Severity("unknown").LessThan(SeverityInfo) {
		t.Error("invalid < valid should be false")
	}
}
