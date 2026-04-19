package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

func assertEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %v, want %v", msg, got, want)
	}
}

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}

	snap := m.Snapshot()
	if snap.StageDurations == nil {
		t.Error("StageDurations should be initialized")
	}

	if snap.DetectorTimes == nil {
		t.Error("DetectorTimes should be initialized")
	}

	if snap.FindingsFound == nil {
		t.Error("FindingsFound should be initialized")
	}
}

func TestMetrics_RecordStage(t *testing.T) {
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordStage("detect", 50*time.Millisecond)
	m.RecordStage("apply", 200*time.Millisecond)

	assertEqual(t, m.StageDuration("detect"), 150*time.Millisecond, "expected detect=150ms")
	assertEqual(t, m.StageDuration("apply"), 200*time.Millisecond, "expected apply=200ms")
}

func TestMetrics_RecordDetector(t *testing.T) {
	m := NewMetrics()
	m.RecordDetector("staticcheck", 50*time.Millisecond, 10)
	m.RecordDetector("staticcheck", 30*time.Millisecond, 5)
	m.RecordDetector("govet", 20*time.Millisecond, 3)

	assertEqual(t, m.DetectorTime("staticcheck"), 80*time.Millisecond, "expected staticcheck=80ms")
	assertEqual(t, m.DetectorFindings("staticcheck"), 15, "expected staticcheck findings=15")
	assertEqual(t, m.DetectorFindings("govet"), 3, "expected govet findings=3")
}

func TestMetrics_RecordFix(t *testing.T) {
	m := NewMetrics()
	m.RecordFix()
	m.RecordFix()
	m.RecordFix()

	if m.TotalFixesApplied() != 3 {
		t.Errorf("expected 3 fixes, got %d", m.TotalFixesApplied())
	}
}

func TestMetrics_StageTiming(t *testing.T) {
	m := NewMetrics()
	done := m.StageTiming("detect")

	time.Sleep(10 * time.Millisecond)
	done()

	if m.StageDuration("detect") < 10*time.Millisecond {
		t.Errorf("expected >= 10ms, got %v", m.StageDuration("detect"))
	}
}

func TestMetrics_TotalDuration(t *testing.T) {
	m := NewMetrics()
	m.SetStart(time.Now())

	time.Sleep(10 * time.Millisecond)

	m.SetEnd(time.Now())

	d := m.TotalDuration()
	if d < 10*time.Millisecond {
		t.Errorf("expected >= 10ms, got %v", d)
	}
}

func TestMetrics_Snapshot(t *testing.T) {
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordDetector("govet", 50*time.Millisecond, 5)
	m.RecordFix()
	m.SetStart(time.Now().Add(-1 * time.Second))
	m.SetEnd(time.Now())

	snap := m.Snapshot()

	assertEqual(
		t,
		snap.StageDurations["detect"],
		100*time.Millisecond,
		"snapshot: expected detect=100ms",
	)
	assertEqual(
		t,
		snap.DetectorTimes["govet"],
		50*time.Millisecond,
		"snapshot: expected govet=50ms",
	)
	assertEqual(t, snap.FindingsFound["govet"], 5, "snapshot: expected govet findings=5")
	assertEqual(t, snap.FixesApplied, 1, "snapshot: expected 1 fix")

	if snap.TotalDuration < 900*time.Millisecond {
		t.Errorf("snapshot: expected >= 900ms, got %v", snap.TotalDuration)
	}

	// Verify snapshot is a copy
	snap2 := m.Snapshot()
	snap2.StageDurations["detect"] = 0
	if m.StageDuration("detect") != 100*time.Millisecond {
		t.Error("snapshot should be a copy, not a reference")
	}
}

func TestPipeline_MetricsIntegration(t *testing.T) {
	m := NewMetrics()
	count := 0
	detector := DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
		count++
		if count > 1 {
			return nil, nil
		}

		return []finding.Finding{
			{ID: "F1", Severity: finding.SeverityWarning, FixStrategy: finding.FixStrategyNone},
		}, nil
	})

	config := Config{
		MaxIterations:     3,
		ParallelDetectors: false,
		Timeout:           5 * time.Second,
		Metrics:           m,
	}

	p := New(config, t.TempDir(), detector)

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Stable {
		t.Error("expected stable result")
	}

	snap := m.Snapshot()
	if snap.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if snap.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	if _, ok := snap.StageDurations["detect"]; !ok {
		t.Error("expected detect stage timing")
	}
}

func TestPipeline_NilMetricsNoPanic(t *testing.T) {
	detector := &mockDetector{
		name:     "test",
		findings: []finding.Finding{},
	}

	config := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		Timeout:           5 * time.Second,
		Metrics:           nil,
	}

	p := New(config, t.TempDir(), detector)

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Stable {
		t.Error("expected stable with no findings")
	}
}
