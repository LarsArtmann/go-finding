package pipeline

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/larsartmann/go-finding"
)

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
	// Processors are chained between detection and triage.
	// Each processor transforms the findings before triage categorizes them.
	Processors []FindingTransformer
	// FixProviders are custom fix providers for domain-specific transformations.
	// If nil, default text-based providers (OffsetProvider, LineProvider, SubstringProvider) are used.
	// Register domain-specific providers (e.g., Go AST, Rust syn) for production accuracy.
	FixProviders []FixProvider
	// DetectorTimeouts configures per-detector timeouts. Map key is detector name,
	// value is the timeout for that detector. Detectors not in the map use the
	// global Timeout.
	DetectorTimeouts map[string]time.Duration
	// Logger provides optional structured logging for pipeline events.
	// If nil, no logging occurs. Use slog.Default() for standard logging.
	Logger *slog.Logger
	// ByteLevelConflictDetection enables precise byte-level conflict detection
	// during triage instead of the default position-based filtering.
	// When true, FixEngine resolves each fix to actual byte offsets and detects
	// overlapping edits. This is more accurate but requires reading file content.
	ByteLevelConflictDetection bool
	// TriageFunc customizes how findings are categorized during triage.
	// If nil, DefaultTriageFunc is used: IsAutoFixable() → Direct,
	// HasFix() → Suggest, else → None.
	TriageFunc TriageFunc
	// StageHooks are called before and after each pipeline stage.
	// Hooks are called in registration order; the first error aborts the pipeline.
	StageHooks []StageHook
}

// TriageFunc categorizes findings into Direct (auto-apply), Suggest (display),
// and None (no fix). Return a TriageResult with the findings distributed.
type TriageFunc func(findings []finding.Finding) *TriageResult

// TriageResult holds findings categorized by fix strategy.
type TriageResult struct {
	Direct  []finding.Finding
	Suggest []finding.Finding
	None    []finding.Finding
}

// DefaultMaxIterations is the default maximum number of pipeline iterations.
const DefaultMaxIterations = 5

// DefaultTimeout is the default pipeline timeout.
const DefaultTimeout = 10 * time.Minute

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	//nolint:exhaustruct
	return Config{
		MaxIterations:     DefaultMaxIterations,
		ParallelDetectors: true,
		Timeout:           DefaultTimeout,
	}
}

// Sentinel validation errors.
var (
	errMaxIterations   = errors.New("max iterations must be >= 0")
	errTimeout         = errors.New("timeout must be >= 0")
	errDetectorTimeout = errors.New("detector timeout must be >= 0")
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
		err := c.Retry.Validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("retry: %w", err))
		}
	}

	for name, d := range c.DetectorTimeouts {
		if d < 0 {
			errs = append(errs, fmt.Errorf("%w: detector %q got %v", errDetectorTimeout, name, d))
		}
	}

	return errors.Join(errs...)
}
