package finding

// FilterFunc is a predicate for filtering findings.
type FilterFunc func(Finding) bool

// Filter returns findings that match all predicates.
func Filter(findings []Finding, predicates ...FilterFunc) []Finding {
	var result []Finding

	for _, f := range findings {
		match := true

		for _, p := range predicates {
			if !p(f) {
				match = false

				break
			}
		}

		if match {
			result = append(result, f)
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
func BySeverityAtLeast(sev Severity) FilterFunc {
	return func(f Finding) bool {
		return f.Severity.GreaterThan(sev) || f.Severity == sev
	}
}

// ByCategory returns a filter for the given category.
func ByCategory(cat string) FilterFunc {
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

	for _, f := range findings {
		key := keyFn(f)
		groups[key] = append(groups[key], f)
	}

	return groups
}

// GroupByFile groups findings by file path.
func GroupByFile(findings []Finding) map[string][]Finding {
	return GroupBy(findings, func(f Finding) string {
		return f.Position.File
	})
}

// GroupBySeverity groups findings by severity.
func GroupBySeverity(findings []Finding) map[Severity][]Finding {
	groups := make(map[Severity][]Finding)
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}

	return groups
}

// GroupByCategory groups findings by category.
func GroupByCategory(findings []Finding) map[string][]Finding {
	return GroupBy(findings, func(f Finding) string {
		return f.Category
	})
}
