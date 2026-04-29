package pipeline

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOnFix_FiresOnlyForAppliedFixes verifies C-1: OnFix callback fires
// exactly once per actually-applied fix, not once per safeFix.
func TestOnFix_FiresOnlyForAppliedFixes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	// File contains two occurrences.
	original := "package main\n\nfunc main() {\n\told()\n\told()\n}\n" //nolint:dupword // test fixture
	writeTestFile(t, testFile, []byte(original))

	// One fix that targets the second occurrence at line 5.
	fix := finding.Finding{
		ID:          "fix1",
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "replace old",
		BeforeCode:  "old()",
		AfterCode:   "new()",
		Position:    finding.Position{File: "fixme.go", Line: 5},
		FixStrategy: finding.FixStrategyDirect,
	}

	var onFixCalls int
	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFix: func(_ finding.Finding, wasApplied bool) {
			if wasApplied {
				onFixCalls++
			}
		},
	}

	det := &mockDetector{name: "tool", findings: []finding.Finding{fix}}
	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// One fix was applied.
	assert.Equal(t, 1, result.Iterations[0].Applied, "expected 1 applied fix")
	// OnFix should have been called exactly once for the applied fix.
	assert.Equal(t, 1, onFixCalls, "OnFix should fire exactly once per applied fix")
}

// TestOnFix_SkipsUnappliedFixes verifies that when a fix is skipped
// (e.g., BeforeCode not found), OnFix is not called with wasApplied=true.
func TestOnFix_SkipsUnappliedFixes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	// Fix with BeforeCode that doesn't exist in the file.
	fix := finding.Finding{
		ID:          "fix1",
		BeforeCode:  "nonexistent",
		AfterCode:   "replacement",
		Position:    finding.Position{File: "fixme.go", Line: 1},
		FixStrategy: finding.FixStrategyDirect,
	}

	var appliedCount int
	cfg := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		OnFix: func(_ finding.Finding, wasApplied bool) {
			if wasApplied {
				appliedCount++
			}
		},
	}

	det := &mockDetector{name: "tool", findings: []finding.Finding{fix}}
	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	assert.Equal(t, 0, appliedCount, "OnFix should not fire for skipped fixes")
}

// TestPipelineRun_ParallelDetectorError verifies that a detector error in
// parallel mode propagates correctly from detectParallel.
func TestPipelineRun_ParallelDetectorError(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ParallelDetectors = true

	expectedErr := errors.New("parallel detector failed")
	d1 := &mockDetector{name: "ok", findings: []finding.Finding{{ID: "1"}}}
	d2 := &mockDetector{name: "fail", err: expectedErr}

	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "parallel detection")
}

// TestApplyTriage_AllConflicts verifies that when all fixes are skipped by
// FilterConflictingFixes (no file info), applyTriage returns nil without
// calling applyDirectFixes.
func TestApplyTriage_AllConflicts(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Fixes without file info are skipped entirely by FilterConflictingFixes.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			Rule:        "r1",
			ToolName:    "t1",
			Message:     "m1",
			Position:    finding.Position{Line: 3, Column: 1}, // File is empty
			FixStrategy: finding.FixStrategyDirect,
			BeforeCode:  "func",
			AfterCode:   "fn",
		},
	}

	cfg := DefaultConfig()
	p, err := New(cfg, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	iter := Iteration{Number: 1}
	err = p.applyTriage(context.Background(), fixes, &iter)
	if err != nil {
		t.Fatalf("applyTriage: %v", err)
	}

	assert.Equal(t, 1, iter.Conflicts, "expected 1 conflict (fix skipped)")
	assert.Equal(t, 0, iter.Applied, "expected 0 applied fixes")
}

// TestApplyTriage_ApplyError verifies that applyTriage returns an error when
// applyDirectFixes fails (e.g., target file does not exist).
func TestApplyTriage_ApplyError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// A single fix pointing to a non-existent file.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			Rule:        "r1",
			ToolName:    "t1",
			Message:     "m1",
			Position:    finding.Position{File: "missing.go", Line: 1, Column: 1},
			FixStrategy: finding.FixStrategyDirect,
			BeforeCode:  "old",
			AfterCode:   "new",
		},
	}

	cfg := DefaultConfig()
	p, err := New(cfg, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	iter := Iteration{Number: 1}
	err = p.applyTriage(context.Background(), fixes, &iter)
	require.Error(t, err)
	assert.ErrorContains(t, err, "apply fixes")
}
