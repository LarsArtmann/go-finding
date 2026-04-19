package finding

import "testing"

// parseIDCase is a test case for ParseID.
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

// stdIDCase creates a parseIDCase with standard tool/rule fields.
func stdIDCase(name, id, file string, line, col int, ok bool) parseIDCase {
	return parseIDCase{
		name:     name,
		id:       id,
		wantTool: "govet", wantRule: "nilcheck",
		wantFile: file, wantLine: line, wantCol: col, wantOK: ok,
	}
}

// testParseIDCase runs a parseIDCase, validating ParseID returns expected values.
func testParseIDCase(t *testing.T, tt parseIDCase) {
	t.Helper()

	p := ParseID(tt.id)
	if p.OK() != tt.wantOK {
		t.Fatalf("ParseID() ok = %v, want %v", p.OK(), tt.wantOK)
	}

	if !tt.wantOK {
		return
	}

	if p.Tool != tt.wantTool || p.Rule != tt.wantRule || p.File != tt.wantFile ||
		p.Line != tt.wantLine || p.Column != tt.wantCol {
		t.Errorf("ParseID() = (%q, %q, %q, %d, %d), want (%q, %q, %q, %d, %d)",
			p.Tool, p.Rule, p.File, p.Line, p.Column,
			tt.wantTool, tt.wantRule, tt.wantFile, tt.wantLine, tt.wantCol)
	}
}

// runParseIDCases runs all parseIDCase tests in a subtest.
func runParseIDCases(t *testing.T, tests []parseIDCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testParseIDCase(t, tt)
		})
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

			got := GenerateID(tt.tool, tt.rule, tt.pos)
			if tt.check != nil {
				tt.check(t, got)

				return
			}

			if got != tt.wantPrefix {
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

func TestGenerateID_ParseID_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		rule string
		pos  Position
	}{
		{"full", "govet", "nilcheck", Position{File: "main.go", Line: 42, Column: 10}},
		{"line only", "govet", "nilcheck", Position{File: "main.go", Line: 42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := GenerateID(tt.tool, tt.rule, tt.pos)
			p := ParseID(id)

			if !p.OK() {
				t.Fatalf("ParseID(%q) returned ok=false", id)
			}

			if p.Tool != tt.tool {
				t.Errorf("tool = %q, want %q", p.Tool, tt.tool)
			}

			if p.Rule != tt.rule {
				t.Errorf("rule = %q, want %q", p.Rule, tt.rule)
			}

			if p.File != tt.pos.File {
				t.Errorf("file = %q, want %q", p.File, tt.pos.File)
			}

			if p.Line != tt.pos.Line {
				t.Errorf("line = %d, want %d", p.Line, tt.pos.Line)
			}

			if p.Column != tt.pos.Column {
				t.Errorf("col = %d, want %d", p.Column, tt.pos.Column)
			}
		})
	}
}

func TestIsHashID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want bool
	}{
		{"tool:rule:0123456789abcdef", true},
		{"tool:rule:main.go:42:10", false},
		{"tool:rule:short", false},
		{"tooshort", false},
		{"a:b:c:d", false},
		{"tool:rule:abcdefghijklmnop", false}, // 16 chars but not hex
		{"tool:rule:0123456789ABCDEF", true},  // uppercase hex
	}

	for _, tt := range tests {
		if got := IsHashID(tt.id); got != tt.want {
			t.Errorf("IsHashID(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestParseID_WindowsPaths(t *testing.T) {
	t.Parallel()

	tests := []parseIDCase{
		{
			name:     "Windows absolute drive letter with line and col",
			id:       "govet:nilcheck:C:/Users/test/file.go:42:10",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "C:/Users/test/file.go", wantLine: 42, wantCol: 10, wantOK: true,
		},
		{
			name:     "Windows absolute drive letter line only",
			id:       "govet:nilcheck:C:/Users/test/file.go:42",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "C:/Users/test/file.go", wantLine: 42, wantCol: 0, wantOK: true,
		},
		{
			name:     "Windows drive letter no position",
			id:       "govet:nilcheck:C:/Users/test/file.go",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "C:/Users/test/file.go", wantLine: 0, wantCol: 0, wantOK: true,
		},
		{
			name:     "Windows path with spaces",
			id:       "govet:nilcheck:C:/Program Files/My App/main.go:15:5",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "C:/Program Files/My App/main.go", wantLine: 15, wantCol: 5, wantOK: true,
		},
		{
			name:     "UNC forward-slash path",
			id:       "govet:nilcheck://server/share/file.go:10:1",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "//server/share/file.go", wantLine: 10, wantCol: 1, wantOK: true,
		},
		{
			name:     "UNC path line only",
			id:       "govet:nilcheck://server/share/dir/file.go:7",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "//server/share/dir/file.go", wantLine: 7, wantCol: 0, wantOK: true,
		},
		{
			name:     "drive root only",
			id:       "govet:nilcheck:C:/:1:1",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "C:/", wantLine: 1, wantCol: 1, wantOK: true,
		},
	}

	runParseIDCases(t, tests)
}

func TestGenerateID_WindowsPathRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tool string
		rule string
		file string
		line int
		col  int
	}{
		{"Windows drive path", "govet", "nilcheck", "C:/Users/test/file.go", 42, 10},
		{"Windows path spaces", "staticcheck", "SA1000", "C:/Program Files/My App/main.go", 15, 5},
		{"UNC path", "govet", "assign", "//server/share/dir/file.go", 7, 0},
		{"drive root", "govet", "nilcheck", "C:/", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pos := Position{File: tt.file, Line: tt.line, Column: tt.col}
			id := GenerateID(tt.tool, tt.rule, pos)

			p := ParseID(id)
			if !p.OK() {
				t.Fatalf("ParseID(%q) returned ok=false", id)
			}

			if p.Tool != tt.tool {
				t.Errorf("tool = %q, want %q", p.Tool, tt.tool)
			}
			if p.Rule != tt.rule {
				t.Errorf("rule = %q, want %q", p.Rule, tt.rule)
			}
			if p.File != tt.file {
				t.Errorf("file = %q, want %q", p.File, tt.file)
			}
			if p.Line != tt.line {
				t.Errorf("line = %d, want %d", p.Line, tt.line)
			}

			if p.Column != tt.col {
				t.Errorf("col = %d, want %d", p.Column, tt.col)
			}
		})
	}
}

func TestExtractFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		parts         []string
		trailingCount int
		want          string
	}{
		{"simple", []string{"tool", "rule", "main.go", "42", "10"}, 2, "main.go"},
		{
			"path with colons",
			[]string{"tool", "rule", "C", "Users", "main.go", "42"},
			1,
			"C:Users:main.go",
		},
		{"too few parts", []string{"tool", "rule"}, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractFile(tt.parts, tt.trailingCount)
			if got != tt.want {
				t.Errorf("extractFile() = %q, want %q", got, tt.want)
			}
		})
	}
}
