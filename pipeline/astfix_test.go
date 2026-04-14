package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestNewASTFixer(t *testing.T) {
	t.Parallel()

	fixer := NewASTFixer()
	if fixer == nil {
		t.Error("NewASTFixer returned nil")
	}
}

func TestASTFixer_Apply(t *testing.T) {
	t.Parallel()

	t.Run("successful text replacement", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "test.go")
		content := "package main\n\nfunc main() {\n\toldCode()\n}\n"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		fixer := NewASTFixer()
		fix := finding.Finding{
			BeforeCode: "oldCode",
			AfterCode:  "newCode",
		}

		result := fixer.Apply(context.Background(), file, fix)

		if !result.Applied {
			t.Errorf("Applied = false, want true; Error: %v", result.Error)
		}
		if result.Method != "text" {
			t.Errorf("Method = %q, want %q", result.Method, "text")
		}

		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		if string(got) != "package main\n\nfunc main() {\n\tnewCode()\n}\n" {
			t.Errorf("file content = %q, unexpected", string(got))
		}
	})

	t.Run("no before/after code returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "test.go")
		if err := os.WriteFile(file, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		fixer := NewASTFixer()
		result := fixer.Apply(context.Background(), file, finding.Finding{})

		if result.Applied {
			t.Error("should not apply without before/after code")
		}
		if result.Error == nil {
			t.Error("expected error")
		}
	})

	t.Run("before code not found returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "test.go")
		if err := os.WriteFile(file, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		fixer := NewASTFixer()
		fix := finding.Finding{
			BeforeCode: "nonexistent",
			AfterCode:  "replacement",
		}

		result := fixer.Apply(context.Background(), file, fix)

		if result.Applied {
			t.Error("should not apply when code not found")
		}
		if result.Error == nil {
			t.Error("expected error")
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		t.Parallel()

		fixer := NewASTFixer()
		fix := finding.Finding{
			BeforeCode: "old",
			AfterCode:  "new",
		}

		result := fixer.Apply(context.Background(), "/nonexistent/file.go", fix)

		if result.Applied {
			t.Error("should not apply for non-existent file")
		}
		if result.Error == nil {
			t.Error("expected error")
		}
	})
}

func TestASTFixer_VerifySyntax(t *testing.T) {
	t.Parallel()

	t.Run("valid Go file returns nil", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "valid.go")
		if err := os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		fixer := NewASTFixer()
		if err := fixer.VerifySyntax(file); err != nil {
			t.Errorf("VerifySyntax returned error: %v", err)
		}
	})

	t.Run("invalid Go file returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "invalid.go")
		if err := os.WriteFile(file, []byte("this is not valid go {{{"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		fixer := NewASTFixer()
		if err := fixer.VerifySyntax(file); err == nil {
			t.Error("expected error for invalid Go")
		}
	})
}

func TestASTFixer_CanApplyASTFix(t *testing.T) {
	t.Parallel()

	t.Run("valid file with before/after returns true", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "test.go")
		if err := os.WriteFile(file, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		fixer := NewASTFixer()
		fix := finding.Finding{BeforeCode: "old", AfterCode: "new"}

		if !fixer.CanApplyASTFix(file, fix) {
			t.Error("CanApplyASTFix = false, want true")
		}
	})

	t.Run("no before/after returns false", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		file := filepath.Join(dir, "test.go")
		if err := os.WriteFile(file, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		fixer := NewASTFixer()
		if fixer.CanApplyASTFix(file, finding.Finding{}) {
			t.Error("CanApplyASTFix = true, want false")
		}
	})

	t.Run("invalid file returns false", func(t *testing.T) {
		t.Parallel()

		fixer := NewASTFixer()
		if fixer.CanApplyASTFix("/nonexistent/file.go", finding.Finding{BeforeCode: "a", AfterCode: "b"}) {
			t.Error("CanApplyASTFix = true for nonexistent file, want false")
		}
	})
}

func TestCalculateStats(t *testing.T) {
	t.Parallel()

	results := []ApplyResult{
		{Applied: true, Method: "text"},
		{Applied: true, Method: "text"},
		{Applied: false, Method: "text", Error: os.ErrNotExist},
		{Applied: false, Method: "skipped"},
	}

	stats := CalculateStats(results)

	if stats.TextApplied != 2 {
		t.Errorf("TextApplied = %d, want 2", stats.TextApplied)
	}
	if stats.Failed != 1 {
		t.Errorf("Failed = %d, want 1", stats.Failed)
	}
	if stats.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", stats.Skipped)
	}
	if stats.Total() != 4 {
		t.Errorf("Total() = %d, want 4", stats.Total())
	}
}
