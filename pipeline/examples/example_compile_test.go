package examples_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExamplesCompile(t *testing.T) {
	t.Parallel()

	tests := []string{"./outcomes"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // intentionally building known example directories
			cmd := exec.CommandContext(t.Context(), "go", "build", "-o", "/dev/null", path)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build %s failed: %v\n%s", path, err, out)
			}

			if strings.TrimSpace(string(out)) != "" {
				t.Logf("go build %s output: %s", path, out)
			}
		})
	}
}
