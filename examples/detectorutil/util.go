package detectorutil

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

func RunTool(ctx context.Context, dir, tool string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return output, nil
		}
		return nil, fmt.Errorf("%s: %w", tool, err)
	}
	return output, nil
}
