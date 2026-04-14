package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestPartialResult_HasErrors(t *testing.T) {
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
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "d1", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "d2", findings: []finding.Finding{{ID: "F2"}}}
	p := New(config, t.TempDir(), d1, d2)

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(result.Findings))
	}
	if result.HasErrors() {
		t.Error("expected no errors")
	}
}

func TestDetectPartial_Sequential_PartialFailure(t *testing.T) {
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "good", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "bad", err: errors.New("boom")}
	p := New(config, t.TempDir(), d1, d2)

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 finding from good detector, got %d", len(result.Findings))
	}
	if !result.HasErrors() {
		t.Error("expected errors from bad detector")
	}
	if result.Errors["bad"] == nil {
		t.Error("expected error for bad detector")
	}
}

func TestDetectPartial_Parallel_PartialFailure(t *testing.T) {
	config := Config{ParallelDetectors: true}
	d1 := &mockDetector{name: "good", findings: []finding.Finding{{ID: "F1"}}}
	d2 := &mockDetector{name: "bad", err: errors.New("boom")}
	p := New(config, t.TempDir(), d1, d2)

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(result.Findings))
	}
	if !result.HasErrors() {
		t.Error("expected errors")
	}
}

func TestDetectPartial_AllFail(t *testing.T) {
	config := Config{ParallelDetectors: false}
	d1 := &mockDetector{name: "d1", err: errors.New("fail1")}
	d2 := &mockDetector{name: "d2", err: errors.New("fail2")}
	p := New(config, t.TempDir(), d1, d2)

	result, err := p.DetectPartial(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestFormatPartialErrors(t *testing.T) {
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
