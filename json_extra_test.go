package finding

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

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

func TestPrettyJSONFiltered(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := NewReport(ToolInfo{Name: "tool"})
	r.AddFinding(Finding{ID: "active", Rule: "r1", Message: "m1", Severity: SeverityWarning})
	r.AddFinding(Finding{
		ID:          "suppressed",
		Rule:        "r2",
		Message:     "m2",
		Severity:    SeverityWarning,
		Suppression: &Suppression{Kind: SuppressionInSource},
	})

	got, err := r.PrettyJSONFiltered()
	g.Expect(err).NotTo(HaveOccurred())

	var parsed Report
	g.Expect(json.Unmarshal([]byte(got), &parsed)).To(Succeed())
	g.Expect(parsed.Findings).To(HaveLen(1))
	g.Expect(parsed.Findings[0].ID).To(Equal("active"))
}

func TestPrettyJSON_ErrorPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	r := MakeSimpleReport("tool")
	r.AddFinding(nanConfidenceFinding())

	_, err := r.PrettyJSON()
	g.Expect(err).To(HaveOccurred())
}

func TestLineJSON_ErrorPath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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
