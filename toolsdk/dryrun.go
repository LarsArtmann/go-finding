package toolsdk

import "context"

type dryRunCtxKey struct{}

// WithDryRun returns a new context that carries the dry-run flag.
// BuildFlow's execution layer calls this when constructing the context for a
// Repair call so the provider honors the user's --dry-run setting.
func WithDryRun(ctx context.Context, dryRun bool) context.Context {
	return context.WithValue(ctx, dryRunCtxKey{}, dryRun)
}

// DryRunFromContext reads the dry-run flag from the context.
// Returns false when no dry-run value was set (including a nil context —
// same miss contract as OptionsFromContext).
func DryRunFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	if v, ok := ctx.Value(dryRunCtxKey{}).(bool); ok {
		return v
	}

	return false
}
