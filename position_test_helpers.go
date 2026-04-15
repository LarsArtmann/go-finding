package finding

// Test helpers for position/range tests.

// rangeLine creates a Range with line-based positions in the same file.
func rangeLine(file string, startLine, endLine int) Range {
	return Range{Start: Position{File: file, Line: startLine}, End: Position{Line: endLine}}
}

// rangeOffset creates a Range with offset-based positions in the same file.
func rangeOffset(file string, startOffset, endOffset int) Range {
	return Range{Start: Position{File: file, Offset: startOffset}, End: Position{Offset: endOffset}}
}

// ptrRange returns a pointer to a Range (for test expected values).
func ptrRange(r Range) *Range {
	return &r
}

// posLine creates a Position with a line number.
func posLine(file string, line int) Position {
	return Position{File: file, Line: line}
}

// posLineCol creates a Position with line and column.
func posLineCol(file string, line, col int) Position {
	return Position{File: file, Line: line, Column: col}
}

// findingWithFix creates a Finding with fix strategy.
func findingWithFix(id, file string, line int, fix FixStrategy) Finding {
	return Finding{
		ID:          id,
		Position:    Position{File: file, Line: line},
		FixStrategy: fix,
	}
}

// findings creates a slice of Findings with IDs and files.
func findings(idsAndFiles ...string) []Finding {
	result := make([]Finding, 0, len(idsAndFiles)/2)
	for i := 0; i < len(idsAndFiles); i += 2 {
		result = append(result, Finding{
			ID:       idsAndFiles[i],
			Position: Position{File: idsAndFiles[i+1]},
		})
	}
	return result
}

// findingsWithSeverity creates Findings with IDs, files, and severities.
func findingsWithSeverity(data []struct {
	id       string
	file     string
	severity Severity
}) []Finding {
	result := make([]Finding, len(data))
	for i, d := range data {
		result[i] = Finding{
			ID:       d.id,
			Position: Position{File: d.file},
			Severity: d.severity,
		}
	}
	return result
}

// findingsWithTool creates Findings with tool names.
func findingsWithTool(ids, toolNames []string) []Finding {
	result := make([]Finding, 0, len(ids))
	for i, id := range ids {
		tool := ""
		if i < len(toolNames) {
			tool = toolNames[i]
		}
		result = append(result, Finding{ID: id, ToolName: tool})
	}
	return result
}
