package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestTriage(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []finding.Finding{
		{ID: "1", FixStrategy: finding.FixStrategyDirect, BeforeCode: "old", AfterCode: "new"},
		{ID: "2", FixStrategy: finding.FixStrategySuggest, AfterCode: "new"},
		{ID: "3", FixStrategy: finding.FixStrategyAI, AfterCode: "new"},
		{ID: "4", FixStrategy: finding.FixStrategyNone},
		{ID: "5", FixStrategy: ""},
	}

	p := &Pipeline{config: DefaultConfig()}
	result := p.triage(findings)

	g.Expect(result.Direct).To(HaveLen(1))

	g.Expect(result.Suggest).To(HaveLen(2))

	g.Expect(result.None).To(HaveLen(2))
}

func TestApplyDirectFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fixes := directFixFinding()

	m := NewMetrics()
	p := &Pipeline{
		config:  Config{Metrics: m},
		rootDir: tempDir,
		metrics: m,
	}

	p.applier, _ = NewFixApplier(tempDir)

	applied, _, err := p.applyDirectFixes(context.Background(), fixes)
	if err != nil {
		t.Fatalf("applyDirectFixes: %v", err)
	}

	g.Expect(applied).To(HaveLen(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	g.Expect(string(got)).To(Equal(want))

	g.Expect(m.TotalFixesApplied()).To(Equal(1))
}

func TestApplyDirectFixes_NoMetrics(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fixes := directFixFinding()

	p := &Pipeline{
		config:  DefaultConfig(),
		rootDir: tempDir,
	}

	var err error

	p.applier, err = NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("create fix applier: %v", err)
	}

	applied, _, err := p.applyDirectFixes(context.Background(), fixes)
	if err != nil {
		t.Fatalf("applyDirectFixes: %v", err)
	}

	g.Expect(applied).To(HaveLen(1))
}

func TestDryRun(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte("package main\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	f := testFinding("1", "r1", "tool", "msg", finding.SeverityError, testFile)
	f.FixStrategy = finding.FixStrategyDirect
	f.BeforeCode = "package main"
	f.AfterCode = "package main // fixed"

	applied := 0
	cfg := Config{
		MaxIterations: 1,
		DryRun:        true,
		OnFix: func(_ finding.Finding, wasApplied bool) {
			if wasApplied {
				applied++
			}
		},
	}

	det := mockDetWithFindings("tool", f)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(applied).To(Equal(0))

	assertIterationsLen(t, result, 1)

	iter := result.Iterations[0]
	g.Expect(iter.DirectFixes).To(Equal(1))

	data, _ := os.ReadFile(testFile)
	g.Expect(string(data)).To(Equal("package main\n"))
}

func TestApplyTriage_DirectFixesApplied(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fix := finding.Finding{
		ID:          "fix1",
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "replace old with new",
		Severity:    finding.SeverityWarning,
		BeforeCode:  "old()",
		AfterCode:   "new()",
		Position:    finding.Position{File: "fixme.go", Line: 4},
		FixStrategy: finding.FixStrategyDirect,
	}

	det := mockDetWithFindings("tool", fix)

	p, err := New(Config{MaxIterations: 3, ParallelDetectors: false}, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Iterations).NotTo(BeEmpty())

	iter := result.Iterations[0]
	g.Expect(iter.DirectFixes).To(Equal(1))

	g.Expect(iter.Applied).To(Equal(1))
}

func TestApplyTriage_ConflictingFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fixes := makeConflictingFixes()

	var (
		conflictFindings []finding.Finding
		appliedFindings  []finding.Finding
	)

	cfg := Config{
		MaxIterations:       1,
		ParallelDetectors:   false,
		GracefulDegradation: true,
		OnFix: func(f finding.Finding, wasApplied bool) {
			if wasApplied {
				appliedFindings = append(appliedFindings, f)
			} else {
				conflictFindings = append(conflictFindings, f)
			}
		},
	}

	det := &mockDetector{name: "tool", findings: fixes}

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Iterations).NotTo(BeEmpty())

	iter := result.Iterations[0]
	g.Expect(iter.Conflicts).NotTo(BeZero())

	g.Expect(conflictFindings).NotTo(BeEmpty())
}

func TestApplyTriage_OnFixCallback(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fix := makeFixFinding("fix1", "old()", "new()", "fixme.go", 4)

	var appliedIDs []string

	cfg := Config{
		MaxIterations:     2,
		ParallelDetectors: false,
		OnFix:             collectingOnFix(&appliedIDs),
	}

	det := mockDetWithFindings("tool", fix)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(appliedIDs).NotTo(BeEmpty())

	g.Expect(appliedIDs[0]).To(Equal("fix1"))
}

func TestApplyTriage_EmptyFixes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	p := &Pipeline{config: DefaultConfig()}
	iter := &Iteration{Number: 1}

	err := p.applyTriage(context.Background(), nil, iter, &PipelineResult{})
	if err != nil {
		t.Fatalf("applyTriage with nil fixes: %v", err)
	}

	g.Expect(iter.Applied).To(Equal(0))

	g.Expect(iter.Conflicts).To(Equal(0))
}

func TestApplyTriage_AllConflicting(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fixes := makeConflictingFixes()

	p := &Pipeline{config: DefaultConfig(), rootDir: tmpDir}
	p.applier, _ = NewFixApplier(tmpDir)

	iter := &Iteration{Number: 1}

	err := p.applyTriage(context.Background(), fixes, iter, &PipelineResult{})
	if err != nil {
		t.Fatalf("applyTriage: %v", err)
	}

	g.Expect(iter.Conflicts).To(Equal(1))

	g.Expect(iter.Applied).To(Equal(1))
}

func directFixFinding() []finding.Finding {
	return []finding.Finding{
		{
			ID:          "fix1",
			BeforeCode:  "old()",
			AfterCode:   "new()",
			Position:    finding.Position{File: "fixme.go", Line: 4},
			FixStrategy: finding.FixStrategyDirect,
		},
	}
}
