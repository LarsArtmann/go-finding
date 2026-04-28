// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
	"golang.org/x/sync/errgroup"
)

// Detector is the interface implemented by tools that can find issues.
type Detector interface {
	// Name returns the detector's name.
	Name() string
	// Detect runs the detector and returns findings.
	Detect(ctx context.Context) ([]finding.Finding, error)
}

// DetectorFunc is an adapter to use ordinary functions as Detectors.
type DetectorFunc func(ctx context.Context) ([]finding.Finding, error)

// Detect implements Detector.
func (f DetectorFunc) Detect(ctx context.Context) ([]finding.Finding, error) {
	return f(ctx)
}

// Name implements Detector. Returns "anonymous" — use NamedDetectorFunc for a custom name.
//
//nolint:revive // receiver unused by design — method exists only to satisfy Detector interface
func (f DetectorFunc) Name() string {
	return "anonymous"
}

// NamedDetectorFunc returns a Detector with the given name wrapping the provided function.
func NamedDetectorFunc(name string, fn DetectorFunc) Detector {
	return &namedDetector{name: name, fn: fn}
}

type namedDetector struct {
	name string
	fn   DetectorFunc
}

func (n *namedDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	return n.fn(ctx)
}

func (n *namedDetector) Name() string {
	return n.name
}

// Config configures the pipeline behavior.
type Config struct {
	// MaxIterations prevents infinite loops.
	MaxIterations int
	// ParallelDetectors runs detectors concurrently.
	ParallelDetectors bool
	// Timeout for the entire pipeline.
	Timeout time.Duration
	// VerifyAfterFix runs a final verification pass after all iterations.
	VerifyAfterFix bool
	// GracefulDegradation continues on detector failures, collecting partial results.
	GracefulDegradation bool
	// DryRun runs detect+triage but skips fix application.
	DryRun bool
	// Retry wraps each detector with retry logic. nil disables retries.
	Retry *RetryConfig
	// OnFinding is called for each finding found.
	OnFinding func(f finding.Finding)
	// OnFix is called when a fix is applied.
	OnFix func(f finding.Finding, applied bool)
	// OnIteration is called at the end of each iteration.
	OnIteration func(iter int, findings []finding.Finding)
	// Metrics collects timing and count data. If nil, no metrics are collected.
	Metrics *Metrics
	// CorrelateFindings runs cross-tool correlation on all findings after detection.
	// Results are stored in PipelineResult.Correlations.
	CorrelateFindings bool
}

const defaultMaxIterations = 5

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	//nolint:exhaustruct
	return Config{
		MaxIterations:     defaultMaxIterations,
		ParallelDetectors: true,
		Timeout:           10 * time.Minute,
	}
}

// Sentinel validation errors.
var (
	errMaxIterations = errors.New("max iterations must be >= 0")
	errTimeout       = errors.New("timeout must be >= 0")
)

// Validate checks the configuration and returns an error if invalid.
func (c Config) Validate() error {
	var errs []error

	if c.MaxIterations < 0 {
		errs = append(errs, fmt.Errorf("%w: got %d", errMaxIterations, c.MaxIterations))
	}

	if c.Timeout < 0 {
		errs = append(errs, fmt.Errorf("%w: got %v", errTimeout, c.Timeout))
	}

	if c.Retry != nil {
		if err := c.Retry.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("retry: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Pipeline orchestrates the detect → triage → fix → verify loop.
type Pipeline struct {
	config     Config
	detectors  []Detector
	rootDir    string
	iterations int
	findings   []finding.Finding
	metrics    *Metrics
}

// New creates a new Pipeline with the given configuration.
// Returns an error if the configuration is invalid.
func New(config Config, rootDir string, detectors ...Detector) (*Pipeline, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Wrap detectors with retry if configured.
	if config.Retry != nil {
		wrapped := make([]Detector, len(detectors))
		for i, d := range detectors {
			wrapped[i] = NewRetryDetector(d, *config.Retry)
		}

		detectors = wrapped
	}

	//nolint:exhaustruct
	return &Pipeline{
		config:    config,
		detectors: detectors,
		rootDir:   rootDir,
		findings:  make([]finding.Finding, 0),
		metrics:   config.Metrics,
	}, nil
}

// stageTiming returns a function that records stage duration when called.
// Returns a no-op if metrics collection is disabled.
func (p *Pipeline) stageTiming(name string) func() {
	if p.metrics == nil {
		return func() {}
	}

	return p.metrics.StageTiming(name)
}

// Run executes the pipeline until stable or max iterations reached.
//
// Run is NOT safe for concurrent use. Create a new Pipeline for each
// concurrent invocation. The returned PipelineResult is safe to read
// concurrently after Run returns.
func (p *Pipeline) Run(ctx context.Context) (*PipelineResult, error) {
	p.findings = p.findings[:0]
	p.iterations = 0

	if p.metrics != nil {
		p.metrics.SetStart(time.Now())
	}

	var metricsResult *PipelineResult

	defer func() {
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
		if isContextDone(ctx) {
			return result, contextError(ctx, "pipeline cancelled")
		}

		iter := Iteration{Number: p.iterations + 1} //nolint:exhaustruct

		// Detect
		detectDone := p.stageTiming("detect")
		detResult, err := p.detect(ctx)

		detectDone()

		if err != nil {
			return result, fmt.Errorf("iteration %d: detect: %w", p.iterations+1, err)
		}

		findings := detResult.Findings

		// Accumulate partial errors across iterations.
		for name, detErr := range detResult.PartialErrors {
			if result.PartialErrors == nil {
				result.PartialErrors = make(map[string]error)
			}

			result.PartialErrors[name] = detErr
		}

		iter.FindingsFound = len(findings)
		iter.findings = findings
		p.findings = append(p.findings, findings...)

		// If no findings, we're done
		if len(findings) == 0 {
			result.Stable = true
			result.Iterations = append(result.Iterations, iter)

			break
		}

		// Triage
		triage := p.triage(findings)
		iter.DirectFixes = len(triage.Direct)
		iter.SuggestFixes = len(triage.Suggest)
		iter.suggest = triage.Suggest
		iter.NoFix = len(triage.None)

		// Apply fixes (with conflict detection)
		if !p.config.DryRun {
			applyDone := p.stageTiming("apply")
			if err := p.applyTriage(ctx, triage.Direct, &iter); err != nil {
				applyDone()

				return result, fmt.Errorf("iteration %d: %w", p.iterations+1, err)
			}

			applyDone()
		}

		result.Iterations = append(result.Iterations, iter)
		p.iterations++

		if p.config.OnIteration != nil {
			p.config.OnIteration(p.iterations, findings)
		}
	}

	result.TotalIterations = len(result.Iterations)

	// Optional cross-tool correlation
	if p.config.CorrelateFindings {
		result.Correlations = finding.Correlate(p.findings)
	}

	// Optional final verification
	if p.config.VerifyAfterFix && len(p.detectors) > 0 {
		verifier := NewVerifier(p.detectors)
		allOriginal := p.collectAllFindings(result)

		verifyResult, err := verifier.Verify(ctx, allOriginal)
		if err != nil {
			return result, fmt.Errorf("verify: %w", err)
		}

		result.Verification = verifyResult
	}

	result.TotalDetected = len(p.findings)

	metricsResult = result

	return result, nil
}

// collectAllFindings gathers all findings from all iterations for verification.
func (*Pipeline) collectAllFindings(
	result *PipelineResult,
) []finding.Finding {
	seen := make(map[string]struct{})

	var all []finding.Finding

	for _, iter := range result.Iterations {
		for _, f := range iter.findings {
			if _, exists := seen[f.ID]; !exists {
				seen[f.ID] = struct{}{}
				all = append(all, f)
			}
		}
	}

	return all
}

// detectResult holds findings and optional partial errors from detection.
type detectResult struct {
	Findings      []finding.Finding
	PartialErrors map[string]error
}

// detect runs all detectors and collects findings.
func (p *Pipeline) detect(ctx context.Context) (*detectResult, error) {
	if p.config.GracefulDegradation {
		result, err := p.DetectPartial(ctx)
		if err != nil {
			return nil, err
		}

		return &detectResult{
			Findings:      result.Findings,
			PartialErrors: result.Errors,
		}, nil
	}

	if p.config.ParallelDetectors {
		findings, err := p.detectParallel(ctx)
		if err != nil {
			return nil, err
		}

		return &detectResult{Findings: findings}, nil //nolint:exhaustruct
	}

	findings, err := p.detectSequential(ctx)
	if err != nil {
		return nil, err
	}

	return &detectResult{Findings: findings}, nil //nolint:exhaustruct
}

// addFindings adds non-suppressed findings to the target slice, calling OnFinding if set.
func (p *Pipeline) addFindings(target, findings []finding.Finding) []finding.Finding {
	for _, f := range findings {
		if !f.IsSuppressed() {
			target = append(target, f)
			p.notifyFinding(f)
		}
	}

	return target
}

// recordDetectorMetrics records timing metrics for a detector if metrics are enabled.
func (p *Pipeline) recordDetectorMetrics(
	name string,
	elapsed time.Duration,
	findings []finding.Finding,
) {
	if p.metrics != nil {
		p.metrics.RecordDetector(name, elapsed, len(findings))
	}
}

// detectSequential runs detectors one at a time.
func (p *Pipeline) detectSequential(ctx context.Context) ([]finding.Finding, error) {
	var allFindings []finding.Finding

	for _, d := range p.detectors {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("detection cancelled: %w", ctx.Err())
		default:
		}

		start := time.Now()
		findings, err := d.Detect(ctx)
		elapsed := time.Since(start)

		if err != nil {
			return nil, fmt.Errorf("detector %s: %w", d.Name(), err)
		}

		p.recordDetectorMetrics(d.Name(), elapsed, findings)

		allFindings = p.addFindings(allFindings, findings)
	}

	return allFindings, nil
}

// detectParallel runs detectors concurrently using errgroup.
func (p *Pipeline) detectParallel(ctx context.Context) ([]finding.Finding, error) {
	var (
		mu          sync.Mutex
		allFindings []finding.Finding
	)

	g, ctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		g.Go(func() error {
			start := time.Now()
			findings, err := d.Detect(ctx)
			elapsed := time.Since(start)

			if err != nil {
				return fmt.Errorf("detector %s: %w", d.Name(), err)
			}

			p.recordDetectorMetrics(d.Name(), elapsed, findings)

			mu.Lock()
			allFindings = p.addFindings(allFindings, findings)
			mu.Unlock()

			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, fmt.Errorf("parallel detection: %w", err)
	}

	return allFindings, nil
}

// TriageResult holds findings categorized by fix strategy.
type TriageResult struct {
	Direct  []finding.Finding
	Suggest []finding.Finding
	None    []finding.Finding
}

// triage categorizes findings by their fix strategy.
//
//nolint:revive // receiver unused by design — method belongs to Pipeline for API cohesion
func (p *Pipeline) triage(findings []finding.Finding) *TriageResult {
	result := &TriageResult{
		Direct:  make([]finding.Finding, 0),
		Suggest: make([]finding.Finding, 0),
		None:    make([]finding.Finding, 0),
	}

	for _, f := range findings {
		switch f.FixStrategy {
		case finding.FixStrategyDirect:
			result.Direct = append(result.Direct, f)
		case finding.FixStrategySuggest, finding.FixStrategyAI:
			result.Suggest = append(result.Suggest, f)
		case finding.FixStrategyNone:
			result.None = append(result.None, f)
		default:
			result.None = append(result.None, f)
		}
	}

	return result
}

// applyTriage handles conflict detection and fix application for one iteration.
func (p *Pipeline) applyTriage(
	ctx context.Context,
	fixes []finding.Finding,
	iter *Iteration,
) error {
	if len(fixes) == 0 {
		return nil
	}

	safeFixes := FilterConflictingFixes(fixes)
	iter.Conflicts = len(fixes) - len(safeFixes)

	if iter.Conflicts > 0 {
		for _, c := range AnalyzeConflicts(fixes) {
			if p.config.OnFix != nil {
				p.config.OnFix(c.Finding, false)
			}
		}
	}

	if len(safeFixes) == 0 {
		return nil
	}

	applied, err := p.applyDirectFixes(ctx, safeFixes)
	if err != nil {
		return fmt.Errorf("apply fixes: %w", err)
	}

	iter.Applied = applied

	if p.config.OnFix != nil {
		for i := 0; i < applied && i < len(safeFixes); i++ {
			p.config.OnFix(safeFixes[i], true)
		}
	}

	return nil
}

// applyDirectFixes applies deterministic fixes to files.
func (p *Pipeline) applyDirectFixes(ctx context.Context, fixes []finding.Finding) (int, error) {
	applier := NewFixApplier(p.rootDir)

	applied, err := applier.Apply(ctx, fixes)
	if err != nil {
		return applied, err
	}

	if p.metrics != nil {
		for range applied {
			p.metrics.RecordFix()
		}
	}

	return applied, nil
}
