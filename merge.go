package finding

import (
	"fmt"
	"maps"
	"slices"
)

// Correlation heuristics constants.
const (
	minFindingsInFile     = 2     // Minimum findings in a file for correlation analysis
	maxLineDiff           = 5     // Maximum line difference for considering findings related
	correlationScoreScale = 5.0   // For converting lineDiff to score
	minCorrelationScore   = 0.5   // Minimum correlation score for matching
	maxCorrelations       = 10000 // Maximum correlations to prevent O(n²) hangs
)

// mergedToolName is the ToolInfo.Name used for reports produced by Combine.
const mergedToolName = "merged"

// emptyToolName is the ToolInfo.Name used for empty reports from Combine.
const emptyToolName = "empty"

// Combine merges multiple reports into a new report with optional deduplication.
// The resulting report has:
//   - Tool.Name = mergedToolName (unless there's only one report)
//   - Findings from all reports
//   - Summary computed from all findings
//
// Use Report.Merge(other) to concatenate one report into another in-place without deduplication.
func Combine(reports []*Report, opts ...MergeOption) *Report {
	if len(reports) == 0 {
		return NewReport(ToolInfo{Name: emptyToolName}) //nolint:exhaustruct
	}

	if len(reports) == 1 {
		r := reports[0]

		result := &Report{ //nolint:exhaustruct
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

	total := 0

	for _, report := range reports {
		if report != nil {
			total += len(report.Findings)
		}
	}

	merged := newReportWithCapacity(ToolInfo{
		Name:    mergedToolName,
		Version: "",
	}, total)

	seen := make(map[string]struct{}, total)

	for _, report := range reports {
		if report == nil {
			continue
		}

		for _, finding := range report.Findings {
			if options.Deduplicate {
				key, ok := dedupKey(finding, options)
				if ok {
					if _, exists := seen[key]; exists {
						continue
					}

					seen[key] = struct{}{}
				}
			}

			merged.addFindingUnchecked(finding.Clone())
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

func dedupKey(finding Finding, opts MergeOptions) (string, bool) {
	switch opts.DeduplicateBy {
	case DeduplicateByID:
		if finding.ID == "" {
			return "", false
		}

		return finding.ID, true
	case DeduplicateByPosition:
		if finding.Position.File == "" {
			return "", false
		}

		return fmt.Sprintf(
			"%s:%s:%d:%d",
			finding.ToolName,
			finding.Position.File,
			finding.Position.Line,
			finding.Position.Column,
		), true
	case DeduplicateByRule:
		if finding.Position.File == "" {
			return "", false
		}

		return fmt.Sprintf(
			"%s:%s:%d:%d",
			finding.Rule,
			finding.Position.File,
			finding.Position.Line,
			finding.Position.Column,
		), true
	default:
		return finding.ID, true
	}
}

// CorrelationScore measures the strength of a correlation between findings.
// Unlike Confidence (which measures certainty of a single finding),
// CorrelationScore measures how strongly two findings are related.
type CorrelationScore float64

// IsValid returns true if the score is in the valid range [0.0, 1.0].
func (s CorrelationScore) IsValid() bool {
	return float64(s) >= 0.0 && float64(s) <= 1.0
}

// String returns a human-readable representation of the correlation score.
func (s CorrelationScore) String() string {
	return fmt.Sprintf("%.2f", float64(s))
}

// Correlation represents a relationship between two or more findings.
type Correlation struct {
	FindingIDs []string         `json:"findingIds"`
	Reason     string           `json:"reason"` // Why they're correlated
	Score      CorrelationScore `json:"score"`  // 0.0-1.0 correlation strength
}

// Correlate finds potentially related findings across tools.
// Currently uses simple heuristics: same file + nearby lines.
//
// This can be used standalone or enabled in Pipeline via Config.CorrelateFindings.
// When enabled, the pipeline populates PipelineResult.Correlations automatically.
//
// # Complexity
//
// Findings are grouped by file, then sorted by line. For each finding, the inner
// loop scans forward until the line difference exceeds maxLineDiff (5 lines),
// then breaks. For well-distributed findings this is effectively O(n) per file.
//
// Worst case: if many findings cluster on the same lines in one file (e.g., 1000
// findings on line 1), the inner loop degrades to O(k²) for that file where k is
// the number of findings in that file. The maxCorrelations constant (10,000) caps
// total output, but silently drops correlations beyond the cap.
//
// For datasets exceeding ~50K findings in a single file, consider pre-filtering
// or raising maxCorrelations (requires source modification).
func Correlate(findings []Finding) []Correlation {
	var correlations []Correlation

	byFile := GroupByFile(findings)

	files := slices.Collect(maps.Keys(byFile))

	slices.Sort(files)

	for _, file := range files {
		fileFindings := byFile[file]
		if len(fileFindings) < minFindingsInFile {
			continue
		}

		slices.SortFunc(fileFindings, func(a, b Finding) int {
			return a.Position.Line - b.Position.Line
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
						Score:      CorrelationScore(confidence),
					})
					if len(correlations) >= maxCorrelations {
						return correlations
					}
				}
			}
		}
	}

	return correlations
}
