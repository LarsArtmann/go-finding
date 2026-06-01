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
		ID:          id,
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
	g.Expect(parsed.Findings).To(HaveLen(1))
}

func TestFindingJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := standardTestFinding()

	data, err := json.Marshal(f)
	g.Expect(err).NotTo(HaveOccurred())

	parsed, err := FromJSON(data)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(parsed.ID).To(Equal(f.ID))
}

func TestSARIFConversion(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{
		ID:          "test:rule1:file.go:10:5",
		Rule:        "rule1",
		ToolName:    "test",
		Message:     "test message",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect,
		AfterCode:   "fixed",
		Category:    CategoryStyle,
		Metadata:    map[string]string{"key": "value"},
	})
	r.ComputeSummary()

	sarif, err := r.ToSARIF()
	g.Expect(err).NotTo(HaveOccurred())

	var log map[string]any
	g.Expect(json.Unmarshal(sarif, &log)).NotTo(HaveOccurred())
	g.Expect(log["version"]).To(Equal("2.1.0"))
}

func TestLSPConversion(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := standardTestFinding()

	lsp := f.ToLSP()
	g.Expect(lsp.Range.Start.Line).To(Equal(9))
	g.Expect(lsp.Severity).To(BeEquivalentTo(1))
}

func TestCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(CategorySecurity.IsValid()).To(BeTrue())
	g.Expect(Category("custom-category").IsValid()).To(BeTrue())
	g.Expect(Category("custom-category").IsStandard()).To(BeFalse())
}

func TestRangeContains(t *testing.T) {
	t.Parallel()

	r := NewRange("test.go", 10, 5, 20, 10)

	tests := []struct {
		name string
		pos  Position
		want bool
	}{
		{"in range", Position{File: "test.go", Line: 15, Column: 7}, true},
		{"before range", Position{File: "test.go", Line: 5, Column: 1}, false},
		{"different file", Position{File: "other.go", Line: 15, Column: 7}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(r.Contains(tt.pos)).To(Equal(tt.want))
		})
	}
}

func TestClampConfidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Confidence
		want Confidence
	}{
		{"negative clamps to 0", -0.5, 0},
		{"zero stays zero", 0, 0},
		{"one stays one", 1, 1},
		{"above one clamps to 1", 1.5, 1},
		{"mid value unchanged", 0.75, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(tt.in.Clamp()).To(BeNumerically("~", tt.want, 1e-9))
		})
	}
}

func TestFinding_HasCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(Finding{}.HasCategory()).To(BeFalse())
	g.Expect(Finding{Category: CategorySecurity}.HasCategory()).To(BeTrue())
}

func TestFinding_String_WithCategory(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		Severity: SeverityError,
		ToolName: "test",
		Rule:     "R1",
		Position: Position{File: "a.go", Line: 10},
		Message:  "something broke",
	}

	g.Expect(f.String()).To(Equal("error test [R1] a.go:10: something broke"))

	f.Category = CategorySecurity
	g.Expect(f.String()).To(Equal("error test [R1] a.go:10: something broke (security)"))
}

func TestFinding_Preview(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(Finding{}.Preview()).To(BeEmpty())

	f := Finding{BeforeCode: "old", AfterCode: "new"}
	g.Expect(f.Preview()).To(Equal("- old\n+ new\n"))

	insertOnly := Finding{AfterCode: "inserted"}
	g.Expect(insertOnly.Preview()).To(Equal("+ inserted\n"))

	deleteOnly := Finding{BeforeCode: "removed"}
	g.Expect(deleteOnly.Preview()).To(Equal("- removed\n"))
}

func TestFinding_HasRange(t *testing.T) {
	t.Parallel()

	t.Run("nil range", func(t *testing.T) {
		t.Parallel()

		f := Finding{Position: Position{File: "a.go", Line: 1}}
		if f.HasRange() {
			t.Error("HasRange() = true for nil range")
		}
	})

	t.Run("valid range", func(t *testing.T) {
		t.Parallel()

		f := Finding{
			Position: Position{File: "a.go", Line: 1},
			Range:    &Range{Start: Position{File: "a.go", Line: 1}, End: Position{Line: 5}},
		}
		if !f.HasRange() {
			t.Error("HasRange() = false for valid range")
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		t.Parallel()

		f := Finding{
			Position: Position{File: "a.go", Line: 1},
			Range:    &Range{},
		}
		if f.HasRange() {
			t.Error("HasRange() = true for invalid range")
		}
	})
}

func TestRelationKind_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     RelationKind
		wantStr  string
	}{
		{"clone-of", RelationCloneOf, "clone-of"},
		{"causes", RelationCauses, "causes"},
		{"wraps", RelationWraps, "wraps"},
		{"related", RelationRelated, "related"},
	}

	for _, tt := range tests {
		if got := string(tt.kind); got != tt.wantStr {
			t.Errorf("RelationKind %s = %q, want %q", tt.name, got, tt.wantStr)
		}
	}
}
