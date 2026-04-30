package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "json")

	parsed := parseJSON(t, &buf)

	tool, _ := parsed["tool"].(map[string]any)
	assert.NotNil(t, tool)
	assert.Equal(t, "test", tool["name"])
}

func TestOutputResults_SARIF(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "sarif")

	parsed := parseJSON(t, &buf)

	version, _ := parsed["$schema"].(string)
	assert.Contains(t, version, "sarif")
}

func TestOutputResults_Text(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "text")

	out := buf.String()
	assert.Contains(t, out, "nilcheck")
	assert.Contains(t, out, "1 finding(s)")
}

func TestOutputText_EmptyReport(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	var buf bytes.Buffer

	outputText(&buf, report)

	assert.Equal(t, "No findings.\n", buf.String())
}

func TestOutputText_WithSuggestion(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFinding(finding.Finding{
		ID: "test:R1:main.go:1:1", Rule: "R1", ToolName: "test",
		Message: "msg", Severity: finding.SeverityError,
		Position:   finding.Pos("main.go", 1, 1),
		Suggestion: "fix it",
	})

	var buf bytes.Buffer
	outputText(&buf, report)

	assert.Contains(t, buf.String(), "Suggestion: fix it")
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

	findings := []finding.Finding{
		{Severity: finding.SeverityInfo, Message: "info"},
		{Severity: finding.SeverityWarning, Message: "warn"},
		{Severity: finding.SeverityError, Message: "err"},
		{Severity: finding.SeverityCritical, Message: "crit"},
	}

	filtered := filterBySeverity(findings, finding.SeverityError)
	assert.Len(t, filtered, 2)
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

			dets := buildDetectors(tt.specs, ".")
			assert.Len(t, dets, tt.wantCount)
		})
	}
}

//nolint:paralleltest // mutates global os.Stderr
func TestFatalf(t *testing.T) {
	var buf bytes.Buffer
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	got := fatalf("testing", errors.New("test error"))

	_ = w.Close()
	os.Stderr = old

	_, _ = io.Copy(&buf, r)

	assert.Equal(t, 1, got)

	want := "Error testing: test error\n"
	assert.Equal(t, want, buf.String())
}

type failingWriter struct {
	err error
}

func (w *failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestOutputResults_WriteError(t *testing.T) {
	t.Parallel()
	report := reportWithFindings()

	err := outputResults(&failingWriter{err: errors.New("disk full")}, report, "json")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "writing JSON")

	err = outputResults(&failingWriter{err: errors.New("disk full")}, report, "sarif")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "writing SARIF")
}

func TestOutputText_WithSummary(t *testing.T) {
	t.Parallel()
	report := finding.NewReport(finding.ToolInfo{Name: "test"})
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
	assert.Contains(t, out, "By severity:")

	assert.Contains(t, out, "2 finding(s)")
}

func reportWithFindings() *finding.Report {
	report := finding.NewReport(finding.ToolInfo{Name: "test", Version: "1.0"})
	report.AddFinding(
		makeTestFinding(
			"test:nilcheck:main.go:10:5",
			"nilcheck",
			"test",
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
	stop, err := setupProfiling("/nonexistent_dir/cpu.prof", "")
	require.Error(t, err)
	assert.Nil(t, stop)
}

//nolint:paralleltest // manipulates global pprof CPU profile state
func TestSetupProfiling_CPUProfileStartFailure(t *testing.T) {
	// Create a temp file that we can't write a valid CPU profile to
	f, err := os.CreateTemp(t.TempDir(), "cpu.prof")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Start a CPU profile already, so the second StartCPUProfile fails.
	f2, err := os.CreateTemp(t.TempDir(), "cpu2.prof")
	require.NoError(t, err)
	defer func() { _ = f2.Close() }()

	require.NoError(t, pprof.StartCPUProfile(f2))
	defer pprof.StopCPUProfile()

	stop, err := setupProfiling(f.Name(), "")
	require.Error(t, err)
	assert.Nil(t, stop)
}

func TestOutputResults_JSONSerializationError(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFinding(finding.Finding{
		ID: "1", Rule: "r1", ToolName: "t", Message: "m",
		Severity: finding.SeverityError, Position: finding.Position{File: "a.go"},
		Confidence: math.NaN(),
	})
	var buf bytes.Buffer

	err := outputResults(&buf, report, "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "serializing JSON")
}

func TestOutputResults_SARIFSerializationError(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFinding(finding.Finding{
		ID: "1", Rule: "r1", ToolName: "t", Message: "m",
		Severity: finding.SeverityError, Position: finding.Position{File: "a.go"},
		Confidence: math.NaN(),
	})
	var buf bytes.Buffer

	err := outputResults(&buf, report, "sarif")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "serializing SARIF")
}

func TestRun_InvalidSeverity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	saveRestoreFlags(t)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"go-finding", "-severity=banana"}

	got := run()
	assert.Equal(t, 1, got)
}

func TestRegisterDetector(t *testing.T) {
	t.Parallel()

	name := uniqueDetName("test-detector")

	// Test registration of a new detector.
	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			"test",
			func(_ context.Context) ([]finding.Finding, error) {
				return nil, nil
			},
		)
	})
	require.NoError(t, err)

	// Verify it can be looked up.
	builder, ok := lookupDetectorBuilder(name)
	require.True(t, ok)
	require.NotNil(t, builder)

	// Test duplicate registration returns error.
	err = RegisterDetector(name, func(_ string) pipeline.Detector {
		return nil
	})
	require.Error(t, err)
	require.ErrorIs(t, err, errDetectorRegistered)
}

func TestWriteOutput_ToFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "output.json")

	report := reportWithFindings()
	err := writeOutput(report, "json", outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"tool"`)
}

func TestWriteOutput_FileCreationError(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()
	err := writeOutput(report, "json", "/nonexistent/dir/out.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating output file")
}
