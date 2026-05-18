package finding

import "slices"

// DiffResult holds the difference between two finding sets.
type DiffResult struct {
	Added     []Finding // Present in "after" but not "before"
	Removed   []Finding // Present in "before" but not "after"
	Unchanged []Finding // Present in both (by ID)
}

func byFindingID(a, b Finding) int {
	if a.ID < b.ID {
		return -1
	}
	if a.ID > b.ID {
		return 1
	}
	return 0
}

// Diff compares two finding sets by ID and categorizes them as added, removed, or unchanged.
// Both slices are sorted by ID in the result.
func Diff(before, after []Finding) DiffResult {
	beforeSet := make(map[string]Finding, len(before))
	for _, f := range before {
		beforeSet[f.ID] = f
	}

	afterSet := make(map[string]Finding, len(after))
	for _, f := range after {
		afterSet[f.ID] = f
	}

	var added, removed, unchanged []Finding

	for id, f := range beforeSet {
		if _, exists := afterSet[id]; !exists {
			removed = append(removed, f)
		} else {
			unchanged = append(unchanged, f)
		}
	}

	for id, f := range afterSet {
		if _, exists := beforeSet[id]; !exists {
			added = append(added, f)
		}
	}

	slices.SortFunc(added, byFindingID)
	slices.SortFunc(removed, byFindingID)
	slices.SortFunc(unchanged, byFindingID)

	return DiffResult{Added: added, Removed: removed, Unchanged: unchanged}
}
