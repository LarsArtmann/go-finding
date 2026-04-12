// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// OnFinding is called for each finding found.
	OnFinding func(f finding.Finding)
	// OnFix is called when a fix is applied.
	OnFix func(f finding.Finding, applied bool)
	// OnIteration is called at the end of each iteration.
	OnIteration func(iter int, findings []finding.Finding)
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
}

// New creates a new Pipeline with the given configuration.
func New(config Config, rootDir string, detectors ...Detector) *Pipeline {
	return &Pipeline{
		config:    config,
		detectors: detectors,
		rootDir:   rootDir,
		findings:  make([]finding.Finding, 0),
	}
}

// Run executes the pipeline until stable or max iterations reached.
func (p *Pipeline) Run(ctx context.Context) (*Result, error) {
	if p.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.Timeout)
		defer cancel()
	}

	result := &Result{
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
		findings, err := p.detect(ctx)
		if err != nil {
			return result, fmt.Errorf("iteration %d: detect: %w", p.iterations+1, err)
		}
		iter.FindingsFound = len(findings)

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

		// Apply fixes
		if len(triage.Direct) > 0 {
			applied, err := p.applyDirectFixes(ctx, triage.Direct)
			if err != nil {
				return result, fmt.Errorf("iteration %d: apply fixes: %w", p.iterations+1, err)
			}
			iter.Applied = applied
		}

		result.Iterations = append(result.Iterations, iter)
		p.iterations++

		if p.config.OnIteration != nil {
			p.config.OnIteration(p.iterations, findings)
		}
	}

	result.TotalIterations = len(result.Iterations)
	result.FinalFindingCount = len(p.findings)
	return result, nil
}

// Result contains the outcome of running the pipeline.
type Result struct {
	Stable            bool
	TotalIterations   int
	Iterations        []Iteration
	FinalFindingCount int
	Error             error
}

// Iteration represents one loop through the pipeline.
type Iteration struct {
	Number        int
	FindingsFound int
	DirectFixes   int
	SuggestFixes  int
	NoFix         int
	Applied       int
	Failed        int
}

// detect runs all detectors and collects findings.
func (p *Pipeline) detect(ctx context.Context) ([]finding.Finding, error) {
	if p.config.ParallelDetectors {
		return p.detectParallel(ctx)
	}
	return p.detectSequential(ctx)
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

		for _, f := range findings {
			if !f.IsSuppressed() {
				allFindings = append(allFindings, f)
				if p.config.OnFinding != nil {
					p.config.OnFinding(f)
				}
			}
		}
	}

	return allFindings, nil
}

// detectParallel runs detectors concurrently using errgroup.
func (p *Pipeline) detectParallel(ctx context.Context) ([]finding.Finding, error) {
	var mu sync.Mutex
	var allFindings []finding.Finding

	g, ctx := errgroup.WithContext(ctx)

	for _, d := range p.detectors {
		d := d // capture loop variable
		g.Go(func() error {
			findings, err := d.Detect(ctx)
			if err != nil {
				return fmt.Errorf("detector %s: %w", d.Name(), err)
			}

			mu.Lock()
			for _, f := range findings {
				if !f.IsSuppressed() {
					allFindings = append(allFindings, f)
					if p.config.OnFinding != nil {
						p.config.OnFinding(f)
					}
				}
			}
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
				return applied, fmt.Errorf("backup %s: %w", path, err)
			}
		}

		// Apply fixes
		if err := a.applyToFile(path, fileFixes); err != nil {
			// Restore from backup on error
			if a.backupEnabled {
				_ = a.restore(path)
			}
			return applied, fmt.Errorf("apply to %s: %w", path, err)
		}

		applied += len(fileFixes)
	}

	return applied, nil
}

// backup creates a backup of the given file.
func (a *FixApplier) backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	backupPath := filepath.Join(a.backupDir, filepath.Base(path)+".bak")
	if err := os.MkdirAll(a.backupDir, 0755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	a.backups[path] = backupPath
	return nil
}

// restore restores a file from its backup.
func (a *FixApplier) restore(path string) error {
	backupPath, ok := a.backups[path]
	if !ok {
		return fmt.Errorf("no backup for %s", path)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("restore file: %w", err)
	}

	return nil
}

// applyToFile applies fixes to a single file.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) error {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// For now, we only support simple text replacement based on BeforeCode/AfterCode
	// This is a simplified implementation - in practice you'd want to use
	// proper AST-based transformations or at least line/column-aware replacements
	result := string(content)
	for _, f := range fixes {
		if f.BeforeCode != "" && f.AfterCode != "" {
			result = replaceCode(result, f.BeforeCode, f.AfterCode)
		}
	}

	// Write updated content
	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// replaceCode performs a simple text replacement.
// This is a placeholder - real implementations should use AST-aware replacements.
func replaceCode(content, before, after string) string {
	return replaceAllString(content, before, after)
}

// replaceAllString replaces all occurrences of old with new in s.
func replaceAllString(s, old, new string) string {
	return stringReplaceAll(s, old, new)
}

// stringReplaceAll is a simple implementation of strings.ReplaceAll for Go < 1.12.
func stringReplaceAll(s, old, new string) string {
	if old == "" {
		return s
	}
	result := ""
	start := 0
	for {
		idx := 0
		for i := start; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == 0 && (start > len(s)-len(old) || s[start:start+len(old)] != old) {
			break
		}
		result += s[start:idx] + new
		start = idx + len(old)
	}
	result += s[start:]
	return result
}
