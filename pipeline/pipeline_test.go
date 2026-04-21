package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

// TestNew tests pipeline creation.
func TestNew(t *testing.T) {
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

	if len(p.detectors) != 1 {
		t.Errorf("expected 1 detector, got %d", len(p.detectors))
	}

	if p.rootDir != "/tmp" {
		t.Errorf("expected rootDir /tmp, got %s", p.rootDir)
	}

	if p.config.MaxIterations != config.MaxIterations {
		t.Error("config not set correctly")
	}
}

// TestPipelineRun_NoFindings tests that pipeline completes when no findings.
func TestPipelineRun_NoFindings(t *testing.T) {
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

	if !result.Stable {
		t.Error("expected stable result")
	}
	// One iteration is recorded even when no findings (the check iteration)
	if result.TotalIterations != 1 {
		t.Errorf("expected 1 iteration (detection check), got %d", result.TotalIterations)
	}
}

// TestPipelineRun_WithFindings tests pipeline with findings.
func TestPipelineRun_WithFindings(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false
	config.MaxIterations = 3 // Limit to avoid running 5 iterations

	findings := []finding.Finding{
		{
			ID:          "test:1",
			Rule:        "test-rule",
			ToolName:    "test-tool",
			Message:     "test message",
			Severity:    finding.SeverityWarning,
			Position:    finding.Position{File: "test.go", Line: 1, Column: 1},
			FixStrategy: finding.FixStrategySuggest,
		},
	}

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

	if result.Stable {
		t.Error("expected non-stable result (findings present but not auto-fixed)")
	}
	// Should run until max iterations since findings persist and can't be auto-fixed
	{
		iterations := result.TotalIterations
		if iterations != config.MaxIterations {
			t.Errorf("expected %d iterations, got %d", config.MaxIterations, iterations)
		}
	}

	if len(result.Iterations) != config.MaxIterations {
		t.Fatalf(
			"expected %d iteration records, got %d",
			config.MaxIterations,
			len(result.Iterations),
		)
	}
	// Check first iteration
	iter := result.Iterations[0]
	if iter.FindingsFound != 1 {
		t.Errorf("expected 1 finding found, got %d", iter.FindingsFound)
	}

	if iter.SuggestFixes != 1 {
		t.Errorf("expected 1 suggest-fix, got %d", iter.SuggestFixes)
	}

	if len(iter.SuggestedFindings()) != 1 {
		t.Errorf("expected 1 suggested finding, got %d", len(iter.SuggestedFindings()))
	}
}

// TestPipelineRun_DetectorError tests error handling from detector.
func TestPipelineRun_DetectorError(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

// TestPipelineRun_ContextCancellation tests context cancellation.
func TestPipelineRun_ContextCancellation(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected context error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestPipelineRun_Timeout tests pipeline timeout.
func TestPipelineRun_Timeout(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.ParallelDetectors = false
	config.Timeout = 50 * time.Millisecond

	detector := &mockDetector{
		name:     "slow",
		delay:    500 * time.Millisecond,
		findings: []finding.Finding{{ID: "t:r:f:1", Rule: "r", ToolName: "t", Message: "m"}},
	}

	p, err := New(config, t.TempDir(), detector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()

	_, err = p.Run(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestPipelineRun_MaxIterations tests max iteration limit.
func TestPipelineRun_MaxIterations(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	config.MaxIterations = 2
	config.ParallelDetectors = false

	// Findings that won't be auto-fixed (suggest strategy)
	findings := []finding.Finding{
		{
			ID:          "test:1",
			Rule:        "test-rule",
			ToolName:    "test-tool",
			Message:     "test message",
			Severity:    finding.SeverityWarning,
			Position:    finding.Position{File: "test.go", Line: 1},
			FixStrategy: finding.FixStrategySuggest,
		},
	}

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

	{
		total := result.TotalIterations
		if total != config.MaxIterations {
			t.Errorf("expected %d iterations, got %d", config.MaxIterations, total)
		}
	}
}

// TestPipelineRun_Parallel tests parallel detection.
func TestPipelineRun_Parallel(t *testing.T) {
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

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	assertFindingsFound(t, result, 2, "expected 2 findings")
}

// TestDetectParallel_SuppressionConsistency verifies that parallel and
// sequential detection produce identical results when findings are suppressed.
func TestDetectParallel_SuppressionConsistency(t *testing.T) {
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

		var all []finding.Finding //nolint:prealloc
		for _, iter := range findings.Iterations {
			all = append(all, iter.Findings()...)
		}

		return all, notified
	}

	seqFindings, seqNotified := runDetect(false)
	parFindings, parNotified := runDetect(true)

	if len(seqFindings) != len(parFindings) {
		t.Errorf(
			"sequential found %d findings, parallel found %d",
			len(seqFindings),
			len(parFindings),
		)
	}

	if len(seqNotified) != len(parNotified) {
		t.Errorf("sequential notified %d, parallel notified %d", len(seqNotified), len(parNotified))
	}

	assertNoSuppressedFindings(t, seqFindings, "sequential")
	assertNoSuppressedFindings(t, parFindings, "parallel")
	assertNoSuppressedNotified(t, seqNotified, "sequential")
	assertNoSuppressedNotified(t, parNotified, "parallel")
}

// TestTriage tests the triage function.
func TestTriage(t *testing.T) {
	t.Parallel()
	findings := []finding.Finding{
		{ID: "1", FixStrategy: finding.FixStrategyDirect},
		{ID: "2", FixStrategy: finding.FixStrategySuggest},
		{ID: "3", FixStrategy: finding.FixStrategyAI},
		{ID: "4", FixStrategy: finding.FixStrategyNone},
		{ID: "5", FixStrategy: ""},
	}

	p := &Pipeline{config: DefaultConfig()}
	result := p.triage(findings)

	if len(result.Direct) != 1 {
		t.Errorf("expected 1 direct, got %d", len(result.Direct))
	}

	if len(result.Suggest) != 2 {
		t.Errorf("expected 2 suggest, got %d", len(result.Suggest))
	}

	if len(result.None) != 2 {
		t.Errorf("expected 2 none, got %d", len(result.None))
	}
}

// TestDetectorFunc tests the DetectorFunc adapter.
func TestDetectorFunc(t *testing.T) {
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

	if f.Name() != "anonymous" {
		t.Errorf("expected name 'anonymous', got %s", f.Name())
	}
}

func TestNamedDetectorFunc(t *testing.T) {
	t.Parallel()

	fn := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{{ID: "test"}}, nil
	})

	d := NamedDetectorFunc("my-linter", fn)

	if d.Name() != "my-linter" {
		t.Errorf("expected name 'my-linter', got %s", d.Name())
	}

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 || findings[0].ID != "test" {
		t.Errorf("expected 1 finding with ID 'test', got %v", findings)
	}
}

// TestDefaultConfig tests default configuration.
func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	if config.MaxIterations != 5 {
		t.Errorf("expected MaxIterations 5, got %d", config.MaxIterations)
	}

	if !config.ParallelDetectors {
		t.Error("expected ParallelDetectors to be true")
	}

	if config.Timeout != 10*time.Minute {
		t.Errorf("expected Timeout 10m, got %v", config.Timeout)
	}
}

// TestFixApplier tests the FixApplier.
func TestFixApplier(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")

	testContent := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := writeFile(testFile, []byte(testContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test backup
	err := applier.backup(testFile)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup exists
	backupPath := applier.backups[testFile]
	if backupPath == "" {
		t.Fatal("backup path not recorded")
	}

	// Test restore
	err = applier.restore(testFile)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	// Test Apply with empty fixes
	applied, err := applier.Apply(context.Background(), nil)
	if err != nil {
		t.Fatalf("apply with nil fixes failed: %v", err)
	}

	if applied != 0 {
		t.Errorf("expected 0 applied, got %d", applied)
	}

	// Test Apply with fixes that have no Position.File
	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{}},
	}

	applied, err = applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if applied != 0 {
		t.Errorf("expected 0 applied (no file), got %d", applied)
	}
}

// TestFixApplier_Apply tests actual fix application.
func TestFixApplier_Apply(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")

	originalContent := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	writeTestFile(t, testFile, []byte(originalContent))

	// Create fix
	fixes := []finding.Finding{
		makeFixFinding("1", `println("hello")`, `println("world")`, "test.go", 0),
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if applied != 1 {
		t.Errorf("expected 1 applied, got %d", applied)
	}

	// Verify content changed
	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tprintln(\"world\")\n}\n"
	if string(content) != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, string(content))
	}
}

func TestFixApplier_RangeBasedFix(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

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
			Position:    finding.Position{File: "test.go", Line: 5, Column: 2},
			Range:       finding.NewRangePtr("test.go", 5, 2, 5, 18),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if applied != 1 {
		t.Fatalf("expected 1 applied, got %d", applied)
	}

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// First println unchanged, second replaced.
	want := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tfmt.Println(\"world\")\n}\n"
	if string(got) != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestFixApplier_MultiLineRangeFix(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

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

	if applied != 1 {
		t.Fatalf("expected 1 applied, got %d", applied)
	}

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}\n"
	if string(got) != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestFixApplier_RangeFixEmptyBeforeCode(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

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

	if applied != 1 {
		t.Fatalf("expected 1 applied, got %d", applied)
	}

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc replaced() { // inserted\n}\n\nfunc main() {}\n"
	if string(got) != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, string(got))
	}
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
	if len(got) != 2 {
		t.Fatalf("Findings() returned %d items, want 2", len(got))
	}

	assertFindingMessage := func(idx int, want string) {
		t.Helper()

		msg := got[idx].Message
		if msg != want {
			t.Errorf("Findings()[%d].Message = %q, want %q", idx, msg, want)
		}
	}
	assertFindingMessage(0, "a")
	assertFindingMessage(1, "b")

	suggested := iter.SuggestedFindings()
	if len(suggested) != 1 {
		t.Fatalf("SuggestedFindings() returned %d items, want 1", len(suggested))
	}

	if suggested[0].Message != "s" {
		t.Errorf("SuggestedFindings()[0].Message = %q, want %q", suggested[0].Message, "s")
	}
}

func TestIteration_Findings_Empty(t *testing.T) {
	t.Parallel()

	iter := Iteration{Number: 1}
	if got := iter.Findings(); got != nil {
		t.Errorf("Findings() = %v, want nil", got)
	}

	if got := iter.SuggestedFindings(); got != nil {
		t.Errorf("SuggestedFindings() = %v, want nil", got)
	}
}

func TestIoErrorAt(t *testing.T) {
	t.Parallel()

	err := ioErrorAt("read failed", os.ErrNotExist, "foo.go")

	var fe *finding.FindingError
	if !errors.As(err, &fe) {
		t.Fatal("expected FindingError")
	}

	if fe.Category != finding.ErrCategoryIO {
		t.Errorf("category = %q, want %q", fe.Category, finding.ErrCategoryIO)
	}

	if fe.Position.File != "foo.go" {
		t.Errorf("file = %q, want %q", fe.Position.File, "foo.go")
	}
}

func TestApplyDirectFixes(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "fixme.go")

	original := "package main\n\nfunc main() {\n\told()\n}\n" //nolint:goconst // test fixture
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

	if applied != 1 {
		t.Fatalf("applied = %d, want 1", applied)
	}

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	if string(got) != want {
		t.Errorf("file content:\nwant:\n%s\ngot:\n%s", want, string(got))
	}

	if m.TotalFixesApplied() != 1 {
		t.Errorf("FixesApplied = %d, want 1", m.TotalFixesApplied())
	}
}

func TestApplyDirectFixes_NoMetrics(t *testing.T) {
	t.Parallel()

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

	if applied != 1 {
		t.Fatalf("applied = %d, want 1", applied)
	}
}

func TestDryRun(t *testing.T) {
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

	det := &mockDetector{name: "tool", findings: []finding.Finding{f}}
	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if applied != 0 {
		t.Errorf("applied = %d, want 0 (dry run should not apply fixes)", applied)
	}

	assertIterationsLen(t, result, 1)

	iter := result.Iterations[0]
	if iter.DirectFixes != 1 {
		t.Errorf("DirectFixes = %d, want 1 (triage should still run)", iter.DirectFixes)
	}

	data, _ := os.ReadFile(testFile)
	if string(data) != "package main\n" {
		t.Errorf("file was modified during dry run: %q", data)
	}
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
			_, err := New(tt.config, t.TempDir())
			if err == nil {
				t.Fatal("expected error for invalid config")
			}
		})
	}
}

func TestNew_ValidConfig_NoError(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()
	p, err := New(config, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
}

func TestPipelineRun_PartialErrorsSurfaced(t *testing.T) {
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

	if result.PartialErrors == nil {
		t.Fatal("expected PartialErrors to be populated")
	}

	if _, ok := result.PartialErrors["bad"]; !ok {
		t.Error("expected 'bad' detector in PartialErrors")
	}

	if _, ok := result.PartialErrors["good"]; ok {
		t.Error("did not expect 'good' detector in PartialErrors")
	}
}

func TestPipelineRun_MetricsInResult(t *testing.T) {
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

	if result.Metrics.TotalDuration == 0 {
		t.Error("expected non-zero TotalDuration in metrics snapshot")
	}

	if result.Metrics.FixesApplied != 0 {
		t.Errorf("FixesApplied = %d, want 0", result.Metrics.FixesApplied)
	}
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

	det := &mockDetector{name: "tool", findings: []finding.Finding{fix}}
	p, err := New(Config{MaxIterations: 3, ParallelDetectors: false}, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	iter := result.Iterations[0]
	if iter.DirectFixes != 1 {
		t.Errorf("DirectFixes = %d, want 1", iter.DirectFixes)
	}

	if iter.Applied != 1 {
		t.Errorf("Applied = %d, want 1", iter.Applied)
	}
}

func TestApplyTriage_ConflictingFixes(t *testing.T) {
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

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	iter := result.Iterations[0]
	if iter.Conflicts == 0 {
		t.Error("expected conflicts from overlapping fixes")
	}

	if len(conflictFindings) == 0 {
		t.Error("expected OnFix callback with conflict finding")
	}
}

func TestApplyTriage_OnFixCallback(t *testing.T) {
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
		OnFix: func(f finding.Finding, wasApplied bool) {
			if wasApplied {
				appliedIDs = append(appliedIDs, f.ID)
			}
		},
	}

	det := &mockDetector{name: "tool", findings: []finding.Finding{fix}}
	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(appliedIDs) == 0 {
		t.Error("expected OnFix callback for applied fix")
	}

	if appliedIDs[0] != "fix1" {
		t.Errorf("applied ID = %q, want %q", appliedIDs[0], "fix1")
	}
}

func TestApplyTriage_EmptyFixes(t *testing.T) {
	t.Parallel()

	p := &Pipeline{config: DefaultConfig()}
	iter := &Iteration{Number: 1}

	err := p.applyTriage(context.Background(), nil, iter)
	if err != nil {
		t.Fatalf("applyTriage with nil fixes: %v", err)
	}

	if iter.Applied != 0 {
		t.Errorf("Applied = %d, want 0", iter.Applied)
	}

	if iter.Conflicts != 0 {
		t.Errorf("Conflicts = %d, want 0", iter.Conflicts)
	}
}

func TestApplyTriage_AllConflicting(t *testing.T) {
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

	if iter.Conflicts != 1 {
		t.Errorf("Conflicts = %d, want 1 (one filtered, one kept)", iter.Conflicts)
	}

	if iter.Applied != 1 {
		t.Errorf("Applied = %d, want 1 (one non-conflicting fix applied)", iter.Applied)
	}
}

func TestPipelineRun_DirectFixStabilizes(t *testing.T) {
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

	if !result.Stable {
		t.Error("expected stable result after fix applied")
	}

	if result.TotalIterations != 2 {
		t.Errorf("TotalIterations = %d, want 2 (detect-fix + verify)", result.TotalIterations)
	}

	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	if string(content) != want {
		t.Errorf("file content:\nwant:\n%s\ngot:\n%s", want, string(content))
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
