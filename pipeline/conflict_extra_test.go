package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

func TestFilterConflictingFixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fixes    []finding.Finding
		expected int
	}{
		{
			name:     "empty input",
			fixes:    nil,
			expected: 0,
		},
		{
			name:     "single fix",
			fixes:    []finding.Finding{findingAt("1", "a.go", 10)},
			expected: 1,
		},
		{
			name:     "no conflicts - different files",
			fixes:    findings("1", "a.go", 10, "2", "b.go", 20),
			expected: 2,
		},
		{
			name:     "no conflicts - same file, different lines",
			fixes:    findings("1", "a.go", 10, "2", "a.go", 50),
			expected: 2,
		},
		{
			name:     "conflict - overlapping ranges keeps first",
			fixes:    overlappingFindings(),
			expected: 1,
		},
		{
			name: "three fixes with one conflict",
			fixes: []finding.Finding{
				findingWithRange("1", "a.go", 10, 10, 20),
				findingWithRange("2", "a.go", 15, 15, 25),
				findingAt("3", "b.go", 30),
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FilterConflictingFixes(tt.fixes)
			if len(result) != tt.expected {
				t.Errorf("FilterConflictingFixes() returned %d, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestAnalyzeConflicts(t *testing.T) {
	t.Parallel()

	t.Run("no conflicts returns empty", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		result := AnalyzeConflicts(findings("1", "a.go", 10, "2", "b.go", 20))
		g.Expect(result).To(BeEmpty())
	})

	t.Run("overlapping fixes detected", func(t *testing.T) {
		t.Parallel()

		fixes := []finding.Finding{
			findingWithRange("1", "a.go", 10, 10, 20),
			findingWithRange("2", "a.go", 15, 15, 25),
		}

		result := AnalyzeConflicts(fixes)
		if len(result) != 1 {
			t.Fatalf("expected 1 conflict, got %d", len(result))
		}

		if result[0].Finding.ID != "2" {
			t.Errorf("conflicting finding ID = %q, want %q", result[0].Finding.ID, "2")
		}

		if result[0].Reason != "overlapping range" {
			t.Errorf("reason = %q, want %q", result[0].Reason, "overlapping range")
		}

		if len(result[0].ConflictsWith) == 0 {
			t.Error("expected ConflictsWith to be populated")
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		result := AnalyzeConflicts(nil)
		g.Expect(result).To(BeEmpty())
	})
}

func TestGetFindingRange(t *testing.T) {
	t.Parallel()

	t.Run("with explicit range", func(t *testing.T) {
		t.Parallel()

		f := findingWithRange("1", "a.go", 10, 10, 20)

		r := getFindingRange(f)
		if r.Start.Line != 10 || r.End.Line != 20 {
			t.Errorf("range = %v-%v, want 10-20", r.Start.Line, r.End.Line)
		}
	})

	t.Run("without range falls back to position", func(t *testing.T) {
		t.Parallel()

		f := findingAt("1", "a.go", 10)

		r := getFindingRange(f)
		if r.Start.Line != 10 {
			t.Errorf("start line = %d, want 10", r.Start.Line)
		}

		if r.End.Line != 0 {
			t.Errorf("end line = %d, want 0 (empty)", r.End.Line)
		}
	})
}

func overlappingAB() []finding.Finding {
	return []finding.Finding{
		findingWithRange("A", "a.go", 10, 10, 15),
		findingWithRange("B", "a.go", 13, 13, 18),
	}
}

func TestConflictDetection_PerMemberCheck(t *testing.T) {
	t.Parallel()

	result := FilterConflictingFixes(append(
		overlappingAB(),
		findingWithRange("C", "a.go", 16, 16, 20),
	))

	if len(result) != 1 {
		t.Errorf("expected 1 non-conflicting fix, got %d", len(result))
	}

	if len(result) > 0 && result[0].ID != "A" {
		t.Errorf("expected A to survive, got %s", result[0].ID)
	}
}

func TestConflictDetection_NonAdjacentNotGrouped(t *testing.T) {
	t.Parallel()

	result := FilterConflictingFixes(append(
		overlappingAB(),
		findingWithRange("D", "a.go", 50, 50, 60),
	))

	if len(result) != 2 {
		t.Errorf("expected 2 non-conflicting fixes (A, D), got %d", len(result))
	}

	ids := make(map[string]bool)
	for _, f := range result {
		ids[f.ID] = true
	}

	if !ids["D"] {
		t.Error("fix D should survive — it's far from A/B")
	}
}
