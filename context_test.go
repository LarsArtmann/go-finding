package finding

import (
	"context"
	"testing"
)

func TestWithWorkingDir(t *testing.T) {
	ctx := WithWorkingDir(context.Background(), "/tmp/test")

	if dir := WorkingDirFromContext(ctx); dir != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %q", dir)
	}
}

func TestWorkingDirFromContext_Empty(t *testing.T) {
	if dir := WorkingDirFromContext(context.Background()); dir != "" {
		t.Errorf("expected empty string, got %q", dir)
	}
}

func TestWorkingDirFromContext_Overwrite(t *testing.T) {
	ctx := WithWorkingDir(context.Background(), "/first")
	ctx = WithWorkingDir(ctx, "/second")

	if dir := WorkingDirFromContext(ctx); dir != "/second" {
		t.Errorf("expected /second (overwrite), got %q", dir)
	}
}
