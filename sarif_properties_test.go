package finding

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
)

func TestSARIF_RoundTripPreservesBeforeCodeAndID(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:          "test:R1:a.go:1:1",
		Rule:        "R1",
		ToolName:    "test",
		Message:     "msg",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		BeforeCode:  "old code",
		AfterCode:   "new code",
		FixStrategy: FixStrategySuggest,
		Related: []RelatedRef{{
			FindingID: "related-123", Relation: "causes", Position: Pos("b.go", 5, 1),
		}},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings, err := FindingsFromSARIF(context.Background(), data)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))

	f := findings[0]

	g.Expect(f.ID).To(gomega.Equal(ID("test:R1:a.go:1:1")))
	g.Expect(f.Rule).To(gomega.Equal(RuleName("R1")))
	g.Expect(f.Message).To(gomega.Equal("msg"))
	g.Expect(f.AfterCode).To(gomega.Equal("new code"))

	g.Expect(f.BeforeCode).To(gomega.Equal("old code"))

	g.Expect(f.Related).To(gomega.HaveLen(1))
	g.Expect(f.Related[0].FindingID).To(gomega.Equal(ID("related-123")))
	g.Expect(f.Related[0].Relation).To(gomega.Equal(RelationKind("causes")))
}

func TestSARIF_TagsRoundTrip(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:       "test:R1:a.go:1:1",
		Rule:     "R1",
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityError,
		Position: Pos("a.go", 1, 1),
		Tags:     []Tag{TagSecurity, "injection", "xss"},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings, err := FindingsFromSARIF(context.Background(), data)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))

	g.Expect(findings[0].Tags).To(gomega.Equal([]Tag{TagSecurity, "injection", "xss"}))
}

func TestSARIF_SuppressedFindingsExcludedFromRoundTrip(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:          "test:R1:a.go:1:1",
		Rule:        "R1",
		ToolName:    "test",
		Message:     "msg",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R1", Reason: "won't fix"},
	})
	report.ComputeSummary()

	data, err := report.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings, err := FindingsFromSARIF(context.Background(), data)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.BeEmpty())
}

func TestSARIF_RoundTrip_EditProperties(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	report := NewReport(ToolInfo{Name: "test"})
	report.AddFinding(Finding{
		ID:       "test:R1:a.go:1:1",
		Rule:     "R1",
		ToolName: "test",
		Message:  "msg",
		Severity: SeverityError,
		Position: Pos("a.go", 1, 1),
		Metadata: map[string]string{
			sarifPropEditPrefix + "offset":      "42",
			sarifPropEditPrefix + "length":      "10",
			sarifPropEditPrefix + "replacement": "new code",
		},
	})

	data, err := report.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings, err := FindingsFromSARIF(context.Background(), data)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))

	f := findings[0]
	g.Expect(f.Metadata).NotTo(gomega.BeNil())
	g.Expect(f.Metadata[sarifPropEditPrefix+"offset"]).To(gomega.Equal("42"))
	g.Expect(f.Metadata[sarifPropEditPrefix+"length"]).To(gomega.Equal("10"))
	g.Expect(f.Metadata[sarifPropEditPrefix+"replacement"]).To(gomega.Equal("new code"))
}

func TestFindingFromSarResult_FixDescriptionWithoutReplacements(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := sarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: sarifMessage{Text: "msg"},
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
			},
		}},
		Fixes: []sarifFix{{
			Description: sarifMessage{Text: "fix it"},
			Changes:     []sarifArtifactChange{},
		}},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Suggestion).To(gomega.Equal("fix it"))
	g.Expect(f.AfterCode).To(gomega.BeEmpty())
	g.Expect(f.FixStrategy).To(gomega.Equal(FixStrategySuggest))
}

func TestFindingFromSarResult_GeneratesIDWithoutProperties(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := sarifResult{
		RuleID:  "R1",
		Level:   "warning",
		Message: sarifMessage{Text: "msg"},
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Region:           &sarifRegion{StartLine: 10, StartColumn: 5},
			},
		}},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.ID).NotTo(gomega.BeEmpty())
	g.Expect(string(f.ID)).To(gomega.ContainSubstring("tool"))
	g.Expect(string(f.ID)).To(gomega.ContainSubstring("R1"))
}

func TestFindingsFromSARIF_FixWithEmptyChanges(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	sarif := `{
		"version": "2.1.0",
		"runs": [{
			"tool": {"driver": {"name": "test"}},
			"results": [{
				"ruleId": "R1",
				"level": "warning",
				"message": {"text": "msg"},
				"locations": [{"physicalLocation": {"artifactLocation": {"uri": "a.go"}}}],
				"fixes": [{"description": {"text": "suggestion only"}, "artifactChanges": []}]
			}]
		}]
	}`

	findings, err := FindingsFromSARIF(context.Background(), []byte(sarif))
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))
	g.Expect(findings[0].Suggestion).To(gomega.Equal("suggestion only"))
	g.Expect(findings[0].AfterCode).To(gomega.BeEmpty())
}

func TestSARIF_SchemaCompliance(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := NewReport(ToolInfo{Name: "schema-test", Version: "1.0.0"})
	r.AddFinding(Finding{
		ID: "t:r:f.go:10:5", Rule: "R1", ToolName: "t", Message: "msg",
		Severity: SeverityError, Position: Position{File: "f.go", Line: 10, Column: 5},
		FixStrategy: FixStrategyDirect, Suggestion: "fix", BeforeCode: "old", AfterCode: "new",
		Category: CategorySecurity, Confidence: 0.9,
		Tags: []Tag{"sec"}, Snippet: "code",
		Range: NewRangePtr("f.go", 10, 5, 10, 10),
		Related: []RelatedRef{
			{FindingID: "other:1", Relation: "clone-of", Position: Position{File: "o.go", Line: 1}},
		},
		Metadata: map[string]string{"key": "val"},
	})
	r.ComputeSummary()

	data, err := r.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	var log map[string]any
	g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

	g.Expect(log["version"]).To(gomega.Equal("2.1.0"))
	g.Expect(log["$schema"]).To(gomega.ContainSubstring("sarif-schema-2.1.0"))
	g.Expect(log["runs"]).NotTo(gomega.BeEmpty())

	runs, ok := log["runs"].([]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(runs).To(gomega.HaveLen(1))

	run, ok := runs[0].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	tool, ok := run["tool"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	driver, ok := tool["driver"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(driver["name"]).To(gomega.Equal("schema-test"))
	g.Expect(driver["version"]).To(gomega.Equal("1.0.0"))

	results, ok := run["results"].([]any)
	g.Expect(ok).To(gomega.BeTrue())

	result, ok := results[0].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(result["ruleId"]).To(gomega.Equal("R1"))
	g.Expect(result["level"]).To(gomega.BeElementOf("none", "note", "warning", "error"))
	g.Expect(result["message"]).NotTo(gomega.BeEmpty())
	g.Expect(result["locations"]).NotTo(gomega.BeEmpty())
	g.Expect(result["properties"]).NotTo(gomega.BeEmpty())
	g.Expect(result["fixes"]).NotTo(gomega.BeEmpty())
	g.Expect(result["relatedLocations"]).NotTo(gomega.BeEmpty())
	g.Expect(result["rank"]).To(gomega.BeNumerically(">=", 0))
	g.Expect(result["rank"]).To(gomega.BeNumerically("<=", 100))

	locs, ok := result["locations"].([]any)
	g.Expect(ok).To(gomega.BeTrue())
	loc, ok := locs[0].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	pl, ok := loc["physicalLocation"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	al, ok := pl["artifactLocation"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(al["uri"]).To(gomega.Equal("f.go"))

	region, ok := pl["region"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(region["startLine"]).To(gomega.Equal(float64(10)))
	g.Expect(region["startColumn"]).To(gomega.Equal(float64(5)))
	g.Expect(region["endLine"]).To(gomega.Equal(float64(10)))
	g.Expect(region["endColumn"]).To(gomega.Equal(float64(10)))

	props, ok := result["properties"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(props[sarifPropID]).To(gomega.Equal("t:r:f.go:10:5"))
	g.Expect(props[sarifPropSeverity]).To(gomega.Equal("error"))
	g.Expect(props[sarifPropFixStrategy]).To(gomega.Equal("direct"))
	g.Expect(props[sarifPropToolName]).To(gomega.Equal("t"))
	g.Expect(props[sarifPropCategory]).To(gomega.Equal("security"))
	g.Expect(props[sarifPropMetaPrefix+"key"]).To(gomega.Equal("val"))
}
