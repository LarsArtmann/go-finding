package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixApplier_Apply_CancelledContext(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "cancel.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fixes := []finding.Finding{
		makeFixFinding("1", "package main", "package main // fixed", "cancel.go", 0),
	}

	applied, err := applier.Apply(ctx, fixes)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, applied)
}

func TestFixApplier_Backup_NonexistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.backup.Backup(filepath.Join(tempDir, "does-not-exist.go"))
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrIO)
}

func TestFixApplier_Restore_WithoutBackup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.backup.Restore(filepath.Join(tempDir, "never-backed-up.go"))
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrInternal)
}

func TestFixApplier_ApplyToFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	fixes := []finding.Finding{makeFixFinding("1", "old", "new", "", 0)}

	applied, err := applier.applyToFile(filepath.Join(tempDir, "missing.go"), fixes)
	require.Error(t, err)
	require.ErrorIs(t, err, finding.ErrIO)
	assert.Equal(t, 0, applied)
}

func TestFixApplier_ApplyToFile_NoMatchingBeforeCode(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "nomatch.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{makeFixFinding("1", "nonexistent_code", "replacement", "", 0)}

	applied, err := applier.applyToFile(testFile, fixes)
	require.NoError(t, err)
	assert.Equal(t, 0, applied, "no match")
}

func TestFixApplier_ApplyToFile_ReadOnlyFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "readonly.go")
	if err := writeFile(testFile, []byte("package main\nold()\n"), 0o444); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{makeFixFinding("1", "old()", "new()", "", 0)}

	applied, err := applier.applyToFile(testFile, fixes)
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrIO)

	_ = applied
}

func TestFixApplier_RollbackAll(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	file1 := filepath.Join(tempDir, "a.go")
	file2 := filepath.Join(tempDir, "b.go")

	orig1 := "package a\n"
	orig2 := "package b\n"

	writeTestFile(t, file1, []byte(orig1))

	writeTestFile(t, file2, []byte(orig2))

	require.NoError(t, applier.backup.Backup(file1))
	require.NoError(t, applier.backup.Backup(file2))

	writeTestFile(t, file1, []byte("modified a\n"))

	writeTestFile(t, file2, []byte("modified b\n"))

	if err := applier.backup.RollbackAll([]string{file1, file2}); err != nil {
		t.Fatalf("rollbackAll: %v", err)
	}

	data1, err := readFile(file1)
	if err != nil {
		t.Fatalf("read a.go: %v", err)
	}

	if string(data1) != orig1 {
		t.Errorf("a.go = %q, want %q", string(data1), orig1)
	}

	data2, err := readFile(file2)
	if err != nil {
		t.Fatalf("read b.go: %v", err)
	}

	if string(data2) != orig2 {
		t.Errorf("b.go = %q, want %q", string(data2), orig2)
	}
}

func TestFixApplier_RollbackAll_PartialFailure(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	goodFile := filepath.Join(tempDir, "good.go")
	writeTestFile(t, goodFile, []byte("package good\n"))

	require.NoError(t, applier.backup.Backup(goodFile))

	writeTestFile(t, goodFile, []byte("modified\n"))

	noBackupFile := filepath.Join(tempDir, "nobackup.go")
	err := applier.backup.RollbackAll([]string{goodFile, noBackupFile})
	require.Error(t, err)

	data, rErr := readFile(goodFile)
	require.NoError(t, rErr)
	assert.Equal(t, "package good\n", string(data))
}

func TestFixApplier_Apply_BackupFailureRollsBack(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	file1 := filepath.Join(tempDir, "first.go")
	writeTestFile(t, file1, []byte("package first\nold1()\n"))

	fixes := []finding.Finding{
		makeFixFinding("1", "old1()", "new1()", "first.go", 0),
		makeFixFinding("2", "old2()", "new2()", "nonexistent_dir/second.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	require.Error(t, err)
	assert.ErrorIs(t, err, finding.ErrIO)

	_ = applied
}

func TestFixApplier_Apply_EmptyFixesList(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	applied, err := applier.Apply(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 0, applied)
}

func TestFixApplier_Apply_ApplyToFileErrorRestoresAndRollsBack(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// File 1: will be successfully modified.
	file1 := filepath.Join(tempDir, "first.go")
	writeTestFile(t, file1, []byte("package first\nold1()\n"))

	// File 2: read-only so applyToFile write fails.
	file2 := filepath.Join(tempDir, "second.go")
	writeTestFile(t, file2, []byte("package second\nold2()\n"))
	errChmod := os.Chmod(file2, 0o444) //nolint:gosec // intentional read-only for test
	require.NoError(t, errChmod)
	t.Cleanup(func() {
		_ = os.Chmod(file2, 0o644) //nolint:gosec // restore permissions in cleanup
	})

	fixes := []finding.Finding{
		makeFixFinding("1", "old1()", "new1()", "first.go", 0),
		makeFixFinding("2", "old2()", "new2()", "second.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	require.Error(t, err)
	require.ErrorIs(t, err, finding.ErrConflict)
	assert.Equal(t, 1, applied,
		"first.go is alphabetically first and should be applied before second.go fails")

	// File 1 should have been applied then rolled back.
	data1, rErr := readFile(file1)
	require.NoError(t, rErr)
	assert.Equal(t, "package first\nold1()\n", string(data1), "file1 should be restored")

	// File 2 should be unchanged (write failed before modification).
	data2, rErr := readFile(file2)
	require.NoError(t, rErr)
	assert.Equal(t, "package second\nold2()\n", string(data2), "file2 should be unchanged")
}

func TestFixApplier_Apply_FixesWithNoFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{File: ""}},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	require.NoError(t, err)
	assert.Equal(t, 0, applied, "no files specified")
}

func TestFixApplier_ApplyToFile_RangeOutOfBounds(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "outofbounds.go")
	content := "package main\n\nfunc main() {}\n"
	writeTestFile(t, testFile, []byte(content))

	fixes := []finding.Finding{
		{
			ID:         "1",
			BeforeCode: "nonexistent",
			AfterCode:  "replacement",
			Range:      finding.NewRangePtr("outofbounds.go", 100, 1, 200, 1),
		},
	}

	applied, err := applier.applyToFile(testFile, fixes)
	require.NoError(t, err)
	assert.Equal(t, 0, applied, "range out of bounds")

	data, rErr := readFile(testFile)
	require.NoError(t, rErr)
	assert.Equal(t, content, string(data), "file modified despite out-of-bounds range")
}

func TestFixApplier_FileHash_Deterministic(t *testing.T) {
	t.Parallel()

	h1 := fileHash("test/path.go")
	h2 := fileHash("test/path.go")
	assert.Equal(t, h1, h2, "fileHash not deterministic")

	h3 := fileHash("different/path.go")
	assert.NotEqual(t, h1, h3, "fileHash collision for different paths")
}

func TestFixApplier_BackupDisabled(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)
	applier.backup.SetEnabled(false)

	testFile := filepath.Join(tempDir, "nobackup.go")
	writeTestFile(t, testFile, []byte("package main\nold()\n"))

	fixes := []finding.Finding{
		makeFixFinding("1", "old()", "new()", "nobackup.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	require.NoError(t, err)
	assert.Equal(t, 1, applied)
	assert.Empty(t, applier.backup.BackupPath("nobackup.go"), "backup disabled")
}

func TestFixApplier_NewFixApplier_Defaults(t *testing.T) {
	t.Parallel()

	applier := NewFixApplier("/tmp/test")

	assert.Equal(t, "/tmp/test", applier.rootDir)
	assert.True(t, applier.backup.IsEnabled(), "backupEnabled should be true")
	assert.NotEmpty(t, applier.backup.backupDir, "backupDir should be set")
}

func TestNewFixApplier_MkdirTempFallback(t *testing.T) {
	// Cannot run in parallel: t.Setenv is incompatible with t.Parallel.
	t.Setenv("TMPDIR", "/etc/passwd")

	applier := NewFixApplier("/tmp/test")
	assert.Equal(t, "/tmp/test", applier.rootDir)
	assert.True(t, applier.backup.IsEnabled(), "backupEnabled should be true")
	assert.NotEmpty(t, applier.backup.backupDir, "backupDir should be set")
}

func TestFixApplier_Apply_RestoreOnApplyError(t *testing.T) {
	t.Parallel()

	applier := newTestApplier(t)
	testBackupRestore(t, applier,
		"package main\nold()\n",
		"package main\nnew()\n")
}

func TestIoErrorAt_WrapsCorrectly(t *testing.T) {
	t.Parallel()

	err := ioErrorAt("test op", os.ErrPermission, "file.go")

	require.ErrorIs(t, err, finding.ErrIO, "ErrIO sentinel match")
	require.ErrorIs(t, err, os.ErrPermission, "os.ErrPermission cause match")

	var fe *finding.FindingError
	require.ErrorAs(t, err, &fe, "expected FindingError")

	assertFindingErrorIO(t, fe, "file.go")
}
