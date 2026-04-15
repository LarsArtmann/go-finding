package testutil

import (
	"testing"
	"time"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

func SeverityFromInt(i int) Severity {
	sevs := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	if i < 0 || i >= len(sevs) {
		return SeverityInfo
	}
	return sevs[i]
}

type Finding struct {
	ID       string
	Rule     string
	ToolName string
	Message  string
	Severity Severity
	Position Position
}

type Position struct {
	File  string
	Line  int
	Column int
}

type Range struct {
	Start Position
	End   Position
}

func MakeTestFinding(id, rule, toolName string, sev Severity) Finding {
	return Finding{
		ID:       id,
		Rule:     rule,
		ToolName: toolName,
		Message:  "test finding",
		Severity: sev,
		Position: Position{File: "test.go", Line: 1, Column: 1},
	}
}

func MakeSimpleFinding(id string) Finding {
	return Finding{
		ID:       id,
		Rule:     "test-rule",
		ToolName: "test-tool",
		Message:  "test message",
		Severity: SeverityError,
		Position: Position{File: "test.go", Line: 1, Column: 1},
	}
}

func MakeTestRange(file string, startLine, startCol, endLine, endCol int) Range {
	return Range{
		Start: Position{File: file, Line: startLine, Column: startCol},
		End:   Position{File: file, Line: endLine, Column: endCol},
	}
}

type SuppressionKind int

const SuppressionInSource SuppressionKind = iota

type Suppression struct {
	Kind      SuppressionKind
	Reason    string
	ExpiresAt *time.Time
}

func MakeTestSuppression(t *testing.T, reason string) *Suppression {
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

func MakeFindingsWithSeverity(ids []string, sevs []Severity) []Finding {
	fs := make([]Finding, len(ids))
	for i := range ids {
		fs[i] = Finding{
			ID:       ids[i],
			Severity: sevs[i],
		}
	}
	return fs
}
