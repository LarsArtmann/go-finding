package finding

import (
	"math"
	"testing"
)

const confTestCustom = "custom"

func TestConfidence_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    Confidence
		want bool
	}{
		{"none is valid", ConfidenceNone, true},
		{"full is valid", ConfidenceFull, true},
		{"medium is valid", ConfidenceMedium, true},
		{"zero value is valid", Confidence(0), true},
		{"fractional is valid", Confidence(0.42), true},
		{"negative is invalid", Confidence(-0.1), false},
		{"above 1 is invalid", Confidence(1.1), false},
		{"NaN is invalid", Confidence(math.NaN()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.c.IsValid(); got != tt.want {
				t.Errorf("Confidence(%v).IsValid() = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestConfidence_Clamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    Confidence
		want Confidence
	}{
		{"within range unchanged", Confidence(0.5), Confidence(0.5)},
		{"zero unchanged", Confidence(0), Confidence(0)},
		{"one unchanged", Confidence(1), Confidence(1)},
		{"negative clamped to zero", Confidence(-0.5), Confidence(0)},
		{"above one clamped to one", Confidence(1.5), Confidence(1)},
		{"large negative clamped to zero", Confidence(-100), Confidence(0)},
		{"large positive clamped to one", Confidence(100), Confidence(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.c.Clamp(); got != tt.want {
				t.Errorf("Confidence(%v).Clamp() = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestConfidence_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    Confidence
		want string
	}{
		{"none", ConfidenceNone, "0.00"},
		{"low", ConfidenceLow, "0.25"},
		{"medium", ConfidenceMedium, "0.50"},
		{"high", ConfidenceHigh, "0.75"},
		{"full", ConfidenceFull, "1.00"},
		{confTestCustom, Confidence(0.42), "0.42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.c.String(); got != tt.want {
				t.Errorf("Confidence(%v).String() = %q, want %q", tt.c, got, tt.want)
			}
		})
	}
}

func TestConfidence_StandardConstants(t *testing.T) {
	t.Parallel()

	if ConfidenceNone != 0.0 {
		t.Errorf("ConfidenceNone = %v, want 0.0", ConfidenceNone)
	}

	if ConfidenceLow != 0.25 {
		t.Errorf("ConfidenceLow = %v, want 0.25", ConfidenceLow)
	}

	if ConfidenceMedium != 0.5 {
		t.Errorf("ConfidenceMedium = %v, want 0.5", ConfidenceMedium)
	}

	if ConfidenceHigh != 0.75 {
		t.Errorf("ConfidenceHigh = %v, want 0.75", ConfidenceHigh)
	}

	if ConfidenceFull != 1.0 {
		t.Errorf("ConfidenceFull = %v, want 1.0", ConfidenceFull)
	}
}
