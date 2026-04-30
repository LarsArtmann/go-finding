// Package main implements the go-finding CLI tool.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
	det "github.com/larsartmann/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/pipeline"
	"gopkg.in/yaml.v3"
)

var version = finding.Version // overridden via -ldflags "-X main.version=..."

func main() {
	os.Exit(run())
}

// fatalf prints an error message and returns exit code 1.
func fatalf(op string, err error) int {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", op, err)

	return 1
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
		showVer    bool
		outputFile string
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
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.StringVar(&outputFile, "output", "", "write output to file (default: stdout)")
	flag.Parse()

	if showVer {
		fmt.Fprintln(os.Stdout, version)

		return 0
	}

	stopProf, err := setupProfiling(cpuprof, memprof)
	if err != nil {
		return 1
	}

	defer stopProf()

	sev, err := parseSeverity(minSev)
	if err != nil {
		return fatalf("parsing severity", err)
	}

	cfg, err := loadConfig(configFile, maxIter, parallel, verify, timeout)
	if err != nil {
		return fatalf("loading config", err)
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
	p, err := pipeline.New(pipelineCfg, dir, detectorList...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pipeline config: %v\n", err)

		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), pipelineCfg.Timeout)
	defer cancel()

	fmt.Fprintf(
		os.Stderr,
		"go-finding v%s: analyzing %s with %d detector(s)\n",
		version,
		dir,
		len(detectorList),
	)

	result, err := p.Run(ctx)
	if err != nil {
		return fatalf("running pipeline", err)
	}

	var allFindings []finding.Finding
	for _, iter := range result.Iterations {
		allFindings = append(allFindings, iter.Findings()...)
	}

	filtered := filterBySeverity(allFindings, sev)

	report := finding.NewReport(finding.ToolInfo{
		Name:    "go-finding",
		Version: version,
	})
	report.AddFindings(filtered)
	report.ComputeSummary()

	w := os.Stdout

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fatalf("creating output file", err)
		}

		defer func() { _ = f.Close() }()
		w = f
	}

	if err := outputResults(w, report, format); err != nil {
		return fatalf("writing output", err)
	}

	fmt.Fprintf(os.Stderr, "\nDone: %d findings (%d iterations, stable=%v)\n",
		len(filtered), result.TotalIterations, result.Stable)

	if result.Metrics.TotalDuration > 0 {
		fmt.Fprintf(os.Stderr, "Metrics: %v total, %d fixes, %d detector(s)\n",
			result.Metrics.TotalDuration,
			result.Metrics.FixesApplied,
			len(result.Metrics.DetectorTimes))
	}

	return 0
}

func setupProfiling(cpuprof, memprof string) (func(), error) {
	var stopFuncs []func()

	if cpuprof != "" {
		f, err := os.Create(cpuprof)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating CPU profile: %v\n", err)

			return nil, fmt.Errorf("create CPU profile: %w", err)
		}

		stopFuncs = append(stopFuncs, func() { _ = f.Close() })

		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()

			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)

			return nil, fmt.Errorf("start CPU profile: %w", err)
		}

		stopFuncs = append(stopFuncs, pprof.StopCPUProfile)
	}

	if memprof != "" {
		stopFuncs = append(stopFuncs, func() {
			f, err := os.Create(memprof)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating memory profile: %v\n", err)

				return
			}

			defer func() { _ = f.Close() }()

			if err := pprof.WriteHeapProfile(f); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing heap profile: %v\n", err)
			}
		})
	}

	return func() {
		for i := len(stopFuncs) - 1; i >= 0; i-- {
			stopFuncs[i]()
		}
	}, nil
}

func loadConfig(
	configFile string,
	maxIter int,
	parallel, verify bool,
	timeout time.Duration,
) (pipelineConfigFile, error) {
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return pipelineConfigFile{}, fmt.Errorf("reading config: %w", err)
		}

		var cfg pipelineConfigFile

		switch ext := filepath.Ext(configFile); ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf("parsing YAML config: %w", err)
			}
		default:
			if err := json.Unmarshal(data, &cfg); err != nil {
				return pipelineConfigFile{}, fmt.Errorf("parsing config: %w", err)
			}
		}

		if err := cfg.validate(); err != nil {
			return pipelineConfigFile{}, fmt.Errorf("invalid config: %w", err)
		}

		return cfg, nil
	}

	return pipelineConfigFile{
		MaxIterations:     maxIter,
		ParallelDetectors: parallel,
		VerifyAfterFix:    verify,
		Timeout:           timeout.String(),
		Detectors:         []detectorSpec{{Name: "govet"}, {Name: "staticcheck"}},
	}, nil
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
		return finding.SeverityWarning, fmt.Errorf(
			"%w %q (use: info, warning, error, critical)",
			errUnknownSeverity,
			s,
		)
	}
}

func filterBySeverity(findings []finding.Finding, minSeverity finding.Severity) []finding.Finding {
	return finding.Filter(findings, finding.BySeverityAtLeast(minSeverity))
}

func buildDetectors(specs []detectorSpec, dir string) []pipeline.Detector {
	var result []pipeline.Detector

	for _, spec := range specs {
		builder, ok := lookupDetectorBuilder(spec.Name)
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: unknown detector %q, skipping\n", spec.Name)

			continue
		}

		result = append(result, builder(dir))
	}

	return result
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
	Name string `json:"name" yaml:"name"`
}

// Sentinel errors for CLI validation.
var (
	errUnknownSeverity    = errors.New("unknown severity")
	errInvalidConfig      = errors.New("invalid config")
	errUnknownDetector    = errors.New("unknown detector")
	errDetectorRegistered = errors.New("detector already registered")
)

var (
	knownDetectorBuildersMu sync.RWMutex
	knownDetectorBuilders   = map[string]func(string) pipeline.Detector{
		"govet":       det.NewGoVetDetector,
		"staticcheck": det.NewStaticcheckDetector,
	}
)

// RegisterDetector registers a custom detector builder by name.
// It is safe for concurrent use. Returns an error if the name is already registered.
func RegisterDetector(name string, builder func(string) pipeline.Detector) error {
	knownDetectorBuildersMu.Lock()
	defer knownDetectorBuildersMu.Unlock()

	if _, exists := knownDetectorBuilders[name]; exists {
		return fmt.Errorf("%w: %q", errDetectorRegistered, name)
	}

	knownDetectorBuilders[name] = builder

	return nil
}

func lookupDetectorBuilder(name string) (func(string) pipeline.Detector, bool) {
	knownDetectorBuildersMu.RLock()
	defer knownDetectorBuildersMu.RUnlock()

	b, ok := knownDetectorBuilders[name]

	return b, ok
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
		maxIter = pipeline.DefaultConfig().MaxIterations
	}

	//nolint:exhaustruct
	return pipeline.Config{
		MaxIterations:     maxIter,
		ParallelDetectors: c.ParallelDetectors,
		VerifyAfterFix:    c.VerifyAfterFix,
		Timeout:           t,
		Metrics:           pipeline.NewMetrics(),
	}
}
