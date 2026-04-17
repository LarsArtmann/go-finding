package finding

import "testing"

func TestSeverity_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want bool
	}{
		{SeverityInfo, true},
		{SeverityWarning, true},
		{SeverityError, true},
		{SeverityCritical, true},
		{Severity(""), false},
		{Severity("unknown"), false},
		{Severity("Info"), false},
		{Severity("WARNING"), false},
	}

	for _, tt := range tests {
		if got := tt.sev.IsValid(); got != tt.want {
			t.Errorf("Severity(%q).IsValid() = %v, want %v", tt.sev, got, tt.want)
		}
	}
}

func TestSeverity_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
	}

	for _, tt := range tests {
		if string(tt.sev) != tt.want {
			t.Errorf("Severity constant = %q, want %q", tt.sev, tt.want)
		}
	}
}

func TestSeverity_GreaterThan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  bool
		label string
	}{
		{SeverityCritical, SeverityError, true, "critical > error"},
		{SeverityError, SeverityWarning, true, "error > warning"},
		{SeverityWarning, SeverityInfo, true, "warning > info"},
		{SeverityInfo, SeverityInfo, false, "info == info"},
		{SeverityInfo, SeverityWarning, false, "info < warning"},
		{SeverityWarning, SeverityError, false, "warning < error"},
		{Severity("unknown"), SeverityInfo, false, "unknown < info (rank -1)"},
		{SeverityInfo, Severity("unknown"), true, "info > unknown (rank -1)"},
	}

	for _, tt := range tests {
		if got := tt.a.GreaterThan(tt.b); got != tt.want {
			t.Errorf(
				"Severity(%q).GreaterThan(%q) = %v, want %v (%s)",
				tt.a,
				tt.b,
				got,
				tt.want,
				tt.label,
			)
		}
	}
}

func TestSeverity_LessThan(t *testing.T) {
	t.Parallel()

	if !SeverityInfo.LessThan(SeverityWarning) {
		t.Error("info should be less than warning")
	}

	if SeverityError.LessThan(SeverityWarning) {
		t.Error("error should not be less than warning")
	}

	if SeverityWarning.LessThan(SeverityWarning) {
		t.Error("warning should not be less than itself")
	}
}

func TestSeverity_Ordering(t *testing.T) {
	t.Parallel()

	ordered := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}

	for i := range ordered {
		for j := range ordered {
			got := ordered[i].GreaterThan(ordered[j])
			want := i > j
			if got != want {
				t.Errorf(
					"Severity(%q).GreaterThan(%q) = %v, want %v",
					ordered[i],
					ordered[j],
					got,
					want,
				)
			}
		}
	}
}
