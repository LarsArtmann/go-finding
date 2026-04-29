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

func FuzzFindingsFromSARIF(f *testing.F) {
	// Valid minimal SARIF
	f.Add([]byte(`{"version":"2.1.0","runs":[]}`))
	// Valid SARIF with a result
	f.Add([]byte(
		`{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"t"}},"results":[{"ruleId":"r1","level":"warning","message":{"text":"m"}}]}]}`,
	))
	// Invalid JSON
	f.Add([]byte(`not json`))
	// Empty input
	f.Add([]byte{})
	// Truncated JSON
	f.Add([]byte(`{"version":"2.1.0","runs":`))
	// Missing required fields
	f.Add([]byte(`{"version":"2.1.0"}`))
	// Nested objects with missing fields
	f.Add([]byte(`{"version":"2.1.0","runs":[{"results":[{"ruleId":"r1"}]}]}`))
	// Large random bytes
	f.Add([]byte(`\x00\x01\x02\x03`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		// Must not panic on any input.
		_, _ = FindingsFromSARIF(data)
	})
}
