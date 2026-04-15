package finding

import (
	"testing"
	"time"
)

func sevFromInt(i int) Severity {
	sevs := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	if i < 0 || i >= len(sevs) {
		return SeverityInfo
	}
	return sevs[i]
}

func testSeverityFromInt(i int) Severity {
	return sevFromInt(i)
}

func makeTestFinding(id, rule, toolName string, sev Severity) Finding {
	return Finding{
		ID:       id,
		Rule:     rule,
		ToolName: toolName,
		Message:  "test finding",
		Severity: sev,
		Position: Position{File: "test.go", Line: 1, Column: 1},
	}
}

func makeTestRange(file string, startLine, startCol, endLine, endCol int) Range {
	return Range{
		Start: Position{File: file, Line: startLine, Column: startCol},
		End:   Position{File: file, Line: endLine, Column: endCol},
	}
}

func makeTestSuppression(t *testing.T, reason string) *Suppression {
	t.Helper()
	return &Suppression{
		Kind:   SuppressionInSource,
		Reason: reason,
		ExpiresAt: func() *time.Time {
			t := time.Now().Add(1 * time.Hour)
			return &t
		}(),
	}
}
