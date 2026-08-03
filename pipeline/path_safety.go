package pipeline

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveRoot resolves rootDir to its canonical form: cleaned + symlinks evaluated.
// Call once per rootDir and reuse the result for multiple resolveSafePathFrom calls
// to avoid redundant EvalSymlinks on the root directory.
func resolveRoot(rootDir string) string {
	cleanRoot := filepath.Clean(rootDir)

	if resolved, err := filepath.EvalSymlinks(cleanRoot); err == nil {
		cleanRoot = resolved
	}

	return cleanRoot
}

// resolveSafePathFrom resolves a relative file path against a pre-resolved root,
// resolving symlinks and verifying the result stays within root. Returns the
// resolved absolute path and true if safe; empty string and false otherwise.
//
// Use resolveRoot to pre-compute the resolved root once, then call this for
// each file path. This avoids redundant EvalSymlinks(rootDir) calls when
// processing many findings against the same root.
func resolveSafePathFrom(resolvedRoot, relPath string) (string, bool) {
	// If relPath is already absolute, use it directly. Some tools (e.g., cqrs-lint)
	// store absolute paths in finding positions, so joining with rootDir would
	// double the path (rootDir + absolutePath = rootDir/rootDir/...).
	// The containment check below still ensures the path stays within rootDir.
	fullPath := relPath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(resolvedRoot, relPath)
	}

	cleanPath := filepath.Clean(fullPath)

	resolved, err := filepath.EvalSymlinks(cleanPath)
	if err == nil {
		cleanPath = resolved
	}

	if cleanPath != resolvedRoot &&
		!strings.HasPrefix(cleanPath, resolvedRoot+string(os.PathSeparator)) {
		return "", false
	}

	return cleanPath, true
}

// resolveSafePath resolves a relative file path against rootDir, resolving
// symlinks and verifying the result stays within rootDir. Returns the resolved
// absolute path and true if safe; empty string and false otherwise.
//
// This is the single source of truth for path containment checks, used by
// both FixApplier.groupFindingsBySafePath and Pipeline.filterByFileEdits to
// prevent path traversal attacks (e.g., Position.File = "../../etc/passwd").
//
// For batch operations processing many file paths against the same root,
// prefer resolveRoot + resolveSafePathFrom to avoid redundant root resolution.
func resolveSafePath(rootDir, relPath string) (string, bool) {
	return resolveSafePathFrom(resolveRoot(rootDir), relPath)
}
