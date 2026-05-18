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
			Severity: SeverityError, Position: Pos("main.go", 42, 5), Suggestion: "add nil check",
		},
	}

	var buf strings.Builder
	FormatText(&buf, findings)

	output := buf.String()
	if !strings.Contains(output, "main.go:42:5") {
		t.Errorf("missing position in output: %q", output)
	}

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("missing severity in output: %q", output)
	}

	if !strings.Contains(output, "nilcheck") {
		t.Errorf("missing rule in output: %q", output)
	}

	if !strings.Contains(output, "Suggestion: add nil check") {
		t.Errorf("missing suggestion in output: %q", output)
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
	FormatMarkdown(&buf, findings)

	output := buf.String()
	if !strings.Contains(output, "| main.go:42:5 | error | nilcheck | possible nil deref |") {
		t.Errorf("unexpected markdown output: %q", output)
	}
}
