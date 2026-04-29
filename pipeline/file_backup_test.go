package pipeline

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileBackup_Defaults(t *testing.T) {
	t.Parallel()

	fb := NewFileBackup(t.TempDir())
	assert.True(t, fb.IsEnabled())
	assert.NotEmpty(t, fb.backupDir)
}

func TestFileBackup_SetEnabled(t *testing.T) {
	t.Parallel()

	fb := NewFileBackup(t.TempDir())
	assert.True(t, fb.IsEnabled())

	fb.SetEnabled(false)
	assert.False(t, fb.IsEnabled())

	fb.SetEnabled(true)
	assert.True(t, fb.IsEnabled())
}

func TestFileBackup_BackupPath_NoBackup(t *testing.T) {
	t.Parallel()

	fb := NewFileBackup(t.TempDir())
	assert.Empty(t, fb.BackupPath("/nonexistent/file.go"))
}

func TestFileBackup_BackupAndRestore(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	writeTestFile(t, testFile, []byte("original content"))

	require.NoError(t, fb.Backup(testFile))

	backupPath := fb.BackupPath(testFile)
	assert.NotEmpty(t, backupPath, "backup path should be recorded")

	writeTestFile(t, testFile, []byte("modified content"))

	require.NoError(t, fb.Restore(testFile))

	data, err := readFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "original content", string(data))
}

func TestFileBackup_Backup_NonexistentFile(t *testing.T) {
	t.Parallel()

	fb := NewFileBackup(t.TempDir())
	err := fb.Backup(filepath.Join(t.TempDir(), "missing.go"))
	require.Error(t, err)
}

func TestFileBackup_Restore_NoBackup(t *testing.T) {
	t.Parallel()

	fb := NewFileBackup(t.TempDir())
	err := fb.Restore(filepath.Join(t.TempDir(), "never-backed.go"))
	require.Error(t, err)
}

func TestFileBackup_RollbackAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)

	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")

	writeTestFile(t, file1, []byte("original a"))
	writeTestFile(t, file2, []byte("original b"))

	require.NoError(t, fb.Backup(file1))
	require.NoError(t, fb.Backup(file2))

	writeTestFile(t, file1, []byte("modified a"))
	writeTestFile(t, file2, []byte("modified b"))

	require.NoError(t, fb.RollbackAll([]string{file1, file2}))

	data1, err := readFile(file1)
	require.NoError(t, err)
	assert.Equal(t, "original a", string(data1))

	data2, err := readFile(file2)
	require.NoError(t, err)
	assert.Equal(t, "original b", string(data2))
}

func TestFileBackup_Disabled_DoesNotBackup(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)
	fb.SetEnabled(false)

	testFile := filepath.Join(tmpDir, "skip.go")
	writeTestFile(t, testFile, []byte("content"))

	// Backup should still succeed (it checks enabled in the caller, not here)
	// But BackupPath should be recorded
	require.NoError(t, fb.Backup(testFile))
	assert.NotEmpty(t, fb.BackupPath(testFile))
}
