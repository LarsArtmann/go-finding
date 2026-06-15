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
	"github.com/larsartmann/go-finding/internal/gotoken"
	"golang.org/x/tools/go/analysis"
)

// DefaultRelation is the default relation type for related information.
const DefaultRelation = finding.RelationRelated

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
		f.Related = append(f.Related, finding.RelatedRef{ //nolint:exhaustruct
			FindingID: relatedID,
			Relation:  DefaultRelation,
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

// ToDiagnostic converts a Finding back to a go/analysis.Diagnostic.
// The fset is used to resolve file positions back to token.Pos values.
// If the file referenced by f.Position.File is not in fset, token.NoPos is used.
//
// The conversion is lossy: Severity, Confidence, Tags, Suppression, Metadata,
// and FixStrategy are not representable in analysis.Diagnostic.
// BeforeCode/AfterCode are converted to a SuggestedFix with TextEdit when present.
func ToDiagnostic(f finding.Finding, fset *token.FileSet) analysis.Diagnostic {
	pos := resolvePos(f.Position, fset)

	diag := analysis.Diagnostic{ //nolint:exhaustruct
		Pos:      pos,
		Message:  f.Message,
		Category: string(f.Category),
	}

	if f.HasFix() && f.Position.File != "" {
		endPos := resolveEndPos(f, fset)
		newText := []byte(f.AfterCode)

		if f.BeforeCode != "" && endPos > pos {
			diag.SuggestedFixes = []analysis.SuggestedFix{
				{
					Message:   f.Suggestion,
					TextEdits: []analysis.TextEdit{{Pos: pos, End: endPos, NewText: newText}},
				},
			}
		} else if f.AfterCode != "" {
			diag.SuggestedFixes = []analysis.SuggestedFix{
				{
					Message:   f.Suggestion,
					TextEdits: []analysis.TextEdit{{Pos: pos, End: pos, NewText: newText}},
				},
			}
		}
	}

	for _, ref := range f.Related {
		relatedPos := resolvePos(ref.Position, fset)
		diag.Related = append(diag.Related, analysis.RelatedInformation{ //nolint:exhaustruct
			Pos:     relatedPos,
			Message: string(ref.Relation),
		})
	}

	return diag
}

// resolvePos converts a finding.Position to a token.Pos by looking up the file in fset.
// Returns token.NoPos if the file is not in the file set or the line is invalid.
func resolvePos(p finding.Position, fset *token.FileSet) token.Pos {
	if p.File == "" || p.Line <= 0 {
		return token.NoPos
	}

	file := gotoken.FindFileByName(fset, p.File)
	if file == nil {
		return token.NoPos
	}

	return gotoken.LineColToPos(file, p.Line, p.Column)
}

// resolveEndPos computes the end token.Pos for a finding's fix range.
func resolveEndPos(f finding.Finding, fset *token.FileSet) token.Pos {
	if f.Range != nil && f.Range.Start.File == f.Position.File {
		return resolvePos(f.Range.End, fset)
	}

	if f.BeforeCode != "" {
		start := resolvePos(f.Position, fset)
		if start != token.NoPos {
			return start + token.Pos(len(f.BeforeCode))
		}
	}

	return token.NoPos
}

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
