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

	tempDir, applier := newTestApplierWithDir(t)

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

	tempDir, applier := newTestApplierWithDir(t)

	err := applier.backup.Backup(filepath.Join(tempDir, "does-not-exist.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
}

func TestFixApplier_Restore_WithoutBackup(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	err := applier.backup.Restore(filepath.Join(tempDir, "never-backed-up.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrInternal)).To(BeTrue())
}

func TestFixApplier_ApplyToFile_PreservesPermissions(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "executable.go")
	original := []byte("#!/usr/bin/env go run\npackage main\nold()\n")
	g.Expect(writeFile(testFile, original, 0o755)).To(Succeed())

	fixes := []finding.Finding{makeFixFinding("1", "old()", "new()", "executable.go", 0)}

	applied, err := applier.Apply(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))

	info, statErr := os.Stat(testFile)
	g.Expect(statErr).NotTo(HaveOccurred())
	g.Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))

	data, readErr := readFile(testFile)
	g.Expect(readErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(Equal("#!/usr/bin/env go run\npackage main\nnew()\n"))
}

func TestFixApplier_ApplyToFile_NonexistentFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	fixes := []finding.Finding{makeFixFinding("1", "old", "new", "", 0)}

	applied, _, err := applier.applyToFile(filepath.Join(tempDir, "missing.go"), fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_ApplyToFile_NoMatchingBeforeCode(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "nomatch.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{makeFixFinding("1", "nonexistent_code", "replacement", "", 0)}

	applied, _, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_PathTraversal_Skipped(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "safe.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{
		makeFixFinding("1", "package main", "package main // fixed", "safe.go", 0),
		{
			ID:          "traversal",
			Rule:        "r",
			ToolName:    "t",
			Message:     "path traversal",
			Severity:    finding.SeverityInfo,
			Position:    finding.Position{File: "../../../etc/passwd"},
			BeforeCode:  "root:",
			AfterCode:   "pwned:",
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))
}

func TestFixApplier_ApplyToFile_ReadOnlyFile(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "readonly.go")

	err := writeFile(testFile, []byte("package main\nold()\n"), 0o444)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{makeFixFinding("1", "old()", "new()", "", 0)}

	applied, _, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())

	_ = applied
}

func TestFixApplier_RollbackAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

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

	err := applier.backup.RollbackAll([]string{file1, file2})
	if err != nil {
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

	tempDir, applier := newTestApplierWithDir(t)

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

	tempDir, applier := newTestApplierWithDir(t)

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

	applier := newTestApplier(t)

	applied, err := applier.Apply(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(0))
}

func TestFixApplier_Apply_ApplyToFileErrorRestoresAndRollsBack(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

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

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{File: ""}},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(0))
}

// TestFixApplier_ApplyWithDetails_ReturnsOnlyApplied is a regression test for
// a bug where ApplyWithDetails returned the input fixes slice instead of the
// actually-applied findings. Findings without a file are silently skipped by
// groupFindingsBySafePath, so they must NOT appear in the returned slice.
func TestFixApplier_ApplyWithDetails_ReturnsOnlyApplied(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir, applier := newTestApplierWithDir(t)

	writeTestFile(t, filepath.Join(tempDir, "real.go"), []byte("package real\nold()\n"))

	fixes := []finding.Finding{
		makeFixFinding("applies", "old()", "new()", "real.go", 0),
		{
			ID:         "skipped",
			BeforeCode: "old",
			AfterCode:  "new",
			Position:   finding.Position{File: ""},
		}, // no file -> skipped
	}

	count, appliedFixes, err := applier.ApplyWithDetails(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(count).To(Equal(1))
	g.Expect(appliedFixes).To(HaveLen(1))
	g.Expect(string(appliedFixes[0].ID)).To(Equal("applies"))
}

func TestFixApplier_ApplyToFile_RangeOutOfBounds(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tempDir, applier := newTestApplierWithDir(t)

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

	applied, _, err := applier.applyToFile(testFile, fixes)
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

	tempDir, applier := newTestApplierWithDir(t)
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

	applier, err := NewFixApplier("/tmp/test")
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	g.Expect(applier.rootDir).To(Equal("/tmp/test"))
	g.Expect(applier.backup.IsEnabled()).To(BeTrue())
	g.Expect(applier.backup.backupDir).NotTo(BeEmpty())
}

func TestNewFixApplier_MkdirTempError(t *testing.T) {
	g := NewWithT(t)
	t.Setenv("TMPDIR", "/etc/passwd")

	_, err := NewFixApplier("/tmp/test")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("create backup directory"))
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

// TestFixApplier_InsertionOnly verifies C-2: findings with only AfterCode
// (and no BeforeCode) are treated as insertions at the finding's line.
func TestFixApplier_InsertionOnly(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	g.Expect(err).NotTo(HaveOccurred())

	want := helloProgram
	g.Expect(string(got)).To(Equal(want))
}

// TestFixApplier_DeletionOnly verifies C-2: findings with only BeforeCode
// (and no AfterCode) are treated as deletions.
func TestFixApplier_DeletionOnly(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	g.Expect(err).NotTo(HaveOccurred())

	// Deletion removes the BeforeCode substring but leaves the trailing newline,
	// resulting in a blank line. To delete the entire line including newline,
	// include "\n" in BeforeCode.
	want := "package main\n\nfunc main() {\n\n}\n"
	g.Expect(string(got)).To(Equal(want))
}

// TestFixApplier_NearestLineReplacement verifies H-7: when BeforeCode
// appears multiple times, the occurrence nearest to the finding's line is replaced.
func TestFixApplier_NearestLineReplacement(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	applied, err := applier.Apply(context.Background(), []finding.Finding{fix})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	g.Expect(err).NotTo(HaveOccurred())

	// Only the second "old()" at line 6 should be replaced.
	want := "package main\n\nfunc main() {\n\told()\n\tother()\n\tnew()\n}\n"
	g.Expect(string(got)).To(Equal(want))
}

func TestFixApplier_ApplyWithDetails(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	applier, err := NewFixApplier(t.TempDir())
	g.Expect(err).NotTo(HaveOccurred())

	count, fixes, err := applier.ApplyWithDetails(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(count).To(Equal(0))
	g.Expect(fixes).To(BeEmpty())
}
