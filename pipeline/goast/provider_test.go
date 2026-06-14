package goast

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-finding"
)

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

	content := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
	fmt.Println("world")
}
`)

	// Replace the FIRST fmt.Println (line 6, col 2)
	// BeforeCode at the exact position: "fmt.Println"
	f := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 6, Column: 2},
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

	edit := edits[0]
	if edit.Length != len("fmt.Println") {
		t.Errorf("edit.Length = %d, want %d", edit.Length, len("fmt.Println"))
	}

	got := applyEdits(content, edits)
	expected := `package main

import "fmt"

func main() {
	log.Printf("hello")
	fmt.Println("world")
}
`

	if string(got) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestProvider_Edits_ReplaceSecondOccurrence(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\told()\n\t")

	// Replace the SECOND old() (line 5, col 2)
	f := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 5, Column: 2},
		BeforeCode: "old()",
		AfterCode:  "new()",
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
	expected := `package main

func main() {
	old()
	new()
}
`

	if string(got) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", got, expected)
	}
}

func TestProvider_Edits_Insert(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

func main() {
}
`)

	// Insert at end of line 4 (before the closing brace)
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

	content := []byte(`package main

func main() {
	old()
}
`)

	// "old()" starts at byte offset 27
	oldIdx := bytes.Index(content, []byte("old()"))
	if oldIdx < 0 {
		t.Fatal("test setup: old() not found")
	}

	f := finding.Finding{
		Range: &finding.Range{
			Start: finding.Position{Offset: oldIdx},
			End:   finding.Position{Offset: oldIdx + len("old()")},
		},
		BeforeCode: "old()",
		AfterCode:  "new()",
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
	if !bytes.Contains(got, []byte("new()")) {
		t.Errorf("expected new() in result, got:\n%s", got)
	}
}
