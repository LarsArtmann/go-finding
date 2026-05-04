package pipeline

import (
	"context"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func makeFinding(id, msg string) finding.Finding {
	return finding.Finding{ID: id, Message: msg}
}

func assertDiffResult(t *testing.T, result *VerifyResult, resolved, remaining, newFindings int) {
	g := NewWithT(t)
	t.Helper()

	g.Expect(result.Resolved).To(Equal(resolved))
	g.Expect(result.Remaining).To(HaveLen(remaining))
	g.Expect(result.NewFindings).To(HaveLen(newFindings))
}

func assertNewFindingID(t *testing.T, result *VerifyResult, wantID string) {
	g := NewWithT(t)
	t.Helper()
	g.Expect(result.NewFindings).To(HaveLen(1))
	g.Expect(result.NewFindings[0].ID).To(Equal(wantID))
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
	assertNewFindingID(t, result, "c:rule:file.go:3")
}

func TestDiffFindings_Mixed(t *testing.T) {
	g := NewWithT(t)
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

	g.Expect(result.Fixed).To(HaveLen(1))
	g.Expect(result.Fixed[0].ID).To(Equal("a:rule:file.go:1"))
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
	g := NewWithT(t)
	t.Parallel()

	callCount := 0
	detector := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
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

	g.Expect(result.Resolved).To(Equal(1))
}

func TestVerifier_Verify_DetectorError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	detector := makeErrorDetector("detector failed")

	v := NewVerifier([]Detector{detector})
	_, err := v.Verify(context.Background(), nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("verify"))
	g.Expect(err.Error()).To(ContainSubstring("detector"))
}

func TestVerifier_Verify_SuppressedFindingsFiltered(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	suppressed := finding.Finding{
		ID:          "suppressed:rule:f.go:1",
		Message:     "suppressed issue",
		Suppression: &finding.Suppression{Kind: "manual", Reason: "won't fix"},
	}
	normal := makeFinding("normal:rule:f.go:2", "normal issue")

	detector := DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{suppressed, normal}, nil
	})

	v := NewVerifier([]Detector{detector})
	result, err := v.Verify(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g.Expect(result.NewFindings).To(HaveLen(1))

	assertNewFindingID(t, result, "normal:rule:f.go:2")
}

func TestFindingKey_EmptyID(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	f := finding.Finding{
		Position: finding.Position{File: "main.go"},
		Rule:     "SA1000",
		Message:  "unused variable",
	}
	key := f.Key()
	g.Expect(key).To(Equal("\x00main.go\x00SA1000\x00unused variable"))
}

func TestFindingKey_WithID(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	f := finding.Finding{ID: "unique-id-123", Position: finding.Position{File: "main.go"}}
	key := f.Key()
	g.Expect(key).To(Equal("unique-id-123"))
}
