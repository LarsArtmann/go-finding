package pipeline

import (
	"context"
	"encoding/json/v2"
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// TestIntegration_ConfigFileToPipelineRun proves that a JSON ConfigFile can be
// loaded, resolved into a Config, combined with a DetectorRegistry, and run
// end-to-end through the Pipeline. This is the "ConfigFile feature actually
// works" proof that was missing.
func TestIntegration_ConfigFileToPipelineRun(t *testing.T) {
	g := NewParallelGomega(t)

	// 1. Register detectors in a registry.
	registry := finding.NewDetectorRegistry()
	registerFindingDetector(registry, "alpha", "alpha-tool", finding.Finding{
		ID: "alpha-1", Rule: "unused-var", ToolName: "alpha-tool",
		Message: "unused variable x", Severity: finding.SeverityWarning,
	})
	registerFindingDetector(registry, "beta", "beta-tool", finding.Finding{
		ID: "beta-1", Rule: "ineffassign", ToolName: "beta-tool",
		Message: "ineffectual assignment", Severity: finding.SeverityError,
	})

	// 2. Parse config file (keep the ConfigFile to resolve detectors).
	configJSON := `{
		"maxIterations": 2,
		"parallelDetectors": false,
		"dryRun": true,
		"detectorNames": ["alpha", "beta"]
	}`

	var cf ConfigFile
	g.Expect(json.Unmarshal([]byte(configJSON), &cf)).To(Succeed())

	// Convert to pipeline Config.
	config, err := ConfigFromFile([]byte(configJSON))
	g.Expect(err).NotTo(HaveOccurred())

	// 3. Resolve detectors via the registry.
	detectors, err := cf.ResolveDetectors(registry)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(detectors).To(HaveLen(2))

	// 4. Build and run the pipeline.
	p, err := New(config, t.TempDir(), detectors...)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.Run(context.Background())
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(result.TotalIterations).To(Equal(2))
	g.Expect(result.Iterations).To(HaveLen(2))
	// DryRun skips fixes, so both findings are suggested.
	g.Expect(result.Iterations[0].FindingsFound).To(Equal(2))
}

// TestIntegration_ConfigFileUnknownDetector verifies that an unknown detector
// name in the config produces a clear error wrapping errResolveDetector.
func TestIntegration_ConfigFileUnknownDetector(t *testing.T) {
	g := NewParallelGomega(t)

	registry := finding.NewDetectorRegistry()

	var file ConfigFile
	g.Expect(json.Unmarshal([]byte(`{"detectorNames": ["ghost"]}`), &file)).To(Succeed())

	_, err := file.ResolveDetectors(registry)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errResolveDetector)).To(BeTrue())
}

// TestIntegration_DetectorRegistryBuildAllToPipelineRun proves that a
// DetectorRegistry populated with constructors can drive a full Pipeline run
// via BuildAll. This was tested in isolation only.
func TestIntegration_DetectorRegistryBuildAllToPipelineRun(t *testing.T) {
	g := NewParallelGomega(t)

	registry := finding.NewDetectorRegistry()
	registerFindingDetector(registry, "govet", "vet", finding.Finding{
		ID: "vet-1", Rule: "printf", ToolName: "vet",
		Message: "invalid format string", Severity: finding.SeverityError,
	})
	registerFindingDetector(registry, "staticcheck", "sc", finding.Finding{
		ID: "sc-1", Rule: "SA1000", ToolName: "sc",
		Message: "invalid regular expression", Severity: finding.SeverityWarning,
	})

	// BuildAll returns detectors in sorted name order.
	detectors, err := registry.BuildAll()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(detectors).To(HaveLen(2))
	g.Expect(detectors[0].Name()).To(Equal("govet"))
	g.Expect(detectors[1].Name()).To(Equal("staticcheck"))

	config := DefaultConfig()
	config.DryRun = true
	config.MaxIterations = 1

	p, err := New(config, t.TempDir(), detectors...)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.Run(context.Background())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result.Iterations[0].FindingsFound).To(Equal(2))
}

// TestIntegration_DetectorRegistryPartialBuild verifies selective Build of a
// subset of registered detectors.
func TestIntegration_DetectorRegistryPartialBuild(t *testing.T) {
	g := NewParallelGomega(t)

	registry := finding.NewDetectorRegistry()
	registry.MustRegister("a", func() finding.Detector { return newMockDetector("a", "ta") })
	registry.MustRegister("b", func() finding.Detector { return newMockDetector("b", "tb") })

	onlyA, err := registry.Build("a")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(onlyA.Name()).To(Equal("a"))

	// Names are sorted.
	g.Expect(registry.Names()).To(Equal([]string{"a", "b"}))
}

// TestIntegration_TypeAliasesCompileMatch verifies that the pipeline package's
// type aliases (Detector, DetectorFunc) are identical to the root finding
// package types. This guards backward compatibility: pipeline.Detector MUST be
// assignable to finding.Detector and vice versa.
func TestIntegration_TypeAliasesCompileMatch(t *testing.T) {
	t.Parallel()

	// pipeline.Detector is `type Detector = finding.Detector` (alias).
	// A value of one type satisfies the other — no conversion needed.
	rootDet := finding.NamedDetectorFunc("root", func(_ context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	var pipeDet Detector = rootDet //nolint:staticcheck // explicit type proves pipeline.Detector == finding.Detector alias

	if rootDet.Name() != "root" {
		t.Fatalf("root detector name = %q, want %q", rootDet.Name(), "root")
	}

	_ = pipeDet

	// DetectorFunc alias: a finding.DetectorFunc satisfies pipeline.Detector.
	df := finding.DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	var _ Detector = df // DetectorFunc implements Detector (via alias)

	// NamedDetectorFunc variable alias is the same function.
	var _ func(string, DetectorFunc) Detector = finding.NamedDetectorFunc //nolint:staticcheck // intentional: verifying alias identity
}

// TestIntegration_ConfigFromReader verifies the streaming reader path produces
// the same Config as the byte-slice path.
func TestIntegration_ConfigFromReader(t *testing.T) {
	g := NewParallelGomega(t)

	raw := `{"maxIterations": 7, "timeout": "30s", "parallelDetectors": true}`

	fromBytes, err := ConfigFromFile([]byte(raw))
	g.Expect(err).NotTo(HaveOccurred())

	r := strings.NewReader(raw)
	fromReader, err := ConfigFromReader(r)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(fromReader.MaxIterations).To(Equal(fromBytes.MaxIterations))
	g.Expect(fromReader.Timeout).To(Equal(fromBytes.Timeout))
	g.Expect(fromReader.ParallelDetectors).To(Equal(fromBytes.ParallelDetectors))
}
