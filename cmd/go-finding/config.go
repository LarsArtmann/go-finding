package main

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/go-faster/yaml"
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/cmd/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/pipeline"
)

type pipelineConfigFile struct {
	MaxIterations     int               `json:"maxIterations"     yaml:"maxIterations"`
	ParallelDetectors bool              `json:"parallelDetectors" yaml:"parallelDetectors"`
	VerifyAfterFix    bool              `json:"verifyAfterFix"    yaml:"verifyAfterFix"`
	Timeout           string            `json:"timeout"           yaml:"timeout"`
	DetectorTimeouts  map[string]string `json:"detectorTimeouts"  yaml:"detectorTimeouts"`
	Detectors         []detectorSpec    `json:"detectors"         yaml:"detectors"`
	// FilterGenerated enables filtering of findings from auto-generated Go source files.
	FilterGenerated  bool     `json:"filterGenerated"  yaml:"filterGenerated"`
	FilterGenTypes   string   `json:"filterGenTypes"   yaml:"filterGenTypes"`
	GeneratedExclude []string `json:"generatedExclude" yaml:"generatedExclude"`
	GeneratedInclude []string `json:"generatedInclude" yaml:"generatedInclude"`
	// ByteLevelConflictDetection enables precise byte-level conflict detection during triage.
	ByteLevelConflictDetection bool `json:"byteLevelConflictDetection" yaml:"byteLevelConflictDetection"`
	// FixRollbackAllFiles opts into all-or-nothing fix application: when a file
	// fails, every file modified earlier in the run is rolled back. By default
	// only the failing file is restored and applied fixes stay on disk.
	FixRollbackAllFiles bool `json:"fixRollbackAllFiles" yaml:"fixRollbackAllFiles"`
	// FixProviders enables named fix providers (e.g., "go-ast") for domain-specific edits.
	FixProviders []string `json:"fixProviders" yaml:"fixProviders"`
	// FlightRecorder enables Go execution trace recording. This is an alternative
	// to the -trace CLI flag for YAML/JSON config users.
	FlightRecorder *flightRecorderFileConfig `json:"flightRecorder" yaml:"flightRecorder"`
}

type flightRecorderFileConfig struct {
	Enabled            bool   `json:"enabled"            yaml:"enabled"`
	OutputDir          string `json:"outputDir"          yaml:"outputDir"`
	SlowStageThreshold string `json:"slowStageThreshold" yaml:"slowStageThreshold"`
	MinAge             string `json:"minAge"             yaml:"minAge"`
	MaxBytes           uint64 `json:"maxBytes"           yaml:"maxBytes"`
}

type detectorSpec struct {
	Name string `json:"name" yaml:"name"`
}

// DetectorName constants referenced from the internal detectors package.
const (
	detectorNameGovet       = detectors.DetectorNameGovet
	detectorNameStaticcheck = detectors.DetectorNameStaticcheck
)

// Sentinel errors for CLI validation.
var (
	errUnknownSeverity    = errors.New("unknown severity")
	errInvalidConfig      = errors.New("invalid config")
	errUnknownDetector    = errors.New("unknown detector")
	errDetectorRegistered = errors.New("detector already registered")
)

func loadConfig(
	configFile string,
	maxIter int,
	parallel, verify bool,
	timeout time.Duration,
) (pipelineConfigFile, error) {
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return pipelineConfigFile{}, fmt.Errorf("reading config %q: %w", configFile, err)
		}

		var cfg pipelineConfigFile

		switch ext := filepath.Ext(configFile); ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf(
					"parsing YAML config %q: %w",
					configFile, err,
				)
			}
		default:
			if err := json.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf(
					"parsing config %q: %w",
					configFile, err,
				)
			}
		}

		if err := cfg.validate(); err != nil {
			return pipelineConfigFile{}, fmt.Errorf("invalid config %q: %w", configFile, err)
		}

		return cfg, nil
	}

	return pipelineConfigFile{ //nolint:exhaustruct_v5
		MaxIterations:     maxIter,
		ParallelDetectors: parallel,
		VerifyAfterFix:    verify,
		Timeout:           timeout.String(),
		Detectors: []detectorSpec{
			{Name: detectorNameGovet},
			{Name: detectorNameStaticcheck},
		},
	}, nil
}

func (c pipelineConfigFile) validate() error {
	if c.MaxIterations < 0 {
		return fmt.Errorf(
			"%w: maxIterations must be >= 0, got %d",
			errInvalidConfig,
			c.MaxIterations,
		)
	}

	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout %q: %w", c.Timeout, err)
		}
	}

	for _, d := range c.Detectors {
		if _, ok := lookupDetectorBuilder(d.Name); !ok {
			return fmt.Errorf("%w: %q (available: %s)", errUnknownDetector, d.Name,
				strings.Join(availableDetectorNames(), ", "))
		}
	}

	for _, name := range c.FixProviders {
		if _, ok := lookupFixProvider(name); !ok {
			return fmt.Errorf("%w: %q (available: %s)", ErrUnknownFixProvider, name,
				strings.Join(availableFixProviderNames(), ", "))
		}
	}

	if c.FlightRecorder != nil {
		if c.FlightRecorder.SlowStageThreshold != "" {
			if _, err := time.ParseDuration(c.FlightRecorder.SlowStageThreshold); err != nil {
				return fmt.Errorf("invalid flightRecorder.slowStageThreshold %q: %w",
					c.FlightRecorder.SlowStageThreshold, err)
			}
		}

		if c.FlightRecorder.MinAge != "" {
			if _, err := time.ParseDuration(c.FlightRecorder.MinAge); err != nil {
				return fmt.Errorf("invalid flightRecorder.minAge %q: %w",
					c.FlightRecorder.MinAge, err)
			}
		}
	}

	return nil
}

func parseSeverity(s string) (finding.Severity, error) {
	sev, err := finding.ParseSeverity(s)
	if err != nil {
		return finding.SeverityWarning, fmt.Errorf(
			"%w %q (use: %s, %s, %s, %s)",
			errUnknownSeverity, s,
			finding.SeverityInfo,
			finding.SeverityWarning,
			finding.SeverityError,
			finding.SeverityCritical,
		)
	}

	return sev, nil
}

func (c pipelineConfigFile) toPipelineConfig() (pipeline.Config, error) {
	t := pipeline.DefaultTimeout

	if c.Timeout != "" {
		d, err := time.ParseDuration(c.Timeout)
		if err != nil {
			return pipeline.Config{}, fmt.Errorf("parse timeout %q: %w", c.Timeout, err)
		}

		t = d
	}

	maxIter := c.MaxIterations
	if maxIter == 0 {
		maxIter = pipeline.DefaultMaxIterations
	}

	detectorTimeouts := make(map[string]time.Duration, len(c.DetectorTimeouts))
	for name, durStr := range c.DetectorTimeouts {
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return pipeline.Config{}, fmt.Errorf(
				"parse detector timeout %q for %q: %w",
				durStr,
				name,
				err,
			)
		}

		detectorTimeouts[name] = d
	}

	return pipeline.Config{ //nolint:exhaustruct_v5
		MaxIterations:              maxIter,
		ParallelDetectors:          c.ParallelDetectors,
		VerifyAfterFix:             c.VerifyAfterFix,
		Timeout:                    t,
		Metrics:                    pipeline.NewMetrics(),
		DetectorTimeouts:           detectorTimeouts,
		ByteLevelConflictDetection: c.ByteLevelConflictDetection,
		FixRollbackAllFiles:        c.FixRollbackAllFiles,
	}, nil
}

func writeOutput(report *finding.Report, format, outputFile string, includeSuppressed bool) error {
	if outputFile == "" {
		return outputResults(os.Stdout, report, format, includeSuppressed)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outputFile, err)
	}

	writeErr := outputResults(f, report, format, includeSuppressed)
	closeErr := f.Close()

	if writeErr != nil {
		return writeErr
	}

	if closeErr != nil {
		return fmt.Errorf("closing output file %s: %w", outputFile, closeErr)
	}

	return nil
}

// supportedFormats is the complete set of CLI output format names.
var supportedFormats = []string{"text", "markdown", "csv", "tsv", "json", "sarif"}

func outputResults(w io.Writer, report *finding.Report, format string, includeSuppressed bool) error {
	if !slices.Contains(supportedFormats, format) {
		return fmt.Errorf("unsupported format %q (supported: %s)", format, strings.Join(supportedFormats, ", "))
	}

	if isGoOutputFormat(format) {
		if err := renderGoOutput(w, report, format); err != nil {
			return fmt.Errorf("writing %s: %w", format, err)
		}

		return nil
	}

	switch format {
	case "json":
		out, err := report.PrettyJSON()
		if err != nil {
			return fmt.Errorf("serializing JSON: %w", err)
		}

		if _, err := fmt.Fprintln(w, out); err != nil {
			return fmt.Errorf("writing JSON: %w", err)
		}
	case "sarif":
		var opts []finding.SARIFOption
		if includeSuppressed {
			opts = append(opts, finding.WithIncludeSuppressed())
		}

		out, err := report.ToSARIFWithOpts(opts...)
		if err != nil {
			return fmt.Errorf("serializing SARIF: %w", err)
		}

		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return fmt.Errorf("writing SARIF: %w", err)
		}
	default:
		outputText(w, report)
	}

	return nil
}

func outputText(w io.Writer, report *finding.Report) {
	findings := report.FindingsSnapshot()

	if len(findings) == 0 {
		_, _ = fmt.Fprintln(w, "No findings.")

		return
	}

	if err := finding.FormatText(w, findings); err != nil {
		_, _ = fmt.Fprintf(w, "warning: %v\n", err)
	}

	_, _ = fmt.Fprintf(w, "\n%d finding(s)\n", len(findings))
	if report.Summary.Total > 0 {
		_, _ = fmt.Fprintf(
			w, "  By severity: %d info, %d warning, %d error, %d critical\n",
			report.Summary.BySeverity[finding.SeverityInfo],
			report.Summary.BySeverity[finding.SeverityWarning],
			report.Summary.BySeverity[finding.SeverityError],
			report.Summary.BySeverity[finding.SeverityCritical],
		)
	}
}

// fatalf prints an error message and returns exit code 1.
func fatalf(op string, err error) int {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", op, err)

	return 1
}
