package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

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

	pc, err := cfg.toPipelineConfig()
	if err != nil {
		t.Fatalf("toPipelineConfig: %v", err)
	}
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
