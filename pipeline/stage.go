package pipeline

// Stage identifies a pipeline stage for metrics and callbacks.
type Stage string

const (
	StageDetect  Stage = "detect"
	StageProcess Stage = "process"
	StageTriage  Stage = "triage"
	StageApply   Stage = "apply"
	StageVerify  Stage = "verify"
)
