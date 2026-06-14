package finding

import (
	"testing"

	. "github.com/onsi/gomega"
)

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
		{"custom", ErrorCategory(confTestCustom), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.c.IsValid()).To(Equal(tt.want))
		})
	}
}

func TestFindingErrorIsCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := NewValidationError("test", nil)
	g.Expect(IsCategory(err, ErrCategoryValidation)).To(BeTrue())
	g.Expect(IsCategory(err, ErrCategoryIO)).To(BeFalse())
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
			g := NewWithT(t)

			g.Expect(r.containsByOffset(tt.p)).To(Equal(tt.want))
		})
	}
}

func TestRangeContainsByOffsetZeroStart(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := Range{
		Start: Position{File: "test.go", Line: 10, Offset: -1},
		End:   Position{File: "test.go", Offset: 200},
	}

	p := Position{File: "test.go", Offset: 150}
	g.Expect(r.containsByOffset(p)).To(BeFalse())
}

func TestSARIFCriticalSeverityPreserved(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	g.Expect(err).NotTo(HaveOccurred())

	t.Logf("SARIF output:\n%s", string(sarif))

	var log struct {
		Runs []struct {
			Results []struct {
				Level      string         `json:"level"`
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}

	unmarshalJSON(t, sarif, &log)

	g.Expect(log.Runs).NotTo(BeEmpty())
	g.Expect(log.Runs[0].Results).NotTo(BeEmpty())

	result := log.Runs[0].Results[0]
	g.Expect(result.Level).To(Equal("error"))

	severity, ok := result.Properties[sarifPropSeverity]
	g.Expect(ok).To(BeTrue())
	g.Expect(severity).To(Equal("critical"))
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
