package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/go-finding"
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
