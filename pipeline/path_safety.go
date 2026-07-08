package pipeline

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveSafePath resolves a relative file path against rootDir, resolving
// symlinks and verifying the result stays within rootDir. Returns the resolved
// absolute path and true if safe; empty string and false otherwise.
//
// This is the single source of truth for path containment checks, used by
// both FixApplier.groupFindingsBySafePath and Pipeline.filterByFileEdits to
// prevent path traversal attacks (e.g., Position.File = "../../etc/passwd").
func resolveSafePath(rootDir, relPath string) (string, bool) {
	cleanRoot := filepath.Clean(rootDir)

	if resolved, err := filepath.EvalSymlinks(cleanRoot); err == nil {
		cleanRoot = resolved
	}

	fullPath := filepath.Join(rootDir, relPath)
	cleanPath := filepath.Clean(fullPath)

	resolved, err := filepath.EvalSymlinks(cleanPath)
	if err == nil {
		cleanPath = resolved
	}

	if cleanPath != cleanRoot &&
		!strings.HasPrefix(cleanPath, cleanRoot+string(os.PathSeparator)) {
		return "", false
	}

	return cleanPath, true
}
