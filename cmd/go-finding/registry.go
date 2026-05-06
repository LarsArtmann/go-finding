package main

import (
	"fmt"
	"os"
	"sync"

	det "github.com/larsartmann/go-finding/internal/detectors"
	"github.com/larsartmann/go-finding/pipeline"
)

var (
	knownDetectorBuildersMu sync.RWMutex
	knownDetectorBuilders   = map[string]func(string) pipeline.Detector{
		"govet":       det.NewGoVetDetector,
		"staticcheck": det.NewStaticcheckDetector,
	}
)

// RegisterDetector registers a custom detector builder by name.
// It is safe for concurrent use. Returns an error if the name is already registered.
func RegisterDetector(name string, builder func(string) pipeline.Detector) error {
	knownDetectorBuildersMu.Lock()
	defer knownDetectorBuildersMu.Unlock()

	if _, exists := knownDetectorBuilders[name]; exists {
		return fmt.Errorf("%w: %q", errDetectorRegistered, name)
	}

	knownDetectorBuilders[name] = builder

	return nil
}

func lookupDetectorBuilder(name string) (func(string) pipeline.Detector, bool) {
	knownDetectorBuildersMu.RLock()
	defer knownDetectorBuildersMu.RUnlock()

	b, ok := knownDetectorBuilders[name]

	return b, ok
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
