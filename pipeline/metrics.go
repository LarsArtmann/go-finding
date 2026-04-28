package pipeline

import (
	"maps"
	"sync"
	"time"
)

// Metrics collects timing and count data from pipeline execution.
// All fields are unexported; use accessor methods for thread-safe reads.
// Use Snapshot() for a point-in-time copy of all metrics.
type Metrics struct {
	mu             sync.Mutex
	stageDurations map[string]time.Duration
	detectorTimes  map[string]time.Duration
	findingsFound  map[string]int
	fixesApplied   int
	startTime      time.Time
	endTime        time.Time
}

// NewMetrics creates a new Metrics collector.
func NewMetrics() *Metrics {
	//nolint:exhaustruct
	return &Metrics{
		stageDurations: make(map[string]time.Duration),
		detectorTimes:  make(map[string]time.Duration),
		findingsFound:  make(map[string]int),
	}
}

// RecordStage records the duration of a pipeline stage.
func (m *Metrics) RecordStage(name string, d time.Duration) {
	m.mu.Lock()
	m.stageDurations[name] += d
	m.mu.Unlock()
}

// RecordDetector records the duration and findings count for a detector.
func (m *Metrics) RecordDetector(name string, d time.Duration, findings int) {
	m.mu.Lock()
	m.detectorTimes[name] += d
	m.findingsFound[name] += findings
	m.mu.Unlock()
}

// RecordFix records a successful fix application.
func (m *Metrics) RecordFix() {
	m.mu.Lock()
	m.fixesApplied++
	m.mu.Unlock()
}

// StageDuration returns the total duration recorded for the named stage.
func (m *Metrics) StageDuration(name string) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stageDurations[name]
}

// DetectorTime returns the total duration recorded for the named detector.
func (m *Metrics) DetectorTime(name string) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.detectorTimes[name]
}

// DetectorFindings returns the total findings count for the named detector.
func (m *Metrics) DetectorFindings(name string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findingsFound[name]
}

// TotalFixesApplied returns the total number of fixes applied.
func (m *Metrics) TotalFixesApplied() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fixesApplied
}

// TotalDuration returns the total pipeline execution time.
// Returns 0 if the pipeline has not completed.
func (m *Metrics) TotalDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.endTime.IsZero() || m.startTime.IsZero() {
		return 0
	}

	return m.endTime.Sub(m.startTime)
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
	StartTime      time.Time
	EndTime        time.Time
	StageDurations map[string]time.Duration
	DetectorTimes  map[string]time.Duration
	FindingsFound  map[string]int
	FixesApplied   int
	TotalDuration  time.Duration
}

// StageDuration returns the duration for the named stage from the snapshot.
func (s MetricsSnapshot) StageDuration(name string) time.Duration {
	return s.StageDurations[name]
}

// SetStart records the pipeline start time.
func (m *Metrics) SetStart(t time.Time) {
	m.mu.Lock()
	m.startTime = t
	m.mu.Unlock()
}

// SetEnd records the pipeline end time.
func (m *Metrics) SetEnd(t time.Time) {
	m.mu.Lock()
	m.endTime = t
	m.mu.Unlock()
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	stages := make(map[string]time.Duration, len(m.stageDurations))
	maps.Copy(stages, m.stageDurations)

	detectors := make(map[string]time.Duration, len(m.detectorTimes))
	maps.Copy(detectors, m.detectorTimes)

	findings := make(map[string]int, len(m.findingsFound))
	maps.Copy(findings, m.findingsFound)

	total := time.Duration(0)
	if !m.endTime.IsZero() {
		total = m.endTime.Sub(m.startTime)
	}

	return MetricsSnapshot{
		StartTime:      m.startTime,
		EndTime:        m.endTime,
		StageDurations: stages,
		DetectorTimes:  detectors,
		FindingsFound:  findings,
		FixesApplied:   m.fixesApplied,
		TotalDuration:  total,
	}
}
