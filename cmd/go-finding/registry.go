package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"

	det "github.com/larsartmann/go-finding/cmd/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/lockutil"
	"github.com/larsartmann/go-finding/pipeline"
)

var (
	knownDetectorBuildersMu sync.RWMutex
	knownDetectorBuilders   = map[string]func(string) pipeline.Detector{
		detectorNameGovet:       det.NewGoVetDetector,
		detectorNameStaticcheck: det.NewStaticcheckDetector,
	}
)

// RegisterDetector registers a custom detector builder by name.
// It is safe for concurrent use. Returns an error if the name is already registered.
func RegisterDetector(name string, builder func(string) pipeline.Detector) error {
	return lockutil.Locked(&knownDetectorBuildersMu, func() error {
		if _, exists := knownDetectorBuilders[name]; exists {
			return fmt.Errorf("%w: %q", errDetectorRegistered, name)
		}

		knownDetectorBuilders[name] = builder

		return nil
	})
}

func lookupDetectorBuilder(name string) (func(string) pipeline.Detector, bool) {
	type lookup struct {
		builder func(string) pipeline.Detector
		ok      bool
	}

	res := lockutil.RLocked(&knownDetectorBuildersMu, func() lookup {
		b, ok := knownDetectorBuilders[name]

		return lookup{builder: b, ok: ok}
	})

	return res.builder, res.ok
}

func availableDetectorNames() []string {
	return lockutil.RLocked(&knownDetectorBuildersMu, func() []string {
		return slices.Sorted(maps.Keys(knownDetectorBuilders))
	})
}

func buildDetectors(specs []detectorSpec, dir string) []pipeline.Detector {
	var result []pipeline.Detector

	for _, spec := range specs {
		builder, ok := lookupDetectorBuilder(spec.Name)
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: unknown detector %q, skipping\n", spec.Name)

			continue
		}

		result = append(result, builder(dir))
	}

	return result
}
