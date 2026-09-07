package finding

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestWriteSARIF_WriterError(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	err := r.WriteSARIF(context.Background(), &failWriter{})
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("encoding SARIF")))
}

func TestWriteSARIFFiltered_WriterError(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	err := r.WriteSARIFWithOpts(context.Background(), &failWriter{}, WithMinSeverity(SeverityWarning))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("encoding SARIF")))
}

func TestWriteTo(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go", Line: 1},
			},
		},
	}

	var buf strings.Builder

	n, err := r.WriteTo(&buf)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(n).To(gomega.BeNumerically(">", 0))

	var log map[string]any
	g.Expect(json.Unmarshal([]byte(buf.String()), &log)).NotTo(gomega.HaveOccurred())
	g.Expect(log["version"]).To(gomega.Equal("2.1.0"))
}

func TestWriteTo_WriterError(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	n, err := r.WriteTo(&failWriter{})
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(n).To(gomega.BeNumerically("==", 0))
}

func TestWriteSARIF_CancelledContext(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf strings.Builder

	err := r.WriteSARIF(ctx, &buf)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("writing SARIF"))
}

func TestWriteSARIFFiltered_CancelledContext(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf strings.Builder

	err := r.WriteSARIFWithOpts(ctx, &buf, WithMinSeverity(SeverityWarning))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("writing SARIF"))
}

func TestFindingsFromSARIF_CancelledContext(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FindingsFromSARIF(ctx, []byte(`{"version":"2.1.0","runs":[]}`))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("reading SARIF"))
}

func TestFindingsFromReader(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	sarif := `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"govet"}},"results":[{"ruleId":"printf","level":"error","message":{"text":"invalid format"}}]}]}`

	findings, err := FindingsFromReader(context.Background(), strings.NewReader(sarif))
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))
	g.Expect(findings[0].Rule).To(gomega.Equal(RuleName("printf")))
	g.Expect(findings[0].ToolName).To(gomega.Equal(ToolName("govet")))
}

func TestFindingsFromReader_CancelledContext(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FindingsFromReader(ctx, strings.NewReader(`{"version":"2.1.0","runs":[]}`))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("reading SARIF"))
}

func TestFindingsFromReader_InvalidJSON(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	_, err := FindingsFromReader(context.Background(), strings.NewReader(`not json`))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("decoding SARIF"))
}

// failWriter is an io.Writer that always returns an error.
type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestToSARIF_RoundTripProperties(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:          "govet:printf:main.go:10:5",
		Rule:        "printf",
		ToolName:    "govet",
		Message:     "invalid format",
		Severity:    SeverityCritical,
		Position:    Position{File: "main.go", Line: 10, Column: 5},
		Category:    CategoryCorrectness,
		Tags:        []Tag{"printf"},
		FixStrategy: FixStrategySuggest,
		Confidence:  0.9,
		Suggestion:  "fix format string",
		Snippet:     "fmt.Sprintf(\"%d\")",
		Metadata:    map[string]string{"custom": "value"},
	}

	report := NewReport(ToolInfo{Name: "govet"})
	report.AddFinding(f)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	var log struct {
		Runs []struct {
			Results []struct {
				Properties map[string]any `json:"properties"`
			} `json:"results"`
		} `json:"runs"`
	}
	unmarshalJSON(t, sarif, &log)

	props := log.Runs[0].Results[0].Properties

	checks := goFindingPropsWithCustom(
		"govet:printf:main.go:10:5",
		"critical",
		"suggest",
		"govet",
		"correctness",
		[]any{"printf"},
		0.9,
		"fix format string",
		`fmt.Sprintf("%d")`,
		sarifPropMetaPrefix+"custom",
		"value",
	)

	for key, want := range checks {
		got, ok := props[key]
		if !ok {
			t.Errorf("missing property %q", key)

			continue
		}
		// JSON numbers unmarshal as float64
		if f64, ok := want.(float64); ok {
			if gotFloat, ok := got.(float64); !ok || gotFloat != f64 {
				t.Errorf("properties[%q] = %v, want %v", key, got, want)
			}
		} else if s, ok := want.([]any); ok {
			gotSlice, ok := got.([]any)
			if !ok || len(gotSlice) != len(s) {
				t.Errorf("properties[%q] = %v, want %v", key, got, want)

				continue
			}

			for i := range s {
				if gotSlice[i] != s[i] {
					t.Errorf("properties[%q][%d] = %v, want %v", key, i, gotSlice[i], s[i])
				}
			}
		} else if got != want {
			t.Errorf("properties[%q] = %v, want %v", key, got, want)
		}
	}
}

func TestFindingFromSarResult_WithFix(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := sarifResult{
		RuleID:  "SA1000",
		Level:   "warning",
		Message: sarifMessage{Text: "unused variable"},
		Locations: []sarifLocation{
			{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: "main.go"},
					Region:           &sarifRegion{StartLine: 10, StartColumn: 5},
				},
			},
		},
		Fixes: []sarifFix{
			{
				Description: sarifMessage{Text: "remove unused variable"},
				Changes: []sarifArtifactChange{
					{
						ArtifactLocation: sarifArtifactLocation{URI: "main.go"},
						Replacements: []sarifReplacement{
							{
								DeletedRegion: sarifRegion{
									StartLine: 10, StartColumn: 5, EndLine: 10, EndColumn: 15,
								},
								InsertedText: sarifMessage{
									Text: "fmt.Println()",
								},
							},
						},
					},
				},
			},
		},
	}

	f := findingFromSarResult(r, "staticcheck")

	g.Expect(f.Rule).To(gomega.Equal(RuleName("SA1000")))
	g.Expect(f.ToolName).To(gomega.Equal(ToolName("staticcheck")))
	g.Expect(f.Message).To(gomega.Equal("unused variable"))
	g.Expect(f.Severity).To(gomega.Equal(SeverityWarning))
	g.Expect(f.Suggestion).To(gomega.Equal("remove unused variable"))
	g.Expect(f.AfterCode).To(gomega.Equal("fmt.Println()"))
	g.Expect(f.FixStrategy).To(gomega.Equal(FixStrategyDirect))
}

func TestFindingFromSarResult_RankAsConfidence(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := sarifResult{
		RuleID:  "R1",
		Level:   "error",
		Message: sarifMessage{Text: "msg"},
		Locations: []sarifLocation{
			{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: "a.go"},
					Region:           &sarifRegion{StartLine: 1},
				},
			},
		},
		Rank: 75.0,
	}

	f := findingFromSarResult(r, "tool")
	g.Expect(f.Confidence).To(gomega.BeNumerically("~", 0.75, 0.01))
}

// TestSARIFRoundTrip_PositionOffset verifies that Position.Offset survives
// the SARIF export → import cycle, including the edge cases of offset=0
// (valid, means byte 0) and offset=-1 (sentinel, means unset/not exported).
func TestSARIFRoundTrip_PositionOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		startOffset int
		endOffset   int
		hasRange    bool
	}{
		{"offset_zero", 0, -1, false},
		{"offset_positive", 42, -1, false},
		{"offset_large", 99999, -1, false},
		{"range_both_offsets", 10, 50, true},
		{"range_start_only", 10, -1, true},
		{"unset_offset", -1, -1, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			f := Finding{
				ID:       ID("test:offset:" + tc.name),
				Rule:     "offset-test",
				ToolName: "test",
				Message:  "test",
				Severity: SeverityWarning,
				Position: Position{
					File:   "test.go",
					Line:   1,
					Column: 1,
					Offset: tc.startOffset,
				},
			}

			if tc.hasRange {
				f.Range = &Range{
					Start: f.Position,
					End: Position{
						File:   "test.go",
						Line:   5,
						Column: 10,
						Offset: tc.endOffset,
					},
				}
			}

			report := NewReport(ToolInfo{Name: "test"})
			report.AddFinding(f)

			sarifBytes, err := report.ToSARIF()
			g.Expect(err).NotTo(gomega.HaveOccurred())

			findings, err := FindingsFromSARIF(context.Background(), sarifBytes)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(findings).To(gomega.HaveLen(1))

			got := findings[0]

			if tc.startOffset >= 0 {
				g.Expect(got.Position.Offset).To(gomega.Equal(tc.startOffset),
					"Position.Offset not preserved")
			} else {
				g.Expect(got.Position.HasOffset()).To(gomega.BeFalse(),
					"unset offset should not be exported/imported")
			}

			if tc.hasRange && tc.endOffset >= 0 {
				g.Expect(got.Range).NotTo(gomega.BeNil())
				g.Expect(got.Range.End.Offset).To(gomega.Equal(tc.endOffset),
					"Range.End.Offset not preserved")
			}
		})
	}
}

// TestSARIFRoundTrip_GroupID verifies that GroupID survives a SARIF
// export/import round-trip via the property bag, and that findings without
// a GroupID do not gain one.
func TestSARIFRoundTrip_GroupID(t *testing.T) {
	g := NewParallelGomega(t)

	f := Finding{
		ID:       ID("art-dupl:clone-detected:a.go:10:1"),
		Rule:     "clone-detected",
		ToolName: "art-dupl",
		Message:  "duplicate code",
		Severity: SeverityWarning,
		Position: Position{File: "a.go", Line: 10, Column: 1},
		GroupID:  "clone-group-1",
	}

	report := NewReport(ToolInfo{Name: "art-dupl"})
	report.AddFinding(f)

	sarifBytes, err := report.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings, err := FindingsFromSARIF(context.Background(), sarifBytes)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))
	g.Expect(findings[0].GroupID).To(gomega.Equal(GroupID("clone-group-1")))

	noGroup := Finding{
		ID:       ID("govet:printf:b.go:1:1"),
		Rule:     "printf",
		ToolName: "govet",
		Message:  "no group",
		Severity: SeverityWarning,
		Position: Position{File: "b.go", Line: 1, Column: 1},
	}

	report2 := NewReport(ToolInfo{Name: "govet"})
	report2.AddFinding(noGroup)

	sarifBytes2, err := report2.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	findings2, err := FindingsFromSARIF(context.Background(), sarifBytes2)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings2).To(gomega.HaveLen(1))
	g.Expect(findings2[0].GroupID).To(gomega.Equal(GroupID("")))
}

// TestSARIFSnippet_BackwardCompat verifies that SARIF region.snippet accepts
// both the spec-compliant object form ({"text":"..."}) and the common
// bare-string shorthand used by many SARIF producers.
func TestSARIFSnippet_BackwardCompat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		snippet  string
		wantText string
	}{
		{
			name:     "object form",
			snippet:  `{"text":"code here"}`,
			wantText: "code here",
		},
		{
			name:     "bare string",
			snippet:  `"bare string snippet"`,
			wantText: "bare string snippet",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			sarif := fmt.Sprintf(`{
				"version": "2.1.0",
				"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
				"runs": [{
					"tool": {"driver": {"name": "test"}},
					"results": [{
						"ruleId": "r1",
						"level": "warning",
						"message": {"text": "msg"},
						"locations": [{
							"physicalLocation": {
								"artifactLocation": {"uri": "a.go"},
								"region": {
									"startLine": 1,
									"snippet": %s
								}
							}
						}]
					}]
				}]
			}`, tc.snippet)

			findings, err := FindingsFromSARIF(context.Background(), []byte(sarif))
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(findings).To(gomega.HaveLen(1))
			g.Expect(findings[0].Snippet).To(gomega.Equal(tc.wantText))
		})
	}
}
