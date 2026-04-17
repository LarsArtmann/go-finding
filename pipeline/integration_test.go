package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

func TestFixApplier_BackupRestoreRoundTrip(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	testFile := filepath.Join(tempDir, "original.go")

	original := "package main\n\nfunc main() {\n\tprintln(\"original\")\n}\n"
	if err := writeFile(testFile, []byte(original), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := applier.backup(testFile); err != nil {
		t.Fatalf("backup: %v", err)
	}

	modified := "package main\n\nfunc main() {\n\tprintln(\"modified\")\n}\n"
	if err := writeFile(testFile, []byte(modified), 0o644); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	if err := applier.restore(testFile); err != nil {
		t.Fatalf("restore: %v", err)
	}

	data, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != original {
		t.Errorf("restored content mismatch:\ngot:  %q\nwant: %q", string(data), original)
	}
}

func TestFixApplier_BackupPathCollision(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	sub1 := filepath.Join(tempDir, "dirA", "file.go")
	sub2 := filepath.Join(tempDir, "dirB", "file.go")

	err := os.MkdirAll(filepath.Dir(sub1), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(filepath.Dir(sub2), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = writeFile(sub1, []byte("A"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = writeFile(sub2, []byte("B"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = applier.backup(sub1)
	if err != nil {
		t.Fatalf("backup sub1: %v", err)
	}

	err = applier.backup(sub2)
	if err != nil {
		t.Fatalf("backup sub2: %v", err)
	}

	if applier.backups[sub1] == applier.backups[sub2] {
		t.Error("backup paths should differ for same-named files in different directories")
	}
}

func TestFixApplier_MultipleFilesConcurrent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	applier := NewFixApplier(tempDir)

	for i := range 5 {
		name := filepath.Join(tempDir, "file"+string(rune('A'+i))+".go")

		content := "package main\n\nfunc main() {\n\tprintln(\"old" + string(
			rune('A'+i),
		) + "\")\n}\n"

		err := writeFile(name, []byte(content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	findings := make([]finding.Finding, 5)
	for i := range findings {
		findings[i] = finding.Finding{
			ID:          string(rune('A' + i)),
			BeforeCode:  "old" + string(rune('A'+i)),
			AfterCode:   "new" + string(rune('A'+i)),
			Position:    finding.Position{File: "file" + string(rune('A'+i)) + ".go"},
			FixStrategy: finding.FixStrategyDirect,
		}
	}

	applied, err := applier.Apply(context.Background(), findings)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if applied != 5 {
		t.Errorf("applied = %d, want 5", applied)
	}
}

func TestPipeline_GracefulDegradation(t *testing.T) {
	t.Parallel()

	config := Config{
		MaxIterations:       1,
		GracefulDegradation: true,
	}

	goodDetector := newMockDetector("good", "good")
	badDetector := &mockDetector{
		name: "bad",
		err:  os.ErrPermission,
	}

	tempDir := t.TempDir()
	p := New(config, tempDir, goodDetector, badDetector)

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Iterations) != 1 {
		t.Fatalf("iterations = %d, want 1", len(result.Iterations))
	}

	if result.Iterations[0].FindingsFound != 1 {
		t.Errorf(
			"FindingsFound = %d, want 1 (from good detector)",
			result.Iterations[0].FindingsFound,
		)
	}
}

func TestPipeline_RetryConfig(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	flaky := NamedDetectorFunc(
		"flaky",
		DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
			count := calls.Add(1)
			if count < 3 {
				return nil, os.ErrDeadlineExceeded
			}

			return []finding.Finding{
				testFinding("1", "r1", "flaky", "m", finding.SeverityError, ""),
			}, nil
		}),
	)

	retryConfig := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
	}
	config := Config{
		MaxIterations: 1,
		Retry:         &retryConfig,
	}

	tempDir := t.TempDir()
	p := New(config, tempDir, flaky)

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result.Iterations[0].FindingsFound != 1 {
		t.Errorf("FindingsFound = %d, want 1 (after retries)", result.Iterations[0].FindingsFound)
	}

	if calls.Load() < 3 {
		t.Errorf("calls = %d, want >= 3", calls.Load())
	}
}

func TestPipeline_VerifyAfterFix(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	det := NamedDetectorFunc(
		"verifiable",
		DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
			c := callCount.Add(1)
			if c == 1 {
				return []finding.Finding{
					testFinding("1", "r1", "v", "m", finding.SeverityError, ""),
				}, nil
			}

			return nil, nil
		}),
	)

	config := Config{
		MaxIterations:     2,
		VerifyAfterFix:    true,
		ParallelDetectors: false,
	}

	tempDir := t.TempDir()
	p := New(config, tempDir, det)

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result.Verification == nil {
		t.Error("expected Verification to be non-nil")
	}
}

func TestPipeline_MetricsRecordsDetector(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics()
	config := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		Metrics:           metrics,
	}

	det := &mockDetector{
		name: "my-detector",
		findings: []finding.Finding{
			{ID: "1", Rule: "r1", ToolName: "test", Message: "m", Severity: finding.SeverityError},
		},
	}

	tempDir := t.TempDir()
	p := New(config, tempDir, det)

	_, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	snap := metrics.Snapshot()
	if snap.DetectorTimes["my-detector"] == 0 {
		t.Error("expected DetectorTimes to be recorded for my-detector")
	}

	if snap.FindingsFound["my-detector"] != 1 {
		t.Errorf("FindingsFound = %d, want 1", snap.FindingsFound["my-detector"])
	}
}

func TestSuppression_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *finding.Suppression
		want bool
	}{
		{"nil", nil, false},
		{"empty", &finding.Suppression{}, false},
		{"kind only", &finding.Suppression{Kind: finding.SuppressionInSource}, false},
		{"rule only", &finding.Suppression{Rule: "SA1000"}, false},
		{"valid", &finding.Suppression{Kind: finding.SuppressionInSource, Rule: "SA1000"}, true},
	}
	for _, tt := range tests {
		if got := tt.s.IsValid(); got != tt.want {
			t.Errorf("%s: IsValid() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
