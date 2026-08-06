package finding

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestToSARIF_WithRelated(t *testing.T) {
	g := NewParallelGomega(t)

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID:       "f1",
				Rule:     "r1",
				Severity: SeverityError,
				Position: Position{File: "a.go", Line: 10},
				Related: []RelatedRef{
					{
						FindingID: "f2", Relation: "causes",
						Position: Position{File: "b.go", Line: 20},
					},
				},
			},
		},
	}

	data, err := r.ToSARIF()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	log := unmarshalSARIF(t, data)
	g.Expect(log.Runs[0].Results).To(gomega.HaveLen(1))

	result := log.Runs[0].Results[0]
	g.Expect(result.Related).To(gomega.HaveLen(1))
	g.Expect(result.Related[0].PhysicalLocation.ArtifactLocation.URI).To(gomega.Equal("b.go"))
	g.Expect(result.Related[0].PhysicalLocation.Region.StartLine).To(gomega.Equal(20))
}

func TestToSARIF_EmptyReport(t *testing.T) {
	t.Parallel()

	r := MakeSimpleReport("tool")

	data, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	log := unmarshalSARIF(t, data)

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(log.Runs[0].Results))
	}
}

func testSARIFNaNHandled(t *testing.T, fn func(*Report) ([]byte, error)) {
	t.Helper()
	g := gomega.NewWithT(t)

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m", Severity: SeverityError,
				Position: Position{File: "a.go"}, Confidence: Confidence(math.NaN()),
			},
		},
	}

	data, err := fn(r)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(data).NotTo(gomega.BeEmpty())
}

func TestToSARIF_NaNConfidence(t *testing.T) {
	t.Parallel()
	testSARIFNaNHandled(t, (*Report).ToSARIF)
}

func TestToSARIFFiltered_NaNConfidence(t *testing.T) {
	t.Parallel()
	testSARIFNaNHandled(t, func(r *Report) ([]byte, error) {
		return r.ToSARIFWithOpts(WithMinSeverity(SeverityError))
	})
}

func TestWriteSARIF(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := simpleSARIFReport()

	var buf strings.Builder

	err := r.WriteSARIF(context.Background(), &buf)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	data := buf.String()
	g.Expect(data).To(gomega.ContainSubstring(`"version": "2.1.0"`))
	g.Expect(data).To(gomega.ContainSubstring(`"tool"`))
}

func TestWriteSARIFFiltered(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	r := &Report{
		Tool: ToolInfo{Name: "tool"},
		findings: []Finding{
			{
				ID: "f1", Rule: "r1", Message: "m",
				Severity: SeverityError, Position: Position{File: "a.go"},
			},
			{
				ID: "f2", Rule: "r2", Message: "m",
				Severity: SeverityInfo, Position: Position{File: "b.go"},
			},
		},
	}

	var buf strings.Builder

	err := r.WriteSARIFWithOpts(context.Background(), &buf, WithMinSeverity(SeverityWarning))
	g.Expect(err).NotTo(gomega.HaveOccurred())

	data := buf.String()
	g.Expect(data).To(gomega.ContainSubstring("f1"))
	g.Expect(data).NotTo(gomega.ContainSubstring("f2"))
}

func TestFindingsFromSARIF_EmptyLog(t *testing.T) {
	t.Parallel()

	findings, err := FindingsFromSARIF(
		context.Background(),
		[]byte(`{"version":"2.1.0","runs":[]}`),
	)
	if err != nil {
		t.Fatalf("FindingsFromSARIF(): %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("FindingsFromSARIF() findings = %v, want empty", findings)
	}
}

func TestFindingsFromSARIF_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := FindingsFromSARIF(context.Background(), []byte(`not json`))
	if err == nil {
		t.Fatal("FindingsFromSARIF() expected error for invalid JSON, got nil")
	}
}

func TestFindingsFromSARIF_RoundTrip(t *testing.T) {
	g := gomega.NewWithT(t)
	t.Parallel()

	original := Finding{
		ID:          "govet:printf:main.go:10:5",
		Rule:        "printf",
		ToolName:    "govet",
		Message:     "invalid format",
		Severity:    SeverityCritical,
		Position:    Pos("main.go", 10, 5),
		Category:    CategoryCorrectness,
		Tags:        []Tag{"printf"},
		FixStrategy: FixStrategySuggest,
		Confidence:  0.9,
		Suggestion:  "fix format string",
		Snippet:     "fmt.Sprintf(\"%d\")",
		Metadata:    map[string]string{"custom": "value"},
		Range:       NewRangePtr("main.go", 10, 5, 10, 20),
	}

	report := NewReport(ToolInfo{Name: "govet"})
	report.AddFinding(original)

	sarif, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF(): %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), sarif)
	if err != nil {
		t.Fatalf("FindingsFromSARIF(): %v", err)
	}

	g.Expect(findings).To(gomega.HaveLen(1))

	got := findings[0]

	g.Expect(got.ID).To(gomega.Equal(original.ID))
	g.Expect(got.Rule).To(gomega.Equal(original.Rule))
	g.Expect(got.ToolName).To(gomega.Equal(original.ToolName))
	g.Expect(got.Severity).To(gomega.Equal(original.Severity))
	g.Expect(got.Message).To(gomega.Equal(original.Message))
	g.Expect(got.Category).To(gomega.Equal(original.Category))
	g.Expect(got.Tags).To(gomega.Equal(original.Tags))
	g.Expect(got.FixStrategy).To(gomega.Equal(original.FixStrategy))
	g.Expect(got.Confidence).To(gomega.BeNumerically("~", original.Confidence, 1e-9))
	g.Expect(got.Suggestion).To(gomega.Equal(original.Suggestion))
	g.Expect(got.Snippet).To(gomega.Equal(original.Snippet))
	g.Expect(got.Position).To(gomega.Equal(original.Position))
	g.Expect(got.Range).NotTo(gomega.BeNil())
	g.Expect(got.Range.End).To(gomega.Equal(original.Range.End))
	g.Expect(got.Metadata["custom"]).To(gomega.Equal("value"))
}

func TestSeverityToSARIFLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sev  Severity
		want string
	}{
		{"info", SeverityInfo, "note"},
		{"warning", SeverityWarning, "warning"},
		{"error", SeverityError, "error"},
		{"critical", SeverityCritical, "error"},
		{"unknown", Severity("unknown"), "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := severityToSARIFLevel(tt.sev)
			if got != tt.want {
				t.Errorf("severityToSARIFLevel(%v) = %q, want %q", tt.sev, got, tt.want)
			}
		})
	}
}

func TestDeterminism_ToSARIF_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "det-sarif", Version: "1.0"})
	r.AddFinding(Finding{
		ID:       "f1",
		Rule:     "r1",
		Severity: SeverityError,
		Position: Position{File: FilePath("a.go"), Line: 10},
		Metadata: map[string]string{"zebra": "z", "alpha": "a", "mike": "m", "delta": "d"},
		Tags:     []Tag{"tag-c", "tag-a", "tag-b"},
	})
	r.AddFinding(Finding{
		ID:       "f2",
		Rule:     "r2",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("b.go"), Line: 20},
		Metadata: map[string]string{"k2": "v2", "k0": "v0", "k1": "v1"},
	})
	r.ComputeSummary()

	first, err := r.ToSARIF()
	if err != nil {
		t.Fatalf("first ToSARIF: %v", err)
	}

	for range 100 {
		got, err := r.ToSARIF()
		if err != nil {
			t.Fatalf("ToSARIF: %v", err)
		}

		if !bytes.Equal(first, got) {
			t.Fatalf("non-deterministic ToSARIF output")
		}
	}
}

func TestDeterminism_WriteSARIF_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "det-sarif", Version: "1.0"})
	r.AddFinding(Finding{
		ID:       "f1",
		Rule:     "r1",
		Severity: SeverityError,
		Position: Position{File: FilePath("a.go"), Line: 10},
		Metadata: map[string]string{"zebra": "z", "alpha": "a", "mike": "m"},
	})
	r.ComputeSummary()

	var first bytes.Buffer
	if err := r.WriteSARIF(context.Background(), &first); err != nil {
		t.Fatalf("first WriteSARIF: %v", err)
	}

	for range 100 {
		var buf bytes.Buffer
		if err := r.WriteSARIF(context.Background(), &buf); err != nil {
			t.Fatalf("WriteSARIF: %v", err)
		}

		if !bytes.Equal(first.Bytes(), buf.Bytes()) {
			t.Fatalf("non-deterministic WriteSARIF output")
		}
	}
}
