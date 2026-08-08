package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

const testToolName = "test"

func detectorSpecs(names ...string) []detectorSpec {
	specs := make([]detectorSpec, len(names))
	for i, name := range names {
		specs[i] = detectorSpec{Name: name}
	}
	return specs
}

func govetConfig(iterations int, timeout ...string) pipelineConfigFile {
	cfg := pipelineConfigFile{
		Detectors:     detectorSpecs(detectorNameGovet),
		MaxIterations: iterations,
	}
	if len(timeout) > 0 {
		cfg.Timeout = timeout[0]
	}
	return cfg
}

func runWithArgs(t *testing.T, args ...string) int {
	t.Helper()
	saveRestoreFlags(t)

	flag.CommandLine = flag.NewFlagSet(toolName, flag.ContinueOnError)
	os.Args = append([]string{toolName}, args...)

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
	if err := outputResults(w, report, format, true); err != nil {
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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	cfg, err := loadConfig("", 2, true, true, 5*time.Minute)
	if err != nil {
		t.Fatalf("loadConfig empty file: %v", err)
	}

	g.Expect(cfg.MaxIterations).To(Equal(2))
	g.Expect(cfg.Detectors).To(HaveLen(2))
}

func TestLoadConfig_MissingFile(t *testing.T) {
	g := NewParallelGomega(t)

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
			name: "invalid flightRecorder slowStageThreshold",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Detectors:     detectorSpecs("govet"),
				FlightRecorder: &flightRecorderFileConfig{
					Enabled:            true,
					SlowStageThreshold: "not-a-duration",
				},
			},
			wantErr: true,
		},
		{
			name: "valid flightRecorder config",
			cfg: pipelineConfigFile{
				MaxIterations: 1,
				Detectors:     detectorSpecs("govet"),
				FlightRecorder: &flightRecorderFileConfig{
					Enabled:            true,
					SlowStageThreshold: "30s",
					OutputDir:          "/tmp/traces",
				},
			},
			wantErr: false,
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
			g := NewParallelGomega(t)

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

			pc, err := tt.cfg.toPipelineConfig()
			if err != nil {
				t.Fatalf("toPipelineConfig: %v", err)
			}

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
