package finding

import "testing"

func TestGenerateID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tool     string
		rule     string
		pos      Position
		wantPrefix string
		check    func(t *testing.T, got string)
	}{
		{
			name:      "full position",
			tool:      "govet",
			rule:      "nilcheck",
			pos:       Position{File: "main.go", Line: 42, Column: 10},
			wantPrefix: "govet:nilcheck:main.go:42:10",
		},
		{
			name:      "line only",
			tool:      "govet",
			rule:      "nilcheck",
			pos:       Position{File: "main.go", Line: 42},
			wantPrefix: "govet:nilcheck:main.go:42",
		},
		{
			name:      "hash-based no line",
			tool:      "govet",
			rule:      "nilcheck",
			pos:       Position{File: "main.go"},
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

	tests := []struct {
		name      string
		id        string
		wantTool  string
		wantRule  string
		wantFile  string
		wantLine  int
		wantCol   int
		wantOK    bool
	}{
		{
			name:     "full ID",
			id:       "govet:nilcheck:main.go:42:10",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "main.go", wantLine: 42, wantCol: 10, wantOK: true,
		},
		{
			name:     "line only",
			id:       "govet:nilcheck:main.go:42",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "main.go", wantLine: 42, wantCol: 0, wantOK: true,
		},
		{
			name:     "no position",
			id:       "govet:nilcheck:main.go",
			wantTool: "govet", wantRule: "nilcheck",
			wantFile: "main.go", wantLine: 0, wantCol: 0, wantOK: true,
		},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tool, rule, file, line, col, ok := ParseID(tt.id)
			if ok != tt.wantOK {
				t.Fatalf("ParseID() ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if tool != tt.wantTool || rule != tt.wantRule || file != tt.wantFile ||
				line != tt.wantLine || col != tt.wantCol {
				t.Errorf("ParseID() = (%q, %q, %q, %d, %d), want (%q, %q, %q, %d, %d)",
					tool, rule, file, line, col,
					tt.wantTool, tt.wantRule, tt.wantFile, tt.wantLine, tt.wantCol)
			}
		})
	}
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
			tool, rule, file, line, col, ok := ParseID(id)

			if !ok {
				t.Fatalf("ParseID(%q) returned ok=false", id)
			}
			if tool != tt.tool {
				t.Errorf("tool = %q, want %q", tool, tt.tool)
			}
			if rule != tt.rule {
				t.Errorf("rule = %q, want %q", rule, tt.rule)
			}
			if file != tt.pos.File {
				t.Errorf("file = %q, want %q", file, tt.pos.File)
			}
			if line != tt.pos.Line {
				t.Errorf("line = %d, want %d", line, tt.pos.Line)
			}
			if col != tt.pos.Column {
				t.Errorf("col = %d, want %d", col, tt.pos.Column)
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
	}

	for _, tt := range tests {
		if got := IsHashID(tt.id); got != tt.want {
			t.Errorf("IsHashID(%q) = %v, want %v", tt.id, got, tt.want)
		}
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
		{"path with colons", []string{"tool", "rule", "C", "Users", "main.go", "42"}, 1, "C:Users:main.go"},
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
