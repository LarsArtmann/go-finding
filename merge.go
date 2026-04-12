package finding

import (
	"fmt"
	"sort"
)

// Merge correlation constants.
const (
	MinFindingsInFile        = 2   // Minimum findings in a file for correlation analysis
	MaxLineDiff              = 5   // Maximum line difference for considering findings related
	CorrelationScoreScale    = 5.0 // For converting lineDiff to score
	MinCorrelationScore      = 0.5 // Minimum correlation score for matching
	FindingPairsSameToolSkip = 2   // Skip first 2 index in findingPairs array
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

		return &Report{
			Tool:     r.Tool,
			Findings: append([]Finding(nil), r.Findings...),
			Summary:  r.Summary,
		}
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
		for _, finding := range report.Findings {
			// Check for duplicates
			if options.Deduplicate {
				key := dedupKey(finding, options)
				if _, exists := seen[key]; exists {
					continue
				}

				seen[key] = struct{}{}
			}

			merged.AddFinding(finding)
		}
	}

	merged.ComputeSummary()

	return merged
}

// MergeOptions controls how reports are merged.
type MergeOptions struct {
	Deduplicate     bool
	DeduplicateBy   DeduplicateBy
	KeepAllRelated  bool
	ConflictHandler ConflictHandler
}

// MergeOption is a functional option for configuring merge behavior.
type MergeOption func(*MergeOptions)

// DeduplicateBy specifies what fields to use for deduplication.
type DeduplicateBy int

const (
	DeduplicateByID       DeduplicateBy = iota // DeduplicateByID ensures exact ID matches.
	DeduplicateByPosition                      // DeduplicateByPosition uses file:line:column.
	DeduplicateByRule                          // DeduplicateByRule matches rule position.
)

// ConflictHandler handles when findings conflict.
type ConflictHandler int

const (
	ConflictKeepFirst   ConflictHandler = iota // ConflictKeepFirst keeps first occurrence.
	ConflictKeepLast                           // ConflictKeepLast keeps last occurrence.
	ConflictKeepHighest                        // ConflictKeepHighest keeps highest severity.
	ConflictKeepAll                            // ConflictKeepAll keeps both findings.
)

func defaultMergeOptions() MergeOptions {
	return MergeOptions{
		Deduplicate:     true,
		DeduplicateBy:   DeduplicateByID,
		KeepAllRelated:  true,
		ConflictHandler: ConflictKeepHighest,
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
	FindingIDs []string // IDs of correlated findings
	Reason     string   // Why they're correlated
	Confidence float64  // 0.0-1.0
}

// Correlate finds potentially related findings across tools.
// Currently uses simple heuristics: same file + overlapping lines.
func Correlate(findings []Finding) []Correlation {
	var correlations []Correlation

	// Group by file
	byFile := GroupByFile(findings)

	for _, fileFindings := range byFile {
		if len(fileFindings) < MinFindingsInFile {
			continue
		}

		// Sort by line
		sort.Slice(fileFindings, func(i, j int) bool {
			return fileFindings[i].Position.Line < fileFindings[j].Position.Line
		})

		// Find nearby findings from different tools
		for i, f1 := range fileFindings {
			for _, f2 := range fileFindings[i+1:] {
				if f1.ToolName == f2.ToolName {
					continue // Same tool, skip
				}

				// Check if lines are close
				lineDiff := f2.Position.Line - f1.Position.Line
				if lineDiff > MaxLineDiff { // Within MaxLineDiff
					break
				}

				confidence := 1.0 - (float64(lineDiff) / CorrelationScoreScale)
				if confidence > MinCorrelationScore {
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
