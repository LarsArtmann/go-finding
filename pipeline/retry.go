package pipeline

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/larsartmann/go-finding"
)

const (
	delayJitterDivisor  = 4
	delayMaxMultiplier  = 100
	delayMaxRetryFactor = 5
)

// RetryConfig configures retry behavior for detectors.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns a sensible retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Duration(delayMaxMultiplier) * time.Millisecond,
		MaxDelay:   time.Duration(delayMaxRetryFactor) * time.Second,
	}
}

// delay calculates the backoff duration for the given attempt with jitter.
func (c RetryConfig) delay(attempt int) time.Duration {
	d := min(time.Duration(math.Pow(2, float64(attempt)))*c.BaseDelay, c.MaxDelay)
	if quarter := int64(d) / delayJitterDivisor; quarter > 0 {
		jitter := time.Duration(rand.Int63n(quarter))
		d += jitter
	}

	return d
}

// RetryDetector wraps a Detector with retry logic on error.
type RetryDetector struct {
	inner  Detector
	config RetryConfig
}

// NewRetryDetector creates a detector that retries on failure with exponential backoff.
func NewRetryDetector(inner Detector, config RetryConfig) *RetryDetector {
	return &RetryDetector{inner: inner, config: config}
}

// Name implements Detector.
func (d *RetryDetector) Name() string {
	return d.inner.Name()
}

// Detect implements Detector, retrying on error up to MaxRetries times.
func (d *RetryDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	var lastErr error

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := d.config.delay(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("retry cancelled: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		findings, err := d.inner.Detect(ctx)
		if err == nil {
			return findings, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf(
		"detector %s failed after %d retries: %w",
		d.inner.Name(),
		d.config.MaxRetries,
		lastErr,
	)
}
