package finding

import (
	"errors"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

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

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("missing severity tag in output: %q", output)
	}

	if !strings.Contains(output, "nilcheck") {
		t.Errorf("missing rule in output: %q", output)
	}

	if !strings.Contains(output, "Suggestion: add nil check") {
		t.Errorf("missing suggestion text in output: %q", output)
	}
}

func TestFormatText_WriterError(t *testing.T) {
	t.Parallel()

	findings := []Finding{{Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1)}}

	err := FormatText(errWriter{}, findings)
	if err == nil {
		t.Fatal("FormatText must propagate writer errors")
	}

	if !strings.Contains(err.Error(), "format text") {
		t.Errorf("error should identify formatter: %v", err)
	}
}

func TestFormatTextRich_CategoryAndWriterError(t *testing.T) {
	t.Parallel()

	findings := []Finding{{
		Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1),
		Category: CategorySecurity, Suggestion: "check nil",
	}}

	var buf strings.Builder
	if err := FormatTextRich(&buf, findings); err != nil {
		t.Fatalf("FormatTextRich: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"[security]", "🟠 ERROR", "💡 check nil"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output: %q", want, out)
		}
	}

	if err := FormatTextRich(errWriter{}, findings); err == nil {
		t.Error("FormatTextRich must propagate writer errors")
	}
}

func TestFormatMarkdown_EscapeAndTruncate(t *testing.T) {
	t.Parallel()

	findings := []Finding{{
		Message: "a|b\nc\rd", Severity: SeverityWarning, Position: Pos("a.go", 1, 1),
	}}

	var buf strings.Builder
	if err := FormatMarkdown(&buf, findings); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}

	if out := buf.String(); !strings.Contains(out, "a\\|b cd") {
		t.Errorf("pipes/newlines/carriage returns not escaped correctly: %q", out)
	}

	truncated := escapeMarkdownCell(strings.Repeat("x", 100), 10)
	if got := utf8.RuneCountInString(truncated); got != 10 {
		t.Errorf("truncated rune count = %d, want 10", got)
	}

	if !strings.HasSuffix(truncated, "...") {
		t.Errorf("truncated cell must end with ellipsis: %q", truncated)
	}

	multibyte := escapeMarkdownCell(strings.Repeat("ä", 100), 10)
	if got := utf8.RuneCountInString(multibyte); got != 10 {
		t.Errorf("multibyte truncation must respect rune boundaries: %q (%d runes)", multibyte, got)
	}
}

func TestFormatTable_WriterError(t *testing.T) {
	t.Parallel()

	findings := []Finding{{Message: "m", Severity: SeverityInfo, Position: Pos("a.go", 1, 1)}}

	err := FormatTable(errWriter{}, findings)
	if err == nil {
		t.Fatal("FormatTable must propagate writer errors")
	}

	if !strings.Contains(err.Error(), "format table") {
		t.Errorf("error should identify formatter: %v", err)
	}
}

func TestFormatTextRich(t *testing.T) {
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
	FormatTextRich(&buf, findings) //nolint:errcheck

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

func TestFormatText_NoSuggestion(t *testing.T) {
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
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("missing severity tag in output: %q", output)
	}

	if strings.Contains(output, "Suggestion:") {
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

// failAfterWriter succeeds for the first failAt writes, then errors. It lets
// tests pin the error-return branch of each individual Fprintf/Fprintln call
// inside the formatters (a plain always-failing writer only ever covers the
// first branch).
type failAfterWriter struct {
	writes int
	failAt int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes > w.failAt {
		return 0, errors.New("write failed")
	}

	return len(p), nil
}

// TestFormatters_PartialWriteErrors drives every write call in every
// formatter to failure in turn: for each formatter the nth write fails and
// the error must propagate with the formatter's label.
func TestFormatters_PartialWriteErrors(t *testing.T) {
	t.Parallel()

	rich := []Finding{{
		Message:    "m",
		Severity:   SeverityWarning,
		Rule:       "r1",
		Category:   CategoryStyle,
		Suggestion: "do this instead",
		Position:   Pos("a.go", 1, 1),
	}}
	plain := []Finding{{Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1)}}

	cases := []struct {
		name     string
		failAt   int
		findings []Finding
		format   func(io.Writer, []Finding) error
	}{
		{"FormatText suggestion write", 2, plain, FormatText},
		{"FormatTextRich category write", 2, rich, FormatTextRich},
		{"FormatTextRich newline write", 3, rich, FormatTextRich},
		{"FormatTextRich suggestion write", 4, rich, FormatTextRich},
		{"FormatMarkdown separator write", 2, plain, FormatMarkdown},
		{"FormatMarkdown row write", 3, plain, FormatMarkdown},
		{"FormatTable row write", 2, plain, FormatTable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := &failAfterWriter{failAt: tc.failAt}

			err := tc.format(w, tc.findings)
			if err == nil {
				t.Fatalf("write %d failing must propagate an error", tc.failAt)
			}
		})
	}
}
