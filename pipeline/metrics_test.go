package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

func assertMetricsDuration(t *testing.T, got, want time.Duration, msg string) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %v, want %v", msg, got, want)
	}
}

func assertMetricsInt(t *testing.T, got, want int, msg string) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %d, want %d", msg, got, want)
	}
}

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}

	if m.StageDurations == nil {
		t.Error("StageDurations should be initialized")
	}

	if m.DetectorTimes == nil {
		t.Error("DetectorTimes should be initialized")
	}

	if m.FindingsFound == nil {
		t.Error("FindingsFound should be initialized")
	}
}

func TestMetrics_RecordStage(t *testing.T) {
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordStage("detect", 50*time.Millisecond)
	m.RecordStage("apply", 200*time.Millisecond)

	assertMetricsDuration(
		t,
		m.StageDurations["detect"],
		150*time.Millisecond,
		"expected detect=150ms",
	)
	assertMetricsDuration(
		t,
		m.StageDurations["apply"],
		200*time.Millisecond,
		"expected apply=200ms",
	)
}

func TestMetrics_RecordDetector(t *testing.T) {
	m := NewMetrics()
	m.RecordDetector("staticcheck", 50*time.Millisecond, 10)
	m.RecordDetector("staticcheck", 30*time.Millisecond, 5)
	m.RecordDetector("govet", 20*time.Millisecond, 3)

	assertMetricsDuration(
		t,
		m.DetectorTimes["staticcheck"],
		80*time.Millisecond,
		"expected staticcheck=80ms",
	)
	assertMetricsInt(t, m.FindingsFound["staticcheck"], 15, "expected staticcheck findings=15")
	assertMetricsInt(t, m.FindingsFound["govet"], 3, "expected govet findings=3")
}

func TestMetrics_RecordFix(t *testing.T) {
	m := NewMetrics()
	m.RecordFix()
	m.RecordFix()
	m.RecordFix()

	if m.FixesApplied != 3 {
		t.Errorf("expected 3 fixes, got %d", m.FixesApplied)
	}
}

func TestMetrics_StageTiming(t *testing.T) {
	m := NewMetrics()
	done := m.StageTiming("detect")

	time.Sleep(10 * time.Millisecond)
	done()

	if m.StageDurations["detect"] < 10*time.Millisecond {
		t.Errorf("expected >= 10ms, got %v", m.StageDurations["detect"])
	}
}

func TestMetrics_TotalDuration(t *testing.T) {
	m := NewMetrics()
	m.StartTime = time.Now()

	time.Sleep(10 * time.Millisecond)

	m.EndTime = time.Now()

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
	m.StartTime = time.Now().Add(-1 * time.Second)
	m.EndTime = time.Now()

	snap := m.Snapshot()

	assertMetricsDuration(
		t,
		snap.StageDurations["detect"],
		100*time.Millisecond,
		"snapshot: expected detect=100ms",
	)
	assertMetricsDuration(
		t,
		snap.DetectorTimes["govet"],
		50*time.Millisecond,
		"snapshot: expected govet=50ms",
	)
	assertMetricsInt(t, snap.FindingsFound["govet"], 5, "snapshot: expected govet findings=5")
	assertMetricsInt(t, snap.FixesApplied, 1, "snapshot: expected 1 fix")

	if snap.TotalDuration < 900*time.Millisecond {
		t.Errorf("snapshot: expected >= 900ms, got %v", snap.TotalDuration)
	}

	// Verify snapshot is a copy
	snap.StageDurations["detect"] = 0
	if m.StageDurations["detect"] != 100*time.Millisecond {
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

	if m.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if m.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	if _, ok := m.StageDurations["detect"]; !ok {
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
