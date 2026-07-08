package finding

import (
	"context"
	"strings"
	"testing"
)

func TestSARIFSuppressionRoundTrip(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "test-tool", Version: "1.0"})
	report.AddFinding(Finding{
		ID:          GenerateID("test-tool", "rule1", Position{File: "main.go", Line: 10, Column: 5}),
		Rule:        "rule1",
		ToolName:    "test-tool",
		Message:     "unused variable",
		Severity:    SeverityWarning,
		Position:    Position{File: FilePath("main.go"), Line: 10, Column: 5},
		FixStrategy: FixStrategyNone,
		Suppression: &Suppression{
			Kind:   SuppressionInSource,
			Rule:   "rule1",
			Reason: "intentionally unused",
		},
	})

	// Default ToSARIF drops suppressed findings.
	data, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF: %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), data)
	if err != nil {
		t.Fatalf("FindingsFromSARIF: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings (suppressed), got %d", len(findings))
	}

	// WithIncludeSuppressed emits them with suppression data.
	dataWithSupp, err := report.ToSARIFWithOpts(WithIncludeSuppressed())
	if err != nil {
		t.Fatalf("ToSARIFWithOpts: %v", err)
	}

	if !strings.Contains(string(dataWithSupp), `"suppressions"`) {
		t.Error("expected suppressions array in SARIF output")
	}

	findingsWithSupp, err := FindingsFromSARIF(context.Background(), dataWithSupp)
	if err != nil {
		t.Fatalf("FindingsFromSARIF with suppressed: %v", err)
	}

	if len(findingsWithSupp) != 1 {
		t.Fatalf("expected 1 finding with suppression, got %d", len(findingsWithSupp))
	}

	f := findingsWithSupp[0]
	if f.Suppression == nil {
		t.Fatal("expected Suppression to be non-nil")
	}

	if f.Suppression.Kind != SuppressionInSource {
		t.Errorf("expected kind %q, got %q", SuppressionInSource, f.Suppression.Kind)
	}

	if f.Suppression.Reason != "intentionally unused" {
		t.Errorf("expected reason %q, got %q", "intentionally unused", f.Suppression.Reason)
	}
}

func TestSARIFSuppressionKindMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		kind       SuppressionKind
		wantSarif  string
		wantStatus string
	}{
		{"in-source", SuppressionInSource, "inSource", "accepted"},
		{"in-config", SuppressionInConfig, "inExternalConfiguration", "accepted"},
		{"in-review", SuppressionInReview, "inExternalConfiguration", "underReview"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotKind := suppressionKindToSARIF(tt.kind)
			gotStatus := suppressionStatusToSARIF(tt.kind)

			if gotKind != tt.wantSarif {
				t.Errorf("kind: got %q, want %q", gotKind, tt.wantSarif)
			}

			if gotStatus != tt.wantStatus {
				t.Errorf("status: got %q, want %q", gotStatus, tt.wantStatus)
			}

			// Round-trip back
			restored := sarifKindToSuppression(gotKind, gotStatus)
			if restored != tt.kind {
				t.Errorf("round-trip: got %q, want %q", restored, tt.kind)
			}
		})
	}
}

// TestSARIFSuppressionRoundTrip_DifferentRule verifies that Suppression.Rule is
// preserved when it differs from Finding.Rule. Regression for the SARIF
// Suppression.Rule loss bug (sarif_export.go:357 / sarif_import.go:247).
// The existing TestSARIFSuppressionRoundTrip has Rule == Finding.Rule, so the
// import fallback (Rule := Finding.Rule) masks the bug.
func TestSARIFSuppressionRoundTrip_DifferentRule(t *testing.T) {
	t.Parallel()

	const (
		findingRule       = RuleName("finding-rule")
		suppressionRule   = RuleName("suppression-rule")
		suppressionReason = "waived by different rule"
	)

	report := NewReport(ToolInfo{Name: "test-tool", Version: "1.0"})
	report.AddFinding(Finding{
		ID:          GenerateID("test-tool", findingRule, Position{File: "main.go", Line: 10}),
		Rule:        findingRule,
		ToolName:    "test-tool",
		Message:     "test",
		Severity:    SeverityWarning,
		Position:    Position{File: FilePath("main.go"), Line: 10},
		FixStrategy: FixStrategyNone,
		Suppression: &Suppression{
			Kind:   SuppressionInConfig,
			Rule:   suppressionRule,
			Reason: suppressionReason,
		},
	})

	data, err := report.ToSARIFWithOpts(WithIncludeSuppressed())
	if err != nil {
		t.Fatalf("ToSARIFWithOpts: %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), data)
	if err != nil {
		t.Fatalf("FindingsFromSARIF: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.Suppression == nil {
		t.Fatal("expected Suppression to be non-nil")
	}

	if f.Suppression.Rule != suppressionRule {
		t.Errorf("Suppression.Rule: got %q, want %q (must NOT fall back to Finding.Rule %q)",
			f.Suppression.Rule, suppressionRule, findingRule)
	}

	if f.Suppression.Reason != suppressionReason {
		t.Errorf("Suppression.Reason: got %q, want %q", f.Suppression.Reason, suppressionReason)
	}
}

func TestSARIFWriteWithIncludeSuppressed(t *testing.T) {
	t.Parallel()

	report := NewReport(ToolInfo{Name: "tool", Version: "1.0"})
	report.AddFinding(Finding{
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "test",
		Severity:    SeverityWarning,
		Position:    Position{File: FilePath("a.go"), Line: 1},
		FixStrategy: FixStrategyNone,
		Suppression: &Suppression{
			Kind:   SuppressionInConfig,
			Rule:   "r1",
			Reason: "config override",
		},
	})

	var buf strings.Builder

	err := report.WriteSARIFWithOpts(context.Background(), &buf, WithIncludeSuppressed())
	if err != nil {
		t.Fatalf("WriteSARIFWithOpts: %v", err)
	}

	findings, err := FindingsFromSARIF(context.Background(), []byte(buf.String()))
	if err != nil {
		t.Fatalf("FindingsFromSARIF: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Suppression == nil {
		t.Fatal("expected suppression to be preserved")
	}

	if findings[0].Suppression.Kind != SuppressionInConfig {
		t.Errorf("expected kind %q, got %q", SuppressionInConfig, findings[0].Suppression.Kind)
	}
}
