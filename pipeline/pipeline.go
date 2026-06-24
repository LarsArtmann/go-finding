// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
)

// errAlreadyRan is the sentinel error returned when Run is called more than once.
var errAlreadyRan = errors.New(
	"pipeline: Run already called; create a new Pipeline for each invocation",
)

// reasonFromContext derives a CompletionReason from the context error.
func reasonFromContext(ctx context.Context) CompletionReason {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ReasonTimeout
	}

	return ReasonCancelled
}

// Pipeline orchestrates the detect → triage → fix → verify loop.
type Pipeline struct {
	config     Config
	detectors  []Detector
	rootDir    string
	iterations int
	findings   []finding.Finding
	metrics    *Metrics
	callbackMu sync.Mutex  // protects OnFinding from parallel goroutines
	applier    *FixApplier // reused across iterations
	ran        bool        // prevents multiple Run calls
}

// New creates a new Pipeline with the given configuration.
// Returns an error if the configuration is invalid.
func New(config Config, rootDir string, detectors ...Detector) (*Pipeline, error) {
	// Wrap detectors with retry if configured.
	if config.Retry != nil {
		wrapped := make([]Detector, len(detectors))
		for i, d := range detectors {
			wrapped[i] = NewRetryDetector(d, *config.Retry)
		}

		detectors = wrapped
	}

	var (
		applier *FixApplier
		err     error
	)

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate config (detectors=%d): %w", len(detectors), err)
	}

	// Create FixApplier eagerly so errors are caught early.
	if len(config.FixProviders) > 0 {
		applier, err = NewFixApplierWithProviders(rootDir, config.FixProviders...)
	} else {
		applier, err = NewFixApplier(rootDir)
	}

	if err != nil {
		return nil, fmt.Errorf("init fix applier: %w", err)
	}

	//nolint:exhaustruct
	return &Pipeline{
		config:    config,
		detectors: detectors,
		rootDir:   rootDir,
		findings:  make([]finding.Finding, 0),
		metrics:   config.Metrics,
		applier:   applier,
	}, nil
}

// stageTiming returns a function that records stage duration when called.
// Returns a no-op if metrics collection is disabled.
func (p *Pipeline) stageTiming(name Stage) func() {
	if p.metrics == nil {
		return func() {}
	}

	return p.metrics.StageTiming(name)
}

// log emits a structured log event if a logger is configured.
func (p *Pipeline) log(ctx context.Context, msg string, attrs ...slog.Attr) {
	if p.config.Logger != nil {
		p.config.Logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	}
}

// fireStageHook sends a StageEvent to all registered StageHooks.
// Returns the first error encountered, which aborts the pipeline.
func (p *Pipeline) fireStageHook(
	ctx context.Context,
	timing StageTiming,
	stage Stage,
	iteration int,
	findings []finding.Finding,
	applied int,
	conflicts int,
) error {
	if len(p.config.StageHooks) == 0 {
		return nil
	}

	event := StageEvent{
		Stage:     stage,
		Timing:    timing,
		Iteration: iteration,
		Findings:  findings,
		Applied:   applied,
		Conflicts: conflicts,
	}

	for _, hook := range p.config.StageHooks {
		err := hook.OnStageEvent(ctx, event)
		if err != nil {
			return fmt.Errorf("stage hook: %w", err)
		}
	}

	return nil
}

// Run executes the pipeline until stable or max iterations reached.
//
// Run is NOT safe for concurrent use. Each Pipeline instance should be used
// for at most one Run call; create a new Pipeline for each invocation.
// Calling Run multiple times on the same Pipeline returns errAlreadyRan.
//
// The returned PipelineResult is safe to read concurrently after Run returns.
func (p *Pipeline) Run(ctx context.Context) (*PipelineResult, error) {
	if p.ran {
		return nil, errAlreadyRan
	}

	p.ran = true
	p.findings = p.findings[:0]
	p.iterations = 0

	if p.metrics != nil {
		p.metrics.SetStart(time.Now())
	}

	var metricsResult *PipelineResult

	defer func() {
		if p.applier != nil {
			_ = p.applier.Close()
			p.applier = nil
		}

		if p.metrics != nil {
			p.metrics.SetEnd(time.Now())

			if metricsResult != nil {
				metricsResult.Metrics = p.metrics.Snapshot()
			}
		}
	}()

	if p.config.Timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	//nolint:exhaustruct
	result := &PipelineResult{
		Iterations: make([]Iteration, 0, p.config.MaxIterations),
	}

	for p.iterations < p.config.MaxIterations {
		err := CheckCanceledWithMsg(ctx, "pipeline cancelled")
		if err != nil {
			result.Reason = reasonFromContext(ctx)

			return result, err
		}

		p.log(
			ctx, "iteration starting",
			slog.Int("iteration", p.iterations+1),
			slog.Int("max_iterations", p.config.MaxIterations),
		)

		done, err := p.runIteration(ctx, result)
		if err != nil {
			if IsContextError(err) {
				result.Reason = reasonFromContext(ctx)
			} else {
				result.Reason = ReasonError
			}

			return result, err
		}

		if done {
			break
		}
	}

	result.TotalIterations = len(result.Iterations)

	if result.Reason == "" {
		result.Reason = ReasonMaxIterations
	}

	// Optional cross-tool correlation
	if p.config.CorrelateFindings {
		result.Correlations = finding.Correlate(p.findings)
	}

	// Optional final verification
	if p.config.VerifyAfterFix && len(p.detectors) > 0 {
		allOriginal := p.collectAllFindings(result)

		_ = p.fireStageHook(ctx, StageBefore, StageVerify, result.TotalIterations, allOriginal, 0, 0)

		verifyDone := p.stageTiming(StageVerify)
		verifyResult, err := Verify(ctx, p.detectors, allOriginal)

		verifyDone()

		if err != nil {
			return result, fmt.Errorf("verify: %w", err)
		}

		_ = p.fireStageHook(ctx, StageAfter, StageVerify, result.TotalIterations, allOriginal, 0, 0)

		result.Verification = verifyResult
	}

	result.TotalDetected = len(p.findings)

	metricsResult = result

	return result, nil
}
