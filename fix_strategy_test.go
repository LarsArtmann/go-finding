package finding

import "testing"

func TestFixStrategy_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fs   FixStrategy
		want bool
	}{
		{FixStrategyNone, true},
		{FixStrategySuggest, true},
		{FixStrategyDirect, true},
		{FixStrategyAI, true},
		{FixStrategy(""), false},
		{FixStrategy("unknown"), false},
		{FixStrategy("DIRECT"), false},
	}

	for _, tt := range tests {
		if got := tt.fs.IsValid(); got != tt.want {
			t.Errorf("FixStrategy(%q).IsValid() = %v, want %v", tt.fs, got, tt.want)
		}
	}
}

func TestFixStrategy_CanAutoApply(t *testing.T) {
	t.Parallel()

	if !FixStrategyDirect.CanAutoApply() {
		t.Error("FixStrategyDirect should be auto-applicable")
	}

	if FixStrategyNone.CanAutoApply() {
		t.Error("FixStrategyNone should not be auto-applicable")
	}

	if FixStrategySuggest.CanAutoApply() {
		t.Error("FixStrategySuggest should not be auto-applicable")
	}

	if FixStrategyAI.CanAutoApply() {
		t.Error("FixStrategyAI should not be auto-applicable")
	}
}

func TestFixStrategy_NeedsAI(t *testing.T) {
	t.Parallel()

	if !FixStrategyAI.NeedsAI() {
		t.Error("FixStrategyAI should need AI")
	}

	if FixStrategyDirect.NeedsAI() {
		t.Error("FixStrategyDirect should not need AI")
	}

	if FixStrategyNone.NeedsAI() {
		t.Error("FixStrategyNone should not need AI")
	}
}

func TestFixStrategy_Constants(t *testing.T) {
	t.Parallel()

	if FixStrategyNone != "none" {
		t.Errorf("FixStrategyNone = %q, want %q", FixStrategyNone, "none")
	}

	if FixStrategySuggest != "suggest" {
		t.Errorf("FixStrategySuggest = %q, want %q", FixStrategySuggest, "suggest")
	}

	if FixStrategyDirect != "direct" {
		t.Errorf("FixStrategyDirect = %q, want %q", FixStrategyDirect, "direct")
	}

	if FixStrategyAI != "ai" {
		t.Errorf("FixStrategyAI = %q, want %q", FixStrategyAI, "ai")
	}
}
