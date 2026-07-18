package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json/v2"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	for _, format := range []string{"text", "markdown", "csv", "tsv", "json", "sarif"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			if err := outputResults(&buf, report, format, true); err != nil {
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

func TestOutputResults_CSV(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer
	requireOutputResults(t, &buf, report, "csv")

	records, err := csv.NewReader(&buf).ReadAll()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(records).To(HaveLen(2)) // header + 1 data row, no footer

	g.Expect(records[0]).To(Equal([]string{"Location", "Severity", "Category", "Rule", "Message", "Fix"}))
	g.Expect(records[1][3]).To(Equal("nilcheck"))
}

func TestOutputResults_TSV(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer
	requireOutputResults(t, &buf, report, "tsv")

	reader := csv.NewReader(&buf)
	reader.Comma = '\t'

	records, err := reader.ReadAll()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(records).To(HaveLen(2)) // header + 1 data row, no footer

	g.Expect(records[0]).To(Equal([]string{"Location", "Severity", "Category", "Rule", "Message", "Fix"}))
	g.Expect(records[1][3]).To(Equal("nilcheck"))
}

func TestOutputResults_MarkdownHasAlignedTable(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	report := reportWithFindings()

	var buf bytes.Buffer
	requireOutputResults(t, &buf, report, "markdown")

	out := buf.String()
	g.Expect(out).To(ContainSubstring("|"))
	g.Expect(out).To(ContainSubstring("Location"))
	g.Expect(out).To(ContainSubstring("Severity"))
	g.Expect(out).To(ContainSubstring("Rule"))
	g.Expect(out).To(ContainSubstring("Message"))
	g.Expect(out).To(ContainSubstring("nilcheck"))
	g.Expect(out).To(ContainSubstring("Findings (1)"))
}

func TestOutputResults_GoOutputEmptyFindings(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "empty", Version: "1.0"})
	report.ComputeSummary()

	for _, format := range []string{"markdown", "csv", "tsv"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			if err := outputResults(&buf, report, format, true); err != nil {
				t.Fatalf("outputResults(%s) error: %v", format, err)
			}

			if buf.Len() != 0 {
				t.Errorf("outputResults(%s) with no findings should produce empty output, got %q", format, buf.String())
			}
		})
	}
}

func TestFindingToTableData(t *testing.T) {
	t.Parallel()

	findings := []finding.Finding{
		{
			Rule:        "R1",
			Message:     "first issue",
			Severity:    finding.SeverityError,
			Position:    finding.Pos("main.go", 10, 5),
			Category:    finding.CategorySecurity,
			FixStrategy: finding.FixStrategyDirect,
		},
		{
			Rule:        "R2",
			Message:     "second issue",
			Severity:    finding.SeverityWarning,
			Position:    finding.Pos("util.go", 3, 1),
			Category:    finding.CategoryStyle,
			FixStrategy: finding.FixStrategySuggest,
		},
	}

	data := findingToTableData(findings)

	if got := data.RowCount(); got != 2 {
		t.Errorf("RowCount = %d, want 2", got)
	}

	if got := data.ColCount(); got != 6 {
		t.Errorf("ColCount = %d, want 6", got)
	}

	headers := data.GetHeaders()
	want := []string{"Location", "Severity", "Category", "Rule", "Message", "Fix"}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("header[%d] = %q, want %q", i, headers[i], h)
		}
	}

	rows := data.GetRows()
	if rows[0][3] != "R1" || rows[1][3] != "R2" {
		t.Errorf("rule column = %v / %v, want R1 / R2", rows[0][3], rows[1][3])
	}

	if rows[0][1] != "ERROR" || rows[1][1] != "WARNING" {
		t.Errorf("severity column = %v / %v, want ERROR / WARNING", rows[0][1], rows[1][1])
	}

	if rows[0][2] != "security" || rows[1][2] != "style" {
		t.Errorf("category column = %v / %v, want security / style", rows[0][2], rows[1][2])
	}

	if rows[0][5] != "direct" || rows[1][5] != "suggest" {
		t.Errorf("fix column = %v / %v, want direct / suggest", rows[0][5], rows[1][5])
	}
}

func TestOutputResults_UnknownFormat(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test", Version: "1.0"})
	report.ComputeSummary()

	var buf bytes.Buffer
	err := outputResults(&buf, report, "xml", true)
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error should mention 'unsupported format', got: %v", err)
	}

	if !strings.Contains(err.Error(), "text") {
		t.Errorf("error should list supported formats, got: %v", err)
	}
}
