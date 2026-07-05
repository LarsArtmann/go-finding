package gotoken

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestLineColToOffset(t *testing.T) {
	t.Parallel()

	src := []byte("package main\n\nfunc main() {}\n")
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tokenFile := fset.File(file.Pos())

	tests := []struct {
		name       string
		line       int
		col        int
		contentLen int
		wantOffset int
		wantOk     bool
	}{
		{"line 1 col 1", 1, 1, len(src), 0, true},
		{"line 3 col 1", 3, 1, len(src), 14, true},
		{"line 3 col 6", 3, 6, len(src), 19, true},
		{"line 0 invalid", 0, 1, len(src), 0, false},
		{"line 99 too high", 99, 1, len(src), 0, false},
		{"offset beyond content", 1, 200, len(src), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offset, ok := LineColToOffset(tokenFile, tt.line, tt.col, tt.contentLen)
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}

			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}

	t.Run("nil file", func(t *testing.T) {
		t.Parallel()

		_, ok := LineColToOffset(nil, 1, 1, 100)
		if ok {
			t.Error("expected ok=false for nil file")
		}
	})
}

func TestLineColToPos(t *testing.T) {
	t.Parallel()

	src := []byte("package main\n\nfunc main() {}\n")
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", src, 0)
	tokenFile := fset.File(file.Pos())

	pos := LineColToPos(tokenFile, 3, 1)
	if pos == token.NoPos {
		t.Error("expected non-NoPos for valid line")
	}

	pos = LineColToPos(tokenFile, 99, 1)
	if pos != token.NoPos {
		t.Error("expected NoPos for invalid line")
	}

	pos = LineColToPos(nil, 1, 1)
	if pos != token.NoPos {
		t.Error("expected NoPos for nil file")
	}
}

func TestFindFileByName(t *testing.T) {
	t.Parallel()

	src := []byte("package main\n")
	fset := token.NewFileSet()
	_, _ = parser.ParseFile(fset, "myfile.go", src, 0)

	tokenFile := FindFileByName(fset, "myfile.go")
	if tokenFile == nil {
		t.Fatal("expected to find myfile.go")
	}

	if tokenFile.Name() != "myfile.go" {
		t.Errorf("got name %q, want %q", tokenFile.Name(), "myfile.go")
	}

	tokenFile = FindFileByName(fset, "nonexistent.go")
	if tokenFile != nil {
		t.Error("expected nil for nonexistent file")
	}
}

func TestFindInnermostNode(t *testing.T) {
	t.Parallel()

	src := []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", src, 0)

	// Find the byte offset of "println"
	// In "func main() {\n\tprintln(..." the println should be at offset ~25
	node := FindInnermostNode(fset, file, 25)
	if node == nil {
		t.Fatal("expected non-nil node at offset 25")
	}

	// Out of range offset
	node = FindInnermostNode(fset, file, 9999)
	if node != nil {
		t.Error("expected nil for out-of-range offset")
	}
}

func TestNodeByteRange(t *testing.T) {
	t.Parallel()

	src := []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "test.go", src, 0)

	// Walk to find the first declaration (the func)
	var funcDecl ast.Node

	for _, decl := range file.Decls {
		funcDecl = decl

		break
	}

	start, end, ok := NodeByteRange(fset, funcDecl)
	if !ok {
		t.Fatal("expected ok=true")
	}

	if start >= end {
		t.Errorf("expected start < end, got start=%d end=%d", start, end)
	}

	// Verify the range contains "func"
	if string(src[start:start+4]) != "func" {
		t.Errorf("expected 'func' at offset %d, got %q", start, src[start:start+4])
	}
}
