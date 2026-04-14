// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Name implements Detector.
func (f DetectorFunc) Name() string {
	return "anonymous"
}

// Stage represents a pipeline stage.
// Stages are pure functions that transform input to output.
type Stage func(ctx context.Context) error

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
	// OnFinding is called for each finding found.
	OnFinding func(f finding.Finding)
	// OnFix is called when a fix is applied.
	OnFix func(f finding.Finding, applied bool)
	// OnIteration is called at the end of each iteration.
	OnIteration func(iter int, findings []finding.Finding)
	// Metrics collects timing and count data. If nil, no metrics are collected.
	Metrics *Metrics
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		MaxIterations:     5,
		ParallelDetectors: true,
		Timeout:           10 * time.Minute,
	}
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
func New(config Config, rootDir string, detectors ...Detector) *Pipeline {
	return &Pipeline{
		config:    config,
		detectors: detectors,
		rootDir:   rootDir,
		findings:  make([]finding.Finding, 0),
		metrics:   config.Metrics,
	}
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
func (p *Pipeline) Run(ctx context.Context) (*PipelineResult, error) {
	if p.metrics != nil {
		p.metrics.StartTime = time.Now()
		defer func() { p.metrics.EndTime = time.Now() }()
	}
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	result := &PipelineResult{
		Iterations: make([]Iteration, 0, p.config.MaxIterations),
	}

	for p.iterations < p.config.MaxIterations {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		iter := Iteration{Number: p.iterations + 1}

		// Detect
		detectDone := p.stageTiming("detect")
		findings, err := p.detect(ctx)
		detectDone()
		if err != nil {
			return result, fmt.Errorf("iteration %d: detect: %w", p.iterations+1, err)
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
		iter.NoFix = len(triage.None)

		// Apply fixes (with conflict detection)
		applyDone := p.stageTiming("apply")
		if err := p.applyTriage(ctx, triage.Direct, &iter); err != nil {
			applyDone()
			return result, fmt.Errorf("iteration %d: %w", p.iterations+1, err)
		}
		applyDone()

		result.Iterations = append(result.Iterations, iter)
		p.iterations++

		if p.config.OnIteration != nil {
			p.config.OnIteration(p.iterations, findings)
		}
	}

	result.TotalIterations = len(result.Iterations)

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

	result.FinalFindingCount = len(p.findings)
	return result, nil
}

// collectAllFindings gathers all findings from all iterations for verification.
func (p *Pipeline) collectAllFindings(result *PipelineResult) []finding.Finding {
	seen := make(map[string]bool)
	var all []finding.Finding
	for _, iter := range result.Iterations {
		for _, f := range iter.findings {
			if !seen[f.ID] {
				seen[f.ID] = true
				all = append(all, f)
			}
		}
	}
	if len(all) == 0 {
		return p.findings
	}
	return all
}

// PipelineResult contains the outcome of running the pipeline.
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
}

// detect runs all detectors and collects findings.
func (p *Pipeline) detect(ctx context.Context) ([]finding.Finding, error) {
	if p.config.ParallelDetectors {
		return p.detectParallel(ctx)
	}
	return p.detectSequential(ctx)
}

// addFindings adds non-suppressed findings to the target slice, calling OnFinding if set.
func (p *Pipeline) addFindings(target []finding.Finding, findings []finding.Finding) []finding.Finding {
	for _, f := range findings {
		if !f.IsSuppressed() {
			target = append(target, f)
			if p.config.OnFinding != nil {
				p.config.OnFinding(f)
			}
		}
	}
	return target
}

// detectSequential runs detectors one at a time.
func (p *Pipeline) detectSequential(ctx context.Context) ([]finding.Finding, error) {
	var allFindings []finding.Finding

	for _, d := range p.detectors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		findings, err := d.Detect(ctx)
		if err != nil {
			return nil, fmt.Errorf("detector %s: %w", d.Name(), err)
		}

		allFindings = p.addFindings(allFindings, findings)
	}

	return allFindings, nil
}

// detectParallel runs detectors concurrently using errgroup.
func (p *Pipeline) detectParallel(ctx context.Context) ([]finding.Finding, error) {
	var mu sync.Mutex
	var allFindings []finding.Finding

	g, ctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		g.Go(func() error {
			findings, err := d.Detect(ctx)
			if err != nil {
				return fmt.Errorf("detector %s: %w", d.Name(), err)
			}

			mu.Lock()
			allFindings = p.addFindings(allFindings, findings)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return allFindings, nil
}

// TriageResult holds findings categorized by fix strategy.
type TriageResult struct {
	Direct  []finding.Finding
	Suggest []finding.Finding
	AI      []finding.Finding
	None    []finding.Finding
}

// triage categorizes findings by their fix strategy.
func (p *Pipeline) triage(findings []finding.Finding) *TriageResult {
	result := &TriageResult{
		Direct:  make([]finding.Finding, 0),
		Suggest: make([]finding.Finding, 0),
		AI:      make([]finding.Finding, 0),
		None:    make([]finding.Finding, 0),
	}

	for _, f := range findings {
		switch f.FixStrategy {
		case finding.FixStrategyDirect:
			result.Direct = append(result.Direct, f)
		case finding.FixStrategySuggest:
			result.Suggest = append(result.Suggest, f)
		case finding.FixStrategyAI:
			result.AI = append(result.AI, f)
		default:
			result.None = append(result.None, f)
		}
	}

	return result
}

// applyTriage handles conflict detection and fix application for one iteration.
func (p *Pipeline) applyTriage(ctx context.Context, fixes []finding.Finding, iter *Iteration) error {
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
	return nil
}

// applyDirectFixes applies deterministic fixes to files.
func (p *Pipeline) applyDirectFixes(ctx context.Context, fixes []finding.Finding) (int, error) {
	applier := NewFixApplier(p.rootDir)
	return applier.Apply(ctx, fixes)
}

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir       string
	backupEnabled bool
	backupDir     string
	backups       map[string]string // original -> backup path
}

// NewFixApplier creates a new FixApplier.
func NewFixApplier(rootDir string) *FixApplier {
	return &FixApplier{
		rootDir:       rootDir,
		backupEnabled: true,
		backupDir:     filepath.Join(os.TempDir(), "go-finding-backups"),
		backups:       make(map[string]string),
	}
}

// Apply applies the given fixes to files and returns the number of successful fixes.
func (a *FixApplier) Apply(ctx context.Context, fixes []finding.Finding) (int, error) {
	// Group fixes by file
	byFile := make(map[string][]finding.Finding)
	for _, f := range fixes {
		if f.Position.File == "" {
			continue
		}
		path := filepath.Join(a.rootDir, f.Position.File)
		byFile[path] = append(byFile[path], f)
	}

	applied := 0
	for path, fileFixes := range byFile {
		select {
		case <-ctx.Done():
			return applied, ctx.Err()
		default:
		}

		// Create backup
		if a.backupEnabled {
			if err := a.backup(path); err != nil {
				return applied, finding.NewIOError(fmt.Sprintf("backup %s", path), err)
			}
		}

		// Apply fixes
		if err := a.applyToFile(path, fileFixes); err != nil {
			// Restore from backup on error
			if a.backupEnabled {
				_ = a.restore(path)
			}
			return applied, finding.NewConflictError(fmt.Sprintf("apply to %s", path), err)
		}

		applied += len(fileFixes)
	}

	return applied, nil
}

// backup creates a backup of the given file.
func (a *FixApplier) backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return finding.NewIOError("read file for backup", err).WithPosition(finding.Position{File: path})
	}

	backupPath := filepath.Join(a.backupDir, filepath.Base(path)+".bak")
	if err := os.MkdirAll(a.backupDir, 0750); err != nil {
		return finding.NewIOError("create backup dir", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return finding.NewIOError("write backup", err).WithPosition(finding.Position{File: path})
	}

	a.backups[path] = backupPath
	return nil
}

// restore restores a file from its backup.
func (a *FixApplier) restore(path string) error {
	backupPath, ok := a.backups[path]
	if !ok {
		return finding.NewInternalError(fmt.Sprintf("no backup for %s", path), nil)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return finding.NewIOError("read backup", err).WithPosition(finding.Position{File: path})
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return finding.NewIOError("restore file", err).WithPosition(finding.Position{File: path})
	}

	return nil
}

// applyToFile applies fixes to a single file.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) error {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return finding.NewIOError("read file", err).WithPosition(finding.Position{File: path})
	}

	// For now, we only support simple text replacement based on BeforeCode/AfterCode
	// This is a simplified implementation - in practice you'd want to use
	// proper AST-based transformations or at least line/column-aware replacements
	result := string(content)
	for _, f := range fixes {
		if f.BeforeCode != "" && f.AfterCode != "" {
			result = strings.ReplaceAll(result, f.BeforeCode, f.AfterCode)
		}
	}

	// Write updated content
	if err := os.WriteFile(path, []byte(result), 0600); err != nil {
		return finding.NewIOError("write file", err).WithPosition(finding.Position{File: path})
	}

	return nil
}
