package finding

import "fmt"

// LSP severity level constants per the LSP specification.
const (
	LSPSeverityError   = 1 // Error
	LSPSeverityWarning = 2 // Warning
	LSPSeverityInfo    = 3 // Information
	LSPSeverityHint    = 4 // Hint
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
				Line:      lspLine(f.Position.Line),
				Character: lspLine(f.Position.Column),
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
			Line:      lspLine(f.Range.End.Line),
			Character: lspLine(f.Range.End.Column),
		}
	} else {
		// Single position diagnostic
		diag.Range.End = diag.Range.Start
	}

	// Add related information
	for _, rel := range f.Related {
		lspPos := LSPPosition{
			Line:      lspLine(rel.Position.Line),
			Character: lspLine(rel.Position.Column),
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

// lspLine converts a 1-based position to a 0-based LSP position, clamping to 0.
func lspLine(n int) int {
	if n <= 0 {
		return 0
	}
	return n - 1
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
		return LSPSeverityError
	case SeverityWarning:
		return LSPSeverityWarning
	case SeverityInfo:
		return LSPSeverityInfo
	default:
		return LSPSeverityWarning
	}
}

func severityFromLSP(sev int) Severity {
	switch sev {
	case LSPSeverityError:
		return SeverityError
	case LSPSeverityWarning:
		return SeverityWarning
	case LSPSeverityInfo, LSPSeverityHint:
		return SeverityInfo
	default:
		return SeverityWarning
	}
}
