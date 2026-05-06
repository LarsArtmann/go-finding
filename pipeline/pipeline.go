// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
	"golang.org/x/sync/errgroup"
)

// Pipeline orchestrates the detect → triage → fix → verify loop.
type Pipeline struct {
	config     Config
	detectors  []Detector
	rootDir    string
	iterations int
	findings   []finding.Finding
	metrics    *Metrics
	callbackMu sync.Mutex // protects OnFinding from parallel goroutines
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
// Run is NOT safe for concurrent use. Each Pipeline instance should be used
// for at most one Run call; create a new Pipeline for each invocation.
// Calling Run multiple times on the same Pipeline resets internal state
// (findings, iterations) but does not re-validate the configuration.
//
// The returned PipelineResult is safe to read concurrently after Run returns.
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
		if err := CheckCanceledWithMsg(ctx, "pipeline cancelled"); err != nil {
			return result, err
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

		// Run processors (filter, enrich, transform)
		for _, proc := range p.config.Processors {
			findings = proc.Process(findings)
		}

		// Accumulate partial errors across iterations.
		for name, detErr := range detResult.Errors {
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

// detect runs all detectors and collects findings.
func (p *Pipeline) detect(ctx context.Context) (*PartialResult, error) {
	if p.config.GracefulDegradation {
		return p.DetectPartial(ctx)
	}

	var findings []finding.Finding
	var err error

	if p.config.ParallelDetectors {
		findings, err = p.detectParallel(ctx)
	} else {
		findings, err = p.detectSequential(ctx)
	}

	if err != nil {
		return nil, err
	}

	return &PartialResult{Findings: findings}, nil //nolint:exhaustruct
}

// filterActive returns non-suppressed findings, calling OnFinding for each.
func (p *Pipeline) filterActive(findings []finding.Finding) []finding.Finding {
	var result []finding.Finding
	for _, f := range findings {
		if !f.IsSuppressed() {
			result = append(result, f)
			p.notifyFinding(f)
		}
	}
	return result
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

// runOneDetector executes a single detector, recording metrics and filtering
// suppressed findings. It returns the active findings or an error.
func (p *Pipeline) runOneDetector(ctx context.Context, d Detector) ([]finding.Finding, error) {
	start := time.Now()
	findings, err := d.Detect(ctx)
	elapsed := time.Since(start)

	if err != nil {
		p.recordDetectorMetrics(d.Name(), elapsed, nil)

		return nil, fmt.Errorf("detector %s: %w", d.Name(), err)
	}

	p.recordDetectorMetrics(d.Name(), elapsed, findings)

	return p.filterActive(findings), nil
}

// detectSequential runs detectors one at a time.
func (p *Pipeline) detectSequential(ctx context.Context) ([]finding.Finding, error) {
	var allFindings []finding.Finding

	for _, d := range p.detectors {
		if err := CheckCanceled(ctx); err != nil {
			return nil, err
		}

		findings, err := p.runOneDetector(ctx, d)
		if err != nil {
			return nil, err
		}

		allFindings = append(allFindings, findings...)
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
			findings, err := p.runOneDetector(ctx, d)
			if err != nil {
				return err
			}

			mu.Lock()
			allFindings = append(allFindings, findings...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
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

	iter.Applied = len(applied)

	if p.config.OnFix != nil {
		for _, f := range applied {
			p.config.OnFix(f, true)
		}
	}

	return nil
}

// applyDirectFixes applies deterministic fixes to files.
func (p *Pipeline) applyDirectFixes(
	ctx context.Context,
	fixes []finding.Finding,
) ([]finding.Finding, error) {
	var applier *FixApplier
	if len(p.config.FixProviders) > 0 {
		applier = NewFixApplierWithProviders(p.rootDir, p.config.FixProviders...)
	} else {
		applier = NewFixApplier(p.rootDir)
	}
	defer func() { _ = applier.Close() }()

	applied, appliedFixes, err := applier.ApplyWithDetails(ctx, fixes)
	if err != nil {
		return nil, err
	}

	if p.metrics != nil {
		for range applied {
			p.metrics.RecordFix()
		}
	}

	return appliedFixes, nil
}
