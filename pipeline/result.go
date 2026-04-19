package pipeline

import (
	"slices"

	"github.com/larsartmann/go-finding"
)

// PipelineResult contains the outcome of running the pipeline.
//
//nolint:revive // stuttering name is intentional for clarity
type PipelineResult struct {
	Stable            bool
	TotalIterations   int
	Iterations        []Iteration
	FinalFindingCount int
	Verification      *VerifyResult
}

// Iteration represents one loop through the pipeline.
type Iteration struct {
	Number        int
	FindingsFound int
	DirectFixes   int
	SuggestFixes  int
	NoFix         int
	Conflicts     int
	Applied       int
	Failed        int
	findings      []finding.Finding
	suggest       []finding.Finding
}

// Findings returns a copy of all findings discovered in this iteration.
func (it Iteration) Findings() []finding.Finding {
	return slices.Clone(it.findings)
}

// SuggestedFindings returns a copy of findings that have FixStrategySuggest.
func (it Iteration) SuggestedFindings() []finding.Finding {
	return slices.Clone(it.suggest)
}
