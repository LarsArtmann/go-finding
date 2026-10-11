package finding

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

var errDetectBoom = errors.New("detect boom")

type coverageDetector struct {
	name     string
	findings []Finding
	err      error
}

func (s *coverageDetector) Name() string { return s.name }

func (s *coverageDetector) Detect(context.Context) ([]Finding, error) {
	return s.findings, s.err
}

type coverageDetectorV2 struct {
	coverageDetector

	result DetectResult
}

func (s *coverageDetectorV2) DetectV2(context.Context) (DetectResult, error) {
	return s.result, s.err
}

func coverageFinding(file string) Finding {
	return NewFinding("rule-1", "testtool", "test finding", SeverityWarning, Position{File: FilePath(file), Line: 1}, ConfidenceHigh)
}

func TestRunDetectPrefersV2(t *testing.T) {
	d := &coverageDetectorV2{
		coverageDetector: coverageDetector{name: "v2"},
		result: DetectResult{
			Findings:       []Finding{coverageFinding("a.go")},
			FilesScanned:   42,
			SkippedModules: []string{"nested/module"},
		},
	}

	res, err := RunDetect(context.Background(), d)
	if err != nil {
		t.Fatalf("RunDetect: %v", err)
	}

	if res.FilesScanned != 42 {
		t.Errorf("FilesScanned = %d, want 42", res.FilesScanned)
	}

	if len(res.Findings) != 1 || len(res.SkippedModules) != 1 {
		t.Errorf("Findings = %d, SkippedModules = %d; want 1 and 1", len(res.Findings), len(res.SkippedModules))
	}
}

func TestRunDetectFallsBackToV1WithZeroedCoverage(t *testing.T) {
	d := &coverageDetector{name: "v1", findings: []Finding{coverageFinding("b.go")}}

	res, err := RunDetect(context.Background(), d)
	if err != nil {
		t.Fatalf("RunDetect: %v", err)
	}

	if res.FilesScanned != 0 {
		t.Errorf("v1 FilesScanned = %d, want 0 (documented zeroed-coverage contract)", res.FilesScanned)
	}

	if len(res.SkippedModules) != 0 {
		t.Errorf("v1 SkippedModules = %v, want empty", res.SkippedModules)
	}

	if len(res.Findings) != 1 {
		t.Errorf("v1 Findings = %d, want 1", len(res.Findings))
	}
}

func TestRunDetectPropagatesErrors(t *testing.T) {
	v1 := &coverageDetector{name: "v1", err: errDetectBoom}
	if _, err := RunDetect(context.Background(), v1); !errors.Is(err, errDetectBoom) {
		t.Errorf("v1 error = %v, want errDetectBoom", err)
	}

	v2 := &coverageDetectorV2{coverageDetector: coverageDetector{name: "v2", err: errDetectBoom}}
	if _, err := RunDetect(context.Background(), v2); !errors.Is(err, errDetectBoom) {
		t.Errorf("v2 error = %v, want errDetectBoom", err)
	}
}

func TestAddDetectResultAccumulatesCoverage(t *testing.T) {
	r := NewReport(ToolInfo{Name: "tool"})

	r.AddDetectResult(DetectResult{
		Findings:       []Finding{coverageFinding("a.go")},
		FilesScanned:   10,
		SkippedModules: []string{"mod/a", "mod/b"},
	})
	r.AddDetectResult(DetectResult{
		Findings:       []Finding{coverageFinding("b.go")},
		FilesScanned:   5,
		SkippedModules: []string{"mod/b", "mod/c"},
	})

	if r.Summary.FilesScanned != 15 {
		t.Errorf("FilesScanned = %d, want 15 (accumulated)", r.Summary.FilesScanned)
	}

	want := []string{"mod/a", "mod/b", "mod/c"}
	if len(r.Summary.SkippedModules) != len(want) {
		t.Fatalf("SkippedModules = %v, want %v", r.Summary.SkippedModules, want)
	}

	for i, m := range want {
		if r.Summary.SkippedModules[i] != m {
			t.Errorf("SkippedModules[%d] = %q, want %q (first-seen order, deduped)", i, r.Summary.SkippedModules[i], m)
		}
	}

	r.ComputeSummaryAt(time.Now())

	if r.Summary.Total != 2 {
		t.Errorf("Total = %d, want 2 findings", r.Summary.Total)
	}

	if r.Summary.FilesScanned != 15 {
		t.Errorf("FilesScanned after ComputeSummary = %d, want 15", r.Summary.FilesScanned)
	}
}

func TestComputeSummaryPreservesCoverage(t *testing.T) {
	r := NewReport(ToolInfo{Name: "tool"})
	r.AddDetectResult(DetectResult{FilesScanned: 7, SkippedModules: []string{"nested"}})

	r.ComputeSummaryAt(time.Now())

	if r.Summary.FilesScanned != 7 {
		t.Errorf("FilesScanned after ComputeSummary = %d, want 7 (coverage must survive recomputation)", r.Summary.FilesScanned)
	}

	if len(r.Summary.SkippedModules) != 1 || r.Summary.SkippedModules[0] != "nested" {
		t.Errorf("SkippedModules after ComputeSummary = %v, want [nested]", r.Summary.SkippedModules)
	}
}

func TestAddDetectResultConcurrent(t *testing.T) {
	r := NewReport(ToolInfo{Name: "tool"})

	const workers = 8
	const runsPerWorker = 50

	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range runsPerWorker {
				r.AddDetectResult(DetectResult{FilesScanned: 1})
			}
		}()
	}

	wg.Wait()

	if r.Summary.FilesScanned != workers*runsPerWorker {
		t.Errorf("FilesScanned = %d, want %d (no lost updates under -race)", r.Summary.FilesScanned, workers*runsPerWorker)
	}
}
