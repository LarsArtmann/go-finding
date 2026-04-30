package pipeline

import (
	"testing"

	"github.com/larsartmann/go-finding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixEngine_Apply_EmptyInput(t *testing.T) {
	t.Parallel()

	engine := NewFixEngine()

	lines, applied := engine.Apply(nil, nil)
	assert.Nil(t, lines)
	assert.Equal(t, 0, applied)

	lines, applied = engine.Apply([]string{}, nil)
	assert.Empty(t, lines)
	assert.Equal(t, 0, applied)
}

func TestFixEngine_Apply_NoMatchingFixes(t *testing.T) {
	t.Parallel()

	engine := NewFixEngine()
	lines := []string{
		"package main",
		"",
		"func main() {}",
	}

	fixes := []finding.Finding{
		{
			BeforeCode: "nonexistent", AfterCode: "replacement",
			Position: finding.Position{File: "a.go", Line: 1},
		},
	}

	result, applied := engine.Apply(lines, fixes)
	assert.Equal(t, lines, result)
	assert.Equal(t, 0, applied)
}

func TestFixEngine_Apply_FixWithNoCode(t *testing.T) {
	t.Parallel()

	engine := NewFixEngine()
	lines := []string{"package main"}

	fixes := []finding.Finding{
		{Position: finding.Position{File: "a.go", Line: 1}},
	}

	result, applied := engine.Apply(lines, fixes)
	assert.Equal(t, lines, result)
	assert.Equal(t, 0, applied)
}

func TestPartitionFixes(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		r, s := partitionFixes(nil)
		assert.Nil(t, r)
		assert.Nil(t, s)
	})

	t.Run("range fix with valid end", func(t *testing.T) {
		t.Parallel()
		fixes := []finding.Finding{
			{
				BeforeCode: "old", AfterCode: "new",
				Range:    finding.NewRangePtr("a.go", 1, 1, 3, 1),
				Position: finding.Pos("a.go", 1, 1),
			},
		}
		r, s := partitionFixes(fixes)
		assert.Len(t, r, 1)
		assert.Empty(t, s)
	})

	t.Run("string fix without range", func(t *testing.T) {
		t.Parallel()
		fixes := []finding.Finding{
			{BeforeCode: "old", AfterCode: "new", Position: finding.Pos("a.go", 1, 1)},
		}
		r, s := partitionFixes(fixes)
		assert.Empty(t, r)
		assert.Len(t, s, 1)
	})

	t.Run("fix with no code is skipped", func(t *testing.T) {
		t.Parallel()
		fixes := []finding.Finding{
			{Position: finding.Pos("a.go", 1, 1)},
		}
		r, s := partitionFixes(fixes)
		assert.Empty(t, r)
		assert.Empty(t, s)
	})
}

func TestApplyRangeFixes_SingleLine(t *testing.T) {
	t.Parallel()

	lines := []string{"package main", "", "func main() {", "\told()", "}"}
	fixes := []finding.Finding{
		{
			BeforeCode: "old()", AfterCode: "new()",
			Range:    finding.NewRangePtr("a.go", 4, 2, 4, 7),
			Position: finding.Pos("a.go", 4, 2),
		},
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	require.Equal(t, 1, applied)
	assert.Equal(t, "\tnew()", result[3])
}

func TestApplyRangeFixes_MultiLine(t *testing.T) {
	t.Parallel()

	lines := []string{"package main", "", "func old() {", "\treturn", "}", "", "func main() {}"}
	fixes := []finding.Finding{
		{
			BeforeCode: "func old() {\n\treturn\n}",
			AfterCode:  "func new() {\n\treturn 42\n}",
			Range:      finding.NewRangePtr("a.go", 3, 1, 5, 1),
			Position:   finding.Pos("a.go", 3, 1),
		},
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	require.Equal(t, 1, applied)
	assert.Equal(t, "func new() {", result[2])
	assert.Equal(t, "\treturn 42", result[3])
	assert.Equal(t, "}", result[4])
}

func TestApplyRangeFixes_OutOfBounds(t *testing.T) {
	t.Parallel()

	lines := []string{"package main"}
	fixes := []finding.Finding{
		{
			BeforeCode: "old", AfterCode: "new",
			Range:    finding.NewRangePtr("a.go", 100, 1, 200, 1),
			Position: finding.Pos("a.go", 100, 1),
		},
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	assert.Equal(t, 0, applied)
	assert.Equal(t, lines, result)
}

func TestApplyRangeFixes_DescendingOrder(t *testing.T) {
	t.Parallel()

	lines := []string{
		"package main",
		"line2: old",
		"line3: old",
		"line4: old",
	}
	fixes := []finding.Finding{
		{
			BeforeCode: "old", AfterCode: "fix1",
			Range:    finding.NewRangePtr("a.go", 2, 8, 2, 11),
			Position: finding.Pos("a.go", 2, 8),
		},
		{
			BeforeCode: "old", AfterCode: "fix2",
			Range:    finding.NewRangePtr("a.go", 4, 8, 4, 11),
			Position: finding.Pos("a.go", 4, 8),
		},
	}

	result, _, applied := applyRangeFixes(lines, fixes)
	require.Equal(t, 2, applied)
	assert.Equal(t, "line2: fix1", result[1])
	assert.Equal(t, "line4: fix2", result[3])
}

func TestApplyStringFixes_ReplaceFirst(t *testing.T) {
	t.Parallel()

	lines := []string{"old code here"}
	fixes := []finding.Finding{
		{BeforeCode: "old", AfterCode: "new", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	require.Equal(t, 1, applied)
	assert.Equal(t, "new code here", result[0])
}

func TestApplyStringFixes_Insertion(t *testing.T) {
	t.Parallel()

	lines := []string{"package main", "", "func main() {}"}
	fixes := []finding.Finding{
		{AfterCode: "\tinserted", Position: finding.Pos("a.go", 3, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	require.Equal(t, 1, applied)
	require.Len(t, result, 4)
	assert.Equal(t, "\tinserted", result[2])
}

func TestApplyStringFixes_NotFound(t *testing.T) {
	t.Parallel()

	lines := []string{"package main"}
	fixes := []finding.Finding{
		{BeforeCode: "nonexistent", AfterCode: "replacement", Position: finding.Pos("a.go", 1, 1)},
	}

	result, _, applied := applyStringFixes(lines, fixes)
	assert.Equal(t, 0, applied)
	assert.Equal(t, lines, result)
}

func TestReplaceNearestToLine(t *testing.T) {
	t.Parallel()

	content := "line1: X\nline2: X\nline3: X"

	t.Run("first occurrence when targetLine is 0", func(t *testing.T) {
		t.Parallel()
		result := replaceNearestToLine(content, "X", "Y", 0)
		assert.Equal(t, "line1: Y\nline2: X\nline3: X", result)
	})

	t.Run("nearest to target line", func(t *testing.T) {
		t.Parallel()
		result := replaceNearestToLine(content, "X", "Y", 3)
		assert.Equal(t, "line1: X\nline2: X\nline3: Y", result)
	})

	t.Run("not found returns unchanged", func(t *testing.T) {
		t.Parallel()
		result := replaceNearestToLine(content, "Z", "Y", 1)
		assert.Equal(t, content, result)
	})
}

func TestLineDistance(t *testing.T) {
	t.Parallel()

	content := "a\nb\nc"

	assert.Equal(t, 0, lineDistance(content, 0, 1))
	assert.Equal(t, 1, lineDistance(content, 0, 2))
	assert.Equal(t, 1, lineDistance(content, 2, 1))
	assert.Equal(t, 2, lineDistance(content, 0, 3))
}
