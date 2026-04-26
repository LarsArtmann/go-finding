package detectors

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{"full position", "main.go:10:5", "/project", "/project/main.go", 10, 5},
		{"file and line only", "main.go:10", "/project", "/project/main.go", 10, 0},
		{"absolute path ignored dir", "/abs/path/main.go:5:1", "/project", "/abs/path/main.go", 5, 1},
		{"no colon returns raw string as file", "just-a-file.go", "", "just-a-file.go", 0, 0},
		{"empty dir with relative path", "pkg/util.go:20:3", "", "pkg/util.go", 20, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pos := parsePosn(tt.posn, tt.dir)
			assert.Equal(t, tt.wantFile, pos.File)
			assert.Equal(t, tt.wantLine, pos.Line)
			assert.Equal(t, tt.wantCol, pos.Column)
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
	assert.Nil(t, findings)
}

func TestParseGoVetJSON_Empty(t *testing.T) {
	t.Parallel()

	findings := parseGoVetJSON([]byte("{}"), "")
	assert.Empty(t, findings)
}

func TestParseStaticcheckJSON(t *testing.T) {
	t.Parallel()

	input := `{"code":"SA1000","severity":"warning","location":{"file":"main.go","line":10,"column":5},"message":"invalid regex pattern"}
{"code":"S1001","severity":"error","location":{"file":"util.go","line":20,"column":1},"message":"should use copy"}`

	findings := parseStaticcheckJSON([]byte(input), "/project")
	require.Len(t, findings, 2)

	f := findings[0]
	assert.Equal(t, "staticcheck", f.ToolName)
	assert.Equal(t, "SA1000", f.Rule)
	assert.Equal(t, finding.SeverityWarning, f.Severity)
	assert.Equal(t, finding.CategoryStyle, f.Category)
	assert.InDelta(t, 0.8, f.Confidence, 0.001)

	f2 := findings[1]
	assert.Equal(t, "S1001", f2.Rule)
	assert.Equal(t, finding.SeverityError, f2.Severity)
	assert.Equal(t, finding.CategoryStyle, f2.Category)
}

func TestParseStaticcheckJSON_Empty(t *testing.T) {
	t.Parallel()

	findings := parseStaticcheckJSON(nil, "")
	assert.Nil(t, findings)
}

func TestParseStaticcheckJSON_LineSkipping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantRule string
	}{
		{"invalid line skipped", "not json at all\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}", 1, "S1001"},
		{"whitespace lines skipped", "\n  \n\t\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}\n\n", 1, "S1001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings := parseStaticcheckJSON([]byte(tt.input), "")
			require.Len(t, findings, tt.wantLen)
			assert.Equal(t, tt.wantRule, findings[0].Rule)
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

			assert.Equal(t, tt.want, staticcheckCategory(tt.code))
		})
	}
}

func TestDetectorNames(t *testing.T) {
	t.Parallel()

	govet := NewGoVetDetector(".")
	assert.Equal(t, "govet", govet.Name())

	sc := NewStaticcheckDetector(".")
	assert.Equal(t, "staticcheck", sc.Name())
}

func TestNewGoVetDetector_CancelledContext(t *testing.T) {
	t.Parallel()
	d := NewGoVetDetector(".")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := d.Detect(ctx)
	require.Error(t, err)
}

func TestNewStaticcheckDetector_CancelledContext(t *testing.T) {
	t.Parallel()
	d := NewStaticcheckDetector(".")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := d.Detect(ctx)
	require.Error(t, err)
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
	require.NoError(t, err)

	t.Logf("govet found %d findings", len(findings))
}

func TestParseGoVetJSON_BadEntry(t *testing.T) {
	t.Parallel()

	input := `{"github.com/example/pkg": "not an array"}`
	findings := parseGoVetJSON([]byte(input), "")
	assert.Nil(t, findings)
}

func TestStaticcheckCategory_A(t *testing.T) {
	t.Parallel()

	assert.Equal(t, finding.CategoryCorrectness, staticcheckCategory("A1000"))
}

func TestParseStaticcheckJSON_AbsolutePath(t *testing.T) {
	t.Parallel()

	input := `{"code":"S1001","severity":"warning","location":{"file":"/abs/path/main.go","line":1,"column":1},"message":"ok"}`
	findings := parseStaticcheckJSON([]byte(input), "/project")
	require.Len(t, findings, 1)

	assert.Equal(t, "/abs/path/main.go", findings[0].Position.File)
}
