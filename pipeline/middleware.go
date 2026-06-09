package pipeline

import (
	"context"
	"slices"
)

// MiddlewareFunc wraps a pipeline's Run execution. Use it to add
// cross-cutting concerns like logging, metrics, tracing, or error
// enrichment without modifying the pipeline itself.
//
// Middleware is applied in registration order: the first registered
// middleware wraps the second, which wraps the third, etc.
type MiddlewareFunc func(next RunFunc) RunFunc

// RunFunc is the function signature for pipeline execution,
// matching [Pipeline.Run].
type RunFunc func(ctx context.Context) (*PipelineResult, error)

// ComposeMiddleware composes multiple middleware into a single middleware.
// The first middleware in the slice is the outermost wrapper.
func ComposeMiddleware(mws ...MiddlewareFunc) MiddlewareFunc {
	return func(next RunFunc) RunFunc {
		for i := range slices.Backward(mws) {
			next = mws[i](next)
		}

		return next
	}
}
