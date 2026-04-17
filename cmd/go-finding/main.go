// Package main implements the go-finding CLI tool.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"slices"
	"time"

	"github.com/larsartmann/go-finding"
	det "github.com/larsartmann/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/pipeline"
	"gopkg.in/yaml.v3"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		dir        string
		format     string
		minSev     string
		maxIter    int
		parallel   bool
		verify     bool
		timeout    time.Duration
		configFile string
		cpuprof    string
		memprof    string
	)

	flag.StringVar(&dir, "dir", ".", "root directory to analyze")
	flag.StringVar(&format, "format", "text", "output format: text, json, sarif")
	flag.StringVar(&minSev, "severity", "info", "minimum severity: info, warning, error, critical")
	flag.IntVar(&maxIter, "max-iterations", 1, "maximum pipeline iterations")
	flag.BoolVar(&parallel, "parallel", true, "run detectors in parallel")
	flag.BoolVar(&verify, "verify", false, "verify fixes by re-running detectors")
	flag.DurationVar(&timeout, "timeout", 10*time.Minute, "pipeline timeout")
	flag.StringVar(&configFile, "config", "", "YAML/JSON config file path")
	flag.StringVar(&cpuprof, "cpuprof", "", "write CPU profile to file")
	flag.StringVar(&memprof, "memprof", "", "write memory profile to file")
	flag.Parse()

	if cpuprof != "" {
		f, err := os.Create(cpuprof)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating CPU profile: %v\n", err)

			return 1
		}

		defer func() { _ = f.Close() }()

		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)

			return 1
		}

		defer pprof.StopCPUProfile()
	}

	defer func() {
		if memprof != "" {
			f, err := os.Create(memprof)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating memory profile: %v\n", err)

				return
			}

			defer func() { _ = f.Close() }()

			if err := pprof.WriteHeapProfile(f); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing heap profile: %v\n", err)
			}
		}
	}()

	sev, err := parseSeverity(minSev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", "parsing severity", err)

		return 1
	}

	var cfg pipelineConfigFile

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error %s: %v\n", "reading config", err)

			return 1
		}

		switch ext := filepath.Ext(configFile); ext {
		case ".yaml", ".yml":
			err := yaml.Unmarshal(data, &cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error %s: %v\n", "parsing YAML config", err)

				return 1
			}
		default:
			err := json.Unmarshal(data, &cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error %s: %v\n", "parsing config", err)

				return 1
			}
		}

		if err := cfg.validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error %s: %v\n", "invalid config", err)

			return 1
		}
	} else {
		cfg = pipelineConfigFile{
			MaxIterations:     maxIter,
			ParallelDetectors: parallel,
			VerifyAfterFix:    verify,
			Timeout:           timeout.String(),
			Detectors:         []detectorSpec{{Name: "govet"}, {Name: "staticcheck"}},
		}
	}

	detectorList := buildDetectors(cfg.Detectors, dir)
	if len(detectorList) == 0 {
		fmt.Fprintln(
			os.Stderr,
			"No detectors configured. Use -config or the default govet+staticcheck detectors.",
		)

		return 1
	}

	pipelineCfg := cfg.toPipelineConfig()
	pipelineCfg.GracefulDegradation = true
	p := pipeline.New(pipelineCfg, dir, detectorList...)

	ctx, cancel := context.WithTimeout(context.Background(), pipelineCfg.Timeout)
	defer cancel()

	fmt.Fprintf(
		os.Stderr,
		"go-finding v0.1.0: analyzing %s with %d detector(s)\n",
		dir,
		len(detectorList),
	)

	result, err := p.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", "running pipeline", err)

		return 1
	}

	var allFindings []finding.Finding
	for _, iter := range result.Iterations {
		allFindings = append(allFindings, iter.Findings()...)
	}

	filtered := filterBySeverity(allFindings, sev)

	report := finding.NewReport(finding.ToolInfo{
		Name:    "go-finding",
		Version: "0.1.0",
	})
	report.AddFindings(filtered)
	report.ComputeSummary()

	if err := outputResults(os.Stdout, report, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", "writing output", err)

		return 1
	}

	fmt.Fprintf(os.Stderr, "\nDone: %d findings (%d iterations, stable=%v)\n",
		len(filtered), result.TotalIterations, result.Stable)

	return 0
}

func parseSeverity(s string) (finding.Severity, error) {
	switch s {
	case "info":
		return finding.SeverityInfo, nil
	case "warning":
		return finding.SeverityWarning, nil
	case "error":
		return finding.SeverityError, nil
	case "critical":
		return finding.SeverityCritical, nil
	default:
		return finding.SeverityWarning, fmt.Errorf("%w %q (use: info, warning, error, critical)", errUnknownSeverity, s)
	}
}

func filterBySeverity(findings []finding.Finding, minSeverity finding.Severity) []finding.Finding {
	levels := map[finding.Severity]int{
		finding.SeverityInfo:     0,
		finding.SeverityWarning:  1,
		finding.SeverityError:    2,
		finding.SeverityCritical: 3,
	}
	minLevel := levels[minSeverity]

	return slices.DeleteFunc(findings, func(f finding.Finding) bool {
		return levels[f.Severity] < minLevel
	})
}

func buildDetectors(specs []detectorSpec, dir string) []pipeline.Detector {
	var result []pipeline.Detector

	for _, spec := range specs {
		switch spec.Name {
		case "govet":
			result = append(result, det.NewGoVetDetector(dir))
		case "staticcheck":
			result = append(result, det.NewStaticcheckDetector(dir))
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown detector %q, skipping\n", spec.Name)
		}
	}

	return result
}

func outputResults(w *os.File, report *finding.Report, format string) error {
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
	default:
		outputText(w, report)
	}

	return nil
}

func outputText(w *os.File, report *finding.Report) {
	if len(report.Findings) == 0 {
		_, _ = fmt.Fprintln(w, "No findings.")

		return
	}

	for _, f := range report.Findings {
		_, _ = fmt.Fprintf(
			w,
			"%s: [%s] %s: %s\n",
			f.Position.String(),
			f.Severity,
			f.Rule,
			f.Message,
		)
		if f.Suggestion != "" {
			_, _ = fmt.Fprintf(w, "  Suggestion: %s\n", f.Suggestion)
		}
	}

	_, _ = fmt.Fprintf(w, "\n%d finding(s)\n", len(report.Findings))
	if report.Summary.Total > 0 {
		_, _ = fmt.Fprintf(w, "  By severity: %d info, %d warning, %d error, %d critical\n",
			report.Summary.BySeverity[finding.SeverityInfo],
			report.Summary.BySeverity[finding.SeverityWarning],
			report.Summary.BySeverity[finding.SeverityError],
			report.Summary.BySeverity[finding.SeverityCritical],
		)
	}
}

type pipelineConfigFile struct {
	MaxIterations     int            `json:"maxIterations"     yaml:"maxIterations"`
	ParallelDetectors bool           `json:"parallelDetectors" yaml:"parallelDetectors"`
	VerifyAfterFix    bool           `json:"verifyAfterFix"    yaml:"verifyAfterFix"`
	Timeout           string         `json:"timeout"           yaml:"timeout"`
	Detectors         []detectorSpec `json:"detectors"         yaml:"detectors"`
}

type detectorSpec struct {
	Name string            `json:"name" yaml:"name"`
	Args map[string]string `json:"args" yaml:"args"`
}

// Sentinel errors for CLI validation.
var (
	errUnknownSeverity = errors.New("unknown severity")
	errInvalidConfig   = errors.New("invalid config")
	errUnknownDetector = errors.New("unknown detector")
)

//nolint:gochecknoglobals // CLI lookup table for validated detector names
var knownDetectors = map[string]bool{
	"govet":       true,
	"staticcheck": true,
}

func (c pipelineConfigFile) validate() error {
	if c.MaxIterations < 0 {
		return fmt.Errorf("%w: maxIterations must be >= 0, got %d", errInvalidConfig, c.MaxIterations)
	}

	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout %q: %w", c.Timeout, err)
		}
	}

	for _, d := range c.Detectors {
		if !knownDetectors[d.Name] {
			return fmt.Errorf("%w %q (available: govet, staticcheck)", errUnknownDetector, d.Name)
		}
	}

	return nil
}

func (c pipelineConfigFile) toPipelineConfig() pipeline.Config {
	t := 10 * time.Minute

	if c.Timeout != "" {
		if d, err := time.ParseDuration(c.Timeout); err == nil {
			t = d
		}
	}

	maxIter := c.MaxIterations
	if maxIter == 0 {
		maxIter = 1
	}

	return pipeline.Config{
		MaxIterations:     maxIter,
		ParallelDetectors: c.ParallelDetectors,
		VerifyAfterFix:    c.VerifyAfterFix,
		Timeout:           t,
		Metrics:           pipeline.NewMetrics(),
	}
}

//nolint:gochecknoinits // CLI log flags setup
func init() {
	log.SetFlags(0)
}
