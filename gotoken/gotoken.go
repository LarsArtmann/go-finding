// Package gotoken provides shared utilities for working with [go/token] and
// [go/ast]. These helpers are used by the analysis/ and pipeline/goast/
// packages to avoid duplicating position resolution logic.
//
// All functions depend only on Go stdlib ([go/token], [go/ast]).
package gotoken

import (
	"go/ast"
	"go/token"
)

// LineColToOffset converts a 1-based line and column to a 0-based byte offset
// using a pre-resolved [*token.File]. Returns ok=false if the line is out of
// range or the resulting offset exceeds contentLen.
func LineColToOffset(tokenFile *token.File, line, col, contentLen int) (int, bool) {
	if tokenFile == nil || line < 1 || line > tokenFile.LineCount() {
		return 0, false
	}

	lineStart := tokenFile.LineStart(line)

	offset := tokenFile.Offset(lineStart)

	if col > 1 {
		offset += col - 1
	}

	if offset > contentLen {
		return 0, false
	}

	return offset, true
}

// LineColToPos converts a 1-based line and column to a [token.Pos] using a
// pre-resolved [*token.File]. Returns [token.NoPos] if the line is out of range.
func LineColToPos(tokenFile *token.File, line, col int) token.Pos {
	if tokenFile == nil || line < 1 || line > tokenFile.LineCount() {
		return token.NoPos
	}

	lineStart := tokenFile.LineStart(line)

	if col > 1 {
		return lineStart + token.Pos(col-1)
	}

	return lineStart
}

// FindFileByName searches the file set for a [*token.File] with the given name.
// Returns nil if not found.
func FindFileByName(fset *token.FileSet, name string) *token.File {
	var result *token.File

	fset.Iterate(func(tokenFile *token.File) bool {
		if tokenFile.Name() == name {
			result = tokenFile

			return false
		}

		return true
	})

	return result
}

// FindInnermostNode returns the deepest AST node whose [Pos, End) range
// contains the given byte offset. Returns nil if no node contains the offset.
func FindInnermostNode(fset *token.FileSet, root ast.Node, byteOffset int) ast.Node {
	tokenFile := fset.File(root.Pos())
	if tokenFile == nil {
		return nil
	}

	pos := tokenFile.Pos(byteOffset)

	var result ast.Node

	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		if n.Pos() <= pos && pos < n.End() {
			result = n

			return true
		}

		return false
	})

	return result
}

// NodeByteRange returns the byte offsets [start, end) of an AST node.
// Returns ok=false if the node positions cannot be resolved.
func NodeByteRange(fset *token.FileSet, node ast.Node) (int, int, bool) {
	tokenFile := fset.File(node.Pos())
	if tokenFile == nil {
		return 0, 0, false
	}

	start := tokenFile.Offset(node.Pos())
	end := tokenFile.Offset(node.End())

	if start > end {
		return 0, 0, false
	}

	return start, end, true
}
