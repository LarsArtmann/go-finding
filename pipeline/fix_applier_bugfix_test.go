package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFixApplier_InsertionOnly verifies C-2: findings with only AfterCode
// (and no BeforeCode) are treated as insertions at the finding's line.
func TestFixApplier_InsertionOnly(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "insert.go")
	content := "package main\n\nfunc main() {\n}\n"
	writeTestFile(t, testFile, []byte(content))

	// Insertion: add "\tprintln(\"hello\")\n" at line 4 (before the closing brace).
	fix := finding.Finding{
		ID:          "insert1",
		AfterCode:   "\tprintln(\"hello\")",
		Position:    finding.Position{File: "insert.go", Line: 4},
		FixStrategy: finding.FixStrategyDirect,
	}

	applier := NewFixApplier(tempDir)
	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	require.NoError(t, err)
	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	require.NoError(t, err)

	want := helloProgram
	assert.Equal(t, want, string(got))
}

// TestFixApplier_DeletionOnly verifies C-2: findings with only BeforeCode
// (and no AfterCode) are treated as deletions.
func TestFixApplier_DeletionOnly(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "delete.go")
	content := helloProgram
	writeTestFile(t, testFile, []byte(content))

	// Deletion: remove "\tprintln(\"hello\")\n".
	fix := finding.Finding{
		ID:          "delete1",
		BeforeCode:  "\tprintln(\"hello\")",
		Position:    finding.Position{File: "delete.go", Line: 4},
		FixStrategy: finding.FixStrategyDirect,
	}

	applier := NewFixApplier(tempDir)
	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	require.NoError(t, err)
	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	require.NoError(t, err)

	// Deletion removes the BeforeCode substring but leaves the trailing newline,
	// resulting in a blank line. To delete the entire line including newline,
	// include "\n" in BeforeCode.
	want := "package main\n\nfunc main() {\n\n}\n"
	assert.Equal(t, want, string(got))
}

// TestFixApplier_NearestLineReplacement verifies H-7: when BeforeCode
// appears multiple times, the occurrence nearest to the finding's line is replaced.
func TestFixApplier_NearestLineReplacement(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nearest.go")
	// "old()" appears at line 4 and line 6.
	content := "package main\n\nfunc main() {\n\told()\n\tother()\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(content))

	// Fix targets line 6 (the second occurrence).
	fix := finding.Finding{
		ID:          "fix1",
		BeforeCode:  "old()",
		AfterCode:   "new()",
		Position:    finding.Position{File: "nearest.go", Line: 6},
		FixStrategy: finding.FixStrategyDirect,
	}

	applier := NewFixApplier(tempDir)
	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	require.NoError(t, err)
	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	require.NoError(t, err)

	// Only the second "old()" at line 6 should be replaced.
	want := "package main\n\nfunc main() {\n\told()\n\tother()\n\tnew()\n}\n"
	assert.Equal(t, want, string(got))
}
