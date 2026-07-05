package finding

import (
	"testing"
)

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

			id := GenerateID(ToolName(tt.tool), RuleName(tt.rule), tt.pos)
			p := ParseID(id)

			if !p.OK() {
				t.Fatalf("ParseID(%q) returned ok=false", id)
			}

			assertRoundTrip(t, &p, tt.tool, tt.rule, string(tt.pos.File), tt.pos.Line, tt.pos.Column)
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
		{"tool:rule:abcdefghijklmnop", false},
		{"tool:rule:0123456789ABCDEF", true},
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
		stdIDCase(
			"Windows absolute drive letter with line and col",
			"govet:nilcheck:C:/Users/test/file.go:42:10",
			"C:/Users/test/file.go",
			42,
			10,
			true,
		),
		stdIDCase(
			"Windows absolute drive letter line only",
			"govet:nilcheck:C:/Users/test/file.go:42",
			"C:/Users/test/file.go",
			42,
			0,
			true,
		),
		stdIDCase(
			"Windows drive letter no position",
			"govet:nilcheck:C:/Users/test/file.go",
			"C:/Users/test/file.go",
			0,
			0,
			true,
		),
		stdIDCase(
			"Windows path with spaces",
			"govet:nilcheck:C:/Program Files/My App/main.go:15:5",
			"C:/Program Files/My App/main.go",
			15,
			5,
			true,
		),
		stdIDCase(
			"UNC forward-slash path",
			"govet:nilcheck://server/share/file.go:10:1",
			"//server/share/file.go",
			10,
			1,
			true,
		),
		stdIDCase(
			"UNC path line only",
			"govet:nilcheck://server/share/dir/file.go:7",
			"//server/share/dir/file.go",
			7,
			0,
			true,
		),
		stdIDCase("drive root only", "govet:nilcheck:C:/:1:1", "C:/", 1, 1, true),
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

			pos := Position{File: FilePath(tt.file), Line: tt.line, Column: tt.col}
			id := GenerateID(ToolName(tt.tool), RuleName(tt.rule), pos)

			p := ParseID(id)
			if !p.OK() {
				t.Fatalf("ParseID(%q) returned ok=false", id)
			}

			assertRoundTrip(t, &p, tt.tool, tt.rule, tt.file, tt.line, tt.col)
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

			if got := extractFile(tt.parts, tt.trailingCount); got != tt.want {
				t.Errorf("extractFile() = %q, want %q", got, tt.want)
			}
		})
	}
}
