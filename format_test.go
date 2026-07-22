package finding

import (
	"strings"
	"testing"
)

func TestFormatText(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			ID: "1", Rule: "nilcheck", ToolName: "govet", Message: "possible nil deref",
			Severity: SeverityError, Position: Pos("main.go", 42, 5),
			Category:   CategorySecurity,
			Suggestion: "add nil check",
		},
	}

	var buf strings.Builder
	FormatText(&buf, findings) //nolint:errcheck

	output := buf.String()
	if !strings.Contains(output, "main.go:42:5") {
		t.Errorf("missing position in output: %q", output)
	}

	if !strings.Contains(output, "🟠 ERROR") {
		t.Errorf("missing severity badge in output: %q", output)
	}

	if !strings.Contains(output, "nilcheck") {
		t.Errorf("missing rule in output: %q", output)
	}

	if !strings.Contains(output, "[security]") {
		t.Errorf("missing category in output: %q", output)
	}

	if !strings.Contains(output, "💡") {
		t.Errorf("missing suggestion indicator in output: %q", output)
	}

	if !strings.Contains(output, "add nil check") {
		t.Errorf("missing suggestion text in output: %q", output)
	}
}

func TestFormatText_NoCategoryNoSuggestion(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			ID: "1", Rule: "unused", ToolName: "test", Message: "x is unused",
			Severity: SeverityInfo, Position: Pos("a.go", 1, 1),
		},
	}

	var buf strings.Builder
	FormatText(&buf, findings) //nolint:errcheck

	output := buf.String()
	if strings.Contains(output, "[") {
		t.Errorf("unexpected category bracket in output: %q", output)
	}

	if strings.Contains(output, "💡") {
		t.Errorf("unexpected suggestion in output: %q", output)
	}
}

func TestFormatMarkdown(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			ID: "1", Rule: "nilcheck", ToolName: "govet", Message: "possible nil deref",
			Severity: SeverityError, Position: Pos("main.go", 42, 5),
		},
	}

	var buf strings.Builder
	FormatMarkdown(&buf, findings) //nolint:errcheck

	output := buf.String()
	if !strings.Contains(output, "| main.go:42:5 | error | nilcheck | possible nil deref |") {
		t.Errorf("unexpected markdown output: %q", output)
	}
}

func TestFormatTable(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			ID: "1", Rule: "nilcheck", ToolName: "govet", Message: "possible nil deref",
			Severity: SeverityError, Position: Pos("main.go", 42, 5),
			Category: CategorySecurity,
		},
		{
			ID: "2", Rule: "unused", ToolName: "test", Message: "x is unused",
			Severity: SeverityInfo, Position: Pos("b.go", 3, 1),
		},
	}

	var buf strings.Builder

	err := FormatTable(&buf, findings)
	if err != nil {
		t.Fatalf("FormatTable() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SEVERITY") {
		t.Errorf("missing header in output: %q", output)
	}

	if !strings.Contains(output, "🟠 ERROR") {
		t.Errorf("missing error badge in output: %q", output)
	}

	if !strings.Contains(output, "🟢 INFO") {
		t.Errorf("missing info badge in output: %q", output)
	}

	if !strings.Contains(output, "main.go:42:5") {
		t.Errorf("missing position in output: %q", output)
	}

	if !strings.Contains(output, "[security]") {
		t.Errorf("missing category in output: %q", output)
	}
}

func TestFormatTable_Empty(t *testing.T) {
	t.Parallel()

	var buf strings.Builder

	err := FormatTable(&buf, nil)
	if err != nil {
		t.Fatalf("FormatTable() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SEVERITY") {
		t.Errorf("missing header in empty table: %q", output)
	}

	lines := strings.Count(output, "\n")
	if lines != 1 {
		t.Errorf("expected 1 line (header only) in empty table, got %d", lines)
	}
}
