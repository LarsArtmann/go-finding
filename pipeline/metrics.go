package pipeline

import (
	"maps"
	"sync"
	"time"
)

// Metrics collects timing and count data from pipeline execution.
type Metrics struct {
	mu             sync.Mutex
	StageDurations map[string]time.Duration
	DetectorTimes  map[string]time.Duration
	FindingsFound  map[string]int
	FixesApplied   int
	StartTime      time.Time
	EndTime        time.Time
}

// NewMetrics creates a new Metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		StageDurations: make(map[string]time.Duration),
		DetectorTimes:  make(map[string]time.Duration),
		FindingsFound:  make(map[string]int),
	}
}

// RecordStage records the duration of a pipeline stage.
func (m *Metrics) RecordStage(name string, d time.Duration) {
	m.mu.Lock()
	m.StageDurations[name] += d
	m.mu.Unlock()
}

// RecordDetector records the duration and findings count for a detector.
func (m *Metrics) RecordDetector(name string, d time.Duration, findings int) {
	m.mu.Lock()
	m.DetectorTimes[name] += d
	m.FindingsFound[name] += findings
	m.mu.Unlock()
}

// RecordFix records a successful fix application.
func (m *Metrics) RecordFix() {
	m.mu.Lock()
	m.FixesApplied++
	m.mu.Unlock()
}

// TotalDuration returns the total pipeline execution time.
func (m *Metrics) TotalDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.EndTime.Sub(m.StartTime)
}

// StageTiming returns a function that records stage duration when called.
func (m *Metrics) StageTiming(name string) func() {
	start := time.Now()

	return func() {
		m.RecordStage(name, time.Since(start))
	}
}

// MetricsSnapshot holds a point-in-time copy of the current metrics.
type MetricsSnapshot struct {
	StageDurations map[string]time.Duration
	DetectorTimes  map[string]time.Duration
	FindingsFound  map[string]int
	FixesApplied   int
	TotalDuration  time.Duration
}

// SetStart records the pipeline start time.
func (m *Metrics) SetStart(t time.Time) {
	m.mu.Lock()
	m.StartTime = t
	m.mu.Unlock()
}

// SetEnd records the pipeline end time.
func (m *Metrics) SetEnd(t time.Time) {
	m.mu.Lock()
	m.EndTime = t
	m.mu.Unlock()
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	stages := make(map[string]time.Duration, len(m.StageDurations))
	maps.Copy(stages, m.StageDurations)

	detectors := make(map[string]time.Duration, len(m.DetectorTimes))
	maps.Copy(detectors, m.DetectorTimes)

	findings := make(map[string]int, len(m.FindingsFound))
	maps.Copy(findings, m.FindingsFound)

	return MetricsSnapshot{
		StageDurations: stages,
		DetectorTimes:  detectors,
		FindingsFound:  findings,
		FixesApplied:   m.FixesApplied,
		TotalDuration:  m.EndTime.Sub(m.StartTime),
	}
}
