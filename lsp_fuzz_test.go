package finding

import (
	"context"
	"testing"
	"time"
)

func FuzzLSPRoundTrip(f *testing.F) {
	f.Add("main.go", "rule1", "tool1", "message", 10, 5)
	f.Fuzz(func(t *testing.T, file, rule, tool, msg string, line, col int) {
		if line <= 0 || col <= 0 || rule == "" || tool == "" {
			return
		}

		original := Finding{
			ID: GenerateID(ToolName(tool), RuleName(rule),
				Position{File: FilePath(file), Line: line, Column: col}),
			Rule:        RuleName(rule),
			ToolName:    ToolName(tool),
			Message:     msg,
			Severity:    SeverityWarning,
			Position:    Position{File: FilePath(file), Line: line, Column: col},
			FixStrategy: FixStrategyNone,
		}

		diag := original.ToLSP()
		restored := FromLSP(FilePath(file), diag)

		if restored.Rule != original.Rule {
			t.Errorf("Rule: got %q, want %q", restored.Rule, original.Rule)
		}

		if restored.ToolName != original.ToolName {
			t.Errorf("ToolName: got %q, want %q", restored.ToolName, original.ToolName)
		}

		if restored.Message != original.Message {
			t.Errorf("Message: got %q, want %q", restored.Message, original.Message)
		}

		if restored.Position.Line != original.Position.Line {
			t.Errorf("Line: got %d, want %d", restored.Position.Line, original.Position.Line)
		}

		if restored.Position.Column != original.Position.Column {
			t.Errorf("Column: got %d, want %d", restored.Position.Column, original.Position.Column)
		}

		if restored.ID != original.ID {
			t.Errorf("ID: got %q, want %q", restored.ID, original.ID)
		}
	})
}

func TestSARIFSuppressionExpiryRoundTrip(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	report := NewReport(ToolInfo{Name: "test", Version: "1.0"})
	report.AddFinding(Finding{
		Rule:        "r1",
		ToolName:    "test",
		Message:     "test",
		Severity:    SeverityWarning,
		Position:    Position{File: FilePath("a.go"), Line: 1},
		FixStrategy: FixStrategyNone,
		Suppression: &Suppression{
			Kind:      SuppressionInConfig,
			Rule:      "r1",
			Reason:    "config override",
			ExpiresAt: &expiry,
		},
	})

	data, err := report.ToSARIFWithOpts(WithIncludeSuppressed())
	if err != nil {
		t.Fatalf("ToSARIFWithOpts: %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), data)
	if err != nil {
		t.Fatalf("FindingsFromSARIF: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.Suppression == nil {
		t.Fatal("expected suppression to be preserved")
	}

	if f.Suppression.Kind != SuppressionInConfig {
		t.Errorf("Kind: got %q, want %q", f.Suppression.Kind, SuppressionInConfig)
	}

	if f.Suppression.Reason != "config override" {
		t.Errorf("Reason: got %q, want %q", f.Suppression.Reason, "config override")
	}

	if f.Suppression.Rule != "r1" {
		t.Errorf("Rule: got %q, want %q", f.Suppression.Rule, "r1")
	}

	if f.Suppression.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be preserved")
	}

	if !f.Suppression.ExpiresAt.Equal(expiry) {
		t.Errorf("ExpiresAt: got %v, want %v", f.Suppression.ExpiresAt, expiry)
	}
}

func TestRangeSameFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    Range
		want bool
	}{
		{
			name: "same file",
			r: Range{
				Start: Position{File: FilePath("a.go"), Line: 1},
				End:   Position{File: FilePath("a.go"), Line: 5},
			},
			want: true,
		},
		{
			name: "empty end file",
			r: Range{
				Start: Position{File: FilePath("a.go"), Line: 1},
				End:   Position{Line: 5},
			},
			want: true,
		},
		{
			name: "different files",
			r: Range{
				Start: Position{File: FilePath("a.go"), Line: 1},
				End:   Position{File: FilePath("b.go"), Line: 5},
			},
			want: false,
		},
		{
			name: "both empty",
			r:    Range{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.r.SameFile(); got != tt.want {
				t.Errorf("SameFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
