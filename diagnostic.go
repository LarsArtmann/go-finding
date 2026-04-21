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
	findingPos := FromTokenPosition(pos)

	// Determine fix strategy from suggested fixes
	fixStrategy := FixStrategyNone

	var suggestion, afterCode string

	if len(d.SuggestedFixes) > 0 {
		fixStrategy = FixStrategyDirect

		suggestion = d.SuggestedFixes[0].Message
		if len(d.SuggestedFixes[0].TextEdits) > 0 {
			afterCode = string(d.SuggestedFixes[0].TextEdits[0].NewText)
		}
	}

	// Build ID from available info
	id := GenerateID(toolName, ruleCode, findingPos)

	//nolint:exhaustruct
	f := Finding{
		ID:          id,
		Rule:        ruleCode,
		ToolName:    toolName,
		Message:     d.Message,
		Severity:    SeverityWarning, // go/analysis doesn't have severity
		Position:    findingPos,
		Category:    Category(d.Category),
		FixStrategy: fixStrategy,
		Suggestion:  suggestion,
		AfterCode:   afterCode,
	}

	// Add related information
	for _, info := range d.Related {
		relatedPos := fset.Position(info.Pos)
		relatedID := GenerateID(toolName, ruleCode, FromTokenPosition(relatedPos))
		f.Related = append(f.Related, RelatedRef{
			FindingID: relatedID,
			Relation:  "related",
			Position:  FromTokenPosition(relatedPos),
		})
	}

	return f
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
		return Position{} //nolint:exhaustruct
	}

	return FromTokenPosition(fset.Position(node.Pos()))
}

// NodeRange returns a Range from an AST node.
func NodeRange(fset *token.FileSet, node ast.Node) Range {
	if node == nil {
		return Range{Start: Position{}, End: Position{}} //nolint:exhaustruct
	}

	startPos := FromTokenPosition(fset.Position(node.Pos()))

	return Range{
		Start: startPos,
		End:   FromTokenPosition(fset.Position(node.End())),
	}
}

// FormatDiagnostic returns a formatted string for a go/analysis diagnostic.
// Similar to how go vet formats output.
func FormatDiagnostic(d *analysis.Diagnostic, fset *token.FileSet, analyzerName string) string {
	pos := fset.Position(d.Pos)

	return fmt.Sprintf(
		"%s:%d:%d: %s: %s",
		pos.Filename,
		pos.Line,
		pos.Column,
		analyzerName,
		d.Message,
	)
}
