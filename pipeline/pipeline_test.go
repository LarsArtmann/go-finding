package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	}
}

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

	assert.Len(t, p.detectors, 1)

	assert.Equal(t, "/tmp", p.rootDir)

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

	assert.True(t, result.Stable)
	assert.Equal(t, 1, result.TotalIterations)
}

// TestPipelineRun_WithFindings tests pipeline with findings.
func TestPipelineRun_WithFindings(t *testing.T) {
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

	assert.False(t, result.Stable)
	assert.Equal(t, config.MaxIterations, result.TotalIterations)

	assert.Len(t, result.Iterations, config.MaxIterations)

	iter := result.Iterations[0]
	assert.Equal(t, 1, iter.FindingsFound)

	assert.Equal(t, 1, iter.SuggestFixes)

	assert.Len(t, iter.SuggestedFindings(), 1)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}

// TestDetectSequential_ContextCancellation verifies that detectSequential
// returns an error when the context is cancelled.
func TestDetectSequential_ContextCancellation(t *testing.T) {
	t.Parallel()

	p := &Pipeline{
		detectors: []Detector{&mockDetector{name: "slow", delay: time.Second}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.detectSequential(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestPipelineRun_MaxIterations tests max iteration limit.
func TestPipelineRun_MaxIterations(t *testing.T) {
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

	assert.Equal(t, config.MaxIterations, result.TotalIterations)
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

	assert.NotEmpty(t, result.Iterations)

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

		var all []finding.Finding
		for _, iter := range findings.Iterations {
			all = append(all, iter.Findings()...)
		}

		return all, notified
	}

	seqFindings, seqNotified := runDetect(false)
	parFindings, parNotified := runDetect(true)

	assert.Len(t, seqFindings, len(parFindings))

	assert.Len(t, seqNotified, len(parNotified))

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

	assert.Len(t, result.Direct, 1)

	assert.Len(t, result.Suggest, 2)

	assert.Len(t, result.None, 2)
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

	assert.Equal(t, "anonymous", f.Name())
}

func TestNamedDetectorFunc(t *testing.T) {
	t.Parallel()

	fn := makeFindingDetectorFunc("test")

	d := NamedDetectorFunc("my-linter", fn)

	assert.Equal(t, "my-linter", d.Name())

	findings, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Len(t, findings, 1)
	assert.Equal(t, "test", findings[0].ID)
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	assert.Equal(t, 5, config.MaxIterations)

	assert.True(t, config.ParallelDetectors)

	assert.Equal(t, 10*time.Minute, config.Timeout)
}

// TestFixApplier tests the FixApplier.
func TestFixApplier(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")

	testContent := helloProgram
	if err := writeFile(testFile, []byte(testContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test backup
	err := applier.backup.Backup(testFile)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Verify backup exists
	backupPath := applier.backup.BackupPath(testFile)
	assert.NotEmpty(t, backupPath)

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

	assert.Equal(t, 0, applied)

	// Test Apply with fixes that have no Position.File
	fixes := []finding.Finding{
		{ID: "1", BeforeCode: "old", AfterCode: "new", Position: finding.Position{}},
	}

	applied, err = applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	assert.Equal(t, 0, applied)
}

// TestFixApplier_Apply tests actual fix application.
func TestFixApplier_Apply(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

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

	assert.Equal(t, 1, applied)

	// Verify content changed
	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "package main\n\nfunc main() {\n\tprintln(\"world\")\n}\n"
	assert.Equal(t, expected, string(content))
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
			Position:    finding.Pos("test.go", 5, 2),
			Range:       finding.NewRangePtr("test.go", 5, 2, 5, 18),
			FixStrategy: finding.FixStrategyDirect,
		},
	}

	applied, err := applier.Apply(context.Background(), fixes)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	// First println unchanged, second replaced.
	want := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tfmt.Println(\"world\")\n}\n"
	assert.Equal(t, want, string(got))
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

	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc new() {\n\treturn 42\n}\n\nfunc main() {}\n"
	assert.Equal(t, want, string(got))
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

	assert.Equal(t, 1, applied)

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc replaced() { // inserted\n}\n\nfunc main() {}\n"
	assert.Equal(t, want, string(got))
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
	assert.Len(t, got, 2)

	assertFindingMessage := func(idx int, want string) {
		t.Helper()

		msg := got[idx].Message
		assert.Equal(t, want, msg)
	}
	assertFindingMessage(0, "a")
	assertFindingMessage(1, "b")

	suggested := iter.SuggestedFindings()
	assert.Len(t, suggested, 1)

	assert.Equal(t, "s", suggested[0].Message)
}

func TestIteration_Findings_Empty(t *testing.T) {
	t.Parallel()

	iter := Iteration{Number: 1}
	assert.Nil(t, iter.Findings())

	assert.Nil(t, iter.SuggestedFindings())
}

func TestIoErrorAt(t *testing.T) {
	t.Parallel()

	err := ioErrorAt("read failed", os.ErrNotExist, "foo.go")

	var fe *finding.FindingError
	require.ErrorAs(t, err, &fe)

	assertFindingErrorIO(t, fe, "foo.go")
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

	assert.Len(t, applied, 1)

	got, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	assert.Equal(t, want, string(got))

	assert.Equal(t, 1, m.TotalFixesApplied())
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

	assert.Len(t, applied, 1)
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

	assert.Equal(t, 0, applied)

	assertIterationsLen(t, result, 1)

	iter := result.Iterations[0]
	assert.Equal(t, 1, iter.DirectFixes)

	data, _ := os.ReadFile(testFile)
	assert.Equal(t, "package main\n", string(data))
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
			assert.Error(t, err)
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

	assert.NotNil(t, p)
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

	assert.NotNil(t, result.PartialErrors)

	_, ok := result.PartialErrors["bad"]
	assert.True(t, ok)

	_, ok = result.PartialErrors["good"]
	assert.False(t, ok)
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

	assert.NotZero(t, result.Metrics.TotalDuration)

	assert.Equal(t, 0, result.Metrics.FixesApplied)
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

	assert.NotEmpty(t, result.Iterations)

	iter := result.Iterations[0]
	assert.Equal(t, 1, iter.DirectFixes)

	assert.Equal(t, 1, iter.Applied)
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

	assert.NotEmpty(t, result.Iterations)

	iter := result.Iterations[0]
	assert.NotZero(t, iter.Conflicts)

	assert.NotEmpty(t, conflictFindings)
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

	assert.NotEmpty(t, appliedIDs)

	assert.Equal(t, "fix1", appliedIDs[0])
}

func TestApplyTriage_EmptyFixes(t *testing.T) {
	t.Parallel()

	p := &Pipeline{config: DefaultConfig()}
	iter := &Iteration{Number: 1}

	err := p.applyTriage(context.Background(), nil, iter)
	if err != nil {
		t.Fatalf("applyTriage with nil fixes: %v", err)
	}

	assert.Equal(t, 0, iter.Applied)

	assert.Equal(t, 0, iter.Conflicts)
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

	assert.Equal(t, 1, iter.Conflicts)

	assert.Equal(t, 1, iter.Applied)
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

	assert.True(t, result.Stable)

	assert.Equal(t, 2, result.TotalIterations)

	content, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	want := "package main\n\nfunc main() {\n\tnew()\n}\n"
	assert.Equal(t, want, string(content))
}

func mockDetectorWithFinding(name, id, rule, tool, msg string) *mockDetector {
	return &mockDetector{
		name: name,
		findings: []finding.Finding{
			{ID: id, Rule: rule, ToolName: tool, Message: msg, Severity: finding.SeverityError},
		},
	}
}

func TestPipelineRun_CorrelateFindings(t *testing.T) {
	t.Parallel()

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

	cfg := Config{MaxIterations: 1, ParallelDetectors: false, CorrelateFindings: true}
	p, err := New(cfg, ".", detA, detB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	assert.NotEmpty(
		t,
		result.Correlations,
		"expected correlations for nearby findings from different tools",
	)
	assert.Equal(t, []string{"a1", "b1"}, result.Correlations[0].FindingIDs)
}

func TestPipelineRun_NoCorrelateWhenDisabled(t *testing.T) {
	t.Parallel()

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

	cfg := Config{MaxIterations: 1, ParallelDetectors: false, CorrelateFindings: false}
	p, err := New(cfg, ".", detA, detB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	assert.Empty(t, result.Correlations, "expected no correlations when disabled")
}
