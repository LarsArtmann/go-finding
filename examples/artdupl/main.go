package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/larsartmann/go-finding"
)

type duplClone struct {
	Name      string       `json:"name"`
	FirstDup  duplFragment `json:"first"`
	SecondDup duplFragment `json:"second"`
}

type duplFragment struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	EndLine   int    `json:"end_line"`
	Content   string `json:"content"`
}

type ArtduplDetector struct {
	dir string
}

func NewArtduplDetector(dir string) *ArtduplDetector {
	return &ArtduplDetector{dir: dir}
}

func (d *ArtduplDetector) Name() string {
	return "artdupl"
}

func (d *ArtduplDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	cmd := exec.CommandContext(ctx, "art-dupl", "-json", "-t", "50", "./...")
	cmd.Dir = d.dir

	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// art-dupl exits non-zero when duplicates found
		} else {
			return nil, fmt.Errorf("art-dupl: %w", err)
		}
	}

	return parseArtduplOutput(output)
}

func parseArtduplOutput(data []byte) ([]finding.Finding, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var clones []duplClone
	if err := json.Unmarshal(data, &clones); err != nil {
		return nil, fmt.Errorf("parse art-dupl output: %w", err)
	}

	findings := make([]finding.Finding, 0, len(clones))
	for i, clone := range clones {
		f := finding.Finding{
			ID:          fmt.Sprintf("artdupl:%s:%d:%d", clone.FirstDup.Path, clone.FirstDup.Line, clone.SecondDup.Line),
			Rule:        "duplicate-code",
			ToolName:    "artdupl",
			Message:     fmt.Sprintf("Duplicate code detected in %s:%d and %s:%d", clone.FirstDup.Path, clone.FirstDup.Line, clone.SecondDup.Path, clone.SecondDup.Line),
			Severity:    finding.SeverityInfo,
			Category:    finding.CategoryDuplication,
			Position:    finding.Position{File: clone.FirstDup.Path, Line: clone.FirstDup.Line},
			FixStrategy: finding.FixStrategySuggest,
			Confidence:  0.7,
			Suggestion:  "Consider extracting the duplicated code into a shared function or method.",
		}
		if clone.FirstDup.EndLine > clone.FirstDup.Line {
			f.Range = &finding.Range{
				Start: finding.Position{File: clone.FirstDup.Path, Line: clone.FirstDup.Line},
				End:   finding.Position{File: clone.FirstDup.Path, Line: clone.FirstDup.EndLine},
			}
		}
		f.Related = []finding.RelatedRef{
			{
				FindingID: fmt.Sprintf("artdupl:%s:%d", clone.SecondDup.Path, clone.SecondDup.Line),
				Relation:  "duplicate-of",
				Position:  finding.Position{File: clone.SecondDup.Path, Line: clone.SecondDup.Line},
			},
		}
		f.Metadata = map[string]string{
			"clone_index": strconv.Itoa(i),
			"clone_name":  clone.Name,
		}
		findings = append(findings, f)
	}

	return findings, nil
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

	detector := NewArtduplDetector(absDir)
	findings, err := detector.Detect(context.Background())
	if err != nil {
		log.Fatalf("detect: %v", err)
	}

	if len(findings) == 0 {
		fmt.Println("No duplicates found.")
		return
	}

	fmt.Printf("Found %d duplicate code blocks:\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  %s:%d: %s [%s]\n", f.Position.File, f.Position.Line, f.Message, f.Severity)
		if len(f.Related) > 0 {
			for _, r := range f.Related {
				fmt.Printf("    duplicate in %s:%d (%s)\n", r.Position.File, r.Position.Line, r.Relation)
			}
		}
	}
}
