package pipeline

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/larsartmann/go-finding"
)

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

// Verify compares original findings against a fresh detection run
// by re-running all detectors and diffing the results.
func Verify(
	ctx context.Context,
	detectors []Detector,
	original []finding.Finding,
) (*VerifyResult, error) {
	var postFindings []finding.Finding

	for _, d := range detectors {
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
		origSet[f.Key()] = f
	}

	postSet := make(map[string]finding.Finding, len(post))
	for _, f := range post {
		postSet[f.Key()] = f
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
		if _, exists := origSet[f.Key()]; exists {
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
