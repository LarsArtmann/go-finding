package finding

import "fmt"

// LSP types for conversion.
// These are simplified representations of LSP Diagnostic types.

type LSPDiagnostic struct {
	Range    LSPRange         `json:"range"`
	Severity int              `json:"severity,omitempty"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Code     string           `json:"code,omitempty"`
	Source   string           `json:"source,omitempty"`
	Message  string           `json:"message"`
	Related  []LSPRelatedInfo `json:"relatedInformation,omitempty"`
}

type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

type LSPPosition struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

type LSPRelatedInfo struct {
	Location LSPLocation `json:"location"`
	Message  string      `json:"message"`
}

type LSPLocation struct {
	URI   string   `json:"uri"`
	Range LSPRange `json:"range"`
}

// ToLSP converts a Finding to LSP Diagnostic format.
// Note: This is a lossy conversion - some fields (FixStrategy, Confidence, etc.) are lost.
func (f Finding) ToLSP() LSPDiagnostic {
	diag := LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{
				Line:      f.Position.Line - 1,   // LSP is 0-based
				Character: f.Position.Column - 1, // LSP is 0-based
			},
		},
		Severity: severityToLSP(f.Severity),
		Code:     f.Rule,
		Source:   f.ToolName,
		Message:  f.Message,
	}

	// Set end position if available
	if f.Range != nil && f.Range.End.Line > 0 {
		diag.Range.End = LSPPosition{
			Line:      f.Range.End.Line - 1,
			Character: f.Range.End.Column - 1,
		}
	} else {
		// Single position diagnostic
		diag.Range.End = diag.Range.Start
	}

	// Add related information
	for _, rel := range f.Related {
		diag.Related = append(diag.Related, LSPRelatedInfo{
			Location: LSPLocation{
				URI: rel.Position.File,
				Range: LSPRange{
					Start: LSPPosition{
						Line:      rel.Position.Line - 1,
						Character: rel.Position.Column - 1,
					},
					End: LSPPosition{
						Line:      rel.Position.Line - 1,
						Character: rel.Position.Column - 1,
					},
				},
			},
			Message: rel.Relation,
		})
	}

	return diag
}

// FromLSP creates a Finding from an LSP Diagnostic.
// Many fields will be empty/default since LSP has less information.
func FromLSP(diag LSPDiagnostic) Finding {
	return Finding{
		ID:       fmt.Sprintf("lsp:%s:%s:%d:%d", diag.Source, diag.Code, diag.Range.Start.Line+1, diag.Range.Start.Character+1),
		Rule:     diag.Code,
		ToolName: diag.Source,
		Message:  diag.Message,
		Severity: severityFromLSP(diag.Severity),
		Position: Position{
			File:   "", // Would need URI from context
			Line:   diag.Range.Start.Line + 1,
			Column: diag.Range.Start.Character + 1,
		},
		FixStrategy: FixStrategyNone, // LSP doesn't specify this
	}
}

func severityToLSP(s Severity) int {
	switch s {
	case SeverityCritical, SeverityError:
		return 1 // Error
	case SeverityWarning:
		return 2 // Warning
	case SeverityInfo:
		return 3 // Info
	default:
		return 2 // Default to warning
	}
}

func severityFromLSP(sev int) Severity {
	switch sev {
	case 1:
		return SeverityError
	case 2:
		return SeverityWarning
	case 3, 4:
		return SeverityInfo
	default:
		return SeverityWarning
	}
}
