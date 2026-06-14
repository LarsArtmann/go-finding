package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

func TestOutputResults_AllFormats(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "integration-test", Version: "1.0"})
	report.AddFinding(finding.Finding{
		ID:          testToolName + ":R1:main.go:1:1",
		Rule:        "R1",
		ToolName:    testToolName,
		Message:     "integration test finding",
		Severity:    finding.SeverityError,
		Position:    finding.Pos("main.go", 1, 1),
		Category:    finding.CategoryCorrectness,
		FixStrategy: finding.FixStrategySuggest,
		Suggestion:  "fix the issue",
	})
	report.ComputeSummary()

	for _, format := range []string{"text", "json", "sarif"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			if err := outputResults(&buf, report, format); err != nil {
				t.Fatalf("outputResults(%s) error: %v", format, err)
			}

			if buf.Len() == 0 {
				t.Errorf("outputResults(%s) produced empty output", format)
			}
		})
	}
}

func TestOutputResults_JSONContainsFindings(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "json")

	var parsed struct {
		Findings []struct {
			Rule string `json:"rule"`
		} `json:"findings"`
	}

	requireJSON(t, &buf, &parsed, "JSON parse error")

	g.Expect(parsed.Findings).To(HaveLen(1))
	g.Expect(parsed.Findings[0].Rule).To(Equal("nilcheck"))
}

func TestOutputResults_SARIFContainsResults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer

	requireOutputResults(t, &buf, report, "sarif")

	var parsed struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []struct {
				RuleID string `json:"ruleId"`
			} `json:"results"`
		} `json:"runs"`
	}

	requireJSON(t, &buf, &parsed, "SARIF parse error")

	g.Expect(parsed.Version).To(Equal("2.1.0"))
	g.Expect(parsed.Runs).To(HaveLen(1))
	g.Expect(parsed.Runs[0].Results).To(HaveLen(1))
	g.Expect(parsed.Runs[0].Results[0].RuleID).To(Equal("nilcheck"))
}

func assertRunFails(t *testing.T, args ...string) {
	t.Helper()
	if got := runWithArgs(t, args...); got != 1 {
		t.Errorf("run() with args %v = %d, want 1", args, got)
	}
}

func requireJSON[T any](t *testing.T, buf *bytes.Buffer, parsed *T, context string) {
	t.Helper()
	if err := json.Unmarshal(buf.Bytes(), parsed); err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}

//nolint:paralleltest // manipulates global flag state
func TestRun_NegativeMaxIterations(t *testing.T) {
	assertRunFails(t, "-max-iterations", "-1", "-dir", t.TempDir())
}

//nolint:paralleltest // manipulates global flag state
func TestRun_PipelineRunError(t *testing.T) {
	g := NewWithT(t)
	name := uniqueDetName("broken-fix")

	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			"broken",
			func(_ context.Context) ([]finding.Finding, error) {
				return []finding.Finding{{
					ID:          "broken:R1:main.go:1:1",
					Rule:        "R1",
					ToolName:    "broken",
					Message:     "broken fix",
					Severity:    finding.SeverityError,
					Position:    finding.Position{File: "nonexistent.go", Line: 1, Column: 1},
					FixStrategy: finding.FixStrategyDirect,
					BeforeCode:  "old",
					AfterCode:   "new",
				}}, nil
			},
		)
	})
	g.Expect(err).NotTo(HaveOccurred())

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfig(t, cfgPath, []byte(
		"maxIterations: 1\ntimeout: 30s\ndetectors:\n  - name: "+name+"\n",
	))

	assertRunFails(t, "-config", cfgPath, "-dir", dir)
}

//nolint:paralleltest // manipulates global flag state
func TestRun_MetricsOutput(t *testing.T) {
	g := NewWithT(t)
	name := uniqueDetName("always-find")

	err := RegisterDetector(name, func(_ string) pipeline.Detector {
		return pipeline.NamedDetectorFunc(
			"find",
			func(_ context.Context) ([]finding.Finding, error) {
				return []finding.Finding{{
					ID:       "find:R1:main.go:1:1",
					Rule:     "R1",
					ToolName: "find",
					Message:  "test finding",
					Severity: finding.SeverityError,
					Position: finding.Position{File: "main.go", Line: 1, Column: 1},
				}}, nil
			},
		)
	})
	g.Expect(err).NotTo(HaveOccurred())

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfig(t, cfgPath, []byte(
		"maxIterations: 1\ntimeout: 30s\ndetectors:\n  - name: "+name+"\n",
	))

	var buf bytes.Buffer
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	saveRestoreFlags(t)
	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-config", cfgPath, "-dir", dir}

	got := run()

	_ = w.Close()
	os.Stderr = old
	_, _ = io.Copy(&buf, r)

	g.Expect(got).To(Equal(0))
	g.Expect(buf.String()).To(ContainSubstring("Metrics:"))
}

//nolint:paralleltest // manipulates global flag state via assertRunFails
func TestRun_BadSeverity(t *testing.T) {
	assertRunFails(t, "-severity", "bogus")
}

//nolint:paralleltest // manipulates global flag state via assertRunFails
func TestRun_MissingConfig(t *testing.T) {
	assertRunFails(t, "-config", "/nonexistent/config.yaml")
}

//nolint:paralleltest // manipulates global flag state
func TestRun_NoDetectors(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
maxIterations: 1
timeout: "30s"
detectors: []
`
	writeConfig(t, cfgPath, []byte(content))

	saveRestoreFlags(t)

	flag.CommandLine = flag.NewFlagSet("go-finding", flag.ContinueOnError)
	os.Args = []string{"go-finding", "-config", cfgPath, "-dir", dir}

	got := run()
	if got != 1 {
		t.Errorf("run() with no detectors = %d, want 1", got)
	}
}
