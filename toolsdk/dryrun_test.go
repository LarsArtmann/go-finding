package toolsdk

import (
	"context"
	"testing"
)

func TestDryRunFromContext_DefaultsFalse(t *testing.T) {
	t.Parallel()

	if DryRunFromContext(context.Background()) {
		t.Error("fresh context must read as dry-run=false")
	}

	if DryRunFromContext(nil) {
		t.Error("nil context must read as dry-run=false, not panic")
	}
}

func TestDryRunRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithDryRun(context.Background(), true)
	if !DryRunFromContext(ctx) {
		t.Error("WithDryRun(true) lost the flag on the way back")
	}

	ctx = WithDryRun(ctx, false)
	if DryRunFromContext(ctx) {
		t.Error("WithDryRun(false) did not overwrite an earlier true value")
	}
}

func TestDryRunContextValueIsolation(t *testing.T) {
	t.Parallel()

	ctx := WithDryRun(context.Background(), true)
	child, cancel := context.WithCancel(ctx)
	defer cancel()

	if !DryRunFromContext(child) {
		t.Error("derived context lost the parent's dry-run flag")
	}
}
