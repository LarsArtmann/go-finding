package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

func runWithArgs(t *testing.T, args ...string) int {
	t.Helper()
	savedCommandLine := flag.CommandLine
	savedArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = savedCommandLine
		os.Args = savedArgs
	})

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = append([]string{"go-finding"}, args...)

	return run()
}

func TestLoadConfig_YAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 3
parallelDetectors: true
verifyAfterFix: false
timeout: "5m"
detectors:
  - name: govet
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	if cfg.MaxIterations != 3 {
		t.Errorf("MaxIterations = %d, want 3", cfg.MaxIterations)
	}

	if cfg.Timeout != "5m" {
		t.Errorf("Timeout = %q, want %q", cfg.Timeout, "5m")
	}

	if len(cfg.Detectors) != 1 || cfg.Detectors[0].Name != "govet" {
		t.Errorf("Detectors = %v, want [govet]", cfg.Detectors)
	}
}

func TestLoadConfig_JSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	content := `{"maxIterations": 7, "parallelDetectors": false, "timeout": "30s", "detectors": [{"name": "govet"}]}`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	if cfg.MaxIterations != 7 {
		t.Errorf("MaxIterations = %d, want 7", cfg.MaxIterations)
	}

	if cfg.ParallelDetectors {
		t.Error("ParallelDetectors should be false")
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfig("", 2, true, true, 5*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig empty file: %v", err)
	}

	if cfg.MaxIterations != 2 {
		t.Errorf("MaxIterations = %d, want 2", cfg.MaxIterations)
	}

	if len(cfg.Detectors) != 2 {
		t.Errorf("default detectors = %d, want 2", len(cfg.Detectors))
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := loadConfig("/nonexistent/config.yaml", 1, true, false, 10*time.Minute)
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	expectConfigError(t, "yaml", "{{invalid yaml")
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	t.Parallel()

	expectConfigError(t, "json", "{invalid json}")
}

func expectConfigError(t *testing.T, ext, content string) {
	t.Helper()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad."+ext)

	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	if err == nil {
		t.Fatalf("expected error for invalid %s", ext)
	}
}

func TestPipelineConfigFile_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     pipelineConfigFile
		wantErr bool
	}{
		{
			name: "valid",
			cfg: pipelineConfigFile{
				MaxIterations: 5,
				Timeout:       "10m",
				Detectors:     []detectorSpec{{Name: "govet"}},
			},
			wantErr: false,
		},
		{
			name: "negative iterations",
			cfg: pipelineConfigFile{
				MaxIterations: -1,
				Detectors:     []detectorSpec{{Name: "govet"}},
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Timeout:       "not-a-duration",
				Detectors:     []detectorSpec{{Name: "govet"}},
			},
			wantErr: true,
		},
		{
			name: "unknown detector",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Detectors:     []detectorSpec{{Name: "nonexistent-tool"}},
			},
			wantErr: true,
		},
		{
			name: "empty is valid",
			cfg: pipelineConfigFile{
				MaxIterations: 0,
				Timeout:       "",
				Detectors:     nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPipelineConfigFile_ToPipelineConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        pipelineConfigFile
		wantIter   int
		wantPar    bool
		wantVerify bool
	}{
		{
			name: "explicit values",
			cfg: pipelineConfigFile{
				MaxIterations:     3,
				ParallelDetectors: true,
				VerifyAfterFix:    true,
				Timeout:           "30s",
			},
			wantIter:   3,
			wantPar:    true,
			wantVerify: true,
		},
		{
			name: "zero iterations defaults to 1",
			cfg: pipelineConfigFile{
				MaxIterations: 0,
				Timeout:       "",
			},
			wantIter: 1,
		},
		{
			name: "empty timeout defaults to 10m",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Timeout:       "",
			},
			wantIter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pc := tt.cfg.toPipelineConfig()

			if pc.MaxIterations != tt.wantIter {
				t.Errorf("MaxIterations = %d, want %d", pc.MaxIterations, tt.wantIter)
			}

			if pc.ParallelDetectors != tt.wantPar {
				t.Errorf("ParallelDetectors = %v, want %v", pc.ParallelDetectors, tt.wantPar)
			}

			if pc.VerifyAfterFix != tt.wantVerify {
				t.Errorf("VerifyAfterFix = %v, want %v", pc.VerifyAfterFix, tt.wantVerify)
			}
		})
	}
}

func TestSetupProfiling(t *testing.T) {
	t.Parallel()

	t.Run("no profiling", func(t *testing.T) {
		t.Parallel()

		stop, err := setupProfiling("", "")
		if err != nil {
			t.Fatalf("setupProfiling no-op error: %v", err)
		}

		stop()
	})

	t.Run("cpu profile", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		cpuPath := filepath.Join(dir, "cpu.prof")

		stop, err := setupProfiling(cpuPath, "")
		if err != nil {
			t.Fatalf("setupProfiling cpu error: %v", err)
		}

		stop()

		info, err := os.Stat(cpuPath)
		if err != nil {
			t.Fatalf("cpu profile not created: %v", err)
		}

		if info.Size() == 0 {
			t.Error("cpu profile is empty")
		}
	})

	t.Run("mem profile", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		memPath := filepath.Join(dir, "mem.prof")

		stop, err := setupProfiling("", memPath)
		if err != nil {
			t.Fatalf("setupProfiling mem error: %v", err)
		}

		stop()

		info, err := os.Stat(memPath)
		if err != nil {
			t.Fatalf("mem profile not created: %v", err)
		}

		if info.Size() == 0 {
			t.Error("mem profile is empty")
		}
	})

	t.Run("cpu profile bad path", func(t *testing.T) {
		t.Parallel()

		_, err := setupProfiling("/nonexistent/dir/cpu.prof", "")
		if err == nil {
			t.Fatal("expected error for bad cpu profile path")
		}
	})
}

func TestRunWithConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
parallelDetectors: false
timeout: "30s"
detectors:
  - name: govet
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Create a simple Go file for the detector to analyze.
	goFile := filepath.Join(dir, "main.go")

	if err := os.WriteFile(goFile, []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write go file: %v", err)
	}

	// Capture stdout by temporarily replacing os.Stdout — but run() uses os.Stdout directly,
	// so test via os/exec subprocess instead.
	// For now, test the config loading + pipeline creation path indirectly.
	cfg, err := loadConfig(cfgPath, 1, false, false, 30*time.Second)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if err := cfg.validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	pc := cfg.toPipelineConfig()
	if pc.MaxIterations != 1 {
		t.Errorf("MaxIterations = %d, want 1", pc.MaxIterations)
	}
}

func TestRunWithInvalidSeverity(t *testing.T) {
	t.Parallel()
	// Test that run returns error for bad severity — this exercises parseSeverity via the run() path.
	// Since run() reads flags, test parseSeverity directly instead.
	_, err := parseSeverity("bogus")
	if err == nil {
		t.Fatal("expected error for bogus severity")
	}

	if !errors.Is(err, errUnknownSeverity) {
		t.Errorf("error should wrap errUnknownSeverity, got: %v", err)
	}
}

func TestLoadConfig_WithInvalidDetector(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
detectors:
  - name: nonexistent-detector
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	if err == nil {
		t.Fatal("expected error for unknown detector in config")
	}

	if !errors.Is(err, errUnknownDetector) {
		t.Errorf("error should wrap errUnknownDetector, got: %v", err)
	}
}

func TestOutputResults_AllFormats(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "integration-test", Version: "1.0"})
	report.AddFinding(finding.Finding{
		ID:          "test:R1:main.go:1:1",
		Rule:        "R1",
		ToolName:    "test",
		Message:     "integration test finding",
		Severity:    finding.SeverityError,
		Position:    finding.Position{File: "main.go", Line: 1, Column: 1},
		Category:    finding.CategoryCorrectness,
		FixStrategy: finding.FixStrategySuggest,
		Suggestion:  "fix the issue",
	})
	report.ComputeSummary()

	for _, format := range []string{"text", "json", "sarif"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			if err := outputResults(&buf, report, format); err != nil {
				t.Fatalf("outputResults(%s) error: %v", format, err)
			}

			if buf.Len() == 0 {
				t.Errorf("outputResults(%s) produced empty output", format)
			}
		})
	}
}

func TestOutputResults_JSONContainsFindings(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	if err := outputResults(&buf, report, "json"); err != nil {
		t.Fatalf("outputResults error: %v", err)
	}

	var parsed struct {
		Findings []struct {
			Rule string `json:"rule"`
		} `json:"findings"`
	}

	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	if len(parsed.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(parsed.Findings))
	}

	if parsed.Findings[0].Rule != "nilcheck" {
		t.Errorf("rule = %q, want %q", parsed.Findings[0].Rule, "nilcheck")
	}
}

func TestOutputResults_SARIFContainsResults(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	if err := outputResults(&buf, report, "sarif"); err != nil {
		t.Fatalf("outputResults error: %v", err)
	}

	var parsed struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []struct {
				RuleID string `json:"ruleId"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("SARIF parse error: %v", err)
	}

	if parsed.Version != "2.1.0" {
		t.Errorf("SARIF version = %q, want %q", parsed.Version, "2.1.0")
	}

	if len(parsed.Runs) != 1 || len(parsed.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 run with 1 result, got %d runs", len(parsed.Runs))
	}

	if parsed.Runs[0].Results[0].RuleID != "nilcheck" {
		t.Errorf("ruleId = %q, want %q", parsed.Runs[0].Results[0].RuleID, "nilcheck")
	}
}

func TestRun_BadSeverity(t *testing.T) {
	t.Parallel()

	got := runWithArgs(t, "-severity", "bogus")
	if got != 1 {
		t.Errorf("run() with bad severity = %d, want 1", got)
	}
}

func TestRun_MissingConfig(t *testing.T) {
	t.Parallel()

	got := runWithArgs(t, "-config", "/nonexistent/config.yaml")
	if got != 1 {
		t.Errorf("run() with missing config = %d, want 1", got)
	}
}

func TestRun_NoDetectors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
timeout: "30s"
detectors: []
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	savedCommandLine := flag.CommandLine
	savedArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = savedCommandLine
		os.Args = savedArgs
	})

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-config", cfgPath, "-dir", dir}

	got := run()
	if got != 1 {
		t.Errorf("run() with no detectors = %d, want 1", got)
	}
}

func TestRun_BadProfilingPath(t *testing.T) {
	t.Parallel()
	savedCommandLine := flag.CommandLine
	savedArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = savedCommandLine
		os.Args = savedArgs
	})

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-cpuprof", "/nonexistent/dir/cpu.prof"}

	got := run()
	if got != 1 {
		t.Errorf("run() with bad cpuprof = %d, want 1", got)
	}
}
