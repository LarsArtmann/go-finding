package pipeline

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/larsartmann/go-finding"
)

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir string
	backup  *FileBackup
	engine  *FixEngine
}

// NewFixApplier creates a new FixApplier.
func NewFixApplier(rootDir string) *FixApplier {
	backupDir, err := os.MkdirTemp("", "go-finding-backups-*")
	if err != nil {
		backupDir = filepath.Join(os.TempDir(), "go-finding-backups")
	}

	return &FixApplier{
		rootDir: rootDir,
		backup:  NewFileBackup(backupDir),
		engine:  NewFixEngine(),
	}
}

// ioErrorAt creates an IO error with position info.
func ioErrorAt(msg string, err error, path string) error {
	pos := finding.Position{File: path} //nolint:exhaustruct
	return finding.NewIOError(msg, err).WithPosition(pos)
}

// Apply applies the given fixes to files and returns the number of successful fixes.
// If an error occurs, all previously modified files are rolled back to their backups.
func (a *FixApplier) Apply(ctx context.Context, fixes []finding.Finding) (int, error) {
	applied, _, err := a.ApplyWithDetails(ctx, fixes)

	return applied, err
}

// ApplyWithDetails applies the given fixes and returns the count of successful fixes,
// the list of successfully applied findings, and any error.
// If an error occurs, all previously modified files are rolled back to their backups.
func (a *FixApplier) ApplyWithDetails(
	ctx context.Context,
	fixes []finding.Finding,
) (int, []finding.Finding, error) {
	// Group fixes by file
	byFile := make(map[string][]finding.Finding)

	for _, f := range fixes {
		if f.Position.File == "" {
			continue
		}

		path := filepath.Join(a.rootDir, f.Position.File)
		byFile[path] = append(byFile[path], f)
	}

	var applied []finding.Finding
	var modified []string

	// Sort file paths for deterministic, reproducible fix application order.
	paths := slices.Collect(maps.Keys(byFile))
	slices.Sort(paths)

	for _, path := range paths {
		fileFixes := byFile[path]
		if err := CheckCanceledWithMsg(ctx, "fix application cancelled"); err != nil {
			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, err
		}

		// Create backup
		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				_ = a.backup.RollbackAll(modified)

				return len(applied), applied, finding.NewIOError("backup "+path, err)
			}
		}

		// Apply fixes
		fileApplied, err := a.applyToFile(path, fileFixes)
		if err != nil {
			// Restore current file from backup
			if a.backup.IsEnabled() {
				_ = a.backup.Restore(path)
			}

			// Restore all previously modified files
			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, finding.NewConflictError("apply to "+path, err)
		}

		modified = append(modified, path)
		applied = append(applied, fileApplied...)
	}

	return len(applied), applied, nil
}

// applyToFile applies fixes to a single file.
// When a finding has a Range with valid end position, it uses line-based replacement
// targeting the exact line range. Otherwise it falls back to string replacement.
// Fixes are sorted descending by position so earlier replacements don't shift later ones.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) ([]finding.Finding, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, ioErrorAt("read file", err, path)
	}

	lines := strings.Split(string(content), "\n")
	newLines, applied, _ := a.engine.ApplyWithDetails(lines, fixes)

	if len(applied) == 0 {
		return nil, nil
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write in fix applier
		path,
		[]byte(strings.Join(newLines, "\n")),
		0o600,
	); err != nil {
		return nil, ioErrorAt("write file", err, path)
	}

	return applied, nil
}
