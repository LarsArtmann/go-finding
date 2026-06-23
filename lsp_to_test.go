package finding

import (
	"testing"
)

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

			if got, want := diag.Code, f.Rule; got != string(want) {
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
