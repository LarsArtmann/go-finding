package pipeline

import (
	"slices"

	"github.com/larsartmann/go-finding"
)

// CompletionReason describes why the pipeline finished.
type CompletionReason string

const (
	// ReasonStable means no findings were detected — the codebase is clean.
	ReasonStable CompletionReason = "stable"
	// ReasonMaxIterations means the pipeline hit the configured iteration limit.
	ReasonMaxIterations CompletionReason = "max-iterations"
	// ReasonCancelled means the pipeline was cancelled via context.
	ReasonCancelled CompletionReason = "cancelled"
	// ReasonTimeout means the pipeline exceeded its configured timeout.
	ReasonTimeout CompletionReason = "timeout"
	// ReasonError means the pipeline stopped due to an unrecoverable error.
	ReasonError CompletionReason = "error"
)

// PipelineResult contains the outcome of running the pipeline.
//
//nolint:revive // stuttering name is intentional for clarity
type PipelineResult struct {
	Reason          CompletionReason
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

// Stable reports whether the pipeline reached a clean state (no findings).
func (r *PipelineResult) Stable() bool {
	return r.Reason == ReasonStable
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
