package examples_test

import (
	"os/exec"
	"testing"
)

func TestExamplesCompile(t *testing.T) {
	t.Parallel()

	tests := []string{"./basic", "./builder", "./pipeline"}
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
