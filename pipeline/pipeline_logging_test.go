package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func mockDetectorWithFinding(name, id, rule, tool, msg string) *mockDetector {
	return &mockDetector{
		name: name,
		findings: []finding.Finding{
			{
				ID:       finding.ID(id),
				Rule:     finding.RuleName(rule),
				ToolName: finding.ToolName(tool),
				Message:  msg,
				Severity: finding.SeverityError,
			},
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
	g.Expect(result.Correlations[0].FindingIDs).To(Equal([]finding.ID{"a1", "b1"}))
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

func TestReasonFromContext(t *testing.T) {
	t.Parallel()

	t.Run("cancelled", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		got := reasonFromContext(ctx)
		g.Expect(got).To(Equal(ReasonCancelled))
	})

	t.Run("timeout", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		<-ctx.Done()

		got := reasonFromContext(ctx)
		g.Expect(got).To(Equal(ReasonTimeout))
	})
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
