package goast

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

// beforeAfterFinding builds a Finding with only position + before/after code set.
// Common shape used across the Edits tests below.
func beforeAfterFinding(file string, line, col int, before, after string) finding.Finding {
	return finding.Finding{
		Position:   finding.Position{File: file, Line: line, Column: col},
		BeforeCode: before,
		AfterCode:  after,
	}
}

func TestProvider_CanHandle(t *testing.T) {
	t.Parallel()

	p := &Provider{}

	tests := []struct {
		name string
		f    finding.Finding
		want bool
	}{
		{
			name: "go file with code change",
			f:    finding.Finding{Position: finding.Position{File: "main.go"}, BeforeCode: "old"},
			want: true,
		},
		{
			name: "go file with after code only",
			f:    finding.Finding{Position: finding.Position{File: "main.go"}, AfterCode: "new"},
			want: true,
		},
		{
			name: "non-go file",
			f:    finding.Finding{Position: finding.Position{File: "main.rs"}, BeforeCode: "old"},
			want: false,
		},
		{
			name: "go file without code change",
			f:    finding.Finding{Position: finding.Position{File: "main.go"}},
			want: false,
		},
		{
			name: "empty file path with code change",
			f:    finding.Finding{BeforeCode: "old"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := p.CanHandle(tt.f); got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Name(t *testing.T) {
	t.Parallel()

	p := &Provider{}

	if name := p.Name(); name != "go-ast" {
		t.Errorf("Name() = %q, want %q", name, "go-ast")
	}
}

func TestProvider_Edits_ReplaceWithPosition(t *testing.T) {
	t.Parallel()

	content := []byte(
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n\tfmt.Println(\"world\")\n}\n",
	)

	f := beforeAfterFinding("test.go", 6, 2, "fmt.Println", "log.Printf")

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}

	edit := edits[0]
	if edit.Length != len("fmt.Println") {
		t.Errorf("edit.Length = %d, want %d", edit.Length, len("fmt.Println"))
	}

	got := applyEdits(content, edits)
	expected := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tlog.Printf(\"hello\")\n\tfmt.Println(\"world\")\n}\n"

	if string(got) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestProvider_Edits_ReplaceSecondOccurrence(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\tfirst()\n\tsecond()\n\tthird()\n}\n")

	f := beforeAfterFinding("test.go", 5, 2, "second()", "replaced()")

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}

	got := applyEdits(content, edits)
	expected := "package main\n\nfunc main() {\n\tfirst()\n\treplaced()\n\tthird()\n}\n"

	if string(got) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestProvider_Edits_Insert(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n}\n")

	f := finding.Finding{
		Position:  finding.Position{File: "test.go", Line: 4, Column: 1},
		AfterCode: "\tprintln(\"inserted\")",
	}

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}

	if !edits[0].IsInsert() {
		t.Errorf("expected insertion edit, got Length=%d", edits[0].Length)
	}
}

func TestProvider_Edits_ByteOffsetRange(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\ttarget()\n}\n")

	targetIdx := bytes.Index(content, []byte("target()"))
	if targetIdx < 0 {
		t.Fatal("test setup: target() not found")
	}

	f := finding.Finding{
		Range: &finding.Range{
			Start: finding.Position{Offset: targetIdx},
			End:   finding.Position{Offset: targetIdx + len("target()")},
		},
		BeforeCode: "target()",
		AfterCode:  "replaced()",
		Position:   finding.Position{File: "test.go"},
	}

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}

	got := applyEdits(content, edits)
	if !bytes.Contains(got, []byte("replaced()")) {
		t.Errorf("expected replaced() in result, got:\n%s", got)
	}
}

func TestProvider_Edits_BeforeCodeOnly(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")

	f := finding.Finding{
		Position:   finding.Position{File: "test.go"},
		BeforeCode: "fmt.Println",
		AfterCode:  "log.Printf",
	}

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}
}

func TestProvider_Edits_ParseFailure(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc {{{ broken }}}\n")

	f := beforeAfterFinding("broken.go", 3, 1, "broken", "fixed")

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() should not return error on parse failure: %v", err)
	}

	if edits != nil {
		t.Errorf("Edits() should return nil edits on parse failure, got %d", len(edits))
	}
}

func TestProvider_Edits_CacheReused(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\talpha()\n\tbeta()\n}\n")

	p := &Provider{}

	f1 := beforeAfterFinding("test.go", 4, 2, "alpha()", "first()")

	edits1, err := p.Edits(content, f1)
	if err != nil || len(edits1) != 1 {
		t.Fatalf("first Edits() failed: err=%v, len=%d", err, len(edits1))
	}

	f2 := beforeAfterFinding("test.go", 5, 2, "beta()", "second()")

	edits2, err := p.Edits(content, f2)
	if err != nil || len(edits2) != 1 {
		t.Fatalf("second Edits() failed: err=%v, len=%d", err, len(edits2))
	}

	if edits1[0].Offset == edits2[0].Offset {
		t.Errorf("expected different offsets for different lines, both at %d", edits1[0].Offset)
	}

	p.mu.Lock()
	cacheHash := p.cache.hash
	p.mu.Unlock()

	if cacheHash == 0 {
		t.Error("expected non-zero cache hash after parsing")
	}
}

func TestProvider_Edits_NoMatchInASTNode(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\tx := 1\n}\n")

	f := beforeAfterFinding("test.go", 4, 2, "nonexistent", "replacement")

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if edits != nil {
		t.Errorf("expected nil edits for unmatched BeforeCode, got %d", len(edits))
	}
}

func TestProvider_Edits_BeforeCodeMismatchAtOffset(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\tvalue := 42\n\tprint(value)\n}\n")

	f := beforeAfterFinding("test.go", 4, 2, "value", "result")

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}
}

func TestProvider_FixEngineIntegration(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\talpha()\n\tbeta()\n\tgamma()\n}\n")

	engine := pipeline.NewFixEngineWithProviders(
		&Provider{},
		&pipeline.SubstringProvider{},
	)

	findings := []finding.Finding{
		{
			Position:   finding.Position{File: "test.go", Line: 4, Column: 2},
			BeforeCode: "alpha()",
			AfterCode:  "first()",
		},
		{
			Position:   finding.Position{File: "test.go", Line: 6, Column: 2},
			BeforeCode: "gamma()",
			AfterCode:  "third()",
		},
	}

	result, _, count := engine.Apply(content, findings)
	if count != 2 {
		t.Errorf("expected 2 applied fixes, got %d", count)
	}

	expected := "package main\n\nfunc main() {\n\tfirst()\n\tbeta()\n\tthird()\n}\n"

	if string(result) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

func applyEdits(content []byte, edits []pipeline.FixEdit) []byte {
	result := content

	for _, edit := range edits {
		result = append(
			append(append([]byte{}, result[:edit.Offset]...), edit.Replacement...),
			result[edit.Offset+edit.Length:]...,
		)
	}

	return result
}
