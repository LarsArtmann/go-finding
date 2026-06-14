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
