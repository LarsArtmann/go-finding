package pipeline

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/larsartmann/go-finding"
)

// FixApplier handles application of fixes to source files.
type FixApplier struct {
	rootDir       string
	backupEnabled bool
	backupDir     string
	backups       map[string]string // original -> backup path
	backupsMu     sync.Mutex
}

// NewFixApplier creates a new FixApplier.
func NewFixApplier(rootDir string) *FixApplier {
	//nolint:exhaustruct
	return &FixApplier{
		rootDir:       rootDir,
		backupEnabled: true,
		backupDir:     filepath.Join(os.TempDir(), "go-finding-backups"),
		backups:       make(map[string]string),
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
			_ = a.rollbackAll(modified)

			return applied, fmt.Errorf("fix application cancelled: %w", ctx.Err())
		default:
		}

		// Create backup
		if a.backupEnabled {
			err := a.backup(path)
			if err != nil {
				_ = a.rollbackAll(modified)

				return applied, finding.NewIOError("backup "+path, err)
			}
		}

		// Apply fixes
		count, err := a.applyToFile(path, fileFixes)
		if err != nil {
			// Restore current file from backup
			if a.backupEnabled {
				_ = a.restore(path)
			}

			// Restore all previously modified files
			_ = a.rollbackAll(modified)

			return applied, finding.NewConflictError("apply to "+path, err)
		}

		modified = append(modified, path)
		applied += count
	}

	return applied, nil
}

// rollbackAll restores all modified files from their backups.
func (a *FixApplier) rollbackAll(paths []string) error {
	var errs []error

	for _, p := range paths {
		if err := a.restore(p); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// fileHash returns the hex-encoded FNV-1a 128-bit hash of s.
func fileHash(s string) string {
	h := fnv.New128a()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// backup creates a backup of the given file.
func (a *FixApplier) backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return ioErrorAt("read file for backup", err, path)
	}

	backupPath := filepath.Join(a.backupDir, fmt.Sprintf("%x.bak", fileHash(path)))
	if err := os.MkdirAll(a.backupDir, 0o750); err != nil {
		return finding.NewIOError("create backup dir", err)
	}

	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return ioErrorAt("write backup", err, path)
	}

	a.backupsMu.Lock()
	a.backups[path] = backupPath
	a.backupsMu.Unlock()

	return nil
}

// restore restores a file from its backup.
func (a *FixApplier) restore(path string) error {
	backupPath, ok := func() (string, bool) {
		a.backupsMu.Lock()
		defer a.backupsMu.Unlock()

		p, exists := a.backups[path]

		return p, exists
	}()
	if !ok {
		return finding.NewInternalError("no backup for "+path, nil)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return ioErrorAt("read backup", err, path)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return ioErrorAt("restore file", err, path)
	}

	return nil
}

// applyToFile applies fixes to a single file.
// When a finding has a Range with valid end position, it uses line-based replacement
// targeting the exact line range. Otherwise it falls back to string replacement.
// Fixes are sorted descending by position so earlier replacements don't shift later ones.
func (*FixApplier) applyToFile(
	path string,
	fixes []finding.Finding,
) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, ioErrorAt("read file", err, path)
	}

	lines := strings.Split(string(content), "\n")
	applied := 0

	// Partition: range-based fixes (apply first, sorted descending) vs string-based.
	var rangeFixes, stringFixes []finding.Finding

	for _, f := range fixes {
		if f.BeforeCode == "" && f.AfterCode == "" {
			continue
		}

		if f.Range != nil && f.Range.HasEnd() && f.Range.Start.Line > 0 && f.Range.End.Line > 0 {
			rangeFixes = append(rangeFixes, f)
		} else if f.BeforeCode != "" && f.AfterCode != "" {
			stringFixes = append(stringFixes, f)
		}
	}

	// Sort range fixes descending so earlier edits don't shift later line numbers.
	slices.SortFunc(rangeFixes, func(a, b finding.Finding) int {
		if a.Range.Start.Line != b.Range.Start.Line {
			return b.Range.Start.Line - a.Range.Start.Line
		}

		return b.Range.Start.Column - a.Range.Start.Column
	})

	// Apply range-based fixes.
	for _, f := range rangeFixes {
		startIdx := f.Range.Start.Line - 1 // 0-indexed
		endIdx := f.Range.End.Line - 1

		if startIdx < 0 || startIdx >= len(lines) {
			continue
		}

		if endIdx >= len(lines) {
			endIdx = len(lines) - 1
		}

		// Verify BeforeCode is present in the range if set.
		rangeContent := strings.Join(lines[startIdx:endIdx+1], "\n")
		if f.BeforeCode != "" {
			if !strings.Contains(rangeContent, f.BeforeCode) {
				continue
			}

			// Targeted replacement: swap BeforeCode→AfterCode within the range,
			// preserving surrounding content like indentation.
			replaced := strings.Replace(rangeContent, f.BeforeCode, f.AfterCode, 1)
			replacementLines := strings.Split(replaced, "\n")
			newLines := make([]string, 0, len(lines)-(endIdx-startIdx+1)+len(replacementLines))
			newLines = append(newLines, lines[:startIdx]...)
			newLines = append(newLines, replacementLines...)
			newLines = append(newLines, lines[endIdx+1:]...)
			lines = newLines
		} else if f.AfterCode != "" {
			// Full replacement: replace entire line range with AfterCode.
			replacement := make([]string, 0, len(lines)-(endIdx-startIdx+1)+1)
			replacement = append(replacement, lines[:startIdx]...)
			replacement = append(replacement, f.AfterCode)
			replacement = append(replacement, lines[endIdx+1:]...)
			lines = replacement
		}

		applied++
	}

	// Apply string-based fixes (fallback, position-independent).
	if len(stringFixes) > 0 {
		joined := strings.Join(lines, "\n")

		for _, f := range stringFixes {
			newContent := strings.Replace(joined, f.BeforeCode, f.AfterCode, 1)
			if newContent != joined {
				joined = newContent
				applied++
			}
		}

		lines = strings.Split(joined, "\n")
	}

	if applied == 0 {
		return 0, nil
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return 0, ioErrorAt("write file", err, path)
	}

	return applied, nil
}
