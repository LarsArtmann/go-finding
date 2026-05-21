package main

import (
	"encoding/json"
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
	"github.com/larsartmann/go-finding/pipeline"
)

type pipelineConfigFile struct {
	MaxIterations     int               `json:"maxIterations"     yaml:"maxIterations"`
	ParallelDetectors bool              `json:"parallelDetectors" yaml:"parallelDetectors"`
	VerifyAfterFix    bool              `json:"verifyAfterFix"    yaml:"verifyAfterFix"`
	Timeout           string            `json:"timeout"           yaml:"timeout"`
	DetectorTimeouts  map[string]string `json:"detectorTimeouts"  yaml:"detectorTimeouts"`
	Detectors         []detectorSpec    `json:"detectors"         yaml:"detectors"`
}

type detectorSpec struct {
	Name string `json:"name" yaml:"name"`
}

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
			return pipelineConfigFile{}, fmt.Errorf(
				"reading config %q (maxIter=%d, timeout=%v, parallel=%v, verify=%v): %w",
				configFile, maxIter, timeout, parallel, verify, err,
			)
		}

		var cfg pipelineConfigFile

		switch ext := filepath.Ext(configFile); ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf(
					"parsing YAML config %q (maxIter=%d, timeout=%v, parallel=%v, verify=%v): %w",
					configFile, maxIter, timeout, parallel, verify, err,
				)
			}
		default:
			if err := json.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf(
					"parsing config %q (maxIter=%d, timeout=%v, parallel=%v, verify=%v): %w",
					configFile, maxIter, timeout, parallel, verify, err,
				)
			}
		}

		if err := cfg.validate(); err != nil {
			return pipelineConfigFile{}, fmt.Errorf(
				"invalid config %q (maxIter=%d, timeout=%v, parallel=%v, verify=%v): %w",
				configFile, maxIter, timeout, parallel, verify, err,
			)
		}

		return cfg, nil
	}

	return pipelineConfigFile{ //nolint:exhaustruct
		MaxIterations:     maxIter,
		ParallelDetectors: parallel,
		VerifyAfterFix:    verify,
		Timeout:           timeout.String(),
		Detectors:         []detectorSpec{{Name: "govet"}, {Name: "staticcheck"}},
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
			return fmt.Errorf("%w %q (available: %s)", errUnknownDetector, d.Name,
				strings.Join(slices.Sorted(slices.Values(availableDetectorNames())), ", "))
		}
	}

	return nil
}

func parseSeverity(s string) (finding.Severity, error) {
	sevMap := map[string]finding.Severity{
		finding.SeverityInfo.String():     finding.SeverityInfo,
		finding.SeverityWarning.String():  finding.SeverityWarning,
		finding.SeverityError.String():    finding.SeverityError,
		finding.SeverityCritical.String(): finding.SeverityCritical,
	}

	if sev, ok := sevMap[s]; ok {
		return sev, nil
	}

	return finding.SeverityWarning, fmt.Errorf(
		"%w %q (use: %s, %s, %s, %s)",
		errUnknownSeverity,
		s,
		finding.SeverityInfo,
		finding.SeverityWarning,
		finding.SeverityError,
		finding.SeverityCritical,
	)
}

func (c pipelineConfigFile) toPipelineConfig() pipeline.Config {
	t := pipeline.DefaultTimeout

	if c.Timeout != "" {
		if d, err := time.ParseDuration(c.Timeout); err == nil {
			t = d
		}
	}

	maxIter := c.MaxIterations
	if maxIter == 0 {
		maxIter = pipeline.DefaultMaxIterations
	}

	detectorTimeouts := make(map[string]time.Duration, len(c.DetectorTimeouts))
	for name, durStr := range c.DetectorTimeouts {
		if d, err := time.ParseDuration(durStr); err == nil {
			detectorTimeouts[name] = d
		}
	}

	return pipeline.Config{ //nolint:exhaustruct
		MaxIterations:     maxIter,
		ParallelDetectors: c.ParallelDetectors,
		VerifyAfterFix:    c.VerifyAfterFix,
		Timeout:           t,
		Metrics:           pipeline.NewMetrics(),
		DetectorTimeouts:  detectorTimeouts,
	}
}

func writeOutput(report *finding.Report, format, outputFile string) error {
	w := os.Stdout

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("creating output file %s: %w", outputFile, err)
		}

		defer func() { _ = f.Close() }()
		w = f
	}

	return outputResults(w, report, format)
}

func outputResults(w io.Writer, report *finding.Report, format string) error {
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
		out, err := report.ToSARIF()
		if err != nil {
			return fmt.Errorf("serializing SARIF: %w", err)
		}

		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return fmt.Errorf("writing SARIF: %w", err)
		}
	case "markdown":
		if err := finding.FormatMarkdown(w, report.Findings); err != nil {
			return fmt.Errorf("writing markdown: %w", err)
		}
	default:
		outputText(w, report)
	}

	return nil
}

func outputText(w io.Writer, report *finding.Report) {
	if len(report.Findings) == 0 {
		_, _ = fmt.Fprintln(w, "No findings.")

		return
	}

	if err := finding.FormatText(w, report.Findings); err != nil {
		_, _ = fmt.Fprintln(w, fmt.Sprintf("warning: %v", err))
	}

	_, _ = fmt.Fprintf(w, "\n%d finding(s)\n", len(report.Findings))
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
