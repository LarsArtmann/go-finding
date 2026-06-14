package finding

import (
	"fmt"
	"testing"
)

func TestReportAll_BreakEarly(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	for i := range 10 {
		r.AddFinding(Finding{ID: string(rune('A' + i)), Message: "finding"})
	}

	count := 0
	for range r.All() {
		count++
		if count == 3 {
			break
		}
	}

	if count != 3 {
		t.Errorf("iterations before break = %d, want 3", count)
	}
}

func TestReport_Filter(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Severity: SeverityError, Message: "m"})
	r.AddFinding(Finding{ID: "2", Severity: SeverityInfo, Message: "m"})
	r.AddFinding(Finding{ID: "3", Severity: SeverityError, Message: "m"})

	filtered := r.Filter(BySeverity(SeverityError))

	if filtered.Tool.Name != string(TagTest) {
		t.Errorf("Tool.Name = %q, want %q", filtered.Tool.Name, string(TagTest))
	}

	AssertFindingsLenAndIDs(t, filtered.Findings, []string{"1", "3"}, "filtered.Findings")
}

func TestReport_Filter_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	filtered := r.Filter(BySeverity(SeverityError))

	if len(filtered.Findings) != 0 {
		t.Errorf("Findings length = %d, want 0", len(filtered.Findings))
	}
}

func TestReport_Map(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	r.AddFinding(Finding{ID: "1", Severity: SeverityError, Message: "m"})
	r.AddFinding(Finding{ID: "2", Severity: SeverityInfo, Message: "m"})

	mapped := r.Map(func(f Finding) Finding {
		f.Severity = SeverityWarning

		return f
	})

	if mapped.Tool.Name != string(TagTest) {
		t.Errorf("Tool.Name = %q, want %q", mapped.Tool.Name, string(TagTest))
	}

	if len(mapped.Findings) != 2 {
		t.Fatalf("Findings length = %d, want 2", len(mapped.Findings))
	}

	for i, f := range mapped.Findings {
		if f.Severity != SeverityWarning {
			t.Errorf("Findings[%d].Severity = %v, want %v", i, f.Severity, SeverityWarning)
		}
	}
}

func TestReport_Map_Empty(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: string(TagTest)})
	mapped := r.Map(func(f Finding) Finding { return f })

	if len(mapped.Findings) != 0 {
		t.Errorf("Findings length = %d, want 0", len(mapped.Findings))
	}
}

func TestReport_CountBySeverity(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "test"})
	r.AddFinding(Finding{Severity: SeverityError})
	r.AddFinding(Finding{Severity: SeverityError})
	r.AddFinding(Finding{Severity: SeverityWarning})
	r.ComputeSummary()

	if r.CountBySeverity(SeverityError) != 2 {
		t.Errorf("CountBySeverity(Error) = %d, want 2", r.CountBySeverity(SeverityError))
	}

	if r.CountBySeverity(SeverityWarning) != 1 {
		t.Errorf("CountBySeverity(Warning) = %d, want 1", r.CountBySeverity(SeverityWarning))
	}

	if r.CountBySeverity(SeverityInfo) != 0 {
		t.Errorf("CountBySeverity(Info) = %d, want 0", r.CountBySeverity(SeverityInfo))
	}
}

func TestReportConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	r := NewReport(ToolInfo{Name: "race-test"})

	const (
		writers      = 10
		readers      = 10
		opsPerWriter = 50
	)

	start := make(chan struct{})
	done := make(chan struct{}, writers+readers)

	for i := range writers {
		go func() {
			<-start

			for j := range opsPerWriter {
				r.AddFinding(Finding{
					ID:       fmt.Sprintf("w%d-f%d", i, j),
					Rule:     "race",
					Severity: SeverityWarning,
				})
			}

			done <- struct{}{}
		}()
	}

	for range readers {
		go func() {
			<-start

			for range opsPerWriter {
				_ = r.Len()
				_ = r.ActiveFindings()
				_ = r.CountBySeverity(SeverityWarning)
				r.ComputeSummary()
			}

			done <- struct{}{}
		}()
	}

	close(start)

	for range writers + readers {
		<-done
	}

	expected := writers * opsPerWriter
	if r.Len() != expected {
		t.Errorf("Len() = %d, want %d", r.Len(), expected)
	}
}

func TestReport_MergeInto(t *testing.T) {
	t.Parallel()

	r1 := NewReport(ToolInfo{Name: "tool-a"})
	r1.AddFinding(Finding{ID: "a1", Rule: "r1", Severity: SeverityError})
	r1.ComputeSummary()

	r2 := NewReport(ToolInfo{Name: "tool-b"})
	r2.AddFinding(Finding{ID: "b1", Rule: "r2", Severity: SeverityWarning})
	r2.ComputeSummary()

	merged := r1.MergeInto(r2)

	if merged.Tool.Name != "tool-a" {
		t.Errorf("Tool.Name = %q, want %q", merged.Tool.Name, "tool-a")
	}

	if merged.Len() != 2 {
		t.Errorf("Len() = %d, want 2", merged.Len())
	}

	if merged.Summary.Total != 2 {
		t.Errorf("Summary.Total = %d, want 2", merged.Summary.Total)
	}

	if r1.Len() != 1 {
		t.Errorf("r1 modified: Len() = %d, want 1", r1.Len())
	}

	if r2.Len() != 1 {
		t.Errorf("r2 modified: Len() = %d, want 1", r2.Len())
	}
}
