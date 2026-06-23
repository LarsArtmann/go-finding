package finding

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestFindingFromSarResult_Rank(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := sarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: sarifMessage{Text: "msg"},
		Rank:    75.0,
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Region:           &sarifRegion{StartLine: 10},
			},
		}},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Confidence).To(gomega.BeNumerically("~", 0.75, 1e-9))
}

func TestFindingFromSarResult_FixesWithReplacements(t *testing.T) {
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
			Changes: []sarifArtifactChange{{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Replacements: []sarifReplacement{{
					DeletedRegion: sarifRegion{StartLine: 5, StartColumn: 1},
					InsertedText:  sarifMessage{Text: "fixed code"},
				}},
			}},
		}},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Suggestion).To(gomega.Equal("fix it"))
	g.Expect(f.AfterCode).To(gomega.Equal("fixed code"))
	g.Expect(f.FixStrategy).To(gomega.Equal(FixStrategyDirect))
}

func TestFindingFromSarResult_FixSuggestionOnly(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	r := sarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: sarifMessage{Text: "consider using slices.Contains"},
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
			},
		}},
		Fixes: []sarifFix{{
			Description: sarifMessage{Text: "use slices.Contains"},
		}},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Suggestion).To(gomega.Equal("use slices.Contains"))
	g.Expect(f.AfterCode).To(gomega.Equal(""))
	g.Expect(f.FixStrategy).To(gomega.Equal(FixStrategySuggest))
}

func TestFindingFromSarResult_RelatedLocations(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	r := sarifResult{
		RuleID:  "r1",
		Level:   "error",
		Message: sarifMessage{Text: "main finding"},
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "main.go"},
			},
		}},
		Related: []sarifRelatedLoc{
			{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: "helper.go"},
					Region:           &sarifRegion{StartLine: 20, StartColumn: 3},
				},
				Message: sarifMessage{Text: "related call"},
			},
			{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: "util.go"},
				},
				Message: sarifMessage{Text: "no region"},
			},
		},
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Related).To(gomega.HaveLen(2))

	if f.Related[0].Relation != RelationKind("related call") {
		t.Errorf("Related[0].Relation = %q, want %q", f.Related[0].Relation, "related call")
	}

	assertRelatedPosition(t, f.Related[0], "helper.go", 20)
	assertRelatedPosition(t, f.Related[1], "util.go", 0)
}

func TestFindingFromSarResult_NoLocations(t *testing.T) {
	t.Parallel()

	r := sarifResult{
		RuleID:  "r1",
		Level:   "warning",
		Message: sarifMessage{Text: "msg"},
	}

	f := findingFromSarResult(r, "tool")
	if f.Position.File != "" {
		t.Errorf("Position.File = %q, want empty (no locations)", f.Position.File)
	}
}

func TestFindingFromSarResult_Properties(t *testing.T) {
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
		Properties: goFindingPropsWithCustom(
			"test-id",
			"critical",
			"direct",
			"scanner",
			"security",
			[]any{"injection"},
			0.85,
			"fix it",
			"code here",
			"custom-key",
			"custom-val",
		),
	}

	f := findingFromSarResult(r, "default-tool")
	if f.ID != "test-id" {
		t.Errorf("ID = %q, want %q", f.ID, "test-id")
	}

	if f.Severity != SeverityCritical {
		t.Errorf("Severity = %v, want %v", f.Severity, SeverityCritical)
	}

	if f.FixStrategy != FixStrategyDirect {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, FixStrategyDirect)
	}

	g.Expect(f.ToolName).To(gomega.Equal(ToolName("scanner")))
	g.Expect(f.Category).To(gomega.Equal(Category("security")))
	g.Expect(f.Tags).To(gomega.Equal([]Tag{"injection"}))
	g.Expect(f.Confidence).To(gomega.BeNumerically("~", 0.85, 1e-9))
	g.Expect(f.Suggestion).To(gomega.Equal("fix it"))
	g.Expect(f.Snippet).To(gomega.Equal("code here"))
	g.Expect(f.Metadata["custom-key"]).To(gomega.Equal("custom-val"))
}

func TestApplySarifPosition_NilRegion(t *testing.T) {
	t.Parallel()

	r := sarifResult{
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Region:           nil,
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)
	assertFindingPosition(t, f, "a.go", 0)

	if f.Range != nil {
		t.Error("Range should be nil with nil region")
	}
}

func TestApplySarifPosition_WithEndPosition(t *testing.T) {
	t.Parallel()

	r := sarifResult{
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Region: &sarifRegion{
					StartLine:   10,
					StartColumn: 5,
					EndLine:     15,
					EndColumn:   20,
				},
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)

	if f.Range == nil {
		t.Fatal("Range should be set with end position")
	}

	if f.Range.End.Line != 15 {
		t.Errorf("Range.End.Line = %d, want 15", f.Range.End.Line)
	}

	if f.Range.End.Column != 20 {
		t.Errorf("Range.End.Column = %d, want 20", f.Range.End.Column)
	}
}

func TestApplySarifPosition_EndColumnOnly(t *testing.T) {
	t.Parallel()

	r := sarifResult{
		Locations: []sarifLocation{{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
				Region: &sarifRegion{
					StartLine:   10,
					StartColumn: 5,
					EndColumn:   20,
				},
			},
		}},
	}

	f := Finding{}
	applySarifPosition(&f, r)

	if f.Range == nil {
		t.Fatal("Range should be set with EndColumn > 0")
	}
}
