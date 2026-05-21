package finding

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// FormatText writes a human-readable text representation of findings to w.
// Each finding is formatted as: file:line:col [SEVERITY] rule: message.
func FormatText(w io.Writer, findings []Finding) error {
	for _, f := range findings {
		if _, err := fmt.Fprintf(
			w, "%s [%s] %s: %s\n",
			f.Position.String(),
			strings.ToUpper(string(f.Severity)),
			f.Rule,
			f.Message,
		); err != nil {
			return fmt.Errorf("format text: %w", err)
		}

		if f.Suggestion != "" {
			if _, err := fmt.Fprintf(w, "  Suggestion: %s\n", f.Suggestion); err != nil {
				return fmt.Errorf("format text suggestion: %w", err)
			}
		}
	}

	return nil
}

// FormatMarkdown writes a markdown table of findings to w.
func FormatMarkdown(w io.Writer, findings []Finding) error {
	if _, err := fmt.Fprintf(w, "| Location | Severity | Rule | Message |\n"); err != nil {
		return fmt.Errorf("format markdown header: %w", err)
	}

	if _, err := fmt.Fprintf(w, "|----------|----------|------|--------|\n"); err != nil {
		return fmt.Errorf("format markdown separator: %w", err)
	}

	const maxMessageLen = 80

	for _, f := range findings {
		msg := escapeMarkdownCell(f.Message, maxMessageLen)
		rule := escapeMarkdownCell(f.Rule, 0)

		if _, err := fmt.Fprintf(
			w, "| %s | %s | %s | %s |\n",
			escapeMarkdownCell(f.Position.String(), 0),
			string(f.Severity),
			rule,
			msg,
		); err != nil {
			return fmt.Errorf("format markdown row: %w", err)
		}
	}

	return nil
}

// escapeMarkdownCell escapes pipe and newline characters in a markdown table cell.
// If maxLen > 0, truncates the string at a rune boundary.
func escapeMarkdownCell(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")

	if maxLen > 0 && utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		s = string(runes[:maxLen-3]) + "..."
	}

	return s
}
