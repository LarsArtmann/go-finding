package pipeline

import (
	"context"
	"testing"

	"github.com/larsartmann/go-finding"
)

func makeFinding(id, msg string) finding.Finding {
	return finding.Finding{ID: id, Message: msg}
}

func assertDiffResult(t *testing.T, result *VerifyResult, resolved, remaining, newFindings int) {
	t.Helper()

	if result.Resolved != resolved {
		t.Errorf("expected %d resolved, got %d", resolved, result.Resolved)
	}

	if len(result.Remaining) != remaining {
		t.Errorf("expected %d remaining, got %d", remaining, len(result.Remaining))
	}

	if len(result.NewFindings) != newFindings {
		t.Errorf("expected %d new findings, got %d", newFindings, len(result.NewFindings))
	}
}

func TestDiffFindings_AllFixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		makeFinding("a:rule:file.go:1", "issue1"),
		makeFinding("b:rule:file.go:2", "issue2"),
	}
	result := DiffFindings(original, nil)
	assertDiffResult(t, result, 2, 0, 0)
}

func TestDiffFindings_NoneFixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{makeFinding("a:rule:file.go:1", "issue1")}
	result := DiffFindings(original, original)
	assertDiffResult(t, result, 0, 1, 0)
}

func TestDiffFindings_NewFindings(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{makeFinding("a:rule:file.go:1", "issue1")}
	post := []finding.Finding{
		makeFinding("a:rule:file.go:1", "issue1"),
		makeFinding("c:rule:file.go:3", "new issue"),
	}
	result := DiffFindings(original, post)
	assertDiffResult(t, result, 0, 1, 1)

	if result.NewFindings[0].ID != "c:rule:file.go:3" {
		t.Errorf("expected new finding ID 'c:rule:file.go:3', got %s", result.NewFindings[0].ID)
	}
}

func TestDiffFindings_Mixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		makeFinding("a:rule:file.go:1", "fixed"),
		makeFinding("b:rule:file.go:2", "remaining"),
	}
	post := []finding.Finding{
		makeFinding("b:rule:file.go:2", "remaining"),
		makeFinding("c:rule:file.go:3", "new"),
	}
	result := DiffFindings(original, post)
	assertDiffResult(t, result, 1, 1, 1)

	if len(result.Fixed) != 1 || result.Fixed[0].ID != "a:rule:file.go:1" {
		t.Errorf("expected fixed finding 'a:rule:file.go:1', got %v", result.Fixed)
	}
}

func TestDiffFindings_EmptyOriginal(t *testing.T) {
	t.Parallel()

	post := []finding.Finding{makeFinding("a:rule:file.go:1", "issue")}
	result := DiffFindings(nil, post)
	assertDiffResult(t, result, 0, 0, 1)
}

func TestDiffFindings_BothEmpty(t *testing.T) {
	t.Parallel()

	result := DiffFindings(nil, nil)
	assertDiffResult(t, result, 0, 0, 0)
}

func TestVerifier_Verify(t *testing.T) {
	t.Parallel()

	callCount := 0
	detector := DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
		callCount++
		if callCount <= 1 {
			return []finding.Finding{makeFinding("a:rule:f.go:1", "issue")}, nil
		}

		return nil, nil
	})
	original, _ := detector.Detect(context.Background())
	v := NewVerifier([]Detector{detector})

	result, err := v.Verify(context.Background(), original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Resolved != 1 {
		t.Errorf("expected 1 resolved, got %d", result.Resolved)
	}
}
