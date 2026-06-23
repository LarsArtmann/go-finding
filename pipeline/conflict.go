package pipeline

import (
	"slices"

	"github.com/larsartmann/go-finding"
)

// Conflict reason constants.
const (
	ReasonOverlappingRange = "overlapping range"
	ReasonOverlappingEdit  = "overlapping edit"
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

// DetectConflicts analyzes fixes and returns groups of non-conflicting fixes
// along with any conflicting fixes that couldn't be grouped.
func DetectConflicts(fixes []finding.Finding) ([]FixGroup, []finding.Finding) {
	byFile := make(map[string][]finding.Finding)

	for _, f := range fixes {
		file := f.Position.File
		if file == "" {
			continue
		}

		byFile[file] = append(byFile[file], f)
	}

	var (
		groups    = make([]FixGroup, 0, len(byFile))
		conflicts []finding.Finding
	)
	for file, fileFixes := range byFile {
		fileGroups, fileConflicts := detectConflictsInFile(file, fileFixes)
		groups = append(groups, fileGroups...)
		conflicts = append(conflicts, fileConflicts...)
	}

	return groups, conflicts
}

// detectConflictsInFile analyzes fixes within a single file.
// Findings are sorted by position and grouped transitively: if A overlaps B
// and B overlaps C (even if A doesn't overlap C), all three join the same group.
// Groups with multiple findings keep only the first; the rest are marked as conflicts.
// This conservative strategy ensures safe application order.
func detectConflictsInFile(
	file string,
	fixes []finding.Finding,
) ([]FixGroup, []finding.Finding) {
	if len(fixes) == 0 {
		return nil, nil
	}

	sorted := make([]finding.Finding, len(fixes))
	copy(sorted, fixes)
	slices.SortFunc(sorted, func(a, b finding.Finding) int {
		return a.Position.Compare(b.Position)
	})

	var (
		groups       []FixGroup
		currentGroup FixGroup
	)

	for _, f := range sorted {
		rangeInfo := getFindingRange(f)

		if len(currentGroup.Fixes) == 0 {
			currentGroup = newFixGroup(file, f, rangeInfo)
		} else {
			overlapsAny := false

			for _, existing := range currentGroup.Fixes {
				if getFindingRange(existing).Overlaps(rangeInfo) {
					overlapsAny = true

					break
				}
			}

			if overlapsAny {
				currentGroup.Fixes = append(currentGroup.Fixes, f)
				currentGroup.Bounds = extendRange(currentGroup.Bounds, rangeInfo)
			} else {
				groups = append(groups, currentGroup)
				currentGroup = newFixGroup(file, f, rangeInfo)
			}
		}
	}

	if len(currentGroup.Fixes) > 0 {
		groups = append(groups, currentGroup)
	}

	var (
		finalGroups []FixGroup
		conflicts   []finding.Finding
	)

	for _, g := range groups {
		if len(g.Fixes) == 1 {
			finalGroups = append(finalGroups, g)
		} else {
			finalGroups = append(finalGroups, FixGroup{
				File:   g.File,
				Fixes:  []finding.Finding{g.Fixes[0]},
				Bounds: getFindingRange(g.Fixes[0]),
			})
			conflicts = append(conflicts, g.Fixes[1:]...)
		}
	}

	return finalGroups, conflicts
}

// getFindingRange extracts the range for a finding.
// Falls back to a single position if no range is specified.
func getFindingRange(
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
	groups, _ := DetectConflicts(fixes)

	result := make([]finding.Finding, 0, len(groups))
	for _, g := range groups {
		result = append(result, g.Fixes...)
	}

	return result
}

// FilterConflictingEdits resolves findings to byte-level edits using the engine,
// then filters overlapping edits at the byte level. This is more precise than
// FilterConflictingFixes because it operates on actual byte offsets rather than
// line/column ranges.
// Returns the non-conflicting findings and any provider errors from edit resolution.
func FilterConflictingEdits(
	content []byte,
	fixes []finding.Finding,
	engine *FixEngine,
) ([]finding.Finding, []error) {
	_, _, conflicts, _, providerErrors := engine.ApplyWithConflicts(content, fixes) //nolint:dogsled
	if len(conflicts) == 0 && len(providerErrors) == 0 {
		return fixes, nil
	}

	conflictIDs := make(map[string]struct{}, len(conflicts))
	for _, c := range conflicts {
		key := c.Finding.Key()
		conflictIDs[key] = struct{}{}
	}

	result := make([]finding.Finding, 0, len(fixes)-len(conflicts))
	for _, f := range fixes {
		if _, isConflict := conflictIDs[f.Key()]; !isConflict {
			result = append(result, f)
		}
	}

	return result, providerErrors
}

// Conflict provides detailed information about conflicts.
type Conflict struct {
	Finding       finding.Finding
	ConflictsWith []finding.Finding
	Reason        string
}

// AnalyzeConflicts provides detailed conflict information.
func AnalyzeConflicts(fixes []finding.Finding) []Conflict {
	groups, conflictingFixes := DetectConflicts(fixes)

	result := make([]Conflict, 0, len(conflictingFixes))

	for _, cf := range conflictingFixes {
		cfRange := getFindingRange(cf)

		var conflictsWith []finding.Finding

		for _, g := range groups {
			if g.Bounds.Overlaps(cfRange) {
				conflictsWith = append(conflictsWith, g.Fixes...)
			}
		}

		result = append(result, Conflict{
			Finding:       cf,
			ConflictsWith: conflictsWith,
			Reason:        ReasonOverlappingRange,
		})
	}

	return result
}
