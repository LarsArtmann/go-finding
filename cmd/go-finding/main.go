package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
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
	flag.IntVar(&maxIter, "max-iterations", 5, "maximum pipeline iterations")
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

	var cfg pipeline.Config
	if configFile != "" {
		cfg, err = loadConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = pipeline.Config{
			MaxIterations:     maxIter,
			ParallelDetectors: parallel,
			VerifyAfterFix:    verify,
			Timeout:           timeout,
			Metrics:           pipeline.NewMetrics(),
		}
	}

	fmt.Fprintf(os.Stderr, "go-finding CLI v0.1.0\n")
	fmt.Fprintf(os.Stderr, "Config: dir=%s format=%s severity=%s\n", dir, format, sev)

	_ = cfg
	fmt.Fprintln(os.Stderr, "No detectors configured yet. See examples/ for detector implementations.")
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

func loadConfig(path string) (pipeline.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return pipeline.Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg pipelineConfigFile

	switch ext := filepath.Ext(path); ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return pipeline.Config{}, fmt.Errorf("parse YAML config: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return pipeline.Config{}, fmt.Errorf("parse config: %w", err)
		}
	}

	return cfg.toPipelineConfig(), nil
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
	timeout := 10 * time.Minute
	if c.Timeout != "" {
		if d, err := time.ParseDuration(c.Timeout); err == nil {
			timeout = d
		}
	}

	maxIter := c.MaxIterations
	if maxIter == 0 {
		maxIter = 5
	}

	return pipeline.Config{
		MaxIterations:     maxIter,
		ParallelDetectors: c.ParallelDetectors,
		VerifyAfterFix:    c.VerifyAfterFix,
		Timeout:           timeout,
		Metrics:           pipeline.NewMetrics(),
	}
}

func init() {
	log.SetFlags(0)
}
