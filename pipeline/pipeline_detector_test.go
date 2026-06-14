package pipeline

import (
	"context"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestDetectParallel_SuppressionConsistency(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	makeFindings := func() []finding.Finding {
		return []finding.Finding{
			{
				ID:       "s1",
				Rule:     "r1",
				ToolName: "t1",
				Message:  "active-1",
				Severity: finding.SeverityError,
			},
			{
				ID:       "s2",
				Rule:     "r2",
				ToolName: "t1",
				Message:  "suppressed-1",
				Severity: finding.SeverityWarning,
				Suppression: &finding.Suppression{
					Kind:   finding.SuppressionInSource,
					Rule:   "r2",
					Reason: "false positive",
				},
			},
			{
				ID:       "s3",
				Rule:     "r3",
				ToolName: "t1",
				Message:  "active-2",
				Severity: finding.SeverityInfo,
			},
			{
				ID:       "s4",
				Rule:     "r4",
				ToolName: "t1",
				Message:  "suppressed-2",
				Severity: finding.SeverityError,
				Suppression: &finding.Suppression{
					Kind:   finding.SuppressionInConfig,
					Rule:   "r4",
					Reason: "legacy",
				},
			},
		}
	}

	runDetect := func(parallel bool) ([]finding.Finding, []string) {
		var notified []string

		config := DefaultConfig()
		config.ParallelDetectors = parallel
		config.OnFinding = func(f finding.Finding) {
			notified = append(notified, f.ID)
		}

		d1 := &mockDetector{name: "d1", findings: makeFindings()}
		d2 := &mockDetector{name: "d2", findings: makeFindings()}

		p, err := New(config, t.TempDir(), d1, d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		findings, err := p.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var all []finding.Finding
		for _, iter := range findings.Iterations {
			all = append(all, iter.Findings()...)
		}

		return all, notified
	}

	seqFindings, seqNotified := runDetect(false)
	parFindings, parNotified := runDetect(true)

	g.Expect(seqFindings).To(HaveLen(len(parFindings)))

	g.Expect(seqNotified).To(HaveLen(len(parNotified)))

	assertNoSuppressedFindings(t, seqFindings, "sequential")
	assertNoSuppressedFindings(t, parFindings, "parallel")
	assertNoSuppressedNotified(t, seqNotified, "sequential")
	assertNoSuppressedNotified(t, parNotified, "parallel")
}

// TestDetectorFunc tests the DetectorFunc adapter.
func TestDetectorFunc(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	called := false
	f := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		called = true

		return nil, nil
	})

	_, err := f.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("function was not called")
	}

	g.Expect(f.Name()).To(Equal("anonymous"))
}

func TestNamedDetectorFunc(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	fn := makeFindingDetectorFunc("test")

	d := NamedDetectorFunc("my-linter", fn)

	g.Expect(d.Name()).To(Equal("my-linter"))

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(findings).To(HaveLen(1))
	g.Expect(findings[0].ID).To(Equal("test"))
}

// TestFixApplier tests the FixApplier.
