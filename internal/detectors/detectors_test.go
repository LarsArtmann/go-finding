package detectors

import (
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestParsePosn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		posn     string
		dir      string
		wantFile string
		wantLine int
		wantCol  int
	}{
		{
			name:     "full position",
			posn:     "main.go:10:5",
			dir:      "/project",
			wantFile: "/project/main.go",
			wantLine: 10,
			wantCol:  5,
		},
		{
			name:     "file and line only",
			posn:     "main.go:10",
			dir:      "/project",
			wantFile: "/project/main.go",
			wantLine: 10,
		},
		{
			name:     "absolute path ignored dir",
			posn:     "/abs/path/main.go:5:1",
			dir:      "/project",
			wantFile: "/abs/path/main.go",
			wantLine: 5,
			wantCol:  1,
		},
		{
			name:     "no colon returns raw string as file",
			posn:     "just-a-file.go",
			dir:      "",
			wantFile: "just-a-file.go",
		},
		{
			name:     "empty dir with relative path",
			posn:     "pkg/util.go:20:3",
			dir:      "",
			wantFile: "pkg/util.go",
			wantLine: 20,
			wantCol:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pos := parsePosn(tt.posn, tt.dir)
			if pos.File != tt.wantFile {
				t.Errorf("File = %q, want %q", pos.File, tt.wantFile)
			}

			if pos.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", pos.Line, tt.wantLine)
			}

			if pos.Column != tt.wantCol {
				t.Errorf("Column = %d, want %d", pos.Column, tt.wantCol)
			}
		})
	}
}

func TestParseGoVetJSON(t *testing.T) {
	t.Parallel()

	input := `{
		"github.com/example/pkg": [
			{"posn": "main.go:10:5", "message": "unused variable x"},
			{"posn": "main.go:20:1", "message": "unreachable code"}
		]
	}`

	findings := parseGoVetJSON([]byte(input), "/project")
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f := findings[0]
	if f.ToolName != "govet" {
		t.Errorf("ToolName = %q, want %q", f.ToolName, "govet")
	}

	if f.Message != "unused variable x" {
		t.Errorf("Message = %q, want %q", f.Message, "unused variable x")
	}
	{
		got := f.Severity
		if got != finding.SeverityWarning {
			t.Errorf("Severity = %v, want %v", got, finding.SeverityWarning)
		}
	}

	if f.Category != finding.CategoryCorrectness {
		t.Errorf("Category = %v, want %v", f.Category, finding.CategoryCorrectness)
	}

	if f.FixStrategy != finding.FixStrategySuggest {
		t.Errorf("FixStrategy = %v, want %v", f.FixStrategy, finding.FixStrategySuggest)
	}

	if f.Position.File != "/project/main.go" {
		t.Errorf("Position.File = %q, want %q", f.Position.File, "/project/main.go")
	}

	if f.Position.Line != 10 {
		t.Errorf("Position.Line = %d, want 10", f.Position.Line)
	}
}

func TestParseGoVetJSON_Invalid(t *testing.T) {
	t.Parallel()

	findings := parseGoVetJSON([]byte("not json"), "")
	if findings != nil {
		t.Errorf("expected nil for invalid JSON, got %v", findings)
	}
}

func TestParseGoVetJSON_Empty(t *testing.T) {
	t.Parallel()

	findings := parseGoVetJSON([]byte("{}"), "")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty object, got %d", len(findings))
	}
}

func TestParseStaticcheckJSON(t *testing.T) {
	t.Parallel()

	input := `{"code":"SA1000","severity":"warning","location":{"file":"main.go","line":10,"column":5},"message":"invalid regex pattern"}
{"code":"S1001","severity":"error","location":{"file":"util.go","line":20,"column":1},"message":"should use copy"}`

	findings := parseStaticcheckJSON([]byte(input), "/project")
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f := findings[0]
	if f.ToolName != "staticcheck" {
		t.Errorf("ToolName = %q, want %q", f.ToolName, "staticcheck")
	}

	if f.Rule != "SA1000" {
		t.Errorf("Rule = %q, want %q", f.Rule, "SA1000")
	}
	{
		actual := f.Severity
		if actual != finding.SeverityWarning {
			t.Errorf("Severity = %v, want %v", actual, finding.SeverityWarning)
		}
	}

	if f.Category != finding.CategoryStyle {
		t.Errorf("Category = %v, want %v", f.Category, finding.CategoryStyle)
	}

	if f.Confidence != 0.8 {
		t.Errorf("Confidence = %f, want 0.8", f.Confidence)
	}

	// Second finding: S1001 = style category, error severity
	f2 := findings[1]
	if f2.Rule != "S1001" {
		t.Errorf("Rule = %q, want %q", f2.Rule, "S1001")
	}
	{
		sev := f2.Severity
		if sev != finding.SeverityError {
			t.Errorf("Severity = %v, want %v", sev, finding.SeverityError)
		}
	}

	if f2.Category != finding.CategoryStyle {
		t.Errorf("Category = %v, want %v", f2.Category, finding.CategoryStyle)
	}
}

func TestParseStaticcheckJSON_Empty(t *testing.T) {
	t.Parallel()

	findings := parseStaticcheckJSON(nil, "")
	if findings != nil {
		t.Errorf("expected nil for empty input, got %v", findings)
	}
}

func TestParseStaticcheckJSON_InvalidLine(t *testing.T) {
	t.Parallel()

	input := "not json at all\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}"

	findings := parseStaticcheckJSON([]byte(input), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (skip invalid line), got %d", len(findings))
	}

	if findings[0].Rule != "S1001" {
		t.Errorf("Rule = %q, want %q", findings[0].Rule, "S1001")
	}
}

func TestStaticcheckCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want finding.Category
	}{
		{"SA1000", finding.CategoryStyle},
		{"S1001", finding.CategoryStyle},
		{"QF1001", finding.CategoryStyle},
		{"U1000", finding.CategoryUnused},
		{"PERF1001", finding.CategoryPerformance},
		{"R1001", finding.CategoryPerformance},
		{"SA", finding.CategoryStyle},
		{"", finding.CategoryCorrectness},
		{"X9999", finding.CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			got := staticcheckCategory(tt.code)
			if got != tt.want {
				t.Errorf("staticcheckCategory(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestDetectorNames(t *testing.T) {
	t.Parallel()

	govet := NewGoVetDetector(".")
	if govet.Name() != "govet" {
		t.Errorf("NewGoVetDetector name = %q, want %q", govet.Name(), "govet")
	}

	sc := NewStaticcheckDetector(".")
	if sc.Name() != "staticcheck" {
		t.Errorf("NewStaticcheckDetector name = %q, want %q", sc.Name(), "staticcheck")
	}
}
