package finding

import (
	"slices"
)

// FilterFunc is a predicate for filtering findings.
type FilterFunc func(Finding) bool

// Filter returns findings that match all predicates.
func Filter(findings []Finding, predicates ...FilterFunc) []Finding {
	result := make([]Finding, 0, len(findings))

	for _, finding := range findings {
		match := true

		for _, p := range predicates {
			if !p(finding) {
				match = false

				break
			}
		}

		if match {
			result = append(result, finding)
		}
	}

	return result
}

// BySeverity returns a filter for the given severity.
func BySeverity(sev Severity) FilterFunc {
	return func(f Finding) bool {
		return f.Severity == sev
	}
}

// BySeverityAtLeast returns a filter for severity >= the given level.
// Findings with invalid severity are excluded (return false).
func BySeverityAtLeast(sev Severity) FilterFunc {
	return func(f Finding) bool {
		return f.Severity.GreaterThanOrEqual(sev)
	}
}

// ByCategory returns a filter for the given category.
func ByCategory(cat Category) FilterFunc {
	return func(f Finding) bool {
		return f.Category == cat
	}
}

// ByFixStrategy returns a filter for the given fix strategy.
func ByFixStrategy(fs FixStrategy) FilterFunc {
	return func(f Finding) bool {
		return f.FixStrategy == fs
	}
}

// ByTool returns a filter for the given tool name.
func ByTool(tool string) FilterFunc {
	return func(f Finding) bool {
		return f.ToolName == tool
	}
}

// ByRule returns a filter for the given rule.
func ByRule(rule string) FilterFunc {
	return func(f Finding) bool {
		return f.Rule == rule
	}
}

// ByFile returns a filter for findings in the given file.
func ByFile(file string) FilterFunc {
	return func(f Finding) bool {
		return f.Position.File == file
	}
}

// NotSuppressed returns a filter for non-suppressed findings.
func NotSuppressed(f Finding) bool {
	return !f.IsSuppressed()
}

// HasFix returns a filter for findings with fixes.
func HasFix(f Finding) bool {
	return f.HasFix()
}

// HasSuggestion returns a filter for findings with suggestions.
func HasSuggestion(f Finding) bool {
	return f.HasSuggestion()
}

// GroupBy groups findings by a key extractor function.
func GroupBy(findings []Finding, keyFn func(Finding) string) map[string][]Finding {
	groups := make(map[string][]Finding)

	for _, finding := range findings {
		key := keyFn(finding)
		groups[key] = append(groups[key], finding)
	}

	return groups
}

// GroupByFile groups findings by file path.
func GroupByFile(findings []Finding) map[string][]Finding {
	return GroupBy(findings, func(finding Finding) string {
		return finding.Position.File
	})
}

// GroupBySeverity groups findings by severity.
func GroupBySeverity(findings []Finding) map[Severity][]Finding {
	groups := make(map[Severity][]Finding)
	for _, finding := range findings {
		groups[finding.Severity] = append(groups[finding.Severity], finding)
	}

	return groups
}

// GroupByCategory groups findings by category.
func GroupByCategory(findings []Finding) map[Category][]Finding {
	groups := make(map[Category][]Finding)

	for _, f := range findings {
		groups[f.Category] = append(groups[f.Category], f)
	}

	return groups
}

// SortByPosition sorts findings by file path, then line, then column.
func SortByPosition(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int {
		return a.Position.Compare(b.Position)
	})
}

// SortBySeverity sorts findings by severity (most severe first).
func SortBySeverity(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int {
		return -a.Severity.Compare(b.Severity)
	})
}
