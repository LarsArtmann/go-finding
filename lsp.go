package finding

import "fmt"

// LSP severity level constants
const (
	LSPErrorSeverity     = 1 // Error
	LSPWarningSeverity   = 2 // Warning
	LSPInfoSeverity      = 3 // Info
	LSPCaseError         = 1 // Error case value
	LSPCaseWarning       = 2 // Warning case value
	LSPCaseInfo          = 3 // Info case value
	LSPCaseInfoAlias     = 4 // Secondary Info case value
)

// LSP types for conversion.
// These are simplified representations of LSP Diagnostic types.

// LSPDiagnostic represents an LSP (Language Server Protocol) diagnostic.
// Used for converting Finding objects to LSP diagnostic format.
type LSPDiagnostic struct {
	Range    LSPRange         `json:"range"`
	Severity int              `json:"severity,omitempty"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Code     string           `json:"code,omitempty"`
	Source   string           `json:"source,omitempty"`
	Message  string           `json:"message"`
	Related  []LSPRelatedInfo `json:"relatedInformation,omitempty"`
}

// LSPRange represents a 0-based character range in a text document.
type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

// LSPPosition represents a 0-based position in a text document.
type LSPPosition struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

// LSPRelatedInfo provides related information for a diagnostic.
type LSPRelatedInfo struct {
	Location LSPLocation `json:"location"`
	Message  string      `json:"message"`
}

// LSPLocation represents the location of a diagnostic.
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
	if f.Range != nil && f.Range.HasEnd() {
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
		lspPos := LSPPosition{
			Line:      rel.Position.Line - 1,
			Character: rel.Position.Column - 1,
		}
		diag.Related = append(diag.Related, LSPRelatedInfo{
			Location: LSPLocation{
				URI:   rel.Position.File,
				Range: LSPRange{Start: lspPos, End: lspPos},
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
		ID: fmt.Sprintf(
			"lsp:%s:%s:%d:%d",
			diag.Source,
			diag.Code,
			diag.Range.Start.Line+1,
			diag.Range.Start.Character+1,
		),
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
		return LSPErrorSeverity
	case SeverityWarning:
		return LSPWarningSeverity
	case SeverityInfo:
		return LSPInfoSeverity
	default:
		return LSPWarningSeverity // Default to warning
	}
}

func severityFromLSP(sev int) Severity {
	switch sev {
	case LSPCaseError:
		return SeverityError
	case LSPCaseWarning:
		return SeverityWarning
	case LSPCaseInfo, LSPCaseInfoAlias:
		return SeverityInfo
	default:
		return SeverityWarning
	}
}
