package finding

import (
	"encoding/json"
	"strings"
	"testing"
)

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

		data, err := json.Marshal(orig)
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

		if len(got.Findings) != 1 {
			t.Fatalf("Findings length = %d, want 1", len(got.Findings))
		}

		if got.Findings[0].ID != "f1" {
			t.Errorf("Findings[0].ID = %q, want %q", got.Findings[0].ID, "f1")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		t.Parallel()

		_, _, err := ReportFromJSON([]byte("{bad"))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("missing tool name returns error", func(t *testing.T) {
		t.Parallel()

		_, _, err := ReportFromJSON([]byte(`{"tool":{"name":""},"findings":[]}`))
		if err == nil {
			t.Error("expected validation error for missing tool name")
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

		data, err := json.Marshal(orig)
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

		if len(got.Findings) != 1 {
			t.Fatalf("Findings length = %d, want 1 (invalid filtered)", len(got.Findings))
		}

		if got.Findings[0].ID != "f1" {
			t.Errorf("Findings[0].ID = %q, want %q", got.Findings[0].ID, "f1")
		}
	})
}

func TestFindingsFromJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid findings slice", func(t *testing.T) {
		t.Parallel()

		orig := []Finding{
			{
				ID:       "f1",
				Rule:     "R1",
				ToolName: "t",
				Message:  "msg",
				Severity: SeverityInfo,
				Position: Position{File: "a.go", Line: 1, Column: 1},
			},
			{
				ID:       "f2",
				Rule:     "R2",
				ToolName: "t",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: "b.go", Line: 2, Column: 1},
			},
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, _, err := FindingsFromJSON(data)
		if err != nil {
			t.Fatalf("FindingsFromJSON: %v", err)
		}

		if len(got) != 2 {
			t.Fatalf("length = %d, want 2", len(got))
		}

		if got[0].ID != "f1" || got[1].ID != "f2" {
			t.Errorf("IDs = [%q, %q], want [f1, f2]", got[0].ID, got[1].ID)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		t.Parallel()

		_, _, err := FindingsFromJSON([]byte("[]]"))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("filters invalid findings", func(t *testing.T) {
		t.Parallel()

		orig := []Finding{
			{
				ID:       "f1",
				Rule:     "R1",
				ToolName: "t",
				Message:  "msg",
				Severity: SeverityInfo,
				Position: Position{File: "a.go", Line: 1, Column: 1},
			},
			{ID: "bad", Severity: SeverityWarning},
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, dropped, err := FindingsFromJSON(data)
		if err != nil {
			t.Fatalf("FindingsFromJSON: %v", err)
		}

		if dropped != 1 {
			t.Errorf("dropped = %d, want 1", dropped)
		}

		if len(got) != 1 {
			t.Fatalf("length = %d, want 1 (invalid finding filtered)", len(got))
		}

		if got[0].ID != "f1" {
			t.Errorf("ID = %q, want %q", got[0].ID, "f1")
		}
	})
}

func TestPrettyJSON(t *testing.T) {
	t.Parallel()

	r := &Report{
		Tool:     ToolInfo{Name: "tool"},
		Findings: []Finding{},
	}

	got, err := r.PrettyJSON()
	if err != nil {
		t.Fatalf("PrettyJSON: %v", err)
	}

	if !strings.Contains(got, "\n") {
		t.Error("PrettyJSON should contain newlines")
	}

	if !strings.Contains(got, "  ") {
		t.Error("PrettyJSON should be indented")
	}

	if !strings.Contains(got, `"tool"`) {
		t.Error("PrettyJSON should contain tool name")
	}
}

func TestLineJSON(t *testing.T) {
	t.Parallel()

	f := Finding{
		ID:       "f1",
		Rule:     "r1",
		Severity: SeverityWarning,
	}

	got, err := f.LineJSON()
	if err != nil {
		t.Fatalf("LineJSON: %v", err)
	}

	if strings.Contains(got, "\n") {
		t.Error("LineJSON should be single line (no newlines)")
	}

	if !strings.Contains(got, `"id"`) {
		t.Error("LineJSON should contain JSON fields")
	}
}
