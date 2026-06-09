package main

import (
	"cmp"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding/pipeline"
)

// filterTypeRegistry maps user-facing type strings to gogenfilter.FilterOption values.
var filterTypeRegistry = map[string]gogenfilter.FilterOption{
	"all":          gogenfilter.FilterAll,
	"sqlc":         gogenfilter.FilterSQLC,
	"templ":        gogenfilter.FilterTempl,
	"go-enum":      gogenfilter.FilterGoEnum,
	"protobuf":     gogenfilter.FilterProtobuf,
	"oapi-codegen": gogenfilter.FilterOapi,
	"deepcopy-gen": gogenfilter.FilterDeepcopy,
	"wire":         gogenfilter.FilterWire,
	"moq":          gogenfilter.FilterMoq,
	"mockgen":      gogenfilter.FilterMockgen,
	"stringer":     gogenfilter.FilterStringer,
	"generic":      gogenfilter.FilterGeneric,
}

// addGeneratedFilter creates a GeneratedFileFilter processor and appends it
// to the pipeline configuration. It is called when either the CLI flag
// -filter-generated or the config field filterGenerated is set.
func addGeneratedFilter(
	pipelineCfg *pipeline.Config,
	cfg pipelineConfigFile,
	filterGenTypes string,
	generatedExclude string,
	generatedInclude string,
) error {
	opts, err := parseFilterGenTypes(filterGenTypes, cfg.FilterGenTypes)
	if err != nil {
		return err
	}

	var configs []gogenfilter.FilterConfig

	optConfig, err := gogenfilter.WithFilterOptions(opts...)
	if err != nil {
		return fmt.Errorf("filter options: %w", err)
	}

	configs = append(configs, optConfig)

	if excl := splitCommaList(generatedExclude); len(excl) > 0 {
		configs = append(configs, gogenfilter.WithExcludePatterns(excl...))
	}
	if excl := cfg.GeneratedExclude; len(excl) > 0 {
		configs = append(configs, gogenfilter.WithExcludePatterns(excl...))
	}

	if incl := splitCommaList(generatedInclude); len(incl) > 0 {
		configs = append(configs, gogenfilter.WithIncludePatterns(incl...))
	}
	if incl := cfg.GeneratedInclude; len(incl) > 0 {
		configs = append(configs, gogenfilter.WithIncludePatterns(incl...))
	}

	logger := cmp.Or(pipelineCfg.Logger, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	filter, err := pipeline.NewGeneratedFileFilter(logger, configs...)
	if err != nil {
		return fmt.Errorf("new generated file filter: %w", err)
	}

	pipelineCfg.Processors = append(pipelineCfg.Processors, filter)

	return nil
}

// parseFilterGenTypes resolves the generator types from CLI or config.
// If CLI types is non-empty "all", it maps to FilterAll. Otherwise it
// parses comma-separated values from CLI or config.
func parseFilterGenTypes(cliTypes, configTypes string) ([]gogenfilter.FilterOption, error) {
	typeStr := cmp.Or(cliTypes, configTypes, "all")

	parts := strings.Split(typeStr, ",")
	opts := make([]gogenfilter.FilterOption, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}

		opt, ok := filterTypeRegistry[p]
		if !ok {
			known := slices.Sorted(slices.Values(mustKeys(filterTypeRegistry)))

			return nil, fmt.Errorf(
				"%w: unknown generated filter type %q (known: %s)",
				errInvalidConfig, p, strings.Join(known, ", "),
			)
		}

		opts = append(opts, opt)
	}

	if len(opts) == 0 {
		opts = append(opts, gogenfilter.FilterAll)
	}

	return opts, nil
}

func mustKeys(m map[string]gogenfilter.FilterOption) []string {
	keys := slices.Collect(maps.Keys(m))
	return keys
}

// splitCommaList parses comma-separated patterns. Returns nil if the string is empty.
func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
