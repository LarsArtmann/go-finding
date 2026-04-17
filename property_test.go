package finding

import (
	"math/rand"
	"testing"
	"testing/quick"
)

func TestProperty_FilterBySeverityAtLeast(t *testing.T) {
	t.Parallel()
	property := func(sev uint8) bool {
		if sev > 3 {
			return true
		}
		findings := []Finding{
			{Severity: SeverityInfo},
			{Severity: SeverityWarning},
			{Severity: SeverityError},
			{Severity: SeverityCritical},
		}
		filtered := Filter(findings, BySeverityAtLeast(sevFromInt(int(sev))))
		for _, f := range filtered {
			if f.Severity.LessThan(sevFromInt(int(sev))) {
				return false
			}
		}
		return true
	}
	checkProperty(t, property)
}

func TestProperty_GroupByFileCoversAll(t *testing.T) {
	t.Parallel()
	property := func(files []string) bool {
		findings := make([]Finding, len(files))
		for i, f := range files {
			findings[i] = Finding{Position: Position{File: f}}
		}
		groups := GroupByFile(findings)
		if len(groups) == 0 && len(findings) > 0 {
			return false
		}
		total := 0
		for _, g := range groups {
			total += len(g)
		}
		return total == len(findings)
	}
	checkProperty(t, property)
}

func TestProperty_MergePreservesAll(t *testing.T) {
	t.Parallel()
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))
		n := rng.Intn(50) + 1
		findings := make([]Finding, n)
		for i := range findings {
			findings[i] = Finding{
				ID: randomSeedRule(
					rng,
				) + ":" + randomSeedFile(
					rng,
				) + ":" + string(
					rune('A'+rng.Intn(26)),
				),
				Severity: sevFromInt(rng.Intn(4)),
			}
		}
		reports := splitSeedFindings(findings, rng)
		merged := Merge(reports, WithDeduplication(false))
		return len(merged.Findings) == len(findings)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_IDRoundTrip(t *testing.T) {
	t.Parallel()
	property := func(tool, rule, file string, line, col uint16) bool {
		if tool == "" || rule == "" || file == "" {
			return true
		}
		id := GenerateID(tool, rule, Position{File: file, Line: int(line), Column: int(col)})
		parsedTool, parsedRule, _, parsedLine, parsedCol, ok := ParseID(id)
		if !ok {
			return false
		}
		if parsedTool != tool {
			return false
		}
		if parsedRule != rule {
			return false
		}
		if int(line) > 0 && parsedLine != int(line) {
			return false
		}
		if int(col) > 0 && parsedCol != int(col) {
			return false
		}
		return true
	}
	if err := quick.Check(property, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_MergeDedupReducesOrPreserves(t *testing.T) {
	t.Parallel()
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		findings := make([]Finding, rng.Intn(20)+2)
		for i := range findings {
			findings[i] = Finding{
				ID:       "dupe-id",
				Severity: sevFromInt(rng.Intn(4)),
			}
		}

		report := NewReport(ToolInfo{Name: "test"})
		for _, f := range findings {
			report.AddFinding(f)
		}

		merged := Merge([]*Report{report, report}, WithDeduplication(true))
		return len(merged.Findings) <= len(findings)*2
	}
	if err := quick.Check(property, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_FilterEmptyReturnsEmpty(t *testing.T) {
	t.Parallel()
	property := func() bool {
		result := Filter(nil, BySeverity(SeverityError))
		return len(result) == 0
	}
	if err := quick.Check(property, nil); err != nil {
		t.Error(err)
	}
}
