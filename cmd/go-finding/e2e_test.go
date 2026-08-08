package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	bin := filepath.Join(tmpDir, "go-finding")
	cmd := exec.CommandContext( //nolint:gosec // test builds its own binary
		context.Background(),
		"go",
		"build",
		"-o",
		bin,
		".",
	)
	cmd.Dir, _ = filepath.Abs(".")
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())

	return bin
}

func initGoModule(t *testing.T, dir string) {
	t.Helper()
	g := NewWithT(t)
	cmd := exec.CommandContext(context.Background(), "go", "mod", "init", "testmod")
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRun_E2E_DefaultDetectors(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {\n\tunused := 42\n}\n"), 0o644)).
		NotTo(HaveOccurred())

	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-dir="+tmpDir,
		"-format=text",
	)
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRun_E2E_ConfigFile(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	g.Expect(os.WriteFile(cfgFile, []byte("maxIterations: 1\ndetectors:\n  - name: govet\n"), 0o644)).
		NotTo(HaveOccurred())

	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-config="+cfgFile,
		"-dir="+tmpDir,
	)
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRun_E2E_SARIFOutput(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	outFile := filepath.Join(tmpDir, "out.sarif")
	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-dir="+tmpDir,
		"-format=sarif",
	)
	out, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(os.WriteFile(outFile, out, 0o644)).NotTo(HaveOccurred())

	data, err := os.ReadFile(outFile)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data)).To(ContainSubstring(`"version": "2.1.0"`))
}

func TestRun_E2E_TraceFlag(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	traceDir := filepath.Join(tmpDir, "traces")

	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-dir="+tmpDir,
		"-trace",
		"-trace-dir="+traceDir,
		"-trace-slow=1ns",
	)
	out, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(out)).To(ContainSubstring("Flight recorder enabled"))

	entries, err := os.ReadDir(traceDir)
	g.Expect(err).NotTo(HaveOccurred())

	var traceFiles []string

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			traceFiles = append(traceFiles, e.Name())
		}
	}

	g.Expect(traceFiles).NotTo(BeEmpty(), "expected at least one .trace file in %s", traceDir)

	for _, name := range traceFiles {
		info, statErr := os.Stat(filepath.Join(traceDir, name))
		g.Expect(statErr).NotTo(HaveOccurred())
		g.Expect(info.Size()).NotTo(BeZero(), "trace file %s is empty", name)
	}
}

func TestRun_E2E_TraceViaConfigFile(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)

	traceDir := filepath.Join(tmpDir, "traces")

	cfgContent := `
maxIterations: 1
detectors:
  - name: govet
flightRecorder:
  enabled: true
  outputDir: "` + traceDir + `"
  slowStageThreshold: "1ns"
`
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	g.Expect(os.WriteFile(cfgFile, []byte(cfgContent), 0o644)).NotTo(HaveOccurred())

	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-config="+cfgFile,
		"-dir="+tmpDir,
	)
	out, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(out)).To(ContainSubstring("Flight recorder enabled via config"))

	entries, readErr := os.ReadDir(traceDir)
	g.Expect(readErr).NotTo(HaveOccurred())

	var hasTrace bool

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			hasTrace = true

			break
		}
	}

	g.Expect(hasTrace).To(BeTrue(), "expected at least one .trace file in config-based trace dir")
}

func TestRun_E2E_TraceViaConfigFile_AllFields(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)

	traceDir := filepath.Join(tmpDir, "traces")

	cfgContent := `
maxIterations: 1
detectors:
  - name: govet
flightRecorder:
  enabled: true
  outputDir: "` + traceDir + `"
  slowStageThreshold: "1ns"
  minAge: "1m"
  maxBytes: 4194304
`
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	g.Expect(os.WriteFile(cfgFile, []byte(cfgContent), 0o644)).NotTo(HaveOccurred())

	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	cmd := exec.CommandContext( //nolint:gosec // E2E test
		context.Background(),
		bin,
		"-config="+cfgFile,
		"-dir="+tmpDir,
	)
	out, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(out)).To(ContainSubstring("Flight recorder enabled via config"))

	entries, readErr := os.ReadDir(traceDir)
	g.Expect(readErr).NotTo(HaveOccurred())

	var hasTrace bool

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			hasTrace = true

			break
		}
	}

	g.Expect(hasTrace).To(BeTrue(), "expected at least one .trace file when all 5 config fields are set")
}

func TestRun_E2E_FilterGenerated(t *testing.T) {
	g := NewParallelGomega(t)

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	g.Expect(os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644)).
		NotTo(HaveOccurred())

	genFile := filepath.Join(tmpDir, "query.sql.go")
	g.Expect(os.WriteFile(genFile, []byte("// Code generated by sqlc. DO NOT EDIT.\npackage main\n"), 0o644)).
		NotTo(HaveOccurred())

	cmd := exec.CommandContext(
		context.Background(),
		bin,
		"-dir="+tmpDir,
		"-format=json",
		"-filter-generated",
	)
	out, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(out)).To(ContainSubstring(`"findings"`))

	cmdTypes := exec.CommandContext(
		context.Background(),
		bin,
		"-dir="+tmpDir,
		"-format=text",
		"-filter-generated",
		"-filter-generated-types=sqlc,mockgen",
	)
	outTypes, err := cmdTypes.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(outTypes)).To(ContainSubstring("Done:"))
}
