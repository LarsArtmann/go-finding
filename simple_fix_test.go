package finding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplySimpleFixes_SingleFix(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	content := "package main\n\noldCode\n"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	findings := []Finding{
		{
			ID:          "1",
			Rule:        "r",
			ToolName:    "t",
			Message:     "m",
			Severity:    SeverityWarning,
			Position:    Position{File: FilePath(filePath), Line: 3},
			BeforeCode:  "oldCode",
			AfterCode:   "newCode",
			FixStrategy: FixStrategyDirect,
		},
	}

	results := ApplySimpleFixes(findings)

	fileResults := results[FilePath(filePath)]
	if len(fileResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(fileResults))
	}

	if !fileResults[0].Applied {
		t.Errorf("fix not applied: %s", fileResults[0].Reason)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != "package main\n\nnewCode\n" {
		t.Errorf("file content = %q, want %q", string(data), "package main\n\nnewCode\n")
	}
}

func TestApplySimpleFixes_MultipleSameFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	content := "foo\nbar\nfoo\n"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	findings := []Finding{
		{
			ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityWarning,
			Position: Position{File: FilePath(filePath), Line: 1},
			BeforeCode: "foo", AfterCode: "baz", FixStrategy: FixStrategyDirect,
		},
		{
			ID: "2", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityWarning,
			Position: Position{File: FilePath(filePath), Line: 2},
			BeforeCode: "bar", AfterCode: "qux", FixStrategy: FixStrategyDirect,
		},
	}

	results := ApplySimpleFixes(findings)

	fileResults := results[FilePath(filePath)]
	if len(fileResults) != 2 {
		t.Fatalf("expected 2 results, got %d", len(fileResults))
	}

	for i, r := range fileResults {
		if !r.Applied {
			t.Errorf("fix[%d] not applied: %s", i, r.Reason)
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	expected := "baz\nqux\nfoo\n"
	if string(data) != expected {
		t.Errorf("file content = %q, want %q", string(data), expected)
	}
}

func TestApplySimpleFixes_NoBeforeCode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	content := "unchanged\n"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	findings := []Finding{
		{
			ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityWarning,
			Position: Position{File: FilePath(filePath), Line: 1},
			AfterCode: "newCode", FixStrategy: FixStrategyDirect,
		},
	}

	results := ApplySimpleFixes(findings)

	fileResults := results[FilePath(filePath)]
	if len(fileResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(fileResults))
	}

	if fileResults[0].Applied {
		t.Error("fix should not be applied without BeforeCode")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != content {
		t.Errorf("file should be unchanged")
	}
}

func TestApplySimpleFixes_NoMatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	content := "hello world\n"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	findings := []Finding{
		{
			ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityWarning,
			Position: Position{File: FilePath(filePath), Line: 1},
			BeforeCode: "nonexistent", AfterCode: "replacement", FixStrategy: FixStrategyDirect,
		},
	}

	results := ApplySimpleFixes(findings)

	fileResults := results[FilePath(filePath)]
	if fileResults[0].Applied {
		t.Error("fix should not be applied when BeforeCode not found")
	}

	if fileResults[0].Reason == "" {
		t.Error("Reason should explain why fix was skipped")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != content {
		t.Errorf("file should be unchanged")
	}
}

func TestApplySimpleFixes_FileNotFound(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityWarning,
			Position: Position{File: FilePath("/nonexistent/path/file.go"), Line: 1},
			BeforeCode: "old", AfterCode: "new", FixStrategy: FixStrategyDirect,
		},
	}

	results := ApplySimpleFixes(findings)

	fileResults := results[FilePath("/nonexistent/path/file.go")]
	if len(fileResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(fileResults))
	}

	if fileResults[0].Applied {
		t.Error("fix should not be applied for nonexistent file")
	}

	if fileResults[0].Reason == "" {
		t.Error("Reason should explain why fix was skipped")
	}
}
