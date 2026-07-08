// Package detectors provides built-in detector implementations
// that wrap external static analysis tools.
package detectors

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// NewGoVetDetector returns a Detector that runs `go vet -json` on the given directory.
func NewGoVetDetector(dir string) pipeline.Detector {
	return pipeline.NamedDetectorFunc(
		DetectorNameGovet,
		func(ctx context.Context) ([]finding.Finding, error) {
			cmd := exec.CommandContext(ctx, "go", "vet", "-json", "./...")
			cmd.Dir = dir

			out, err := cmd.Output()
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) && len(out) > 0 {
					return parseGoVetJSON(out, dir), nil
				}

				return nil, fmt.Errorf("run go vet: %w", err)
			}

			return parseGoVetJSON(out, dir), nil
		},
	)
}

func parseGoVetJSON(data []byte, dir string) []finding.Finding {
	var diagnostics map[string]jsontext.Value

	err := json.Unmarshal(data, &diagnostics)
	if err != nil {
		return nil
	}

	var findings []finding.Finding

	for name, raw := range diagnostics {
		var entries []struct {
			Posn    string `json:"posn"`
			Message string `json:"message"`
		}

		err := json.Unmarshal(raw, &entries)
		if err != nil {
			continue
		}

		for _, e := range entries {
			pos := parsePosn(e.Posn, dir)
			findings = append(findings, finding.Finding{ //nolint:exhaustruct
				ID:          finding.GenerateID(DetectorNameGovet, finding.RuleName(name), pos),
				Rule:        finding.RuleName(name),
				ToolName:    DetectorNameGovet,
				Message:     e.Message,
				Severity:    finding.SeverityWarning,
				Position:    pos,
				Category:    finding.CategoryCorrectness,
				FixStrategy: finding.FixStrategySuggest,
			})
		}
	}

	return findings
}

func parsePosn(posn, dir string) finding.Position {
	posnFieldCount := 4

	parts := strings.SplitN(posn, ":", posnFieldCount)
	if len(parts) < 2 {
		return finding.Position{File: finding.FilePath(posn)} //nolint:exhaustruct
	}

	pos := finding.Position{File: finding.FilePath(resolvePath(dir, parts[0]))} //nolint:exhaustruct

	if len(parts) >= 2 {
		line, err := strconv.Atoi(parts[1])
		if err == nil {
			pos.Line = line
		}
	}

	if len(parts) >= 3 {
		col, err := strconv.Atoi(parts[2])
		if err == nil {
			pos.Column = col
		}
	}

	return pos
}
