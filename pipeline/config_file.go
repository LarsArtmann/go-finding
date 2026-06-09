package pipeline

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ConfigFile represents a pipeline configuration that can be loaded from JSON.
// Duration fields are strings (e.g., "10m", "30s") parsed via time.ParseDuration.
// Zero values inherit defaults from [DefaultConfig].
type ConfigFile struct {
	MaxIterations              int               `json:"maxIterations"`
	ParallelDetectors          bool              `json:"parallelDetectors"`
	VerifyAfterFix             bool              `json:"verifyAfterFix"`
	Timeout                    string            `json:"timeout"`
	DetectorTimeouts           map[string]string `json:"detectorTimeouts"`
	GracefulDegradation        bool              `json:"gracefulDegradation"`
	DryRun                     bool              `json:"dryRun"`
	CorrelateFindings          bool              `json:"correlateFindings"`
	ByteLevelConflictDetection bool              `json:"byteLevelConflictDetection"`
}

// ConfigFromFile parses a JSON [ConfigFile] and converts it to a [Config].
// Returns an error if any duration strings are malformed.
func ConfigFromFile(data []byte) (Config, error) {
	var cf ConfigFile

	err := json.Unmarshal(data, &cf)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cf.toConfig()
}

// ConfigFromReader parses a JSON [ConfigFile] from an [io.Reader].
func ConfigFromReader(r io.Reader) (Config, error) {
	var cf ConfigFile

	err := json.NewDecoder(r).Decode(&cf)
	if err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	return cf.toConfig()
}

func (cf ConfigFile) toConfig() (Config, error) {
	timeout := DefaultTimeout

	if cf.Timeout != "" {
		d, err := time.ParseDuration(cf.Timeout)
		if err != nil {
			return Config{}, fmt.Errorf("parse timeout %q: %w", cf.Timeout, err)
		}

		timeout = d
	}

	maxIter := cf.MaxIterations
	if maxIter == 0 {
		maxIter = DefaultMaxIterations
	}

	detectorTimeouts := make(map[string]time.Duration, len(cf.DetectorTimeouts))
	for name, durStr := range cf.DetectorTimeouts {
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return Config{}, fmt.Errorf("parse detector timeout %q for %q: %w", durStr, name, err)
		}

		detectorTimeouts[name] = d
	}

	return Config{ //nolint:exhaustruct
		MaxIterations:              maxIter,
		ParallelDetectors:          cf.ParallelDetectors,
		VerifyAfterFix:             cf.VerifyAfterFix,
		Timeout:                    timeout,
		GracefulDegradation:        cf.GracefulDegradation,
		DryRun:                     cf.DryRun,
		CorrelateFindings:          cf.CorrelateFindings,
		ByteLevelConflictDetection: cf.ByteLevelConflictDetection,
		DetectorTimeouts:           detectorTimeouts,
	}, nil
}
