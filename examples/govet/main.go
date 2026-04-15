// Package main demonstrates wrapping go vet output as a Detector
// that produces Finding structs compatible with the pipeline.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
)

// goVetOutput represents the JSON structure from go vet -json.
// Map of package -> checker -> []diagnostic.
type goVetOutput map[string]map[string][]goVetDiagnostic

type goVetDiagnostic struct {
	Posn    string `json:"posn"`
	End     string `json:"end"`
	Message string `json:"message"`
}

// GoVetDetector runs go vet and converts results to Findings.
type GoVetDetector struct {
	dir string
}

// NewGoVetDetector creates a detector that runs go vet in the given directory.
func NewGoVetDetector(dir string) *GoVetDetector {
	return &GoVetDetector{dir: dir}
}

// Name implements Detector.
func (d *GoVetDetector) Name() string {
	return "govet"
}

// Detect runs go vet -json and converts the output to Findings.
func (d *GoVetDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	cmd := exec.CommandContext(ctx, "go", "vet", "-json", "./...")
	cmd.Dir = d.dir

	output, err := cmd.Output()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
			// go vet exits non-zero when issues found — that's fine
		} else {
			return nil, fmt.Errorf("go vet: %w", err)
		}
	}

	return parseGoVetOutput(output)
}

func parseGoVetOutput(data []byte) ([]finding.Finding, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var result goVetOutput
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse go vet output: %w", err)
	}

	var findings []finding.Finding
	for pkg, checkers := range result {
		for checker, diagnostics := range checkers {
			for i, diag := range diagnostics {
				pos, end := parsePosition(diag.Posn, diag.End)
				f := finding.Finding{
					ID:          fmt.Sprintf("govet/%s/%d", checker, i),
					Rule:        checker,
					Message:     diag.Message,
					Severity:    finding.SeverityWarning,
					Category:    "go-vet",
					ToolName:    "govet",
					FixStrategy: finding.FixStrategySuggest,
					Position:    pos,
				}
				if end.Line > 0 {
					f.Range = &finding.Range{Start: pos, End: end}
				}
				parts := strings.Split(pkg, "/")
				f.Metadata = map[string]string{"package": parts[len(parts)-1]}
				findings = append(findings, f)
			}
		}
	}

	return findings, nil
}

// parsePosition parses "file:line:col" format from go vet output.
func parsePosition(posn, end string) (finding.Position, finding.Position) {
	pos := parsePosn(posn)
	endPos := parsePosn(end)
	return pos, endPos
}

func parsePosn(posn string) finding.Position {
	if posn == "" {
		return finding.Position{}
	}
	parts := strings.Split(posn, ":")
	if len(parts) < 3 {
		return finding.Position{File: posn}
	}
	line, _ := strconv.Atoi(parts[1])
	col, _ := strconv.Atoi(parts[2])
	return finding.Position{
		File:   parts[0],
		Line:   line,
		Column: col,
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

	detector := NewGoVetDetector(absDir)
	findings, err := detector.Detect(context.Background())
	if err != nil {
		log.Fatalf("detect: %v", err)
	}

	if len(findings) == 0 {
		log.New(os.Stderr, "", 0).Println("No issues found.")
		return
	}

	w := log.New(os.Stdout, "", 0)
	w.Printf("Found %d issues:", len(findings))
	for _, f := range findings {
		loc := ""
		if f.Position.File != "" {
			loc = fmt.Sprintf("%s:%d:%d: ", f.Position.File, f.Position.Line, f.Position.Column)
		}
		w.Printf("  %s%s [%s]", loc, f.Message, f.Rule)
	}
}
