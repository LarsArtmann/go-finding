package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// pipelineTestFinding creates a Finding with test values at the specified column.
func pipelineTestFinding(column int) finding.Finding {
	return finding.Finding{
		ID:          "test:1",
		Rule:        "test-rule",
		ToolName:    "test-tool",
		Message:     "test message",
		Severity:    finding.SeverityWarning,
		Position:    finding.Pos("test.go", 1, column),
		FixStrategy: finding.FixStrategySuggest,
		AfterCode:   "fixed()",
	}
}

// TestNew tests pipeline creation.
func TestNew(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	detector := &mockDetector{name: "test", findings: nil}

	p, err := New(config, "/tmp", detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}

	g.Expect(p.detectors).To(HaveLen(1))

	g.Expect(p.rootDir).To(Equal("/tmp"))

	if p.config.MaxIterations != config.MaxIterations {
		t.Error("config not set correctly")
	}
}

// TestPipelineRun_NoFindings tests that pipeline completes when no findings.
func TestPipelineRun_NoFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false
	detector := &mockDetector{name: "test", findings: nil}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Stable()).To(BeTrue())
	g.Expect(result.TotalIterations).To(Equal(1))
}

// TestPipelineRun_WithFindings tests pipeline with findings.
func TestPipelineRun_WithFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false
	config.MaxIterations = 3 // Limit to avoid running 5 iterations

	findings := []finding.Finding{pipelineTestFinding(1)}

	detector := &mockDetector{name: "test", findings: findings}
	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Stable()).To(BeFalse())
	g.Expect(result.TotalIterations).To(Equal(config.MaxIterations))

	g.Expect(result.Iterations).To(HaveLen(config.MaxIterations))

	iter := result.Iterations[0]
	g.Expect(iter.FindingsFound).To(Equal(1))

	g.Expect(iter.SuggestFixes).To(Equal(1))

	g.Expect(iter.SuggestedFindings()).To(HaveLen(1))
}

// TestPipelineRun_DetectorError tests error handling from detector.
func TestPipelineRun_DetectorError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false

	expectedErr := errors.New("detector failed")
	detector := &mockDetector{name: "test", err: expectedErr}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	_, err = p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, expectedErr)).To(BeTrue())
}

// TestDetectSequential_ContextCancellation verifies that detectSequential
// returns an error when the context is cancelled.
func TestDetectSequential_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	p := &Pipeline{
		detectors: []Detector{&mockDetector{name: "slow", delay: time.Second}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.detectSequential(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

// TestPipelineRun_ContextCancellation tests context cancellation.
func TestPipelineRun_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false

	// Detector that takes time
	detector := &mockDetector{
		name:  "slow",
		delay: 100 * time.Millisecond,
	}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

// TestPipelineRun_Timeout tests pipeline timeout.
func TestPipelineRun_Timeout(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false
	config.Timeout = 50 * time.Millisecond

	detector := slowTestDetector("slow", 500*time.Millisecond)

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	_, err = p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
}

// TestPipelineRun_MaxIterations tests max iteration limit.
func TestPipelineRun_MaxIterations(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.MaxIterations = 2
	config.ParallelDetectors = false

	// Findings that won't be auto-fixed (suggest strategy)
	findings := []finding.Finding{pipelineTestFinding(0)}

	detector := &mockDetector{name: "test", findings: findings}
	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.TotalIterations).To(Equal(config.MaxIterations))
}

// TestPipelineRun_Parallel tests parallel detection.
func TestPipelineRun_Parallel(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = true

	d1 := &mockDetector{
		name: "d1",
		findings: []finding.Finding{
			testFinding("1", "r1", "t1", "m1", finding.SeverityInfo, "a.go"),
		},
	}
	d2 := &mockDetector{
		name: "d2",
		findings: []finding.Finding{
			{
				ID:       "2",
				Rule:     "r2",
				ToolName: "t2",
				Message:  "m2",
				Severity: finding.SeverityWarning,
				Position: finding.Position{File: "b.go"},
			},
		},
	}

	p, err := New(config, t.TempDir(), d1, d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Iterations).NotTo(BeEmpty())

	assertFindingsFound(t, result, 2, "expected 2 findings")
}

// TestDetectParallel_SuppressionConsistency verifies that parallel and
// sequential detection produce identical results when findings are suppressed.
func TestDetectParallel_SuppressionConsistency(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	makeFindings := func() []finding.Finding {
		return []finding.Finding{
			{
				ID:       "s1",
				Rule:     "r1",
				ToolName: "t1",
				Message:  "active-1",
				Severity: finding.SeverityError,
			},
			{
				ID:       "s2",
				Rule:     "r2",
				ToolName: "t1",
				Message:  "suppressed-1",
				Severity: finding.SeverityWarning,
				Suppression: &finding.Suppression{
					Kind:   finding.SuppressionInSource,
					Rule:   "r2",
					Reason: "false positive",
				},
			},
			{
				ID:       "s3",
				Rule:     "r3",
				ToolName: "t1",
				Message:  "active-2",
				Severity: finding.SeverityInfo,
			},
			{
				ID:       "s4",
				Rule:     "r4",
				ToolName: "t1",
				Message:  "suppressed-2",
				Severity: finding.SeverityError,
				Suppression: &finding.Suppression{
					Kind:   finding.SuppressionInConfig,
					Rule:   "r4",
					Reason: "legacy",
				},
			},
		}
	}

	runDetect := func(parallel bool) ([]finding.Finding, []string) {
		var notified []string

		config := DefaultConfig()
		config.ParallelDetectors = parallel
		config.OnFinding = func(f finding.Finding) {
			notified = append(notified, f.ID)
		}

		d1 := &mockDetector{name: "d1", findings: makeFindings()}
		d2 := &mockDetector{name: "d2", findings: makeFindings()}

		p, err := New(config, t.TempDir(), d1, d2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		findings, err := p.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var all []finding.Finding
		for _, iter := range findings.Iterations {
			all = append(all, iter.Findings()...)
		}

		return all, notified
	}

	seqFindings, seqNotified := runDetect(false)
	parFindings, parNotified := runDetect(true)

	g.Expect(seqFindings).To(HaveLen(len(parFindings)))

	g.Expect(seqNotified).To(HaveLen(len(parNotified)))

	assertNoSuppressedFindings(t, seqFindings, "sequential")
	assertNoSuppressedFindings(t, parFindings, "parallel")
	assertNoSuppressedNotified(t, seqNotified, "sequential")
	assertNoSuppressedNotified(t, parNotified, "parallel")
}

// TestTriage tests the triage function.
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

// TestDetectorFunc tests the DetectorFunc adapter.
func TestDetectorFunc(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	called := false
	f := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		called = true

		return nil, nil
	})

	_, err := f.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("function was not called")
	}

	g.Expect(f.Name()).To(Equal("anonymous"))
}

func TestNamedDetectorFunc(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	fn := makeFindingDetectorFunc("test")

	d := NamedDetectorFunc("my-linter", fn)

	g.Expect(d.Name()).To(Equal("my-linter"))

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(findings).To(HaveLen(1))
	g.Expect(findings[0].ID).To(Equal("test"))
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	config := DefaultConfig()

	g.Expect(config.MaxIterations).To(Equal(5))

	g.Expect(config.ParallelDetectors).To(BeTrue())

	g.Expect(config.Timeout).To(Equal(10 * time.Minute))
}

// TestFixApplier tests the FixApplier.
func TestFixApplier(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")

	testContent := helloProgram
	if err := writeFile(testFile, []byte(testContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test backup
	err = applier.backup.Backup(testFile)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup exists
	backupPath := applier.backup.BackupPath(testFile)
	g.Expect(backupPath).NotTo(BeEmpty())

	// Test restore
	err = applier.backup.Restore(testFile)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Test Apply with empty fixes
	applied, err := applier.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("apply with nil fixes failed: %v", err)
	}

	g.Expect(applied).To(Equal(0))

	// Test Apply with fixes that have no Position.File
	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{}},
	}

	applied, err = applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(0))
}

// TestFixApplier_Apply tests actual fix application.
func TestFixApplier_Apply(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()
	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")

	originalContent := helloProgram
	writeTestFile(t, testFile, []byte(originalContent))

	// Create fix
	fixes := []finding.Finding{
		makeFixFinding("1", `println("hello")`, `println("world")`, "test.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	// Verify content changed
	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tprintln(\"world\")\n}\n"
	g.Expect(string(content)).To(Equal(expected))
}

func TestFixApplier_RangeBasedFix(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()
	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tprintln(\"hello\")\n}\n" //nolint:dupword // test fixture intentionally duplicates println
	if writeErr := writeFile(testFile, []byte(content), 0o644); writeErr != nil {
		t.Fatalf("create test file: %v", writeErr)
	}

	// Fix only the second "println" at line 5, using Range for precision.
	// Without Range, strings.Replace would hit the first occurrence at line 4.
	fixes := []finding.Finding{
		{
			ID:          "fix1",
			BeforeCode:  "println(\"hello\")",
			AfterCode:   "fmt.Println(\"world\")",
			Position:    finding.Pos("test.go", 5, 2),
			Range:       finding.NewRangePtr("test.go", 5, 2, 5, 18),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// First println unchanged, second replaced.
	want := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tfmt.Println(\"world\")\n}\n"
	g.Expect(string(got)).To(Equal(want))
}

func TestFixApplier_MultiLineRangeFix(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()
	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc old() {\n\treturn\n}\n\nfunc main() {}\n"
	writeTestFile(t, testFile, []byte(content))

	// Replace lines 3-5 (func old) with new content.
	fixes := []finding.Finding{
		{
			BeforeCode:  "func old() {\n\treturn\n}",
			AfterCode:   "func new() {\n\treturn 42\n}",
			Position:    finding.Position{File: "test.go", Line: 3},
			Range:       finding.NewRangePtr("test.go", 3, 1, 5, 2),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}\n"
	g.Expect(string(got)).To(Equal(want))
}

func TestFixApplier_RangeFixEmptyBeforeCode(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	tempDir := t.TempDir()
	applier, err := NewFixApplier(tempDir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")

	content := "package main\n\nfunc old() {\n\treturn\n}\n\nfunc main() {}\n"
	writeTestFile(t, testFile, []byte(content))

	// Replace lines 3-5 with new content using only Range + AfterCode (no BeforeCode).
	fixes := []finding.Finding{
		{
			AfterCode:   "func replaced() { // inserted\n}",
			Position:    finding.Position{File: "test.go", Line: 3},
			Range:       finding.NewRangePtr("test.go", 3, 1, 5, 2),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	g.Expect(applied).To(Equal(1))

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc replaced() { // inserted\n}\n\nfunc main() {}\n"
	g.Expect(string(got)).To(Equal(want))
}

// BenchmarkParallelDetection benchmarks parallel vs sequential detection.
func BenchmarkParallelDetection(b *testing.B) {
	for _, parallel := range []bool{false, true} {
		name := "sequential"
		if parallel {
			name = "parallel"
		}

		b.Run(name, func(b *testing.B) {
			config := DefaultConfig()
			config.ParallelDetectors = parallel

			// Create multiple slow detectors
			detectors := make([]Detector, 5)
			for i := range detectors {
				detectors[i] = &mockDetector{
					name:  "bench",
					delay: 1 * time.Millisecond,
					findings: []finding.Finding{
						testFinding("bench", "r", "t", "m", finding.SeverityInfo, "x.go"),
					},
				}
			}

			b.ResetTimer()

			for range b.N {
				p, err := New(config, b.TempDir(), detectors...)
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = p.Run(ctx)

				cancel()
			}
		})
	}
}

func TestIteration_Findings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	findings := []finding.Finding{
		{Message: "a", Severity: finding.SeverityError},
		{Message: "b", Severity: finding.SeverityWarning},
	}
	iter := Iteration{
		Number:   1,
		findings: findings,
		suggest:  []finding.Finding{{Message: "s", FixStrategy: finding.FixStrategySuggest}},
	}

	got := iter.Findings()
	g.Expect(got).To(HaveLen(2))

	assertFindingMessage := func(idx int, want string) {
		t.Helper()

		msg := got[idx].Message
		g.Expect(msg).To(Equal(want))
	}
	assertFindingMessage(0, "a")
	assertFindingMessage(1, "b")

	suggested := iter.SuggestedFindings()
	g.Expect(suggested).To(HaveLen(1))

	g.Expect(suggested[0].Message).To(Equal("s"))
}

func TestIteration_Findings_Empty(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	iter := Iteration{Number: 1}
	g.Expect(iter.Findings()).To(BeNil())

	g.Expect(iter.SuggestedFindings()).To(BeNil())
}

func TestIoErrorAt(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	err := ioErrorAt("read failed", os.ErrNotExist, "foo.go")

	var fe *finding.FindingError
	g.Expect(errors.As(err, &fe)).To(BeTrue())

	assertFindingErrorIO(t, fe, "foo.go")
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

	applied, err := p.applyDirectFixes(context.Background(), fixes)
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

	applied, err := p.applyDirectFixes(context.Background(), fixes)
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
	if err := os.WriteFile(testFile, []byte("package main\n"), 0o644); err != nil {
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

func TestNew_RejectsInvalidConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config Config
	}{
		{"negative iterations", Config{MaxIterations: -1}},
		{"negative timeout", Config{MaxIterations: 1, Timeout: -1 * time.Second}},
		{"invalid retry", Config{MaxIterations: 1, Retry: &RetryConfig{MaxRetries: -1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			_, err := New(tt.config, t.TempDir())
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNew_ValidConfig_NoError(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	config := DefaultConfig()
	p, err := New(config, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(p).NotTo(BeNil())
}

func TestPipelineRun_PartialErrorsSurfaced(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	goodDetector := mockDetectorWithFinding("good", "F1", "r", "good", "m")
	badDetector := &mockDetector{name: "bad", err: errors.New("boom")}

	config := Config{
		MaxIterations:       1,
		GracefulDegradation: true,
	}

	p, err := New(config, t.TempDir(), goodDetector, badDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.PartialErrors).NotTo(BeNil())

	_, ok := result.PartialErrors["bad"]
	g.Expect(ok).To(BeTrue())

	_, ok = result.PartialErrors["good"]
	g.Expect(ok).To(BeFalse())
}

func TestPipelineRun_MetricsInResult(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	m := NewMetrics()
	config := Config{
		MaxIterations: 1,
		Metrics:       m,
	}

	detector := &mockDetector{name: "test", findings: nil}
	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.Metrics.TotalDuration).NotTo(BeZero())

	g.Expect(result.Metrics.FixesApplied).To(Equal(0))
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

	var conflictFindings []finding.Finding
	var appliedFindings []finding.Finding

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

	err := p.applyTriage(context.Background(), nil, iter)
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
	iter := &Iteration{Number: 1}

	err := p.applyTriage(context.Background(), fixes, iter)
	if err != nil {
		t.Fatalf("applyTriage: %v", err)
	}

	g.Expect(iter.Conflicts).To(Equal(1))

	g.Expect(iter.Applied).To(Equal(1))
}

func TestPipelineRun_DirectFixStabilizes(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n"
	writeTestFile(t, testFile, []byte(original))

	fix := makeFixFinding("fix1", "old()", "new()", "fixme.go", 4)

	callCount := 0
	det := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		callCount++
		if callCount == 1 {
			return []finding.Finding{fix}, nil
		}

		return nil, nil
	})

	cfg := Config{MaxIterations: 5, ParallelDetectors: false}
	p, err := New(cfg, tmpDir, NamedDetectorFunc("tool", det))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Stable()).To(BeTrue())

	g.Expect(result.TotalIterations).To(Equal(2))

	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	g.Expect(string(content)).To(Equal(want))
}

func testContextErrorCancels(t *testing.T, parallel bool) {
	t.Helper()
	g := NewWithT(t)

	goodDetector := &mockDetector{
		name:     "good",
		findings: []finding.Finding{{ID: "f1", Rule: "r", ToolName: "t", Message: "m"}},
	}
	badDetector := &mockDetector{
		name: "bad",
		err:  context.Canceled,
	}

	config := DefaultConfig()
	config.GracefulDegradation = true
	config.ParallelDetectors = parallel

	p, err := New(config, t.TempDir(), goodDetector, badDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())

	g.Expect(result.PartialErrors).
		To(BeNil(), "context errors should not be stored as partial errors")
}

func TestDetectPartialSequential_ContextErrorPropagates(t *testing.T) {
	t.Parallel()
	testContextErrorCancels(t, false)
}

func TestDetectPartialParallel_ContextErrorPropagates(t *testing.T) {
	t.Parallel()
	testContextErrorCancels(t, true)
}

func TestVerify_ContextCancellation(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	slowDetector := &mockDetector{name: "slow", delay: time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Verify(ctx, []Detector{slowDetector}, nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.Canceled)).To(BeTrue())
}

func TestIsContextError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"canceled", context.Canceled, true},
		{"deadline", context.DeadlineExceeded, true},
		{"wrapped canceled", fmt.Errorf("wrap: %w", context.Canceled), true},
		{"other error", errors.New("other"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsContextError(tt.err); got != tt.want {
				t.Errorf("IsContextError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func mockDetectorWithFinding(name, id, rule, tool, msg string) *mockDetector {
	return &mockDetector{
		name: name,
		findings: []finding.Finding{
			{ID: id, Rule: rule, ToolName: tool, Message: msg, Severity: finding.SeverityError},
		},
	}
}

func correlateTestDetectors() (*mockDetector, *mockDetector) {
	detA := &mockDetector{
		name: "tool-a",
		findings: []finding.Finding{
			{
				ID: "a1", Rule: "R1", ToolName: "tool-a", Message: "msg a",
				Severity: finding.SeverityError, Position: finding.Pos("file.go", 10, 1),
			},
		},
	}
	detB := &mockDetector{
		name: "tool-b",
		findings: []finding.Finding{
			{
				ID: "b1", Rule: "R2", ToolName: "tool-b", Message: "msg b",
				Severity: finding.SeverityWarning, Position: finding.Pos("file.go", 12, 1),
			},
		},
	}
	return detA, detB
}

func TestPipelineRun_PerDetectorTimeout(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ParallelDetectors = false
	cfg.DetectorTimeouts = map[string]time.Duration{
		"slow": 10 * time.Millisecond,
	}

	slowDetector := slowTestDetector("slow", 500*time.Millisecond)

	p, err := New(cfg, t.TempDir(), slowDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	_, err = p.Run(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
}

func TestPipelineRun_PerDetectorTimeout_UnaffectedDetector(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ParallelDetectors = false
	cfg.DetectorTimeouts = map[string]time.Duration{
		"other": 10 * time.Millisecond, // timeout for a different detector
	}

	slowDetector := slowMockDetector(
		"slow", 50*time.Millisecond,
		finding.Finding{ID: "s1", Rule: "r", ToolName: "t", Message: "m"},
	)

	p, err := New(cfg, t.TempDir(), slowDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	result, err := p.Run(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result.Iterations).NotTo(BeEmpty())
	g.Expect(result.Iterations[0].FindingsFound).To(Equal(1))
}

func TestPipelineRun_CorrelateFindings(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	detA, detB := correlateTestDetectors()

	cfg := Config{MaxIterations: 1, ParallelDetectors: false, CorrelateFindings: true}
	p, err := New(cfg, ".", detA, detB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Correlations).NotTo(BeEmpty())
	g.Expect(result.Correlations[0].FindingIDs).To(Equal([]string{"a1", "b1"}))
}

func TestPipelineRun_NoCorrelateWhenDisabled(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	detA, detB := correlateTestDetectors()

	cfg := Config{MaxIterations: 1, ParallelDetectors: false, CorrelateFindings: false}
	p, err := New(cfg, ".", detA, detB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(result.Correlations).To(BeEmpty())
}

func TestPipelineRun_StructuredLogging(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := DefaultConfig()
	cfg.ParallelDetectors = false
	cfg.MaxIterations = 2
	cfg.Logger = logger

	findings := []finding.Finding{pipelineTestFinding(1)}
	detector := &mockDetector{name: "test", findings: findings}

	p, err := New(cfg, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	g.Expect(output).NotTo(BeEmpty())

	// Verify structured log messages are emitted
	type logEntry struct {
		Msg string `json:"msg"`
	}

	var entries []logEntry
	for _, line := range splitLines(output) {
		if line == "" {
			continue
		}

		var entry logEntry
		if json.Unmarshal([]byte(line), &entry) == nil {
			entries = append(entries, entry)
		}
	}

	messages := make(map[string]bool)
	for _, e := range entries {
		messages[e.Msg] = true
	}

	g.Expect(messages["iteration starting"]).To(BeTrue())
	g.Expect(messages["triage complete"]).To(BeTrue())
}

func TestPipelineRun_OnStage(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	var stages []string
	cfg := DefaultConfig()
	cfg.ParallelDetectors = false
	cfg.MaxIterations = 1
	cfg.OnStage = func(stage Stage, iteration, count int) {
		stages = append(stages, fmt.Sprintf("%s:%d:%d", stage, iteration, count))
	}

	findings := []finding.Finding{pipelineTestFinding(1)}
	detector := &mockDetector{name: "test", findings: findings}

	p, err := New(cfg, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	g.Expect(stages).NotTo(BeEmpty())
	g.Expect(stages[0]).To(ContainSubstring("detect:1:"))
	g.Expect(stages).To(ContainElement(ContainSubstring("triage:1:")))
}

func splitLines(s string) []string {
	var lines []string
	for line := range strings.SplitSeq(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func TestPipelineRun_SingleUse(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	p, err := New(cfg, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Run(context.Background())
	if err == nil {
		t.Fatal("expected error on second Run call")
	}
}

func TestPipelineRun_SingleUse_ErrorMessage(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	p, err := New(cfg, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	_, _ = p.Run(context.Background())
	_, err = p.Run(context.Background())

	if err.Error() != "pipeline: Run already called; create a new Pipeline for each invocation" {
		t.Errorf("unexpected error message: %v", err)
	}
}
