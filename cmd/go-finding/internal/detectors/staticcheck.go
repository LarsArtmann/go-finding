package detectors

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// NewStaticcheckDetector returns a Detector that runs `staticcheck -f json` on the given directory.
func NewStaticcheckDetector(dir string) pipeline.Detector {
	return pipeline.NamedDetectorFunc(
		DetectorNameStaticcheck,
		func(ctx context.Context) ([]finding.Finding, error) {
			cmd := exec.CommandContext(ctx, "staticcheck", "-f", "json", "./...")
			cmd.Dir = dir

			out, err := cmd.Output()
			if err != nil {
				if _, ok := errors.AsType[*exec.ExitError](err); ok && len(out) > 0 { //nolint:erraudit // ok-pattern type assertion
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
			// Before/After are an optional go-finding extension to the
			// staticcheck JSON line format: when both are present the finding
			// is auto-fixable (FixStrategyDirect with literal replacement)
			// instead of suggest-only. Real staticcheck output omits them.
			Before string `json:"before"`
			After  string `json:"after"`
		}

		err := json.Unmarshal([]byte(line), &entry)
		if err != nil { //nolint:erraudit // malformed staticcheck lines are intentionally skipped
			continue
		}

		pos := finding.Position{
			File:   finding.FilePath(resolvePath(dir, entry.Location.File)),
			Line:   entry.Location.Line,
			Column: entry.Location.Column,
		}

		sev := finding.SeverityWarning
		if entry.Severity == "error" {
			sev = finding.SeverityError
		}

		cat := staticcheckCategory(entry.Code)

		strategy := finding.FixStrategySuggest
		var before, after string
		if entry.Before != "" && entry.After != "" {
			strategy = finding.FixStrategyDirect
			before, after = entry.Before, entry.After
		}

		findings = append(findings, finding.Finding{
			ID:          finding.GenerateID(DetectorNameStaticcheck, finding.RuleName(entry.Code), pos),
			Rule:        finding.RuleName(entry.Code),
			ToolName:    DetectorNameStaticcheck,
			Message:     entry.Message,
			Severity:    sev,
			Position:    pos,
			Category:    cat,
			FixStrategy: strategy,
			BeforeCode:  before,
			AfterCode:   after,
			Confidence:  defaultStaticcheckConfidence,
		})
	}

	return findings
}

func staticcheckCategory(code string) finding.Category {
	if len(code) == 0 {
		return finding.CategoryCorrectness
	}

	// SA codes are static-analysis correctness checks, not style.
	if len(code) >= 2 && code[0] == 'S' && code[1] == 'A' {
		return finding.CategoryCorrectness
	}

	switch code[0] {
	case 'S', 'Q':
		return finding.CategoryStyle
	case 'U':
		return finding.CategoryUnused
	case 'P', 'R', 'F':
		return finding.CategoryPerformance
	default:
		return finding.CategoryCorrectness
	}
}
