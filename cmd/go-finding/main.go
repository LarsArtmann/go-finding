// Package main implements the go-finding CLI tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"slices"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

var version = finding.Version // overridden via -ldflags "-X main.version=..."

const defaultMaxIterations = 5

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
		_, _ = fmt.Fprintln(os.Stdout, version)

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

	if err := writeOutput(report, format, outputFile); err != nil {
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
		for _, v := range slices.Backward(stopFuncs) {
			v()
		}
	}, nil
}

func filterBySeverity(findings []finding.Finding, minSeverity finding.Severity) []finding.Finding {
	return finding.Filter(findings, finding.BySeverityAtLeast(minSeverity))
}
