package finding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzToSARIF(f *testing.F) {
	f.Add("tool1", "rule1", "msg1", "file1.go", 1, 1, "error")
	f.Add("", "", "", "", 0, 0, "warning")

	f.Fuzz(func(t *testing.T, tool, rule, msg, file string, line, col int, level string) {
		if line < 0 || col < 0 {
			t.Skip()
		}

		r := NewReport(ToolInfo{Name: tool})
		r.AddFinding(Finding{
			ID:       "id",
			Rule:     rule,
			ToolName: tool,
			Message:  msg,
			Severity: FromSARIFLevel(level),
			Position: Position{File: file, Line: line, Column: col},
		})
		r.ComputeSummary()

		data, err := r.ToSARIF()
		require.NoError(t, err, "ToSARIF should not fail")

		var log map[string]any
		require.NoError(t, json.Unmarshal(data, &log), "SARIF output should be valid JSON")

		// Round-trip should not panic.
		_, _ = FindingsFromSARIF(data)
	})
}
