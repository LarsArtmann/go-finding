package pipeline

import (
	"context"

	"github.com/larsartmann/go-finding"
)

// StageTiming indicates whether a stage hook fires before or after the stage executes.
type StageTiming string

const (
	// StageBefore fires before the stage starts executing.
	StageBefore StageTiming = "before"
	// StageAfter fires after the stage completes successfully.
	StageAfter StageTiming = "after"
)

// StageEvent carries context about a pipeline stage boundary.
type StageEvent struct {
	Stage     Stage
	Timing    StageTiming
	Iteration int
	Findings  []finding.Finding
	Applied   int // Fixes applied (StageAfter + StageApply only)
	// Conflicts holds the number of conflicts detected during fix application
	// (StageAfter + StageApply only).
	Conflicts int
}

// StageHook receives stage boundary notifications. Implementations must be
// safe for concurrent use if ParallelDetectors is enabled.
//
// Returning a non-nil error from a StageBefore hook aborts the pipeline.
// Returning a non-nil error from a StageAfter hook also aborts the pipeline.
type StageHook interface {
	OnStageEvent(ctx context.Context, event StageEvent) error
}

// StageHookFunc adapts a function to the StageHook interface.
type StageHookFunc func(ctx context.Context, event StageEvent) error

// OnStageEvent calls the underlying function.
func (f StageHookFunc) OnStageEvent(ctx context.Context, event StageEvent) error {
	return f(ctx, event)
}
