package detectors

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
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
	assert.Len(t, findings, 2)

	f := findings[0]
	assert.Equal(t, "govet", f.ToolName)
	assert.Equal(t, "github.com/example/pkg", f.Rule)
	assert.Equal(t, "unused variable x", f.Message)
	assert.Equal(t, finding.SeverityWarning, f.Severity)
	assert.Equal(t, finding.CategoryCorrectness, f.Category)
	assert.Equal(t, finding.FixStrategySuggest, f.FixStrategy)
	assert.Equal(t, "/project/main.go", f.Position.File)
	assert.Equal(t, 10, f.Position.Line)
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
	assert.Equal(t, "staticcheck", f.ToolName)

	assert.Equal(t, "SA1000", f.Rule)
	assert.Equal(t, finding.SeverityWarning, f.Severity)
	assert.Equal(t, finding.CategoryStyle, f.Category)
	assert.InDelta(t, 0.8, f.Confidence, 0.001)

	// Second finding: S1001 = style category, error severity
	f2 := findings[1]
	assert.Equal(t, "S1001", f2.Rule)
	assert.Equal(t, finding.SeverityError, f2.Severity)
	assert.Equal(t, finding.CategoryStyle, f2.Category)
}

func TestParseStaticcheckJSON_Empty(t *testing.T) {
	t.Parallel()

	findings := parseStaticcheckJSON(nil, "")
	if findings != nil {
		t.Errorf("expected nil for empty input, got %v", findings)
	}
}

func TestParseStaticcheckJSON_LineSkipping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantLen     int
		wantRule    string
		description string
	}{
		{
			name:        "invalid line skipped",
			input:       "not json at all\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}",
			wantLen:     1,
			wantRule:    "S1001",
			description: "skip invalid line",
		},
		{
			name:        "whitespace lines skipped",
			input:       "\n  \n\t\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}\n\n",
			wantLen:     1,
			wantRule:    "S1001",
			description: "skip whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := parseStaticcheckJSON([]byte(tt.input), "")
			if len(findings) != tt.wantLen {
				t.Fatalf("expected %d findings, got %d", tt.wantLen, len(findings))
			}

			if findings[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", findings[0].Rule, tt.wantRule)
			}
		})
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
		{"F1001", finding.CategoryPerformance},
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

func TestNewGoVetDetector_CancelledContext(t *testing.T) {
	t.Parallel()
	d := NewGoVetDetector(".")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := d.Detect(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestNewStaticcheckDetector_CancelledContext(t *testing.T) {
	t.Parallel()
	d := NewStaticcheckDetector(".")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := d.Detect(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestNewGoVetDetector_ValidProject(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	d := NewGoVetDetector("../../")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("govet: %v", err)
	}

	t.Logf("govet found %d findings", len(findings))
}

func TestParseGoVetJSON_BadEntry(t *testing.T) {
	t.Parallel()

	input := `{"github.com/example/pkg": "not an array"}`
	findings := parseGoVetJSON([]byte(input), "")
	if findings != nil {
		t.Errorf("expected nil for bad entry, got %v", findings)
	}
}

func TestStaticcheckCategory_A(t *testing.T) {
	t.Parallel()

	got := staticcheckCategory("A1000")
	if got != finding.CategoryCorrectness {
		t.Errorf("staticcheckCategory(A1000) = %v, want %v", got, finding.CategoryCorrectness)
	}
}

func TestParseStaticcheckJSON_AbsolutePath(t *testing.T) {
	t.Parallel()

	input := `{"code":"S1001","severity":"warning","location":{"file":"/abs/path/main.go","line":1,"column":1},"message":"ok"}`
	findings := parseStaticcheckJSON([]byte(input), "/project")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Position.File != "/abs/path/main.go" {
		t.Errorf("File = %q, want %q", findings[0].Position.File, "/abs/path/main.go")
	}
}
