package pipeline

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/larsartmann/go-finding"
)

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir string
	backup  *FileBackup
	engine  *FixEngine
}

// NewFixApplier creates a new FixApplier with default text-based providers.
func NewFixApplier(rootDir string) (*FixApplier, error) {
	backupDir, err := os.MkdirTemp("", "go-finding-backups-*")
	if err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	return &FixApplier{
		rootDir: rootDir,
		backup:  NewFileBackup(backupDir),
		engine:  NewFixEngine(),
	}, nil
}

// NewFixApplierWithProviders creates a FixApplier with custom fix providers.
// Use this to register domain-specific providers (e.g., Go AST, Rust syn).
func NewFixApplierWithProviders(rootDir string, providers ...FixProvider) (*FixApplier, error) {
	backupDir, err := os.MkdirTemp("", "go-finding-backups-*")
	if err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	return &FixApplier{
		rootDir: rootDir,
		backup:  NewFileBackup(backupDir),
		engine:  NewFixEngineWithProviders(providers...),
	}, nil
}

// Close removes the temporary backup directory. Implement io.Closer.
func (a *FixApplier) Close() error {
	if a.backup != nil && a.backup.IsEnabled() {
		if err := os.RemoveAll(a.backup.backupDir); err != nil {
			return fmt.Errorf("removing backup dir: %w", err)
		}
	}

	return nil
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

	paths := slices.Collect(maps.Keys(byFile))
	slices.Sort(paths)

	for _, path := range paths {
		fileFixes := byFile[path]
		if err := CheckCanceledWithMsg(ctx, "fix application cancelled"); err != nil {
			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, err
		}

		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				_ = a.backup.RollbackAll(modified)

				return len(applied), applied, finding.NewIOError("backup "+path, err)
			}
		}

		fileApplied, err := a.applyToFile(path, fileFixes)
		if err != nil {
			if a.backup.IsEnabled() {
				_ = a.backup.Restore(path)
			}

			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, finding.NewConflictError("apply to "+path, err)
		}

		modified = append(modified, path)
		applied = append(applied, fileApplied...)
	}

	return len(applied), applied, nil
}

// applyToFile applies fixes to a single file using byte-level edits.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) ([]finding.Finding, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, ioErrorAt("read file", err, path)
	}

	appliedFixes, _, newContent := a.engine.ApplyWithConflicts(content, fixes)

	if len(appliedFixes) == 0 {
		return nil, nil
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write
		path,
		newContent,
		0o600,
	); err != nil {
		return nil, ioErrorAt("write file", err, path)
	}

	return appliedFixes, nil
}
