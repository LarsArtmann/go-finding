package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

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
	if tool == nil || tool["name"] != "test" {
		t.Errorf("expected tool.name=test, got %v", tool)
	}
}

func TestOutputResults_SARIF(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "sarif")

	parsed := parseJSON(t, &buf)

	version, _ := parsed["$schema"].(string)
	if !strings.Contains(version, "sarif") {
		t.Errorf("expected $schema to contain 'sarif', got %q", version)
	}
}

func TestOutputResults_Text(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "text")

	out := buf.String()
	if !strings.Contains(out, "nilcheck") {
		t.Errorf("text output should contain rule name, got: %s", out)
	}
	if !strings.Contains(out, "1 finding(s)") {
		t.Errorf("text output should contain finding count, got: %s", out)
	}
}

func TestOutputText_EmptyReport(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	var buf bytes.Buffer

	outputText(&buf, report)

	if buf.String() != "No findings.\n" {
		t.Errorf("empty report text = %q, want %q", buf.String(), "No findings.\n")
	}
}

func TestOutputText_WithSuggestion(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFinding(finding.Finding{
		ID: "test:R1:main.go:1:1", Rule: "R1", ToolName: "test",
		Message: "msg", Severity: finding.SeverityError,
		Position:   finding.Position{File: "main.go", Line: 1, Column: 1},
		Suggestion: "fix it",
	})

	var buf bytes.Buffer
	outputText(&buf, report)

	if !strings.Contains(buf.String(), "Suggestion: fix it") {
		t.Errorf("text output should contain suggestion, got: %s", buf.String())
	}
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
	if len(filtered) != 2 {
		t.Errorf("filterBySeverity(Error) returned %d findings, want 2", len(filtered))
	}
}

func TestBuildDetectors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		specs     []detectorSpec
		wantCount int
	}{
		{
			name: "known detectors",
			specs: []detectorSpec{
				{Name: "govet"},
				{Name: "staticcheck"},
			},
			wantCount: 2,
		},
		{
			name: "unknown detector skipped",
			specs: []detectorSpec{
				{Name: "govet"},
				{Name: "nonexistent"},
			},
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
			if len(dets) != tt.wantCount {
				t.Errorf("buildDetectors returned %d, want %d", len(dets), tt.wantCount)
			}
		})
	}
}

func TestFatalf(t *testing.T) {
	t.Parallel()
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	got := fatalf("testing", errors.New("test error"))

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if got != 1 {
		t.Errorf("fatalf returned %d, want 1", got)
	}

	want := "Error testing: test error\n"
	if buf.String() != want {
		t.Errorf("fatalf output = %q, want %q", buf.String(), want)
	}
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
	if err == nil {
		t.Error("expected error for JSON write failure")
	}

	if !strings.Contains(err.Error(), "writing JSON") {
		t.Errorf("error = %v, want writing JSON", err)
	}

	err = outputResults(&failingWriter{err: errors.New("disk full")}, report, "sarif")
	if err == nil {
		t.Error("expected error for SARIF write failure")
	}

	if !strings.Contains(err.Error(), "writing SARIF") {
		t.Errorf("error = %v, want writing SARIF", err)
	}
}

func TestOutputText_WithSummary(t *testing.T) {
	t.Parallel()
	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFinding(finding.Finding{
		Severity: finding.SeverityError, Rule: "R1", Message: "err1",
		Position: finding.Position{File: "a.go", Line: 1, Column: 1},
		ID:       "test:R1:a.go:1:1",
	})
	report.AddFinding(finding.Finding{
		Severity: finding.SeverityInfo, Rule: "R2", Message: "info1",
		Position: finding.Position{File: "b.go", Line: 2, Column: 1},
		ID:       "test:R2:b.go:2:1",
	})
	report.ComputeSummary()

	var buf bytes.Buffer
	outputText(&buf, report)

	out := buf.String()
	if !strings.Contains(out, "By severity:") {
		t.Errorf("text output should contain severity summary, got: %s", out)
	}

	if !strings.Contains(out, "2 finding(s)") {
		t.Errorf("text output should contain finding count, got: %s", out)
	}
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
