package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestByteLevelConflictDetection_ConflictingEdits(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := "fixme.go"
	writeTestFile(t, filepath.Join(tmpDir, testFile),
		[]byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"))

	fix1 := makeFixFinding("f1", "hello", "world", testFile, 4)
	fix2 := makeFixFinding("f2", "hello", "universe", testFile, 4)

	cfg := Config{
		MaxIterations:              1,
		ByteLevelConflictDetection: true,
		OnFix: func(f finding.Finding, applied bool) {
			if f.ID == "f1" && !applied {
				t.Errorf("f1 should have been applied")
			}
		},
	}

	det := mockDetWithFindings("tool", fix1, fix2)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	if result.Iterations[0].Conflicts != 1 {
		t.Errorf("expected 1 conflict from overlapping edits, got %d", result.Iterations[0].Conflicts)
	}

	if result.Iterations[0].Applied != 1 {
		t.Errorf("expected 1 fix applied, got %d", result.Iterations[0].Applied)
	}
}

func TestByteLevelConflictDetection_NonConflictingEdits(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := "fixme.go"
	writeTestFile(t, filepath.Join(tmpDir, testFile),
		[]byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"))

	fix1 := makeFixFinding("f1", "hello", "world", testFile, 4)
	fix2 := makeFixFinding("f2", "main", "myApp", testFile, 3)

	cfg := Config{
		MaxIterations:              1,
		ByteLevelConflictDetection: true,
	}

	det := mockDetWithFindings("tool", fix1, fix2)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	if result.Iterations[0].Conflicts != 0 {
		t.Errorf("expected no conflicts, got %d", result.Iterations[0].Conflicts)
	}

	if result.Iterations[0].Applied != 2 {
		t.Errorf("expected 2 applied, got %d", result.Iterations[0].Applied)
	}
}

func TestByteLevelConflictDetection_MissingFile_FailOpen(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	fix1 := makeFixFinding("f1", "hello", "world", "nonexistent.go", 4)

	cfg := Config{
		MaxIterations:              1,
		ByteLevelConflictDetection: true,
	}

	det := mockDetWithFindings("tool", fix1)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = p.Run(context.Background())
	if err == nil {
		t.Fatal("expected error when applying fix to nonexistent file")
	}
}
