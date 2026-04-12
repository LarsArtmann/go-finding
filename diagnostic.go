package finding

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// FromDiagnostic converts a go/analysis.Diagnostic to a Finding.
// The toolName parameter identifies which analyzer produced this.
// The ruleCode parameter provides a rule identifier (since go/analysis.Diagnostic doesn't have Code).
func FromDiagnostic(
	d *analysis.Diagnostic,
	fset *token.FileSet,
	toolName, ruleCode string,
) Finding {
	pos := fset.Position(d.Pos)

	// Determine fix strategy from suggested fixes
	fixStrategy := FixStrategyNone
	if len(d.SuggestedFixes) > 0 {
		fixStrategy = FixStrategyDirect
	}

	// Build ID from available info
	id := GenerateID(toolName, ruleCode, Position{
		File:   pos.Filename,
		Line:   pos.Line,
		Column: pos.Column,
	})

	f := Finding{
		ID:       id,
		Rule:     ruleCode,
		ToolName: toolName,
		Message:  d.Message,
		Severity: SeverityWarning, // go/analysis doesn't have severity
		Position: Position{
			File:   pos.Filename,
			Line:   pos.Line,
			Column: pos.Column,
			Offset: pos.Offset,
		},
		Category:    d.Category,
		FixStrategy: fixStrategy,
	}

	// Add related information
	for _, info := range d.Related {
		relatedPos := fset.Position(info.Pos)
		f.Related = append(f.Related, RelatedRef{
			Relation: "related",
			Position: Position{
				File:   relatedPos.Filename,
				Line:   relatedPos.Line,
				Column: relatedPos.Column,
				Offset: relatedPos.Offset,
			},
		})
	}

	return f
}

// AnalysisDiagnostic creates a Diagnostic from a Finding for go/analysis framework.
// This is the reverse of FromDiagnostic.
// Note: This is lossy - many Finding fields have no equivalent in analysis.Diagnostic.
func (f Finding) AnalysisDiagnostic() analysis.Diagnostic {
	d := analysis.Diagnostic{
		Pos:      token.NoPos, // Would need fset
		Message:  f.Message,
		Category: f.Category,
	}

	if f.FixStrategy == FixStrategyDirect && f.AfterCode != "" {
		// We can't accurately set Pos/End without fset
		// This is a simplified representation
		d.SuggestedFixes = []analysis.SuggestedFix{
			{
				Message: f.Suggestion,
				TextEdits: []analysis.TextEdit{
					{
						NewText: []byte(f.AfterCode),
					},
				},
			},
		}
	}

	return d
}

// NodePosition returns a Position from an AST node.
func NodePosition(fset *token.FileSet, node ast.Node) Position {
	if node == nil {
		return Position{}
	}

	pos := fset.Position(node.Pos())

	return Position{
		File:   pos.Filename,
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
	}
}

// NodeRange returns a Range from an AST node.
func NodeRange(fset *token.FileSet, node ast.Node) Range {
	if node == nil {
		return Range{}
	}

	startPos := fset.Position(node.Pos())
	endPos := fset.Position(node.End())

	return Range{
		Start: Position{
			File:   startPos.Filename,
			Line:   startPos.Line,
			Column: startPos.Column,
			Offset: startPos.Offset,
		},
		End: Position{
			File:   endPos.Filename,
			Line:   endPos.Line,
			Column: endPos.Column,
			Offset: endPos.Offset,
		},
	}
}

// FormatDiagnostic returns a formatted string for a go/analysis diagnostic.
// Similar to how go vet formats output.
func FormatDiagnostic(d *analysis.Diagnostic, fset *token.FileSet, analyzerName string) string {
	pos := fset.Position(d.Pos)

	return fmt.Sprintf(
		"%s:%d:%d: %s: %s\n",
		pos.Filename,
		pos.Line,
		pos.Column,
		analyzerName,
		d.Message,
	)
}
