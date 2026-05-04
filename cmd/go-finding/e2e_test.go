package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
	cmd.Dir = "/home/lars/projects/go-finding/cmd/go-finding"
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())

	return bin
}

func initGoModule(t *testing.T, dir string) {
	g := NewWithT(t)
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "go", "mod", "init", "testmod")
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRun_E2E_DefaultDetectors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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
	t.Parallel()
	g := NewWithT(t)

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
