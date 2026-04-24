package pipeline

import (
	"slices"

	"github.com/larsartmann/go-finding"
)

// FixGroup represents fixes that can be safely applied together.
// Fixes in the same group don't conflict with each other.
type FixGroup struct {
	File  string
	Fixes []finding.Finding
	// Bounds is the combined range covering all fixes in this group
	Bounds finding.Range
}

// newFixGroup creates a FixGroup with a single fix.
func newFixGroup(file string, f finding.Finding, bounds finding.Range) FixGroup {
	return FixGroup{
		File:   file,
		Fixes:  []finding.Finding{f},
		Bounds: bounds,
	}
}

// ConflictDetector identifies conflicting fixes.
type ConflictDetector struct{}

// NewConflictDetector creates a new conflict detector.
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{}
}

// DetectConflicts analyzes fixes and returns groups of non-conflicting fixes
// along with any conflicting fixes that couldn't be grouped.
func (c *ConflictDetector) DetectConflicts(
	fixes []finding.Finding,
) ([]FixGroup, []finding.Finding) {
	// Group by file first
	byFile := make(map[string][]finding.Finding)

	for _, f := range fixes {
		file := f.Position.File
		if file == "" {
			continue // Skip fixes without file info
		}

		byFile[file] = append(byFile[file], f)
	}

	var (
		groups    = make([]FixGroup, 0, len(byFile))
		conflicts []finding.Finding
	)

	for file, fileFixes := range byFile {
		fileGroups, fileConflicts := c.detectConflictsInFile(file, fileFixes)
		groups = append(groups, fileGroups...)
		conflicts = append(conflicts, fileConflicts...)
	}

	return groups, conflicts
}

// detectConflictsInFile analyzes fixes within a single file.
func (c *ConflictDetector) detectConflictsInFile(
	file string,
	fixes []finding.Finding,
) ([]FixGroup, []finding.Finding) {
	if len(fixes) == 0 {
		return nil, nil
	}

	// Sort fixes by start position
	sorted := make([]finding.Finding, len(fixes))
	copy(sorted, fixes)
	slices.SortFunc(sorted, func(a, b finding.Finding) int {
		return a.Position.Compare(b.Position)
	})

	var (
		groups       []FixGroup
		currentGroup FixGroup
		conflicts    []finding.Finding
	)

	for _, f := range sorted {
		rangeInfo := c.getFindingRange(f)

		if len(currentGroup.Fixes) == 0 {
			currentGroup = newFixGroup(file, f, rangeInfo)
		} else {
			// Check if this fix conflicts with current group
			if currentGroup.Bounds.Overlaps(rangeInfo) {
				// Conflict detected - add to current group and extend bounds
				currentGroup.Fixes = append(currentGroup.Fixes, f)
				currentGroup.Bounds = extendRange(currentGroup.Bounds, rangeInfo)
			} else {
				// No conflict - finalize current group and start new one
				groups = append(groups, currentGroup)
				currentGroup = newFixGroup(file, f, rangeInfo)
			}
		}
	}

	// Don't forget the last group
	if len(currentGroup.Fixes) > 0 {
		groups = append(groups, currentGroup)
	}

	// Split groups with multiple fixes into individual fixes as conflicts
	// (for now - we could implement merge logic later)
	var finalGroups []FixGroup

	for _, g := range groups {
		if len(g.Fixes) == 1 {
			finalGroups = append(finalGroups, g)
		} else {
			// Group has conflicts - only keep the first fix, mark others as conflicts
			finalGroups = append(finalGroups, FixGroup{
				File:   g.File,
				Fixes:  []finding.Finding{g.Fixes[0]},
				Bounds: c.getFindingRange(g.Fixes[0]),
			})
			conflicts = append(conflicts, g.Fixes[1:]...)
		}
	}

	return finalGroups, conflicts
}

// getFindingRange extracts the range for a finding.
// Falls back to a single position if no range is specified.
func (*ConflictDetector) getFindingRange(
	f finding.Finding,
) finding.Range {
	if f.Range != nil && f.Range.IsValid() {
		return *f.Range
	}

	// Create a single-position range
	return finding.Range{
		Start: f.Position,
		End:   finding.Position{}, //nolint:exhaustruct // Empty end means single position
	}
}

// extendRange returns a range that covers both input ranges.
func extendRange(r1, r2 finding.Range) finding.Range {
	result := r1

	// Extend start if r2 starts earlier
	if r2.Start.Compare(result.Start) < 0 {
		result.Start = r2.Start
	}

	// Determine effective end positions (single-point ranges end at their start)
	r1End := r1.End
	if r1End.Line == 0 {
		r1End = r1.Start
	}

	r2End := r2.End
	if r2End.Line == 0 {
		r2End = r2.Start
	}

	// Extend end if r2 ends later
	if r2End.Compare(r1End) > 0 {
		result.End = r2End
	}

	return result
}

// FilterConflictingFixes returns only non-conflicting fixes.
func FilterConflictingFixes(fixes []finding.Finding) []finding.Finding {
	detector := NewConflictDetector()
	groups, _ := detector.DetectConflicts(fixes)

	result := make([]finding.Finding, 0, len(groups))
	for _, g := range groups {
		result = append(result, g.Fixes...)
	}

	return result
}

// ConflictInfo provides detailed information about conflicts.
type ConflictInfo struct {
	Finding       finding.Finding
	ConflictsWith []finding.Finding
	Reason        string
}

// AnalyzeConflicts provides detailed conflict information.
func AnalyzeConflicts(fixes []finding.Finding) []ConflictInfo {
	detector := NewConflictDetector()
	groups, conflictingFixes := detector.DetectConflicts(fixes)

	result := make([]ConflictInfo, 0, len(conflictingFixes))

	// For each conflicting fix, find what it conflicts with
	for _, cf := range conflictingFixes {
		cfRange := detector.getFindingRange(cf)

		var conflictsWith []finding.Finding

		for _, g := range groups {
			if g.Bounds.Overlaps(cfRange) {
				conflictsWith = append(conflictsWith, g.Fixes...)
			}
		}

		result = append(result, ConflictInfo{
			Finding:       cf,
			ConflictsWith: conflictsWith,
			Reason:        "overlapping range",
		})
	}

	return result
}
