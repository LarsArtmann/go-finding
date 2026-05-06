package finding

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"testing"

	. "github.com/onsi/gomega"
)

func assertSingleFindingWithID(t *testing.T, got *Report, wantID string) {
	t.Helper()
	if len(got.Findings) != 1 {
		t.Fatalf("Findings length = %d, want 1", len(got.Findings))
	}
	if got.Findings[0].ID != wantID {
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

func TestFindingsFromJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid findings slice", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		orig := []Finding{
			MakeFindingWithPos("f1", "R1", "t", "msg", SeverityInfo, "a.go", 1, 1),
			MakeFindingWithPos("f2", "R2", "t", "msg", SeverityError, "b.go", 2, 1),
		}

		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		got, _, err := FindingsFromJSON(data)
		if err != nil {
			t.Fatalf("FindingsFromJSON: %v", err)
		}

		g.Expect(got).To(HaveLen(2))

		if got[0].ID != "f1" || got[1].ID != "f2" {
			t.Errorf("IDs = [%q, %q], want [f1, f2]", got[0].ID, got[1].ID)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		t.Parallel()

		expectJSONError(t, func() error {
			_, _, err := FindingsFromJSON([]byte("[]]"))
			return err
		}, "invalid JSON")
	})

	t.Run("filters invalid findings", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		orig := []Finding{
			MakeFindingWithPos("f1", "R1", "t", "msg", SeverityInfo, "a.go", 1, 1),
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

		g.Expect(got).To(HaveLen(1))

		if got[0].ID != "f1" {
			t.Errorf("ID = %q, want %q", got[0].ID, "f1")
		}
	})
}

func TestPrettyJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := MakeSimpleReport("tool")

	got, err := r.PrettyJSON()
	if err != nil {
		t.Fatalf("PrettyJSON: %v", err)
	}

	g.Expect(got).To(ContainSubstring("\n"))
	g.Expect(got).To(ContainSubstring("  "))
	g.Expect(got).To(ContainSubstring(`"tool"`))
}

func TestLineJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:       "f1",
		Rule:     "r1",
		Severity: SeverityWarning,
	}

	got, err := f.LineJSON()
	if err != nil {
		t.Fatalf("LineJSON: %v", err)
	}

	g.Expect(got).NotTo(ContainSubstring("\n"))
	g.Expect(got).To(ContainSubstring(`"id"`))
}

func TestPrettyJSON_ErrorPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := MakeSimpleReport("tool")
	r.AddFinding(Finding{
		ID: "f1", Rule: "r1", ToolName: "t", Message: "m",
		Severity: SeverityWarning, Position: Position{File: "a.go"},
		Confidence: Confidence(math.NaN()),
	})

	_, err := r.PrettyJSON()
	g.Expect(err).To(HaveOccurred())
}

func TestLineJSON_ErrorPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:         "f1",
		Rule:       "r1",
		ToolName:   "t",
		Message:    "m",
		Severity:   SeverityWarning,
		Position:   Position{File: "a.go"},
		Confidence: Confidence(math.NaN()),
	}

	_, err := f.LineJSON()
	g.Expect(err).To(HaveOccurred())
}

func expectJSONError(t *testing.T, fn func() error, context string) {
	t.Helper()

	if err := fn(); err == nil {
		t.Errorf("expected error for %s", context)
	}
}

func TestFinding_WriteJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{
		ID:       "f1",
		Rule:     "r1",
		Severity: SeverityWarning,
	}

	var buf bytes.Buffer
	err := f.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	got := buf.String()
	g.Expect(got).To(ContainSubstring(`"id":"f1"`))
	g.Expect(got).To(ContainSubstring("\n"))
}

func TestReport_WriteJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := MakeSimpleReport("tool")

	var buf bytes.Buffer
	err := r.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	got := buf.String()
	g.Expect(got).To(ContainSubstring("\n"))
	g.Expect(got).To(ContainSubstring("  "))
	g.Expect(got).To(ContainSubstring(`"tool"`))
}

type failingWriter struct{ err error }

func (w *failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestFinding_WriteJSON_Error(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := Finding{ID: "f1", Rule: "r1", Severity: SeverityWarning}
	err := f.WriteJSON(&failingWriter{err: errors.New("write failed")})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("encoding finding JSON")))
}

func TestReport_WriteJSON_Error(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := MakeSimpleReport("tool")
	err := r.WriteJSON(&failingWriter{err: errors.New("write failed")})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("encoding report JSON")))
}
