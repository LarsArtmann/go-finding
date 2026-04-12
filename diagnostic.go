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
		ID:          id,
		Rule:        ruleCode,
		ToolName:    toolName,
		Message:     d.Message,
		Severity:    SeverityWarning, // go/analysis doesn't have severity
		Position:    FromTokenPosition(pos),
		Category:    d.Category,
		FixStrategy: fixStrategy,
	}

	// Add related information
	for _, info := range d.Related {
		relatedPos := fset.Position(info.Pos)
		f.Related = append(f.Related, RelatedRef{
			Relation: "related",
			Position: FromTokenPosition(relatedPos),
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

// FromTokenPosition creates a Position from a token.Position.
func FromTokenPosition(pos token.Position) Position {
	return Position{
		File:   pos.Filename,
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
	}
}

// NodePosition returns a Position from an AST node.
func NodePosition(fset *token.FileSet, node ast.Node) Position {
	if node == nil {
		return Position{}
	}

	return FromTokenPosition(fset.Position(node.Pos()))
}

// NodeRange returns a Range from an AST node.
func NodeRange(fset *token.FileSet, node ast.Node) Range {
	if node == nil {
		return Range{}
	}

	return Range{
		Start: FromTokenPosition(fset.Position(node.Pos())),
		End:   FromTokenPosition(fset.Position(node.End())),
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
