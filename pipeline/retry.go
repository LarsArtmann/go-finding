package pipeline

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/larsartmann/go-finding"
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
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}
}

// delay calculates the backoff duration for the given attempt.
func (c RetryConfig) delay(attempt int) time.Duration {
	d := time.Duration(math.Pow(2, float64(attempt))) * c.BaseDelay
	if d > c.MaxDelay {
		return c.MaxDelay
	}
	return d
}

// RetryDetector wraps a Detector with retry logic on error.
type RetryDetector struct {
	inner Detector
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
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		findings, err := d.inner.Detect(ctx)
		if err == nil {
			return findings, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("detector %s failed after %d retries: %w", d.inner.Name(), d.config.MaxRetries, lastErr)
}
