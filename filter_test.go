package finding

import "testing"

func TestFilter_Empty(t *testing.T) {
	t.Parallel()

	result := Filter(nil, BySeverity(SeverityError))
	if len(result) != 0 {
		t.Errorf("Filter(nil) = %d, want 0", len(result))
	}
}

func TestFilter_NoPredicates(t *testing.T) {
	t.Parallel()

	findings := []Finding{{ID: "1"}, {ID: "2"}}
	result := Filter(findings)
	if len(result) != 2 {
		t.Errorf("Filter with no predicates = %d, want 2", len(result))
	}
}

func TestFilter_MultiplePredicates(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError, ToolName: "tool1"},
		{ID: "2", Severity: SeverityError, ToolName: "tool2"},
		{ID: "3", Severity: SeverityWarning, ToolName: "tool1"},
	}

	result := Filter(findings, BySeverity(SeverityError), ByTool("tool1"))
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("Filter(Severity=error, Tool=tool1) = %v, want [1]", result)
	}
}

func TestBySeverity(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityError},
	}

	result := Filter(findings, BySeverity(SeverityError))
	if len(result) != 2 {
		t.Errorf("BySeverity(error) = %d, want 2", len(result))
	}
}

func TestBySeverityAtLeast(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityInfo},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityError},
		{ID: "4", Severity: SeverityCritical},
	}

	result := Filter(findings, BySeverityAtLeast(SeverityWarning))
	if len(result) != 3 {
		t.Errorf("BySeverityAtLeast(warning) = %d, want 3", len(result))
	}
}

func TestByCategory(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Category: CategorySecurity},
		{ID: "2", Category: CategoryStyle},
		{ID: "3", Category: CategorySecurity},
	}

	result := Filter(findings, ByCategory(CategorySecurity))
	if len(result) != 2 {
		t.Errorf("ByCategory(security) = %d, want 2", len(result))
	}
}

func TestByFixStrategy(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", FixStrategy: FixStrategyDirect},
		{ID: "2", FixStrategy: FixStrategyNone},
		{ID: "3", FixStrategy: FixStrategyDirect},
	}

	result := Filter(findings, ByFixStrategy(FixStrategyDirect))
	if len(result) != 2 {
		t.Errorf("ByFixStrategy(direct) = %d, want 2", len(result))
	}
}

func TestByTool(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "govet"},
		{ID: "2", ToolName: "staticcheck"},
	}

	result := Filter(findings, ByTool("govet"))
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("ByTool(govet) = %v, want [1]", result)
	}
}

func TestByRule(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Rule: "nilcheck"},
		{ID: "2", Rule: "unused"},
	}

	result := Filter(findings, ByRule("nilcheck"))
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("ByRule(nilcheck) = %v, want [1]", result)
	}
}

func TestByFile(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Position: Position{File: "a.go"}},
		{ID: "2", Position: Position{File: "b.go"}},
	}

	result := Filter(findings, ByFile("a.go"))
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("ByFile(a.go) = %v, want [1]", result)
	}
}

func TestNotSuppressed(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1"},
		{ID: "2", Suppression: &Suppression{Kind: SuppressionInSource}},
	}

	result := Filter(findings, NotSuppressed)
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("NotSuppressed = %v, want [1]", result)
	}
}

func TestHasFix(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", FixStrategy: FixStrategyDirect},
		{ID: "2", FixStrategy: FixStrategyNone},
		{ID: "3", FixStrategy: FixStrategyAI},
	}

	result := Filter(findings, HasFix)
	if len(result) != 2 {
		t.Errorf("HasFix = %d, want 2", len(result))
	}
}

func TestHasSuggestion(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Suggestion: "fix it"},
		{ID: "2", BeforeCode: "old", AfterCode: "new"},
		{ID: "3"},
	}

	result := Filter(findings, HasSuggestion)
	if len(result) != 2 {
		t.Errorf("HasSuggestion = %d, want 2", len(result))
	}
}

func TestGroupBy(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", ToolName: "a"},
		{ID: "2", ToolName: "b"},
		{ID: "3", ToolName: "a"},
	}

	groups := GroupBy(findings, func(f Finding) string { return f.ToolName })
	if len(groups["a"]) != 2 {
		t.Errorf("GroupBy tool 'a' = %d, want 2", len(groups["a"]))
	}
	if len(groups["b"]) != 1 {
		t.Errorf("GroupBy tool 'b' = %d, want 1", len(groups["b"]))
	}
}

func TestGroupBy_Empty(t *testing.T) {
	t.Parallel()

	groups := GroupBy(nil, func(f Finding) string { return f.ToolName })
	if len(groups) != 0 {
		t.Errorf("GroupBy(nil) = %d groups, want 0", len(groups))
	}
}

func TestGroupByFile(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Position: Position{File: "a.go"}},
		{ID: "2", Position: Position{File: "b.go"}},
		{ID: "3", Position: Position{File: "a.go"}},
	}

	groups := GroupByFile(findings)
	if len(groups["a.go"]) != 2 {
		t.Errorf("GroupByFile 'a.go' = %d, want 2", len(groups["a.go"]))
	}
}

func TestGroupBySeverity(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityError},
	}

	groups := GroupBySeverity(findings)
	if len(groups[SeverityError]) != 2 {
		t.Errorf("GroupBySeverity error = %d, want 2", len(groups[SeverityError]))
	}
	if len(groups[SeverityWarning]) != 1 {
		t.Errorf("GroupBySeverity warning = %d, want 1", len(groups[SeverityWarning]))
	}
}

func TestGroupByCategory(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{ID: "1", Category: CategorySecurity},
		{ID: "2", Category: CategoryStyle},
		{ID: "3", Category: CategorySecurity},
	}

	groups := GroupByCategory(findings)
	if len(groups[CategorySecurity]) != 2 {
		t.Errorf("GroupByCategory security = %d, want 2", len(groups[CategorySecurity]))
	}
}
