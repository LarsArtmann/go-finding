package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
)

func TestFixApplier_BackupRestoreRoundTrip(t *testing.T) {
	t.Parallel()

	applier := NewFixApplier(t.TempDir())
	testBackupRestore(t, applier,
		"package main\n\nfunc main() {\n\tprintln(\"original\")\n}\n",
		"package main\n\nfunc main() {\n\tprintln(\"modified\")\n}\n")
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
		assert.Fail(t, "backup paths should differ for same-named files in different directories")
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

	assert.Equal(t, 5, applied)
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
	p, err := New(config, tempDir, goodDetector, badDetector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	assert.NotNil(t, result)

	assertIterationsLen(t, result, 1)

	assertFindingsFound(t, result, 1, "FindingsFound")
}

func TestPipeline_RetryConfig(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	flaky := NamedDetectorFunc(
		"flaky",
		DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
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
	p, err := New(config, tempDir, flaky)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	assert.Equal(t, 1, result.Iterations[0].FindingsFound)

	assert.GreaterOrEqual(t, calls.Load(), int32(3))
}

func TestPipeline_VerifyAfterFix(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	det := NamedDetectorFunc(
		"verifiable",
		DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
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
	p, err := New(config, tempDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	assert.NotNil(t, result.Verification)
}

func TestPipeline_MetricsRecordsDetector(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics()
	config := Config{
		MaxIterations:     1,
		ParallelDetectors: false,
		Metrics:           metrics,
	}

	det := mockDetectorWithFinding("my-detector", "1", "r1", "test", "m")

	tempDir := t.TempDir()
	p, err := New(config, tempDir, det)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	snap := metrics.Snapshot()
	assert.NotZero(t, snap.DetectorTimes["my-detector"])

	assert.Equal(t, 1, snap.FindingsFound["my-detector"])
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.s.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}
