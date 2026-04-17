package finding

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFindingIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{
			name: "valid finding",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 1},
			},
			want: true,
		},
		{
			name: "missing ID",
			f: Finding{
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 1},
			},
			want: false,
		},
		{
			name: "missing Rule",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 1},
			},
			want: false,
		},
		{
			name: "missing ToolName",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 1},
			},
			want: false,
		},
		{
			name: "missing Message",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Severity: SeverityError,
				Position: Position{File: "file.go", Line: 1},
			},
			want: false,
		},
		{
			name: "invalid position",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{},
			},
			want: false,
		},
		{
			name: "invalid severity",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: Severity("bogus"),
				Position: Position{File: "file.go", Line: 1},
			},
			want: false,
		},
		{
			name: "empty finding",
			f:    Finding{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.f.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindingHasFix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"direct", Finding{FixStrategy: FixStrategyDirect}, true},
		{"ai", Finding{FixStrategy: FixStrategyAI}, true},
		{"none", Finding{FixStrategy: FixStrategyNone}, false},
		{"suggest", Finding{FixStrategy: FixStrategySuggest}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.f.HasFix(); got != tt.want {
				t.Errorf("HasFix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindingHasSuggestion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"with suggestion text", Finding{Suggestion: "fix it"}, true},
		{"with before and after code", Finding{BeforeCode: "old", AfterCode: "new"}, true},
		{"with only before code", Finding{BeforeCode: "old"}, false},
		{"with only after code", Finding{AfterCode: "new"}, false},
		{"empty", Finding{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.f.HasSuggestion(); got != tt.want {
				t.Errorf("HasSuggestion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindingIsSuppressed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"no suppression", Finding{}, false},
		{"nil suppression", Finding{Suppression: nil}, false},
		{
			name: "active suppression",
			f:    Finding{Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R1"}},
			want: true,
		},
		{
			name: "expired suppression",
			f: Finding{
				Suppression: &Suppression{
					Kind:      SuppressionInSource,
					Rule:      "R1",
					ExpiresAt: new(time.Now().Add(-time.Hour)),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.f.IsSuppressed(); got != tt.want {
				t.Errorf("IsSuppressed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelatedRefIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    RelatedRef
		want bool
	}{
		{"with finding ID", RelatedRef{FindingID: "abc123"}, true},
		{"empty finding ID", RelatedRef{FindingID: ""}, false},
		{
			"with relation and position",
			RelatedRef{
				FindingID: "abc",
				Relation:  "clone-of",
				Position:  Position{File: "f.go", Line: 1},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuppressionIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *Suppression
		want bool
	}{
		{"valid in-source", &Suppression{Kind: SuppressionInSource, Rule: "SA1000"}, true},
		{"valid in-config", &Suppression{Kind: SuppressionInConfig, Rule: "unused"}, true},
		{"valid in-review", &Suppression{Kind: SuppressionInReview, Rule: "dup"}, true},
		{"missing kind", &Suppression{Rule: "SA1000"}, false},
		{"missing rule", &Suppression{Kind: SuppressionInSource}, false},
		{"nil suppression", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuppressionIsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *Suppression
		want bool
	}{
		{"no expiry", &Suppression{Kind: SuppressionInSource, Rule: "R1"}, false},
		{
			"future expiry",
			&Suppression{
				Kind:      SuppressionInSource,
				Rule:      "R1",
				ExpiresAt: new(time.Now().Add(time.Hour)),
			},
			false,
		},
		{
			name: "past expiry",
			s: &Suppression{
				Kind:      SuppressionInSource,
				Rule:      "R1",
				ExpiresAt: new(time.Now().Add(-time.Hour)),
			},
			want: true,
		},
		{"nil suppression", nil, false},
	}

	now := time.Now()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.s.IsExpired(now); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCategoryIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    ErrorCategory
		want bool
	}{
		{"validation", ErrCategoryValidation, true},
		{"io", ErrCategoryIO, true},
		{"parse", ErrCategoryParse, true},
		{"conflict", ErrCategoryConflict, true},
		{"internal", ErrCategoryInternal, true},
		{"empty", ErrorCategory(""), false},
		{"custom", ErrorCategory("custom"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.c.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindingErrorIsCategory(t *testing.T) {
	t.Parallel()

	err := NewValidationError("test", nil)
	if !IsCategory(err, ErrCategoryValidation) {
		t.Error("expected IsCategory to match validation")
	}

	if IsCategory(err, ErrCategoryIO) {
		t.Error("expected IsCategory to not match io")
	}
}

func TestRangeContainsByOffset(t *testing.T) {
	t.Parallel()

	r := Range{
		Start: Position{File: "test.go", Offset: 100},
		End:   Position{File: "test.go", Offset: 200},
	}

	tests := []struct {
		name string
		p    Position
		want bool
	}{
		{"within range", Position{File: "test.go", Offset: 150}, true},
		{"at start", Position{File: "test.go", Offset: 100}, true},
		{"at end", Position{File: "test.go", Offset: 200}, true},
		{"before range", Position{File: "test.go", Offset: 50}, false},
		{"after range", Position{File: "test.go", Offset: 250}, false},
		{"no offset on position", Position{File: "test.go", Line: 1}, false},
		{"no offset on range start", Position{File: "test.go", Offset: 150}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := r.containsByOffset(tt.p); got != tt.want {
				t.Errorf("containsByOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeContainsByOffsetZeroStart(t *testing.T) {
	t.Parallel()

	r := Range{
		Start: Position{File: "test.go", Line: 10},
		End:   Position{File: "test.go", Offset: 200},
	}

	p := Position{File: "test.go", Offset: 150}
	if r.containsByOffset(p) {
		t.Error("expected false when range start has no offset")
	}
}

func TestSARIFCriticalSeverityPreserved(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "tool:rule:file.go:1:1",
		Rule:        "rule",
		ToolName:    "tool",
		Message:     "critical issue",
		Severity:    SeverityCritical,
		Position:    Position{File: "file.go", Line: 1},
		FixStrategy: FixStrategyNone,
	}

	report := NewReport(ToolInfo{Name: "tool"})
	report.AddFinding(f)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF failed: %v", err)
	}

	t.Logf("SARIF output:\n%s", string(sarif))

	var log struct {
		Runs []struct {
			Results []struct {
				Level      string         `json:"level"`
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}

	if uErr := json.Unmarshal(sarif, &log); uErr != nil {
		t.Fatalf("unmarshal: %v", uErr)
	}

	if len(log.Runs) == 0 || len(log.Runs[0].Results) == 0 {
		t.Fatal("no results in SARIF")
	}

	result := log.Runs[0].Results[0]
	if result.Level != "error" {
		t.Errorf("SARIF level = %q, want %q", result.Level, "error")
	}

	severity, ok := result.Properties["go-finding/severity"]
	if !ok {
		t.Fatal("missing go-finding/severity property")
	}

	if severity != "critical" {
		t.Errorf("go-finding/severity = %v, want %q", severity, "critical")
	}
}

func TestCategoryString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryCorrectness, "correctness"},
		{CategoryStyle, "style"},
		{CategoryPerformance, "performance"},
		{CategorySecurity, "security"},
		{CategoryUnused, "unused"},
		{Category("custom"), "custom"},
	}
	for _, tt := range tests {
		if got := tt.cat.String(); got != tt.want {
			t.Errorf("Category(%q).String() = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestFixStrategyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fs   FixStrategy
		want string
	}{
		{FixStrategyNone, "none"},
		{FixStrategySuggest, "suggest"},
		{FixStrategyDirect, "direct"},
		{FixStrategyAI, "ai"},
		{FixStrategy("custom"), "custom"},
	}
	for _, tt := range tests {
		if got := tt.fs.String(); got != tt.want {
			t.Errorf("FixStrategy(%q).String() = %q, want %q", tt.fs, got, tt.want)
		}
	}
}
