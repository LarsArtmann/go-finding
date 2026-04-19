package finding

import (
	"fmt"
	"sort"
)

// Merge correlation constants.
const (
	minFindingsInFile     = 2   // Minimum findings in a file for correlation analysis
	maxLineDiff           = 5   // Maximum line difference for considering findings related
	correlationScoreScale = 5.0 // For converting lineDiff to score
	minCorrelationScore   = 0.5 // Minimum correlation score for matching
)

// Merge combines multiple reports into one.
// The merged report has:
// - Tool.Name = "merged" (unless there's only one report)
// - Findings from all reports
// - Summary computed from all findings
// Options control deduplication and conflict resolution.
func Merge(reports []*Report, opts ...MergeOption) *Report {
	if len(reports) == 0 {
		return NewReport(ToolInfo{Name: "empty"})
	}

	if len(reports) == 1 {
		r := reports[0]

		result := &Report{
			Tool:     r.Tool,
			Findings: cloneFindings(r.Findings),
		}
		result.ComputeSummary()

		return result
	}

	options := defaultMergeOptions()
	for _, opt := range opts {
		opt(&options)
	}

	merged := NewReport(ToolInfo{
		Name:    "merged",
		Version: "",
	})

	seen := make(map[string]struct{})

	for _, report := range reports {
		if report == nil {
			continue
		}

		for _, finding := range report.Findings {
			// Check for duplicates
			if options.Deduplicate {
				key := dedupKey(finding, options)
				if _, exists := seen[key]; exists {
					continue
				}

				seen[key] = struct{}{}
			}

			merged.AddFinding(finding.Clone())
		}
	}

	merged.ComputeSummary()

	return merged
}

func cloneFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return nil
	}

	cloned := make([]Finding, len(findings))
	for i, f := range findings {
		cloned[i] = f.Clone()
	}

	return cloned
}

// MergeOptions controls how reports are merged.
type MergeOptions struct {
	Deduplicate   bool
	DeduplicateBy DeduplicateBy
}

// MergeOption is a functional option for configuring merge behavior.
type MergeOption func(*MergeOptions)

// DeduplicateBy specifies what fields to use for deduplication.
type DeduplicateBy int

// Deduplication strategies control how findings are matched during merge.
const (
	DeduplicateByID       DeduplicateBy = iota // Exact ID matches.
	DeduplicateByPosition                      // File:line:column matching.
	DeduplicateByRule                          // Rule + position matching.
)

func defaultMergeOptions() MergeOptions {
	return MergeOptions{
		Deduplicate:   true,
		DeduplicateBy: DeduplicateByID,
	}
}

// WithDeduplication enables/disables deduplication.
func WithDeduplication(enabled bool) MergeOption {
	return func(o *MergeOptions) {
		o.Deduplicate = enabled
	}
}

// WithDeduplicateBy sets the deduplication strategy.
func WithDeduplicateBy(by DeduplicateBy) MergeOption {
	return func(o *MergeOptions) {
		o.DeduplicateBy = by
	}
}

func dedupKey(finding Finding, opts MergeOptions) string {
	switch opts.DeduplicateBy {
	case DeduplicateByID:
		return finding.ID
	case DeduplicateByPosition:
		return fmt.Sprintf(
			"%s:%d:%d",
			finding.Position.File,
			finding.Position.Line,
			finding.Position.Column,
		)
	case DeduplicateByRule:
		return fmt.Sprintf(
			"%s:%s:%d:%d",
			finding.Rule,
			finding.Position.File,
			finding.Position.Line,
			finding.Position.Column,
		)
	default:
		return finding.ID
	}
}

// Correlation links related findings from different tools.
type Correlation struct {
	FindingIDs []string `json:"finding_ids"` //nolint:tagliatelle // SARIF uses snake_case
	Reason     string   `json:"reason"`      // Why they're correlated
	Confidence float64  `json:"confidence"`  // 0.0-1.0
}

// Correlate finds potentially related findings across tools.
// Currently uses simple heuristics: same file + overlapping lines.
//
// This is a standalone utility — it is not wired into Pipeline.Run().
// Call it directly on merged findings when cross-tool correlation is needed.
func Correlate(findings []Finding) []Correlation {
	var correlations []Correlation

	byFile := GroupByFile(findings)

	files := make([]string, 0, len(byFile))
	for f := range byFile {
		files = append(files, f)
	}

	sort.Strings(files)

	for _, file := range files {
		fileFindings := byFile[file]
		if len(fileFindings) < minFindingsInFile {
			continue
		}

		sort.Slice(fileFindings, func(i, j int) bool {
			return fileFindings[i].Position.Line < fileFindings[j].Position.Line
		})

		for i, f1 := range fileFindings {
			for _, f2 := range fileFindings[i+1:] {
				if f1.ToolName == f2.ToolName {
					continue
				}

				lineDiff := f2.Position.Line - f1.Position.Line
				if lineDiff > maxLineDiff {
					break
				}

				confidence := 1.0 - (float64(lineDiff) / correlationScoreScale)
				if confidence > minCorrelationScore {
					correlations = append(correlations, Correlation{
						FindingIDs: []string{f1.ID, f2.ID},
						Reason:     "same file, nearby lines",
						Confidence: confidence,
					})
				}
			}
		}
	}

	return correlations
}
