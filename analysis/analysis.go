// Package analysis provides integration between go/analysis diagnostics
// and finding.Finding values. Importing this package adds a dependency on
// golang.org/x/tools — consumers that don't need go/analysis integration
// can use the core finding package without this dependency.
package analysis

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
)

// FromDiagnostic converts a go/analysis.Diagnostic to a Finding.
// The toolName parameter identifies which analyzer produced this.
// The ruleCode parameter provides a rule identifier (since go/analysis.Diagnostic doesn't have Code).
// The defaultSeverity is used because go/analysis.Diagnostics don't carry severity;
// if empty, SeverityWarning is used.
func FromDiagnostic(
	d *analysis.Diagnostic,
	fset *token.FileSet,
	toolName, ruleCode string,
	defaultSeverity ...finding.Severity,
) finding.Finding {
	sev := finding.SeverityWarning
	if len(defaultSeverity) > 0 && defaultSeverity[0].IsValid() {
		sev = defaultSeverity[0]
	}

	pos := fset.Position(d.Pos)
	findingPos := FromTokenPosition(pos)

	fixStrategy := finding.FixStrategyNone

	var suggestion, afterCode string

	if len(d.SuggestedFixes) > 0 {
		fixStrategy = finding.FixStrategyDirect

		suggestion = d.SuggestedFixes[0].Message
		if len(d.SuggestedFixes[0].TextEdits) > 0 {
			afterCode = string(d.SuggestedFixes[0].TextEdits[0].NewText)
		}
	}

	id := finding.GenerateID(toolName, ruleCode, findingPos)

	//nolint:exhaustruct
	f := finding.Finding{
		ID:          id,
		Rule:        ruleCode,
		ToolName:    toolName,
		Message:     d.Message,
		Severity:    sev,
		Position:    findingPos,
		Category:    finding.Category(d.Category),
		FixStrategy: fixStrategy,
		Suggestion:  suggestion,
		AfterCode:   afterCode,
	}

	for _, info := range d.Related {
		relatedPos := fset.Position(info.Pos)
		relatedID := finding.GenerateID(toolName, ruleCode, FromTokenPosition(relatedPos))
		f.Related = append(f.Related, finding.RelatedRef{
			FindingID: relatedID,
			Relation:  "related",
			Position:  FromTokenPosition(relatedPos),
		})
	}

	return f
}

// FromTokenPosition creates a Position from a token.Position.
func FromTokenPosition(pos token.Position) finding.Position {
	return finding.Position{
		File:   pos.Filename,
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
	}
}

// NodePosition returns a Position from an AST node.
func NodePosition(fset *token.FileSet, node ast.Node) finding.Position {
	return FromTokenPosition(nodeStartPos(fset, node))
}

// NodeRange returns a Range from an AST node.
func NodeRange(fset *token.FileSet, node ast.Node) finding.Range {
	if node == nil {
		return finding.Range{
			Start: finding.Position{}, //nolint:exhaustruct
			End:   finding.Position{}, //nolint:exhaustruct
		}
	}

	return finding.Range{
		Start: FromTokenPosition(nodeStartPos(fset, node)),
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

// nodeStartPos returns the start token.Position of an AST node, or a zero position if node is nil.
func nodeStartPos(fset *token.FileSet, node ast.Node) token.Position {
	if node == nil {
		return token.Position{
			Filename: "",
			Offset:   0,
			Line:     0,
			Column:   0,
		}
	}

	return fset.Position(node.Pos())
}
