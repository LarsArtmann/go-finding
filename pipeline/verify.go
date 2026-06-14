package pipeline

import (
	"context"
	"fmt"

	"github.com/larsartmann/go-finding"
)

// VerifyResult holds the outcome of verifying fixes by re-running detectors.
type VerifyResult struct {
	Fixed    []finding.Finding // Findings that were resolved
	Resolved int               // Count of resolved findings
	// Remaining findings that still exist after fixes
	Remaining []finding.Finding
	// New findings introduced by the fixes
	NewFindings []finding.Finding
	// Modified findings — same key but different content
	Modified []finding.Finding
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
		err := CheckCanceled(ctx)
		if err != nil {
			return nil, err
		}

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

	fixed := make([]finding.Finding, 0, len(original))

	for id, f := range origSet {
		if _, exists := postSet[id]; !exists {
			fixed = append(fixed, f)
		}
	}

	finding.SortFindingsByID(fixed)

	newFindings := make([]finding.Finding, 0, len(post))

	for id, f := range postSet {
		if _, exists := origSet[id]; !exists {
			newFindings = append(newFindings, f)
		}
	}

	finding.SortFindingsByID(newFindings)

	var (
		remaining = make([]finding.Finding, 0, len(post))
		modified  = make([]finding.Finding, 0, len(post))
	)

	for _, f := range post {
		orig, exists := origSet[f.Key()]
		if !exists {
			continue
		}

		if !orig.Equal(f) {
			modified = append(modified, f)
		} else {
			remaining = append(remaining, f)
		}
	}

	finding.SortFindingsByID(modified)

	return &VerifyResult{
		Fixed:       fixed,
		Resolved:    len(fixed),
		Remaining:   remaining,
		NewFindings: newFindings,
		Modified:    modified,
	}
}
