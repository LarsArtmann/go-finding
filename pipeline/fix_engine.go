package pipeline

import (
	"slices"
	"strings"

	"github.com/larsartmann/go-finding"
)

// FixEngine applies code transformations purely in memory.
// It has no file-system or backup dependencies.
type FixEngine struct{}

// NewFixEngine creates a new FixEngine.
func NewFixEngine() *FixEngine {
	return &FixEngine{}
}

// Apply runs all applicable fixes against the given lines and returns the
// modified lines together with the number of fixes successfully applied.
func (*FixEngine) Apply(lines []string, fixes []finding.Finding) ([]string, int) {
	rangeFixes, stringFixes := partitionFixes(fixes)

	lines, applied := applyRangeFixes(lines, rangeFixes)
	lines, strApplied := applyStringFixes(lines, stringFixes)

	return lines, applied + strApplied
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
