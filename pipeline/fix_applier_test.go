package pipeline

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFixApplier_Apply_CancelledContext(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	err := applier.backup.Backup(filepath.Join(tempDir, "does-not-exist.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
}

func TestFixApplier_Restore_WithoutBackup(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	err := applier.backup.Restore(filepath.Join(tempDir, "never-backed-up.go"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrInternal)).To(BeTrue())
}

func TestFixApplier_ApplyToFile_PreservesPermissions(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	fixes := []finding.Finding{makeFixFinding("1", "old", "new", "", 0)}

	applied, _, _, err := applier.applyToFile(filepath.Join(tempDir, "missing.go"), fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_ApplyToFile_NoMatchingBeforeCode(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "nomatch.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{makeFixFinding("1", "nonexistent_code", "replacement", "", 0)}

	applied, _, _, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(BeNil())
}

func TestFixApplier_PathTraversal_Skipped(t *testing.T) {
	g := NewParallelGomega(t)

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

	// Since the outcome-surfacing change, the traversal finding is reported
	// as a soft failure (joined validation error) instead of being silently
	// dropped. The safe fix still applies and /etc/passwd is untouched.
	g.Expect(err).To(HaveOccurred())
	g.Expect(applied).To(Equal(1))

	data, readErr := os.ReadFile(testFile)
	g.Expect(readErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(ContainSubstring("// fixed"))
}

func TestFixApplier_ApplyToFile_ReadOnlyFile(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "readonly.go")

	err := writeFile(testFile, []byte("package main\nold()\n"), 0o444)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{makeFixFinding("1", "old()", "new()", "", 0)}

	applied, _, _, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, finding.ErrIO)).To(BeTrue())

	_ = applied
}

func TestFixApplier_RollbackAll(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	applier := newTestApplier(t)

	applied, err := applier.Apply(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(Equal(0))
}

// TestFixApplier_Apply_FileError_DefaultKeepsEarlierFiles verifies the
// default rollback policy: when one file fails, only the failing file is
// restored and files applied earlier keep their fixes (issue #28).
func TestFixApplier_Apply_FileError_DefaultKeepsEarlierFiles(t *testing.T) {
	g := NewParallelGomega(t)

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

	// Default policy (RollbackPolicyFailingFile): file 1 keeps its applied fix.
	data1, rErr := readFile(file1)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data1)).To(Equal("package first\nnew1()\n"))

	// File 2 should be unchanged (write failed, restored from backup).
	data2, rErr := readFile(file2)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data2)).To(Equal("package second\nold2()\n"))
}

// TestFixApplier_Apply_FileError_RollbackPolicyAllFiles verifies the
// all-or-nothing contract: when a file fails, files modified earlier in the
// run are rolled back to their backups.
func TestFixApplier_Apply_FileError_RollbackPolicyAllFiles(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)
	applier.SetRollbackPolicy(RollbackPolicyAllFiles)

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

// TestFixApplier_ApplyWithReport_SoftErrorKeepsOtherFiles verifies the
// issue #28 scenario: one unresolvable finding must not discard clean fixes
// in the same file or in other files. The failed finding is reported via
// outcomes and the returned error, while all applied fixes stay on disk.
func TestFixApplier_ApplyWithReport_SoftErrorKeepsOtherFiles(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	file1 := filepath.Join(tempDir, "first.go")
	writeTestFile(t, file1, []byte("package first\nold1()\nold2()\n"))

	file2 := filepath.Join(tempDir, "second.go")
	writeTestFile(t, file2, []byte("package second\nold3()\n"))

	fixes := []finding.Finding{
		makeFixFinding("good-1", "old1()", "new1()", "first.go", 0),
		// Position beyond EOF and BeforeCode not in the file: the line
		// provider errors and the substring provider refuses, so this
		// finding fails resolution.
		makeFixFinding("unresolvable", "no-such-text", "fixed()", "first.go", 999),
		makeFixFinding("good-2", "old3()", "new3()", "second.go", 0),
	}

	report, err := applier.ApplyWithReport(context.Background(), fixes)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("unresolvable"))

	g.Expect(report.Applied).To(Equal(2))
	g.Expect(report.AppliedFixes).To(HaveLen(2))
	g.Expect(report.RolledBack).To(BeEmpty())

	failed := report.FailedOutcomes()
	g.Expect(failed).To(HaveLen(1))
	g.Expect(string(failed[0].Finding.ID)).To(Equal("unresolvable"))
	g.Expect(failed[0].Status).To(Equal(FixOutcomeFailed))

	data1, rErr := readFile(file1)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data1)).To(Equal("package first\nnew1()\nold2()\n"))

	data2, rErr := readFile(file2)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data2)).To(Equal("package second\nnew3()\n"))
}

// TestFixApplier_ApplyWithReport_RefusedFindingsReported verifies that
// findings a provider matched but refused are visible in outcomes without
// failing the run.
func TestFixApplier_ApplyWithReport_RefusedFindingsReported(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	testFile := filepath.Join(tempDir, "only.go")
	writeTestFile(t, testFile, []byte("package only\nold()\n"))

	fixes := []finding.Finding{
		makeFixFinding("good", "old()", "new()", "only.go", 0),
		makeFixFinding("refused", "no-such-text", "new()", "only.go", 0),
	}

	report, err := applier.ApplyWithReport(context.Background(), fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(report.Applied).To(Equal(1))

	g.Expect(report.Outcomes).To(HaveLen(2))

	byID := map[string]FixOutcome{}
	for _, o := range report.Outcomes {
		byID[string(o.Finding.ID)] = o
	}

	g.Expect(byID["good"].Status).To(Equal(FixOutcomeApplied))
	g.Expect(byID["refused"].Status).To(Equal(FixOutcomeRefused))
	g.Expect(byID["refused"].Err).NotTo(HaveOccurred())
}

func TestFixApplier_Apply_FixesWithNoFile(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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

	applied, _, _, err := applier.applyToFile(testFile, fixes)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(applied).To(BeNil())

	data, rErr := readFile(testFile)
	g.Expect(rErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(Equal(content))
}

func TestFixApplier_FileHash_Deterministic(t *testing.T) {
	g := NewParallelGomega(t)

	h1 := fileHash("test/path.go")
	h2 := fileHash("test/path.go")
	g.Expect(h2).To(Equal(h1))

	h3 := fileHash("different/path.go")
	g.Expect(h3).NotTo(Equal(h1))
}

func TestFixApplier_BackupDisabled(t *testing.T) {
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	applier, err := NewFixApplier(t.TempDir())
	g.Expect(err).NotTo(HaveOccurred())

	count, fixes, err := applier.ApplyWithDetails(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(count).To(Equal(0))
	g.Expect(fixes).To(BeEmpty())
}

// TestFixApplier_RollbackErrorNotSwallowed verifies that when applyToFile
// fails hard (write error on a read-only file) AND Restore fails (because the
// .bak file was deleted), the returned error includes BOTH the apply failure
// and the rollback failure — proving rollback errors are not silently swallowed.
func TestFixApplier_RollbackErrorNotSwallowed(t *testing.T) {
	g := NewParallelGomega(t)

	root := t.TempDir()

	testFile := filepath.Join(root, "target.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	resolvedPath, ok := ResolveSafePath(root, "target.go")
	g.Expect(ok).To(BeTrue())

	backupDir := t.TempDir()
	backup := NewFileBackup(backupDir)

	provider := &saboteurProvider{backup: backup, targetPath: resolvedPath}

	applier := &FixApplier{
		rootDir: root,
		backup:  backup,
		engine:  NewFixEngineWithProviders(provider),
	}

	fix := finding.Finding{
		ID:          "rollback-test",
		BeforeCode:  "package main",
		AfterCode:   "package main // fixed",
		Position:    finding.Position{File: "target.go", Line: 1},
		FixStrategy: finding.FixStrategyDirect,
	}

	errChmod := os.Chmod(testFile, 0o444) //nolint:gosec // intentional read-only for test
	g.Expect(errChmod).NotTo(HaveOccurred())
	t.Cleanup(func() {
		_ = os.Chmod(testFile, 0o644) //nolint:gosec // restore permissions in cleanup
	})

	_, err := applier.Apply(context.Background(), []finding.Finding{fix})
	g.Expect(err).To(HaveOccurred())

	errMsg := err.Error()
	g.Expect(errMsg).To(ContainSubstring("rollback also failed"),
		"error should mention rollback failure: %s", errMsg)
	g.Expect(errMsg).To(ContainSubstring("restore"),
		"error should mention restore failure: %s", errMsg)
}

// TestFixApplier_ApplyWithReport_CancelledContext verifies that a cancelled
// context stops fix application before the next file is touched, no files are
// modified, and no outcomes are fabricated for unprocessed files — under both
// rollback policies.
func TestFixApplier_ApplyWithReport_CancelledContext(t *testing.T) {
	scenarios := []struct {
		name    string
		rollAll bool
	}{{name: "default", rollAll: false}, {name: "AllFiles", rollAll: true}}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			tempDir, applier := newTestApplierWithDir(t)
			if tc.rollAll {
				applier.SetRollbackPolicy(RollbackPolicyAllFiles)
			}
			t.Cleanup(func() { _ = applier.Close() })

			fileA := filepath.Join(tempDir, "a.go")
			writeTestFile(t, fileA, []byte("package a\noldA()\n"))

			fileB := filepath.Join(tempDir, "b.go")
			writeTestFile(t, fileB, []byte("package b\noldB()\n"))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			report, err := applier.ApplyWithReport(ctx, []finding.Finding{
				makeFixFinding("1", "oldA()", "newA()", "a.go", 0),
				makeFixFinding("2", "oldB()", "newB()", "b.go", 0),
			})
			g.Expect(err).To(HaveOccurred())
			g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())

			g.Expect(report.Applied).To(Equal(0))
			g.Expect(report.Outcomes).To(BeEmpty())
			g.Expect(report.RolledBack).To(BeEmpty())

			for _, path := range []string{fileA, fileB} {
				data, readErr := readFile(path)
				g.Expect(readErr).NotTo(HaveOccurred())
				g.Expect(data).NotTo(ContainSubstring("new"))
			}
		})
	}
}

// TestFixApplier_ApplyWithReport_BackupFailure verifies the rollback policy on
// backup failures: by default earlier files keep their applied fixes; the
// AllFiles policy rolls them back.
func TestFixApplier_ApplyWithReport_BackupFailure(t *testing.T) {
	scenarios := []struct {
		name      string
		rollAll   bool
		wantFileA string
	}{
		{
			name:      "default keeps earlier file fixes",
			rollAll:   false,
			wantFileA: "package first\nnew1()\n",
		},
		{
			name:      "AllFiles rolls back earlier files",
			rollAll:   true,
			wantFileA: "package first\nold1()\n",
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			tempDir, applier := newTestApplierWithDir(t)
			if tc.rollAll {
				applier.SetRollbackPolicy(RollbackPolicyAllFiles)
			}
			t.Cleanup(func() { _ = applier.Close() })

			fileA := filepath.Join(tempDir, "first.go")
			writeTestFile(t, fileA, []byte("package first\nold1()\n"))

			fileB := filepath.Join(tempDir, "second.go")
			writeTestFile(t, fileB, []byte("package second\nold2()\n"))
			errChmod := os.Chmod(
				fileB,
				0o000,
			) //nolint:gosec // unreadable on purpose: backup open fails
			g.Expect(errChmod).NotTo(HaveOccurred())
			t.Cleanup(func() {
				_ = os.Chmod(fileB, 0o644) //nolint:gosec // restore permissions in cleanup
			})

			report, err := applier.ApplyWithReport(context.Background(), []finding.Finding{
				makeFixFinding("1", "old1()", "new1()", "first.go", 0),
				makeFixFinding("2", "old2()", "new2()", "second.go", 0),
			})
			g.Expect(err).To(HaveOccurred())

			dataA, readErr := readFile(fileA)
			g.Expect(readErr).NotTo(HaveOccurred())
			g.Expect(string(dataA)).To(Equal(tc.wantFileA))

			if tc.rollAll {
				g.Expect(report.RolledBack).To(ContainElement(fileA))
			} else {
				g.Expect(report.RolledBack).To(BeEmpty())
			}
		})
	}
}

// TestFixApplier_BackupHygiene verifies that backups stay inside the backup
// directory (never scatter .bak files into the target tree) and that Close
// removes the backup directory entirely.
func TestFixApplier_BackupHygiene(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)

	fileA := filepath.Join(tempDir, "first.go")
	writeTestFile(t, fileA, []byte("package first\nold1()\n"))

	fileB := filepath.Join(tempDir, "second.go")
	writeTestFile(t, fileB, []byte("package second\nold2()\n"))
	errChmod := os.Chmod(fileB, 0o444) //nolint:gosec // write fails: failing file is restored
	g.Expect(errChmod).NotTo(HaveOccurred())
	t.Cleanup(func() {
		_ = os.Chmod(fileB, 0o644) //nolint:gosec // restore permissions in cleanup
	})

	_, err := applier.ApplyWithReport(context.Background(), []finding.Finding{
		makeFixFinding("1", "old1()", "new1()", "first.go", 0),
		makeFixFinding("2", "old2()", "new2()", "second.go", 0),
	})
	g.Expect(err).To(HaveOccurred())

	backupDir := applier.backup.backupDir

	var stray []string
	walkErr := filepath.WalkDir(tempDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".bak") {
			stray = append(stray, path)
		}

		return nil
	})
	g.Expect(walkErr).NotTo(HaveOccurred())
	g.Expect(stray).To(BeEmpty(), "no .bak files may leak into the target tree")

	g.Expect(backupDir).To(BeADirectory())

	g.Expect(applier.Close()).To(Succeed())
	g.Expect(backupDir).NotTo(BeAnExistingFile())
	g.Expect(backupDir).NotTo(BeADirectory())
}

// cancelingProvider cancels the run's context when it sees a specific
// BeforeCode, and produces valid edits otherwise. It lets tests place the
// cancellation exactly between two files of a multi-file run.
type cancelingProvider struct {
	cancel context.CancelFunc
	before string
}

func (*cancelingProvider) Name() string                       { return "canceling" }
func (p *cancelingProvider) CanHandle(f finding.Finding) bool { return f.HasCodeChange() }
func (p *cancelingProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	if f.BeforeCode == p.before {
		p.cancel()

		return nil, errors.New("cancelled during edit resolution")
	}

	idx := bytes.Index(content, []byte(f.BeforeCode))
	if idx < 0 {
		return nil, nil
	}

	return []FixEdit{newReplacementEdit(idx, len(f.BeforeCode), f)}, nil
}

// TestFixApplier_RolledBackPathsInErrorText verifies that ApplyWithReport
// embeds the exact file paths that were rolled back in the returned error
// text, for all three failure paths that trigger rollback: cancellation,
// backup failure, and hard file failure.
func TestFixApplier_RolledBackPathsInErrorText(t *testing.T) {
	t.Run("cancelled after first file", func(t *testing.T) {
		g := NewParallelGomega(t)

		tempDir, _ := newTestApplierWithDir(t)

		fileA := filepath.Join(tempDir, "a.go")
		writeTestFile(t, fileA, []byte("package a\noldA()\n"))

		fileB := filepath.Join(tempDir, "b.go")
		writeTestFile(t, fileB, []byte("package b\noldB()\n"))

		fileC := filepath.Join(tempDir, "c.go")
		writeTestFile(t, fileC, []byte("package c\noldC()\n"))

		ctx, cancel := context.WithCancel(context.Background())
		applier, err := NewFixApplierWithProviders(tempDir, &cancelingProvider{cancel: cancel, before: "oldB()"})
		g.Expect(err).NotTo(HaveOccurred())
		applier.SetRollbackPolicy(RollbackPolicyAllFiles)
		t.Cleanup(func() { _ = applier.Close() })

		report, err := applier.ApplyWithReport(ctx, []finding.Finding{
			makeFixFinding("1", "oldA()", "newA()", "a.go", 2),
			makeFixFinding("2", "oldB()", "newB()", "b.go", 2),
			makeFixFinding("3", "oldC()", "newC()", "c.go", 2),
		})
		g.Expect(err).To(HaveOccurred())
		g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())

		errMsg := err.Error()
		g.Expect(errMsg).To(ContainSubstring("(rolled back: "),
			"error should list the rolled-back paths: %s", errMsg)
		g.Expect(errMsg).To(ContainSubstring(fileA),
			"error should list the applied-then-rolled-back path: %s", errMsg)
		// b.go was backed up (so restored as a content no-op) even though its
		// finding soft-failed; the note lists every restored path.
		g.Expect(errMsg).To(ContainSubstring(fileB),
			"error should list the backed-up-then-restored path: %s", errMsg)

		g.Expect(report.RolledBack).To(ContainElements(fileA, fileB))

		data, readErr := readFile(fileA)
		g.Expect(readErr).NotTo(HaveOccurred())
		g.Expect(string(data)).To(Equal("package a\noldA()\n"))
	})

	t.Run("backup failure lists earlier files", func(t *testing.T) {
		g := NewParallelGomega(t)

		tempDir, applier := newTestApplierWithDir(t)
		applier.SetRollbackPolicy(RollbackPolicyAllFiles)
		t.Cleanup(func() { _ = applier.Close() })

		fileA := filepath.Join(tempDir, "first.go")
		writeTestFile(t, fileA, []byte("package first\nold1()\n"))

		fileB := filepath.Join(tempDir, "second.go")
		writeTestFile(t, fileB, []byte("package second\nold2()\n"))
		errChmod := os.Chmod(fileB, 0o000) //nolint:gosec // unreadable on purpose: backup open fails
		g.Expect(errChmod).NotTo(HaveOccurred())
		t.Cleanup(func() {
			_ = os.Chmod(fileB, 0o644) //nolint:gosec // restore permissions in cleanup
		})

		report, err := applier.ApplyWithReport(context.Background(), []finding.Finding{
			makeFixFinding("1", "old1()", "new1()", "first.go", 2),
			makeFixFinding("2", "old2()", "new2()", "second.go", 2),
		})
		g.Expect(err).To(HaveOccurred())

		errMsg := err.Error()
		g.Expect(errMsg).To(ContainSubstring("(rolled back: "+fileA+")"),
			"error should list the rolled-back path: %s", errMsg)
		g.Expect(report.RolledBack).To(ContainElement(fileA))
	})

	t.Run("hard file failure lists restored files alongside rollback failure", func(t *testing.T) {
		g := NewParallelGomega(t)

		tempDir, applier := newTestApplierWithDir(t)
		applier.SetRollbackPolicy(RollbackPolicyAllFiles)
		t.Cleanup(func() { _ = applier.Close() })

		fileA := filepath.Join(tempDir, "first.go")
		writeTestFile(t, fileA, []byte("package first\nold1()\n"))

		fileB := filepath.Join(tempDir, "second.go")
		writeTestFile(t, fileB, []byte("package second\nold2()\n"))
		errChmod := os.Chmod(fileB, 0o444) //nolint:gosec // read-only on purpose: backup ok, write fails
		g.Expect(errChmod).NotTo(HaveOccurred())
		t.Cleanup(func() {
			_ = os.Chmod(fileB, 0o644) //nolint:gosec // restore permissions in cleanup
		})

		report, err := applier.ApplyWithReport(context.Background(), []finding.Finding{
			makeFixFinding("1", "old1()", "new1()", "first.go", 2),
			makeFixFinding("2", "old2()", "new2()", "second.go", 2),
		})
		g.Expect(err).To(HaveOccurred())

		errMsg := err.Error()
		g.Expect(errMsg).To(ContainSubstring("rollback also failed"),
			"error should mention the failed restore of the read-only file: %s", errMsg)
		g.Expect(errMsg).To(ContainSubstring("(rolled back: "+fileA+")"),
			"error should list the successfully rolled-back path: %s", errMsg)
		g.Expect(report.RolledBack).To(ContainElement(fileA))
	})
}

// TestFixApplier_ShiftMapMixedOutcomes verifies the per-file line shift map
// reflects exactly the applied edits, even when other findings in the same
// file were refused — refused findings contribute no shift entries.
func TestFixApplier_ShiftMapMixedOutcomes(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)
	t.Cleanup(func() { _ = applier.Close() })

	file := filepath.Join(tempDir, "m.go")
	writeTestFile(t, file, []byte("package m\nold1()\nold2()\n"))

	report, err := applier.ApplyWithReport(context.Background(), []finding.Finding{
		makeFixFinding("1", "old1()", "new1()\nnew1b()", "m.go", 2),
		makeFixFinding("2", "never-present()", "x()", "m.go", 3),
	})
	g.Expect(err).NotTo(HaveOccurred())

	byID := make(map[finding.ID]FixOutcome, len(report.Outcomes))
	for _, o := range report.Outcomes {
		byID[o.Finding.ID] = o
	}

	g.Expect(byID["1"].Status).To(Equal(FixOutcomeApplied))
	g.Expect(byID["2"].Status).To(Equal(FixOutcomeRefused))
	g.Expect(report.Applied).To(Equal(1))

	g.Expect(report.ShiftMaps).To(HaveLen(1))
	shift, ok := report.ShiftMaps["m.go"]
	g.Expect(ok).To(BeTrue())
	g.Expect(shift).NotTo(BeNil())

	// The replacement on line 2 adds one extra line: line 3 moved to 4.
	g.Expect(shift.ShiftedLine(3)).To(Equal(4))
	g.Expect(shift.Entries()).To(HaveLen(1))
}

// TestFixApplier_UnsafePathReportsFailedOutcome verifies that findings whose
// path fails the containment check (traversal outside root) surface as failed
// outcomes with a validation error and a joined error return, instead of being
// silently dropped during grouping.
func TestFixApplier_UnsafePathReportsFailedOutcome(t *testing.T) {
	g := NewParallelGomega(t)

	tempDir, applier := newTestApplierWithDir(t)
	t.Cleanup(func() { _ = applier.Close() })

	testFile := filepath.Join(tempDir, "safe.go")
	writeTestFile(t, testFile, []byte("package main\n"))

	fixes := []finding.Finding{
		makeFixFinding("safe-1", "package main", "package fixed", "safe.go", 0),
		makeFixFinding("evil-1", "a", "b", "../../../etc/passwd", 1),
	}

	report, err := applier.ApplyWithReport(context.Background(), fixes)

	g.Expect(err).To(HaveOccurred(), "unsafe path must surface in the joined error return")
	g.Expect(report.Applied).To(Equal(1), "the safe finding still applies")

	var outcome *FixOutcome
	for i := range report.Outcomes {
		if report.Outcomes[i].Finding.ID == finding.ID("evil-1") {
			outcome = &report.Outcomes[i]
		}
	}

	g.Expect(outcome).NotTo(BeNil(), "unsafe finding gets an outcome, not silence")
	g.Expect(outcome.Status).To(Equal(FixOutcomeFailed))

	var ferr *finding.FindingError
	g.Expect(errors.As(outcome.Err, &ferr)).To(BeTrue())
	g.Expect(ferr.Category).To(Equal(finding.ErrCategoryValidation))
	g.Expect(ferr.Position).NotTo(BeNil())
	g.Expect(string(ferr.Position.File)).To(Equal("../../../etc/passwd"))

	data, readErr := os.ReadFile(testFile)
	g.Expect(readErr).NotTo(HaveOccurred())
	g.Expect(string(data)).To(ContainSubstring("package fixed"))
}
