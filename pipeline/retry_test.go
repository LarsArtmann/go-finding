package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
)

func newRetryDet(d Detector, maxRetries int) *RetryDetector {
	return NewRetryDetector(d, RetryConfig{MaxRetries: maxRetries, BaseDelay: time.Millisecond})
}

func TestRetryConfig_delay(t *testing.T) {
	t.Parallel()
	c := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 5 * time.Second}

	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 100 * time.Millisecond, 125 * time.Millisecond},
		{1, 200 * time.Millisecond, 250 * time.Millisecond},
		{2, 400 * time.Millisecond, 500 * time.Millisecond},
		{3, 800 * time.Millisecond, 1000 * time.Millisecond},
	}

	for _, tt := range tests {
		got := c.delay(tt.attempt)
		if got < tt.min || got > tt.max {
			t.Errorf("delay(%d) = %v, want [%v, %v]", tt.attempt, got, tt.min, tt.max)
		}
	}
}

func TestRetryConfig_delay_maxCap(t *testing.T) {
	t.Parallel()
	c := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 300 * time.Millisecond}

	got := c.delay(10)
	if got < 300*time.Millisecond || got > 375*time.Millisecond {
		t.Errorf(
			"delay(10) = %v, want [%v, %v] (capped + jitter)",
			got,
			300*time.Millisecond,
			375*time.Millisecond,
		)
	}
}

func TestRetryDetector_SuccessOnFirstTry(t *testing.T) {
	t.Parallel()
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{{ID: "F1"}}, nil
	})
	rd := newRetryDet(inner, 3)

	findings, err := rd.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertIntEq(t, len(findings), 1, "findings")
}

func TestRetryDetector_SuccessAfterRetries(t *testing.T) {
	t.Parallel()
	calls := 0
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("transient")
		}

		return []finding.Finding{{ID: "F1"}}, nil
	})

	rd := newRetryDet(inner, 3)

	findings, err := rd.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertIntEq(t, len(findings), 1, "findings")

	assert.Equal(t, 3, calls)
}

func TestRetryDetector_ExhaustedRetries(t *testing.T) {
	t.Parallel()
	inner := makeErrorDetector("permanent")

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond})

	_, err := rd.Detect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	assert.ErrorContains(t, err, "permanent")
}

func TestRetryDetector_ContextCancellation(t *testing.T) {
	t.Parallel()
	calls := 0
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		calls++

		return nil, errors.New("fail")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 10, BaseDelay: time.Millisecond})

	_, err := rd.Detect(ctx)
	if err == nil {
		t.Fatal("expected error")
	}

	assert.ErrorIs(t, err, context.Canceled)
}

func TestRetryDetector_Name(t *testing.T) {
	t.Parallel()
	inner := &mockDetector{name: "my-detector"}

	rd := NewRetryDetector(inner, DefaultRetryConfig())
	assert.Equal(t, "my-detector", rd.Name())
}

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()
	c := DefaultRetryConfig()
	assert.Equal(t, 3, c.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, c.BaseDelay)
	assert.Equal(t, 5*time.Second, c.MaxDelay)
}

func TestRetryConfig_Validate_BaseDelayExceedsMax(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: 2 * time.Second, MaxDelay: 1 * time.Second}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when BaseDelay > MaxDelay")
	}

	assert.ErrorIs(t, err, errBaseDelayExceedsMax)
}

func TestRetryConfig_Validate_MaxDelayZeroWithBaseDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 0}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when MaxDelay=0 with BaseDelay>0")
	}

	assert.ErrorIs(t, err, errMaxDelayZero)
}

func TestRetryConfig_Validate_NegativeMaxRetries(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxRetries: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxRetries")
	}

	assert.ErrorIs(t, err, errMaxRetriesNegative)
}

func TestRetryConfig_Validate_NegativeBaseDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative BaseDelay")
	}

	assert.ErrorIs(t, err, errBaseDelayNegative)
}

func TestRetryConfig_Validate_NegativeMaxDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxDelay: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxDelay")
	}

	assert.ErrorIs(t, err, errMaxDelayNegative)
}

func TestRetryConfig_Validate_Valid(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxRetries: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: 5 * time.Second}
	assert.NoError(t, c.Validate())
}
