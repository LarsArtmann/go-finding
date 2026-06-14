package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFixApplier(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")

	testContent := helloProgram

	err = writeFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test backup
	err = applier.backup.Backup(testFile)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup exists
	backupPath := applier.backup.BackupPath(testFile)
	g.Expect(backupPath).NotTo(BeEmpty())

	// Test restore
	err = applier.backup.Restore(testFile)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Test Apply with empty fixes
	applied, err := applier.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("apply with nil fixes failed: %v", err)
	}

	g.Expect(applied).To(Equal(0))

	// Test Apply with fixes that have no Position.File
	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{}},
	}

	applied, err = applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(0))
}

// TestFixApplier_Apply tests actual fix application.
func TestFixApplier_Apply(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")

	originalContent := helloProgram
	writeTestFile(t, testFile, []byte(originalContent))

	// Create fix
	fixes := []finding.Finding{
		makeFixFinding("1", `println("hello")`, `println("world")`, "test.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	// Verify content changed
	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tprintln(\"world\")\n}\n"
	g.Expect(string(content)).To(Equal(expected))
}

func TestFixApplier_RangeBasedFix(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tprintln(\"hello\")\n}\n" //nolint:dupword // test fixture intentionally duplicates println

	writeErr := writeFile(testFile, []byte(content), 0o644)
	if writeErr != nil {
		t.Fatalf("create test file: %v", writeErr)
	}

	// Fix only the second "println" at line 5, using Range for precision.
	// Without Range, strings.Replace would hit the first occurrence at line 4.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			BeforeCode:  "println(\"hello\")",
			AfterCode:   "fmt.Println(\"world\")",
			Position:    finding.Pos("test.go", 5, 2),
			Range:       finding.NewRangePtr("test.go", 5, 2, 5, 18),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// First println unchanged, second replaced.
	want := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tfmt.Println(\"world\")\n}\n"
	g.Expect(string(got)).To(Equal(want))
}

func TestFixApplier_MultiLineRangeFix(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc old() {\n\treturn\n}\n\nfunc main() {}\n"
	writeTestFile(t, testFile, []byte(content))

	// Replace lines 3-5 (func old) with new content.
	fixes := []finding.Finding{
		{
			BeforeCode:  "func old() {\n\treturn\n}",
			AfterCode:   "func new() {\n\treturn 42\n}",
			Position:    finding.Position{File: "test.go", Line: 3},
			Range:       finding.NewRangePtr("test.go", 3, 1, 5, 2),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}\n"
	g.Expect(string(got)).To(Equal(want))
}

func TestFixApplier_RangeFixEmptyBeforeCode(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()

	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc old() {\n\treturn\n}\n\nfunc main() {}\n"
	writeTestFile(t, testFile, []byte(content))

	// Replace lines 3-5 with new content using only Range + AfterCode (no BeforeCode).
	fixes := []finding.Finding{
		{
			AfterCode:   "func replaced() { // inserted\n}",
			Position:    finding.Position{File: "test.go", Line: 3},
			Range:       finding.NewRangePtr("test.go", 3, 1, 5, 2),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc replaced() { // inserted\n}\n\nfunc main() {}\n"
	g.Expect(string(got)).To(Equal(want))
}

// BenchmarkParallelDetection benchmarks parallel vs sequential detection.
