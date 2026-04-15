package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/examples/detectorutil"
)

type staticcheckDiagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Pos      struct {
		Filename string `json:"filename"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Offset   int    `json:"offset"`
	} `json:"pos"`
	End struct {
		Filename string `json:"filename"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
	} `json:"end"`
}

type StaticcheckDetector struct {
	dir string
}

func NewStaticcheckDetector(dir string) *StaticcheckDetector {
	return &StaticcheckDetector{dir: dir}
}

func (d *StaticcheckDetector) Name() string {
	return "staticcheck"
}

func (d *StaticcheckDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	output, err := detectorutil.RunTool(ctx, d.dir, "staticcheck", "-f", "json", "./...")
	if err != nil {
		return nil, err
	}
	return parseStaticcheckOutput(output)
}

func parseStaticcheckOutput(data []byte) ([]finding.Finding, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var diagnostics []staticcheckDiagnostic
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var diag staticcheckDiagnostic
		if err := json.Unmarshal([]byte(line), &diag); err != nil {
			continue
		}
		diagnostics = append(diagnostics, diag)
	}

	findings := make([]finding.Finding, 0, len(diagnostics))
	for i, diag := range diagnostics {
		f := finding.Finding{
			ID:          fmt.Sprintf("staticcheck:%s:%s:%d:%d", diag.Code, diag.Pos.Filename, diag.Pos.Line, diag.Pos.Column),
			Rule:        diag.Code,
			ToolName:    "staticcheck",
			Message:     diag.Message,
			Severity:    mapStaticcheckSeverity(diag.Severity),
			Position:    finding.Position{File: diag.Pos.Filename, Line: diag.Pos.Line, Column: diag.Pos.Column, Offset: diag.Pos.Offset},
			Category:    mapStaticcheckCategory(diag.Code),
			FixStrategy: finding.FixStrategySuggest,
			Confidence:  0.8,
		}
		if diag.End.Line > 0 {
			f.Range = &finding.Range{
				Start: finding.Position{File: diag.Pos.Filename, Line: diag.Pos.Line, Column: diag.Pos.Column},
				End:   finding.Position{File: diag.End.Filename, Line: diag.End.Line, Column: diag.End.Column},
			}
		}
		f.Metadata = map[string]string{"index": strconv.Itoa(i)}
		findings = append(findings, f)
	}

	return findings, nil
}

func mapStaticcheckSeverity(sev string) finding.Severity {
	switch sev {
	case "error":
		return finding.SeverityError
	case "warning", "deprecated":
		return finding.SeverityWarning
	case "info", "ignored":
		return finding.SeverityInfo
	default:
		return finding.SeverityWarning
	}
}

func mapStaticcheckCategory(code string) finding.Category {
	if len(code) == 0 {
		return ""
	}
	switch code[0] {
	case 'S': // style, simplicity
		return finding.CategoryStyle
	case 'U': // unused
		return finding.CategoryCorrectness
	case 'Q': // correctness
		return finding.CategoryCorrectness
	case 'A': // performance
		return finding.CategoryPerformance
	case 'T': // testing
		return finding.CategoryTesting
	default:
		return ""
	}
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolve path: %v", err)
	}

	detector := NewStaticcheckDetector(absDir)

	findings, err := detector.Detect(context.Background())
	if err != nil {
		log.Fatalf("detect: %v", err)
	}

	if len(findings) == 0 {
		fmt.Println("No issues found.")
		return
	}

	fmt.Printf("Found %d issues:\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  %s: %s [%s] (%s)\n", f.Position, f.Message, f.Rule, f.Severity)
	}
}
