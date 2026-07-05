package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
	"golang.org/x/sync/errgroup"
)

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
// If a per-detector timeout is configured, it takes precedence over the global timeout.
func (p *Pipeline) runOneDetector(ctx context.Context, d Detector) ([]finding.Finding, error) {
	if timeout, ok := p.config.DetectorTimeouts[d.Name()]; ok && timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

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
		err := CheckCanceled(ctx)
		if err != nil {
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

	err := g.Wait()
	if err != nil {
		return nil, fmt.Errorf("parallel detection: %w", err)
	}

	return allFindings, nil
}

// triage categorizes findings using the configured TriageFunc (or default).
func (p *Pipeline) triage(findings []finding.Finding) *TriageResult {
	fn := p.config.TriageFunc
	if fn == nil {
		fn = DefaultTriageFunc
	}

	return fn(findings)
}

// DefaultTriageFunc categorizes findings using HasFix() as the canonical source of truth.
// - IsAutoFixable() → Direct (auto-apply via FixEngine)
// - HasFix() but not auto-fixable → Suggest (display suggestion)
// - No fix available → None.
func DefaultTriageFunc(findings []finding.Finding) *TriageResult {
	result := &TriageResult{
		Direct:  make([]finding.Finding, 0),
		Suggest: make([]finding.Finding, 0),
		None:    make([]finding.Finding, 0),
	}

	for _, f := range findings {
		if f.IsAutoFixable() {
			result.Direct = append(result.Direct, f)
		} else if f.HasFix() {
			result.Suggest = append(result.Suggest, f)
		} else {
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

	var (
		safeFixes      []finding.Finding
		providerErrors []error
	)

	if p.config.ByteLevelConflictDetection {
		engine := p.byteConflictEngine()
		safeFixes, providerErrors = p.filterByFileEdits(ctx, fixes, engine)
	} else {
		safeFixes = FilterConflictingFixes(fixes)
	}

	iter.Conflicts = len(fixes) - len(safeFixes)

	if len(providerErrors) > 0 {
		p.log(
			ctx, "provider errors during conflict detection",
			slog.Int("errors", len(providerErrors)),
		)
	}

	if iter.Conflicts > 0 {
		p.log(
			ctx, "conflicts detected",
			slog.Int("total", len(fixes)),
			slog.Int("conflicts", iter.Conflicts),
			slog.Int("safe", len(safeFixes)),
		)

		for _, c := range AnalyzeConflicts(fixes) {
			if p.config.OnFix != nil {
				p.config.OnFix(c.Finding, false)
			}
		}
	}

	if len(safeFixes) == 0 {
		return nil
	}

	applied, shiftMaps, err := p.applyDirectFixes(ctx, safeFixes)
	if err != nil {
		return fmt.Errorf("apply fixes: %w", err)
	}

	iter.Applied = len(applied)

	// Shift remaining findings' positions and ranges based on applied edits.
	for file, shiftMap := range shiftMaps {
		for i := range iter.findings {
			f := &iter.findings[i]
			if f.Position.File == finding.FilePath(file) {
				f.Position = shiftMap.ShiftedPosition(f.Position)
				f.Range = shiftMap.ShiftedRange(f.Range)
			}
		}
	}

	if p.config.OnFix != nil {
		appliedSet := make(map[string]struct{}, len(applied))

		for _, f := range applied {
			appliedSet[f.Key()] = struct{}{}
		}

		for _, f := range safeFixes {
			if _, ok := appliedSet[f.Key()]; ok {
				p.config.OnFix(f, true)
			} else {
				p.config.OnFix(f, false)
			}
		}
	}

	return nil
}

// applyDirectFixes applies deterministic fixes to files and returns the applied findings
// and a per-file line shift map for updating remaining findings' line numbers.
func (p *Pipeline) applyDirectFixes(
	ctx context.Context,
	fixes []finding.Finding,
) ([]finding.Finding, map[string]*LineShiftMap, error) {
	applied, appliedFixes, shiftMaps, err := p.applier.ApplyWithShiftMap(ctx, fixes)
	if err != nil {
		return nil, nil, err
	}

	if p.metrics != nil && applied > 0 {
		p.metrics.RecordFixes(uint(applied))
	}

	return appliedFixes, shiftMaps, nil
}

// byteConflictEngine returns a FixEngine with custom providers if configured,
// otherwise the default engine with standard text-based providers.
func (p *Pipeline) byteConflictEngine() *FixEngine {
	if len(p.config.FixProviders) > 0 {
		return NewFixEngineWithProviders(p.config.FixProviders...)
	}

	return NewFixEngine()
}

// filterByFileEdits groups fixes by file, reads each file's content,
// and uses byte-level FilterConflictingEdits for precise conflict detection.
func (p *Pipeline) filterByFileEdits(
	_ context.Context,
	fixes []finding.Finding,
	engine *FixEngine,
) ([]finding.Finding, []error) {
	byFile := make(map[finding.FilePath][]finding.Finding, len(fixes))
	for _, f := range fixes {
		byFile[f.Position.File] = append(byFile[f.Position.File], f)
	}

	var (
		result    []finding.Finding
		allErrors []error
	)

	for file, fileFixes := range byFile {
		fullPath := filepath.Join(p.rootDir, string(file))

		content, err := os.ReadFile(filepath.Clean(fullPath))
		if err != nil {
			result = append(result, fileFixes...)

			continue
		}

		filtered, providerErrs := FilterConflictingEdits(content, fileFixes, engine)
		result = append(result, filtered...)
		allErrors = append(allErrors, providerErrs...)
	}

	return result, allErrors
}
