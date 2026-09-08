package pipeline

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/larsartmann/go-finding"
)

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir        string
	backup         *FileBackup
	engine         *FixEngine
	rollbackPolicy RollbackPolicy
}

// RollbackPolicy controls which files are restored when a file fails during a
// multi-file fix run.
type RollbackPolicy int

const (
	// RollbackPolicyFailingFile keeps every file applied so far and restores
	// only the failing file. Per-finding soft failures (provider resolve
	// errors, refused findings) never abort the run: they are reported in
	// ApplyReport.Outcomes while all successfully applied fixes stay on disk.
	// This is the default and matches consumers that batch independent fixes
	// (e.g., a lint --fix run across many files).
	RollbackPolicyFailingFile RollbackPolicy = iota

	// RollbackPolicyAllFiles provides all-or-nothing semantics: any file
	// failure restores the failing file and every file modified earlier in
	// the run, preserving the pre-call state of all inputs.
	RollbackPolicyAllFiles
)

// SetRollbackPolicy sets the failure behavior for multi-file runs.
// The default is RollbackPolicyFailingFile.
func (a *FixApplier) SetRollbackPolicy(p RollbackPolicy) {
	a.rollbackPolicy = p
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
	pos := finding.Position{File: finding.FilePath(path), Offset: -1}

	return finding.NewIOError(msg, err).WithPosition(pos)
}

// Apply applies the given fixes to files and returns the number of successful fixes.
// On a file failure, rollback scope is governed by the applier's RollbackPolicy:
// by default only the failing file is restored and earlier files keep their fixes.
func (a *FixApplier) Apply(ctx context.Context, fixes []finding.Finding) (int, error) {
	report, err := a.ApplyWithReport(ctx, fixes)

	return report.Applied, err
}

// ApplyWithDetails applies the given fixes and returns the count of successful fixes,
// the list of successfully applied findings, and any error.
// On a file failure, rollback scope is governed by the applier's RollbackPolicy:
// by default only the failing file is restored and earlier files keep their fixes.
func (a *FixApplier) ApplyWithDetails(
	ctx context.Context,
	fixes []finding.Finding,
) (int, []finding.Finding, error) {
	report, err := a.ApplyWithReport(ctx, fixes)

	return report.Applied, report.AppliedFixes, err
}

// ApplyWithShiftMap applies fixes and returns the count, applied findings,
// a per-file line shift map, and any error. The shift map can be used to
// update remaining findings' line numbers after fixes are applied.
// On a file failure, rollback scope is governed by the applier's RollbackPolicy:
// by default only the failing file is restored and earlier files keep their fixes.
func (a *FixApplier) ApplyWithShiftMap(
	ctx context.Context,
	fixes []finding.Finding,
) (int, []finding.Finding, map[string]*LineShiftMap, error) {
	report, err := a.ApplyWithReport(ctx, fixes)

	return report.Applied, report.AppliedFixes, report.ShiftMaps, err
}

// ApplyReport is the detailed result of one FixApplier run.
type ApplyReport struct {
	// Applied is the number of findings successfully written to disk.
	Applied int
	// AppliedFixes lists the applied findings in application order.
	AppliedFixes []finding.Finding
	// ShiftMaps maps relative file paths to line shift maps for files with
	// applied edits.
	ShiftMaps map[string]*LineShiftMap
	// Outcomes holds one entry per fixable input finding, in processing
	// order, distinguishing applied / no-change / refused / conflict /
	// invalid / failed findings. Failed outcomes carry the provider error.
	Outcomes []FixOutcome
	// RolledBack lists the file paths restored from backup due to a failure.
	RolledBack []string
}

// FailedOutcomes returns the outcomes with Status FixOutcomeFailed.
func (r ApplyReport) FailedOutcomes() []FixOutcome {
	failed := make([]FixOutcome, 0, len(r.Outcomes))
	for _, o := range r.Outcomes {
		if o.Status == FixOutcomeFailed {
			failed = append(failed, o)
		}
	}

	return failed
}

// ApplyWithReport applies fixes and returns a detailed report (applied fixes,
// per-finding outcomes, line shift maps, rolled-back files) plus any error.
// Files are processed in sorted path order. Soft per-finding failures are
// collected in the report and returned as a joined error at the end; a hard
// file failure stops the run and restores files according to RollbackPolicy.
func (a *FixApplier) ApplyWithReport(
	ctx context.Context,
	fixes []finding.Finding,
) (ApplyReport, error) {
	byFile := a.groupFindingsBySafePath(fixes)

	report := ApplyReport{ShiftMaps: make(map[string]*LineShiftMap)}
	var modified []string

	paths := slices.Sorted(maps.Keys(byFile))

	for _, path := range paths {
		fileFixes := byFile[path]

		err := CheckCanceledWithMsg(ctx, "fix application cancelled")
		if err != nil {
			if a.rollbackPolicy == RollbackPolicyAllFiles {
				if rollbackErr := a.backup.RollbackAll(modified); rollbackErr != nil {
					return report, fmt.Errorf("%w (rollback also failed: %w)", err, rollbackErr)
				}

				report.RolledBack = append(report.RolledBack, modified...)
			}

			return report, err
		}

		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				backupErr := finding.NewIOError("backup "+path, err)

				if a.rollbackPolicy == RollbackPolicyAllFiles {
					if rollbackErr := a.backup.RollbackAll(modified); rollbackErr != nil {
						return report, fmt.Errorf("%w (rollback also failed: %w)", backupErr, rollbackErr)
					}

					report.RolledBack = append(report.RolledBack, modified...)
				}

				return report, backupErr
			}
		}

		fileApplied, shiftMap, outcomes, err := a.applyToFile(path, fileFixes)
		report.Outcomes = append(report.Outcomes, outcomes...)
		if err != nil {
			return report, a.handleFileError(&report, path, modified, err)
		}

		modified = append(modified, path)
		report.AppliedFixes = append(report.AppliedFixes, fileApplied...)
		report.Applied = len(report.AppliedFixes)

		a.recordShiftMap(shiftMap, fileFixes, report.ShiftMaps)
	}

	report.Applied = len(report.AppliedFixes)

	var softErrs []error
	for _, o := range report.Outcomes {
		if o.Err != nil {
			softErrs = append(softErrs, o.Err)
		}
	}

	return report, errors.Join(softErrs...)
}

// handleFileError restores the failing file, applies the rollback policy to
// files modified earlier in the run, and wraps the failure.
func (a *FixApplier) handleFileError(report *ApplyReport, path string, modified []string, err error) error {
	var rollbackErrs []error

	if a.backup.IsEnabled() {
		if restoreErr := a.backup.Restore(path); restoreErr != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("restore %s: %w", path, restoreErr))
		} else {
			report.RolledBack = append(report.RolledBack, path)
		}
	}

	if a.rollbackPolicy == RollbackPolicyAllFiles {
		if rollbackErr := a.backup.RollbackAll(modified); rollbackErr != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("rollback: %w", rollbackErr))
		} else {
			report.RolledBack = append(report.RolledBack, modified...)
		}
	}

	applyErr := finding.NewConflictError("apply to "+path, err)
	if len(rollbackErrs) > 0 {
		return fmt.Errorf("%w (rollback also failed: %w)", applyErr, errors.Join(rollbackErrs...))
	}

	return applyErr
}

// groupFindingsBySafePath groups findings by their resolved filesystem path,
// skipping findings without a file or with unsafe path traversal.
// Uses the resolved path (not the raw join) as the map key, preventing
// TOCTOU races where a symlink is swapped between validation and file I/O.
//
// The root directory is resolved once and each unique file path is resolved
// at most once, avoiding redundant EvalSymlinks syscalls when many findings
// target the same file.
func (a *FixApplier) groupFindingsBySafePath(fixes []finding.Finding) map[string][]finding.Finding {
	byFile := make(map[string][]finding.Finding)
	resolvedRoot := ResolveRoot(a.rootDir)
	pathCache := make(map[string]string, len(fixes))

	for _, f := range fixes {
		if f.Position.File == "" {
			continue
		}

		rawPath := string(f.Position.File)

		safePath, cached := pathCache[rawPath]
		if !cached {
			safePath, _ = ResolveSafePathFrom(resolvedRoot, rawPath)
			pathCache[rawPath] = safePath
		}

		if safePath == "" {
			continue
		}

		byFile[safePath] = append(byFile[safePath], f)
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
// Returns the applied findings, an optional line shift map, and per-finding
// outcomes. Soft failures (provider resolve errors, refused findings) are
// reflected in the outcomes and never fail the file: applied edits stay
// written. The returned error is reserved for hard I/O failures (stat, read,
// write).
func (a *FixApplier) applyToFile(path string, fixes []finding.Finding) ([]finding.Finding, *LineShiftMap, []FixOutcome, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, nil, ioErrorAt("stat file", err, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, ioErrorAt("read file", err, path)
	}

	result := a.engine.ApplyWithOutcomes(content, fixes)
	if len(result.AppliedEdits) == 0 {
		return nil, nil, result.Outcomes, nil
	}

	shiftMap := NewLineShiftMap(content, result.AppliedEdits)

	err = os.WriteFile( //nolint:gosec // intentional file write
		path,
		result.Content,
		info.Mode(),
	)
	if err != nil {
		return nil, nil, nil, ioErrorAt("write file", err, path)
	}

	return result.Applied, shiftMap, result.Outcomes, nil
}
