package pipeline

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

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
// multi-file fix run. See docs/DOMAIN_LANGUAGE.md ("Fix Application") for the
// shared vocabulary. The default is documented on RollbackPolicyFailingFile.
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
	byFile, droppedOutcomes := a.groupFindingsBySafePath(fixes)

	report := ApplyReport{ShiftMaps: make(map[string]*LineShiftMap), Outcomes: droppedOutcomes}
	var modified []string

	paths := slices.Sorted(maps.Keys(byFile))

	for _, path := range paths {
		fileFixes := byFile[path]

		err := CheckCanceledWithMsg(ctx, "fix application cancelled")
		if err != nil {
			if a.rollbackPolicy == RollbackPolicyAllFiles {
				if rollbackErr := a.backup.RollbackAll(modified); rollbackErr != nil {
					return report, fmt.Errorf(
						"%w (rollback also failed: %w)%s",
						err,
						rollbackErr,
						rolledBackNote(report.RolledBack),
					)
				}

				report.RolledBack = append(report.RolledBack, modified...)

				return report, fmt.Errorf("%w%s", err, rolledBackNote(report.RolledBack))
			}

			return report, err
		}

		if a.backup.IsEnabled() {
			err := a.backup.Backup(path)
			if err != nil {
				backupErr := finding.NewIOError("backup "+path, err)

				if a.rollbackPolicy == RollbackPolicyAllFiles {
					if rollbackErr := a.backup.RollbackAll(modified); rollbackErr != nil {
						return report, fmt.Errorf(
							"%w (rollback also failed: %w)%s",
							backupErr,
							rollbackErr,
							rolledBackNote(report.RolledBack),
						)
					}

					report.RolledBack = append(report.RolledBack, modified...)

					return report, fmt.Errorf("%w%s", backupErr, rolledBackNote(report.RolledBack))
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

// ApplyDryRun resolves every fix against the current file contents and
// reports the outcomes and counts a real run would produce — without
// writing, backing up, or rolling back anything (D4). Files are read only.
//
// The report semantics match ApplyWithReport with two differences:
// Applied/AppliedFixes are the WOULD-apply counts against the current
// content, and RolledBack is always empty because nothing is modified.
// Soft per-finding failures (provider errors, unsafe paths) surface as
// failed outcomes and in the joined error return, exactly like a real run.
// Because nothing is written, edits in one file cannot conflict with edits
// resolved later in the same dry run beyond per-file overlap rules.
func (a *FixApplier) ApplyDryRun(
	ctx context.Context,
	fixes []finding.Finding,
) (ApplyReport, error) {
	byFile, droppedOutcomes := a.groupFindingsBySafePath(fixes)

	report := ApplyReport{ShiftMaps: make(map[string]*LineShiftMap), Outcomes: droppedOutcomes}

	for _, path := range slices.Sorted(maps.Keys(byFile)) {
		if err := CheckCanceledWithMsg(ctx, "dry run cancelled"); err != nil {
			return report, err
		}

		result, original, _, err := a.loadAndResolve(path, byFile[path])
		if err != nil {
			return report, err
		}

		report.Outcomes = append(report.Outcomes, result.Outcomes...)

		if len(result.AppliedEdits) == 0 {
			continue
		}

		report.AppliedFixes = append(report.AppliedFixes, result.Applied...)

		a.recordShiftMap(NewLineShiftMap(original, result.AppliedEdits), byFile[path], report.ShiftMaps)
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
		return fmt.Errorf(
			"%w (rollback also failed: %w)%s",
			applyErr,
			errors.Join(rollbackErrs...),
			rolledBackNote(report.RolledBack),
		)
	}

	return fmt.Errorf("%w%s", applyErr, rolledBackNote(report.RolledBack))
}

// rolledBackNote renders the rolled-back file list for error messages so
// callers can see exactly which paths were restored after a failure.
func rolledBackNote(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	return " (rolled back: " + strings.Join(paths, ", ") + ")"
}

// groupFindingsBySafePath groups findings by their resolved filesystem path,
// skipping findings without a file or with unsafe path traversal.
// Uses the resolved path (not the raw join) as the map key, preventing
// TOCTOU races where a symlink is swapped between validation and file I/O.
//
// The root directory is resolved once and each unique file path is resolved
// at most once, avoiding redundant EvalSymlinks syscalls when many findings
// target the same file.
//
// Findings whose path fails the containment check (traversal outside the
// root) are dropped from the map and returned as failed outcomes carrying a
// validation error, so callers report them instead of silently losing them.
func (a *FixApplier) groupFindingsBySafePath(
	fixes []finding.Finding,
) (map[string][]finding.Finding, []FixOutcome) {
	byFile := make(map[string][]finding.Finding)
	resolvedRoot := ResolveRoot(a.rootDir)
	pathCache := make(map[string]string, len(fixes))

	var dropped []FixOutcome

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
			dropped = append(dropped, FixOutcome{
				Finding: f,
				Status:  FixOutcomeFailed,
				Err: finding.NewValidationError(
					fmt.Sprintf("unsafe path %q for finding %s: resolves outside root %s", rawPath, f.ID, a.rootDir),
					nil,
				).WithPosition(f.Position),
			})

			continue
		}

		byFile[safePath] = append(byFile[safePath], f)
	}

	return byFile, dropped
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
func (a *FixApplier) applyToFile(
	path string,
	fixes []finding.Finding,
) ([]finding.Finding, *LineShiftMap, []FixOutcome, error) {
	result, original, info, err := a.loadAndResolve(path, fixes)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(result.AppliedEdits) == 0 {
		return nil, nil, result.Outcomes, nil
	}

	shiftMap := NewLineShiftMap(original, result.AppliedEdits)

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

// loadAndResolve stats and reads the file, then resolves the fixes against
// its content without modifying anything. Shared by the applying and the
// dry-run paths.
func (a *FixApplier) loadAndResolve(path string, fixes []finding.Finding) (FixApplyResult, []byte, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FixApplyResult{}, nil, nil, ioErrorAt("stat file", err, path)
	}

	content, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	if err != nil {
		return FixApplyResult{}, nil, nil, ioErrorAt("read file", err, path)
	}

	result := a.engine.ApplyWithOutcomes(content, fixes)

	return result, content, info, nil
}
