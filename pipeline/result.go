package pipeline

import (
	"slices"

	"github.com/larsartmann/go-finding"
)

// PipelineResult contains the outcome of running the pipeline.
//
//nolint:revive // stuttering name is intentional for clarity
type PipelineResult struct {
	Stable          bool
	TotalIterations int
	Iterations      []Iteration
	// TotalDetected is the total number of findings detected across all iterations.
	// This includes findings that were fixed in subsequent iterations.
	TotalDetected int
	Verification  *VerifyResult
	PartialErrors map[string]error
	Metrics       MetricsSnapshot
	// Correlations holds cross-tool finding correlations when
	// Config.CorrelateFindings is enabled.
	Correlations []finding.Correlation
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
