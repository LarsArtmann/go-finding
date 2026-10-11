package finding

import (
	"testing"
)

type parseIDCase struct {
	name     string
	id       string
	wantTool string
	wantRule string
	wantFile string
	wantLine int
	wantCol  int
	wantOK   bool
}

func stdIDCase(name, id, file string, line, col int, ok bool) parseIDCase {
	return parseIDCase{
		name:     name,
		id:       id,
		wantTool: "govet", wantRule: "nilcheck",
		wantFile: file, wantLine: line, wantCol: col, wantOK: ok,
	}
}

func testParseIDCase(t *testing.T, tt parseIDCase) {
	t.Helper()

	p := ParseID(ID(tt.id))
	if p.OK() != tt.wantOK {
		t.Fatalf("OK() = %v, want %v", p.OK(), tt.wantOK)
	}

	if !tt.wantOK {
		return
	}

	if string(p.Tool) != tt.wantTool {
		t.Errorf("Tool = %q, want %q", p.Tool, tt.wantTool)
	}

	if string(p.Rule) != tt.wantRule {
		t.Errorf("Rule = %q, want %q", p.Rule, tt.wantRule)
	}

	if string(p.File) != tt.wantFile {
		t.Errorf("File = %q, want %q", p.File, tt.wantFile)
	}

	if p.Line != tt.wantLine {
		t.Errorf("Line = %d, want %d", p.Line, tt.wantLine)
	}

	if p.Column != tt.wantCol {
		t.Errorf("Column = %d, want %d", p.Column, tt.wantCol)
	}
}

func runParseIDCases(t *testing.T, tests []parseIDCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testParseIDCase(t, tt)
		})
	}
}

func assertRoundTrip(t *testing.T, p *ParsedID, tool, rule, file string, line, col int) {
	t.Helper()

	if string(p.Tool) != tool {
		t.Errorf("Tool = %q, want %q", p.Tool, tool)
	}

	if string(p.Rule) != rule {
		t.Errorf("Rule = %q, want %q", p.Rule, rule)
	}

	if string(p.File) != file {
		t.Errorf("File = %q, want %q", p.File, file)
	}

	if p.Line != line {
		t.Errorf("Line = %d, want %d", p.Line, line)
	}

	if p.Column != col {
		t.Errorf("Column = %d, want %d", p.Column, col)
	}
}

func TestGenerateID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tool       string
		rule       string
		pos        Position
		wantPrefix string
		check      func(t *testing.T, got string)
	}{
		{
			name:       "full position",
			tool:       "govet",
			rule:       "nilcheck",
			pos:        Position{File: "main.go", Line: 42, Column: 10},
			wantPrefix: "govet:nilcheck:main.go:42:10",
		},
		{
			name:       "line only",
			tool:       "govet",
			rule:       "nilcheck",
			pos:        Position{File: "main.go", Line: 42},
			wantPrefix: "govet:nilcheck:main.go:42",
		},
		{
			name: "hash-based no line",
			tool: "govet",
			rule: "nilcheck",
			pos:  Position{File: "main.go"},
			check: func(t *testing.T, got string) {
				t.Helper()

				if !IsHashID(got) {
					t.Errorf("expected hash-based ID, got %q", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GenerateID(ToolName(tt.tool), RuleName(tt.rule), tt.pos)
			if tt.check != nil {
				tt.check(t, string(got))

				return
			}

			if string(got) != tt.wantPrefix {
				t.Errorf("GenerateID() = %q, want %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	t.Parallel()

	pos := Position{File: "main.go", Line: 10, Column: 5}
	a := GenerateID("tool", "rule", pos)
	b := GenerateID("tool", "rule", pos)

	if a != b {
		t.Errorf("GenerateID not deterministic: %q != %q", a, b)
	}
}

func TestGenerateID_HashDeterministic(t *testing.T) {
	t.Parallel()

	pos := Position{File: "main.go"}
	a := GenerateID("tool", "rule", pos)
	b := GenerateID("tool", "rule", pos)

	if a != b {
		t.Errorf("hash-based GenerateID not deterministic: %q != %q", a, b)
	}
}

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []parseIDCase{
		stdIDCase("full ID", "govet:nilcheck:main.go:42:10", "main.go", 42, 10, true),
		stdIDCase("line only", "govet:nilcheck:main.go:42", "main.go", 42, 0, true),
		stdIDCase("no position", "govet:nilcheck:main.go", "main.go", 0, 0, true),
		{
			name:     "hash-based",
			id:       "tool:rule:0123456789abcdef",
			wantTool: "tool", wantRule: "rule",
			wantFile: "", wantLine: 0, wantCol: 0, wantOK: true,
		},
		{
			name:   "too few parts",
			id:     "onlyone",
			wantOK: false,
		},
		{
			name:   "empty string",
			id:     "",
			wantOK: false,
		},
	}

	runParseIDCases(t, tests)
}
