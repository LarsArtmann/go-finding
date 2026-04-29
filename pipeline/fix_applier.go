package pipeline

import (
	"context"
	"fmt"
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
}

// NewFixApplier creates a new FixApplier.
func NewFixApplier(rootDir string) *FixApplier {
	return &FixApplier{
		rootDir: rootDir,
		backup:  NewFileBackup(filepath.Join(os.TempDir(), "go-finding-backups")),
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
func (*FixApplier) applyToFile(path string, fixes []finding.Finding) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, ioErrorAt("read file", err, path)
	}

	lines := strings.Split(string(content), "\n")
	rangeFixes, stringFixes := partitionFixes(fixes)

	lines, applied := applyRangeFixes(lines, rangeFixes)
	lines, strApplied := applyStringFixes(lines, stringFixes)
	applied += strApplied

	if applied == 0 {
		return 0, nil
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write in fix applier
		path,
		[]byte(strings.Join(lines, "\n")),
		0o600,
	); err != nil {
		return 0, ioErrorAt("write file", err, path)
	}

	return applied, nil
}

// partitionFixes splits fixes into range-based and string-based categories.
func partitionFixes(fixes []finding.Finding) ([]finding.Finding, []finding.Finding) {
	var rangeFixes, stringFixes []finding.Finding

	for _, f := range fixes {
		if f.BeforeCode == "" && f.AfterCode == "" {
			continue
		}

		if f.Range != nil && f.Range.HasEnd() && f.Range.Start.Line > 0 && f.Range.End.Line > 0 {
			rangeFixes = append(rangeFixes, f)
		} else if f.BeforeCode != "" || f.AfterCode != "" {
			stringFixes = append(stringFixes, f)
		}
	}

	return rangeFixes, stringFixes
}

// applyRangeFixes applies line-range replacements and returns the updated lines.
func applyRangeFixes(lines []string, fixes []finding.Finding) ([]string, int) {
	slices.SortFunc(fixes, func(a, b finding.Finding) int {
		if a.Range.Start.Line != b.Range.Start.Line {
			return b.Range.Start.Line - a.Range.Start.Line
		}

		return b.Range.Start.Column - a.Range.Start.Column
	})

	applied := 0

	for _, f := range fixes {
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

	return lines, applied
}

// applyStringFixes applies fallback string replacements and insertions.
func applyStringFixes(lines []string, fixes []finding.Finding) ([]string, int) {
	joined := strings.Join(lines, "\n")
	joinedChanged := false
	applied := 0

	for _, f := range fixes {
		if f.BeforeCode == "" {
			// Insertion: place AfterCode at the finding's line.
			lineIdx := f.Position.Line - 1
			if lineIdx >= 0 && lineIdx <= len(lines) {
				lines = slices.Insert(lines, lineIdx, f.AfterCode)
				applied++
			}

			continue
		}

		newContent := replaceNearestToLine(joined, f.BeforeCode, f.AfterCode, f.Position.Line)
		if newContent != joined {
			joined = newContent
			joinedChanged = true
			applied++
		}
	}

	if joinedChanged {
		lines = strings.Split(joined, "\n")
	}

	return lines, applied
}

// replaceNearestToLine replaces the occurrence of old nearest to the given line number.
// If targetLine is 0 or no line bias can be determined, replaces the first occurrence.
func replaceNearestToLine(content, old, replacement string, targetLine int) string {
	idx := strings.Index(content, old)
	if idx < 0 {
		return content
	}

	// If no line info, use first occurrence.
	if targetLine <= 0 {
		return strings.Replace(content, old, replacement, 1)
	}

	// Find all occurrences and pick the one nearest to targetLine.
	best := idx
	bestDist := lineDistance(content, idx, targetLine)

	for {
		next := strings.Index(content[idx+len(old):], old)
		if next < 0 {
			break
		}

		idx = idx + len(old) + next
		dist := lineDistance(content, idx, targetLine)
		if dist < bestDist {
			bestDist = dist
			best = idx
		}
	}

	return content[:best] + replacement + content[best+len(old):]
}

// lineDistance counts how many newlines appear before position pos in content,
// then returns the absolute difference from targetLine.
func lineDistance(content string, pos int, targetLine int) int {
	line := 1
	for i := 0; i < pos && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}

	diff := line - targetLine
	if diff < 0 {
		return -diff
	}

	return diff
}
