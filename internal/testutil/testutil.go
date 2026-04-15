package testutil

import (
	"testing"
	"time"

	"github.com/lane-ax/finding"
)

func SeverityFromInt(i int) finding.Severity {
	sevs := []finding.Severity{finding.SeverityInfo, finding.SeverityWarning, finding.SeverityError, finding.SeverityCritical}
	if i < 0 || i >= len(sevs) {
		return finding.SeverityInfo
	}
	return sevs[i]
}

func MakeTestFinding(id, rule, toolName string, sev finding.Severity) finding.Finding {
	return finding.Finding{
		ID:       id,
		Rule:     rule,
		ToolName: toolName,
		Message:  "test finding",
		Severity: sev,
		Position: finding.Position{File: "test.go", Line: 1, Column: 1},
	}
}

func MakeSimpleFinding(id string) finding.Finding {
	return finding.Finding{
		ID:       id,
		Rule:     "test-rule",
		ToolName: "test-tool",
		Message:  "test message",
		Severity: finding.SeverityError,
		Position: finding.Position{File: "test.go", Line: 1, Column: 1},
	}
}

func MakeTestRange(file string, startLine, startCol, endLine, endCol int) finding.Range {
	return finding.Range{
		Start: finding.Position{File: file, Line: startLine, Column: startCol},
		End:   finding.Position{File: file, Line: endLine, Column: endCol},
	}
}

func MakeTestSuppression(t *testing.T, reason string) *finding.Suppression {
	t.Helper()
	return &finding.Suppression{
		Kind:   finding.SuppressionInSource,
		Reason: reason,
		ExpiresAt: func() *time.Time {
			t := time.Now().Add(1 * time.Hour)
			return &t
		}(),
	}
}

func MakeFindingsWithSeverity(ids []string, sevs []finding.Severity) []finding.Finding {
	fs := make([]finding.Finding, len(ids))
	for i := range ids {
		fs[i] = finding.Finding{
			ID:       ids[i],
			Severity: sevs[i],
		}
	}
	return fs
}
