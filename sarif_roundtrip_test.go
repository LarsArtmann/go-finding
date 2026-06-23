package finding

import (
	"context"
	"encoding/json"
	"errors"
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

	err := r.WriteSARIFFiltered(context.Background(), &failWriter{}, SeverityWarning)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("encoding SARIF")))
}

func TestWriteTo(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		Findings: []Finding{
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

	err := r.WriteSARIFFiltered(ctx, &buf, SeverityWarning)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("writing SARIF filtered"))
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
