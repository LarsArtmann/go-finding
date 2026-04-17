package finding

import (
	"math/rand"
	"testing"
)

func FuzzMergeRandom(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed int64, dedup bool, dedupBy int) {
		rng := rand.New(rand.NewSource(seed))

		n := rng.Intn(20) + 1
		findings := make([]Finding, n)
		for i := range findings {
			findings[i] = Finding{
				ID:       randomSeedID(rng),
				Rule:     randomSeedRule(rng),
				Severity: randomSeverity(rng),
				Position: Position{
					File:   randomSeedFile(rng),
					Line:   rng.Intn(1000),
					Column: rng.Intn(200),
				},
				ToolName: randomSeedTool(rng),
			}
		}

		reports := splitSeedFindings(findings, rng)

		var opts []MergeOption
		opts = append(opts, WithDeduplication(dedup))
		if dedupBy >= 0 && dedupBy <= 2 {
			opts = append(opts, WithDeduplicateBy(DeduplicateBy(dedupBy)))
		}

		merged := Merge(reports, opts...)
		if merged == nil {
			t.Fatal("Merge returned nil")
		}

		if !dedup && len(merged.Findings) != len(findings) {
			t.Fatalf(
				"merge without dedup: got %d findings, want %d",
				len(merged.Findings),
				len(findings),
			)
		}
	})
}

func FuzzFilterBySeverity(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed int64, minSev int) {
		if minSev < 0 || minSev > 3 {
			t.Skip()
		}

		rng := rand.New(rand.NewSource(seed))
		n := rng.Intn(50) + 1
		findings := make([]Finding, n)
		for i := range findings {
			findings[i] = Finding{
				Severity: randomSeverity(rng),
				Category: randomSeedCategory(rng),
			}
		}

		filtered := Filter(findings, BySeverityAtLeast(sevFromInt(minSev)))
		for _, f := range filtered {
			if f.Severity.LessThan(sevFromInt(minSev)) {
				t.Fatalf("filtered finding has severity %v < %v", f.Severity, sevFromInt(minSev))
			}
		}
	})
}

func randomSeedID(rng *rand.Rand) string {
	return randomSeedTool(rng) + ":" + randomSeedRule(rng) + ":" + randomSeedFile(rng)
}

func randomSeedRule(rng *rand.Rand) string {
	rules := []string{"SA1000", "S1001", "ST1000", "QF1001", "U1000"}
	return rules[rng.Intn(len(rules))]
}

func randomSeedFile(rng *rand.Rand) string {
	files := []string{"main.go", "foo.go", "bar.go", "baz.go", "util.go"}
	return files[rng.Intn(len(files))]
}

func randomSeverity(rng *rand.Rand) Severity {
	sevs := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}
	return sevs[rng.Intn(len(sevs))]
}

func randomSeedTool(rng *rand.Rand) string {
	tools := []string{"govet", "staticcheck", "artdupl", "branching"}
	return tools[rng.Intn(len(tools))]
}

func randomSeedCategory(rng *rand.Rand) Category {
	cats := []Category{
		CategorySecurity,
		CategoryStyle,
		CategoryPerformance,
		CategoryCorrectness,
		CategoryTesting,
	}
	return cats[rng.Intn(len(cats))]
}

func splitSeedFindings(all []Finding, rng *rand.Rand) []*Report {
	n := rng.Intn(3) + 1
	reports := make([]*Report, n)
	for i := range reports {
		reports[i] = NewReport(ToolInfo{Name: "test"})
	}
	for _, f := range all {
		idx := rng.Intn(n)
		reports[idx].AddFinding(f)
	}
	return reports
}
