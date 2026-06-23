package finding

import (
	"testing"
)

func TestSeverityToLSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     LSPSeverity
	}{
		{"critical maps to error", SeverityCritical, LSPSeverityError},
		{"error maps to error", SeverityError, LSPSeverityError},
		{"warning maps to warning", SeverityWarning, LSPSeverityWarning},
		{"info maps to info", SeverityInfo, LSPSeverityInfo},
		{"unknown maps to warning", Severity("unknown"), LSPSeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := severityToLSP(tt.severity); got != tt.want {
				t.Errorf("severityToLSP(%v) = %d, want %d", tt.severity, int(got), int(tt.want))
			}
		})
	}
}

func TestSeverityFromLSP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sev  LSPSeverity
		want Severity
	}{
		{"error(1) maps to SeverityError", LSPSeverityError, SeverityError},
		{"warning(2) maps to SeverityWarning", LSPSeverityWarning, SeverityWarning},
		{"info(3) maps to SeverityInfo", LSPSeverityInfo, SeverityInfo},
		{"hint(4) maps to SeverityInfo", LSPSeverityHint, SeverityInfo},
		{"unknown(0) maps to SeverityWarning", 0, SeverityWarning},
		{"unknown(99) maps to SeverityWarning", 99, SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := severityFromLSP(tt.sev); got != tt.want {
				t.Errorf("severityFromLSP(%d) = %v, want %v", int(tt.sev), got, tt.want)
			}
		})
	}
}

func lspDiag(line, startChar, endChar int) LSPDiagnostic {
	return LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: line, Character: startChar},
			End:   LSPPosition{Line: line, Character: endChar},
		},
	}
}

func TestFromLSP(t *testing.T) {
	t.Parallel()

	diag := LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: 4, Character: 9},
			End:   LSPPosition{Line: 4, Character: 15},
		},
		Severity: LSPSeverityError,
		Code:     "unused-var",
		Source:   "golangci-lint",
		Message:  "unused variable: x",
	}

	f := FromLSP("file:///test.go", diag)

	if got, want := f.ID, "golangci-lint:unused-var:file:///test.go:5:10"; string(got) != want {
		t.Errorf("FromLSP ID = %q, want %q", got, want)
	}

	if got, want := f.Rule, "unused-var"; string(got) != want {
		t.Errorf("FromLSP Rule = %q, want %q", got, want)
	}

	if got, want := f.ToolName, "golangci-lint"; string(got) != want {
		t.Errorf("FromLSP ToolName = %q, want %q", got, want)
	}

	if got, want := f.Message, "unused variable: x"; got != want {
		t.Errorf("FromLSP Message = %q, want %q", got, want)
	}

	if got, want := f.Severity, SeverityError; got != want {
		t.Errorf("FromLSP Severity = %v, want %v", got, want)
	}

	if got, want := f.Position.Line, 5; got != want {
		t.Errorf("FromLSP Position.Line = %d, want %d (0-based→1-based)", got, want)
	}

	if got, want := f.Position.Column, 10; got != want {
		t.Errorf("FromLSP Position.Column = %d, want %d (0-based→1-based)", got, want)
	}

	if got, want := f.Position.File, "file:///test.go"; got != want {
		t.Errorf("FromLSP Position.File = %q, want %q", got, want)
	}

	if got, want := f.FixStrategy, FixStrategyNone; got != want {
		t.Errorf("FromLSP FixStrategy = %v, want %v", got, want)
	}

	if f.Range == nil {
		t.Fatal("FromLSP Range = nil, want non-nil (end differs from start)")
	}

	if got, want := f.Range.End.Line, 5; got != want {
		t.Errorf("FromLSP Range.End.Line = %d, want %d", got, want)
	}

	if got, want := f.Range.End.Column, 16; got != want {
		t.Errorf("FromLSP Range.End.Column = %d, want %d", got, want)
	}

	if got, want := f.Metadata[LSPSeverityKey], "1"; got != want {
		t.Errorf("FromLSP Metadata[lsp-severity] = %q, want %q", got, want)
	}

	p := ParseID(f.ID)
	if !p.OK() {
		t.Fatalf("ParseID(%q) failed", f.ID)
	}

	if p.Tool != "golangci-lint" {
		t.Errorf("ParseID Tool = %q, want %q", p.Tool, "golangci-lint")
	}

	if p.Rule != "unused-var" {
		t.Errorf("ParseID Rule = %q, want %q", p.Rule, "unused-var")
	}

	if p.File != "file:///test.go" {
		t.Errorf("ParseID File = %q, want %q", p.File, "file:///test.go")
	}

	if p.Line != 5 {
		t.Errorf("ParseID Line = %d, want %d", p.Line, 5)
	}

	if p.Column != 10 {
		t.Errorf("ParseID Column = %d, want %d", p.Column, 10)
	}
}

func TestFromLSPNoEndRange(t *testing.T) {
	t.Parallel()

	diag := lspDiag(4, 9, 9)
	diag.Severity = LSPSeverityWarning
	diag.Code = "simplify"
	diag.Source = "gofmt"
	diag.Message = "can simplify"

	f := FromLSP("file:///test.go", diag)

	if f.Range != nil {
		t.Errorf("FromLSP Range = %v, want nil (end same as start)", f.Range)
	}
}

func TestFromLSPRelated(t *testing.T) {
	t.Parallel()

	diag := lspDiag(0, 0, 5)
	diag.Severity = LSPSeverityWarning
	diag.Code = "dupl"
	diag.Source = "dupl"
	diag.Message = "clone detected"
	diag.Related = []LSPRelated{
		{
			Location: LSPLocation{
				URI:   "file:///other.go",
				Range: LSPRange{Start: LSPPosition{Line: 9, Character: 0}},
			},
			Message: "clone-of",
		},
	}

	f := FromLSP("file:///test.go", diag)

	if got, want := len(f.Related), 1; got != want {
		t.Fatalf("FromLSP Related length = %d, want %d", got, want)
	}

	if got, want := f.Related[0].Relation, RelationKind("clone-of"); got != want {
		t.Errorf("FromLSP Related[0].Relation = %q, want %q", got, want)
	}

	if got, want := f.Related[0].Position.File, "file:///other.go"; got != want {
		t.Errorf("FromLSP Related[0].File = %q, want %q", got, want)
	}

	if got, want := f.Related[0].Position.Line, 10; got != want {
		t.Errorf("FromLSP Related[0].Line = %d, want %d", got, want)
	}
}
