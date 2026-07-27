package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/onsi/gomega"
)

func TestStageHook_RecordsEvents(t *testing.T) {
	g := NewParallelGomega(t)

	var (
		mu     sync.Mutex
		events []StageEvent
	)

	hook := StageHookFunc(func(_ context.Context, event StageEvent) error {
		mu.Lock()

		events = append(events, event)

		mu.Unlock()

		return nil
	})

	d := newMockDetector("test", "tool", finding.NewFinding(
		"rule", "tool", "msg", finding.SeverityInfo,
		finding.Position{}, finding.ConfidenceHigh,
	))

	cfg := DefaultConfig()
	cfg.MaxIterations = 1
	cfg.StageHooks = []StageHook{hook}

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result, err := p.Run(context.Background())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(result.Reason).To(gomega.Equal(ReasonMaxIterations))

	mu.Lock()
	defer mu.Unlock()

	g.Expect(events).ToNot(gomega.BeEmpty())

	timings := make(map[StageTiming]bool)
	stages := make(map[Stage]bool)

	for _, e := range events {
		timings[e.Timing] = true
		stages[e.Stage] = true
	}

	g.Expect(timings).To(gomega.HaveKey(StageBefore))
	g.Expect(timings).To(gomega.HaveKey(StageAfter))
	g.Expect(stages).To(gomega.HaveKey(StageDetect))
	g.Expect(stages).To(gomega.HaveKey(StageTriage))
}

func TestStageHook_BeforeHookAborts(t *testing.T) {
	g := NewParallelGomega(t)

	abortErr := errors.New("hook abort")

	hook := StageHookFunc(func(_ context.Context, event StageEvent) error {
		if event.Timing == StageBefore && event.Stage == StageDetect {
			return abortErr
		}

		return nil
	})

	d := newMockDetector("test", "tool", finding.NewFinding(
		"rule", "tool", "msg", finding.SeverityInfo,
		finding.Position{}, finding.ConfidenceHigh,
	))

	cfg := DefaultConfig()
	cfg.MaxIterations = 1
	cfg.StageHooks = []StageHook{hook}

	p, err := New(cfg, t.TempDir(), d)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	_, err = p.Run(context.Background())
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("stage hook"))
}
