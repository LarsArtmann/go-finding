package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	"golang.org/x/tools/go/analysis"
)

func TestToDiagnostic_WithSuggestedFix(t *testing.T) {
	t.Parallel()

	src := "package main\n\nfunc main() {\n\told()\n}\n"

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, "fixme.go", src, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	f := finding.Finding{
		Rule:        "R1",
		ToolName:    "tool",
		Message:     "replace old with new",
		BeforeCode:  "old()",
		AfterCode:   "new()",
		FixStrategy: finding.FixStrategyDirect,
		Position:    finding.Position{File: "fixme.go", Line: 4, Column: 2},
		Suggestion:  "use new()",
	}

	diag := ToDiagnostic(f, fset)

	if len(diag.SuggestedFixes) != 1 {
		t.Fatalf("SuggestedFixes = %d, want 1", len(diag.SuggestedFixes))
	}

	if diag.SuggestedFixes[0].Message != "use new()" {
		t.Errorf("SuggestedFix.Message = %q, want %q", diag.SuggestedFixes[0].Message, "use new()")
	}

	if len(diag.SuggestedFixes[0].TextEdits) != 1 {
		t.Fatalf("TextEdits = %d, want 1", len(diag.SuggestedFixes[0].TextEdits))
	}

	edit := diag.SuggestedFixes[0].TextEdits[0]
	if string(edit.NewText) != "new()" {
		t.Errorf("NewText = %q, want %q", string(edit.NewText), "new()")
	}
}

func TestToDiagnostic_FileNotInFset(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f := finding.Finding{
		Message: "orphan finding",
		Position: finding.Position{
			File: "nonexistent.go",
			Line: 10,
		},
	}

	diag := ToDiagnostic(f, fset)

	if diag.Message != "orphan finding" {
		t.Errorf("Message = %q, want %q", diag.Message, "orphan finding")
	}

	if diag.Pos != token.NoPos {
		t.Errorf("Pos should be NoPos for file not in fset, got %v", diag.Pos)
	}
}

func TestToDiagnostic_RoundTrip(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	orig := &analysis.Diagnostic{
		Pos:            f.Name.Pos(),
		Message:        "round trip test",
		Category:       "test",
		SuggestedFixes: suggestedFix("fix it"),
	}

	// Forward: Diagnostic -> Finding
	got := FromDiagnostic(orig, fset, "tool", "R1")

	// Reverse: Finding -> Diagnostic
	result := ToDiagnostic(got, fset)

	if result.Message != orig.Message {
		t.Errorf("Message: got %q, want %q", result.Message, orig.Message)
	}

	if result.Category != orig.Category {
		t.Errorf("Category: got %q, want %q", result.Category, orig.Category)
	}

	// Position should round-trip exactly
	origPos := fset.Position(orig.Pos)

	resultPos := fset.Position(result.Pos)
	if origPos.Line != resultPos.Line || origPos.Column != resultPos.Column {
		t.Errorf("Position: got %d:%d, want %d:%d",
			resultPos.Line, resultPos.Column, origPos.Line, origPos.Column)
	}
}

func TestToDiagnostic_WithRelated(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, testFilename, testSrcMain, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	pos := fset.Position(f.Name.Pos())

	finding := finding.Finding{
		Message:  "main issue",
		Position: FromTokenPosition(pos),
		Related: []finding.RelatedRef{
			{
				Relation: "causes",
				Position: FromTokenPosition(pos),
			},
		},
	}

	diag := ToDiagnostic(finding, fset)

	if len(diag.Related) != 1 {
		t.Fatalf("Related = %d, want 1", len(diag.Related))
	}

	if diag.Related[0].Message != "causes" {
		t.Errorf("Related[0].Message = %q, want %q", diag.Related[0].Message, "causes")
	}
}

// multiEditSource is the erraudit legacyerrors shape: a fix that removes a
// declaration AND rewrites a condition in the same file.
const multiEditSource = "package main\n\nvar useLegacy = true\n\nfunc main() {\n\tif useLegacy {\n\t\tpanic(1)\n\t}\n}\n"

func multiEditDiagnostic(t *testing.T, fset *token.FileSet, file *ast.File) *analysis.Diagnostic {
	t.Helper()

	tokenFile := fset.File(file.Pos())

	declOffset := strings.Index(multiEditSource, "var useLegacy = true\n")
	if declOffset < 0 {
		t.Fatalf("decl not found in source")
	}

	condOff := strings.Index(multiEditSource, "useLegacy {")
	if condOff < 0 {
		t.Fatalf("condition not found in source")
	}

	return &analysis.Diagnostic{
		Pos:     tokenFile.Pos(condOff),
		Message: "legacy flag must go",
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "inline the condition and drop the declaration",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     tokenFile.Pos(declOffset),
						End:     tokenFile.Pos(declOffset + len("var useLegacy = true\n")),
						NewText: nil,
					},
					{
						Pos:     tokenFile.Pos(condOff),
						End:     tokenFile.Pos(condOff + len("useLegacy {")),
						NewText: []byte("false {"),
					},
				},
			},
		},
	}
}

func TestFromDiagnosticWithSource_MultiEditFixCarriesAllEdits(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "multi.go", multiEditSource, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	d := multiEditDiagnostic(t, fset, file)

	f := FromDiagnosticWithSource(d, fset, []byte(multiEditSource), "tool", "R1")

	if len(f.Edits) != 2 {
		t.Fatalf("Edits = %d, want 2 (all edits of the fix must survive the bridge)", len(f.Edits))
	}

	first := f.Edits[0]
	declOffset := strings.Index(multiEditSource, "var useLegacy = true\n")

	if first.Start.Offset != declOffset || first.End.Offset != declOffset+len("var useLegacy = true\n") {
		t.Errorf("Edits[0] span = [%d, %d), want [%d, %d)",
			first.Start.Offset, first.End.Offset,
			declOffset, declOffset+len("var useLegacy = true\n"))
	}

	if first.NewText != "" {
		t.Errorf("Edits[0].NewText = %q, want empty (pure deletion)", first.NewText)
	}

	second := f.Edits[1]
	if second.NewText != "false {" {
		t.Errorf("Edits[1].NewText = %q, want %q", second.NewText, "false {")
	}

	if second.Start.File != "multi.go" || second.End.File != "multi.go" {
		t.Errorf("Edits[1] file = %q/%q, want multi.go", second.Start.File, second.End.File)
	}

	if f.BeforeCode != "var useLegacy = true\n" {
		t.Errorf("BeforeCode = %q, want the first edit's span text", f.BeforeCode)
	}

	if f.AfterCode != "" {
		t.Errorf("AfterCode = %q, want empty (first edit is a deletion)", f.AfterCode)
	}
}

func TestToDiagnostic_MultiEditRoundTrip(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "multi.go", multiEditSource, 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	orig := multiEditDiagnostic(t, fset, file)

	f := FromDiagnosticWithSource(orig, fset, []byte(multiEditSource), "tool", "R1")

	roundTripped := ToDiagnostic(f, fset)

	if len(roundTripped.SuggestedFixes) != 1 {
		t.Fatalf("SuggestedFixes = %d, want 1", len(roundTripped.SuggestedFixes))
	}

	got := roundTripped.SuggestedFixes[0].TextEdits
	if len(got) != 2 {
		t.Fatalf("TextEdits = %d, want 2 (round-trip must preserve both edits)", len(got))
	}

	for i := range got {
		want := orig.SuggestedFixes[0].TextEdits[i]

		gotStart, wantStart := fset.Position(got[i].Pos), fset.Position(want.Pos)
		gotEnd, wantEnd := fset.Position(got[i].End), fset.Position(want.End)

		if gotStart.Offset != wantStart.Offset || gotEnd.Offset != wantEnd.Offset {
			t.Errorf("TextEdits[%d] span = [%d, %d), want [%d, %d)",
				i, gotStart.Offset, gotEnd.Offset, wantStart.Offset, wantEnd.Offset)
		}

		if string(got[i].NewText) != string(want.NewText) {
			t.Errorf("TextEdits[%d].NewText = %q, want %q", i, got[i].NewText, want.NewText)
		}
	}
}

func TestToDiagnostic_InsertionEditRoundTrip(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	_, err := parser.ParseFile(fset, "insert.go", "package main\n", 0)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	f := finding.Finding{
		Rule:        "R1",
		ToolName:    "tool",
		Message:     "add import",
		FixStrategy: finding.FixStrategyDirect,
		Position:    finding.Position{File: "insert.go", Line: 1, Column: 1},
		Suggestion:  "insert the import",
		Edits: []finding.TextEdit{
			{
				Start:   finding.Position{File: "insert.go", Line: 1, Column: 1, Offset: 0},
				End:     finding.Position{Offset: -1},
				NewText: "import \"fmt\"\n",
			},
		},
	}

	diag := ToDiagnostic(f, fset)

	if len(diag.SuggestedFixes) != 1 || len(diag.SuggestedFixes[0].TextEdits) != 1 {
		t.Fatalf("SuggestedFixes/TextEdits = %d/%d, want 1/1",
			len(diag.SuggestedFixes), len(diag.SuggestedFixes[0].TextEdits))
	}

	edit := diag.SuggestedFixes[0].TextEdits[0]
	if edit.Pos != edit.End {
		t.Errorf("insertion edit must resolve End to Pos, got Pos %v != End %v", edit.Pos, edit.End)
	}

	if string(edit.NewText) != "import \"fmt\"\n" {
		t.Errorf("NewText = %q, want the import line", edit.NewText)
	}
}
