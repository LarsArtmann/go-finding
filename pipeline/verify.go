package pipeline

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/larsartmann/go-finding"
)

func findingKey(f finding.Finding) string {
	if f.ID != "" {
		return f.ID
	}
	return f.Position.File + "\x00" + f.Rule + "\x00" + f.Message
}

func byFindingID(a, b finding.Finding) int { return cmp.Compare(a.ID, b.ID) }

// VerifyResult holds the outcome of verifying fixes by re-running detectors.
type VerifyResult struct {
	Fixed    []finding.Finding // Findings that were resolved
	Resolved int               // Count of resolved findings
	// Remaining findings that still exist after fixes
	Remaining []finding.Finding
	// New findings introduced by the fixes
	NewFindings []finding.Finding
}

// Verifier re-runs detectors after fixes to verify what was resolved.
type Verifier struct {
	detectors []Detector
}

// NewVerifier creates a verifier that uses the same detectors as the pipeline.
func NewVerifier(detectors []Detector) *Verifier {
	return &Verifier{detectors: detectors}
}

// Verify compares original findings against a fresh detection run.
func (v *Verifier) Verify(ctx context.Context, original []finding.Finding) (*VerifyResult, error) {
	// Re-run all detectors
	var postFindings []finding.Finding

	for _, d := range v.detectors {
		findings, err := d.Detect(ctx)
		if err != nil {
			return nil, fmt.Errorf("verify: detector %s: %w", d.Name(), err)
		}

		for _, f := range findings {
			if !f.IsSuppressed() {
				postFindings = append(postFindings, f)
			}
		}
	}

	return DiffFindings(original, postFindings), nil
}

// DiffFindings compares original and post-fix findings to categorize them.
func DiffFindings(original, post []finding.Finding) *VerifyResult {
	origSet := make(map[string]finding.Finding, len(original))
	for _, f := range original {
		origSet[findingKey(f)] = f
	}

	postSet := make(map[string]finding.Finding, len(post))
	for _, f := range post {
		postSet[findingKey(f)] = f
	}

	var fixed []finding.Finding

	for id, f := range origSet {
		if _, exists := postSet[id]; !exists {
			fixed = append(fixed, f)
		}
	}

	slices.SortFunc(fixed, byFindingID)

	var newFindings []finding.Finding

	for id, f := range postSet {
		if _, exists := origSet[id]; !exists {
			newFindings = append(newFindings, f)
		}
	}

	slices.SortFunc(newFindings, byFindingID)

	var remaining []finding.Finding

	for _, f := range post {
		if _, exists := origSet[findingKey(f)]; exists {
			remaining = append(remaining, f)
		}
	}

	return &VerifyResult{
		Fixed:       fixed,
		Resolved:    len(fixed),
		Remaining:   remaining,
		NewFindings: newFindings,
	}
}
