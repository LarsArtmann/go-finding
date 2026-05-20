package finding

import (
	"fmt"
	"io"
	"strings"
)

// FormatText writes a human-readable text representation of findings to w.
// Each finding is formatted as: file:line:col [SEVERITY] rule: message.
func FormatText(w io.Writer, findings []Finding) {
	for _, f := range findings {
		_, _ = fmt.Fprintf(
			w, "%s [%s] %s: %s\n",
			f.Position.String(),
			strings.ToUpper(string(f.Severity)),
			f.Rule,
			f.Message,
		)

		if f.Suggestion != "" {
			_, _ = fmt.Fprintf(w, "  Suggestion: %s\n", f.Suggestion)
		}
	}
}

// FormatMarkdown writes a markdown table of findings to w.
func FormatMarkdown(w io.Writer, findings []Finding) {
	_, _ = fmt.Fprintf(w, "| Location | Severity | Rule | Message |\n")
	_, _ = fmt.Fprintf(w, "|----------|----------|------|--------|\n")

	const maxMessageLen = 80

	for _, f := range findings {
		msg := f.Message
		if len(msg) > maxMessageLen {
			msg = msg[:77] + "..."
		}

		msg = strings.ReplaceAll(msg, "|", "\\|")
		msg = strings.ReplaceAll(msg, "\n", " ")

		_, _ = fmt.Fprintf(
			w, "| %s | %s | %s | %s |\n",
			f.Position.String(),
			string(f.Severity),
			f.Rule,
			msg,
		)
	}
}
