package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
)

// findingWithRange creates a Finding with a Range in the same file.
func findingWithRange(id string, file string, line, startLine, endLine int) finding.Finding {
	return finding.Finding{
		ID: id,
		Position: finding.Position{File: file, Line: line},
		Range: &finding.Range{
			Start: finding.Position{File: file, Line: startLine},
			End:   finding.Position{File: file, Line: endLine},
		},
	}
}

func TestConflictDetectorDetectConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		fixes           []finding.Finding
		expectedGroups  int
		expectedConflicts int
	}{
		{
			name: "no conflicts - separate files",
			fixes: []finding.Finding{
				{ID: "1", Position: finding.Position{File: "a.go", Line: 10}},
				{ID: "2", Position: finding.Position{File: "b.go", Line: 10}},
			},
			expectedGroups:  2,
			expectedConflicts: 0,
		},
		{
			name: "no conflicts - same file, different lines",
			fixes: []finding.Finding{
				{ID: "1", Position: finding.Position{File: "a.go", Line: 10}},
				{ID: "2", Position: finding.Position{File: "a.go", Line: 20}},
			},
			expectedGroups:  2,
			expectedConflicts: 0,
		},
		{
			name: "conflict - overlapping ranges",
			fixes: []finding.Finding{
				findingWithRange("1", "a.go", 10, 10, 20),
				findingWithRange("2", "a.go", 15, 15, 25),
			},
			expectedGroups:  1,
			expectedConflicts: 1,
		},
		{
			name: "conflict - adjacent ranges",
			fixes: []finding.Finding{
				findingWithRange("1", "a.go", 10, 10, 20),
				findingWithRange("2", "a.go", 20, 20, 30),
			},
			expectedGroups:  1,
			expectedConflicts: 1,
		},
		{
			name: "empty fixes",
			fixes: []finding.Finding{},
			expectedGroups:  0,
			expectedConflicts: 0,
		},
		{
			name: "single fix",
			fixes: []finding.Finding{
				{ID: "1", Position: finding.Position{File: "a.go", Line: 10}},
			},
			expectedGroups:  1,
			expectedConflicts: 0,
		},
		{
			name: "fix without file skipped",
			fixes: []finding.Finding{
				{ID: "1", Position: finding.Position{Line: 10}},
			},
			expectedGroups:  0,
			expectedConflicts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			detector := NewConflictDetector()
			groups, conflicts := detector.DetectConflicts(tt.fixes)

			if len(groups) != tt.expectedGroups {
				t.Errorf("len(groups) = %d, want %d", len(groups), tt.expectedGroups)
			}

			if len(conflicts) != tt.expectedConflicts {
				t.Errorf("len(conflicts) = %d, want %d", len(conflicts), tt.expectedConflicts)
			}
		})
	}
}

func TestHasConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fixes    []finding.Finding
		expected bool
	}{
		{
			name:     "no conflicts",
			fixes:    []finding.Finding{
				{ID: "1", Position: finding.Position{File: "a.go", Line: 10}},
				{ID: "2", Position: finding.Position{File: "a.go", Line: 20}},
			},
			expected: false,
		},
		{
			name:     "has conflicts",
			fixes: []finding.Finding{
				findingWithRange("1", "a.go", 10, 10, 20),
				findingWithRange("2", "a.go", 15, 15, 25),
			},
			expected: true,
		},
		{
			name:     "empty",
			fixes:    []finding.Finding{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := HasConflicts(tt.fixes)
			if got != tt.expected {
				t.Errorf("HasConflicts() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFilterConflictingFixes(t *testing.T) {
	t.Parallel()

	// Create fixes with overlapping ranges (need File in Range.Start for overlap detection)
	fixes := []finding.Finding{
		findingWithRange("1", "a.go", 10, 10, 15),
		findingWithRange("2", "a.go", 14, 14, 20), // Overlaps with 1
		{ID: "3", Position: finding.Position{File: "a.go", Line: 30}}, // No conflict
	}

	result := FilterConflictingFixes(fixes)

	// Should get 2 non-conflicting fixes (1 and 3, or 2 and 3)
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

	// ID 3 should always be in the result (never conflicts)
	foundID3 := false
	for _, f := range result {
		if f.ID == "3" {
			foundID3 = true
			break
		}
	}
	if !foundID3 {
		t.Error("expected ID 3 to be in result")
	}
}

func TestGroupFixesByConflict(t *testing.T) {
	t.Parallel()

	fixes := []finding.Finding{
		{ID: "1", Position: finding.Position{File: "a.go", Line: 10}},
		{ID: "2", Position: finding.Position{File: "a.go", Line: 20}},
		{ID: "3", Position: finding.Position{File: "b.go", Line: 5}},
	}

	groups := GroupFixesByConflict(fixes)

	// Should get 3 groups (all non-conflicting)
	if len(groups) != 3 {
		t.Errorf("len(groups) = %d, want 3", len(groups))
	}

	// Verify groups are sorted by file
	if groups[0][0].Position.File != "a.go" || groups[1][0].Position.File != "a.go" {
		t.Error("expected a.go fixes to be grouped first")
	}
}

func TestPositionLess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a, b     finding.Position
		expected bool
	}{
		{
			name:     "different lines",
			a:        finding.Position{Line: 10},
			b:        finding.Position{Line: 20},
			expected: true,
		},
		{
			name:     "same line, different columns",
			a:        finding.Position{Line: 10, Column: 5},
			b:        finding.Position{Line: 10, Column: 10},
			expected: true,
		},
		{
			name:     "same line and column, different offset",
			a:        finding.Position{Line: 10, Column: 5, Offset: 100},
			b:        finding.Position{Line: 10, Column: 5, Offset: 200},
			expected: true,
		},
		{
			name:     "equal positions",
			a:        finding.Position{Line: 10, Column: 5},
			b:        finding.Position{Line: 10, Column: 5},
			expected: false,
		},
		{
			name:     "greater line",
			a:        finding.Position{Line: 20},
			b:        finding.Position{Line: 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := positionLess(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("positionLess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtendRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		r1, r2   finding.Range
		expected finding.Range
	}{
		{
			name: "r2 extends r1",
			r1:   finding.Range{Start: finding.Position{Line: 10}, End: finding.Position{Line: 20}},
			r2:   finding.Range{Start: finding.Position{Line: 15}, End: finding.Position{Line: 30}},
			expected: finding.Range{Start: finding.Position{Line: 10}, End: finding.Position{Line: 30}},
		},
		{
			name: "r2 contained in r1",
			r1:   finding.Range{Start: finding.Position{Line: 10}, End: finding.Position{Line: 30}},
			r2:   finding.Range{Start: finding.Position{Line: 15}, End: finding.Position{Line: 20}},
			expected: finding.Range{Start: finding.Position{Line: 10}, End: finding.Position{Line: 30}},
		},
		{
			name: "single position ranges",
			r1:   finding.Range{Start: finding.Position{Line: 10}},
			r2:   finding.Range{Start: finding.Position{Line: 20}},
			expected: finding.Range{Start: finding.Position{Line: 10}, End: finding.Position{Line: 20}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extendRange(tt.r1, tt.r2)
			if got.Start.Line != tt.expected.Start.Line || got.End.Line != tt.expected.End.Line {
				t.Errorf("extendRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}
