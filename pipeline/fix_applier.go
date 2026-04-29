package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// Group fixes by file
	byFile := make(map[string][]finding.Finding)

	for _, f := range fixes {
		if f.Position.File == "" {
			continue
		}

		path := filepath.Join(a.rootDir, f.Position.File)
		byFile[path] = append(byFile[path], f)
	}

	applied := 0
	var modified []string

	for path, fileFixes := range byFile {
		select {
		case <-ctx.Done():
			_ = a.backup.RollbackAll(modified)

			return applied, fmt.Errorf("fix application cancelled: %w", ctx.Err())
		default:
		}

		// Create backup
		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				_ = a.backup.RollbackAll(modified)

				return applied, finding.NewIOError("backup "+path, err)
			}
		}

		// Apply fixes
		count, err := a.applyToFile(path, fileFixes)
		if err != nil {
			// Restore current file from backup
			if a.backup.IsEnabled() {
				_ = a.backup.Restore(path)
			}

			// Restore all previously modified files
			_ = a.backup.RollbackAll(modified)

			return applied, finding.NewConflictError("apply to "+path, err)
		}

		modified = append(modified, path)
		applied += count
	}

	return applied, nil
}

// applyToFile applies fixes to a single file.
// When a finding has a Range with valid end position, it uses line-based replacement
// targeting the exact line range. Otherwise it falls back to string replacement.
// Fixes are sorted descending by position so earlier replacements don't shift later ones.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, ioErrorAt("read file", err, path)
	}

	lines := strings.Split(string(content), "\n")
	newLines, applied := a.engine.Apply(lines, fixes)

	if applied == 0 {
		return 0, nil
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write in fix applier
		path,
		[]byte(strings.Join(newLines, "\n")),
		0o600,
	); err != nil {
		return 0, ioErrorAt("write file", err, path)
	}

	return applied, nil
}
