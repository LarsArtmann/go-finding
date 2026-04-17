// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

// Name implements Detector. Returns "anonymous" — use NamedDetectorFunc for a custom name.
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
	// Wrap detectors with retry if configured.
	if config.Retry != nil {
		wrapped := make([]Detector, len(detectors))
		for i, d := range detectors {
			wrapped[i] = NewRetryDetector(d, *config.Retry)
		}
		detectors = wrapped
	}

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
	p.findings = p.findings[:0]
	p.iterations = 0

	if p.metrics != nil {
		p.metrics.SetStart(time.Now())
		defer func() { p.metrics.SetEnd(time.Now()) }()
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
		if isContextDone(ctx) {
			return result, ctx.Err()
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
		iter.suggest = triage.Suggest
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
	suggest       []finding.Finding
}

// Findings returns all findings discovered in this iteration.
func (it Iteration) Findings() []finding.Finding {
	return it.findings
}

// SuggestedFindings returns findings that have FixStrategySuggest.
func (it Iteration) SuggestedFindings() []finding.Finding {
	return it.suggest
}

// detect runs all detectors and collects findings.
func (p *Pipeline) detect(ctx context.Context) ([]finding.Finding, error) {
	if p.config.GracefulDegradation {
		result, err := p.DetectPartial(ctx)
		if err != nil {
			return nil, err
		}
		// Partial detector errors are non-fatal with GracefulDegradation.
		// Attach them to PipelineResult if callers need them.
		return result.Findings, nil
	}
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
			p.notifyFinding(f)
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

		start := time.Now()
		findings, err := d.Detect(ctx)
		elapsed := time.Since(start)
		if err != nil {
			return nil, fmt.Errorf("detector %s: %w", d.Name(), err)
		}

		if p.metrics != nil {
			p.metrics.RecordDetector(d.Name(), elapsed, len(findings))
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
			start := time.Now()
			findings, err := d.Detect(ctx)
			elapsed := time.Since(start)
			if err != nil {
				return fmt.Errorf("detector %s: %w", d.Name(), err)
			}

			if p.metrics != nil {
				p.metrics.RecordDetector(d.Name(), elapsed, len(findings))
			}

			mu.Lock()
			allFindings = append(allFindings, findings...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var filtered []finding.Finding
	for i := range allFindings {
		if !allFindings[i].IsSuppressed() {
			filtered = append(filtered, allFindings[i])
			p.notifyFinding(allFindings[i])
		}
	}

	return filtered, nil
}

// TriageResult holds findings categorized by fix strategy.
type TriageResult struct {
	Direct  []finding.Finding
	Suggest []finding.Finding
	None    []finding.Finding
}

// triage categorizes findings by their fix strategy.
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
		case finding.FixStrategySuggest:
			result.Suggest = append(result.Suggest, f)
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

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir       string
	backupEnabled bool
	backupDir     string
	backups       map[string]string // original -> backup path
	backupsMu     sync.Mutex
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

// ioErrorAt creates an IO error with position info.
func ioErrorAt(msg string, err error, path string) error {
	return finding.NewIOError(msg, err).WithPosition(finding.Position{File: path})
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
		count, err := a.applyToFile(path, fileFixes)
		if err != nil {
			// Restore from backup on error
			if a.backupEnabled {
				_ = a.restore(path)
			}
			return applied, finding.NewConflictError(fmt.Sprintf("apply to %s", path), err)
		}

		applied += count
	}

	return applied, nil
}

// fileHash returns the hex-encoded SHA256 hash of s.
func fileHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// backup creates a backup of the given file.
func (a *FixApplier) backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return ioErrorAt("read file for backup", err, path)
	}

	backupPath := filepath.Join(a.backupDir, fmt.Sprintf("%x.bak", fileHash(path)))
	if err := os.MkdirAll(a.backupDir, 0750); err != nil {
		return finding.NewIOError("create backup dir", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return ioErrorAt("write backup", err, path)
	}

	a.backupsMu.Lock()
	a.backups[path] = backupPath
	a.backupsMu.Unlock()
	return nil
}

// restore restores a file from its backup.
func (a *FixApplier) restore(path string) error {
	backupPath, ok := func() (string, bool) {
		a.backupsMu.Lock()
		defer a.backupsMu.Unlock()
		p, exists := a.backups[path]
		return p, exists
	}()
	if !ok {
		return finding.NewInternalError(fmt.Sprintf("no backup for %s", path), nil)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return ioErrorAt("read backup", err, path)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return ioErrorAt("restore file", err, path)
	}

	return nil
}

// applyToFile applies fixes to a single file.
// When a finding has a Range with valid end position, it uses line-based replacement
// targeting the exact line range. Otherwise it falls back to string replacement.
// Fixes are sorted descending by position so earlier replacements don't shift later ones.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, ioErrorAt("read file", err, path)
	}

	lines := strings.Split(string(content), "\n")
	applied := 0

	// Partition: range-based fixes (apply first, sorted descending) vs string-based.
	var rangeFixes, stringFixes []finding.Finding
	for _, f := range fixes {
		if f.BeforeCode == "" && f.AfterCode == "" {
			continue
		}
		if f.Range != nil && f.Range.HasEnd() && f.Range.Start.Line > 0 && f.Range.End.Line > 0 {
			rangeFixes = append(rangeFixes, f)
		} else if f.BeforeCode != "" && f.AfterCode != "" {
			stringFixes = append(stringFixes, f)
		}
	}

	// Sort range fixes descending so earlier edits don't shift later line numbers.
	slices.SortFunc(rangeFixes, func(a, b finding.Finding) int {
		if a.Range.Start.Line != b.Range.Start.Line {
			return b.Range.Start.Line - a.Range.Start.Line
		}
		return b.Range.Start.Column - a.Range.Start.Column
	})

	// Apply range-based fixes.
	for _, f := range rangeFixes {
		startIdx := f.Range.Start.Line - 1 // 0-indexed
		endIdx := f.Range.End.Line - 1

		if startIdx < 0 || startIdx >= len(lines) {
			continue
		}
		if endIdx >= len(lines) {
			endIdx = len(lines) - 1
		}

		// Verify BeforeCode is present in the range if set.
		rangeContent := strings.Join(lines[startIdx:endIdx+1], "\n")
		if f.BeforeCode != "" {
			if !strings.Contains(rangeContent, f.BeforeCode) {
				continue
			}
			// Targeted replacement: swap BeforeCode→AfterCode within the range,
			// preserving surrounding content like indentation.
			replaced := strings.Replace(rangeContent, f.BeforeCode, f.AfterCode, 1)
			replacementLines := strings.Split(replaced, "\n")
			newLines := make([]string, 0, len(lines)-(endIdx-startIdx+1)+len(replacementLines))
			newLines = append(newLines, lines[:startIdx]...)
			newLines = append(newLines, replacementLines...)
			newLines = append(newLines, lines[endIdx+1:]...)
			lines = newLines
		} else if f.AfterCode != "" {
			// Full replacement: replace entire line range with AfterCode.
			replacement := make([]string, 0, len(lines)-(endIdx-startIdx+1)+1)
			replacement = append(replacement, lines[:startIdx]...)
			replacement = append(replacement, f.AfterCode)
			replacement = append(replacement, lines[endIdx+1:]...)
			lines = replacement
		}
		applied++
	}

	// Apply string-based fixes (fallback, position-independent).
	if len(stringFixes) > 0 {
		joined := strings.Join(lines, "\n")
		for _, f := range stringFixes {
			newContent := strings.Replace(joined, f.BeforeCode, f.AfterCode, 1)
			if newContent != joined {
				joined = newContent
				applied++
			}
		}
		lines = strings.Split(joined, "\n")
	}

	if applied == 0 {
		return 0, nil
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600); err != nil {
		return 0, ioErrorAt("write file", err, path)
	}

	return applied, nil
}
