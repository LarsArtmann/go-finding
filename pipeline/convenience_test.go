package pipeline

import (
	"context"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/onsi/gomega"
)

func TestDetect_EmptyDetectors(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	findings, err := Detect(context.Background())
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.BeNil())
}

func TestDetect_SingleDetector(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	expected := []finding.Finding{
		{ID: "f1", Rule: "r1", ToolName: "test", Severity: finding.SeverityWarning},
		{ID: "f2", Rule: "r1", ToolName: "test", Severity: finding.SeverityError},
	}

	detector := finding.NamedDetectorFunc("test-detector", func(_ context.Context) ([]finding.Finding, error) {
		return expected, nil
	})

	findings, err := Detect(context.Background(), detector)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(2))
	g.Expect(string(findings[0].ID)).To(gomega.Equal("f1"))
	g.Expect(string(findings[1].ID)).To(gomega.Equal("f2"))
}

func TestDetect_MultipleDetectorsParallel(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	d1 := finding.NamedDetectorFunc("d1", func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{{ID: "a", ToolName: "d1"}}, nil
	})
	d2 := finding.NamedDetectorFunc("d2", func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{{ID: "b", ToolName: "d2"}}, nil
	})

	findings, err := Detect(context.Background(), d1, d2)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(2))
}

func TestDetect_FiltersSuppressed(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	suppressed := finding.Finding{
		ID:          "supp",
		ToolName:    "test",
		Suppression: &finding.Suppression{Kind: finding.SuppressionInConfig},
	}
	active := finding.Finding{ID: "act", ToolName: "test"}

	detector := finding.NamedDetectorFunc("test", func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{suppressed, active}, nil
	})

	findings, err := Detect(context.Background(), detector)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(findings).To(gomega.HaveLen(1))
	g.Expect(string(findings[0].ID)).To(gomega.Equal("act"))
}

func TestDetect_ContextCancellation(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	detector := finding.NamedDetectorFunc("test", func(_ context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	_, err := Detect(ctx, detector)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestDetect_DetectorError(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	detector := finding.NamedDetectorFunc("failing", func(_ context.Context) ([]finding.Finding, error) {
		return nil, context.DeadlineExceeded
	})

	_, err := Detect(context.Background(), detector)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("failing"))
}

func TestApplyToContent_SingleFix(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	content := []byte("package main\n\noldFunc()\n")
	fix := finding.Finding{
		ID:          "f1",
		Rule:        "rename",
		ToolName:    "test",
		BeforeCode:  "oldFunc()",
		AfterCode:   "newFunc()",
		FixStrategy: finding.FixStrategyDirect,
	}

	result, applied := ApplyToContent(content, []finding.Finding{fix})
	g.Expect(applied).To(gomega.Equal(1))
	g.Expect(string(result)).To(gomega.ContainSubstring("newFunc()"))
	g.Expect(string(result)).NotTo(gomega.ContainSubstring("oldFunc()"))
}

func TestApplyToContent_NoFixes(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	content := []byte("package main\n")

	result, applied := ApplyToContent(content, nil)
	g.Expect(applied).To(gomega.Equal(0))
	g.Expect(result).To(gomega.Equal(content))
}

func TestApplyToContent_NoMatch(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	content := []byte("package main\n")
	fix := finding.Finding{
		ID:          "f1",
		ToolName:    "test",
		BeforeCode:  "nonexistent",
		AfterCode:   "replacement",
		FixStrategy: finding.FixStrategyDirect,
	}

	result, applied := ApplyToContent(content, []finding.Finding{fix})
	g.Expect(applied).To(gomega.Equal(0))
	g.Expect(string(result)).To(gomega.Equal("package main\n"))
}
