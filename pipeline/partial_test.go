package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
)

func assertPartialFindingsLen(t *testing.T, result *PartialResult, want int, msg string) {
	t.Helper()

	if len(result.Findings) != want {
		t.Errorf("%s = %d, want %d", msg, len(result.Findings), want)
	}
}

func TestPartialResult_HasErrors(t *testing.T) {
	t.Parallel()
	r := &PartialResult{Errors: make(map[string]error)}
	if r.HasErrors() {
		t.Error("empty errors should not report HasErrors")
	}

	r.Errors["det1"] = errors.New("fail")
	if !r.HasErrors() {
		t.Error("should report HasErrors with one error")
	}
}

func TestDetectPartial_Sequential_AllSucceed(t *testing.T) {
	t.Parallel()
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "d1", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "d2", findings: []finding.Finding{{ID: "F2"}}}
	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertPartialFindingsLen(t, result, 2, "expected 2 findings")

	if result.HasErrors() {
		t.Error("expected no errors")
	}
}

func TestDetectPartial_Sequential_PartialFailure(t *testing.T) {
	t.Parallel()
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "good", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "bad", err: errors.New("boom")}
	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertPartialFindingsLen(t, result, 1, "expected 1 finding from good detector")

	if !result.HasErrors() {
		t.Error("expected errors from bad detector")
	}

	if result.Errors["bad"] == nil {
		t.Error("expected error for bad detector")
	}
}

func TestDetectPartial_Parallel_PartialFailure(t *testing.T) {
	t.Parallel()
	config := Config{ParallelDetectors: true}
	d1 := &mockDetector{name: "good", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "bad", err: errors.New("boom")}
	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertPartialFindingsLen(t, result, 1, "expected 1 finding")

	if !result.HasErrors() {
		t.Error("expected errors")
	}
}

func TestDetectPartial_AllFail(t *testing.T) {
	t.Parallel()
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "d1", err: errors.New("fail1")}
	d2 := &mockDetector{name: "d2", err: errors.New("fail2")}
	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertPartialFindingsLen(t, result, 0, "expected 0 findings")

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestFormatPartialErrors(t *testing.T) {
	t.Parallel()
	if err := FormatPartialErrors(nil); err != nil {
		t.Errorf("nil map should return nil error, got %v", err)
	}

	if err := FormatPartialErrors(map[string]error{}); err != nil {
		t.Errorf("empty map should return nil error, got %v", err)
	}

	errs := map[string]error{
		"det1": errors.New("fail1"),
		"det2": errors.New("fail2"),
	}

	err := FormatPartialErrors(errs)
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
