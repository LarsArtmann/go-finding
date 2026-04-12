package finding

import (
	"fmt"
	"sort"
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

	for _, r := range reports {
		for _, f := range r.Findings {
			// Check for duplicates
			if options.Deduplicate {
				key := dedupKey(f, options)
				if _, exists := seen[key]; exists {
					continue
				}

				seen[key] = struct{}{}
			}

			merged.AddFinding(f)
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

type MergeOption func(*MergeOptions)

// DeduplicateBy specifies what fields to use for deduplication.
type DeduplicateBy int

const (
	DeduplicateByID       DeduplicateBy = iota // Exact ID match
	DeduplicateByPosition                      // Same file, line, column
	DeduplicateByRule                          // Same rule + position
)

// ConflictHandler handles when findings conflict.
type ConflictHandler int

const (
	ConflictKeepFirst   ConflictHandler = iota // Keep first occurrence
	ConflictKeepLast                           // Keep last occurrence
	ConflictKeepHighest                        // Keep highest severity
	ConflictKeepAll                            // Keep both (don't deduplicate)
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

func dedupKey(f Finding, opts MergeOptions) string {
	switch opts.DeduplicateBy {
	case DeduplicateByID:
		return f.ID
	case DeduplicateByPosition:
		return fmt.Sprintf("%s:%d:%d", f.Position.File, f.Position.Line, f.Position.Column)
	case DeduplicateByRule:
		return fmt.Sprintf(
			"%s:%s:%d:%d",
			f.Rule,
			f.Position.File,
			f.Position.Line,
			f.Position.Column,
		)
	default:
		return f.ID
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
		if len(fileFindings) < 2 {
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
				if lineDiff > 5 { // Within 5 lines
					break
				}

				confidence := 1.0 - (float64(lineDiff) / 5.0)
				if confidence > 0.5 {
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
