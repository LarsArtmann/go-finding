package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestOutputResults_WriteError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	report := reportWithFindings()

	err := outputResults(&failingWriter{err: errors.New("disk full")}, report, "json")
	g.Expect(err).To(HaveOccurred())

	g.Expect(err.Error()).To(ContainSubstring("writing JSON"))

	err = outputResults(&failingWriter{err: errors.New("disk full")}, report, "sarif")
	g.Expect(err).To(HaveOccurred())

	g.Expect(err.Error()).To(ContainSubstring("writing SARIF"))
}

func TestOutputText_WithSummary(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	report := finding.NewReport(finding.ToolInfo{Name: testToolName})
	report.AddFinding(finding.Finding{
		Severity: finding.SeverityError, Rule: "R1", Message: "err1",
		Position: finding.Pos("a.go", 1, 1),
		ID:       "test:R1:a.go:1:1",
	})
	report.AddFinding(finding.Finding{
		Severity: finding.SeverityInfo, Rule: "R2", Message: "info1",
		Position: finding.Pos("b.go", 2, 1),
		ID:       "test:R2:b.go:2:1",
	})
	report.ComputeSummary()

	var buf bytes.Buffer
	outputText(&buf, report)

	out := buf.String()
	g.Expect(out).To(ContainSubstring("By severity:"))

	g.Expect(out).To(ContainSubstring("2 finding(s)"))
}

func reportWithFindings() *finding.Report {
	report := finding.NewReport(finding.ToolInfo{Name: testToolName, Version: "1.0"})
	report.AddFinding(
		makeTestFinding(
			"test:nilcheck:main.go:10:5",
			"nilcheck",
			testToolName,
			"possible nil dereference",
			finding.SeverityWarning,
			"main.go",
			10,
			5,
		),
	)
	report.ComputeSummary()

	return report
}

func makeTestFinding(
	id, rule, tool, msg string,
	sev finding.Severity,
	file string,
	line, col int,
) finding.Finding {
	return finding.Finding{
		ID:       id,
		Rule:     rule,
		ToolName: tool,
		Message:  msg,
		Severity: sev,
		Position: finding.Position{File: file, Line: line, Column: col},
	}
}

//nolint:paralleltest // writes to global os.Stderr; may race with TestFatalf
func TestSetupProfiling_BadCPUProfilePath(t *testing.T) {
	g := NewWithT(t)
	stop, err := setupProfiling("/nonexistent_dir/cpu.prof", "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(stop).To(BeNil())
}

//nolint:paralleltest // manipulates global pprof CPU profile state
func TestSetupProfiling_CPUProfileStartFailure(t *testing.T) {
	g := NewWithT(t)
	// Create a temp file that we can't write a valid CPU profile to
	f, err := os.CreateTemp(t.TempDir(), "cpu.prof")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(f.Close()).NotTo(HaveOccurred())

	// Start a CPU profile already, so the second StartCPUProfile fails.
	f2, err := os.CreateTemp(t.TempDir(), "cpu2.prof")
	g.Expect(err).NotTo(HaveOccurred())
	defer func() { _ = f2.Close() }()

	g.Expect(pprof.StartCPUProfile(f2)).NotTo(HaveOccurred())
	defer pprof.StopCPUProfile()

	stop, err := setupProfiling(f.Name(), "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(stop).To(BeNil())
}

func reportWithNaNConfidence() *finding.Report {
	report := finding.NewReport(finding.ToolInfo{Name: testToolName})
	report.AddFinding(finding.Finding{
		ID: "1", Rule: "r1", ToolName: "t", Message: "m",
		Severity: finding.SeverityError, Position: finding.Position{File: "a.go"},
		Confidence: finding.Confidence(math.NaN()),
	})
	return report
}

func testOutputSerializationError(t *testing.T, format, substr string) {
	t.Helper()
	g := NewWithT(t)

	var buf bytes.Buffer
	err := outputResults(&buf, reportWithNaNConfidence(), format)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring(substr))
}

func TestOutputResults_JSONSerializationError(t *testing.T) {
	t.Parallel()
	testOutputSerializationError(t, "json", "serializing JSON")
}

func TestOutputResults_SARIFNaNConfidenceHandled(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	var buf bytes.Buffer
	err := outputResults(&buf, reportWithNaNConfidence(), "sarif")
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRun_InvalidSeverity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()
	g := NewWithT(t)

	saveRestoreFlags(t)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{toolName, "-severity=banana"}

	got := run()
	g.Expect(got).To(Equal(1))
}

func TestRegisterDetector(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	name := uniqueDetName("test-detector")

	// Test registration of a new detector.
	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			testToolName,
			func(_ context.Context) ([]finding.Finding, error) {
				return nil, nil
			},
		)
	})
	g.Expect(err).NotTo(HaveOccurred())

	// Verify it can be looked up.
	builder, ok := lookupDetectorBuilder(name)
	g.Expect(ok).To(BeTrue())
	g.Expect(builder).NotTo(BeNil())

	// Test duplicate registration returns error.
	err = RegisterDetector(name, func(_ string) pipeline.Detector {
		return nil
	})
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errDetectorRegistered)).To(BeTrue())
}

func TestWriteOutput_ToFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "output.json")

	report := reportWithFindings()
	err := writeOutput(report, "json", outPath)
	g.Expect(err).NotTo(HaveOccurred())

	data, err := os.ReadFile(outPath)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data)).To(ContainSubstring(`"tool"`))
}

func TestWriteOutput_FileCreationError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()
	err := writeOutput(report, "json", "/nonexistent/dir/out.json")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("creating output file"))
}
