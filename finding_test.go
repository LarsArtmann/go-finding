package finding

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func standardTestFinding() Finding {
	return Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: benchFile, Line: 10, Column: 5, Offset: 0},
		Category:    "",
		FixStrategy: FixStrategyNone,
		Suggestion:  "",
		BeforeCode:  "",
		AfterCode:   "",
		Range:       nil,
		Snippet:     "",
		Confidence:  ConfidenceFull,
		Related:     []RelatedRef{},
		Suppression: nil,
		Metadata:    map[string]string{},
	}
}

func addFindingForTest(t *testing.T, r *Report, id string, sev Severity, file string) {
	t.Helper()

	r.AddFinding(Finding{
		ID:          ID(id),
		Severity:    sev,
		Position:    Position{File: file, Line: 0, Column: 0, Offset: 0},
		Category:    "",
		FixStrategy: FixStrategyNone,
		Suggestion:  "",
		BeforeCode:  "",
		AfterCode:   "",
		Range:       nil,
		Snippet:     "",
		Confidence:  ConfidenceFull,
		Related:     []RelatedRef{},
		Suppression: nil,
		Metadata:    map[string]string{},
	})
}

func TestSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		valid    bool
	}{
		{SeverityInfo.String(), SeverityInfo, true},
		{"warning", SeverityWarning, true},
		{SeverityError.String(), SeverityError, true},
		{SeverityCritical.String(), SeverityCritical, true},
		{"invalid", Severity("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(tt.severity.IsValid()).To(Equal(tt.valid))
		})
	}
}

func TestSeverityOrdering(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(SeverityInfo.GreaterThan(SeverityWarning)).To(BeFalse())
	g.Expect(SeverityError.GreaterThan(SeverityWarning)).To(BeTrue())
	g.Expect(SeverityCritical.GreaterThan(SeverityError)).To(BeTrue())
}

func TestFixStrategy(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(FixStrategyNone.IsValid()).To(BeTrue())
	g.Expect(FixStrategyDirect.CanAutoApply()).To(BeTrue())
	g.Expect(FixStrategySuggest.CanAutoApply()).To(BeFalse())
	g.Expect(FixStrategyAI.NeedsAI()).To(BeTrue())
}

func TestPosition(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	p := Position{File: "test.go", Line: 42, Column: 5}
	g.Expect(p.IsValid()).To(BeTrue())
	g.Expect(p.String()).To(Equal("test.go:42:5"))
}

func TestPositionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pos  Position
		want string
	}{
		{Position{File: "test.go"}, "test.go"},
		{Position{File: "test.go", Line: 42}, "test.go:42"},
		{Position{File: "test.go", Line: 42, Column: 5}, "test.go:42:5"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(tt.pos.String()).To(Equal(tt.want))
		})
	}
}

func TestFinding(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect,
		BeforeCode:  "old",
		AfterCode:   "new",
	}

	g.Expect(f.IsValid()).To(BeTrue())
	g.Expect(f.HasFix()).To(BeTrue())
	g.Expect(f.HasSuggestion()).To(BeTrue())
	g.Expect(f.IsSuppressed()).To(BeFalse())
}

func TestFindingSuppressed(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityWarning,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		Suppression: &Suppression{Kind: SuppressionInSource, Reason: "intentional"},
	}

	g.Expect(f.IsSuppressed()).To(BeTrue())
}

func TestReport(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := NewReport(ToolInfo{Name: "test-tool", Version: "1.0.0"})

	addFindingForTest(t, r, "1", SeverityError, "a.go")
	addFindingForTest(t, r, "2", SeverityWarning, "b.go")
	addFindingForTest(t, r, "3", SeverityError, "a.go")

	r.ComputeSummary()

	g.Expect(r.Summary.Total).To(Equal(3))
	g.Expect(r.Summary.FilesAffected).To(Equal(2))
	g.Expect(r.Summary.BySeverity[SeverityError]).To(Equal(2))

	errors := r.BySeverity(SeverityError)
	g.Expect(errors).To(HaveLen(2))
}

func TestReportJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{
		ID:       "test:rule1:file.go:10:5",
		Rule:     "rule1",
		ToolName: "test",
		Message:  "test message",
		Severity: SeverityError,
		Position: Position{File: "file.go", Line: 10, Column: 5},
	})
	r.ComputeSummary()

	data, err := json.Marshal(r)
	g.Expect(err).NotTo(HaveOccurred())

	parsed, _, err := ReportFromJSON(data)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(parsed.Tool.Name).To(Equal("test"))
	g.Expect(parsed.FindingsSnapshot()).To(HaveLen(1))
}
