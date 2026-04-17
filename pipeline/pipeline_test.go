package pipeline

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

// mockDetector is a test detector that returns predefined findings.
type mockDetector struct {
	name     string
	findings []finding.Finding
	err      error
	delay    time.Duration
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.findings, m.err
}

func newMockDetector(name, toolName string, findings ...finding.Finding) *mockDetector {
	if len(findings) == 0 {
		findings = []finding.Finding{
			{ID: "1", Rule: "r1", ToolName: toolName, Message: "m", Severity: finding.SeverityError},
		}
	}
	return &mockDetector{name: name, findings: findings}
}

// testFinding creates a Finding for testing with common fields.
func testFinding(id, rule, tool, msg string, sev finding.Severity, file string) finding.Finding {
	return finding.Finding{
		ID:       id,
		Rule:     rule,
		ToolName: tool,
		Message:  msg,
		Severity: sev,
		Position: finding.Position{File: file},
	}
}

// TestNew tests pipeline creation.
func TestNew(t *testing.T) {
	config := DefaultConfig()
	detector := &mockDetector{name: "test", findings: nil}

	p := New(config, "/tmp", detector)

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
	config := DefaultConfig()
	config.ParallelDetectors = false
	detector := &mockDetector{name: "test", findings: nil}

	p := New(config, t.TempDir(), detector)
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
	p := New(config, t.TempDir(), detector)
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stable {
		t.Error("expected non-stable result (findings present but not auto-fixed)")
	}
	// Should run until max iterations since findings persist and can't be auto-fixed
	if result.TotalIterations != config.MaxIterations {
		t.Errorf("expected %d iterations, got %d", config.MaxIterations, result.TotalIterations)
	}
	if len(result.Iterations) != config.MaxIterations {
		t.Fatalf("expected %d iteration records, got %d", config.MaxIterations, len(result.Iterations))
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
	config := DefaultConfig()
	config.ParallelDetectors = false

	expectedErr := errors.New("detector failed")
	detector := &mockDetector{name: "test", err: expectedErr}

	p := New(config, t.TempDir(), detector)
	ctx := context.Background()

	_, err := p.Run(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

// TestPipelineRun_ContextCancellation tests context cancellation.
func TestPipelineRun_ContextCancellation(t *testing.T) {
	config := DefaultConfig()
	config.ParallelDetectors = false

	// Detector that takes time
	detector := &mockDetector{
		name:  "slow",
		delay: 100 * time.Millisecond,
	}

	p := New(config, t.TempDir(), detector)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := p.Run(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestPipelineRun_Timeout tests pipeline timeout.
func TestPipelineRun_Timeout(t *testing.T) {
	config := DefaultConfig()
	config.ParallelDetectors = false
	config.Timeout = 50 * time.Millisecond

	detector := &mockDetector{
		name:     "slow",
		delay:    500 * time.Millisecond,
		findings: []finding.Finding{{ID: "t:r:f:1", Rule: "r", ToolName: "t", Message: "m"}},
	}

	p := New(config, t.TempDir(), detector)
	ctx := context.Background()

	_, err := p.Run(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestPipelineRun_MaxIterations tests max iteration limit.
func TestPipelineRun_MaxIterations(t *testing.T) {
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
	p := New(config, t.TempDir(), detector)
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalIterations != config.MaxIterations {
		t.Errorf("expected %d iterations, got %d", config.MaxIterations, result.TotalIterations)
	}
}

// TestPipelineRun_Parallel tests parallel detection.
func TestPipelineRun_Parallel(t *testing.T) {
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
			{ID: "2", Rule: "r2", ToolName: "t2", Message: "m2", Severity: finding.SeverityWarning, Position: finding.Position{File: "b.go"}},
		},
	}

	p := New(config, t.TempDir(), d1, d2)
	ctx := context.Background()

	result, err := p.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}
	if result.Iterations[0].FindingsFound != 2 {
		t.Errorf("expected 2 findings, got %d", result.Iterations[0].FindingsFound)
	}
}

// TestTriage tests the triage function.
func TestTriage(t *testing.T) {
	findings := []finding.Finding{
		{ID: "1", FixStrategy: finding.FixStrategyDirect},
		{ID: "2", FixStrategy: finding.FixStrategySuggest},
		{ID: "3", FixStrategy: finding.FixStrategyAI},
		{ID: "4", FixStrategy: finding.FixStrategyNone},
		{ID: "5", FixStrategy: ""}, // Empty strategy
	}

	p := &Pipeline{config: DefaultConfig()}
	result := p.triage(findings)

	if len(result.Direct) != 1 {
		t.Errorf("expected 1 direct, got %d", len(result.Direct))
	}
	if len(result.Suggest) != 1 {
		t.Errorf("expected 1 suggest, got %d", len(result.Suggest))
	}
	if len(result.None) != 3 {
		t.Errorf("expected 3 none (including AI), got %d", len(result.None))
	}
}

// TestDetectorFunc tests the DetectorFunc adapter.
func TestDetectorFunc(t *testing.T) {
	called := false
	f := DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
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

	fn := DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
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
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := writeFile(testFile, []byte(testContent), 0644); err != nil {
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
	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")
	originalContent := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := writeFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create fix
	fixes := []finding.Finding{
		{
			ID:          "1",
			BeforeCode:  "println(\"hello\")",
			AfterCode:   "println(\"world\")",
			Position:    finding.Position{File: "test.go"},
			FixStrategy: finding.FixStrategyDirect,
		},
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
			for i := 0; i < b.N; i++ {
				p := New(config, b.TempDir(), detectors...)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				p.Run(ctx)
				cancel()
			}
		})
	}
}

// Helper functions for file operations (avoiding banned imports).

func writeFile(path string, data []byte, perm uint32) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(perm))
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var result []byte
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return result, nil
}
