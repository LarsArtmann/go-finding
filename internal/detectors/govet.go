// Package detectors provides built-in detector implementations
// that wrap external static analysis tools.
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

// NewGoVetDetector returns a Detector that runs `go vet -json` on the given directory.
func NewGoVetDetector(dir string) pipeline.Detector {
	return pipeline.NamedDetectorFunc(
		"govet",
		func(ctx context.Context) ([]finding.Finding, error) {
			cmd := exec.CommandContext(ctx, "go", "vet", "-json", "./...")
			cmd.Dir = dir
			cmd.Stderr = nil

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
	var diagnostics map[string]json.RawMessage

	err := json.Unmarshal(data, &diagnostics)
	if err != nil {
		return nil
	}

	var findings []finding.Finding

	for _, raw := range diagnostics {
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
			findings = append(findings, finding.Finding{
				ID:          finding.GenerateID("govet", "", pos),
				ToolName:    "govet",
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
		return finding.Position{File: posn}
	}

	pos := finding.Position{File: parts[0]}
	if dir != "" && !filepath.IsAbs(pos.File) {
		pos.File = filepath.Join(dir, parts[0])
	}

	if len(parts) >= 2 {
		_, _ = fmt.Sscanf(parts[1], "%d", &pos.Line)
	}

	if len(parts) >= 3 {
		_, _ = fmt.Sscanf(parts[2], "%d", &pos.Column)
	}

	return pos
}
