package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestFixApplier_Apply_CancelledContext(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "cancel.go")
	if err := writeFile(testFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fixes := []finding.Finding{
		makeFixFinding("1", "package main", "package main // fixed", "cancel.go", 0),
	}

	applied, err := applier.Apply(ctx, fixes)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0", applied)
	}
}

func TestFixApplier_Backup_NonexistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.backup(filepath.Join(tempDir, "does-not-exist.go"))
	if err == nil {
		t.Fatal("expected error backing up nonexistent file")
	}

	if !errors.Is(err, finding.ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}
}

func TestFixApplier_Restore_WithoutBackup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	err := applier.restore(filepath.Join(tempDir, "never-backed-up.go"))
	if err == nil {
		t.Fatal("expected error restoring without backup")
	}

	if !errors.Is(err, finding.ErrInternal) {
		t.Errorf("error = %v, want ErrInternal", err)
	}
}

func TestFixApplier_ApplyToFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	fixes := []finding.Finding{
		{
			ID:         "1",
			BeforeCode: "old",
			AfterCode:  "new",
		},
	}

	applied, err := applier.applyToFile(filepath.Join(tempDir, "missing.go"), fixes)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}

	if !errors.Is(err, finding.ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0", applied)
	}
}

func TestFixApplier_ApplyToFile_NoMatchingBeforeCode(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "nomatch.go")
	if err := writeFile(testFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{
		{
			ID:         "1",
			BeforeCode: "nonexistent_code",
			AfterCode:  "replacement",
		},
	}

	applied, err := applier.applyToFile(testFile, fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0 (no match)", applied)
	}
}

func TestFixApplier_ApplyToFile_ReadOnlyFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "readonly.go")
	if err := writeFile(testFile, []byte("package main\nold()\n"), 0o444); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{
		{
			ID:         "1",
			BeforeCode: "old()",
			AfterCode:  "new()",
		},
	}

	applied, err := applier.applyToFile(testFile, fixes)
	if err == nil {
		t.Fatal("expected error writing to read-only file")
	}

	if !errors.Is(err, finding.ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}

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

	if err := writeFile(file1, []byte(orig1), 0o644); err != nil {
		t.Fatalf("create a.go: %v", err)
	}

	if err := writeFile(file2, []byte(orig2), 0o644); err != nil {
		t.Fatalf("create b.go: %v", err)
	}

	if err := applier.backup(file1); err != nil {
		t.Fatalf("backup a.go: %v", err)
	}

	if err := applier.backup(file2); err != nil {
		t.Fatalf("backup b.go: %v", err)
	}

	if err := writeFile(file1, []byte("modified a\n"), 0o644); err != nil {
		t.Fatalf("modify a.go: %v", err)
	}

	if err := writeFile(file2, []byte("modified b\n"), 0o644); err != nil {
		t.Fatalf("modify b.go: %v", err)
	}

	if err := applier.rollbackAll([]string{file1, file2}); err != nil {
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
	if err := writeFile(goodFile, []byte("package good\n"), 0o644); err != nil {
		t.Fatalf("create good.go: %v", err)
	}

	if err := applier.backup(goodFile); err != nil {
		t.Fatalf("backup good.go: %v", err)
	}

	if err := writeFile(goodFile, []byte("modified\n"), 0o644); err != nil {
		t.Fatalf("modify good.go: %v", err)
	}

	noBackupFile := filepath.Join(tempDir, "nobackup.go")
	err := applier.rollbackAll([]string{goodFile, noBackupFile})
	if err == nil {
		t.Fatal("expected partial rollback error")
	}

	data, rErr := readFile(goodFile)
	if rErr != nil {
		t.Fatalf("read good.go: %v", rErr)
	}

	if string(data) != "package good\n" {
		t.Errorf("good.go not restored: %q", string(data))
	}
}

func TestFixApplier_Apply_BackupFailureRollsBack(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	file1 := filepath.Join(tempDir, "first.go")
	if err := writeFile(file1, []byte("package first\nold1()\n"), 0o644); err != nil {
		t.Fatalf("create first.go: %v", err)
	}

	fixes := []finding.Finding{
		{
			ID:          "1",
			BeforeCode:  "old1()",
			AfterCode:   "new1()",
			Position:    finding.Position{File: "first.go"},
			FixStrategy: finding.FixStrategyDirect,
		},
		{
			ID:          "2",
			BeforeCode:  "old2()",
			AfterCode:   "new2()",
			Position:    finding.Position{File: "nonexistent_dir/second.go"},
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err == nil {
		t.Fatal("expected error from backup failure")
	}

	if !errors.Is(err, finding.ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}

	_ = applied
}

func TestFixApplier_Apply_EmptyFixesList(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	applied, err := applier.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0", applied)
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0 (no files specified)", applied)
	}
}

func TestFixApplier_ApplyToFile_RangeOutOfBounds(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "outofbounds.go")
	content := "package main\n\nfunc main() {}\n"
	if err := writeFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{
		{
			ID:         "1",
			BeforeCode: "nonexistent",
			AfterCode:  "replacement",
			Range:      finding.NewRangePtr("outofbounds.go", 100, 1, 200, 1),
		},
	}

	applied, err := applier.applyToFile(testFile, fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0 (range out of bounds)", applied)
	}

	data, rErr := readFile(testFile)
	if rErr != nil {
		t.Fatalf("read: %v", rErr)
	}

	if string(data) != content {
		t.Errorf("file modified despite out-of-bounds range: %q", string(data))
	}
}

func TestFixApplier_FileHash_Deterministic(t *testing.T) {
	t.Parallel()

	h1 := fileHash("test/path.go")
	h2 := fileHash("test/path.go")
	if h1 != h2 {
		t.Errorf("fileHash not deterministic: %q != %q", h1, h2)
	}

	h3 := fileHash("different/path.go")
	if h1 == h3 {
		t.Error("fileHash collision for different paths")
	}
}

func TestFixApplier_BackupDisabled(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)
	applier.backupEnabled = false

	testFile := filepath.Join(tempDir, "nobackup.go")
	if err := writeFile(testFile, []byte("package main\nold()\n"), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	fixes := []finding.Finding{
		makeFixFinding("1", "old()", "new()", "nobackup.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}

	if len(applier.backups) != 0 {
		t.Errorf("backups = %d, want 0 (backup disabled)", len(applier.backups))
	}
}

func TestFixApplier_NewFixApplier_Defaults(t *testing.T) {
	t.Parallel()

	applier := NewFixApplier("/tmp/test")

	if applier.rootDir != "/tmp/test" {
		t.Errorf("rootDir = %q, want /tmp/test", applier.rootDir)
	}

	if !applier.backupEnabled {
		t.Error("backupEnabled = false, want true")
	}

	if applier.backups == nil {
		t.Error("backups map is nil")
	}
}

func TestFixApplier_Apply_RestoreOnApplyError(t *testing.T) {
	t.Parallel()

	applier := NewFixApplier(t.TempDir())
	testBackupRestore(t, applier,
		"package main\nold()\n",
		"package main\nnew()\n")
}

func TestIoErrorAt_WrapsCorrectly(t *testing.T) {
	t.Parallel()

	err := ioErrorAt("test op", os.ErrPermission, "file.go")

	if !errors.Is(err, finding.ErrIO) {
		t.Error("expected ErrIO sentinel match")
	}

	if !errors.Is(err, os.ErrPermission) {
		t.Error("expected os.ErrPermission cause match")
	}

	var fe *finding.FindingError
	if !errors.As(err, &fe) {
		t.Fatal("expected FindingError")
	}

	if fe.Category != finding.ErrCategoryIO {
		t.Errorf("category = %q, want %q", fe.Category, finding.ErrCategoryIO)
	}

	if fe.Position.File != "file.go" {
		t.Errorf("file = %q, want %q", fe.Position.File, "file.go")
	}
}
