package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"slices"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/pipeline"
)

func main() {
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
			os.Exit(1)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	defer func() {
		if memprof != "" {
			f, err := os.Create(memprof)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating memory profile: %v\n", err)
				return
			}
			defer f.Close()
			pprof.WriteHeapProfile(f)
		}
	}()

	sev, err := parseSeverity(minSev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var cfg pipelineConfigFile
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}
		switch ext := filepath.Ext(configFile); ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing YAML config: %v\n", err)
				os.Exit(1)
			}
		default:
			if err := json.Unmarshal(data, &cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
				os.Exit(1)
			}
		}
		if err := cfg.validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
			os.Exit(1)
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

	detectors := buildDetectors(cfg.Detectors, dir)
	if len(detectors) == 0 {
		fmt.Fprintln(os.Stderr, "No detectors configured. Use -config or the default govet+staticcheck detectors.")
		os.Exit(1)
	}

	pipelineCfg := cfg.toPipelineConfig()
	pipelineCfg.GracefulDegradation = true
	p := pipeline.New(pipelineCfg, dir, detectors...)

	ctx, cancel := context.WithTimeout(context.Background(), pipelineCfg.Timeout)
	defer cancel()

	fmt.Fprintf(os.Stderr, "go-finding v0.1.0: analyzing %s with %d detector(s)\n", dir, len(detectors))

	result, err := p.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\nDone: %d findings (%d iterations, stable=%v)\n",
		len(filtered), result.TotalIterations, result.Stable)
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
		return finding.SeverityWarning, fmt.Errorf("unknown severity %q (use: info, warning, error, critical)", s)
	}
}

func filterBySeverity(findings []finding.Finding, min finding.Severity) []finding.Finding {
	levels := map[finding.Severity]int{
		finding.SeverityInfo:     0,
		finding.SeverityWarning:  1,
		finding.SeverityError:    2,
		finding.SeverityCritical: 3,
	}
	minLevel := levels[min]
	return slices.DeleteFunc(findings, func(f finding.Finding) bool {
		return levels[f.Severity] < minLevel
	})
}

func buildDetectors(specs []detectorSpec, dir string) []pipeline.Detector {
	var detectors []pipeline.Detector
	for _, spec := range specs {
		switch spec.Name {
		case "govet":
			detectors = append(detectors, detectors.NewGoVetDetector(dir))
		case "staticcheck":
			detectors = append(detectors, detectors.NewStaticcheckDetector(dir))
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown detector %q, skipping\n", spec.Name)
		}
	}
	return detectors
}

func outputResults(w *os.File, report *finding.Report, format string) error {
	switch format {
	case "json":
		out, err := report.PrettyJSON()
		if err != nil {
			return err
		}
		fmt.Fprintln(w, out)
	case "sarif":
		out, err := report.ToSARIF()
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
	default:
		outputText(w, report)
	}
	return nil
}

func outputText(w *os.File, report *finding.Report) {
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "No findings.")
		return
	}
	for _, f := range report.Findings {
		fmt.Fprintf(w, "%s: [%s] %s: %s\n", f.Position.String(), f.Severity, f.Rule, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(w, "  Suggestion: %s\n", f.Suggestion)
		}
	}
	fmt.Fprintf(w, "\n%d finding(s)\n", len(report.Findings))
	if report.Summary.Total > 0 {
		fmt.Fprintf(w, "  By severity: %d info, %d warning, %d error, %d critical\n",
			report.Summary.BySeverity[finding.SeverityInfo],
			report.Summary.BySeverity[finding.SeverityWarning],
			report.Summary.BySeverity[finding.SeverityError],
			report.Summary.BySeverity[finding.SeverityCritical],
		)
	}
}

type pipelineConfigFile struct {
	MaxIterations     int            `json:"maxIterations" yaml:"maxIterations"`
	ParallelDetectors bool           `json:"parallelDetectors" yaml:"parallelDetectors"`
	VerifyAfterFix    bool           `json:"verifyAfterFix" yaml:"verifyAfterFix"`
	Timeout           string         `json:"timeout" yaml:"timeout"`
	Detectors         []detectorSpec `json:"detectors" yaml:"detectors"`
}

type detectorSpec struct {
	Name string            `json:"name" yaml:"name"`
	Args map[string]string `json:"args" yaml:"args"`
}

var knownDetectors = map[string]bool{
	"govet":       true,
	"staticcheck": true,
}

func (c pipelineConfigFile) validate() error {
	if c.MaxIterations < 0 {
		return fmt.Errorf("maxIterations must be >= 0, got %d", c.MaxIterations)
	}
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout %q: %w", c.Timeout, err)
		}
	}
	for _, d := range c.Detectors {
		if !knownDetectors[d.Name] {
			return fmt.Errorf("unknown detector %q (available: govet, staticcheck)", d.Name)
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

func init() {
	log.SetFlags(0)
}
