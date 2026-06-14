package goast

import (
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func TestProvider_Edits_BeforeCodeOnly(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

func main() {
	fmt.Println("hello")
}
`)

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

	// Intentionally invalid Go syntax
	content := []byte(`package main

func {{{ broken }}}
`)

	f := finding.Finding{
		Position:   finding.Position{File: "broken.go", Line: 3, Column: 1},
		BeforeCode: "broken",
		AfterCode:  "fixed",
	}

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

	content := []byte("package main\n\nfunc main() {\n\told()\n\t")

	p := &Provider{}

	// First call triggers parse
	f1 := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 4, Column: 2},
		BeforeCode: "old()",
		AfterCode:  "new()",
	}

	edits1, err := p.Edits(content, f1)
	if err != nil || len(edits1) != 1 {
		t.Fatalf("first Edits() failed: err=%v, len=%d", err, len(edits1))
	}

	// Second call should reuse cache
	f2 := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 5, Column: 2},
		BeforeCode: "old()",
		AfterCode:  "newer()",
	}

	edits2, err := p.Edits(content, f2)
	if err != nil || len(edits2) != 1 {
		t.Fatalf("second Edits() failed: err=%v, len=%d", err, len(edits2))
	}

	// Verify different offsets (different lines)
	if edits1[0].Offset == edits2[0].Offset {
		t.Errorf("expected different offsets for different lines, both at %d", edits1[0].Offset)
	}

	// Verify cache hash matches
	p.mu.Lock()
	cacheHash := p.cache.hash
	p.mu.Unlock()

	if cacheHash == 0 {
		t.Error("expected non-zero cache hash after parsing")
	}
}

func TestProvider_Edits_NoMatchInASTNode(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

func main() {
	x := 1
}
`)

	// BeforeCode doesn't exist in the content at the given position
	f := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 4, Column: 2},
		BeforeCode: "nonexistent",
		AfterCode:  "replacement",
	}

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

	content := []byte(`package main

func main() {
	value := 42
	print(value)
}
`)

	// Position points to line 4, but BeforeCode "print" is on line 5
	// The AST node search should find it within the containing block
	f := finding.Finding{
		Position:   finding.Position{File: "test.go", Line: 4, Column: 2},
		BeforeCode: "value",
		AfterCode:  "result",
	}

	p := &Provider{}

	edits, err := p.Edits(content, f)
	if err != nil {
		t.Fatalf("Edits() error: %v", err)
	}

	// Should find "value" within the AST node at line 4
	if len(edits) != 1 {
		t.Fatalf("Edits() returned %d edits, want 1", len(edits))
	}
}

func TestProvider_FixEngineIntegration(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc main() {\n\told()\n\t")

	// Register GoASTProvider before text providers
	engine := pipeline.NewFixEngineWithProviders(
		&Provider{},
		&pipeline.SubstringProvider{},
	)

	findings := []finding.Finding{
		{
			Position:   finding.Position{File: "test.go", Line: 4, Column: 2},
			BeforeCode: "old()",
			AfterCode:  "first()",
		},
		{
			Position:   finding.Position{File: "test.go", Line: 6, Column: 2},
			BeforeCode: "old()",
			AfterCode:  "third()",
		},
	}

	result, applied, count := engine.Apply(content, findings)
	if count != 2 {
		t.Errorf("expected 2 applied fixes, got %d", len(applied))
	}

	expected := `package main

func main() {
	first()
	old()
	third()
}
`

	if string(result) != expected {
		t.Errorf("result mismatch:\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

// applyEdits is a test helper that applies edits to content.
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
