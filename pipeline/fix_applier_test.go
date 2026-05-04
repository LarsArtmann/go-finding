package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFixApplier_Apply_CancelledContext(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
	g.Expect(applied).To(Equal(0))
}

func TestFixApplier_Backup_NonexistentFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.backup.Backup(filepath.Join(tempDir, "does-not-exist.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
}

func TestFixApplier_Restore_WithoutBackup(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.backup.Restore(filepath.Join(tempDir, "never-backed-up.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrInternal)).To(BeTrue())
}

func TestFixApplier_ApplyToFile_NonexistentFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	fixes := []finding.Finding{makeFixFinding("1", "old", "new", "", 0)}

	applied, err := applier.applyToFile(filepath.Join(tempDir, "missing.go"), fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_ApplyToFile_NoMatchingBeforeCode(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "nomatch.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{makeFixFinding("1", "nonexistent_code", "replacement", "", 0)}

	applied, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_ApplyToFile_ReadOnlyFile(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "readonly.go")
	if err := writeFile(testFile, []byte("package main\nold()\n"), 0o444); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{makeFixFinding("1", "old()", "new()", "", 0)}

	applied, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())

	_ = applied
}

func TestFixApplier_RollbackAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	file1 := filepath.Join(tempDir, "a.go")
	file2 := filepath.Join(tempDir, "b.go")

	orig1 := "package a\n"
	orig2 := "package b\n"

	writeTestFile(t, file1, []byte(orig1))

	writeTestFile(t, file2, []byte(orig2))

	g.Expect(applier.backup.Backup(file1)).NotTo(HaveOccurred())
	g.Expect(applier.backup.Backup(file2)).NotTo(HaveOccurred())

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
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	goodFile := filepath.Join(tempDir, "good.go")
	writeTestFile(t, goodFile, []byte("package good\n"))

	g.Expect(applier.backup.Backup(goodFile)).NotTo(HaveOccurred())

	writeTestFile(t, goodFile, []byte("modified\n"))

	noBackupFile := filepath.Join(tempDir, "nobackup.go")
	err := applier.backup.RollbackAll([]string{goodFile, noBackupFile})
	g.Expect(err).To(HaveOccurred())

	data, rErr := readFile(goodFile)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(Equal("package good\n"))
}

func TestFixApplier_Apply_BackupFailureRollsBack(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())

	_ = applied
}

func TestFixApplier_Apply_EmptyFixesList(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	applied, err := applier.Apply(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(0))
}

func TestFixApplier_Apply_ApplyToFileErrorRestoresAndRollsBack(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// File 1: will be successfully modified.
	file1 := filepath.Join(tempDir, "first.go")
	writeTestFile(t, file1, []byte("package first\nold1()\n"))

	// File 2: read-only so applyToFile write fails.
	file2 := filepath.Join(tempDir, "second.go")
	writeTestFile(t, file2, []byte("package second\nold2()\n"))
	errChmod := os.Chmod(file2, 0o444) //nolint:gosec // intentional read-only for test
	g.Expect(errChmod).NotTo(HaveOccurred())
	t.Cleanup(func() {
		_ = os.Chmod(file2, 0o644) //nolint:gosec // restore permissions in cleanup
	})

	fixes := []finding.Finding{
		makeFixFinding("1", "old1()", "new1()", "first.go", 0),
		makeFixFinding("2", "old2()", "new2()", "second.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrConflict)).To(BeTrue())
	g.Expect(applied).To(Equal(1))

	// File 1 should have been applied then rolled back.
	data1, rErr := readFile(file1)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data1)).To(Equal("package first\nold1()\n"))

	// File 2 should be unchanged (write failed before modification).
	data2, rErr := readFile(file2)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data2)).To(Equal("package second\nold2()\n"))
}

func TestFixApplier_Apply_FixesWithNoFile(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{File: ""}},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(0))
}

func TestFixApplier_ApplyToFile_RangeOutOfBounds(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(BeNil())

	data, rErr := readFile(testFile)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(Equal(content))
}

func TestFixApplier_FileHash_Deterministic(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	h1 := fileHash("test/path.go")
	h2 := fileHash("test/path.go")
	g.Expect(h2).To(Equal(h1))

	h3 := fileHash("different/path.go")
	g.Expect(h3).NotTo(Equal(h1))
}

func TestFixApplier_BackupDisabled(t *testing.T) {
	g := NewWithT(t)
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
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))
	g.Expect(applier.backup.BackupPath("nobackup.go")).To(BeEmpty())
}

func TestFixApplier_NewFixApplier_Defaults(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	applier := NewFixApplier("/tmp/test")

	g.Expect(applier.rootDir).To(Equal("/tmp/test"))
	g.Expect(applier.backup.IsEnabled()).To(BeTrue())
	g.Expect(applier.backup.backupDir).NotTo(BeEmpty())
}

func TestNewFixApplier_MkdirTempFallback(t *testing.T) {
	g := NewWithT(t)
	t.Setenv("TMPDIR", "/etc/passwd")

	applier := NewFixApplier("/tmp/test")
	g.Expect(applier.rootDir).To(Equal("/tmp/test"))
	g.Expect(applier.backup.IsEnabled()).To(BeTrue())
	g.Expect(applier.backup.backupDir).NotTo(BeEmpty())
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
	g := NewWithT(t)

	err := ioErrorAt("test op", os.ErrPermission, "file.go")

	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
	g.Expect(errors.Is(err, os.ErrPermission)).To(BeTrue())

	var fe *finding.FindingError
	g.Expect(errors.As(err, &fe)).To(BeTrue())

	assertFindingErrorIO(t, fe, "file.go")
}
