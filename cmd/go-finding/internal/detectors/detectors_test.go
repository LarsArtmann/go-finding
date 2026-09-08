package detectors

import (
	"context"
	"os"
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
			g := NewParallelGomega(t)

			pos := parsePosn(tt.posn, tt.dir)
			g.Expect(pos.File).To(Equal(finding.FilePath(tt.wantFile)))
			g.Expect(pos.Line).To(Equal(tt.wantLine))
			g.Expect(pos.Column).To(Equal(tt.wantCol))
		})
	}
}

func TestParseGoVetJSON(t *testing.T) {
	g := NewParallelGomega(t)

	input := `{
		"github.com/example/pkg": [
			{"posn": "main.go:10:5", "message": "unused variable x"},
			{"posn": "main.go:20:1", "message": "unreachable code"}
		]
	}`

	findings := parseGoVetJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(2))

	f := findings[0]
	g.Expect(f.ToolName).To(Equal(finding.ToolName("govet")))
	g.Expect(f.Rule).To(Equal(finding.RuleName("github.com/example/pkg")))
	g.Expect(f.Message).To(Equal("unused variable x"))
	g.Expect(f.Severity).To(Equal(finding.SeverityWarning))
	g.Expect(f.Category).To(Equal(finding.CategoryCorrectness))
	g.Expect(f.FixStrategy).To(Equal(finding.FixStrategySuggest))
	g.Expect(f.Position.File).To(Equal(finding.FilePath("/project/main.go")))
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
			g := NewParallelGomega(t)

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
	g := NewParallelGomega(t)

	input := `{"code":"SA1000","severity":"warning","location":{"file":"main.go","line":10,"column":5},"message":"invalid regex pattern"}
{"code":"S1001","severity":"error","location":{"file":"util.go","line":20,"column":1},"message":"should use copy"}`

	findings := parseStaticcheckJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(2))

	f := findings[0]
	g.Expect(f.ToolName).To(Equal(finding.ToolName("staticcheck")))
	g.Expect(f.Rule).To(Equal(finding.RuleName("SA1000")))
	g.Expect(f.Severity).To(Equal(finding.SeverityWarning))
	g.Expect(f.Category).To(Equal(finding.CategoryCorrectness))
	g.Expect(f.Confidence).To(BeNumerically("~", 0.8, 0.001))

	f2 := findings[1]
	g.Expect(f2.Rule).To(Equal(finding.RuleName("S1001")))
	g.Expect(f2.Severity).To(Equal(finding.SeverityError))
	g.Expect(f2.Category).To(Equal(finding.CategoryStyle))
}

func TestParseStaticcheckJSON_Empty(t *testing.T) {
	g := NewParallelGomega(t)

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
			g := NewParallelGomega(t)

			findings := parseStaticcheckJSON([]byte(tt.input), "")
			g.Expect(findings).To(HaveLen(tt.wantLen))
			g.Expect(findings[0].Rule).To(Equal(finding.RuleName(tt.wantRule)))
		})
	}
}

func TestStaticcheckCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want finding.Category
	}{
		{"SA1000", finding.CategoryCorrectness},
		{"S1001", finding.CategoryStyle},
		{"QF1001", finding.CategoryStyle},
		{"U1000", finding.CategoryUnused},
		{"PERF1001", finding.CategoryPerformance},
		{"R1001", finding.CategoryPerformance},
		{"F1001", finding.CategoryPerformance},
		{"SA", finding.CategoryCorrectness},
		{"", finding.CategoryCorrectness},
		{"X9999", finding.CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			g := NewParallelGomega(t)

			g.Expect(staticcheckCategory(tt.code)).To(Equal(tt.want))
		})
	}
}

func TestDetectorNames(t *testing.T) {
	g := NewParallelGomega(t)

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
			g := NewParallelGomega(t)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := tt.det.Detect(ctx)
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNewGoVetDetector_ValidProject(t *testing.T) {
	g := NewParallelGomega(t)
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
	g := NewParallelGomega(t)

	input := `{"github.com/example/pkg": "not an array"}`
	findings := parseGoVetJSON([]byte(input), "")
	g.Expect(findings).To(BeNil())
}

func TestStaticcheckCategory_A(t *testing.T) {
	g := NewParallelGomega(t)

	g.Expect(staticcheckCategory("A1000")).To(Equal(finding.CategoryCorrectness))
}

func TestParseStaticcheckJSON_AbsolutePath(t *testing.T) {
	g := NewParallelGomega(t)

	input := `{"code":"S1001","severity":"warning","location":{"file":"/abs/path/main.go","line":1,"column":1},"message":"ok"}`
	findings := parseStaticcheckJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(1))

	g.Expect(findings[0].Position.File).To(Equal(finding.FilePath("/abs/path/main.go")))
}

func TestParseStaticcheckJSON_FixExtension(t *testing.T) {
	g := NewParallelGomega(t)

	input := `{"code":"ST1000","severity":"warning","location":{"file":"main.go","line":4,"column":2},"message":"old() should be new()","before":"old()","after":"new()"}`
	findings := parseStaticcheckJSON([]byte(input), "/project")
	g.Expect(findings).To(HaveLen(1))

	f := findings[0]
	g.Expect(f.FixStrategy).To(Equal(finding.FixStrategyDirect))
	g.Expect(f.BeforeCode).To(Equal("old()"))
	g.Expect(f.AfterCode).To(Equal("new()"))
	g.Expect(f.IsAutoFixable()).To(BeTrue())
}

func TestParseStaticcheckJSON_FixExtensionRequiresBoth(t *testing.T) {
	g := NewParallelGomega(t)

	input := `{"code":"S1000","severity":"warning","location":{"file":"main.go","line":4,"column":2},"message":"half a fix","before":"old()"}`
	findings := parseStaticcheckJSON([]byte(input), "")
	g.Expect(findings).To(HaveLen(1))

	g.Expect(findings[0].FixStrategy).To(Equal(finding.FixStrategySuggest),
		"before without after must stay suggest-only")
	g.Expect(findings[0].BeforeCode).To(BeEmpty())
}

// TestParseStaticcheckJSON_RealCorpus parses REAL staticcheck 2026.2.1 output
// (captured with `staticcheck -f json ./...` against the flawed fixture in
// testdata/staticcheck-corpus/; location paths relativized for portability —
// see corpus.go for regeneration instructions). Guards the parser against
// real-output drift: field layout, severity mapping, category mapping, and
// the absence of before/after in genuine tool output.
func TestParseStaticcheckJSON_RealCorpus(t *testing.T) {
	g := NewParallelGomega(t)

	data, err := os.ReadFile("testdata/staticcheck-corpus/corpus.jsonl")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).NotTo(BeEmpty(), "corpus fixture must exist")

	findings := parseStaticcheckJSON(data, "")
	g.Expect(findings).To(HaveLen(3), "real corpus carries exactly three findings")

	byCode := map[finding.RuleName]finding.Finding{}
	for _, f := range findings {
		byCode[f.Rule] = f
	}
	g.Expect(byCode).To(HaveKey(finding.RuleName("S1002")))
	g.Expect(byCode).To(HaveKey(finding.RuleName("SA4006")))
	g.Expect(byCode).To(HaveKey(finding.RuleName("S1025")))

	for code, f := range byCode {
		g.Expect(f.ToolName).To(Equal(finding.ToolName("staticcheck")), code)
		g.Expect(f.Severity).To(Equal(finding.SeverityError),
			"%s: real 2026.2.1 output marks all three as error", code)
		g.Expect(f.Message).NotTo(BeEmpty(), code)
		g.Expect(f.Position.File).To(Equal(finding.FilePath("corpus.go")), code)
		g.Expect(f.Position.Line).To(BeNumerically(">", 0), code)
		g.Expect(f.Position.Column).To(BeNumerically(">", 0), code)
		g.Expect(f.FixStrategy).To(Equal(finding.FixStrategySuggest),
			"%s: real staticcheck output has no before/after", code)
		g.Expect(f.BeforeCode).To(BeEmpty(), code)
		g.Expect(f.AfterCode).To(BeEmpty(), code)
	}

	g.Expect(byCode[finding.RuleName("SA4006")].Category).
		To(Equal(finding.CategoryCorrectness), "SA* codes are correctness")
	g.Expect(byCode[finding.RuleName("S1002")].Category).
		To(Equal(finding.CategoryStyle), "S* (non-SA) codes are style")
	g.Expect(byCode[finding.RuleName("S1025")].Category).
		To(Equal(finding.CategoryStyle))

	g.Expect(byCode[finding.RuleName("S1002")].Position.Line).To(Equal(15))
	g.Expect(byCode[finding.RuleName("SA4006")].Position.Line).To(Equal(20))
	g.Expect(byCode[finding.RuleName("S1025")].Position.Line).To(Equal(27))
}
