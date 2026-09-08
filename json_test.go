package finding

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"math"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
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

	if len(got.findings) != 1 {
		t.Fatalf("Findings length = %d, want 1", len(got.findings))
	}

	if string(got.findings[0].ID) != wantID {
		t.Errorf("Findings[0].ID = %q, want %q", got.findings[0].ID, wantID)
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
			findings: []Finding{
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
			findings: []Finding{
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
		g := NewParallelGomega(t)

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
		g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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

func TestPrettyJSONFiltered(t *testing.T) {
	g := NewParallelGomega(t)

	r := NewReport(ToolInfo{Name: "tool"})
	r.AddFinding(Finding{ID: "active", Rule: "r1", Message: "m1", Severity: SeverityWarning})
	r.AddFinding(Finding{
		ID:          "suppressed",
		Rule:        "r2",
		Message:     "m2",
		Severity:    SeverityWarning,
		Suppression: &Suppression{Kind: SuppressionInSource, Rule: "r2"},
	})

	got, err := r.PrettyJSONFiltered()
	g.Expect(err).NotTo(HaveOccurred())

	var parsed Report
	g.Expect(json.Unmarshal([]byte(got), &parsed)).To(Succeed())
	g.Expect(parsed.FindingsSnapshot()).To(HaveLen(1))
	g.Expect(parsed.FindingsSnapshot()[0].ID).To(Equal(ID("active")))
}

func TestPrettyJSON_ErrorPath(t *testing.T) {
	g := NewParallelGomega(t)

	r := MakeSimpleReport("tool")
	r.AddFinding(nanConfidenceFinding())

	_, err := r.PrettyJSON()
	g.Expect(err).To(HaveOccurred())
}

func TestLineJSON_ErrorPath(t *testing.T) {
	g := NewParallelGomega(t)

	f := nanConfidenceFinding()

	_, err := f.LineJSON()
	g.Expect(err).To(HaveOccurred())
}

func expectJSONError(t *testing.T, fn func() error, context string) {
	t.Helper()

	err := fn()
	if err == nil {
		t.Errorf("expected error for %s", context)
	}
}

func TestFinding_WriteJSON(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	f := Finding{ID: "f1", Rule: "r1", Severity: SeverityWarning}
	err := f.WriteJSON(&failingWriter{err: errors.New("write failed")})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("encoding finding JSON")))
}

func TestReport_WriteJSON_Error(t *testing.T) {
	g := NewParallelGomega(t)

	r := MakeSimpleReport("tool")
	err := r.WriteJSON(&failingWriter{err: errors.New("write failed")})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("encoding report JSON")))
}

// TestReport_UnmarshalJSON_ConcurrentSafety verifies that UnmarshalJSON is safe
// for concurrent use. Run with -race to detect data races.
// Regression for the UnmarshalJSON data race (json.go:43).
func TestReport_UnmarshalJSON_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "concurrent", Version: "1.0"})
	report.AddFinding(Finding{
		ID:         "f1",
		Rule:       "rule-a",
		ToolName:   "concurrent",
		Message:    "test",
		Severity:   SeverityWarning,
		Position:   Position{File: FilePath("main.go"), Line: 1},
		Confidence: ConfidenceHigh,
	})

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	target := NewReport(ToolInfo{})

	var wg sync.WaitGroup

	// Concurrent writers — all unmarshalling into the same *Report.
	for range 10 {
		wg.Go(func() {
			_ = target.UnmarshalJSON(data)
		})
	}

	// Concurrent readers — snapshot while writers may be active.
	for range 10 {
		wg.Go(func() {
			_ = target.FindingsSnapshot()
		})
	}

	wg.Wait()
}

// findingWithMaps builds a Finding whose JSON output includes map-bearing fields
// (Metadata, Tags) that would serialize in non-deterministic key order without
// json.Deterministic(true).
func findingWithMaps() Finding {
	return Finding{
		ID:       "tool:rule:file.go:10:5",
		Rule:     "rule1",
		ToolName: "tool1",
		Message:  "test message",
		Severity: SeverityError,
		Position: Position{File: FilePath("file.go"), Line: 10, Column: 5},
		Category: "security",
		Metadata: map[string]string{
			"zebra": "z", "alpha": "a", "mike": "m", "delta": "d", "sierra": "s",
			"bravo": "b", "tango": "t", "hotel": "h", "india": "i", "kilo": "k",
			"lima": "l", "november": "n", "oscar": "o", "papa": "p", "quebec": "q",
			"romeo": "r", "uniform": "u", "victor": "v", "whiskey": "w", "xray": "x",
		},
		Tags: []Tag{"tag-c", "tag-a", "tag-b"},
	}
}

// reportWithMaps builds a Report with multiple map-bearing findings for
// determinism testing.
func reportWithMaps() *Report {
	r := NewReport(ToolInfo{Name: "det-test", Version: "1.0"})
	r.AddFinding(findingWithMaps())
	r.AddFinding(Finding{
		ID:       "f2",
		Rule:     "r2",
		Severity: SeverityWarning,
		Position: Position{File: FilePath("b.go"), Line: 2},
		Metadata: map[string]string{
			"k0": "v0", "k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4",
			"k5": "v5", "k6": "v6", "k7": "v7", "k8": "v8", "k9": "v9",
		},
	})
	r.ComputeSummary()

	return r
}

func TestDeterminism_FindingJSON_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	f := findingWithMaps()

	first, err := f.LineJSON()
	if err != nil {
		t.Fatalf("first LineJSON: %v", err)
	}

	for range 100 {
		got, err := f.LineJSON()
		if err != nil {
			t.Fatalf("LineJSON: %v", err)
		}

		if got != first {
			t.Fatalf("non-deterministic Finding JSON output:\n  run 1: %s\n  run N: %s", first, got)
		}
	}
}

func TestDeterminism_FindingWriteJSON_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	f := findingWithMaps()

	var first bytes.Buffer
	if err := f.WriteJSON(&first); err != nil {
		t.Fatalf("first WriteJSON: %v", err)
	}

	for range 100 {
		var buf bytes.Buffer
		if err := f.WriteJSON(&buf); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		if !bytes.Equal(first.Bytes(), buf.Bytes()) {
			t.Fatalf("non-deterministic Finding.WriteJSON output")
		}
	}
}

func TestDeterminism_ReportJSON_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := reportWithMaps()

	first, err := r.JSON()
	if err != nil {
		t.Fatalf("first JSON: %v", err)
	}

	for range 100 {
		got, err := r.JSON()
		if err != nil {
			t.Fatalf("JSON: %v", err)
		}

		if got != first {
			t.Fatalf("non-deterministic Report JSON output")
		}
	}
}

func TestDeterminism_ReportPrettyJSON_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "det-test", Version: "1.0"})
	r.AddFinding(findingWithMaps())
	r.ComputeSummary()

	first, err := r.PrettyJSON()
	if err != nil {
		t.Fatalf("first PrettyJSON: %v", err)
	}

	for range 100 {
		got, err := r.PrettyJSON()
		if err != nil {
			t.Fatalf("PrettyJSON: %v", err)
		}

		if got != first {
			t.Fatalf("non-deterministic Report.PrettyJSON output")
		}
	}
}

func TestDeterminism_ReportPrettyJSONFiltered_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := reportWithMaps()

	first, err := r.PrettyJSONFiltered()
	if err != nil {
		t.Fatalf("first PrettyJSONFiltered: %v", err)
	}

	for range 100 {
		got, err := r.PrettyJSONFiltered()
		if err != nil {
			t.Fatalf("PrettyJSONFiltered: %v", err)
		}

		if got != first {
			t.Fatalf("non-deterministic Report.PrettyJSONFiltered output")
		}
	}
}

func TestDeterminism_ReportWriteJSON_RawBytesIdentical(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "det-test", Version: "1.0"})
	r.AddFinding(findingWithMaps())
	r.ComputeSummary()

	var first bytes.Buffer
	if err := r.WriteJSON(&first); err != nil {
		t.Fatalf("first WriteJSON: %v", err)
	}

	for range 100 {
		var buf bytes.Buffer
		if err := r.WriteJSON(&buf); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		if !bytes.Equal(first.Bytes(), buf.Bytes()) {
			t.Fatalf("non-deterministic Report.WriteJSON output")
		}
	}
}

// TestFindingJSON_GoldenWire pins the exact JSON wire format for Finding so
// any change to field names, ordering, or the groupId encoding is a conscious
// decision. Consumers and downstream tools parse these bytes verbatim.
func TestFindingJSON_GoldenWire(t *testing.T) {
	g := NewParallelGomega(t)

	goldenNoGroup := `{"id":"test:rule1:file.go:10:5","rule":"rule1","toolName":"test","message":"test message","severity":"error","position":{"file":"file.go","line":10,"column":5,"offset":0},"fixStrategy":"none","confidence":1}`
	goldenWithGroup := `{"id":"test:rule1:file.go:10:5","rule":"rule1","toolName":"test","message":"test message","severity":"error","position":{"file":"file.go","line":10,"column":5,"offset":0},"fixStrategy":"none","confidence":1,"groupId":"clone-group-1"}`

	t.Run("groupId absent", func(t *testing.T) {
		t.Parallel()

		data, err := json.Marshal(standardTestFinding())
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(data)).To(Equal(goldenNoGroup))
	})

	t.Run("groupId present", func(t *testing.T) {
		t.Parallel()

		f := standardTestFinding()
		f.GroupID = "clone-group-1"

		data, err := json.Marshal(f)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(data)).To(Equal(goldenWithGroup))

		parsed, err := FromJSON(data)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(parsed.Equal(f)).To(BeTrue())
	})
}
