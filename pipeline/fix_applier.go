package pipeline

import (
	"context"
	"errors"
	"fmt"
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

// NewFixApplier creates a new FixApplier with default text-based providers.
func NewFixApplier(rootDir string) (*FixApplier, error) {
	return NewFixApplierWithProviders(rootDir)
}

// NewFixApplierWithProviders creates a FixApplier with custom fix providers.
// Use this to register domain-specific providers (e.g., Go AST, Rust syn).
// When called with no providers, uses the default provider chain.
func NewFixApplierWithProviders(rootDir string, providers ...FixProvider) (*FixApplier, error) {
	backupDir, err := os.MkdirTemp("", "go-finding-backups-*")
	if err != nil {
		return nil, fmt.Errorf(
			"create backup directory for rootDir=%s providers=%d: %w",
			rootDir,
			len(providers),
			err,
		)
	}

	var engine *FixEngine
	if len(providers) == 0 {
		engine = NewFixEngine()
	} else {
		engine = NewFixEngineWithProviders(providers...)
	}

	return &FixApplier{
		rootDir: rootDir,
		backup:  NewFileBackup(backupDir),
		engine:  engine,
	}, nil
}

// Close removes the temporary backup directory. Implement io.Closer.
func (a *FixApplier) Close() error {
	if a.backup != nil && a.backup.IsEnabled() {
		err := os.RemoveAll(a.backup.backupDir)
		if err != nil {
			return fmt.Errorf("removing backup dir: %w", err)
		}
	}

	return nil
}

// ioErrorAt creates an IO error with position info.
func ioErrorAt(msg string, err error, path string) error {
	pos := finding.Position{File: finding.FilePath(path), Offset: -1} //nolint:exhaustruct

	return finding.NewIOError(msg, err).WithPosition(pos)
}

// Apply applies the given fixes to files and returns the number of successful fixes.
// If an error occurs, all previously modified files are rolled back to their backups.
func (a *FixApplier) Apply(ctx context.Context, fixes []finding.Finding) (int, error) {
	applied, _, _, err := a.ApplyWithShiftMap(ctx, fixes)

	return applied, err
}

// ApplyWithDetails applies the given fixes and returns the count of successful fixes,
// the list of successfully applied findings, and any error.
// If an error occurs, all previously modified files are rolled back to their backups.
func (a *FixApplier) ApplyWithDetails(
	ctx context.Context,
	fixes []finding.Finding,
) (int, []finding.Finding, error) {
	applied, appliedFixes, _, err := a.ApplyWithShiftMap(ctx, fixes)

	return applied, appliedFixes, err
}

// ApplyWithShiftMap applies fixes and returns the count, applied findings,
// a per-file line shift map, and any error. The shift map can be used to
// update remaining findings' line numbers after fixes are applied.
func (a *FixApplier) ApplyWithShiftMap(
	ctx context.Context,
	fixes []finding.Finding,
) (int, []finding.Finding, map[string]*LineShiftMap, error) {
	byFile := a.groupFindingsBySafePath(fixes)

	var (
		applied   []finding.Finding
		modified  []string
		shiftMaps = make(map[string]*LineShiftMap)
	)

	paths := slices.Sorted(maps.Keys(byFile))

	for _, path := range paths {
		fileFixes := byFile[path]

		err := CheckCanceledWithMsg(ctx, "fix application cancelled")
		if err != nil {
			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, shiftMaps, err
		}

		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				_ = a.backup.RollbackAll(modified)

				return len(applied), applied, shiftMaps, finding.NewIOError("backup "+path, err)
			}
		}

		fileApplied, shiftMap, err := a.applyToFile(path, fileFixes)
		if err != nil {
			if a.backup.IsEnabled() {
				_ = a.backup.Restore(path)
			}

			_ = a.backup.RollbackAll(modified)

			return len(applied), applied, shiftMaps, finding.NewConflictError("apply to "+path, err)
		}

		modified = append(modified, path)
		applied = append(applied, fileApplied...)

		a.recordShiftMap(shiftMap, fileFixes, shiftMaps)
	}

	return len(applied), applied, shiftMaps, nil
}

// groupFindingsBySafePath groups findings by their resolved filesystem path,
// skipping findings without a file or with unsafe path traversal.
func (a *FixApplier) groupFindingsBySafePath(fixes []finding.Finding) map[string][]finding.Finding {
	byFile := make(map[string][]finding.Finding)

	// Resolve root once instead of per-finding (was O(N) syscalls).
	cleanRoot := filepath.Clean(a.rootDir)

	if resolved, err := filepath.EvalSymlinks(cleanRoot); err == nil {
		cleanRoot = resolved
	}

	for _, f := range fixes {
		if f.Position.File == "" {
			continue
		}

		path := filepath.Join(a.rootDir, string(f.Position.File))

		cleanPath := filepath.Clean(path)

		resolved, err := filepath.EvalSymlinks(cleanPath)
		if err == nil {
			cleanPath = resolved
		}

		if cleanPath != cleanRoot &&
			!strings.HasPrefix(cleanPath, cleanRoot+string(os.PathSeparator)) {
			continue
		}

		byFile[path] = append(byFile[path], f)
	}

	return byFile
}

// recordShiftMap stores the shift map indexed by relative file path.
func (*FixApplier) recordShiftMap(
	shiftMap *LineShiftMap,
	fileFixes []finding.Finding,
	shiftMaps map[string]*LineShiftMap,
) {
	if shiftMap == nil || len(shiftMap.entries) == 0 {
		return
	}

	var relPath string

	for _, f := range fileFixes {
		if f.Position.File != "" {
			relPath = string(f.Position.File)

			break
		}
	}

	if relPath != "" {
		shiftMaps[relPath] = shiftMap
	}
}

// applyToFile applies fixes to a single file using byte-level edits.
// Returns the applied findings and an optional line shift map.
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) ([]finding.Finding, *LineShiftMap, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, ioErrorAt("stat file", err, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, ioErrorAt("read file", err, path)
	}

	appliedFixes, appliedEdits, _, newContent, resolveErrors := a.engine.ApplyWithConflicts(content, fixes)

	if len(appliedFixes) == 0 {
		return nil, nil, errors.Join(resolveErrors...)
	}

	var shiftMap *LineShiftMap
	if len(appliedEdits) > 0 {
		shiftMap = NewLineShiftMap(content, appliedEdits)
	}

	err = os.WriteFile( //nolint:gosec // intentional file write
		path,
		newContent,
		info.Mode(),
	)
	if err != nil {
		return nil, nil, ioErrorAt("write file", err, path)
	}

	return appliedFixes, shiftMap, errors.Join(resolveErrors...)
}
