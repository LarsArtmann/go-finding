package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/pprof"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/larsartmann/go-finding"
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
			detectors = append(detectors, pipeline.NamedDetectorFunc("govet", goVetDetector(dir)))
		case "staticcheck":
			detectors = append(detectors, pipeline.NamedDetectorFunc("staticcheck", staticcheckDetector(dir)))
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown detector %q, skipping\n", spec.Name)
		}
	}
	return detectors
}

func goVetDetector(dir string) pipeline.DetectorFunc {
	return func(ctx context.Context) ([]finding.Finding, error) {
		cmd := exec.CommandContext(ctx, "go", "vet", "-json", "./...")
		cmd.Dir = dir
		cmd.Stderr = nil
		out, err := cmd.Output()
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return nil, nil
			}
			return nil, fmt.Errorf("run go vet: %w", err)
		}
		return parseGoVetJSON(out, dir), nil
	}
}

func parseGoVetJSON(data []byte, dir string) []finding.Finding {
	var diagnostics map[string]json.RawMessage
	if err := json.Unmarshal(data, &diagnostics); err != nil {
		return nil
	}

	var findings []finding.Finding
	for pkg, raw := range diagnostics {
		var entries []struct {
			Posn    string `json:"posn"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(raw, &entries); err != nil {
			continue
		}
		for _, e := range entries {
			pos := parsePosn(e.Posn, dir)
			findings = append(findings, finding.Finding{
				ID:          finding.GenerateID("govet", "", pos),
				ToolName:    "govet",
				Message:     e.Message,
				Severity:    finding.SeverityWarning,
				Position:    pos,
				Category:    finding.CategoryCorrectness,
				FixStrategy: finding.FixStrategySuggest,
			})
		}
		_ = pkg
	}
	return findings
}

func parsePosn(posn, dir string) finding.Position {
	parts := strings.SplitN(posn, ":", 4)
	if len(parts) < 2 {
		return finding.Position{File: posn}
	}
	pos := finding.Position{File: parts[0]}
	if dir != "" && !filepath.IsAbs(pos.File) {
		pos.File = filepath.Join(dir, parts[0])
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &pos.Line)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &pos.Column)
	}
	return pos
}

func staticcheckDetector(dir string) pipeline.DetectorFunc {
	return func(ctx context.Context) ([]finding.Finding, error) {
		cmd := exec.CommandContext(ctx, "staticcheck", "-f", "json", "./...")
		cmd.Dir = dir
		cmd.Stderr = nil
		out, err := cmd.Output()
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return parseStaticcheckJSON(out, dir), nil
			}
			return nil, fmt.Errorf("run staticcheck: %w", err)
		}
		return parseStaticcheckJSON(out, dir), nil
	}
}

func parseStaticcheckJSON(data []byte, dir string) []finding.Finding {
	if len(data) == 0 {
		return nil
	}

	var findings []finding.Finding
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			Code     string `json:"code"`
			Severity string `json:"severity"`
			Location struct {
				File   string `json:"file"`
				Line   int    `json:"line"`
				Column int    `json:"column"`
			} `json:"location"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		pos := finding.Position{
			File:   entry.Location.File,
			Line:   entry.Location.Line,
			Column: entry.Location.Column,
		}
		if dir != "" && !filepath.IsAbs(pos.File) {
			pos.File = filepath.Join(dir, entry.Location.File)
		}

		sev := finding.SeverityWarning
		if entry.Severity == "error" {
			sev = finding.SeverityError
		}

		cat := staticcheckCategory(entry.Code)

		findings = append(findings, finding.Finding{
			ID:          finding.GenerateID("staticcheck", entry.Code, pos),
			Rule:        entry.Code,
			ToolName:    "staticcheck",
			Message:     entry.Message,
			Severity:    sev,
			Position:    pos,
			Category:    cat,
			FixStrategy: finding.FixStrategySuggest,
			Confidence:  0.8,
		})
	}
	return findings
}

func staticcheckCategory(code string) finding.Category {
	if len(code) == 0 {
		return finding.CategoryCorrectness
	}
	switch code[0] {
	case 'S', 'Q':
		return finding.CategoryStyle
	case 'U':
		return finding.CategoryUnused
	case 'P', 'R', 'F':
		return finding.CategoryPerformance
	case 'A':
		return finding.CategoryCorrectness
	default:
		return finding.CategoryCorrectness
	}
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
