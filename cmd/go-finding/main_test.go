package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestOutputResults_JSON(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()
	var buf bytes.Buffer

	if err := outputResults(&buf, report, "json"); err != nil {
		t.Fatalf("outputResults(json) error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, buf.String())
	}

	tool, _ := parsed["tool"].(map[string]any)
	if tool == nil || tool["name"] != "test" {
		t.Errorf("expected tool.name=test, got %v", tool)
	}
}

func TestOutputResults_SARIF(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()
	var buf bytes.Buffer

	if err := outputResults(&buf, report, "sarif"); err != nil {
		t.Fatalf("outputResults(sarif) error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid SARIF JSON: %v\noutput: %s", err, buf.String())
	}

	version, _ := parsed["$schema"].(string)
	if !strings.Contains(version, "sarif") {
		t.Errorf("expected $schema to contain 'sarif', got %q", version)
	}
}

func TestOutputResults_Text(t *testing.T) {
	t.Parallel()

	report := reportWithFindings()
	var buf bytes.Buffer

	if err := outputResults(&buf, report, "text"); err != nil {
		t.Fatalf("outputResults(text) error: %v", err)
	}

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

func reportWithFindings() *finding.Report {
	report := finding.NewReport(finding.ToolInfo{Name: "test", Version: "1.0"})
	report.AddFinding(finding.Finding{
		ID: "test:nilcheck:main.go:10:5", Rule: "nilcheck", ToolName: "test",
		Message: "possible nil dereference", Severity: finding.SeverityWarning,
		Position: finding.Position{File: "main.go", Line: 10, Column: 5},
	})
	report.ComputeSummary()

	return report
}
