package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func assertPartialFindingsLen(t *testing.T, result *PartialResult, want int) {
	t.Helper()
	g := NewWithT(t)

	g.Expect(result.Findings).To(HaveLen(want))
}

// goodThenBadDetectors returns two mock detectors: one that succeeds with a finding,
// and one that fails with an error. Used to test partial success scenarios.
func goodThenBadDetectors() (*mockDetector, *mockDetector) {
	d1 := mockDet("good", "F1")
	d2 := &mockDetector{name: "bad", err: errors.New("boom")}
	return d1, d2
}

func runDetectPartial(t *testing.T, parallel bool, d1, d2 *mockDetector) *PartialResult {
	t.Helper()
	g := NewWithT(t)

	config := Config{ParallelDetectors: parallel}
	p, err := New(config, t.TempDir(), d1, d2)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.DetectPartial(context.Background())
	g.Expect(err).NotTo(HaveOccurred())

	return result
}

func TestPartialResult_HasErrors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	r := &PartialResult{Errors: make(map[string]error)}
	g.Expect(r.HasErrors()).To(BeFalse())

	r.Errors["det1"] = errors.New("fail")
	g.Expect(r.HasErrors()).To(BeTrue())
}

func TestDetectPartial_Sequential_AllSucceed(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	d1 := mockDet("d1", "F1")
	d2 := mockDet("d2", "F2")

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 2)
	g.Expect(result.HasErrors()).To(BeFalse())
}

func TestDetectPartial_Sequential_PartialFailure(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	d1, d2 := goodThenBadDetectors()

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 1)
	g.Expect(result.HasErrors()).To(BeTrue())
	g.Expect(result.Errors["bad"]).To(HaveOccurred())
}

func TestDetectPartial_Parallel_PartialFailure(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	d1, d2 := goodThenBadDetectors()

	result := runDetectPartial(t, true, d1, d2)

	assertPartialFindingsLen(t, result, 1)
	g.Expect(result.HasErrors()).To(BeTrue())
}

func TestDetectPartial_AllFail(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	d1 := &mockDetector{name: "d1", err: errors.New("fail1")}
	d2 := &mockDetector{name: "d2", err: errors.New("fail2")}

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 0)
	g.Expect(result.Errors).To(HaveLen(2))
}

func TestFormatPartialErrors(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	g.Expect(FormatPartialErrors(nil)).NotTo(HaveOccurred())
	g.Expect(FormatPartialErrors(map[string]error{})).NotTo(HaveOccurred())

	errs := map[string]error{
		"det1": errors.New("fail1"),
		"det2": errors.New("fail2"),
	}

	err := FormatPartialErrors(errs)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).NotTo(BeEmpty())
}

func TestFormatPartialErrors_ErrorWrapping(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	inner := errors.New("inner error")
	errs := map[string]error{"det1": inner}

	wrapped := FormatPartialErrors(errs)
	g.Expect(wrapped).To(HaveOccurred())
	g.Expect(errors.Is(wrapped, ErrPartialDetection)).To(BeTrue())
	g.Expect(errors.Is(wrapped, inner)).To(BeTrue())
}

func TestDetectPartial_Sequential_CancelBeforeSecond(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	d1 := mockDet("fast", "F1")
	d2 := &mockDetector{
		name:     "slow",
		delay:    5 * time.Second,
		findings: []finding.Finding{{ID: "F2"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := Config{ParallelDetectors: false}
	p, err := New(config, t.TempDir(), d1, d2)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.DetectPartial(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).NotTo(BeNil())
}

func TestDetectPartial_Parallel_CancelReturnsPartial(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	d1 := mockDet("fast", "F1")
	d2 := &mockDetector{
		name:     "slow",
		delay:    5 * time.Second,
		findings: []finding.Finding{{ID: "F2"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := Config{ParallelDetectors: true}
	p, err := New(config, t.TempDir(), d1, d2)
	g.Expect(err).NotTo(HaveOccurred())

	result, err := p.DetectPartial(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).NotTo(BeNil())
}
