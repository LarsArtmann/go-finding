package detectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// NewStaticcheckDetector returns a Detector that runs `staticcheck -f json` on the given directory.
func NewStaticcheckDetector(dir string) pipeline.Detector {
	return pipeline.NamedDetectorFunc(
		"staticcheck",
		func(ctx context.Context) ([]finding.Finding, error) {
			cmd := exec.CommandContext(ctx, "staticcheck", "-f", "json", "./...")
			cmd.Dir = dir

			out, err := cmd.Output()
			if err != nil {
				exitError := &exec.ExitError{} //nolint:exhaustruct
				if errors.As(err, &exitError) && len(out) > 0 {
					return parseStaticcheckJSON(out, dir), nil
				}

				return nil, fmt.Errorf("run staticcheck: %w", err)
			}

			return parseStaticcheckJSON(out, dir), nil
		},
	)
}

const defaultStaticcheckConfidence = 0.8

func parseStaticcheckJSON(data []byte, dir string) []finding.Finding {
	if len(data) == 0 {
		return nil
	}

	var findings []finding.Finding

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry struct {
			Code     string `json:"code"`
			Severity string `json:"severity"`
			Location struct {
				File   string `json:"file"`
				Line   int    `json:"line"`
				Column int    `json:"column"`
			} `json:"location"`
			Message string `json:"message"`
		}

		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			continue
		}

		pos := finding.Position{ //nolint:exhaustruct
			File:   entry.Location.File,
			Line:   entry.Location.Line,
			Column: entry.Location.Column,
		}
		if dir != "" && !filepath.IsAbs(pos.File) {
			pos.File = filepath.Join(dir, entry.Location.File)
		}

		sev := finding.SeverityWarning
		if entry.Severity == "error" {
			sev = finding.SeverityError
		}

		cat := staticcheckCategory(entry.Code)

		//nolint:exhaustruct
		findings = append(findings, finding.Finding{ //nolint:exhaustruct
			ID:          finding.GenerateID("staticcheck", entry.Code, pos),
			Rule:        entry.Code,
			ToolName:    "staticcheck",
			Message:     entry.Message,
			Severity:    sev,
			Position:    pos,
			Category:    cat,
			FixStrategy: finding.FixStrategySuggest,
			Confidence:  defaultStaticcheckConfidence,
		})
	}

	return findings
}

func staticcheckCategory(code string) finding.Category {
	if len(code) == 0 {
		return finding.CategoryCorrectness
	}

	switch code[0] {
	case 'S', 'Q':
		return finding.CategoryStyle
	case 'U':
		return finding.CategoryUnused
	case 'P', 'R', 'F':
		return finding.CategoryPerformance
	case 'A':
		return finding.CategoryCorrectness
	default:
		return finding.CategoryCorrectness
	}
}
