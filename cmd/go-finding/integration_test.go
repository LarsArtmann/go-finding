package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

func detectorSpecs(names ...string) []detectorSpec {
	specs := make([]detectorSpec, len(names))
	for i, name := range names {
		specs[i] = detectorSpec{Name: name}
	}
	return specs
}

func govetConfig(iterations int, timeout ...string) pipelineConfigFile {
	cfg := pipelineConfigFile{Detectors: detectorSpecs("govet"), MaxIterations: iterations}
	if len(timeout) > 0 {
		cfg.Timeout = timeout[0]
	}
	return cfg
}

func runWithArgs(t *testing.T, args ...string) int {
	t.Helper()
	saveRestoreFlags(t)

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = append([]string{"go-finding"}, args...)

	return run()
}

func writeConfig(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

// saveRestoreFlags saves the current flag state and registers a cleanup to restore it.
func saveRestoreFlags(t *testing.T) {
	t.Helper()
	savedCommandLine := flag.CommandLine
	savedArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = savedCommandLine
		os.Args = savedArgs
	})
}

func requireOutputResults(t *testing.T, w *bytes.Buffer, report *finding.Report, format string) {
	t.Helper()
	if err := outputResults(w, report, format); err != nil {
		t.Fatalf("outputResults error: %v", err)
	}
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
	writeConfig(t, cfgPath, []byte(content))

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
	g := NewWithT(t)
	t.Parallel()
		g := NewWithT(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	content := `{"maxIterations": 7, "parallelDetectors": false, "timeout": "30s", "detectors": [{"name": "govet"}]}`
	writeConfig(t, cfgPath, []byte(content))

	cfg, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	g.Expect(cfg.MaxIterations).To(Equal(7))
	g.Expect(cfg.ParallelDetectors).To(BeFalse())
}

func TestLoadConfig_NoFile(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
		g := NewWithT(t)

	cfg, err := loadConfig("", 2, true, true, 5*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig empty file: %v", err)
	}

	g.Expect(cfg.MaxIterations).To(Equal(2))
	g.Expect(cfg.Detectors).To(HaveLen(2))
}

func TestLoadConfig_MissingFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	_, err := loadConfig("/nonexistent/config.yaml", 1, true, false, 10*time.Minute)
	g.Expect(err).To(HaveOccurred())
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
	g := NewWithT(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad."+ext)

	writeConfig(t, cfgPath, []byte(content))

	_, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	g.Expect(err).To(HaveOccurred())
}

func TestPipelineConfigFile_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     pipelineConfigFile
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     govetConfig(5, "10m"),
			wantErr: false,
		},
		{
			name:    "negative iterations",
			cfg:     govetConfig(-1),
			wantErr: true,
		},
		{
			name: "invalid timeout",
			cfg: pipelineConfigFile{
				Detectors:     detectorSpecs("govet"),
				MaxIterations: 1,
				Timeout:       "not-a-duration",
			},
			wantErr: true,
		},
		{
			name: "unknown detector",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Detectors:     detectorSpecs("nonexistent-tool"),
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
			g := NewWithT(t)

			err := tt.cfg.validate()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
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
			name: "zero iterations defaults to pipeline default",
			cfg: pipelineConfigFile{
				MaxIterations: 0,
				Timeout:       "",
			},
			wantIter: 5,
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

//nolint:paralleltest // manipulates global pprof CPU profile state and os.Stderr
func TestSetupProfiling(t *testing.T) {
	g := NewWithT(t)

	t.Run("no profiling", func(t *testing.T) {
		stop, err := setupProfiling("", "")
		if err != nil {
			t.Fatalf("setupProfiling no-op error: %v", err)
		}

		stop()
	})

	t.Run("cpu profile", func(t *testing.T) {
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

		g.Expect(info.Size()).NotTo(BeZero())
	})

	t.Run("mem profile", func(t *testing.T) {
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

		g.Expect(info.Size()).NotTo(BeZero())
	})

	t.Run("cpu profile bad path", func(t *testing.T) {
		_, err := setupProfiling("/nonexistent/dir/cpu.prof", "")
		if err == nil {
			t.Fatal("expected error for bad cpu profile path")
		}
	})
}

func TestRunWithConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
parallelDetectors: false
timeout: "30s"
detectors:
  - name: govet
`
	writeConfig(t, cfgPath, []byte(content))

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
	g.Expect(pc.MaxIterations).To(Equal(1))
}

func TestRunWithInvalidSeverity(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	// Test that run returns error for bad severity — this exercises parseSeverity via the run() path.
	// Since run() reads flags, test parseSeverity directly instead.
	_, err := parseSeverity("bogus")
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errUnknownSeverity)).To(BeTrue())
}

func TestLoadConfig_WithInvalidDetector(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
detectors:
  - name: nonexistent-detector
`
	writeConfig(t, cfgPath, []byte(content))

	_, err := loadConfig(cfgPath, 1, true, false, 10*time.Minute)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errUnknownDetector)).To(BeTrue())
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
		Position:    finding.Pos("main.go", 1, 1),
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
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "json")

	var parsed struct {
		Findings []struct {
			Rule string `json:"rule"`
		} `json:"findings"`
	}

	requireJSON(t, &buf, &parsed, "JSON parse error")

	g.Expect(parsed.Findings).To(HaveLen(1))
	g.Expect(parsed.Findings[0].Rule).To(Equal("nilcheck"))
}

func TestOutputResults_SARIFContainsResults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "sarif")

	var parsed struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []struct {
				RuleID string `json:"ruleId"`
			} `json:"results"`
		} `json:"runs"`
	}

	requireJSON(t, &buf, &parsed, "SARIF parse error")

	g.Expect(parsed.Version).To(Equal("2.1.0"))
	g.Expect(parsed.Runs).To(HaveLen(1))
	g.Expect(parsed.Runs[0].Results).To(HaveLen(1))
	g.Expect(parsed.Runs[0].Results[0].RuleID).To(Equal("nilcheck"))
}

func assertRunFails(t *testing.T, args ...string) {
	t.Helper()
	if got := runWithArgs(t, args...); got != 1 {
		t.Errorf("run() with args %v = %d, want 1", args, got)
	}
}

func requireJSON[T any](t *testing.T, buf *bytes.Buffer, parsed *T, context string) {
	t.Helper()
	if err := json.Unmarshal(buf.Bytes(), parsed); err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}

//nolint:paralleltest // manipulates global flag state
func TestRun_NegativeMaxIterations(t *testing.T) {
	assertRunFails(t, "-max-iterations", "-1", "-dir", t.TempDir())
}

//nolint:paralleltest // manipulates global flag state
func TestRun_PipelineRunError(t *testing.T) {
	g := NewWithT(t)
	name := uniqueDetName("broken-fix")

	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			"broken",
			func(_ context.Context) ([]finding.Finding, error) {
				return []finding.Finding{{
					ID:          "broken:R1:main.go:1:1",
					Rule:        "R1",
					ToolName:    "broken",
					Message:     "broken fix",
					Severity:    finding.SeverityError,
					Position:    finding.Position{File: "nonexistent.go", Line: 1, Column: 1},
					FixStrategy: finding.FixStrategyDirect,
					BeforeCode:  "old",
					AfterCode:   "new",
				}}, nil
			},
		)
	})
	g.Expect(err).NotTo(HaveOccurred())

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfig(t, cfgPath, []byte(
		"maxIterations: 1\ntimeout: 30s\ndetectors:\n  - name: "+name+"\n",
	))

	assertRunFails(t, "-config", cfgPath, "-dir", dir)
}

//nolint:paralleltest // manipulates global flag state
func TestRun_MetricsOutput(t *testing.T) {
	g := NewWithT(t)
	name := uniqueDetName("always-find")

	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			"find",
			func(_ context.Context) ([]finding.Finding, error) {
				return []finding.Finding{{
					ID:       "find:R1:main.go:1:1",
					Rule:     "R1",
					ToolName: "find",
					Message:  "test finding",
					Severity: finding.SeverityError,
					Position: finding.Position{File: "main.go", Line: 1, Column: 1},
				}}, nil
			},
		)
	})
	g.Expect(err).NotTo(HaveOccurred())

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfig(t, cfgPath, []byte(
		"maxIterations: 1\ntimeout: 30s\ndetectors:\n  - name: "+name+"\n",
	))

	var buf bytes.Buffer
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	saveRestoreFlags(t)
	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-config", cfgPath, "-dir", dir}

	got := run()

	_ = w.Close()
	os.Stderr = old
	_, _ = io.Copy(&buf, r)

	g.Expect(got).To(Equal(0))
	g.Expect(buf.String()).To(ContainSubstring("Metrics:"))
}

//nolint:paralleltest // manipulates global flag state via assertRunFails
func TestRun_BadSeverity(t *testing.T) {
	assertRunFails(t, "-severity", "bogus")
}

//nolint:paralleltest // manipulates global flag state via assertRunFails
func TestRun_MissingConfig(t *testing.T) {
	assertRunFails(t, "-config", "/nonexistent/config.yaml")
}

//nolint:paralleltest // manipulates global flag state
func TestRun_NoDetectors(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
timeout: "30s"
detectors: []
`
	writeConfig(t, cfgPath, []byte(content))

	saveRestoreFlags(t)

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-config", cfgPath, "-dir", dir}

	got := run()
	if got != 1 {
		t.Errorf("run() with no detectors = %d, want 1", got)
	}
}
