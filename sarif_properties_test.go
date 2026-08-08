package finding

import (
	"context"
	"encoding/json/v2"
	"testing"

	"github.com/onsi/gomega"
)

func TestSARIF_RoundTripPreservesBeforeCodeAndID(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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

func TestSARIF_SchemaCompliance_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("multiple_findings_produce_multiple_results", func(t *testing.T) {
		g := gomega.NewWithT(t)

		r := NewReport(ToolInfo{Name: "multi"})
		r.AddFinding(Finding{
			ID: "a:r:f.go:1:1", Rule: "R1", ToolName: "t", Message: "first",
			Severity: SeverityWarning, Position: Pos("f.go", 1, 1),
		})
		r.AddFinding(Finding{
			ID: "b:r:f.go:2:1", Rule: "R2", ToolName: "t", Message: "second",
			Severity: SeverityError, Position: Pos("f.go", 2, 1),
		})
		r.ComputeSummary()

		data, err := r.ToSARIF()
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var log map[string]any
		g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

		runs := log["runs"].([]any)
		run := runs[0].(map[string]any)
		results := run["results"].([]any)
		g.Expect(results).To(gomega.HaveLen(2))
	})

	t.Run("file_level_position_omits_region", func(t *testing.T) {
		g := gomega.NewWithT(t)

		r := NewReport(ToolInfo{Name: "file-level"})
		r.AddFinding(Finding{
			ID: "x:r:config.yaml:0:0", Rule: "R1", ToolName: "t", Message: "config issue",
			Severity: SeverityInfo, Position: FilePos("config.yaml"),
		})
		r.ComputeSummary()

		data, err := r.ToSARIF()
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var log map[string]any
		g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

		runs := log["runs"].([]any)
		run := runs[0].(map[string]any)
		results := run["results"].([]any)
		result := results[0].(map[string]any)

		locs := result["locations"].([]any)
		loc := locs[0].(map[string]any)
		pl := loc["physicalLocation"].(map[string]any)
		al := pl["artifactLocation"].(map[string]any)
		g.Expect(al["uri"]).To(gomega.Equal("config.yaml"))

		_, hasRegion := pl["region"]
		if hasRegion {
			region := pl["region"].(map[string]any)
			_, hasStartLine := region["startLine"]
			g.Expect(hasStartLine).To(gomega.BeFalse(),
				"file-level position region should not have startLine (Line=0 omitted by omitempty)")
		}
	})

	t.Run("minimal_finding_omits_optional_sections", func(t *testing.T) {
		g := gomega.NewWithT(t)

		r := NewReport(ToolInfo{Name: "minimal"})
		r.AddFinding(Finding{
			ID: "m:r:f.go:1:1", Rule: "R1", ToolName: "t", Message: "minimal",
			Severity: SeverityInfo, Position: Pos("f.go", 1, 1),
		})
		r.ComputeSummary()

		data, err := r.ToSARIF()
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var log map[string]any
		g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

		runs := log["runs"].([]any)
		run := runs[0].(map[string]any)
		results := run["results"].([]any)
		result := results[0].(map[string]any)

		_, hasFixes := result["fixes"]
		g.Expect(hasFixes).To(gomega.BeFalse(), "minimal finding should not emit fixes")

		_, hasRelated := result["relatedLocations"]
		g.Expect(hasRelated).To(gomega.BeFalse(), "minimal finding should not emit relatedLocations")
	})

	t.Run("empty_report_produces_valid_sarif", func(t *testing.T) {
		g := gomega.NewWithT(t)

		r := NewReport(ToolInfo{Name: "empty", Version: "0.0.0"})
		r.ComputeSummary()

		data, err := r.ToSARIF()
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var log map[string]any
		g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

		g.Expect(log["version"]).To(gomega.Equal("2.1.0"))
		g.Expect(log["$schema"]).To(gomega.ContainSubstring("sarif-schema-2.1.0"))

		runs := log["runs"].([]any)
		g.Expect(runs).To(gomega.HaveLen(1))
		run := runs[0].(map[string]any)

		tool := run["tool"].(map[string]any)
		driver := tool["driver"].(map[string]any)
		g.Expect(driver["name"]).To(gomega.Equal("empty"))

		results, hasResults := run["results"]
		if hasResults {
			g.Expect(results.([]any)).To(gomega.BeEmpty())
		}
	})

	t.Run("suppressed_finding_emits_suppressions_array", func(t *testing.T) {
		g := gomega.NewWithT(t)

		r := NewReport(ToolInfo{Name: "suppressed"})
		r.AddFinding(Finding{
			ID: "s:r:f.go:1:1", Rule: "R1", ToolName: "t", Message: "suppressed issue",
			Severity:   SeverityWarning,
			Position:   Pos("f.go", 1, 1),
			Suppression: &Suppression{Kind: SuppressionInSource, Rule: "nolint", Reason: "intentional"},
		})
		r.ComputeSummary()

		data, err := r.ToSARIFWithOpts(WithIncludeSuppressed())
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var log map[string]any
		g.Expect(json.Unmarshal(data, &log)).NotTo(gomega.HaveOccurred())

		runs := log["runs"].([]any)
		run := runs[0].(map[string]any)
		results := run["results"].([]any)
		g.Expect(results).To(gomega.HaveLen(1))
		result := results[0].(map[string]any)

		suppressions, ok := result["suppressions"].([]any)
		g.Expect(ok).To(gomega.BeTrue(), "suppressed finding should emit suppressions array")
		g.Expect(suppressions).To(gomega.HaveLen(1))
		sup := suppressions[0].(map[string]any)
		g.Expect(sup["kind"]).To(gomega.Equal("inSource"))
	})
}
