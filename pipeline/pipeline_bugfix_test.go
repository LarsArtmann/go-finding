package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
)

// TestOnFix_FiresOnlyForAppliedFixes verifies C-1: OnFix callback fires
// exactly once per actually-applied fix, not once per safeFix.
func TestOnFix_FiresOnlyForAppliedFixes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")
	// File has two occurrences of "old()".
	original := "package main\n\nfunc main() {\n\told()\n\told()\n}\n"
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
