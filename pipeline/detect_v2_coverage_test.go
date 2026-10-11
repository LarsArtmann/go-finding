package pipeline

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/larsartmann/go-finding"
)

// v2Detector is a mockDetector that also measures its own coverage, the
// minimal DetectorV2 implementation.
type v2Detector struct {
	mockDetector
	filesScanned   int
	skippedModules []string
}

func (d *v2Detector) DetectV2(ctx context.Context) (finding.DetectResult, error) {
	findings, err := d.Detect(ctx)

	return finding.DetectResult{
		Findings:       findings,
		FilesScanned:   d.filesScanned,
		SkippedModules: d.skippedModules,
	}, err
}

func TestPipelineSurfacesV2Coverage(t *testing.T) {
	t.Parallel()

	first := &v2Detector{
		mockDetector:   mockDetector{name: "v2-first"},
		filesScanned:   3,
		skippedModules: []string{"mod-a", "mod-b"},
	}

	second := &v2Detector{
		mockDetector:   mockDetector{name: "v2-second"},
		filesScanned:   2,
		skippedModules: []string{"mod-b", "mod-c"},
	}

	p, err := New(Config{MaxIterations: DefaultMaxIterations}, t.TempDir(), first, second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.FilesScanned != 5 {
		t.Errorf("FilesScanned = %d, want 5 (3 + 2)", result.FilesScanned)
	}

	want := []string{"mod-a", "mod-b", "mod-c"}
	if !slices.Equal(result.SkippedModules, want) {
		t.Errorf("SkippedModules = %v, want first-seen union %v", result.SkippedModules, want)
	}
}

func TestPipelineV1DetectorZeroesCoverage(t *testing.T) {
	t.Parallel()

	p, err := New(Config{MaxIterations: DefaultMaxIterations}, t.TempDir(), newMockDetector("v1-only", "tool"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.FilesScanned != 0 {
		t.Errorf("FilesScanned = %d, want 0 (v1 contract: no measurement)", result.FilesScanned)
	}

	if len(result.SkippedModules) != 0 {
		t.Errorf("SkippedModules = %v, want empty", result.SkippedModules)
	}
}

func TestDetectPartialAccumulatesCoverageFromSuccessfulDetectors(t *testing.T) {
	t.Parallel()

	failing := &mockDetector{name: "boom", err: errors.New("boom")}

	healthy := &v2Detector{
		mockDetector:   mockDetector{name: "v2-healthy"},
		filesScanned:   7,
		skippedModules: []string{"nested"},
	}

	p, err := New(Config{MaxIterations: DefaultMaxIterations}, t.TempDir(), failing, healthy)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	partial, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("DetectPartial: %v", err)
	}

	if !partial.HasErrors() {
		t.Fatal("partial result should record the failing detector")
	}

	if partial.FilesScanned != 7 {
		t.Errorf("FilesScanned = %d, want 7 (successes only)", partial.FilesScanned)
	}

	if !slices.Equal(partial.SkippedModules, []string{"nested"}) {
		t.Errorf("SkippedModules = %v, want [nested]", partial.SkippedModules)
	}
}

func TestRetryDetectorPreservesV2Coverage(t *testing.T) {
	t.Parallel()

	inner := &v2Detector{
		mockDetector:   mockDetector{name: "v2-inner"},
		filesScanned:   4,
		skippedModules: []string{"vendored"},
	}

	retry := NewRetryDetector(inner, RetryConfig{MaxRetries: 2})

	result, err := retry.DetectV2(context.Background())
	if err != nil {
		t.Fatalf("DetectV2: %v", err)
	}

	if result.FilesScanned != 4 {
		t.Errorf("FilesScanned = %d, want 4 (v2 survives the decorator)", result.FilesScanned)
	}

	if !slices.Equal(result.SkippedModules, []string{"vendored"}) {
		t.Errorf("SkippedModules = %v, want [vendored]", result.SkippedModules)
	}

	findings, err := retry.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != len(result.Findings) {
		t.Errorf("Detect returned %d findings, DetectV2 returned %d; they must agree",
			len(findings), len(result.Findings))
	}
}

func TestConvenienceDetectRunsV2Detector(t *testing.T) {
	t.Parallel()

	v2 := &v2Detector{
		mockDetector: mockDetector{
			name:     "v2-convenience",
			findings: []finding.Finding{{ID: "1"}},
		},
		filesScanned: 9,
	}

	findings, err := Detect(context.Background(), v2)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 {
		t.Errorf("Detect returned %d findings, want 1", len(findings))
	}
}
