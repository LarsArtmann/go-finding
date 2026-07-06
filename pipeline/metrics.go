package pipeline

import (
	"maps"
	"sync"
	"time"

	"github.com/larsartmann/go-finding/lockutil"
)

// Metrics collects timing and count data from pipeline execution.
// All fields are unexported; use accessor methods for thread-safe reads.
// Use Snapshot() for a point-in-time copy of all metrics.
type Metrics struct {
	mu             sync.Mutex
	stageDurations map[Stage]time.Duration
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
		stageDurations: make(map[Stage]time.Duration),
		detectorTimes:  make(map[string]time.Duration),
		findingsFound:  make(map[string]int),
	}
}

// record runs fn while holding m.mu and returns its result. Consolidates
// the m.mu.Lock()/m.mu.Unlock() boilerplate for write-side Metrics
// mutations. Returns struct{} when no value is needed.
func record[T any](m *Metrics, fn func() T) T {
	return lockutil.Locked(&m.mu, fn)
}

// read runs fn while holding m.mu and returns its result. Consolidates
// the read-side m.mu.Lock()/defer m.mu.Unlock() boilerplate for Metrics.
func readMetrics[T any](m *Metrics, fn func() T) T {
	return lockutil.Locked(&m.mu, fn)
}

// RecordStage records the duration of a pipeline stage.
func (m *Metrics) RecordStage(name Stage, d time.Duration) {
	record(m, func() struct{} {
		m.stageDurations[name] += d

		return struct{}{}
	})
}

// RecordDetector records the duration and findings count for a detector.
func (m *Metrics) RecordDetector(name string, d time.Duration, findings int) {
	record(m, func() struct{} {
		m.detectorTimes[name] += d
		m.findingsFound[name] += findings

		return struct{}{}
	})
}

// RecordFixes records multiple successful fix applications in a single mutex acquisition.
func (m *Metrics) RecordFixes(count uint) {
	record(m, func() struct{} {
		m.fixesApplied += int(count)

		return struct{}{}
	})
}

// StageDuration returns the total duration recorded for the named stage.
func (m *Metrics) StageDuration(name Stage) time.Duration {
	return readMetrics(m, func() time.Duration {
		return m.stageDurations[name]
	})
}

// DetectorTime returns the total duration recorded for the named detector.
func (m *Metrics) DetectorTime(name string) time.Duration {
	return readMetrics(m, func() time.Duration {
		return m.detectorTimes[name]
	})
}

// DetectorFindings returns the total findings count for the named detector.
func (m *Metrics) DetectorFindings(name string) int {
	return readMetrics(m, func() int {
		return m.findingsFound[name]
	})
}

// TotalFixesApplied returns the total number of fixes applied.
func (m *Metrics) TotalFixesApplied() int {
	return readMetrics(m, func() int {
		return m.fixesApplied
	})
}

// TotalDuration returns the total pipeline execution time.
// Returns 0 if the pipeline has not completed.
func (m *Metrics) TotalDuration() time.Duration {
	return readMetrics(m, func() time.Duration {
		if m.endTime.IsZero() || m.startTime.IsZero() {
			return 0
		}

		return max(m.endTime.Sub(m.startTime), 0)
	})
}

// StageTiming returns a function that records stage duration when called.
func (m *Metrics) StageTiming(name Stage) func() {
	start := time.Now()

	return func() {
		m.RecordStage(name, time.Since(start))
	}
}

// MetricsSnapshot holds a point-in-time copy of the current metrics.
type MetricsSnapshot struct {
	StartTime      time.Time
	EndTime        time.Time
	StageDurations map[Stage]time.Duration
	DetectorTimes  map[string]time.Duration
	FindingsFound  map[string]int
	FixesApplied   int
	TotalDuration  time.Duration
}

// StageDuration returns the duration for the named stage from the snapshot.
func (s MetricsSnapshot) StageDuration(name Stage) time.Duration {
	return s.StageDurations[name]
}

// SetStart records the pipeline start time.
func (m *Metrics) SetStart(t time.Time) {
	record(m, func() struct{} {
		m.startTime = t

		return struct{}{}
	})
}

// SetEnd records the pipeline end time.
func (m *Metrics) SetEnd(t time.Time) {
	record(m, func() struct{} {
		m.endTime = t

		return struct{}{}
	})
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return readMetrics(m, func() MetricsSnapshot {
		stages := make(map[Stage]time.Duration, len(m.stageDurations))
		maps.Copy(stages, m.stageDurations)

		detectors := make(map[string]time.Duration, len(m.detectorTimes))
		maps.Copy(detectors, m.detectorTimes)

		findings := make(map[string]int, len(m.findingsFound))
		maps.Copy(findings, m.findingsFound)

		total := time.Duration(0)

		if !m.endTime.IsZero() && !m.startTime.IsZero() {
			if d := m.endTime.Sub(m.startTime); d > 0 {
				total = d
			}
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
	})
}
