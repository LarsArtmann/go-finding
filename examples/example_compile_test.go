package examples_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExamplesCompile(t *testing.T) {
	t.Parallel()

	tests := []string{"./basic", "./builder"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // intentionally building known example directories
			cmd := exec.CommandContext(t.Context(), "go", "build", "-o", "/dev/null", path)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build %s failed: %v\n%s", path, err, out)
			}
		})
	}
}

func TestExamplesRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path    string
		contain string
	}{
		{"./basic", "Report has 1 finding(s)"},
		{"./builder", "Fix available: oldPattern → newPattern"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // intentionally running known example directories
			cmd := exec.CommandContext(t.Context(), "go", "run", tt.path)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go run %s failed: %v\n%s", tt.path, err, out)
			}

			if !strings.Contains(string(out), tt.contain) {
				t.Fatalf("go run %s: output missing %q\nGot:\n%s", tt.path, tt.contain, out)
			}
		})
	}
}
