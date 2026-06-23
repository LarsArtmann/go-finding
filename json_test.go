package finding

import (
	"encoding/json"
	"math"
	"testing"
)

func nanConfidenceFinding() Finding {
	return Finding{
		ID:         "f1",
		Rule:       "r1",
		ToolName:   "t",
		Message:    "m",
		Severity:   SeverityWarning,
		Position:   Position{File: "a.go"},
		Confidence: Confidence(math.NaN()),
	}
}

func assertSingleFindingWithID(t *testing.T, got *Report, wantID string) {
	t.Helper()

	if len(got.Findings) != 1 {
		t.Fatalf("Findings length = %d, want 1", len(got.Findings))
	}

	if string(got.Findings[0].ID) != wantID {
		t.Errorf("Findings[0].ID = %q, want %q", got.Findings[0].ID, wantID)
	}
}

func TestFromJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid finding", func(t *testing.T) {
		t.Parallel()

		orig := Finding{
			ID:          "tool:rule:file.go:10:5",
			Rule:        "rule1",
			ToolName:    "tool1",
			Message:     "test message",
			Severity:    SeverityError,
			Position:    Position{File: "file.go", Line: 10, Column: 5},
			Category:    "security",
			FixStrategy: FixStrategyDirect,
			BeforeCode:  "old",
			AfterCode:   "new",
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, err := FromJSON(data)
		if err != nil {
			t.Fatalf("FromJSON: %v", err)
		}

		if got.ID != orig.ID {
			t.Errorf("ID = %q, want %q", got.ID, orig.ID)
		}

		if got.Rule != orig.Rule {
			t.Errorf("Rule = %q, want %q", got.Rule, orig.Rule)
		}

		if got.Severity != orig.Severity {
			t.Errorf("Severity = %v, want %v", got.Severity, orig.Severity)
		}

		if got.Position.File != orig.Position.File {
			t.Errorf("Position.File = %q, want %q", got.Position.File, orig.Position.File)
		}

		if got.FixStrategy != orig.FixStrategy {
			t.Errorf("FixStrategy = %v, want %v", got.FixStrategy, orig.FixStrategy)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		t.Parallel()

		_, err := FromJSON([]byte("not json"))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("missing required fields returns error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			json string
		}{
			{"empty object", `{}`},
			{
				"missing position",
				`{"id":"x","rule":"r","toolName":"t","message":"m","severity":"warning"}`,
			},
			{
				"missing id",
				`{"rule":"r","toolName":"t","message":"m","severity":"warning","position":{"file":"a.go"}}`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := FromJSON([]byte(tt.json))
				if err == nil {
					t.Error("expected validation error for incomplete finding")
				}
			})
		}
	})
}

func TestReportFromJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid report", func(t *testing.T) {
		t.Parallel()

		orig := Report{
			Tool: ToolInfo{Name: "test-tool", Version: "1.0"},
			Findings: []Finding{
				{
					ID:       "f1",
					Rule:     "R1",
					ToolName: "test-tool",
					Message:  "msg",
					Severity: SeverityWarning,
					Position: Position{File: "a.go", Line: 1, Column: 1},
				},
			},
		}

		data, err := json.Marshal(&orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, _, err := ReportFromJSON(data)
		if err != nil {
			t.Fatalf("ReportFromJSON: %v", err)
		}

		if got.Tool.Name != orig.Tool.Name {
			t.Errorf("Tool.Name = %q, want %q", got.Tool.Name, orig.Tool.Name)
		}

		assertSingleFindingWithID(t, got, "f1")
	})

	t.Run("invalid input returns error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			input string
			msg   string
		}{
			{"invalid JSON", "{bad", "error: invalid JSON"},
			{"missing tool name", `{"tool":{"name":""},"findings":[]}`, "error: empty tool"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, _, err := ReportFromJSON([]byte(tt.input))
				if err == nil {
					t.Error(tt.msg)
				}
			})
		}
	})

	t.Run("filters invalid findings in report", func(t *testing.T) {
		t.Parallel()

		orig := Report{
			Tool: ToolInfo{Name: "test-tool"},
			Findings: []Finding{
				{
					ID:       "f1",
					Rule:     "R1",
					ToolName: "test-tool",
					Message:  "msg",
					Severity: SeverityWarning,
					Position: Position{File: "a.go", Line: 1, Column: 1},
				},
				{ID: "bad", Severity: SeverityWarning},
			},
		}

		data, err := json.Marshal(&orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, dropped, err := ReportFromJSON(data)
		if err != nil {
			t.Fatalf("ReportFromJSON: %v", err)
		}

		if dropped != 1 {
			t.Errorf("dropped = %d, want 1", dropped)
		}

		assertSingleFindingWithID(t, got, "f1")
	})
}
