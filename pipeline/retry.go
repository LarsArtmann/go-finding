package pipeline

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/larsartmann/go-finding"
)

const (
	delayJitterDivisor  = 4
	delayMaxMultiplier  = 100
	delayMaxRetryFactor = 5
)

var (
	errMaxRetriesNegative  = errors.New("MaxRetries must be >= 0")
	errBaseDelayNegative   = errors.New("BaseDelay must be >= 0")
	errMaxDelayNegative    = errors.New("MaxDelay must be >= 0")
	errBaseDelayExceedsMax = errors.New("BaseDelay must not exceed MaxDelay")
	errMaxDelayZero        = errors.New("MaxDelay must be > 0 when BaseDelay is set")
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

// Validate checks the retry configuration and returns an error if invalid.
func (c RetryConfig) Validate() error {
	var errs []error

	if c.MaxRetries < 0 {
		errs = append(errs, errMaxRetriesNegative)
	}

	if c.BaseDelay < 0 {
		errs = append(errs, errBaseDelayNegative)
	}

	if c.MaxDelay < 0 {
		errs = append(errs, errMaxDelayNegative)
	}

	if c.BaseDelay > 0 && c.MaxDelay > 0 && c.BaseDelay > c.MaxDelay {
		errs = append(errs, errBaseDelayExceedsMax)
	}

	if c.MaxDelay == 0 && c.BaseDelay > 0 {
		errs = append(errs, errMaxDelayZero)
	}

	return errors.Join(errs...)
}

// delay calculates the backoff duration for the given attempt with jitter.
func (c RetryConfig) delay(attempt int) time.Duration {
	d := min(time.Duration(1<<attempt)*c.BaseDelay, c.MaxDelay)
	if quarter := int64(d) / delayJitterDivisor; quarter > 0 {
		jitter := time.Duration(rand.Int64N(quarter))
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
