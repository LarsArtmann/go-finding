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

	if got, want := f.ID, "golangci-lint:unused-var:file:///test.go:5:10"; got != want {
		t.Errorf("FromLSP ID = %q, want %q", got, want)
	}

	if got, want := f.Rule, "unused-var"; got != want {
		t.Errorf("FromLSP Rule = %q, want %q", got, want)
	}

	if got, want := f.ToolName, "golangci-lint"; got != want {
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

	if got, want := f.Metadata["go-finding/lsp-severity"], "1"; got != want {
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
	diag.Related = []LSPRelatedInfo{
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

func TestToLSP(t *testing.T) { //nolint:gocognit,funlen // comprehensive table-driven test
	t.Parallel()

	t.Run("with range", func(t *testing.T) {
		t.Run("with start/end positions", func(t *testing.T) {
			t.Parallel()

			f := Finding{
				Rule:     "SA1000",
				ToolName: "staticcheck",
				Message:  "invalid printf format",
				Severity: SeverityWarning,
				Position: Position{File: "main.go", Line: 10, Column: 5},
				Range:    NewRangePtr("main.go", 10, 5, 10, 20),
			}

			diag := f.ToLSP()

			if got, want := diag.Severity, LSPSeverityWarning; got != want {
				t.Errorf("ToLSP Severity = %d, want %d", got, want)
			}

			if got, want := diag.Code, f.Rule; got != want {
				t.Errorf("ToLSP Code = %q, want %q", got, want)
			}

			if got, want := diag.Source, "staticcheck"; got != want {
				t.Errorf("ToLSP Source = %q, want %q", got, want)
			}

			if got, want := diag.Message, "invalid printf format"; got != want {
				t.Errorf("ToLSP Message = %q, want %q", got, want)
			}

			if got, want := diag.Range.Start.Line, 9; got != want {
				t.Errorf("ToLSP Start.Line = %d, want %d (1-based→0-based)", got, want)
			}

			if got, want := diag.Range.Start.Character, 4; got != want {
				t.Errorf("ToLSP Start.Character = %d, want %d (1-based→0-based)", got, want)
			}

			if got, want := diag.Range.End.Line, 9; got != want {
				t.Errorf("ToLSP End.Line = %d, want %d", got, want)
			}

			if got, want := diag.Range.End.Character, 19; got != want {
				t.Errorf("ToLSP End.Character = %d, want %d", got, want)
			}
		})

		t.Run("zero-based conversion with zero line", func(t *testing.T) {
			t.Parallel()

			f := Finding{
				Rule:     "r1",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityWarning,
				Position: Position{File: "a.go", Line: 0, Column: 0},
			}

			diag := f.ToLSP()
			if diag.Range.Start.Line != 0 {
				t.Errorf("ToLSP Line=0 → Start.Line=0, got %d", diag.Range.Start.Line)
			}

			if diag.Range.Start.Character != 0 {
				t.Errorf("ToLSP Column=0 → Start.Character=0, got %d", diag.Range.Start.Character)
			}
		})
	})

	t.Run("without range uses start position for end", func(t *testing.T) {
		t.Parallel()

		f := Finding{
			Rule:     "vet",
			ToolName: "go vet",
			Message:  "missing printf argument",
			Severity: SeverityError,
			Position: Position{File: "main.go", Line: 3, Column: 7},
		}

		diag := f.ToLSP()

		if got, want := diag.Range.End.Line, diag.Range.Start.Line; got != want {
			t.Errorf("ToLSP End.Line = %d, want %d (same as start)", got, want)
		}

		if got, want := diag.Range.End.Character, diag.Range.Start.Character; got != want {
			t.Errorf("ToLSP End.Character = %d, want %d (same as start)", got, want)
		}
	})

	t.Run("with related info", func(t *testing.T) {
		t.Parallel()

		f := Finding{
			Rule:     "clone-detected",
			ToolName: "dupl",
			Message:  "clone detected",
			Severity: SeverityInfo,
			Position: Position{File: "a.go", Line: 5, Column: 1},
			Related: []RelatedRef{
				{
					FindingID: "dupl:clone-detected:b.go:10:1",
					Relation:  "clone-of",
					Position:  Position{File: "b.go", Line: 10, Column: 1},
				},
			},
		}

		diag := f.ToLSP()

		if got, want := len(diag.Related), 1; got != want {
			t.Fatalf("ToLSP Related length = %d, want %d", got, want)
		}

		rel := diag.Related[0]
		if got, want := rel.Message, "clone-of"; got != want {
			t.Errorf("ToLSP Related[0].Message = %q, want %q", got, want)
		}

		if got, want := rel.Location.URI, "b.go"; got != want {
			t.Errorf("ToLSP Related[0].URI = %q, want %q", got, want)
		}

		if got, want := rel.Location.Range.Start.Line, 9; got != want {
			t.Errorf("ToLSP Related[0].Line = %d, want %d (1-based→0-based)", got, want)
		}
	})

	t.Run("round-trip preserves core fields", func(t *testing.T) {
		t.Parallel()

		orig := lspDiag(0, 0, 5)
		orig.Severity = LSPSeverityWarning
		orig.Code = "rule1"
		orig.Source = "tool1"
		orig.Message = "msg1"

		f := FromLSP("file:///test.go", orig)
		roundTrip := f.ToLSP()

		if got, want := roundTrip.Code, orig.Code; got != want {
			t.Errorf("round-trip Code = %q, want %q", got, want)
		}

		if got, want := roundTrip.Source, orig.Source; got != want {
			t.Errorf("round-trip Source = %q, want %q", got, want)
		}

		if got, want := roundTrip.Message, orig.Message; got != want {
			t.Errorf("round-trip Message = %q, want %q", got, want)
		}

		if got, want := roundTrip.Severity, orig.Severity; got != want {
			t.Errorf("round-trip Severity = %d, want %d", got, want)
		}

		if got, want := roundTrip.Range.Start.Line, orig.Range.Start.Line; got != want {
			t.Errorf("round-trip Start.Line = %d, want %d", got, want)
		}

		if roundTrip.Range.Start.Character != orig.Range.Start.Character {
			t.Errorf("round-trip Start.Character = %d, want %d",
				roundTrip.Range.Start.Character, orig.Range.Start.Character)
		}
	})
}
