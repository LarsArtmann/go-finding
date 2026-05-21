package pipeline

import (
	"context"
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
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	g.Expect(m).NotTo(BeNil())

	snap := m.Snapshot()
	g.Expect(snap.StageDurations).NotTo(BeNil())
	g.Expect(snap.DetectorTimes).NotTo(BeNil())
	g.Expect(snap.FindingsFound).NotTo(BeNil())
}

func TestMetrics_RecordStage(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordStage("detect", 50*time.Millisecond)
	m.RecordStage("apply", 200*time.Millisecond)

	g.Expect(m.StageDuration("detect")).To(Equal(150 * time.Millisecond))
	g.Expect(m.StageDuration("apply")).To(Equal(200 * time.Millisecond))
}

func TestMetrics_RecordDetector(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.RecordDetector("staticcheck", 50*time.Millisecond, 10)
	m.RecordDetector("staticcheck", 30*time.Millisecond, 5)
	m.RecordDetector("govet", 20*time.Millisecond, 3)

	g.Expect(m.DetectorTime("staticcheck")).To(Equal(80 * time.Millisecond))
	g.Expect(m.DetectorFindings("staticcheck")).To(Equal(15))
	g.Expect(m.DetectorFindings("govet")).To(Equal(3))
}

func TestMetrics_RecordFix(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.RecordFix()
	m.RecordFix()
	m.RecordFix()

	g.Expect(m.TotalFixesApplied()).To(Equal(3))
}

func TestMetrics_StageTiming(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	done := m.StageTiming("detect")

	time.Sleep(10 * time.Millisecond)
	done()

	g.Expect(m.StageDuration("detect")).To(BeNumerically(">=", 10*time.Millisecond))
}

func TestMetrics_TotalDuration(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.SetStart(time.Now())

	time.Sleep(10 * time.Millisecond)

	m.SetEnd(time.Now())

	d := m.TotalDuration()
	g.Expect(d).To(BeNumerically(">=", 10*time.Millisecond))
}

func TestMetrics_Snapshot(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.RecordStage("detect", 100*time.Millisecond)
	m.RecordDetector("govet", 50*time.Millisecond, 5)
	m.RecordFix()
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
	t.Parallel()
	g := NewWithT(t)
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
	g.Expect(result.Stable).To(BeTrue())

	snap := m.Snapshot()
	g.Expect(snap.StartTime.IsZero()).To(BeFalse())
	g.Expect(snap.EndTime.IsZero()).To(BeFalse())

	_, ok := snap.StageDurations["detect"]
	g.Expect(ok).To(BeTrue())
}

func TestMetrics_TotalDuration_NoEnd(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	m := NewMetrics()
	m.SetStart(time.Now())

	d := m.TotalDuration()
	g.Expect(d).To(Equal(time.Duration(0)))
}

func TestMetricsSnapshot_StageDuration(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	snap := MetricsSnapshot{
		StageDurations: map[string]time.Duration{
			"detect": 100 * time.Millisecond,
		},
	}

	g.Expect(snap.StageDuration("detect")).To(Equal(100 * time.Millisecond))
	g.Expect(snap.StageDuration("nonexistent")).To(Equal(time.Duration(0)))
}

func TestPipeline_NilMetricsNoPanic(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	detector := &mockDetector{
		name:     "test",
		findings: []finding.Finding{},
	}

	result := runPipelineWithMetrics(t, testConfig(1, nil), detector)
	g.Expect(result.Stable).To(BeTrue())
}

func TestMetrics_TotalDuration_NegativeGuard(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	m := NewMetrics()
	m.SetStart(time.Now())
	m.SetEnd(time.Now().Add(-1 * time.Hour))

	d := m.TotalDuration()
	g.Expect(d).To(Equal(time.Duration(0)))
}

func TestMetrics_Snapshot_NegativeGuard(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	m := NewMetrics()
	m.SetStart(time.Now())
	m.SetEnd(time.Now().Add(-1 * time.Hour))

	snap := m.Snapshot()
	g.Expect(snap.TotalDuration).To(Equal(time.Duration(0)))
}
