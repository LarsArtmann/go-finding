package main

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/larsartmann/go-finding/lockutil"
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
var ErrUnknownFixProvider = errors.New("unknown fix provider")

func lookupFixProvider(name string) (func() pipeline.FixProvider, bool) {
	type lookup struct {
		builder func() pipeline.FixProvider
		ok      bool
	}

	res := lockutil.RLocked(&knownFixProvidersMu, func() lookup {
		b, ok := knownFixProviders[name]

		return lookup{builder: b, ok: ok}
	})

	return res.builder, res.ok
}

func availableFixProviderNames() []string {
	return lockutil.RLocked(&knownFixProvidersMu, func() []string {
		return slices.Sorted(maps.Keys(knownFixProviders))
	})
}

// resolveFixProviders converts provider names into a []pipeline.FixProvider slice.
// When non-empty, the returned slice prepends the named providers to the default
// text-based providers (OffsetProvider, LineProvider, SubstringProvider).
// Returns an error listing the valid names if any name is unknown.
func resolveFixProviders(names []string) ([]pipeline.FixProvider, error) {
	if len(names) == 0 {
		return nil, nil
	}

	providers := make([]pipeline.FixProvider, 0, len(names)+3)

	for _, name := range names {
		builder, ok := lookupFixProvider(name)
		if !ok {
			return nil, fmt.Errorf("%w: %q (available: %s)",
				ErrUnknownFixProvider, name,
				strings.Join(availableFixProviderNames(), ", "))
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

	return providers, nil
}
