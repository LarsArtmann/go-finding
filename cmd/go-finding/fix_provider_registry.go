package main

import (
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/go-finding/pipeline/goast"
)

// fixProviderName constants for built-in providers.
const (
	fixProviderGoAST = "go-ast"
)

var (
	knownFixProvidersMu sync.RWMutex
	knownFixProviders   = map[string]func() pipeline.FixProvider{
		fixProviderGoAST: func() pipeline.FixProvider { return &goast.Provider{} },
	}
)

// ErrUnknownFixProvider is returned when a fix provider name is not recognized.
var ErrUnknownFixProvider = fmt.Errorf("unknown fix provider")

func lookupFixProvider(name string) (func() pipeline.FixProvider, bool) {
	knownFixProvidersMu.RLock()
	defer knownFixProvidersMu.RUnlock()

	b, ok := knownFixProviders[name]

	return b, ok
}

func availableFixProviderNames() []string {
	knownFixProvidersMu.RLock()
	defer knownFixProvidersMu.RUnlock()

	return slices.Collect(maps.Keys(knownFixProviders))
}

// resolveFixProviders converts provider names into a []pipeline.FixProvider slice.
// When non-empty, the returned slice prepends the named providers to the default
// text-based providers (OffsetProvider, LineProvider, SubstringProvider).
func resolveFixProviders(names []string) []pipeline.FixProvider {
	if len(names) == 0 {
		return nil
	}

	providers := make([]pipeline.FixProvider, 0, len(names)+3)

	for _, name := range names {
		builder, ok := lookupFixProvider(name)
		if !ok {
			continue
		}

		providers = append(providers, builder())
	}

	// Append default text-based providers as fallbacks.
	providers = append(
		providers,
		&pipeline.OffsetProvider{},
		&pipeline.LineProvider{},
		&pipeline.SubstringProvider{},
	)

	return providers
}
