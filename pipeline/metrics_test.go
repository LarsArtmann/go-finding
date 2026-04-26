package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	assert.NotNil(t, m)

	snap := m.Snapshot()
	assert.NotNil(t, snap.StageDurations)
	assert.NotNil(t, snap.DetectorTimes)
	assert.NotNil(t, snap.FindingsFound)
}

func TestMetrics_RecordStage(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordStage("detect", 50*time.Millisecond)
	m.RecordStage("apply", 200*time.Millisecond)

	assert.Equal(t, 150*time.Millisecond, m.StageDuration("detect"))
	assert.Equal(t, 200*time.Millisecond, m.StageDuration("apply"))
}

func TestMetrics_RecordDetector(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.RecordDetector("staticcheck", 50*time.Millisecond, 10)
	m.RecordDetector("staticcheck", 30*time.Millisecond, 5)
	m.RecordDetector("govet", 20*time.Millisecond, 3)

	assert.Equal(t, 80*time.Millisecond, m.DetectorTime("staticcheck"))
	assert.Equal(t, 15, m.DetectorFindings("staticcheck"))
	assert.Equal(t, 3, m.DetectorFindings("govet"))
}

func TestMetrics_RecordFix(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.RecordFix()
	m.RecordFix()
	m.RecordFix()

	assert.Equal(t, 3, m.TotalFixesApplied())
}

func TestMetrics_StageTiming(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	done := m.StageTiming("detect")

	time.Sleep(10 * time.Millisecond)
	done()

	assert.GreaterOrEqual(t, m.StageDuration("detect"), 10*time.Millisecond)
}

func TestMetrics_TotalDuration(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.SetStart(time.Now())

	time.Sleep(10 * time.Millisecond)

	m.SetEnd(time.Now())

	d := m.TotalDuration()
	assert.GreaterOrEqual(t, d, 10*time.Millisecond)
}

func TestMetrics_Snapshot(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordDetector("govet", 50*time.Millisecond, 5)
	m.RecordFix()
	m.SetStart(time.Now().Add(-1 * time.Second))
	m.SetEnd(time.Now())

	snap := m.Snapshot()

	assert.Equal(t, 100*time.Millisecond, snap.StageDurations["detect"])
	assert.Equal(t, 50*time.Millisecond, snap.DetectorTimes["govet"])
	assert.Equal(t, 5, snap.FindingsFound["govet"])
	assert.Equal(t, 1, snap.FixesApplied)
	assert.GreaterOrEqual(t, snap.TotalDuration, 900*time.Millisecond)

	// Verify snapshot is a copy
	snap2 := m.Snapshot()
	snap2.StageDurations["detect"] = 0

	assert.Equal(t, 100*time.Millisecond, m.StageDuration("detect"))
}

func TestPipeline_MetricsIntegration(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	count := 0
	detector := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
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

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.True(t, result.Stable)

	snap := m.Snapshot()
	assert.False(t, snap.StartTime.IsZero())
	assert.False(t, snap.EndTime.IsZero())

	_, ok := snap.StageDurations["detect"]
	assert.True(t, ok)
}

func TestMetrics_TotalDuration_NoEnd(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.SetStart(time.Now())

	d := m.TotalDuration()
	assert.Equal(t, time.Duration(0), d)
}

func TestMetricsSnapshot_StageDuration(t *testing.T) {
	t.Parallel()
	snap := MetricsSnapshot{
		StageDurations: map[string]time.Duration{
			"detect": 100 * time.Millisecond,
		},
	}

	assert.Equal(t, 100*time.Millisecond, snap.StageDuration("detect"))
	assert.Equal(t, time.Duration(0), snap.StageDuration("nonexistent"))
}

func TestPipeline_NilMetricsNoPanic(t *testing.T) {
	t.Parallel()
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

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.True(t, result.Stable)
}
