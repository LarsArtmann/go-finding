package pipeline

import (
	"context"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestDiffFindings_AllFixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "issue1"},
		{ID: "b:rule:file.go:2", Message: "issue2"},
	}

	result := DiffFindings(original, nil)

	if result.Resolved != 2 {
		t.Errorf("expected 2 resolved, got %d", result.Resolved)
	}
	if len(result.Remaining) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(result.Remaining))
	}
	if len(result.NewFindings) != 0 {
		t.Errorf("expected 0 new findings, got %d", len(result.NewFindings))
	}
}

func TestDiffFindings_NoneFixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "issue1"},
	}

	result := DiffFindings(original, original)

	if result.Resolved != 0 {
		t.Errorf("expected 0 resolved, got %d", result.Resolved)
	}
	if len(result.Remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(result.Remaining))
	}
}

func TestDiffFindings_NewFindings(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "issue1"},
	}
	post := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "issue1"},
		{ID: "c:rule:file.go:3", Message: "new issue"},
	}

	result := DiffFindings(original, post)

	if result.Resolved != 0 {
		t.Errorf("expected 0 resolved, got %d", result.Resolved)
	}
	if len(result.NewFindings) != 1 {
		t.Errorf("expected 1 new finding, got %d", len(result.NewFindings))
	}
	if result.NewFindings[0].ID != "c:rule:file.go:3" {
		t.Errorf("expected new finding ID 'c:rule:file.go:3', got %s", result.NewFindings[0].ID)
	}
}

func TestDiffFindings_Mixed(t *testing.T) {
	t.Parallel()

	original := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "fixed"},
		{ID: "b:rule:file.go:2", Message: "remaining"},
	}
	post := []finding.Finding{
		{ID: "b:rule:file.go:2", Message: "remaining"},
		{ID: "c:rule:file.go:3", Message: "new"},
	}

	result := DiffFindings(original, post)

	if result.Resolved != 1 {
		t.Errorf("expected 1 resolved, got %d", result.Resolved)
	}
	if len(result.Remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(result.Remaining))
	}
	if len(result.NewFindings) != 1 {
		t.Errorf("expected 1 new finding, got %d", len(result.NewFindings))
	}
	if len(result.Fixed) != 1 || result.Fixed[0].ID != "a:rule:file.go:1" {
		t.Errorf("expected fixed finding 'a:rule:file.go:1', got %v", result.Fixed)
	}
}

func TestDiffFindings_EmptyOriginal(t *testing.T) {
	t.Parallel()

	post := []finding.Finding{
		{ID: "a:rule:file.go:1", Message: "issue"},
	}

	result := DiffFindings(nil, post)

	if result.Resolved != 0 {
		t.Errorf("expected 0 resolved, got %d", result.Resolved)
	}
	if len(result.NewFindings) != 1 {
		t.Errorf("expected 1 new finding, got %d", len(result.NewFindings))
	}
}

func TestDiffFindings_BothEmpty(t *testing.T) {
	t.Parallel()

	result := DiffFindings(nil, nil)

	if result.Resolved != 0 {
		t.Errorf("expected 0 resolved, got %d", result.Resolved)
	}
	if len(result.Remaining) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(result.Remaining))
	}
	if len(result.NewFindings) != 0 {
		t.Errorf("expected 0 new findings, got %d", len(result.NewFindings))
	}
}

func TestVerifier_Verify(t *testing.T) {
	t.Parallel()

	callCount := 0
	detector := DetectorFunc(func(ctx context.Context) ([]finding.Finding, error) {
		callCount++
		if callCount <= 1 {
			return []finding.Finding{
				{ID: "a:rule:f.go:1", Message: "issue"},
			}, nil
		}
		return nil, nil
	})

	// First call to get original findings
	original, _ := detector.Detect(context.Background())

	// Second call (inside Verify) returns empty — all fixed
	v := NewVerifier([]Detector{detector})
	result, err := v.Verify(context.Background(), original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Resolved != 1 {
		t.Errorf("expected 1 resolved, got %d", result.Resolved)
	}
}
