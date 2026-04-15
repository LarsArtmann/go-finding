package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
)

func findingWithRange(id string, file string, line, startLine, endLine int) finding.Finding {
	return finding.Finding{
		ID:       id,
		Position: finding.Position{File: file, Line: line},
		Range: &finding.Range{
			Start: finding.Position{File: file, Line: startLine},
			End:   finding.Position{File: file, Line: endLine},
		},
	}
}

func findingAt(id, file string, line int) finding.Finding {
	return finding.Finding{ID: id, Position: finding.Position{File: file, Line: line}}
}

func findings(fixSpecs ...any) []finding.Finding {
	var fixes []finding.Finding
	for i := 0; i < len(fixSpecs); i += 3 {
		fixes = append(fixes, findingAt(
			fixSpecs[i].(string),
			fixSpecs[i+1].(string),
			fixSpecs[i+2].(int),
		))
	}
	return fixes
}

func TestConflictDetectorDetectConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fixes            []finding.Finding
		expectedGroups    int
		expectedConflicts int
	}{
		{"no conflicts - separate files", findings("1", "a.go", 10, "2", "b.go", 10), 2, 0},
		{"no conflicts - same file, different lines", findings("1", "a.go", 10, "2", "a.go", 20), 2, 0},
		{"conflict - overlapping ranges", []finding.Finding{findingWithRange("1", "a.go", 10, 10, 20), findingWithRange("2", "a.go", 15, 15, 25)}, 1, 1},
		{"conflict - adjacent ranges", []finding.Finding{findingWithRange("1", "a.go", 10, 10, 20), findingWithRange("2", "a.go", 20, 20, 30)}, 1, 1},
		{"empty fixes", []finding.Finding{}, 0, 0},
		{"single fix", findings("1", "a.go", 10), 1, 0},
		{"fix without file skipped", []finding.Finding{{ID: "1", Position: finding.Position{Line: 10}}}, 0, 0},
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
		{"no conflicts", findings("1", "a.go", 10, "2", "a.go", 20), false},
		{"has conflicts", []finding.Finding{findingWithRange("1", "a.go", 10, 10, 20), findingWithRange("2", "a.go", 15, 15, 25)}, true},
		{"empty", []finding.Finding{}, false},
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

	fixes := []finding.Finding{
		findingWithRange("1", "a.go", 10, 10, 15),
		findingWithRange("2", "a.go", 14, 14, 20),
		findingAt("3", "a.go", 30),
	}

	result := FilterConflictingFixes(fixes)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

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

	fixes := findings("1", "a.go", 10, "2", "a.go", 20, "3", "b.go", 5)

	groups := GroupFixesByConflict(fixes)

	if len(groups) != 3 {
		t.Errorf("len(groups) = %d, want 3", len(groups))
	}

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
		{"different lines", finding.Position{Line: 10}, finding.Position{Line: 20}, true},
		{"same line, different columns", finding.Position{Line: 10, Column: 5}, finding.Position{Line: 10, Column: 10}, true},
		{"same line and column, different offset", finding.Position{Line: 10, Column: 5, Offset: 100}, finding.Position{Line: 10, Column: 5, Offset: 200}, true},
		{"equal positions", finding.Position{Line: 10, Column: 5}, finding.Position{Line: 10, Column: 5}, false},
		{"greater line", finding.Position{Line: 20}, finding.Position{Line: 10}, false},
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

	makeRange := func(startLine, endLine int) finding.Range {
		return finding.Range{Start: finding.Position{Line: startLine}, End: finding.Position{Line: endLine}}
	}

	tests := []struct {
		name     string
		r1, r2   finding.Range
		expected finding.Range
	}{
		{"r2 extends r1", makeRange(10, 20), makeRange(15, 30), makeRange(10, 30)},
		{"r2 contained in r1", makeRange(10, 30), makeRange(15, 20), makeRange(10, 30)},
		{"single position ranges", finding.Range{Start: finding.Position{Line: 10}}, finding.Range{Start: finding.Position{Line: 20}}, makeRange(10, 20)},
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
