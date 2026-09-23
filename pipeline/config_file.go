package pipeline

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/larsartmann/go-finding"
)

// ConfigFile represents a pipeline configuration that can be loaded from JSON.
// Duration fields are strings (e.g., "10m", "30s") parsed via time.ParseDuration.
// Zero values inherit defaults from [DefaultConfig].
//
// DetectorNames and ProviderNames store names to be resolved via a
// [DetectorRegistry] and fix provider map at construction time — the config
// file itself cannot construct function values.
// FlightRecorderFileConfig is the JSON-friendly representation of flight
// recorder settings for use in a [ConfigFile]. Duration fields use string
// syntax (e.g. "30s", "2m") parsed via time.ParseDuration, matching the
// convention used elsewhere in ConfigFile.
type FlightRecorderFileConfig struct {
	Enabled            bool   `json:"enabled"`
	OutputDir          string `json:"outputDir"`
	SlowStageThreshold string `json:"slowStageThreshold"`
	MinAge             string `json:"minAge"`
	MaxBytes           uint64 `json:"maxBytes"`
	MaxFiles           int    `json:"maxFiles"`
	Compress           bool   `json:"compress"`
}

type ConfigFile struct {
	MaxIterations              int                       `json:"maxIterations"`
	ParallelDetectors          bool                      `json:"parallelDetectors"`
	VerifyAfterFix             bool                      `json:"verifyAfterFix"`
	Timeout                    string                    `json:"timeout"`
	DetectorTimeouts           map[string]string         `json:"detectorTimeouts"`
	GracefulDegradation        bool                      `json:"gracefulDegradation"`
	DryRun                     bool                      `json:"dryRun"`
	CorrelateFindings          bool                      `json:"correlateFindings"`
	ByteLevelConflictDetection bool                      `json:"byteLevelConflictDetection"`
	FixRollbackAllFiles        bool                      `json:"fixRollbackAllFiles"`
	Severity                   string                    `json:"severity"`
	DetectorNames              []string                  `json:"detectorNames"`
	ProviderNames              []string                  `json:"providerNames"`
	FlightRecorder             *FlightRecorderFileConfig `json:"flightRecorder"`
}

// Sentinel errors for config file resolution.
var (
	errUnknownProvider = errors.New("unknown provider")
	errResolveDetector = errors.New("resolve detector")
)

// ConfigFromFile parses a JSON [ConfigFile] and converts it to a [Config].
// Returns an error if any duration strings are malformed.
func ConfigFromFile(data []byte) (Config, error) {
	var cf ConfigFile

	return decodeConfig(cf, json.Unmarshal(data, &cf))
}

// ConfigFromReader parses a JSON [ConfigFile] from an [io.Reader].
func ConfigFromReader(r io.Reader) (Config, error) {
	var cf ConfigFile

	return decodeConfig(cf, json.UnmarshalRead(r, &cf))
}

// decodeConfig wraps an unmarshal error and converts the [ConfigFile] to a [Config].
func decodeConfig(cf ConfigFile, err error) (Config, error) {
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
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

	return Config{ //nolint:exhaustruct_v5
		MaxIterations:              maxIter,
		ParallelDetectors:          cf.ParallelDetectors,
		VerifyAfterFix:             cf.VerifyAfterFix,
		Timeout:                    timeout,
		GracefulDegradation:        cf.GracefulDegradation,
		DryRun:                     cf.DryRun,
		CorrelateFindings:          cf.CorrelateFindings,
		ByteLevelConflictDetection: cf.ByteLevelConflictDetection,
		FixRollbackAllFiles:        cf.FixRollbackAllFiles,
		DetectorTimeouts:           detectorTimeouts,
	}, nil
}

// SeverityFilter returns the parsed minimum severity, or ok=false if not set.
func (cf ConfigFile) SeverityFilter() (finding.Severity, bool) {
	if cf.Severity == "" {
		return finding.SeverityInfo, false
	}

	sev, err := finding.ParseSeverity(cf.Severity)
	if err != nil {
		return finding.SeverityInfo, false
	}

	return sev, true
}

// ResolveDetectors builds detectors from the config's DetectorNames using the
// given registry. Returns an error if any name is not registered.
func (cf ConfigFile) ResolveDetectors(registry *finding.DetectorRegistry) ([]finding.Detector, error) {
	if len(cf.DetectorNames) == 0 {
		return nil, nil
	}

	detectors := make([]finding.Detector, 0, len(cf.DetectorNames))

	for _, name := range cf.DetectorNames {
		d, err := registry.Build(name)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", errResolveDetector, name, err)
		}

		detectors = append(detectors, d)
	}

	return detectors, nil
}

// ResolveFlightRecorderConfig maps the JSON-friendly file representation
// onto a [FlightRecorderConfig]: string-encoded durations are parsed and
// unset fields fall back to [DefaultFlightRecorderConfig]. A nil section
// yields the defaults. Duration errors read
// "parse flightRecorder.<field> %q" — that wording is contract.
func ResolveFlightRecorderConfig(fc *FlightRecorderFileConfig) (FlightRecorderConfig, error) {
	config := DefaultFlightRecorderConfig()

	if fc == nil {
		return config, nil
	}

	if fc.OutputDir != "" {
		config.OutputDir = fc.OutputDir
	}

	if fc.SlowStageThreshold != "" {
		d, err := time.ParseDuration(fc.SlowStageThreshold)
		if err != nil {
			return FlightRecorderConfig{}, fmt.Errorf(
				"parse flightRecorder.slowStageThreshold %q: %w",
				fc.SlowStageThreshold,
				err,
			)
		}

		config.SlowStageThreshold = d
	}

	if fc.MinAge != "" {
		d, err := time.ParseDuration(fc.MinAge)
		if err != nil {
			return FlightRecorderConfig{}, fmt.Errorf(
				"parse flightRecorder.minAge %q: %w",
				fc.MinAge,
				err,
			)
		}

		config.MinAge = d
	}

	if fc.MaxBytes != 0 {
		config.MaxBytes = fc.MaxBytes
	}

	if fc.MaxFiles != 0 {
		config.MaxFiles = fc.MaxFiles
	}

	config.Compress = fc.Compress

	return config, nil
}

// ResolveFlightRecorder constructs a [FlightRecorderHook] from the config's
// FlightRecorder section. Returns (nil, nil) when flight recording is not
// enabled. The caller is responsible for appending the returned hook to
// [Config.StageHooks] and calling Close() when done.
func (cf ConfigFile) ResolveFlightRecorder() (*FlightRecorderHook, error) {
	if cf.FlightRecorder == nil || !cf.FlightRecorder.Enabled {
		return nil, nil
	}

	config, err := ResolveFlightRecorderConfig(cf.FlightRecorder)
	if err != nil {
		return nil, err
	}

	return NewFlightRecorderHook(config)
}

// ResolveProviders builds fix providers from the config's ProviderNames using
// the given map. Returns an error if any name is not found.
func (cf ConfigFile) ResolveProviders(
	providers map[string]FixProvider,
) ([]FixProvider, error) {
	if len(cf.ProviderNames) == 0 {
		return nil, nil
	}

	result := make([]FixProvider, 0, len(cf.ProviderNames))

	for _, name := range cf.ProviderNames {
		p, ok := providers[name]
		if !ok {
			return nil, fmt.Errorf("%w: resolve provider %q", errUnknownProvider, name)
		}

		result = append(result, p)
	}

	return result, nil
}
