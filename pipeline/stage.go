package pipeline

// Stage identifies a pipeline stage for metrics and callbacks.
type Stage string

// Pipeline stage identifiers used in metrics and OnStage callbacks.
const (
	// StageDetect is the detection stage where detectors run.
	StageDetect Stage = "detect"
	// StageProcess is the processor stage where FindingTransformers transform results.
	StageProcess Stage = "process"
	// StageTriage is the triage stage where findings are categorized by fix strategy.
	StageTriage Stage = "triage"
	// StageApply is the application stage where fixes are written to disk.
	StageApply Stage = "apply"
	// StageVerify is the verification stage where detectors re-run to confirm fixes.
	StageVerify Stage = "verify"
)
