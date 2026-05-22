package finding

import (
	"errors"
	"testing"
)

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
		if tt.sev.IsValid() != tt.want {
			t.Errorf("IsValid(%q) = %v, want %v", tt.sev, !tt.want, tt.want)
		}
	}
}

func TestSeverity_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want Severity
	}{
		{SeverityInfo, SeverityInfo},
		{SeverityWarning, SeverityWarning},
		{SeverityError, SeverityError},
		{SeverityCritical, SeverityCritical},
	}

	for _, tt := range tests {
		if string(tt.sev) != string(tt.want) {
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
		{Severity("unknown"), SeverityInfo, false, "invalid < valid returns false"},
		{SeverityInfo, Severity("unknown"), false, "valid > invalid returns false"},
		{Severity("unknown"), Severity("unknown"), false, "invalid > invalid returns false"},
	}

	for _, tt := range tests {
		if tt.a.GreaterThan(tt.b) != tt.want {
			t.Errorf("GreaterThan: %s = %v, want %v", tt.label, !tt.want, tt.want)
		}
	}
}

func TestSeverity_LessThan(t *testing.T) {
	t.Parallel()

	if !SeverityInfo.LessThan(SeverityWarning) {
		t.Error("info < warning should be true")
	}
	if SeverityError.LessThan(SeverityWarning) {
		t.Error("error < warning should be false")
	}
	if SeverityWarning.LessThan(SeverityWarning) {
		t.Error("warning < warning should be false")
	}
	if Severity("unknown").LessThan(SeverityInfo) {
		t.Error("invalid < valid should be false")
	}
	if SeverityInfo.LessThan(Severity("unknown")) {
		t.Error("valid < invalid should be false")
	}
	if Severity("unknown").LessThan(Severity("other")) {
		t.Error("invalid < other invalid should be false")
	}
}

func TestSeverity_Ordering(t *testing.T) {
	t.Parallel()

	ordered := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}

	for i := range ordered {
		for j := range ordered {
			want := i > j
			if ordered[i].GreaterThan(ordered[j]) != want {
				t.Errorf("GreaterThan(%s, %s) = %v, want %v", ordered[i], ordered[j], !want, want)
			}
		}
	}
}

func TestSeverity_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b  Severity
		want  int
		label string
	}{
		{SeverityInfo, SeverityInfo, 0, "info == info"},
		{SeverityWarning, SeverityInfo, 1, "warning > info"},
		{SeverityInfo, SeverityWarning, -1, "info < warning"},
		{SeverityCritical, SeverityError, 1, "critical > error"},
		{SeverityError, SeverityCritical, -1, "error < critical"},
		{Severity("unknown"), SeverityInfo, -1, "invalid < valid"},
		{SeverityInfo, Severity("unknown"), 1, "valid > invalid"},
		{Severity("foo"), Severity("bar"), 1, "different invalids ordered lexicographically"},
		{
			Severity("bar"),
			Severity("foo"),
			-1,
			"different invalids ordered lexicographically reverse",
		},
		{Severity("foo"), Severity("foo"), 0, "same invalid == same invalid"},
	}

	for _, tt := range tests {
		if got := tt.a.Compare(tt.b); got != tt.want {
			t.Errorf("Compare %s: got %d, want %d", tt.label, got, tt.want)
		}
	}
}

func TestSeverity_CompareOp_DefaultCase(t *testing.T) {
	t.Parallel()

	result := SeverityInfo.compareOp(SeverityWarning, comparisonOp(99))
	if result {
		t.Error("invalid comparisonOp should return false")
	}
}

func TestParseSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Severity
		isErr bool
	}{
		{"info", SeverityInfo, false},
		{"warning", SeverityWarning, false},
		{"error", SeverityError, false},
		{"critical", SeverityCritical, false},
		{"unknown", "", true},
		{"", "", true},
		{"INFO", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := ParseSeverity(tt.input)
			if tt.isErr {
				if err == nil {
					t.Errorf("ParseSeverity(%q) = nil error, want error", tt.input)
				}

				if !errors.Is(err, errInvalidSeverity) {
					t.Errorf(
						"ParseSeverity(%q) error = %v, want wrapping errInvalidSeverity",
						tt.input, err,
					)
				}
			} else {
				if err != nil {
					t.Errorf("ParseSeverity(%q) = unexpected error: %v", tt.input, err)
				}

				if got != tt.want {
					t.Errorf("ParseSeverity(%q) = %q, want %q", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestMustParseSeverity(t *testing.T) {
	t.Parallel()

	sev := MustParseSeverity("warning")
	if sev != SeverityWarning {
		t.Errorf("MustParseSeverity(%q) = %q, want %q", "warning", sev, SeverityWarning)
	}
}

func TestMustParseSeverity_Panics(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("MustParseSeverity(%q) should panic", "invalid")
		}
	}()

	MustParseSeverity("invalid")
}
