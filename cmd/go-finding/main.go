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

const toolName = "go-finding"

func main() {
	os.Exit(run())
}

type cliFlags struct {
	dir              string
	format           string
	minSev           string
	maxIter          int
	parallel         bool
	verify           bool
	timeout          time.Duration
	configFile       string
	cpuprof          string
	memprof          string
	showVer          bool
	outputFile       string
	filterGenerated  bool
	filterGenTypes   string
	generatedExclude string
	generatedInclude string
}

func parseFlags() cliFlags {
	var f cliFlags

	flag.StringVar(&f.dir, "dir", ".", "root directory to analyze")
	flag.StringVar(&f.format, "format", "text", "output format: text, markdown, json, sarif")
	flag.StringVar(&f.minSev, "severity", "info", "minimum severity: info, warning, error, critical")
	flag.IntVar(&f.maxIter, "max-iterations", 1, "maximum pipeline iterations")
	flag.BoolVar(&f.parallel, "parallel", true, "run detectors in parallel")
	flag.BoolVar(&f.verify, "verify", false, "verify fixes by re-running detectors")
	flag.DurationVar(&f.timeout, "timeout", pipeline.DefaultTimeout, "pipeline timeout")
	flag.StringVar(&f.configFile, "config", "", "YAML/JSON config file path")
	flag.StringVar(&f.cpuprof, "cpuprof", "", "write CPU profile to file")
	flag.StringVar(&f.memprof, "memprof", "", "write memory profile to file")
	flag.BoolVar(&f.showVer, "version", false, "print version and exit")
	flag.StringVar(&f.outputFile, "output", "", "write output to file (default: stdout)")
	flag.BoolVar(&f.filterGenerated, "filter-generated", false, "filter out findings from auto-generated files")
	flag.StringVar(
		&f.filterGenTypes, "filter-generated-types", "all",
		"comma-separated generator types to filter (all, sqlc, templ, mockgen, protobuf, ...)",
	)
	flag.StringVar(
		&f.generatedExclude, "generated-exclude", "",
		"comma-separated glob patterns for files to exclude from generated filtering",
	)
	flag.StringVar(
		&f.generatedInclude, "generated-include", "",
		"comma-separated glob patterns restricting generated-filtering scope",
	)
	flag.Parse()

	return f
}

func run() int {
	f := parseFlags()

	if f.showVer {
		_, _ = fmt.Fprintln(os.Stdout, version)

		return 0
	}

	stopProf, err := setupProfiling(f.cpuprof, f.memprof)
	if err != nil {
		return 1
	}

	defer stopProf()

	sev, err := parseSeverity(f.minSev)
	if err != nil {
		return fatalf("parsing severity", err)
	}

	cfg, err := loadConfig(f.configFile, f.maxIter, f.parallel, f.verify, f.timeout)
	if err != nil {
		return fatalf("loading config", err)
	}

	detectorList := buildDetectors(cfg.Detectors, f.dir)
	if len(detectorList) == 0 {
		fmt.Fprintln(
			os.Stderr,
			"No detectors configured. Use -config or the default govet+staticcheck detectors.",
		)

		return 1
	}

	pipelineCfg := cfg.toPipelineConfig()
	pipelineCfg.GracefulDegradation = true

	if f.filterGenerated || cfg.FilterGenerated {
		if err := addGeneratedFilter(
			&pipelineCfg, cfg,
			f.filterGenTypes, f.generatedExclude, f.generatedInclude,
		); err != nil {
			return fatalf("configuring generated file filter", err)
		}
	}

	p, err := pipeline.New(pipelineCfg, f.dir, detectorList...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pipeline config: %v\n", err)

		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), pipelineCfg.Timeout)
	defer cancel()

	fmt.Fprintf(
		os.Stderr,
		"%s v%s: analyzing %s with %d detector(s)\n",
		toolName,
		version,
		f.dir,
		len(detectorList),
	)

	result, err := p.Run(ctx)
	if err != nil {
		return fatalf("running pipeline", err)
	}

	return writeResults(result, sev, f.format, f.outputFile)
}

func writeResults(
	result *pipeline.PipelineResult,
	minSev finding.Severity,
	format string,
	outputFile string,
) int {
	var allFindings []finding.Finding
	for _, iter := range result.Iterations {
		allFindings = append(allFindings, iter.Findings()...)
	}

	filtered := filterBySeverity(allFindings, minSev)

	report := finding.NewReport(finding.ToolInfo{
		Name:    toolName,
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

			return nil, fmt.Errorf(
				"create CPU profile %s (memprof=%q): %w",
				cpuprof, memprof, err,
			)
		}

		stopFuncs = append(stopFuncs, func() { _ = f.Close() })

		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()

			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)

			return nil, fmt.Errorf("start CPU profile %s: %w", cpuprof, err)
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
