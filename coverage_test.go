package finding

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func expiredSuppression() *Suppression {
	return &Suppression{
		Kind:      SuppressionInSource,
		Rule:      "R1",
		ExpiresAt: new(time.Now().Add(-time.Hour)),
	}
}

func runIsValidTests(t *testing.T, tests []struct {
	name string
	f    Finding
	want bool
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.f.IsValid()).To(Equal(tt.want))
		})
	}
}

func TestFindingIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{
			name: "valid finding",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: benchFile, Line: 1},
			},
			want: true,
		},
		{
			name: "missing ID",
			f: Finding{
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: benchFile, Line: 1},
			},
			want: false,
		},
		{
			name: "missing Rule",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: benchFile, Line: 1},
			},
			want: false,
		},
		{
			name: "missing ToolName",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{File: benchFile, Line: 1},
			},
			want: false,
		},
		{
			name: "missing Message",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Severity: SeverityError,
				Position: Position{File: benchFile, Line: 1},
			},
			want: false,
		},
		{
			name: "invalid position",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: SeverityError,
				Position: Position{},
			},
			want: false,
		},
		{
			name: "invalid severity",
			f: Finding{
				ID:       "tool:rule:file.go:1:1",
				Rule:     "rule",
				ToolName: "tool",
				Message:  "msg",
				Severity: Severity("bogus"),
				Position: Position{File: benchFile, Line: 1},
			},
			want: false,
		},
		{
			name: "empty finding",
			f:    Finding{},
			want: false,
		},
	}

	runIsValidTests(t, tests)
}

func TestFindingHasFix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"direct", Finding{FixStrategy: FixStrategyDirect, AfterCode: exportTestFixed}, true},
		{"direct-no-code", Finding{FixStrategy: FixStrategyDirect}, false},
		{"ai-no-code", Finding{FixStrategy: FixStrategyAI}, false},
		{"ai-with-code", Finding{FixStrategy: FixStrategyAI, AfterCode: exportTestFixed}, true},
		{"none", Finding{FixStrategy: FixStrategyNone}, false},
		{"suggest-no-code", Finding{FixStrategy: FixStrategySuggest}, false},
		{
			"suggest-with-code",
			Finding{FixStrategy: FixStrategySuggest, AfterCode: exportTestFixed},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.f.HasFix()).To(Equal(tt.want))
		})
	}
}

func TestFindingHasSuggestion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"with suggestion text", Finding{Suggestion: exportTestFixIt}, true},
		{
			"with before and after code",
			Finding{BeforeCode: exportTestOld, AfterCode: exportTestNew},
			true,
		},
		{"with only before code", Finding{BeforeCode: "old"}, false},
		{"with only after code", Finding{AfterCode: "new"}, false},
		{"empty", Finding{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.f.HasSuggestion()).To(Equal(tt.want))
		})
	}
}

func TestFindingIsSuppressed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"no suppression", Finding{}, false},
		{"nil suppression", Finding{Suppression: nil}, false},
		{
			name: "active suppression",
			f:    Finding{Suppression: &Suppression{Kind: SuppressionInSource, Rule: "R1"}},
			want: true,
		},
		{
			name: "expired suppression",
			f:    Finding{Suppression: expiredSuppression()},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.f.IsSuppressed()).To(Equal(tt.want))
		})
	}
}

func TestRelatedRefIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    RelatedRef
		want bool
	}{
		{"with finding ID", RelatedRef{FindingID: "abc123", Relation: RelationRelated}, true},
		{"empty finding ID", RelatedRef{FindingID: ""}, false},
		{
			"with relation and position",
			RelatedRef{
				FindingID: "abc",
				Relation:  exportTestCloneOf,
				Position:  Position{File: exportTestFileF, Line: 1},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.r.IsValid()).To(Equal(tt.want))
		})
	}
}

func TestSuppressionIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *Suppression
		want bool
	}{
		{"valid in-source", &Suppression{Kind: SuppressionInSource, Rule: benchRule}, true},
		{"valid in-config", &Suppression{Kind: SuppressionInConfig, Rule: "unused"}, true},
		{"valid in-review", &Suppression{Kind: SuppressionInReview, Rule: "dup"}, true},
		{"missing kind", &Suppression{Rule: benchRule}, false},
		{"missing rule", &Suppression{Kind: SuppressionInSource}, false},
		{"nil suppression", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.s.IsValid()).To(Equal(tt.want))
		})
	}
}

func TestSuppressionIsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    *Suppression
		want bool
	}{
		{"no expiry", &Suppression{Kind: SuppressionInSource, Rule: "R1"}, false},
		{
			"future expiry",
			&Suppression{
				Kind:      SuppressionInSource,
				Rule:      "R1",
				ExpiresAt: new(time.Now().Add(time.Hour)),
			},
			false,
		},
		{
			name: "past expiry",
			s:    expiredSuppression(),
			want: true,
		},
		{"nil suppression", nil, false},
	}

	now := time.Now()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			g.Expect(tt.s.IsExpired(now)).To(Equal(tt.want))
		})
	}
}
