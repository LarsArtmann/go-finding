package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func testConfig(maxIter int, m *Metrics) Config {
	return Config{
		MaxIterations:     maxIter,
		ParallelDetectors: false,
		Timeout:           5 * time.Second,
		Metrics:           m,
	}
}

func runPipelineWithMetrics(t *testing.T, config Config, det Detector) *PipelineResult {
	t.Helper()
	g := NewWithT(t)

	p, err := New(config, t.TempDir(), det)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.Run(context.Background())
	g.Expect(err).NotTo(HaveOccurred())

	return result
}

func TestNewMetrics(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	g.Expect(m).NotTo(BeNil())

	snap := m.Snapshot()
	g.Expect(snap.StageDurations).NotTo(BeNil())
	g.Expect(snap.DetectorTimes).NotTo(BeNil())
	g.Expect(snap.FindingsFound).NotTo(BeNil())
}

func TestMetrics_RecordStage(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	m.RecordStage(StageDetect, 100*time.Millisecond)
	m.RecordStage(StageDetect, 50*time.Millisecond)
	m.RecordStage(StageApply, 200*time.Millisecond)

	g.Expect(m.StageDuration(StageDetect)).To(Equal(150 * time.Millisecond))
	g.Expect(m.StageDuration(StageApply)).To(Equal(200 * time.Millisecond))
}

func TestMetrics_RecordDetector(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	m.RecordDetector("staticcheck", 50*time.Millisecond, 10)
	m.RecordDetector("staticcheck", 30*time.Millisecond, 5)
	m.RecordDetector("govet", 20*time.Millisecond, 3)

	g.Expect(m.DetectorTime("staticcheck")).To(Equal(80 * time.Millisecond))
	g.Expect(m.DetectorFindings("staticcheck")).To(Equal(15))
	g.Expect(m.DetectorFindings("govet")).To(Equal(3))
}

func TestMetrics_RecordFixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		fixes []uint
		want  int
	}{
		{"zero", []uint{0}, 0},
		{"single batch of 5", []uint{5}, 5},
		{"deprecated single mixed with batch", []uint{1, 3, 1}, 5},
		{"three single recordings", []uint{1, 1, 1}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)
			m := NewMetrics()

			for _, count := range tt.fixes {
				if count == 1 {
					m.RecordFixes(1)
				} else {
					m.RecordFixes(count)
				}
			}

			g.Expect(m.TotalFixesApplied()).To(Equal(tt.want))
		})
	}
}

func TestMetrics_StageTiming(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	done := m.StageTiming(StageDetect)

	time.Sleep(10 * time.Millisecond)
	done()

	g.Expect(m.StageDuration(StageDetect)).To(BeNumerically(">=", 10*time.Millisecond))
}

func TestMetrics_TotalDuration(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	m.SetStart(time.Now())

	time.Sleep(10 * time.Millisecond)

	m.SetEnd(time.Now())

	d := m.TotalDuration()
	g.Expect(d).To(BeNumerically(">=", 10*time.Millisecond))
}

func TestMetrics_Snapshot(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordDetector("govet", 50*time.Millisecond, 5)
	m.RecordFixes(1)
	m.SetStart(time.Now().Add(-1 * time.Second))
	m.SetEnd(time.Now())

	snap := m.Snapshot()

	g.Expect(snap.StageDurations["detect"]).To(Equal(100 * time.Millisecond))
	g.Expect(snap.DetectorTimes["govet"]).To(Equal(50 * time.Millisecond))
	g.Expect(snap.FindingsFound["govet"]).To(Equal(5))
	g.Expect(snap.FixesApplied).To(Equal(1))
	g.Expect(snap.TotalDuration).To(BeNumerically(">=", 900*time.Millisecond))

	// Verify snapshot is a copy
	snap2 := m.Snapshot()
	snap2.StageDurations["detect"] = 0

	g.Expect(m.StageDuration("detect")).To(Equal(100 * time.Millisecond))
}

func TestPipeline_MetricsIntegration(t *testing.T) {
	g := NewParallelGomega(t)
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

	result := runPipelineWithMetrics(t, testConfig(3, m), detector)
	g.Expect(result.Stable()).To(BeTrue())

	snap := m.Snapshot()
	g.Expect(snap.StartTime.IsZero()).To(BeFalse())
	g.Expect(snap.EndTime.IsZero()).To(BeFalse())

	_, ok := snap.StageDurations["detect"]
	g.Expect(ok).To(BeTrue())
}

func TestMetrics_TotalDuration_NoEnd(t *testing.T) {
	g := NewParallelGomega(t)
	m := NewMetrics()
	m.SetStart(time.Now())

	d := m.TotalDuration()
	g.Expect(d).To(Equal(time.Duration(0)))
}

func TestMetricsSnapshot_StageDuration(t *testing.T) {
	g := NewParallelGomega(t)

	snap := MetricsSnapshot{
		StageDurations: map[Stage]time.Duration{
			StageDetect: 100 * time.Millisecond,
		},
	}

	g.Expect(snap.StageDuration(StageDetect)).To(Equal(100 * time.Millisecond))
	g.Expect(snap.StageDuration("nonexistent")).To(Equal(time.Duration(0)))
}

func TestPipeline_NilMetricsNoPanic(t *testing.T) {
	g := NewParallelGomega(t)

	detector := &mockDetector{
		name:     "test",
		findings: []finding.Finding{},
	}

	result := runPipelineWithMetrics(t, testConfig(1, nil), detector)
	g.Expect(result.Stable()).To(BeTrue())
}

func TestMetrics_TotalDuration_NegativeGuard(t *testing.T) {
	g := NewParallelGomega(t)

	m := NewMetrics()
	m.SetStart(time.Now())
	m.SetEnd(time.Now().Add(-1 * time.Hour))

	d := m.TotalDuration()
	g.Expect(d).To(Equal(time.Duration(0)))
}

func TestMetrics_Snapshot_NegativeGuard(t *testing.T) {
	g := NewParallelGomega(t)

	m := NewMetrics()
	m.SetStart(time.Now())
	m.SetEnd(time.Now().Add(-1 * time.Hour))

	snap := m.Snapshot()
	g.Expect(snap.TotalDuration).To(Equal(time.Duration(0)))
}

func TestMetrics_RecordOutcome(t *testing.T) {
	g := NewParallelGomega(t)

	m := NewMetrics()
	m.RecordOutcome(FixOutcomeApplied)
	m.RecordOutcome(FixOutcomeApplied)
	m.RecordOutcome(FixOutcomeRefused)

	g.Expect(m.OutcomeCounts()).To(Equal(map[FixOutcomeStatus]int{
		FixOutcomeApplied: 2,
		FixOutcomeRefused: 1,
	}))

	snap := m.Snapshot()
	g.Expect(snap.OutcomeCounts).To(Equal(map[FixOutcomeStatus]int{
		FixOutcomeApplied: 2,
		FixOutcomeRefused: 1,
	}))

	// Snapshot must be a copy: mutating it must not affect the metrics.
	snap.OutcomeCounts[FixOutcomeApplied] = 99
	g.Expect(m.OutcomeCounts()[FixOutcomeApplied]).To(Equal(2))
}

func TestMetrics_RecordOutcome_Concurrent(t *testing.T) {
	g := NewParallelGomega(t)

	m := NewMetrics()

	statuses := []FixOutcomeStatus{FixOutcomeApplied, FixOutcomeRefused, FixOutcomeFailed}

	var wg sync.WaitGroup

	for _, status := range statuses {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 50 {
				m.RecordOutcome(status)
			}
		}()
	}

	wg.Wait()

	g.Expect(m.OutcomeCounts()).To(Equal(map[FixOutcomeStatus]int{
		FixOutcomeApplied: 50,
		FixOutcomeRefused: 50,
		FixOutcomeFailed:  50,
	}))
}
