package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPartialFindingsLen(t *testing.T, result *PartialResult, want int, msg string) {
	t.Helper()

	assert.Len(t, result.Findings, want, msg)
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

	config := Config{ParallelDetectors: parallel}
	p, err := New(config, t.TempDir(), d1, d2)
	require.NoError(t, err)

	result, err := p.DetectPartial(context.Background())
	require.NoError(t, err)

	return result
}

func TestPartialResult_HasErrors(t *testing.T) {
	t.Parallel()
	r := &PartialResult{Errors: make(map[string]error)}
	assert.False(t, r.HasErrors(), "empty errors should not report HasErrors")

	r.Errors["det1"] = errors.New("fail")
	assert.True(t, r.HasErrors(), "should report HasErrors with one error")
}

func TestDetectPartial_Sequential_AllSucceed(t *testing.T) {
	t.Parallel()
	d1 := mockDet("d1", "F1")
	d2 := mockDet("d2", "F2")

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 2, "expected 2 findings")
	assert.False(t, result.HasErrors(), "expected no errors")
}

func TestDetectPartial_Sequential_PartialFailure(t *testing.T) {
	t.Parallel()
	d1, d2 := goodThenBadDetectors()

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 1, "expected 1 finding from good detector")
	assert.True(t, result.HasErrors(), "expected errors from bad detector")
	assert.Error(t, result.Errors["bad"], "expected error for bad detector")
}

func TestDetectPartial_Parallel_PartialFailure(t *testing.T) {
	t.Parallel()
	d1, d2 := goodThenBadDetectors()

	result := runDetectPartial(t, true, d1, d2)

	assertPartialFindingsLen(t, result, 1, "expected 1 finding")
	assert.True(t, result.HasErrors(), "expected errors")
}

func TestDetectPartial_AllFail(t *testing.T) {
	t.Parallel()
	d1 := &mockDetector{name: "d1", err: errors.New("fail1")}
	d2 := &mockDetector{name: "d2", err: errors.New("fail2")}

	result := runDetectPartial(t, false, d1, d2)

	assertPartialFindingsLen(t, result, 0, "expected 0 findings")
	assert.Len(t, result.Errors, 2)
}

func TestFormatPartialErrors(t *testing.T) {
	t.Parallel()
	assert.NoError(t, FormatPartialErrors(nil))
	assert.NoError(t, FormatPartialErrors(map[string]error{}))

	errs := map[string]error{
		"det1": errors.New("fail1"),
		"det2": errors.New("fail2"),
	}

	err := FormatPartialErrors(errs)
	require.Error(t, err)
	assert.NotEmpty(t, err.Error())
}
