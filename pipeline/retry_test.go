package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

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
	inner := &mockDetector{
		name:     "test",
		findings: []finding.Finding{{ID: "F1"}},
	}
	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})

	findings, err := rd.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
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

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})

	findings, err := rd.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}

	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryDetector_ExhaustedRetries(t *testing.T) {
	t.Parallel()
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return nil, errors.New("permanent")
	})

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond})

	_, err := rd.Detect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errors.New("permanent")) {
		// The wrapped error should contain "permanent"
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
	}
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

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetryDetector_Name(t *testing.T) {
	t.Parallel()
	inner := &mockDetector{name: "my-detector"}

	rd := NewRetryDetector(inner, DefaultRetryConfig())
	if rd.Name() != "my-detector" {
		t.Errorf("expected my-detector, got %s", rd.Name())
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()
	c := DefaultRetryConfig()
	if c.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", c.MaxRetries)
	}

	if c.BaseDelay != 100*time.Millisecond {
		t.Errorf("expected BaseDelay=100ms, got %v", c.BaseDelay)
	}

	if c.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay=5s, got %v", c.MaxDelay)
	}
}

func TestRetryConfig_Validate_BaseDelayExceedsMax(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: 2 * time.Second, MaxDelay: 1 * time.Second}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when BaseDelay > MaxDelay")
	}

	if !errors.Is(err, errBaseDelayExceedsMax) {
		t.Errorf("expected errBaseDelayExceedsMax, got %v", err)
	}
}

func TestRetryConfig_Validate_MaxDelayZeroWithBaseDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 0}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when MaxDelay=0 with BaseDelay>0")
	}

	if !errors.Is(err, errMaxDelayZero) {
		t.Errorf("expected errMaxDelayZero, got %v", err)
	}
}

func TestRetryConfig_Validate_NegativeMaxRetries(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxRetries: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxRetries")
	}

	if !errors.Is(err, errMaxRetriesNegative) {
		t.Errorf("expected errMaxRetriesNegative, got %v", err)
	}
}

func TestRetryConfig_Validate_NegativeBaseDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{BaseDelay: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative BaseDelay")
	}

	if !errors.Is(err, errBaseDelayNegative) {
		t.Errorf("expected errBaseDelayNegative, got %v", err)
	}
}

func TestRetryConfig_Validate_NegativeMaxDelay(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxDelay: -1}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxDelay")
	}

	if !errors.Is(err, errMaxDelayNegative) {
		t.Errorf("expected errMaxDelayNegative, got %v", err)
	}
}

func TestRetryConfig_Validate_Valid(t *testing.T) {
	t.Parallel()

	c := RetryConfig{MaxRetries: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: 5 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
