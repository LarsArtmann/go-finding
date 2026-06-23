package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding"
)

// GeneratedFileFilter is a FindingTransformer that removes findings from
// auto-generated Go source files (sqlc, protobuf, mockgen, stringer, etc.).
//
// It wraps gogenfilter.Filter and evaluates each finding's Position.File.
// Findings from files that are detected as generated are dropped.
//
// When a file cannot be read for content-based detection (e.g. file removed
// between detection and filtering), the finding is kept unchanged and the
// error is logged rather than propagated, preventing pipeline abort on I/O
// edge cases.
type GeneratedFileFilter struct {
	filter *gogenfilter.Filter
	logger *slog.Logger
}

// NewGeneratedFileFilter creates a FindingTransformer that filters out findings
// from generated Go source files.
//
// Pass gogenfilter configuration options via configs. Common configs:
//   - gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
//   - gogenfilter.WithFilterOptions(gogenfilter.FilterSQLC, gogenfilter.FilterTempl)
//   - gogenfilter.WithExcludePatterns("**/vendor/**")
//   - gogenfilter.WithIncludePatterns("pkg/**")
//
// If no config is provided a disabled filter is used (no findings removed).
func NewGeneratedFileFilter(
	logger *slog.Logger,
	configs ...gogenfilter.FilterConfig,
) (*GeneratedFileFilter, error) {
	filter, err := gogenfilter.NewFilter(configs...)
	if err != nil {
		return nil, fmt.Errorf("new generated file filter: %w", err)
	}

	return &GeneratedFileFilter{
		filter: filter,
		logger: logger,
	}, nil
}

// Name returns the processor name.
func (*GeneratedFileFilter) Name() string {
	return "generated-file-filter"
}

// Transform filters out findings whose source file is auto-generated.
func (g *GeneratedFileFilter) Transform(
	ctx context.Context,
	findings []finding.Finding,
) ([]finding.Finding, error) {
	if !g.filter.IsEnabled() {
		return findings, nil
	}

	result := make([]finding.Finding, 0, len(findings))

	for _, f := range findings {
		err := ctx.Err()
		if err != nil {
			return nil, fmt.Errorf("generated filter cancelled: %w", err)
		}

		if f.Position.File == "" {
			result = append(result, f)

			continue
		}

		filtered, err := g.filter.Filter(f.Position.File)
		if err != nil {
			g.logFilterError(ctx, f.Position.File, err)
			result = append(result, f)

			continue
		}

		if !filtered {
			result = append(result, f)
		}
	}

	return result, nil
}

func (g *GeneratedFileFilter) logFilterError(
	ctx context.Context,
	file string,
	err error,
) {
	if g.logger != nil {
		g.logger.LogAttrs(
			ctx,
			slog.LevelWarn,
			"generated-file-filter: could not evaluate file, keeping finding",
			slog.String("file", file),
			slog.String("error", err.Error()),
		)
	}
}
