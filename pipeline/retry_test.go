package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
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
	g := NewParallelGomega(t)

	inner := makeFindingDetectorFunc("F1")
	rd := newRetryDet(inner, 3)

	findings, err := rd.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(findings).To(HaveLen(1))
}

func TestRetryDetector_SuccessAfterRetries(t *testing.T) {
	g := NewParallelGomega(t)

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

	g.Expect(findings).To(HaveLen(1))

	g.Expect(calls).To(Equal(3))
}

func TestRetryDetector_ExhaustedRetries(t *testing.T) {
	g := NewParallelGomega(t)

	inner := makeErrorDetector("permanent")

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond})

	_, err := rd.Detect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	g.Expect(err).To(MatchError(ContainSubstring("permanent")))
}

func TestRetryDetector_ContextCancellation(t *testing.T) {
	g := NewParallelGomega(t)

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

	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

func TestRetryDetector_ContextErrorFromInnerNotRetried(t *testing.T) {
	g := NewParallelGomega(t)

	calls := 0
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		calls++

		return nil, context.Canceled
	})

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})

	_, err := rd.Detect(context.Background())
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())

	g.Expect(calls).To(Equal(1), "context.Canceled should not be retried")
}

func TestRetryDetector_DeadlineExceededFromInnerNotRetried(t *testing.T) {
	g := NewParallelGomega(t)

	calls := 0
	inner := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		calls++

		return nil, context.DeadlineExceeded
	})

	rd := NewRetryDetector(inner, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond})

	_, err := rd.Detect(context.Background())
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())

	g.Expect(calls).To(Equal(1), "context.DeadlineExceeded should not be retried")
}

func TestRetryDetector_Name(t *testing.T) {
	g := NewParallelGomega(t)
	inner := &mockDetector{name: "my-detector"}

	rd := NewRetryDetector(inner, DefaultRetryConfig())
	g.Expect(rd.Name()).To(Equal("my-detector"))
}

func TestDefaultRetryConfig(t *testing.T) {
	g := NewParallelGomega(t)
	c := DefaultRetryConfig()
	g.Expect(c.MaxRetries).To(Equal(3))
	g.Expect(c.BaseDelay).To(Equal(100 * time.Millisecond))
	g.Expect(c.MaxDelay).To(Equal(5 * time.Second))
}

func TestRetryConfig_Validate_BaseDelayExceedsMax(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{BaseDelay: 2 * time.Second, MaxDelay: 1 * time.Second}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when BaseDelay > MaxDelay")
	}

	g.Expect(errors.Is(err, errBaseDelayExceedsMax)).To(BeTrue())
}

func TestRetryConfig_Validate_MaxDelayZeroWithBaseDelay(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{BaseDelay: 100 * time.Millisecond, MaxDelay: 0}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error when MaxDelay=0 with BaseDelay>0")
	}

	g.Expect(errors.Is(err, errMaxDelayZero)).To(BeTrue())
}

func TestRetryConfig_Validate_NegativeMaxRetries(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{MaxRetries: -1}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxRetries")
	}

	g.Expect(errors.Is(err, errMaxRetriesNegative)).To(BeTrue())
}

func TestRetryConfig_Validate_NegativeBaseDelay(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{BaseDelay: -1}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative BaseDelay")
	}

	g.Expect(errors.Is(err, errBaseDelayNegative)).To(BeTrue())
}

func TestRetryConfig_Validate_NegativeMaxDelay(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{MaxDelay: -1}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxDelay")
	}

	g.Expect(errors.Is(err, errMaxDelayNegative)).To(BeTrue())
}

func TestRetryConfig_Validate_Valid(t *testing.T) {
	g := NewParallelGomega(t)

	c := RetryConfig{MaxRetries: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: 5 * time.Second}
	g.Expect(c.Validate()).NotTo(HaveOccurred())
}
