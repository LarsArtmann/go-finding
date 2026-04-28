package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	bin := filepath.Join(tmpDir, "go-finding")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", bin, ".") //nolint:gosec // test builds its own binary
	cmd.Dir = "/home/lars/projects/go-finding/cmd/go-finding"
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", out)

	return bin
}

func initGoModule(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "go", "mod", "init", "testmod") //nolint:gosec // test helper
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go mod init failed: %s", out)
}

func TestRun_E2E_DefaultDetectors(t *testing.T) {
	t.Parallel()

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	goFile := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nfunc main() {\n\tunused := 42\n}\n"), 0o644))

	cmd := exec.CommandContext(context.Background(), bin, "-dir="+tmpDir, "-format=text") //nolint:gosec // E2E test
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "output: %s", out)
}

func TestRun_E2E_ConfigFile(t *testing.T) {
	t.Parallel()

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("maxIterations: 1\ndetectors:\n  - name: govet\n"), 0o644))

	goFile := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644))

	cmd := exec.CommandContext(context.Background(), bin, "-config="+cfgFile, "-dir="+tmpDir) //nolint:gosec // E2E test
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "output: %s", out)
}

func TestRun_E2E_SARIFOutput(t *testing.T) {
	t.Parallel()

	bin := buildBinary(t)
	tmpDir := t.TempDir()
	initGoModule(t, tmpDir)
	goFile := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644))

	outFile := filepath.Join(tmpDir, "out.sarif")
	cmd := exec.CommandContext(context.Background(), bin, "-dir="+tmpDir, "-format=sarif") //nolint:gosec // E2E test
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "output: %s", out)
	require.NoError(t, os.WriteFile(outFile, out, 0o644))

	data, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"version": "2.1.0"`)
}
