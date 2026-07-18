package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

var testDetCounter atomic.Int64

func uniqueDetName(prefix string) string {
	return prefix + "-" + strconv.FormatInt(testDetCounter.Add(1), 10)
}

func parseJSON(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, buf.String())
	}
	return parsed
}

func TestOutputResults_JSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "json")

	parsed := parseJSON(t, &buf)

	tool, _ := parsed["tool"].(map[string]any)
	g.Expect(tool).NotTo(BeNil())
	g.Expect(tool["name"]).To(Equal(testToolName))
}

func TestOutputResults_SARIF(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "sarif")

	parsed := parseJSON(t, &buf)

	version, _ := parsed["$schema"].(string)
	g.Expect(version).To(ContainSubstring("sarif"))
}

func TestOutputResults_Text(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "text")

	out := buf.String()
	g.Expect(out).To(ContainSubstring("nilcheck"))
	g.Expect(out).To(ContainSubstring("1 finding(s)"))
}

func TestOutputText_EmptyReport(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := finding.NewReport(finding.ToolInfo{Name: testToolName})
	var buf bytes.Buffer

	outputText(&buf, report)

	g.Expect(buf.String()).To(Equal("No findings.\n"))
}

func TestOutputText_WithSuggestion(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := finding.NewReport(finding.ToolInfo{Name: testToolName})
	report.AddFinding(finding.Finding{
		ID: testToolName + ":R1:main.go:1:1", Rule: "R1", ToolName: testToolName,
		Message: "msg", Severity: finding.SeverityError,
		Position:   finding.Pos("main.go", 1, 1),
		Suggestion: "fix it",
	})

	var buf bytes.Buffer
	outputText(&buf, report)

	g.Expect(buf.String()).To(ContainSubstring("Suggestion: fix it"))
}

func TestParseSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  finding.Severity
		ok    bool
	}{
		{"info", finding.SeverityInfo, true},
		{"warning", finding.SeverityWarning, true},
		{"error", finding.SeverityError, true},
		{"critical", finding.SeverityCritical, true},
		{"unknown", finding.SeverityWarning, false},
		{"", finding.SeverityWarning, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got, err := parseSeverity(tt.input)
			if tt.ok && err != nil {
				t.Fatalf("parseSeverity(%q) error: %v", tt.input, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("parseSeverity(%q) expected error, got nil", tt.input)
			}
			if got != tt.want {
				t.Errorf("parseSeverity(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterBySeverity(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	findings := []finding.Finding{
		{Severity: finding.SeverityInfo, Message: "info"},
		{Severity: finding.SeverityWarning, Message: "warn"},
		{Severity: finding.SeverityError, Message: "err"},
		{Severity: finding.SeverityCritical, Message: "crit"},
	}

	filtered := filterBySeverity(findings, finding.SeverityError)
	g.Expect(filtered).To(HaveLen(2))
}

func TestBuildDetectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		specs     []detectorSpec
		wantCount int
	}{
		{
			name:      "known detectors",
			specs:     detectorSpecs("govet", "staticcheck"),
			wantCount: 2,
		},
		{
			name:      "unknown detector skipped",
			specs:     detectorSpecs("govet", "nonexistent"),
			wantCount: 1,
		},
		{
			name:      "empty specs",
			specs:     nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			dets := buildDetectors(tt.specs, ".")
			g.Expect(dets).To(HaveLen(tt.wantCount))
		})
	}
}

//nolint:paralleltest // mutates global os.Stderr
func TestFatalf(t *testing.T) {
	g := NewWithT(t)
	var buf bytes.Buffer
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	got := fatalf("testing", errors.New("test error"))

	_ = w.Close()
	os.Stderr = old

	_, _ = io.Copy(&buf, r)

	g.Expect(got).To(Equal(1))

	want := "Error testing: test error\n"
	g.Expect(buf.String()).To(Equal(want))
}

type failingWriter struct {
	err error
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestOutputResults_WriteError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	report := reportWithFindings()

	err := outputResults(&failingWriter{err: errors.New("disk full")}, report, "json", true)
	g.Expect(err).To(HaveOccurred())

	g.Expect(err.Error()).To(ContainSubstring("writing JSON"))

	err = outputResults(&failingWriter{err: errors.New("disk full")}, report, "sarif", true)
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
		ID:       finding.ID(id),
		Rule:     finding.RuleName(rule),
		ToolName: finding.ToolName(tool),
		Message:  msg,
		Severity: sev,
		Position: finding.Position{File: finding.FilePath(file), Line: line, Column: col},
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
	err := outputResults(&buf, reportWithNaNConfidence(), format, true)
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
	err := outputResults(&buf, reportWithNaNConfidence(), "sarif", true)
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
	err := writeOutput(report, "json", outPath, true)
	g.Expect(err).NotTo(HaveOccurred())

	data, err := os.ReadFile(outPath)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data)).To(ContainSubstring(`"tool"`))
}

func TestWriteOutput_FileCreationError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()
	err := writeOutput(report, "json", "/nonexistent/dir/out.json", true)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("creating output file"))
}

func TestResolveFixProviders_UnknownReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	_, err := resolveFixProviders([]string{"nonexistent-provider"})
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, ErrUnknownFixProvider)).To(BeTrue())
}

func TestResolveFixProviders_KnownAndDefaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	providers, err := resolveFixProviders([]string{"go-ast"})
	g.Expect(err).NotTo(HaveOccurred())
	// go-ast + 3 default providers (offset, line, substring)
	g.Expect(providers).To(HaveLen(4))
	g.Expect(providers[0].Name()).To(Equal("go-ast"))
}

func TestResolveFixProviders_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	providers, err := resolveFixProviders(nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(providers).To(BeNil())
}

func TestToPipelineConfig_BadTimeout(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{Timeout: "not-a-duration"}
	_, err := cfg.toPipelineConfig()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("parse timeout"))
}

func TestToPipelineConfig_BadDetectorTimeout(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{
		DetectorTimeouts: map[string]string{"govet": "bad"},
	}
	_, err := cfg.toPipelineConfig()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("parse detector timeout"))
}

func TestToPipelineConfig_Success(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{
		MaxIterations:    3,
		Timeout:          "30s",
		DetectorTimeouts: map[string]string{"govet": "5s"},
	}
	pc, err := cfg.toPipelineConfig()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(pc.MaxIterations).To(Equal(3))
	g.Expect(pc.DetectorTimeouts).To(HaveKey("govet"))
}

func TestOutputResults_Markdown(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	var buf bytes.Buffer
	report := reportWithFindings()

	err := outputResults(&buf, report, "markdown", true)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).To(ContainSubstring("nil dereference"))
}

func TestValidate_InvalidMaxIterations(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{MaxIterations: -1}
	err := cfg.validate()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("maxIterations"))
}
