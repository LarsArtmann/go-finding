package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestNewFileBackup_Defaults(t *testing.T) {
	g := NewParallelGomega(t)

	fb := NewFileBackup(t.TempDir())
	g.Expect(fb.IsEnabled()).To(BeTrue())
	g.Expect(fb.backupDir).NotTo(BeEmpty())
}

func TestFileBackup_SetEnabled(t *testing.T) {
	g := NewParallelGomega(t)

	fb := NewFileBackup(t.TempDir())
	g.Expect(fb.IsEnabled()).To(BeTrue())

	fb.SetEnabled(false)
	g.Expect(fb.IsEnabled()).To(BeFalse())

	fb.SetEnabled(true)
	g.Expect(fb.IsEnabled()).To(BeTrue())
}

func TestFileBackup_BackupPath_NoBackup(t *testing.T) {
	g := NewParallelGomega(t)

	fb := NewFileBackup(t.TempDir())
	g.Expect(fb.BackupPath("/nonexistent/file.go")).To(BeEmpty())
}

func TestFileBackup_BackupAndRestore(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	writeTestFile(t, testFile, []byte("original content"))

	g.Expect(fb.Backup(testFile)).NotTo(HaveOccurred())

	backupPath := fb.BackupPath(testFile)
	g.Expect(backupPath).NotTo(BeEmpty())

	writeTestFile(t, testFile, []byte("modified content"))

	g.Expect(fb.Restore(testFile)).NotTo(HaveOccurred())

	data, err := readFile(testFile)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data)).To(Equal("original content"))
}

func TestFileBackup_Backup_NonexistentFile(t *testing.T) {
	g := NewParallelGomega(t)

	fb := NewFileBackup(t.TempDir())
	err := fb.Backup(filepath.Join(t.TempDir(), "missing.go"))
	g.Expect(err).To(HaveOccurred())
}

func TestFileBackup_Restore_NoBackup(t *testing.T) {
	g := NewParallelGomega(t)

	fb := NewFileBackup(t.TempDir())
	err := fb.Restore(filepath.Join(t.TempDir(), "never-backed.go"))
	g.Expect(err).To(HaveOccurred())
}

func TestFileBackup_RollbackAll(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)

	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")

	writeTestFile(t, file1, []byte("original a"))
	writeTestFile(t, file2, []byte("original b"))

	g.Expect(fb.Backup(file1)).NotTo(HaveOccurred())
	g.Expect(fb.Backup(file2)).NotTo(HaveOccurred())

	writeTestFile(t, file1, []byte("modified a"))
	writeTestFile(t, file2, []byte("modified b"))

	g.Expect(fb.RollbackAll([]string{file1, file2})).NotTo(HaveOccurred())

	data1, err := readFile(file1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data1)).To(Equal("original a"))

	data2, err := readFile(file2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data2)).To(Equal("original b"))
}

func TestFileBackup_Disabled_DoesNotBackup(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)
	fb.SetEnabled(false)

	testFile := filepath.Join(tmpDir, "skip.go")
	writeTestFile(t, testFile, []byte("content"))

	// Backup should still succeed (it checks enabled in the caller, not here)
	// But BackupPath should be recorded
	g.Expect(fb.Backup(testFile)).NotTo(HaveOccurred())
	g.Expect(fb.BackupPath(testFile)).NotTo(BeEmpty())
}

func TestFileBackup_Backup_MkdirAllError(t *testing.T) {
	g := NewParallelGomega(t)

	// Create a file (not a directory) to use as backupDir parent.
	// MkdirAll will fail because it cannot create a directory inside a file.
	badDir := filepath.Join(t.TempDir(), "not-a-dir")
	writeTestFile(t, badDir, []byte("I am a file"))

	fb := NewFileBackup(badDir)
	testFile := filepath.Join(t.TempDir(), "test.go")
	writeTestFile(t, testFile, []byte("content"))

	err := fb.Backup(testFile)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
}

func TestFileBackup_Restore_ReadBackupError(t *testing.T) {
	g := NewParallelGomega(t)

	tmpDir := t.TempDir()
	fb := NewFileBackup(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	writeTestFile(t, testFile, []byte("original"))

	g.Expect(fb.Backup(testFile)).NotTo(HaveOccurred())

	// Delete the backup file to trigger read error on restore.
	backupPath := fb.BackupPath(testFile)
	g.Expect(os.Remove(backupPath)).NotTo(HaveOccurred())

	err := fb.Restore(testFile)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
}
