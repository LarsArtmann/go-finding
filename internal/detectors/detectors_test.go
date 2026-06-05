package detectors

import (
	"context"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

const (
	detectorsTestProject = "/project"
	detectorsTestS1001   = "S1001"
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
		{"full position", "main.go:10:5", detectorsTestProject, "/project/main.go", 10, 5},
		{"file and line only", "main.go:10", detectorsTestProject, "/project/main.go", 10, 0},
		{
			"absolute path ignored dir",
			"/abs/path/main.go:5:1",
			detectorsTestProject,
			"/abs/path/main.go",
			5,
			1,
		},
		{"no colon returns raw string as file", "just-a-file.go", "", "just-a-file.go", 0, 0},
		{"empty dir with relative path", "pkg/util.go:20:3", "", "pkg/util.go", 20, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			pos := parsePosn(tt.posn, tt.dir)
			g.Expect(pos.File).To(Equal(tt.wantFile))
			g.Expect(pos.Line).To(Equal(tt.wantLine))
			g.Expect(pos.Column).To(Equal(tt.wantCol))
		})
	}
}

func TestParseGoVetJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := `{
		"github.com/example/pkg": [
			{"posn": "main.go:10:5", "message": "unused variable x"},
			{"posn": "main.go:20:1", "message": "unreachable code"}
		]
	}`

	findings := parseGoVetJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(2))

	f := findings[0]
	g.Expect(f.ToolName).To(Equal("govet"))
	g.Expect(f.Rule).To(Equal("github.com/example/pkg"))
	g.Expect(f.Message).To(Equal("unused variable x"))
	g.Expect(f.Severity).To(Equal(finding.SeverityWarning))
	g.Expect(f.Category).To(Equal(finding.CategoryCorrectness))
	g.Expect(f.FixStrategy).To(Equal(finding.FixStrategySuggest))
	g.Expect(f.Position.File).To(Equal("/project/main.go"))
	g.Expect(f.Position.Line).To(Equal(10))
}

func TestParseGoVetJSON_InvalidAndEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		empty bool
	}{
		{"invalid JSON", "not json", false},
		{"empty JSON", "{}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			findings := parseGoVetJSON([]byte(tt.input), "")
			if tt.empty {
				g.Expect(findings).To(BeEmpty())
			} else {
				g.Expect(findings).To(BeNil())
			}
		})
	}
}

func TestParseStaticcheckJSON(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := `{"code":"SA1000","severity":"warning","location":{"file":"main.go","line":10,"column":5},"message":"invalid regex pattern"}
{"code":"S1001","severity":"error","location":{"file":"util.go","line":20,"column":1},"message":"should use copy"}`

	findings := parseStaticcheckJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(2))

	f := findings[0]
	g.Expect(f.ToolName).To(Equal("staticcheck"))
	g.Expect(f.Rule).To(Equal("SA1000"))
	g.Expect(f.Severity).To(Equal(finding.SeverityWarning))
	g.Expect(f.Category).To(Equal(finding.CategoryStyle))
	g.Expect(f.Confidence).To(BeNumerically("~", 0.8, 0.001))

	f2 := findings[1]
	g.Expect(f2.Rule).To(Equal("S1001"))
	g.Expect(f2.Severity).To(Equal(finding.SeverityError))
	g.Expect(f2.Category).To(Equal(finding.CategoryStyle))
}

func TestParseStaticcheckJSON_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	findings := parseStaticcheckJSON(nil, "")
	g.Expect(findings).To(BeNil())
}

func TestParseStaticcheckJSON_LineSkipping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantRule string
	}{
		{
			"invalid line skipped",
			"not json at all\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}",
			1,
			detectorsTestS1001,
		},
		{
			"whitespace lines skipped",
			"\n  \n\t\n{\"code\":\"S1001\",\"severity\":\"warning\",\"location\":{\"file\":\"a.go\",\"line\":1,\"column\":1},\"message\":\"ok\"}\n\n",
			1,
			"S1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			findings := parseStaticcheckJSON([]byte(tt.input), "")
			g.Expect(findings).To(HaveLen(tt.wantLen))
			g.Expect(findings[0].Rule).To(Equal(tt.wantRule))
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
			g := NewWithT(t)

			g.Expect(staticcheckCategory(tt.code)).To(Equal(tt.want))
		})
	}
}

func TestDetectorNames(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	govet := NewGoVetDetector(".")
	g.Expect(govet.Name()).To(Equal("govet"))

	sc := NewStaticcheckDetector(".")
	g.Expect(sc.Name()).To(Equal("staticcheck"))
}

func TestDetector_CancelledContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		det  pipeline.Detector
	}{
		{"govet", NewGoVetDetector(".")},
		{"staticcheck", NewStaticcheckDetector(".")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := tt.det.Detect(ctx)
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNewGoVetDetector_ValidProject(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	d := NewGoVetDetector("../../")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findings, err := d.Detect(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	t.Logf("govet found %d findings", len(findings))
}

func TestParseGoVetJSON_BadEntry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := `{"github.com/example/pkg": "not an array"}`
	findings := parseGoVetJSON([]byte(input), "")
	g.Expect(findings).To(BeNil())
}

func TestStaticcheckCategory_A(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	g.Expect(staticcheckCategory("A1000")).To(Equal(finding.CategoryCorrectness))
}

func TestParseStaticcheckJSON_AbsolutePath(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	input := `{"code":"S1001","severity":"warning","location":{"file":"/abs/path/main.go","line":1,"column":1},"message":"ok"}`
	findings := parseStaticcheckJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(1))

	g.Expect(findings[0].Position.File).To(Equal("/abs/path/main.go"))
}
