package pipeline

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestNew(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()
	detector := &mockDetector{name: "test", findings: nil}

	p, err := New(config, "/tmp", detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}

	g.Expect(p.detectors).To(HaveLen(1))

	g.Expect(p.rootDir).To(Equal("/tmp"))

	if p.config.MaxIterations != config.MaxIterations {
		t.Error("config not set correctly")
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	config := DefaultConfig()

	g.Expect(config.MaxIterations).To(Equal(5))

	g.Expect(config.ParallelDetectors).To(BeTrue())

	g.Expect(config.Timeout).To(Equal(10 * time.Minute))
}

func TestNew_RejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config Config
	}{
		{"negative iterations", Config{MaxIterations: -1}},
		{"negative timeout", Config{MaxIterations: 1, Timeout: -1 * time.Second}},
		{"invalid retry", Config{MaxIterations: 1, Retry: &RetryConfig{MaxRetries: -1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			_, err := New(tt.config, t.TempDir())
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNew_ValidConfig_NoError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	config := DefaultConfig()

	p, err := New(config, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(p).NotTo(BeNil())
}
